package appliance

func ensureEnterpriseDefaults(data *AppData) bool {
	seed := SeedData()
	changed := false
	if data.SchemaVersion < 2 {
		data.SchemaVersion = 2
		changed = true
	}
	if data.Next == nil {
		data.Next = map[string]int64{}
		changed = true
	}
	for key, value := range seed.Next {
		if data.Next[key] < value {
			data.Next[key] = value
			changed = true
		}
	}
	if normalizeStandaloneOperationPlatform(data) {
		changed = true
	}
	if len(data.Companies) == 0 {
		data.Companies = seed.Companies
		changed = true
	}
	for i := range data.Companies {
		if data.Companies[i].Status == "" {
			data.Companies[i].Status = "active"
			changed = true
		}
	}
	if len(data.Departments) == 0 {
		data.Departments = seed.Departments
		changed = true
	}
	if len(data.Plants) == 0 {
		data.Plants = seed.Plants
		changed = true
	}
	if len(data.Warehouses) == 0 {
		data.Warehouses = seed.Warehouses
		changed = true
	}
	if len(data.Silos) == 0 {
		data.Silos = seed.Silos
		changed = true
	}
	for _, silo := range seed.Silos {
		if !hasSiloDefault(*data, silo.WarehouseID, silo.MaterialID) {
			silo.ID = nextID(data, "silo")
			data.Silos = append(data.Silos, silo)
			changed = true
		}
	}
	if len(data.CustomerContacts) == 0 {
		data.CustomerContacts = seed.CustomerContacts
		changed = true
	}
	if len(data.CustomerProfiles) == 0 {
		data.CustomerProfiles = seed.CustomerProfiles
		changed = true
	}
	if len(data.CustomerComplaints) == 0 {
		data.CustomerComplaints = seed.CustomerComplaints
		changed = true
	}
	if len(data.Contracts) == 0 {
		data.Contracts = seed.Contracts
		changed = true
	}
	if len(data.ContractAttachments) == 0 {
		data.ContractAttachments = seed.ContractAttachments
		changed = true
	}
	if len(data.TaxRates) == 0 {
		data.TaxRates = seed.TaxRates
		changed = true
	}
	if len(data.PricePolicies) == 0 {
		data.PricePolicies = seed.PricePolicies
		changed = true
	}
	if len(data.Carriers) == 0 {
		data.Carriers = seed.Carriers
		changed = true
	}
	if len(data.VehicleDevices) == 0 {
		data.VehicleDevices = seed.VehicleDevices
		changed = true
	}
	for _, device := range seed.VehicleDevices {
		found := false
		for _, existing := range data.VehicleDevices {
			if existing.DeviceNo == device.DeviceNo {
				found = true
				break
			}
		}
		if !found {
			data.VehicleDevices = append(data.VehicleDevices, device)
			changed = true
		}
	}
	if len(data.DeviceCredentials) == 0 {
		data.DeviceCredentials = seed.DeviceCredentials
		changed = true
	}
	if len(data.GatewayRoutes) == 0 {
		data.GatewayRoutes = seed.GatewayRoutes
		changed = true
	}
	for _, credential := range seed.DeviceCredentials {
		found := false
		for _, existing := range data.DeviceCredentials {
			if existing.DeviceNo == credential.DeviceNo {
				found = true
				break
			}
		}
		if !found {
			data.DeviceCredentials = append(data.DeviceCredentials, credential)
			changed = true
		}
	}
	if len(data.SecurityPolicies) == 0 {
		data.SecurityPolicies = seed.SecurityPolicies
		changed = true
	}
	for _, policy := range seed.SecurityPolicies {
		if !hasSecurityPolicyDefault(data.SecurityPolicies, policy.Type, policy.Value) {
			if hasSecurityPolicyID(data.SecurityPolicies, policy.ID) {
				policy.ID = nextID(data, "securityPolicy")
			}
			data.SecurityPolicies = append(data.SecurityPolicies, policy)
			changed = true
		}
	}
	if len(data.OIDCProviders) == 0 {
		data.OIDCProviders = seed.OIDCProviders
		changed = true
	}
	if len(data.SCIMProviders) == 0 {
		data.SCIMProviders = seed.SCIMProviders
		changed = true
	}
	if len(data.ProductInstances) == 0 {
		data.ProductInstances = seed.ProductInstances
		changed = true
	}
	if len(data.SystemAlerts) == 0 {
		data.SystemAlerts = seed.SystemAlerts
		changed = true
	}
	if len(data.ProductRenewalTasks) == 0 {
		data.ProductRenewalTasks = seed.ProductRenewalTasks
		changed = true
	}
	if len(data.ProductRenewalQuotes) == 0 {
		data.ProductRenewalQuotes = seed.ProductRenewalQuotes
		changed = true
	}
	if len(data.ProductRenewalContracts) == 0 {
		data.ProductRenewalContracts = seed.ProductRenewalContracts
		changed = true
	}
	if len(data.ProductRenewalPayments) == 0 {
		data.ProductRenewalPayments = seed.ProductRenewalPayments
		changed = true
	}
	if len(data.ProductRenewalApprovals) == 0 {
		data.ProductRenewalApprovals = seed.ProductRenewalApprovals
		changed = true
	}
	if len(data.ProductRenewalInvoices) == 0 {
		data.ProductRenewalInvoices = seed.ProductRenewalInvoices
		changed = true
	}
	if len(data.ProductRenewalESigns) == 0 {
		data.ProductRenewalESigns = seed.ProductRenewalESigns
		changed = true
	}
	if len(data.ProductRenewalIntegrations) == 0 {
		data.ProductRenewalIntegrations = seed.ProductRenewalIntegrations
		changed = true
	}
	if len(data.ProductRenewalSyncRecords) == 0 {
		data.ProductRenewalSyncRecords = seed.ProductRenewalSyncRecords
		changed = true
	}
	if len(data.ProductProbeReports) == 0 {
		data.ProductProbeReports = seed.ProductProbeReports
		changed = true
	}
	if len(data.ProductTelemetryEvents) == 0 {
		data.ProductTelemetryEvents = seed.ProductTelemetryEvents
		changed = true
	}
	if len(data.ProductMonitoringIntegrations) == 0 {
		data.ProductMonitoringIntegrations = seed.ProductMonitoringIntegrations
		changed = true
	}
	if len(data.ProductAlertRules) == 0 {
		data.ProductAlertRules = seed.ProductAlertRules
		changed = true
	}
	if len(data.ProductMonitoringEvents) == 0 {
		data.ProductMonitoringEvents = seed.ProductMonitoringEvents
		changed = true
	}
	if len(data.ProductAlertPolicies) == 0 {
		data.ProductAlertPolicies = seed.ProductAlertPolicies
		changed = true
	}
	if len(data.ProductAlertChannels) == 0 {
		data.ProductAlertChannels = seed.ProductAlertChannels
		changed = true
	}
	if len(data.ProductAlertNotifications) == 0 {
		data.ProductAlertNotifications = seed.ProductAlertNotifications
		changed = true
	}
	if len(data.ProductUpdateRollouts) == 0 {
		data.ProductUpdateRollouts = seed.ProductUpdateRollouts
		changed = true
	}
	if len(data.ProductUpdateExecutions) == 0 {
		data.ProductUpdateExecutions = seed.ProductUpdateExecutions
		changed = true
	}
	if len(data.ProductSystemUpdateTasks) == 0 {
		data.ProductSystemUpdateTasks = seed.ProductSystemUpdateTasks
		changed = true
	}
	for i := range data.ProductInstances {
		if data.ProductInstances[i].ProbeToken == "" {
			data.ProductInstances[i].ProbeToken = productProbeToken(data.ProductInstances[i])
			data.ProductInstances[i].ProbeEnabled = true
			changed = true
		}
		if data.ProductInstances[i].HealthStatus == "" {
			data.ProductInstances[i].HealthStatus = fallback(data.ProductInstances[i].Status, "unknown")
			changed = true
		}
	}
	if len(data.FieldPolicies) == 0 {
		data.FieldPolicies = seed.FieldPolicies
		changed = true
	}
	for _, policy := range seed.FieldPolicies {
		if !hasFieldPolicyDefault(data.FieldPolicies, policy.RoleCode, policy.Resource, policy.Field) {
			data.FieldPolicies = append(data.FieldPolicies, policy)
			changed = true
		}
	}
	if len(data.PurchaseRequests) == 0 {
		data.PurchaseRequests = seed.PurchaseRequests
		changed = true
	}
	if len(data.PurchaseOrders) == 0 {
		data.PurchaseOrders = seed.PurchaseOrders
		changed = true
	}
	if len(data.RawMaterialReceipts) == 0 {
		data.RawMaterialReceipts = seed.RawMaterialReceipts
		changed = true
	}
	if len(data.InventoryFlows) == 0 {
		data.InventoryFlows = seed.InventoryFlows
		changed = true
	}
	if len(data.DispatchSchedules) == 0 {
		data.DispatchSchedules = seed.DispatchSchedules
		changed = true
	}
	for _, item := range seed.Inventory {
		if !hasInventoryDefault(*data, item.SiteID, item.MaterialID) {
			item.ID = nextID(data, "inventory")
			data.Inventory = append(data.Inventory, item)
			changed = true
		}
	}
	if len(data.ProductionTasks) == 0 {
		data.ProductionTasks = seed.ProductionTasks
		changed = true
	}
	if len(data.ProductionBatches) == 0 {
		data.ProductionBatches = seed.ProductionBatches
		changed = true
	}
	if len(data.ProductionReports) == 0 {
		data.ProductionReports = seed.ProductionReports
		changed = true
	}
	if len(data.ScaleDevices) == 0 {
		data.ScaleDevices = seed.ScaleDevices
		changed = true
	}
	if len(data.ScaleTickets) == 0 {
		data.ScaleTickets = seed.ScaleTickets
		changed = true
	}
	if len(data.ScaleWeightRecords) == 0 {
		data.ScaleWeightRecords = seed.ScaleWeightRecords
		changed = true
	}
	if len(data.DeliveryNotes) == 0 {
		data.DeliveryNotes = seed.DeliveryNotes
		changed = true
	}
	if len(data.DeliverySignLinks) == 0 {
		data.DeliverySignLinks = seed.DeliverySignLinks
		changed = true
	}
	if len(data.TicketPrintLogs) == 0 {
		data.TicketPrintLogs = seed.TicketPrintLogs
		changed = true
	}
	if len(data.TicketVoidLogs) == 0 {
		data.TicketVoidLogs = seed.TicketVoidLogs
		changed = true
	}
	if len(data.DeliverySigns) == 0 {
		data.DeliverySigns = seed.DeliverySigns
		changed = true
	}
	if len(data.DeliverySignAttachments) == 0 {
		data.DeliverySignAttachments = seed.DeliverySignAttachments
		changed = true
	}
	if backfillRawMaterialTickets(data) {
		changed = true
	}
	if len(data.SalesInvoices) == 0 {
		data.SalesInvoices = seed.SalesInvoices
		changed = true
	}
	for i := range data.SalesInvoices {
		if data.SalesInvoices[i].InvoiceType == "" {
			data.SalesInvoices[i].InvoiceType = "blue"
			changed = true
		}
		if data.SalesInvoices[i].InvoiceCategory == "" {
			if data.SalesInvoices[i].InvoiceType == "red" {
				data.SalesInvoices[i].InvoiceCategory = "red_vat_special"
			} else {
				data.SalesInvoices[i].InvoiceCategory = "blue_vat_special"
			}
			changed = true
		}
	}
	if len(data.TaxGatewaySubmissions) == 0 {
		data.TaxGatewaySubmissions = seed.TaxGatewaySubmissions
		changed = true
	}
	for i := range data.TaxGatewaySubmissions {
		if data.TaxGatewaySubmissions[i].Action == "" {
			data.TaxGatewaySubmissions[i].Action = "issue"
			changed = true
		}
	}
	if len(data.Receivables) == 0 {
		data.Receivables = seed.Receivables
		changed = true
	}
	if len(data.Receipts) == 0 {
		data.Receipts = seed.Receipts
		changed = true
	}
	if len(data.PaymentPlans) == 0 {
		data.PaymentPlans = seed.PaymentPlans
		changed = true
	}
	if len(data.CollectionTemplates) == 0 {
		data.CollectionTemplates = seed.CollectionTemplates
		changed = true
	}
	if len(data.SupplierStatements) == 0 {
		data.SupplierStatements = seed.SupplierStatements
		changed = true
	}
	if len(data.Payables) == 0 {
		data.Payables = seed.Payables
		changed = true
	}
	if len(data.TransportSettlements) == 0 {
		data.TransportSettlements = seed.TransportSettlements
		changed = true
	}
	if len(data.TransportSettlementItems) == 0 {
		data.TransportSettlementItems = seed.TransportSettlementItems
		changed = true
	}
	if len(data.CostCalcs) == 0 {
		data.CostCalcs = seed.CostCalcs
		changed = true
	}
	if len(data.ProjectProfits) == 0 {
		data.ProjectProfits = seed.ProjectProfits
		changed = true
	}
	if len(data.RuleDefinitions) == 0 {
		data.RuleDefinitions = seed.RuleDefinitions
		changed = true
	}
	if len(data.Notifications) == 0 {
		data.Notifications = seed.Notifications
		changed = true
	}
	if len(data.IntegrationEndpoints) == 0 {
		data.IntegrationEndpoints = seed.IntegrationEndpoints
		changed = true
	}
	for _, endpoint := range seed.IntegrationEndpoints {
		if endpoint.Type != "collection" {
			continue
		}
		found := false
		for _, existing := range data.IntegrationEndpoints {
			if existing.Type == endpoint.Type && existing.Protocol == endpoint.Protocol {
				found = true
				break
			}
		}
		if !found {
			data.IntegrationEndpoints = append(data.IntegrationEndpoints, endpoint)
			changed = true
		}
	}
	if len(data.ApprovalFlows) == 0 {
		data.ApprovalFlows = seed.ApprovalFlows
		changed = true
	}
	for _, flow := range seed.ApprovalFlows {
		found := false
		for _, existing := range data.ApprovalFlows {
			if existing.Code == flow.Code {
				found = true
				break
			}
		}
		if !found {
			data.ApprovalFlows = append(data.ApprovalFlows, flow)
			changed = true
		}
	}
	if len(data.DataDictionaries) == 0 {
		data.DataDictionaries = seed.DataDictionaries
		changed = true
	}
	for _, item := range seed.DataDictionaries {
		if !hasDataDictionaryDefault(data.DataDictionaries, item.Type, item.Code) {
			data.DataDictionaries = append(data.DataDictionaries, item)
			changed = true
		}
	}
	for i := range data.Plugins {
		if data.Plugins[i].Checksum == "" {
			for _, plugin := range seed.Plugins {
				if plugin.ID == data.Plugins[i].ID {
					data.Plugins[i].Checksum = plugin.Checksum
					data.Plugins[i].Signature = plugin.Signature
					changed = true
				}
			}
		}
		for _, plugin := range seed.Plugins {
			if plugin.ID != data.Plugins[i].ID {
				continue
			}
			if data.Plugins[i].Runtime == "" {
				data.Plugins[i].Runtime = plugin.Runtime
				data.Plugins[i].Entrypoint = plugin.Entrypoint
				data.Plugins[i].Sandbox = plugin.Sandbox
				changed = true
			}
		}
	}
	for i := range data.Updates {
		if data.Updates[i].Component == "" {
			data.Updates[i].Component = "server"
			changed = true
		}
		if data.Updates[i].PackageType == "" {
			data.Updates[i].PackageType = "full"
			changed = true
		}
		if data.Updates[i].Signature == "" {
			for _, update := range seed.Updates {
				if update.ID == data.Updates[i].ID {
					data.Updates[i].Signature = update.Signature
					data.Updates[i].RollbackVersion = update.RollbackVersion
					data.Updates[i].Component = fallback(data.Updates[i].Component, update.Component)
					changed = true
				}
			}
		}
	}
	for i := range data.Roles {
		for _, role := range seed.Roles {
			if role.Code == data.Roles[i].Code && len(data.Roles[i].Permissions) < len(role.Permissions) {
				data.Roles[i].Permissions = role.Permissions
				data.Roles[i].DataScope = role.DataScope
				changed = true
			}
		}
	}
	for _, role := range seed.Roles {
		found := false
		for _, existing := range data.Roles {
			if existing.Code == role.Code {
				found = true
				break
			}
		}
		if !found {
			data.Roles = append(data.Roles, role)
			changed = true
		}
	}
	for _, user := range seed.Users {
		found := false
		for _, existing := range data.Users {
			if existing.Username == user.Username {
				found = true
				break
			}
		}
		if !found {
			data.Users = append(data.Users, user)
			changed = true
		}
	}
	return changed
}

