package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

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
