package appliance

import (
	"errors"
	"math"
	"sort"
)

func findCustomer(data AppData, id int64) (Customer, bool) {
	for _, item := range data.Customers {
		if item.ID == id {
			return item, true
		}
	}
	return Customer{}, false
}

func findProject(data AppData, id int64) (Project, bool) {
	for _, item := range data.Projects {
		if item.ID == id {
			return item, true
		}
	}
	return Project{}, false
}

func findProduct(data AppData, id int64) (Product, bool) {
	for _, item := range data.Products {
		if item.ID == id {
			return item, true
		}
	}
	return Product{}, false
}

func findOrder(data AppData, id int64) (SalesOrder, bool) {
	for _, item := range data.Orders {
		if item.ID == id {
			return item, true
		}
	}
	return SalesOrder{}, false
}

func findVehicle(data AppData, id int64) (Vehicle, bool) {
	for _, item := range data.Vehicles {
		if item.ID == id {
			return item, true
		}
	}
	return Vehicle{}, false
}

func findVehicleByPlate(data AppData, plate string) (Vehicle, bool) {
	for _, item := range data.Vehicles {
		if item.PlateNo == plate {
			return item, true
		}
	}
	return Vehicle{}, false
}

func findDispatch(data AppData, id int64) (DispatchOrder, bool) {
	for _, item := range data.DispatchOrders {
		if item.ID == id {
			return item, true
		}
	}
	return DispatchOrder{}, false
}

func activeContract(data AppData, customerID, projectID, productID int64) (Contract, bool) {
	for _, contract := range data.Contracts {
		if contract.CustomerID == customerID && contract.ProjectID == projectID && contract.Status == "active" {
			for _, item := range contract.Items {
				if item.ProductID == productID {
					return contract, true
				}
			}
		}
	}
	return Contract{}, false
}

func approvedMixDesign(data AppData, productID, siteID int64) (MixDesign, bool) {
	for _, item := range data.MixDesigns {
		if item.ProductID == productID && mixDesignMatchesSite(item, siteID) && item.Status == "approved" && item.IsCurrent {
			return item, true
		}
	}
	for _, item := range data.MixDesigns {
		if item.ProductID == productID && mixDesignMatchesSite(item, siteID) && item.Status == "approved" {
			return item, true
		}
	}
	return MixDesign{}, false
}

func mixDesignMatchesSite(item MixDesign, siteID int64) bool {
	return item.SiteID == siteID || item.SiteID == 0 || siteID == 0
}

func inventoryStatus(data AppData) string {
	for _, item := range data.Inventory {
		if item.AvailableStatus == "warning" {
			return "warning"
		}
	}
	return "ok"
}

func inventoryWarnings(data AppData) []InventoryItem {
	out := []InventoryItem{}
	for _, item := range data.Inventory {
		material, ok := findMaterial(data, item.MaterialID)
		if item.AvailableStatus == "warning" || (ok && item.Quantity < material.SafeStock) {
			out = append(out, item)
		}
	}
	return out
}

func findMaterial(data AppData, id int64) (Material, bool) {
	for _, item := range data.Materials {
		if item.ID == id {
			return item, true
		}
	}
	return Material{}, false
}

func lastOrders(orders []SalesOrder, limit int) []SalesOrder {
	copied := append([]SalesOrder(nil), orders...)
	sort.Slice(copied, func(i, j int) bool { return copied[i].ID > copied[j].ID })
	if len(copied) > limit {
		return copied[:limit]
	}
	return copied
}

func nextDispatchStatus(status string) string {
	flow := []string{"assigned", "accepted", "arrived_site", "loading", "loaded", "departed", "in_transit", "arrived_project", "unloading", "signed", "completed"}
	for i, item := range flow {
		if item == status && i < len(flow)-1 {
			return flow[i+1]
		}
	}
	if status == "" || status == "pending" {
		return "assigned"
	}
	return status
}

func vehicleStatusForDispatch(status string) string {
	switch status {
	case "assigned", "accepted":
		return "assigned"
	case "arrived_site":
		return "waiting_load"
	case "loading":
		return "loading"
	case "loaded", "departed", "in_transit":
		return "in_transit"
	case "arrived_project":
		return "arrived"
	case "unloading", "signed":
		return "unloading"
	case "completed":
		return "returning"
	case "cancelled":
		return "idle"
	default:
		return status
	}
}

func updateVehicleStatus(data *AppData, vehicleID int64, status string) {
	if status == "" {
		return
	}
	for i := range data.Vehicles {
		if data.Vehicles[i].ID == vehicleID {
			data.Vehicles[i].BusinessStatus = status
		}
	}
	for i := range data.LatestLocations {
		if data.LatestLocations[i].VehicleID == vehicleID {
			data.LatestLocations[i].TransportStatus = status
		}
	}
}

