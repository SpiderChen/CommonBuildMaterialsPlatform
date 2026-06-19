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
	msg := newMessage("PLC,PLC-1,HZS,telemetry,temp=42", "test", "tcp", "plant-csv", time.Now())
	if err := f.Forward(context.Background(), msg); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	if got.Raw != msg.Raw || got.Channel != "tcp" || got.Protocol != "plant-csv" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
}

func TestUnwrapBodyEnvelope(t *testing.T) {
	raw, protocol, source := unwrapBody([]byte(`{"raw":"PLC,PLC-1,HZS,telemetry,temp=42","protocol":"plant-csv","source":"adapter-a"}`), "default", "source")
	if !strings.HasPrefix(raw, "PLC,") || protocol != "plant-csv" || source != "adapter-a" {
		t.Fatalf("unexpected unwrap values: %s %s %s", raw, protocol, source)
	}
}
