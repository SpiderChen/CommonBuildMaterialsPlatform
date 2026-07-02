package appliance

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	workflowCategoryApproval          = "approval"
	workflowDeliveryProcessingTimeout = 15 * time.Minute
)

type workflowStartRequest struct {
	DefinitionCode string
	TriggerEventID int64
	Resource       string
	ResourceID     int64
	ResourceNo     string
	Title          string
	Applicant      string
	Reason         string
	Variables      map[string]string
}

type workflowTaskEscalationTarget struct {
	TaskID   int64
	TaskNo   string
	FromRole string
	ToRole   string
}

type workflowActionRequest struct {
	Action        string
	Actor         string
	ActorRole     string
	Comment       string
	AllowOverride bool
}

type workflowEventRequest struct {
	EventType  string
	Source     string
	EventKey   string
	Resource   string
	ResourceID int64
	ResourceNo string
	Title      string
	Actor      string
	Reason     string
	Variables  map[string]string
	ReplayOfID int64
}

type WorkflowEventPreview struct {
	Event              WorkflowEvent               `json:"event"`
	MatchedDefinitions []string                    `json:"matchedDefinitions"`
	Matches            []WorkflowEventPreviewMatch `json:"matches"`
	WillStart          int                         `json:"willStart"`
	Duplicate          bool                        `json:"duplicate"`
	DuplicateEventNo   string                      `json:"duplicateEventNo,omitempty"`
	Warnings           []string                    `json:"warnings,omitempty"`
}

type WorkflowEventPreviewMatch struct {
	DefinitionID      int64  `json:"definitionId"`
	DefinitionCode    string `json:"definitionCode"`
	DefinitionName    string `json:"definitionName"`
	Category          string `json:"category"`
	Resource          string `json:"resource"`
	StepCount         int    `json:"stepCount"`
	FirstRole         string `json:"firstRole,omitempty"`
	WillStart         bool   `json:"willStart"`
	PendingInstanceID int64  `json:"pendingInstanceId,omitempty"`
	PendingInstanceNo string `json:"pendingInstanceNo,omitempty"`
	Reason            string `json:"reason,omitempty"`
}

func ensureWorkflowDefaults(data *AppData) bool {
	changed := false
	for _, flow := range data.ApprovalFlows {
		if updated, _ := upsertApprovalWorkflowDefinition(data, flow); updated {
			changed = true
		}
	}
	return changed
}

func workflowDefinitionsFromApprovalFlows(flows []ApprovalFlow) []WorkflowDefinition {
	definitions := make([]WorkflowDefinition, 0, len(flows))
	for _, flow := range flows {
		definitions = append(definitions, approvalWorkflowDefinition(flow))
	}
	return definitions
}

func upsertApprovalWorkflowDefinition(data *AppData, flow ApprovalFlow) (bool, error) {
	if strings.TrimSpace(flow.Code) == "" {
		return false, nil
	}
	desired := approvalWorkflowDefinition(flow)
	if desired.ID == 0 {
		desired.ID = nextID(data, "workflowDefinition")
	}
	ensureCounterAtLeast(data, "workflowDefinition", desired.ID)
	if i := workflowDefinitionIndexForCode(*data, desired.Category, desired.Code); i >= 0 {
		existing := data.WorkflowDefinitions[i]
		desired.ID = data.WorkflowDefinitions[i].ID
		desired.Version = nonZeroIntAsInt(existing.Version, desired.Version)
		ensureCounterAtLeast(data, "workflowDefinition", desired.ID)
		if !workflowDefinitionUsesDefaultTrigger(existing) {
			desired.Trigger = existing.Trigger
		}
		desired.Steps = preserveWorkflowStepRuntimeConfig(existing.Steps, desired.Steps)
		if err := validateWorkflowDefinitionChange(*data, existing, desired); err != nil {
			return false, err
		}
		if reflect.DeepEqual(data.WorkflowDefinitions[i], desired) {
			return false, nil
		}
		data.WorkflowDefinitions[i] = desired
		return true, nil
	}
	data.WorkflowDefinitions = append(data.WorkflowDefinitions, desired)
	return true, nil
}

func preserveWorkflowStepRuntimeConfig(existing []WorkflowStep, desired []WorkflowStep) []WorkflowStep {
	out := append([]WorkflowStep{}, desired...)
	for i := range out {
		if out[i].SLAHours != 0 {
			continue
		}
		for _, previous := range existing {
			if previous.Seq == out[i].Seq && previous.RoleCode == out[i].RoleCode && previous.Action == out[i].Action {
				out[i].SLAHours = previous.SLAHours
				break
			}
		}
	}
	return out
}

func approvalWorkflowDefinition(flow ApprovalFlow) WorkflowDefinition {
	definition := WorkflowDefinition{
		ID:       flow.ID,
		Code:     strings.TrimSpace(flow.Code),
		Name:     strings.TrimSpace(flow.Name),
		Category: workflowCategoryApproval,
		Resource: strings.TrimSpace(flow.Resource),
		Steps:    workflowStepsFromApprovalSteps(flow.Code, flow.Steps),
		Status:   strings.TrimSpace(flow.Status),
		Version:  1,
	}
	if definition.Trigger.EventType == "" && definition.Trigger.Resource == "" {
		definition.Trigger = defaultWorkflowTrigger(definition)
	}
	return definition
}

func workflowDefinitionUsesDefaultTrigger(definition WorkflowDefinition) bool {
	trigger := definition.Trigger
	if trigger.EventType == "" && trigger.Resource == "" && len(trigger.Conditions) == 0 {
		return true
	}
	return reflect.DeepEqual(trigger, defaultWorkflowTrigger(definition))
}

func validateWorkflowDefinitionChange(data AppData, current WorkflowDefinition, next WorkflowDefinition) error {
	if pendingWorkflowInstanceCount(data, current) == 0 {
		return nil
	}
	if strings.TrimSpace(next.Status) != "active" {
		return fmt.Errorf("工作流定义存在待处理实例，不能停用")
	}
	if strings.TrimSpace(current.Code) != strings.TrimSpace(next.Code) {
		return fmt.Errorf("工作流定义存在待处理实例，不能修改编码")
	}
	if strings.TrimSpace(current.Category) != strings.TrimSpace(next.Category) {
		return fmt.Errorf("工作流定义存在待处理实例，不能修改类型")
	}
	if strings.TrimSpace(current.Resource) != strings.TrimSpace(next.Resource) {
		return fmt.Errorf("工作流定义存在待处理实例，不能修改业务对象")
	}
	if !reflect.DeepEqual(sortedWorkflowSteps(current.Steps), sortedWorkflowSteps(next.Steps)) {
		return fmt.Errorf("工作流定义存在待处理实例，不能修改步骤")
	}
	return nil
}

func pendingWorkflowInstanceCount(data AppData, definition WorkflowDefinition) int {
	count := 0
	for _, instance := range data.WorkflowInstances {
		if instance.Status != "pending" {
			continue
		}
		if definition.ID > 0 && instance.DefinitionID == definition.ID {
			count++
			continue
		}
		if strings.TrimSpace(definition.Code) != "" && instance.DefinitionCode == definition.Code {
			count++
		}
	}
	return count
}

func defaultWorkflowTrigger(definition WorkflowDefinition) WorkflowTrigger {
	resource := strings.TrimSpace(definition.Resource)
	trigger := WorkflowTrigger{
		EventType: strings.Trim(strings.TrimSpace(resource)+".submitted", "."),
		Resource:  resource,
	}
	switch strings.TrimSpace(definition.Code) {
	case "order_credit_risk":
		trigger.EventType = "sales_order.risk_detected"
		trigger.Resource = "sales_order"
		trigger.Conditions = []WorkflowCondition{{Field: "riskFlags", Operator: "contains", Value: "credit_limit"}}
	case "price_below_floor":
		trigger.EventType = "sales_order.risk_detected"
		trigger.Resource = "sales_order"
		trigger.Conditions = []WorkflowCondition{{Field: "riskFlags", Operator: "equals", Value: "price_below_floor"}}
	case "inventory_transfer":
		trigger.EventType = "inventory_transfer.submitted"
		trigger.Resource = "inventory_transfer"
	case "contract_version":
		trigger.EventType = "contract.submitted"
		trigger.Resource = "contract"
	}
	return trigger
}

