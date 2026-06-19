package appliance

import (
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

func (a *App) dispatchCenter(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if (len(parts) == 0 || (len(parts) == 1 && parts[0] == "overview")) && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, buildDispatchCenterOverview(data))
		return
	}
	writeError(w, http.StatusNotFound, "unknown dispatch center route")
}

func buildDispatchCenterOverview(data AppData) DispatchCenterOverview {
	locationsByVehicle := map[int64]VehicleLatestLocation{}
	for _, location := range data.LatestLocations {
		locationsByVehicle[location.VehicleID] = location
	}

	activeVehicleIDs := map[int64]bool{}
	progressByOrder := map[int64]*DispatchCenterSiteProgress{}
	overview := DispatchCenterOverview{
		SiteProgress:      []DispatchCenterSiteProgress{},
		VehicleQueue:      []DispatchCenterQueueItem{},
		ProductionTasks:   []DispatchCenterProductionTask{},
		AvailableVehicles: []DispatchCenterVehicle{},
		LatestLocations:   data.LatestLocations,
	}

	for _, vehicle := range data.Vehicles {
		overview.KPIs.TotalVehicles++
		online := vehicle.OnlineStatus
		business := vehicle.BusinessStatus
		if location, ok := locationsByVehicle[vehicle.ID]; ok {
			online = fallback(online, location.OnlineStatus)
			business = fallback(business, location.TransportStatus)
		}
		if online == "online" {
			overview.KPIs.OnlineVehicles++
		}
		switch business {
		case "", "idle", "returning":
			overview.KPIs.IdleVehicles++
		case "waiting_load", "assigned", "accepted":
			overview.KPIs.QueueVehicles++
		case "loading":
			overview.KPIs.LoadingVehicles++
		case "loaded", "departed", "in_transit":
			overview.KPIs.InTransitVehicles++
		case "arrived", "arrived_project", "unloading", "signed":
			overview.KPIs.ArrivedVehicles++
		}
	}
	overview.KPIs.VehicleOnlineRate = percent(overview.KPIs.OnlineVehicles, overview.KPIs.TotalVehicles)

	for _, order := range data.Orders {
		if !openSupplyOrder(order) {
			continue
		}
		item := newDispatchCenterSiteProgress(data, order)
		progressByOrder[order.ID] = &item
		overview.KPIs.OpenSupplyOrders++
	}

	for _, plan := range data.ProductionPlans {
		item, ok := progressByOrder[plan.OrderID]
		if !ok {
			continue
		}
		item.ProducedQty = round(item.ProducedQty + plan.ProducedQty)
		item.ProducedPercent = percentFloat(item.ProducedQty, item.PlanQuantity)
	}

	for _, dispatch := range data.DispatchOrders {
		progress, hasProgress := progressByOrder[dispatch.OrderID]
		var queueItem DispatchCenterQueueItem
		if activeDispatchStatus(dispatch.Status) {
			activeVehicleIDs[dispatch.VehicleID] = true
			overview.KPIs.ActiveDispatches++
			queueItem = dispatchCenterQueueItem(data, locationsByVehicle, dispatch)
			overview.VehicleQueue = append(overview.VehicleQueue, queueItem)
		}
		if !hasProgress {
			continue
		}
		progress.LoadedQty = round(progress.LoadedQty + dispatch.LoadedQty)
		if activeDispatchStatus(dispatch.Status) {
			progress.ActiveDispatches++
			if queueDispatchStatus(dispatch.Status) || dispatch.Status == "loading" {
				progress.QueueVehicles++
			}
			if inTransitDispatchStatus(dispatch.Status) {
				progress.InTransitVehicles++
			}
			if queueItem.ETA != "" && (progress.NextETA == "" || queueItem.ETA < progress.NextETA) {
				progress.NextETA = queueItem.ETA
			}
		}
	}

	for orderID, progress := range progressByOrder {
		if order, ok := findOrder(data, orderID); ok {
			progress.DispatchedQty = round(order.DispatchedQty)
			progress.SignedQty = round(order.SignedQty)
			progress.RemainingQty = round(order.PlanQuantity - order.SignedQty)
			progress.DispatchedPercent = percentFloat(progress.DispatchedQty, progress.PlanQuantity)
			progress.LoadedPercent = percentFloat(progress.LoadedQty, progress.PlanQuantity)
			progress.SignedPercent = percentFloat(progress.SignedQty, progress.PlanQuantity)
			progress.Status = supplyProgressStatus(order, *progress)
		}
		overview.SiteProgress = append(overview.SiteProgress, *progress)
	}

	for _, task := range data.ProductionTasks {
		if task.Status == "completed" {
			continue
		}
		overview.ProductionTasks = append(overview.ProductionTasks, dispatchCenterProductionTask(data, task))
		overview.KPIs.ActiveProductionTasks++
	}

	today := todayString()
	for _, vehicle := range data.Vehicles {
		if vehicle.Status != "active" || activeVehicleIDs[vehicle.ID] {
			continue
		}
		overview.AvailableVehicles = append(overview.AvailableVehicles, dispatchCenterVehicle(data, locationsByVehicle, vehicle, today))
	}

	sort.Slice(overview.SiteProgress, func(i, j int) bool {
		if overview.SiteProgress[i].Status != overview.SiteProgress[j].Status {
			return dispatchCenterStatusRank(overview.SiteProgress[i].Status) < dispatchCenterStatusRank(overview.SiteProgress[j].Status)
		}
		return overview.SiteProgress[i].OrderID > overview.SiteProgress[j].OrderID
	})
	sort.Slice(overview.VehicleQueue, func(i, j int) bool {
		if overview.VehicleQueue[i].Status != overview.VehicleQueue[j].Status {
			return dispatchQueueRank(overview.VehicleQueue[i].Status) < dispatchQueueRank(overview.VehicleQueue[j].Status)
		}
		if overview.VehicleQueue[i].ETA != overview.VehicleQueue[j].ETA {
			return overview.VehicleQueue[i].ETA < overview.VehicleQueue[j].ETA
		}
		return overview.VehicleQueue[i].DispatchID < overview.VehicleQueue[j].DispatchID
	})
	sort.Slice(overview.ProductionTasks, func(i, j int) bool {
		if overview.ProductionTasks[i].Status != overview.ProductionTasks[j].Status {
			return productionTaskRank(overview.ProductionTasks[i].Status) < productionTaskRank(overview.ProductionTasks[j].Status)
		}
		return overview.ProductionTasks[i].TaskID < overview.ProductionTasks[j].TaskID
	})
	sort.Slice(overview.AvailableVehicles, func(i, j int) bool {
		if overview.AvailableVehicles[i].OnlineStatus != overview.AvailableVehicles[j].OnlineStatus {
			return overview.AvailableVehicles[i].OnlineStatus == "online"
		}
		return overview.AvailableVehicles[i].VehicleID < overview.AvailableVehicles[j].VehicleID
	})

	return overview
}

