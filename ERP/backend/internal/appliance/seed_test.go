package appliance

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestSeedDataDictionariesComplete(t *testing.T) {
	data := SeedData()
	expectedTypes := map[string]int{
		"product_line":         6,
		"dispatch_status":      10,
		"ticket_type":          5,
		"invoice_type":         6,
		"quality_status":       4,
		"quality_result":       4,
		"yard_type":            4,
		"buffer_type":          6,
		"vehicle_type":         6,
		"plant_status":         5,
		"resource_status":      6,
		"shift_type":           4,
		"delivery_channel":     5,
		"payment_method":       6,
		"org_company_level":    5,
		"data_scope":           7,
		"config_status":        4,
		"account_status":       4,
		"laboratory_test_type": 8,
		"severity_level":       4,
		"sample_source_type":   6,
		"settlement_status":    6,
	}
	requiredCodes := map[string][]string{
		"dispatch_status":      {"assigned", "waiting_load", "in_transit", "signed", "completed", "cancelled"},
		"payment_method":       {"bank", "cash", "bill", "acceptance", "wechat", "alipay"},
		"quality_result":       {"pending", "passed", "failed", "retest"},
		"settlement_status":    {"unpaid", "partial", "settled", "overdue"},
		"sample_source_type":   {"manual", "production_batch", "raw_inspection", "quality_inspection"},
		"vehicle_type":         {"搅拌车", "自卸车", "沥青运输车", "粉料罐车"},
		"laboratory_test_type": {"marshall_stability", "asphalt_content", "gradation", "slump", "raw_material"},
	}

	seen := map[string]DataDictionary{}
	counts := map[string]int{}
	maxID := int64(0)
	for _, item := range data.DataDictionaries {
		if item.ID <= 0 {
			t.Fatalf("dictionary has invalid id: %+v", item)
		}
		if item.Type == "" || item.Code == "" || item.Label == "" {
			t.Fatalf("dictionary has blank key field: %+v", item)
		}
		if item.Sort <= 0 {
			t.Fatalf("dictionary has invalid sort: %+v", item)
		}
		key := item.Type + "/" + item.Code
		if previous, ok := seen[key]; ok {
			t.Fatalf("duplicate dictionary key %s: %+v and %+v", key, previous, item)
		}
		seen[key] = item
		counts[item.Type]++
		if item.ID > maxID {
			maxID = item.ID
		}
	}

	for dictType, minCount := range expectedTypes {
		if counts[dictType] < minCount {
			t.Fatalf("dictionary type %s has %d items, expected at least %d", dictType, counts[dictType], minCount)
		}
	}
	for dictType, codes := range requiredCodes {
		for _, code := range codes {
			if _, ok := seen[dictType+"/"+code]; !ok {
				t.Fatalf("dictionary type %s missing required code %s", dictType, code)
			}
		}
	}
	if data.Next["dict"] < maxID {
		t.Fatalf("dict counter %d is below max dictionary id %d", data.Next["dict"], maxID)
	}
}

