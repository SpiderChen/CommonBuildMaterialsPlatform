package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type scimRolePayload struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Primary bool   `json:"primary"`
}

type scimUserPayload struct {
	Schemas     []string          `json:"schemas"`
	UserName    string            `json:"userName"`
	DisplayName string            `json:"displayName"`
	Active      *bool             `json:"active"`
	Roles       []scimRolePayload `json:"roles"`
}

type scimPatchRequest struct {
	Operations []scimPatchOperation `json:"Operations"`
}

type scimPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type scimUserUpdate struct {
	Username    string
	DisplayName string
	Active      *bool
	RoleCode    string
}

func (a *App) systemSCIM(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "providers" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, publicSCIMProviders(a.mustSnapshot().SCIMProviders))
		return
	}
	if len(parts) == 1 && parts[0] == "providers" && r.Method == http.MethodPost {
		a.upsertSCIMProvider(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "providers" && parts[2] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req struct {
			Status string `json:"status"`
		}
		_ = readJSON(r, &req)
		var updated SCIMProvider
		topic := "system.scim.status"
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.SCIMProviders {
				if data.SCIMProviders[i].ID == id {
					status := fallback(strings.TrimSpace(req.Status), "enabled")
					if hasPendingWorkflowForResource(*data, "scim_provider", id) {
						updated = publicSCIMProvider(data.SCIMProviders[i])
						topic = "system.scim.status_requested"
						return nil
					}
					_, instances, err := publishSCIMProviderStatusWorkflow(data, data.SCIMProviders[i], status, session.User.Username)
					if err != nil {
						return err
					}
					if len(instances) > 0 {
						updated = publicSCIMProvider(data.SCIMProviders[i])
						topic = "system.scim.status_requested"
						addAudit(data, session.User.Username, "request_status", "scim_provider", id, data.SCIMProviders[i].Code+"/"+status, clientIP(r))
						return nil
					}
					next, err := applySCIMProviderStatusLocked(data, id, status)
					if err != nil {
						return err
					}
					updated = publicSCIMProvider(next)
					addAudit(data, session.User.Username, "status", "scim_provider", id, updated.Code+"/"+updated.Status, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("SCIM 提供商不存在")
		})
		a.respondMutation(w, err, updated, topic)
		return
	}
	if len(parts) == 2 && parts[0] == "providers" && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var deleted SCIMProvider
		err := a.store.Mutate(func(data *AppData) error {
			for i, item := range data.SCIMProviders {
				if item.ID != id {
					continue
				}
				if item.Status == "enabled" {
					return fmt.Errorf("启用中的 SCIM 提供商不能删除，请先停用")
				}
				if hasPendingWorkflowForResource(*data, "scim_provider", id) {
					return fmt.Errorf("SCIM 提供商存在待审批状态变更，不能删除")
				}
				deleted = publicSCIMProvider(item)
				data.SCIMProviders = append(data.SCIMProviders[:i], data.SCIMProviders[i+1:]...)
				addAudit(data, session.User.Username, "delete", "scim_provider", id, item.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("SCIM 提供商不存在")
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("system.scim.provider.deleted", deleted)
		writeJSON(w, http.StatusOK, deleted)
		return
	}
	writeError(w, http.StatusNotFound, "unknown scim system route")
}

func publishSCIMProviderStatusWorkflow(data *AppData, item SCIMProvider, targetStatus string, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "scim_provider.status_change_requested",
		Source:     "system",
		Resource:   "scim_provider",
		ResourceID: item.ID,
		ResourceNo: item.Code,
		Title:      "SCIM 提供商状态变更 " + item.Code,
		Actor:      actor,
		Reason:     "SCIM 提供商状态变更审批",
		Variables: map[string]string{
			"targetStatus":    targetStatus,
			"currentStatus":   item.Status,
			"code":            item.Code,
			"name":            item.Name,
			"defaultRoleCode": item.DefaultRoleCode,
			"companyId":       fmt.Sprintf("%d", item.CompanyID),
			"siteId":          fmt.Sprintf("%d", item.SiteID),
		},
	})
}

func applySCIMProviderStatusLocked(data *AppData, id int64, status string) (SCIMProvider, error) {
	status = fallback(strings.TrimSpace(status), "enabled")
	for i := range data.SCIMProviders {
		if data.SCIMProviders[i].ID == id {
			data.SCIMProviders[i].Status = status
			return data.SCIMProviders[i], nil
		}
	}
	return SCIMProvider{}, fmt.Errorf("SCIM 提供商不存在")
}

func (a *App) upsertSCIMProvider(w http.ResponseWriter, r *http.Request, session Session) {
	var req SCIMProvider
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid scim provider payload")
		return
	}
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "SCIM provider name/code required")
		return
	}
	var updated SCIMProvider
	err := a.store.Mutate(func(data *AppData) error {
		req.Code = strings.TrimSpace(req.Code)
		req.Name = strings.TrimSpace(req.Name)
		req.Status = fallback(req.Status, "enabled")
		req.DefaultRoleCode = fallback(req.DefaultRoleCode, "customer")
		if req.CompanyID == 0 {
			req.CompanyID = 1
		}
		for i := range data.SCIMProviders {
			if data.SCIMProviders[i].ID == req.ID || data.SCIMProviders[i].Code == req.Code {
				if req.BearerToken == "" {
					req.BearerToken = data.SCIMProviders[i].BearerToken
				}
				req.ID = data.SCIMProviders[i].ID
				req.CreatedAt = data.SCIMProviders[i].CreatedAt
				req.LastSyncAt = data.SCIMProviders[i].LastSyncAt
				data.SCIMProviders[i] = req
				updated = publicSCIMProvider(req)
				addAudit(data, session.User.Username, "update", "scim_provider", req.ID, req.Code, clientIP(r))
				return nil
			}
		}
		if req.BearerToken == "" {
			req.BearerToken = tokenString()
		}
		req.ID = nextID(data, "scimProvider")
		req.CreatedAt = nowString()
		data.SCIMProviders = append(data.SCIMProviders, req)
		updated = publicSCIMProvider(req)
		addAudit(data, session.User.Username, "create", "scim_provider", req.ID, req.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, "system.scim.provider")
}

func (a *App) scim(w http.ResponseWriter, r *http.Request, parts []string) {
	provider, ok := a.scimProviderFromRequest(r)
	if !ok {
		unauthorized(w)
		return
	}
	if len(parts) == 0 || parts[0] != "Users" {
		writeError(w, http.StatusNotFound, "unknown scim resource")
		return
	}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			a.listSCIMUsers(w, provider)
		case http.MethodPost:
			a.createOrUpdateSCIMUser(w, r, provider)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid scim user id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.getSCIMUser(w, provider, userID)
	case http.MethodPut:
		a.replaceSCIMUser(w, r, provider, userID)
	case http.MethodPatch:
		a.patchSCIMUser(w, r, provider, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) scimProviderFromRequest(r *http.Request) (SCIMProvider, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" || token == header {
		return SCIMProvider{}, false
	}
	for _, provider := range a.mustSnapshot().SCIMProviders {
		if provider.Status == "enabled" && provider.BearerToken != "" && provider.BearerToken == token {
			return provider, true
		}
	}
	return SCIMProvider{}, false
}

func (a *App) listSCIMUsers(w http.ResponseWriter, provider SCIMProvider) {
	users := []map[string]interface{}{}
	for _, user := range a.mustSnapshot().Users {
		if scimUserInProvider(user, provider) {
			users = append(users, scimUserResponse(user))
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": len(users),
		"Resources":    users,
		"startIndex":   1,
		"itemsPerPage": len(users),
	})
}

func (a *App) getSCIMUser(w http.ResponseWriter, provider SCIMProvider, id int64) {
	for _, user := range a.mustSnapshot().Users {
		if user.ID == id && scimUserInProvider(user, provider) {
			writeJSON(w, http.StatusOK, scimUserResponse(user))
			return
		}
	}
	writeError(w, http.StatusNotFound, "SCIM user not found")
}

func (a *App) createOrUpdateSCIMUser(w http.ResponseWriter, r *http.Request, provider SCIMProvider) {
	var payload scimUserPayload
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid scim user payload")
		return
	}
	update := scimUpdateFromPayload(payload, provider)
	if update.Username == "" {
		writeError(w, http.StatusBadRequest, "SCIM userName required")
		return
	}
	var user User
	statusCode := http.StatusCreated
	err := a.store.Mutate(func(data *AppData) error {
		index := -1
		for i := range data.Users {
			if data.Users[i].Username == update.Username && scimUserInProvider(data.Users[i], provider) {
				index = i
				break
			}
		}
		action := "create"
		if index >= 0 {
			statusCode = http.StatusOK
			previous := data.Users[index]
			applySCIMUpdate(&data.Users[index], update)
			action = scimAction(previous.Status, data.Users[index].Status, "update")
			user = publicUser(data.Users[index])
		} else {
			status := "active"
			if update.Active != nil && !*update.Active {
				status = "disabled"
			}
			salt, hash := makePassword(tokenString())
			user = User{
				ID: nextID(data, "user"), CompanyID: provider.CompanyID, SiteID: provider.SiteID,
				Username: update.Username, DisplayName: fallback(update.DisplayName, update.Username), RoleCode: update.RoleCode,
				PasswordSalt: salt, PasswordHash: hash, Status: status,
			}
			data.Users = append(data.Users, user)
			user = publicUser(user)
		}
		recordSCIMEvent(data, provider, user, action, "success", update.Username, "scim:"+provider.Code, clientIP(r))
		addAudit(data, "scim:"+provider.Code, action, "scim_user", user.ID, user.Username, clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, statusCode, scimUserResponse(user))
}

func (a *App) replaceSCIMUser(w http.ResponseWriter, r *http.Request, provider SCIMProvider, id int64) {
	var payload scimUserPayload
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid scim user payload")
		return
	}
	update := scimUpdateFromPayload(payload, provider)
	a.updateSCIMUser(w, r, provider, id, update)
}

func (a *App) patchSCIMUser(w http.ResponseWriter, r *http.Request, provider SCIMProvider, id int64) {
	var payload scimPatchRequest
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid scim patch payload")
		return
	}
	update := scimUpdateFromPatch(payload, provider)
	a.updateSCIMUser(w, r, provider, id, update)
}

