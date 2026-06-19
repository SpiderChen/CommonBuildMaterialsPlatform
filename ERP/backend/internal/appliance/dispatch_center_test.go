package appliance

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestDispatchCenterOverviewForDispatcher(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "dispatcher", "dispatch123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/dispatch-center/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dispatch center overview status %d: %s", rec.Code, rec.Body.String())
	}

	var overview DispatchCenterOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode dispatch center overview: %v", err)
	}
	if overview.KPIs.TotalVehicles == 0 || overview.KPIs.ActiveDispatches == 0 {
		t.Fatalf("expected vehicle and active dispatch KPIs, got %+v", overview.KPIs)
	}
	if len(overview.SiteProgress) == 0 {
		t.Fatalf("expected site supply progress")
	}
	if overview.SiteProgress[0].PlanQuantity == 0 || overview.SiteProgress[0].ProducedPercent == 0 {
		t.Fatalf("expected quantity and production progress, got %+v", overview.SiteProgress[0])
	}
	if len(overview.VehicleQueue) == 0 || overview.VehicleQueue[0].QueueNo == "" || overview.VehicleQueue[0].ProjectName == "" {
		t.Fatalf("expected dispatch queue with project context, got %+v", overview.VehicleQueue)
	}
	queueItem := overview.VehicleQueue[0]
	if queueItem.ETA == "" || queueItem.PlannedETA == "" || queueItem.ETASource == "" || queueItem.ETAConfidence == "" {
		t.Fatalf("expected dispatch queue ETA details, got %+v", queueItem)
	}
	if queueItem.ETADistanceKm < 0 || queueItem.ETAMinutes < 0 {
		t.Fatalf("expected non-negative ETA distance and minutes, got %+v", queueItem)
	}
	if queueItem.ETASource != "planned" && queueItem.ETASpeedKPH <= 0 {
		t.Fatalf("expected computed ETA speed, got %+v", queueItem)
	}
	if len(overview.ProductionTasks) == 0 || overview.ProductionTasks[0].RemainingQty == 0 {
		t.Fatalf("expected active production tasks, got %+v", overview.ProductionTasks)
	}
}
