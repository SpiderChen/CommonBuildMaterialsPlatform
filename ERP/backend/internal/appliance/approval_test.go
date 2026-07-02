package appliance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newWorkflowWebhookEndpoint(t *testing.T, status int) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(status)
		if status >= 200 && status < 300 {
			_, _ = w.Write([]byte("workflow accepted"))
			return
		}
		_, _ = w.Write([]byte("workflow rejected"))
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func workflowEndpointBody(body string, endpoint string) string {
	return strings.NewReplacer(
		"mock://success", endpoint,
		"mock://ok", endpoint,
		"mock://fail", endpoint,
	).Replace(body)
}

func TestCreditRiskOrderCreatesApprovalTaskAndBlocksDirectApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":5000,"planTime":"2026-06-18 16:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create risky order status %d: %s", rec.Code, rec.Body.String())
	}
	var order SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode order: %v", err)
	}
	if order.Status != "pending_approval" || order.RiskFlag != "credit_limit" {
		t.Fatalf("expected pending credit approval order, got %+v", order)
	}

	orderID := strconv.FormatInt(order.ID, 10)
	rec = testRequest(t, app, token, http.MethodPost, "/api/orders/"+orderID+"/approve", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected direct approval rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	tasks := fetchApprovalTasks(t, app, token)
	if len(tasks) != 1 {
		t.Fatalf("expected one approval task, got %+v", tasks)
	}
	if tasks[0].Resource != "sales_order" || tasks[0].ResourceID != order.ID || tasks[0].CurrentRole != "dispatcher" {
		t.Fatalf("unexpected approval task: %+v", tasks[0])
	}
	if tasks[0].WorkflowInstanceID == 0 || tasks[0].WorkflowTaskID == 0 {
		t.Fatalf("expected approval task backed by workflow engine, got %+v", tasks[0])
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || len(snapshot.WorkflowTasks) != 1 {
		t.Fatalf("expected one workflow instance and one workflow task, got %+v / %+v", snapshot.WorkflowInstances, snapshot.WorkflowTasks)
	}
	if len(snapshot.WorkflowEvents) != 1 || snapshot.WorkflowEvents[0].EventType != "sales_order.risk_detected" || snapshot.WorkflowEvents[0].Status != "handled" {
		t.Fatalf("expected handled workflow event, got %+v", snapshot.WorkflowEvents)
	}
	if len(snapshot.WorkflowEvents[0].MatchedDefinitions) != 1 || snapshot.WorkflowEvents[0].MatchedDefinitions[0] != "order_credit_risk" {
		t.Fatalf("expected workflow event to match order_credit_risk, got %+v", snapshot.WorkflowEvents[0])
	}
	if snapshot.WorkflowInstances[0].DefinitionCode != "order_credit_risk" || snapshot.WorkflowInstances[0].CurrentRole != "dispatcher" {
		t.Fatalf("unexpected workflow instance: %+v", snapshot.WorkflowInstances[0])
	}
	if len(snapshot.WorkflowLogs) != 2 || snapshot.WorkflowLogs[0].Action != "instance_started" || snapshot.WorkflowLogs[1].Action != "task_created" {
		t.Fatalf("expected workflow start logs, got %+v", snapshot.WorkflowLogs)
	}

	taskID := strconv.FormatInt(tasks[0].ID, 10)
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+taskID+"/act", `{"action":"approve","comment":"销售确认"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first approval status %d: %s", rec.Code, rec.Body.String())
	}
	var task ApprovalTask
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode first approval: %v", err)
	}
	if task.Status != "pending" || task.CurrentRole != "boss" || len(task.Actions) != 1 {
		t.Fatalf("expected second approval step, got %+v", task)
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.WorkflowTasks) != 2 || snapshot.WorkflowInstances[0].CurrentRole != "boss" {
		t.Fatalf("expected workflow engine to create second task, got %+v / %+v", snapshot.WorkflowInstances, snapshot.WorkflowTasks)
	}
	if len(snapshot.WorkflowLogs) != 4 || snapshot.WorkflowLogs[2].Action != "task_approved" || snapshot.WorkflowLogs[3].Action != "task_created" {
		t.Fatalf("expected workflow step logs, got %+v", snapshot.WorkflowLogs)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+taskID+"/act", `{"action":"approve","comment":"高管确认"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("final approval status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode final approval: %v", err)
	}
	if task.Status != "approved" || task.CurrentRole != "" || len(task.Actions) != 2 {
		t.Fatalf("expected approved task, got %+v", task)
	}
	snapshot = app.mustSnapshot()
	if snapshot.WorkflowInstances[0].Status != "approved" || snapshot.WorkflowInstances[0].CurrentTaskID != 0 || len(snapshot.WorkflowInstances[0].Actions) != 2 {
		t.Fatalf("expected workflow instance approved, got %+v", snapshot.WorkflowInstances[0])
	}
	if len(snapshot.WorkflowLogs) != 7 || snapshot.WorkflowLogs[5].Action != "instance_approved" || snapshot.WorkflowLogs[6].Action != workflowResultAppliedAction {
		t.Fatalf("expected workflow completion logs, got %+v", snapshot.WorkflowLogs)
	}
	if len(snapshot.WorkflowOutbox) != 7 || snapshot.WorkflowOutbox[6].EventType != "workflow.result_applied" || snapshot.WorkflowOutbox[6].Payload["var.resource"] != "sales_order" {
		t.Fatalf("expected workflow result outbox event, got %+v", snapshot.WorkflowOutbox)
	}

	orders := fetchOrders(t, app, token)
	var updated SalesOrder
	for _, item := range orders {
		if item.ID == order.ID {
			updated = item
		}
	}
	if updated.Status != "submitted" {
		t.Fatalf("expected order back to submitted after approval, got %+v", updated)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/orders/"+orderID+"/approve", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("business approval status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestWorkflowDefinitionsMirrorApprovalFlowConfig(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/system/workflows", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("workflows status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Definitions []WorkflowDefinition `json:"definitions"`
		Instances   []WorkflowInstance   `json:"instances"`
		Tasks       []WorkflowTask       `json:"tasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode workflows: %v", err)
	}
	if len(overview.Definitions) < 5 || !hasWorkflowDefinition(overview.Definitions, "order_credit_risk", "active") {
		t.Fatalf("expected seeded workflow definitions, got %+v", overview.Definitions)
	}
	for _, definition := range overview.Definitions {
		if definition.Code == "order_credit_risk" && definition.Trigger.EventType != "sales_order.risk_detected" {
			t.Fatalf("expected order_credit_risk event trigger, got %+v", definition.Trigger)
		}
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/approval-flows", `{"code":"quality_exception","name":"质量异常审批","resource":"quality_inspection","steps":[{"roleCode":"quality","action":"approve"},{"roleCode":"boss","action":"approve"}],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create approval flow status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/definitions", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("workflow definitions status %d: %s", rec.Code, rec.Body.String())
	}
	var definitions []WorkflowDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &definitions); err != nil {
		t.Fatalf("decode workflow definitions: %v", err)
	}
	if !hasWorkflowDefinition(definitions, "quality_exception", "active") {
		t.Fatalf("expected workflow definition mirrored from approval flow, got %+v", definitions)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"contract_archive","name":"合同归档审批","category":"approval","resource":"contract","steps":[{"seq":1,"roleCode":"dispatcher","action":"approve","name":"业务确认"},{"seq":2,"roleCode":"boss","action":"approve","name":"归档确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	var savedDefinition WorkflowDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &savedDefinition); err != nil {
		t.Fatalf("decode workflow definition: %v", err)
	}
	if savedDefinition.ID == 0 || savedDefinition.Code != "contract_archive" || len(savedDefinition.Steps) != 2 {
		t.Fatalf("unexpected saved workflow definition: %+v", savedDefinition)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/approval-flows", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("approval flows after workflow save status %d: %s", rec.Code, rec.Body.String())
	}
	var mirroredFlows []ApprovalFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &mirroredFlows); err != nil {
		t.Fatalf("decode mirrored approval flows: %v", err)
	}
	if !hasApprovalFlow(mirroredFlows, "contract_archive") {
		t.Fatalf("expected workflow definition mirrored to approval flow, got %+v", mirroredFlows)
	}
}

func TestWorkflowInboxEndpointFiltersPendingTasksByRole(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	dispatcherToken := testLogin(t, app, "dispatcher", "dispatch123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"inbox_dispatcher","name":"待办调度审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"workflow.inbox.dispatcher","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"dispatcher","action":"approve","name":"调度确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dispatcher inbox workflow status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"inbox_quality","name":"待办质检审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"workflow.inbox.quality","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quality inbox workflow status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/events", `{"eventType":"workflow.inbox.dispatcher","source":"test","eventKey":"workflow-inbox-dispatcher","resource":"quality_exception","resourceId":9701,"resourceNo":"QE-INBOX-DISPATCHER","title":"调度待办"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish dispatcher inbox event status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/events", `{"eventType":"workflow.inbox.quality","source":"test","eventKey":"workflow-inbox-quality","resource":"quality_exception","resourceId":9702,"resourceNo":"QE-INBOX-QUALITY","title":"质检待办"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish quality inbox event status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, dispatcherToken, http.MethodGet, "/api/system/workflows/inbox", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dispatcher workflow inbox status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatcherInbox []WorkflowInboxItem
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatcherInbox); err != nil {
		t.Fatalf("decode dispatcher workflow inbox: %v", err)
	}
	if len(dispatcherInbox) != 1 || dispatcherInbox[0].Task.RoleCode != "dispatcher" || !dispatcherInbox[0].CanAct || dispatcherInbox[0].Instance.ResourceNo != "QE-INBOX-DISPATCHER" {
		t.Fatalf("expected dispatcher inbox to contain only dispatcher task, got %+v", dispatcherInbox)
	}

	rec = testRequest(t, app, dispatcherToken, http.MethodGet, "/api/system/workflows", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dispatcher workflow overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Inbox []WorkflowInboxItem `json:"inbox"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode dispatcher workflow overview: %v", err)
	}
	if len(overview.Inbox) != 1 || overview.Inbox[0].Task.RoleCode != "dispatcher" {
		t.Fatalf("expected workflow overview inbox to be role filtered, got %+v", overview.Inbox)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/workflows/inbox", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("admin workflow inbox status %d: %s", rec.Code, rec.Body.String())
	}
	var adminInbox []WorkflowInboxItem
	if err := json.Unmarshal(rec.Body.Bytes(), &adminInbox); err != nil {
		t.Fatalf("decode admin workflow inbox: %v", err)
	}
	if len(adminInbox) != 2 {
		t.Fatalf("expected admin inbox to contain both pending tasks, got %+v", adminInbox)
	}
}

func TestWorkflowEngineRejectsAndChecksRole(t *testing.T) {
	data := SeedData()
	task, err := submitApprovalTask(&data, "inventory_transfer", "inventory_transfer", 99, "IT000000000099", "库存调拨审批", "admin", "测试拒绝")
	if err != nil {
		t.Fatalf("submit approval task: %v", err)
	}
	if task.WorkflowInstanceID == 0 {
		t.Fatalf("expected workflow-backed task, got %+v", task)
	}

	if _, err := actWorkflowInstance(&data, task.WorkflowInstanceID, workflowActionRequest{Action: "approve", Actor: "customer", ActorRole: "customer"}); err == nil {
		t.Fatalf("expected role check to reject workflow action")
	}
	instance, err := actWorkflowInstance(&data, task.WorkflowInstanceID, workflowActionRequest{Action: "reject", Actor: "dispatcher", ActorRole: "dispatcher", Comment: "不同意"})
	if err != nil {
		t.Fatalf("reject workflow: %v", err)
	}
	if instance.Status != "rejected" || instance.CurrentTaskID != 0 || len(instance.Actions) != 1 {
		t.Fatalf("expected rejected workflow instance, got %+v", instance)
	}
	if len(data.WorkflowTasks) != 1 || data.WorkflowTasks[0].Status != "rejected" {
		t.Fatalf("expected workflow task rejected, got %+v", data.WorkflowTasks)
	}
}

func TestWorkflowEventReplayStartsWorkflowAfterDefinitionIsAdded(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	var original WorkflowEvent
	if err := app.store.Mutate(func(data *AppData) error {
		event, instances, err := publishWorkflowEvent(data, workflowEventRequest{
			EventType:  "quality_exception.submitted",
			Source:     "external-qms",
			EventKey:   "qms:QE-9001",
			Resource:   "quality_exception",
			ResourceID: 9001,
			ResourceNo: "QE-9001",
			Title:      "质量异常审批",
			Actor:      "admin",
			Reason:     "试验结果异常",
			Variables:  map[string]string{"severity": "high"},
		})
		if err != nil {
			return err
		}
		if len(instances) != 0 || event.Status != "ignored" {
			t.Fatalf("expected ignored event before workflow definition, got %+v / %+v", event, instances)
		}
		original = event
		return nil
	}); err != nil {
		t.Fatalf("publish ignored event: %v", err)
	}

	rec := testRequest(t, app, token, http.MethodGet, "/api/system/workflows/events?recovery=true&source=external-qms&eventKey=qms:QE-9001", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter recovery workflow events status %d: %s", rec.Code, rec.Body.String())
	}
	var filteredEvents []WorkflowEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredEvents); err != nil {
		t.Fatalf("decode recovery workflow events: %v", err)
	}
	if len(filteredEvents) != 1 || filteredEvents[0].ID != original.ID || filteredEvents[0].Status != "ignored" {
		t.Fatalf("expected recovery filter to find original ignored event, got %+v", filteredEvents)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_event","name":"质量异常事件审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.submitted","resource":"quality_exception","conditions":[{"field":"severity","operator":"equals","value":"high"}]},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"},{"seq":2,"roleCode":"boss","action":"approve","name":"管理确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events/"+strconv.FormatInt(original.ID, 10)+"/replay", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("replay workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	var replayed WorkflowEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &replayed); err != nil {
		t.Fatalf("decode replayed event: %v", err)
	}
	if replayed.Status != "handled" || replayed.Source != "external-qms" || replayed.EventKey != "qms:QE-9001" || replayed.ReplayOfID != original.ID || len(replayed.MatchedDefinitions) != 1 || replayed.MatchedDefinitions[0] != "quality_exception_event" {
		t.Fatalf("expected handled replayed event, got %+v", replayed)
	}

	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].DefinitionCode != "quality_exception_event" {
		t.Fatalf("expected replay to start quality workflow, got %+v", snapshot.WorkflowInstances)
	}
	if snapshot.WorkflowInstances[0].TriggerEventID != replayed.ID {
		t.Fatalf("expected workflow instance linked to replayed event, got %+v", snapshot.WorkflowInstances[0])
	}
	if len(snapshot.ApprovalTasks) != 1 || snapshot.ApprovalTasks[0].WorkflowInstanceID != snapshot.WorkflowInstances[0].ID {
		t.Fatalf("expected replay to create approval task, got %+v", snapshot.ApprovalTasks)
	}
	if len(snapshot.WorkflowLogs) != 2 || snapshot.WorkflowLogs[0].TriggerEventID != replayed.ID || snapshot.WorkflowLogs[1].Action != "task_created" {
		t.Fatalf("expected workflow logs for replayed event, got %+v", snapshot.WorkflowLogs)
	}
	var originalAfterReplay WorkflowEvent
	for _, event := range snapshot.WorkflowEvents {
		if event.ID == original.ID {
			originalAfterReplay = event
			break
		}
	}
	if originalAfterReplay.Status != "resolved" || originalAfterReplay.RecoveredByEventID != replayed.ID || originalAfterReplay.ResolvedBy != "admin" || originalAfterReplay.ResolvedAt == "" {
		t.Fatalf("expected original event resolved by replay, got %+v", originalAfterReplay)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/events?replayOfId="+strconv.FormatInt(original.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter replayed workflow events status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredEvents); err != nil {
		t.Fatalf("decode replayed workflow events: %v", err)
	}
	if len(filteredEvents) != 1 || filteredEvents[0].ID != replayed.ID || filteredEvents[0].EventKey != "qms:QE-9001" {
		t.Fatalf("expected replayOfId filter to find replayed event, got %+v", filteredEvents)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/events?status=resolved&recoveredByEventId="+strconv.FormatInt(replayed.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter resolved workflow events status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredEvents); err != nil {
		t.Fatalf("decode resolved workflow events: %v", err)
	}
	if len(filteredEvents) != 1 || filteredEvents[0].ID != original.ID {
		t.Fatalf("expected recoveredByEventId filter to find original event, got %+v", filteredEvents)
	}
}

func TestWorkflowEventManualResolve(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	var original WorkflowEvent
	if err := app.store.Mutate(func(data *AppData) error {
		event, instances, err := publishWorkflowEvent(data, workflowEventRequest{
			EventType:  "quality_exception.submitted",
			Resource:   "quality_exception",
			ResourceID: 9002,
			ResourceNo: "QE-9002",
			Title:      "质量异常审批",
			Actor:      "admin",
			Variables:  map[string]string{"severity": "low"},
		})
		if err != nil {
			return err
		}
		if len(instances) != 0 || event.Status != "ignored" {
			t.Fatalf("expected ignored event, got %+v / %+v", event, instances)
		}
		original = event
		return nil
	}); err != nil {
		t.Fatalf("publish ignored event: %v", err)
	}

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events/"+strconv.FormatInt(original.ID, 10)+"/resolve", `{"resolution":"manual triage"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("resolve workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	var resolved WorkflowEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &resolved); err != nil {
		t.Fatalf("decode resolved event: %v", err)
	}
	if resolved.Status != "resolved" || resolved.Resolution != "manual triage" || resolved.ResolvedBy != "admin" || resolved.ResolvedAt == "" {
		t.Fatalf("expected resolved workflow event, got %+v", resolved)
	}
}

func TestWorkflowEventPreviewDoesNotCreateWork(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_preview","name":"质量异常预览审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.preview","resource":"quality_exception","conditions":[{"field":"severity","operator":"equals","value":"high"}]},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}

	body := `{"eventType":"quality_exception.preview","source":"integration","eventKey":"QE-PREVIEW-1","actor":"qms-bot","resource":"quality_exception","resourceId":9301,"resourceNo":"QE-PREVIEW-1","title":"预览质量异常","reason":"预览测试","variables":{"severity":"high"}}`
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events/preview", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("preview workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	var preview WorkflowEventPreview
	if err := json.Unmarshal(rec.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview.Event.Status != "preview" || preview.Event.EventNo != "" || preview.Event.Actor != "qms-bot" || preview.WillStart != 1 || len(preview.Matches) != 1 || !preview.Matches[0].WillStart {
		t.Fatalf("expected preview to match and start one workflow, got %+v", preview)
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowEvents) != 0 || len(snapshot.WorkflowInstances) != 0 || len(snapshot.WorkflowTasks) != 0 || len(snapshot.WorkflowLogs) != 0 || len(snapshot.WorkflowOutbox) != 0 {
		t.Fatalf("expected preview not to mutate workflow state, got events=%+v instances=%+v tasks=%+v logs=%+v outbox=%+v", snapshot.WorkflowEvents, snapshot.WorkflowInstances, snapshot.WorkflowTasks, snapshot.WorkflowLogs, snapshot.WorkflowOutbox)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	var published WorkflowEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &published); err != nil {
		t.Fatalf("decode published event: %v", err)
	}
	if published.Actor != "qms-bot" {
		t.Fatalf("expected published workflow event to preserve external actor, got %+v", published)
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].Applicant != "qms-bot" {
		t.Fatalf("expected workflow instance applicant from external actor, got %+v", snapshot.WorkflowInstances)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events/preview", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("duplicate preview workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode duplicate preview: %v", err)
	}
	if !preview.Duplicate || preview.DuplicateEventNo != published.EventNo || preview.WillStart != 0 || len(preview.Matches) != 1 || preview.Matches[0].WillStart {
		t.Fatalf("expected duplicate preview not to start workflow, published=%+v preview=%+v", published, preview)
	}
}

func TestWorkflowEventPublishEndpointDeduplicatesEventKey(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_publish","name":"质量异常发布审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.submitted","resource":"quality_exception","conditions":[{"field":"severity","operator":"equals","value":"high"}]},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}

	body := `{"eventType":"quality_exception.submitted","source":"integration","eventKey":"QE-EXT-1","resource":"quality_exception","resourceId":9101,"resourceNo":"QE-EXT-1","title":"外部质量异常","reason":"外部系统推送","variables":{"severity":"high"}}`
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	var first WorkflowEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode first event: %v", err)
	}
	if first.Status != "handled" || first.Source != "integration" || first.EventKey != "QE-EXT-1" || len(first.MatchedDefinitions) != 1 {
		t.Fatalf("expected handled published event, got %+v", first)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("duplicate workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	var duplicate WorkflowEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &duplicate); err != nil {
		t.Fatalf("decode duplicate event: %v", err)
	}
	if duplicate.ID != first.ID || duplicate.EventNo != first.EventNo {
		t.Fatalf("expected duplicate publish to return original event, got first=%+v duplicate=%+v", first, duplicate)
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowEvents) != 1 || len(snapshot.WorkflowInstances) != 1 || len(snapshot.ApprovalTasks) != 1 || len(snapshot.WorkflowLogs) != 2 || len(snapshot.WorkflowOutbox) != 2 {
		t.Fatalf("expected duplicate event key not to create new workflow work, got events=%+v instances=%+v tasks=%+v logs=%+v outbox=%+v", snapshot.WorkflowEvents, snapshot.WorkflowInstances, snapshot.ApprovalTasks, snapshot.WorkflowLogs, snapshot.WorkflowOutbox)
	}
	if snapshot.WorkflowOutbox[0].Status != "pending" || snapshot.WorkflowOutbox[0].EventType != "workflow.instance_started" {
		t.Fatalf("expected pending workflow outbox event, got %+v", snapshot.WorkflowOutbox)
	}
	if snapshot.WorkflowOutbox[0].Payload["triggerEventNo"] != first.EventNo || snapshot.WorkflowOutbox[0].Payload["triggerEventType"] != first.EventType || snapshot.WorkflowOutbox[0].Payload["triggerSource"] != "integration" || snapshot.WorkflowOutbox[0].Payload["triggerEventKey"] != "QE-EXT-1" {
		t.Fatalf("expected workflow outbox to include trigger event identity, event=%+v outbox=%+v", first, snapshot.WorkflowOutbox[0])
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(snapshot.WorkflowOutbox[0].ID, 10)+"/ack", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("ack workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var acked WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &acked); err != nil {
		t.Fatalf("decode acked outbox: %v", err)
	}
	if acked.Status != "sent" || acked.AcknowledgedBy != "admin" || acked.AcknowledgedAt == "" {
		t.Fatalf("expected sent outbox, got %+v", acked)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(acked.ID, 10)+"/reset", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reset workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var reset WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &reset); err != nil {
		t.Fatalf("decode reset outbox: %v", err)
	}
	if reset.Status != "pending" || reset.Attempts != 1 || reset.AcknowledgedAt != "" {
		t.Fatalf("expected reset pending outbox, got %+v", reset)
	}
}

func TestWorkflowOutboxClaimFailAndRetryGuards(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_outbox","name":"质量异常出口测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.outbox_test","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.outbox_test","source":"test","eventKey":"outbox-claim-1","resource":"quality_exception","resourceId":9301,"resourceNo":"QE-OUTBOX-1","title":"出口测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowOutbox) < 1 {
		t.Fatalf("expected workflow outbox events, got %+v", snapshot.WorkflowOutbox)
	}
	outboxID := snapshot.WorkflowOutbox[0].ID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(outboxID, 10)+"/claim", `{"consumer":"delivery-worker"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("claim workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var claimed WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &claimed); err != nil {
		t.Fatalf("decode claimed outbox: %v", err)
	}
	if claimed.Status != "processing" || claimed.Attempts != 1 || claimed.ClaimedBy != "delivery-worker" || claimed.ClaimedAt == "" {
		t.Fatalf("expected processing claimed outbox, got %+v", claimed)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(outboxID, 10)+"/claim", `{"consumer":"delivery-worker-2"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate claim rejection, got status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(outboxID, 10)+"/fail", `{"error":"delivery timeout","retryAfterMinutes":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("fail workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var failed WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &failed); err != nil {
		t.Fatalf("decode failed outbox: %v", err)
	}
	if failed.Status != "failed" || failed.LastError != "delivery timeout" || failed.NextAttemptAt == "" || failed.Attempts != 1 {
		t.Fatalf("expected failed outbox with retry guard, got %+v", failed)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/outbox?status=failed", ``)
	if rec.Code != http.StatusOK {
		t.Fatalf("list failed workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var failedItems []WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &failedItems); err != nil {
		t.Fatalf("decode failed outbox list: %v", err)
	}
	if len(failedItems) != 1 || failedItems[0].ID != outboxID {
		t.Fatalf("expected failed outbox filter to return item %d, got %+v", outboxID, failedItems)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(outboxID, 10)+"/claim", `{"consumer":"delivery-worker"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected retry guard rejection, got status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(outboxID, 10)+"/reset", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reset workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var reset WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &reset); err != nil {
		t.Fatalf("decode reset outbox: %v", err)
	}
	if reset.Status != "pending" || reset.Attempts != 2 || reset.LastError != "" || reset.NextAttemptAt != "" || reset.ClaimedBy != "" {
		t.Fatalf("expected reset outbox to clear failure state, got %+v", reset)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/outbox/"+strconv.FormatInt(outboxID, 10)+"/claim", `{"consumer":"delivery-worker"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reclaim workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	var reclaimed WorkflowOutbox
	if err := json.Unmarshal(rec.Body.Bytes(), &reclaimed); err != nil {
		t.Fatalf("decode reclaimed outbox: %v", err)
	}
	if reclaimed.Status != "processing" || reclaimed.Attempts != 3 || reclaimed.ClaimedBy != "delivery-worker" {
		t.Fatalf("expected reclaimed processing outbox, got %+v", reclaimed)
	}
}

func TestWorkflowSubscriptionCreatesDeliveriesAndDispatchesWebhook(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-success","name":"工作流成功投递","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	var subscription WorkflowSubscription
	if err := json.Unmarshal(rec.Body.Bytes(), &subscription); err != nil {
		t.Fatalf("decode workflow subscription: %v", err)
	}
	if subscription.ID == 0 || subscription.Code != "workflow-success" {
		t.Fatalf("expected saved workflow subscription, got %+v", subscription)
	}
	if subscription.TimeoutSeconds != 5 {
		t.Fatalf("expected default workflow subscription timeout, got %+v", subscription)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_delivery","name":"质量异常投递测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.delivery_test","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.delivery_test","source":"test","eventKey":"delivery-1","resource":"quality_exception","resourceId":9401,"resourceNo":"QE-DELIVERY-1","title":"投递测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowOutbox) != 2 || len(snapshot.WorkflowDeliveries) != 2 {
		t.Fatalf("expected outbox deliveries for workflow logs, outbox=%+v deliveries=%+v", snapshot.WorkflowOutbox, snapshot.WorkflowDeliveries)
	}
	if snapshot.WorkflowDeliveries[0].Status != "pending" || snapshot.WorkflowDeliveries[0].SubscriptionCode != "workflow-success" {
		t.Fatalf("expected pending workflow delivery, got %+v", snapshot.WorkflowDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/deliveries?status=pending", ``)
	if rec.Code != http.StatusOK {
		t.Fatalf("list pending deliveries status %d: %s", rec.Code, rec.Body.String())
	}
	var pending []WorkflowDelivery
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending deliveries: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected two pending deliveries, got %+v", pending)
	}

	deliveryID := snapshot.WorkflowDeliveries[0].ID
	outboxID := snapshot.WorkflowDeliveries[0].OutboxID
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/"+strconv.FormatInt(deliveryID, 10)+"/dispatch", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dispatch workflow delivery status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatched WorkflowDelivery
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatched); err != nil {
		t.Fatalf("decode dispatched delivery: %v", err)
	}
	if dispatched.Status != "succeeded" || dispatched.ResponseStatus != http.StatusOK || dispatched.CompletedAt == "" || dispatched.RequestPayload == "" {
		t.Fatalf("expected succeeded workflow delivery, got %+v", dispatched)
	}
	snapshot = app.mustSnapshot()
	for _, item := range snapshot.WorkflowOutbox {
		if item.ID == outboxID && (item.Status != "sent" || item.AcknowledgedBy != "workflow-delivery") {
			t.Fatalf("expected workflow outbox to be acknowledged by delivery, got %+v", item)
		}
	}
}

func TestWorkflowSubscriptionDeleteDisablesPendingDeliveries(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-delete","name":"待删除订阅","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	var subscription WorkflowSubscription
	if err := json.Unmarshal(rec.Body.Bytes(), &subscription); err != nil {
		t.Fatalf("decode workflow subscription: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_delete_subscription","name":"质量异常删除订阅测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.delete_subscription","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.delete_subscription","source":"test","eventKey":"delete-subscription-1","resource":"quality_exception","resourceId":9501,"resourceNo":"QE-DELETE-SUB-1","title":"删除订阅测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	pendingDeliveryCount := 0
	for _, delivery := range snapshot.WorkflowDeliveries {
		if delivery.SubscriptionID == subscription.ID && delivery.Status == "pending" {
			pendingDeliveryCount++
		}
	}
	if pendingDeliveryCount == 0 {
		t.Fatalf("expected pending deliveries for subscription %d, got %+v", subscription.ID, snapshot.WorkflowDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/workflows/subscriptions/"+strconv.FormatInt(subscription.ID, 10), "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "启用中的工作流订阅不能删除") {
		t.Fatalf("active workflow subscription delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions/"+strconv.FormatInt(subscription.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("disable workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/workflows/subscriptions/"+strconv.FormatInt(subscription.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/subscriptions", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("workflow subscriptions after delete status %d: %s", rec.Code, rec.Body.String())
	}
	var subscriptions []WorkflowSubscription
	if err := json.Unmarshal(rec.Body.Bytes(), &subscriptions); err != nil {
		t.Fatalf("decode workflow subscriptions after delete: %v", err)
	}
	for _, item := range subscriptions {
		if item.ID == subscription.ID {
			t.Fatalf("deleted workflow subscription still listed: %+v", item)
		}
	}

	snapshot = app.mustSnapshot()
	for _, delivery := range snapshot.WorkflowDeliveries {
		if delivery.SubscriptionID != subscription.ID {
			continue
		}
		if delivery.Status != "dead" || delivery.LastError != "订阅已删除" {
			t.Fatalf("expected pending delivery to be closed after subscription delete, got %+v", delivery)
		}
	}
}

func TestWorkflowSubscriptionDispatchesSignedWebhook(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	secret := "workflow-webhook-secret"
	type receivedWebhook struct {
		EventType        string
		DeliveryNo       string
		SubscriptionCode string
		Timestamp        string
		Signature        string
		Body             []byte
	}
	received := receivedWebhook{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read workflow webhook body: %v", err)
		}
		received = receivedWebhook{
			EventType:        r.Header.Get("X-CBMP-Workflow-Event"),
			DeliveryNo:       r.Header.Get("X-CBMP-Workflow-Delivery"),
			SubscriptionCode: r.Header.Get("X-CBMP-Workflow-Subscription"),
			Timestamp:        r.Header.Get("X-CBMP-Timestamp"),
			Signature:        r.Header.Get("X-CBMP-Signature"),
			Body:             body,
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("workflow accepted"))
	}))
	defer server.Close()

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", `{"code":"workflow-http-signed","name":"签名 webhook","eventType":"workflow.*","targetType":"webhook","endpoint":"`+server.URL+`","secret":"`+secret+`","retryLimit":3,"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save signed workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_signed_delivery","name":"签名投递测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.signed_delivery","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create signed delivery workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.signed_delivery","source":"test","eventKey":"signed-delivery-1","resource":"quality_exception","resourceId":9411,"resourceNo":"QE-SIGNED-DELIVERY","title":"签名投递测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish signed delivery workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowDeliveries) == 0 {
		t.Fatalf("expected signed workflow delivery, got %+v", snapshot.WorkflowDeliveries)
	}
	deliveryID := snapshot.WorkflowDeliveries[0].ID
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/"+strconv.FormatInt(deliveryID, 10)+"/dispatch", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dispatch signed workflow delivery status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatched WorkflowDelivery
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatched); err != nil {
		t.Fatalf("decode signed dispatched delivery: %v", err)
	}
	if dispatched.Status != "succeeded" || dispatched.ResponseStatus != http.StatusAccepted || dispatched.ResponseBody != "workflow accepted" || dispatched.RequestPayload == "" {
		t.Fatalf("expected signed webhook delivery success, got %+v", dispatched)
	}
	if received.EventType != dispatched.EventType || received.DeliveryNo != dispatched.DeliveryNo || received.SubscriptionCode != "workflow-http-signed" {
		t.Fatalf("unexpected signed webhook headers: %+v delivery=%+v", received, dispatched)
	}
	if received.Timestamp == "" || received.Signature == "" {
		t.Fatalf("expected signed webhook timestamp and signature, got %+v", received)
	}
	if got, want := received.Signature, workflowDeliverySignature(secret, received.Timestamp, received.Body); got != want {
		t.Fatalf("unexpected signed webhook signature got %s want %s", got, want)
	}
	var payload struct {
		DeliveryNo       string            `json:"deliveryNo"`
		OutboxNo         string            `json:"outboxNo"`
		SubscriptionCode string            `json:"subscriptionCode"`
		EventType        string            `json:"eventType"`
		Resource         string            `json:"resource"`
		ResourceID       int64             `json:"resourceId"`
		Payload          map[string]string `json:"payload"`
		Attempts         int               `json:"attempts"`
	}
	if err := json.Unmarshal(received.Body, &payload); err != nil {
		t.Fatalf("decode signed webhook payload: %v", err)
	}
	if payload.DeliveryNo != dispatched.DeliveryNo || payload.EventType != dispatched.EventType || payload.Resource != "quality_exception" || payload.ResourceID != 9411 || payload.Attempts != 1 {
		t.Fatalf("unexpected signed webhook payload: %+v delivery=%+v", payload, dispatched)
	}
}

func TestWorkflowSubscriptionDispatchHonorsTimeoutSeconds(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("too slow"))
	}))
	defer server.Close()

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", `{"code":"workflow-http-timeout","name":"超时 webhook","eventType":"workflow.*","targetType":"webhook","endpoint":"`+server.URL+`","retryLimit":1,"timeoutSeconds":1,"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save timeout workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	var subscription WorkflowSubscription
	if err := json.Unmarshal(rec.Body.Bytes(), &subscription); err != nil {
		t.Fatalf("decode timeout workflow subscription: %v", err)
	}
	if subscription.TimeoutSeconds != 1 {
		t.Fatalf("expected workflow subscription timeout 1, got %+v", subscription)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_timeout_delivery","name":"超时投递测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.timeout_delivery","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save timeout workflow definition status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.timeout_delivery","source":"test","eventKey":"timeout-delivery-1","resource":"quality_exception","resourceId":9412,"resourceNo":"QE-TIMEOUT-DELIVERY","title":"超时投递测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish timeout workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowDeliveries) == 0 {
		t.Fatalf("expected timeout workflow delivery")
	}
	deliveryID := snapshot.WorkflowDeliveries[0].ID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/"+strconv.FormatInt(deliveryID, 10)+"/dispatch", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dispatch timeout workflow delivery status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatched WorkflowDelivery
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatched); err != nil {
		t.Fatalf("decode timeout dispatched workflow delivery: %v", err)
	}
	if dispatched.Status != "dead" || dispatched.ResponseStatus != 0 || dispatched.LastError == "" || dispatched.CompletedAt == "" {
		t.Fatalf("expected timed out delivery marked dead, got %+v", dispatched)
	}
	if dispatched.Payload["timeoutSeconds"] != "1" {
		t.Fatalf("expected timeout seconds in delivery payload, got %+v", dispatched.Payload)
	}
}

func TestWorkflowSubscriptionFiltersDeliveriesByResourceDefinitionAndEventType(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-filter-resource","name":"按资源订阅","eventType":"workflow.*","resource":"quality_exception","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save resource workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-filter-definition","name":"按流程订阅","eventType":"workflow.*","definitionCode":"quality_exception_subscription_filter","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save definition workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-filter-task-created","name":"按事件订阅","eventType":"workflow.task_created","resource":"quality_exception","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save event type workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-filter-miss","name":"不匹配订阅","eventType":"workflow.*","resource":"sales_order","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save nonmatching workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_subscription_filter","name":"质量异常订阅过滤测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.subscription_filter","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quality workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"ticket_void_subscription_filter","name":"磅单作废订阅过滤测试","category":"approval","resource":"ticket_void","trigger":{"eventType":"ticket_void.subscription_filter","resource":"ticket_void"},"steps":[{"seq":1,"roleCode":"dispatcher","action":"approve","name":"调度确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create ticket workflow definition status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.subscription_filter","source":"test","eventKey":"subscription-filter-quality","resource":"quality_exception","resourceId":9501,"resourceNo":"QE-SUB-FILTER","title":"质量异常订阅过滤"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish quality workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowOutbox) != 2 || len(snapshot.WorkflowDeliveries) != 5 {
		t.Fatalf("expected five filtered deliveries for quality workflow, outbox=%+v deliveries=%+v", snapshot.WorkflowOutbox, snapshot.WorkflowDeliveries)
	}
	counts := map[string]int{}
	for _, delivery := range snapshot.WorkflowDeliveries {
		counts[delivery.SubscriptionCode]++
		if delivery.SubscriptionCode == "workflow-filter-miss" {
			t.Fatalf("nonmatching subscription should not receive delivery, got %+v", delivery)
		}
		if delivery.SubscriptionCode == "workflow-filter-task-created" && delivery.EventType != "workflow.task_created" {
			t.Fatalf("event type filtered subscription received wrong event, got %+v", delivery)
		}
	}
	if counts["workflow-filter-resource"] != 2 || counts["workflow-filter-definition"] != 2 || counts["workflow-filter-task-created"] != 1 {
		t.Fatalf("unexpected filtered delivery counts: %+v deliveries=%+v", counts, snapshot.WorkflowDeliveries)
	}

	var filteredOutbox []WorkflowOutbox
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/outbox?eventType=workflow.task_created&resource=quality_exception&triggerEventKey=subscription-filter-quality", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow outbox by event/resource/trigger key status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredOutbox); err != nil {
		t.Fatalf("decode filtered workflow outbox: %v", err)
	}
	if len(filteredOutbox) != 1 || filteredOutbox[0].EventType != "workflow.task_created" || filteredOutbox[0].Payload["triggerEventKey"] != "subscription-filter-quality" {
		t.Fatalf("expected task-created outbox for quality trigger, got %+v", filteredOutbox)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/outbox?definitionCode=quality_exception_subscription_filter&resourceId=9501&triggerEventId="+strconv.FormatInt(filteredOutbox[0].TriggerEventID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow outbox by definition/resource id/trigger id status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredOutbox); err != nil {
		t.Fatalf("decode definition-filtered workflow outbox: %v", err)
	}
	if len(filteredOutbox) != 2 {
		t.Fatalf("expected two outbox events for quality workflow definition, got %+v", filteredOutbox)
	}

	var filteredInstances []WorkflowInstance
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/instances?resource=quality_exception&definitionCode=quality_exception_subscription_filter&status=pending&resourceId=9501", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow instances by resource/definition/status status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredInstances); err != nil {
		t.Fatalf("decode filtered workflow instances: %v", err)
	}
	if len(filteredInstances) != 1 || filteredInstances[0].ResourceNo != "QE-SUB-FILTER" || filteredInstances[0].TriggerEventID != filteredOutbox[0].TriggerEventID {
		t.Fatalf("expected one pending quality workflow instance, got %+v", filteredInstances)
	}

	var filteredTasks []WorkflowTask
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/tasks?resource=quality_exception&definitionCode=quality_exception_subscription_filter&roleCode=quality&status=pending&instanceId="+strconv.FormatInt(filteredInstances[0].ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow tasks by resource/definition/role/status status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredTasks); err != nil {
		t.Fatalf("decode filtered workflow tasks: %v", err)
	}
	if len(filteredTasks) != 1 || filteredTasks[0].TaskNo == "" || filteredTasks[0].RoleCode != "quality" {
		t.Fatalf("expected one pending quality workflow task, got %+v", filteredTasks)
	}

	if err := app.store.Mutate(func(data *AppData) error {
		for i := range data.WorkflowTasks {
			if data.WorkflowTasks[i].ID == filteredTasks[0].ID {
				data.WorkflowTasks[i].DueAt = "2000-01-01 00:00:00"
				return nil
			}
		}
		t.Fatalf("pending quality workflow task not found in store")
		return nil
	}); err != nil {
		t.Fatalf("mark workflow task overdue: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/tasks?overdue=true&roleCode=quality&resource=quality_exception", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter overdue workflow tasks status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredTasks); err != nil {
		t.Fatalf("decode overdue workflow tasks: %v", err)
	}
	if len(filteredTasks) != 1 || filteredTasks[0].DueAt != "2000-01-01 00:00:00" {
		t.Fatalf("expected overdue quality workflow task, got %+v", filteredTasks)
	}

	var filteredLogs []WorkflowLog
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/logs?resource=quality_exception&definitionCode=quality_exception_subscription_filter&instanceId="+strconv.FormatInt(filteredInstances[0].ID, 10)+"&taskId="+strconv.FormatInt(filteredTasks[0].ID, 10)+"&action=task_created&triggerEventId="+strconv.FormatInt(filteredOutbox[0].TriggerEventID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow logs by resource/definition/instance/task/action status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredLogs); err != nil {
		t.Fatalf("decode filtered workflow logs: %v", err)
	}
	if len(filteredLogs) != 1 || filteredLogs[0].Action != "task_created" || filteredLogs[0].TaskID != filteredTasks[0].ID {
		t.Fatalf("expected one task-created workflow log, got %+v", filteredLogs)
	}

	var filteredDeliveries []WorkflowDelivery
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/deliveries?subscriptionCode=workflow-filter-task-created&eventType=workflow.task_created&resource=quality_exception", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow deliveries by subscription/event/resource status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredDeliveries); err != nil {
		t.Fatalf("decode filtered workflow deliveries: %v", err)
	}
	if len(filteredDeliveries) != 1 || filteredDeliveries[0].SubscriptionCode != "workflow-filter-task-created" || filteredDeliveries[0].EventType != "workflow.task_created" {
		t.Fatalf("expected one task-created delivery, got %+v", filteredDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/deliveries?subscriptionCode=workflow-filter-definition&definitionCode=quality_exception_subscription_filter&triggerEventKey=subscription-filter-quality&targetType=webhook", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter workflow deliveries by definition/trigger/target status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredDeliveries); err != nil {
		t.Fatalf("decode definition-filtered workflow deliveries: %v", err)
	}
	if len(filteredDeliveries) != 2 {
		t.Fatalf("expected two definition-filtered deliveries, got %+v", filteredDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"ticket_void.subscription_filter","source":"test","eventKey":"subscription-filter-ticket","resource":"ticket_void","resourceId":9502,"resourceNo":"TV-SUB-FILTER","title":"磅单作废订阅过滤"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish ticket workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.WorkflowOutbox) != 4 || len(snapshot.WorkflowDeliveries) != 5 {
		t.Fatalf("ticket workflow should not create deliveries for quality-only subscriptions, outbox=%+v deliveries=%+v", snapshot.WorkflowOutbox, snapshot.WorkflowDeliveries)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/outbox?resource=ticket_void&definitionCode=ticket_void_subscription_filter", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter ticket workflow outbox status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredOutbox); err != nil {
		t.Fatalf("decode ticket workflow outbox: %v", err)
	}
	if len(filteredOutbox) != 2 {
		t.Fatalf("expected two ticket workflow outbox events, got %+v", filteredOutbox)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows/deliveries?resource=ticket_void", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filter ticket workflow deliveries status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredDeliveries); err != nil {
		t.Fatalf("decode ticket workflow deliveries: %v", err)
	}
	if len(filteredDeliveries) != 0 {
		t.Fatalf("ticket workflow should not have deliveries for quality subscriptions, got %+v", filteredDeliveries)
	}
}

func TestWorkflowSubscriptionBackfillsExistingPendingOutbox(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_delivery_backfill","name":"质量异常补投测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.delivery_backfill","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.delivery_backfill","source":"test","eventKey":"delivery-backfill-1","resource":"quality_exception","resourceId":9404,"resourceNo":"QE-DELIVERY-BACKFILL-1","title":"补投测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowOutbox) != 2 || len(snapshot.WorkflowDeliveries) != 0 {
		t.Fatalf("expected pending outbox without deliveries before subscription, outbox=%+v deliveries=%+v", snapshot.WorkflowOutbox, snapshot.WorkflowDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-backfill","name":"工作流补投","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.WorkflowDeliveries) != 2 {
		t.Fatalf("expected active subscription to backfill deliveries, got %+v", snapshot.WorkflowDeliveries)
	}
	for _, delivery := range snapshot.WorkflowDeliveries {
		if delivery.SubscriptionCode != "workflow-backfill" || delivery.Status != "pending" {
			t.Fatalf("expected pending backfilled delivery for workflow-backfill, got %+v", delivery)
		}
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-backfill","name":"工作流补投更新","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("update workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	if snapshot = app.mustSnapshot(); len(snapshot.WorkflowDeliveries) != 2 {
		t.Fatalf("expected subscription update not to duplicate deliveries, got %+v", snapshot.WorkflowDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-backfill-enabled","name":"工作流启用补投","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"disabled"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save disabled workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	var disabled WorkflowSubscription
	if err := json.Unmarshal(rec.Body.Bytes(), &disabled); err != nil {
		t.Fatalf("decode disabled workflow subscription: %v", err)
	}
	if snapshot = app.mustSnapshot(); len(snapshot.WorkflowDeliveries) != 2 {
		t.Fatalf("expected disabled subscription not to backfill deliveries, got %+v", snapshot.WorkflowDeliveries)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions/"+strconv.FormatInt(disabled.ID, 10)+"/status", `{"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("enable workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	if snapshot = app.mustSnapshot(); len(snapshot.WorkflowDeliveries) != 4 {
		t.Fatalf("expected enabling subscription to backfill deliveries, got %+v", snapshot.WorkflowDeliveries)
	}
}

func TestWorkflowDispatchDueDeliveriesProcessesReadyBatch(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)
	failEndpoint := newWorkflowWebhookEndpoint(t, http.StatusBadGateway)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-batch-success","name":"工作流批量成功投递","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save success workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-batch-fail","name":"工作流批量失败投递","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://fail","retryLimit":1,"status":"active"}`, failEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save failure workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_delivery_batch","name":"质量异常批量投递测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.delivery_batch","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.delivery_batch","source":"test","eventKey":"delivery-batch-1","resource":"quality_exception","resourceId":9403,"resourceNo":"QE-DELIVERY-BATCH-1","title":"批量投递测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowDeliveries) != 4 {
		t.Fatalf("expected four pending workflow deliveries, got %+v", snapshot.WorkflowDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/dispatch-due?limit=10", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dispatch due workflow deliveries status %d: %s", rec.Code, rec.Body.String())
	}
	var batch WorkflowDeliveryDispatchBatch
	if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
		t.Fatalf("decode dispatch due batch: %v", err)
	}
	if batch.Total != 4 || batch.Dispatched != 4 || batch.Succeeded != 2 || batch.Failed != 2 || batch.Skipped != 0 {
		t.Fatalf("unexpected dispatch due batch result: %+v", batch)
	}

	snapshot = app.mustSnapshot()
	succeeded := 0
	dead := 0
	for _, item := range snapshot.WorkflowDeliveries {
		switch item.Status {
		case "succeeded":
			succeeded++
		case "dead":
			dead++
		}
	}
	if succeeded != 2 || dead != 2 {
		t.Fatalf("expected two succeeded and two dead deliveries, got %+v", snapshot.WorkflowDeliveries)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/dispatch-due?limit=10", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("redispatch due workflow deliveries status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
		t.Fatalf("decode redispatch due batch: %v", err)
	}
	if batch.Total != 0 || batch.Dispatched != 0 {
		t.Fatalf("expected no due deliveries after batch dispatch, got %+v", batch)
	}
}

func TestWorkflowDispatchDueRecoverStaleProcessingDeliveries(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-stale-processing","name":"工作流卡死投递恢复","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_delivery_stale","name":"质量异常卡死投递测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.delivery_stale","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.delivery_stale","source":"test","eventKey":"delivery-stale-1","resource":"quality_exception","resourceId":9405,"resourceNo":"QE-DELIVERY-STALE-1","title":"卡死投递测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	staleAt := time.Now().Add(-workflowDeliveryProcessingTimeout - time.Minute).Format("2006-01-02 15:04:05")
	err := app.store.Mutate(func(data *AppData) error {
		if len(data.WorkflowDeliveries) == 0 {
			return fmt.Errorf("expected workflow delivery")
		}
		item := &data.WorkflowDeliveries[0]
		item.Status = "processing"
		item.Attempts = 1
		item.LastAttemptAt = staleAt
		item.UpdatedAt = staleAt
		item.NextAttemptAt = ""
		syncWorkflowOutboxDeliveryStatus(data, item.OutboxID)
		return nil
	})
	if err != nil {
		t.Fatalf("prepare stale workflow delivery: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/dispatch-due?limit=10", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dispatch due workflow deliveries status %d: %s", rec.Code, rec.Body.String())
	}
	var batch WorkflowDeliveryDispatchBatch
	if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
		t.Fatalf("decode stale dispatch batch: %v", err)
	}
	if batch.Recovered != 1 || batch.Dispatched == 0 || batch.Succeeded == 0 {
		t.Fatalf("expected stale delivery recovered and dispatched, got %+v", batch)
	}
	snapshot := app.mustSnapshot()
	if snapshot.WorkflowDeliveries[0].Status != "succeeded" || snapshot.WorkflowDeliveries[0].Payload["recoveredBy"] != "admin" {
		t.Fatalf("expected stale delivery succeeded after recovery, got %+v", snapshot.WorkflowDeliveries[0])
	}
	foundAudit := false
	for _, item := range snapshot.AuditLogs {
		if item.Action == "recover" && item.Resource == "workflow_delivery" && item.ResourceID == snapshot.WorkflowDeliveries[0].ID {
			foundAudit = true
			break
		}
	}
	if !foundAudit {
		t.Fatalf("expected workflow delivery recovery audit log, got %+v", snapshot.AuditLogs)
	}
}

func TestWorkflowDeliveryFailureIsAuditedAndMarksOutboxFailed(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	failEndpoint := newWorkflowWebhookEndpoint(t, http.StatusBadGateway)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-fail","name":"工作流失败投递","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://fail","retryLimit":1,"status":"active"}`, failEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save workflow subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_delivery_fail","name":"质量异常失败投递测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.delivery_fail","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.delivery_fail","source":"test","eventKey":"delivery-fail-1","resource":"quality_exception","resourceId":9402,"resourceNo":"QE-DELIVERY-FAIL-1","title":"失败投递测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowDeliveries) == 0 {
		t.Fatalf("expected workflow deliveries, got %+v", snapshot.WorkflowDeliveries)
	}
	deliveryID := snapshot.WorkflowDeliveries[0].ID
	outboxID := snapshot.WorkflowDeliveries[0].OutboxID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/"+strconv.FormatInt(deliveryID, 10)+"/dispatch", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dispatch failed workflow delivery status %d: %s", rec.Code, rec.Body.String())
	}
	var failed WorkflowDelivery
	if err := json.Unmarshal(rec.Body.Bytes(), &failed); err != nil {
		t.Fatalf("decode failed delivery: %v", err)
	}
	if failed.Status != "dead" || failed.ResponseStatus != http.StatusBadGateway || failed.LastError == "" || failed.CompletedAt == "" {
		t.Fatalf("expected failed workflow delivery to be audited as dead, got %+v", failed)
	}
	snapshot = app.mustSnapshot()
	for _, item := range snapshot.WorkflowOutbox {
		if item.ID == outboxID && (item.Status != "failed" || item.LastError == "") {
			t.Fatalf("expected workflow outbox failure state, got %+v", item)
		}
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/deliveries/"+strconv.FormatInt(deliveryID, 10)+"/reset", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reset workflow delivery status %d: %s", rec.Code, rec.Body.String())
	}
	var reset WorkflowDelivery
	if err := json.Unmarshal(rec.Body.Bytes(), &reset); err != nil {
		t.Fatalf("decode reset delivery: %v", err)
	}
	if reset.Status != "pending" || reset.LastError != "" || reset.CompletedAt != "" || reset.ResponseStatus != 0 || reset.Payload["resetBy"] != "admin" {
		t.Fatalf("expected reset workflow delivery to be pending, got %+v", reset)
	}
	snapshot = app.mustSnapshot()
	var resetOutbox WorkflowOutbox
	for _, item := range snapshot.WorkflowOutbox {
		if item.ID == outboxID {
			resetOutbox = item
			break
		}
	}
	if resetOutbox.Status != "pending" {
		t.Fatalf("expected outbox pending after delivery reset, got %+v", resetOutbox)
	}
	foundResetAudit := false
	for _, item := range snapshot.AuditLogs {
		if item.Action == "reset" && item.Resource == "workflow_delivery" && item.ResourceID == deliveryID {
			foundResetAudit = true
			break
		}
	}
	if !foundResetAudit {
		t.Fatalf("expected reset workflow delivery audit log, got %+v", snapshot.AuditLogs)
	}
}

