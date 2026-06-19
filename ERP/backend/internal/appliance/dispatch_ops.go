package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) createDispatchSchedule(w http.ResponseWriter, r *http.Request, session Session) {
	var item DispatchSchedule
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid dispatch schedule")
		return
	}
	var created DispatchSchedule
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findSite(*data, item.SiteID); !ok {
			return fmt.Errorf("站点不存在")
		}
		vehicle, ok := findVehicle(*data, item.VehicleID)
		if !ok {
			return fmt.Errorf("车辆不存在")
		}
		driverID := nonZeroInt(item.DriverID, vehicle.DriverID)
		if _, ok := findDriver(*data, driverID); !ok {
			return fmt.Errorf("司机不存在")
		}
		carrierID := item.CarrierID
		if carrierID == 0 {
			carrierID = carrierIDForVehicle(*data, vehicle)
		}
		if carrierID == 0 {
			return fmt.Errorf("承运商不存在")
		}
		shiftDate := fallback(item.ShiftDate, todayString())
		shift := fallback(item.Shift, "day")
		if activeScheduleConflict(*data, vehicle.ID, driverID, shiftDate, shift) {
			return fmt.Errorf("车辆或司机已有同班次排班")
		}
		item.ID = nextID(data, "dispatchSchedule")
		item.ScheduleNo = number("DS", item.ID)
		item.DriverID = driverID
		item.CarrierID = carrierID
		item.ShiftDate = shiftDate
		item.Shift = shift
		item.CapacityQty = nonZero(item.CapacityQty, 120)
		item.AssignedQty = 0
		item.Status = fallback(item.Status, "active")
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		data.DispatchSchedules = append(data.DispatchSchedules, item)
		created = item
		addAudit(data, session.User.Username, "create", "dispatch_schedule", item.ID, item.ScheduleNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, created, "dispatch.schedule.created")
}

