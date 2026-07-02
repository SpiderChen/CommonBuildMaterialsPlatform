package ops

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const licenseIssuerName = "CBMP OperationsPlatform"

var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
)

type licenseSignedPayload struct {
	LicenseID    string   `json:"licenseId"`
	CustomerName string   `json:"customerName"`
	Watermark    string   `json:"watermark"`
	ExpiresAt    string   `json:"expiresAt"`
	Edition      string   `json:"edition"`
	Modules      []string `json:"modules"`
	MaxSites     int      `json:"maxSites"`
	MaxVehicles  int      `json:"maxVehicles"`
	IssuedAt     string   `json:"issuedAt"`
	Issuer       string   `json:"issuer"`
}

func Summary(data AppData) OperationsSummary {
	now := time.Now()
	result := OperationsSummary{}
	for _, customer := range data.Customers {
		if customer.RenewalStatus == "active" || customer.RenewalStatus == "expiring" {
			result.ActiveCustomers++
		}
		if expiredOrExpiring(customer.ExpiresAt, now, 45) == "expired" {
			result.ExpiredLicenses++
		}
		if expiredOrExpiring(customer.ExpiresAt, now, 45) == "expiring" {
			result.ExpiringLicenses++
		}
	}
	for _, alert := range data.Alerts {
		if alert.Status == "open" || alert.Status == "acknowledged" {
			result.OpenAlerts++
			if alert.Severity == "critical" {
				result.CriticalAlerts++
			}
		}
	}
	for _, pkg := range data.UpdatePackages {
		if pkg.Target == "client" {
			result.ClientPackages++
		}
		if pkg.Target == "server" {
			result.ServerPackages++
		}
	}
	for _, assignment := range data.Assignments {
		if assignment.Status == "assigned" || assignment.Status == "running" || assignment.Status == "downloaded" {
			result.PendingUpdateRollouts++
		}
	}
	return result
}

func CreateCustomerDeployment(data *AppData, req CreateCustomerDeploymentRequest) (CustomerDeployment, error) {
	customerName := strings.TrimSpace(req.CustomerName)
	licenseID := strings.TrimSpace(req.LicenseID)
	expiresAt := strings.TrimSpace(req.ExpiresAt)
	if customerName == "" || licenseID == "" || expiresAt == "" {
		return CustomerDeployment{}, fmt.Errorf("%w: customerName, licenseId and expiresAt are required", ErrBadRequest)
	}
	if _, err := time.Parse("2006-01-02", expiresAt); err != nil {
		return CustomerDeployment{}, fmt.Errorf("%w: expiresAt must use YYYY-MM-DD", ErrBadRequest)
	}
	for _, customer := range data.Customers {
		if strings.EqualFold(strings.TrimSpace(customer.LicenseID), licenseID) {
			return CustomerDeployment{}, fmt.Errorf("%w: licenseId already exists", ErrBadRequest)
		}
		if strings.TrimSpace(req.UpdaterToken) != "" && strings.EqualFold(strings.TrimSpace(customer.UpdaterToken), strings.TrimSpace(req.UpdaterToken)) {
			return CustomerDeployment{}, fmt.Errorf("%w: updaterToken already exists", ErrBadRequest)
		}
	}
	id := nextID(data, "customer")
	renewalStatus := expiredOrExpiring(expiresAt, time.Now(), 45)
	if renewalStatus == "" {
		renewalStatus = "active"
	}
	item := CustomerDeployment{
		ID: id, CustomerName: customerName,
		ProductName: fallback(req.ProductName, "CommonBuildMaterialsPlatform"),
		LicenseID:   licenseID, UpdaterToken: fallback(req.UpdaterToken, fmt.Sprintf("ops-updater-%s-%d", safeTokenPart(licenseID), id)),
		Edition:        fallback(req.Edition, "Enterprise Appliance"),
		DeploymentMode: fallback(req.DeploymentMode, "private_server"),
		Environment:    fallback(req.Environment, "production"),
		ServerEndpoint: strings.TrimSpace(req.ServerEndpoint),
		ContactName:    strings.TrimSpace(req.ContactName), ContactPhone: strings.TrimSpace(req.ContactPhone),
		ExpiresAt: expiresAt, RenewalStatus: renewalStatus,
		Modules:              normalizedModules(req.Modules),
		MaxSites:             positiveOrDefault(req.MaxSites, 1),
		MaxVehicles:          positiveOrDefault(req.MaxVehicles, 1),
		CurrentClientVersion: strings.TrimSpace(req.CurrentClientVersion),
		CurrentServerVersion: strings.TrimSpace(req.CurrentServerVersion),
		TargetClientVersion:  fallback(req.TargetClientVersion, req.CurrentClientVersion),
		TargetServerVersion:  fallback(req.TargetServerVersion, req.CurrentServerVersion),
		HealthStatus:         normalizeCustomerHealth(""),
		LastHeartbeatAt:      "",
		Notes:                strings.TrimSpace(req.Notes),
	}
	data.Customers = append([]CustomerDeployment{item}, data.Customers...)
	appendAudit(data, fallback(req.Operator, "ops"), "customer.created", item.CustomerName, item.LicenseID)
	return item, nil
}

