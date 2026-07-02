package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestCustomerBlacklistReleaseWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"customer_blacklist_release_review","name":"客户黑名单解除审批","category":"approval","resource":"customer_blacklist_release","trigger":{"eventType":"customer_blacklist_release.requested","resource":"customer_blacklist_release","conditions":[{"field":"severity","operator":"equals","value":"high"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"风控复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create blacklist release workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-blacklists", `{"customerId":1,"reason":"逾期回款停供","severity":"high","blockOrders":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer blacklist status %d: %s", rec.Code, rec.Body.String())
	}
	var blacklist CustomerBlacklist
	if err := json.Unmarshal(rec.Body.Bytes(), &blacklist); err != nil {
		t.Fatalf("decode customer blacklist: %v", err)
	}
	if blacklist.Status != "active" || customerStatus(app.mustSnapshot().Customers, blacklist.CustomerID) != "blocked" {
		t.Fatalf("expected active blacklist and blocked customer, got blacklist %+v customers %+v", blacklist, app.mustSnapshot().Customers)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-blacklists/"+strconv.FormatInt(blacklist.ID, 10)+"/release", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request blacklist release workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode workflow instance: %v", err)
	}
	if instance.Resource != "customer_blacklist_release" || instance.ResourceID != blacklist.ID || instance.Status != "pending" {
		t.Fatalf("unexpected workflow instance: %+v", instance)
	}
	snapshot := app.mustSnapshot()
	if customerBlacklistStatus(snapshot.CustomerBlacklists, blacklist.ID) != "active" || customerStatus(snapshot.Customers, blacklist.CustomerID) != "blocked" {
		t.Fatalf("blacklist should remain active before approval, got blacklist %+v customers %+v", snapshot.CustomerBlacklists, snapshot.Customers)
	}

	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.InstanceID == instance.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending release workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"风控复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve blacklist release workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if customerBlacklistStatus(snapshot.CustomerBlacklists, blacklist.ID) != "released" || customerStatus(snapshot.Customers, blacklist.CustomerID) != "active" {
		t.Fatalf("expected released blacklist and active customer, got blacklist %+v customers %+v", snapshot.CustomerBlacklists, snapshot.Customers)
	}
}

func TestCustomerBlacklistCreateWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"customer_blacklist_create_review","name":"客户黑名单审批","category":"approval","resource":"customer_blacklist","trigger":{"eventType":"customer_blacklist.submitted","resource":"customer_blacklist","conditions":[{"field":"severity","operator":"equals","value":"high"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"风控复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create blacklist workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-blacklists", `{"customerId":1,"reason":"逾期回款停供","severity":"high","blockOrders":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request customer blacklist workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var blacklist CustomerBlacklist
	if err := json.Unmarshal(rec.Body.Bytes(), &blacklist); err != nil {
		t.Fatalf("decode customer blacklist: %v", err)
	}
	if blacklist.Status != "pending_approval" {
		t.Fatalf("expected pending blacklist before workflow approval, got %+v", blacklist)
	}
	snapshot := app.mustSnapshot()
	if customerStatus(snapshot.Customers, blacklist.CustomerID) == "blocked" {
		t.Fatalf("customer should not be blocked before workflow approval, got %+v", snapshot.Customers)
	}

	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "customer_blacklist" && task.ResourceID == blacklist.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending blacklist workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"风控复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve blacklist workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if customerBlacklistStatus(snapshot.CustomerBlacklists, blacklist.ID) != "active" || customerStatus(snapshot.Customers, blacklist.CustomerID) != "blocked" {
		t.Fatalf("expected active blacklist and blocked customer after workflow approval, got blacklist %+v customers %+v", snapshot.CustomerBlacklists, snapshot.Customers)
	}
}

func customerBlacklistStatus(items []CustomerBlacklist, id int64) string {
	for _, item := range items {
		if item.ID == id {
			return item.Status
		}
	}
	return ""
}

func customerStatus(items []Customer, id int64) string {
	for _, item := range items {
		if item.ID == id {
			return item.Status
		}
	}
	return ""
}
