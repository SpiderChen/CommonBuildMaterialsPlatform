//go:build legacy_product_ops

package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type productSystemUpdateTaskRequest struct {
	Action     string `json:"action"`
	InstanceID int64  `json:"instanceId"`
	Remark     string `json:"remark"`
}

type productSystemUpdatePollRequest struct {
	UpdaterToken string `json:"updaterToken"`
	Watermark    string `json:"watermark"`
}

type productSystemUpdateReportRequest struct {
	UpdaterToken   string `json:"updaterToken"`
	Status         string `json:"status"`
	Progress       int    `json:"progress"`
	Step           string `json:"step"`
	Message        string `json:"message"`
	Error          string `json:"error"`
	CurrentVersion string `json:"currentVersion"`
	UpdaterVersion string `json:"updaterVersion"`
}

type productSystemUpdateTaskResponse struct {
	Accepted bool                      `json:"accepted"`
	Instance ProductInstance           `json:"instance"`
	Tasks    []ProductSystemUpdateTask `json:"tasks"`
}

func queueProductSystemUpdateTask(data *AppData, rollout *ProductUpdateRollout, update UpdatePackage, req productSystemUpdateTaskRequest, actor string) (ProductSystemUpdateTask, error) {
	action := fallback(strings.ToLower(strings.TrimSpace(req.Action)), "apply")
	if action != "apply" && action != "rollback" {
		return ProductSystemUpdateTask{}, fmt.Errorf("端内系统更新动作必须是 apply 或 rollback")
	}
	itemIndex := productRolloutItemIndexForExecution(*rollout, req.InstanceID, action)
	if itemIndex < 0 {
		return ProductSystemUpdateTask{}, fmt.Errorf("批次中没有可下发的客户实例")
	}
	item := &rollout.Items[itemIndex]
	instanceIndex := productInstanceIndexByID(*data, item.InstanceID)
	if instanceIndex < 0 {
		return ProductSystemUpdateTask{}, fmt.Errorf("客户实例不存在")
	}
	instance := &data.ProductInstances[instanceIndex]
	if !updatePackageVerified(update) {
		return ProductSystemUpdateTask{}, fmt.Errorf("更新包验签失败")
	}
	now := nowString()
	targetVersion := rollout.Version
	if action == "rollback" {
		targetVersion = item.FromVersion
	}
	if rollout.Status == "pending" {
		rollout.Status = "running"
		rollout.StartedAt = now
	}
	if item.StartedAt == "" {
		item.StartedAt = now
	}
	item.Status = "running"
	item.Message = fallback(strings.TrimSpace(req.Remark), "已下发端内系统更新任务")

	executionID := nextID(data, "updateExecution")
	execution := ProductUpdateExecution{
		ID:               executionID,
		ExecutionNo:      number("UE", executionID),
		RolloutID:        rollout.ID,
		RolloutNo:        rollout.RolloutNo,
		UpdateID:         update.ID,
		InstanceID:       instance.ID,
		CustomerName:     instance.CustomerName,
		Component:        rollout.Component,
		Version:          targetVersion,
		Action:           action,
		Status:           "queued_for_updater",
		ArtifactFileName: fallback(update.FileName, updatePackageFileName(update)),
		ChecksumVerified: true,
		StartedBy:        actor,
		StartedAt:        now,
		PrecheckSummary:  fmt.Sprintf("已生成端内系统更新任务，目标客户 %s，组件 %s，动作 %s", instance.CustomerName, rollout.Component, action),
		Steps: []ProductUpdateExecutionStep{
			newUpdateExecutionStep(data, "验签更新包", "succeeded", "checksum/signature verified", now, 3000),
			newUpdateExecutionStep(data, "下发端内系统更新任务", "queued", "waiting for client/server updater polling", now, 0),
		},
	}
	data.ProductUpdateExecutions = append(data.ProductUpdateExecutions, execution)

	taskID := nextID(data, "systemUpdateTask")
	task := ProductSystemUpdateTask{
		ID:               taskID,
		TaskNo:           number("SU", taskID),
		ExecutionID:      execution.ID,
		ExecutionNo:      execution.ExecutionNo,
		RolloutID:        rollout.ID,
		RolloutNo:        rollout.RolloutNo,
		RolloutItemID:    item.ID,
		UpdateID:         update.ID,
		InstanceID:       instance.ID,
		CustomerName:     instance.CustomerName,
		Watermark:        instance.Watermark,
		Component:        rollout.Component,
		Version:          targetVersion,
		FromVersion:      item.FromVersion,
		Action:           action,
		Status:           "queued",
		ArtifactFileName: execution.ArtifactFileName,
		Checksum:         update.Checksum,
		Signature:        update.Signature,
		DownloadURL:      fmt.Sprintf("/api/system/updates/%d/download", update.ID),
		UpdaterTokenHint: tokenHint(instance.ProbeToken),
		CreatedBy:        actor,
		CreatedAt:        now,
		Remark:           strings.TrimSpace(req.Remark),
	}
	task.Logs = append(task.Logs, ProductSystemUpdateTaskLog{
		ID: nextID(data, "systemUpdateTaskLog"), Status: "queued", Progress: 0, Step: "queued",
		Message: "平台已下发端内系统更新任务", CreatedAt: now,
	})
	data.ProductSystemUpdateTasks = append(data.ProductSystemUpdateTasks, task)
	recomputeProductRollout(rollout)
	return task, nil
}