func publishWorkflowEvent(data *AppData, req workflowEventRequest) (WorkflowEvent, []WorkflowInstance, error) {
	ensureWorkflowDefaults(data)
	draft := workflowEventFromRequest(req, 0)
	if draft.EventKey != "" {
		if existing, ok := findDuplicateWorkflowEvent(*data, draft); ok {
			return existing, nil, nil
		}
	}
	eventID := nextID(data, "workflowEvent")
	event := workflowEventFromRequest(req, eventID)
	event.Status = "ignored"
	if event.Resource == "" {
		event.Status = "failed"
		event.Error = "workflow event resource is required"
		data.WorkflowEvents = append(data.WorkflowEvents, event)
		return event, nil, fmt.Errorf("%s", event.Error)
	}
	definitions := matchingWorkflowDefinitions(*data, event)
	data.WorkflowEvents = append(data.WorkflowEvents, event)
	eventIndex := len(data.WorkflowEvents) - 1
	instances := make([]WorkflowInstance, 0, len(definitions))
	for _, definition := range definitions {
		instance, err := startWorkflowInstance(data, workflowStartRequest{
			DefinitionCode: definition.Code,
			TriggerEventID: event.ID,
			Resource:       event.Resource,
			ResourceID:     event.ResourceID,
			ResourceNo:     event.ResourceNo,
			Title:          fallback(event.Title, definition.Name),
			Applicant:      event.Actor,
			Reason:         event.Reason,
			Variables:      event.Variables,
		})
		if err != nil {
			event.Status = "failed"
			event.Error = err.Error()
			data.WorkflowEvents[eventIndex] = event
			return event, instances, err
		}
		event.MatchedDefinitions = append(event.MatchedDefinitions, definition.Code)
		instances = append(instances, instance)
		if definition.Category == workflowCategoryApproval {
			syncApprovalTaskForWorkflowInstance(data, instance)
		}
	}
	if len(instances) > 0 {
		event.Status = "handled"
	}
	data.WorkflowEvents[eventIndex] = event
	return event, instances, nil
}

func previewWorkflowEvent(data AppData, req workflowEventRequest) (WorkflowEventPreview, error) {
	ensureWorkflowDefaults(&data)
	event := workflowEventFromRequest(req, 0)
	event.Status = "preview"
	if event.Resource == "" {
		return WorkflowEventPreview{}, fmt.Errorf("workflow event resource is required")
	}
	preview := WorkflowEventPreview{
		Event:    event,
		Matches:  []WorkflowEventPreviewMatch{},
		Warnings: []string{},
	}
	if event.EventKey != "" {
		if existing, ok := findDuplicateWorkflowEvent(data, event); ok {
			preview.Duplicate = true
			preview.DuplicateEventNo = existing.EventNo
			preview.Warnings = append(preview.Warnings, fmt.Sprintf("事件去重键已存在: %s", existing.EventNo))
		}
	}
	for _, definition := range matchingWorkflowDefinitions(data, event) {
		steps := sortedWorkflowSteps(definition.Steps)
		match := WorkflowEventPreviewMatch{
			DefinitionID:   definition.ID,
			DefinitionCode: definition.Code,
			DefinitionName: definition.Name,
			Category:       definition.Category,
			Resource:       definition.Resource,
			StepCount:      len(steps),
			WillStart:      true,
		}
		if len(steps) > 0 {
			match.FirstRole = steps[0].RoleCode
		}
		if preview.Duplicate {
			match.WillStart = false
			match.Reason = "去重键已存在，发布会返回已有事件"
		}
		if pending, ok := findPendingWorkflowInstance(data, definition.Code, event.Resource, event.ResourceID); ok {
			match.WillStart = false
			match.PendingInstanceID = pending.ID
			match.PendingInstanceNo = pending.InstanceNo
			match.Reason = "已存在相同业务对象的待处理实例"
		}
		if len(steps) == 0 {
			match.WillStart = false
			match.Reason = "流程定义没有步骤"
		}
		preview.MatchedDefinitions = append(preview.MatchedDefinitions, definition.Code)
		if match.WillStart && !preview.Duplicate {
			preview.WillStart++
		}
		preview.Matches = append(preview.Matches, match)
	}
	if len(preview.Matches) == 0 {
		preview.Warnings = append(preview.Warnings, "没有匹配到启用的工作流定义")
	}
	return preview, nil
}

func workflowEventFromRequest(req workflowEventRequest, eventID int64) WorkflowEvent {
	variables := map[string]string{}
	for key, value := range req.Variables {
		key = strings.TrimSpace(key)
		if key != "" {
			variables[key] = strings.TrimSpace(value)
		}
	}
	event := WorkflowEvent{
		ID:         eventID,
		EventType:  strings.TrimSpace(req.EventType),
		Source:     fallback(strings.TrimSpace(req.Source), "erp"),
		EventKey:   strings.TrimSpace(req.EventKey),
		Resource:   strings.TrimSpace(req.Resource),
		ResourceID: req.ResourceID,
		ResourceNo: strings.TrimSpace(req.ResourceNo),
		Title:      strings.TrimSpace(req.Title),
		Actor:      strings.TrimSpace(req.Actor),
		Reason:     strings.TrimSpace(req.Reason),
		Variables:  variables,
		ReplayOfID: req.ReplayOfID,
		CreatedAt:  nowString(),
	}
	if event.ID > 0 {
		event.EventNo = number("WFE", event.ID)
	}
	if event.EventType == "" {
		event.EventType = strings.Trim(strings.TrimSpace(event.Resource)+".submitted", ".")
	}
	return event
}

func findDuplicateWorkflowEvent(data AppData, event WorkflowEvent) (WorkflowEvent, bool) {
	if event.ReplayOfID > 0 {
		return WorkflowEvent{}, false
	}
	if event.EventKey == "" {
		return WorkflowEvent{}, false
	}
	source := fallback(strings.TrimSpace(event.Source), "erp")
	for _, existing := range data.WorkflowEvents {
		if existing.EventKey == event.EventKey && fallback(strings.TrimSpace(existing.Source), "erp") == source && existing.ReplayOfID == 0 {
			return existing, true
		}
	}
	return WorkflowEvent{}, false
}

func replayWorkflowEvent(data *AppData, eventID int64, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	for _, event := range data.WorkflowEvents {
		if event.ID != eventID {
			continue
		}
		if !workflowEventNeedsRecovery(event.Status) {
			return WorkflowEvent{}, nil, fmt.Errorf("只有失败或忽略的工作流事件可以重放")
		}
		variables := map[string]string{}
		for key, value := range event.Variables {
			variables[key] = value
		}
		if event.EventNo != "" {
			variables["replayOf"] = event.EventNo
		}
		replayed, instances, err := publishWorkflowEvent(data, workflowEventRequest{
			EventType:  event.EventType,
			Source:     event.Source,
			EventKey:   event.EventKey,
			Resource:   event.Resource,
			ResourceID: event.ResourceID,
			ResourceNo: event.ResourceNo,
			Title:      event.Title,
			Actor:      fallback(strings.TrimSpace(actor), event.Actor),
			Reason:     event.Reason,
			Variables:  variables,
			ReplayOfID: event.ID,
		})
		if err != nil {
			return replayed, instances, err
		}
		if replayed.Status == "handled" {
			if _, err := resolveWorkflowEvent(data, event.ID, actor, fmt.Sprintf("replayed as %s", replayed.EventNo), replayed.ID); err != nil {
				return replayed, instances, err
			}
		}
		return replayed, instances, nil
	}
	return WorkflowEvent{}, nil, fmt.Errorf("工作流事件不存在")
}

func resolveWorkflowEvent(data *AppData, eventID int64, actor string, resolution string, recoveredByEventID int64) (WorkflowEvent, error) {
	for i := range data.WorkflowEvents {
		if data.WorkflowEvents[i].ID != eventID {
			continue
		}
		if !workflowEventNeedsRecovery(data.WorkflowEvents[i].Status) {
			return WorkflowEvent{}, fmt.Errorf("只有失败或忽略的工作流事件可以标记处理")
		}
		data.WorkflowEvents[i].Status = "resolved"
		data.WorkflowEvents[i].Resolution = fallback(strings.TrimSpace(resolution), "resolved")
		data.WorkflowEvents[i].ResolvedBy = strings.TrimSpace(actor)
		data.WorkflowEvents[i].ResolvedAt = nowString()
		data.WorkflowEvents[i].RecoveredByEventID = recoveredByEventID
		return data.WorkflowEvents[i], nil
	}
	return WorkflowEvent{}, fmt.Errorf("工作流事件不存在")
}

func workflowEventNeedsRecovery(status string) bool {
	return status == "failed" || status == "ignored"
}

func matchingWorkflowDefinitions(data AppData, event WorkflowEvent) []WorkflowDefinition {
	definitions := make([]WorkflowDefinition, 0)
	for _, definition := range data.WorkflowDefinitions {
		if definition.Status != "active" {
			continue
		}
		if workflowDefinitionMatchesEvent(definition, event) {
			definition.Steps = sortedWorkflowSteps(definition.Steps)
			definitions = append(definitions, definition)
		}
	}
	sort.SliceStable(definitions, func(i, j int) bool {
		leftConditions := len(definitions[i].Trigger.Conditions)
		rightConditions := len(definitions[j].Trigger.Conditions)
		if leftConditions != rightConditions {
			return leftConditions > rightConditions
		}
		return definitions[i].Code < definitions[j].Code
	})
	return definitions
}

func workflowDefinitionMatchesEvent(definition WorkflowDefinition, event WorkflowEvent) bool {
	trigger := definition.Trigger
	if trigger.EventType == "" && trigger.Resource == "" {
		trigger = defaultWorkflowTrigger(definition)
	}
	if trigger.EventType != "" && trigger.EventType != event.EventType {
		return false
	}
	triggerResource := fallback(strings.TrimSpace(trigger.Resource), strings.TrimSpace(definition.Resource))
	if triggerResource != "" && triggerResource != event.Resource {
		return false
	}
	for _, condition := range trigger.Conditions {
		if !workflowConditionMatchesEvent(condition, event) {
			return false
		}
	}
	return true
}

