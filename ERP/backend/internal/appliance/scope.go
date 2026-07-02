package appliance

import "strings"

func scopedData(data AppData, user User) AppData {
	switch normalizeDataScope(roleDataScope(data.Roles, user.RoleCode)) {
	case "customer":
		return applyFieldPolicies(scopeCustomer(data, user.CustomerID), user)
	case "driver":
		return applyFieldPolicies(scopeDriver(data, user.DriverID), user)
	case "site":
		return applyFieldPolicies(scopeSite(data, user.SiteID), user)
	case "company":
		return applyFieldPolicies(scopeCompany(data, user.CompanyID), user)
	case "device":
		return applyFieldPolicies(data, user)
	case "group":
		return applyFieldPolicies(data, user)
	}
	switch user.RoleCode {
	case "customer":
		if user.CustomerID > 0 {
			return applyFieldPolicies(scopeCustomer(data, user.CustomerID), user)
		}
	case "driver":
		if user.DriverID > 0 {
			return applyFieldPolicies(scopeDriver(data, user.DriverID), user)
		}
	case "dispatcher", "quality":
		if user.SiteID > 0 {
			return applyFieldPolicies(scopeSite(data, user.SiteID), user)
		}
	}
	return applyFieldPolicies(data, user)
}

func roleDataScope(roles []Role, roleCode string) string {
	for _, role := range roles {
		if role.Code == roleCode {
			return role.DataScope
		}
	}
	return ""
}

func normalizeDataScope(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "platform", "tenant", "group":
		return "group"
	case "company", "site", "customer", "driver", "device":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "group"
	}
}

func scopeCustomer(data AppData, customerID int64) AppData {
	data.Customers = filter(data.Customers, func(item Customer) bool { return item.ID == customerID })
	data.CustomerContacts = filter(data.CustomerContacts, func(item CustomerContact) bool { return item.CustomerID == customerID })
	data.CustomerBlacklists = filter(data.CustomerBlacklists, func(item CustomerBlacklist) bool { return item.CustomerID == customerID })
	data.CustomerProfiles = filter(data.CustomerProfiles, func(item CustomerProfile) bool { return item.CustomerID == customerID })
	data.CustomerComplaints = filter(data.CustomerComplaints, func(item CustomerComplaint) bool { return item.CustomerID == customerID })
	data.Projects = filter(data.Projects, func(item Project) bool { return item.CustomerID == customerID })
	data.Contracts = filter(data.Contracts, func(item Contract) bool { return item.CustomerID == customerID })
	data.ContractAttachments = filter(data.ContractAttachments, func(item ContractAttachment) bool { return item.CustomerID == customerID })
	data.Orders = filter(data.Orders, func(item SalesOrder) bool { return item.CustomerID == customerID })
	data.ProductionPlans = filter(data.ProductionPlans, func(item ProductionPlan) bool {
		order, ok := findOrder(data, item.OrderID)
		return ok && order.CustomerID == customerID
	})
	visiblePlans := productionPlanIDs(data.ProductionPlans)
	data.ProductionTasks = filter(data.ProductionTasks, func(item ProductionTask) bool { return visiblePlans[item.PlanID] })
	data.ProductionBatches = filter(data.ProductionBatches, func(item ProductionBatch) bool { return visiblePlans[item.PlanID] })
	data.ProductionReports = nil
	data.QualityInspections = nil
	data.QualitySamples = nil
	data.RawMaterialInspections = nil
	data.MixDesignPlantProfiles = nil
	data.MixDesignTrialRuns = nil
	data.LaboratorySamples = nil
	data.LaboratoryTests = nil
	data.LaboratoryEquipment = nil
	data.LaboratoryCalibrations = nil
	data.QualityExceptions = nil
	data.DispatchOrders = filter(data.DispatchOrders, func(item DispatchOrder) bool {
		order, ok := findOrder(data, item.OrderID)
		return ok && order.CustomerID == customerID
	})
	customerDispatchIDs := dispatchOrderIDs(data.DispatchOrders)
	data.DispatchSchedules = nil
	data.TransportSettlementItems = filter(data.TransportSettlementItems, func(item TransportSettlementItem) bool { return customerDispatchIDs[item.DispatchID] })
	data.TransportSettlements = nil
	data.ScaleTickets = filter(data.ScaleTickets, func(item ScaleTicket) bool {
		order, ok := findOrder(data, item.OrderID)
		return ok && order.CustomerID == customerID
	})
	data.DeliverySigns = filter(data.DeliverySigns, func(item DeliverySign) bool { return item.CustomerID == customerID })
	data.DeliverySignLinks = filter(data.DeliverySignLinks, func(item DeliverySignLink) bool { return item.CustomerID == customerID })
	customerSignIDs := deliverySignIDs(data.DeliverySigns)
	data.DeliverySignAttachments = filter(data.DeliverySignAttachments, func(item DeliverySignAttachment) bool {
		return customerSignIDs[item.SignID] || customerDispatchIDs[item.DispatchID]
	})
	data.Statements = filter(data.Statements, func(item Statement) bool { return item.CustomerID == customerID })
	data.SalesInvoices = filter(data.SalesInvoices, func(item SalesInvoice) bool { return item.CustomerID == customerID })
	visibleInvoices := invoiceIDs(data.SalesInvoices)
	data.RedLetterInfos = filter(data.RedLetterInfos, func(item RedLetterInfo) bool {
		return item.CustomerID == customerID || visibleInvoices[item.OriginalInvoiceID] || visibleInvoices[item.RedInvoiceID]
	})
	data.TaxGatewaySubmissions = filter(data.TaxGatewaySubmissions, func(item TaxGatewaySubmission) bool { return visibleInvoices[item.InvoiceID] })
	data.Receivables = filter(data.Receivables, func(item Receivable) bool { return item.CustomerID == customerID })
	data.Receipts = filter(data.Receipts, func(item Receipt) bool { return item.CustomerID == customerID })
	data.PaymentPlans = filter(data.PaymentPlans, func(item PaymentPlan) bool { return item.CustomerID == customerID })
	data.CollectionTasks = filter(data.CollectionTasks, func(item CollectionTask) bool { return item.CustomerID == customerID })
	data.CollectionDispatches = filter(data.CollectionDispatches, func(item CollectionDispatch) bool { return item.CustomerID == customerID })
	data.LatestLocations = filter(data.LatestLocations, func(item VehicleLatestLocation) bool { return item.CurrentCustomerID == customerID })
	data.VehicleAlarms = filter(data.VehicleAlarms, func(item VehicleAlarm) bool {
		dispatch, ok := findDispatch(data, item.DispatchID)
		if !ok {
			return false
		}
		order, ok := findOrder(data, dispatch.OrderID)
		return ok && order.CustomerID == customerID
	})
	data.Locations = filter(data.Locations, func(item VehicleLocationEvent) bool {
		dispatch, ok := findDispatch(data, item.DispatchID)
		if !ok {
			return false
		}
		order, ok := findOrder(data, dispatch.OrderID)
		return ok && order.CustomerID == customerID
	})
	data.Vehicles = filter(data.Vehicles, func(item Vehicle) bool {
		for _, dispatch := range data.DispatchOrders {
			if dispatch.VehicleID == item.ID {
				return true
			}
		}
		for _, latest := range data.LatestLocations {
			if latest.VehicleID == item.ID && latest.CurrentCustomerID == customerID {
				return true
			}
		}
		return false
	})
	ids := collectScopeIDs(data)
	ids.customers[customerID] = true
	data = filterRelatedMasters(data, ids)
	data = filterTicketArtifacts(data)
	data.ApprovalTasks = filterApprovalTasksForScopedData(data)
	return data
}