func PollUpdateTasks(data *AppData, req UpdateTaskPollRequest) (UpdateTaskPollResponse, error) {
	customerIndex := findCustomerByUpdaterToken(*data, req.UpdaterToken)
	if customerIndex < 0 {
		return UpdateTaskPollResponse{}, ErrNotFound
	}
	now := nowString()
	customer := &data.Customers[customerIndex]
	customer.LastHeartbeatAt = now
	customer.HealthStatus = normalizeCustomerHealth(customer.HealthStatus)
	tasks := []ProductSystemUpdateTask{}
	for _, assignment := range data.Assignments {
		if assignment.CustomerID != customer.ID {
			continue
		}
		pkgIndex := findPackageIndex(*data, assignment.PackageID)
		if pkgIndex < 0 {
			continue
		}
		pkg := data.UpdatePackages[pkgIndex]
		if pkg.Status != "published" {
			continue
		}
		status := fallback(assignment.Status, "assigned")
		if status != "assigned" && status != "running" && status != "downloaded" {
			continue
		}
		tasks = append(tasks, assignmentTask(*customer, assignment, pkg))
	}
	return UpdateTaskPollResponse{
		Accepted: true,
		Instance: ProductInstance{
			ID: customer.ID, CustomerName: customer.CustomerName, Watermark: customer.LicenseID,
			ClientVersion: customer.CurrentClientVersion, ServerVersion: customer.CurrentServerVersion,
			Endpoint: customer.ServerEndpoint, ProbeToken: customer.UpdaterToken,
			HealthStatus: customer.HealthStatus, LastHeartbeatAt: customer.LastHeartbeatAt,
		},
		Tasks: tasks,
	}, nil
}

func DownloadAssignedPackage(data *AppData, packageID, assignmentID int64, updaterToken string) (UpdatePackageDownload, error) {
	customerIndex := findCustomerByUpdaterToken(*data, updaterToken)
	if customerIndex < 0 {
		return UpdatePackageDownload{}, ErrNotFound
	}
	customer := &data.Customers[customerIndex]
	pkgIndex := findPackageIndex(*data, packageID)
	if pkgIndex < 0 {
		return UpdatePackageDownload{}, ErrNotFound
	}
	assignmentIndex := findAssignmentIndex(*data, assignmentID)
	if assignmentIndex < 0 {
		assignmentIndex = findAssignmentByPackageAndCustomer(*data, packageID, customer.ID)
	}
	if assignmentIndex < 0 {
		return UpdatePackageDownload{}, ErrNotFound
	}
	assignment := &data.Assignments[assignmentIndex]
	if assignment.CustomerID != customer.ID || assignment.PackageID != packageID {
		return UpdatePackageDownload{}, ErrNotFound
	}
	pkg := &data.UpdatePackages[pkgIndex]
	if pkg.Status != "published" {
		return UpdatePackageDownload{}, fmt.Errorf("%w: package is not published", ErrBadRequest)
	}
	now := nowString()
	assignment.Status = "running"
	assignment.Progress = maxInt(assignment.Progress, 40)
	assignment.Step = "downloaded"
	assignment.Message = "更新包已下载"
	assignment.UpdatedAt = now
	if assignment.DownloadedAt == "" {
		assignment.DownloadedAt = now
	}
	pkg.DownloadCount++
	pkg.LastDownloadedAt = now
	customer.LastHeartbeatAt = now
	if !opsUpdatePackageVerified(*pkg) {
		return UpdatePackageDownload{}, fmt.Errorf("%w: update package signature verification failed", ErrBadRequest)
	}
	if strings.TrimSpace(pkg.ArtifactContentBase64) == "" {
		return UpdatePackageDownload{}, fmt.Errorf("%w: update package artifact is missing", ErrBadRequest)
	}
	artifact, err := decodeUpdateArtifact(pkg.ArtifactContentBase64)
	if err != nil {
		return UpdatePackageDownload{}, fmt.Errorf("%w: update package artifact is invalid", ErrBadRequest)
	}
	artifactSHA := updateArtifactSHA256(artifact)
	if pkg.ArtifactSHA256 != "" && pkg.ArtifactSHA256 != artifactSHA {
		return UpdatePackageDownload{}, fmt.Errorf("%w: update package artifact checksum mismatch", ErrBadRequest)
	}
	return UpdatePackageDownload{
		FileName: pkg.FileName, ContentType: fallback(pkg.ArtifactContentType, "application/octet-stream"), Verified: true, GeneratedAt: now,
		ArtifactFileName: fallback(pkg.ArtifactFileName, pkg.FileName), ArtifactContentType: fallback(pkg.ArtifactContentType, "application/octet-stream"),
		ArtifactSizeBytes: int64(len(artifact)), ArtifactSHA256: artifactSHA, ArtifactContentBase64: pkg.ArtifactContentBase64,
		Manifest: map[string]string{
			"assignmentId": strconv.FormatInt(assignment.ID, 10),
			"customerId":   strconv.FormatInt(customer.ID, 10),
			"target":       pkg.Target,
			"version":      pkg.Version,
		},
		Package: sanitizeUpdatePackageForResponse(*pkg),
	}, nil
}