func upsertStatement(data *AppData, sign DeliverySign, order SalesOrder) {
	period := periodString()
	productID := order.ProductID
	unitPrice := order.UnitPrice
	if sign.LineID != 0 {
		if line, ok := findOrderLine(order, sign.LineID); ok {
			productID = line.ProductID
			unitPrice = line.UnitPrice
		}
	}
	item := StatementItem{
		SignID:    sign.ID,
		OrderID:   order.ID,
		LineID:    sign.LineID,
		TicketID:  sign.TicketID,
		ProductID: productID,
		Quantity:  sign.SignedQty,
		UnitPrice: unitPrice,
		Amount:    round(sign.SignedQty * unitPrice),
	}
	for i := range data.Statements {
		if data.Statements[i].CustomerID == order.CustomerID && data.Statements[i].ProjectID == order.ProjectID && data.Statements[i].Period == period && data.Statements[i].Status == "draft" {
			data.Statements[i].Items = append(data.Statements[i].Items, item)
			data.Statements[i].TotalQty = round(data.Statements[i].TotalQty + item.Quantity)
			data.Statements[i].TotalAmount = round(data.Statements[i].TotalAmount + item.Amount)
			return
		}
	}
	statement := Statement{
		ID:          nextID(data, "statement"),
		StatementNo: number("CS", data.Next["statement"]),
		CustomerID:  order.CustomerID,
		ProjectID:   order.ProjectID,
		Period:      period,
		TotalQty:    item.Quantity,
		TotalAmount: item.Amount,
		Status:      "draft",
		Items:       []StatementItem{item},
	}
	data.Statements = append(data.Statements, statement)
}

func upsertLatestLocation(data *AppData, latest VehicleLatestLocation) {
	for i := range data.LatestLocations {
		if data.LatestLocations[i].VehicleID == latest.VehicleID {
			if latest.TransportStatus == "" {
				latest.TransportStatus = data.LatestLocations[i].TransportStatus
			}
			data.LatestLocations[i] = latest
			return
		}
	}
	data.LatestLocations = append(data.LatestLocations, latest)
}

func currentTripContext(data AppData, vehicleID int64) (dispatchID, orderID, projectID, customerID int64) {
	for _, dispatch := range data.DispatchOrders {
		if dispatch.VehicleID != vehicleID {
			continue
		}
		if dispatch.Status == "completed" || dispatch.Status == "cancelled" {
			continue
		}
		dispatchID = dispatch.ID
		orderID = dispatch.OrderID
		projectID = dispatch.ProjectID
		if order, ok := findOrder(data, dispatch.OrderID); ok {
			customerID = order.CustomerID
		}
		return
	}
	return
}

func inferAddress(data AppData, longitude, latitude float64) string {
	nearestName := "未知道路"
	nearestDistance := math.MaxFloat64
	for _, site := range data.Sites {
		d := distanceMeters(latitude, longitude, site.Latitude, site.Longitude)
		if d < nearestDistance {
			nearestDistance = d
			nearestName = site.Name
		}
	}
	for _, project := range data.Projects {
		d := distanceMeters(latitude, longitude, project.Latitude, project.Longitude)
		if d < nearestDistance {
			nearestDistance = d
			nearestName = project.Name
		}
	}
	if nearestDistance < 700 {
		return nearestName
	}
	return "运输途中"
}

func createFenceEvents(data *AppData, event VehicleLocationEvent) {
	for _, fence := range data.GeoFences {
		fence = normalizeGeoFence(fence)
		if fence.Status != "active" {
			continue
		}
		inside := geoFenceContains(fence, event.Longitude, event.Latitude)
		wasInside := latestFenceInsideState(data.GeoFenceEvents, event.VehicleID, fence.ID)
		if inside == wasInside {
			continue
		}
		eventType := "leave"
		if inside {
			eventType = "enter"
		}
		data.GeoFenceEvents = append(data.GeoFenceEvents, GeoFenceEvent{
			ID:         nextID(data, "fenceEvent"),
			VehicleID:  event.VehicleID,
			FenceID:    fence.ID,
			EventType:  eventType,
			DispatchID: event.DispatchID,
			EventTime:  fallback(event.LocationTime, nowString()),
		})
		if inside && event.DispatchID > 0 {
			updateDispatchForFenceEnter(data, event.VehicleID, event.DispatchID, fence)
		}
	}
}

func updateDispatchForFenceEnter(data *AppData, vehicleID, dispatchID int64, fence GeoFence) {
	for i := range data.DispatchOrders {
		if data.DispatchOrders[i].ID != dispatchID {
			continue
		}
		if fence.Type == "site" && data.DispatchOrders[i].Status == "assigned" {
			data.DispatchOrders[i].Status = "arrived_site"
			updateVehicleStatus(data, vehicleID, "waiting_load")
		}
		if fence.Type == "project" && (data.DispatchOrders[i].Status == "in_transit" || data.DispatchOrders[i].Status == "departed") {
			data.DispatchOrders[i].Status = "arrived_project"
			updateVehicleStatus(data, vehicleID, "arrived")
		}
	}
}

func latestFenceInsideState(events []GeoFenceEvent, vehicleID, fenceID int64) bool {
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.VehicleID != vehicleID || event.FenceID != fenceID {
			continue
		}
		return event.EventType == "enter"
	}
	return false
}