func newDispatchCenterSiteProgress(data AppData, order SalesOrder) DispatchCenterSiteProgress {
	customer, _ := findCustomer(data, order.CustomerID)
	project, _ := findProject(data, order.ProjectID)
	site, _ := findSite(data, order.SiteID)
	return DispatchCenterSiteProgress{
		OrderID:       order.ID,
		OrderNo:       order.OrderNo,
		CustomerID:    order.CustomerID,
		CustomerName:  customer.Name,
		ProjectID:     order.ProjectID,
		ProjectName:   project.Name,
		SiteID:        order.SiteID,
		SiteName:      site.Name,
		ProductID:     order.ProductID,
		ProductName:   dispatchCenterProductName(data, order.ProductID),
		Unit:          order.Unit,
		PlanQuantity:  round(order.PlanQuantity),
		DispatchedQty: round(order.DispatchedQty),
		SignedQty:     round(order.SignedQty),
		RemainingQty:  round(order.PlanQuantity - order.SignedQty),
		Status:        order.Status,
	}
}

func dispatchCenterQueueItem(data AppData, locations map[int64]VehicleLatestLocation, dispatch DispatchOrder) DispatchCenterQueueItem {
	order, _ := findOrder(data, dispatch.OrderID)
	customer, _ := findCustomer(data, order.CustomerID)
	project, _ := findProject(data, dispatch.ProjectID)
	site, _ := findSite(data, dispatch.SiteID)
	vehicle, _ := findVehicle(data, dispatch.VehicleID)
	driver, _ := findDriver(data, dispatch.DriverID)
	location := locations[dispatch.VehicleID]
	eta := estimateDispatchETA(data, dispatch, location)
	return DispatchCenterQueueItem{
		DispatchID:     dispatch.ID,
		DispatchNo:     dispatch.DispatchNo,
		OrderID:        dispatch.OrderID,
		OrderNo:        order.OrderNo,
		CustomerName:   customer.Name,
		ProjectID:      dispatch.ProjectID,
		ProjectName:    project.Name,
		SiteID:         dispatch.SiteID,
		SiteName:       site.Name,
		ProductName:    dispatchCenterProductName(data, order.ProductID),
		VehicleID:      dispatch.VehicleID,
		PlateNo:        vehicle.PlateNo,
		DriverID:       dispatch.DriverID,
		DriverName:     driver.Name,
		QueueNo:        dispatch.QueueNo,
		ETA:            eta.ArrivalAt,
		PlannedETA:     dispatch.ETA,
		ETASource:      eta.Source,
		ETAMinutes:     eta.Minutes,
		ETADistanceKm:  eta.DistanceKm,
		ETAConfidence:  eta.Confidence,
		ETATarget:      eta.Target,
		ETASpeedKPH:    eta.SpeedKPH,
		Status:         dispatch.Status,
		PlanQuantity:   round(dispatch.PlanQuantity),
		LoadedQty:      round(dispatch.LoadedQty),
		SignedQty:      round(dispatch.SignedQty),
		OnlineStatus:   fallback(vehicle.OnlineStatus, location.OnlineStatus),
		BusinessStatus: fallback(vehicle.BusinessStatus, location.TransportStatus),
		LastLocationAt: location.LastLocationTime,
		UpdatedAt:      dispatch.UpdatedAt,
	}
}

