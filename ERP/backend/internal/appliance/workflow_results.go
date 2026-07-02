package appliance

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	workflowResultAppliedAction = "result_applied"
	workflowResultFailedAction  = "result_failed"
)

type workflowResultHandler func(data *AppData, instance WorkflowInstance) error

var workflowResultHandlers = map[string]workflowResultHandler{
	"sales_order":                applySalesOrderWorkflowResult,
	"inventory_transfer":         applyInventoryTransferWorkflowResult,
	"inventory_stocktake":        applyInventoryStocktakeWorkflowResult,
	"contract":                   applyContractWorkflowResult,
	"statement":                  applyCustomerStatementWorkflowResult,
	"supplier_statement":         applySupplierStatementWorkflowResult,
	"red_letter_info":            applyRedLetterInfoWorkflowResult,
	"delivery_note":              applyDeliveryNoteWorkflowResult,
	"delivery_sign":              applyDeliverySignWorkflowResult,
	"ticket_void":                applyTicketVoidWorkflowResult,
	"quality_exception":          applyQualityExceptionWorkflowResult,
	"raw_material_inspection":    applyRawMaterialInspectionWorkflowResult,
	"laboratory_test":            applyLaboratoryTestWorkflowResult,
	"mix_design":                 applyMixDesignWorkflowResult,
	"mix_design_plant_profile":   applyMixDesignPlantProfileWorkflowResult,
	"production_plan":            applyProductionPlanWorkflowResult,
	"system_user":                applySystemUserWorkflowResult,
	"oidc_provider":              applyOIDCProviderWorkflowResult,
	"scim_provider":              applySCIMProviderWorkflowResult,
	"gateway_route":              applyGatewayRouteWorkflowResult,
	"customer_blacklist":         applyCustomerBlacklistWorkflowResult,
	"customer_blacklist_release": applyCustomerBlacklistReleaseWorkflowResult,
	"plant_buffer_adjustment":    applyPlantBufferAdjustmentWorkflowResult,
	"stock_yard_adjustment":      applyStockYardAdjustmentWorkflowResult,
}

func applyWorkflowResult(data *AppData, instance WorkflowInstance) error {
	if !workflowResultStatus(instance.Status) {
		return nil
	}
	resource := strings.TrimSpace(instance.Resource)
	if resource == "quality_exception" && workflowTriggerEventType(*data, instance) != "quality_exception.close_requested" {
		return nil
	}
	handler, ok := workflowResultHandlers[resource]
	if !ok {
		return nil
	}
	if workflowResultAlreadyApplied(*data, instance) {
		return nil
	}
	if err := handler(data, instance); err != nil {
		return err
	}
	appendWorkflowLog(data, WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		TriggerEventID: instance.TriggerEventID,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         workflowResultAppliedAction,
		Status:         instance.Status,
		Actor:          lastWorkflowActor(instance),
		Message:        "workflow result synced to business resource",
		Variables: map[string]string{
			"resource":   resource,
			"resourceId": strconv.FormatInt(instance.ResourceID, 10),
			"resourceNo": instance.ResourceNo,
			"status":     instance.Status,
		},
		CreatedAt: nowString(),
	})
	return nil
}

func appendWorkflowResultFailureLog(data *AppData, instance WorkflowInstance, err error) WorkflowLog {
	taskID := instance.CurrentTaskID
	if taskID == 0 && len(instance.Actions) > 0 {
		taskID = instance.Actions[len(instance.Actions)-1].TaskID
	}
	taskNo := ""
	for _, task := range data.WorkflowTasks {
		if task.ID == taskID {
			taskNo = task.TaskNo
			break
		}
	}
	resourceID := ""
	if instance.ResourceID != 0 {
		resourceID = strconv.FormatInt(instance.ResourceID, 10)
	}
	return appendWorkflowLog(data, WorkflowLog{
		InstanceID:     instance.ID,
		InstanceNo:     instance.InstanceNo,
		TriggerEventID: instance.TriggerEventID,
		TaskID:         taskID,
		TaskNo:         taskNo,
		DefinitionCode: instance.DefinitionCode,
		Resource:       instance.Resource,
		ResourceID:     instance.ResourceID,
		Action:         workflowResultFailedAction,
		Status:         "failed",
		Actor:          lastWorkflowActor(instance),
		Message:        errorString(err),
		Variables: map[string]string{
			"resource":       strings.TrimSpace(instance.Resource),
			"resourceId":     resourceID,
			"resourceNo":     instance.ResourceNo,
			"intendedStatus": instance.Status,
			"error":          errorString(err),
		},
		CreatedAt: nowString(),
	})
}

func workflowResultStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "approved", "rejected", "cancelled":
		return true
	default:
		return false
	}
}

func workflowResultAlreadyApplied(data AppData, instance WorkflowInstance) bool {
	for _, log := range data.WorkflowLogs {
		if log.Action != workflowResultAppliedAction {
			continue
		}
		if instance.ID > 0 && log.InstanceID == instance.ID && log.Status == instance.Status {
			return true
		}
		if instance.ID == 0 && log.Resource == instance.Resource && log.ResourceID == instance.ResourceID && log.Status == instance.Status {
			return true
		}
	}
	return false
}

func lastWorkflowActor(instance WorkflowInstance) string {
	for i := len(instance.Actions) - 1; i >= 0; i-- {
		if actor := strings.TrimSpace(instance.Actions[i].Actor); actor != "" {
			return actor
		}
	}
	return strings.TrimSpace(instance.Applicant)
}

func workflowInstanceFromApprovalTask(task ApprovalTask) WorkflowInstance {
	return WorkflowInstance{
		ID:             task.WorkflowInstanceID,
		InstanceNo:     task.TaskNo,
		DefinitionCode: task.FlowCode,
		DefinitionName: task.FlowName,
		Category:       workflowCategoryApproval,
		Resource:       task.Resource,
		ResourceID:     task.ResourceID,
		ResourceNo:     task.ResourceNo,
		Title:          task.Title,
		Applicant:      task.Applicant,
		CurrentTaskID:  task.WorkflowTaskID,
		CurrentStep:    task.CurrentStep,
		CurrentRole:    task.CurrentRole,
		Status:         task.Status,
		Reason:         task.Reason,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
		Actions:        workflowActionsFromApprovalActions(task.Actions),
	}
}

func applyApprovalResult(data *AppData, task ApprovalTask) error {
	if task.WorkflowInstanceID != 0 {
		if index := findWorkflowInstanceIndex(*data, task.WorkflowInstanceID); index >= 0 {
			return applyWorkflowResult(data, data.WorkflowInstances[index])
		}
	}
	return applyWorkflowResult(data, workflowInstanceFromApprovalTask(task))
}

func applySalesOrderWorkflowResult(data *AppData, instance WorkflowInstance) error {
	for i := range data.Orders {
		if data.Orders[i].ID != instance.ResourceID {
			continue
		}
		switch instance.Status {
		case "approved":
			data.Orders[i].Status = "submitted"
		case "rejected", "cancelled":
			data.Orders[i].Status = instance.Status
		}
		return nil
	}
	return fmt.Errorf("工作流关联订单不存在")
}

func applyInventoryTransferWorkflowResult(data *AppData, instance WorkflowInstance) error {
	for i := range data.InventoryTransfers {
		if data.InventoryTransfers[i].ID != instance.ResourceID {
			continue
		}
		switch instance.Status {
		case "approved":
			data.InventoryTransfers[i].Status = "approved"
		case "rejected", "cancelled":
			data.InventoryTransfers[i].Status = instance.Status
		}
		return nil
	}
	return fmt.Errorf("工作流关联调拨单不存在")
}

func applyInventoryStocktakeWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyInventoryStocktakeReviewLocked(data, instance.ResourceID)
		return err
	case "rejected", "cancelled":
		index := inventoryStocktakeIndex(*data, instance.ResourceID)
		if index < 0 {
			return fmt.Errorf("工作流关联盘点单不存在")
		}
		data.InventoryStocktakes[index].Status = instance.Status
		return nil
	default:
		return nil
	}
}

func applyContractWorkflowResult(data *AppData, instance WorkflowInstance) error {
	index := contractIndex(*data, instance.ResourceID)
	if index < 0 {
		return fmt.Errorf("工作流关联合同不存在")
	}
	switch instance.Status {
	case "rejected", "cancelled":
		data.Contracts[index].Status = instance.Status
		return nil
	case "approved":
	default:
		return nil
	}
	contractNo := data.Contracts[index].ContractNo
	for i := range data.Contracts {
		if i != index && data.Contracts[i].ContractNo == contractNo && data.Contracts[i].Status == "active" {
			data.Contracts[i].Status = "superseded"
		}
	}
	data.Contracts[index].Status = "active"
	data.Contracts[index].ApprovedAt = nowString()
	data.Contracts[index].ApprovedBy = lastWorkflowActor(instance)
	if data.Contracts[index].Version <= 0 {
		data.Contracts[index].Version = 1
	}
	return nil
}

func applyCustomerStatementWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := confirmStatementLocked(data, instance.ResourceID, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		for i := range data.Statements {
			if data.Statements[i].ID == instance.ResourceID {
				data.Statements[i].Status = instance.Status
				return nil
			}
		}
		return fmt.Errorf("工作流关联客户对账单不存在")
	default:
		return nil
	}
}

func applySupplierStatementWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := approveSupplierStatementLocked(data, instance.ResourceID, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		for i := range data.SupplierStatements {
			if data.SupplierStatements[i].ID == instance.ResourceID {
				data.SupplierStatements[i].Status = instance.Status
				return nil
			}
		}
		return fmt.Errorf("工作流关联供应商对账单不存在")
	default:
		return nil
	}
}

func applyRedLetterInfoWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := approveRedLetterInfoLocked(data, instance.ResourceID, lastWorkflowActor(instance), "", "")
		return err
	case "rejected", "cancelled":
		for i := range data.RedLetterInfos {
			if data.RedLetterInfos[i].ID == instance.ResourceID {
				data.RedLetterInfos[i].Status = instance.Status
				return nil
			}
		}
		return fmt.Errorf("工作流关联红字信息表不存在")
	default:
		return nil
	}
}

func applyDeliveryNoteWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		targetStatus := normalizeDeliveryNoteStatus(instance.Variables["targetStatus"])
		_, err := applyDeliveryNoteStatusLocked(data, instance.ResourceID, targetStatus)
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyDeliverySignWorkflowResult(data *AppData, instance WorkflowInstance) error {
	for i := range data.DeliverySigns {
		if data.DeliverySigns[i].ID != instance.ResourceID {
			continue
		}
		switch instance.Status {
		case "approved", "rejected", "cancelled":
			data.DeliverySigns[i].ReviewStatus = instance.Status
			data.DeliverySigns[i].ReviewedBy = lastWorkflowActor(instance)
			data.DeliverySigns[i].ReviewedAt = nowString()
			return nil
		default:
			return nil
		}
	}
	return fmt.Errorf("工作流关联签收记录不存在")
}

func applyTicketVoidWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyTicketVoidDecisionLocked(data, instance.ResourceID, true, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		_, err := applyTicketVoidDecisionLocked(data, instance.ResourceID, false, lastWorkflowActor(instance))
		return err
	default:
		return nil
	}
}

func applyRawMaterialInspectionWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		req := RawMaterialInspection{
			Moisture:   optionalWorkflowFloat(instance, "moisture"),
			MudContent: optionalWorkflowFloat(instance, "mudContent"),
			Fineness:   strings.TrimSpace(instance.Variables["fineness"]),
			Result:     strings.TrimSpace(instance.Variables["result"]),
			Remark:     fallback(strings.TrimSpace(instance.Variables["remark"]), instance.Reason),
		}
		_, err := applyRawMaterialInspectionReviewLocked(data, instance.ResourceID, req, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyLaboratoryTestWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		req := LaboratoryTestRecord{
			Result:   strings.TrimSpace(instance.Variables["result"]),
			Reviewer: strings.TrimSpace(instance.Variables["reviewer"]),
			Remark:   strings.TrimSpace(instance.Variables["remark"]),
		}
		if value, err := strconv.ParseFloat(strings.TrimSpace(instance.Variables["value"]), 64); err == nil {
			req.Value = value
		}
		_, err := applyLaboratoryTestReviewLocked(data, instance.ResourceID, req, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		index := laboratoryTestIndex(*data, instance.ResourceID)
		if index < 0 {
			return fmt.Errorf("工作流关联试验记录不存在")
		}
		data.LaboratoryTests[index].Status = instance.Status
		return nil
	default:
		return nil
	}
}

func applyMixDesignWorkflowResult(data *AppData, instance WorkflowInstance) error {
	action := strings.TrimSpace(instance.Variables["workflowAction"])
	switch instance.Status {
	case "approved":
		if action == "retire" {
			_, err := applyMixDesignRetireLocked(data, instance.ResourceID)
			return err
		}
		req := mixDesignApprovalRequest{
			TrialRunID:    optionalWorkflowInt(instance, "trialRunId"),
			EffectiveFrom: strings.TrimSpace(instance.Variables["effectiveFrom"]),
			EffectiveTo:   strings.TrimSpace(instance.Variables["effectiveTo"]),
		}
		_, err := applyMixDesignApprovalLocked(data, instance.ResourceID, req, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		if action == "retire" {
			return nil
		}
		for i := range data.MixDesigns {
			if data.MixDesigns[i].ID == instance.ResourceID {
				data.MixDesigns[i].Status = instance.Status
				data.MixDesigns[i].UpdatedAt = nowString()
				return nil
			}
		}
		return fmt.Errorf("工作流关联配比不存在")
	default:
		return nil
	}
}

func applyMixDesignPlantProfileWorkflowResult(data *AppData, instance WorkflowInstance) error {
	action := strings.TrimSpace(instance.Variables["workflowAction"])
	switch instance.Status {
	case "approved":
		if action == "retire" {
			_, err := applyMixDesignPlantProfileRetireLocked(data, instance.ResourceID)
			return err
		}
		req := mixDesignPlantProfileApprovalRequest{
			EffectiveFrom: strings.TrimSpace(instance.Variables["effectiveFrom"]),
			EffectiveTo:   strings.TrimSpace(instance.Variables["effectiveTo"]),
		}
		_, err := applyMixDesignPlantProfileApprovalLocked(data, instance.ResourceID, req, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		if action == "retire" {
			return nil
		}
		for i := range data.MixDesignPlantProfiles {
			if data.MixDesignPlantProfiles[i].ID == instance.ResourceID {
				data.MixDesignPlantProfiles[i].Status = instance.Status
				data.MixDesignPlantProfiles[i].IsCurrent = false
				data.MixDesignPlantProfiles[i].UpdatedAt = nowString()
				return nil
			}
		}
		return fmt.Errorf("工作流关联生产线配比不存在")
	default:
		return nil
	}
}

func applyProductionPlanWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyProductionPlanCancelLocked(data, instance.ResourceID)
		return err
	case "rejected", "cancelled":
		index := productionPlanIndex(*data, instance.ResourceID)
		if index < 0 {
			return fmt.Errorf("工作流关联生产计划不存在")
		}
		if data.ProductionPlans[index].Status == "pending_approval" {
			data.ProductionPlans[index].Status = normalizeProductionPlanStatus(*data, data.ProductionPlans[index])
			data.ProductionPlans[index].UpdatedAt = nowString()
		}
		return nil
	default:
		return nil
	}
}

func applySystemUserWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applySystemUserStatusLocked(data, instance.ResourceID, instance.Variables["targetStatus"])
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyOIDCProviderWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyOIDCProviderStatusLocked(data, instance.ResourceID, instance.Variables["targetStatus"])
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applySCIMProviderWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applySCIMProviderStatusLocked(data, instance.ResourceID, instance.Variables["targetStatus"])
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyGatewayRouteWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyGatewayRouteStatusLocked(data, instance.ResourceID, instance.Variables["targetStatus"], lastWorkflowActor(instance), "")
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyQualityExceptionWorkflowResult(data *AppData, instance WorkflowInstance) error {
	if workflowTriggerEventType(*data, instance) != "quality_exception.close_requested" {
		return nil
	}
	switch instance.Status {
	case "approved":
		_, err := applyQualityExceptionCloseLocked(data, instance.ResourceID, QualityException{
			RootCause:        instance.Variables["rootCause"],
			CorrectiveAction: instance.Variables["correctiveAction"],
			Responsible:      instance.Variables["responsible"],
		}, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyCustomerBlacklistReleaseWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyCustomerBlacklistReleaseLocked(data, instance.ResourceID)
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyCustomerBlacklistWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		_, err := applyCustomerBlacklistActivationLocked(data, instance.ResourceID)
		return err
	case "rejected", "cancelled":
		for i := range data.CustomerBlacklists {
			if data.CustomerBlacklists[i].ID == instance.ResourceID {
				data.CustomerBlacklists[i].Status = instance.Status
				return nil
			}
		}
		return fmt.Errorf("工作流关联客户黑名单不存在")
	default:
		return nil
	}
}

func applyPlantBufferAdjustmentWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		actual, err := requiredWorkflowFloat(instance, "actualQty")
		if err != nil {
			return err
		}
		req := plantBufferAdjustmentRequest{
			BufferID:      instance.ResourceID,
			BufferCode:    instance.ResourceNo,
			ActualQty:     actual,
			MoistureRate:  optionalWorkflowFloat(instance, "moistureRate"),
			QualityStatus: strings.TrimSpace(instance.Variables["qualityStatus"]),
			Status:        strings.TrimSpace(instance.Variables["status"]),
			Remark:        fallback(strings.TrimSpace(instance.Variables["remark"]), instance.Reason),
		}
		_, err = applyPlantBufferAdjustmentLocked(data, instance.ResourceID, req, instance.ID, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func applyStockYardAdjustmentWorkflowResult(data *AppData, instance WorkflowInstance) error {
	switch instance.Status {
	case "approved":
		actual, err := requiredWorkflowFloat(instance, "actualQty")
		if err != nil {
			return err
		}
		req := stockYardAdjustmentRequest{
			PileID:        instance.ResourceID,
			PileCode:      instance.ResourceNo,
			ActualQty:     actual,
			MoistureRate:  optionalWorkflowFloat(instance, "moistureRate"),
			QualityStatus: strings.TrimSpace(instance.Variables["qualityStatus"]),
			Status:        strings.TrimSpace(instance.Variables["status"]),
			Remark:        fallback(strings.TrimSpace(instance.Variables["remark"]), instance.Reason),
		}
		_, err = applyStockYardAdjustmentLocked(data, instance.ResourceID, req, instance.ID, lastWorkflowActor(instance))
		return err
	case "rejected", "cancelled":
		return nil
	default:
		return nil
	}
}

func requiredWorkflowFloat(instance WorkflowInstance, key string) (float64, error) {
	value := strings.TrimSpace(instance.Variables[key])
	if value == "" {
		return 0, fmt.Errorf("工作流变量缺少 %s", key)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("工作流变量 %s 不合法", key)
	}
	return parsed, nil
}

func optionalWorkflowFloat(instance WorkflowInstance, key string) float64 {
	value := strings.TrimSpace(instance.Variables[key])
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func optionalWorkflowInt(instance WorkflowInstance, key string) int64 {
	value := strings.TrimSpace(instance.Variables[key])
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