func TestEnterpriseDefaultsBackfillMissingDictionaryItems(t *testing.T) {
	data := AppData{
		SchemaVersion: 2,
		Next:          map[string]int64{"dict": 1},
		DataDictionaries: []DataDictionary{
			{ID: 1, Type: "payment_method", Code: "bank", Label: "银行", Sort: 1, Status: "active"},
		},
	}

	if !ensureEnterpriseDefaults(&data) {
		t.Fatal("expected enterprise defaults to backfill missing dictionary items")
	}

	seen := map[string]bool{}
	maxID := int64(0)
	for _, item := range data.DataDictionaries {
		seen[item.Type+"/"+item.Code] = true
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	for _, key := range []string{
		"payment_method/acceptance",
		"payment_method/wechat",
		"dispatch_status/waiting_load",
		"settlement_status/overdue",
		"laboratory_test_type/asphalt_content",
	} {
		if !seen[key] {
			t.Fatalf("expected dictionary backfill for %s", key)
		}
	}
	if data.Next["dict"] < maxID {
		t.Fatalf("dict counter %d is below max dictionary id %d after backfill", data.Next["dict"], maxID)
	}
}

func TestInitialDataDoesNotShipDemoBusinessRecords(t *testing.T) {
	data := InitialData()
	emptyBusinessFields := []string{
		"Companies",
		"Sites",
		"Departments",
		"Plants",
		"Warehouses",
		"Silos",
		"Customers",
		"CustomerContacts",
		"CustomerBlacklists",
		"CustomerProfiles",
		"CustomerComplaints",
		"PricePolicies",
		"TaxRates",
		"Suppliers",
		"Carriers",
		"Projects",
		"Products",
		"Materials",
		"Vehicles",
		"Drivers",
		"VehicleDevices",
		"DeviceCredentials",
		"Contracts",
		"Orders",
		"ProductionPlans",
		"ProductionTasks",
		"ProductionBatches",
		"ProductionReports",
		"QualityInspections",
		"QualitySamples",
		"RawMaterialInspections",
		"LaboratorySamples",
		"LaboratoryTests",
		"LaboratoryEquipment",
		"LaboratoryCalibrations",
		"QualityExceptions",
		"Inventory",
		"StockYards",
		"StockYardPiles",
		"StockYardFlows",
		"InventoryTransfers",
		"InventoryStocktakes",
		"InventoryBatchTraces",
		"PurchaseRequests",
		"PurchaseOrders",
		"RawMaterialReceipts",
		"InventoryFlows",
		"DispatchOrders",
		"DispatchSchedules",
		"ScaleDevices",
		"ScaleTickets",
		"ScaleWeightRecords",
		"ScaleDeviceEvents",
		"DeliveryNotes",
		"DeliverySignLinks",
		"TicketPrintLogs",
		"TicketVoidLogs",
		"DeliverySigns",
		"DeliverySignAttachments",
		"Statements",
		"SalesInvoices",
		"RedLetterInfos",
		"TaxGatewaySubmissions",
		"Receivables",
		"Receipts",
		"PaymentPlans",
		"CollectionTasks",
		"CollectionTemplates",
		"CollectionDispatches",
		"SupplierStatements",
		"Payables",
		"Payments",
		"TransportSettlements",
		"TransportSettlementItems",
		"CostCalcs",
		"ProjectProfits",
		"Locations",
		"LatestLocations",
		"GeoFences",
		"GeoFenceEvents",
		"VehicleAlarms",
		"RuleDefinitions",
		"Notifications",
		"IntegrationEndpoints",
		"ApprovalFlows",
		"ApprovalTasks",
		"WorkflowDefinitions",
		"WorkflowInstances",
		"WorkflowTasks",
		"WorkflowEvents",
		"WorkflowLogs",
		"WorkflowOutbox",
		"WorkflowSubscriptions",
		"WorkflowDeliveries",
	}
	value := reflect.ValueOf(data)
	for _, field := range emptyBusinessFields {
		slice := value.FieldByName(field)
		if !slice.IsValid() {
			t.Fatalf("initial data test references unknown field %s", field)
		}
		if slice.Len() != 0 {
			t.Fatalf("initial data must not ship demo business records in %s, got %d", field, slice.Len())
		}
	}
	if data.GroupProfile.Code != "UNCONFIGURED" || data.License.LastVerificationState != "missing" {
		t.Fatalf("initial data must start as unconfigured customer appliance, got group=%+v license=%+v", data.GroupProfile, data.License)
	}
	if len(data.DataDictionaries) == 0 || len(data.Modules) == 0 || len(data.Roles) != 1 || len(data.Users) != 1 {
		t.Fatalf("initial data must keep only runtime foundation, got dict=%d modules=%d roles=%d users=%d", len(data.DataDictionaries), len(data.Modules), len(data.Roles), len(data.Users))
	}
}

func TestDemoBusinessSeedPurgeKeepsModifiedBusinessData(t *testing.T) {
	seed := SeedData()
	data := SeedData()
	data.Orders = append(data.Orders, SalesOrder{ID: 999, OrderNo: "SO-CUSTOMER-REAL", CustomerID: 1, ProjectID: 1, ProductID: 1, SiteID: 1})
	if purgeDemoBusinessSeed(&data, seed) {
		t.Fatalf("modified business data must not be purged")
	}
	if len(data.Orders) != len(seed.Orders)+1 {
		t.Fatalf("modified order was lost: %+v", data.Orders)
	}
}

func TestSeedDataExternalIntegrationsRequireRealConfiguration(t *testing.T) {
	data := SeedData()
	assertNoMockEndpoint := func(kind, code, endpoint string) {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(endpoint)), "mock://") {
			t.Fatalf("%s %s must not use mock endpoint in seed data: %s", kind, code, endpoint)
		}
	}
	assertNoDemoSecret := func(kind, code, value string) {
		if strings.Contains(strings.ToLower(strings.TrimSpace(value)), "demo") {
			t.Fatalf("%s %s must not ship demo credential in seed data", kind, code)
		}
	}
	assertNoLocalSimulator := func(kind, code, value string) {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if strings.Contains(normalized, "local-simulator") || strings.Contains(normalized, "local-tax") {
			t.Fatalf("%s %s must not ship local simulator integration metadata: %s", kind, code, value)
		}
	}
	if data.License.LicenseID != "" || data.License.Signature != "" || data.License.LastVerificationState != "missing" {
		t.Fatalf("seed license must require customer license import, got %+v", data.License)
	}
	if len(data.TaxGatewaySubmissions) != 0 {
		t.Fatalf("seed data must not include pre-submitted tax gateway records, got %+v", data.TaxGatewaySubmissions)
	}

	for _, item := range data.ProductRenewalIntegrations {
		assertNoMockEndpoint("renewal integration", item.Code, item.Endpoint)
		assertNoDemoSecret("renewal integration token", item.Code, item.Token)
		assertNoDemoSecret("renewal integration secret", item.Code, item.Secret)
		if item.Status == "active" && strings.TrimSpace(item.Endpoint) == "" {
			t.Fatalf("active renewal integration %s must have a real endpoint", item.Code)
		}
	}
	for _, item := range data.SCIMProviders {
		assertNoDemoSecret("scim provider token", item.Code, item.BearerToken)
		if item.Status == "enabled" && strings.TrimSpace(item.BearerToken) == "" {
			t.Fatalf("enabled scim provider %s must have a bearer token", item.Code)
		}
	}
	for _, item := range data.DeviceCredentials {
		if item.Status == "active" {
			t.Fatalf("seed device credential %s must stay disabled until a customer key is configured", item.DeviceNo)
		}
		if item.KeyHash != "" {
			t.Fatalf("seed device credential %s must not ship a reusable key hash", item.DeviceNo)
		}
	}
	for _, item := range data.ContractAttachments {
		assertNoDemoSecret("contract attachment checksum", item.FileName, item.Checksum)
	}
	for _, item := range data.IntegrationEndpoints {
		assertNoDemoSecret("integration endpoint url", item.Name, item.URL)
		assertNoLocalSimulator("integration endpoint url", item.Name, item.URL)
		normalizedURL := strings.ToLower(strings.TrimSpace(item.URL))
		if strings.Contains(normalizedURL, "127.0.0.1") || strings.Contains(normalizedURL, "localhost") {
			t.Fatalf("integration endpoint %s must not ship a development host url: %s", item.Name, item.URL)
		}
		if item.Type == "collection" && item.Status == "online" && strings.TrimSpace(item.URL) == "" {
			t.Fatalf("collection endpoint %s must not be online without a configured url", item.Name)
		}
	}
	for _, item := range data.ProductRenewalInvoices {
		assertNoLocalSimulator("renewal invoice external request", item.InvoiceNo, item.ExternalRequest)
		if item.TaxStatus == "accepted" && strings.TrimSpace(item.ExternalRequest) == "" {
			t.Fatalf("accepted renewal invoice %s must have a real external request id", item.InvoiceNo)
		}
	}
	for _, item := range data.ProductRenewalSyncRecords {
		if item.Scenario == "tax" && item.ResourceType == "invoice" && item.Status == "succeeded" {
			if strings.TrimSpace(item.ExternalRequestID) == "" {
				t.Fatalf("succeeded renewal tax sync %s must have a real external request id", item.SyncNo)
			}
			if strings.Contains(strings.ToLower(item.ExternalRequestID), "local") {
				t.Fatalf("succeeded renewal tax sync %s must not ship local request metadata: %s", item.SyncNo, item.ExternalRequestID)
			}
		}
	}
	for _, item := range data.ProductInstances {
		assertNoDemoSecret("product instance probe token", item.Watermark, item.ProbeToken)
		if item.ProbeEnabled && strings.TrimSpace(item.ProbeToken) == "" {
			t.Fatalf("probe-enabled product instance %s must have a probe token", item.Watermark)
		}
	}
	for _, item := range data.ProductMonitoringIntegrations {
		assertNoMockEndpoint("monitoring integration", item.Code, item.Endpoint)
		assertNoDemoSecret("monitoring integration token", item.Code, item.Token)
		if item.Status == "active" && strings.TrimSpace(item.Token) == "" {
			t.Fatalf("active monitoring integration %s must have a token", item.Code)
		}
	}
	for _, item := range data.ProductAlertChannels {
		assertNoMockEndpoint("alert channel", item.Code, item.Endpoint)
		assertNoDemoSecret("alert channel token", item.Code, item.Token)
		assertNoDemoSecret("alert channel secret", item.Code, item.Secret)
		if item.Status == "active" && item.Type != "sse" && item.Type != "local" {
			t.Fatalf("external alert channel %s must stay disabled until a real endpoint is configured", item.Code)
		}
	}
	for _, item := range data.ProductAlertRules {
		for _, channel := range item.NotifyChannels {
			if channel != "sse" && channel != "local" {
				t.Fatalf("seed alert rule %s must not notify external channel %s by default", item.RuleNo, channel)
			}
		}
	}
	for _, item := range data.ProductAlertPolicies {
		for _, channel := range item.NotifyChannels {
			if channel != "sse" && channel != "local" {
				t.Fatalf("seed alert policy %s must not notify external channel %s by default", item.PolicyNo, channel)
			}
		}
	}
}