func normalizeGeoFence(fence GeoFence) GeoFence {
	if fence.Shape == "" {
		fence.Shape = "circle"
	}
	if fence.Status == "" {
		fence.Status = "active"
	}
	if fence.Shape == "circle" && fence.Radius <= 0 {
		fence.Radius = 300
	}
	return fence
}

func geoFenceContains(fence GeoFence, longitude, latitude float64) bool {
	fence = normalizeGeoFence(fence)
	if fence.Shape == "polygon" && len(fence.Polygon) >= 3 {
		return pointInPolygon(longitude, latitude, fence.Polygon)
	}
	return distanceMeters(latitude, longitude, fence.Latitude, fence.Longitude) <= fence.Radius
}

func pointInPolygon(longitude, latitude float64, polygon []GeoPoint) bool {
	inside := false
	j := len(polygon) - 1
	for i := range polygon {
		xi, yi := polygon[i].Longitude, polygon[i].Latitude
		xj, yj := polygon[j].Longitude, polygon[j].Latitude
		intersects := ((yi > latitude) != (yj > latitude)) &&
			(longitude < (xj-xi)*(latitude-yi)/(yj-yi)+xi)
		if intersects {
			inside = !inside
		}
		j = i
	}
	return inside
}

func findGeoFence(data AppData, id int64) (GeoFence, bool) {
	for _, item := range data.GeoFences {
		if item.ID == id {
			return normalizeGeoFence(item), true
		}
	}
	return GeoFence{}, false
}

func ensureGeoFenceMatchesMaster(data AppData, fence *GeoFence) error {
	switch fence.Type {
	case "site":
		if fence.SiteID == 0 {
			return errors.New("站点围栏必须选择厂区/站点")
		}
		site, ok := findSite(data, fence.SiteID)
		if !ok {
			return errors.New("厂区/站点不存在")
		}
		fence.ProjectID = 0
		if fence.Name == "" {
			fence.Name = site.Name + "围栏"
		}
		if fence.Longitude == 0 && fence.Latitude == 0 {
			fence.Longitude = site.Longitude
			fence.Latitude = site.Latitude
		}
	case "project":
		if fence.ProjectID == 0 {
			return errors.New("工地围栏必须选择项目/工地")
		}
		project, ok := findProject(data, fence.ProjectID)
		if !ok {
			return errors.New("项目/工地不存在")
		}
		fence.SiteID = 0
		if fence.Name == "" {
			fence.Name = project.Name + "围栏"
		}
		if fence.Longitude == 0 && fence.Latitude == 0 {
			fence.Longitude = project.Longitude
			fence.Latitude = project.Latitude
		}
	case "yard", "customer", "supplier", "route":
		if fence.Name == "" {
			return errors.New("自定义围栏必须填写名称")
		}
	default:
		return errors.New("围栏类型不支持")
	}
	return nil
}

func validateGeoFence(data AppData, fence *GeoFence) error {
	fence.Type = fallback(fence.Type, "site")
	fence.Shape = fallback(fence.Shape, "circle")
	fence.Status = fallback(fence.Status, "active")
	if fence.Status != "active" && fence.Status != "inactive" {
		return errors.New("围栏状态不支持")
	}
	if fence.Shape != "circle" && fence.Shape != "polygon" {
		return errors.New("围栏形状不支持")
	}
	if err := ensureGeoFenceMatchesMaster(data, fence); err != nil {
		return err
	}
	if fence.Shape == "polygon" {
		if len(fence.Polygon) < 3 {
			return errors.New("多边形围栏至少需要 3 个坐标点")
		}
		if fence.Longitude == 0 && fence.Latitude == 0 {
			var longitude, latitude float64
			for _, point := range fence.Polygon {
				longitude += point.Longitude
				latitude += point.Latitude
			}
			fence.Longitude = longitude / float64(len(fence.Polygon))
			fence.Latitude = latitude / float64(len(fence.Polygon))
		}
		return nil
	}
	if fence.Radius <= 0 {
		fence.Radius = 300
	}
	if fence.Longitude == 0 || fence.Latitude == 0 {
		return errors.New("圆形围栏必须填写中心经纬度")
	}
	fence.Polygon = nil
	return nil
}

func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000
	toRad := func(v float64) float64 { return v * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func vehicleEfficiency(data AppData) []map[string]interface{} {
	counts := map[int64]int{}
	for _, dispatch := range data.DispatchOrders {
		if dispatch.Status == "completed" {
			counts[dispatch.VehicleID]++
		}
	}
	out := []map[string]interface{}{}
	for _, vehicle := range data.Vehicles {
		out = append(out, map[string]interface{}{
			"vehicleId":      vehicle.ID,
			"plateNo":        vehicle.PlateNo,
			"completedTrips": counts[vehicle.ID],
			"businessStatus": vehicle.BusinessStatus,
			"onlineStatus":   vehicle.OnlineStatus,
		})
	}
	return out
}