func ReportUpdateTask(data *AppData, taskNo string, req UpdateTaskReportRequest) (UpdateAssignment, error) {
	assignmentID, err := assignmentIDFromTaskNo(taskNo)
	if err != nil {
		return UpdateAssignment{}, fmt.Errorf("%w: invalid taskNo", ErrBadRequest)
	}
	assignmentIndex := findAssignmentIndex(*data, assignmentID)
	if assignmentIndex < 0 {
		return UpdateAssignment{}, ErrNotFound
	}
	assignment := &data.Assignments[assignmentIndex]
	customerIndex := findCustomerIndex(*data, assignment.CustomerID)
	if customerIndex < 0 || !customerMatchesUpdaterToken(data.Customers[customerIndex], req.UpdaterToken) {
		return UpdateAssignment{}, ErrNotFound
	}
	pkgIndex := findPackageIndex(*data, assignment.PackageID)
	if pkgIndex < 0 {
		return UpdateAssignment{}, ErrNotFound
	}
	status := normalizeAssignmentStatus(req.Status)
	now := nowString()
	assignment.Status = status
	assignment.Progress = clamp(req.Progress, 0, 100)
	assignment.Step = strings.TrimSpace(req.Step)
	assignment.Message = strings.TrimSpace(req.Message)
	assignment.Error = strings.TrimSpace(req.Error)
	assignment.UpdaterVersion = strings.TrimSpace(req.UpdaterVersion)
	assignment.UpdatedAt = now
	customer := &data.Customers[customerIndex]
	customer.LastHeartbeatAt = now
	pkg := data.UpdatePackages[pkgIndex]
	if status == "applied" {
		assignment.Progress = 100
		assignment.AppliedAt = now
		assignment.Error = ""
		customer.HealthStatus = "healthy"
		if strings.TrimSpace(req.CurrentVersion) != "" {
			if pkg.Target == "server" {
				customer.CurrentServerVersion = strings.TrimSpace(req.CurrentVersion)
			} else {
				customer.CurrentClientVersion = strings.TrimSpace(req.CurrentVersion)
			}
		} else if pkg.Target == "server" {
			customer.CurrentServerVersion = pkg.Version
		} else {
			customer.CurrentClientVersion = pkg.Version
		}
		appendAudit(data, "updater", "package.applied", taskNo, customer.CustomerName+" "+pkg.Target+" "+pkg.Version)
	} else if status == "failed" || status == "rolled_back" {
		assignment.Progress = 100
		customer.HealthStatus = "degraded"
		appendAudit(data, "updater", "package."+status, taskNo, fallback(assignment.Error, assignment.Message))
	}
	return *assignment, nil
}

func RenewCustomer(data *AppData, customerID int64, req RenewLicenseRequest) (LicenseRenewal, error) {
	if strings.TrimSpace(req.ExpiresAt) == "" {
		return LicenseRenewal{}, fmt.Errorf("%w: expiresAt is required", ErrBadRequest)
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(req.ExpiresAt)); err != nil {
		return LicenseRenewal{}, fmt.Errorf("%w: expiresAt must use YYYY-MM-DD", ErrBadRequest)
	}
	index := findCustomerIndex(*data, customerID)
	if index < 0 {
		return LicenseRenewal{}, ErrNotFound
	}
	now := nowString()
	customer := &data.Customers[index]
	oldExpiresAt := customer.ExpiresAt
	customer.ExpiresAt = req.ExpiresAt
	customer.RenewalStatus = "active"
	customer.HealthStatus = normalizeCustomerHealth(customer.HealthStatus)
	if strings.TrimSpace(req.Edition) != "" {
		customer.Edition = strings.TrimSpace(req.Edition)
	}
	if req.MaxSites > 0 {
		customer.MaxSites = req.MaxSites
	}
	if req.MaxVehicles > 0 {
		customer.MaxVehicles = req.MaxVehicles
	}
	if strings.TrimSpace(req.Note) != "" {
		customer.Notes = strings.TrimSpace(req.Note)
	}

	id := nextID(data, "renewal")
	renewal := LicenseRenewal{
		ID: id, RenewalNo: number("RN", id), CustomerID: customer.ID, LicenseID: customer.LicenseID,
		OldExpiresAt: oldExpiresAt, NewExpiresAt: customer.ExpiresAt, Edition: customer.Edition,
		Modules: append([]string{}, customer.Modules...), MaxSites: customer.MaxSites, MaxVehicles: customer.MaxVehicles,
		LicensePackageNo: number("LP", id), Watermark: licenseWatermark(*customer), Issuer: licenseIssuerName, IssuedAt: dateOnly(now),
		Status:   "completed",
		Operator: fallback(req.Operator, "ops"), Note: req.Note, CreatedAt: now,
	}
	pkg, err := signLicensePackageForRenewal(*customer, renewal)
	if err != nil {
		return LicenseRenewal{}, err
	}
	renewal.PublicKey = pkg.PublicKey
	renewal.PublicKeyFingerprint = pkg.PublicKeyFingerprint
	renewal.Signature = pkg.Signature
	data.Renewals = append([]LicenseRenewal{renewal}, data.Renewals...)
	appendAudit(data, renewal.Operator, "license.renewed", customer.CustomerName, fmt.Sprintf("%s -> %s", oldExpiresAt, customer.ExpiresAt))
	return renewal, nil
}