func scopeDriver(data AppData, driverID int64) AppData {
	data.Drivers = filter(data.Drivers, func(item Driver) bool { return item.ID == driverID })
	data.Vehicles = filter(data.Vehicles, func(item Vehicle) bool { return item.DriverID == driverID })
	data.VehicleDevices = filter(data.VehicleDevices, func(item VehicleDevice) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.DriverID == driverID
	})
	data.DispatchOrders = filter(data.DispatchOrders, func(item DispatchOrder) bool { return item.DriverID == driverID })
	visibleOrders := orderIDsFromDispatches(data.DispatchOrders)
	driverDispatchIDs := dispatchOrderIDs(data.DispatchOrders)
	data.DispatchSchedules = filter(data.DispatchSchedules, func(item DispatchSchedule) bool { return item.DriverID == driverID })
	data.TransportSettlementItems = filter(data.TransportSettlementItems, func(item TransportSettlementItem) bool { return driverDispatchIDs[item.DispatchID] })
	data.TransportSettlements = nil
	data.Orders = filter(data.Orders, func(item SalesOrder) bool { return visibleOrders[item.ID] })
	data.ProductionPlans = filter(data.ProductionPlans, func(item ProductionPlan) bool { return visibleOrders[item.OrderID] })
	visiblePlans := productionPlanIDs(data.ProductionPlans)
	data.ProductionTasks = filter(data.ProductionTasks, func(item ProductionTask) bool { return visiblePlans[item.PlanID] })
	data.ProductionBatches = filter(data.ProductionBatches, func(item ProductionBatch) bool { return visiblePlans[item.PlanID] })
	data.ProductionReports = nil
	data.QualityInspections = nil
	data.QualitySamples = nil
	data.RawMaterialInspections = nil
	data.MixDesignPlantProfiles = nil
	data.MixDesignTrialRuns = nil
	data.LaboratorySamples = nil
	data.LaboratoryTests = nil
	data.LaboratoryEquipment = nil
	data.LaboratoryCalibrations = nil
	data.QualityExceptions = nil
	data.ScaleTickets = filter(data.ScaleTickets, func(item ScaleTicket) bool {
		dispatch, ok := findDispatch(data, item.DispatchID)
		return ok && dispatch.DriverID == driverID
	})
	data.DeliverySigns = filter(data.DeliverySigns, func(item DeliverySign) bool {
		dispatch, ok := findDispatch(data, item.DispatchID)
		return ok && dispatch.DriverID == driverID
	})
	data.DeliverySignLinks = filter(data.DeliverySignLinks, func(item DeliverySignLink) bool { return driverDispatchIDs[item.DispatchID] })
	driverSignIDs := deliverySignIDs(data.DeliverySigns)
	data.DeliverySignAttachments = filter(data.DeliverySignAttachments, func(item DeliverySignAttachment) bool {
		return driverSignIDs[item.SignID] || driverDispatchIDs[item.DispatchID]
	})
	data.LatestLocations = filter(data.LatestLocations, func(item VehicleLatestLocation) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.DriverID == driverID
	})
	data.Locations = filter(data.Locations, func(item VehicleLocationEvent) bool { return item.DriverID == driverID })
	data.VehicleAlarms = filter(data.VehicleAlarms, func(item VehicleAlarm) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.DriverID == driverID
	})
	ids := collectScopeIDs(data)
	data = filterRelatedMasters(data, ids)
	data = filterTicketArtifacts(data)
	data.ApprovalTasks = filterApprovalTasksForScopedData(data)
	return data
}

