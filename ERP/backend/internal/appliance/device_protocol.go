package appliance

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type locationReportPayload struct {
	DeviceNo     string  `json:"deviceNo"`
	PlateNo      string  `json:"plateNo"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	Speed        float64 `json:"speed"`
	Direction    float64 `json:"direction"`
	Mileage      float64 `json:"mileage"`
	AccStatus    int     `json:"accStatus"`
	LocationTime string  `json:"locationTime"`
	SourceType   string  `json:"sourceType"`
}

type protocolFrameIngestRequest struct {
	Channel  string          `json:"channel"`
	Protocol string          `json:"protocol"`
	Raw      string          `json:"raw"`
	Payload  json.RawMessage `json:"payload"`
}

type protocolIngestResponse struct {
	Frame           DeviceProtocolFrame   `json:"frame"`
	Location        *VehicleLocationEvent `json:"location,omitempty"`
	ScaleEvent      *ScaleDeviceEvent     `json:"scaleEvent,omitempty"`
	ProductionBatch *ProductionBatch      `json:"productionBatch,omitempty"`
}

func (a *App) ingestGPSProtocolFrame(w http.ResponseWriter, r *http.Request, session Session) {
	req, err := readProtocolFrameIngestRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Channel = fallback(strings.TrimSpace(req.Channel), "mqtt")
	req.Protocol = fallback(strings.TrimSpace(req.Protocol), "gps-json")
	response, err := a.processGPSProtocolFrame(r, session, req)
	if err != nil {
		writeProtocolFrameError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) processGPSProtocolFrame(r *http.Request, session Session, req protocolFrameIngestRequest) (protocolIngestResponse, error) {
	location, err := parseGPSProtocolFrame(req)
	frame := baseProtocolFrame(session, req, "vehicle_location")
	frame.DeviceNo = fallback(location.DeviceNo, deviceNoFromSession(session))
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	event, latest, err := a.recordLocationReport(r, session, location)
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	a.runtime.CacheLatestLocation(latest)
	a.runtime.StoreTrackPoint(event)
	frame.ParsedID = event.ID
	frame.Status = "accepted"
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		return protocolIngestResponse{}, err
	}
	a.emit("vehicle.location.update", event)
	a.emit("device.protocol.frame", saved)
	return protocolIngestResponse{Frame: saved, Location: &event}, nil
}

func (a *App) ingestScaleProtocolFrame(w http.ResponseWriter, r *http.Request, session Session) {
	req, err := readProtocolFrameIngestRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Channel = fallback(strings.TrimSpace(req.Channel), "serial")
	req.Protocol = fallback(strings.TrimSpace(req.Protocol), "scale-csv")
	response, err := a.processScaleProtocolFrame(r, session, req)
	if err != nil {
		writeProtocolFrameError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) processScaleProtocolFrame(r *http.Request, session Session, req protocolFrameIngestRequest) (protocolIngestResponse, error) {
	scaleEvent, err := parseScaleProtocolFrame(req)
	frame := baseProtocolFrame(session, req, "scale_device_event")
	frame.DeviceNo = fallback(scaleEvent.DeviceCode, deviceNoFromSession(session))
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	item, err := a.recordScaleDeviceEvent(r, session, scaleEvent)
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	frame.ParsedID = item.ID
	frame.Status = "accepted"
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		return protocolIngestResponse{}, err
	}
	a.emit("scale.device_event.reported", item)
	a.emit("device.protocol.frame", saved)
	return protocolIngestResponse{Frame: saved, ScaleEvent: &item}, nil
}

func readProtocolFrameIngestRequest(r *http.Request) (protocolFrameIngestRequest, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		return protocolFrameIngestRequest{}, fmt.Errorf("read protocol frame failed")
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return protocolFrameIngestRequest{}, fmt.Errorf("protocol frame is empty")
	}
	if strings.HasPrefix(text, "{") {
		var req protocolFrameIngestRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return protocolFrameIngestRequest{}, fmt.Errorf("invalid protocol frame json")
		}
		if strings.TrimSpace(req.Raw) == "" && len(req.Payload) > 0 {
			req.Raw = strings.TrimSpace(string(req.Payload))
		}
		if strings.TrimSpace(req.Raw) == "" {
			req.Raw = text
		}
		return req, nil
	}
	return protocolFrameIngestRequest{Raw: text}, nil
}

func parseGPSProtocolFrame(req protocolFrameIngestRequest) (locationReportPayload, error) {
	raw := strings.TrimSpace(req.Raw)
	var payload locationReportPayload
	if len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return payload, fmt.Errorf("invalid gps json payload")
		}
		return normalizeLocationPayload(payload, req), nil
	}
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return payload, fmt.Errorf("invalid gps json frame")
		}
		return normalizeLocationPayload(payload, req), nil
	}
	fields, err := parseProtocolCSV(raw)
	if err != nil {
		return payload, fmt.Errorf("invalid gps csv frame: %w", err)
	}
	offset := 0
	if len(fields) > 0 && (strings.EqualFold(fields[0], "GPS") || strings.EqualFold(fields[0], "LOCATION")) {
		offset = 1
	}
	if len(fields)-offset < 8 {
		return payload, fmt.Errorf("gps frame requires device, plate, lon, lat, speed, direction, mileage and acc")
	}
	payload.DeviceNo = fields[offset]
	payload.PlateNo = fields[offset+1]
	payload.Longitude, err = parseFloatField(fields[offset+2], "longitude")
	if err != nil {
		return payload, err
	}
	payload.Latitude, err = parseFloatField(fields[offset+3], "latitude")
	if err != nil {
		return payload, err
	}
	payload.Speed, err = parseFloatField(fields[offset+4], "speed")
	if err != nil {
		return payload, err
	}
	payload.Direction, err = parseFloatField(fields[offset+5], "direction")
	if err != nil {
		return payload, err
	}
	payload.Mileage, err = parseFloatField(fields[offset+6], "mileage")
	if err != nil {
		return payload, err
	}
	acc, err := strconv.Atoi(fields[offset+7])
	if err != nil {
		return payload, fmt.Errorf("invalid accStatus")
	}
	payload.AccStatus = acc
	if len(fields)-offset > 8 {
		payload.LocationTime = fields[offset+8]
	}
	return normalizeLocationPayload(payload, req), nil
}

func parseScaleProtocolFrame(req protocolFrameIngestRequest) (ScaleDeviceEvent, error) {
	raw := strings.TrimSpace(req.Raw)
	var payload ScaleDeviceEvent
	if len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return payload, fmt.Errorf("invalid scale json payload")
		}
		return normalizeScalePayload(payload), nil
	}
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return payload, fmt.Errorf("invalid scale json frame")
		}
		return normalizeScalePayload(payload), nil
	}
	fields, err := parseProtocolCSV(raw)
	if err != nil {
		return payload, fmt.Errorf("invalid scale csv frame: %w", err)
	}
	offset := 0
	if len(fields) > 0 && strings.EqualFold(fields[0], "SCALE") {
		offset = 1
	}
	if len(fields)-offset < 7 {
		return payload, fmt.Errorf("scale frame requires device, ticket, plate, recognizedPlate, weight, weightType and stable")
	}
	payload.DeviceCode = fields[offset]
	if fields[offset+1] != "" {
		ticketID, err := strconv.ParseInt(fields[offset+1], 10, 64)
		if err != nil {
			return payload, fmt.Errorf("invalid ticketId")
		}
		payload.TicketID = ticketID
	}
	payload.PlateNo = fields[offset+2]
	payload.RecognizedPlateNo = fields[offset+3]
	payload.Weight, err = parseFloatField(fields[offset+4], "weight")
	if err != nil {
		return payload, err
	}
	payload.WeightType = fields[offset+5]
	payload.Stable, err = parseBoolField(fields[offset+6], "stable")
	if err != nil {
		return payload, err
	}
	if len(fields)-offset > 7 {
		payload.SnapshotURL = fields[offset+7]
	}
	return normalizeScalePayload(payload), nil
}

func normalizeLocationPayload(payload locationReportPayload, req protocolFrameIngestRequest) locationReportPayload {
	payload.DeviceNo = strings.TrimSpace(payload.DeviceNo)
	payload.PlateNo = strings.TrimSpace(payload.PlateNo)
	payload.SourceType = fallback(strings.TrimSpace(payload.SourceType), strings.TrimSpace(req.Channel)+"_"+strings.TrimSpace(req.Protocol))
	if payload.SourceType == "_" {
		payload.SourceType = "gps_protocol"
	}
	return payload
}

func normalizeScalePayload(payload ScaleDeviceEvent) ScaleDeviceEvent {
	payload.DeviceCode = strings.TrimSpace(payload.DeviceCode)
	payload.PlateNo = strings.TrimSpace(payload.PlateNo)
	payload.RecognizedPlateNo = strings.TrimSpace(payload.RecognizedPlateNo)
	payload.WeightType = strings.TrimSpace(payload.WeightType)
	payload.SnapshotURL = strings.TrimSpace(payload.SnapshotURL)
	return payload
}

func parseProtocolCSV(raw string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(raw))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	fields, err := reader.Read()
	if err != nil {
		return nil, err
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields, nil
}

func parseFloatField(value string, name string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return parsed, nil
}

func parseBoolField(value string, name string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "stable", "ok":
		return true, nil
	case "0", "false", "no", "n", "unstable", "fail":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s", name)
	}
}

func baseProtocolFrame(session Session, req protocolFrameIngestRequest, resource string) DeviceProtocolFrame {
	return DeviceProtocolFrame{
		Channel:        strings.TrimSpace(req.Channel),
		Protocol:       strings.TrimSpace(req.Protocol),
		Raw:            truncateProtocolRaw(req.Raw),
		ParsedResource: resource,
		Status:         "rejected",
		Actor:          session.User.Username,
	}
}

func (a *App) rejectProtocolFrame(w http.ResponseWriter, r *http.Request, frame DeviceProtocolFrame, cause error) {
	frame.Status = "rejected"
	frame.Error = cause.Error()
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.emit("device.protocol.frame", saved)
	writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": cause.Error(), "frame": saved})
}

type protocolFrameError struct {
	message string
	frame   DeviceProtocolFrame
}

func (e protocolFrameError) Error() string {
	return e.message
}

func writeProtocolFrameError(w http.ResponseWriter, err error) {
	var frameErr protocolFrameError
	if ok := asProtocolFrameError(err, &frameErr); ok {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": frameErr.message, "frame": frameErr.frame})
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func asProtocolFrameError(err error, target *protocolFrameError) bool {
	if err == nil {
		return false
	}
	if frameErr, ok := err.(protocolFrameError); ok {
		*target = frameErr
		return true
	}
	return false
}

func (a *App) rejectProtocolFrameData(r *http.Request, frame DeviceProtocolFrame, cause error) error {
	frame.Status = "rejected"
	frame.Error = cause.Error()
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		return err
	}
	a.emit("device.protocol.frame", saved)
	return protocolFrameError{message: cause.Error(), frame: saved}
}

func (a *App) recordDeviceProtocolFrame(r *http.Request, frame DeviceProtocolFrame) (DeviceProtocolFrame, error) {
	err := a.store.Mutate(func(data *AppData) error {
		frame.ID = nextID(data, "protocolFrame")
		frame.FrameNo = number("DPF", frame.ID)
		frame.ReceivedAt = fallback(frame.ReceivedAt, nowString())
		frame.Actor = fallback(frame.Actor, "system")
		data.DeviceProtocolFrames = append(data.DeviceProtocolFrames, frame)
		if len(data.DeviceProtocolFrames) > 1000 {
			data.DeviceProtocolFrames = data.DeviceProtocolFrames[len(data.DeviceProtocolFrames)-1000:]
		}
		addAudit(data, frame.Actor, "ingest", "device_protocol_frame", frame.ID, frame.Channel+"/"+frame.Protocol+"/"+frame.Status, clientIP(r))
		return nil
	})
	return frame, err
}

func truncateProtocolRaw(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= 4096 {
		return raw
	}
	return raw[:4096]
}

func deviceNoFromSession(session Session) string {
	return strings.TrimPrefix(session.User.Username, "device:")
}

func findVehicleByDeviceNo(data AppData, deviceNo string) (Vehicle, bool) {
	for _, device := range data.VehicleDevices {
		if device.DeviceNo != deviceNo {
			continue
		}
		return findVehicle(data, device.VehicleID)
	}
	return Vehicle{}, false
}
