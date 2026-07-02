package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestDeliveryNoteManagementFlow(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/delivery/notes", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delivery notes status %d: %s", rec.Code, rec.Body.String())
	}
	var notes []DeliveryNote
	if err := json.Unmarshal(rec.Body.Bytes(), &notes); err != nil {
		t.Fatalf("decode delivery notes: %v", err)
	}
	if len(notes) == 0 || notes[0].NoteNo == "" {
		t.Fatalf("expected seeded delivery notes, got %+v", notes)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes", `{"dispatchId":2}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("create delivery note status %d: %s", rec.Code, rec.Body.String())
	}
	var created DeliveryNote
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created delivery note: %v", err)
	}
	if created.ID == 0 || created.DispatchID != 2 || created.Status != "issued" {
		t.Fatalf("expected issued note for dispatch 2, got %+v", created)
	}
	if created.TicketID != 0 {
		t.Fatalf("dispatch 2 has no ticket in seed, expected ticket 0, got %+v", created)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes/"+strconvID(created.ID)+"/sign-link", `{"channel":"qr","phone":"13800010002"}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("create delivery note sign link status %d: %s", rec.Code, rec.Body.String())
	}
	var link DeliverySignLink
	if err := json.Unmarshal(rec.Body.Bytes(), &link); err != nil {
		t.Fatalf("decode delivery sign link: %v", err)
	}
	if link.DispatchID != created.DispatchID || link.QRCode == "" {
		t.Fatalf("expected sign link for note dispatch, got %+v", link)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/delivery/notes/"+strconvID(created.ID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get delivery note status %d: %s", rec.Code, rec.Body.String())
	}
	var updated DeliveryNote
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated delivery note: %v", err)
	}
	if updated.Status != "pending" || updated.QRCode != link.QRCode {
		t.Fatalf("expected pending note with link QR, got %+v / %+v", updated, link)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes/"+strconvID(created.ID)+"/reprint", `{}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("reprint delivery note status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes/"+strconvID(created.ID)+"/status", `{"status":"void"}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("void delivery note status %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"void"`) {
		t.Fatalf("expected void status response: %s", rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes/"+strconvID(created.ID)+"/status", `{"status":"reopen"}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("reopen delivery note status %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"issued"`) {
		t.Fatalf("expected reopened issued status response: %s", rec.Body.String())
	}
}

func TestDeliveryNoteVoidWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"delivery_note_void_review","name":"送货单作废审批","category":"approval","resource":"delivery_note","trigger":{"eventType":"delivery_note.status_change_requested","resource":"delivery_note","conditions":[{"field":"targetStatus","operator":"equals","value":"void"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"作废复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create delivery note void workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes", `{"dispatchId":2}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("create workflow delivery note status %d: %s", rec.Code, rec.Body.String())
	}
	var note DeliveryNote
	if err := json.Unmarshal(rec.Body.Bytes(), &note); err != nil {
		t.Fatalf("decode workflow delivery note: %v", err)
	}
	if note.Status != "issued" {
		t.Fatalf("expected issued note before void request, got %+v", note)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes/"+strconvID(note.ID)+"/status", `{"status":"void"}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("request delivery note void workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending DeliveryNote
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending delivery note: %v", err)
	}
	if pending.Status != "issued" {
		t.Fatalf("delivery note should stay issued before workflow approval, got %+v", pending)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/notes/"+strconvID(note.ID)+"/sign-link", `{"channel":"qr","phone":"13800010002"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected pending void workflow to block sign link, got %d: %s", rec.Code, rec.Body.String())
	}

	snapshot := app.mustSnapshot()
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "delivery_note" && task.ResourceID == note.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending delivery note workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconvID(taskID)+"/act", `{"action":"approve","comment":"作废复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve delivery note void workflow status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/delivery/notes/"+strconvID(note.ID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get approved void delivery note status %d: %s", rec.Code, rec.Body.String())
	}
	var voided DeliveryNote
	if err := json.Unmarshal(rec.Body.Bytes(), &voided); err != nil {
		t.Fatalf("decode voided delivery note: %v", err)
	}
	if voided.Status != "void" {
		t.Fatalf("expected workflow-approved void delivery note, got %+v", voided)
	}
}

func TestDeliverySignPublishesWorkflowEvent(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"delivery_sign_archive","name":"签收归档流程","category":"approval","resource":"delivery_sign","trigger":{"eventType":"delivery_sign.completed","resource":"delivery_sign","conditions":[{"field":"signedQty","operator":"greater_than","value":"0"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"签收归档"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create delivery sign workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/sign", `{"dispatchId":2,"signer":"赵工","signedQty":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("sign delivery status %d: %s", rec.Code, rec.Body.String())
	}
	var sign DeliverySign
	if err := json.Unmarshal(rec.Body.Bytes(), &sign); err != nil {
		t.Fatalf("decode delivery sign: %v", err)
	}
	if sign.ID == 0 || sign.SignNo == "" || sign.SignedQty != 1 {
		t.Fatalf("unexpected delivery sign: %+v", sign)
	}
	if sign.ReviewStatus != "pending_review" {
		t.Fatalf("expected delivery sign pending review, got %+v", sign)
	}

	snapshot := app.mustSnapshot()
	eventFound := false
	var taskID int64
	for _, event := range snapshot.WorkflowEvents {
		if event.EventType == "delivery_sign.completed" && event.Resource == "delivery_sign" && event.ResourceID == sign.ID && event.Status == "handled" {
			eventFound = true
			break
		}
	}
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "delivery_sign" && task.ResourceID == sign.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if !eventFound || taskID == 0 {
		t.Fatalf("expected delivery sign workflow event and task, events=%+v tasks=%+v", snapshot.WorkflowEvents, snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconvID(taskID)+"/act", `{"action":"approve","comment":"签收归档通过"}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("approve delivery sign workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	updated, ok := findDeliverySign(snapshot, sign.ID)
	if !ok {
		t.Fatalf("expected delivery sign in snapshot")
	}
	if updated.ReviewStatus != "approved" || updated.ReviewedBy != "admin" || updated.ReviewedAt == "" {
		t.Fatalf("expected approved delivery sign review result, got %+v", updated)
	}
	resultApplied := false
	for _, log := range snapshot.WorkflowLogs {
		if log.Action == workflowResultAppliedAction && log.Resource == "delivery_sign" && log.ResourceID == sign.ID && log.Status == "approved" {
			resultApplied = true
			break
		}
	}
	if !resultApplied {
		t.Fatalf("expected delivery sign result_applied log, logs=%+v", snapshot.WorkflowLogs)
	}
}

func TestDeliveryNotePermissionsFollowRole(t *testing.T) {
	app := newTestHTTPApp(t)
	dispatcherToken := testLogin(t, app, "dispatcher", "dispatch123")

	rec := testRequest(t, app, dispatcherToken, http.MethodPost, "/api/delivery/notes", `{"dispatchId":2}`)
	if !deliveryNoteTestStatusOK(rec.Code) {
		t.Fatalf("dispatcher should manage delivery notes, status %d: %s", rec.Code, rec.Body.String())
	}

	customerToken := testLogin(t, app, "customer", "customer123")
	rec = testRequest(t, app, customerToken, http.MethodGet, "/api/delivery/notes", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("customer should read delivery notes, status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, customerToken, http.MethodPost, "/api/delivery/notes", `{"dispatchId":1}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("customer should not manage delivery notes, status %d: %s", rec.Code, rec.Body.String())
	}
}

func strconvID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func deliveryNoteTestStatusOK(code int) bool {
	return code >= 200 && code < 300
}
