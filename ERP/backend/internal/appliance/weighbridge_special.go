package appliance

import (
	"fmt"
	"net/http"
)

func (a *App) createTransferTicket(w http.ResponseWriter, r *http.Request, session Session) {
	var req ScaleTicket
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid transfer ticket")
		return
	}
	var item ScaleTicket
	err := a.store.Mutate(func(data *AppData) error {
		transfer, ok := findInventoryTransfer(*data, req.TransferID)
		if !ok {
			return fmt.Errorf("调拨单不存在")
		}
		if transfer.Status != "completed" {
			return fmt.Errorf("调拨单未完成，不能生成调拨过磅记录")
		}
		item = baseScaleTicket(data, "inventory_transfer", transfer.FromSiteID, req)
		item.TransferID = transfer.ID
		item.MaterialID = transfer.MaterialID
		item.Unit = fallback(req.Unit, fallback(transfer.Unit, "t"))
		item.PlateNo = fallback(req.PlateNo, fmt.Sprintf("调拨车-%s", transfer.TransferNo))
		item.NetWeight = measuredNetWeight(item, transfer.Quantity)
		item.Remark = fallback(req.Remark, "库存调拨过磅")
		appendScaleTicketWithWeights(data, item)
		addAudit(data, session.User.Username, "create", "scale_ticket", item.ID, item.TicketNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "ticket.transfer.created")
}

func (a *App) createReturnTicket(w http.ResponseWriter, r *http.Request, session Session) {
	var req ScaleTicket
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid return ticket")
		return
	}
	var item ScaleTicket
	err := a.store.Mutate(func(data *AppData) error {
		var dispatch DispatchOrder
		var order SalesOrder
		var vehicle Vehicle
		var relatedTicket ScaleTicket
		if req.RelatedTicketID != 0 {
			var ok bool
			relatedTicket, ok = findScaleTicket(*data, req.RelatedTicketID)
			if !ok {
				return fmt.Errorf("原出厂过磅记录不存在")
			}
			req.DispatchID = nonZeroInt(req.DispatchID, relatedTicket.DispatchID)
		}
		var ok bool
		dispatch, ok = findDispatch(*data, req.DispatchID)
		if !ok {
			return fmt.Errorf("派车单不存在")
		}
		order, ok = findOrder(*data, dispatch.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		vehicle, ok = findVehicle(*data, dispatch.VehicleID)
		if !ok {
			return fmt.Errorf("车辆不存在")
		}
		item = baseScaleTicket(data, "product_return", dispatch.SiteID, req)
		item.DispatchID = dispatch.ID
		item.OrderID = order.ID
		item.VehicleID = vehicle.ID
		item.RelatedTicketID = req.RelatedTicketID
		item.PlateNo = fallback(req.PlateNo, vehicle.PlateNo)
		item.Unit = fallback(req.Unit, order.Unit)
		item.NetWeight = measuredNetWeight(item, 0)
		if item.NetWeight <= 0 {
			return fmt.Errorf("退料净重必须大于 0")
		}
		item.Remark = fallback(req.Remark, "工地退料回站")
		appendScaleTicketWithWeights(data, item)
		addAudit(data, session.User.Username, "create", "scale_ticket", item.ID, item.TicketNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "ticket.return.created")
}

func (a *App) createWasteTicket(w http.ResponseWriter, r *http.Request, session Session) {
	var req ScaleTicket
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid waste ticket")
		return
	}
	var item ScaleTicket
	err := a.store.Mutate(func(data *AppData) error {
		if req.SiteID == 0 {
			return fmt.Errorf("废料站点不能为空")
		}
		if _, ok := findMaterial(*data, req.MaterialID); !ok {
			return fmt.Errorf("废料物料不存在")
		}
		item = baseScaleTicket(data, "waste_out", req.SiteID, req)
		item.MaterialID = req.MaterialID
		item.PlateNo = fallback(req.PlateNo, "废料车")
		item.Unit = fallback(req.Unit, "t")
		item.NetWeight = measuredNetWeight(item, 0)
		if item.NetWeight <= 0 {
			return fmt.Errorf("废料净重必须大于 0")
		}
		balance, _, ok := decreaseInventory(data, req.SiteID, req.MaterialID, item.NetWeight)
		if !ok {
			return fmt.Errorf("废料出库库存不足")
		}
		item.Remark = fallback(req.Remark, "废料出库过磅")
		appendScaleTicketWithWeights(data, item)
		flowID := nextID(data, "inventoryFlow")
		data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
			ID: flowID, FlowNo: number("IF", flowID), SiteID: req.SiteID, MaterialID: req.MaterialID,
			SourceType: "waste_ticket", SourceID: item.ID, Direction: "out", Quantity: item.NetWeight,
			BalanceQty: balance, Remark: item.Remark, CreatedAt: item.CreatedAt,
		})
		addAudit(data, session.User.Username, "create", "scale_ticket", item.ID, item.TicketNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "ticket.waste.created")
}

func baseScaleTicket(data *AppData, ticketType string, siteID int64, req ScaleTicket) ScaleTicket {
	id := nextID(data, "ticket")
	ticketNo := number("ST", id)
	return ScaleTicket{
		ID: id, TicketNo: ticketNo, TicketType: ticketType, SiteID: siteID,
		GrossWeight: req.GrossWeight, TareWeight: req.TareWeight, Unit: fallback(req.Unit, "t"),
		SnapshotURL: fallback(req.SnapshotURL, scaleCaptureURL(defaultScaleDeviceForSite(*data, siteID).Code, ticketNo, "gross")),
		PrintCount:  1, SignStatus: "not_required", SettlementStatus: "pending", Status: "valid", CreatedAt: nowString(),
	}
}

func appendScaleTicketWithWeights(data *AppData, item ScaleTicket) {
	data.ScaleTickets = append(data.ScaleTickets, item)
	device := defaultScaleDeviceForSite(*data, item.SiteID)
	appendWeightRecord(data, device.ID, item.ID, item.PlateNo, item.TareWeight, "tare", scaleCaptureURL(device.Code, item.TicketNo, "tare"), item.CreatedAt)
	appendWeightRecord(data, device.ID, item.ID, item.PlateNo, item.GrossWeight, "gross", item.SnapshotURL, item.CreatedAt)
}

func measuredNetWeight(item ScaleTicket, fallbackQty float64) float64 {
	net := round(item.GrossWeight - item.TareWeight)
	if net > 0 {
		return net
	}
	return round(fallbackQty)
}

func findScaleTicket(data AppData, id int64) (ScaleTicket, bool) {
	for _, item := range data.ScaleTickets {
		if item.ID == id {
			return item, true
		}
	}
	return ScaleTicket{}, false
}
