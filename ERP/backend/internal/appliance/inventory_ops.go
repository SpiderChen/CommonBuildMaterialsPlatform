package appliance

import (
	"fmt"
	"net/http"
)

func (a *App) createInventoryTransfer(w http.ResponseWriter, r *http.Request, session Session) {
	var item InventoryTransfer
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory transfer")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.FromSiteID == 0 || item.ToSiteID == 0 || item.FromSiteID == item.ToSiteID {
			return fmt.Errorf("调拨站点不合法")
		}
		if _, ok := findMaterial(*data, item.MaterialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("调拨数量必须大于 0")
		}
		item.ID = nextID(data, "inventoryTransfer")
		item.TransferNo = number("IT", item.ID)
		item.Unit = fallback(item.Unit, "t")
		item.Status = "pending_approval"
		item.CreatedAt = nowString()
		data.InventoryTransfers = append(data.InventoryTransfers, item)
		reason := fmt.Sprintf("调拨物料 %d: %d -> %d, 数量 %.2f%s", item.MaterialID, item.FromSiteID, item.ToSiteID, item.Quantity, item.Unit)
		_, instances, err := publishWorkflowEvent(data, workflowEventRequest{
			EventType:  "inventory_transfer.submitted",
			Resource:   "inventory_transfer",
			ResourceID: item.ID,
			ResourceNo: item.TransferNo,
			Title:      "库存调拨审批",
			Actor:      session.User.Username,
			Reason:     reason,
			Variables: map[string]string{
				"materialId": fmt.Sprintf("%d", item.MaterialID),
				"fromSiteId": fmt.Sprintf("%d", item.FromSiteID),
				"toSiteId":   fmt.Sprintf("%d", item.ToSiteID),
				"quantity":   fmt.Sprintf("%.2f", item.Quantity),
				"unit":       item.Unit,
			},
		})
		if err != nil {
			return err
		}
		if len(instances) == 0 {
			return fmt.Errorf("库存调拨工作流未配置")
		}
		addAudit(data, session.User.Username, "create", "inventory_transfer", item.ID, item.TransferNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "inventory.transfer.created")
}

func (a *App) completeInventoryTransfer(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item InventoryTransfer
	err := a.store.Mutate(func(data *AppData) error {
		idx := -1
		for i := range data.InventoryTransfers {
			if data.InventoryTransfers[i].ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("调拨单不存在")
		}
		if data.InventoryTransfers[idx].Status == "completed" {
			item = data.InventoryTransfers[idx]
			return nil
		}
		if data.InventoryTransfers[idx].Status != "approved" {
			return fmt.Errorf("调拨单未审批通过")
		}
		item = data.InventoryTransfers[idx]
		if _, lots, ok := decreaseInventory(data, item.FromSiteID, item.MaterialID, item.Quantity); !ok {
			return fmt.Errorf("调出库存不足")
		} else {
			for _, lot := range lots {
				increaseInventoryWithLot(data, item.ToSiteID, item.MaterialID, lot.SupplierID, lot.Quantity, lot.BatchNo, lot.RawReceiptID)
			}
		}
		inBalance := inventoryBalance(*data, item.ToSiteID, item.MaterialID)
		outBalance := inventoryBalance(*data, item.FromSiteID, item.MaterialID)
		outFlowID := nextID(data, "inventoryFlow")
		data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
			ID: outFlowID, FlowNo: number("IF", outFlowID), SiteID: item.FromSiteID, MaterialID: item.MaterialID,
			SourceType: "inventory_transfer", SourceID: item.ID, Direction: "out", Quantity: item.Quantity,
			BalanceQty: outBalance, Remark: "站点调拨出库", CreatedAt: nowString(),
		})
		inFlowID := nextID(data, "inventoryFlow")
		data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
			ID: inFlowID, FlowNo: number("IF", inFlowID), SiteID: item.ToSiteID, MaterialID: item.MaterialID,
			SourceType: "inventory_transfer", SourceID: item.ID, Direction: "in", Quantity: item.Quantity,
			BalanceQty: inBalance, Remark: "站点调拨入库", CreatedAt: nowString(),
		})
		data.InventoryTransfers[idx].Status = "completed"
		data.InventoryTransfers[idx].CompletedAt = nowString()
		item = data.InventoryTransfers[idx]
		addAudit(data, session.User.Username, "complete", "inventory_transfer", item.ID, item.TransferNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "inventory.transfer.completed")
}

