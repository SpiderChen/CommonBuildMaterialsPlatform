package ops

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
)

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
		if assignment.Status == "assigned" || assignment.Status == "downloaded" {
			result.PendingUpdateRollouts++
		}
	}
	return result
}

func RenewCustomer(data *AppData, customerID int64, req RenewLicenseRequest) (LicenseRenewal, error) {
	if strings.TrimSpace(req.ExpiresAt) == "" {
		return LicenseRenewal{}, fmt.Errorf("%w: expiresAt is required", ErrBadRequest)
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
		MaxSites: customer.MaxSites, MaxVehicles: customer.MaxVehicles, Status: "completed",
		Operator: fallback(req.Operator, "ops"), Note: req.Note, CreatedAt: now,
	}
	data.Renewals = append([]LicenseRenewal{renewal}, data.Renewals...)
	appendAudit(data, renewal.Operator, "license.renewed", customer.CustomerName, fmt.Sprintf("%s -> %s", oldExpiresAt, customer.ExpiresAt))
	return renewal, nil
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

func CreatePackage(data *AppData, req CreateUpdatePackageRequest) (UpdatePackage, error) {
	req.Target = strings.ToLower(strings.TrimSpace(req.Target))
	if req.Target != "client" && req.Target != "server" {
		return UpdatePackage{}, fmt.Errorf("%w: target must be client or server", ErrBadRequest)
	}
	if strings.TrimSpace(req.Version) == "" || strings.TrimSpace(req.FileName) == "" {
		return UpdatePackage{}, fmt.Errorf("%w: version and fileName are required", ErrBadRequest)
	}
	id := nextID(data, "package")
	pkg := UpdatePackage{
		ID: id, PackageNo: number("UP", id), Target: req.Target,
		ProductName: fallback(req.ProductName, "CommonBuildMaterialsPlatform"),
		Version: strings.TrimSpace(req.Version), Channel: fallback(req.Channel, "stable"), Status: "staged",
		FileName: strings.TrimSpace(req.FileName), Checksum: strings.TrimSpace(req.Checksum),
		MinVersion: fallback(req.MinVersion, "1.0.0"), RolloutPct: clamp(req.RolloutPct, 0, 100),
		ReleaseNotes: strings.TrimSpace(req.ReleaseNotes), UploadedAt: nowString(),
	}
	data.UpdatePackages = append([]UpdatePackage{pkg}, data.UpdatePackages...)
	appendAudit(data, "ops", "package.created", pkg.PackageNo, pkg.Target+" "+pkg.Version)
	return pkg, nil
}

func PublishPackage(data *AppData, packageID int64) (UpdatePackage, error) {
	index := findPackageIndex(*data, packageID)
	if index < 0 {
		return UpdatePackage{}, ErrNotFound
	}
	pkg := &data.UpdatePackages[index]
	pkg.Status = "published"
	pkg.PublishedAt = nowString()
	appendAudit(data, "ops", "package.published", pkg.PackageNo, pkg.Target+" "+pkg.Version)
	return *pkg, nil
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
		item := UpdateAssignment{ID: id, PackageID: packageID, CustomerID: customerID, Status: "assigned", AssignedAt: now}
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

func findAlertIndex(data AppData, id int64) int {
	for i, item := range data.Alerts {
		if item.ID == id {
			return i
		}
	}
	return -1
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

func normalizeCustomerHealth(value string) string {
	if value == "critical" {
		return "degraded"
	}
	return fallback(value, "healthy")
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