func scopeSite(data AppData, siteID int64) AppData {
	data.Sites = filter(data.Sites, func(item Site) bool { return item.ID == siteID })
	data.Plants = filter(data.Plants, func(item Plant) bool { return item.SiteID == siteID })
	data.PlantBufferLocations = filter(data.PlantBufferLocations, func(item PlantBufferLocation) bool { return item.SiteID == siteID })
	data.PlantBufferFlows = filter(data.PlantBufferFlows, func(item PlantBufferFlow) bool { return item.SiteID == siteID })
	data.StockYards = filter(data.StockYards, func(item StockYard) bool { return item.SiteID == siteID })
	data.StockYardPiles = filter(data.StockYardPiles, func(item StockYardPile) bool { return item.SiteID == siteID })
	data.StockYardFlows = filter(data.StockYardFlows, func(item StockYardFlow) bool { return item.SiteID == siteID })
	data.Warehouses = filter(data.Warehouses, func(item Warehouse) bool { return item.SiteID == siteID })
	data.Inventory = filter(data.Inventory, func(item InventoryItem) bool { return item.SiteID == siteID })
	data.InventoryFlows = filter(data.InventoryFlows, func(item InventoryFlow) bool { return item.SiteID == siteID })
	data.InventoryBatchTraces = filter(data.InventoryBatchTraces, func(item InventoryBatchTrace) bool { return item.SiteID == siteID })
	data.InventoryTransfers = filter(data.InventoryTransfers, func(item InventoryTransfer) bool {
		return item.FromSiteID == siteID || item.ToSiteID == siteID
	})
	data.InventoryStocktakes = filter(data.InventoryStocktakes, func(item InventoryStocktake) bool { return item.SiteID == siteID })
	data.Orders = filter(data.Orders, func(item SalesOrder) bool { return item.SiteID == siteID })
	data.ProductionPlans = filter(data.ProductionPlans, func(item ProductionPlan) bool { return item.SiteID == siteID })
	data.ProductionTasks = filter(data.ProductionTasks, func(item ProductionTask) bool { return item.SiteID == siteID })
	data.ProductionBatches = filter(data.ProductionBatches, func(item ProductionBatch) bool { return item.SiteID == siteID })
	data.ProductionReports = filter(data.ProductionReports, func(item ProductionDailyReport) bool { return item.SiteID == siteID })
	data.MixDesigns = filter(data.MixDesigns, func(item MixDesign) bool { return item.SiteID == siteID || item.SiteID == 0 })
	data.MixDesignPlantProfiles = filter(data.MixDesignPlantProfiles, func(item MixDesignPlantProfile) bool { return item.SiteID == siteID })
	data.QualityInspections = filter(data.QualityInspections, func(item QualityInspection) bool { return item.SiteID == siteID })
	visibleInspections := qualityInspectionIDs(data.QualityInspections)
	data.QualitySamples = filter(data.QualitySamples, func(item QualitySample) bool { return visibleInspections[item.InspectionID] })
	data.RawMaterialInspections = filter(data.RawMaterialInspections, func(item RawMaterialInspection) bool { return item.SiteID == siteID })
	data.MixDesignTrialRuns = filter(data.MixDesignTrialRuns, func(item MixDesignTrialRun) bool { return item.SiteID == siteID })
	data.LaboratorySamples = filter(data.LaboratorySamples, func(item LaboratorySample) bool { return item.SiteID == siteID })
	data.LaboratoryTests = filter(data.LaboratoryTests, func(item LaboratoryTestRecord) bool { return item.SiteID == siteID })
	data.LaboratoryEquipment = filter(data.LaboratoryEquipment, func(item LaboratoryEquipment) bool { return item.SiteID == siteID })
	data.LaboratoryCalibrations = filter(data.LaboratoryCalibrations, func(item LaboratoryCalibration) bool { return item.SiteID == siteID })
	data.QualityExceptions = filter(data.QualityExceptions, func(item QualityException) bool { return item.SiteID == siteID })
	data.DispatchOrders = filter(data.DispatchOrders, func(item DispatchOrder) bool { return item.SiteID == siteID })
	siteDispatchIDs := dispatchOrderIDs(data.DispatchOrders)
	data.DispatchSchedules = filter(data.DispatchSchedules, func(item DispatchSchedule) bool { return item.SiteID == siteID })
	data.TransportSettlementItems = filter(data.TransportSettlementItems, func(item TransportSettlementItem) bool { return siteDispatchIDs[item.DispatchID] })
	settlementIDs := transportSettlementIDs(data.TransportSettlementItems)
	data.TransportSettlements = filter(data.TransportSettlements, func(item TransportSettlement) bool { return settlementIDs[item.ID] })
	data.Vehicles = filter(data.Vehicles, func(item Vehicle) bool { return item.SiteID == siteID })
	data.Drivers = filter(data.Drivers, func(item Driver) bool {
		for _, vehicle := range data.Vehicles {
			if vehicle.DriverID == item.ID {
				return true
			}
		}
		return false
	})
	data.VehicleDevices = filter(data.VehicleDevices, func(item VehicleDevice) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.SiteID == siteID
	})
	data.ScaleDevices = filter(data.ScaleDevices, func(item ScaleDevice) bool { return item.SiteID == siteID })
	data.ScaleDeviceEvents = filter(data.ScaleDeviceEvents, func(item ScaleDeviceEvent) bool {
		device, ok := findScaleDeviceByCode(data, item.DeviceCode)
		return ok && device.SiteID == siteID
	})
	data.ScaleTickets = filter(data.ScaleTickets, func(item ScaleTicket) bool {
		if item.SiteID == siteID {
			return true
		}
		dispatch, ok := findDispatch(data, item.DispatchID)
		if ok && dispatch.SiteID == siteID {
			return true
		}
		return item.TicketType == "raw_material_in" && rawTicketSiteID(data, item) == siteID
	})
	data.RawMaterialReceipts = filter(data.RawMaterialReceipts, func(item RawMaterialReceipt) bool { return item.SiteID == siteID })
	data.LatestLocations = filter(data.LatestLocations, func(item VehicleLatestLocation) bool { return item.CurrentSiteID == siteID })
	data.Locations = filter(data.Locations, func(item VehicleLocationEvent) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.SiteID == siteID
	})
	data.DeliverySigns = filter(data.DeliverySigns, func(item DeliverySign) bool {
		dispatch, ok := findDispatch(data, item.DispatchID)
		return ok && dispatch.SiteID == siteID
	})
	data.DeliverySignLinks = filter(data.DeliverySignLinks, func(item DeliverySignLink) bool { return siteDispatchIDs[item.DispatchID] })
	siteSignIDs := deliverySignIDs(data.DeliverySigns)
	data.DeliverySignAttachments = filter(data.DeliverySignAttachments, func(item DeliverySignAttachment) bool {
		return siteSignIDs[item.SignID] || siteDispatchIDs[item.DispatchID]
	})
	data.VehicleAlarms = filter(data.VehicleAlarms, func(item VehicleAlarm) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.SiteID == siteID
	})
	ids := collectScopeIDs(data)
	ids.sites[siteID] = true
	data = filterRelatedMasters(data, ids)
	data.GeoFences = filter(data.GeoFences, func(item GeoFence) bool {
		return item.SiteID == siteID || ids.projects[item.ProjectID]
	})
	data.GeoFenceEvents = filter(data.GeoFenceEvents, func(item GeoFenceEvent) bool {
		vehicle, ok := findVehicle(data, item.VehicleID)
		return ok && vehicle.SiteID == siteID
	})
	data = filterTicketArtifacts(data)
	data.ApprovalTasks = filterApprovalTasksForScopedData(data)
	return data
}