func TestWorkflowInstanceCancelClosesTasksAndApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_cancel","name":"质量异常取消测试","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.cancel_test","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.cancel_test","source":"test","eventKey":"cancel-1","resource":"quality_exception","resourceId":9201,"resourceNo":"QE-CANCEL-1","title":"取消测试","variables":{"severity":"high"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].Status != "pending" {
		t.Fatalf("expected pending workflow instance, got %+v", snapshot.WorkflowInstances)
	}
	instanceID := snapshot.WorkflowInstances[0].ID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/instances/"+strconv.FormatInt(instanceID, 10)+"/cancel", `{"reason":"business cancelled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("cancel workflow instance status %d: %s", rec.Code, rec.Body.String())
	}
	var cancelled WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &cancelled); err != nil {
		t.Fatalf("decode cancelled instance: %v", err)
	}
	if cancelled.Status != "cancelled" || cancelled.CurrentTaskID != 0 || cancelled.CompletedAt == "" || len(cancelled.Actions) != 1 || cancelled.Actions[0].Action != "cancel" {
		t.Fatalf("expected cancelled workflow instance, got %+v", cancelled)
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].Status != "cancelled" || snapshot.WorkflowTasks[0].CompletedAt == "" {
		t.Fatalf("expected workflow task cancelled, got %+v", snapshot.WorkflowTasks)
	}
	if len(snapshot.ApprovalTasks) != 1 || snapshot.ApprovalTasks[0].Status != "cancelled" {
		t.Fatalf("expected approval task cancelled, got %+v", snapshot.ApprovalTasks)
	}
	if len(snapshot.WorkflowLogs) != 4 || snapshot.WorkflowLogs[2].Action != "task_cancelled" || snapshot.WorkflowLogs[3].Action != "instance_cancelled" {
		t.Fatalf("expected cancellation logs, got %+v", snapshot.WorkflowLogs)
	}
}

