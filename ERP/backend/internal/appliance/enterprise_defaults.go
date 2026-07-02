package appliance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

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
	shouldBackfillIntegrationEndpoints := len(data.IntegrationEndpoints) == 0 && data.Next["integration"] == 0 && data.Next["integrationEndpoint"] == 0
	for key, value := range seed.Next {
		if data.Next[key] < value {
			data.Next[key] = value
			changed = true
		}
	}
	if normalizeStandaloneOperationPlatform(data) {
		changed = true
	}
	if !demoSeedEnabled() {
		if purgeDemoBusinessSeed(data, seed) {
			changed = true
		}
		if ensureRuntimeDefaults(data, seed) {
			changed = true
		}
		return changed
	}
	if data.GroupProfile.Name == "" {
		data.GroupProfile = seed.GroupProfile
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
		for _, company := range seed.Companies {
			if company.ID != data.Companies[i].ID {
				continue
			}
			if data.Companies[i].Level == "" {
				data.Companies[i].Level = company.Level
				changed = true
			}
			if data.Companies[i].Region == "" {
				data.Companies[i].Region = company.Region
				changed = true
			}
			if data.Companies[i].ParentID == 0 && company.ParentID != 0 {
				data.Companies[i].ParentID = company.ParentID
				changed = true
			}
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
	for i := range data.Vehicles {
		if data.Vehicles[i].InternalNo == "" {
			data.Vehicles[i].InternalNo = fmt.Sprintf("V%03d", data.Vehicles[i].ID)
			changed = true
		}
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
		if hasDemoCredential(data.ProductInstances[i].ProbeToken) {
			data.ProductInstances[i].ProbeToken = productProbeToken(data.ProductInstances[i])
			changed = true
		}
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
	if sanitizeDeliveryLicenseDefaults(data) {
		changed = true
	}
	if sanitizeDeliveryExternalDefaults(data) {
		changed = true
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
	if len(data.PlantBufferLocations) == 0 {
		data.PlantBufferLocations = seed.PlantBufferLocations
		changed = true
	}
	if len(data.PlantBufferFlows) == 0 {
		data.PlantBufferFlows = seed.PlantBufferFlows
		changed = true
	}
	if len(data.StockYards) == 0 {
		data.StockYards = seed.StockYards
		changed = true
	}
	if len(data.StockYardPiles) == 0 {
		data.StockYardPiles = seed.StockYardPiles
		changed = true
	}
	if len(data.StockYardFlows) == 0 {
		data.StockYardFlows = seed.StockYardFlows
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
	if shouldBackfillIntegrationEndpoints {
		data.IntegrationEndpoints = seed.IntegrationEndpoints
		changed = true
	}
	if shouldBackfillIntegrationEndpoints {
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
	}
	for _, endpoint := range data.IntegrationEndpoints {
		if data.Next["integration"] < endpoint.ID {
			ensureCounterAtLeast(data, "integration", endpoint.ID)
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
		if data.Next == nil || data.Next["dict"] < item.ID {
			ensureCounterAtLeast(data, "dict", item.ID)
			changed = true
		}
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
		scope := normalizeDataScope(data.Roles[i].DataScope)
		if data.Roles[i].DataScope != scope {
			data.Roles[i].DataScope = scope
			changed = true
		}
	}
	if ensureBuiltinRolePermissions(data, "quality", "approval:read", "approval:write") {
		changed = true
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
	if ensureBuiltinAdminSuperAdmin(data) {
		changed = true
	}
	return changed
}

func ensureRuntimeDefaults(data *AppData, seed AppData) bool {
	changed := false
	if data.GroupProfile.Name == "" {
		data.GroupProfile = unconfiguredGroupProfile()
		changed = true
	}
	if len(data.Modules) == 0 {
		data.Modules = seed.Modules
		changed = true
	}
	if len(data.Plugins) == 0 {
		data.Plugins = seed.Plugins
		changed = true
	}
	if len(data.SecurityPolicies) == 0 {
		data.SecurityPolicies = seed.SecurityPolicies
		changed = true
	}
	if len(data.GatewayRoutes) == 0 {
		data.GatewayRoutes = seed.GatewayRoutes
		changed = true
	}
	if len(data.DataDictionaries) == 0 {
		data.DataDictionaries = seed.DataDictionaries
		changed = true
	}
	for _, item := range seed.DataDictionaries {
		if data.Next == nil || data.Next["dict"] < item.ID {
			ensureCounterAtLeast(data, "dict", item.ID)
			changed = true
		}
		if !hasDataDictionaryDefault(data.DataDictionaries, item.Type, item.Code) {
			data.DataDictionaries = append(data.DataDictionaries, item)
			changed = true
		}
	}
	if sanitizeDeliveryLicenseDefaults(data) {
		changed = true
	}
	if sanitizeProductInstanceDefaults(data) {
		changed = true
	}
	if sanitizeDeliveryExternalDefaults(data) {
		changed = true
	}
	if ensureBuiltinRolePermissions(data, "quality", "approval:read", "approval:write") {
		changed = true
	}
	if ensureBuiltinAdminSuperAdmin(data) {
		changed = true
	}
	return changed
}

func sanitizeProductInstanceDefaults(data *AppData) bool {
	changed := false
	for i := range data.ProductInstances {
		if hasDemoCredential(data.ProductInstances[i].ProbeToken) {
			data.ProductInstances[i].ProbeToken = productProbeToken(data.ProductInstances[i])
			changed = true
		}
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
	return changed
}

func productProbeToken(instance ProductInstance) string {
	seed := strings.Join([]string{"probe", instance.Watermark, instance.LicenseID, instance.CustomerName}, "|")
	sum := sha256.Sum256([]byte(seed))
	return "probe-" + hex.EncodeToString(sum[:])[:24]
}

const (
	builtinAdminUsername      = "admin"
	builtinSuperAdminRoleCode = "boss"
	builtinSuperAdminRoleName = "超级管理员"
)

func ensureBuiltinAdminSuperAdmin(data *AppData) bool {
	changed := false
	foundRole := false
	for i := range data.Roles {
		if data.Roles[i].Code != builtinSuperAdminRoleCode {
			continue
		}
		foundRole = true
		if data.Roles[i].Name != builtinSuperAdminRoleName {
			data.Roles[i].Name = builtinSuperAdminRoleName
			changed = true
		}
		if !hasOnlyWildcardPermission(data.Roles[i].Permissions) {
			data.Roles[i].Permissions = []string{"*"}
			changed = true
		}
		if normalizeDataScope(data.Roles[i].DataScope) != "group" {
			data.Roles[i].DataScope = "group"
			changed = true
		}
	}
	if !foundRole {
		data.Roles = append(data.Roles, builtinSuperAdminRole(nextID(data, "role")))
		changed = true
	}
	foundAdmin := false
	for i := range data.Users {
		if data.Users[i].Username != builtinAdminUsername {
			continue
		}
		foundAdmin = true
		if data.Users[i].DisplayName != builtinSuperAdminRoleName {
			data.Users[i].DisplayName = builtinSuperAdminRoleName
			changed = true
		}
		if data.Users[i].RoleCode != builtinSuperAdminRoleCode {
			data.Users[i].RoleCode = builtinSuperAdminRoleCode
			changed = true
		}
	}
	if !foundAdmin {
		admin := initialAdminUser(nil)
		admin.ID = nextID(data, "user")
		data.Users = append(data.Users, admin)
		changed = true
	}
	return changed
}

func builtinSuperAdminRole(id int64) Role {
	return Role{
		ID:          id,
		Code:        builtinSuperAdminRoleCode,
		Name:        builtinSuperAdminRoleName,
		Permissions: []string{"*"},
		DataScope:   "group",
	}
}

func hasOnlyWildcardPermission(items []string) bool {
	return len(items) == 1 && items[0] == "*"
}

func ensureBuiltinRolePermissions(data *AppData, roleCode string, permissions ...string) bool {
	changed := false
	for i := range data.Roles {
		if data.Roles[i].Code != roleCode || permissionGranted(data.Roles[i].Permissions, "*") {
			continue
		}
		for _, permission := range permissions {
			if permissionGranted(data.Roles[i].Permissions, permission) {
				continue
			}
			data.Roles[i].Permissions = append(data.Roles[i].Permissions, permission)
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
		scope := normalizeDataScope(data.Roles[i].DataScope)
		if data.Roles[i].DataScope != scope {
			data.Roles[i].DataScope = scope
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
		if hasDemoCredential(data.SCIMProviders[i].BearerToken) {
			data.SCIMProviders[i].BearerToken = ""
			if data.SCIMProviders[i].Status == "enabled" {
				data.SCIMProviders[i].Status = "disabled"
			}
			data.SCIMProviders[i].LastSyncAt = ""
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

func sanitizeDeliveryExternalDefaults(data *AppData) bool {
	changed := false
	for i := range data.DeviceCredentials {
		item := &data.DeviceCredentials[i]
		knownDemoKey := hasKnownDemoDeviceCredentialHash(item.KeyHash)
		activeWithoutKey := item.Status == "active" && strings.TrimSpace(item.KeyHash) == ""
		if !knownDemoKey && !activeWithoutKey {
			continue
		}
		if item.KeyHash != "" && knownDemoKey {
			item.KeyHash = ""
			changed = true
		}
		if item.Status == "active" {
			item.Status = "disabled"
			changed = true
		}
		if item.LastUsedAt != "" {
			item.LastUsedAt = ""
			changed = true
		}
	}
	for i := range data.ProductRenewalIntegrations {
		item := &data.ProductRenewalIntegrations[i]
		if !hasMockEndpoint(item.Endpoint) && !hasDemoCredential(item.Token) && !hasDemoCredential(item.Secret) {
			continue
		}
		if item.Endpoint != "" {
			item.Endpoint = ""
			changed = true
		}
		if item.Token != "" {
			item.Token = ""
			changed = true
		}
		if item.Secret != "" {
			item.Secret = ""
			changed = true
		}
		if item.Status == "active" {
			item.Status = "disabled"
			changed = true
		}
		if item.LastSyncAt != "" {
			item.LastSyncAt = ""
			changed = true
		}
		if item.LastError == "" {
			item.LastError = "待配置真实 " + fallback(item.Scenario, "external") + " endpoint"
			changed = true
		}
	}
	for i := range data.ProductMonitoringIntegrations {
		item := &data.ProductMonitoringIntegrations[i]
		if !hasMockEndpoint(item.Endpoint) && !hasDemoCredential(item.Token) {
			continue
		}
		if item.Endpoint != "" {
			item.Endpoint = ""
			changed = true
		}
		if item.Token != "" {
			item.Token = ""
			changed = true
		}
		if item.Status == "active" {
			item.Status = "disabled"
			changed = true
		}
		if item.LastEventAt != "" {
			item.LastEventAt = ""
			changed = true
		}
	}
	for i := range data.ProductAlertChannels {
		item := &data.ProductAlertChannels[i]
		external := item.Type != "sse" && item.Type != "local"
		invalidActiveExternal := external && item.Status == "active" && strings.TrimSpace(item.Endpoint) == ""
		if !hasMockEndpoint(item.Endpoint) && !hasDemoCredential(item.Token) && !hasDemoCredential(item.Secret) && !invalidActiveExternal {
			continue
		}
		if item.Endpoint != "" && (hasMockEndpoint(item.Endpoint) || invalidActiveExternal) {
			item.Endpoint = ""
			changed = true
		}
		if hasDemoCredential(item.Token) {
			item.Token = ""
			changed = true
		}
		if hasDemoCredential(item.Secret) {
			item.Secret = ""
			changed = true
		}
		if item.Status == "active" && external {
			item.Status = "disabled"
			changed = true
		}
		if item.LastDeliveredAt != "" {
			item.LastDeliveredAt = ""
			changed = true
		}
		if item.LastError == "" && external {
			item.LastError = "待配置真实 " + fallback(item.Type, item.Code) + " endpoint"
			changed = true
		}
	}
	for i := range data.IntegrationEndpoints {
		item := &data.IntegrationEndpoints[i]
		if !hasMockEndpoint(item.URL) {
			continue
		}
		item.URL = ""
		if item.Status == "online" {
			item.Status = "standby"
		}
		item.LastSyncAt = ""
		changed = true
	}
	activeChannels := activeProductAlertChannels(*data)
	for i := range data.ProductAlertRules {
		next, ok := sanitizeNotifyChannels(data.ProductAlertRules[i].NotifyChannels, activeChannels)
		if ok {
			data.ProductAlertRules[i].NotifyChannels = next
			changed = true
		}
	}
	for i := range data.ProductAlertPolicies {
		next, ok := sanitizeNotifyChannels(data.ProductAlertPolicies[i].NotifyChannels, activeChannels)
		if ok {
			data.ProductAlertPolicies[i].NotifyChannels = next
			changed = true
		}
	}
	return changed
}

func sanitizeDeliveryLicenseDefaults(data *AppData) bool {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(data.License.LicenseID)), "local-demo") &&
		!strings.HasPrefix(strings.ToLower(strings.TrimSpace(data.License.Signature)), "local-demo") {
		return false
	}
	data.License = LicenseInfo{
		CustomerName:          "待导入授权客户",
		Edition:               fallback(data.License.Edition, "ERP Appliance"),
		LastVerificationState: "missing",
		LastVerificationError: "待导入客户授权包",
	}
	return true
}

func hasMockEndpoint(endpoint string) bool {
	normalized := strings.ToLower(strings.TrimSpace(endpoint))
	return strings.HasPrefix(normalized, "mock://") ||
		strings.HasPrefix(normalized, "tax://") ||
		strings.Contains(normalized, "local-simulator")
}

func hasDemoCredential(value string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(value)), "demo")
}

func hasKnownDemoDeviceCredentialHash(keyHash string) bool {
	switch strings.TrimSpace(keyHash) {
	case sha256Hex("device-demo-key-1"),
		sha256Hex("device-demo-key-2"),
		sha256Hex("device-demo-key-3"),
		sha256Hex("driver-app-demo-key"),
		sha256Hex("scale-demo-key-1"),
		sha256Hex("plant-demo-key-1"),
		sha256Hex("gps-forwarder-demo-key"):
		return true
	default:
		return false
	}
}

func activeProductAlertChannels(data AppData) map[string]bool {
	out := map[string]bool{}
	for _, item := range data.ProductAlertChannels {
		if item.Status != "active" {
			continue
		}
		if code := strings.TrimSpace(item.Code); code != "" {
			out[code] = true
		}
		if typ := strings.TrimSpace(item.Type); typ != "" {
			out[typ] = true
		}
	}
	return out
}

func sanitizeNotifyChannels(channels []string, activeChannels map[string]bool) ([]string, bool) {
	next := []string{}
	seen := map[string]bool{}
	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if channel == "" || seen[channel] || !activeChannels[channel] {
			continue
		}
		seen[channel] = true
		next = append(next, channel)
	}
	if len(next) == 0 && activeChannels["sse"] {
		next = []string{"sse"}
	}
	if len(next) != len(channels) {
		return next, true
	}
	for i := range next {
		if next[i] != channels[i] {
			return next, true
		}
	}
	return channels, false
}

func hasDataDictionaryDefault(items []DataDictionary, typ, code string) bool {
	for _, item := range items {
		if item.Type == typ && item.Code == code {
			return true
		}
	}
	return false
}