func scopeCompany(data AppData, companyID int64) AppData {
	companyIDs := descendantCompanyIDs(data.Companies, companyID)
	siteIDs := siteIDsForCompanies(data.Sites, companyIDs)
	customerIDs := customerIDsForCompanies(data.Customers, companyIDs)
	projectIDs := projectIDsForCustomers(data.Projects, customerIDs)

	data.Companies = filter(data.Companies, func(item Company) bool { return companyIDs[item.ID] })
	data.Departments = filter(data.Departments, func(item Department) bool { return companyIDs[item.CompanyID] })
	data.Sites = filter(data.Sites, func(item Site) bool { return siteIDs[item.ID] })
	data.Plants = filter(data.Plants, func(item Plant) bool { return siteIDs[item.SiteID] })
	data.PlantBufferLocations = filter(data.PlantBufferLocations, func(item PlantBufferLocation) bool { return siteIDs[item.SiteID] })
	data.PlantBufferFlows = filter(data.PlantBufferFlows, func(item PlantBufferFlow) bool { return siteIDs[item.SiteID] })
	data.StockYards = filter(data.StockYards, func(item StockYard) bool { return siteIDs[item.SiteID] })
	data.StockYardPiles = filter(data.StockYardPiles, func(item StockYardPile) bool { return siteIDs[item.SiteID] })
	data.StockYardFlows = filter(data.StockYardFlows, func(item StockYardFlow) bool { return siteIDs[item.SiteID] })
	data.Warehouses = filter(data.Warehouses, func(item Warehouse) bool { return siteIDs[item.SiteID] })
	warehouseIDs := map[int64]bool{}
	for _, warehouse := range data.Warehouses {
		warehouseIDs[warehouse.ID] = true
	}
	data.Silos = filter(data.Silos, func(item Silo) bool { return warehouseIDs[item.WarehouseID] })
	data.Customers = filter(data.Customers, func(item Customer) bool { return customerIDs[item.ID] })
	data.CustomerContacts = filter(data.CustomerContacts, func(item CustomerContact) bool { return customerIDs[item.CustomerID] })
	data.CustomerBlacklists = filter(data.CustomerBlacklists, func(item CustomerBlacklist) bool { return customerIDs[item.CustomerID] })
	data.CustomerProfiles = filter(data.CustomerProfiles, func(item CustomerProfile) bool { return customerIDs[item.CustomerID] })
	data.CustomerComplaints = filter(data.CustomerComplaints, func(item CustomerComplaint) bool { return customerIDs[item.CustomerID] })
	data.Projects = filter(data.Projects, func(item Project) bool { return projectIDs[item.ID] || customerIDs[item.CustomerID] })
	data.Contracts = filter(data.Contracts, func(item Contract) bool { return customerIDs[item.CustomerID] })
	data.ContractAttachments = filter(data.ContractAttachments, func(item ContractAttachment) bool { return customerIDs[item.CustomerID] })
	data.PricePolicies = filter(data.PricePolicies, func(item PricePolicy) bool {
		return customerIDs[item.CustomerID] || projectIDs[item.ProjectID] || (item.CustomerID == 0 && item.ProjectID == 0)
	})
	data.Orders = filter(data.Orders, func(item SalesOrder) bool {
		return siteIDs[item.SiteID] || customerIDs[item.CustomerID] || projectIDs[item.ProjectID]
	})
	visibleOrders := salesOrderIDs(data.Orders)
	data.ProductionPlans = filter(data.ProductionPlans, func(item ProductionPlan) bool {
		return siteIDs[item.SiteID] || visibleOrders[item.OrderID]
	})
	visiblePlans := productionPlanIDs(data.ProductionPlans)
	data.ProductionTasks = filter(data.ProductionTasks, func(item ProductionTask) bool {
		return siteIDs[item.SiteID] || visiblePlans[item.PlanID]
	})
	data.ProductionBatches = filter(data.ProductionBatches, func(item ProductionBatch) bool {
		return siteIDs[item.SiteID] || visiblePlans[item.PlanID]
	})
	data.ProductionReports = filter(data.ProductionReports, func(item ProductionDailyReport) bool { return siteIDs[item.SiteID] })
	data.PlantBufferLocations = filter(data.PlantBufferLocations, func(item PlantBufferLocation) bool { return siteIDs[item.SiteID] })
	data.PlantBufferFlows = filter(data.PlantBufferFlows, func(item PlantBufferFlow) bool { return siteIDs[item.SiteID] })
	data.StockYards = filter(data.StockYards, func(item StockYard) bool { return siteIDs[item.SiteID] })
	data.StockYardPiles = filter(data.StockYardPiles, func(item StockYardPile) bool { return siteIDs[item.SiteID] })
	data.StockYardFlows = filter(data.StockYardFlows, func(item StockYardFlow) bool { return siteIDs[item.SiteID] })
	data.MixDesigns = filter(data.MixDesigns, func(item MixDesign) bool { return item.SiteID == 0 || siteIDs[item.SiteID] })
	data.MixDesignPlantProfiles = filter(data.MixDesignPlantProfiles, func(item MixDesignPlantProfile) bool { return siteIDs[item.SiteID] })
	data.QualityInspections = filter(data.QualityInspections, func(item QualityInspection) bool { return siteIDs[item.SiteID] })
	visibleInspections := qualityInspectionIDs(data.QualityInspections)
	data.QualitySamples = filter(data.QualitySamples, func(item QualitySample) bool { return visibleInspections[item.InspectionID] })
	data.RawMaterialInspections = filter(data.RawMaterialInspections, func(item RawMaterialInspection) bool { return siteIDs[item.SiteID] })
	data.MixDesignTrialRuns = filter(data.MixDesignTrialRuns, func(item MixDesignTrialRun) bool { return siteIDs[item.SiteID] })
	data.LaboratorySamples = filter(data.LaboratorySamples, func(item LaboratorySample) bool { return siteIDs[item.SiteID] })
	data.LaboratoryTests = filter(data.LaboratoryTests, func(item LaboratoryTestRecord) bool { return siteIDs[item.SiteID] })
	data.LaboratoryEquipment = filter(data.LaboratoryEquipment, func(item LaboratoryEquipment) bool { return siteIDs[item.SiteID] })
	data.LaboratoryCalibrations = filter(data.LaboratoryCalibrations, func(item LaboratoryCalibration) bool { return siteIDs[item.SiteID] })
	data.QualityExceptions = filter(data.QualityExceptions, func(item QualityException) bool { return siteIDs[item.SiteID] })
	data.Inventory = filter(data.Inventory, func(item InventoryItem) bool { return siteIDs[item.SiteID] })
	data.InventoryFlows = filter(data.InventoryFlows, func(item InventoryFlow) bool { return siteIDs[item.SiteID] })
	data.InventoryBatchTraces = filter(data.InventoryBatchTraces, func(item InventoryBatchTrace) bool { return siteIDs[item.SiteID] })
	data.InventoryTransfers = filter(data.InventoryTransfers, func(item InventoryTransfer) bool {
		return siteIDs[item.FromSiteID] || siteIDs[item.ToSiteID]
	})
	data.InventoryStocktakes = filter(data.InventoryStocktakes, func(item InventoryStocktake) bool { return siteIDs[item.SiteID] })
	data.PurchaseRequests = filter(data.PurchaseRequests, func(item PurchaseRequest) bool { return siteIDs[item.SiteID] })
	visibleRequests := purchaseRequestIDs(data.PurchaseRequests)
	data.PurchaseOrders = filter(data.PurchaseOrders, func(item PurchaseOrder) bool { return visibleRequests[item.RequestID] })
	data.RawMaterialReceipts = filter(data.RawMaterialReceipts, func(item RawMaterialReceipt) bool { return siteIDs[item.SiteID] })
	data.DispatchOrders = filter(data.DispatchOrders, func(item DispatchOrder) bool {
		return siteIDs[item.SiteID] || visibleOrders[item.OrderID]
	})
	visibleDispatches := dispatchOrderIDs(data.DispatchOrders)
	data.DispatchSchedules = filter(data.DispatchSchedules, func(item DispatchSchedule) bool { return siteIDs[item.SiteID] })
	data.TransportSettlementItems = filter(data.TransportSettlementItems, func(item TransportSettlementItem) bool { return visibleDispatches[item.DispatchID] })
	visibleSettlements := transportSettlementIDs(data.TransportSettlementItems)
	data.TransportSettlements = filter(data.TransportSettlements, func(item TransportSettlement) bool { return visibleSettlements[item.ID] })
	data.Vehicles = filter(data.Vehicles, func(item Vehicle) bool { return siteIDs[item.SiteID] })
	visibleVehicleIDs := map[int64]bool{}
	visibleDriverIDs := map[int64]bool{}
	for _, vehicle := range data.Vehicles {
		visibleVehicleIDs[vehicle.ID] = true
		visibleDriverIDs[vehicle.DriverID] = true
	}
	for _, dispatch := range data.DispatchOrders {
		visibleVehicleIDs[dispatch.VehicleID] = true
		visibleDriverIDs[dispatch.DriverID] = true
	}
	data.Drivers = filter(data.Drivers, func(item Driver) bool { return visibleDriverIDs[item.ID] })
	data.VehicleDevices = filter(data.VehicleDevices, func(item VehicleDevice) bool { return visibleVehicleIDs[item.VehicleID] })
	data.ScaleDevices = filter(data.ScaleDevices, func(item ScaleDevice) bool { return siteIDs[item.SiteID] })
	data.ScaleDeviceEvents = filter(data.ScaleDeviceEvents, func(item ScaleDeviceEvent) bool {
		device, ok := findScaleDeviceByCode(data, item.DeviceCode)
		return ok && siteIDs[device.SiteID]
	})
	data.ScaleTickets = filter(data.ScaleTickets, func(item ScaleTicket) bool {
		if siteIDs[item.SiteID] || visibleDispatches[item.DispatchID] {
			return true
		}
		return item.TicketType == "raw_material_in" && siteIDs[rawTicketSiteID(data, item)]
	})
	data.LatestLocations = filter(data.LatestLocations, func(item VehicleLatestLocation) bool {
		return siteIDs[item.CurrentSiteID] || customerIDs[item.CurrentCustomerID] || visibleVehicleIDs[item.VehicleID]
	})
	data.Locations = filter(data.Locations, func(item VehicleLocationEvent) bool { return visibleVehicleIDs[item.VehicleID] })
	data.DeliverySigns = filter(data.DeliverySigns, func(item DeliverySign) bool {
		return customerIDs[item.CustomerID] || projectIDs[item.ProjectID] || visibleDispatches[item.DispatchID]
	})
	data.DeliverySignLinks = filter(data.DeliverySignLinks, func(item DeliverySignLink) bool {
		return visibleDispatches[item.DispatchID] || visibleOrders[item.OrderID]
	})
	visibleSigns := deliverySignIDs(data.DeliverySigns)
	data.DeliverySignAttachments = filter(data.DeliverySignAttachments, func(item DeliverySignAttachment) bool {
		return visibleSigns[item.SignID] || visibleDispatches[item.DispatchID]
	})
	data.Statements = filter(data.Statements, func(item Statement) bool { return customerIDs[item.CustomerID] })
	data.SalesInvoices = filter(data.SalesInvoices, func(item SalesInvoice) bool { return customerIDs[item.CustomerID] })
	visibleInvoices := invoiceIDs(data.SalesInvoices)
	data.RedLetterInfos = filter(data.RedLetterInfos, func(item RedLetterInfo) bool {
		return customerIDs[item.CustomerID] || visibleInvoices[item.OriginalInvoiceID] || visibleInvoices[item.RedInvoiceID]
	})
	data.TaxGatewaySubmissions = filter(data.TaxGatewaySubmissions, func(item TaxGatewaySubmission) bool { return visibleInvoices[item.InvoiceID] })
	data.Receivables = filter(data.Receivables, func(item Receivable) bool { return customerIDs[item.CustomerID] })
	data.Receipts = filter(data.Receipts, func(item Receipt) bool { return customerIDs[item.CustomerID] })
	data.PaymentPlans = filter(data.PaymentPlans, func(item PaymentPlan) bool { return customerIDs[item.CustomerID] })
	data.CollectionTasks = filter(data.CollectionTasks, func(item CollectionTask) bool { return customerIDs[item.CustomerID] })
	data.CollectionDispatches = filter(data.CollectionDispatches, func(item CollectionDispatch) bool { return customerIDs[item.CustomerID] })
	data.VehicleAlarms = filter(data.VehicleAlarms, func(item VehicleAlarm) bool {
		return visibleVehicleIDs[item.VehicleID] || visibleDispatches[item.DispatchID]
	})
	data.Users = filter(data.Users, func(item User) bool { return companyIDs[item.CompanyID] })
	data.OIDCProviders = filter(data.OIDCProviders, func(item OIDCProvider) bool { return item.CompanyID == 0 || companyIDs[item.CompanyID] })
	data.SCIMProviders = filter(data.SCIMProviders, func(item SCIMProvider) bool { return item.CompanyID == 0 || companyIDs[item.CompanyID] })

	ids := collectScopeIDs(data)
	for id := range siteIDs {
		ids.sites[id] = true
	}
	for id := range customerIDs {
		ids.customers[id] = true
	}
	for id := range projectIDs {
		ids.projects[id] = true
	}
	data = filterRelatedMasters(data, ids)
	data.GeoFences = filter(data.GeoFences, func(item GeoFence) bool {
		return siteIDs[item.SiteID] || projectIDs[item.ProjectID]
	})
	visibleGeoFences := geoFenceIDs(data.GeoFences)
	data.GeoFenceEvents = filter(data.GeoFenceEvents, func(item GeoFenceEvent) bool {
		return visibleGeoFences[item.FenceID] || visibleVehicleIDs[item.VehicleID]
	})
	data = filterTicketArtifacts(data)
	data.ApprovalTasks = filterApprovalTasksForScopedData(data)
	return data
}