type dispatchETAEstimate struct {
	ArrivalAt  string
	Source     string
	Minutes    float64
	DistanceKm float64
	Confidence string
	Target     string
	SpeedKPH   float64
}

type dispatchETATarget struct {
	Name      string
	Latitude  float64
	Longitude float64
}

func estimateDispatchETA(data AppData, dispatch DispatchOrder, location VehicleLatestLocation) dispatchETAEstimate {
	estimate := dispatchETAEstimate{
		ArrivalAt:  dispatch.ETA,
		Source:     "planned",
		Confidence: "low",
		Target:     "计划时间",
	}
	target, ok := resolveDispatchETATarget(data, dispatch)
	if ok {
		estimate.Target = target.Name
	}
	if !ok || location.VehicleID == 0 || location.Latitude == 0 || location.Longitude == 0 {
		return estimate
	}
	distance := distanceMeters(location.Latitude, location.Longitude, target.Latitude, target.Longitude)
	speed, source, confidence := dispatchETASpeed(data, dispatch.VehicleID, location)
	minutes := distance/1000/speed*60 + dispatchETABufferMinutes(dispatch.Status)
	if math.IsNaN(minutes) || math.IsInf(minutes, 0) || minutes < 0 {
		return estimate
	}
	estimate.ArrivalAt = time.Now().Add(time.Duration(math.Ceil(minutes)) * time.Minute).Format("2006-01-02 15:04:05")
	estimate.Source = source
	estimate.Minutes = round(minutes)
	estimate.DistanceKm = round(distance / 1000)
	estimate.Confidence = dispatchETAConfidence(location.LastLocationTime, confidence)
	estimate.SpeedKPH = round(speed)
	return estimate
}

