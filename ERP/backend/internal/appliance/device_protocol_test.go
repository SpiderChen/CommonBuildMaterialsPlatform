package appliance

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGPSProtocolFrameIngestRecordsFrameAndLocation(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	payload := `{"channel":"mqtt","protocol":"gps-json","raw":"{\"deviceNo\":\"GPS1000002\",\"plateNo\":\"粤B22336\",\"longitude\":113.9412,\"latitude\":22.5428,\"speed\":38.5,\"direction\":96,\"mileage\":2001.5,\"accStatus\":1,\"locationTime\":\"2026-06-18 12:05:00\"}"}`
	rec := testDeviceRequest(t, app, "device-demo-key-2", http.MethodPost, "/api/iot/protocols/gps/ingest", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gps protocol ingest status %d: %s", rec.Code, rec.Body.String())
	}
	var response protocolIngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode gps protocol response: %v", err)
	}
	if response.Frame.Status != "accepted" || response.Frame.ParsedResource != "vehicle_location" || response.Frame.ParsedID == 0 {
		t.Fatalf("unexpected gps protocol frame: %+v", response.Frame)
	}
	if response.Location == nil || response.Location.PlateNo != "粤B22336" || response.Location.DeviceID != "GPS1000002" {
		t.Fatalf("unexpected gps location: %+v", response.Location)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/integrations/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("integrations overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		ProtocolFrames []DeviceProtocolFrame `json:"protocolFrames"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode integrations overview: %v", err)
	}
	if len(overview.ProtocolFrames) != 1 || overview.ProtocolFrames[0].Status != "accepted" {
		t.Fatalf("expected accepted frame in overview, got %+v", overview.ProtocolFrames)
	}
}

func TestScaleProtocolFrameIngestAndRejectedFrame(t *testing.T) {
	app := newTestHTTPApp(t)
	payload := `{"channel":"serial","protocol":"scale-csv","raw":"SCALE,NS-SCALE-01,1,粤B12345,粤B12345,31.8,gross,true,capture://ns-scale-01/protocol-gross.jpg"}`
	rec := testDeviceRequest(t, app, "scale-demo-key-1", http.MethodPost, "/api/weighbridge/protocols/scale/ingest", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("scale protocol ingest status %d: %s", rec.Code, rec.Body.String())
	}
	var response protocolIngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode scale protocol response: %v", err)
	}
	if response.Frame.Status != "accepted" || response.Frame.ParsedResource != "scale_device_event" || response.Frame.ParsedID == 0 {
		t.Fatalf("unexpected scale protocol frame: %+v", response.Frame)
	}
	if response.ScaleEvent == nil || response.ScaleEvent.DeviceCode != "NS-SCALE-01" || response.ScaleEvent.Weight != 31.8 {
		t.Fatalf("unexpected scale event: %+v", response.ScaleEvent)
	}

	rec = testDeviceRequest(t, app, "scale-demo-key-1", http.MethodPost, "/api/weighbridge/protocols/scale/ingest", `bad-scale-frame`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected rejected bad frame, got %d: %s", rec.Code, rec.Body.String())
	}
	var failed struct {
		Error string              `json:"error"`
		Frame DeviceProtocolFrame `json:"frame"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &failed); err != nil {
		t.Fatalf("decode failed frame: %v", err)
	}
	if failed.Error == "" || failed.Frame.Status != "rejected" || failed.Frame.FrameNo == "" {
		t.Fatalf("expected rejected frame details, got %+v", failed)
	}

	rec = testDeviceRequest(t, app, "device-demo-key-1", http.MethodPost, "/api/weighbridge/protocols/scale/ingest", payload)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected gps key denied for scale protocol, got %d: %s", rec.Code, rec.Body.String())
	}
}