func TestSeedDataBusinessClosureReferencesAreConsistent(t *testing.T) {
	data := SeedData()

	orders := map[int64]SalesOrder{}
	for _, item := range data.Orders {
		orders[item.ID] = item
	}
	order := orders[1]
	if order.ID == 0 || order.CustomerID == 0 || order.ProjectID == 0 || order.ProductID == 0 || order.SiteID == 0 {
		t.Fatalf("seed sales order 1 must anchor customer/project/product/site references, got %+v", order)
	}
	if !seedFloatEqual(order.TotalAmount, order.PlanQuantity*order.UnitPrice) {
		t.Fatalf("order amount mismatch: %+v", order)
	}

	var plan ProductionPlan
	for _, item := range data.ProductionPlans {
		if item.OrderID == order.ID && item.ProductID == order.ProductID {
			plan = item
			break
		}
	}
	if plan.ID == 0 || plan.PlanQuantity <= 0 || plan.ProducedQty <= 0 {
		t.Fatalf("order %s must have an active production plan, got %+v", order.OrderNo, plan)
	}

	var task ProductionTask
	for _, item := range data.ProductionTasks {
		if item.PlanID == plan.ID && item.OrderID == order.ID && item.ProductID == order.ProductID {
			task = item
			break
		}
	}
	if task.ID == 0 || task.PlanQty <= 0 || task.ProducedQty <= 0 {
		t.Fatalf("production plan %s must have a task, got %+v", plan.PlanNo, task)
	}

	var batch ProductionBatch
	for _, item := range data.ProductionBatches {
		if item.TaskID == task.ID && item.PlanID == plan.ID && item.OrderID == order.ID {
			batch = item
			break
		}
	}
	if batch.ID == 0 || batch.Quantity <= 0 || batch.QualityStatus != "passed" {
		t.Fatalf("production task %s must have a passed batch, got %+v", task.TaskNo, batch)
	}

	var dispatch DispatchOrder
	for _, item := range data.DispatchOrders {
		if item.OrderID == order.ID && item.ProductID == order.ProductID && item.SignedQty > 0 && item.Status == "completed" {
			dispatch = item
			break
		}
	}
	if dispatch.ID == 0 || dispatch.LineID == 0 || dispatch.LineSeq == 0 {
		t.Fatalf("order %s must have a completed signed dispatch, got %+v", order.OrderNo, dispatch)
	}
	if dispatch.SignedQty > dispatch.LoadedQty || dispatch.LoadedQty > dispatch.PlanQuantity {
		t.Fatalf("dispatch quantities must progress plan >= loaded >= signed, got %+v", dispatch)
	}

	var ticket ScaleTicket
	for _, item := range data.ScaleTickets {
		if item.TicketType == "product_out" && item.DispatchID == dispatch.ID && item.OrderID == order.ID {
			ticket = item
			break
		}
	}
	if ticket.ID == 0 || ticket.VehicleID != dispatch.VehicleID || ticket.NetWeight <= 0 || ticket.Status != "valid" || ticket.SignStatus != "signed" {
		t.Fatalf("dispatch %s must have a valid signed product-out ticket, got %+v", dispatch.DispatchNo, ticket)
	}

	var note DeliveryNote
	for _, item := range data.DeliveryNotes {
		if item.TicketID == ticket.ID && item.DispatchID == dispatch.ID && item.OrderID == order.ID {
			note = item
			break
		}
	}
	if note.ID == 0 || note.Status != "signed" {
		t.Fatalf("ticket %s must have a signed delivery note, got %+v", ticket.TicketNo, note)
	}

	var link DeliverySignLink
	for _, item := range data.DeliverySignLinks {
		if item.TicketID == ticket.ID && item.DispatchID == dispatch.ID && item.OrderID == order.ID {
			link = item
			break
		}
	}
	if link.ID == 0 || link.Status != "used" || link.CustomerID != order.CustomerID || link.ProjectID != order.ProjectID {
		t.Fatalf("delivery note %s must have a used customer sign link, got %+v", note.NoteNo, link)
	}

	var sign DeliverySign
	for _, item := range data.DeliverySigns {
		if item.LinkID == link.ID && item.TicketID == ticket.ID && item.DispatchID == dispatch.ID && item.OrderID == order.ID {
			sign = item
			break
		}
	}
	if sign.ID == 0 || !seedFloatEqual(sign.SignedQty, dispatch.SignedQty) || sign.SignedAt == "" {
		t.Fatalf("dispatch %s must have a matching delivery sign, got %+v", dispatch.DispatchNo, sign)
	}

	var statement Statement
	var statementItem StatementItem
	for _, item := range data.Statements {
		for _, line := range item.Items {
			if line.SignID == sign.ID && line.OrderID == order.ID && line.TicketID == ticket.ID && line.ProductID == order.ProductID {
				statement = item
				statementItem = line
				break
			}
		}
		if statement.ID != 0 {
			break
		}
	}
	if statement.ID == 0 || statement.CustomerID != order.CustomerID || statement.ProjectID != order.ProjectID {
		t.Fatalf("delivery sign %s must feed a customer statement, got statement=%+v item=%+v", sign.SignNo, statement, statementItem)
	}
	if !seedFloatEqual(statementItem.Quantity, sign.SignedQty) || !seedFloatEqual(statementItem.Amount, statementItem.Quantity*statementItem.UnitPrice) {
		t.Fatalf("statement item must match signed quantity and price, got %+v", statementItem)
	}
	if !seedFloatEqual(statement.TotalQty, statementItem.Quantity) || !seedFloatEqual(statement.TotalAmount, statementItem.Amount) {
		t.Fatalf("statement totals must match items, got statement=%+v item=%+v", statement, statementItem)
	}
	if statement.Status != "invoiced" || statement.ConfirmedBy == "" || statement.ConfirmedAt == "" {
		t.Fatalf("statement %s must be confirmed and invoiced before invoice/receivable seed records, got %+v", statement.StatementNo, statement)
	}

	var invoice SalesInvoice
	for _, item := range data.SalesInvoices {
		if item.StatementID == statement.ID && item.CustomerID == order.CustomerID {
			invoice = item
			break
		}
	}
	if invoice.ID == 0 || !seedFloatEqual(invoice.Amount, statement.TotalAmount) || !seedFloatEqual(invoice.TaxAmount, round(statement.TotalAmount*invoice.TaxRate)) {
		t.Fatalf("statement %s must have a matching invoice, got %+v", statement.StatementNo, invoice)
	}
	if invoice.TaxStatus != "pending" || invoice.TaxControlNo != "" || invoice.FileURL != "" || invoice.IssuedAt != "" {
		t.Fatalf("seed invoice must wait for customer tax gateway configuration, got %+v", invoice)
	}
	if len(data.TaxGatewaySubmissions) != 0 {
		t.Fatalf("seed data must not pre-submit tax gateway records, got %+v", data.TaxGatewaySubmissions)
	}

	var receivable Receivable
	for _, item := range data.Receivables {
		if item.StatementID == statement.ID && item.InvoiceID == invoice.ID && item.CustomerID == order.CustomerID {
			receivable = item
			break
		}
	}
	if receivable.ID == 0 || !seedFloatEqual(receivable.Amount, invoice.Amount) || receivable.ReceivedAmount <= 0 || receivable.ReceivedAmount > receivable.Amount {
		t.Fatalf("invoice %s must feed a partially paid receivable, got %+v", invoice.InvoiceNo, receivable)
	}

	var receipt Receipt
	for _, item := range data.Receipts {
		if item.ReceivableID == receivable.ID && item.CustomerID == order.CustomerID {
			receipt = item
			break
		}
	}
	if receipt.ID == 0 || receipt.Status != "confirmed" || !seedFloatEqual(receipt.Amount, receivable.ReceivedAmount) {
		t.Fatalf("receivable %s must have a confirmed receipt, got %+v", receivable.BillNo, receipt)
	}

	var paymentPlan PaymentPlan
	for _, item := range data.PaymentPlans {
		if item.ReceivableID == receivable.ID && item.CustomerID == order.CustomerID {
			paymentPlan = item
			break
		}
	}
	remaining := receivable.Amount - receivable.ReceivedAmount
	if paymentPlan.ID == 0 || paymentPlan.Amount <= 0 || paymentPlan.Amount > remaining || paymentPlan.Status != "planned" {
		t.Fatalf("receivable %s must have a valid follow-up payment plan, got %+v", receivable.BillNo, paymentPlan)
	}
}

