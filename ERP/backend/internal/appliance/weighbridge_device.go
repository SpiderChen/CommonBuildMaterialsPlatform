package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) scaleDeviceEvents(w http.ResponseWriter, r *http.Request, session Session) {
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, data.ScaleDeviceEvents)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req ScaleDeviceEvent
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid scale device event")
		return
	}
	item, err := a.recordScaleDeviceEvent(r, session, req)
	a.respondMutation(w, err, item, "scale.device_event.reported")
}

func (a *App) recordScaleDeviceEvent(r *http.Request, session Session, req ScaleDeviceEvent) (ScaleDeviceEvent, error) {
	var item ScaleDeviceEvent
	err := a.store.Mutate(func(data *AppData) error {
		if strings.HasPrefix(session.User.Username, "device:") {
			deviceNo := strings.TrimPrefix(session.User.Username, "device:")
			if req.DeviceCode != "" && req.DeviceCode != deviceNo {
				return fmt.Errorf("device key does not match scale payload")
			}
			req.DeviceCode = deviceNo
		}
		device, ok := findScaleDeviceByCode(*data, req.DeviceCode)
		if !ok {
			return fmt.Errorf("地磅设备不存在")
		}
		if req.Weight <= 0 {
			return fmt.Errorf("设备重量必须大于 0")
		}
		req.WeightType = strings.TrimSpace(req.WeightType)
		if req.WeightType != "gross" && req.WeightType != "tare" {
			return fmt.Errorf("称重类型必须是 gross 或 tare")
		}
		var ticket ScaleTicket
		if req.TicketID != 0 {
			var found bool
			ticket, found = findScaleTicket(*data, req.TicketID)
			if !found {
				return fmt.Errorf("关联票据不存在")
			}
		}
		item = buildScaleDeviceEvent(data, device, ticket, req)
		data.ScaleDeviceEvents = append(data.ScaleDeviceEvents, item)
		if req.TicketID != 0 {
			applyScaleDeviceEventToTicket(data, item)
		}
		for i := range data.ScaleDevices {
			if data.ScaleDevices[i].ID == device.ID {
				data.ScaleDevices[i].Status = "online"
			}
		}
		addAudit(data, session.User.Username, "report", "scale_device_event", item.ID, item.EventNo, clientIP(r))
		return nil
	})
	return item, err
}

func buildScaleDeviceEvent(data *AppData, device ScaleDevice, ticket ScaleTicket, req ScaleDeviceEvent) ScaleDeviceEvent {
	reasons := []string{}
	recognizedPlate := strings.TrimSpace(req.RecognizedPlateNo)
	if recognizedPlate == "" {
		recognizedPlate = strings.TrimSpace(req.PlateNo)
	}
	if req.SnapshotURL == "" {
		reasons = append(reasons, "missing_snapshot")
	}
	if !req.Stable {
		reasons = append(reasons, "unstable_weight")
	}
	if ticket.ID != 0 {
		expectedPlate := strings.TrimSpace(ticket.PlateNo)
		if expectedPlate != "" && recognizedPlate != "" && expectedPlate != recognizedPlate {
			reasons = append(reasons, "plate_mismatch")
		}
		if ticket.SiteID != 0 && ticket.SiteID != device.SiteID {
			reasons = append(reasons, "site_mismatch")
		}
	}
	id := nextID(data, "scaleDeviceEvent")
	status := "accepted"
	if len(reasons) > 0 {
		status = "blocked"
	}
	return ScaleDeviceEvent{
		ID: id, EventNo: number("SDE", id), DeviceID: device.ID, DeviceCode: device.Code,
		TicketID: req.TicketID, PlateNo: strings.TrimSpace(req.PlateNo), RecognizedPlateNo: recognizedPlate,
		Weight: round(req.Weight), WeightType: req.WeightType, Stable: req.Stable, SnapshotURL: req.SnapshotURL,
		CheatFlag: len(reasons) > 0, CheatReason: strings.Join(reasons, ","), Status: status, ReceivedAt: nowString(),
	}
}

func applyScaleDeviceEventToTicket(data *AppData, event ScaleDeviceEvent) {
	for i := range data.ScaleTickets {
		if data.ScaleTickets[i].ID != event.TicketID {
			continue
		}
		if event.CheatFlag {
			data.ScaleTickets[i].Status = "abnormal"
		}
		if event.RecognizedPlateNo != "" {
			data.ScaleTickets[i].PlateNo = fallback(data.ScaleTickets[i].PlateNo, event.RecognizedPlateNo)
		}
		if event.WeightType == "gross" {
			data.ScaleTickets[i].GrossWeight = event.Weight
			data.ScaleTickets[i].SnapshotURL = event.SnapshotURL
		}
		if event.WeightType == "tare" {
			data.ScaleTickets[i].TareWeight = event.Weight
		}
		if data.ScaleTickets[i].GrossWeight > data.ScaleTickets[i].TareWeight {
			data.ScaleTickets[i].NetWeight = round(data.ScaleTickets[i].GrossWeight - data.ScaleTickets[i].TareWeight)
		}
		appendWeightRecord(data, event.DeviceID, event.TicketID, data.ScaleTickets[i].PlateNo, event.Weight, event.WeightType, event.SnapshotURL, event.ReceivedAt)
		return
	}
}

func findScaleDeviceByCode(data AppData, code string) (ScaleDevice, bool) {
	for _, item := range data.ScaleDevices {
		if item.Code == code {
			return item, true
		}
	}
	return ScaleDevice{}, false
}
