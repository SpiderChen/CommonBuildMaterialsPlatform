package appliance

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type DeviceGatewayEndpointConfig struct {
	Name       string
	Kind       string
	Address    string
	FilePath   string
	BrokerURL  string
	Topic      string
	Username   string
	Password   string
	ClientID   string
	DeviceKey  string
	QOS        byte
	Parser     string
	Channel    string
	Protocol   string
	Permission string
}

type DeviceGatewayStatus struct {
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	Address        string `json:"address"`
	Source         string `json:"source"`
	Channel        string `json:"channel"`
	Protocol       string `json:"protocol"`
	Status         string `json:"status"`
	LastError      string `json:"lastError"`
	AcceptedFrames int64  `json:"acceptedFrames"`
	RejectedFrames int64  `json:"rejectedFrames"`
	LastFrameAt    string `json:"lastFrameAt"`
}

type DeviceGateway struct {
	app          *App
	configs      []DeviceGatewayEndpointConfig
	pollInterval time.Duration
	mu           sync.RWMutex
	statuses     map[string]DeviceGatewayStatus
	listeners    []net.Listener
	fileOffsets  map[string]int64
}

func NewDeviceGatewayFromEnv(app *App) *DeviceGateway {
	configs := []DeviceGatewayEndpointConfig{}
	if addr := strings.TrimSpace(os.Getenv("CBMP_GPS_TCP_ADDR")); addr != "" {
		configs = append(configs, DeviceGatewayEndpointConfig{
			Name: "gps-tcp", Kind: "tcp", Address: addr, Parser: "gps",
			Channel: "tcp", Protocol: "gps-csv", Permission: "location:report",
		})
	}
	if addr := strings.TrimSpace(os.Getenv("CBMP_SCALE_TCP_ADDR")); addr != "" {
		configs = append(configs, DeviceGatewayEndpointConfig{
			Name: "scale-tcp", Kind: "tcp", Address: addr, Parser: "scale",
			Channel: "tcp", Protocol: "scale-csv", Permission: "scale:report",
		})
	}
	if addr := strings.TrimSpace(os.Getenv("CBMP_PLANT_TCP_ADDR")); addr != "" {
		configs = append(configs, DeviceGatewayEndpointConfig{
			Name: "plant-tcp", Kind: "tcp", Address: addr, Parser: "plant",
			Channel: "tcp", Protocol: "plant-csv", Permission: "plant:report",
		})
	}
	if path := strings.TrimSpace(os.Getenv("CBMP_GPS_SERIAL_FILE")); path != "" {
		configs = append(configs, DeviceGatewayEndpointConfig{
			Name: "gps-serial-file", Kind: "file", FilePath: path, Parser: "gps",
			Channel: "serial", Protocol: "gps-csv", Permission: "location:report",
		})
	}
	if path := strings.TrimSpace(os.Getenv("CBMP_SCALE_SERIAL_FILE")); path != "" {
		configs = append(configs, DeviceGatewayEndpointConfig{
			Name: "scale-serial-file", Kind: "file", FilePath: path, Parser: "scale",
			Channel: "serial", Protocol: "scale-csv", Permission: "scale:report",
		})
	}
	if path := strings.TrimSpace(os.Getenv("CBMP_PLANT_SERIAL_FILE")); path != "" {
		configs = append(configs, DeviceGatewayEndpointConfig{
			Name: "plant-serial-file", Kind: "file", FilePath: path, Parser: "plant",
			Channel: "serial", Protocol: "plant-csv", Permission: "plant:report",
		})
	}
	if brokerURL := strings.TrimSpace(os.Getenv("CBMP_MQTT_BROKER_URL")); brokerURL != "" {
		username := strings.TrimSpace(os.Getenv("CBMP_MQTT_USERNAME"))
		password := os.Getenv("CBMP_MQTT_PASSWORD")
		clientID := fallback(strings.TrimSpace(os.Getenv("CBMP_MQTT_CLIENT_ID")), "cbmp-device-gateway")
		qos := parseGatewayQOS(os.Getenv("CBMP_MQTT_QOS"))
		commonKey := strings.TrimSpace(os.Getenv("CBMP_MQTT_DEVICE_KEY"))
		if topic := strings.TrimSpace(os.Getenv("CBMP_GPS_MQTT_TOPIC")); topic != "" {
			configs = append(configs, DeviceGatewayEndpointConfig{
				Name: "gps-mqtt", Kind: "mqtt", BrokerURL: brokerURL, Topic: topic, Username: username,
				Password: password, ClientID: clientID + "-gps", DeviceKey: fallback(strings.TrimSpace(os.Getenv("CBMP_GPS_MQTT_DEVICE_KEY")), commonKey),
				QOS: qos, Parser: "gps", Channel: "mqtt", Protocol: fallback(strings.TrimSpace(os.Getenv("CBMP_GPS_MQTT_PROTOCOL")), "gps-json"), Permission: "location:report",
			})
		}
		if topic := strings.TrimSpace(os.Getenv("CBMP_SCALE_MQTT_TOPIC")); topic != "" {
			configs = append(configs, DeviceGatewayEndpointConfig{
				Name: "scale-mqtt", Kind: "mqtt", BrokerURL: brokerURL, Topic: topic, Username: username,
				Password: password, ClientID: clientID + "-scale", DeviceKey: fallback(strings.TrimSpace(os.Getenv("CBMP_SCALE_MQTT_DEVICE_KEY")), commonKey),
				QOS: qos, Parser: "scale", Channel: "mqtt", Protocol: fallback(strings.TrimSpace(os.Getenv("CBMP_SCALE_MQTT_PROTOCOL")), "scale-csv"), Permission: "scale:report",
			})
		}
		if topic := strings.TrimSpace(os.Getenv("CBMP_PLANT_MQTT_TOPIC")); topic != "" {
			configs = append(configs, DeviceGatewayEndpointConfig{
				Name: "plant-mqtt", Kind: "mqtt", BrokerURL: brokerURL, Topic: topic, Username: username,
				Password: password, ClientID: clientID + "-plant", DeviceKey: fallback(strings.TrimSpace(os.Getenv("CBMP_PLANT_MQTT_DEVICE_KEY")), commonKey),
				QOS: qos, Parser: "plant", Channel: "mqtt", Protocol: fallback(strings.TrimSpace(os.Getenv("CBMP_PLANT_MQTT_PROTOCOL")), "plant-json"), Permission: "plant:report",
			})
		}
	}
	poll := time.Second
	if raw := strings.TrimSpace(os.Getenv("CBMP_DEVICE_GATEWAY_POLL_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			poll = time.Duration(ms) * time.Millisecond
		}
	}
	return NewDeviceGateway(app, configs, poll)
}