func TestCollectionEndpointRequiresRealEndpointExceptManualPhone(t *testing.T) {
	data := SeedData()
	endpoint, endpointIndex := ensureCollectionEndpoint(&data, "sms")
	if endpointIndex < 0 || endpoint.URL != "" || endpoint.Status != "standby" {
		t.Fatalf("seed collection sms endpoint must start as standby without a url, got %+v", endpoint)
	}
	if err := validateCollectionEndpointConfigured(endpoint, "sms"); err == nil {
		t.Fatalf("expected unconfigured sms collection endpoint to be rejected")
	}

	phoneEndpoint, _ := ensureCollectionEndpoint(&data, "phone")
	if err := validateCollectionEndpointConfigured(phoneEndpoint, "phone"); err != nil {
		t.Fatalf("manual phone collection should not require an external endpoint: %v", err)
	}

	data.IntegrationEndpoints[endpointIndex].URL = "https://sms-provider.example.test/send"
	data.IntegrationEndpoints[endpointIndex].Status = "online"
	endpoint, _ = ensureCollectionEndpoint(&data, "sms")
	if endpoint.URL != "https://sms-provider.example.test/send" || endpoint.Status != "online" {
		t.Fatalf("expected configured real sms collection endpoint, got %+v", endpoint)
	}
	if err := validateCollectionEndpointConfigured(endpoint, "sms"); err != nil {
		t.Fatalf("real sms collection endpoint should validate: %v", err)
	}
}

