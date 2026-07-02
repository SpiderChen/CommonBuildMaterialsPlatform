package appliance

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMenuPermissionMarksCoverCurrentWorkbenchPages(t *testing.T) {
	requiredKeys := []string{
		"production-plans",
		"production-tasks",
		"production-batches",
		"production-reports",
		"dispatch-schedules",
		"dispatch-queue",
		"stock-yards",
		"customer-risk",
		"sales-pricing",
		"portal-customer",
		"raw-material-receipts",
		"inventory-transfers",
		"inventory-stocktakes",
		"raw-material-inspections",
		"finance-receivables",
		"finance-invoices",
		"finance-collections",
		"finance-suppliers",
		"finance-carriers",
		"master-carriers",
		"portal-driver",
		"delivery-signs",
		"system-license",
		"system-maintenance",
		"system-gateway",
		"system-security",
		"system-identity",
		"system-plugins",
		"system-rules",
		"system-integrations",
		"system-audit",
	}
	marks := menuPermissionMarks()
	editable := editableMenuLabelKeys()
	for _, key := range requiredKeys {
		if marks[key] == "" {
			t.Fatalf("menuPermissionMarks missing %s", key)
		}
		if _, ok := editable[key]; !ok {
			t.Fatalf("editableMenuLabelKeys missing %s", key)
		}
	}
}

func TestSystemMenuLabelsOnlyUpdateKnownLabels(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/menu-labels", `{"key":"orders","label":"Order Center"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save menu label status %d: %s", rec.Code, rec.Body.String())
	}

	var labels map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &labels); err != nil {
		t.Fatalf("decode menu labels: %v", err)
	}
	if labels["orders"] != "Order Center" {
		t.Fatalf("expected updated menu label, got %+v", labels)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/menu-labels", `{"key":"custom-menu","label":"Custom"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown menu key should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	var bootstrap struct {
		MenuLabels map[string]string `json:"menuLabels"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if bootstrap.MenuLabels["orders"] != "Order Center" {
		t.Fatalf("bootstrap should expose saved label, got %+v", bootstrap.MenuLabels)
	}
	if _, ok := bootstrap.MenuLabels["custom-menu"]; ok {
		t.Fatalf("unknown menu key must not be persisted: %+v", bootstrap.MenuLabels)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/menu-labels/orders", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("reset menu label status %d: %s", rec.Code, rec.Body.String())
	}
	labels = nil
	if err := json.Unmarshal(rec.Body.Bytes(), &labels); err != nil {
		t.Fatalf("decode reset menu labels: %v", err)
	}
	if _, ok := labels["orders"]; ok {
		t.Fatalf("orders menu label should be reset, got %+v", labels)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/menu-labels/custom-menu", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown menu key reset should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap after reset status %d: %s", rec.Code, rec.Body.String())
	}
	bootstrap = struct {
		MenuLabels map[string]string `json:"menuLabels"`
	}{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("decode bootstrap after reset: %v", err)
	}
	if _, ok := bootstrap.MenuLabels["orders"]; ok {
		t.Fatalf("bootstrap should not expose reset label, got %+v", bootstrap.MenuLabels)
	}
}