func (a *App) productOpsSystemUpdate(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 && parts[0] == "tasks" && (r.Method == http.MethodGet || r.Method == http.MethodPost) {
		var req productSystemUpdatePollRequest
		if r.Method == http.MethodPost {
			_ = readJSON(r, &req)
		}
		token := updaterTokenFromRequest(r, req.UpdaterToken)
		if token == "" {
			unauthorized(w)
			return
		}
		var response productSystemUpdateTaskResponse
		err := a.store.Mutate(func(data *AppData) error {
			instanceIndex := productInstanceIndexByUpdaterToken(*data, token, req.Watermark)
			if instanceIndex < 0 {
				return fmt.Errorf("updater token invalid")
			}
			instance := &data.ProductInstances[instanceIndex]
			now := nowString()
			instance.LastHeartbeatAt = now
			for i := range data.ProductSystemUpdateTasks {
				task := &data.ProductSystemUpdateTasks[i]
				if task.InstanceID != instance.ID || (task.Status != "queued" && task.Status != "assigned" && task.Status != "running") {
					continue
				}
				if task.Status == "queued" {
					task.Status = "assigned"
					task.ClaimedAt = now
					task.LastHeartbeatAt = now
					task.Logs = append(task.Logs, ProductSystemUpdateTaskLog{
						ID: nextID(data, "systemUpdateTaskLog"), Status: task.Status, Progress: task.Progress, Step: "claimed",
						Message: "端内更新器已拉取任务", CreatedAt: now,
					})
					markUpdateExecutionSystemStep(data, task.ExecutionID, "running", "端内更新器已拉取任务", now)
				}
				response.Tasks = append(response.Tasks, *task)
			}
			response.Accepted = true
			response.Instance = *instance
			addAudit(data, "updater:"+instance.Watermark, "poll", "system_update_task", instance.ID, fmt.Sprintf("%d task(s)", len(response.Tasks)), clientIP(r))
			return nil
		})
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, response)
		return
	}
	if len(parts) == 3 && parts[0] == "tasks" && parts[2] == "report" && r.Method == http.MethodPost {
		taskNo := strings.TrimSpace(parts[1])
		var req productSystemUpdateReportRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid system update report payload")
			return
		}
		token := updaterTokenFromRequest(r, req.UpdaterToken)
		if token == "" {
			unauthorized(w)
			return
		}
		var updated ProductSystemUpdateTask
		err := a.store.Mutate(func(data *AppData) error {
			taskIndex := productSystemUpdateTaskIndex(*data, taskNo)
			if taskIndex < 0 {
				return fmt.Errorf("端内系统更新任务不存在")
			}
			task := &data.ProductSystemUpdateTasks[taskIndex]
			instanceIndex := productInstanceIndexByID(*data, task.InstanceID)
			if instanceIndex < 0 || data.ProductInstances[instanceIndex].ProbeToken != token {
				return fmt.Errorf("updater token invalid")
			}
			now := nowString()
			status := normalizeSystemUpdateStatus(req.Status)
			if status == "" {
				status = task.Status
			}
			task.Status = status
			task.Progress = clampProgress(req.Progress, task.Progress)
			task.LastHeartbeatAt = now
			task.Logs = append(task.Logs, ProductSystemUpdateTaskLog{
				ID: nextID(data, "systemUpdateTaskLog"), Status: task.Status, Progress: task.Progress, Step: fallback(strings.TrimSpace(req.Step), task.Status),
				Message: fallback(strings.TrimSpace(req.Message), strings.TrimSpace(req.Error)), CreatedAt: now,
			})
			if task.StartedAt == "" && (status == "running" || status == "succeeded" || status == "failed" || status == "rolled_back") {
				task.StartedAt = now
			}
			if status == "succeeded" || status == "failed" || status == "rolled_back" {
				task.CompletedAt = now
				task.Progress = 100
				task.Result = strings.TrimSpace(req.Message)
				task.Error = strings.TrimSpace(req.Error)
				completeSystemUpdateTask(data, task, &data.ProductInstances[instanceIndex], req)
			} else {
				markUpdateExecutionSystemStep(data, task.ExecutionID, "running", fallback(req.Message, "端内更新器执行中"), now)
			}
			updated = *task
			addAudit(data, "updater:"+task.Watermark, "report", "system_update_task", task.ID, task.TaskNo+" "+task.Status, clientIP(r))
			return nil
		})
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		a.emit("product_ops.system_update.reported", updated)
		writeJSON(w, http.StatusCreated, updated)
		return
	}
	writeError(w, http.StatusNotFound, "unknown system update route")
}

