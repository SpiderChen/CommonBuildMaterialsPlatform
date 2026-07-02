package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPForwarderProtocolFrameMode(t *testing.T) {
	var got ProtocolFrameEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CBMP-Signature") == "" {
			t.Error("missing signature")
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	f := &HTTPForwarder{
		target: server.URL,
		secret: "secret",
		mode:   "protocol-frame",
		client: server.Client(),
	}
	msg := newMessage("PLC,PLC-1,AMP,telemetry,temp=42", "test", "tcp", "plant-csv", time.Now())
	if err := f.Forward(context.Background(), msg); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	if got.Raw != msg.Raw || got.Channel != "tcp" || got.Protocol != "plant-csv" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
	if len(got.Payload) != 0 {
		t.Fatalf("expected no ERP payload for incomplete plant frame, got %s", string(got.Payload))
	}
}

func TestHTTPForwarderProtocolFrameModeBuildsPlantPayload(t *testing.T) {
	var got ProtocolFrameEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	f := &HTTPForwarder{
		target: server.URL,
		mode:   "protocol-frame",
		client: server.Client(),
	}
	raw := `{"deviceNo":"PLANT-NS-AMP240","plantCode":"NS-AMP240","taskId":42,"batchNo":"PLC-BATCH-001","quantity":6,"completedAt":"2026-06-20 09:20:00","eventType":"batch"}`
	msg, err := ParseFrame(raw, "test", "tcp", "plant-json", time.Now())
	if err != nil {
		t.Fatalf("ParseFrame returned error: %v", err)
	}
	if err := f.Forward(context.Background(), msg); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	var payload struct {
		DeviceNo  string  `json:"deviceNo"`
		TaskID    int64   `json:"taskId"`
		BatchNo   string  `json:"batchNo"`
		PlantCode string  `json:"plantCode"`
		Quantity  float64 `json:"quantity"`
	}
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("decode plant payload: %v", err)
	}
	if payload.DeviceNo != "PLANT-NS-AMP240" || payload.TaskID != 42 || payload.BatchNo != "PLC-BATCH-001" || payload.PlantCode != "NS-AMP240" || payload.Quantity != 6 {
		t.Fatalf("unexpected plant payload: %+v", payload)
	}
}

func TestHTTPForwarderProtocolFrameModeBuildsBufferPayload(t *testing.T) {
	var got ProtocolFrameEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	f := &HTTPForwarder{
		target: server.URL,
		mode:   "protocol-frame",
		client: server.Client(),
	}
	raw := `{"deviceNo":"PLANT-NS-AMP240","plantCode":"NS-AMP240","bufferCode":"NS-AMP240-AGG-01","materialId":3,"quantity":32.5,"moistureRate":4.2,"qualityStatus":"passed","eventType":"buffer_level","reportedAt":"2026-06-21 22:50:00"}`
	msg, err := ParseFrame(raw, "test", "tcp", "buffer-json", time.Now())
	if err != nil {
		t.Fatalf("ParseFrame returned error: %v", err)
	}
	if err := f.Forward(context.Background(), msg); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	var payload struct {
		DeviceNo     string  `json:"deviceNo"`
		PlantCode    string  `json:"plantCode"`
		BufferCode   string  `json:"bufferCode"`
		MaterialID   int64   `json:"materialId"`
		Quantity     float64 `json:"quantity"`
		MoistureRate float64 `json:"moistureRate"`
	}
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("decode buffer payload: %v", err)
	}
	if payload.DeviceNo != "PLANT-NS-AMP240" || payload.PlantCode != "NS-AMP240" || payload.BufferCode != "NS-AMP240-AGG-01" || payload.MaterialID != 3 || payload.Quantity != 32.5 || payload.MoistureRate != 4.2 {
		t.Fatalf("unexpected buffer payload: %+v", payload)
	}
}

func TestHTTPForwarderProtocolFrameModeBuildsYardPayload(t *testing.T) {
	var got ProtocolFrameEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	f := &HTTPForwarder{
		target: server.URL,
		mode:   "protocol-frame",
		client: server.Client(),
	}
	raw := `{"deviceNo":"YARD-NS-AGG","yardCode":"NS-YARD-AGG","pileCode":"NS-YARD-SAND-01","materialId":3,"quantity":538.5,"moistureRate":4.7,"qualityStatus":"passed","eventType":"yard_level","reportedAt":"2026-06-21 23:10:00"}`
	msg, err := ParseFrame(raw, "test", "tcp", "yard-json", time.Now())
	if err != nil {
		t.Fatalf("ParseFrame returned error: %v", err)
	}
	if err := f.Forward(context.Background(), msg); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	var payload struct {
		DeviceNo     string  `json:"deviceNo"`
		YardCode     string  `json:"yardCode"`
		PileCode     string  `json:"pileCode"`
		MaterialID   int64   `json:"materialId"`
		Quantity     float64 `json:"quantity"`
		MoistureRate float64 `json:"moistureRate"`
	}
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("decode yard payload: %v", err)
	}
	if payload.DeviceNo != "YARD-NS-AGG" || payload.YardCode != "NS-YARD-AGG" || payload.PileCode != "NS-YARD-SAND-01" || payload.MaterialID != 3 || payload.Quantity != 538.5 || payload.MoistureRate != 4.7 {
		t.Fatalf("unexpected yard payload: %+v", payload)
	}
}

func TestUnwrapBodyEnvelope(t *testing.T) {
	raw, protocol, source := unwrapBody([]byte(`{"raw":"PLC,PLC-1,AMP,telemetry,temp=42","protocol":"plant-csv","source":"adapter-a"}`), "default", "source")
	if !strings.HasPrefix(raw, "PLC,") || protocol != "plant-csv" || source != "adapter-a" {
		t.Fatalf("unexpected unwrap values: %s %s %s", raw, protocol, source)
	}
}
