package appliance

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
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
					if _, err := upsertApprovalWorkflowDefinition(data, item); err != nil {
						return err
					}
					saved = item
					addAudit(data, session.User.Username, "update", "approval_flow", item.ID, item.Code, clientIP(r))
					return nil
				}
			}
			item.ID = nextID(data, "approvalFlow")
			data.ApprovalFlows = append(data.ApprovalFlows, item)
			if _, err := upsertApprovalWorkflowDefinition(data, item); err != nil {
				return err
			}
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
				if _, err := upsertApprovalWorkflowDefinition(data, updated); err != nil {
					return err
				}
				addAudit(data, session.User.Username, "status", "approval_flow", id, updated.Code+"/"+status, clientIP(r))
				return nil
			}
			return fmt.Errorf("审批流模板不存在")
		})
		a.respondMutation(w, err, updated, "system.approval_flow.updated")
		return
	}
	if len(parts) == 1 && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var deleted ApprovalFlow
		err := a.store.Mutate(func(data *AppData) error {
			for i, item := range data.ApprovalFlows {
				if item.ID != id {
					continue
				}
				if item.Status == "active" {
					return fmt.Errorf("启用中的审批流不能删除，请先禁用或转草稿")
				}
				if err := deleteApprovalWorkflowDefinitions(data, item); err != nil {
					return err
				}
				deleted = item
				data.ApprovalFlows = append(data.ApprovalFlows[:i], data.ApprovalFlows[i+1:]...)
				addAudit(data, session.User.Username, "delete", "approval_flow", id, item.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("审批流模板不存在")
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("system.approval_flow.deleted", deleted)
		writeJSON(w, http.StatusOK, deleted)
		return
	}
	writeError(w, http.StatusNotFound, "unknown approval flow route")
}

func deleteApprovalWorkflowDefinitions(data *AppData, flow ApprovalFlow) error {
	code := strings.TrimSpace(flow.Code)
	if code == "" {
		return nil
	}
	for _, definition := range data.WorkflowDefinitions {
		if definition.Category != workflowCategoryApproval || strings.TrimSpace(definition.Code) != code {
			continue
		}
		if pendingWorkflowInstanceCount(*data, definition) > 0 {
			return fmt.Errorf("审批流存在待处理工作流实例，不能删除")
		}
	}
	next := make([]WorkflowDefinition, 0, len(data.WorkflowDefinitions))
	for _, definition := range data.WorkflowDefinitions {
		if definition.Category == workflowCategoryApproval && strings.TrimSpace(definition.Code) == code {
			continue
		}
		next = append(next, definition)
	}
	data.WorkflowDefinitions = next
	return nil
}

type workflowEventResolutionRequest struct {
	Resolution string `json:"resolution"`
}

type workflowEventPublishRequest struct {
	EventType  string            `json:"eventType"`
	Source     string            `json:"source"`
	EventKey   string            `json:"eventKey"`
	Actor      string            `json:"actor"`
	Resource   string            `json:"resource"`
	ResourceID int64             `json:"resourceId"`
	ResourceNo string            `json:"resourceNo"`
	Title      string            `json:"title"`
	Reason     string            `json:"reason"`
	Variables  map[string]string `json:"variables"`
}

type workflowInstanceCancelRequest struct {
	Reason string `json:"reason"`
}

type workflowTaskActionRequest struct {
	Action  string `json:"action"`
	Comment string `json:"comment"`
}

type workflowTaskReassignRequest struct {
	RoleCode string `json:"roleCode"`
	Reason   string `json:"reason"`
}

type workflowOutboxClaimRequest struct {
	Consumer string `json:"consumer"`
}

type workflowOutboxFailRequest struct {
	Error             string `json:"error"`
	RetryAfterMinutes int    `json:"retryAfterMinutes"`
}

func (a *App) systemWorkflows(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "definitions" && r.Method == http.MethodPost {
		var item WorkflowDefinition
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow definition")
			return
		}
		var saved WorkflowDefinition
		err := a.store.Mutate(func(data *AppData) error {
			next, err := saveWorkflowDefinition(data, item)
			if err != nil {
				return err
			}
			saved = next
			addAudit(data, session.User.Username, "save", "workflow_definition", saved.ID, saved.Code, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "system.workflow_definition.saved")
		return
	}
	if len(parts) == 3 && parts[0] == "definitions" && parts[2] == "publish" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var item WorkflowDefinition
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow definition")
			return
		}
		var published WorkflowDefinition
		err := a.store.Mutate(func(data *AppData) error {
			next, err := publishWorkflowDefinitionVersion(data, id, item)
			if err != nil {
				return err
			}
			published = next
			addAudit(data, session.User.Username, "publish", "workflow_definition", published.ID, fmt.Sprintf("%s/v%d", published.Code, published.Version), clientIP(r))
			return nil
		})
		a.respondMutation(w, err, published, "system.workflow_definition.published")
		return
	}
	if len(parts) == 3 && parts[0] == "definitions" && parts[2] == "rollback" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var rolledBack WorkflowDefinition
		err := a.store.Mutate(func(data *AppData) error {
			next, err := rollbackWorkflowDefinitionVersion(data, id)
			if err != nil {
				return err
			}
			rolledBack = next
			addAudit(data, session.User.Username, "rollback", "workflow_definition", rolledBack.ID, fmt.Sprintf("%s/v%d", rolledBack.Code, rolledBack.Version), clientIP(r))
			return nil
		})
		a.respondMutation(w, err, rolledBack, "system.workflow_definition.rolled_back")
		return
	}
	if len(parts) == 1 && parts[0] == "subscriptions" && r.Method == http.MethodPost {
		var item WorkflowSubscription
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow subscription")
			return
		}
		var saved WorkflowSubscription
		err := a.store.Mutate(func(data *AppData) error {
			next, err := saveWorkflowSubscription(data, item)
			if err != nil {
				return err
			}
			saved = next
			addAudit(data, session.User.Username, "save", "workflow_subscription", saved.ID, saved.Code, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "system.workflow_subscription.saved")
		return
	}
	if len(parts) == 3 && parts[0] == "subscriptions" && parts[2] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req struct {
			Status string `json:"status"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow subscription status")
			return
		}
		var updated WorkflowSubscription
		err := a.store.Mutate(func(data *AppData) error {
			next, err := setWorkflowSubscriptionStatus(data, id, req.Status)
			if err != nil {
				return err
			}
			updated = next
			addAudit(data, session.User.Username, "status", "workflow_subscription", updated.ID, updated.Status, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, updated, "system.workflow_subscription.updated")
		return
	}
	if len(parts) == 2 && parts[0] == "subscriptions" && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var deleted WorkflowSubscription
		err := a.store.Mutate(func(data *AppData) error {
			for i, item := range data.WorkflowSubscriptions {
				if item.ID != id {
					continue
				}
				if item.Status == "active" || item.Status == "enabled" {
					return fmt.Errorf("启用中的工作流订阅不能删除，请先停用")
				}
				now := nowString()
				outboxIDs := map[int64]bool{}
				for j := range data.WorkflowDeliveries {
					delivery := &data.WorkflowDeliveries[j]
					if delivery.SubscriptionID != id {
						continue
					}
					if delivery.Status == "processing" {
						return fmt.Errorf("工作流订阅存在执行中的投递，请稍后再删除")
					}
					if delivery.Status == "succeeded" || delivery.Status == "dead" {
						continue
					}
					delivery.Status = "dead"
					delivery.LastError = "订阅已删除"
					delivery.NextAttemptAt = ""
					delivery.CompletedAt = now
					delivery.UpdatedAt = now
					outboxIDs[delivery.OutboxID] = true
				}
				deleted = item
				data.WorkflowSubscriptions = append(data.WorkflowSubscriptions[:i], data.WorkflowSubscriptions[i+1:]...)
				for outboxID := range outboxIDs {
					syncWorkflowOutboxDeliveryStatus(data, outboxID)
				}
				addAudit(data, session.User.Username, "delete", "workflow_subscription", id, item.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("工作流订阅不存在")
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("system.workflow_subscription.deleted", deleted)
		writeJSON(w, http.StatusOK, deleted)
		return
	}
	if len(parts) == 3 && parts[0] == "tasks" && parts[2] == "act" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req workflowTaskActionRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow task action")
			return
		}
		req.Action = strings.TrimSpace(req.Action)
		if req.Action != "approve" && req.Action != "reject" {
			writeError(w, http.StatusBadRequest, "工作流动作只能是 approve 或 reject")
			return
		}
		var updated WorkflowInstance
		var auditAction string
		var resultFailure WorkflowInstance
		var resultErr error
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.WorkflowTasks {
				if data.WorkflowTasks[i].ID != id {
					continue
				}
				task := data.WorkflowTasks[i]
				if task.Status != "pending" {
					return fmt.Errorf("工作流任务已结束")
				}
				instance, err := actWorkflowInstance(data, task.InstanceID, workflowActionRequest{
					Action:        req.Action,
					Actor:         session.User.Username,
					ActorRole:     session.User.RoleCode,
					Comment:       req.Comment,
					AllowOverride: canAccess(*data, session.User, "*"),
				})
				if err != nil {
					return err
				}
				updated = instance
				auditAction = "workflow_" + req.Action
				if instance.Category == workflowCategoryApproval {
					syncApprovalTaskForWorkflowInstance(data, instance)
				}
				if workflowResultStatus(instance.Status) {
					if err := applyWorkflowResult(data, instance); err != nil {
						resultFailure = instance
						resultErr = err
						return err
					}
				}
				addAudit(data, session.User.Username, auditAction, "workflow_task", task.ID, task.TaskNo, clientIP(r))
				return nil
			}
			return fmt.Errorf("工作流任务不存在")
		})
		a.recordWorkflowResultFailure(resultFailure, resultErr)
		a.respondMutation(w, err, updated, "system.workflow_task.updated")
		return
	}
	if len(parts) == 3 && parts[0] == "tasks" && parts[2] == "reassign" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req workflowTaskReassignRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow task reassign")
			return
		}
		var updated WorkflowInstance
		err := a.store.Mutate(func(data *AppData) error {
			if !canAccess(*data, session.User, "*") {
				return fmt.Errorf("无权改派工作流任务")
			}
			instance, err := reassignWorkflowTask(data, id, req.RoleCode, session.User.Username, req.Reason)
			if err != nil {
				return err
			}
			updated = instance
			if instance.Category == workflowCategoryApproval {
				syncApprovalTaskForWorkflowInstance(data, instance)
			}
			addAudit(data, session.User.Username, "reassign", "workflow_task", id, req.RoleCode, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, updated, "system.workflow_task.reassigned")
		return
	}
	if len(parts) == 3 && parts[0] == "tasks" && parts[2] == "escalate" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req workflowTaskReassignRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow task escalation")
			return
		}
		var updated WorkflowInstance
		err := a.store.Mutate(func(data *AppData) error {
			if !canAccess(*data, session.User, "*") {
				return fmt.Errorf("无权升级工作流任务")
			}
			instance, err := escalateWorkflowTask(data, id, req.RoleCode, session.User.Username, req.Reason)
			if err != nil {
				return err
			}
			updated = instance
			if instance.Category == workflowCategoryApproval {
				syncApprovalTaskForWorkflowInstance(data, instance)
			}
			addAudit(data, session.User.Username, "escalate", "workflow_task", id, req.RoleCode, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, updated, "system.workflow_task.escalated")
		return
	}
	if len(parts) == 3 && parts[0] == "outbox" && (parts[2] == "ack" || parts[2] == "reset" || parts[2] == "claim" || parts[2] == "fail") && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var claimReq workflowOutboxClaimRequest
		var failReq workflowOutboxFailRequest
		if parts[2] == "claim" {
			if err := readJSON(r, &claimReq); err != nil {
				writeError(w, http.StatusBadRequest, "invalid workflow outbox claim")
				return
			}
		}
		if parts[2] == "fail" {
			if err := readJSON(r, &failReq); err != nil {
				writeError(w, http.StatusBadRequest, "invalid workflow outbox failure")
				return
			}
		}
		var updated WorkflowOutbox
		err := a.store.Mutate(func(data *AppData) error {
			var item WorkflowOutbox
			var err error
			switch parts[2] {
			case "ack":
				item, err = acknowledgeWorkflowOutbox(data, id, session.User.Username)
			case "reset":
				item, err = resetWorkflowOutbox(data, id, session.User.Username)
			case "claim":
				item, err = claimWorkflowOutbox(data, id, session.User.Username, claimReq.Consumer)
			case "fail":
				item, err = failWorkflowOutbox(data, id, session.User.Username, failReq.Error, failReq.RetryAfterMinutes)
			}
			if err != nil {
				return err
			}
			updated = item
			addAudit(data, session.User.Username, parts[2], "workflow_outbox", updated.ID, updated.OutboxNo, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, updated, "system.workflow_outbox.updated")
		return
	}
	if len(parts) == 2 && parts[0] == "deliveries" && parts[1] == "dispatch-due" && r.Method == http.MethodPost {
		limit := workflowDeliveryDispatchLimit(r)
		recovered := a.recoverStaleWorkflowDeliveries(session.User.Username, clientIP(r))
		ids := dueWorkflowDeliveryIDs(a.mustSnapshot(), limit, time.Now())
		batch := a.dispatchWorkflowDeliveries(ids, session.User.Username, clientIP(r))
		batch.Recovered = len(recovered)
		writeJSON(w, http.StatusCreated, batch)
		return
	}
	if len(parts) == 2 && parts[0] == "automation" && parts[1] == "run" && r.Method == http.MethodPost {
		run := WorkflowAutomationRun{
			Escalations: []WorkflowTaskEscalationResult{},
		}
		deliveryLimit := workflowQueryLimit(r, "deliveryLimit", 20, 100)
		escalationLimit := workflowQueryLimit(r, "escalationLimit", 20, 100)
		recovered := a.recoverStaleWorkflowDeliveries(session.User.Username, clientIP(r))
		run.Deliveries = a.dispatchWorkflowDeliveries(dueWorkflowDeliveryIDs(a.mustSnapshot(), deliveryLimit, time.Now()), session.User.Username, clientIP(r))
		run.Deliveries.Recovered = len(recovered)
		targets := dueWorkflowTaskEscalationTargets(a.mustSnapshot(), escalationLimit, time.Now())
		for _, target := range targets {
			result := WorkflowTaskEscalationResult{
				TaskID:   target.TaskID,
				TaskNo:   target.TaskNo,
				FromRole: target.FromRole,
				ToRole:   target.ToRole,
				Status:   "skipped",
			}
			var updated WorkflowInstance
			err := a.store.Mutate(func(data *AppData) error {
				instance, err := escalateWorkflowTask(data, target.TaskID, target.ToRole, session.User.Username, "自动 SLA 超时升级")
				if err != nil {
					return err
				}
				updated = instance
				if instance.Category == workflowCategoryApproval {
					syncApprovalTaskForWorkflowInstance(data, instance)
				}
				addAudit(data, session.User.Username, "auto_escalate", "workflow_task", target.TaskID, target.ToRole, clientIP(r))
				return nil
			})
			if err != nil {
				result.Error = err.Error()
				run.Skipped++
				run.Escalations = append(run.Escalations, result)
				continue
			}
			result.InstanceID = updated.ID
			result.InstanceNo = updated.InstanceNo
			result.Status = "escalated"
			run.Escalated++
			run.Escalations = append(run.Escalations, result)
		}
		writeJSON(w, http.StatusCreated, run)
		return
	}
	if len(parts) == 3 && parts[0] == "deliveries" && parts[2] == "dispatch" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		updated, err := a.dispatchWorkflowDelivery(id, session.User.Username, clientIP(r))
		a.respondMutation(w, err, updated, "system.workflow_delivery.dispatched")
		return
	}
	if len(parts) == 3 && parts[0] == "deliveries" && parts[2] == "reset" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var updated WorkflowDelivery
		err := a.store.Mutate(func(data *AppData) error {
			next, err := resetWorkflowDelivery(data, id, session.User.Username)
			if err != nil {
				return err
			}
			updated = next
			addAudit(data, session.User.Username, "reset", "workflow_delivery", updated.ID, updated.DeliveryNo, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, updated, "system.workflow_delivery.reset")
		return
	}
	if len(parts) == 3 && parts[0] == "instances" && parts[2] == "cancel" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req workflowInstanceCancelRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow cancel request")
			return
		}
		var cancelled WorkflowInstance
		var resultFailure WorkflowInstance
		var resultErr error
		err := a.store.Mutate(func(data *AppData) error {
			instance, err := cancelWorkflowInstance(data, id, session.User.Username, req.Reason)
			if err != nil {
				return err
			}
			cancelled = instance
			if instance.Category == workflowCategoryApproval {
				syncApprovalTaskForWorkflowInstance(data, instance)
			}
			if err := applyWorkflowResult(data, instance); err != nil {
				resultFailure = instance
				resultErr = err
				return err
			}
			addAudit(data, session.User.Username, "cancel", "workflow_instance", cancelled.ID, cancelled.InstanceNo, clientIP(r))
			return nil
		})
		a.recordWorkflowResultFailure(resultFailure, resultErr)
		a.respondMutation(w, err, cancelled, "system.workflow_instance.cancelled")
		return
	}
	if len(parts) == 2 && parts[0] == "events" && parts[1] == "preview" && r.Method == http.MethodPost {
		var req workflowEventPublishRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow event")
			return
		}
		preview, err := previewWorkflowEvent(a.mustSnapshot(), workflowEventRequest{
			EventType:  req.EventType,
			Source:     req.Source,
			EventKey:   req.EventKey,
			Resource:   req.Resource,
			ResourceID: req.ResourceID,
			ResourceNo: req.ResourceNo,
			Title:      req.Title,
			Actor:      fallback(strings.TrimSpace(req.Actor), session.User.Username),
			Reason:     req.Reason,
			Variables:  req.Variables,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, preview)
		return
	}
	if len(parts) == 1 && parts[0] == "events" && r.Method == http.MethodPost {
		var req workflowEventPublishRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow event")
			return
		}
		var published WorkflowEvent
		err := a.store.Mutate(func(data *AppData) error {
			event, _, err := publishWorkflowEvent(data, workflowEventRequest{
				EventType:  req.EventType,
				Source:     req.Source,
				EventKey:   req.EventKey,
				Resource:   req.Resource,
				ResourceID: req.ResourceID,
				ResourceNo: req.ResourceNo,
				Title:      req.Title,
				Actor:      fallback(strings.TrimSpace(req.Actor), session.User.Username),
				Reason:     req.Reason,
				Variables:  req.Variables,
			})
			if err != nil {
				return err
			}
			published = event
			addAudit(data, session.User.Username, "publish", "workflow_event", published.ID, published.EventNo, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, published, "system.workflow_event.published")
		return
	}
	if len(parts) == 3 && parts[0] == "events" && parts[2] == "replay" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var replayed WorkflowEvent
		err := a.store.Mutate(func(data *AppData) error {
			event, _, err := replayWorkflowEvent(data, id, session.User.Username)
			if err != nil {
				return err
			}
			replayed = event
			addAudit(data, session.User.Username, "replay", "workflow_event", replayed.ID, replayed.EventNo, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, replayed, "system.workflow_event.replayed")
		return
	}
	if len(parts) == 3 && parts[0] == "events" && parts[2] == "resolve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req workflowEventResolutionRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid workflow event resolution")
			return
		}
		var resolved WorkflowEvent
		err := a.store.Mutate(func(data *AppData) error {
			event, err := resolveWorkflowEvent(data, id, session.User.Username, req.Resolution, 0)
			if err != nil {
				return err
			}
			resolved = event
			addAudit(data, session.User.Username, "resolve", "workflow_event", resolved.ID, resolved.EventNo, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, resolved, "system.workflow_event.resolved")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "unknown workflow route")
		return
	}
	data := a.mustSnapshot()
	if len(parts) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"definitions":   data.WorkflowDefinitions,
			"instances":     data.WorkflowInstances,
			"tasks":         data.WorkflowTasks,
			"inbox":         workflowInboxItems(data, session.User),
			"events":        data.WorkflowEvents,
			"logs":          data.WorkflowLogs,
			"outbox":        data.WorkflowOutbox,
			"subscriptions": data.WorkflowSubscriptions,
			"deliveries":    data.WorkflowDeliveries,
			"catalog":       workflowCatalog(),
		})
		return
	}
	if len(parts) == 1 && parts[0] == "catalog" {
		writeJSON(w, http.StatusOK, workflowCatalog())
		return
	}
	if len(parts) == 1 && parts[0] == "definitions" {
		writeJSON(w, http.StatusOK, data.WorkflowDefinitions)
		return
	}
	if len(parts) == 1 && parts[0] == "instances" {
		writeJSON(w, http.StatusOK, workflowInstancesForQuery(data.WorkflowInstances, r))
		return
	}
	if len(parts) == 1 && parts[0] == "tasks" {
		writeJSON(w, http.StatusOK, workflowTasksForQuery(data.WorkflowTasks, r))
		return
	}
	if len(parts) == 1 && parts[0] == "inbox" {
		writeJSON(w, http.StatusOK, workflowInboxItems(data, session.User))
		return
	}
	if len(parts) == 1 && parts[0] == "events" {
		writeJSON(w, http.StatusOK, workflowEventsForQuery(data.WorkflowEvents, r))
		return
	}
	if len(parts) == 1 && parts[0] == "logs" {
		writeJSON(w, http.StatusOK, workflowLogsForQuery(data.WorkflowLogs, r))
		return
	}
	if len(parts) == 1 && parts[0] == "outbox" {
		writeJSON(w, http.StatusOK, workflowOutboxForQuery(data.WorkflowOutbox, r))
		return
	}
	if len(parts) == 1 && parts[0] == "subscriptions" {
		writeJSON(w, http.StatusOK, data.WorkflowSubscriptions)
		return
	}
	if len(parts) == 1 && parts[0] == "deliveries" {
		writeJSON(w, http.StatusOK, workflowDeliveriesForQuery(data.WorkflowDeliveries, r))
		return
	}
	writeError(w, http.StatusNotFound, "unknown workflow route")
}

func saveWorkflowDefinition(data *AppData, item WorkflowDefinition) (WorkflowDefinition, error) {
	if err := normalizeWorkflowDefinition(&item); err != nil {
		return WorkflowDefinition{}, err
	}
	for i := range data.WorkflowDefinitions {
		if item.ID > 0 && data.WorkflowDefinitions[i].ID == item.ID {
			item.ID = data.WorkflowDefinitions[i].ID
			if err := validateWorkflowDefinitionChange(*data, data.WorkflowDefinitions[i], item); err != nil {
				return WorkflowDefinition{}, err
			}
			data.WorkflowDefinitions[i] = item
			syncApprovalFlowFromCurrentWorkflowDefinition(data, item)
			ensureCounterAtLeast(data, "workflowDefinition", item.ID)
			return item, nil
		}
	}
	if item.ID == 0 {
		if i := workflowDefinitionIndexForCode(*data, item.Category, item.Code); i >= 0 {
			item.ID = data.WorkflowDefinitions[i].ID
			if err := validateWorkflowDefinitionChange(*data, data.WorkflowDefinitions[i], item); err != nil {
				return WorkflowDefinition{}, err
			}
			data.WorkflowDefinitions[i] = item
			syncApprovalFlowFromCurrentWorkflowDefinition(data, item)
			ensureCounterAtLeast(data, "workflowDefinition", item.ID)
			return item, nil
		}
	}
	item.ID = nextID(data, "workflowDefinition")
	data.WorkflowDefinitions = append(data.WorkflowDefinitions, item)
	syncApprovalFlowFromCurrentWorkflowDefinition(data, item)
	return item, nil
}

func workflowDeliveryDispatchLimit(r *http.Request) int {
	return workflowQueryLimit(r, "limit", 20, 100)
}

func workflowInstancesForQuery(items []WorkflowInstance, r *http.Request) []WorkflowInstance {
	query := r.URL.Query()
	status := strings.TrimSpace(query.Get("status"))
	category := strings.TrimSpace(query.Get("category"))
	resource := strings.TrimSpace(query.Get("resource"))
	resourceNo := strings.TrimSpace(query.Get("resourceNo"))
	definitionCode := strings.TrimSpace(query.Get("definitionCode"))
	instanceNo := strings.TrimSpace(query.Get("instanceNo"))
	currentRole := strings.TrimSpace(query.Get("currentRole"))
	resourceID := workflowInt64QueryValue(query.Get("resourceId"))
	definitionID := workflowInt64QueryValue(query.Get("definitionId"))
	triggerEventID := workflowInt64QueryValue(query.Get("triggerEventId"))
	if status == "" && category == "" && resource == "" && resourceNo == "" && definitionCode == "" && instanceNo == "" && currentRole == "" && resourceID == 0 && definitionID == 0 && triggerEventID == 0 {
		return items
	}
	filtered := []WorkflowInstance{}
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if category != "" && item.Category != category {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		if resourceNo != "" && item.ResourceNo != resourceNo {
			continue
		}
		if definitionCode != "" && item.DefinitionCode != definitionCode {
			continue
		}
		if instanceNo != "" && item.InstanceNo != instanceNo {
			continue
		}
		if currentRole != "" && item.CurrentRole != currentRole {
			continue
		}
		if resourceID != 0 && item.ResourceID != resourceID {
			continue
		}
		if definitionID != 0 && item.DefinitionID != definitionID {
			continue
		}
		if triggerEventID != 0 && item.TriggerEventID != triggerEventID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func workflowTasksForQuery(items []WorkflowTask, r *http.Request) []WorkflowTask {
	query := r.URL.Query()
	status := strings.TrimSpace(query.Get("status"))
	resource := strings.TrimSpace(query.Get("resource"))
	definitionCode := strings.TrimSpace(query.Get("definitionCode"))
	taskNo := strings.TrimSpace(query.Get("taskNo"))
	roleCode := strings.TrimSpace(query.Get("roleCode"))
	action := strings.TrimSpace(query.Get("action"))
	resourceID := workflowInt64QueryValue(query.Get("resourceId"))
	instanceID := workflowInt64QueryValue(query.Get("instanceId"))
	overdueOnly := workflowTruthy(query.Get("overdue"))
	if status == "" && resource == "" && definitionCode == "" && taskNo == "" && roleCode == "" && action == "" && resourceID == 0 && instanceID == 0 && !overdueOnly {
		return items
	}
	filtered := []WorkflowTask{}
	now := time.Now()
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		if definitionCode != "" && item.DefinitionCode != definitionCode {
			continue
		}
		if taskNo != "" && item.TaskNo != taskNo {
			continue
		}
		if roleCode != "" && item.RoleCode != roleCode {
			continue
		}
		if action != "" && item.Action != action {
			continue
		}
		if resourceID != 0 && item.ResourceID != resourceID {
			continue
		}
		if instanceID != 0 && item.InstanceID != instanceID {
			continue
		}
		if overdueOnly && !workflowTaskOverdue(item, now) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func workflowLogsForQuery(items []WorkflowLog, r *http.Request) []WorkflowLog {
	query := r.URL.Query()
	action := strings.TrimSpace(query.Get("action"))
	status := strings.TrimSpace(query.Get("status"))
	resource := strings.TrimSpace(query.Get("resource"))
	definitionCode := strings.TrimSpace(query.Get("definitionCode"))
	instanceNo := strings.TrimSpace(query.Get("instanceNo"))
	taskNo := strings.TrimSpace(query.Get("taskNo"))
	actor := strings.TrimSpace(query.Get("actor"))
	resourceID := workflowInt64QueryValue(query.Get("resourceId"))
	instanceID := workflowInt64QueryValue(query.Get("instanceId"))
	triggerEventID := workflowInt64QueryValue(query.Get("triggerEventId"))
	taskID := workflowInt64QueryValue(query.Get("taskId"))
	if action == "" && status == "" && resource == "" && definitionCode == "" && instanceNo == "" && taskNo == "" && actor == "" && resourceID == 0 && instanceID == 0 && triggerEventID == 0 && taskID == 0 {
		return items
	}
	filtered := []WorkflowLog{}
	for _, item := range items {
		if action != "" && item.Action != action {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		if definitionCode != "" && item.DefinitionCode != definitionCode {
			continue
		}
		if instanceNo != "" && item.InstanceNo != instanceNo {
			continue
		}
		if taskNo != "" && item.TaskNo != taskNo {
			continue
		}
		if actor != "" && item.Actor != actor {
			continue
		}
		if resourceID != 0 && item.ResourceID != resourceID {
			continue
		}
		if instanceID != 0 && item.InstanceID != instanceID {
			continue
		}
		if triggerEventID != 0 && item.TriggerEventID != triggerEventID {
			continue
		}
		if taskID != 0 && item.TaskID != taskID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func workflowEventsForQuery(items []WorkflowEvent, r *http.Request) []WorkflowEvent {
	query := r.URL.Query()
	status := strings.TrimSpace(query.Get("status"))
	source := strings.TrimSpace(query.Get("source"))
	eventType := strings.TrimSpace(query.Get("eventType"))
	resource := strings.TrimSpace(query.Get("resource"))
	eventKey := strings.TrimSpace(query.Get("eventKey"))
	resourceID := int64(0)
	if raw := strings.TrimSpace(query.Get("resourceId")); raw != "" {
		resourceID, _ = strconv.ParseInt(raw, 10, 64)
	}
	replayOfID := int64(0)
	if raw := strings.TrimSpace(query.Get("replayOfId")); raw != "" {
		replayOfID, _ = strconv.ParseInt(raw, 10, 64)
	}
	recoveredByEventID := int64(0)
	if raw := strings.TrimSpace(query.Get("recoveredByEventId")); raw != "" {
		recoveredByEventID, _ = strconv.ParseInt(raw, 10, 64)
	}
	recoveryOnly := workflowTruthy(query.Get("recovery"))
	if status == "" && source == "" && eventType == "" && resource == "" && eventKey == "" && resourceID == 0 && replayOfID == 0 && recoveredByEventID == 0 && !recoveryOnly {
		return items
	}
	filtered := []WorkflowEvent{}
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if source != "" && item.Source != source {
			continue
		}
		if eventType != "" && item.EventType != eventType {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		if eventKey != "" && item.EventKey != eventKey {
			continue
		}
		if resourceID != 0 && item.ResourceID != resourceID {
			continue
		}
		if replayOfID != 0 && item.ReplayOfID != replayOfID {
			continue
		}
		if recoveredByEventID != 0 && item.RecoveredByEventID != recoveredByEventID {
			continue
		}
		if recoveryOnly && !workflowEventNeedsRecovery(item.Status) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func workflowOutboxForQuery(items []WorkflowOutbox, r *http.Request) []WorkflowOutbox {
	query := r.URL.Query()
	status := strings.TrimSpace(query.Get("status"))
	eventType := strings.TrimSpace(query.Get("eventType"))
	resource := strings.TrimSpace(query.Get("resource"))
	definitionCode := strings.TrimSpace(query.Get("definitionCode"))
	outboxNo := strings.TrimSpace(query.Get("outboxNo"))
	triggerEventKey := strings.TrimSpace(query.Get("triggerEventKey"))
	resourceID := workflowInt64QueryValue(query.Get("resourceId"))
	triggerEventID := workflowInt64QueryValue(query.Get("triggerEventId"))
	if status == "" && eventType == "" && resource == "" && definitionCode == "" && outboxNo == "" && triggerEventKey == "" && resourceID == 0 && triggerEventID == 0 {
		return items
	}
	filtered := []WorkflowOutbox{}
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if eventType != "" && item.EventType != eventType {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		if definitionCode != "" && item.DefinitionCode != definitionCode {
			continue
		}
		if outboxNo != "" && item.OutboxNo != outboxNo {
			continue
		}
		if triggerEventKey != "" && item.Payload["triggerEventKey"] != triggerEventKey {
			continue
		}
		if resourceID != 0 && item.ResourceID != resourceID {
			continue
		}
		if triggerEventID != 0 && item.TriggerEventID != triggerEventID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func workflowDeliveriesForQuery(items []WorkflowDelivery, r *http.Request) []WorkflowDelivery {
	query := r.URL.Query()
	status := strings.TrimSpace(query.Get("status"))
	eventType := strings.TrimSpace(query.Get("eventType"))
	resource := strings.TrimSpace(query.Get("resource"))
	definitionCode := strings.TrimSpace(query.Get("definitionCode"))
	subscriptionCode := strings.TrimSpace(query.Get("subscriptionCode"))
	targetType := strings.TrimSpace(query.Get("targetType"))
	deliveryNo := strings.TrimSpace(query.Get("deliveryNo"))
	outboxNo := strings.TrimSpace(query.Get("outboxNo"))
	triggerEventKey := strings.TrimSpace(query.Get("triggerEventKey"))
	resourceID := workflowInt64QueryValue(query.Get("resourceId"))
	outboxID := workflowInt64QueryValue(query.Get("outboxId"))
	subscriptionID := workflowInt64QueryValue(query.Get("subscriptionId"))
	triggerEventID := workflowInt64QueryValue(query.Get("triggerEventId"))
	if status == "" && eventType == "" && resource == "" && definitionCode == "" && subscriptionCode == "" && targetType == "" && deliveryNo == "" && outboxNo == "" && triggerEventKey == "" && resourceID == 0 && outboxID == 0 && subscriptionID == 0 && triggerEventID == 0 {
		return items
	}
	filtered := []WorkflowDelivery{}
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if eventType != "" && item.EventType != eventType {
			continue
		}
		if resource != "" && item.Resource != resource {
			continue
		}
		if definitionCode != "" && item.Payload["definitionCode"] != definitionCode {
			continue
		}
		if subscriptionCode != "" && item.SubscriptionCode != subscriptionCode {
			continue
		}
		if targetType != "" && item.TargetType != targetType {
			continue
		}
		if deliveryNo != "" && item.DeliveryNo != deliveryNo {
			continue
		}
		if outboxNo != "" && item.OutboxNo != outboxNo {
			continue
		}
		if triggerEventKey != "" && item.Payload["triggerEventKey"] != triggerEventKey {
			continue
		}
		if resourceID != 0 && item.ResourceID != resourceID {
			continue
		}
		if outboxID != 0 && item.OutboxID != outboxID {
			continue
		}
		if subscriptionID != 0 && item.SubscriptionID != subscriptionID {
			continue
		}
		if triggerEventID != 0 && item.Payload["triggerEventId"] != strconv.FormatInt(triggerEventID, 10) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func workflowInt64QueryValue(value string) int64 {
	if raw := strings.TrimSpace(value); raw != "" {
		parsed, _ := strconv.ParseInt(raw, 10, 64)
		return parsed
	}
	return 0
}

func workflowTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func (a *App) recordWorkflowResultFailure(instance WorkflowInstance, resultErr error) {
	if resultErr == nil {
		return
	}
	_ = a.store.Mutate(func(data *AppData) error {
		appendWorkflowResultFailureLog(data, instance, resultErr)
		return nil
	})
}

func workflowQueryLimit(r *http.Request, key string, defaultValue int, maxValue int) int {
	limit := defaultValue
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	if limit <= 0 {
		return defaultValue
	}
	if maxValue > 0 && limit > maxValue {
		return maxValue
	}
	return limit
}

func (a *App) dispatchWorkflowDeliveries(ids []int64, actor string, requestIP string) WorkflowDeliveryDispatchBatch {
	batch := WorkflowDeliveryDispatchBatch{
		Total:   len(ids),
		Results: []WorkflowDeliveryDispatchResult{},
	}
	for _, id := range ids {
		updated, err := a.dispatchWorkflowDelivery(id, actor, requestIP)
		result := WorkflowDeliveryDispatchResult{
			DeliveryID: id,
			Status:     "skipped",
		}
		if err != nil {
			result.Error = err.Error()
			batch.Skipped++
			batch.Results = append(batch.Results, result)
			continue
		}
		result.DeliveryID = updated.ID
		result.DeliveryNo = updated.DeliveryNo
		result.OutboxNo = updated.OutboxNo
		result.EventType = updated.EventType
		result.Status = updated.Status
		result.ResponseStatus = updated.ResponseStatus
		result.Error = updated.LastError
		batch.Dispatched++
		if updated.Status == "succeeded" {
			batch.Succeeded++
		} else {
			batch.Failed++
		}
		batch.Results = append(batch.Results, result)
	}
	return batch
}

func (a *App) recoverStaleWorkflowDeliveries(actor string, requestIP string) []WorkflowDelivery {
	recovered := []WorkflowDelivery{}
	_ = a.store.Mutate(func(data *AppData) error {
		recovered = recoverStaleWorkflowDeliveries(data, workflowDeliveryProcessingTimeout, time.Now(), actor)
		for _, item := range recovered {
			addAudit(data, actor, "recover", "workflow_delivery", item.ID, item.DeliveryNo, requestIP)
		}
		return nil
	})
	return recovered
}

func (a *App) dispatchWorkflowDelivery(id int64, actor string, requestIP string) (WorkflowDelivery, error) {
	var prepared WorkflowDelivery
	var subscription WorkflowSubscription
	err := a.store.Mutate(func(data *AppData) error {
		next, sub, err := prepareWorkflowDeliveryDispatch(data, id)
		if err != nil {
			return err
		}
		prepared = next
		subscription = sub
		addAudit(data, actor, "dispatch", "workflow_delivery", prepared.ID, prepared.DeliveryNo, requestIP)
		return nil
	})
	if err != nil {
		return prepared, err
	}
	requestPayload, responseStatus, responseBody, dispatchErr := sendWorkflowDelivery(prepared, subscription)
	var updated WorkflowDelivery
	err = a.store.Mutate(func(data *AppData) error {
		next, err := finishWorkflowDeliveryDispatch(data, prepared.ID, requestPayload, responseStatus, responseBody, dispatchErr)
		if err != nil {
			return err
		}
		updated = next
		return nil
	})
	return updated, err
}

func sendWorkflowDelivery(delivery WorkflowDelivery, subscription WorkflowSubscription) (string, int, string, error) {
	body := map[string]interface{}{
		"deliveryNo":       delivery.DeliveryNo,
		"outboxNo":         delivery.OutboxNo,
		"subscriptionCode": delivery.SubscriptionCode,
		"eventType":        delivery.EventType,
		"resource":         delivery.Resource,
		"resourceId":       delivery.ResourceID,
		"payload":          delivery.Payload,
		"attempts":         delivery.Attempts,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", 0, "", err
	}
	requestPayload := string(raw)
	endpoint := strings.TrimSpace(delivery.Endpoint)
	if err := validateNoMockEndpoint(endpoint, "工作流投递 endpoint"); err != nil {
		return requestPayload, 0, "", err
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return requestPayload, 0, "", fmt.Errorf("unsupported workflow delivery endpoint: %s", endpoint)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return requestPayload, 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Workflow-Event", delivery.EventType)
	req.Header.Set("X-CBMP-Workflow-Delivery", delivery.DeliveryNo)
	req.Header.Set("X-CBMP-Workflow-Subscription", subscription.Code)
	if secret := strings.TrimSpace(subscription.Secret); secret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Set("X-CBMP-Timestamp", timestamp)
		req.Header.Set("X-CBMP-Signature", workflowDeliverySignature(secret, timestamp, raw))
	}
	timeoutSeconds := subscription.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}
	client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return requestPayload, 0, "", err
	}
	defer resp.Body.Close()
	responseRaw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	responseBody := string(responseRaw)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return requestPayload, resp.StatusCode, responseBody, fmt.Errorf("workflow delivery failed with HTTP %d", resp.StatusCode)
	}
	return requestPayload, resp.StatusCode, responseBody, nil
}

func workflowDeliverySignature(secret string, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func publishWorkflowDefinitionVersion(data *AppData, baseID int64, item WorkflowDefinition) (WorkflowDefinition, error) {
	base, ok := findWorkflowDefinitionByID(*data, baseID)
	if !ok {
		return WorkflowDefinition{}, fmt.Errorf("工作流定义不存在")
	}
	if strings.TrimSpace(item.Name) == "" {
		item = base
	} else {
		item.Code = base.Code
		item.Category = base.Category
		item.Resource = base.Resource
		if len(item.Steps) == 0 {
			item.Steps = base.Steps
		}
		if item.Trigger.EventType == "" && item.Trigger.Resource == "" && len(item.Trigger.Conditions) == 0 {
			item.Trigger = base.Trigger
		}
	}
	item.ID = 0
	item.Code = base.Code
	item.Category = base.Category
	item.Resource = base.Resource
	item.Version = maxWorkflowDefinitionVersion(*data, base.Category, base.Code) + 1
	item.Status = "active"
	if err := normalizeWorkflowDefinition(&item); err != nil {
		return WorkflowDefinition{}, err
	}
	for i := range data.WorkflowDefinitions {
		if data.WorkflowDefinitions[i].Category == item.Category && data.WorkflowDefinitions[i].Code == item.Code && data.WorkflowDefinitions[i].Status == "active" {
			data.WorkflowDefinitions[i].Status = "disabled"
		}
	}
	item.ID = nextID(data, "workflowDefinition")
	data.WorkflowDefinitions = append(data.WorkflowDefinitions, item)
	syncApprovalFlowFromCurrentWorkflowDefinition(data, item)
	return item, nil
}

func rollbackWorkflowDefinitionVersion(data *AppData, id int64) (WorkflowDefinition, error) {
	targetIndex := -1
	for i := range data.WorkflowDefinitions {
		if data.WorkflowDefinitions[i].ID == id {
			targetIndex = i
			break
		}
	}
	if targetIndex < 0 {
		return WorkflowDefinition{}, fmt.Errorf("工作流定义不存在")
	}
	target := data.WorkflowDefinitions[targetIndex]
	for i := range data.WorkflowDefinitions {
		if data.WorkflowDefinitions[i].Category == target.Category && data.WorkflowDefinitions[i].Code == target.Code {
			data.WorkflowDefinitions[i].Status = "disabled"
		}
	}
	data.WorkflowDefinitions[targetIndex].Status = "active"
	syncApprovalFlowFromCurrentWorkflowDefinition(data, data.WorkflowDefinitions[targetIndex])
	return data.WorkflowDefinitions[targetIndex], nil
}

func maxWorkflowDefinitionVersion(data AppData, category string, code string) int {
	maxVersion := 0
	for _, definition := range data.WorkflowDefinitions {
		if definition.Category == category && definition.Code == code && workflowDefinitionVersion(definition) > maxVersion {
			maxVersion = workflowDefinitionVersion(definition)
		}
	}
	return maxVersion
}

func normalizeWorkflowDefinition(item *WorkflowDefinition) error {
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.Category = fallback(strings.TrimSpace(item.Category), workflowCategoryApproval)
	item.Resource = strings.TrimSpace(item.Resource)
	item.Trigger = normalizeWorkflowTrigger(item.Trigger, *item)
	item.Status = strings.TrimSpace(item.Status)
	if item.Code == "" || item.Name == "" {
		return fmt.Errorf("工作流定义必须包含编码和名称")
	}
	if len(item.Steps) == 0 {
		return fmt.Errorf("工作流定义至少需要一个步骤")
	}
	status, err := normalizeConfigStatus(item.Status)
	if err != nil {
		return err
	}
	item.Status = status
	if item.Version <= 0 {
		item.Version = 1
	}
	seen := map[int]bool{}
	for i := range item.Steps {
		item.Steps[i].Code = strings.TrimSpace(item.Steps[i].Code)
		item.Steps[i].Name = strings.TrimSpace(item.Steps[i].Name)
		item.Steps[i].Type = fallback(strings.TrimSpace(item.Steps[i].Type), item.Category)
		item.Steps[i].RoleCode = strings.TrimSpace(item.Steps[i].RoleCode)
		item.Steps[i].Action = fallback(strings.TrimSpace(item.Steps[i].Action), "approve")
		if item.Steps[i].SLAHours < 0 {
			item.Steps[i].SLAHours = 0
		}
		if item.Steps[i].RoleCode == "" {
			return fmt.Errorf("工作流步骤必须包含处理角色")
		}
		if item.Steps[i].Seq <= 0 {
			item.Steps[i].Seq = i + 1
		}
		if item.Steps[i].Code == "" {
			item.Steps[i].Code = fmt.Sprintf("%s.step.%d", item.Code, item.Steps[i].Seq)
		}
		if item.Steps[i].Name == "" {
			item.Steps[i].Name = fmt.Sprintf("步骤 %d", item.Steps[i].Seq)
		}
		if seen[item.Steps[i].Seq] {
			return fmt.Errorf("工作流步骤序号重复")
		}
		seen[item.Steps[i].Seq] = true
	}
	item.Steps = sortedWorkflowSteps(item.Steps)
	return nil
}

func syncApprovalFlowFromWorkflowDefinition(data *AppData, item WorkflowDefinition) {
	flow := ApprovalFlow{
		Code:     item.Code,
		Name:     item.Name,
		Resource: item.Resource,
		Steps:    approvalStepsFromWorkflowSteps(item.Steps),
		Status:   item.Status,
	}
	for i := range data.ApprovalFlows {
		if data.ApprovalFlows[i].Code == flow.Code {
			flow.ID = data.ApprovalFlows[i].ID
			data.ApprovalFlows[i] = flow
			ensureCounterAtLeast(data, "approvalFlow", flow.ID)
			return
		}
	}
	flow.ID = nextID(data, "approvalFlow")
	data.ApprovalFlows = append(data.ApprovalFlows, flow)
}

func syncApprovalFlowFromCurrentWorkflowDefinition(data *AppData, item WorkflowDefinition) {
	if item.Category != workflowCategoryApproval {
		return
	}
	if active, ok := currentActiveWorkflowDefinition(*data, item.Category, item.Code); ok {
		syncApprovalFlowFromWorkflowDefinition(data, active)
		return
	}
	syncApprovalFlowFromWorkflowDefinition(data, item)
}

func currentActiveWorkflowDefinition(data AppData, category string, code string) (WorkflowDefinition, bool) {
	var selected WorkflowDefinition
	found := false
	for _, definition := range data.WorkflowDefinitions {
		if definition.Category != category || definition.Code != code || definition.Status != "active" {
			continue
		}
		if !found || workflowDefinitionVersion(definition) > workflowDefinitionVersion(selected) {
			selected = definition
			found = true
		}
	}
	if found {
		selected.Steps = sortedWorkflowSteps(selected.Steps)
	}
	return selected, found
}

func approvalStepsFromWorkflowSteps(steps []WorkflowStep) []ApprovalStep {
	out := make([]ApprovalStep, 0, len(steps))
	for _, step := range sortedWorkflowSteps(steps) {
		out = append(out, ApprovalStep{
			Seq:      step.Seq,
			RoleCode: step.RoleCode,
			Action:   fallback(step.Action, "approve"),
		})
	}
	return out
}

func normalizeWorkflowTrigger(trigger WorkflowTrigger, definition WorkflowDefinition) WorkflowTrigger {
	trigger.EventType = strings.TrimSpace(trigger.EventType)
	trigger.Resource = strings.TrimSpace(trigger.Resource)
	if trigger.EventType == "" && trigger.Resource == "" && len(trigger.Conditions) == 0 {
		return defaultWorkflowTrigger(definition)
	}
	if trigger.Resource == "" {
		trigger.Resource = strings.TrimSpace(definition.Resource)
	}
	conditions := make([]WorkflowCondition, 0, len(trigger.Conditions))
	for _, condition := range trigger.Conditions {
		condition.Field = strings.TrimSpace(condition.Field)
		condition.Operator = fallback(strings.TrimSpace(condition.Operator), "equals")
		condition.Value = strings.TrimSpace(condition.Value)
		if condition.Field == "" {
			continue
		}
		conditions = append(conditions, condition)
	}
	trigger.Conditions = conditions
	return trigger
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
	if len(parts) == 1 && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var deleted DataDictionary
		err := a.store.Mutate(func(data *AppData) error {
			for i, item := range data.DataDictionaries {
				if item.ID != id {
					continue
				}
				if item.Status == "active" {
					return fmt.Errorf("启用中的字典项不能删除，请先禁用或转草稿")
				}
				deleted = item
				data.DataDictionaries = append(data.DataDictionaries[:i], data.DataDictionaries[i+1:]...)
				addAudit(data, session.User.Username, "delete", "data_dictionary", id, item.Type+"/"+item.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("数据字典不存在")
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("system.dictionary.deleted", deleted)
		writeJSON(w, http.StatusOK, deleted)
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
	for _, existingItem := range existing {
		if item.ID > 0 && existingItem.ID != item.ID && existingItem.Type == item.Type && existingItem.Code == item.Code {
			return fmt.Errorf("数据字典类型和编码已存在")
		}
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