func workflowConditionMatchesEvent(condition WorkflowCondition, event WorkflowEvent) bool {
	field := strings.TrimSpace(condition.Field)
	operator := fallback(strings.TrimSpace(condition.Operator), "equals")
	expected := strings.TrimSpace(condition.Value)
	actual, exists := workflowEventFieldValue(event, field)
	switch operator {
	case "exists":
		return exists && actual != ""
	case "missing":
		return !exists || actual == ""
	case "not_equals":
		return actual != expected
	case "contains":
		return strings.Contains(actual, expected)
	case "not_contains":
		return !strings.Contains(actual, expected)
	case "greater_than", "gt":
		return workflowNumericConditionMatches(actual, expected, func(left, right float64) bool { return left > right })
	case "greater_or_equal", "gte":
		return workflowNumericConditionMatches(actual, expected, func(left, right float64) bool { return left >= right })
	case "less_than", "lt":
		return workflowNumericConditionMatches(actual, expected, func(left, right float64) bool { return left < right })
	case "less_or_equal", "lte":
		return workflowNumericConditionMatches(actual, expected, func(left, right float64) bool { return left <= right })
	default:
		return actual == expected
	}
}

func workflowNumericConditionMatches(actual string, expected string, compare func(float64, float64) bool) bool {
	left, leftErr := strconv.ParseFloat(strings.TrimSpace(actual), 64)
	right, rightErr := strconv.ParseFloat(strings.TrimSpace(expected), 64)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return compare(left, right)
}

func workflowEventFieldValue(event WorkflowEvent, field string) (string, bool) {
	switch field {
	case "eventType":
		return event.EventType, event.EventType != ""
	case "resource":
		return event.Resource, event.Resource != ""
	case "resourceId":
		if event.ResourceID == 0 {
			return "", false
		}
		return fmt.Sprintf("%d", event.ResourceID), true
	case "resourceNo":
		return event.ResourceNo, event.ResourceNo != ""
	case "title":
		return event.Title, event.Title != ""
	case "actor":
		return event.Actor, event.Actor != ""
	case "reason":
		return event.Reason, event.Reason != ""
	}
	value, ok := event.Variables[field]
	return value, ok
}

func workflowStepsFromApprovalSteps(flowCode string, steps []ApprovalStep) []WorkflowStep {
	out := make([]WorkflowStep, 0, len(steps))
	for i, step := range steps {
		seq := step.Seq
		if seq <= 0 {
			seq = i + 1
		}
		action := fallback(strings.TrimSpace(step.Action), "approve")
		out = append(out, WorkflowStep{
			Seq:      seq,
			Code:     fmt.Sprintf("%s.step.%d", strings.TrimSpace(flowCode), seq),
			Name:     fmt.Sprintf("Step %d", seq),
			Type:     workflowCategoryApproval,
			RoleCode: strings.TrimSpace(step.RoleCode),
			Action:   action,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Seq < out[j].Seq
	})
	return out
}

func startWorkflowInstance(data *AppData, req workflowStartRequest) (WorkflowInstance, error) {
	ensureWorkflowDefaults(data)
	req.DefinitionCode = strings.TrimSpace(req.DefinitionCode)
	req.Resource = strings.TrimSpace(req.Resource)
	if req.DefinitionCode == "" {
		return WorkflowInstance{}, fmt.Errorf("工作流定义编码不能为空")
	}
	definition, ok := findWorkflowDefinition(*data, req.DefinitionCode, true)
	if !ok {
		return WorkflowInstance{}, fmt.Errorf("工作流定义未配置: %s", req.DefinitionCode)
	}
	if req.Resource == "" {
		req.Resource = definition.Resource
	}
	for _, instance := range data.WorkflowInstances {
		if instance.DefinitionCode == req.DefinitionCode && instance.Resource == req.Resource && instance.ResourceID == req.ResourceID && instance.Status == "pending" {
			return instance, nil
		}
	}
	steps := sortedWorkflowSteps(definition.Steps)
	if len(steps) == 0 {
		return WorkflowInstance{}, fmt.Errorf("工作流定义至少需要一个步骤: %s", req.DefinitionCode)
	}
	now := nowString()
	instanceID := nextID(data, "workflowInstance")
	taskID := nextID(data, "workflowTask")
	instance := WorkflowInstance{
		ID:             instanceID,
		InstanceNo:     number("WF", instanceID),
		DefinitionID:   definition.ID,
		DefinitionCode: definition.Code,
		DefinitionName: definition.Name,
		Category:       definition.Category,
		TriggerEventID: req.TriggerEventID,
		Resource:       req.Resource,
		ResourceID:     req.ResourceID,
		ResourceNo:     req.ResourceNo,
		Title:          req.Title,
		Applicant:      req.Applicant,
		CurrentTaskID:  taskID,
		CurrentStep:    steps[0].Seq,
		CurrentRole:    steps[0].RoleCode,
		Status:         "pending",
		Reason:         req.Reason,
		Variables:      req.Variables,
		CreatedAt:      now,
		UpdatedAt:      now,
		Actions:        []WorkflowAction{},
	}
	data.WorkflowInstances = append(data.WorkflowInstances, instance)
	task := workflowTaskFromStep(taskID, instance, steps[0], now)
	data.WorkflowTasks = append(data.WorkflowTasks, task)
	appendWorkflowLog(data, WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		TriggerEventID: instance.TriggerEventID,
		TaskID:         task.ID,
		TaskNo:         task.TaskNo,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         "instance_started",
		Status:         instance.Status,
		Actor:          instance.Applicant,
		Message:        instance.Title,
		Variables:      instance.Variables,
		CreatedAt:      now,
	})
	appendWorkflowLog(data, workflowLogFromTask(instance, task, "task_created", instance.Applicant, now))
	return instance, nil
}

func actWorkflowInstance(data *AppData, instanceID int64, req workflowActionRequest) (WorkflowInstance, error) {
	req.Action = strings.TrimSpace(req.Action)
	if req.Action != "approve" && req.Action != "reject" {
		return WorkflowInstance{}, fmt.Errorf("工作流动作只能是 approve 或 reject")
	}
	instanceIndex := findWorkflowInstanceIndex(*data, instanceID)
	if instanceIndex < 0 {
		return WorkflowInstance{}, fmt.Errorf("工作流实例不存在")
	}
	instance := &data.WorkflowInstances[instanceIndex]
	if instance.Status != "pending" {
		return WorkflowInstance{}, fmt.Errorf("工作流实例已结束")
	}
	taskIndex := findCurrentWorkflowTaskIndex(*data, *instance)
	if taskIndex < 0 {
		return WorkflowInstance{}, fmt.Errorf("工作流当前任务不存在")
	}
	task := &data.WorkflowTasks[taskIndex]
	if task.Status != "pending" {
		return WorkflowInstance{}, fmt.Errorf("工作流任务已结束")
	}
	if !req.AllowOverride && task.RoleCode != req.ActorRole {
		return WorkflowInstance{}, fmt.Errorf("无权处理当前工作流任务")
	}

	definition, ok := findWorkflowDefinitionForInstance(*data, *instance)
	if !ok {
		return WorkflowInstance{}, fmt.Errorf("工作流定义不存在: %s", instance.DefinitionCode)
	}

	now := nowString()
	action := WorkflowAction{
		Seq:        len(instance.Actions) + 1,
		TaskID:     task.ID,
		Step:       task.Step,
		StepCode:   task.StepCode,
		RoleCode:   task.RoleCode,
		Action:     req.Action,
		Actor:      req.Actor,
		Comment:    strings.TrimSpace(req.Comment),
		ActedAt:    now,
		FromStatus: instance.Status,
	}
	task.CompletedAt = now
	instance.UpdatedAt = now
	if req.Action == "reject" {
		task.Status = "rejected"
		instance.Status = "rejected"
		instance.CurrentTaskID = 0
		instance.CurrentRole = ""
		instance.CompletedAt = now
		action.ToStatus = instance.Status
		instance.Actions = append(instance.Actions, action)
		appendWorkflowLog(data, workflowLogFromTask(*instance, *task, "task_rejected", req.Actor, now))
		appendWorkflowLog(data, WorkflowLog{
			InstanceID:     instance.ID,
			InstanceNo:     instance.InstanceNo,
			TriggerEventID: instance.TriggerEventID,
			TaskID:         task.ID,
			TaskNo:         task.TaskNo,
			DefinitionCode: instance.DefinitionCode,
			Resource:       instance.Resource,
			ResourceID:     instance.ResourceID,
			Action:         "instance_rejected",
			Status:         instance.Status,
			Actor:          req.Actor,
			Message:        action.Comment,
			CreatedAt:      now,
		})
		return *instance, nil
	}

	task.Status = "completed"
	next, hasNext := nextWorkflowStep(definition, instance.CurrentStep)
	if hasNext {
		appendWorkflowLog(data, workflowLogFromTask(*instance, *task, "task_approved", req.Actor, now))
		nextTaskID := nextID(data, "workflowTask")
		instance.CurrentTaskID = nextTaskID
		instance.CurrentStep = next.Seq
		instance.CurrentRole = next.RoleCode
		instance.Status = "pending"
		nextTask := workflowTaskFromStep(nextTaskID, *instance, next, now)
		data.WorkflowTasks = append(data.WorkflowTasks, nextTask)
		action.ToStatus = instance.Status
		instance.Actions = append(instance.Actions, action)
		appendWorkflowLog(data, workflowLogFromTask(*instance, nextTask, "task_created", req.Actor, now))
		return *instance, nil
	}

	appendWorkflowLog(data, workflowLogFromTask(*instance, *task, "task_approved", req.Actor, now))
	instance.Status = "approved"
	instance.CurrentTaskID = 0
	instance.CurrentRole = ""
	instance.CompletedAt = now
	action.ToStatus = instance.Status
	instance.Actions = append(instance.Actions, action)
	appendWorkflowLog(data, WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		TriggerEventID: instance.TriggerEventID,
		TaskID:         task.ID,
		TaskNo:         task.TaskNo,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         "instance_approved",
		Status:         instance.Status,
		Actor:          req.Actor,
		Message:        action.Comment,
		CreatedAt:      now,
	})
	return *instance, nil
}

