package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type deliveryNoteRequest struct {
	DispatchID int64  `json:"dispatchId"`
	TicketID   int64  `json:"ticketId"`
	Status     string `json:"status"`
}

type deliveryNoteStatusRequest struct {
	Status string `json:"status"`
}

func (a *App) listDeliveryNotes(w http.ResponseWriter, r *http.Request, session Session) {
	writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).DeliveryNotes)
}

func (a *App) getDeliveryNote(w http.ResponseWriter, r *http.Request, session Session, noteID int64) {
	for _, note := range scopedData(a.mustSnapshot(), session.User).DeliveryNotes {
		if note.ID == noteID {
			writeJSON(w, http.StatusOK, note)
			return
		}
	}
	writeError(w, http.StatusNotFound, "送货单不存在")
}

func (a *App) createDeliveryNote(w http.ResponseWriter, r *http.Request, session Session) {
	var req deliveryNoteRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid delivery note")
		return
	}
	var note DeliveryNote
	err := a.store.Mutate(func(data *AppData) error {
		next, err := upsertDeliveryNote(data, req)
		if err != nil {
			return err
		}
		note = next
		addAudit(data, session.User.Username, "upsert", "delivery_note", note.ID, note.NoteNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, note, "delivery.note.saved")
}

func (a *App) updateDeliveryNoteStatus(w http.ResponseWriter, r *http.Request, session Session, noteID int64) {
	var req deliveryNoteStatusRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid delivery note status")
		return
	}
	var note DeliveryNote
	status := normalizeDeliveryNoteStatus(req.Status)
	topic := "delivery.note.status_changed"
	err := a.store.Mutate(func(data *AppData) error {
		index := deliveryNoteIndex(*data, noteID)
		if index < 0 {
			return fmt.Errorf("送货单不存在")
		}
		current := data.DeliveryNotes[index]
		if current.Status == "signed" && status == "void" {
			return fmt.Errorf("已签收送货单不能作废")
		}
		if status == "" {
			return fmt.Errorf("送货单状态无效")
		}
		if hasPendingWorkflowForResource(*data, "delivery_note", noteID) {
			return fmt.Errorf("送货单正在工作流审批中")
		}
		if deliveryNoteStatusNeedsWorkflow(status) {
			_, instances, err := publishDeliveryNoteStatusWorkflow(data, current, status, session.User.Username)
			if err != nil {
				return err
			}
			if len(instances) > 0 {
				note = current
				topic = "delivery.note.status_requested"
				addAudit(data, session.User.Username, "request_status", "delivery_note", note.ID, status, clientIP(r))
				return nil
			}
		}
		next, err := applyDeliveryNoteStatusLocked(data, noteID, status)
		if err != nil {
			return err
		}
		note = next
		addAudit(data, session.User.Username, "status", "delivery_note", note.ID, note.Status, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, note, topic)
}

func (a *App) reprintDeliveryNote(w http.ResponseWriter, r *http.Request, session Session, noteID int64) {
	var note DeliveryNote
	err := a.store.Mutate(func(data *AppData) error {
		index := deliveryNoteIndex(*data, noteID)
		if index < 0 {
			return fmt.Errorf("送货单不存在")
		}
		note = data.DeliveryNotes[index]
		if note.TicketID != 0 {
			for i := range data.ScaleTickets {
				if data.ScaleTickets[i].ID == note.TicketID {
					data.ScaleTickets[i].PrintCount++
					break
				}
			}
			logID := nextID(data, "ticketPrintLog")
			data.TicketPrintLogs = append(data.TicketPrintLogs, TicketPrintLog{
				ID: logID, TicketID: note.TicketID, PrintedBy: session.User.Username, PrintedAt: nowString(),
			})
		}
		addAudit(data, session.User.Username, "reprint", "delivery_note", note.ID, note.NoteNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, note, "delivery.note.reprinted")
}

func (a *App) createDeliveryNoteSignLink(w http.ResponseWriter, r *http.Request, session Session, noteID int64) {
	var req deliverySignLinkRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sign link")
		return
	}
	var link DeliverySignLink
	err := a.store.Mutate(func(data *AppData) error {
		note, ok := findDeliveryNote(*data, noteID)
		if !ok {
			return fmt.Errorf("送货单不存在")
		}
		if note.Status == "void" || note.Status == "cancelled" {
			return fmt.Errorf("送货单已作废，不能生成签收链接")
		}
		if hasPendingWorkflowForResource(*data, "delivery_note", note.ID) {
			return fmt.Errorf("送货单正在工作流审批中")
		}
		req.DispatchID = note.DispatchID
		req.TicketID = nonZeroInt(req.TicketID, note.TicketID)
		next, err := buildDeliverySignLink(data, req, session.User.Username)
		if err != nil {
			return err
		}
		link = next
		data.DeliverySignLinks = append(data.DeliverySignLinks, link)
		upsertDeliveryNoteQRCode(data, link)
		addAudit(data, session.User.Username, "send", "delivery_note_sign_link", note.ID, link.LinkNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, link, "delivery.note.sign_link.sent")
}

func upsertDeliveryNote(data *AppData, req deliveryNoteRequest) (DeliveryNote, error) {
	dispatch, ok := findDispatch(*data, req.DispatchID)
	if !ok {
		return DeliveryNote{}, fmt.Errorf("派车单不存在")
	}
	order, ok := findOrder(*data, dispatch.OrderID)
	if !ok {
		return DeliveryNote{}, fmt.Errorf("订单不存在")
	}
	ticketID, err := resolveDeliveryNoteTicket(*data, dispatch.ID, req.TicketID)
	if err != nil {
		return DeliveryNote{}, err
	}
	status := normalizeDeliveryNoteStatus(req.Status)
	if status == "" {
		status = "issued"
	}
	if index, note, ok := findDeliveryNoteByDispatch(*data, dispatch.ID); ok {
		if ticketID != 0 {
			data.DeliveryNotes[index].TicketID = ticketID
		}
		data.DeliveryNotes[index].OrderID = order.ID
		if data.DeliveryNotes[index].QRCode == "" {
			data.DeliveryNotes[index].QRCode = deliveryNoteQRCode(*data, dispatch.ID, note.NoteNo)
		}
		if req.Status != "" {
			data.DeliveryNotes[index].Status = status
		}
		return data.DeliveryNotes[index], nil
	}
	id := nextID(data, "deliveryNote")
	noteNo := number("DN", id)
	note := DeliveryNote{
		ID: id, NoteNo: noteNo, TicketID: ticketID, OrderID: order.ID, DispatchID: dispatch.ID,
		QRCode: deliveryNoteQRCode(*data, dispatch.ID, noteNo), Status: status, CreatedAt: nowString(),
	}
	data.DeliveryNotes = append(data.DeliveryNotes, note)
	return note, nil
}

func deliveryNoteStatusNeedsWorkflow(status string) bool {
	return status == "void" || status == "cancelled"
}

func publishDeliveryNoteStatusWorkflow(data *AppData, item DeliveryNote, targetStatus string, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "delivery_note.status_change_requested",
		Source:     "delivery",
		Resource:   "delivery_note",
		ResourceID: item.ID,
		ResourceNo: item.NoteNo,
		Title:      "送货单状态变更 " + item.NoteNo,
		Actor:      actor,
		Reason:     "送货单状态变更审批",
		Variables: map[string]string{
			"targetStatus":  targetStatus,
			"currentStatus": item.Status,
			"ticketId":      fmt.Sprintf("%d", item.TicketID),
			"orderId":       fmt.Sprintf("%d", item.OrderID),
			"dispatchId":    fmt.Sprintf("%d", item.DispatchID),
		},
	})
}

func applyDeliveryNoteStatusLocked(data *AppData, noteID int64, status string) (DeliveryNote, error) {
	index := deliveryNoteIndex(*data, noteID)
	if index < 0 {
		return DeliveryNote{}, fmt.Errorf("送货单不存在")
	}
	if status == "" {
		return DeliveryNote{}, fmt.Errorf("送货单状态无效")
	}
	if data.DeliveryNotes[index].Status == "signed" && status == "void" {
		return DeliveryNote{}, fmt.Errorf("已签收送货单不能作废")
	}
	data.DeliveryNotes[index].Status = status
	return data.DeliveryNotes[index], nil
}

func resolveDeliveryNoteTicket(data AppData, dispatchID, ticketID int64) (int64, error) {
	if ticketID != 0 {
		ticket, ok := findScaleTicket(data, ticketID)
		if !ok {
			return 0, fmt.Errorf("过磅记录不存在")
		}
		if ticket.DispatchID != 0 && ticket.DispatchID != dispatchID {
			return 0, fmt.Errorf("过磅记录不属于该派车单")
		}
		return ticket.ID, nil
	}
	for _, ticket := range data.ScaleTickets {
		if ticket.DispatchID == dispatchID && ticket.Status == "valid" {
			return ticket.ID, nil
		}
	}
	return 0, nil
}

func deliveryNoteQRCode(data AppData, dispatchID int64, noteNo string) string {
	if link, ok := latestDeliverySignLinkForDispatch(data, dispatchID); ok && link.QRCode != "" {
		return link.QRCode
	}
	return "qr://" + noteNo
}

func latestDeliverySignLinkForDispatch(data AppData, dispatchID int64) (DeliverySignLink, bool) {
	var latest DeliverySignLink
	for _, link := range data.DeliverySignLinks {
		if link.DispatchID == dispatchID && link.ID > latest.ID {
			latest = link
		}
	}
	return latest, latest.ID != 0
}

func findDeliveryNote(data AppData, id int64) (DeliveryNote, bool) {
	for _, note := range data.DeliveryNotes {
		if note.ID == id {
			return note, true
		}
	}
	return DeliveryNote{}, false
}

func findDeliveryNoteByDispatch(data AppData, dispatchID int64) (int, DeliveryNote, bool) {
	for i, note := range data.DeliveryNotes {
		if note.DispatchID == dispatchID {
			return i, note, true
		}
	}
	return -1, DeliveryNote{}, false
}

func deliveryNoteIndex(data AppData, id int64) int {
	for i, note := range data.DeliveryNotes {
		if note.ID == id {
			return i
		}
	}
	return -1
}

func parseDeliveryNoteID(parts []string) (int64, bool) {
	if len(parts) < 2 || parts[0] != "notes" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	return id, err == nil && id > 0
}

func normalizeDeliveryNoteStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "", "issued", "pending", "signed", "void", "cancelled":
		return strings.TrimSpace(value)
	case "reopen", "reopened":
		return "issued"
	default:
		return ""
	}
}