func (a *App) generateCarrierSettlement(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		CarrierID   int64   `json:"carrierId"`
		Period      string  `json:"period"`
		RatePerTrip float64 `json:"ratePerTrip"`
		RatePerUnit float64 `json:"ratePerUnit"`
	}
	_ = readJSON(r, &req)
	var result struct {
		Settlement TransportSettlement       `json:"settlement"`
		Items      []TransportSettlementItem `json:"items"`
	}
	err := a.store.Mutate(func(data *AppData) error {
		carrierID := req.CarrierID
		if carrierID == 0 && len(data.Carriers) > 0 {
			carrierID = data.Carriers[0].ID
		}
		if _, ok := findCarrier(*data, carrierID); !ok {
			return fmt.Errorf("承运商不存在")
		}
		period := fallback(req.Period, periodString())
		ratePerTrip := nonZero(req.RatePerTrip, 480)
		ratePerUnit := req.RatePerUnit
		var items []TransportSettlementItem
		total := 0.0
		for _, dispatch := range data.DispatchOrders {
			if dispatch.Status != "completed" || settlementItemExists(*data, dispatch.ID) {
				continue
			}
			vehicle, ok := findVehicle(*data, dispatch.VehicleID)
			if !ok || carrierIDForVehicle(*data, vehicle) != carrierID {
				continue
			}
			quantity := nonZero(dispatch.SignedQty, dispatch.PlanQuantity)
			amount := ratePerTrip
			if ratePerUnit > 0 {
				amount = round(quantity * ratePerUnit)
			}
			itemID := nextID(data, "transportSettlementItem")
			item := TransportSettlementItem{
				ID: itemID, DispatchID: dispatch.ID, DispatchNo: dispatch.DispatchNo, CarrierID: carrierID,
				VehicleID: dispatch.VehicleID, DriverID: dispatch.DriverID, Quantity: quantity, Amount: amount,
				Status: "pending", CreatedAt: nowString(),
			}
			items = append(items, item)
			total = round(total + amount)
		}
		if len(items) == 0 {
			return fmt.Errorf("没有可生成对账的已完成派车单")
		}
		settlementID := nextID(data, "transportSettlement")
		settlement := TransportSettlement{
			ID: settlementID, SettlementNo: number("TS", settlementID), CarrierID: carrierID,
			Period: period, TripCount: len(items), Amount: total, Status: "draft",
		}
		for i := range items {
			items[i].SettlementID = settlement.ID
			data.TransportSettlementItems = append(data.TransportSettlementItems, items[i])
		}
		data.TransportSettlements = append(data.TransportSettlements, settlement)
		result.Settlement = settlement
		result.Items = items
		addAudit(data, session.User.Username, "generate", "transport_settlement", settlement.ID, settlement.SettlementNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, result, "dispatch.carrier_settlement.generated")
}

func reserveDispatchSchedule(data *AppData, dispatch DispatchOrder) error {
	if len(data.DispatchSchedules) == 0 {
		return nil
	}
	shiftDate := todayString()
	if order, ok := findOrder(*data, dispatch.OrderID); ok && len(order.PlanTime) >= 10 {
		shiftDate = order.PlanTime[:10]
	}
	index := matchingDispatchScheduleIndex(*data, dispatch.VehicleID, dispatch.DriverID, shiftDate)
	if index < 0 {
		return nil
	}
	schedule := data.DispatchSchedules[index]
	if schedule.Status != "active" {
		return fmt.Errorf("车辆排班不可用")
	}
	if schedule.AssignedQty+dispatch.PlanQuantity > schedule.CapacityQty {
		return fmt.Errorf("车辆排班容量不足")
	}
	data.DispatchSchedules[index].AssignedQty = round(schedule.AssignedQty + dispatch.PlanQuantity)
	data.DispatchSchedules[index].UpdatedAt = nowString()
	return nil
}

func activeScheduleConflict(data AppData, vehicleID, driverID int64, shiftDate, shift string) bool {
	for _, item := range data.DispatchSchedules {
		if item.Status != "active" || item.ShiftDate != shiftDate || item.Shift != shift {
			continue
		}
		if item.VehicleID == vehicleID || item.DriverID == driverID {
			return true
		}
	}
	return false
}

func matchingDispatchScheduleIndex(data AppData, vehicleID, driverID int64, shiftDate string) int {
	for i, item := range data.DispatchSchedules {
		if item.VehicleID == vehicleID && item.DriverID == driverID && item.ShiftDate == shiftDate {
			return i
		}
	}
	for i, item := range data.DispatchSchedules {
		if item.VehicleID == vehicleID && item.ShiftDate == shiftDate {
			return i
		}
	}
	return -1
}

func carrierIDForVehicle(data AppData, vehicle Vehicle) int64 {
	needle := strings.TrimSpace(vehicle.Carrier)
	for _, carrier := range data.Carriers {
		if carrier.ID > 0 && (carrier.Name == needle || fmt.Sprint(carrier.ID) == needle) {
			return carrier.ID
		}
	}
	if len(data.Carriers) > 0 {
		return data.Carriers[0].ID
	}
	return 0
}

func findCarrier(data AppData, id int64) (Carrier, bool) {
	for _, item := range data.Carriers {
		if item.ID == id {
			return item, true
		}
	}
	return Carrier{}, false
}

func findDriver(data AppData, id int64) (Driver, bool) {
	for _, item := range data.Drivers {
		if item.ID == id {
			return item, true
		}
	}
	return Driver{}, false
}

func findSite(data AppData, id int64) (Site, bool) {
	for _, item := range data.Sites {
		if item.ID == id {
			return item, true
		}
	}
	return Site{}, false
}

func settlementItemExists(data AppData, dispatchID int64) bool {
	for _, item := range data.TransportSettlementItems {
		if item.DispatchID == dispatchID && item.Status != "cancelled" {
			return true
		}
	}
	return false
}