func normalizeStandaloneOperationPlatform(data *AppData) bool {
	changed := false
	if len(data.Tenants) > 0 {
		data.Tenants = nil
		changed = true
	}
	if len(data.TenantPolicies) > 0 {
		data.TenantPolicies = nil
		changed = true
	}
	for i := range data.Roles {
		if data.Roles[i].DataScope == "tenant" {
			data.Roles[i].DataScope = "platform"
			changed = true
		}
	}
	for i := range data.Users {
		if data.Users[i].TenantID != 0 {
			data.Users[i].TenantID = 0
			changed = true
		}
	}
	for i := range data.Companies {
		if data.Companies[i].TenantID != 0 {
			data.Companies[i].TenantID = 0
			changed = true
		}
	}
	for i := range data.OIDCProviders {
		if data.OIDCProviders[i].TenantID != 0 {
			data.OIDCProviders[i].TenantID = 0
			changed = true
		}
	}
	for i := range data.SCIMProviders {
		if data.SCIMProviders[i].TenantID != 0 {
			data.SCIMProviders[i].TenantID = 0
			changed = true
		}
	}
	return changed
}

func hasSiloDefault(data AppData, warehouseID, materialID int64) bool {
	for _, item := range data.Silos {
		if item.WarehouseID == warehouseID && item.MaterialID == materialID {
			return true
		}
	}
	return false
}

func hasInventoryDefault(data AppData, siteID, materialID int64) bool {
	for _, item := range data.Inventory {
		if item.SiteID == siteID && item.MaterialID == materialID {
			return true
		}
	}
	return false
}

func hasFieldPolicyDefault(items []FieldPolicy, roleCode, resource, field string) bool {
	for _, item := range items {
		if item.RoleCode == roleCode && item.Resource == resource && item.Field == field {
			return true
		}
	}
	return false
}

func hasSecurityPolicyDefault(items []SecurityPolicy, policyType, value string) bool {
	for _, item := range items {
		if item.Type == policyType && item.Value == value {
			return true
		}
	}
	return false
}

func hasSecurityPolicyID(items []SecurityPolicy, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasDataDictionaryDefault(items []DataDictionary, typ, code string) bool {
	for _, item := range items {
		if item.Type == typ && item.Code == code {
			return true
		}
	}
	return false
}