func resolveDispatchETATarget(data AppData, dispatch DispatchOrder) (dispatchETATarget, bool) {
	if dispatchInTransitToProject(dispatch.Status) {
		project, ok := findProject(data, dispatch.ProjectID)
		if ok && project.Latitude != 0 && project.Longitude != 0 {
			return dispatchETATarget{Name: project.Name, Latitude: project.Latitude, Longitude: project.Longitude}, true
		}
	}
	site, ok := findSite(data, dispatch.SiteID)
	if ok && site.Latitude != 0 && site.Longitude != 0 {
		return dispatchETATarget{Name: site.Name, Latitude: site.Latitude, Longitude: site.Longitude}, true
	}
	project, ok := findProject(data, dispatch.ProjectID)
	if ok && project.Latitude != 0 && project.Longitude != 0 {
		return dispatchETATarget{Name: project.Name, Latitude: project.Latitude, Longitude: project.Longitude}, true
	}
	return dispatchETATarget{}, false
}

func dispatchInTransitToProject(status string) bool {
	switch status {
	case "loading", "loaded", "departed", "in_transit", "arrived_project", "unloading", "signed":
		return true
	default:
		return false
	}
}

func dispatchETASpeed(data AppData, vehicleID int64, location VehicleLatestLocation) (float64, string, string) {
	if location.Speed >= 5 {
		return location.Speed, "gps_current", "high"
	}
	events := []VehicleLocationEvent{}
	for _, event := range data.Locations {
		if event.VehicleID == vehicleID && event.Speed >= 5 {
			events = append(events, event)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].LocationTime == events[j].LocationTime {
			return events[i].ID > events[j].ID
		}
		return events[i].LocationTime > events[j].LocationTime
	})
	sum := 0.0
	count := 0
	for _, event := range events {
		sum += event.Speed
		count++
		if count == 6 {
			break
		}
	}
	if count > 0 {
		return sum / float64(count), "gps_average", "medium"
	}
	return 35, "rule_default", "low"
}

func dispatchETABufferMinutes(status string) float64 {
	switch status {
	case "assigned", "accepted":
		return 8
	case "arrived_site", "waiting_load":
		return 12
	case "loading":
		return 18
	case "loaded":
		return 3
	default:
		return 0
	}
}

func dispatchETAConfidence(locationTime, base string) string {
	parsed, err := parseTime(locationTime)
	if err != nil {
		return "low"
	}
	if time.Since(parsed) > 6*time.Hour && base == "high" {
		return "medium"
	}
	return base
}

func dispatchCenterProductionTask(data AppData, task ProductionTask) DispatchCenterProductionTask {
	plan, _ := findProductionPlan(data, task.PlanID)
	order, _ := findOrder(data, task.OrderID)
	customer, _ := findCustomer(data, order.CustomerID)
	project, _ := findProject(data, order.ProjectID)
	site, _ := findSite(data, task.SiteID)
	mix, _ := findMixDesign(data, task.MixDesignID)
	remaining := round(task.PlanQty - task.ProducedQty)
	return DispatchCenterProductionTask{
		TaskID:        task.ID,
		TaskNo:        task.TaskNo,
		PlanID:        task.PlanID,
		PlanNo:        plan.PlanNo,
		OrderID:       task.OrderID,
		OrderNo:       order.OrderNo,
		CustomerName:  customer.Name,
		ProjectID:     order.ProjectID,
		ProjectName:   project.Name,
		SiteID:        task.SiteID,
		SiteName:      site.Name,
		ProductID:     task.ProductID,
		ProductName:   dispatchCenterProductName(data, task.ProductID),
		MixDesignCode: mix.Code,
		PlanQty:       round(task.PlanQty),
		ProducedQty:   round(task.ProducedQty),
		RemainingQty:  remaining,
		Progress:      percentFloat(task.ProducedQty, task.PlanQty),
		Status:        task.Status,
		StartedAt:     task.StartedAt,
		UpdatedAt:     task.UpdatedAt,
	}
}