func DownloadRenewalLicensePackage(data *AppData, renewalID int64) (LicensePackageDownload, error) {
	renewalIndex := findRenewalIndex(*data, renewalID)
	if renewalIndex < 0 {
		return LicensePackageDownload{}, ErrNotFound
	}
	renewal := &data.Renewals[renewalIndex]
	customerIndex := findCustomerIndex(*data, renewal.CustomerID)
	if customerIndex < 0 {
		return LicensePackageDownload{}, ErrNotFound
	}
	customer := data.Customers[customerIndex]
	if renewal.LicensePackageNo == "" {
		renewal.LicensePackageNo = number("LP", renewal.ID)
	}
	if renewal.Watermark == "" {
		renewal.Watermark = licenseWatermark(customer)
	}
	if renewal.Issuer == "" {
		renewal.Issuer = licenseIssuerName
	}
	if renewal.IssuedAt == "" {
		renewal.IssuedAt = dateOnly(renewal.CreatedAt)
	}
	if len(renewal.Modules) == 0 {
		renewal.Modules = append([]string{}, customer.Modules...)
	}
	pkg, err := licensePackageFromRenewal(customer, *renewal)
	if err != nil || pkg.Signature == "" || pkg.PublicKey == "" {
		pkg, err = signLicensePackageForRenewal(customer, *renewal)
		if err != nil {
			return LicensePackageDownload{}, err
		}
		renewal.PublicKey = pkg.PublicKey
		renewal.PublicKeyFingerprint = pkg.PublicKeyFingerprint
		renewal.Signature = pkg.Signature
	}
	renewal.DownloadCount++
	renewal.LastDownloadedAt = nowString()
	pkg.Status = "issued"
	payload, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return LicensePackageDownload{}, err
	}
	appendAudit(data, "ops", "license.package.downloaded", renewal.RenewalNo, customer.CustomerName)
	return LicensePackageDownload{
		FileName:      licensePackageFileName(customer, *renewal),
		ContentType:   "application/json",
		GeneratedAt:   renewal.LastDownloadedAt,
		ContentBase64: base64.StdEncoding.EncodeToString(payload),
		Customer:      customer,
		Renewal:       *renewal,
		Package:       pkg,
	}, nil
}

func signLicensePackageForRenewal(customer CustomerDeployment, renewal LicenseRenewal) (LicensePackage, error) {
	publicKey, privateKey, err := licenseIssuerKeyPair()
	if err != nil {
		return LicensePackage{}, err
	}
	renewal.PublicKey = encodeLicenseKey(publicKey)
	renewal.PublicKeyFingerprint = licensePublicKeyFingerprint(renewal.PublicKey)
	pkg, err := licensePackageFromRenewal(customer, renewal)
	if err != nil {
		return LicensePackage{}, err
	}
	payload, err := licenseCanonicalPayload(pkg)
	if err != nil {
		return LicensePackage{}, err
	}
	pkg.Signature = "ed25519:" + base64.RawStdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	return pkg, nil
}

func licensePackageFromRenewal(customer CustomerDeployment, renewal LicenseRenewal) (LicensePackage, error) {
	modules := append([]string{}, renewal.Modules...)
	if len(modules) == 0 {
		modules = append([]string{}, customer.Modules...)
	}
	if len(modules) == 0 {
		return LicensePackage{}, fmt.Errorf("%w: license modules are required", ErrBadRequest)
	}
	pkg := LicensePackage{
		ID:                   renewal.ID,
		LicenseID:            fallback(renewal.LicenseID, customer.LicenseID),
		CustomerName:         customer.CustomerName,
		Watermark:            fallback(renewal.Watermark, licenseWatermark(customer)),
		ExpiresAt:            renewal.NewExpiresAt,
		Edition:              fallback(renewal.Edition, customer.Edition),
		Modules:              modules,
		MaxSites:             renewal.MaxSites,
		MaxVehicles:          renewal.MaxVehicles,
		IssuedAt:             fallback(renewal.IssuedAt, dateOnly(renewal.CreatedAt)),
		Issuer:               fallback(renewal.Issuer, licenseIssuerName),
		PublicKey:            renewal.PublicKey,
		PublicKeyFingerprint: renewal.PublicKeyFingerprint,
		Signature:            renewal.Signature,
		Status:               "issued",
	}
	if pkg.MaxSites <= 0 || pkg.MaxVehicles <= 0 {
		return LicensePackage{}, fmt.Errorf("%w: license quota must be positive", ErrBadRequest)
	}
	if pkg.PublicKey != "" && pkg.PublicKeyFingerprint == "" {
		pkg.PublicKeyFingerprint = licensePublicKeyFingerprint(pkg.PublicKey)
	}
	return pkg, nil
}