func descendantCompanyIDs(companies []Company, companyID int64) map[int64]bool {
	ids := map[int64]bool{}
	if companyID == 0 {
		return ids
	}
	ids[companyID] = true
	changed := true
	for changed {
		changed = false
		for _, company := range companies {
			if ids[company.ID] || !ids[company.ParentID] {
				continue
			}
			ids[company.ID] = true
			changed = true
		}
	}
	return ids
}

func siteIDsForCompanies(sites []Site, companyIDs map[int64]bool) map[int64]bool {
	ids := map[int64]bool{}
	for _, site := range sites {
		if companyIDs[site.CompanyID] {
			ids[site.ID] = true
		}
	}
	return ids
}

func customerIDsForCompanies(customers []Customer, companyIDs map[int64]bool) map[int64]bool {
	ids := map[int64]bool{}
	for _, customer := range customers {
		if companyIDs[customer.CompanyID] {
			ids[customer.ID] = true
		}
	}
	return ids
}

func projectIDsForCustomers(projects []Project, customerIDs map[int64]bool) map[int64]bool {
	ids := map[int64]bool{}
	for _, project := range projects {
		if customerIDs[project.CustomerID] {
			ids[project.ID] = true
		}
	}
	return ids
}

type scopeIDSet struct {
	customers map[int64]bool
	projects  map[int64]bool
	products  map[int64]bool
	sites     map[int64]bool
	vehicles  map[int64]bool
	drivers   map[int64]bool
}