func cancelWorkflowInstance(data *AppData, instanceID int64, actor string, reason string) (WorkflowInstance, error) {
	instanceIndex := findWorkflowInstanceIndex(*data, instanceID)
	if instanceIndex < 0 {
		return WorkflowInstance{}, fmt.Errorf("工作流实例不存在")
	}
	instance := &data.WorkflowInstances[instanceIndex]
	if instance.Status != "pending" {
		return WorkflowInstance{}, fmt.Errorf("只有待处理工作流实例可以取消")
	}
	taskIndex := findCurrentWorkflowTaskIndex(*data, *instance)
	if taskIndex < 0 {
		return WorkflowInstance{}, fmt.Errorf("工作流当前任务不存在")
	}
	task := &data.WorkflowTasks[taskIndex]
	if task.Status != "pending" {
		return WorkflowInstance{}, fmt.Errorf("工作流任务已结束")
	}

	now := nowString()
	comment := strings.TrimSpace(reason)
	task.Status = "cancelled"
	task.CompletedAt = now
	instance.Status = "cancelled"
	instance.CurrentTaskID = 0
	instance.CurrentRole = ""
	instance.UpdatedAt = now
	instance.CompletedAt = now
	instance.Actions = append(instance.Actions, WorkflowAction{
		Seq:        len(instance.Actions) + 1,
		TaskID:     task.ID,
		Step:       task.Step,
		StepCode:   task.StepCode,
		RoleCode:   task.RoleCode,
		Action:     "cancel",
		Actor:      strings.TrimSpace(actor),
		Comment:    comment,
		ActedAt:    now,
		FromStatus: "pending",
		ToStatus:   instance.Status,
	})
	appendWorkflowLog(data, workflowLogFromTask(*instance, *task, "task_cancelled", actor, now))
	appendWorkflowLog(data, WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		TriggerEventID: instance.TriggerEventID,
		TaskID:         task.ID,
		TaskNo:         task.TaskNo,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         "instance_cancelled",
		Status:         instance.Status,
		Actor:          strings.TrimSpace(actor),
		Message:        comment,
		CreatedAt:      now,
	})
	return *instance, nil
}

func reassignWorkflowTask(data *AppData, taskID int64, roleCode string, actor string, reason string) (WorkflowInstance, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return WorkflowInstance{}, fmt.Errorf("改派角色不能为空")
	}
	if !workflowRoleExists(*data, roleCode) {
		return WorkflowInstance{}, fmt.Errorf("改派角色不存在: %s", roleCode)
	}
	for i := range data.WorkflowTasks {
		if data.WorkflowTasks[i].ID != taskID {
			continue
		}
		task := &data.WorkflowTasks[i]
		if task.Status != "pending" {
			return WorkflowInstance{}, fmt.Errorf("只有待处理工作流任务可以改派")
		}
		instanceIndex := findWorkflowInstanceIndex(*data, task.InstanceID)
		if instanceIndex < 0 {
			return WorkflowInstance{}, fmt.Errorf("工作流实例不存在")
		}
		instance := &data.WorkflowInstances[instanceIndex]
		if instance.Status != "pending" || instance.CurrentTaskID != task.ID {
			return WorkflowInstance{}, fmt.Errorf("工作流当前任务不匹配")
		}
		oldRole := task.RoleCode
		if oldRole == roleCode {
			return *instance, nil
		}
		now := nowString()
		comment := strings.TrimSpace(reason)
		task.RoleCode = roleCode
		instance.CurrentRole = roleCode
		instance.UpdatedAt = now
		instance.Actions = append(instance.Actions, WorkflowAction{
			Seq:        len(instance.Actions) + 1,
			TaskID:     task.ID,
			Step:       task.Step,
			StepCode:   task.StepCode,
			RoleCode:   oldRole,
			Action:     "reassign",
			Actor:      strings.TrimSpace(actor),
			Comment:    comment,
			ActedAt:    now,
			FromStatus: "pending",
			ToStatus:   "pending",
		})
		appendWorkflowLog(data, WorkflowLog{
			InstanceID:     instance.ID,
			InstanceNo:     instance.InstanceNo,
			TriggerEventID: instance.TriggerEventID,
			TaskID:         task.ID,
			TaskNo:         task.TaskNo,
			DefinitionCode: instance.DefinitionCode,
			Resource:       instance.Resource,
			ResourceID:     instance.ResourceID,
			Action:         "task_reassigned",
			Status:         task.Status,
			Actor:          strings.TrimSpace(actor),
			Message:        comment,
			Variables: map[string]string{
				"oldRole": oldRole,
				"newRole": roleCode,
			},
			CreatedAt: now,
		})
		return *instance, nil
	}
	return WorkflowInstance{}, fmt.Errorf("工作流任务不存在")
}

func escalateWorkflowTask(data *AppData, taskID int64, roleCode string, actor string, reason string) (WorkflowInstance, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return WorkflowInstance{}, fmt.Errorf("升级角色不能为空")
	}
	if !workflowRoleExists(*data, roleCode) {
		return WorkflowInstance{}, fmt.Errorf("升级角色不存在: %s", roleCode)
	}
	for i := range data.WorkflowTasks {
		if data.WorkflowTasks[i].ID != taskID {
			continue
		}
		task := &data.WorkflowTasks[i]
		if task.Status != "pending" {
			return WorkflowInstance{}, fmt.Errorf("只有待处理工作流任务可以升级")
		}
		if !workflowTaskOverdue(*task, time.Now()) {
			return WorkflowInstance{}, fmt.Errorf("工作流任务尚未超时")
		}
		instanceIndex := findWorkflowInstanceIndex(*data, task.InstanceID)
		if instanceIndex < 0 {
			return WorkflowInstance{}, fmt.Errorf("工作流实例不存在")
		}
		instance := &data.WorkflowInstances[instanceIndex]
		if instance.Status != "pending" || instance.CurrentTaskID != task.ID {
			return WorkflowInstance{}, fmt.Errorf("工作流当前任务不匹配")
		}
		oldRole := task.RoleCode
		now := nowString()
		comment := fallback(strings.TrimSpace(reason), "SLA 超时升级")
		task.RoleCode = roleCode
		task.EscalatedAt = now
		task.EscalatedBy = strings.TrimSpace(actor)
		task.EscalatedFromRole = oldRole
		task.EscalationReason = comment
		instance.CurrentRole = roleCode
		instance.UpdatedAt = now
		instance.Actions = append(instance.Actions, WorkflowAction{
			Seq:        len(instance.Actions) + 1,
			TaskID:     task.ID,
			Step:       task.Step,
			StepCode:   task.StepCode,
			RoleCode:   oldRole,
			Action:     "escalate",
			Actor:      strings.TrimSpace(actor),
			Comment:    comment,
			ActedAt:    now,
			FromStatus: "pending",
			ToStatus:   "pending",
		})
		appendWorkflowLog(data, WorkflowLog{
			InstanceID:     instance.ID,
			InstanceNo:     instance.InstanceNo,
			TriggerEventID: instance.TriggerEventID,
			TaskID:         task.ID,
			TaskNo:         task.TaskNo,
			DefinitionCode: instance.DefinitionCode,
			Resource:       instance.Resource,
			ResourceID:     instance.ResourceID,
			Action:         "task_escalated",
			Status:         task.Status,
			Actor:          strings.TrimSpace(actor),
			Message:        comment,
			Variables: map[string]string{
				"oldRole": oldRole,
				"newRole": roleCode,
				"dueAt":   task.DueAt,
			},
			CreatedAt: now,
		})
		return *instance, nil
	}
	return WorkflowInstance{}, fmt.Errorf("工作流任务不存在")
}

func workflowRoleExists(data AppData, roleCode string) bool {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return false
	}
	for _, role := range data.Roles {
		if strings.TrimSpace(role.Code) == roleCode {
			return true
		}
	}
	return false
}

func workflowTaskFromStep(id int64, instance WorkflowInstance, step WorkflowStep, now string) WorkflowTask {
	return WorkflowTask{
		ID:             id,
		TaskNo:         number("WFT", id),
		InstanceID:     instance.ID,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Step:           step.Seq,
		StepCode:       step.Code,
		StepName:       step.Name,
		RoleCode:       step.RoleCode,
		Action:         step.Action,
		Status:         "pending",
		CreatedAt:      now,
		DueAt:          workflowTaskDueAt(step, now),
	}
}

