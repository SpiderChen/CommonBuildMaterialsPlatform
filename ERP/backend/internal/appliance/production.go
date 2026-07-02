package appliance

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (a *App) production(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		if r.Method == http.MethodGet {
			data := scopedData(a.mustSnapshot(), session.User)
			writeJSON(w, http.StatusOK, enrichedProductionPlans(data))
			return
		}
		if r.Method == http.MethodPost {
			a.createProductionPlan(w, r, session)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if len(parts) == 3 && parts[0] == "protocols" && parts[1] == "plant" && parts[2] == "ingest" && r.Method == http.MethodPost {
		a.ingestPlantProtocolFrame(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "protocols" && parts[1] == "buffer" && parts[2] == "ingest" && r.Method == http.MethodPost {
		a.ingestBufferProtocolFrame(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "protocols" && parts[1] == "yard" && parts[2] == "ingest" && r.Method == http.MethodPost {
		a.ingestYardProtocolFrame(w, r, session)
		return
	}
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		if len(parts) == 1 {
			if id, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				plan, ok := findProductionPlan(data, id)
				if !ok {
					writeError(w, http.StatusNotFound, "生产计划不存在")
					return
				}
				writeJSON(w, http.StatusOK, enrichProductionPlan(data, plan))
				return
			}
		}
		switch parts[0] {
		case "overview":
			writeJSON(w, http.StatusOK, productionPayload(data))
		case "tasks":
			writeJSON(w, http.StatusOK, data.ProductionTasks)
		case "batches":
			writeJSON(w, http.StatusOK, data.ProductionBatches)
		case "reports":
			writeJSON(w, http.StatusOK, data.ProductionReports)
		case "mix-designs":
			writeJSON(w, http.StatusOK, data.MixDesigns)
		case "mix-design-plant-profiles":
			writeJSON(w, http.StatusOK, data.MixDesignPlantProfiles)
		case "buffer-balances", "plant-buffer-locations":
			writeJSON(w, http.StatusOK, data.PlantBufferLocations)
		case "buffer-flows":
			writeJSON(w, http.StatusOK, data.PlantBufferFlows)
		case "stock-yards":
			writeJSON(w, http.StatusOK, data.StockYards)
		case "yard-balances", "stock-yard-piles":
			writeJSON(w, http.StatusOK, data.StockYardPiles)
		case "yard-flows", "stock-yard-flows":
			writeJSON(w, http.StatusOK, data.StockYardFlows)
		default:
			writeError(w, http.StatusNotFound, "unknown production resource")
		}
		return
	}
	if len(parts) == 1 && parts[0] == "buffer-transfers" && r.Method == http.MethodPost {
		a.createPlantBufferTransfer(w, r, session)
		return
	}
	if len(parts) == 1 && parts[0] == "buffer-adjustments" && r.Method == http.MethodPost {
		a.createPlantBufferAdjustment(w, r, session)
		return
	}
	if len(parts) == 3 && parts[1] == "tasks" && parts[2] == "auto" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.autoCreateProductionTasks(w, r, session, id)
		return
	}
	if len(parts) == 2 && parts[1] == "tasks" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.createProductionTask(w, r, session, id)
		return
	}
	if len(parts) == 1 && (r.Method == http.MethodPut || r.Method == http.MethodPatch) {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.updateProductionPlan(w, r, session, id)
		return
	}
	if len(parts) == 2 && parts[1] == "cancel" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.cancelProductionPlan(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "tasks" && parts[2] == "batches" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.createProductionBatch(w, r, session, id)
		return
	}
	if len(parts) == 2 && parts[0] == "reports" && parts[1] == "generate" && r.Method == http.MethodPost {
		a.generateProductionReport(w, r, session)
		return
	}
	writeError(w, http.StatusNotFound, "unknown production route")
}

func productionPayload(data AppData) map[string]interface{} {
	plans := enrichedProductionPlans(data)
	return map[string]interface{}{
		"plans":                  plans,
		"tasks":                  data.ProductionTasks,
		"batches":                data.ProductionBatches,
		"reports":                data.ProductionReports,
		"mixDesigns":             data.MixDesigns,
		"mixDesignPlantProfiles": data.MixDesignPlantProfiles,
		"plants":                 plantsWithGatewayStatus(data),
		"plantBufferLocations":   data.PlantBufferLocations,
		"plantBufferFlows":       data.PlantBufferFlows,
		"stockYards":             data.StockYards,
		"stockYardPiles":         data.StockYardPiles,
		"stockYardFlows":         data.StockYardFlows,
		"traces":                 data.InventoryBatchTraces,
		"kpis":                   productionKPIs(data, plans),
		"planSummaries":          productionPlanSummaries(data, plans),
	}
}

func (a *App) createProductionPlan(w http.ResponseWriter, r *http.Request, session Session) {
	var item ProductionPlan
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid production plan")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		order, ok := findOrder(*data, item.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		if order.Status != "approved" && order.Status != "scheduled" && order.Status != "dispatching" {
			return fmt.Errorf("订单未审批不能排产")
		}
		plant, err := selectProductionPlant(*data, order.SiteID, item.PlantID)
		if err != nil {
			return err
		}
		if _, ok := approvedMixDesign(*data, order.ProductID, order.SiteID); !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		remaining := round(order.PlanQuantity - plannedOrderQty(*data, order.ID, 0))
		if remaining <= 0 {
			return fmt.Errorf("订单已无可排产余量")
		}
		item.ID = nextID(data, "productionPlan")
		item.PlanNo = number("PP", item.ID)
		item.ProductID = order.ProductID
		item.SiteID = order.SiteID
		item.PlantID = plant.ID
		item.PlantCode = plant.Code
		item.PlanQuantity = nonZero(item.PlanQuantity, remaining)
		if item.PlanQuantity <= 0 {
			return fmt.Errorf("计划方量必须大于 0")
		}
		if item.PlanQuantity > remaining {
			return fmt.Errorf("计划方量超过订单剩余量")
		}
		item.PlanDate = fallback(item.PlanDate, todayString())
		item.Shift = fallback(item.Shift, "白班")
		mix, profile, ok := productionMixDesignForPlant(*data, item.ProductID, item.SiteID, item.PlantID, item.PlanDate)
		if !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		item.MixDesignID = mix.ID
		item.MixProfileID = profile.ID
		item.CapacityStatus, item.InventoryStatus, item.RecipeStatus, item.RiskReason = evaluateProductionPlanReadiness(*data, item)
		if item.RecipeStatus == "missing" {
			return fmt.Errorf("生产线与配比不匹配：%s", fallback(item.RiskReason, "所选生产线无法承接当前配比"))
		}
		item.Status = "scheduled"
		item.CreatedAt = fallback(item.CreatedAt, nowString())
		item.UpdatedAt = item.CreatedAt
		data.ProductionPlans = append(data.ProductionPlans, item)
		for i := range data.Orders {
			if data.Orders[i].ID == order.ID {
				data.Orders[i].Status = "scheduled"
			}
		}
		addAudit(data, session.User.Username, "create", "production_plan", item.ID, item.PlanNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "production.plan.created")
}

