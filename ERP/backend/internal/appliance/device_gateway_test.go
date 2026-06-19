package appliance

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDeviceGatewayTCPIngestsGPSFrame(t *testing.T) {
	app := newTestHTTPApp(t)
	gateway := NewDeviceGateway(app, []DeviceGatewayEndpointConfig{
		{
			Name: "gps-tcp-test", Kind: "tcp", Address: "127.0.0.1:0", Parser: "gps",
			Channel: "tcp", Protocol: "gps-csv", Permission: "location:report",
		},
	}, 20*time.Millisecond)
	app.gateway = gateway
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := gateway.Start(ctx); err != nil {
		t.Fatalf("start gateway: %v", err)
	}
	statuses := gateway.Status()
	if len(statuses) != 1 || statuses[0].Address == "" {
		t.Fatalf("expected listening gateway status, got %+v", statuses)
	}
	conn, err := net.Dial("tcp", statuses[0].Address)
	if err != nil {
		t.Fatalf("dial gateway: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	if _, err := fmt.Fprintln(conn, "AUTH device-demo-key-2"); err != nil {
		t.Fatalf("write auth: %v", err)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read auth response: %v", err)
	}
	if strings.TrimSpace(line) != "OK AUTH" {
		t.Fatalf("unexpected auth response: %q", line)
	}
	if _, err := fmt.Fprintln(conn, "GPS,GPS1000002,粤B22336,113.9412,22.5428,38.5,96,2001.5,1,2026-06-18 12:05:00"); err != nil {
		t.Fatalf("write frame: %v", err)
	}
	line, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read frame response: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(line), "OK DPF") {
		t.Fatalf("unexpected frame response: %q", line)
	}
	data := app.mustSnapshot()
	if len(data.DeviceProtocolFrames) != 1 || data.DeviceProtocolFrames[0].Status != "accepted" {
		t.Fatalf("expected accepted gateway frame, got %+v", data.DeviceProtocolFrames)
	}
	if !hasLatestLocation(data.LatestLocations, "粤B22336") {
		t.Fatalf("expected latest location for gateway frame, got %+v", data.LatestLocations)
	}
	statuses = gateway.Status()
	if statuses[0].AcceptedFrames != 1 || statuses[0].RejectedFrames != 0 {
		t.Fatalf("unexpected gateway counters: %+v", statuses[0])
	}
}

func TestDeviceGatewaySerialFileIngestsScaleFrame(t *testing.T) {
	app := newTestHTTPApp(t)
	dir := t.TempDir()
	serialFile := filepath.Join(dir, "scale-serial.log")
	if err := os.WriteFile(serialFile, []byte("KEY=scale-demo-key-1|SCALE,NS-SCALE-01,1,粤B12345,粤B12345,32.1,gross,true,capture://scale/file-gross.jpg\n"), 0600); err != nil {
		t.Fatalf("write serial file: %v", err)
	}
	gateway := NewDeviceGateway(app, []DeviceGatewayEndpointConfig{
		{
			Name: "scale-file-test", Kind: "file", FilePath: serialFile, Parser: "scale",
			Channel: "serial", Protocol: "scale-csv", Permission: "scale:report",
		},
	}, 10*time.Millisecond)
	app.gateway = gateway
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := gateway.Start(ctx); err != nil {
		t.Fatalf("start file gateway: %v", err)
	}
	waitForCondition(t, 500*time.Millisecond, func() bool {
		data := app.mustSnapshot()
		return len(data.DeviceProtocolFrames) == 1 && len(data.ScaleDeviceEvents) == 1
	})
	data := app.mustSnapshot()
	if data.DeviceProtocolFrames[0].Status != "accepted" || data.ScaleDeviceEvents[0].Weight != 32.1 {
		t.Fatalf("unexpected serial gateway result: frames=%+v events=%+v", data.DeviceProtocolFrames, data.ScaleDeviceEvents)
	}
	statuses := gateway.Status()
	if len(statuses) != 1 || statuses[0].AcceptedFrames != 1 {
		t.Fatalf("unexpected serial gateway status: %+v", statuses)
	}
}