func TestWorkflowTaskActionEndpointCompletesApprovalAndBusinessResult(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":5000,"planTime":"2026-06-18 16:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create risky order status %d: %s", rec.Code, rec.Body.String())
	}
	var order SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode order: %v", err)
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].Status != "pending" {
		t.Fatalf("expected first workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[0].ID, 10)+"/act", `{"action":"approve","comment":"workflow task step"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("act first workflow task status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode first workflow action: %v", err)
	}
	if instance.Status != "pending" || instance.CurrentRole != "boss" {
		t.Fatalf("expected workflow to advance to boss, got %+v", instance)
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.ApprovalTasks) != 1 || snapshot.ApprovalTasks[0].CurrentRole != "boss" || snapshot.ApprovalTasks[0].Status != "pending" {
		t.Fatalf("expected approval task synced to boss, got %+v", snapshot.ApprovalTasks)
	}
	if len(snapshot.WorkflowTasks) != 2 || snapshot.WorkflowTasks[1].Status != "pending" {
		t.Fatalf("expected second workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[1].ID, 10)+"/act", `{"action":"approve","comment":"workflow task final"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("act final workflow task status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode final workflow action: %v", err)
	}
	if instance.Status != "approved" || instance.CurrentTaskID != 0 {
		t.Fatalf("expected workflow approved, got %+v", instance)
	}
	orders := fetchOrders(t, app, token)
	for _, item := range orders {
		if item.ID == order.ID {
			if item.Status != "submitted" {
				t.Fatalf("expected order submitted after workflow task action, got %+v", item)
			}
			return
		}
	}
	t.Fatalf("order not found after workflow task action")
}

func TestWorkflowTaskActionRollsBackWhenBusinessResultFails(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"delivery_sign_missing_result","name":"缺失签收回写测试","category":"approval","resource":"delivery_sign","trigger":{"eventType":"delivery_sign.missing_result","resource":"delivery_sign"},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"签收归档"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"delivery_sign.missing_result","source":"test","eventKey":"missing-result-1","resource":"delivery_sign","resourceId":909901,"resourceNo":"DS-MISSING","title":"缺失签收回写"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowInstances[0].Status != "pending" || snapshot.WorkflowTasks[0].Status != "pending" {
		t.Fatalf("expected pending workflow before failed result, instances=%+v tasks=%+v", snapshot.WorkflowInstances, snapshot.WorkflowTasks)
	}
	taskID := snapshot.WorkflowTasks[0].ID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"触发缺失业务对象"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected failed workflow action status, got %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].Status != "pending" || snapshot.WorkflowInstances[0].CurrentTaskID != taskID || len(snapshot.WorkflowInstances[0].Actions) != 0 {
		t.Fatalf("expected workflow instance rollback after failed result, got %+v", snapshot.WorkflowInstances)
	}
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].Status != "pending" || snapshot.WorkflowTasks[0].CompletedAt != "" {
		t.Fatalf("expected workflow task rollback after failed result, got %+v", snapshot.WorkflowTasks)
	}
	if len(snapshot.ApprovalTasks) != 1 || snapshot.ApprovalTasks[0].Status != "pending" || len(snapshot.ApprovalTasks[0].Actions) != 0 {
		t.Fatalf("expected approval task rollback after failed result, got %+v", snapshot.ApprovalTasks)
	}
	resultFailedLog := false
	for _, log := range snapshot.WorkflowLogs {
		if log.Action == "instance_approved" || log.Action == workflowResultAppliedAction {
			t.Fatalf("unexpected terminal workflow log after failed result: %+v", snapshot.WorkflowLogs)
		}
		if log.Action == workflowResultFailedAction && log.InstanceID == snapshot.WorkflowInstances[0].ID && log.Status == "failed" && log.Variables["intendedStatus"] == "approved" {
			resultFailedLog = true
		}
	}
	if !resultFailedLog {
		t.Fatalf("expected result_failed workflow log after failed result, got %+v", snapshot.WorkflowLogs)
	}
	resultFailedOutbox := false
	for _, outbox := range snapshot.WorkflowOutbox {
		if outbox.EventType == "workflow.result_failed" && outbox.InstanceID == snapshot.WorkflowInstances[0].ID && outbox.Payload["var.intendedStatus"] == "approved" {
			resultFailedOutbox = true
			break
		}
	}
	if !resultFailedOutbox {
		t.Fatalf("expected result_failed workflow outbox after failed result, got %+v", snapshot.WorkflowOutbox)
	}
}

