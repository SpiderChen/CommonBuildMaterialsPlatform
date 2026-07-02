package appliance

type WorkflowCatalog struct {
	Events          []WorkflowCatalogEvent       `json:"events"`
	OutboxEvents    []WorkflowCatalogOutboxEvent `json:"outboxEvents"`
	Resources       []WorkflowCatalogResource    `json:"resources"`
	ConditionFields []WorkflowCatalogField       `json:"conditionFields"`
}

type WorkflowCatalogEvent struct {
	Code        string                     `json:"code"`
	Label       string                     `json:"label"`
	Name        string                     `json:"name"`
	Resource    string                     `json:"resource"`
	EventType   string                     `json:"eventType"`
	Source      string                     `json:"source"`
	Description string                     `json:"description"`
	Triggers    []WorkflowCatalogTrigger   `json:"triggers"`
	Integration WorkflowCatalogIntegration `json:"integration"`
	Conditions  []WorkflowCondition        `json:"conditions"`
	Steps       []WorkflowStep             `json:"steps"`
	Variables   []WorkflowCatalogField     `json:"variables"`
}

type WorkflowCatalogTrigger struct {
	Module      string `json:"module"`
	Action      string `json:"action"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Handler     string `json:"handler"`
	Description string `json:"description"`
}

type WorkflowCatalogIntegration struct {
	HasTrigger       bool   `json:"hasTrigger"`
	HasResultHandler bool   `json:"hasResultHandler"`
	ResultPolicy     string `json:"resultPolicy"`
	Status           string `json:"status"`
}

type WorkflowCatalogOutboxEvent struct {
	EventType     string                 `json:"eventType"`
	Label         string                 `json:"label"`
	Description   string                 `json:"description"`
	PayloadFields []WorkflowCatalogField `json:"payloadFields"`
}

type WorkflowCatalogResource struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type WorkflowCatalogField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

func workflowCatalog() WorkflowCatalog {
	resources := []WorkflowCatalogResource{
		{"sales_order", "销售订单"},
		{"inventory_transfer", "库存调拨"},
		{"inventory_stocktake", "库存盘点"},
		{"contract", "客户合同"},
		{"statement", "客户对账单"},
		{"supplier_statement", "供应商对账单"},
		{"red_letter_info", "红字信息表"},
		{"delivery_note", "发货单"},
		{"delivery_sign", "送货签收"},
		{"ticket_void", "磅单作废"},
		{"raw_material_inspection", "原材料检验"},
		{"laboratory_test", "试验检测"},
		{"mix_design", "配合比"},
		{"mix_design_plant_profile", "生产线配合比"},
		{"production_plan", "生产计划"},
		{"quality_exception", "质量异常"},
		{"system_user", "系统用户"},
		{"oidc_provider", "SSO 提供商"},
		{"scim_provider", "SCIM 提供商"},
		{"gateway_route", "网关路由"},
		{"customer_blacklist", "客户黑名单"},
		{"customer_blacklist_release", "黑名单解除"},
		{"plant_buffer_adjustment", "筒仓校正"},
		{"stock_yard_adjustment", "堆场校正"},
	}
	fields := []WorkflowCatalogField{
		{"eventType", "事件类型"},
		{"resource", "业务对象"},
		{"resourceNo", "业务编号"},
		{"actor", "发起人"},
		{"reason", "原因"},
		{"riskFlags", "风险标记"},
		{"riskReasons", "风险原因"},
		{"material", "物料"},
		{"fromSite", "来源站点"},
		{"toSite", "目标站点"},
		{"qty", "数量"},
		{"customer", "客户"},
		{"project", "项目"},
		{"version", "版本"},
		{"severity", "严重程度"},
		{"targetStatus", "目标状态"},
		{"currentStatus", "当前状态"},
		{"code", "编码"},
		{"name", "名称"},
		{"pathPrefix", "网关路径"},
		{"companyId", "公司"},
		{"siteId", "站点"},
	}
	events := []WorkflowCatalogEvent{
		catalogEvent("order_credit_risk", "销售风险审批", "销售订单风险审批", "sales_order", "sales_order.risk_detected", "sales", "订单触发信用、底价等风险时进入审批", []WorkflowCondition{{Field: "riskFlags", Operator: "contains", Value: "credit_limit"}}, catalogSteps("dispatcher", "业务确认", "boss", "管理确认"), "riskFlags", "riskReasons", "customer", "project", "totalAmount"),
		catalogEvent("inventory_transfer", "库存调拨审批", "库存调拨审批", "inventory_transfer", "inventory_transfer.submitted", "inventory", "库存调拨提交后进入调拨复核", nil, catalogSteps("dispatcher", "调拨确认", "boss", "管理确认"), "fromSite", "toSite", "material", "qty"),
		catalogEvent("inventory_stocktake_review", "库存盘点审批", "库存盘点复核", "inventory_stocktake", "inventory_stocktake.review_requested", "inventory", "盘点结果提交后进入复核", nil, catalogSteps("dispatcher", "盘点复核", "boss", "管理确认"), "siteId", "materialId", "actualQty", "varianceQty"),
		catalogEvent("contract_version", "合同版本审批", "客户合同版本审批", "contract", "contract.submitted", "contract", "合同版本提交后审批生效", nil, catalogSteps("dispatcher", "合同明细确认", "boss", "生效确认"), "customerId", "projectId", "version", "amount"),
		catalogEvent("statement_confirm_review", "对账确认审批", "客户对账确认审批", "statement", "statement.confirm_requested", "finance", "客户对账单确认前进入财务复核", nil, catalogSteps("finance", "财务复核", "boss", "管理确认"), "customerId", "projectId", "period", "totalAmount"),
		catalogEvent("supplier_statement_review", "供应商对账审批", "供应商对账审批", "supplier_statement", "supplier_statement.submitted", "finance", "供应商对账单生成后进入审批", nil, catalogSteps("finance", "财务复核", "boss", "管理确认"), "supplierId", "period", "amount"),
		catalogEvent("red_letter_review", "红字信息审批", "红字信息表审批", "red_letter_info", "red_letter_info.requested", "finance", "红字信息表提交后进入审批", nil, catalogSteps("finance", "财务复核", "boss", "管理确认"), "invoiceId", "reason", "amount"),
		catalogEvent("delivery_note_void_review", "送货单作废审批", "送货单作废审批", "delivery_note", "delivery_note.status_change_requested", "delivery", "送货单作废或取消前进入审批", []WorkflowCondition{{Field: "targetStatus", Operator: "equals", Value: "void"}}, catalogSteps("dispatcher", "调度复核", "boss", "作废确认"), "targetStatus", "currentStatus", "noteNo"),
		catalogEvent("ticket_void_review", "磅单作废审批", "磅单作废审批", "ticket_void", "ticket_void.requested", "weighbridge", "磅单作废申请进入复核", nil, catalogSteps("dispatcher", "调度复核", "boss", "作废确认"), "ticketId", "ticketNo", "ticketType", "plateNo"),
		catalogEvent("raw_material_review", "原材料复检审批", "原材料检验复核", "raw_material_inspection", "raw_material_inspection.review_requested", "quality", "原材料检验结果提交后进入复核", nil, catalogSteps("quality", "质检复核", "boss", "质量确认"), "materialId", "supplierId", "result", "qualityStatus"),
		catalogEvent("laboratory_test_review", "试验检测审批", "试验检测复核", "laboratory_test", "laboratory_test.review_requested", "laboratory", "实验室检测结果提交后进入复核", nil, catalogSteps("quality", "检测复核", "boss", "质量确认"), "sampleId", "result", "value", "equipmentId"),
		catalogEvent("mix_design_approval", "配合比审批", "生产配比审批", "mix_design", "mix_design.submitted", "laboratory", "配合比批准前进入审批", nil, catalogSteps("quality", "试配确认", "boss", "生产确认"), "productId", "siteId", "version", "trialRunId"),
		catalogEvent("mix_design_retire", "配合比停用审批", "配合比停用审批", "mix_design", "mix_design.retire_requested", "laboratory", "配合比停用前进入审批", nil, catalogSteps("quality", "停用确认", "boss", "管理确认"), "productId", "siteId", "version"),
		catalogEvent("mix_design_plant_profile", "生产线配比审批", "生产线配比审批", "mix_design_plant_profile", "mix_design_plant_profile.submitted", "laboratory", "生产线配比启用前进入审批", nil, catalogSteps("quality", "生产线试配确认", "boss", "生产确认"), "mixDesignId", "plantId", "effectiveFrom"),
		catalogEvent("mix_design_plant_profile_retire", "生产线配比停用审批", "生产线配比停用审批", "mix_design_plant_profile", "mix_design_plant_profile.retire_requested", "laboratory", "生产线配比停用前进入审批", nil, catalogSteps("quality", "停用确认", "boss", "管理确认"), "mixDesignId", "plantId", "plantCode", "version", "isCurrent"),
		catalogEvent("production_plan_cancel", "生产计划取消审批", "生产计划取消审批", "production_plan", "production_plan.cancel_requested", "production", "生产计划取消前进入审批", nil, catalogSteps("dispatcher", "生产复核", "boss", "取消确认"), "planNo", "reason", "status"),
		catalogEvent("quality_exception_submitted", "质量异常事件", "质量异常事件审批", "quality_exception", "quality_exception.submitted", "laboratory", "质量异常创建后发布事件，可按严重程度进入审批", []WorkflowCondition{{Field: "severity", Operator: "equals", Value: "high"}}, catalogSteps("quality", "质检确认", "boss", "管理确认"), "severity", "sourceType", "sourceNo"),
		catalogEvent("quality_exception_close_review", "质量异常关闭审批", "质量异常关闭审批", "quality_exception", "quality_exception.close_requested", "laboratory", "质量异常闭环关闭前进入审批", []WorkflowCondition{{Field: "targetStatus", Operator: "equals", Value: "closed"}}, catalogSteps("quality", "质检确认", "boss", "闭环确认"), "severity", "rootCause", "correctiveAction", "targetStatus"),
		catalogEvent("delivery_sign_completed", "送货签收事件", "送货签收复核", "delivery_sign", "delivery_sign.completed", "delivery", "送货签收完成后发布事件，可用于交付复核或外部通知", nil, catalogSteps("dispatcher", "签收复核", "", ""), "dispatchId", "ticketId", "orderId", "signedQty"),
		catalogEvent("system_user_status_review", "用户状态审批", "用户状态变更审批", "system_user", "system_user.status_change_requested", "system", "系统用户启停用前进入审批", []WorkflowCondition{{Field: "targetStatus", Operator: "equals", Value: "disabled"}}, catalogSteps("boss", "账号状态复核", "", ""), "targetStatus", "currentStatus", "username", "roleCode"),
		catalogEvent("oidc_provider_status_review", "SSO 状态审批", "SSO 状态变更审批", "oidc_provider", "oidc_provider.status_change_requested", "system", "SSO 提供商启停用前进入审批", []WorkflowCondition{{Field: "targetStatus", Operator: "equals", Value: "disabled"}}, catalogSteps("boss", "SSO 状态复核", "", ""), "targetStatus", "currentStatus", "code", "issuer"),
		catalogEvent("scim_provider_status_review", "SCIM 状态审批", "SCIM 状态变更审批", "scim_provider", "scim_provider.status_change_requested", "system", "SCIM 提供商启停用前进入审批", []WorkflowCondition{{Field: "targetStatus", Operator: "equals", Value: "disabled"}}, catalogSteps("boss", "SCIM 状态复核", "", ""), "targetStatus", "currentStatus", "code", "defaultRoleCode"),
		catalogEvent("gateway_route_status_review", "网关路由审批", "网关路由状态变更审批", "gateway_route", "gateway_route.status_change_requested", "system", "API 网关路由启停用前进入审批", []WorkflowCondition{{Field: "targetStatus", Operator: "equals", Value: "disabled"}}, catalogSteps("boss", "网关路由复核", "", ""), "targetStatus", "currentStatus", "pathPrefix", "stableUpstream"),
		catalogEvent("customer_blacklist_review", "客户黑名单审批", "客户黑名单审批", "customer_blacklist", "customer_blacklist.submitted", "master", "客户加入黑名单前进入审批", nil, catalogSteps("finance", "风险复核", "boss", "管理确认"), "customerId", "reason", "riskLevel"),
		catalogEvent("customer_blacklist_release", "黑名单解除审批", "黑名单解除审批", "customer_blacklist_release", "customer_blacklist_release.requested", "master", "客户黑名单解除前进入审批", nil, catalogSteps("finance", "解除复核", "boss", "管理确认"), "customerId", "reason"),
		catalogEvent("plant_buffer_adjustment", "筒仓校正审批", "筒仓校正审批", "plant_buffer_adjustment", "plant_buffer_adjustment.requested", "production", "筒仓库存校正前进入审批", nil, catalogSteps("dispatcher", "现场复核", "boss", "管理确认"), "actualQty", "moistureRate", "qualityStatus"),
		catalogEvent("stock_yard_adjustment", "堆场校正审批", "堆场校正审批", "stock_yard_adjustment", "stock_yard_adjustment.requested", "inventory", "堆场库存校正前进入审批", nil, catalogSteps("dispatcher", "现场复核", "boss", "管理确认"), "actualQty", "moistureRate", "qualityStatus"),
	}
	for i := range events {
		events[i].Triggers = workflowCatalogTriggers(events[i].EventType)
		events[i].Integration = workflowCatalogIntegration(events[i])
	}
	return WorkflowCatalog{Events: events, OutboxEvents: workflowCatalogOutboxEvents(), Resources: resources, ConditionFields: fields}
}

func workflowCatalogIntegration(event WorkflowCatalogEvent) WorkflowCatalogIntegration {
	policy := workflowCatalogResultPolicy(event)
	integration := WorkflowCatalogIntegration{
		HasTrigger:       len(event.Triggers) > 0,
		HasResultHandler: workflowCatalogHasResultHandler(event),
		ResultPolicy:     policy,
		Status:           "ready",
	}
	if !integration.HasTrigger {
		integration.Status = "missing_trigger"
		return integration
	}
	if policy == "event_only" {
		integration.Status = "event_only"
		return integration
	}
	if !integration.HasResultHandler {
		integration.Status = "missing_result_handler"
	}
	return integration
}

func workflowCatalogResultPolicy(event WorkflowCatalogEvent) string {
	switch event.EventType {
	case "quality_exception.submitted":
		return "event_only"
	default:
		return "write_back"
	}
}

func workflowCatalogHasResultHandler(event WorkflowCatalogEvent) bool {
	_, ok := workflowResultHandlers[event.Resource]
	return ok
}

func catalogEvent(code, label, name, resource, eventType, source, description string, conditions []WorkflowCondition, steps []WorkflowStep, variableKeys ...string) WorkflowCatalogEvent {
	variables := make([]WorkflowCatalogField, 0, len(variableKeys))
	for _, key := range variableKeys {
		variables = append(variables, WorkflowCatalogField{Key: key, Label: key})
	}
	return WorkflowCatalogEvent{
		Code: code, Label: label, Name: name, Resource: resource, EventType: eventType, Source: source,
		Description: description, Conditions: conditions, Steps: steps, Variables: variables,
	}
}

func workflowCatalogOutboxEvents() []WorkflowCatalogOutboxEvent {
	commonFields := []WorkflowCatalogField{
		{Key: "outboxNo", Label: "出口事件编号"},
		{Key: "eventType", Label: "出口事件类型"},
		{Key: "definitionCode", Label: "流程编码"},
		{Key: "resource", Label: "业务对象"},
		{Key: "resourceId", Label: "业务对象 ID"},
		{Key: "triggerEventId", Label: "触发事件 ID"},
		{Key: "triggerEventNo", Label: "触发事件编号"},
		{Key: "triggerEventType", Label: "触发事件类型"},
		{Key: "triggerSource", Label: "触发事件来源"},
		{Key: "triggerEventKey", Label: "触发事件去重键"},
		{Key: "triggerReplayOfId", Label: "重放来源事件 ID"},
		{Key: "action", Label: "动作"},
		{Key: "status", Label: "状态"},
		{Key: "actor", Label: "操作人"},
		{Key: "message", Label: "消息"},
	}
	return []WorkflowCatalogOutboxEvent{
		catalogOutboxEvent("workflow.*", "全部工作流事件", "订阅所有工作流出口事件，适合审计、消息总线或外部 BI 同步", commonFields),
		catalogOutboxEvent("workflow.instance_started", "流程实例启动", "工作流事件命中定义并创建实例后发布", commonFields),
		catalogOutboxEvent("workflow.task_created", "审批任务创建", "每个新审批步骤生成待办时发布", commonFields),
		catalogOutboxEvent("workflow.task_approved", "审批任务通过", "用户通过当前审批任务后发布", commonFields),
		catalogOutboxEvent("workflow.task_rejected", "审批任务驳回", "用户驳回当前审批任务后发布", commonFields),
		catalogOutboxEvent("workflow.instance_approved", "流程通过结束", "最后一个审批步骤通过、实例结束后发布", commonFields),
		catalogOutboxEvent("workflow.instance_rejected", "流程驳回结束", "任一步骤驳回、实例结束后发布", commonFields),
		catalogOutboxEvent("workflow.task_cancelled", "审批任务取消", "流程被取消并关闭当前任务时发布", commonFields),
		catalogOutboxEvent("workflow.instance_cancelled", "流程取消结束", "流程实例取消结束后发布", commonFields),
		catalogOutboxEvent("workflow.task_reassigned", "审批任务改派", "管理员将待办改派到其他角色时发布", commonFields),
		catalogOutboxEvent("workflow.task_escalated", "审批任务升级", "SLA 超时或人工升级待办角色时发布", commonFields),
		catalogOutboxEvent("workflow.result_applied", "审批结果回写", "流程结束并把结果同步到业务对象后发布", commonFields),
		catalogOutboxEvent("workflow.result_failed", "审批结果回写失败", "流程结果同步业务对象失败后发布，实例保持待处理以便重试", commonFields),
	}
}

func catalogOutboxEvent(eventType, label, description string, fields []WorkflowCatalogField) WorkflowCatalogOutboxEvent {
	return WorkflowCatalogOutboxEvent{EventType: eventType, Label: label, Description: description, PayloadFields: fields}
}

func workflowCatalogTriggers(eventType string) []WorkflowCatalogTrigger {
	switch eventType {
	case "sales_order.risk_detected":
		return []WorkflowCatalogTrigger{catalogTrigger("销售订单", "创建订单并命中信用或底价风险", "POST", "/api/orders", "App.orders", "订单创建时自动做风险检查，命中后发布销售风险事件")}
	case "inventory_transfer.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("堆场管理", "创建库存调拨", "POST", "/api/procurement/transfers", "App.createInventoryTransfer", "调拨单创建后进入调拨复核")}
	case "inventory_stocktake.review_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("堆场管理", "提交库存盘点复核", "POST", "/api/procurement/stocktakes/{id}/review", "App.reviewInventoryStocktake", "盘点结果提交后发布复核事件")}
	case "contract.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("合同", "提交合同版本", "POST", "/api/contracts/{id}/submit", "App.submitContractVersion", "合同版本提交后进入生效审批")}
	case "statement.confirm_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("财务对账", "确认客户对账单", "POST", "/api/statements/{id}/confirm", "App.confirmStatement", "客户或财务确认对账单前发布确认审批事件")}
	case "supplier_statement.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("采购结算", "生成供应商对账单", "POST", "/api/finance/supplier-statements", "App.createSupplierStatement", "供应商对账单生成后发布审批事件")}
	case "red_letter_info.requested":
		return []WorkflowCatalogTrigger{catalogTrigger("财务发票", "申请红字信息表", "POST", "/api/finance/red-letter", "App.createRedLetterInfo", "红字信息表申请后发布财务复核事件")}
	case "delivery_note.status_change_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("交付", "作废或取消送货单", "POST", "/api/delivery/notes/{id}/status", "App.updateDeliveryNoteStatus", "送货单作废或取消前发布状态变更审批事件")}
	case "delivery_sign.completed":
		return []WorkflowCatalogTrigger{
			catalogTrigger("交付", "司机或调度录入签收", "POST", "/api/delivery/sign", "App.signDelivery", "签收完成后发布交付事件"),
			catalogTrigger("交付", "公开签收链接提交", "POST", "/api/public/delivery-sign/{token}", "App.signDeliveryByToken", "外部签收完成后发布交付事件"),
		}
	case "ticket_void.requested":
		return []WorkflowCatalogTrigger{catalogTrigger("过磅", "申请磅单作废", "POST", "/api/tickets/{id}/void", "App.requestTicketVoid", "磅单作废申请后发布复核事件")}
	case "raw_material_inspection.review_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("质量检验", "提交原材料检验复核", "POST", "/api/quality/raw-inspections/{id}/review", "App.reviewRawMaterialInspection", "原材料检验结果提交后发布复核事件")}
	case "laboratory_test.review_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "提交试验检测复核", "POST", "/api/laboratory/tests/{id}/review", "App.reviewLaboratoryTest", "试验检测结果提交后发布复核事件")}
	case "mix_design.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "提交配合比审批", "POST", "/api/laboratory/mix-designs/{id}/approve", "App.approveLaboratoryMixDesign", "配合比批准前发布审批事件")}
	case "mix_design.retire_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "申请停用配合比", "POST", "/api/laboratory/mix-designs/{id}/retire", "App.retireLaboratoryMixDesign", "配合比停用前发布审批事件")}
	case "mix_design_plant_profile.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "提交生产线配比审批", "POST", "/api/laboratory/mix-design-plant-profiles/{id}/approve", "App.approveMixDesignPlantProfile", "生产线配比启用前发布审批事件")}
	case "mix_design_plant_profile.retire_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "申请停用生产线配比", "POST", "/api/laboratory/mix-design-plant-profiles/{id}/retire", "App.retireMixDesignPlantProfile", "生产线配比停用前发布审批事件")}
	case "production_plan.cancel_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("生产", "申请取消生产计划", "POST", "/api/production-plans/{id}/cancel", "App.cancelProductionPlan", "生产计划取消前发布审批事件")}
	case "quality_exception.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "创建质量异常", "POST", "/api/laboratory/exceptions", "App.createLaboratoryException", "质量异常创建后发布事件")}
	case "quality_exception.close_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("实验室", "申请关闭质量异常", "POST", "/api/laboratory/exceptions/{id}/handle", "App.handleLaboratoryException", "质量异常闭环关闭前发布审批事件")}
	case "system_user.status_change_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("系统", "启停用系统用户", "POST", "/api/system/users/{id}/status", "App.systemUsers", "用户状态变更前发布审批事件")}
	case "oidc_provider.status_change_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("系统", "启停用 SSO 提供商", "POST", "/api/system/oidc/providers/{id}/status", "App.systemOIDCProviders", "SSO 提供商状态变更前发布审批事件")}
	case "scim_provider.status_change_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("系统", "启停用 SCIM 提供商", "POST", "/api/system/scim/providers/{id}/status", "App.systemSCIMProviders", "SCIM 提供商状态变更前发布审批事件")}
	case "gateway_route.status_change_requested":
		return []WorkflowCatalogTrigger{catalogTrigger("系统", "启停用网关路由", "POST", "/api/system/gateway/routes/{id}/status", "App.systemGatewayRoutes", "网关路由状态变更前发布审批事件")}
	case "customer_blacklist.submitted":
		return []WorkflowCatalogTrigger{catalogTrigger("客户风控", "提交客户黑名单", "POST", "/api/customers/blacklist", "App.createCustomerBlacklist", "客户加入黑名单前发布风控审批事件")}
	case "customer_blacklist_release.requested":
		return []WorkflowCatalogTrigger{catalogTrigger("客户风控", "申请解除黑名单", "POST", "/api/customers/blacklist/{id}/release", "App.requestCustomerBlacklistRelease", "客户黑名单解除前发布审批事件")}
	case "plant_buffer_adjustment.requested":
		return []WorkflowCatalogTrigger{catalogTrigger("生产", "提交筒仓盘点校正", "POST", "/api/production-plans/buffers/{id}/adjust", "App.adjustPlantBuffer", "筒仓库存校正前发布审批事件")}
	case "stock_yard_adjustment.requested":
		return []WorkflowCatalogTrigger{catalogTrigger("堆场管理", "提交堆位盘点校正", "POST", "/api/procurement/stock-yards/piles/{id}/adjust", "App.adjustStockYardPile", "堆场库存校正前发布审批事件")}
	default:
		return nil
	}
}

func catalogTrigger(module, action, method, path, handler, description string) WorkflowCatalogTrigger {
	return WorkflowCatalogTrigger{Module: module, Action: action, Method: method, Path: path, Handler: handler, Description: description}
}

func catalogSteps(firstRole, firstName, secondRole, secondName string) []WorkflowStep {
	steps := []WorkflowStep{}
	if firstRole != "" {
		steps = append(steps, WorkflowStep{Seq: 1, Name: firstName, Type: workflowCategoryApproval, RoleCode: firstRole, Action: "approve", SLAHours: 24})
	}
	if secondRole != "" {
		steps = append(steps, WorkflowStep{Seq: 2, Name: secondName, Type: workflowCategoryApproval, RoleCode: secondRole, Action: "approve", SLAHours: 24})
	}
	return steps
}