func productSystemUpdateTaskIndex(data AppData, taskNo string) int {
	for i := range data.ProductSystemUpdateTasks {
		task := data.ProductSystemUpdateTasks[i]
		if task.TaskNo == taskNo || strconv.FormatInt(task.ID, 10) == taskNo {
			return i
		}
	}
	return -1
}

func normalizeSystemUpdateStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "assigned", "running", "succeeded", "failed", "rolled_back":
		return strings.ToLower(strings.TrimSpace(status))
	case "success", "done", "completed", "installed":
		return "succeeded"
	case "rollback", "rolledback":
		return "rolled_back"
	case "error":
		return "failed"
	default:
		return ""
	}
}

func clampProgress(next, current int) int {
	if next <= 0 {
		return current
	}
	if next > 100 {
		return 100
	}
	if next < current {
		return current
	}
	return next
}

func completeSystemUpdateTask(data *AppData, task *ProductSystemUpdateTask, instance *ProductInstance, req productSystemUpdateReportRequest) {
	status := task.Status
	now := fallback(task.CompletedAt, nowString())
	executionIndex := productUpdateExecutionIndexByID(*data, task.ExecutionID)
	if executionIndex >= 0 {
		execution := &data.ProductUpdateExecutions[executionIndex]
		execution.CompletedAt = now
		execution.Status = status
		execution.DurationMs = elapsedMs(execution.StartedAt, now)
		execution.Result = fallback(strings.TrimSpace(req.Message), task.Result)
		execution.Error = fallback(strings.TrimSpace(req.Error), task.Error)
		stepStatus := "succeeded"
		if status == "failed" {
			stepStatus = "failed"
		}
		execution.Steps = append(execution.Steps, newUpdateExecutionStep(data, "端内更新回执", stepStatus, fallback(req.Message, req.Error), now, 0))
	}
	rolloutIndex := productUpdateRolloutIndexByID(*data, task.RolloutID)
	if rolloutIndex < 0 {
		return
	}
	rollout := &data.ProductUpdateRollouts[rolloutIndex]
	itemIndex := productRolloutItemIndex(*rollout, task.InstanceID)
	if itemIndex < 0 {
		return
	}
	item := &rollout.Items[itemIndex]
	item.Message = fallback(req.Message, task.Result)
	switch status {
	case "succeeded":
		item.Status = "applied"
		item.AppliedAt = now
		version := fallback(strings.TrimSpace(req.CurrentVersion), task.Version)
		_ = productRolloutSyncInstance(data, instance.ID, task.Component, version)
	case "rolled_back":
		item.Status = "rolled_back"
		item.RolledBackAt = now
		version := fallback(strings.TrimSpace(req.CurrentVersion), task.FromVersion)
		_ = productRolloutSyncInstance(data, instance.ID, task.Component, version)
	case "failed":
		item.Status = "failed"
		item.Message = fallback(req.Error, req.Message)
	}
	recomputeProductRollout(rollout)
}

func markUpdateExecutionSystemStep(data *AppData, executionID int64, status, message, now string) {
	index := productUpdateExecutionIndexByID(*data, executionID)
	if index < 0 {
		return
	}
	execution := &data.ProductUpdateExecutions[index]
	execution.Status = "updater_" + status
	execution.Steps = append(execution.Steps, newUpdateExecutionStep(data, "端内更新进度", status, message, now, 0))
}

func productUpdateExecutionIndexByID(data AppData, id int64) int {
	for i := range data.ProductUpdateExecutions {
		if data.ProductUpdateExecutions[i].ID == id {
			return i
		}
	}
	return -1
}

func productUpdateRolloutIndexByID(data AppData, id int64) int {
	for i := range data.ProductUpdateRollouts {
		if data.ProductUpdateRollouts[i].ID == id {
			return i
		}
	}
	return -1
}

func elapsedMs(startText, endText string) int64 {
	start, ok := parseLocalDateTime(startText)
	if !ok {
		return 0
	}
	end, ok := parseLocalDateTime(endText)
	if !ok || end.Before(start) {
		return 0
	}
	return end.Sub(start).Milliseconds()
}

func tokenHint(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 10 {
		return token
	}
	return token[:6] + "..." + token[len(token)-4:]
}
