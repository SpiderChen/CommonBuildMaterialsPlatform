package appliance

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestBuildTrackReplayLinksStopsTicketsSignsAndFenceEvents(t *testing.T) {
	data := SeedData()
	data.Locations = append(data.Locations,
		VehicleLocationEvent{ID: 99, VehicleID: 1, PlateNo: "粤B12345", DriverID: 1, DispatchID: 1, DeviceID: "GPS1000001", SourceType: "gps_device", Longitude: 113.9368, Latitude: 22.5420, Speed: 18, Direction: 160, Mileage: 123450, AccStatus: 1, OnlineStatus: "online", Address: "南山沥青站出入口", LocationTime: "2026-06-18 10:00:00", ReceiveTime: "2026-06-18 10:00:01"},
		VehicleLocationEvent{ID: 100, VehicleID: 1, PlateNo: "粤B12345", DriverID: 1, DispatchID: 1, DeviceID: "GPS1000001", SourceType: "gps_device", Longitude: 113.9422, Latitude: 22.5386, Speed: 0, Direction: 160, Mileage: 123463, AccStatus: 1, OnlineStatus: "online", Address: "高新园临停点", LocationTime: "2026-06-18 11:10:00", ReceiveTime: "2026-06-18 11:10:01"},
		VehicleLocationEvent{ID: 101, VehicleID: 1, PlateNo: "粤B12345", DriverID: 1, DispatchID: 1, DeviceID: "GPS1000001", SourceType: "gps_device", Longitude: 113.9423, Latitude: 22.5385, Speed: 0, Direction: 160, Mileage: 123463, AccStatus: 1, OnlineStatus: "online", Address: "高新园临停点", LocationTime: "2026-06-18 11:18:00", ReceiveTime: "2026-06-18 11:18:01"},
		VehicleLocationEvent{ID: 102, VehicleID: 1, PlateNo: "粤B12345", DriverID: 1, DispatchID: 1, DeviceID: "GPS1000001", SourceType: "gps_device", Longitude: 113.9446, Latitude: 22.5364, Speed: 32, Direction: 155, Mileage: 123466, AccStatus: 1, OnlineStatus: "online", Address: "科技园二期工地外", LocationTime: "2026-06-18 11:45:00", ReceiveTime: "2026-06-18 11:45:01"},
	)
	data.GeoFenceEvents = append(data.GeoFenceEvents, GeoFenceEvent{ID: 1, VehicleID: 1, FenceID: 3, EventType: "enter", DispatchID: 1, EventTime: "2026-06-18 11:20:00"})

	replay := buildTrackReplay(data, 1, "2026-06-18 09:55:00", "2026-06-18 11:50:00")

	if replay.VehicleID != 1 || replay.PlateNo != "粤B12345" {
		t.Fatalf("unexpected vehicle in replay: %+v", replay)
	}
	if len(replay.Points) < 6 {
		t.Fatalf("expected replay points, got %d", len(replay.Points))
	}
	if replay.DistanceKm <= 0 || replay.DurationMinutes <= 0 || replay.AverageSpeed <= 0 || replay.MaxSpeed <= 0 {
		t.Fatalf("expected non-zero replay statistics: %+v", replay)
	}
	if replay.StopCount != 1 || len(replay.Stops) != 1 {
		t.Fatalf("expected one stop, got %+v", replay.Stops)
	}
	if replay.Stops[0].DurationMinutes != 8 {
		t.Fatalf("expected 8 minute stop, got %.2f", replay.Stops[0].DurationMinutes)
	}
	if len(replay.Tickets) != 1 || len(replay.Signs) != 1 || len(replay.FenceEvents) != 1 {
		t.Fatalf("expected linked ticket, sign and fence event, got tickets=%d signs=%d fences=%d", len(replay.Tickets), len(replay.Signs), len(replay.FenceEvents))
	}
	if replay.Compression.RawPointCount != len(replay.Points) || replay.Compression.CompressedPointCount != len(replay.CompressedPoints) {
		t.Fatalf("expected compression summary to match points, got %+v", replay.Compression)
	}
	if replay.Compression.CompressedPointCount == 0 || replay.Compression.CompressedPointCount > replay.Compression.RawPointCount {
		t.Fatalf("expected compressed points not to exceed raw points, got %+v", replay.Compression)
	}
	if replay.Compression.Algorithm != "rdp" || replay.Compression.PreservedStops != replay.StopCount {
		t.Fatalf("expected rdp compression preserving stops, got %+v", replay.Compression)
	}
	if replay.ExportName == "" {
		t.Fatalf("expected export name")
	}
}

func TestTrackReplayHTTPAPI(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/vehicle/track/replay?vehicleId=1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("track replay status %d: %s", rec.Code, rec.Body.String())
	}
	var replay TrackReplay
	if err := json.Unmarshal(rec.Body.Bytes(), &replay); err != nil {
		t.Fatalf("decode replay: %v", err)
	}
	if replay.VehicleID != 1 || len(replay.Points) == 0 {
		t.Fatalf("expected vehicle 1 replay points, got %+v", replay)
	}
	if len(replay.CompressedPoints) == 0 || replay.Compression.RawPointCount == 0 {
		t.Fatalf("expected compressed replay points, got %+v", replay.Compression)
	}
}