func (a *App) updateSCIMUser(w http.ResponseWriter, r *http.Request, provider SCIMProvider, id int64, update scimUserUpdate) {
	var user User
	err := a.store.Mutate(func(data *AppData) error {
		index := -1
		for i := range data.Users {
			if data.Users[i].ID == id && scimUserInProvider(data.Users[i], provider) {
				index = i
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("SCIM user not found")
		}
		if update.Username != "" {
			for i := range data.Users {
				if i != index && data.Users[i].Username == update.Username {
					return fmt.Errorf("SCIM userName already exists")
				}
			}
		}
		previous := data.Users[index]
		applySCIMUpdate(&data.Users[index], update)
		user = publicUser(data.Users[index])
		action := scimAction(previous.Status, user.Status, "update")
		recordSCIMEvent(data, provider, user, action, "success", user.Username, "scim:"+provider.Code, clientIP(r))
		addAudit(data, "scim:"+provider.Code, action, "scim_user", user.ID, user.Username, clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, scimUserResponse(user))
}

func scimUpdateFromPayload(payload scimUserPayload, provider SCIMProvider) scimUserUpdate {
	role := scimRoleFromRoles(payload.Roles)
	return scimUserUpdate{
		Username:    strings.TrimSpace(payload.UserName),
		DisplayName: strings.TrimSpace(payload.DisplayName),
		Active:      payload.Active,
		RoleCode:    fallback(role, fallback(provider.DefaultRoleCode, "customer")),
	}
}

func scimUpdateFromPatch(payload scimPatchRequest, provider SCIMProvider) scimUserUpdate {
	update := scimUserUpdate{RoleCode: ""}
	for _, operation := range payload.Operations {
		path := strings.ToLower(strings.TrimSpace(operation.Path))
		if path == "" {
			applySCIMPatchObject(&update, operation.Value, provider)
			continue
		}
		applySCIMPatchValue(&update, path, operation.Value, provider)
	}
	return update
}

func applySCIMUpdate(user *User, update scimUserUpdate) {
	if update.Username != "" {
		user.Username = update.Username
	}
	if update.DisplayName != "" {
		user.DisplayName = update.DisplayName
	}
	if update.RoleCode != "" {
		user.RoleCode = update.RoleCode
	}
	if update.Active != nil {
		if *update.Active {
			user.Status = "active"
		} else {
			user.Status = "disabled"
		}
	}
}

func applySCIMPatchObject(update *scimUserUpdate, value interface{}, provider SCIMProvider) {
	fields, ok := value.(map[string]interface{})
	if !ok {
		return
	}
	for key, item := range fields {
		applySCIMPatchValue(update, strings.ToLower(key), item, provider)
	}
}

func applySCIMPatchValue(update *scimUserUpdate, path string, value interface{}, provider SCIMProvider) {
	switch strings.ToLower(strings.TrimSpace(path)) {
	case "active":
		if boolValue, ok := value.(bool); ok {
			update.Active = &boolValue
		}
	case "displayname":
		if text, ok := value.(string); ok {
			update.DisplayName = strings.TrimSpace(text)
		}
	case "username", "userName":
		if text, ok := value.(string); ok {
			update.Username = strings.TrimSpace(text)
		}
	case "roles":
		if role := scimRoleFromValue(value); role != "" {
			update.RoleCode = fallback(role, fallback(provider.DefaultRoleCode, "customer"))
		}
	}
}

func scimRoleFromRoles(roles []scimRolePayload) string {
	for _, role := range roles {
		if strings.TrimSpace(role.Value) != "" {
			return strings.TrimSpace(role.Value)
		}
	}
	return ""
}

func scimRoleFromValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]interface{}:
		if text, ok := typed["value"].(string); ok {
			return strings.TrimSpace(text)
		}
	case []interface{}:
		for _, item := range typed {
			if role := scimRoleFromValue(item); role != "" {
				return role
			}
		}
	}
	return ""
}