func TestEnterpriseDefaultsSanitizesLegacyMockExternalIntegrations(t *testing.T) {
	data := AppData{
		SchemaVersion: 2,
		Next:          map[string]int64{"productInstance": 1},
		License: LicenseInfo{
			LicenseID: "local-demo", CustomerName: "演示客户", Watermark: "CBMP-DEMO", ExpiresAt: "2027-12-31",
			Edition: "ERP Appliance", Modules: []string{"erp"}, MaxSites: 20, MaxVehicles: 5000, Issuer: "local-demo", Signature: "local-demo-signature",
		},
		ProductInstances: []ProductInstance{{
			ID: 1, CustomerName: "测试客户", LicenseID: "LIC-TEST", Watermark: "WM-TEST", Status: "online", ProbeEnabled: true, ProbeToken: "probe-demo-token",
		}},
		SCIMProviders: []SCIMProvider{{
			ID: 1, Code: "enterprise-scim", Name: "Enterprise SCIM", BearerToken: "demo-scim-token", Status: "enabled",
		}},
		DeviceCredentials: []DeviceCredential{
			{ID: 1, DeviceNo: "GPS1000001", KeyHash: sha256Hex("device-demo-key-1"), Scopes: []string{"location:report"}, Status: "active", LastUsedAt: "2026-06-19 09:00:00"},
			{ID: 99, DeviceNo: "CUSTOM-DEVICE", KeyHash: sha256Hex("customer-device-key"), Scopes: []string{"location:report"}, Status: "active", LastUsedAt: "2026-06-19 09:05:00"},
		},
		ProductRenewalIntegrations: []ProductRenewalIntegration{{
			ID: 1, Code: "tax_gateway", Provider: "tax", Scenario: "tax", Endpoint: "mock://success", Token: "tax-demo-token", Secret: "tax-demo-secret", Status: "active", LastSyncAt: "2026-06-19 09:20:00",
		}},
		ProductMonitoringIntegrations: []ProductMonitoringIntegration{{
			ID: 1, Code: "prometheus-site", Provider: "prometheus", Endpoint: "mock://operations-platform-monitoring", Token: "mon-demo-token", Status: "active", LastEventAt: "2026-06-19 08:50:00",
		}},
		ProductAlertChannels: []ProductAlertChannel{
			{ID: 1, Code: "sse", Type: "sse", Status: "active"},
			{ID: 2, Code: "webhook", Type: "webhook", Endpoint: "mock://success", Status: "active", LastDeliveredAt: "2026-06-19 08:10:00"},
			{ID: 3, Code: "enterprise_wechat", Type: "enterprise_wechat", Token: "wecom-demo-token", Secret: "wecom-demo-secret", Status: "active"},
			{ID: 4, Code: "sms", Type: "sms", Status: "active"},
		},
		IntegrationEndpoints: []IntegrationEndpoint{
			{ID: 1, Name: "财务税控接口", Type: "finance", Protocol: "rest/http", URL: "tax://local-simulator", Status: "online", LastSyncAt: "2026-06-19 09:30:00"},
			{ID: 2, Name: "催收短信通道", Type: "collection", Protocol: "sms", URL: "collection://local-simulator/sms", Status: "online", LastSyncAt: "2026-06-19 09:31:00"},
		},
		ProductAlertRules: []ProductAlertRule{{
			ID: 1, RuleNo: "AR-LEGACY", NotifyChannels: []string{"sse", "webhook", "sms"}, Status: "active",
		}},
		ProductAlertPolicies: []ProductAlertPolicy{{
			ID: 1, PolicyNo: "AP-LEGACY", NotifyChannels: []string{"webhook", "sse"}, Status: "active",
		}},
	}

	if !ensureEnterpriseDefaults(&data) {
		t.Fatal("expected enterprise defaults to sanitize legacy mock integrations")
	}
	if strings.Contains(data.ProductInstances[0].ProbeToken, "demo") || data.ProductInstances[0].ProbeToken == "" {
		t.Fatalf("expected regenerated probe token, got %+v", data.ProductInstances[0])
	}
	if data.License.LicenseID != "" || data.License.Signature != "" || data.License.LastVerificationState != "missing" {
		t.Fatalf("expected local demo license sanitized to missing state, got %+v", data.License)
	}
	if data.SCIMProviders[0].BearerToken != "" || data.SCIMProviders[0].Status != "disabled" {
		t.Fatalf("expected disabled scim provider without demo token, got %+v", data.SCIMProviders[0])
	}
	if data.DeviceCredentials[0].KeyHash != "" || data.DeviceCredentials[0].Status != "disabled" || data.DeviceCredentials[0].LastUsedAt != "" {
		t.Fatalf("expected disabled device credential without demo key hash, got %+v", data.DeviceCredentials[0])
	}
	if data.DeviceCredentials[1].KeyHash == "" || data.DeviceCredentials[1].Status != "active" || data.DeviceCredentials[1].LastUsedAt == "" {
		t.Fatalf("expected customer device credential to remain active, got %+v", data.DeviceCredentials[1])
	}
	renewal := data.ProductRenewalIntegrations[0]
	if renewal.Status != "disabled" || renewal.Endpoint != "" || renewal.Token != "" || renewal.Secret != "" || renewal.LastSyncAt != "" || renewal.LastError == "" {
		t.Fatalf("expected disabled sanitized renewal integration, got %+v", renewal)
	}
	monitoring := data.ProductMonitoringIntegrations[0]
	if monitoring.Status != "disabled" || monitoring.Endpoint != "" || monitoring.Token != "" || monitoring.LastEventAt != "" {
		t.Fatalf("expected disabled sanitized monitoring integration, got %+v", monitoring)
	}
	for _, channel := range data.ProductAlertChannels {
		if channel.Type == "sse" {
			continue
		}
		if channel.Status != "disabled" || channel.Endpoint != "" || strings.Contains(channel.Token, "demo") || strings.Contains(channel.Secret, "demo") {
			t.Fatalf("expected disabled sanitized alert channel, got %+v", channel)
		}
	}
	for _, endpoint := range data.IntegrationEndpoints {
		if endpoint.URL != "" || endpoint.Status != "standby" || endpoint.LastSyncAt != "" {
			t.Fatalf("expected legacy simulator integration endpoint to be cleared, got %+v", endpoint)
		}
	}
	if got := strings.Join(data.ProductAlertRules[0].NotifyChannels, ","); got != "sse" {
		t.Fatalf("expected rule channels sanitized to sse, got %s", got)
	}
	if got := strings.Join(data.ProductAlertPolicies[0].NotifyChannels, ","); got != "sse" {
		t.Fatalf("expected policy channels sanitized to sse, got %s", got)
	}
}

func seedFloatEqual(left, right float64) bool {
	return math.Abs(left-right) < 0.000001
}