func newScopeIDSet() scopeIDSet {
	return scopeIDSet{
		customers: map[int64]bool{},
		projects:  map[int64]bool{},
		products:  map[int64]bool{},
		sites:     map[int64]bool{},
		vehicles:  map[int64]bool{},
		drivers:   map[int64]bool{},
	}
}

func collectScopeIDs(data AppData) scopeIDSet {
	ids := newScopeIDSet()
	for _, order := range data.Orders {
		ids.customers[order.CustomerID] = true
		ids.projects[order.ProjectID] = true
		ids.products[order.ProductID] = true
		ids.sites[order.SiteID] = true
	}
	for _, plan := range data.ProductionPlans {
		ids.products[plan.ProductID] = true
		ids.sites[plan.SiteID] = true
	}
	for _, task := range data.ProductionTasks {
		ids.products[task.ProductID] = true
		ids.sites[task.SiteID] = true
	}
	for _, item := range data.MixDesigns {
		ids.products[item.ProductID] = true
		ids.sites[item.SiteID] = true
	}
	for _, item := range data.MixDesignPlantProfiles {
		ids.products[item.ProductID] = true
		ids.sites[item.SiteID] = true
	}
	for _, item := range data.MixDesignTrialRuns {
		ids.products[item.ProductID] = true
		ids.sites[item.SiteID] = true
	}
	for _, item := range data.LaboratorySamples {
		ids.products[item.ProductID] = true
		ids.sites[item.SiteID] = true
	}
	for _, dispatch := range data.DispatchOrders {
		ids.sites[dispatch.SiteID] = true
		ids.vehicles[dispatch.VehicleID] = true
		ids.drivers[dispatch.DriverID] = true
	}
	for _, vehicle := range data.Vehicles {
		ids.sites[vehicle.SiteID] = true
		ids.vehicles[vehicle.ID] = true
		ids.drivers[vehicle.DriverID] = true
	}
	for _, sign := range data.DeliverySigns {
		ids.customers[sign.CustomerID] = true
		ids.projects[sign.ProjectID] = true
	}
	for _, latest := range data.LatestLocations {
		ids.customers[latest.CurrentCustomerID] = true
		ids.projects[latest.CurrentProjectID] = true
		ids.sites[latest.CurrentSiteID] = true
		ids.vehicles[latest.VehicleID] = true
	}
	return ids
}

