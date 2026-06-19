package forwarder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPSinkProtocolFrameMode(t *testing.T) {
	var got ProtocolFrameEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-CBMP-Signature") == "" {
			t.Error("missing signature")
		}
		if r.Header.Get("X-Device-Key") != "device-demo-key-2" {
			t.Errorf("missing device key")
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	sink := &HTTPSink{
		target:    server.URL,
		deviceKey: "device-demo-key-2",
		secret:    "secret",
		mode:      "protocol-frame",
		client:    server.Client(),
	}
	loc := newLocation("GPS,GPS1,粤B1,113.9,22.5,30,90,10,1,2026-06-18 12:00:00", "test", "tcp", "gps-csv", time.Now())
	if err := sink.Write(context.Background(), loc); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if got.Raw != loc.Raw || got.Channel != "tcp" || got.Protocol != "gps-csv" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
}

func TestDeduper(t *testing.T) {
	d := NewDeduper(time.Minute)
	loc := Location{DeviceNo: "GPS1", Longitude: 113.9, Latitude: 22.5, LocationTime: "2026-06-18T12:00:00Z"}
	now := time.Now()
	if d.Seen(loc, now) {
		t.Fatal("first point should not be duplicate")
	}
	if !d.Seen(loc, now.Add(time.Second)) {
		t.Fatal("second point should be duplicate")
	}
}
