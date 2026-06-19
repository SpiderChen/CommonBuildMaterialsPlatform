package gateway

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

var knownJSONFields = map[string]bool{
	"asset": true, "assetNo": true, "channel": true, "device": true, "deviceId": true,
	"deviceNo": true, "eventTime": true, "eventType": true, "kind": true, "line": true,
	"measurements": true, "protocol": true, "readings": true, "source": true,
	"station": true, "stationNo": true, "timestamp": true, "time": true, "ts": true, "type": true,
}

func ParseFrame(raw, source, channel, defaultProtocol string, now time.Time) (Message, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Message{}, fmt.Errorf("empty frame")
	}
	if strings.HasPrefix(raw, "{") {
		return parseJSONFrame(raw, source, channel, defaultProtocol, now)
	}
	return parseCSVFrame(raw, source, channel, defaultProtocol, now)
}

func parseJSONFrame(raw, source, channel, defaultProtocol string, now time.Time) (Message, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return Message{}, fmt.Errorf("invalid json frame")
	}
	protocol := firstString(data, "protocol")
	if protocol == "" {
		protocol = defaultProtocol
	}
	msg := newMessage(raw, source, channel, protocol, now)
	msg.DeviceNo = firstString(data, "deviceNo", "device", "deviceId", "stationNo")
	msg.AssetNo = firstString(data, "assetNo", "asset", "line", "station")
	msg.EventType = fallbackString(firstString(data, "eventType", "type", "kind"), "telemetry")
	if ts := firstString(data, "eventTime", "timestamp", "time", "ts"); ts != "" {
		msg.EventTime = normalizeTime(ts, now)
	}
	msg.Readings = appendReadings(msg.Readings, data["readings"])
	msg.Readings = appendReadings(msg.Readings, data["measurements"])
	for key, value := range data {
		if knownJSONFields[key] {
			continue
		}
		if n, ok := numeric(value); ok {
			msg.Readings = append(msg.Readings, Reading{Name: key, Value: n})
			continue
		}
		if s, ok := value.(string); ok && s != "" {
			msg.Tags[key] = s
		}
	}
	return finalizeMessage(msg)
}

func parseCSVFrame(raw, source, channel, defaultProtocol string, now time.Time) (Message, error) {
	reader := csv.NewReader(strings.NewReader(raw))
	reader.FieldsPerRecord = -1
	fields, err := reader.Read()
	if err != nil && err != io.EOF {
		return Message{}, fmt.Errorf("invalid csv frame")
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	if len(fields) < 4 {
		return Message{}, fmt.Errorf("csv frame requires at least 4 fields")
	}
	protocol := defaultProtocol
	offset := 0
	if marker := strings.ToUpper(fields[0]); marker == "ICG" || marker == "PLC" || marker == "OPC" || marker == "MODBUS" || marker == "PLANT" {
		protocol = strings.ToLower(marker) + "-csv"
		offset = 1
	}
	if len(fields) < offset+4 {
		return Message{}, fmt.Errorf("csv frame requires device, asset, event and readings")
	}
	msg := newMessage(raw, source, channel, protocol, now)
	msg.DeviceNo = fields[offset]
	msg.AssetNo = fields[offset+1]
	msg.EventType = fallbackString(fields[offset+2], "telemetry")
	msg.Readings = parseReadingList(fields[offset+3])
	if len(fields) > offset+4 && fields[offset+4] != "" {
		msg.EventTime = normalizeTime(fields[offset+4], now)
	}
	return finalizeMessage(msg)
}

func appendReadings(existing []Reading, value any) []Reading {
	switch v := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			item := v[key]
			if n, ok := numeric(item); ok {
				existing = append(existing, Reading{Name: key, Value: n})
				continue
			}
			if nested, ok := item.(map[string]any); ok {
				if n, ok := numeric(nested["value"]); ok {
					existing = append(existing, Reading{
						Name:    key,
						Value:   n,
						Unit:    fmt.Sprint(nested["unit"]),
						Quality: fmt.Sprint(nested["quality"]),
					})
				}
			}
		}
	case []any:
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name := firstString(m, "name", "code", "point")
			n, ok := numeric(m["value"])
			if name == "" || !ok {
				continue
			}
			existing = append(existing, Reading{
				Name:    name,
				Value:   n,
				Unit:    firstString(m, "unit"),
				Quality: firstString(m, "quality"),
			})
		}
	}
	return existing
}

func parseReadingList(raw string) []Reading {
	parts := strings.Split(raw, "|")
	readings := make([]Reading, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, rest, ok := strings.Cut(part, "=")
		if !ok {
			name, rest, ok = strings.Cut(part, ":")
		}
		if !ok {
			continue
		}
		segs := strings.Split(rest, ":")
		value, err := strconv.ParseFloat(strings.TrimSpace(segs[0]), 64)
		if err != nil {
			continue
		}
		reading := Reading{Name: strings.TrimSpace(name), Value: value}
		if len(segs) > 1 {
			reading.Unit = strings.TrimSpace(segs[1])
		}
		if len(segs) > 2 {
			reading.Quality = strings.TrimSpace(segs[2])
		}
		readings = append(readings, reading)
	}
	return readings
}

func finalizeMessage(msg Message) (Message, error) {
	msg.DeviceNo = strings.TrimSpace(msg.DeviceNo)
	if msg.DeviceNo == "" {
		return Message{}, fmt.Errorf("deviceNo is required")
	}
	if len(msg.Readings) == 0 {
		return Message{}, fmt.Errorf("at least one numeric reading is required")
	}
	if len(msg.Tags) == 0 {
		msg.Tags = nil
	}
	for i := range msg.Readings {
		msg.Readings[i].Name = strings.TrimSpace(msg.Readings[i].Name)
		msg.Readings[i].Unit = strings.TrimSpace(msg.Readings[i].Unit)
		msg.Readings[i].Quality = strings.TrimSpace(msg.Readings[i].Quality)
	}
	return msg, nil
}

func firstString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if s := strings.TrimSpace(fmt.Sprint(value)); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func numeric(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func normalizeTime(value string, now time.Time) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return now.UTC().Format(time.RFC3339)
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006/01/02 15:04:05", "2006-01-02T15:04:05"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return value
}