func TestWorkflowTaskReassignSyncsInstanceAndApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":5000,"planTime":"2026-06-18 16:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create risky order status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].RoleCode != "dispatcher" {
		t.Fatalf("expected dispatcher workflow task, got %+v", snapshot.WorkflowTasks)
	}
	taskID := snapshot.WorkflowTasks[0].ID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/reassign", `{"roleCode":"boss","reason":"manager takeover"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reassign workflow task status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode reassigned instance: %v", err)
	}
	if instance.CurrentRole != "boss" || len(instance.Actions) != 1 || instance.Actions[0].Action != "reassign" {
		t.Fatalf("expected reassigned workflow instance, got %+v", instance)
	}
	snapshot = app.mustSnapshot()
	if snapshot.WorkflowTasks[0].RoleCode != "boss" || snapshot.WorkflowInstances[0].CurrentRole != "boss" {
		t.Fatalf("expected workflow task and instance role synced, got %+v / %+v", snapshot.WorkflowTasks, snapshot.WorkflowInstances)
	}
	if len(snapshot.ApprovalTasks) != 1 || snapshot.ApprovalTasks[0].CurrentRole != "boss" {
		t.Fatalf("expected approval task role synced, got %+v", snapshot.ApprovalTasks)
	}
	if len(snapshot.WorkflowLogs) != 3 || snapshot.WorkflowLogs[2].Action != "task_reassigned" || snapshot.WorkflowLogs[2].Variables["oldRole"] != "dispatcher" || snapshot.WorkflowLogs[2].Variables["newRole"] != "boss" {
		t.Fatalf("expected reassign workflow log, got %+v", snapshot.WorkflowLogs)
	}
	if len(snapshot.WorkflowOutbox) != 3 || snapshot.WorkflowOutbox[2].EventType != "workflow.task_reassigned" {
		t.Fatalf("expected reassign outbox event, got %+v", snapshot.WorkflowOutbox)
	}
}