func scimAction(previousStatus, nextStatus, fallbackAction string) string {
	if previousStatus == "active" && nextStatus != "active" {
		return "deactivate"
	}
	if previousStatus != "active" && nextStatus == "active" {
		return "reactivate"
	}
	return fallbackAction
}

func scimUserInProvider(user User, provider SCIMProvider) bool {
	if provider.CompanyID != 0 && user.CompanyID != provider.CompanyID {
		return false
	}
	if provider.SiteID != 0 && user.SiteID != provider.SiteID {
		return false
	}
	return true
}

func scimUserResponse(user User) map[string]interface{} {
	return map[string]interface{}{
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"id":          strconv.FormatInt(user.ID, 10),
		"userName":    user.Username,
		"displayName": user.DisplayName,
		"active":      user.Status == "active",
		"roles":       []map[string]interface{}{{"value": user.RoleCode, "display": user.RoleCode, "primary": true}},
		"meta":        map[string]interface{}{"resourceType": "User"},
	}
}

func recordSCIMEvent(data *AppData, provider SCIMProvider, user User, action, status, detail, actor, ip string) {
	createdAt := nowString()
	for i := range data.SCIMProviders {
		if data.SCIMProviders[i].ID == provider.ID {
			data.SCIMProviders[i].LastSyncAt = createdAt
			break
		}
	}
	id := nextID(data, "scimEvent")
	data.SCIMEvents = append(data.SCIMEvents, SCIMProvisioningEvent{
		ID: id, EventNo: number("SCIM", id), ProviderID: provider.ID, ProviderCode: provider.Code,
		UserID: user.ID, Username: user.Username, Action: action, Status: status, Detail: detail,
		CreatedAt: createdAt, Actor: actor, IP: ip,
	})
	if len(data.SCIMEvents) > 500 {
		data.SCIMEvents = data.SCIMEvents[len(data.SCIMEvents)-500:]
	}
}

func publicSCIMProviders(items []SCIMProvider) []SCIMProvider {
	out := make([]SCIMProvider, 0, len(items))
	for _, item := range items {
		out = append(out, publicSCIMProvider(item))
	}
	return out
}

func publicSCIMProvider(item SCIMProvider) SCIMProvider {
	item.TenantID = 0
	item.BearerToken = ""
	return item
}