func licenseCanonicalPayload(item LicensePackage) ([]byte, error) {
	return json.Marshal(licenseSignedPayload{
		LicenseID: item.LicenseID, CustomerName: item.CustomerName, Watermark: item.Watermark,
		ExpiresAt: item.ExpiresAt, Edition: item.Edition, Modules: item.Modules,
		MaxSites: item.MaxSites, MaxVehicles: item.MaxVehicles, IssuedAt: item.IssuedAt, Issuer: item.Issuer,
	})
}

func licenseIssuerKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	value := strings.TrimPrefix(strings.TrimSpace(os.Getenv("CBM_OPS_LICENSE_ISSUER_PRIVATE_KEY")), "ed25519:")
	if value == "" {
		return nil, nil, fmt.Errorf("%w: license issuer private key is required", ErrBadRequest)
	}
	decoded, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: license issuer private key is invalid", ErrBadRequest)
	}
	switch len(decoded) {
	case ed25519.SeedSize:
		privateKey := ed25519.NewKeyFromSeed(decoded)
		return privateKey.Public().(ed25519.PublicKey), privateKey, nil
	case ed25519.PrivateKeySize:
		privateKey := ed25519.PrivateKey(decoded)
		return privateKey.Public().(ed25519.PublicKey), privateKey, nil
	default:
		return nil, nil, fmt.Errorf("%w: license issuer private key length is invalid", ErrBadRequest)
	}
}

func encodeLicenseKey(key ed25519.PublicKey) string {
	return "ed25519:" + base64.RawStdEncoding.EncodeToString(key)
}

func licensePublicKeyFingerprint(value string) string {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return ""
	}
	sum := sha256.Sum256(decoded)
	return hex.EncodeToString(sum[:])[:16]
}

func licenseWatermark(customer CustomerDeployment) string {
	return fallback(customer.UpdaterToken, customer.LicenseID)
}

func licensePackageFileName(customer CustomerDeployment, renewal LicenseRenewal) string {
	name := safeTokenPart(customer.LicenseID)
	if name == "customer" {
		name = safeTokenPart(customer.CustomerName)
	}
	return fmt.Sprintf("%s-%s-license.json", name, strings.ToLower(renewal.LicensePackageNo))
}

func dateOnly(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= len("2006-01-02") {
		return value[:len("2006-01-02")]
	}
	return time.Now().Format("2006-01-02")
}

func CreateOrUpdateAlert(data *AppData, req CreateAlertRequest) (SystemAlert, error) {
	customerID, err := alertCustomerID(*data, req)
	if err != nil {
		return SystemAlert{}, err
	}
	source := fallback(req.Source, "customer")
	severity := normalizeAlertSeverity(req.Severity)
	title := strings.TrimSpace(req.Title)
	message := strings.TrimSpace(req.Message)
	if title == "" || message == "" {
		return SystemAlert{}, fmt.Errorf("%w: title and message are required", ErrBadRequest)
	}
	now := fallback(req.OccurredAt, nowString())
	if req.AutoResolve {
		index := findOpenAlertByFingerprint(*data, customerID, source, title)
		if index < 0 {
			return SystemAlert{}, ErrNotFound
		}
		alert := &data.Alerts[index]
		alert.Status = "resolved"
		alert.ResolvedAt = now
		alert.LastSeenAt = now
		alert.Resolution = fallback(req.Resolution, "客户侧已恢复")
		appendAudit(data, fallback(req.Operator, "probe"), "alert.resolved", alert.AlertNo, alert.Resolution)
		return *alert, nil
	}
	if index := findOpenAlertByFingerprint(*data, customerID, source, title); index >= 0 {
		alert := &data.Alerts[index]
		alert.Severity = mostSevere(alert.Severity, severity)
		alert.Message = message
		alert.LastSeenAt = now
		alert.Assignee = fallback(req.Assignee, alert.Assignee)
		alert.Resolution = ""
		if alert.Status == "" || alert.Status == "resolved" {
			alert.Status = "open"
		}
		appendAudit(data, fallback(req.Operator, "probe"), "alert.updated", alert.AlertNo, alert.Title)
		return *alert, nil
	}
	id := nextID(data, "alert")
	alert := SystemAlert{
		ID: id, AlertNo: number("AL", id), CustomerID: customerID, Source: source, Severity: severity,
		Title: title, Message: message, Status: "open", FirstSeenAt: now, LastSeenAt: now,
		Assignee: strings.TrimSpace(req.Assignee),
	}
	data.Alerts = append([]SystemAlert{alert}, data.Alerts...)
	appendAudit(data, fallback(req.Operator, "probe"), "alert.created", alert.AlertNo, alert.Title)
	return alert, nil
}

