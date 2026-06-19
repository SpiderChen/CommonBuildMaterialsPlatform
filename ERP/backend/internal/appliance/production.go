package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (a *App) production(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).ProductionPlans)
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
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
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
		default:
			writeError(w, http.StatusNotFound, "unknown production resource")
		}
		return
	}
	if len(parts) == 2 && parts[1] == "tasks" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.createProductionTask(w, r, session, id)
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
	return map[string]interface{}{
		"plans":      data.ProductionPlans,
		"tasks":      data.ProductionTasks,
		"batches":    data.ProductionBatches,
		"reports":    data.ProductionReports,
		"mixDesigns": data.MixDesigns,
		"plants":     data.Plants,
		"traces":     data.InventoryBatchTraces,
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
		if _, ok := approvedMixDesign(*data, order.ProductID, order.SiteID); !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		item.ID = nextID(data, "productionPlan")
		item.PlanNo = number("PP", item.ID)
		item.ProductID = order.ProductID
		item.SiteID = order.SiteID
		item.PlanQuantity = nonZero(item.PlanQuantity, order.PlanQuantity)
		item.PlanDate = fallback(item.PlanDate, todayString())
		item.Shift = fallback(item.Shift, "白班")
		item.CapacityStatus = "ok"
		item.InventoryStatus = inventoryStatus(*data)
		item.RecipeStatus = "ok"
		item.Status = "scheduled"
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
		order, ok := findOrder(*data, plan.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		mix, ok := approvedMixDesign(*data, plan.ProductID, plan.SiteID)
		if !ok {
			return fmt.Errorf("产品无已审批配方")
		}
		remaining := round(plan.PlanQuantity - plannedTaskQty(*data, plan.ID))
		item = req
		item.ID = nextID(data, "productionTask")
		item.TaskNo = number("PT", item.ID)
		item.PlanID = plan.ID
		item.OrderID = plan.OrderID
		item.SiteID = plan.SiteID
		item.ProductID = plan.ProductID
		item.MixDesignID = mix.ID
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
			if data.ProductionPlans[i].ID == plan.ID && data.ProductionPlans[i].Status == "scheduled" {
				data.ProductionPlans[i].Status = "producing"
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
		mix, ok := findMixDesign(*data, task.MixDesignID)
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
		item.ProductID = task.ProductID
		item.MixDesignID = task.MixDesignID
		item.Quantity = round(quantity)
		item.PlantCode = fallback(req.PlantCode, defaultPlantCode(*data, task.SiteID))
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
		consumptions = append(consumptions, productionMaterialConsumption{MaterialID: material.MaterialID, Quantity: required, Unit: material.Unit})
	}
	return consumeBatchInventoryMaterials(data, batch, consumptions, "生产批次理论扣减")
}

func consumeBatchInventoryMaterials(data *AppData, batch ProductionBatch, materials []productionMaterialConsumption, remark string) error {
	for _, material := range materials {
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