func filterRelatedMasters(data AppData, ids scopeIDSet) AppData {
	data.Sites = filter(data.Sites, func(item Site) bool { return ids.sites[item.ID] })
	data.Plants = filter(data.Plants, func(item Plant) bool { return ids.sites[item.SiteID] })
	data.PlantBufferLocations = filter(data.PlantBufferLocations, func(item PlantBufferLocation) bool { return ids.sites[item.SiteID] })
	data.PlantBufferFlows = filter(data.PlantBufferFlows, func(item PlantBufferFlow) bool { return ids.sites[item.SiteID] })
	data.StockYards = filter(data.StockYards, func(item StockYard) bool { return ids.sites[item.SiteID] })
	data.StockYardPiles = filter(data.StockYardPiles, func(item StockYardPile) bool { return ids.sites[item.SiteID] })
	data.StockYardFlows = filter(data.StockYardFlows, func(item StockYardFlow) bool { return ids.sites[item.SiteID] })
	data.Warehouses = filter(data.Warehouses, func(item Warehouse) bool { return ids.sites[item.SiteID] })
	warehouseIDs := map[int64]bool{}
	for _, warehouse := range data.Warehouses {
		warehouseIDs[warehouse.ID] = true
	}
	data.Silos = filter(data.Silos, func(item Silo) bool { return warehouseIDs[item.WarehouseID] })
	data.Inventory = filter(data.Inventory, func(item InventoryItem) bool { return ids.sites[item.SiteID] })
	data.InventoryBatchTraces = filter(data.InventoryBatchTraces, func(item InventoryBatchTrace) bool { return ids.sites[item.SiteID] })
	data.ScaleDevices = filter(data.ScaleDevices, func(item ScaleDevice) bool { return ids.sites[item.SiteID] })
	data.ScaleDeviceEvents = filter(data.ScaleDeviceEvents, func(item ScaleDeviceEvent) bool {
		device, ok := findScaleDeviceByCode(data, item.DeviceCode)
		return ok && ids.sites[device.SiteID]
	})
	data.Customers = filter(data.Customers, func(item Customer) bool { return ids.customers[item.ID] })
	data.CustomerContacts = filter(data.CustomerContacts, func(item CustomerContact) bool { return ids.customers[item.CustomerID] })
	data.CustomerBlacklists = filter(data.CustomerBlacklists, func(item CustomerBlacklist) bool { return ids.customers[item.CustomerID] })
	data.CustomerProfiles = filter(data.CustomerProfiles, func(item CustomerProfile) bool { return ids.customers[item.CustomerID] })
	data.CustomerComplaints = filter(data.CustomerComplaints, func(item CustomerComplaint) bool { return ids.customers[item.CustomerID] })
	data.Projects = filter(data.Projects, func(item Project) bool {
		return ids.projects[item.ID] || ids.customers[item.CustomerID]
	})
	data.Contracts = filter(data.Contracts, func(item Contract) bool { return ids.customers[item.CustomerID] })
	data.ContractAttachments = filter(data.ContractAttachments, func(item ContractAttachment) bool { return ids.customers[item.CustomerID] })
	data.PricePolicies = filter(data.PricePolicies, func(item PricePolicy) bool {
		return ids.customers[item.CustomerID] || ids.projects[item.ProjectID] || (item.CustomerID == 0 && item.ProjectID == 0)
	})
	data.Products = filter(data.Products, func(item Product) bool {
		return ids.products[item.ID] || len(ids.products) == 0
	})
	data.MixDesigns = filter(data.MixDesigns, func(item MixDesign) bool {
		return ids.products[item.ProductID] || len(ids.products) == 0
	})
	data.MixDesignPlantProfiles = filter(data.MixDesignPlantProfiles, func(item MixDesignPlantProfile) bool {
		return ids.products[item.ProductID] || len(ids.products) == 0
	})
	data.MixDesignTrialRuns = filter(data.MixDesignTrialRuns, func(item MixDesignTrialRun) bool {
		return ids.products[item.ProductID] || len(ids.products) == 0
	})
	data.Vehicles = filter(data.Vehicles, func(item Vehicle) bool {
		return ids.vehicles[item.ID] || ids.sites[item.SiteID]
	})
	data.Drivers = filter(data.Drivers, func(item Driver) bool { return ids.drivers[item.ID] })
	data.VehicleDevices = filter(data.VehicleDevices, func(item VehicleDevice) bool { return ids.vehicles[item.VehicleID] })
	data = filterRelatedOrganizations(data)
	return data
}

func filterRelatedOrganizations(data AppData) AppData {
	companyIDs := map[int64]bool{}
	for _, site := range data.Sites {
		companyIDs[site.CompanyID] = true
	}
	for _, customer := range data.Customers {
		companyIDs[customer.CompanyID] = true
	}
	data.Companies = filter(data.Companies, func(item Company) bool {
		if companyIDs[item.ID] {
			return true
		}
		return false
	})
	data.Tenants = nil
	data.TenantPolicies = nil
	data.Departments = filter(data.Departments, func(item Department) bool { return companyIDs[item.CompanyID] })
	return data
}

func orderIDsFromDispatches(items []DispatchOrder) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.OrderID] = true
	}
	return ids
}

func salesOrderIDs(items []SalesOrder) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func purchaseRequestIDs(items []PurchaseRequest) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func dispatchOrderIDs(items []DispatchOrder) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func transportSettlementIDs(items []TransportSettlementItem) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.SettlementID] = true
	}
	return ids
}

func geoFenceIDs(items []GeoFence) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func productionPlanIDs(items []ProductionPlan) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func qualityInspectionIDs(items []QualityInspection) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func filterTicketArtifacts(data AppData) AppData {
	visible := visibleTicketIDs(data)
	data.ScaleWeightRecords = filter(data.ScaleWeightRecords, func(item ScaleWeightRecord) bool { return visible[item.TicketID] })
	data.DeliveryNotes = filter(data.DeliveryNotes, func(item DeliveryNote) bool { return item.TicketID == 0 || visible[item.TicketID] })
	data.DeliverySignLinks = filter(data.DeliverySignLinks, func(item DeliverySignLink) bool { return item.TicketID == 0 || visible[item.TicketID] })
	data.DeliverySignAttachments = filter(data.DeliverySignAttachments, func(item DeliverySignAttachment) bool { return item.TicketID == 0 || visible[item.TicketID] })
	data.TicketPrintLogs = filter(data.TicketPrintLogs, func(item TicketPrintLog) bool { return visible[item.TicketID] })
	data.TicketVoidLogs = filter(data.TicketVoidLogs, func(item TicketVoidLog) bool { return visible[item.TicketID] })
	return data
}

func deliverySignIDs(items []DeliverySign) map[int64]bool {
	ids := map[int64]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	return ids
}