func AcknowledgeAlert(data *AppData, alertID int64, operator string) (SystemAlert, error) {
	index := findAlertIndex(*data, alertID)
	if index < 0 {
		return SystemAlert{}, ErrNotFound
	}
	alert := &data.Alerts[index]
	if alert.Status == "resolved" {
		return *alert, nil
	}
	alert.Status = "acknowledged"
	alert.AcknowledgedAt = nowString()
	alert.Assignee = fallback(operator, alert.Assignee)
	appendAudit(data, fallback(operator, "ops"), "alert.acknowledged", alert.AlertNo, alert.Title)
	return *alert, nil
}

func ResolveAlert(data *AppData, alertID int64, req ResolveAlertRequest) (SystemAlert, error) {
	index := findAlertIndex(*data, alertID)
	if index < 0 {
		return SystemAlert{}, ErrNotFound
	}
	alert := &data.Alerts[index]
	alert.Status = "resolved"
	alert.ResolvedAt = nowString()
	alert.Resolution = fallback(req.Resolution, "已处理")
	alert.Assignee = fallback(req.Operator, alert.Assignee)
	appendAudit(data, fallback(req.Operator, "ops"), "alert.resolved", alert.AlertNo, alert.Resolution)
	return *alert, nil
}

func decodeUpdateArtifact(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("artifact is empty")
	}
	if out, err := base64.StdEncoding.DecodeString(value); err == nil {
		return out, nil
	}
	return base64.RawStdEncoding.DecodeString(value)
}

func updateArtifactSHA256(artifact []byte) string {
	sum := sha256.Sum256(artifact)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func strictUpdateChecksum(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	digest := strings.TrimPrefix(value, "sha256:")
	if len(digest) != 64 || !strings.HasPrefix(value, "sha256:") {
		return "", false
	}
	for _, ch := range digest {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return "", false
		}
	}
	return digest, true
}

func signOpsUpdatePackage(item UpdatePackage) (string, error) {
	secret := strings.TrimSpace(os.Getenv("CBM_OPS_UPDATE_SIGNING_SECRET"))
	if secret == "" {
		return "", fmt.Errorf("%w: update signing secret is required", ErrBadRequest)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(opsUpdatePackageSignaturePayload(item)))
	return "hmac-sha256:" + hex.EncodeToString(mac.Sum(nil)), nil
}

func opsUpdatePackageVerified(item UpdatePackage) bool {
	signature := strings.ToLower(strings.TrimSpace(item.Signature))
	if !strings.HasPrefix(signature, "hmac-sha256:") {
		return false
	}
	expected := strings.TrimPrefix(signature, "hmac-sha256:")
	if _, ok := strictUpdateChecksum("sha256:" + expected); !ok {
		return false
	}
	actual, err := signOpsUpdatePackage(item)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(strings.TrimPrefix(actual, "hmac-sha256:")))
}

func opsUpdatePackageSignaturePayload(item UpdatePackage) string {
	return strings.Join([]string{
		item.Version,
		fallback(item.Component, item.Target),
		item.Channel,
		fallback(item.PackageType, "full"),
		"",
		"",
		item.Checksum,
		item.ArtifactSHA256,
		item.ArtifactFileName,
		fmt.Sprintf("%d", item.ArtifactSizeBytes),
		"",
		"",
	}, "\n")
}

func CreatePackage(data *AppData, req CreateUpdatePackageRequest) (UpdatePackage, error) {
	req.Target = strings.ToLower(strings.TrimSpace(req.Target))
	if req.Target != "client" && req.Target != "server" {
		return UpdatePackage{}, fmt.Errorf("%w: target must be client or server", ErrBadRequest)
	}
	if strings.TrimSpace(req.Version) == "" || strings.TrimSpace(req.FileName) == "" {
		return UpdatePackage{}, fmt.Errorf("%w: version and fileName are required", ErrBadRequest)
	}
	artifact, err := decodeUpdateArtifact(req.ArtifactContentBase64)
	if err != nil || len(artifact) == 0 {
		return UpdatePackage{}, fmt.Errorf("%w: artifactContentBase64 is required", ErrBadRequest)
	}
	artifactSHA := updateArtifactSHA256(artifact)
	checksum := strings.TrimSpace(req.Checksum)
	if checksum == "" {
		checksum = artifactSHA
	}
	if expected, ok := strictUpdateChecksum(checksum); !ok || expected != strings.TrimPrefix(artifactSHA, "sha256:") {
		return UpdatePackage{}, fmt.Errorf("%w: checksum must match artifact sha256", ErrBadRequest)
	}
	id := nextID(data, "package")
	pkg := UpdatePackage{
		ID: id, PackageNo: number("UP", id), Target: req.Target, Component: req.Target,
		ProductName: fallback(req.ProductName, "CommonBuildMaterialsPlatform"),
		Version:     strings.TrimSpace(req.Version), Channel: fallback(req.Channel, "stable"), Status: "staged",
		PackageType:           "full",
		FileName:              strings.TrimSpace(req.FileName),
		Checksum:              checksum,
		Signature:             strings.TrimSpace(req.Signature),
		ArtifactFileName:      strings.TrimSpace(req.FileName),
		ArtifactContentType:   fallback(strings.TrimSpace(req.ArtifactContentType), "application/octet-stream"),
		ArtifactSizeBytes:     int64(len(artifact)),
		ArtifactSHA256:        artifactSHA,
		ArtifactContentBase64: base64.StdEncoding.EncodeToString(artifact),
		MinVersion:            fallback(req.MinVersion, "1.0.0"), RolloutPct: clamp(req.RolloutPct, 0, 100),
		ReleaseNotes: strings.TrimSpace(req.ReleaseNotes), UploadedAt: nowString(),
	}
	if pkg.Signature == "" {
		signature, err := signOpsUpdatePackage(pkg)
		if err != nil {
			return UpdatePackage{}, err
		}
		pkg.Signature = signature
	}
	if !opsUpdatePackageVerified(pkg) {
		return UpdatePackage{}, fmt.Errorf("%w: update package signature verification failed", ErrBadRequest)
	}
	data.UpdatePackages = append([]UpdatePackage{pkg}, data.UpdatePackages...)
	appendAudit(data, "ops", "package.created", pkg.PackageNo, pkg.Target+" "+pkg.Version)
	return sanitizeUpdatePackageForResponse(pkg), nil
}

