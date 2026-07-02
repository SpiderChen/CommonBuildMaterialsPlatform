package ops

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCustomersAPIAllowsCreatingDeploymentFromEmptyStore(t *testing.T) {
	t.Setenv("CBM_OPS_SEED_DEMO", "")

	app := NewApp(NewStore(filepath.Join(t.TempDir(), "ops.json")), "")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/customers", strings.NewReader(`{
		"customerName":"真实客户 API",
		"licenseId":"CBMP-API-001",
		"serverEndpoint":"https://erp.api-customer.example.com",
		"expiresAt":"2027-12-31",
		"maxSites":2,
		"maxVehicles":20
	}`))
	req.Header.Set("Content-Type", "application/json")
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create customer status %d: %s", rec.Code, rec.Body.String())
	}
	var created CustomerDeployment
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created customer: %v", err)
	}
	if created.ID == 0 || created.CustomerName != "真实客户 API" || created.LicenseID != "CBMP-API-001" {
		t.Fatalf("unexpected created customer: %+v", created)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/customers", nil)
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CBMP-API-001") {
		t.Fatalf("created customer not listed, status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdaterCompatibilityAPICompletesAssignedPackage(t *testing.T) {
	t.Setenv("CBM_OPS_SEED_DEMO", "")
	setTestUpdateSigning(t)

	app := NewApp(NewStore(filepath.Join(t.TempDir(), "ops.json")), "")
	customer := postJSONForTest[CustomerDeployment](t, app, "/api/customers", `{
		"customerName":"Updater 客户",
		"licenseId":"CBMP-UPDATER-001",
		"updaterToken":"api-updater-token",
		"expiresAt":"2027-12-31",
		"currentClientVersion":"1.0.0",
		"maxSites":1,
		"maxVehicles":1
	}`)
	pkg := postJSONForTest[UpdatePackage](t, app, "/api/update-packages", `{"target":"client","version":"1.1.0","fileName":"cbmp-client-1.1.0.json","artifactContentBase64":"`+testUpdateArtifactBase64("cbmp client 1.1.0 artifact")+`"}`)
	if pkg.ArtifactContentBase64 != "" {
		t.Fatalf("create package response must not expose artifact content")
	}
	pkg = postJSONForTest[UpdatePackage](t, app, "/api/update-packages/"+strconvFormat(pkg.ID)+"/publish", `{}`)
	if pkg.ArtifactContentBase64 != "" {
		t.Fatalf("publish package response must not expose artifact content")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/update-packages", nil)
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list packages status %d: %s", rec.Code, rec.Body.String())
	}
	var listed []UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode listed packages: %v", err)
	}
	if len(listed) != 1 || listed[0].ArtifactContentBase64 != "" {
		t.Fatalf("list packages must not expose artifact content: %+v", listed)
	}
	assignments := postJSONForTest[[]UpdateAssignment](t, app, "/api/update-packages/"+strconvFormat(pkg.ID)+"/assign", `{"customerIds":[`+strconvFormat(customer.ID)+`]}`)
	if len(assignments) != 1 {
		t.Fatalf("expected one assignment: %+v", assignments)
	}

	poll := postJSONForTest[UpdateTaskPollResponse](t, app, "/api/product-ops/system-updates/tasks", `{"updaterToken":"api-updater-token"}`)
	if !poll.Accepted || len(poll.Tasks) != 1 {
		t.Fatalf("unexpected poll: %+v", poll)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/system/updates/"+strconvFormat(pkg.ID)+"/download?assignmentId="+strconvFormat(assignments[0].ID), nil)
	req.Header.Set("X-CBMP-Updater-Token", "api-updater-token")
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"verified":true`) {
		t.Fatalf("download status %d: %s", rec.Code, rec.Body.String())
	}
	var download UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode download: %v", err)
	}
	if download.ArtifactContentBase64 == "" || download.Package.ArtifactContentBase64 != "" {
		t.Fatalf("download should expose artifact only at envelope level: %+v", download)
	}

	report := postJSONForTest[UpdateAssignment](t, app, "/api/product-ops/system-updates/tasks/"+poll.Tasks[0].TaskNo+"/report", `{"updaterToken":"api-updater-token","status":"applied","progress":100,"currentVersion":"1.1.0"}`)
	if report.Status != "applied" || report.AppliedAt == "" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRenewalLicensePackageDownloadAPI(t *testing.T) {
	t.Setenv("CBM_OPS_SEED_DEMO", "")
	setTestLicenseIssuer(t)

	app := NewApp(NewStore(filepath.Join(t.TempDir(), "ops.json")), "")
	customer := postJSONForTest[CustomerDeployment](t, app, "/api/customers", `{
		"customerName":"授权客户",
		"licenseId":"CBMP-LICENSE-001",
		"updaterToken":"license-token",
		"expiresAt":"2027-12-31",
		"modules":["erp","dispatch","finance"],
		"maxSites":2,
		"maxVehicles":20
	}`)
	renewal := postJSONForTest[LicenseRenewal](t, app, "/api/customers/"+strconvFormat(customer.ID)+"/renewals", `{
		"expiresAt":"2028-12-31",
		"edition":"Enterprise Appliance",
		"maxSites":4,
		"maxVehicles":40,
		"operator":"ops"
	}`)
	if renewal.Signature == "" || renewal.LicensePackageNo == "" {
		t.Fatalf("renewal did not issue package metadata: %+v", renewal)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/renewals/"+strconvFormat(renewal.ID)+"/license-package", nil)
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download license package status %d: %s", rec.Code, rec.Body.String())
	}
	var download LicensePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode download envelope: %v", err)
	}
	if download.FileName == "" || download.Package.ExpiresAt != "2028-12-31" || download.Package.MaxSites != 4 {
		t.Fatalf("unexpected download envelope: %+v", download)
	}
	verifySignedLicensePackage(t, download.Package)
	decoded, err := base64.StdEncoding.DecodeString(download.ContentBase64)
	if err != nil {
		t.Fatalf("decode content base64: %v", err)
	}
	var pkg LicensePackage
	if err := json.Unmarshal(decoded, &pkg); err != nil {
		t.Fatalf("decode license package content: %v", err)
	}
	if pkg.LicenseID != customer.LicenseID || pkg.Signature != download.Package.Signature {
		t.Fatalf("content package mismatch: %+v", pkg)
	}
}

