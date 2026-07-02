package appliance

import (
	"net/http"
	"testing"
)

func TestWorkbenchGETAPISurfaceIsRoutable(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	endpoints := []string{
		"/api/bootstrap",
		"/api/dashboard",
		"/api/reports",
		"/api/dispatch-center/overview",
		"/api/orders",
		"/api/portal/overview",
		"/api/portal/complaints",
		"/api/quality/overview",
		"/api/procurement/overview",
		"/api/finance/overview",
		"/api/rules",
		"/api/integrations/overview",
		"/api/contracts",
		"/api/dispatch-orders",
		"/api/dispatch-orders/schedules",
		"/api/dispatch-orders/carrier-settlements",
		"/api/weighbridge/tickets",
		"/api/weighbridge/ticket-prints",
		"/api/weighbridge/ticket-voids",
		"/api/weighbridge/weight-records",
		"/api/weighbridge/device-events",
		"/api/delivery/notes",
		"/api/delivery/sign",
		"/api/delivery/sign-links",
		"/api/delivery/sign-attachments",
		"/api/statements",
		"/api/vehicle/location/latest",
		"/api/vehicle/alarms",
		"/api/vehicle/fences",
		"/api/system/runtime",
		"/api/system/backups",
		"/api/system/backups/drills",
		"/api/system/gateway",
		"/api/system/users",
		"/api/system/roles",
		"/api/system/dictionaries",
		"/api/system/audit",
		"/api/system/modules",
		"/api/system/org",
		"/api/system/workflows",
		"/api/system/workflows/catalog",
		"/api/system/workflows/inbox",
		"/api/system/workflows/instances",
		"/api/system/workflows/tasks",
		"/api/system/workflows/logs",
		"/api/system/workflows/events",
		"/api/system/workflows/outbox",
		"/api/system/workflows/deliveries",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			rec := testRequest(t, app, token, http.MethodGet, endpoint, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s status %d: %s", endpoint, rec.Code, rec.Body.String())
			}
		})
	}
}
