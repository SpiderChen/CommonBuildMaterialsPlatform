package ops

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func setTestLicenseIssuer(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	seed := sha256.Sum256([]byte("cbmp-ops-test-license-issuer"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	t.Setenv("CBM_OPS_LICENSE_ISSUER_PRIVATE_KEY", "ed25519:"+base64.RawStdEncoding.EncodeToString(privateKey))
	return privateKey
}

func setTestUpdateSigning(t *testing.T) {
	t.Helper()
	t.Setenv("CBM_OPS_UPDATE_SIGNING_SECRET", "cbmp-ops-test-update-signing-secret")
}

func testUpdateArtifactBase64(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func TestCustomerDeploymentCreationStartsOperationsFlow(t *testing.T) {
	setTestLicenseIssuer(t)
	setTestUpdateSigning(t)
	data := EmptyData()
	customer, err := CreateCustomerDeployment(&data, CreateCustomerDeploymentRequest{
		CustomerName: "真实客户 A", LicenseID: "CBMP-REAL-001", ExpiresAt: "2027-12-31",
		ServerEndpoint: "https://erp.customer-a.example.com", ContactName: "王工", ContactPhone: "13800138000",
		UpdaterToken: "real-updater-token", MaxSites: 3, MaxVehicles: 80, Operator: "ops",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	if customer.ID == 0 || customer.RenewalStatus != "active" || customer.HealthStatus != "healthy" {
		t.Fatalf("unexpected customer defaults: %+v", customer)
	}
	if len(data.Customers) != 1 || data.Customers[0].LicenseID != "CBMP-REAL-001" {
		t.Fatalf("customer was not persisted: %+v", data.Customers)
	}

	if _, err := CreateCustomerDeployment(&data, CreateCustomerDeploymentRequest{
		CustomerName: "重复授权客户", LicenseID: "CBMP-REAL-001", ExpiresAt: "2027-12-31",
	}); err == nil {
		t.Fatalf("duplicate license should be rejected")
	}

	if _, err := RenewCustomer(&data, customer.ID, RenewLicenseRequest{ExpiresAt: "2028-12-31", Operator: "ops"}); err != nil {
		t.Fatalf("renew created customer: %v", err)
	}
	pkg, err := CreatePackage(&data, CreateUpdatePackageRequest{
		Target: "client", Version: "1.5.0", FileName: "cbmp-client-1.5.0.exe",
		ArtifactContentBase64: testUpdateArtifactBase64("cbmp client 1.5.0 artifact"),
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	if _, err := PublishPackage(&data, pkg.ID); err != nil {
		t.Fatalf("publish package: %v", err)
	}
	assignments, err := AssignPackage(&data, pkg.ID, AssignUpdatePackageRequest{CustomerIDs: []int64{customer.ID}})
	if err != nil {
		t.Fatalf("assign package: %v", err)
	}
	if len(assignments) != 1 || assignments[0].CustomerID != customer.ID {
		t.Fatalf("unexpected assignments: %+v", assignments)
	}
	poll, err := PollUpdateTasks(&data, UpdateTaskPollRequest{UpdaterToken: "real-updater-token"})
	if err != nil {
		t.Fatalf("poll update tasks: %v", err)
	}
	if !poll.Accepted || len(poll.Tasks) != 1 || poll.Tasks[0].DownloadURL == "" {
		t.Fatalf("unexpected poll response: %+v", poll)
	}
	if _, err := PollUpdateTasks(&data, UpdateTaskPollRequest{UpdaterToken: "CBMP-REAL-001"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("license id must not authenticate updater poll, got %v", err)
	}
	download, err := DownloadAssignedPackage(&data, pkg.ID, assignments[0].ID, "real-updater-token")
	if err != nil {
		t.Fatalf("download assigned package: %v", err)
	}
	if _, err := DownloadAssignedPackage(&data, pkg.ID, assignments[0].ID, "CBMP-REAL-001"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("license id must not authenticate package download, got %v", err)
	}
	if !download.Verified || download.Package.ID != pkg.ID || download.ArtifactContentBase64 == "" {
		t.Fatalf("unexpected download envelope: %+v", download)
	}
	if data.Assignments[findAssignmentIndex(data, assignments[0].ID)].Status != "running" || data.UpdatePackages[findPackageIndex(data, pkg.ID)].DownloadCount != 1 {
		t.Fatalf("download did not update rollout state: assignments=%+v packages=%+v", data.Assignments, data.UpdatePackages)
	}
	report, err := ReportUpdateTask(&data, poll.Tasks[0].TaskNo, UpdateTaskReportRequest{UpdaterToken: "real-updater-token", Status: "applied", Progress: 100, CurrentVersion: "1.5.0"})
	if err != nil {
		t.Fatalf("report update task: %v", err)
	}
	if report.Status != "applied" || data.Customers[findCustomerIndex(data, customer.ID)].CurrentClientVersion != "1.5.0" {
		t.Fatalf("report did not apply version: report=%+v customer=%+v", report, data.Customers[findCustomerIndex(data, customer.ID)])
	}
}

func TestRenewCustomerRequiresConfiguredLicenseIssuer(t *testing.T) {
	t.Setenv("CBM_OPS_LICENSE_ISSUER_PRIVATE_KEY", "")
	data := EmptyData()
	customer, err := CreateCustomerDeployment(&data, CreateCustomerDeploymentRequest{
		CustomerName: "未配置签发客户", LicenseID: "CBMP-NO-ISSUER-001", ExpiresAt: "2027-12-31",
		MaxSites: 1, MaxVehicles: 1, Operator: "ops",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	_, err = RenewCustomer(&data, customer.ID, RenewLicenseRequest{ExpiresAt: "2028-12-31", Operator: "ops"})
	if !errors.Is(err, ErrBadRequest) || !strings.Contains(err.Error(), "license issuer private key is required") {
		t.Fatalf("expected missing issuer private key rejection, got %v", err)
	}
}

func TestRenewCustomerUpdatesLicenseAndAudit(t *testing.T) {
	setTestLicenseIssuer(t)
	data := SeedData()
	renewal, err := RenewCustomer(&data, 3, RenewLicenseRequest{
		ExpiresAt: "2027-06-30", Edition: "Enterprise Appliance", MaxSites: 16, MaxVehicles: 2400, Operator: "ops", Note: "已完成年度续费",
	})
	if err != nil {
		t.Fatalf("renew customer: %v", err)
	}
	if renewal.OldExpiresAt != "2026-06-10" || renewal.NewExpiresAt != "2027-06-30" {
		t.Fatalf("unexpected renewal: %+v", renewal)
	}
	if renewal.LicensePackageNo == "" || renewal.Signature == "" || renewal.PublicKeyFingerprint == "" || len(renewal.Modules) == 0 {
		t.Fatalf("renewal did not issue a license package: %+v", renewal)
	}
	customer := data.Customers[findCustomerIndex(data, 3)]
	if customer.RenewalStatus != "active" || customer.MaxSites != 16 || customer.MaxVehicles != 2400 {
		t.Fatalf("customer was not updated: %+v", customer)
	}
	if len(data.AuditLogs) == 0 || data.AuditLogs[0].Action != "license.renewed" {
		t.Fatalf("missing audit log: %+v", data.AuditLogs)
	}
	download, err := DownloadRenewalLicensePackage(&data, renewal.ID)
	if err != nil {
		t.Fatalf("download renewal license package: %v", err)
	}
	if download.Package.LicenseID != renewal.LicenseID || download.Package.ExpiresAt != renewal.NewExpiresAt || download.ContentBase64 == "" {
		t.Fatalf("unexpected license package download: %+v", download)
	}
	verifySignedLicensePackage(t, download.Package)
	decoded, err := base64.StdEncoding.DecodeString(download.ContentBase64)
	if err != nil {
		t.Fatalf("decode license package content: %v", err)
	}
	var content LicensePackage
	if err := json.Unmarshal(decoded, &content); err != nil {
		t.Fatalf("unmarshal license package content: %v", err)
	}
	if content.LicenseID != renewal.LicenseID || content.Signature != download.Package.Signature {
		t.Fatalf("download content does not match package: %+v", content)
	}
	if data.Renewals[findRenewalIndex(data, renewal.ID)].DownloadCount != 1 || data.AuditLogs[0].Action != "license.package.downloaded" {
		t.Fatalf("download was not audited: renewals=%+v audit=%+v", data.Renewals, data.AuditLogs)
	}
}

func verifySignedLicensePackage(t *testing.T, pkg LicensePackage) {
	t.Helper()
	publicKey, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(pkg.PublicKey, "ed25519:"))
	if err != nil {
		t.Fatalf("decode public key: %v", err)
	}
	signature, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(pkg.Signature, "ed25519:"))
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	payload, err := licenseCanonicalPayload(pkg)
	if err != nil {
		t.Fatalf("canonical license payload: %v", err)
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), payload, signature) {
		t.Fatalf("license package signature is invalid")
	}
}

func TestAlertLifecycle(t *testing.T) {
	data := SeedData()
	created, err := CreateOrUpdateAlert(&data, CreateAlertRequest{
		CustomerID: 1, Source: "server", Severity: "warning", Title: "API 错误率升高", Message: "5xx 超过阈值", Assignee: "交付运维", Operator: "probe",
	})
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}
	updated, err := CreateOrUpdateAlert(&data, CreateAlertRequest{
		CustomerID: 1, Source: "server", Severity: "critical", Title: "API 错误率升高", Message: "5xx 继续升高", Operator: "probe",
	})
	if err != nil {
		t.Fatalf("update alert: %v", err)
	}
	if updated.ID != created.ID || updated.Severity != "critical" || updated.Message != "5xx 继续升高" {
		t.Fatalf("alert was not deduplicated/escalated: created=%+v updated=%+v", created, updated)
	}
	resolved, err := CreateOrUpdateAlert(&data, CreateAlertRequest{
		CustomerID: 1, Source: "server", Title: "API 错误率升高", Message: "恢复", AutoResolve: true, Resolution: "错误率恢复正常",
	})
	if err != nil {
		t.Fatalf("auto resolve alert: %v", err)
	}
	if resolved.ID != created.ID || resolved.Status != "resolved" {
		t.Fatalf("alert was not auto-resolved: %+v", resolved)
	}

	alert, err := AcknowledgeAlert(&data, 1, "交付运维")
	if err != nil {
		t.Fatalf("ack alert: %v", err)
	}
	if alert.Status != "acknowledged" || alert.AcknowledgedAt == "" {
		t.Fatalf("unexpected acknowledged alert: %+v", alert)
	}
	alert, err = ResolveAlert(&data, 1, ResolveAlertRequest{Operator: "交付运维", Resolution: "已完成服务端重启和补丁发布"})
	if err != nil {
		t.Fatalf("resolve alert: %v", err)
	}
	if alert.Status != "resolved" || alert.ResolvedAt == "" || alert.Resolution == "" {
		t.Fatalf("unexpected resolved alert: %+v", alert)
	}
}

func TestPackagePublishAndAssign(t *testing.T) {
	setTestUpdateSigning(t)
	data := SeedData()
	pkg, err := CreatePackage(&data, CreateUpdatePackageRequest{
		Target: "server", Version: "1.4.4", FileName: "cbmp-appliance-1.4.4.tar.gz", RolloutPct: 50,
		ArtifactContentBase64: testUpdateArtifactBase64("cbmp server 1.4.4 artifact"),
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	pkg, err = PublishPackage(&data, pkg.ID)
	if err != nil {
		t.Fatalf("publish package: %v", err)
	}
	if pkg.Status != "published" || pkg.PublishedAt == "" {
		t.Fatalf("unexpected package: %+v", pkg)
	}
	assignments, err := AssignPackage(&data, pkg.ID, AssignUpdatePackageRequest{CustomerIDs: []int64{1, 2}})
	if err != nil {
		t.Fatalf("assign package: %v", err)
	}
	if len(assignments) != 2 {
		t.Fatalf("expected two assignments, got %+v", assignments)
	}
}

func TestCreatePackageRequiresArtifactAndSigningSecret(t *testing.T) {
	t.Setenv("CBM_OPS_UPDATE_SIGNING_SECRET", "")
	data := EmptyData()
	if _, err := CreatePackage(&data, CreateUpdatePackageRequest{
		Target: "server", Version: "1.4.4", FileName: "cbmp-appliance-1.4.4.tar.gz",
	}); !errors.Is(err, ErrBadRequest) || !strings.Contains(err.Error(), "artifactContentBase64 is required") {
		t.Fatalf("expected missing artifact rejection, got %v", err)
	}
	if _, err := CreatePackage(&data, CreateUpdatePackageRequest{
		Target: "server", Version: "1.4.4", FileName: "cbmp-appliance-1.4.4.tar.gz",
		ArtifactContentBase64: testUpdateArtifactBase64("cbmp server 1.4.4 artifact"),
	}); !errors.Is(err, ErrBadRequest) || !strings.Contains(err.Error(), "update signing secret is required") {
		t.Fatalf("expected missing signing secret rejection, got %v", err)
	}
}
