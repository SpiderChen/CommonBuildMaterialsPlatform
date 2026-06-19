package appliance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestTicketReprintAndVoidApprovalFlow(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/1/reprint", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reprint status %d: %s", rec.Code, rec.Body.String())
	}
	var printLog TicketPrintLog
	if err := json.Unmarshal(rec.Body.Bytes(), &printLog); err != nil {
		t.Fatalf("decode print log: %v", err)
	}
	if printLog.TicketID != 1 || printLog.PrintedBy != "admin" {
		t.Fatalf("unexpected print log: %+v", printLog)
	}

	tickets := fetchTickets(t, app, token)
	if tickets[0].PrintCount != 2 {
		t.Fatalf("expected print count 2, got %d", tickets[0].PrintCount)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/1/void/request", `{"reason":"地磅重量争议复核"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("void request status %d: %s", rec.Code, rec.Body.String())
	}
	var voidLog TicketVoidLog
	if err := json.Unmarshal(rec.Body.Bytes(), &voidLog); err != nil {
		t.Fatalf("decode void log: %v", err)
	}
	if voidLog.Status != "pending" || voidLog.Reason == "" {
		t.Fatalf("unexpected pending void log: %+v", voidLog)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/weighbridge/ticket-voids", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("ticket voids status %d: %s", rec.Code, rec.Body.String())
	}
	var voidLogs []TicketVoidLog
	if err := json.Unmarshal(rec.Body.Bytes(), &voidLogs); err != nil {
		t.Fatalf("decode void logs: %v", err)
	}
	if len(voidLogs) != 1 || voidLogs[0].Status != "pending" {
		t.Fatalf("expected pending void log list, got %+v", voidLogs)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/1/void/approve", `{"approved":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("void approve status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &voidLog); err != nil {
		t.Fatalf("decode approved void log: %v", err)
	}
	if voidLog.Status != "approved" || voidLog.ApprovedBy != "admin" {
		t.Fatalf("unexpected approved void log: %+v", voidLog)
	}

	tickets = fetchTickets(t, app, token)
	if tickets[0].Status != "void" || tickets[0].SettlementStatus != "void" || tickets[0].SignStatus != "void" {
		t.Fatalf("expected void ticket status, got %+v", tickets[0])
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/1/reprint", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected void ticket reprint rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRawMaterialReceiptCreatesInboundScaleTicketAndSnapshots(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/procurement/receipts", `{"purchaseOrderId":1,"siteId":1,"plateNo":"粤B原料88","grossWeight":56.8,"tareWeight":18.2,"qualityStatus":"passed","status":"stocked"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("raw receipt status %d: %s", rec.Code, rec.Body.String())
	}
	var receipt RawMaterialReceipt
	if err := json.Unmarshal(rec.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	if receipt.TicketID == 0 || receipt.PlateNo != "粤B原料88" {
		t.Fatalf("expected receipt to link inbound ticket, got %+v", receipt)
	}

	tickets := fetchTickets(t, app, token)
	ticket := findTicket(tickets, receipt.TicketID)
	if ticket.TicketType != "raw_material_in" || ticket.ReceiptID != receipt.ID || ticket.MaterialID != receipt.MaterialID || ticket.SupplierID != receipt.SupplierID {
		t.Fatalf("unexpected raw inbound ticket: %+v", ticket)
	}
	if ticket.NetWeight != receipt.NetWeight || ticket.SnapshotURL == "" || ticket.SignStatus != "not_required" {
		t.Fatalf("expected ticket weights and snapshot, got %+v", ticket)
	}

	records := fetchWeightRecords(t, app, token)
	if countWeightRecords(records, ticket.ID) != 2 {
		t.Fatalf("expected tare and gross snapshot records, got %+v", records)
	}

	dispatcherToken := testLogin(t, app, "dispatcher", "dispatch123")
	dispatcherTickets := fetchTickets(t, app, dispatcherToken)
	if findTicket(dispatcherTickets, ticket.ID).ID == 0 {
		t.Fatalf("expected site dispatcher to see raw inbound ticket, got %+v", dispatcherTickets)
	}
}

func TestTransferReturnAndWasteScaleTickets(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/procurement/transfers", `{"fromSiteId":1,"toSiteId":2,"materialId":3,"quantity":10,"remark":"调拨过磅测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create transfer status %d: %s", rec.Code, rec.Body.String())
	}
	var transfer InventoryTransfer
	if err := json.Unmarshal(rec.Body.Bytes(), &transfer); err != nil {
		t.Fatalf("decode transfer: %v", err)
	}
	task := findApprovalTaskForResource(fetchApprovalTasks(t, app, token), "inventory_transfer", transfer.ID)
	if task.ID == 0 {
		t.Fatalf("transfer approval task not found")
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+itoa(task.ID)+"/act", `{"action":"approve","comment":"调度审批"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first transfer approval status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+itoa(task.ID)+"/act", `{"action":"approve","comment":"高管审批"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("final transfer approval status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/transfers/"+itoa(transfer.ID)+"/complete", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("complete transfer status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/transfer", `{"transferId":`+itoa(transfer.ID)+`,"grossWeight":28,"tareWeight":18}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("transfer ticket status %d: %s", rec.Code, rec.Body.String())
	}
	var transferTicket ScaleTicket
	if err := json.Unmarshal(rec.Body.Bytes(), &transferTicket); err != nil {
		t.Fatalf("decode transfer ticket: %v", err)
	}
	if transferTicket.TicketType != "inventory_transfer" || transferTicket.TransferID != transfer.ID || transferTicket.NetWeight != 10 {
		t.Fatalf("unexpected transfer ticket: %+v", transferTicket)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/return", `{"relatedTicketId":1,"grossWeight":24.2,"tareWeight":18.2}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("return ticket status %d: %s", rec.Code, rec.Body.String())
	}
	var returnTicket ScaleTicket
	if err := json.Unmarshal(rec.Body.Bytes(), &returnTicket); err != nil {
		t.Fatalf("decode return ticket: %v", err)
	}
	if returnTicket.TicketType != "product_return" || returnTicket.RelatedTicketID != 1 || returnTicket.DispatchID != 1 || returnTicket.NetWeight != 6 {
		t.Fatalf("unexpected return ticket: %+v", returnTicket)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets/waste", `{"siteId":1,"materialId":3,"grossWeight":21.5,"tareWeight":20.5,"remark":"含泥废砂外运"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("waste ticket status %d: %s", rec.Code, rec.Body.String())
	}
	var wasteTicket ScaleTicket
	if err := json.Unmarshal(rec.Body.Bytes(), &wasteTicket); err != nil {
		t.Fatalf("decode waste ticket: %v", err)
	}
	if wasteTicket.TicketType != "waste_out" || wasteTicket.MaterialID != 3 || wasteTicket.NetWeight != 1 {
		t.Fatalf("unexpected waste ticket: %+v", wasteTicket)
	}

	records := fetchWeightRecords(t, app, token)
	for _, ticket := range []ScaleTicket{transferTicket, returnTicket, wasteTicket} {
		if countWeightRecords(records, ticket.ID) != 2 {
			t.Fatalf("expected gross/tare records for %+v, got %+v", ticket, records)
		}
	}
	procurement := fetchProcurementOverview(t, app, token)
	if !hasInventoryQty(procurement.Inventory, 1, 3, 829) {
		t.Fatalf("expected waste ticket to deduct site inventory to 829, got %+v", procurement.Inventory)
	}
	if countInventoryFlows(procurement.Flows, "waste_ticket", wasteTicket.ID) != 1 {
		t.Fatalf("expected waste inventory flow, got %+v", procurement.Flows)
	}
}

func TestScaleDeviceEventReportsCaptureAndCheatResult(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testDeviceRequest(t, app, "scale-demo-key-1", http.MethodPost, "/api/weighbridge/device-events", `{"deviceCode":"NS-SCALE-01","ticketId":1,"plateNo":"粤B12345","recognizedPlateNo":"粤B12345","weight":31.6,"weightType":"gross","stable":true,"snapshotUrl":"capture://ns-scale-01/device-gross.jpg"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("device event status %d: %s", rec.Code, rec.Body.String())
	}
	var event ScaleDeviceEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &event); err != nil {
		t.Fatalf("decode device event: %v", err)
	}
	if event.Status != "accepted" || event.CheatFlag || event.DeviceCode != "NS-SCALE-01" {
		t.Fatalf("unexpected accepted device event: %+v", event)
	}

	tickets := fetchTickets(t, app, token)
	ticket := findTicket(tickets, 1)
	if ticket.GrossWeight != 31.6 || ticket.SnapshotURL != event.SnapshotURL {
		t.Fatalf("expected device event to update ticket gross weight, got %+v", ticket)
	}
	records := fetchWeightRecords(t, app, token)
	if countWeightRecords(records, 1) < 3 {
		t.Fatalf("expected device capture record to be appended, got %+v", records)
	}

	rec = testDeviceRequest(t, app, "scale-demo-key-1", http.MethodPost, "/api/weighbridge/device-events", `{"deviceCode":"NS-SCALE-01","ticketId":1,"plateNo":"粤B00000","recognizedPlateNo":"粤B00000","weight":18.7,"weightType":"tare","stable":true,"snapshotUrl":"capture://ns-scale-01/bad-tare.jpg"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("mismatch device event status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &event); err != nil {
		t.Fatalf("decode mismatch device event: %v", err)
	}
	if event.Status != "blocked" || !event.CheatFlag || event.CheatReason != "plate_mismatch" {
		t.Fatalf("expected plate mismatch cheat result, got %+v", event)
	}
	tickets = fetchTickets(t, app, token)
	if findTicket(tickets, 1).Status != "abnormal" {
		t.Fatalf("expected mismatched ticket to be abnormal, got %+v", findTicket(tickets, 1))
	}

	events := fetchScaleDeviceEvents(t, app, token)
	if len(events) != 2 {
		t.Fatalf("expected two device events, got %+v", events)
	}

	rec = testDeviceRequest(t, app, "device-demo-key-1", http.MethodPost, "/api/weighbridge/device-events", `{"deviceCode":"NS-SCALE-01","ticketId":1,"weight":31.6,"weightType":"gross","stable":true,"snapshotUrl":"capture://bad.jpg"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected device scope rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func fetchTickets(t *testing.T, app *App, token string) []ScaleTicket {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/weighbridge/tickets", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("tickets status %d: %s", rec.Code, rec.Body.String())
	}
	var tickets []ScaleTicket
	if err := json.Unmarshal(rec.Body.Bytes(), &tickets); err != nil {
		t.Fatalf("decode tickets: %v", err)
	}
	if len(tickets) == 0 {
		t.Fatalf("expected tickets")
	}
	return tickets
}

func fetchWeightRecords(t *testing.T, app *App, token string) []ScaleWeightRecord {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/weighbridge/weight-records", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("weight records status %d: %s", rec.Code, rec.Body.String())
	}
	var records []ScaleWeightRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &records); err != nil {
		t.Fatalf("decode weight records: %v", err)
	}
	return records
}

func fetchScaleDeviceEvents(t *testing.T, app *App, token string) []ScaleDeviceEvent {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/weighbridge/device-events", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("scale device events status %d: %s", rec.Code, rec.Body.String())
	}
	var events []ScaleDeviceEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatalf("decode scale device events: %v", err)
	}
	return events
}

func testDeviceRequest(t *testing.T, app *App, key, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("X-Device-Key", key)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	return rec
}

func findTicket(items []ScaleTicket, id int64) ScaleTicket {
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	return ScaleTicket{}
}

func countWeightRecords(items []ScaleWeightRecord, ticketID int64) int {
	count := 0
	for _, item := range items {
		if item.TicketID == ticketID && item.SnapshotURL != "" {
			count++
		}
	}
	return count
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