func NewDeviceGateway(app *App, configs []DeviceGatewayEndpointConfig, pollInterval time.Duration) *DeviceGateway {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	gateway := &DeviceGateway{
		app:          app,
		configs:      append([]DeviceGatewayEndpointConfig(nil), configs...),
		pollInterval: pollInterval,
		statuses:     map[string]DeviceGatewayStatus{},
		fileOffsets:  map[string]int64{},
	}
	for _, cfg := range gateway.configs {
		gateway.statuses[cfg.Name] = DeviceGatewayStatus{
			Name: cfg.Name, Kind: cfg.Kind, Address: gatewayAddress(cfg), Source: gatewaySource(cfg),
			Channel: cfg.Channel, Protocol: cfg.Protocol, Status: "configured",
		}
	}
	return gateway
}

func (g *DeviceGateway) Start(ctx context.Context) error {
	if g == nil || len(g.configs) == 0 {
		return nil
	}
	for _, cfg := range g.configs {
		switch cfg.Kind {
		case "tcp":
			if err := g.startTCP(ctx, cfg); err != nil {
				return err
			}
		case "file":
			g.setStatus(cfg.Name, "online", "")
			go g.fileLoop(ctx, cfg)
		case "mqtt":
			g.setStatus(cfg.Name, "connecting", "")
			go g.mqttLoop(ctx, cfg)
		default:
			return fmt.Errorf("unknown device gateway kind: %s", cfg.Kind)
		}
	}
	return nil
}