func (a *App) createInventoryStocktake(w http.ResponseWriter, r *http.Request, session Session) {
	var item InventoryStocktake
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory stocktake")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findMaterial(*data, item.MaterialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		item.ID = nextID(data, "stocktake")
		item.StocktakeNo = number("STK", item.ID)
		item.BookQty = inventoryBalance(*data, item.SiteID, item.MaterialID)
		item.DiffQty = round(item.ActualQty - item.BookQty)
		item.Unit = fallback(item.Unit, "t")
		item.Status = "pending_review"
		item.CreatedAt = nowString()
		data.InventoryStocktakes = append(data.InventoryStocktakes, item)
		addAudit(data, session.User.Username, "create", "inventory_stocktake", item.ID, item.StocktakeNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "inventory.stocktake.created")
}

func (a *App) reviewInventoryStocktake(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item InventoryStocktake
	topic := "inventory.stocktake.reviewed"
	err := a.store.Mutate(func(data *AppData) error {
		idx := inventoryStocktakeIndex(*data, id)
		if idx < 0 {
			return fmt.Errorf("盘点单不存在")
		}
		if data.InventoryStocktakes[idx].Status == "completed" {
			item = data.InventoryStocktakes[idx]
			return nil
		}
		_, instances, err := publishInventoryStocktakeReviewWorkflow(data, data.InventoryStocktakes[idx], session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			data.InventoryStocktakes[idx].Status = "pending_approval"
			item = data.InventoryStocktakes[idx]
			topic = "inventory.stocktake.workflow_requested"
			addAudit(data, session.User.Username, "request_review", "inventory_stocktake", item.ID, item.StocktakeNo, clientIP(r))
			return nil
		}
		next, err := applyInventoryStocktakeReviewLocked(data, id)
		if err != nil {
			return err
		}
		item = next
		addAudit(data, session.User.Username, "review", "inventory_stocktake", item.ID, item.StocktakeNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishInventoryStocktakeReviewWorkflow(data *AppData, item InventoryStocktake, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "inventory_stocktake.review_requested",
		Source:     "inventory",
		Resource:   "inventory_stocktake",
		ResourceID: item.ID,
		ResourceNo: item.StocktakeNo,
		Title:      "库存盘点复核 " + item.StocktakeNo,
		Actor:      actor,
		Reason:     fallback(item.Remark, "库存盘点复核"),
		Variables: map[string]string{
			"siteId":     fmt.Sprintf("%d", item.SiteID),
			"materialId": fmt.Sprintf("%d", item.MaterialID),
			"bookQty":    fmt.Sprintf("%.2f", item.BookQty),
			"actualQty":  fmt.Sprintf("%.2f", item.ActualQty),
			"diffQty":    fmt.Sprintf("%.2f", item.DiffQty),
			"unit":       item.Unit,
		},
	})
}

func applyInventoryStocktakeReviewLocked(data *AppData, id int64) (InventoryStocktake, error) {
	idx := inventoryStocktakeIndex(*data, id)
	if idx < 0 {
		return InventoryStocktake{}, fmt.Errorf("盘点单不存在")
	}
	if data.InventoryStocktakes[idx].Status == "completed" {
		return data.InventoryStocktakes[idx], nil
	}
	item := data.InventoryStocktakes[idx]
	if !setInventoryQuantity(data, item.SiteID, item.MaterialID, item.ActualQty) {
		return InventoryStocktake{}, fmt.Errorf("库存不存在")
	}
	direction := "in"
	quantity := item.DiffQty
	if quantity < 0 {
		direction = "out"
		quantity = -quantity
	}
	if quantity > 0 {
		flowID := nextID(data, "inventoryFlow")
		data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
			ID: flowID, FlowNo: number("IF", flowID), SiteID: item.SiteID, MaterialID: item.MaterialID,
			SourceType: "inventory_stocktake", SourceID: item.ID, Direction: direction, Quantity: round(quantity),
			BalanceQty: item.ActualQty, Remark: "库存盘点调整", CreatedAt: nowString(),
		})
	}
	data.InventoryStocktakes[idx].Status = "completed"
	data.InventoryStocktakes[idx].ReviewedAt = nowString()
	return data.InventoryStocktakes[idx], nil
}

func inventoryStocktakeIndex(data AppData, id int64) int {
	for i := range data.InventoryStocktakes {
		if data.InventoryStocktakes[i].ID == id {
			return i
		}
	}
	return -1
}

func inventoryBalance(data AppData, siteID, materialID int64) float64 {
	total := 0.0
	for _, item := range data.Inventory {
		if item.SiteID == siteID && item.MaterialID == materialID {
			total += item.Quantity
		}
	}
	return round(total)
}

func setInventoryQuantity(data *AppData, siteID, materialID int64, quantity float64) bool {
	current := inventoryBalance(*data, siteID, materialID)
	diff := round(quantity - current)
	if diff == 0 {
		return current == quantity
	}
	if diff > 0 {
		increaseInventory(data, siteID, materialID, 0, diff)
		return true
	}
	_, _, ok := decreaseInventory(data, siteID, materialID, -diff)
	return ok
}