func PublishPackage(data *AppData, packageID int64) (UpdatePackage, error) {
	index := findPackageIndex(*data, packageID)
	if index < 0 {
		return UpdatePackage{}, ErrNotFound
	}
	pkg := &data.UpdatePackages[index]
	if !opsUpdatePackageVerified(*pkg) {
		return UpdatePackage{}, fmt.Errorf("%w: update package signature verification failed", ErrBadRequest)
	}
	if strings.TrimSpace(pkg.ArtifactContentBase64) == "" {
		return UpdatePackage{}, fmt.Errorf("%w: update package artifact is missing", ErrBadRequest)
	}
	pkg.Status = "published"
	pkg.PublishedAt = nowString()
	appendAudit(data, "ops", "package.published", pkg.PackageNo, pkg.Target+" "+pkg.Version)
	return sanitizeUpdatePackageForResponse(*pkg), nil
}

func sanitizeUpdatePackageForResponse(item UpdatePackage) UpdatePackage {
	item.ArtifactContentBase64 = ""
	return item
}

func sanitizeUpdatePackagesForResponse(items []UpdatePackage) []UpdatePackage {
	out := make([]UpdatePackage, len(items))
	for i := range items {
		out[i] = sanitizeUpdatePackageForResponse(items[i])
	}
	return out
}

func AssignPackage(data *AppData, packageID int64, req AssignUpdatePackageRequest) ([]UpdateAssignment, error) {
	pkgIndex := findPackageIndex(*data, packageID)
	if pkgIndex < 0 {
		return nil, ErrNotFound
	}
	if len(req.CustomerIDs) == 0 {
		return nil, fmt.Errorf("%w: customerIds is required", ErrBadRequest)
	}
	now := nowString()
	created := []UpdateAssignment{}
	for _, customerID := range req.CustomerIDs {
		if findCustomerIndex(*data, customerID) < 0 || assignmentExists(*data, packageID, customerID) {
			continue
		}
		id := nextID(data, "assignment")
		item := UpdateAssignment{ID: id, PackageID: packageID, CustomerID: customerID, Status: "assigned", AssignedAt: now, Progress: 0, Step: "assigned", UpdatedAt: now}
		data.Assignments = append([]UpdateAssignment{item}, data.Assignments...)
		created = append(created, item)
	}
	appendAudit(data, "ops", "package.assigned", data.UpdatePackages[pkgIndex].PackageNo, fmt.Sprintf("%d customer(s)", len(created)))
	return created, nil
}

func SortedCustomers(data AppData) []CustomerDeployment {
	items := append([]CustomerDeployment{}, data.Customers...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].ExpiresAt < items[j].ExpiresAt
	})
	return items
}

func nextID(data *AppData, key string) int64 {
	ensureNext(data)
	data.Next[key]++
	return data.Next[key]
}

func appendAudit(data *AppData, actor, action, target, detail string) {
	id := nextID(data, "audit")
	data.AuditLogs = append([]AuditLog{{
		ID: id, Actor: actor, Action: action, Target: target, Detail: detail, CreatedAt: nowString(),
	}}, data.AuditLogs...)
}

func number(prefix string, id int64) string {
	return fmt.Sprintf("%s%012d", prefix, id)
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func expiredOrExpiring(value string, now time.Time, days int) string {
	day, err := time.Parse("2006-01-02", value)
	if err != nil {
		return ""
	}
	if day.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		return "expired"
	}
	if day.Before(now.AddDate(0, 0, days)) {
		return "expiring"
	}
	return "active"
}