func (g *DeviceGateway) mqttLoop(ctx context.Context, cfg DeviceGatewayEndpointConfig) {
	for {
		select {
		case <-ctx.Done():
			g.setStatus(cfg.Name, "stopped", "")
			return
		default:
		}
		opts := mqtt.NewClientOptions().
			AddBroker(cfg.BrokerURL).
			SetClientID(cfg.ClientID).
			SetCleanSession(true).
			SetAutoReconnect(true).
			SetConnectRetry(false).
			SetOrderMatters(false).
			SetKeepAlive(30 * time.Second).
			SetPingTimeout(5 * time.Second).
			SetConnectTimeout(5 * time.Second)
		if cfg.Username != "" {
			opts.SetUsername(cfg.Username)
		}
		if cfg.Password != "" {
			opts.SetPassword(cfg.Password)
		}
		opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			if err != nil {
				g.setStatus(cfg.Name, "degraded", err.Error())
			}
		})
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			token := client.Subscribe(cfg.Topic, cfg.QOS, func(_ mqtt.Client, msg mqtt.Message) {
				go g.processMQTTPayload(cfg, msg.Topic(), msg.Payload())
			})
			if token.WaitTimeout(5*time.Second) && token.Error() == nil {
				g.setStatus(cfg.Name, "online", "")
				return
			}
			errText := "mqtt subscribe timeout"
			if token.Error() != nil {
				errText = token.Error().Error()
			}
			g.setStatus(cfg.Name, "degraded", errText)
		})
		client := mqtt.NewClient(opts)
		token := client.Connect()
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			<-ctx.Done()
			client.Disconnect(250)
			g.setStatus(cfg.Name, "stopped", "")
			return
		}
		errText := "mqtt connect timeout"
		if token.Error() != nil {
			errText = token.Error().Error()
		}
		g.setStatus(cfg.Name, "degraded", errText)
		client.Disconnect(250)
		select {
		case <-ctx.Done():
			g.setStatus(cfg.Name, "stopped", "")
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (a *App) StartDeviceGateways(ctx context.Context) error {
	if a.gateway == nil {
		return nil
	}
	return a.gateway.Start(ctx)
}

func (g *DeviceGateway) Status() []DeviceGatewayStatus {
	if g == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]DeviceGatewayStatus, 0, len(g.statuses))
	for _, item := range g.statuses {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (g *DeviceGateway) startTCP(ctx context.Context, cfg DeviceGatewayEndpointConfig) error {
	ln, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		g.setStatus(cfg.Name, "degraded", err.Error())
		return err
	}
	g.mu.Lock()
	g.listeners = append(g.listeners, ln)
	status := g.statuses[cfg.Name]
	status.Address = ln.Addr().String()
	status.Status = "online"
	status.LastError = ""
	g.statuses[cfg.Name] = status
	g.mu.Unlock()
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	go g.acceptLoop(ctx, cfg, ln)
	return nil
}

func (g *DeviceGateway) acceptLoop(ctx context.Context, cfg DeviceGatewayEndpointConfig, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				g.setStatus(cfg.Name, "stopped", "")
				return
			default:
				g.setStatus(cfg.Name, "degraded", err.Error())
				continue
			}
		}
		go g.handleConn(ctx, cfg, conn)
	}
}

func (g *DeviceGateway) handleConn(ctx context.Context, cfg DeviceGatewayEndpointConfig, conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024), 64*1024)
	writer := bufio.NewWriter(conn)
	deviceKey := ""
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line := strings.TrimSpace(scanner.Text())
		nextKey, raw, authOnly := extractGatewayLine(deviceKey, line)
		if nextKey != deviceKey {
			deviceKey = nextKey
		}
		if authOnly {
			_, _ = writer.WriteString("OK AUTH\n")
			_ = writer.Flush()
			continue
		}
		if raw == "" {
			continue
		}
		ok, message := g.processRawFrame(cfg, deviceKey, raw, remote)
		if ok {
			_, _ = writer.WriteString("OK " + message + "\n")
		} else {
			_, _ = writer.WriteString("ERR " + message + "\n")
		}
		_ = writer.Flush()
	}
	if err := scanner.Err(); err != nil {
		g.setStatus(cfg.Name, "degraded", err.Error())
	}
}

