package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (a *App) approvals(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, visibleApprovalTasks(data, session.User))
		return
	}
	if len(parts) == 2 && parts[1] == "act" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.actApproval(w, r, session, id)
		return
	}
	writeError(w, http.StatusNotFound, "unknown approval route")
}

func (a *App) actApproval(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Action  string `json:"action"`
		Comment string `json:"comment"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid approval action")
		return
	}
	req.Action = strings.TrimSpace(req.Action)
	if req.Action != "approve" && req.Action != "reject" {
		writeError(w, http.StatusBadRequest, "审批动作只能是 approve 或 reject")
		return
	}

	var updated ApprovalTask
	var auditAction string
	var resultFailure WorkflowInstance
	var resultErr error
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.ApprovalTasks {
			if data.ApprovalTasks[i].ID != id {
				continue
			}
			if data.ApprovalTasks[i].Status != "pending" {
				return fmt.Errorf("审批任务已结束")
			}
			if !userCanActApproval(*data, session.User, data.ApprovalTasks[i]) {
				return fmt.Errorf("无权审批当前步骤")
			}
			task := &data.ApprovalTasks[i]
			if err := ensureWorkflowForApprovalTask(data, task); err != nil {
				return err
			}
			instance, err := actWorkflowInstance(data, task.WorkflowInstanceID, workflowActionRequest{
				Action:        req.Action,
				Actor:         session.User.Username,
				ActorRole:     session.User.RoleCode,
				Comment:       req.Comment,
				AllowOverride: canAccess(*data, session.User, "*"),
			})
			if err != nil {
				return err
			}
			syncApprovalTaskFromWorkflow(task, instance)
			switch task.Status {
			case "rejected":
				auditAction = "reject"
				if err := applyApprovalResult(data, *task); err != nil {
					resultFailure = workflowInstanceFromApprovalTask(*task)
					resultErr = err
					return err
				}
			case "approved":
				auditAction = "approve"
				if err := applyApprovalResult(data, *task); err != nil {
					resultFailure = workflowInstanceFromApprovalTask(*task)
					resultErr = err
					return err
				}
			default:
				auditAction = "approve_step"
			}
			updated = *task
			addAudit(data, session.User.Username, auditAction, "approval_task", task.ID, task.TaskNo, clientIP(r))
			return nil
		}
		return fmt.Errorf("审批任务不存在")
	})
	a.recordWorkflowResultFailure(resultFailure, resultErr)
	a.respondMutation(w, err, updated, "approval.task.updated")
}

func submitApprovalTask(data *AppData, flowCode, resource string, resourceID int64, resourceNo, title, applicant, reason string) (ApprovalTask, error) {
	for i := range data.ApprovalTasks {
		if data.ApprovalTasks[i].Resource == resource && data.ApprovalTasks[i].ResourceID == resourceID && data.ApprovalTasks[i].Status == "pending" {
			if err := ensureWorkflowForApprovalTask(data, &data.ApprovalTasks[i]); err != nil {
				return ApprovalTask{}, err
			}
			return data.ApprovalTasks[i], nil
		}
	}
	flow, ok := activeApprovalFlow(*data, flowCode)
	if !ok || len(flow.Steps) == 0 {
		return ApprovalTask{}, fmt.Errorf("审批流未配置: %s", flowCode)
	}
	if _, err := upsertApprovalWorkflowDefinition(data, flow); err != nil {
		return ApprovalTask{}, err
	}
	instance, err := startWorkflowInstance(data, workflowStartRequest{
		DefinitionCode: flowCode,
		Resource:       resource,
		ResourceID:     resourceID,
		ResourceNo:     resourceNo,
		Title:          title,
		Applicant:      applicant,
		Reason:         reason,
	})
	if err != nil {
		return ApprovalTask{}, err
	}
	taskID := nextID(data, "approvalTask")
	task := ApprovalTask{
		ID:      taskID,
		TaskNo:  number("APV", taskID),
		Actions: []ApprovalTaskAction{},
	}
	syncApprovalTaskFromWorkflow(&task, instance)
	data.ApprovalTasks = append(data.ApprovalTasks, task)
	return task, nil
}

func visibleApprovalTasks(data AppData, user User) []ApprovalTask {
	if canAccess(data, user, "*") {
		return data.ApprovalTasks
	}
	out := []ApprovalTask{}
	for _, task := range data.ApprovalTasks {
		if task.Applicant == user.Username || task.CurrentRole == user.RoleCode {
			out = append(out, task)
			continue
		}
		if task.Resource == "sales_order" {
			if _, ok := findOrder(data, task.ResourceID); ok {
				out = append(out, task)
			}
		}
		if task.Resource == "inventory_transfer" {
			if _, ok := findInventoryTransfer(data, task.ResourceID); ok {
				out = append(out, task)
			}
		}
		if task.Resource == "contract" {
			if _, ok := findContract(data, task.ResourceID); ok {
				out = append(out, task)
			}
		}
	}
	return out
}

func userCanActApproval(data AppData, user User, task ApprovalTask) bool {
	if canAccess(data, user, "*") {
		return true
	}
	return task.CurrentRole == user.RoleCode
}

func activeApprovalFlow(data AppData, code string) (ApprovalFlow, bool) {
	for _, flow := range data.ApprovalFlows {
		if flow.Code == code && flow.Status == "active" {
			return flow, true
		}
	}
	return ApprovalFlow{}, false
}

func nextApprovalStep(data AppData, task ApprovalTask) (ApprovalStep, bool) {
	flow, ok := activeApprovalFlow(data, task.FlowCode)
	if !ok {
		return ApprovalStep{}, false
	}
	for _, step := range flow.Steps {
		if step.Seq > task.CurrentStep {
			return step, true
		}
	}
	return ApprovalStep{}, false
}

func findInventoryTransfer(data AppData, id int64) (InventoryTransfer, bool) {
	for _, item := range data.InventoryTransfers {
		if item.ID == id {
			return item, true
		}
	}
	return InventoryTransfer{}, false
}
