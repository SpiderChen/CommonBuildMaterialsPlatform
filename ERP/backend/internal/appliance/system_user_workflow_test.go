package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestSystemUserCreateRequiresInitialPassword(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/users", `{"username":"api.user","displayName":"API 用户","roleCode":"dispatcher","status":"active"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "初始密码") {
		t.Fatalf("expected initial password rejection, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, ok := userByUsername(app.mustSnapshot().Users, "api.user"); ok {
		t.Fatalf("user should not be created without an explicit initial password")
	}
}

func TestSystemRoleDeleteRequiresNoReferences(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/roles", `{"code":"temp-auditor","name":"临时审计员","dataScope":"site","permissions":["system:audit"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create removable role status %d: %s", rec.Code, rec.Body.String())
	}
	var removable Role
	if err := json.Unmarshal(rec.Body.Bytes(), &removable); err != nil {
		t.Fatalf("decode removable role: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/roles/"+strconv.FormatInt(removable.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete removable role status %d: %s", rec.Code, rec.Body.String())
	}
	if _, ok := roleByCode(app.mustSnapshot().Roles, removable.Code); ok {
		t.Fatalf("role should be deleted: %+v", removable)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/roles/1", "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "内置超级管理员") {
		t.Fatalf("builtin role delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}
	dispatcher, ok := roleByCode(app.mustSnapshot().Roles, "dispatcher")
	if !ok {
		t.Fatal("expected dispatcher role in seed data")
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/roles/"+strconv.FormatInt(dispatcher.ID, 10), "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "用户引用") {
		t.Fatalf("referenced role delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSystemUserStatusWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"system_user_status_review","name":"用户状态变更审批","category":"approval","resource":"system_user","trigger":{"eventType":"system_user.status_change_requested","resource":"system_user","conditions":[{"field":"targetStatus","operator":"equals","value":"disabled"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"账号状态复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create system user status workflow status %d: %s", rec.Code, rec.Body.String())
	}

	user, ok := userByUsername(app.mustSnapshot().Users, "quality")
	if !ok || user.Status != "active" {
		t.Fatalf("expected active quality user seed, got found=%v user=%+v", ok, user)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/users/"+strconv.FormatInt(user.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request user status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending User
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending user: %v", err)
	}
	if pending.Status != "active" {
		t.Fatalf("user should remain active before workflow approval, got %+v", pending)
	}

	snapshot := app.mustSnapshot()
	current, ok := userByUsername(snapshot.Users, "quality")
	if !ok || current.Status != "active" {
		t.Fatalf("expected persisted active user before approval, got found=%v user=%+v", ok, current)
	}
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "system_user" && task.ResourceID == user.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending system user workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"账号状态复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve user status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	updated, ok := userByUsername(app.mustSnapshot().Users, "quality")
	if !ok || updated.Status != "disabled" {
		t.Fatalf("expected disabled user after workflow approval, got found=%v user=%+v", ok, updated)
	}
}

func TestOIDCProviderStatusWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	providerPayload := `{"name":"审批 OIDC","code":"workflow-oidc","issuer":"https://idp.example.com/cbmp","clientId":"cbmp-desktop","clientSecret":"workflow-oidc-secret","authUrl":"https://idp.example.com/oauth2/v1/authorize","tokenUrl":"https://idp.example.com/oauth2/v1/token","redirectUri":"http://127.0.0.1:8088/api/auth/sso/workflow-oidc/callback","scopes":["openid","profile","email"],"usernameClaim":"preferred_username","displayNameClaim":"name","roleCode":"boss","companyId":1,"autoProvision":true,"status":"enabled"}`
	rec := testRequest(t, app, token, http.MethodPost, "/api/system/sso/providers", providerPayload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create oidc provider status %d: %s", rec.Code, rec.Body.String())
	}
	var provider OIDCProvider
	if err := json.Unmarshal(rec.Body.Bytes(), &provider); err != nil {
		t.Fatalf("decode oidc provider: %v", err)
	}
	if provider.ID == 0 || provider.Status != "enabled" {
		t.Fatalf("unexpected oidc provider: %+v", provider)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"oidc_provider_status_review","name":"SSO 状态变更审批","category":"approval","resource":"oidc_provider","trigger":{"eventType":"oidc_provider.status_change_requested","resource":"oidc_provider","conditions":[{"field":"targetStatus","operator":"equals","value":"disabled"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"SSO 状态复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create oidc status workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/sso/providers/"+strconv.FormatInt(provider.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request oidc status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending OIDCProvider
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending oidc provider: %v", err)
	}
	if pending.Status != "enabled" {
		t.Fatalf("oidc provider should remain enabled before workflow approval, got %+v", pending)
	}

	snapshot := app.mustSnapshot()
	current, ok := oidcProviderByCode(snapshot.OIDCProviders, "workflow-oidc")
	if !ok || current.Status != "enabled" {
		t.Fatalf("expected persisted enabled oidc provider before approval, got found=%v provider=%+v", ok, current)
	}
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "oidc_provider" && task.ResourceID == provider.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending oidc provider workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"SSO 状态复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve oidc status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	updated, ok := oidcProviderByCode(app.mustSnapshot().OIDCProviders, "workflow-oidc")
	if !ok || updated.Status != "disabled" {
		t.Fatalf("expected disabled oidc provider after workflow approval, got found=%v provider=%+v", ok, updated)
	}
}

func TestSCIMProviderStatusWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/scim/providers", `{"name":"Workflow SCIM","code":"workflow-scim","companyId":1,"defaultRoleCode":"customer","status":"enabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create scim provider status %d: %s", rec.Code, rec.Body.String())
	}
	var provider SCIMProvider
	if err := json.Unmarshal(rec.Body.Bytes(), &provider); err != nil {
		t.Fatalf("decode scim provider: %v", err)
	}
	if provider.ID == 0 || provider.Status != "enabled" {
		t.Fatalf("unexpected scim provider: %+v", provider)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"scim_provider_status_review","name":"SCIM status review","category":"approval","resource":"scim_provider","trigger":{"eventType":"scim_provider.status_change_requested","resource":"scim_provider","conditions":[{"field":"targetStatus","operator":"equals","value":"disabled"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"SCIM status review"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create scim status workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/scim/providers/"+strconv.FormatInt(provider.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request scim status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending SCIMProvider
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending scim provider: %v", err)
	}
	if pending.Status != "enabled" {
		t.Fatalf("scim provider should remain enabled before workflow approval, got %+v", pending)
	}

	snapshot := app.mustSnapshot()
	current, ok := scimProviderByCode(snapshot.SCIMProviders, "workflow-scim")
	if !ok || current.Status != "enabled" {
		t.Fatalf("expected persisted enabled scim provider before approval, got found=%v provider=%+v", ok, current)
	}
	taskID := pendingWorkflowTaskID(snapshot, "scim_provider", provider.ID)
	if taskID == 0 {
		t.Fatalf("expected pending scim provider workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"SCIM status approved"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve scim status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	updated, ok := scimProviderByCode(app.mustSnapshot().SCIMProviders, "workflow-scim")
	if !ok || updated.Status != "disabled" {
		t.Fatalf("expected disabled scim provider after workflow approval, got found=%v provider=%+v", ok, updated)
	}
}

func TestGatewayRouteStatusWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	route, ok := gatewayRouteByPath(app.mustSnapshot().GatewayRoutes, "/api/")
	if !ok || route.Status != "active" {
		t.Fatalf("expected active /api/ gateway route seed, got found=%v route=%+v", ok, route)
	}

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"gateway_route_status_review","name":"Gateway route status review","category":"approval","resource":"gateway_route","trigger":{"eventType":"gateway_route.status_change_requested","resource":"gateway_route","conditions":[{"field":"targetStatus","operator":"equals","value":"disabled"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"Gateway route status review"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create gateway route status workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/gateway/routes/"+strconv.FormatInt(route.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request gateway route status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending GatewayRoute
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending gateway route: %v", err)
	}
	if pending.Status != "active" {
		t.Fatalf("gateway route should remain active before workflow approval, got %+v", pending)
	}

	snapshot := app.mustSnapshot()
	current, ok := gatewayRouteByPath(snapshot.GatewayRoutes, "/api/")
	if !ok || current.Status != "active" {
		t.Fatalf("expected persisted active gateway route before approval, got found=%v route=%+v", ok, current)
	}
	taskID := pendingWorkflowTaskID(snapshot, "gateway_route", route.ID)
	if taskID == 0 {
		t.Fatalf("expected pending gateway route workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"Gateway route status approved"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve gateway route status workflow status %d: %s", rec.Code, rec.Body.String())
	}
	updated, ok := gatewayRouteByPath(app.mustSnapshot().GatewayRoutes, "/api/")
	if !ok || updated.Status != "disabled" {
		t.Fatalf("expected disabled gateway route after workflow approval, got found=%v route=%+v", ok, updated)
	}
}

func roleByCode(items []Role, code string) (Role, bool) {
	for _, item := range items {
		if item.Code == code {
			return item, true
		}
	}
	return Role{}, false
}

func userByUsername(items []User, username string) (User, bool) {
	for _, item := range items {
		if item.Username == username {
			return item, true
		}
	}
	return User{}, false
}

func oidcProviderByCode(items []OIDCProvider, code string) (OIDCProvider, bool) {
	for _, item := range items {
		if item.Code == code {
			return item, true
		}
	}
	return OIDCProvider{}, false
}

func scimProviderByCode(items []SCIMProvider, code string) (SCIMProvider, bool) {
	for _, item := range items {
		if item.Code == code {
			return item, true
		}
	}
	return SCIMProvider{}, false
}

func gatewayRouteByPath(items []GatewayRoute, path string) (GatewayRoute, bool) {
	for _, item := range items {
		if item.PathPrefix == path {
			return item, true
		}
	}
	return GatewayRoute{}, false
}

func pendingWorkflowTaskID(data AppData, resource string, resourceID int64) int64 {
	for _, task := range data.WorkflowTasks {
		if task.Resource == resource && task.ResourceID == resourceID && task.Status == "pending" {
			return task.ID
		}
	}
	return 0
}