func (g *DeviceGateway) fileLoop(ctx context.Context, cfg DeviceGatewayEndpointConfig) {
	ticker := time.NewTicker(g.pollInterval)
	defer ticker.Stop()
	for {
		g.drainFile(cfg)
		select {
		case <-ctx.Done():
			g.setStatus(cfg.Name, "stopped", "")
			return
		case <-ticker.C:
		}
	}
}

func (g *DeviceGateway) drainFile(cfg DeviceGatewayEndpointConfig) {
	file, err := os.Open(cfg.FilePath)
	if err != nil {
		g.setStatus(cfg.Name, "degraded", err.Error())
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		g.setStatus(cfg.Name, "degraded", err.Error())
		return
	}
	offset := g.fileOffsets[cfg.Name]
	if info.Size() < offset {
		offset = 0
	}
	if _, err := file.Seek(offset, 0); err != nil {
		g.setStatus(cfg.Name, "degraded", err.Error())
		return
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 64*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		deviceKey, raw, authOnly := extractGatewayLine("", line)
		if authOnly || raw == "" {
			continue
		}
		g.processRawFrame(cfg, deviceKey, raw, "serial-file")
	}
	if err := scanner.Err(); err != nil {
		g.setStatus(cfg.Name, "degraded", err.Error())
		return
	}
	nextOffset, _ := file.Seek(0, 1)
	g.fileOffsets[cfg.Name] = nextOffset
	g.setStatus(cfg.Name, "online", "")
}

func (g *DeviceGateway) processRawFrame(cfg DeviceGatewayEndpointConfig, deviceKey string, raw string, remoteAddr string) (bool, string) {
	req := protocolFrameIngestRequest{Channel: cfg.Channel, Protocol: cfg.Protocol, Raw: raw}
	return g.processFrameRequest(cfg, deviceKey, req, remoteAddr)
}

func (g *DeviceGateway) processMQTTPayload(cfg DeviceGatewayEndpointConfig, topic string, payload []byte) (bool, string) {
	text := strings.TrimSpace(string(payload))
	deviceKey, raw, authOnly := extractGatewayLine(cfg.DeviceKey, text)
	if authOnly {
		return true, "auth ignored"
	}
	req := protocolFrameIngestRequest{Channel: cfg.Channel, Protocol: cfg.Protocol}
	if strings.HasPrefix(text, "{") {
		var envelope struct {
			DeviceKey string          `json:"deviceKey"`
			Channel   string          `json:"channel"`
			Protocol  string          `json:"protocol"`
			Raw       string          `json:"raw"`
			Payload   json.RawMessage `json:"payload"`
		}
		if json.Unmarshal(payload, &envelope) == nil && (envelope.DeviceKey != "" || envelope.Raw != "" || len(envelope.Payload) > 0) {
			deviceKey = fallback(strings.TrimSpace(envelope.DeviceKey), cfg.DeviceKey)
			req.Channel = fallback(strings.TrimSpace(envelope.Channel), cfg.Channel)
			req.Protocol = fallback(strings.TrimSpace(envelope.Protocol), cfg.Protocol)
			req.Raw = strings.TrimSpace(envelope.Raw)
			req.Payload = envelope.Payload
			if req.Raw == "" && len(envelope.Payload) > 0 {
				req.Raw = strings.TrimSpace(string(envelope.Payload))
			}
		}
	}
	if req.Raw == "" && len(req.Payload) == 0 {
		req.Raw = raw
	}
	if req.Raw == "" && len(req.Payload) == 0 {
		req.Raw = text
	}
	return g.processFrameRequest(cfg, deviceKey, req, "mqtt:"+topic)
}

