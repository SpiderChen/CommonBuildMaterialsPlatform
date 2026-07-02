package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestWildcardPermissionGrantsAccess(t *testing.T) {
	data := AppData{
		Roles: []Role{{Code: "boss", Name: "Super Admin", Permissions: []string{"*"}, DataScope: "group"}},
	}
	user := User{RoleCode: "boss"}

	if !canAccess(data, user, "system:read") {
		t.Fatal("wildcard permission should grant access")
	}
}

func TestEmptyRolePermissionsGrantNoAccess(t *testing.T) {
	data := AppData{
		Roles: []Role{{Code: "empty-admin", Name: "Empty Admin", Permissions: nil, DataScope: "group"}},
	}
	user := User{RoleCode: "empty-admin"}

	if canAccess(data, user, "system:read") {
		t.Fatal("empty role permissions should not grant access")
	}
}

func TestEnsureEnterpriseDefaultsRepairsBuiltinAdminSuperAdmin(t *testing.T) {
	data := SeedData()
	for i := range data.Roles {
		if data.Roles[i].Code == builtinSuperAdminRoleCode {
			data.Roles[i].Name = "集团管理员"
			data.Roles[i].Permissions = nil
			data.Roles[i].DataScope = "site"
		}
	}
	for i := range data.Users {
		if data.Users[i].Username == builtinAdminUsername {
			data.Users[i].DisplayName = "集团管理员"
			data.Users[i].RoleCode = "dispatcher"
		}
	}

	if !ensureEnterpriseDefaults(&data) {
		t.Fatal("expected default repair to change broken builtin admin configuration")
	}
	var repairedRole Role
	for _, role := range data.Roles {
		if role.Code == builtinSuperAdminRoleCode {
			repairedRole = role
			break
		}
	}
	if repairedRole.Name != builtinSuperAdminRoleName || !hasOnlyWildcardPermission(repairedRole.Permissions) || repairedRole.DataScope != "group" {
		t.Fatalf("builtin super admin role was not repaired: %+v", repairedRole)
	}
	var repairedAdmin User
	for _, user := range data.Users {
		if user.Username == builtinAdminUsername {
			repairedAdmin = user
			break
		}
	}
	if repairedAdmin.DisplayName != builtinSuperAdminRoleName || repairedAdmin.RoleCode != builtinSuperAdminRoleCode {
		t.Fatalf("builtin admin user was not repaired: %+v", repairedAdmin)
	}
}

func TestEnsureEnterpriseDefaultsGrantsQualityWorkflowTaskPermissions(t *testing.T) {
	data := SeedData()
	for i := range data.Roles {
		if data.Roles[i].Code == "quality" {
			data.Roles[i].Permissions = []string{"bootstrap:read", "dashboard:read", "production:read", "quality:*", "procurement:read"}
		}
	}

	if !ensureEnterpriseDefaults(&data) {
		t.Fatal("expected default repair to grant quality workflow permissions")
	}
	qualityUser := User{RoleCode: "quality"}
	if !canAccess(data, qualityUser, "approval:read") || !canAccess(data, qualityUser, "approval:write") {
		t.Fatalf("quality role should be able to read and act workflow tasks: %+v", data.Roles)
	}
}

func TestRegularRoleStillRequiresPermissionGrant(t *testing.T) {
	data := AppData{
		Roles: []Role{{Code: "dispatcher", Name: "Dispatcher", Permissions: []string{"dispatch:read"}, DataScope: "site"}},
	}
	user := User{RoleCode: "dispatcher"}

	if canAccess(data, user, "system:read") {
		t.Fatal("regular roles should not bypass stored permissions")
	}
}

func TestWorkflowRoutesUseApprovalPermissionsForBusinessUse(t *testing.T) {
	if perm := requiredPermission([]string{"system", "workflows"}, "GET"); perm != "approval:read" {
		t.Fatalf("workflow overview should require approval read, got %s", perm)
	}
	if perm := requiredPermission([]string{"system", "workflows", "tasks", "12", "act"}, "POST"); perm != "approval:write" {
		t.Fatalf("workflow task action should require approval write, got %s", perm)
	}
	if perm := requiredPermission([]string{"system", "workflows", "definitions"}, "POST"); perm != "system:write" {
		t.Fatalf("workflow configuration should still require system write, got %s", perm)
	}
}

