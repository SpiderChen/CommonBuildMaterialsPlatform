package gateway

import (
	"testing"
	"time"
)

func TestParseJSONFrame(t *testing.T) {
	raw := `{"deviceNo":"PLC-1","assetNo":"AMP240-A","eventType":"batch","timestamp":"2026-06-18 10:20:00","readings":{"bitumen":{"value":8.32,"unit":"t"},"temperature":165},"shift":"A"}`
	msg, err := ParseFrame(raw, "test", "http", "industrial-json", time.Date(2026, 6, 18, 10, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ParseFrame returned error: %v", err)
	}
	if msg.DeviceNo != "PLC-1" || msg.AssetNo != "AMP240-A" || msg.EventType != "batch" {
		t.Fatalf("unexpected identity: %+v", msg)
	}
	if len(msg.Readings) != 2 {
		t.Fatalf("expected two readings, got %d", len(msg.Readings))
	}
	if msg.Tags["shift"] != "A" {
		t.Fatalf("expected shift tag, got %+v", msg.Tags)
	}
}

func TestParseCSVFrame(t *testing.T) {
	raw := "PLC,PLC-2,AMP160-B,telemetry,temp=165:C:good|pressure=1.28:MPa,2026-06-18 11:00:00"
	msg, err := ParseFrame(raw, "test", "tcp", "industrial-csv", time.Now())
	if err != nil {
		t.Fatalf("ParseFrame returned error: %v", err)
	}
	if msg.DeviceNo != "PLC-2" || msg.Protocol != "plc-csv" {
		t.Fatalf("unexpected message: %+v", msg)
	}
	if len(msg.Readings) != 2 || msg.Readings[0].Name != "temp" || msg.Readings[0].Unit != "C" {
		t.Fatalf("unexpected readings: %+v", msg.Readings)
	}
}

func TestParseRejectsMissingReadings(t *testing.T) {
	_, err := ParseFrame(`{"deviceNo":"PLC-1","status":"online"}`, "test", "http", "industrial-json", time.Now())
	if err == nil {
		t.Fatal("expected missing readings error")
	}
}
