package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestCustomerStatementWorkflowConfirmation(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"customer_statement_confirm","name":"客户对账确认","category":"approval","resource":"statement","trigger":{"eventType":"statement.confirm_requested","resource":"statement","conditions":[{"field":"totalAmount","operator":"greater_than","value":"0"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"财务确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer statement workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/statements/1/confirm", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("confirm statement status %d: %s", rec.Code, rec.Body.String())
	}
	var statement Statement
	if err := json.Unmarshal(rec.Body.Bytes(), &statement); err != nil {
		t.Fatalf("decode statement: %v", err)
	}
	if statement.Status != "pending_approval" {
		t.Fatalf("expected statement pending workflow approval, got %+v", statement)
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowEvents) != 1 || snapshot.WorkflowEvents[0].EventType != "statement.confirm_requested" || snapshot.WorkflowEvents[0].Resource != "statement" || snapshot.WorkflowEvents[0].Status != "handled" {
		t.Fatalf("expected handled statement workflow event, got %+v", snapshot.WorkflowEvents)
	}
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].Resource != "statement" || snapshot.WorkflowTasks[0].ResourceID != statement.ID || snapshot.WorkflowTasks[0].Status != "pending" {
		t.Fatalf("expected pending statement workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/statements/1/confirm", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("direct confirm should not bypass pending workflow, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[0].ID, 10)+"/act", `{"action":"approve","comment":"财务确认"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("act statement workflow status %d: %s", rec.Code, rec.Body.String())
	}
	for _, item := range app.mustSnapshot().Statements {
		if item.ID != statement.ID {
			continue
		}
		if item.Status != "confirmed" || item.ConfirmedBy == "" || item.ConfirmedAt == "" {
			t.Fatalf("expected confirmed statement after workflow approval, got %+v", item)
		}
		return
	}
	t.Fatalf("statement not found after workflow approval")
}
