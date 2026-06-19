package appliance

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRuntimeStoresTrackPointToClickHouse(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		received = string(raw)
		_, _ = w.Write([]byte("Ok.\n"))
	}))
	defer server.Close()

	runtime := &RuntimeServices{
		clickhouseHTTP: server.URL,
		httpClient: &http.Client{Timeout: time.Second},
	}
	runtime.StoreTrackPoint(VehicleLocationEvent{
		ID: 99, VehicleID: 1, PlateNo: "粤B12345", DriverID: 1, DispatchID: 2,
		DeviceID: "GPS1000001", SourceType: "gps_device", Longitude: 113.95,
		Latitude: 22.53, Speed: 42, Direction: 180, Mileage: 123.4,
		OnlineStatus: "online", Address: "测试道路", LocationTime: "2026-06-18 10:00:00", ReceiveTime: "2026-06-18 10:00:01",
	})
	if !strings.Contains(received, "cbmp_vehicle_track_point") {
		t.Fatalf("expected table DDL in clickhouse payload: %s", received)
	}
	if !strings.Contains(received, `"plate_no":"粤B12345"`) {
		t.Fatalf("expected track JSON row in payload: %s", received)
	}
}