func TestSystemSecurityConfigurationAPIs(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/security/policies", `{"name":"测试会话上限","type":"test_session_limit","value":"3","enabled":true,"remark":"test"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save security policy status %d: %s", rec.Code, rec.Body.String())
	}
	var policy SecurityPolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &policy); err != nil {
		t.Fatalf("decode security policy: %v", err)
	}
	if policy.ID == 0 || policy.Type != "test_session_limit" || !policy.Enabled {
		t.Fatalf("unexpected security policy: %+v", policy)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/security/policies/"+strconv.FormatInt(policy.ID, 10), "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "启用中的安全策略不能删除") {
		t.Fatalf("enabled security policy delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/security/policies/"+strconv.FormatInt(policy.ID, 10)+"/toggle", `{"enabled":false}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("toggle security policy status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &policy); err != nil {
		t.Fatalf("decode toggled security policy: %v", err)
	}
	if policy.Enabled {
		t.Fatalf("security policy should be disabled: %+v", policy)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/security/policies/"+strconv.FormatInt(policy.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete security policy status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("security after policy delete status %d: %s", rec.Code, rec.Body.String())
	}
	var security struct {
		Policies          []SecurityPolicy   `json:"policies"`
		DeviceCredentials []DeviceCredential `json:"deviceCredentials"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode security after policy delete: %v", err)
	}
	for _, item := range security.Policies {
		if item.ID == policy.ID {
			t.Fatalf("deleted security policy still listed: %+v", item)
		}
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/security/device-credentials", `{"deviceNo":"QA-SCALE-01","deviceKey":"qa-device-key","scopes":["scale:report"],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save device credential status %d: %s", rec.Code, rec.Body.String())
	}
	var credential DeviceCredential
	if err := json.Unmarshal(rec.Body.Bytes(), &credential); err != nil {
		t.Fatalf("decode device credential: %v", err)
	}
	if credential.ID == 0 || credential.DeviceNo != "QA-SCALE-01" || credential.KeyHash != "" || credential.Status != "active" {
		t.Fatalf("unexpected public device credential: %+v", credential)
	}
	snapshot := app.mustSnapshot()
	var stored DeviceCredential
	for _, item := range snapshot.DeviceCredentials {
		if item.ID == credential.ID {
			stored = item
			break
		}
	}
	if stored.KeyHash != sha256Hex("qa-device-key") {
		t.Fatalf("device key hash was not stored correctly: %+v", stored)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/security/device-credentials/"+strconv.FormatInt(credential.ID, 10), "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "启用中的设备凭证不能删除") {
		t.Fatalf("active device credential delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/security/device-credentials/"+strconv.FormatInt(credential.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("disable device credential status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &credential); err != nil {
		t.Fatalf("decode disabled device credential: %v", err)
	}
	if credential.Status != "disabled" || credential.KeyHash != "" {
		t.Fatalf("device credential should be disabled and sanitized: %+v", credential)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/system/security/device-credentials/"+strconv.FormatInt(credential.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete device credential status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("security after credential delete status %d: %s", rec.Code, rec.Body.String())
	}
	security = struct {
		Policies          []SecurityPolicy   `json:"policies"`
		DeviceCredentials []DeviceCredential `json:"deviceCredentials"`
	}{}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode security after credential delete: %v", err)
	}
	for _, item := range security.DeviceCredentials {
		if item.ID == credential.ID {
			t.Fatalf("deleted device credential still listed: %+v", item)
		}
		if item.KeyHash != "" {
			t.Fatalf("public device credential must not expose key hash: %+v", item)
		}
	}
}

func TestIntegrationEndpointConfigurationAPIs(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/integrations/endpoints", `{"name":"催收短信供应商","type":"collection_sms","protocol":"rest/http","url":"mock://sms","status":"online"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("mock endpoint should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/integrations/endpoints", `{"name":"催收短信供应商","type":"collection_sms","protocol":"rest/http","url":"https://sms.example.test/send","status":"online"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save integration endpoint status %d: %s", rec.Code, rec.Body.String())
	}
	var endpoint IntegrationEndpoint
	if err := json.Unmarshal(rec.Body.Bytes(), &endpoint); err != nil {
		t.Fatalf("decode integration endpoint: %v", err)
	}
	if endpoint.ID == 0 || endpoint.URL != "https://sms.example.test/send" || endpoint.Status != "online" {
		t.Fatalf("unexpected integration endpoint: %+v", endpoint)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/integrations/endpoints/"+strconv.FormatInt(endpoint.ID, 10), "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "启用中的集成端点不能删除") {
		t.Fatalf("active integration endpoint delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/integrations/endpoints/"+strconv.FormatInt(endpoint.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("disable integration endpoint status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &endpoint); err != nil {
		t.Fatalf("decode disabled integration endpoint: %v", err)
	}
	if endpoint.Status != "disabled" {
		t.Fatalf("integration endpoint should be disabled: %+v", endpoint)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/integrations/endpoints/"+strconv.FormatInt(endpoint.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete integration endpoint status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/integrations/endpoints", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("integration endpoints after delete status %d: %s", rec.Code, rec.Body.String())
	}
	var endpoints []IntegrationEndpoint
	if err := json.Unmarshal(rec.Body.Bytes(), &endpoints); err != nil {
		t.Fatalf("decode integration endpoints after delete: %v", err)
	}
	for _, item := range endpoints {
		if item.ID == endpoint.ID {
			t.Fatalf("deleted integration endpoint still listed: %+v", item)
		}
	}
}

func TestRuleDefinitionConfigurationAPIs(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/rules/definitions", `{"code":"qa_speed_limit","name":"测试超速规则","category":"vehicle","metric":"speed","operator":">","threshold":88,"level":"warning","enabled":true,"notifyRoles":["dispatcher"],"description":"test rule"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save rule definition status %d: %s", rec.Code, rec.Body.String())
	}
	var rule RuleDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &rule); err != nil {
		t.Fatalf("decode rule definition: %v", err)
	}
	if rule.ID == 0 || rule.Code != "qa_speed_limit" || !rule.Enabled || rule.Threshold != 88 {
		t.Fatalf("unexpected rule definition: %+v", rule)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/rules/definitions/"+strconv.FormatInt(rule.ID, 10), "")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "启用中的规则不能删除") {
		t.Fatalf("enabled rule definition delete should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/rules/definitions", `{"id":`+strconv.FormatInt(rule.ID, 10)+`,"code":"qa_speed_limit","name":"测试超速规则更新","category":"vehicle","metric":"speed","operator":">=","threshold":92,"level":"critical","enabled":true,"notifyRoles":["boss"],"description":"updated"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("update rule definition status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rule); err != nil {
		t.Fatalf("decode updated rule definition: %v", err)
	}
	if rule.Name != "测试超速规则更新" || rule.Level != "critical" || rule.NotifyRoles[0] != "boss" {
		t.Fatalf("rule definition was not updated: %+v", rule)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/rules/definitions/"+strconv.FormatInt(rule.ID, 10)+"/status", `{"enabled":false}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("disable rule definition status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rule); err != nil {
		t.Fatalf("decode disabled rule definition: %v", err)
	}
	if rule.Enabled {
		t.Fatalf("rule definition should be disabled: %+v", rule)
	}

	rec = testRequest(t, app, token, http.MethodDelete, "/api/rules/definitions/"+strconv.FormatInt(rule.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete rule definition status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/rules/definitions", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("rule definitions after delete status %d: %s", rec.Code, rec.Body.String())
	}
	var rules []RuleDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &rules); err != nil {
		t.Fatalf("decode rule definitions after delete: %v", err)
	}
	for _, item := range rules {
		if item.ID == rule.ID {
			t.Fatalf("deleted rule definition still listed: %+v", item)
		}
	}
}