func filterApprovalTasksForScopedData(data AppData) []ApprovalTask {
	orderIDs := map[int64]bool{}
	for _, order := range data.Orders {
		orderIDs[order.ID] = true
	}
	transferIDs := map[int64]bool{}
	for _, transfer := range data.InventoryTransfers {
		transferIDs[transfer.ID] = true
	}
	return filter(data.ApprovalTasks, func(item ApprovalTask) bool {
		if item.Resource == "sales_order" {
			return orderIDs[item.ResourceID]
		}
		if item.Resource == "inventory_transfer" {
			return transferIDs[item.ResourceID]
		}
		return true
	})
}

func applyFieldPolicies(data AppData, user User) AppData {
	for _, policy := range data.FieldPolicies {
		if !policy.Enabled || policy.RoleCode != user.RoleCode {
			continue
		}
		data = applyFieldPolicy(data, policy)
	}
	return data
}

func applyFieldPolicy(data AppData, policy FieldPolicy) AppData {
	switch policy.Resource {
	case "customers":
		for i := range data.Customers {
			if policy.Field == "phone" || policy.Field == "*" {
				data.Customers[i].Phone = maskField(data.Customers[i].Phone, policy.Mask)
			}
		}
	case "customerContacts":
		for i := range data.CustomerContacts {
			if policy.Field == "phone" || policy.Field == "*" {
				data.CustomerContacts[i].Phone = maskField(data.CustomerContacts[i].Phone, policy.Mask)
			}
		}
	case "projects":
		for i := range data.Projects {
			if policy.Field == "phone" || policy.Field == "*" {
				data.Projects[i].Phone = maskField(data.Projects[i].Phone, policy.Mask)
			}
		}
	case "orders":
		for i := range data.Orders {
			if policy.Field == "phone" || policy.Field == "*" {
				data.Orders[i].Phone = maskField(data.Orders[i].Phone, policy.Mask)
			}
		}
	case "deliverySigns":
		for i := range data.DeliverySigns {
			if policy.Field == "phone" || policy.Field == "*" {
				data.DeliverySigns[i].Phone = maskField(data.DeliverySigns[i].Phone, policy.Mask)
			}
		}
	case "drivers":
		for i := range data.Drivers {
			if policy.Field == "phone" || policy.Field == "*" {
				data.Drivers[i].Phone = maskField(data.Drivers[i].Phone, policy.Mask)
			}
			if policy.Field == "licenseNo" || policy.Field == "*" {
				data.Drivers[i].LicenseNo = maskField(data.Drivers[i].LicenseNo, policy.Mask)
			}
		}
	case "*":
		for i := range data.Customers {
			data.Customers[i].Phone = maskField(data.Customers[i].Phone, policy.Mask)
		}
		for i := range data.CustomerContacts {
			data.CustomerContacts[i].Phone = maskField(data.CustomerContacts[i].Phone, policy.Mask)
		}
		for i := range data.Projects {
			data.Projects[i].Phone = maskField(data.Projects[i].Phone, policy.Mask)
		}
		for i := range data.Orders {
			data.Orders[i].Phone = maskField(data.Orders[i].Phone, policy.Mask)
		}
		for i := range data.DeliverySigns {
			data.DeliverySigns[i].Phone = maskField(data.DeliverySigns[i].Phone, policy.Mask)
		}
		for i := range data.Drivers {
			data.Drivers[i].Phone = maskField(data.Drivers[i].Phone, policy.Mask)
			data.Drivers[i].LicenseNo = maskField(data.Drivers[i].LicenseNo, policy.Mask)
		}
	}
	return data
}

func maskField(value, mode string) string {
	switch mode {
	case "code":
		return maskCode(value)
	case "redact":
		if value == "" {
			return ""
		}
		return "***"
	default:
		return maskPhone(value)
	}
}

func maskPhone(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	if len(value) <= 7 {
		return value[:1] + "****" + value[len(value)-1:]
	}
	return value[:3] + "****" + value[len(value)-4:]
}

func maskCode(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

func visibleTicketPrintLogs(data AppData) []TicketPrintLog {
	visible := visibleTicketIDs(data)
	return filter(data.TicketPrintLogs, func(item TicketPrintLog) bool { return visible[item.TicketID] })
}

func visibleTicketVoidLogs(data AppData) []TicketVoidLog {
	visible := visibleTicketIDs(data)
	return filter(data.TicketVoidLogs, func(item TicketVoidLog) bool { return visible[item.TicketID] })
}

func visibleTicketIDs(data AppData) map[int64]bool {
	visible := map[int64]bool{}
	for _, ticket := range data.ScaleTickets {
		visible[ticket.ID] = true
	}
	return visible
}

func userCanAccessTicket(data AppData, user User, ticket ScaleTicket) bool {
	switch user.RoleCode {
	case "customer":
		if user.CustomerID == 0 {
			return false
		}
		order, ok := findOrder(data, ticket.OrderID)
		return ok && order.CustomerID == user.CustomerID
	case "driver":
		if user.DriverID == 0 {
			return false
		}
		dispatch, ok := findDispatch(data, ticket.DispatchID)
		return ok && dispatch.DriverID == user.DriverID
	case "dispatcher":
		if user.SiteID == 0 {
			return false
		}
		if ticket.SiteID == user.SiteID {
			return true
		}
		dispatch, ok := findDispatch(data, ticket.DispatchID)
		if ok && dispatch.SiteID == user.SiteID {
			return true
		}
		return ticket.TicketType == "raw_material_in" && rawTicketSiteID(data, ticket) == user.SiteID
	default:
		return true
	}
}

func rawTicketSiteID(data AppData, ticket ScaleTicket) int64 {
	if ticket.ReceiptID == 0 {
		return 0
	}
	receipt, ok := findRawMaterialReceipt(data, ticket.ReceiptID)
	if !ok {
		return 0
	}
	return receipt.SiteID
}

func userCanAccessInvoice(data AppData, user User, invoice SalesInvoice) bool {
	switch user.RoleCode {
	case "customer":
		return user.CustomerID > 0 && invoice.CustomerID == user.CustomerID
	case "driver", "device":
		return false
	default:
		return true
	}
}

func invoiceIDs(items []SalesInvoice) map[int64]bool {
	out := map[int64]bool{}
	for _, item := range items {
		out[item.ID] = true
	}
	return out
}

func filter[T any](items []T, keep func(T) bool) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			out = append(out, item)
		}
	}
	return out
}