func workflowTaskDueAt(step WorkflowStep, createdAt string) string {
	if step.SLAHours <= 0 {
		return ""
	}
	base, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(createdAt), time.Local)
	if err != nil {
		base = time.Now()
	}
	return base.Add(time.Duration(step.SLAHours) * time.Hour).Format("2006-01-02 15:04:05")
}

func workflowTaskOverdue(task WorkflowTask, now time.Time) bool {
	if strings.TrimSpace(task.DueAt) == "" {
		return false
	}
	dueAt, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(task.DueAt), time.Local)
	if err != nil {
		return false
	}
	return now.After(dueAt)
}

func workflowInboxItems(data AppData, user User) []WorkflowInboxItem {
	admin := canAccess(data, user, "*")
	now := time.Now()
	items := []WorkflowInboxItem{}
	for _, task := range data.WorkflowTasks {
		if task.Status != "pending" {
			continue
		}
		canAct := admin || task.RoleCode == user.RoleCode
		if !canAct {
			continue
		}
		instanceIndex := findWorkflowInstanceIndex(data, task.InstanceID)
		if instanceIndex < 0 {
			continue
		}
		items = append(items, WorkflowInboxItem{
			Task:     task,
			Instance: data.WorkflowInstances[instanceIndex],
			CanAct:   canAct,
			Overdue:  workflowTaskOverdue(task, now),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Overdue != items[j].Overdue {
			return items[i].Overdue
		}
		leftDue := strings.TrimSpace(items[i].Task.DueAt)
		rightDue := strings.TrimSpace(items[j].Task.DueAt)
		if leftDue != rightDue {
			if leftDue == "" {
				return false
			}
			if rightDue == "" {
				return true
			}
			return leftDue < rightDue
		}
		return items[i].Task.CreatedAt < items[j].Task.CreatedAt
	})
	return items
}

func dueWorkflowTaskEscalationTargets(data AppData, limit int, now time.Time) []workflowTaskEscalationTarget {
	targets := []workflowTaskEscalationTarget{}
	for _, task := range data.WorkflowTasks {
		if task.Status != "pending" || strings.TrimSpace(task.EscalatedAt) != "" {
			continue
		}
		if !workflowTaskOverdue(task, now) {
			continue
		}
		roleCode := automaticWorkflowEscalationRole(data, task.RoleCode)
		if roleCode == "" {
			continue
		}
		targets = append(targets, workflowTaskEscalationTarget{
			TaskID:   task.ID,
			TaskNo:   task.TaskNo,
			FromRole: task.RoleCode,
			ToRole:   roleCode,
		})
		if limit > 0 && len(targets) >= limit {
			return targets
		}
	}
	return targets
}

func automaticWorkflowEscalationRole(data AppData, currentRole string) string {
	currentRole = strings.TrimSpace(currentRole)
	for _, roleCode := range []string{"boss", "company-manager"} {
		if roleCode != currentRole && workflowRoleExists(data, roleCode) {
			return roleCode
		}
	}
	for _, role := range data.Roles {
		if role.Code != currentRole && permissionGranted(role.Permissions, "*") {
			return role.Code
		}
	}
	return ""
}

func appendWorkflowLog(data *AppData, item WorkflowLog) WorkflowLog {
	logID := nextID(data, "workflowLog")
	item.ID = logID
	item.LogNo = number("WFL", logID)
	item.Action = strings.TrimSpace(item.Action)
	item.CreatedAt = fallback(item.CreatedAt, nowString())
	if item.Variables != nil {
		variables := map[string]string{}
		for key, value := range item.Variables {
			if strings.TrimSpace(key) != "" {
				variables[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
		}
		item.Variables = variables
	}
	data.WorkflowLogs = append(data.WorkflowLogs, item)
	appendWorkflowOutbox(data, item)
	return item
}

func appendWorkflowOutbox(data *AppData, log WorkflowLog) WorkflowOutbox {
	outboxID := nextID(data, "workflowOutbox")
	now := fallback(log.CreatedAt, nowString())
	payload := map[string]string{
		"action":         log.Action,
		"status":         log.Status,
		"actor":          log.Actor,
		"message":        log.Message,
		"definitionCode": log.DefinitionCode,
	}
	if log.TriggerEventID > 0 {
		payload["triggerEventId"] = strconv.FormatInt(log.TriggerEventID, 10)
		for _, event := range data.WorkflowEvents {
			if event.ID != log.TriggerEventID {
				continue
			}
			payload["triggerEventNo"] = event.EventNo
			payload["triggerEventType"] = event.EventType
			payload["triggerSource"] = event.Source
			payload["triggerEventKey"] = event.EventKey
			if event.ReplayOfID > 0 {
				payload["triggerReplayOfId"] = strconv.FormatInt(event.ReplayOfID, 10)
			}
			break
		}
	}
	for key, value := range log.Variables {
		if strings.TrimSpace(key) != "" {
			payload["var."+strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	item := WorkflowOutbox{
		ID:             outboxID,
		OutboxNo:       number("WFO", outboxID),
		LogID:          log.ID,
		EventType:      "workflow." + log.Action,
		InstanceID:     log.InstanceID,
		InstanceNo:     log.InstanceNo,
		TriggerEventID: log.TriggerEventID,
		TaskID:         log.TaskID,
		TaskNo:         log.TaskNo,
		DefinitionCode: log.DefinitionCode,
		Resource:       log.Resource,
		ResourceID:     log.ResourceID,
		Status:         "pending",
		Payload:        payload,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	data.WorkflowOutbox = append(data.WorkflowOutbox, item)
	appendWorkflowDeliveriesForOutbox(data, item)
	return item
}

func acknowledgeWorkflowOutbox(data *AppData, id int64, actor string) (WorkflowOutbox, error) {
	for i := range data.WorkflowOutbox {
		if data.WorkflowOutbox[i].ID != id {
			continue
		}
		now := nowString()
		data.WorkflowOutbox[i].Status = "sent"
		data.WorkflowOutbox[i].LastError = ""
		data.WorkflowOutbox[i].NextAttemptAt = ""
		data.WorkflowOutbox[i].AcknowledgedBy = strings.TrimSpace(actor)
		data.WorkflowOutbox[i].AcknowledgedAt = now
		data.WorkflowOutbox[i].UpdatedAt = now
		return data.WorkflowOutbox[i], nil
	}
	return WorkflowOutbox{}, fmt.Errorf("工作流出口事件不存在")
}

func claimWorkflowOutbox(data *AppData, id int64, actor string, consumer string) (WorkflowOutbox, error) {
	for i := range data.WorkflowOutbox {
		if data.WorkflowOutbox[i].ID != id {
			continue
		}
		item := &data.WorkflowOutbox[i]
		switch item.Status {
		case "sent":
			return WorkflowOutbox{}, fmt.Errorf("工作流出口事件已确认")
		case "processing":
			return WorkflowOutbox{}, fmt.Errorf("工作流出口事件已被领取")
		case "failed":
			if !workflowOutboxRetryReady(*item, time.Now()) {
				return WorkflowOutbox{}, fmt.Errorf("工作流出口事件尚未到重试时间")
			}
		case "pending", "":
		default:
			return WorkflowOutbox{}, fmt.Errorf("工作流出口事件状态不可领取: %s", item.Status)
		}
		now := nowString()
		item.Status = "processing"
		item.Attempts++
		item.LastError = ""
		item.NextAttemptAt = ""
		item.ClaimedBy = fallback(strings.TrimSpace(consumer), strings.TrimSpace(actor))
		item.ClaimedAt = now
		item.UpdatedAt = now
		if item.Payload == nil {
			item.Payload = map[string]string{}
		}
		item.Payload["claimedBy"] = item.ClaimedBy
		return *item, nil
	}
	return WorkflowOutbox{}, fmt.Errorf("工作流出口事件不存在")
}

func failWorkflowOutbox(data *AppData, id int64, actor string, errorMessage string, retryAfterMinutes int) (WorkflowOutbox, error) {
	for i := range data.WorkflowOutbox {
		if data.WorkflowOutbox[i].ID != id {
			continue
		}
		item := &data.WorkflowOutbox[i]
		if item.Status == "sent" {
			return WorkflowOutbox{}, fmt.Errorf("工作流出口事件已确认")
		}
		if retryAfterMinutes <= 0 {
			retryAfterMinutes = 5
		}
		now := nowString()
		item.Status = "failed"
		item.LastError = fallback(strings.TrimSpace(errorMessage), "delivery failed")
		item.NextAttemptAt = time.Now().Add(time.Duration(retryAfterMinutes) * time.Minute).Format("2006-01-02 15:04:05")
		if item.ClaimedBy == "" {
			item.ClaimedBy = strings.TrimSpace(actor)
		}
		if item.ClaimedAt == "" {
			item.ClaimedAt = now
		}
		item.UpdatedAt = now
		if item.Payload == nil {
			item.Payload = map[string]string{}
		}
		item.Payload["failedBy"] = strings.TrimSpace(actor)
		item.Payload["retryAfterMinutes"] = strconv.Itoa(retryAfterMinutes)
		return *item, nil
	}
	return WorkflowOutbox{}, fmt.Errorf("工作流出口事件不存在")
}

func resetWorkflowOutbox(data *AppData, id int64, actor string) (WorkflowOutbox, error) {
	for i := range data.WorkflowOutbox {
		if data.WorkflowOutbox[i].ID != id {
			continue
		}
		data.WorkflowOutbox[i].Status = "pending"
		data.WorkflowOutbox[i].Attempts++
		data.WorkflowOutbox[i].LastError = ""
		data.WorkflowOutbox[i].ClaimedBy = ""
		data.WorkflowOutbox[i].ClaimedAt = ""
		data.WorkflowOutbox[i].NextAttemptAt = ""
		data.WorkflowOutbox[i].AcknowledgedBy = ""
		data.WorkflowOutbox[i].AcknowledgedAt = ""
		data.WorkflowOutbox[i].UpdatedAt = nowString()
		if data.WorkflowOutbox[i].Payload == nil {
			data.WorkflowOutbox[i].Payload = map[string]string{}
		}
		data.WorkflowOutbox[i].Payload["resetBy"] = strings.TrimSpace(actor)
		return data.WorkflowOutbox[i], nil
	}
	return WorkflowOutbox{}, fmt.Errorf("工作流出口事件不存在")
}

func workflowOutboxRetryReady(item WorkflowOutbox, now time.Time) bool {
	next := strings.TrimSpace(item.NextAttemptAt)
	if next == "" {
		return true
	}
	nextAttemptAt, err := time.ParseInLocation("2006-01-02 15:04:05", next, time.Local)
	if err != nil {
		return true
	}
	return !now.Before(nextAttemptAt)
}

func saveWorkflowSubscription(data *AppData, item WorkflowSubscription) (WorkflowSubscription, error) {
	now := nowString()
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.EventType = strings.TrimSpace(item.EventType)
	item.Resource = strings.TrimSpace(item.Resource)
	item.DefinitionCode = strings.TrimSpace(item.DefinitionCode)
	item.TargetType = fallback(strings.TrimSpace(item.TargetType), "webhook")
	item.Endpoint = strings.TrimSpace(item.Endpoint)
	item.Status = fallback(strings.TrimSpace(item.Status), "active")
	if item.RetryLimit <= 0 {
		item.RetryLimit = 3
	}
	if item.TimeoutSeconds <= 0 {
		item.TimeoutSeconds = 5
	}
	if item.Code == "" {
		return WorkflowSubscription{}, fmt.Errorf("工作流订阅编码不能为空")
	}
	if item.Name == "" {
		item.Name = item.Code
	}
	if item.EventType == "" {
		item.EventType = "workflow.*"
	}
	if item.Endpoint == "" {
		return WorkflowSubscription{}, fmt.Errorf("工作流订阅端点不能为空")
	}
	if err := validateNoMockEndpoint(item.Endpoint, "工作流订阅端点"); err != nil {
		return WorkflowSubscription{}, err
	}
	for i := range data.WorkflowSubscriptions {
		existing := &data.WorkflowSubscriptions[i]
		if (item.ID > 0 && existing.ID == item.ID) || (item.ID == 0 && existing.Code == item.Code) {
			if item.ID == 0 {
				item.ID = existing.ID
			}
			item.CreatedAt = fallback(existing.CreatedAt, now)
			item.UpdatedAt = now
			data.WorkflowSubscriptions[i] = item
			appendWorkflowDeliveriesForSubscription(data, item)
			return item, nil
		}
	}
	item.ID = nextID(data, "workflowSubscription")
	item.CreatedAt = now
	item.UpdatedAt = now
	data.WorkflowSubscriptions = append(data.WorkflowSubscriptions, item)
	appendWorkflowDeliveriesForSubscription(data, item)
	return item, nil
}

func setWorkflowSubscriptionStatus(data *AppData, id int64, status string) (WorkflowSubscription, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return WorkflowSubscription{}, fmt.Errorf("工作流订阅状态不能为空")
	}
	for i := range data.WorkflowSubscriptions {
		if data.WorkflowSubscriptions[i].ID != id {
			continue
		}
		data.WorkflowSubscriptions[i].Status = status
		data.WorkflowSubscriptions[i].UpdatedAt = nowString()
		appendWorkflowDeliveriesForSubscription(data, data.WorkflowSubscriptions[i])
		return data.WorkflowSubscriptions[i], nil
	}
	return WorkflowSubscription{}, fmt.Errorf("工作流订阅不存在")
}

func appendWorkflowDeliveriesForOutbox(data *AppData, outbox WorkflowOutbox) []WorkflowDelivery {
	deliveries := []WorkflowDelivery{}
	for _, subscription := range data.WorkflowSubscriptions {
		if !workflowSubscriptionMatches(subscription, outbox) {
			continue
		}
		if workflowDeliveryExists(*data, outbox.ID, subscription.ID) {
			continue
		}
		deliveries = append(deliveries, appendWorkflowDelivery(data, outbox, subscription))
	}
	return deliveries
}

func appendWorkflowDeliveriesForSubscription(data *AppData, subscription WorkflowSubscription) []WorkflowDelivery {
	deliveries := []WorkflowDelivery{}
	for _, outbox := range data.WorkflowOutbox {
		if !workflowOutboxCanBackfillDelivery(outbox) {
			continue
		}
		if !workflowSubscriptionMatches(subscription, outbox) {
			continue
		}
		if workflowDeliveryExists(*data, outbox.ID, subscription.ID) {
			continue
		}
		deliveries = append(deliveries, appendWorkflowDelivery(data, outbox, subscription))
	}
	return deliveries
}

func appendWorkflowDelivery(data *AppData, outbox WorkflowOutbox, subscription WorkflowSubscription) WorkflowDelivery {
	deliveryID := nextID(data, "workflowDelivery")
	now := nowString()
	payload := map[string]string{
		"outboxNo":         outbox.OutboxNo,
		"eventType":        outbox.EventType,
		"definitionCode":   outbox.DefinitionCode,
		"resource":         outbox.Resource,
		"resourceId":       strconv.FormatInt(outbox.ResourceID, 10),
		"subscriptionCode": subscription.Code,
		"timeoutSeconds":   strconv.Itoa(subscription.TimeoutSeconds),
	}
	for key, value := range outbox.Payload {
		if strings.TrimSpace(key) != "" {
			payload[key] = value
		}
	}
	item := WorkflowDelivery{
		ID:               deliveryID,
		DeliveryNo:       number("WFD", deliveryID),
		OutboxID:         outbox.ID,
		OutboxNo:         outbox.OutboxNo,
		SubscriptionID:   subscription.ID,
		SubscriptionCode: subscription.Code,
		SubscriptionName: subscription.Name,
		EventType:        outbox.EventType,
		Resource:         outbox.Resource,
		ResourceID:       outbox.ResourceID,
		TargetType:       fallback(subscription.TargetType, "webhook"),
		Endpoint:         subscription.Endpoint,
		Status:           "pending",
		RetryLimit:       subscription.RetryLimit,
		Payload:          payload,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	data.WorkflowDeliveries = append(data.WorkflowDeliveries, item)
	return item
}

func workflowSubscriptionMatches(subscription WorkflowSubscription, outbox WorkflowOutbox) bool {
	if subscription.Status != "active" && subscription.Status != "enabled" {
		return false
	}
	if !workflowPatternMatches(subscription.EventType, outbox.EventType) {
		return false
	}
	if subscription.Resource != "" && subscription.Resource != outbox.Resource {
		return false
	}
	if subscription.DefinitionCode != "" && subscription.DefinitionCode != outbox.DefinitionCode {
		return false
	}
	return true
}

func workflowOutboxCanBackfillDelivery(outbox WorkflowOutbox) bool {
	switch strings.TrimSpace(outbox.Status) {
	case "", "pending", "failed":
		return true
	default:
		return false
	}
}

func workflowPatternMatches(pattern string, value string) bool {
	pattern = strings.TrimSpace(pattern)
	value = strings.TrimSpace(value)
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(value, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == value
}

func workflowDeliveryExists(data AppData, outboxID int64, subscriptionID int64) bool {
	for _, item := range data.WorkflowDeliveries {
		if item.OutboxID == outboxID && item.SubscriptionID == subscriptionID {
			return true
		}
	}
	return false
}

func prepareWorkflowDeliveryDispatch(data *AppData, id int64) (WorkflowDelivery, WorkflowSubscription, error) {
	for i := range data.WorkflowDeliveries {
		if data.WorkflowDeliveries[i].ID != id {
			continue
		}
		item := &data.WorkflowDeliveries[i]
		switch item.Status {
		case "succeeded":
			return WorkflowDelivery{}, WorkflowSubscription{}, fmt.Errorf("工作流投递已完成")
		case "processing":
			return WorkflowDelivery{}, WorkflowSubscription{}, fmt.Errorf("工作流投递正在执行")
		case "failed":
			if !workflowDeliveryRetryReady(*item, time.Now()) {
				return WorkflowDelivery{}, WorkflowSubscription{}, fmt.Errorf("工作流投递尚未到重试时间")
			}
		case "dead":
			return WorkflowDelivery{}, WorkflowSubscription{}, fmt.Errorf("工作流投递已超过重试次数")
		}
		subscription, ok := findWorkflowSubscription(*data, item.SubscriptionID)
		if !ok {
			return WorkflowDelivery{}, WorkflowSubscription{}, fmt.Errorf("工作流订阅不存在")
		}
		now := nowString()
		item.Status = "processing"
		item.Attempts++
		item.LastAttemptAt = now
		item.NextAttemptAt = ""
		item.LastError = ""
		item.UpdatedAt = now
		return *item, subscription, nil
	}
	return WorkflowDelivery{}, WorkflowSubscription{}, fmt.Errorf("工作流投递记录不存在")
}

func finishWorkflowDeliveryDispatch(data *AppData, id int64, requestPayload string, responseStatus int, responseBody string, dispatchErr error) (WorkflowDelivery, error) {
	for i := range data.WorkflowDeliveries {
		if data.WorkflowDeliveries[i].ID != id {
			continue
		}
		item := &data.WorkflowDeliveries[i]
		now := nowString()
		item.RequestPayload = requestPayload
		item.ResponseStatus = responseStatus
		item.ResponseBody = responseBody
		item.UpdatedAt = now
		if dispatchErr == nil && responseStatus >= 200 && responseStatus < 300 {
			item.Status = "succeeded"
			item.LastError = ""
			item.NextAttemptAt = ""
			item.CompletedAt = now
			syncWorkflowOutboxDeliveryStatus(data, item.OutboxID)
			return *item, nil
		}
		item.Status = "failed"
		item.LastError = fallback(errorString(dispatchErr), fmt.Sprintf("HTTP %d", responseStatus))
		if item.RetryLimit > 0 && item.Attempts >= item.RetryLimit {
			item.Status = "dead"
			item.NextAttemptAt = ""
			item.CompletedAt = now
		} else {
			item.NextAttemptAt = time.Now().Add(workflowDeliveryBackoff(item.Attempts)).Format("2006-01-02 15:04:05")
		}
		syncWorkflowOutboxDeliveryStatus(data, item.OutboxID)
		return *item, nil
	}
	return WorkflowDelivery{}, fmt.Errorf("工作流投递记录不存在")
}

func recoverStaleWorkflowDeliveries(data *AppData, timeout time.Duration, now time.Time, actor string) []WorkflowDelivery {
	if timeout <= 0 {
		timeout = workflowDeliveryProcessingTimeout
	}
	recovered := []WorkflowDelivery{}
	outboxIDs := map[int64]bool{}
	for i := range data.WorkflowDeliveries {
		item := &data.WorkflowDeliveries[i]
		if item.Status != "processing" {
			continue
		}
		startedAt, ok := workflowDeliveryAttemptTime(*item)
		if ok && now.Sub(startedAt) < timeout {
			continue
		}
		nowText := now.Format("2006-01-02 15:04:05")
		item.Status = "failed"
		item.LastError = "workflow delivery processing timeout"
		item.NextAttemptAt = nowText
		item.UpdatedAt = nowText
		if item.RetryLimit > 0 && item.Attempts >= item.RetryLimit {
			item.Status = "dead"
			item.NextAttemptAt = ""
			item.CompletedAt = nowText
		}
		if item.Payload == nil {
			item.Payload = map[string]string{}
		}
		item.Payload["recoveredBy"] = strings.TrimSpace(actor)
		item.Payload["recoveredAt"] = nowText
		recovered = append(recovered, *item)
		outboxIDs[item.OutboxID] = true
	}
	for outboxID := range outboxIDs {
		syncWorkflowOutboxDeliveryStatus(data, outboxID)
	}
	return recovered
}

func workflowDeliveryAttemptTime(item WorkflowDelivery) (time.Time, bool) {
	for _, value := range []string{item.LastAttemptAt, item.UpdatedAt, item.CreatedAt} {
		if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func resetWorkflowDelivery(data *AppData, id int64, actor string) (WorkflowDelivery, error) {
	for i := range data.WorkflowDeliveries {
		if data.WorkflowDeliveries[i].ID != id {
			continue
		}
		item := &data.WorkflowDeliveries[i]
		if item.Status == "processing" {
			return WorkflowDelivery{}, fmt.Errorf("工作流投递正在执行")
		}
		item.Status = "pending"
		item.LastError = ""
		item.NextAttemptAt = ""
		item.CompletedAt = ""
		item.RequestPayload = ""
		item.ResponseStatus = 0
		item.ResponseBody = ""
		item.UpdatedAt = nowString()
		if item.Payload == nil {
			item.Payload = map[string]string{}
		}
		item.Payload["resetBy"] = strings.TrimSpace(actor)
		syncWorkflowOutboxDeliveryStatus(data, item.OutboxID)
		return *item, nil
	}
	return WorkflowDelivery{}, fmt.Errorf("工作流投递记录不存在")
}

func workflowDeliveryRetryReady(item WorkflowDelivery, now time.Time) bool {
	next := strings.TrimSpace(item.NextAttemptAt)
	if next == "" {
		return true
	}
	nextAttemptAt, err := time.ParseInLocation("2006-01-02 15:04:05", next, time.Local)
	if err != nil {
		return true
	}
	return !now.Before(nextAttemptAt)
}

func dueWorkflowDeliveryIDs(data AppData, limit int, now time.Time) []int64 {
	ids := []int64{}
	for _, item := range data.WorkflowDeliveries {
		switch item.Status {
		case "", "pending":
		case "failed":
			if !workflowDeliveryRetryReady(item, now) {
				continue
			}
		default:
			continue
		}
		ids = append(ids, item.ID)
		if limit > 0 && len(ids) >= limit {
			return ids
		}
	}
	return ids
}

func workflowDeliveryBackoff(attempts int) time.Duration {
	if attempts <= 0 {
		attempts = 1
	}
	minutes := 5 * attempts
	if minutes > 60 {
		minutes = 60
	}
	return time.Duration(minutes) * time.Minute
}

func findWorkflowSubscription(data AppData, id int64) (WorkflowSubscription, bool) {
	for _, item := range data.WorkflowSubscriptions {
		if item.ID == id {
			return item, true
		}
	}
	return WorkflowSubscription{}, false
}

func syncWorkflowOutboxDeliveryStatus(data *AppData, outboxID int64) {
	if outboxID == 0 {
		return
	}
	total := 0
	succeeded := 0
	failed := 0
	processing := 0
	lastError := ""
	nextAttemptAt := ""
	for _, delivery := range data.WorkflowDeliveries {
		if delivery.OutboxID != outboxID {
			continue
		}
		total++
		switch delivery.Status {
		case "succeeded":
			succeeded++
		case "processing":
			processing++
		case "failed", "dead":
			failed++
			lastError = fallback(delivery.LastError, lastError)
			nextAttemptAt = fallback(delivery.NextAttemptAt, nextAttemptAt)
		}
	}
	if total == 0 {
		return
	}
	for i := range data.WorkflowOutbox {
		if data.WorkflowOutbox[i].ID != outboxID {
			continue
		}
		now := nowString()
		if succeeded == total {
			data.WorkflowOutbox[i].Status = "sent"
			data.WorkflowOutbox[i].AcknowledgedBy = "workflow-delivery"
			data.WorkflowOutbox[i].AcknowledgedAt = now
			data.WorkflowOutbox[i].LastError = ""
			data.WorkflowOutbox[i].NextAttemptAt = ""
		} else if failed > 0 {
			data.WorkflowOutbox[i].Status = "failed"
			data.WorkflowOutbox[i].LastError = lastError
			data.WorkflowOutbox[i].NextAttemptAt = nextAttemptAt
		} else if processing > 0 {
			data.WorkflowOutbox[i].Status = "processing"
			data.WorkflowOutbox[i].LastError = ""
			data.WorkflowOutbox[i].NextAttemptAt = ""
		} else {
			data.WorkflowOutbox[i].Status = "pending"
			data.WorkflowOutbox[i].LastError = ""
			data.WorkflowOutbox[i].NextAttemptAt = ""
		}
		data.WorkflowOutbox[i].UpdatedAt = now
		return
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

func workflowLogFromTask(instance WorkflowInstance, task WorkflowTask, action string, actor string, at string) WorkflowLog {
	return WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		TriggerEventID: instance.TriggerEventID,
		TaskID:         task.ID,
		TaskNo:         task.TaskNo,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         action,
		Status:         task.Status,
		Actor:          strings.TrimSpace(actor),
		Message:        task.StepName,
		CreatedAt:      at,
	}
}

func ensureWorkflowForApprovalTask(data *AppData, task *ApprovalTask) error {
	if task.WorkflowInstanceID != 0 {
		return nil
	}
	ensureWorkflowDefaults(data)
	definition, ok := findWorkflowDefinition(*data, task.FlowCode, false)
	if !ok {
		return fmt.Errorf("工作流定义不存在: %s", task.FlowCode)
	}
	now := fallback(task.UpdatedAt, nowString())
	createdAt := fallback(task.CreatedAt, now)
	instanceID := nextID(data, "workflowInstance")
	instance := WorkflowInstance{
		ID:             instanceID,
		InstanceNo:     number("WF", instanceID),
		DefinitionID:   definition.ID,
		DefinitionCode: task.FlowCode,
		DefinitionName: task.FlowName,
		Category:       workflowCategoryApproval,
		Resource:       task.Resource,
		ResourceID:     task.ResourceID,
		ResourceNo:     task.ResourceNo,
		Title:          task.Title,
		Applicant:      task.Applicant,
		CurrentStep:    task.CurrentStep,
		CurrentRole:    task.CurrentRole,
		Status:         fallback(task.Status, "pending"),
		Reason:         task.Reason,
		CreatedAt:      createdAt,
		UpdatedAt:      now,
		Actions:        workflowActionsFromApprovalActions(task.Actions),
	}
	if instance.Status == "pending" {
		_, ok := workflowStepBySeq(definition.Steps, task.CurrentStep)
		if !ok {
			return fmt.Errorf("工作流当前步骤不存在: %d", task.CurrentStep)
		}
		taskID := nextID(data, "workflowTask")
		instance.CurrentTaskID = taskID
	}
	data.WorkflowInstances = append(data.WorkflowInstances, instance)
	appendWorkflowLog(data, WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         "instance_started",
		Status:         instance.Status,
		Actor:          instance.Applicant,
		Message:        instance.Title,
		Variables:      instance.Variables,
		CreatedAt:      createdAt,
	})
	if instance.Status == "pending" && instance.CurrentTaskID != 0 {
		step, _ := workflowStepBySeq(definition.Steps, task.CurrentStep)
		workflowTask := workflowTaskFromStep(instance.CurrentTaskID, instance, step, now)
		data.WorkflowTasks = append(data.WorkflowTasks, workflowTask)
		appendWorkflowLog(data, workflowLogFromTask(instance, workflowTask, "task_created", task.Applicant, now))
	}
	task.WorkflowInstanceID = instance.ID
	task.WorkflowTaskID = instance.CurrentTaskID
	return nil
}

func syncApprovalTaskFromWorkflow(task *ApprovalTask, instance WorkflowInstance) {
	task.WorkflowInstanceID = instance.ID
	task.WorkflowTaskID = instance.CurrentTaskID
	task.FlowCode = instance.DefinitionCode
	task.FlowName = instance.DefinitionName
	task.Resource = instance.Resource
	task.ResourceID = instance.ResourceID
	task.ResourceNo = instance.ResourceNo
	task.Title = instance.Title
	task.Applicant = instance.Applicant
	task.CurrentStep = instance.CurrentStep
	task.CurrentRole = instance.CurrentRole
	task.Status = instance.Status
	task.Reason = instance.Reason
	task.CreatedAt = fallback(task.CreatedAt, instance.CreatedAt)
	task.UpdatedAt = instance.UpdatedAt
	task.Actions = approvalActionsFromWorkflowActions(instance.Actions)
}

func syncApprovalTaskForWorkflowInstance(data *AppData, instance WorkflowInstance) ApprovalTask {
	for i := range data.ApprovalTasks {
		if data.ApprovalTasks[i].WorkflowInstanceID == instance.ID || (data.ApprovalTasks[i].Resource == instance.Resource && data.ApprovalTasks[i].ResourceID == instance.ResourceID && data.ApprovalTasks[i].Status == "pending") {
			syncApprovalTaskFromWorkflow(&data.ApprovalTasks[i], instance)
			return data.ApprovalTasks[i]
		}
	}
	taskID := nextID(data, "approvalTask")
	task := ApprovalTask{
		ID:      taskID,
		TaskNo:  number("APV", taskID),
		Actions: []ApprovalTaskAction{},
	}
	syncApprovalTaskFromWorkflow(&task, instance)
	data.ApprovalTasks = append(data.ApprovalTasks, task)
	return task
}

func approvalActionsFromWorkflowActions(actions []WorkflowAction) []ApprovalTaskAction {
	out := make([]ApprovalTaskAction, 0, len(actions))
	for _, action := range actions {
		out = append(out, ApprovalTaskAction{
			Seq:      action.Seq,
			Step:     action.Step,
			RoleCode: action.RoleCode,
			Action:   action.Action,
			Actor:    action.Actor,
			Comment:  action.Comment,
			ActedAt:  action.ActedAt,
		})
	}
	return out
}

func workflowActionsFromApprovalActions(actions []ApprovalTaskAction) []WorkflowAction {
	out := make([]WorkflowAction, 0, len(actions))
	for _, action := range actions {
		out = append(out, WorkflowAction{
			Seq:      action.Seq,
			Step:     action.Step,
			RoleCode: action.RoleCode,
			Action:   action.Action,
			Actor:    action.Actor,
			Comment:  action.Comment,
			ActedAt:  action.ActedAt,
		})
	}
	return out
}

func findWorkflowDefinition(data AppData, code string, activeOnly bool) (WorkflowDefinition, bool) {
	code = strings.TrimSpace(code)
	bestIndex := -1
	for i, definition := range data.WorkflowDefinitions {
		if definition.Code != code {
			continue
		}
		if activeOnly && definition.Status != "active" {
			continue
		}
		if bestIndex < 0 || workflowDefinitionVersion(definition) > workflowDefinitionVersion(data.WorkflowDefinitions[bestIndex]) {
			bestIndex = i
		}
	}
	if bestIndex < 0 {
		return WorkflowDefinition{}, false
	}
	definition := data.WorkflowDefinitions[bestIndex]
	definition.Steps = sortedWorkflowSteps(definition.Steps)
	return definition, true
}

func findWorkflowDefinitionByID(data AppData, id int64) (WorkflowDefinition, bool) {
	if id <= 0 {
		return WorkflowDefinition{}, false
	}
	for _, definition := range data.WorkflowDefinitions {
		if definition.ID != id {
			continue
		}
		definition.Steps = sortedWorkflowSteps(definition.Steps)
		return definition, true
	}
	return WorkflowDefinition{}, false
}

func findWorkflowDefinitionForInstance(data AppData, instance WorkflowInstance) (WorkflowDefinition, bool) {
	if definition, ok := findWorkflowDefinitionByID(data, instance.DefinitionID); ok {
		return definition, true
	}
	return findWorkflowDefinition(data, instance.DefinitionCode, false)
}

func workflowDefinitionIndexForCode(data AppData, category string, code string) int {
	category = strings.TrimSpace(category)
	code = strings.TrimSpace(code)
	bestIndex := -1
	for i, definition := range data.WorkflowDefinitions {
		if definition.Category != category || definition.Code != code {
			continue
		}
		if bestIndex < 0 {
			bestIndex = i
			continue
		}
		current := data.WorkflowDefinitions[bestIndex]
		if definition.Status == "active" && current.Status != "active" {
			bestIndex = i
			continue
		}
		if definition.Status == current.Status && workflowDefinitionVersion(definition) > workflowDefinitionVersion(current) {
			bestIndex = i
		}
	}
	return bestIndex
}

func workflowDefinitionVersion(definition WorkflowDefinition) int {
	if definition.Version <= 0 {
		return 1
	}
	return definition.Version
}

func findWorkflowInstanceIndex(data AppData, id int64) int {
	for i, instance := range data.WorkflowInstances {
		if instance.ID == id {
			return i
		}
	}
	return -1
}

func findPendingWorkflowInstance(data AppData, definitionCode string, resource string, resourceID int64) (WorkflowInstance, bool) {
	definitionCode = strings.TrimSpace(definitionCode)
	resource = strings.TrimSpace(resource)
	for _, instance := range data.WorkflowInstances {
		if instance.Status == "pending" && instance.DefinitionCode == definitionCode && instance.Resource == resource && instance.ResourceID == resourceID {
			return instance, true
		}
	}
	return WorkflowInstance{}, false
}

func hasPendingWorkflowForEvent(data AppData, resource string, resourceID int64, eventType string) bool {
	resource = strings.TrimSpace(resource)
	eventType = strings.TrimSpace(eventType)
	for _, instance := range data.WorkflowInstances {
		if instance.Status != "pending" || instance.Resource != resource || instance.ResourceID != resourceID {
			continue
		}
		if eventType == "" {
			return true
		}
		for _, event := range data.WorkflowEvents {
			if event.ID == instance.TriggerEventID && event.EventType == eventType {
				return true
			}
		}
	}
	return false
}

func workflowTriggerEventType(data AppData, instance WorkflowInstance) string {
	for _, event := range data.WorkflowEvents {
		if event.ID == instance.TriggerEventID {
			return event.EventType
		}
	}
	return ""
}

func findCurrentWorkflowTaskIndex(data AppData, instance WorkflowInstance) int {
	if instance.CurrentTaskID > 0 {
		for i, task := range data.WorkflowTasks {
			if task.ID == instance.CurrentTaskID {
				return i
			}
		}
	}
	for i, task := range data.WorkflowTasks {
		if task.InstanceID == instance.ID && task.Status == "pending" {
			return i
		}
	}
	return -1
}

func nextWorkflowStep(definition WorkflowDefinition, currentStep int) (WorkflowStep, bool) {
	for _, step := range sortedWorkflowSteps(definition.Steps) {
		if step.Seq > currentStep {
			return step, true
		}
	}
	return WorkflowStep{}, false
}

func workflowStepBySeq(steps []WorkflowStep, seq int) (WorkflowStep, bool) {
	for _, step := range steps {
		if step.Seq == seq {
			return step, true
		}
	}
	return WorkflowStep{}, false
}

func sortedWorkflowSteps(steps []WorkflowStep) []WorkflowStep {
	out := append([]WorkflowStep{}, steps...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Seq < out[j].Seq
	})
	return out
}

func ensureCounterAtLeast(data *AppData, key string, value int64) {
	if data.Next == nil {
		data.Next = map[string]int64{}
	}
	if data.Next[key] < value {
		data.Next[key] = value
	}
}