func dispatchCenterVehicle(data AppData, locations map[int64]VehicleLatestLocation, vehicle Vehicle, today string) DispatchCenterVehicle {
	site, _ := findSite(data, vehicle.SiteID)
	driver, _ := findDriver(data, vehicle.DriverID)
	location := locations[vehicle.ID]
	scheduleNo := ""
	remaining := 0.0
	for _, schedule := range data.DispatchSchedules {
		if schedule.VehicleID == vehicle.ID && schedule.Status == "active" && schedule.ShiftDate == today {
			scheduleNo = schedule.ScheduleNo
			remaining = round(schedule.CapacityQty - schedule.AssignedQty)
			break
		}
	}
	return DispatchCenterVehicle{
		VehicleID:         vehicle.ID,
		PlateNo:           vehicle.PlateNo,
		VehicleType:       vehicle.VehicleType,
		Capacity:          vehicle.Capacity,
		Carrier:           vehicle.Carrier,
		SiteID:            vehicle.SiteID,
		SiteName:          site.Name,
		DriverID:          vehicle.DriverID,
		DriverName:        driver.Name,
		OnlineStatus:      fallback(vehicle.OnlineStatus, location.OnlineStatus),
		BusinessStatus:    fallback(vehicle.BusinessStatus, location.TransportStatus),
		ScheduleNo:        scheduleNo,
		ScheduleRemaining: remaining,
		LastLocationAt:    location.LastLocationTime,
	}
}

func openSupplyOrder(order SalesOrder) bool {
	if order.PlanQuantity <= 0 || order.SignedQty >= order.PlanQuantity {
		return false
	}
	switch order.Status {
	case "cancelled", "completed", "rejected":
		return false
	default:
		return true
	}
}

func activeDispatchStatus(status string) bool {
	switch status {
	case "", "completed", "cancelled":
		return false
	default:
		return true
	}
}

func queueDispatchStatus(status string) bool {
	switch status {
	case "assigned", "accepted", "arrived_site":
		return true
	default:
		return false
	}
}

func inTransitDispatchStatus(status string) bool {
	switch status {
	case "loaded", "departed", "in_transit":
		return true
	default:
		return false
	}
}

func arrivedDispatchStatus(status string) bool {
	switch status {
	case "arrived_project", "unloading", "signed":
		return true
	default:
		return false
	}
}

func supplyProgressStatus(order SalesOrder, progress DispatchCenterSiteProgress) string {
	if progress.SignedQty >= progress.PlanQuantity {
		return "completed"
	}
	if progress.InTransitVehicles > 0 {
		return "in_transit"
	}
	if progress.QueueVehicles > 0 {
		return "dispatching"
	}
	if progress.ProducedQty > progress.SignedQty {
		return "produced"
	}
	return order.Status
}

func dispatchCenterProductName(data AppData, id int64) string {
	product, ok := findProduct(data, id)
	if !ok {
		return ""
	}
	if strings.TrimSpace(product.Spec) == "" {
		return product.Name
	}
	return product.Name + " " + product.Spec
}

func percentFloat(part, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return round(part / total * 100)
}

func dispatchCenterStatusRank(status string) int {
	switch status {
	case "in_transit":
		return 1
	case "dispatching":
		return 2
	case "produced", "producing", "scheduled":
		return 3
	case "approved":
		return 4
	default:
		return 5
	}
}

func dispatchQueueRank(status string) int {
	switch status {
	case "arrived_site":
		return 1
	case "loading":
		return 2
	case "loaded", "departed", "in_transit":
		return 3
	case "arrived_project", "unloading", "signed":
		return 4
	case "accepted":
		return 5
	case "assigned":
		return 6
	default:
		return 9
	}
}

func productionTaskRank(status string) int {
	switch status {
	case "running", "producing":
		return 1
	case "pending":
		return 2
	default:
		return 3
	}
}