func TestAlertReportAPICreatesAndClosesCustomerAlert(t *testing.T) {
	t.Setenv("CBM_OPS_SEED_DEMO", "")

	app := NewApp(NewStore(filepath.Join(t.TempDir(), "ops.json")), "")
	customer := postJSONForTest[CustomerDeployment](t, app, "/api/customers", `{
		"customerName":"告警客户",
		"licenseId":"CBMP-ALERT-001",
		"updaterToken":"alert-token",
		"expiresAt":"2027-12-31",
		"maxSites":1,
		"maxVehicles":1
	}`)
	otherCustomer := postJSONForTest[CustomerDeployment](t, app, "/api/customers", `{
		"customerName":"另一个告警客户",
		"licenseId":"CBMP-ALERT-002",
		"updaterToken":"other-alert-token",
		"expiresAt":"2027-12-31",
		"maxSites":1,
		"maxVehicles":1
	}`)
	alert := postJSONForTest[SystemAlert](t, app, "/api/product-ops/alerts/report", `{
		"updaterToken":"alert-token",
		"customerId":`+strconvFormat(otherCustomer.ID)+`,
		"source":"server",
		"severity":"critical",
		"title":"服务端错误率升高",
		"message":"5xx 超过阈值"
	}`)
	if alert.ID == 0 || alert.Status != "open" || alert.Severity != "critical" || alert.CustomerID != customer.ID {
		t.Fatalf("unexpected alert: %+v", alert)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/product-ops/alerts/report", strings.NewReader(`{
		"updaterToken":"CBMP-ALERT-001",
		"source":"server",
		"severity":"critical",
		"title":"伪造授权号上报",
		"message":"不应接受授权号作为 token"
	}`))
	req.Header.Set("Content-Type", "application/json")
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected license id token rejected, got %d: %s", rec.Code, rec.Body.String())
	}
	alert = postJSONForTest[SystemAlert](t, app, "/api/alerts/"+strconvFormat(alert.ID)+"/ack", `{}`)
	if alert.Status != "acknowledged" || alert.AcknowledgedAt == "" {
		t.Fatalf("alert not acknowledged: %+v", alert)
	}
	alert = postJSONForTest[SystemAlert](t, app, "/api/alerts/"+strconvFormat(alert.ID)+"/resolve", `{"operator":"ops","resolution":"已恢复"}`)
	if alert.Status != "resolved" || alert.ResolvedAt == "" {
		t.Fatalf("alert not resolved: %+v", alert)
	}
}

func postJSONForTest[T any](t *testing.T, app *App, path string, body string) T {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("post %s status %d: %s", path, rec.Code, rec.Body.String())
	}
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return out
}

func strconvFormat(value int64) string {
	return strconv.FormatInt(value, 10)
}
