package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (a *App) systemApprovalFlows(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, a.mustSnapshot().ApprovalFlows)
		return
	}
	if len(parts) == 0 && r.Method == http.MethodPost {
		var item ApprovalFlow
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid approval flow")
			return
		}
		var saved ApprovalFlow
		err := a.store.Mutate(func(data *AppData) error {
			if err := normalizeApprovalFlow(&item); err != nil {
				return err
			}
			for i := range data.ApprovalFlows {
				if (item.ID > 0 && data.ApprovalFlows[i].ID == item.ID) || (item.ID == 0 && data.ApprovalFlows[i].Code == item.Code) {
					item.ID = data.ApprovalFlows[i].ID
					data.ApprovalFlows[i] = item
					saved = item
					addAudit(data, session.User.Username, "update", "approval_flow", item.ID, item.Code, clientIP(r))
					return nil
				}
			}
			item.ID = nextID(data, "approvalFlow")
			data.ApprovalFlows = append(data.ApprovalFlows, item)
			saved = item
			addAudit(data, session.User.Username, "create", "approval_flow", item.ID, item.Code, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "system.approval_flow.saved")
		return
	}
	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Status string `json:"status"`
		}
		_ = readJSON(r, &req)
		var updated ApprovalFlow
		err := a.store.Mutate(func(data *AppData) error {
			status, err := normalizeConfigStatus(req.Status)
			if err != nil {
				return err
			}
			for i := range data.ApprovalFlows {
				if data.ApprovalFlows[i].ID != id {
					continue
				}
				data.ApprovalFlows[i].Status = status
				updated = data.ApprovalFlows[i]
				addAudit(data, session.User.Username, "status", "approval_flow", id, updated.Code+"/"+status, clientIP(r))
				return nil
			}
			return fmt.Errorf("审批流模板不存在")
		})
		a.respondMutation(w, err, updated, "system.approval_flow.updated")
		return
	}
	writeError(w, http.StatusNotFound, "unknown approval flow route")
}

func (a *App) systemDictionaries(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, a.mustSnapshot().DataDictionaries)
		return
	}
	if len(parts) == 0 && r.Method == http.MethodPost {
		var item DataDictionary
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid dictionary")
			return
		}
		var saved DataDictionary
		err := a.store.Mutate(func(data *AppData) error {
			if err := normalizeDataDictionary(&item, data.DataDictionaries); err != nil {
				return err
			}
			for i := range data.DataDictionaries {
				if (item.ID > 0 && data.DataDictionaries[i].ID == item.ID) || (item.ID == 0 && data.DataDictionaries[i].Type == item.Type && data.DataDictionaries[i].Code == item.Code) {
					item.ID = data.DataDictionaries[i].ID
					data.DataDictionaries[i] = item
					saved = item
					addAudit(data, session.User.Username, "update", "data_dictionary", item.ID, item.Type+"/"+item.Code, clientIP(r))
					return nil
				}
			}
			item.ID = nextID(data, "dict")
			data.DataDictionaries = append(data.DataDictionaries, item)
			saved = item
			addAudit(data, session.User.Username, "create", "data_dictionary", item.ID, item.Type+"/"+item.Code, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "system.dictionary.saved")
		return
	}
	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Status string `json:"status"`
		}
		_ = readJSON(r, &req)
		var updated DataDictionary
		err := a.store.Mutate(func(data *AppData) error {
			status, err := normalizeConfigStatus(req.Status)
			if err != nil {
				return err
			}
			for i := range data.DataDictionaries {
				if data.DataDictionaries[i].ID != id {
					continue
				}
				data.DataDictionaries[i].Status = status
				updated = data.DataDictionaries[i]
				addAudit(data, session.User.Username, "status", "data_dictionary", id, updated.Type+"/"+updated.Code+"/"+status, clientIP(r))
				return nil
			}
			return fmt.Errorf("数据字典不存在")
		})
		a.respondMutation(w, err, updated, "system.dictionary.updated")
		return
	}
	writeError(w, http.StatusNotFound, "unknown dictionary route")
}

func normalizeApprovalFlow(item *ApprovalFlow) error {
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.Resource = strings.TrimSpace(item.Resource)
	item.Status = strings.TrimSpace(item.Status)
	if item.Code == "" || item.Name == "" || item.Resource == "" {
		return fmt.Errorf("审批流模板必须包含编码、名称和业务资源")
	}
	if len(item.Steps) == 0 {
		return fmt.Errorf("审批流模板至少需要一个步骤")
	}
	status, err := normalizeConfigStatus(item.Status)
	if err != nil {
		return err
	}
	item.Status = status
	seen := map[int]bool{}
	for i := range item.Steps {
		item.Steps[i].RoleCode = strings.TrimSpace(item.Steps[i].RoleCode)
		item.Steps[i].Action = fallback(strings.TrimSpace(item.Steps[i].Action), "approve")
		if item.Steps[i].RoleCode == "" {
			return fmt.Errorf("审批步骤必须包含角色")
		}
		if item.Steps[i].Seq <= 0 {
			item.Steps[i].Seq = i + 1
		}
		if seen[item.Steps[i].Seq] {
			return fmt.Errorf("审批步骤序号重复")
		}
		seen[item.Steps[i].Seq] = true
	}
	return nil
}

func normalizeDataDictionary(item *DataDictionary, existing []DataDictionary) error {
	item.Type = strings.TrimSpace(item.Type)
	item.Code = strings.TrimSpace(item.Code)
	item.Label = strings.TrimSpace(item.Label)
	item.Status = strings.TrimSpace(item.Status)
	if item.Type == "" || item.Code == "" || item.Label == "" {
		return fmt.Errorf("数据字典必须包含类型、编码和名称")
	}
	status, err := normalizeConfigStatus(item.Status)
	if err != nil {
		return err
	}
	item.Status = status
	if item.Sort <= 0 {
		for _, existingItem := range existing {
			if existingItem.Type == item.Type && existingItem.Sort >= item.Sort {
				item.Sort = existingItem.Sort + 1
			}
		}
		if item.Sort <= 0 {
			item.Sort = 1
		}
	}
	return nil
}

func normalizeConfigStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active", nil
	}
	switch status {
	case "active", "disabled", "draft":
		return status, nil
	default:
		return "", fmt.Errorf("状态只能是 active、disabled 或 draft")
	}
}
