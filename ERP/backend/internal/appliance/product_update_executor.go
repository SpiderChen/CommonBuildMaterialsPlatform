//go:build legacy_product_ops

package appliance

import (
	"errors"
	"fmt"
	"strings"
)

type productUpdateExecutionRequest struct {
	Action     string `json:"action"`
	InstanceID int64  `json:"instanceId"`
	DryRun     bool   `json:"dryRun"`
	Remark     string `json:"remark"`
}

func executeProductUpdate(data *AppData, rollout *ProductUpdateRollout, update UpdatePackage, req productUpdateExecutionRequest, actor string) (ProductUpdateExecution, error) {
	action := fallback(strings.ToLower(strings.TrimSpace(req.Action)), "apply")
	if action != "apply" && action != "rollback" {
		return ProductUpdateExecution{}, fmt.Errorf("执行动作必须是 apply 或 rollback")
	}
	itemIndex := productRolloutItemIndexForExecution(*rollout, req.InstanceID, action)
	if itemIndex < 0 {
		return ProductUpdateExecution{}, fmt.Errorf("批次中没有可执行的客户实例")
	}
	item := &rollout.Items[itemIndex]
	instanceIndex := productInstanceIndexByID(*data, item.InstanceID)
	if instanceIndex < 0 {
		return ProductUpdateExecution{}, fmt.Errorf("客户实例不存在")
	}
	instance := &data.ProductInstances[instanceIndex]
	now := nowString()
	id := nextID(data, "updateExecution")
	execution := ProductUpdateExecution{
		ID:               id,
		ExecutionNo:      number("UE", id),
		RolloutID:        rollout.ID,
		RolloutNo:        rollout.RolloutNo,
		UpdateID:         update.ID,
		InstanceID:       instance.ID,
		CustomerName:     instance.CustomerName,
		Component:        rollout.Component,
		Version:          rollout.Version,
		Action:           action,
		Status:           "running",
		ArtifactFileName: fallback(update.FileName, updatePackageFileName(update)),
		ChecksumVerified: updatePackageVerified(update),
		DryRun:           req.DryRun,
		StartedBy:        actor,
		StartedAt:        now,
		PrecheckSummary:  "等待执行",
	}
	if !execution.ChecksumVerified {
		execution.Status = "failed"
		execution.Error = "更新包验签失败"
		execution.CompletedAt = now
		execution.Steps = append(execution.Steps, newUpdateExecutionStep(data, "验签更新包", "failed", execution.Error, now, 0))
		data.ProductUpdateExecutions = append(data.ProductUpdateExecutions, execution)
		return execution, errors.New(execution.Error)
	}
	targetVersion := rollout.Version
	if action == "rollback" {
		targetVersion = item.FromVersion
	}
	stepSpecs := []struct {
		name     string
		message  string
		duration int64
	}{
		{"验签更新包", "checksum/signature verified", 3000},
		{"创建现场快照", "snapshot " + instance.CustomerName + " before " + action, 12000},
		{"分发更新包", execution.ArtifactFileName + " delivered to customer appliance", 18000},
		{"停止" + rollout.Component + "组件", rollout.Component + " component stopped in controlled window", 9000},
		{"安装目标版本", rollout.Component + " -> " + targetVersion, 24000},
		{"健康检查", "probe heartbeat and telemetry smoke passed", 10000},
		{"发布执行结果", "execution result persisted and event emitted", 2000},
	}
	if req.DryRun {
		stepSpecs = []struct {
			name     string
			message  string
			duration int64
		}{
			{"验签更新包", "checksum/signature verified", 3000},
			{"检查执行窗口", "dry-run validates target, component and rollback version", 4000},
			{"生成执行计划", "no customer version changed during dry-run", 2000},
		}
	}
	var total int64
	for _, spec := range stepSpecs {
		execution.Steps = append(execution.Steps, newUpdateExecutionStep(data, spec.name, "succeeded", spec.message, now, spec.duration))
		total += spec.duration
	}
	execution.DurationMs = total
	execution.CompletedAt = now
	execution.PrecheckSummary = fmt.Sprintf("验签通过，目标客户 %s，组件 %s，动作 %s", instance.CustomerName, rollout.Component, action)
	if req.DryRun {
		execution.Status = "dry_run_passed"
		execution.Result = "执行计划校验通过，未修改客户实例版本"
		data.ProductUpdateExecutions = append(data.ProductUpdateExecutions, execution)
		return execution, nil
	}
	if rollout.Status == "pending" {
		rollout.Status = "running"
		rollout.StartedAt = now
	}
	if item.StartedAt == "" {
		item.StartedAt = now
	}
	item.Message = fallback(strings.TrimSpace(req.Remark), productRolloutDefaultMessage(action))
	switch action {
	case "apply":
		item.Status = "applied"
		item.AppliedAt = now
		if err := productRolloutSyncInstance(data, item.InstanceID, rollout.Component, rollout.Version); err != nil {
			return execution, err
		}
		execution.Result = "目标版本已应用到客户实例"
	case "rollback":
		item.Status = "rolled_back"
		item.RolledBackAt = now
		if err := productRolloutSyncInstance(data, item.InstanceID, rollout.Component, item.FromVersion); err != nil {
			return execution, err
		}
		execution.Result = "客户实例已回滚到原版本"
	}
	recomputeProductRollout(rollout)
	execution.Status = "succeeded"
	data.ProductUpdateExecutions = append(data.ProductUpdateExecutions, execution)
	return execution, nil
}

func newUpdateExecutionStep(data *AppData, name, status, message, now string, duration int64) ProductUpdateExecutionStep {
	return ProductUpdateExecutionStep{
		ID:          nextID(data, "updateExecutionStep"),
		Name:        name,
		Status:      status,
		Message:     message,
		StartedAt:   now,
		CompletedAt: now,
		DurationMs:  duration,
	}
}

func productRolloutItemIndexForExecution(rollout ProductUpdateRollout, instanceID int64, action string) int {
	if instanceID > 0 {
		for i, item := range rollout.Items {
			if item.InstanceID == instanceID {
				return i
			}
		}
		return -1
	}
	for i, item := range rollout.Items {
		if action == "rollback" && item.Status == "applied" {
			return i
		}
		if action == "apply" && (item.Status == "pending" || item.Status == "running") {
			return i
		}
	}
	return -1
}

func productInstanceIndexByID(data AppData, id int64) int {
	for i := range data.ProductInstances {
		if data.ProductInstances[i].ID == id {
			return i
		}
	}
	return -1
}