func TestWorkflowTaskEscalationRequiresOverdueAndLogs(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_sla","name":"质量异常 SLA 审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.sla","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认","slaHours":1}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.sla","source":"test","eventKey":"sla-1","resource":"quality_exception","resourceId":9401,"resourceNo":"QE-SLA-1","title":"SLA 测试","variables":{"severity":"high"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].DueAt == "" || snapshot.WorkflowTasks[0].RoleCode != "quality" {
		t.Fatalf("expected workflow task with SLA due date, got %+v", snapshot.WorkflowTasks)
	}
	taskID := snapshot.WorkflowTasks[0].ID

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/escalate", `{"roleCode":"boss","reason":"SLA overdue"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected non-overdue escalation blocked, status %d: %s", rec.Code, rec.Body.String())
	}

	if err := app.store.Mutate(func(data *AppData) error {
		for i := range data.WorkflowTasks {
			if data.WorkflowTasks[i].ID == taskID {
				data.WorkflowTasks[i].DueAt = "2000-01-01 00:00:00"
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("force workflow task overdue: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/escalate", `{"roleCode":"boss","reason":"SLA overdue"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("escalate workflow task status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode escalated instance: %v", err)
	}
	if instance.CurrentRole != "boss" || len(instance.Actions) != 1 || instance.Actions[0].Action != "escalate" {
		t.Fatalf("expected escalated workflow instance, got %+v", instance)
	}
	snapshot = app.mustSnapshot()
	if snapshot.WorkflowTasks[0].RoleCode != "boss" || snapshot.WorkflowTasks[0].EscalatedFromRole != "quality" || snapshot.WorkflowTasks[0].EscalatedAt == "" {
		t.Fatalf("expected workflow task escalation metadata, got %+v", snapshot.WorkflowTasks[0])
	}
	if len(snapshot.ApprovalTasks) != 1 || snapshot.ApprovalTasks[0].CurrentRole != "boss" {
		t.Fatalf("expected approval task current role synced, got %+v", snapshot.ApprovalTasks)
	}
	if len(snapshot.WorkflowLogs) != 3 || snapshot.WorkflowLogs[2].Action != "task_escalated" || snapshot.WorkflowLogs[2].Variables["oldRole"] != "quality" || snapshot.WorkflowLogs[2].Variables["newRole"] != "boss" {
		t.Fatalf("expected task escalated workflow log, got %+v", snapshot.WorkflowLogs)
	}
	if len(snapshot.WorkflowOutbox) != 3 || snapshot.WorkflowOutbox[2].EventType != "workflow.task_escalated" {
		t.Fatalf("expected task escalated outbox event, got %+v", snapshot.WorkflowOutbox)
	}
}

func TestWorkflowAutomationRunDispatchesDueDeliveriesAndEscalatesOverdueTasks(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	successEndpoint := newWorkflowWebhookEndpoint(t, http.StatusOK)

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/subscriptions", workflowEndpointBody(`{"code":"workflow-automation-success","name":"工作流自动投递","eventType":"workflow.*","targetType":"webhook","endpoint":"mock://success","retryLimit":3,"status":"active"}`, successEndpoint))
	if rec.Code != http.StatusCreated {
		t.Fatalf("save workflow automation subscription status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_automation","name":"质量异常自动化审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.automation","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认","slaHours":1}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow automation definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.automation","source":"test","eventKey":"automation-1","resource":"quality_exception","resourceId":9801,"resourceNo":"QE-AUTO-1","title":"自动化测试","variables":{"severity":"high"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow automation event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowTasks) != 1 || len(snapshot.WorkflowDeliveries) != 2 {
		t.Fatalf("expected pending workflow task and deliveries, got tasks=%+v deliveries=%+v", snapshot.WorkflowTasks, snapshot.WorkflowDeliveries)
	}
	taskID := snapshot.WorkflowTasks[0].ID
	if err := app.store.Mutate(func(data *AppData) error {
		for i := range data.WorkflowTasks {
			if data.WorkflowTasks[i].ID == taskID {
				data.WorkflowTasks[i].DueAt = "2000-01-01 00:00:00"
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("force workflow automation task overdue: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/automation/run?deliveryLimit=10&escalationLimit=10", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("run workflow automation status %d: %s", rec.Code, rec.Body.String())
	}
	var run WorkflowAutomationRun
	if err := json.Unmarshal(rec.Body.Bytes(), &run); err != nil {
		t.Fatalf("decode workflow automation run: %v", err)
	}
	if run.Deliveries.Total != 2 || run.Deliveries.Succeeded != 2 || run.Escalated != 1 || len(run.Escalations) != 1 || run.Escalations[0].ToRole != "boss" {
		t.Fatalf("unexpected workflow automation result: %+v", run)
	}
	snapshot = app.mustSnapshot()
	if snapshot.WorkflowTasks[0].RoleCode != "boss" || snapshot.WorkflowTasks[0].EscalatedFromRole != "quality" || snapshot.WorkflowTasks[0].EscalatedAt == "" {
		t.Fatalf("expected workflow automation to escalate task, got %+v", snapshot.WorkflowTasks[0])
	}
	if snapshot.ApprovalTasks[0].CurrentRole != "boss" {
		t.Fatalf("expected approval task synced after automation escalation, got %+v", snapshot.ApprovalTasks)
	}
	sent := 0
	for _, delivery := range snapshot.WorkflowDeliveries {
		if delivery.Status == "succeeded" {
			sent++
		}
	}
	if sent != 2 {
		t.Fatalf("expected initial workflow deliveries sent by automation, got %+v", snapshot.WorkflowDeliveries)
	}
}

func TestWorkflowDefinitionVersionPublishAndRollbackIsolateRunningInstances(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_version","name":"质量异常版本审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.version","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"},{"seq":2,"roleCode":"boss","action":"approve","name":"管理确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}
	var v1 WorkflowDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &v1); err != nil {
		t.Fatalf("decode v1 definition: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.version","source":"test","eventKey":"version-1","resource":"quality_exception","resourceId":9501,"resourceNo":"QE-V1","title":"版本隔离 V1","variables":{"severity":"high"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish v1 event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].DefinitionID != v1.ID || snapshot.WorkflowInstances[0].CurrentRole != "quality" {
		t.Fatalf("expected v1 instance assigned to quality, got %+v", snapshot.WorkflowInstances)
	}
	oldTaskID := snapshot.WorkflowInstances[0].CurrentTaskID

	publishBody := `{"name":"质量异常版本审批新版","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.version","resource":"quality_exception"},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"管理确认"}],"status":"active"}`
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions/"+strconv.FormatInt(v1.ID, 10)+"/publish", publishBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish workflow version status %d: %s", rec.Code, rec.Body.String())
	}
	var v2 WorkflowDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &v2); err != nil {
		t.Fatalf("decode v2 definition: %v", err)
	}
	if v2.ID == v1.ID || v2.Version != 2 || v2.Status != "active" || len(v2.Steps) != 1 || v2.Steps[0].RoleCode != "boss" {
		t.Fatalf("expected active v2 definition, got %+v", v2)
	}
	snapshot = app.mustSnapshot()
	for _, definition := range snapshot.WorkflowDefinitions {
		if definition.ID == v1.ID && definition.Status != "disabled" {
			t.Fatalf("expected v1 disabled after publishing v2, got %+v", definition)
		}
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(oldTaskID, 10)+"/act", `{"action":"approve","comment":"old version step"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("act old version task status %d: %s", rec.Code, rec.Body.String())
	}
	var oldInstance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &oldInstance); err != nil {
		t.Fatalf("decode old version instance: %v", err)
	}
	if oldInstance.Status != "pending" || oldInstance.DefinitionID != v1.ID || oldInstance.CurrentStep != 2 || oldInstance.CurrentRole != "boss" {
		t.Fatalf("expected old instance to keep v1 two-step definition, got %+v", oldInstance)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.version","source":"test","eventKey":"version-2","resource":"quality_exception","resourceId":9502,"resourceNo":"QE-V2","title":"版本隔离 V2","variables":{"severity":"high"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish v2 event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	var newInstance WorkflowInstance
	for _, instance := range snapshot.WorkflowInstances {
		if instance.ResourceID == 9502 {
			newInstance = instance
			break
		}
	}
	if newInstance.ID == 0 || newInstance.DefinitionID != v2.ID || newInstance.CurrentRole != "boss" {
		t.Fatalf("expected new instance to use v2 definition, got %+v", newInstance)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions/"+strconv.FormatInt(v1.ID, 10)+"/rollback", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("rollback workflow version status %d: %s", rec.Code, rec.Body.String())
	}
	var rolledBack WorkflowDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &rolledBack); err != nil {
		t.Fatalf("decode rolled back definition: %v", err)
	}
	if rolledBack.ID != v1.ID || rolledBack.Status != "active" {
		t.Fatalf("expected v1 active after rollback, got %+v", rolledBack)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/events", `{"eventType":"quality_exception.version","source":"test","eventKey":"version-3","resource":"quality_exception","resourceId":9503,"resourceNo":"QE-V3","title":"版本隔离 V3","variables":{"severity":"high"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish rollback event status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	var rollbackInstance WorkflowInstance
	for _, instance := range snapshot.WorkflowInstances {
		if instance.ResourceID == 9503 {
			rollbackInstance = instance
			break
		}
	}
	if rollbackInstance.ID == 0 || rollbackInstance.DefinitionID != v1.ID || rollbackInstance.CurrentRole != "quality" {
		t.Fatalf("expected rollback instance to use v1 definition, got %+v", rollbackInstance)
	}
}

func TestWorkflowDefinitionProtectsPendingInstances(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":5000,"planTime":"2026-06-18 16:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create risky order status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	var definition WorkflowDefinition
	for _, item := range snapshot.WorkflowDefinitions {
		if item.Code == "order_credit_risk" {
			definition = item
			break
		}
	}
	if definition.ID == 0 || len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].Status != "pending" {
		t.Fatalf("expected pending order_credit_risk instance, got definition=%+v instances=%+v", definition, snapshot.WorkflowInstances)
	}

	renamed := definition
	renamed.Name = "销售订单风险审批-新版"
	renamed.Version++
	renamePayload, err := json.Marshal(renamed)
	if err != nil {
		t.Fatalf("marshal renamed workflow definition: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", string(renamePayload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected non-breaking workflow definition rename to succeed, status %d: %s", rec.Code, rec.Body.String())
	}

	changedSteps := renamed
	changedSteps.Steps = []WorkflowStep{{Seq: 1, RoleCode: "boss", Action: "approve", Name: "管理确认"}}
	stepChangePayload, err := json.Marshal(changedSteps)
	if err != nil {
		t.Fatalf("marshal changed workflow definition: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", string(stepChangePayload))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected destructive workflow definition change blocked, status %d: %s", rec.Code, rec.Body.String())
	}

	var flow ApprovalFlow
	for _, item := range snapshot.ApprovalFlows {
		if item.Code == "order_credit_risk" {
			flow = item
			break
		}
	}
	if flow.ID == 0 {
		t.Fatalf("approval flow not found: %+v", snapshot.ApprovalFlows)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/approval-flows/"+strconv.FormatInt(flow.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected approval flow disable blocked while pending, status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestWorkflowConditionNumericOperators(t *testing.T) {
	event := WorkflowEvent{
		Variables: map[string]string{
			"amount": "1250.5",
			"score":  "82",
		},
	}
	cases := []struct {
		name      string
		condition WorkflowCondition
		want      bool
	}{
		{name: "greater than", condition: WorkflowCondition{Field: "amount", Operator: "greater_than", Value: "1000"}, want: true},
		{name: "greater or equal", condition: WorkflowCondition{Field: "score", Operator: "greater_or_equal", Value: "82"}, want: true},
		{name: "less than false", condition: WorkflowCondition{Field: "amount", Operator: "less_than", Value: "1000"}, want: false},
		{name: "less or equal alias", condition: WorkflowCondition{Field: "score", Operator: "lte", Value: "82"}, want: true},
		{name: "invalid number", condition: WorkflowCondition{Field: "amount", Operator: "greater_than", Value: "high"}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := workflowConditionMatchesEvent(tc.condition, event); got != tc.want {
				t.Fatalf("workflowConditionMatchesEvent() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWorkflowResultDispatcherAppliesBusinessStatusAndIsIdempotent(t *testing.T) {
	data := AppData{
		Next: map[string]int64{},
		Orders: []SalesOrder{{
			ID:      42,
			OrderNo: "SO-DISPATCHER",
			Status:  "pending_approval",
		}},
	}
	instance := WorkflowInstance{
		ID:             7,
		InstanceNo:     "WF-DISPATCHER",
		DefinitionCode: "manual_dispatcher_test",
		Resource:       "sales_order",
		ResourceID:     42,
		ResourceNo:     "SO-DISPATCHER",
		Status:         "approved",
		Actions:        []WorkflowAction{{Actor: "boss", Action: "approve"}},
	}

	if err := applyWorkflowResult(&data, instance); err != nil {
		t.Fatalf("apply workflow result: %v", err)
	}
	if data.Orders[0].Status != "submitted" {
		t.Fatalf("expected order status submitted, got %+v", data.Orders[0])
	}
	if len(data.WorkflowLogs) != 1 || data.WorkflowLogs[0].Action != workflowResultAppliedAction || data.WorkflowLogs[0].Actor != "boss" {
		t.Fatalf("expected one result-applied log, got %+v", data.WorkflowLogs)
	}
	if len(data.WorkflowOutbox) != 1 || data.WorkflowOutbox[0].EventType != "workflow.result_applied" || data.WorkflowOutbox[0].Status != "pending" || data.WorkflowOutbox[0].Payload["var.resource"] != "sales_order" {
		t.Fatalf("expected one result outbox event, got %+v", data.WorkflowOutbox)
	}

	if err := applyWorkflowResult(&data, instance); err != nil {
		t.Fatalf("reapply workflow result: %v", err)
	}
	if len(data.WorkflowLogs) != 1 || len(data.WorkflowOutbox) != 1 {
		t.Fatalf("expected idempotent result dispatch, got logs=%+v outbox=%+v", data.WorkflowLogs, data.WorkflowOutbox)
	}
}

func fetchApprovalTasks(t *testing.T, app *App, token string) []ApprovalTask {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/approvals", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("approvals status %d: %s", rec.Code, rec.Body.String())
	}
	var tasks []ApprovalTask
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("decode approvals: %v", err)
	}
	return tasks
}

func fetchOrders(t *testing.T, app *App, token string) []SalesOrder {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/orders", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("orders status %d: %s", rec.Code, rec.Body.String())
	}
	var orders []SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &orders); err != nil {
		t.Fatalf("decode orders: %v", err)
	}
	return orders
}

func hasWorkflowDefinition(items []WorkflowDefinition, code, status string) bool {
	for _, item := range items {
		if item.Code == code && item.Status == status && len(item.Steps) > 0 {
			return true
		}
	}
	return false
}

func hasApprovalFlow(items []ApprovalFlow, code string) bool {
	for _, item := range items {
		if item.Code == code && len(item.Steps) > 0 {
			return true
		}
	}
	return false
}
