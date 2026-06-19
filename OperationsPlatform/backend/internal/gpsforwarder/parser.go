package forwarder

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func ParseLocation(raw, source, channel, defaultProtocol string, now time.Time) (Location, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Location{}, fmt.Errorf("empty location frame")
	}
	if strings.HasPrefix(raw, "{") {
		return parseJSONLocation(raw, source, channel, defaultProtocol, now)
	}
	return parseCSVLocation(raw, source, channel, defaultProtocol, now)
}

func parseJSONLocation(raw, source, channel, defaultProtocol string, now time.Time) (Location, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return Location{}, fmt.Errorf("invalid gps json frame")
	}
	protocol := firstString(data, "protocol")
	if protocol == "" {
		protocol = defaultProtocol
	}
	loc := newLocation(raw, source, channel, protocol, now)
	loc.DeviceNo = firstString(data, "deviceNo", "device", "deviceId", "terminalNo", "imei")
	loc.PlateNo = firstString(data, "plateNo", "plate", "vehicleNo")
	loc.Longitude = firstFloat(data, "longitude", "lon", "lng", "x")
	loc.Latitude = firstFloat(data, "latitude", "lat", "y")
	loc.Speed = firstFloat(data, "speed", "velocity")
	loc.Direction = firstFloat(data, "direction", "course", "heading")
	loc.Mileage = firstFloat(data, "mileage", "odometer")
	loc.AccStatus = int(firstFloat(data, "accStatus", "acc", "ignition"))
	if ts := firstString(data, "locationTime", "time", "timestamp", "gpsTime"); ts != "" {
		loc.LocationTime = normalizeTime(ts, now)
	}
	return finalizeLocation(loc)
}

func parseCSVLocation(raw, source, channel, defaultProtocol string, now time.Time) (Location, error) {
	reader := csv.NewReader(strings.NewReader(raw))
	reader.FieldsPerRecord = -1
	fields, err := reader.Read()
	if err != nil && err != io.EOF {
		return Location{}, fmt.Errorf("invalid gps csv frame")
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	if len(fields) < 8 {
		return Location{}, fmt.Errorf("gps csv frame requires at least 8 fields")
	}
	offset := 0
	protocol := defaultProtocol
	if marker := strings.ToUpper(fields[0]); marker == "GPS" || marker == "LOCATION" || marker == "BD" {
		protocol = strings.ToLower(marker) + "-csv"
		offset = 1
	}
	if len(fields) < offset+8 {
		return Location{}, fmt.Errorf("gps csv frame requires device, plate, lon, lat, speed, direction, mileage, acc")
	}
	loc := newLocation(raw, source, channel, protocol, now)
	loc.DeviceNo = fields[offset]
	loc.PlateNo = fields[offset+1]
	loc.Longitude = parseFloat(fields[offset+2])
	loc.Latitude = parseFloat(fields[offset+3])
	loc.Speed = parseFloat(fields[offset+4])
	loc.Direction = parseFloat(fields[offset+5])
	loc.Mileage = parseFloat(fields[offset+6])
	loc.AccStatus = int(parseFloat(fields[offset+7]))
	if len(fields) > offset+8 && fields[offset+8] != "" {
		loc.LocationTime = normalizeTime(fields[offset+8], now)
	}
	return finalizeLocation(loc)
}

func finalizeLocation(loc Location) (Location, error) {
	loc.DeviceNo = strings.TrimSpace(loc.DeviceNo)
	loc.PlateNo = strings.TrimSpace(loc.PlateNo)
	if loc.DeviceNo == "" {
		return Location{}, fmt.Errorf("deviceNo is required")
	}
	if loc.Longitude < -180 || loc.Longitude > 180 || loc.Latitude < -90 || loc.Latitude > 90 {
		return Location{}, fmt.Errorf("invalid coordinate")
	}
	if loc.Longitude == 0 && loc.Latitude == 0 {
		return Location{}, fmt.Errorf("zero coordinate is not accepted")
	}
	if loc.Speed < 0 {
		return Location{}, fmt.Errorf("speed must be >= 0")
	}
	if loc.Direction < 0 || loc.Direction >= 360 {
		loc.Direction = 0
	}
	return loc, nil
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

func firstFloat(data map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			switch v := value.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case json.Number:
				n, _ := v.Float64()
				return n
			case string:
				return parseFloat(v)
			}
		}
	}
	return 0
}

func parseFloat(value string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return n
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
