package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestGeoFenceHTTPAPISupportsCreateUpdateArchiveAndEvents(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/vehicle/fences", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list fences status %d: %s", rec.Code, rec.Body.String())
	}
	var fences []GeoFence
	if err := json.Unmarshal(rec.Body.Bytes(), &fences); err != nil {
		t.Fatalf("decode fences: %v", err)
	}
	if len(fences) < 4 || fences[0].Shape == "" {
		t.Fatalf("expected normalized seed fences, got %+v", fences)
	}

	body := `{"name":"科技园二期多边形围栏","type":"project","projectId":1,"shape":"polygon","polygon":[{"longitude":113.944,"latitude":22.535},{"longitude":113.946,"latitude":22.535},{"longitude":113.946,"latitude":22.537},{"longitude":113.944,"latitude":22.537}],"status":"active"}`
	rec = testRequest(t, app, token, http.MethodPost, "/api/vehicle/fences", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create fence status %d: %s", rec.Code, rec.Body.String())
	}
	var created GeoFence
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created fence: %v", err)
	}
	if created.ID == 0 || created.Shape != "polygon" || created.ProjectID != 1 || created.Longitude == 0 || created.Latitude == 0 {
		t.Fatalf("unexpected created fence: %+v", created)
	}

	update := `{"name":"科技园二期圆形围栏","type":"project","projectId":1,"shape":"circle","longitude":113.9452,"latitude":22.5358,"radius":520,"status":"active"}`
	rec = testRequest(t, app, token, http.MethodPut, "/api/vehicle/fences/"+geoFenceItoa(created.ID), update)
	if rec.Code != http.StatusOK {
		t.Fatalf("update fence status %d: %s", rec.Code, rec.Body.String())
	}
	var updated GeoFence
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated fence: %v", err)
	}
	if updated.Shape != "circle" || updated.Radius != 520 || len(updated.Polygon) != 0 {
		t.Fatalf("unexpected updated fence: %+v", updated)
	}

	report := `{"plateNo":"粤B12345","longitude":113.9452,"latitude":22.5358,"speed":18,"direction":120,"locationTime":"2026-06-18 12:00:00"}`
	rec = testRequest(t, app, token, http.MethodPost, "/api/iot/vehicle/location/report", report)
	if rec.Code != http.StatusCreated {
		t.Fatalf("report location status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/vehicle/fence-events?vehicleId=1&limit=10", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("fence events status %d: %s", rec.Code, rec.Body.String())
	}
	var events []GeoFenceEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatalf("decode fence events: %v", err)
	}
	if len(events) == 0 || events[0].EventType != "enter" {
		t.Fatalf("expected enter fence event, got %+v", events)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/vehicle/fences/"+geoFenceItoa(created.ID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("archive fence status %d: %s", rec.Code, rec.Body.String())
	}
	var archived GeoFence
	if err := json.Unmarshal(rec.Body.Bytes(), &archived); err != nil {
		t.Fatalf("decode archived fence: %v", err)
	}
	if archived.Status != "inactive" {
		t.Fatalf("expected inactive fence, got %+v", archived)
	}
}

func TestCreateFenceEventsDeduplicatesEnterAndRecordsLeave(t *testing.T) {
	data := SeedData()
	data.GeoFences = []GeoFence{{ID: 1, Name: "测试围栏", Type: "site", SiteID: 1, Longitude: 113.9345, Latitude: 22.5431, Radius: 500, Shape: "circle", Status: "active"}}
	data.GeoFenceEvents = nil

	inside := VehicleLocationEvent{ID: 1, VehicleID: 1, PlateNo: "粤B12345", Longitude: 113.9345, Latitude: 22.5431, LocationTime: "2026-06-18 12:00:00"}
	createFenceEvents(&data, inside)
	createFenceEvents(&data, inside)
	if len(data.GeoFenceEvents) != 1 || data.GeoFenceEvents[0].EventType != "enter" {
		t.Fatalf("expected one deduplicated enter event, got %+v", data.GeoFenceEvents)
	}

	outside := VehicleLocationEvent{ID: 2, VehicleID: 1, PlateNo: "粤B12345", Longitude: 113.9800, Latitude: 22.5800, LocationTime: "2026-06-18 12:10:00"}
	createFenceEvents(&data, outside)
	if len(data.GeoFenceEvents) != 2 || data.GeoFenceEvents[1].EventType != "leave" {
		t.Fatalf("expected leave event, got %+v", data.GeoFenceEvents)
	}
}

func geoFenceItoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