func findCustomerIndex(data AppData, id int64) int {
	for i, item := range data.Customers {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func findRenewalIndex(data AppData, id int64) int {
	for i, item := range data.Renewals {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func findCustomerByUpdaterToken(data AppData, token string) int {
	for i, item := range data.Customers {
		if customerMatchesUpdaterToken(item, token) {
			return i
		}
	}
	return -1
}

func customerMatchesUpdaterToken(customer CustomerDeployment, token string) bool {
	normalized := strings.TrimSpace(token)
	return normalized != "" && normalized == strings.TrimSpace(customer.UpdaterToken)
}

func findAlertIndex(data AppData, id int64) int {
	for i, item := range data.Alerts {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func findOpenAlertByFingerprint(data AppData, customerID int64, source, title string) int {
	normalizedSource := strings.ToLower(strings.TrimSpace(source))
	normalizedTitle := strings.ToLower(strings.TrimSpace(title))
	for i, item := range data.Alerts {
		if item.CustomerID == customerID &&
			item.Status != "resolved" &&
			strings.ToLower(strings.TrimSpace(item.Source)) == normalizedSource &&
			strings.ToLower(strings.TrimSpace(item.Title)) == normalizedTitle {
			return i
		}
	}
	return -1
}

func alertCustomerID(data AppData, req CreateAlertRequest) (int64, error) {
	if req.CustomerID > 0 {
		if findCustomerIndex(data, req.CustomerID) < 0 {
			return 0, ErrNotFound
		}
		return req.CustomerID, nil
	}
	index := findCustomerByUpdaterToken(data, req.UpdaterToken)
	if index < 0 {
		return 0, ErrNotFound
	}
	return data.Customers[index].ID, nil
}

func normalizeAlertSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical", "warning", "info":
		return strings.ToLower(strings.TrimSpace(value))
	case "error", "high", "fatal":
		return "critical"
	case "warn", "medium":
		return "warning"
	default:
		return "info"
	}
}

func mostSevere(left, right string) string {
	rank := map[string]int{"info": 1, "warning": 2, "critical": 3}
	if rank[normalizeAlertSeverity(right)] > rank[normalizeAlertSeverity(left)] {
		return normalizeAlertSeverity(right)
	}
	return normalizeAlertSeverity(left)
}

func findPackageIndex(data AppData, id int64) int {
	for i, item := range data.UpdatePackages {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func assignmentExists(data AppData, packageID, customerID int64) bool {
	for _, item := range data.Assignments {
		if item.PackageID == packageID && item.CustomerID == customerID {
			return true
		}
	}
	return false
}

func findAssignmentIndex(data AppData, id int64) int {
	for i, item := range data.Assignments {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func findAssignmentByPackageAndCustomer(data AppData, packageID, customerID int64) int {
	for i, item := range data.Assignments {
		if item.PackageID == packageID && item.CustomerID == customerID {
			return i
		}
	}
	return -1
}

func assignmentTask(customer CustomerDeployment, assignment UpdateAssignment, pkg UpdatePackage) ProductSystemUpdateTask {
	fromVersion := customer.CurrentClientVersion
	if pkg.Target == "server" {
		fromVersion = customer.CurrentServerVersion
	}
	return ProductSystemUpdateTask{
		ID: assignment.ID, TaskNo: taskNoForAssignment(assignment.ID),
		ExecutionID: assignment.ID, ExecutionNo: taskNoForAssignment(assignment.ID),
		RolloutID: pkg.ID, RolloutNo: pkg.PackageNo, UpdateID: pkg.ID,
		InstanceID: customer.ID, CustomerName: customer.CustomerName, Watermark: customer.LicenseID,
		Component: pkg.Target, Version: pkg.Version, FromVersion: fromVersion, Action: "apply",
		Status: fallback(assignment.Status, "assigned"), Progress: assignment.Progress,
		ArtifactFileName: pkg.FileName, Checksum: pkg.Checksum,
		DownloadURL: fmt.Sprintf("/api/system/updates/%d/download?assignmentId=%d", pkg.ID, assignment.ID),
		CreatedAt:   assignment.AssignedAt, Remark: assignment.Message,
	}
}

func taskNoForAssignment(id int64) string {
	return number("UA", id)
}

func assignmentIDFromTaskNo(taskNo string) (int64, error) {
	value := strings.TrimPrefix(strings.TrimSpace(taskNo), "UA")
	return strconv.ParseInt(value, 10, 64)
}

func normalizeAssignmentStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "assigned":
		return "assigned"
	case "running", "downloaded":
		return "running"
	case "applied", "completed", "succeeded", "success":
		return "applied"
	case "rolled_back":
		return "rolled_back"
	case "failed", "error":
		return "failed"
	default:
		return "running"
	}
}

func safeTokenPart(value string) string {
	normalized := strings.Builder{}
	for _, ch := range strings.ToLower(value) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			normalized.WriteRune(ch)
		}
	}
	result := normalized.String()
	if len(result) > 12 {
		return result[:12]
	}
	return fallback(result, "customer")
}

func normalizeCustomerHealth(value string) string {
	if value == "critical" {
		return "degraded"
	}
	return fallback(value, "healthy")
}

func normalizedModules(items []string) []string {
	if len(items) == 0 {
		return []string{"erp"}
	}
	result := []string{}
	seen := map[string]bool{}
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	if len(result) == 0 {
		return []string{"erp"}
	}
	return result
}

func positiveOrDefault(value, fallbackValue int) int {
	if value > 0 {
		return value
	}
	return fallbackValue
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return strings.TrimSpace(value)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
