package gateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Forwarder interface {
	Forward(context.Context, Message) error
}

func NewForwarder(cfg Config, logger *log.Logger) (Forwarder, error) {
	forwarders := []Forwarder{}
	for _, target := range cfg.HTTPTargets {
		forwarders = append(forwarders, &HTTPForwarder{
			target:     target,
			token:      cfg.HTTPBearer,
			secret:     cfg.SharedSecret,
			mode:       strings.ToLower(cfg.ForwardMode),
			includeRaw: cfg.IncludeRaw,
			client:     &http.Client{Timeout: cfg.HTTPTimeout},
		})
	}
	if cfg.OutputFile != "" {
		fileForwarder, err := NewFileForwarder(cfg.OutputFile, cfg.IncludeRaw)
		if err != nil {
			return nil, err
		}
		forwarders = append(forwarders, fileForwarder)
	}
	if len(forwarders) == 0 {
		forwarders = append(forwarders, &StdoutForwarder{logger: logger, includeRaw: cfg.IncludeRaw})
	}
	return MultiForwarder(forwarders), nil
}

type MultiForwarder []Forwarder

func (m MultiForwarder) Forward(ctx context.Context, msg Message) error {
	var errs []string
	for _, forwarder := range m {
		if err := forwarder.Forward(ctx, msg); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

type HTTPForwarder struct {
	target     string
	token      string
	secret     string
	mode       string
	includeRaw bool
	client     *http.Client
}

func (h *HTTPForwarder) Forward(ctx context.Context, msg Message) error {
	body, err := h.body(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.target, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	if h.secret != "" {
		ts := time.Now().UTC().Format(time.RFC3339)
		req.Header.Set("X-CBMP-Timestamp", ts)
		req.Header.Set("X-CBMP-Signature", signBody(h.secret, ts, body))
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http target %s returned %s", h.target, resp.Status)
	}
	return nil
}

func (h *HTTPForwarder) body(msg Message) ([]byte, error) {
	if h.mode == "protocol-frame" {
		envelope := ProtocolFrameEnvelope{
			Channel:  msg.Channel,
			Protocol: msg.Protocol,
			Raw:      msg.Raw,
		}
		if payload, ok := plantProtocolPayload(msg); ok {
			envelope.Payload = payload
		} else if payload, ok := bufferProtocolPayload(msg); ok {
			envelope.Payload = payload
		} else if payload, ok := yardProtocolPayload(msg); ok {
			envelope.Payload = payload
		}
		return json.Marshal(envelope)
	}
	if !h.includeRaw {
		msg = msg.withoutRaw()
	}
	return json.Marshal(msg)
}

type erpPlantPayload struct {
	DeviceNo      string  `json:"deviceNo,omitempty"`
	TaskID        int64   `json:"taskId,omitempty"`
	TaskNo        string  `json:"taskNo,omitempty"`
	BatchNo       string  `json:"batchNo,omitempty"`
	PlantCode     string  `json:"plantCode,omitempty"`
	Quantity      float64 `json:"quantity,omitempty"`
	Operator      string  `json:"operator,omitempty"`
	QualityStatus string  `json:"qualityStatus,omitempty"`
	Status        string  `json:"status,omitempty"`
	StartedAt     string  `json:"startedAt,omitempty"`
	CompletedAt   string  `json:"completedAt,omitempty"`
}

func plantProtocolPayload(msg Message) (json.RawMessage, bool) {
	if !isPlantProtocolMessage(msg) {
		return nil, false
	}
	payload := erpPlantPayload{
		DeviceNo:      strings.TrimSpace(msg.DeviceNo),
		TaskNo:        firstTag(msg.Tags, "taskNo", "task_no", "task"),
		BatchNo:       firstTag(msg.Tags, "batchNo", "batch_no", "batch"),
		PlantCode:     firstTag(msg.Tags, "plantCode", "plant_code", "plant", "line", "station"),
		Operator:      firstTag(msg.Tags, "operator", "operatorName"),
		QualityStatus: firstTag(msg.Tags, "qualityStatus", "quality_status", "quality"),
		Status:        firstTag(msg.Tags, "status", "batchStatus", "batch_status"),
		StartedAt:     firstTag(msg.Tags, "startedAt", "started_at", "startTime", "start_time"),
		CompletedAt:   firstTag(msg.Tags, "completedAt", "completed_at", "completeTime", "complete_time"),
	}
	if payload.PlantCode == "" {
		payload.PlantCode = strings.TrimSpace(msg.AssetNo)
	}
	if payload.TaskID == 0 {
		payload.TaskID = int64Reading(msg.Readings, "taskId", "task_id", "task")
	}
	if payload.TaskID == 0 {
		payload.TaskID = int64Tag(msg.Tags, "taskId", "task_id")
	}
	if payload.Quantity == 0 {
		if quantity, ok := floatReading(msg.Readings, "quantity", "qty", "batchQty", "batchQuantity", "producedQty", "actualQuantity", "volume"); ok {
			payload.Quantity = quantity
		}
	}
	if payload.CompletedAt == "" {
		payload.CompletedAt = strings.TrimSpace(msg.EventTime)
	}
	if payload.PlantCode == "" || payload.Quantity <= 0 || (payload.TaskID == 0 && payload.TaskNo == "") {
		return nil, false
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}
	return body, true
}

type erpBufferPayload struct {
	DeviceNo      string  `json:"deviceNo,omitempty"`
	PlantCode     string  `json:"plantCode,omitempty"`
	BufferCode    string  `json:"bufferCode,omitempty"`
	MaterialID    int64   `json:"materialId,omitempty"`
	Quantity      float64 `json:"quantity"`
	MoistureRate  float64 `json:"moistureRate,omitempty"`
	QualityStatus string  `json:"qualityStatus,omitempty"`
	Status        string  `json:"status,omitempty"`
	ReportedAt    string  `json:"reportedAt,omitempty"`
}

func bufferProtocolPayload(msg Message) (json.RawMessage, bool) {
	if !isBufferProtocolMessage(msg) {
		return nil, false
	}
	payload := erpBufferPayload{
		DeviceNo:      strings.TrimSpace(msg.DeviceNo),
		PlantCode:     firstTag(msg.Tags, "plantCode", "plant_code", "plant", "line", "station"),
		BufferCode:    firstTag(msg.Tags, "bufferCode", "buffer_code", "binCode", "bin_code", "bin", "silo"),
		QualityStatus: firstTag(msg.Tags, "qualityStatus", "quality_status", "quality"),
		Status:        firstTag(msg.Tags, "status", "binStatus", "bin_status"),
		ReportedAt:    firstTag(msg.Tags, "reportedAt", "reported_at", "eventTime", "time", "timestamp"),
	}
	if payload.PlantCode == "" {
		payload.PlantCode = strings.TrimSpace(msg.AssetNo)
	}
	if payload.MaterialID == 0 {
		payload.MaterialID = int64Reading(msg.Readings, "materialId", "material_id")
	}
	if payload.MaterialID == 0 {
		payload.MaterialID = int64Tag(msg.Tags, "materialId", "material_id")
	}
	if quantity, ok := floatReading(msg.Readings, "quantity", "qty", "currentQty", "currentQuantity", "levelQty", "binQty", "weight"); ok {
		payload.Quantity = quantity
	}
	if moisture, ok := floatReading(msg.Readings, "moistureRate", "moisture_rate", "moisture", "waterRate"); ok {
		payload.MoistureRate = moisture
	}
	if payload.ReportedAt == "" {
		payload.ReportedAt = strings.TrimSpace(msg.EventTime)
	}
	if payload.BufferCode == "" || payload.Quantity < 0 {
		return nil, false
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}
	return body, true
}

type erpYardPayload struct {
	DeviceNo      string  `json:"deviceNo,omitempty"`
	YardCode      string  `json:"yardCode,omitempty"`
	PileCode      string  `json:"pileCode,omitempty"`
	MaterialID    int64   `json:"materialId,omitempty"`
	Quantity      float64 `json:"quantity"`
	MoistureRate  float64 `json:"moistureRate,omitempty"`
	QualityStatus string  `json:"qualityStatus,omitempty"`
	Status        string  `json:"status,omitempty"`
	ReportedAt    string  `json:"reportedAt,omitempty"`
}

func yardProtocolPayload(msg Message) (json.RawMessage, bool) {
	if !isYardProtocolMessage(msg) {
		return nil, false
	}
	payload := erpYardPayload{
		DeviceNo:      strings.TrimSpace(msg.DeviceNo),
		YardCode:      firstTag(msg.Tags, "yardCode", "yard_code", "yard", "stockYard", "stock_yard"),
		PileCode:      firstTag(msg.Tags, "pileCode", "pile_code", "pile", "stockpileCode", "stockpile_code", "stockpile"),
		QualityStatus: firstTag(msg.Tags, "qualityStatus", "quality_status", "quality"),
		Status:        firstTag(msg.Tags, "status", "pileStatus", "pile_status"),
		ReportedAt:    firstTag(msg.Tags, "reportedAt", "reported_at", "eventTime", "time", "timestamp"),
	}
	if payload.YardCode == "" {
		payload.YardCode = strings.TrimSpace(msg.AssetNo)
	}
	if payload.MaterialID == 0 {
		payload.MaterialID = int64Reading(msg.Readings, "materialId", "material_id")
	}
	if payload.MaterialID == 0 {
		payload.MaterialID = int64Tag(msg.Tags, "materialId", "material_id")
	}
	if quantity, ok := floatReading(msg.Readings, "quantity", "qty", "currentQty", "currentQuantity", "levelQty", "pileQty", "stockpileQty", "weight"); ok {
		payload.Quantity = quantity
	}
	if moisture, ok := floatReading(msg.Readings, "moistureRate", "moisture_rate", "moisture", "waterRate"); ok {
		payload.MoistureRate = moisture
	}
	if payload.ReportedAt == "" {
		payload.ReportedAt = strings.TrimSpace(msg.EventTime)
	}
	if payload.PileCode == "" || payload.Quantity < 0 {
		return nil, false
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}
	return body, true
}

func isPlantProtocolMessage(msg Message) bool {
	protocol := strings.ToLower(strings.TrimSpace(msg.Protocol))
	eventType := strings.ToLower(strings.TrimSpace(msg.EventType))
	return strings.HasPrefix(protocol, "plant") || strings.Contains(protocol, "plant-") || eventType == "batch" || eventType == "production_batch"
}

func isBufferProtocolMessage(msg Message) bool {
	protocol := strings.ToLower(strings.TrimSpace(msg.Protocol))
	eventType := strings.ToLower(strings.TrimSpace(msg.EventType))
	return strings.HasPrefix(protocol, "buffer") || strings.Contains(protocol, "buffer-") || eventType == "buffer_level" || eventType == "bin_level" || eventType == "silo_level"
}

func isYardProtocolMessage(msg Message) bool {
	protocol := strings.ToLower(strings.TrimSpace(msg.Protocol))
	eventType := strings.ToLower(strings.TrimSpace(msg.EventType))
	return strings.HasPrefix(protocol, "yard") || strings.Contains(protocol, "yard-") || eventType == "yard_level" || eventType == "stockpile_level" || eventType == "pile_level"
}

func firstTag(tags map[string]string, keys ...string) string {
	for _, key := range keys {
		for tagKey, value := range tags {
			if sameGatewayFieldName(tagKey, key) {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func int64Tag(tags map[string]string, keys ...string) int64 {
	value := firstTag(tags, keys...)
	if value == "" {
		return 0
	}
	out, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return out
}

func int64Reading(readings []Reading, names ...string) int64 {
	value, ok := floatReading(readings, names...)
	if !ok {
		return 0
	}
	return int64(value)
}

func floatReading(readings []Reading, names ...string) (float64, bool) {
	for _, name := range names {
		for _, reading := range readings {
			if sameGatewayFieldName(reading.Name, name) {
				return reading.Value, true
			}
		}
	}
	return 0, false
}

func sameGatewayFieldName(left string, right string) bool {
	return normalizeGatewayFieldName(left) == normalizeGatewayFieldName(right)
}

func normalizeGatewayFieldName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Map(func(r rune) rune {
		if r == '_' || r == '-' || r == ' ' {
			return -1
		}
		return r
	}, value)
}

type FileForwarder struct {
	path       string
	includeRaw bool
	mu         sync.Mutex
}

func NewFileForwarder(path string, includeRaw bool) (*FileForwarder, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &FileForwarder{path: path, includeRaw: includeRaw}, nil
}

func (f *FileForwarder) Forward(_ context.Context, msg Message) error {
	if !f.includeRaw {
		msg = msg.withoutRaw()
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(body, '\n'))
	return err
}

type StdoutForwarder struct {
	logger     *log.Logger
	includeRaw bool
}

func (s *StdoutForwarder) Forward(_ context.Context, msg Message) error {
	if !s.includeRaw {
		msg = msg.withoutRaw()
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.logger.Print(string(body))
	return nil
}

func signBody(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