func TestDeviceGatewayMQTTPayloadIngestsGPSAndScaleFrames(t *testing.T) {
	app := newTestHTTPApp(t)
	gateway := NewDeviceGateway(app, []DeviceGatewayEndpointConfig{
		{
			Name: "gps-mqtt-test", Kind: "mqtt", BrokerURL: "mqtt://broker.example:1883", Topic: "cbmp/gps/+/up", Parser: "gps",
			Channel: "mqtt", Protocol: "gps-json", Permission: "location:report",
		},
		{
			Name: "scale-mqtt-test", Kind: "mqtt", BrokerURL: "mqtt://broker.example:1883", Topic: "cbmp/scale/+/up", Parser: "scale",
			Channel: "mqtt", Protocol: "scale-csv", Permission: "scale:report",
		},
	}, 10*time.Millisecond)
	app.gateway = gateway
	gpsCfg := gateway.configs[0]
	scaleCfg := gateway.configs[1]

	ok, message := gateway.processMQTTPayload(gpsCfg, "cbmp/gps/GPS1000002/up", []byte(`{
		"deviceKey": "device-demo-key-2",
		"payload": {
			"deviceNo": "GPS1000002",
			"plateNo": "粤B22336",
			"longitude": 113.9412,
			"latitude": 22.5428,
			"speed": 38.5,
			"direction": 96,
			"mileage": 2001.5,
			"accStatus": 1,
			"locationTime": "2026-06-18 12:05:00"
		}
	}`))
	if !ok || !strings.HasPrefix(message, "DPF") {
		t.Fatalf("expected accepted mqtt gps frame, ok=%v message=%q", ok, message)
	}
	ok, message = gateway.processMQTTPayload(scaleCfg, "cbmp/scale/NS-SCALE-01/up", []byte("KEY=scale-demo-key-1|SCALE,NS-SCALE-01,1,粤B12345,粤B12345,33.6,gross,true,capture://scale/mqtt-gross.jpg"))
	if !ok || !strings.HasPrefix(message, "DPF") {
		t.Fatalf("expected accepted mqtt scale frame, ok=%v message=%q", ok, message)
	}

	data := app.mustSnapshot()
	if len(data.DeviceProtocolFrames) != 2 {
		t.Fatalf("expected 2 protocol frames, got %+v", data.DeviceProtocolFrames)
	}
	if !hasLatestLocation(data.LatestLocations, "粤B22336") {
		t.Fatalf("expected latest location from mqtt gps frame, got %+v", data.LatestLocations)
	}
	if len(data.ScaleDeviceEvents) == 0 || data.ScaleDeviceEvents[len(data.ScaleDeviceEvents)-1].Weight != 33.6 {
		t.Fatalf("expected mqtt scale event, got %+v", data.ScaleDeviceEvents)
	}
	statuses := gateway.Status()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 gateway statuses, got %+v", statuses)
	}
	for _, status := range statuses {
		if status.Kind != "mqtt" || status.AcceptedFrames != 1 || status.RejectedFrames != 0 {
			t.Fatalf("unexpected mqtt gateway status: %+v", status)
		}
	}
}

func TestDeviceGatewayFromEnvConfiguresMQTTBrokerWithoutLeakingSecret(t *testing.T) {
	app := newTestHTTPApp(t)
	t.Setenv("CBMP_MQTT_BROKER_URL", "mqtt://user:secret@broker.example:1883")
	t.Setenv("CBMP_MQTT_USERNAME", "gateway-user")
	t.Setenv("CBMP_MQTT_PASSWORD", "gateway-password")
	t.Setenv("CBMP_MQTT_CLIENT_ID", "plant-a")
	t.Setenv("CBMP_MQTT_QOS", "1")
	t.Setenv("CBMP_MQTT_DEVICE_KEY", "device-demo-key-2")
	t.Setenv("CBMP_GPS_MQTT_TOPIC", "cbmp/gps/+/up")
	t.Setenv("CBMP_SCALE_MQTT_TOPIC", "cbmp/scale/+/up")

	gateway := NewDeviceGatewayFromEnv(app)
	statuses := gateway.Status()
	if len(statuses) != 2 {
		t.Fatalf("expected gps and scale mqtt statuses, got %+v", statuses)
	}
	for _, status := range statuses {
		if status.Kind != "mqtt" || status.Address == "" || status.Source == "" {
			t.Fatalf("unexpected mqtt status metadata: %+v", status)
		}
		if strings.Contains(status.Address, "secret") || strings.Contains(status.Address, "gateway-password") {
			t.Fatalf("mqtt status leaked broker secret: %+v", status)
		}
	}
	if gateway.configs[0].QOS != 1 || gateway.configs[0].ClientID != "plant-a-gps" {
		t.Fatalf("unexpected mqtt config: %+v", gateway.configs[0])
	}
}

func hasLatestLocation(items []VehicleLatestLocation, plateNo string) bool {
	for _, item := range items {
		if item.PlateNo == plateNo {
			return true
		}
	}
	return false
}

func waitForCondition(t *testing.T, timeout time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not satisfied within %s", timeout)
}