func (a *App) updateProductionPlan(w http.ResponseWriter, r *http.Request, session Session, planID int64) {
	var req ProductionPlan
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid production plan")
		return
	}
	var updated ProductionPlan
	err := a.store.Mutate(func(data *AppData) error {
		index := -1
		for i := range data.ProductionPlans {
			if data.ProductionPlans[i].ID == planID {
				index = i
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("生产计划不存在")
		}
		plan := data.ProductionPlans[index]
		if plan.Status == "cancelled" {
			return fmt.Errorf("生产计划已取消")
		}
		if plan.Status == "completed" && (req.PlanQuantity > 0 || req.PlanDate != "" || req.Shift != "") {
			return fmt.Errorf("生产计划已完成不能调整")
		}
		if err := ensureProductionPlanPlant(*data, &plan); err != nil {
			return err
		}
		order, ok := findOrder(*data, plan.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		if req.PlantID != 0 && req.PlantID != plan.PlantID {
			if plan.ProducedQty > 0 || plannedTaskQty(*data, plan.ID) > 0 {
				return fmt.Errorf("生产计划已有生产或任务，不能更换生产线")
			}
			plant, err := selectProductionPlant(*data, plan.SiteID, req.PlantID)
			if err != nil {
				return err
			}
			plan.PlantID = plant.ID
			plan.PlantCode = plant.Code
		}
		if req.PlanDate != "" {
			plan.PlanDate = req.PlanDate
		}
		if req.Shift != "" {
			plan.Shift = req.Shift
		}
		if req.PlanQuantity > 0 {
			nextQty := round(req.PlanQuantity)
			if nextQty < plan.ProducedQty {
				return fmt.Errorf("计划方量不能小于已生产量")
			}
			allocatedQty := plannedTaskQty(*data, plan.ID)
			if nextQty < allocatedQty {
				return fmt.Errorf("计划方量不能小于已下达任务量")
			}
			maxQty := round(order.PlanQuantity - plannedOrderQty(*data, order.ID, plan.ID))
			if nextQty > maxQty {
				return fmt.Errorf("计划方量超过订单剩余量")
			}
			plan.PlanQuantity = nextQty
		}
		mix, profile, ok := productionMixDesignForPlant(*data, plan.ProductID, plan.SiteID, plan.PlantID, plan.PlanDate)
		if !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		plan.MixDesignID = mix.ID
		plan.MixProfileID = profile.ID
		plan.CapacityStatus, plan.InventoryStatus, plan.RecipeStatus, plan.RiskReason = evaluateProductionPlanReadiness(*data, plan)
		if plan.RecipeStatus == "missing" {
			return fmt.Errorf("生产线与配比不匹配：%s", fallback(plan.RiskReason, "所选生产线无法承接当前配比"))
		}
		plan.Status = normalizeProductionPlanStatus(*data, plan)
		plan.UpdatedAt = nowString()
		data.ProductionPlans[index] = plan
		updated = enrichProductionPlan(*data, plan)
		addAudit(data, session.User.Username, "update", "production_plan", plan.ID, plan.PlanNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, "production.plan.updated")
}

func (a *App) cancelProductionPlan(w http.ResponseWriter, r *http.Request, session Session, planID int64) {
	var updated ProductionPlan
	topic := "production.plan.cancelled"
	err := a.store.Mutate(func(data *AppData) error {
		index := productionPlanIndex(*data, planID)
		if index < 0 {
			return fmt.Errorf("生产计划不存在")
		}
		plan := data.ProductionPlans[index]
		if plan.Status == "cancelled" {
			updated = enrichProductionPlan(*data, plan)
			return nil
		}
		if hasPendingWorkflowForResource(*data, "production_plan", plan.ID) {
			updated = enrichProductionPlan(*data, plan)
			return nil
		}
		if plan.ProducedQty > 0 || planHasBatches(*data, plan.ID) {
			return fmt.Errorf("生产计划已有生产批次不能取消")
		}
		_, instances, err := publishProductionPlanCancelWorkflow(data, plan, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			data.ProductionPlans[index].Status = "pending_approval"
			data.ProductionPlans[index].UpdatedAt = nowString()
			updated = enrichProductionPlan(*data, data.ProductionPlans[index])
			topic = "production.plan.cancel_requested"
			addAudit(data, session.User.Username, "request_cancel", "production_plan", plan.ID, plan.PlanNo, clientIP(r))
			return nil
		}
		next, err := applyProductionPlanCancelLocked(data, plan.ID)
		if err != nil {
			return err
		}
		updated = enrichProductionPlan(*data, next)
		addAudit(data, session.User.Username, "cancel", "production_plan", plan.ID, plan.PlanNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, topic)
}

func publishProductionPlanCancelWorkflow(data *AppData, plan ProductionPlan, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "production_plan.cancel_requested",
		Source:     "production",
		Resource:   "production_plan",
		ResourceID: plan.ID,
		ResourceNo: plan.PlanNo,
		Title:      "生产计划取消 " + plan.PlanNo,
		Actor:      actor,
		Reason:     "生产计划取消审批",
		Variables: map[string]string{
			"orderId":      fmt.Sprintf("%d", plan.OrderID),
			"siteId":       fmt.Sprintf("%d", plan.SiteID),
			"plantId":      fmt.Sprintf("%d", plan.PlantID),
			"plantCode":    plan.PlantCode,
			"productId":    fmt.Sprintf("%d", plan.ProductID),
			"planQuantity": fmt.Sprintf("%.2f", plan.PlanQuantity),
			"producedQty":  fmt.Sprintf("%.2f", plan.ProducedQty),
			"planDate":     plan.PlanDate,
			"shift":        plan.Shift,
			"status":       plan.Status,
		},
	})
}

func applyProductionPlanCancelLocked(data *AppData, planID int64) (ProductionPlan, error) {
	index := productionPlanIndex(*data, planID)
	if index < 0 {
		return ProductionPlan{}, fmt.Errorf("生产计划不存在")
	}
	plan := data.ProductionPlans[index]
	if plan.Status == "cancelled" {
		return plan, nil
	}
	if plan.ProducedQty > 0 || planHasBatches(*data, plan.ID) {
		return ProductionPlan{}, fmt.Errorf("生产计划已有生产批次不能取消")
	}
	plan.Status = "cancelled"
	plan.UpdatedAt = nowString()
	data.ProductionPlans[index] = plan
	for i := range data.ProductionTasks {
		if data.ProductionTasks[i].PlanID == plan.ID && data.ProductionTasks[i].Status != "completed" {
			data.ProductionTasks[i].Status = "cancelled"
			data.ProductionTasks[i].UpdatedAt = nowString()
		}
	}
	refreshOrderProductionStatus(data, plan.OrderID)
	return plan, nil
}

func (a *App) autoCreateProductionTasks(w http.ResponseWriter, r *http.Request, session Session, planID int64) {
	var req struct {
		TaskQty float64 `json:"taskQty"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid production task")
		return
	}
	created := []ProductionTask{}
	err := a.store.Mutate(func(data *AppData) error {
		plan, ok := findProductionPlan(*data, planID)
		if !ok {
			return fmt.Errorf("生产计划不存在")
		}
		if plan.Status == "cancelled" {
			return fmt.Errorf("生产计划已取消")
		}
		if plan.Status == "pending_approval" {
			return fmt.Errorf("生产计划正在流程审批中")
		}
		if plan.Status == "completed" {
			return fmt.Errorf("生产计划已完成")
		}
		if err := ensureProductionPlanPlant(*data, &plan); err != nil {
			return err
		}
		if _, ok := findOrder(*data, plan.OrderID); !ok {
			return fmt.Errorf("订单不存在")
		}
		mix, profile, ok := productionMixDesignByIDs(*data, plan.MixDesignID, plan.MixProfileID, plan.PlantID, plan.PlanDate)
		if !ok {
			mix, profile, ok = productionMixDesignForPlant(*data, plan.ProductID, plan.SiteID, plan.PlantID, plan.PlanDate)
		}
		if !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		plan.MixDesignID = mix.ID
		plan.MixProfileID = profile.ID
		_, _, recipeStatus, recipeReason := evaluateProductionPlanReadiness(*data, plan)
		if recipeStatus == "missing" {
			return fmt.Errorf("生产线与配比不匹配：%s", fallback(recipeReason, "所选生产线无法承接当前配比"))
		}
		remaining := round(plan.PlanQuantity - plannedTaskQty(*data, plan.ID))
		if remaining <= 0 {
			return fmt.Errorf("计划剩余量不足")
		}
		chunkQty := round(req.TaskQty)
		if chunkQty <= 0 || chunkQty > remaining {
			chunkQty = remaining
		}
		for remaining > 0 {
			qty := chunkQty
			if qty > remaining {
				qty = remaining
			}
			id := nextID(data, "productionTask")
			task := ProductionTask{
				ID: id, TaskNo: number("PT", id), PlanID: plan.ID, OrderID: plan.OrderID,
				SiteID: plan.SiteID, PlantID: plan.PlantID, PlantCode: plan.PlantCode,
				ProductID: plan.ProductID, MixDesignID: mix.ID, MixProfileID: profile.ID,
				PlanQty: round(qty), Status: "pending", CreatedAt: nowString(), UpdatedAt: nowString(),
			}
			data.ProductionTasks = append(data.ProductionTasks, task)
			created = append(created, task)
			remaining = round(remaining - qty)
		}
		for i := range data.ProductionPlans {
			if data.ProductionPlans[i].ID == plan.ID {
				data.ProductionPlans[i].PlantID = plan.PlantID
				data.ProductionPlans[i].PlantCode = plan.PlantCode
				data.ProductionPlans[i].MixDesignID = plan.MixDesignID
				data.ProductionPlans[i].MixProfileID = plan.MixProfileID
				if data.ProductionPlans[i].Status == "scheduled" {
					data.ProductionPlans[i].Status = "producing"
					data.ProductionPlans[i].UpdatedAt = nowString()
				}
			}
		}
		addAudit(data, session.User.Username, "auto_tasks", "production_plan", plan.ID, plan.PlanNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, created, "production.tasks.created")
}

func (a *App) createProductionTask(w http.ResponseWriter, r *http.Request, session Session, planID int64) {
	var req ProductionTask
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid production task")
		return
	}
	var item ProductionTask
	err := a.store.Mutate(func(data *AppData) error {
		plan, ok := findProductionPlan(*data, planID)
		if !ok {
			return fmt.Errorf("生产计划不存在")
		}
		if plan.Status == "cancelled" {
			return fmt.Errorf("生产计划已取消")
		}
		if plan.Status == "pending_approval" {
			return fmt.Errorf("生产计划正在流程审批中")
		}
		if plan.Status == "completed" {
			return fmt.Errorf("生产计划已完成")
		}
		if err := ensureProductionPlanPlant(*data, &plan); err != nil {
			return err
		}
		order, ok := findOrder(*data, plan.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		mix, profile, ok := productionMixDesignByIDs(*data, plan.MixDesignID, plan.MixProfileID, plan.PlantID, plan.PlanDate)
		if !ok {
			mix, profile, ok = productionMixDesignForPlant(*data, plan.ProductID, plan.SiteID, plan.PlantID, plan.PlanDate)
		}
		if !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		plan.MixDesignID = mix.ID
		plan.MixProfileID = profile.ID
		_, _, recipeStatus, recipeReason := evaluateProductionPlanReadiness(*data, plan)
		if recipeStatus == "missing" {
			return fmt.Errorf("生产线与配比不匹配：%s", fallback(recipeReason, "所选生产线无法承接当前配比"))
		}
		remaining := round(plan.PlanQuantity - plannedTaskQty(*data, plan.ID))
		item = req
		item.ID = nextID(data, "productionTask")
		item.TaskNo = number("PT", item.ID)
		item.PlanID = plan.ID
		item.OrderID = plan.OrderID
		item.SiteID = plan.SiteID
		item.PlantID = plan.PlantID
		item.PlantCode = plan.PlantCode
		item.ProductID = plan.ProductID
		item.MixDesignID = mix.ID
		item.MixProfileID = profile.ID
		item.PlanQty = nonZero(req.PlanQty, remaining)
		if item.PlanQty <= 0 {
			return fmt.Errorf("计划剩余量不足")
		}
		if item.PlanQty > remaining {
			return fmt.Errorf("任务量超过计划剩余量")
		}
		item.Status = fallback(req.Status, "pending")
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		data.ProductionTasks = append(data.ProductionTasks, item)
		for i := range data.ProductionPlans {
			if data.ProductionPlans[i].ID == plan.ID {
				data.ProductionPlans[i].PlantID = plan.PlantID
				data.ProductionPlans[i].PlantCode = plan.PlantCode
				data.ProductionPlans[i].MixDesignID = plan.MixDesignID
				data.ProductionPlans[i].MixProfileID = plan.MixProfileID
				if data.ProductionPlans[i].Status == "scheduled" {
					data.ProductionPlans[i].Status = "producing"
				}
			}
		}
		for i := range data.Orders {
			if data.Orders[i].ID == order.ID && data.Orders[i].Status == "scheduled" {
				data.Orders[i].Status = "scheduled"
			}
		}
		addAudit(data, session.User.Username, "create", "production_task", item.ID, item.TaskNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "production.task.created")
}

func (a *App) createProductionBatch(w http.ResponseWriter, r *http.Request, session Session, taskID int64) {
	var req ProductionBatch
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid production batch")
		return
	}
	var item ProductionBatch
	err := a.store.Mutate(func(data *AppData) error {
		task, ok := findProductionTask(*data, taskID)
		if !ok {
			return fmt.Errorf("生产任务不存在")
		}
		if task.Status == "completed" {
			return fmt.Errorf("生产任务已完成")
		}
		plan, ok := findProductionPlan(*data, task.PlanID)
		if !ok {
			return fmt.Errorf("生产计划不存在")
		}
		if err := ensureProductionTaskPlant(*data, &task, plan); err != nil {
			return err
		}
		mix, profile, ok := productionMixDesignByIDs(*data, task.MixDesignID, task.MixProfileID, task.PlantID, plan.PlanDate)
		if !ok {
			mix, profile, ok = productionMixDesignForPlant(*data, task.ProductID, task.SiteID, task.PlantID, plan.PlanDate)
		}
		if !ok {
			return fmt.Errorf("配方不存在")
		}
		remaining := round(task.PlanQty - task.ProducedQty)
		quantity := nonZero(req.Quantity, remaining)
		if quantity <= 0 {
			return fmt.Errorf("批次数量必须大于 0")
		}
		if quantity > remaining {
			return fmt.Errorf("批次数量超过任务剩余量")
		}
		item = req
		item.ID = nextID(data, "productionBatch")
		item.BatchNo = number("PB", item.ID)
		item.TaskID = task.ID
		item.PlanID = task.PlanID
		item.OrderID = task.OrderID
		item.SiteID = task.SiteID
		item.PlantID = task.PlantID
		item.ProductID = task.ProductID
		item.MixDesignID = mix.ID
		item.MixProfileID = profile.ID
		item.Quantity = round(quantity)
		if req.PlantID != 0 && req.PlantID != task.PlantID {
			return fmt.Errorf("生产批次生产线必须与任务一致")
		}
		if strings.TrimSpace(req.PlantCode) != "" && !strings.EqualFold(req.PlantCode, task.PlantCode) {
			return fmt.Errorf("生产批次生产线必须与任务一致")
		}
		item.PlantCode = task.PlantCode
		item.Operator = fallback(req.Operator, session.User.DisplayName)
		item.QualityStatus = fallback(req.QualityStatus, "pending")
		item.Status = fallback(req.Status, "produced")
		item.StartedAt = fallback(req.StartedAt, nowString())
		item.CompletedAt = fallback(req.CompletedAt, nowString())
		if err := consumeBatchInventory(data, item, mix); err != nil {
			return err
		}
		data.ProductionBatches = append(data.ProductionBatches, item)
		updateProductionProgress(data, task.ID, plan.ID, item.Quantity)
		addAudit(data, session.User.Username, "create", "production_batch", item.ID, item.BatchNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "production.batch.created")
}

func (a *App) generateProductionReport(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		SiteID     int64  `json:"siteId"`
		ReportDate string `json:"reportDate"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid production report")
		return
	}
	var report ProductionDailyReport
	err := a.store.Mutate(func(data *AppData) error {
		req.ReportDate = fallback(req.ReportDate, todayString())
		req.SiteID = nonZeroInt(req.SiteID, session.User.SiteID)
		if req.SiteID == 0 && len(data.Sites) > 0 {
			req.SiteID = data.Sites[0].ID
		}
		if req.SiteID == 0 {
			return fmt.Errorf("站点不存在")
		}
		report = buildProductionDailyReport(*data, req.SiteID, req.ReportDate)
		found := false
		for i := range data.ProductionReports {
			if data.ProductionReports[i].SiteID == req.SiteID && data.ProductionReports[i].ReportDate == req.ReportDate {
				report.ID = data.ProductionReports[i].ID
				report.ReportNo = data.ProductionReports[i].ReportNo
				data.ProductionReports[i] = report
				found = true
				break
			}
		}
		if !found {
			report.ID = nextID(data, "productionReport")
			report.ReportNo = number("PDR", report.ID)
			data.ProductionReports = append(data.ProductionReports, report)
		}
		addAudit(data, session.User.Username, "generate", "production_report", report.ID, report.ReportNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, report, "production.report.generated")
}

func findProductionPlan(data AppData, id int64) (ProductionPlan, bool) {
	for _, item := range data.ProductionPlans {
		if item.ID == id {
			return item, true
		}
	}
	return ProductionPlan{}, false
}

func productionPlanIndex(data AppData, id int64) int {
	for i := range data.ProductionPlans {
		if data.ProductionPlans[i].ID == id {
			return i
		}
	}
	return -1
}

func findProductionTask(data AppData, id int64) (ProductionTask, bool) {
	for _, item := range data.ProductionTasks {
		if item.ID == id {
			return item, true
		}
	}
	return ProductionTask{}, false
}

func findMixDesign(data AppData, id int64) (MixDesign, bool) {
	for _, item := range data.MixDesigns {
		if item.ID == id {
			return item, true
		}
	}
	return MixDesign{}, false
}

func plannedTaskQty(data AppData, planID int64) float64 {
	total := 0.0
	for _, item := range data.ProductionTasks {
		if item.PlanID == planID && item.Status != "cancelled" {
			total += item.PlanQty
		}
	}
	return round(total)
}

func plannedOrderQty(data AppData, orderID, excludePlanID int64) float64 {
	total := 0.0
	for _, item := range data.ProductionPlans {
		if item.OrderID == orderID && item.ID != excludePlanID && item.Status != "cancelled" {
			total += item.PlanQuantity
		}
	}
	return round(total)
}

func enrichedProductionPlans(data AppData) []ProductionPlan {
	out := make([]ProductionPlan, 0, len(data.ProductionPlans))
	for _, plan := range data.ProductionPlans {
		out = append(out, enrichProductionPlan(data, plan))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PlanDate == out[j].PlanDate {
			return out[i].ID > out[j].ID
		}
		return out[i].PlanDate > out[j].PlanDate
	})
	return out
}

func enrichProductionPlan(data AppData, plan ProductionPlan) ProductionPlan {
	_ = ensureProductionPlanPlant(data, &plan)
	if plan.MixDesignID == 0 {
		if mix, profile, ok := productionMixDesignForPlant(data, plan.ProductID, plan.SiteID, plan.PlantID, plan.PlanDate); ok {
			plan.MixDesignID = mix.ID
			plan.MixProfileID = profile.ID
		}
	}
	plan.PlannedTaskQty = plannedTaskQty(data, plan.ID)
	plan.RemainingQty = round(plan.PlanQuantity - plan.ProducedQty)
	if plan.RemainingQty < 0 {
		plan.RemainingQty = 0
	}
	if plan.PlanQuantity > 0 {
		plan.Progress = round(plan.ProducedQty / plan.PlanQuantity * 100)
	}
	if plan.CapacityStatus == "" || plan.InventoryStatus == "" || plan.RecipeStatus == "" || plan.RiskReason == "" {
		capacity, inventory, recipe, reason := evaluateProductionPlanReadiness(data, plan)
		plan.CapacityStatus = fallback(plan.CapacityStatus, capacity)
		plan.InventoryStatus = fallback(plan.InventoryStatus, inventory)
		plan.RecipeStatus = fallback(plan.RecipeStatus, recipe)
		plan.RiskReason = fallback(plan.RiskReason, reason)
	}
	return plan
}

func productionKPIs(data AppData, plans []ProductionPlan) ProductionOverviewKPI {
	today := todayString()
	out := ProductionOverviewKPI{PlanCount: len(plans), TaskCount: len(data.ProductionTasks), BatchCount: len(data.ProductionBatches)}
	for _, plan := range plans {
		if plan.Status != "cancelled" {
			out.ActivePlanCount++
			out.ScheduledQty = round(out.ScheduledQty + plan.PlanQuantity)
			out.ProducedQty = round(out.ProducedQty + plan.ProducedQty)
			out.RemainingQty = round(out.RemainingQty + plan.RemainingQty)
		}
		if plan.PlanDate == today && plan.Status != "cancelled" {
			out.TodayPlannedQty = round(out.TodayPlannedQty + plan.PlanQuantity)
			out.TodayProducedQty = round(out.TodayProducedQty + plan.ProducedQty)
		}
		if plan.CapacityStatus == "warning" || plan.CapacityStatus == "overbooked" || plan.CapacityStatus == "no_capacity" {
			out.CapacityWarnings++
		}
		if plan.InventoryStatus == "warning" || plan.InventoryStatus == "shortage" {
			out.InventoryWarnings++
		}
	}
	for _, task := range data.ProductionTasks {
		if task.Status == "running" || task.Status == "pending" {
			out.RunningTaskCount++
		}
	}
	return out
}

func productionPlanSummaries(data AppData, plans []ProductionPlan) []ProductionPlanSummary {
	type key struct {
		date   string
		siteID int64
	}
	index := map[key]int{}
	out := []ProductionPlanSummary{}
	for _, plan := range plans {
		if plan.Status == "cancelled" {
			continue
		}
		k := key{date: plan.PlanDate, siteID: plan.SiteID}
		pos, ok := index[k]
		if !ok {
			out = append(out, ProductionPlanSummary{PlanDate: plan.PlanDate, SiteID: plan.SiteID, Status: "scheduled"})
			pos = len(out) - 1
			index[k] = pos
		}
		summary := &out[pos]
		summary.PlanCount++
		summary.PlannedQty = round(summary.PlannedQty + plan.PlanQuantity)
		summary.ProducedQty = round(summary.ProducedQty + plan.ProducedQty)
		summary.RemainingQty = round(summary.RemainingQty + plan.RemainingQty)
		summary.Status = mergeProductionSummaryStatus(summary.Status, plan.Status, plan.CapacityStatus, plan.InventoryStatus, plan.RecipeStatus)
		if summary.RiskReason == "" && plan.RiskReason != "" {
			summary.RiskReason = plan.RiskReason
		}
	}
	for i := range out {
		for _, task := range data.ProductionTasks {
			plan, ok := findProductionPlan(data, task.PlanID)
			if ok && plan.PlanDate == out[i].PlanDate && plan.SiteID == out[i].SiteID && task.Status != "cancelled" {
				out[i].TaskCount++
			}
		}
		for _, batch := range data.ProductionBatches {
			if batch.SiteID == out[i].SiteID && datePart(batch.CompletedAt) == out[i].PlanDate {
				out[i].BatchCount++
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PlanDate == out[j].PlanDate {
			return out[i].SiteID < out[j].SiteID
		}
		return out[i].PlanDate > out[j].PlanDate
	})
	return out
}

func mergeProductionSummaryStatus(current, planStatus, capacityStatus, inventoryStatus, recipeStatus string) string {
	if capacityStatus == "overbooked" || capacityStatus == "no_capacity" || inventoryStatus == "shortage" || recipeStatus != "ok" {
		return "warning"
	}
	if current == "warning" {
		return current
	}
	if planStatus == "producing" || planStatus == "running" {
		return "producing"
	}
	if current == "producing" {
		return current
	}
	if planStatus == "scheduled" || planStatus == "pending" {
		return "scheduled"
	}
	if planStatus == "completed" && current == "scheduled" {
		return "completed"
	}
	return current
}

func evaluateProductionPlanReadiness(data AppData, plan ProductionPlan) (string, string, string, string) {
	reasons := []string{}
	if err := ensureProductionPlanPlant(data, &plan); err != nil {
		reasons = append(reasons, err.Error())
	}
	capacityStatus, capacityReason := productionCapacityStatus(data, plan)
	if capacityReason != "" {
		reasons = append(reasons, capacityReason)
	}
	var mix MixDesign
	var ok bool
	if plan.MixDesignID > 0 {
		mix, _, ok = productionMixDesignByIDs(data, plan.MixDesignID, plan.MixProfileID, plan.PlantID, plan.PlanDate)
	} else {
		mix, _, ok = productionMixDesignForPlant(data, plan.ProductID, plan.SiteID, plan.PlantID, plan.PlanDate)
	}
	if !ok {
		reasons = append(reasons, "缺少已审批配方")
		return capacityStatus, "unknown", "missing", strings.Join(reasons, "；")
	}
	recipeStatus, recipeReason := productionLineRecipeStatus(data, plan, mix)
	if recipeReason != "" {
		reasons = append(reasons, recipeReason)
	}
	inventoryStatus, inventoryReason := productionInventoryStatus(data, plan.SiteID, plan.PlanQuantity, mix)
	if inventoryReason != "" {
		reasons = append(reasons, inventoryReason)
	}
	return capacityStatus, inventoryStatus, recipeStatus, strings.Join(reasons, "；")
}

func productionCapacityStatus(data AppData, plan ProductionPlan) (string, string) {
	capacity := siteShiftCapacity(data, plan.SiteID)
	if plan.PlantID != 0 {
		plant, ok := findPlantByID(data, plan.PlantID)
		if !ok {
			return "no_capacity", "生产线不存在"
		}
		capacity = plantShiftCapacity(plant)
	}
	if capacity <= 0 {
		return "no_capacity", "生产线未配置可用产能"
	}
	planned := plan.PlanQuantity
	for _, item := range data.ProductionPlans {
		if item.ID == plan.ID || item.Status == "cancelled" {
			continue
		}
		if item.SiteID != plan.SiteID || item.PlanDate != plan.PlanDate || item.Shift != plan.Shift {
			continue
		}
		if plan.PlantID != 0 {
			if err := ensureProductionPlanPlant(data, &item); err != nil || item.PlantID != plan.PlantID {
				continue
			}
		}
		if item.SiteID == plan.SiteID {
			planned += item.PlanQuantity
		}
	}
	planned = round(planned)
	if planned > capacity {
		return "overbooked", fmt.Sprintf("班次计划 %.2f 超过产能 %.2f", planned, capacity)
	}
	if planned > round(capacity*0.85) {
		return "warning", fmt.Sprintf("班次计划 %.2f 接近产能 %.2f", planned, capacity)
	}
	return "ok", ""
}

func siteShiftCapacity(data AppData, siteID int64) float64 {
	total := 0.0
	for _, plant := range data.Plants {
		if plant.SiteID != siteID || (plant.Status != "running" && plant.Status != "active") {
			continue
		}
		capacity := plantShiftCapacity(plant)
		if capacity <= 0 {
			continue
		}
		total += capacity
	}
	return round(total)
}

func plantShiftCapacity(plant Plant) float64 {
	capacity := parseCapacityNumber(plant.Capacity)
	if capacity <= 0 {
		return 0
	}
	if strings.Contains(strings.ToLower(plant.Capacity), "/h") {
		capacity *= 8
	}
	return round(capacity)
}

func parseCapacityNumber(value string) float64 {
	var b strings.Builder
	for _, r := range value {
		if (r >= '0' && r <= '9') || r == '.' {
			b.WriteRune(r)
			continue
		}
		if b.Len() > 0 {
			break
		}
	}
	if b.Len() == 0 {
		return 0
	}
	out, _ := strconv.ParseFloat(b.String(), 64)
	return out
}

func productionInventoryStatus(data AppData, siteID int64, quantity float64, mix MixDesign) (string, string) {
	shortages := []string{}
	warnings := []string{}
	for _, material := range productionMaterialRequirements(mix, quantity) {
		balance := inventoryBalance(data, siteID, material.MaterialID)
		name := fmt.Sprintf("物料 %d", material.MaterialID)
		safeStock := 0.0
		if item, ok := findMaterial(data, material.MaterialID); ok {
			name = item.Name
			safeStock = item.SafeStock
		}
		if balance < material.Quantity {
			shortages = append(shortages, fmt.Sprintf("%s需%.2f余%.2f", name, material.Quantity, balance))
			continue
		}
		if safeStock > 0 && round(balance-material.Quantity) < safeStock {
			warnings = append(warnings, name)
		}
	}
	if len(shortages) > 0 {
		return "shortage", "库存不足：" + strings.Join(shortages, "、")
	}
	if len(warnings) > 0 {
		return "warning", "生产后低于安全库存：" + strings.Join(warnings, "、")
	}
	return "ok", ""
}

func productionMaterialRequirements(mix MixDesign, quantity float64) []productionMaterialConsumption {
	out := []productionMaterialConsumption{}
	for _, material := range mix.Materials {
		required := round(material.Dosage * quantity / 1000)
		if required <= 0 {
			continue
		}
		out = append(out, productionMaterialConsumption{MaterialID: material.MaterialID, Quantity: required, Unit: material.Unit, BufferCode: material.BufferCode})
	}
	return out
}

func ensureProductionPlanPlant(data AppData, plan *ProductionPlan) error {
	plant, err := selectProductionPlant(data, plan.SiteID, plan.PlantID)
	if err != nil {
		return err
	}
	plan.PlantID = plant.ID
	plan.PlantCode = plant.Code
	return nil
}

func ensureProductionTaskPlant(data AppData, task *ProductionTask, plan ProductionPlan) error {
	if err := ensureProductionPlanPlant(data, &plan); err != nil {
		return err
	}
	if task.PlantID != 0 && task.PlantID != plan.PlantID {
		return fmt.Errorf("生产任务生产线与计划不一致")
	}
	if strings.TrimSpace(task.PlantCode) != "" && !strings.EqualFold(task.PlantCode, plan.PlantCode) {
		return fmt.Errorf("生产任务生产线与计划不一致")
	}
	task.PlantID = plan.PlantID
	task.PlantCode = plan.PlantCode
	return nil
}

func selectProductionPlant(data AppData, siteID, requestedID int64) (Plant, error) {
	if requestedID != 0 {
		plant, ok := findPlantByID(data, requestedID)
		if !ok {
			return Plant{}, fmt.Errorf("生产线不存在")
		}
		if plant.SiteID != siteID {
			return Plant{}, fmt.Errorf("生产线必须属于计划站点")
		}
		if !productionPlantAvailable(plant) {
			return Plant{}, fmt.Errorf("生产线未启用")
		}
		return plant, nil
	}
	for _, plant := range data.Plants {
		if plant.SiteID == siteID && productionPlantAvailable(plant) {
			return plant, nil
		}
	}
	return Plant{}, fmt.Errorf("站点未配置可用生产线")
}

func productionPlantAvailable(plant Plant) bool {
	return plant.Status == "running" || plant.Status == "active"
}

func productionLineRecipeStatus(data AppData, plan ProductionPlan, mix MixDesign) (string, string) {
	if plan.PlantID == 0 {
		return "missing", "生产计划未选择生产线"
	}
	missing := []string{}
	warnings := []string{}
	for _, material := range productionMaterialRequirements(mix, plan.PlanQuantity) {
		if !siteUsesLineBufferForMaterial(data, plan.SiteID, material.MaterialID) {
			continue
		}
		name := productionMaterialName(data, material.MaterialID)
		buffers := productionLineBuffersForMaterial(data, plan.PlantID, material.MaterialID)
		if len(buffers) == 0 {
			missing = append(missing, fmt.Sprintf("%s未配置到%s", name, fallback(plan.PlantCode, "所选生产线")))
			continue
		}
		usableQty := 0.0
		hasUsableBuffer := false
		for _, buffer := range buffers {
			if buffer.MaterialID != material.MaterialID {
				continue
			}
			if buffer.Status != "active" && buffer.Status != "running" {
				continue
			}
			if buffer.QualityStatus != "" && buffer.QualityStatus != "passed" {
				continue
			}
			hasUsableBuffer = true
			usableQty = round(usableQty + buffer.CurrentQty)
		}
		if !hasUsableBuffer {
			missing = append(missing, fmt.Sprintf("%s未在筒仓装入可用物料", name))
			continue
		}
		if usableQty < material.Quantity {
			warnings = append(warnings, fmt.Sprintf("%s筒仓余量%.2f少于计划需%.2f", name, usableQty, material.Quantity))
		}
	}
	if len(missing) > 0 {
		return "missing", "生产线配比不匹配：" + strings.Join(missing, "、")
	}
	if len(warnings) > 0 {
		return "warning", "筒仓风险：" + strings.Join(warnings, "、")
	}
	return "ok", ""
}

func siteUsesLineBufferForMaterial(data AppData, siteID, materialID int64) bool {
	for _, buffer := range data.PlantBufferLocations {
		if buffer.SiteID == siteID && productionLineBufferCanCarryMaterial(buffer, materialID) {
			return true
		}
	}
	return false
}

func productionLineBuffersForMaterial(data AppData, plantID, materialID int64) []PlantBufferLocation {
	out := []PlantBufferLocation{}
	for _, buffer := range data.PlantBufferLocations {
		if buffer.PlantID == plantID && productionLineBufferCanCarryMaterial(buffer, materialID) {
			out = append(out, buffer)
		}
	}
	return out
}

func productionLineBufferCanCarryMaterial(buffer PlantBufferLocation, materialID int64) bool {
	if buffer.MaterialID == materialID {
		return true
	}
	for _, id := range buffer.AllowedMaterialIDs {
		if id == materialID {
			return true
		}
	}
	return false
}

func productionMaterialName(data AppData, materialID int64) string {
	if item, ok := findMaterial(data, materialID); ok {
		return item.Name
	}
	return fmt.Sprintf("物料%d", materialID)
}

func normalizeProductionPlanStatus(data AppData, plan ProductionPlan) string {
	if plan.Status == "cancelled" {
		return "cancelled"
	}
	if plan.PlanQuantity > 0 && plan.ProducedQty >= plan.PlanQuantity {
		return "completed"
	}
	if plan.ProducedQty > 0 || plannedTaskQty(data, plan.ID) > 0 {
		return "producing"
	}
	return "scheduled"
}

func planHasBatches(data AppData, planID int64) bool {
	for _, item := range data.ProductionBatches {
		if item.PlanID == planID {
			return true
		}
	}
	return false
}

func refreshOrderProductionStatus(data *AppData, orderID int64) {
	hasActivePlan := false
	for _, plan := range data.ProductionPlans {
		if plan.OrderID == orderID && plan.Status != "cancelled" {
			hasActivePlan = true
			break
		}
	}
	for i := range data.Orders {
		if data.Orders[i].ID != orderID {
			continue
		}
		if !hasActivePlan && data.Orders[i].Status == "scheduled" {
			data.Orders[i].Status = "approved"
		}
		return
	}
}

func defaultPlantCode(data AppData, siteID int64) string {
	for _, plant := range data.Plants {
		if plant.SiteID == siteID && plant.Status == "running" {
			return plant.Code
		}
	}
	return "manual"
}

func consumeBatchInventory(data *AppData, batch ProductionBatch, mix MixDesign) error {
	consumptions := make([]productionMaterialConsumption, 0, len(mix.Materials))
	for _, material := range mix.Materials {
		required := round(material.Dosage * batch.Quantity / 1000)
		if required <= 0 {
			continue
		}
		consumptions = append(consumptions, productionMaterialConsumption{MaterialID: material.MaterialID, Quantity: required, Unit: material.Unit, BufferCode: material.BufferCode})
	}
	return consumeBatchInventoryMaterials(data, batch, consumptions, "生产批次理论扣减")
}

func consumeBatchInventoryMaterials(data *AppData, batch ProductionBatch, materials []productionMaterialConsumption, remark string) error {
	for _, material := range materials {
		if consumed, err := consumeBatchBufferMaterial(data, batch, material); consumed || err != nil {
			if err != nil {
				return err
			}
			continue
		}
		required := round(material.Quantity)
		if required <= 0 {
			return fmt.Errorf("物料消耗量必须大于 0")
		}
		if material.MaterialID <= 0 {
			return fmt.Errorf("物料编号不能为空")
		}
		balance, lots, ok := decreaseInventory(data, batch.SiteID, material.MaterialID, required)
		if !ok {
			name := fmt.Sprintf("物料 %d", material.MaterialID)
			if item, found := findMaterial(*data, material.MaterialID); found {
				name = item.Name
			}
			return fmt.Errorf("%s 库存不足", name)
		}
		flowID := nextID(data, "inventoryFlow")
		data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
			ID: flowID, FlowNo: number("IF", flowID), SiteID: batch.SiteID, MaterialID: material.MaterialID,
			SourceType: "production_batch", SourceID: batch.ID, Direction: "out", Quantity: required,
			BalanceQty: balance, Remark: remark, CreatedAt: batch.CompletedAt,
		})
		appendInventoryBatchTraces(data, batch, lots)
	}
	return nil
}

func decreaseInventory(data *AppData, siteID, materialID int64, quantity float64) (float64, []inventoryLotConsumption, bool) {
	balance := inventoryBalance(*data, siteID, materialID)
	if quantity <= 0 {
		return balance, nil, true
	}
	if balance < quantity {
		return balance, nil, false
	}
	remaining := quantity
	lots := []inventoryLotConsumption{}
	for i := range data.Inventory {
		if data.Inventory[i].SiteID != siteID || data.Inventory[i].MaterialID != materialID {
			continue
		}
		if remaining <= 0 {
			break
		}
		used := data.Inventory[i].Quantity
		if used > remaining {
			used = remaining
		}
		data.Inventory[i].Quantity = round(data.Inventory[i].Quantity - used)
		data.Inventory[i].UpdatedAt = nowString()
		adjustSiloCurrentQty(data, siteID, data.Inventory[i].Silo, -used)
		lots = append(lots, inventoryLotConsumption{
			InventoryItemID: data.Inventory[i].ID,
			RawReceiptID:    data.Inventory[i].RawReceiptID,
			SiteID:          data.Inventory[i].SiteID,
			MaterialID:      data.Inventory[i].MaterialID,
			SupplierID:      data.Inventory[i].SupplierID,
			BatchNo:         data.Inventory[i].BatchNo,
			Warehouse:       data.Inventory[i].Warehouse,
			Silo:            data.Inventory[i].Silo,
			Quantity:        round(used),
			Unit:            data.Inventory[i].Unit,
		})
		remaining = round(remaining - used)
	}
	balance = inventoryBalance(*data, siteID, materialID)
	for i := range data.Inventory {
		if data.Inventory[i].SiteID != siteID || data.Inventory[i].MaterialID != materialID {
			continue
		}
		if material, ok := findMaterial(*data, materialID); ok && balance < material.SafeStock {
			data.Inventory[i].AvailableStatus = "warning"
		}
	}
	return balance, lots, true
}

func appendInventoryBatchTraces(data *AppData, batch ProductionBatch, lots []inventoryLotConsumption) {
	for _, lot := range lots {
		if lot.Quantity <= 0 {
			continue
		}
		id := nextID(data, "inventoryTrace")
		data.InventoryBatchTraces = append(data.InventoryBatchTraces, InventoryBatchTrace{
			ID: id, TraceNo: number("IBT", id), ProductionBatchID: batch.ID, ProductionBatchNo: batch.BatchNo,
			RawReceiptID: lot.RawReceiptID, InventoryItemID: lot.InventoryItemID, SiteID: lot.SiteID,
			MaterialID: lot.MaterialID, SupplierID: lot.SupplierID, BatchNo: lot.BatchNo,
			Warehouse: lot.Warehouse, Silo: lot.Silo, Quantity: lot.Quantity, Unit: fallback(lot.Unit, "t"),
			CreatedAt: batch.CompletedAt,
		})
	}
}

func updateProductionProgress(data *AppData, taskID, planID int64, quantity float64) {
	for i := range data.ProductionTasks {
		if data.ProductionTasks[i].ID == taskID {
			data.ProductionTasks[i].ProducedQty = round(data.ProductionTasks[i].ProducedQty + quantity)
			if data.ProductionTasks[i].StartedAt == "" {
				data.ProductionTasks[i].StartedAt = nowString()
			}
			data.ProductionTasks[i].UpdatedAt = nowString()
			if data.ProductionTasks[i].ProducedQty >= data.ProductionTasks[i].PlanQty {
				data.ProductionTasks[i].Status = "completed"
				data.ProductionTasks[i].CompletedAt = nowString()
			} else {
				data.ProductionTasks[i].Status = "running"
			}
		}
	}
	for i := range data.ProductionPlans {
		if data.ProductionPlans[i].ID == planID {
			data.ProductionPlans[i].ProducedQty = round(data.ProductionPlans[i].ProducedQty + quantity)
			if data.ProductionPlans[i].ProducedQty >= data.ProductionPlans[i].PlanQuantity {
				data.ProductionPlans[i].Status = "completed"
			} else {
				data.ProductionPlans[i].Status = "producing"
			}
		}
	}
}

func buildProductionDailyReport(data AppData, siteID int64, reportDate string) ProductionDailyReport {
	report := ProductionDailyReport{
		SiteID: siteID, ReportDate: reportDate, Status: "generated", GeneratedAt: nowString(),
	}
	for _, plan := range data.ProductionPlans {
		if plan.SiteID == siteID && plan.PlanDate == reportDate {
			report.PlannedQty = round(report.PlannedQty + plan.PlanQuantity)
		}
	}
	for _, batch := range data.ProductionBatches {
		if batch.SiteID != siteID || datePart(batch.CompletedAt) != reportDate {
			continue
		}
		report.ProducedQty = round(report.ProducedQty + batch.Quantity)
		report.BatchCount++
		if batch.QualityStatus == "passed" {
			report.QualityPassed++
		} else {
			report.QualityPending++
		}
		if product, ok := findProduct(data, batch.ProductID); ok {
			report.MaterialCost = round(report.MaterialCost + product.CostPrice*batch.Quantity)
		}
	}
	return report
}

func datePart(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return strings.TrimSpace(value)
}