func (g *DeviceGateway) processFrameRequest(cfg DeviceGatewayEndpointConfig, deviceKey string, req protocolFrameIngestRequest, remoteAddr string) (bool, string) {
	r := gatewayRequest(remoteAddr)
	session, ok := g.app.deviceSessionFromKey(strings.TrimSpace(deviceKey), clientIP(r))
	if !ok {
		frame := baseProtocolFrame(Session{User: User{Username: "gateway:" + cfg.Name}}, req, gatewayResource(cfg))
		frame.Error = "device key required or invalid"
		g.recordGatewayRejected(cfg, r, frame)
		return false, frame.Error
	}
	if !permissionGranted(session.DeviceScopes, cfg.Permission) {
		frame := baseProtocolFrame(session, req, gatewayResource(cfg))
		frame.DeviceNo = deviceNoFromSession(session)
		frame.Error = "permission denied: " + cfg.Permission
		g.recordGatewayRejected(cfg, r, frame)
		return false, frame.Error
	}
	var response protocolIngestResponse
	var err error
	switch cfg.Parser {
	case "gps":
		response, err = g.app.processGPSProtocolFrame(r, session, req)
	case "scale":
		response, err = g.app.processScaleProtocolFrame(r, session, req)
	case "plant":
		response, err = g.app.processPlantProtocolFrame(r, session, req)
	default:
		err = fmt.Errorf("unknown parser: %s", cfg.Parser)
	}
	if err != nil {
		g.recordRejected(cfg, err.Error())
		return false, err.Error()
	}
	g.recordAccepted(cfg)
	return true, response.Frame.FrameNo
}

func (g *DeviceGateway) recordGatewayRejected(cfg DeviceGatewayEndpointConfig, r *http.Request, frame DeviceProtocolFrame) {
	frame.Status = "rejected"
	saved, err := g.app.recordDeviceProtocolFrame(r, frame)
	if err == nil {
		g.app.emit("device.protocol.frame", saved)
	}
	g.recordRejected(cfg, frame.Error)
}

func extractGatewayLine(currentKey string, line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return currentKey, "", false
	}
	if strings.HasPrefix(strings.ToUpper(line), "AUTH ") {
		return strings.TrimSpace(line[5:]), "", true
	}
	if strings.HasPrefix(line, "KEY=") {
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(strings.TrimPrefix(parts[0], "KEY=")), strings.TrimSpace(parts[1]), false
		}
		return "", "", false
	}
	return currentKey, line, false
}

func gatewayRequest(remoteAddr string) *http.Request {
	return &http.Request{RemoteAddr: remoteAddr, Header: http.Header{}}
}

func gatewayResource(cfg DeviceGatewayEndpointConfig) string {
	if cfg.Parser == "scale" {
		return "scale_device_event"
	}
	return "vehicle_location"
}

func gatewayAddress(cfg DeviceGatewayEndpointConfig) string {
	if cfg.Kind == "mqtt" {
		return maskedBrokerURL(cfg.BrokerURL)
	}
	return cfg.Address
}

func gatewaySource(cfg DeviceGatewayEndpointConfig) string {
	if cfg.Kind == "mqtt" {
		return cfg.Topic
	}
	return cfg.FilePath
}

func maskedBrokerURL(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "<configured>"
	}
	if parsed.User != nil {
		parsed.User = url.User("configured")
	}
	return parsed.String()
}

func parseGatewayQOS(raw string) byte {
	switch strings.TrimSpace(raw) {
	case "1":
		return 1
	case "2":
		return 2
	default:
		return 0
	}
}

func (g *DeviceGateway) setStatus(name string, status string, lastError string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	item := g.statuses[name]
	item.Status = status
	item.LastError = lastError
	g.statuses[name] = item
}

func (g *DeviceGateway) recordAccepted(cfg DeviceGatewayEndpointConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	item := g.statuses[cfg.Name]
	item.Status = "online"
	item.LastError = ""
	item.AcceptedFrames++
	item.LastFrameAt = nowString()
	g.statuses[cfg.Name] = item
}

func (g *DeviceGateway) recordRejected(cfg DeviceGatewayEndpointConfig, lastError string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	item := g.statuses[cfg.Name]
	item.Status = "online"
	item.LastError = lastError
	item.RejectedFrames++
	item.LastFrameAt = nowString()
	g.statuses[cfg.Name] = item
}
