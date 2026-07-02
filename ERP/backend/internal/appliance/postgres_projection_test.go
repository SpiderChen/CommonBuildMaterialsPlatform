package appliance

import (
	"strings"
	"testing"
)

func TestBusinessProjectionInsertsCoverCoreEnterpriseTables(t *testing.T) {
	data := AppData{
		CustomerContacts: []CustomerContact{{
			ID: 1, CustomerID: 2, Name: "王经理", Phone: "13800010001", Role: "项目负责人", IsDefault: true, Status: "active",
		}},
		CustomerBlacklists: []CustomerBlacklist{{
			ID: 2, CustomerID: 2, CustomerName: "测试客户", Reason: "回款风险", Scope: "sales_order", Severity: "high",
			BlockOrders: true, Status: "active", CreatedAt: "2026-06-18 12:00:00", Actor: "admin",
		}},
		CustomerProfiles: []CustomerProfile{{
			ID: 3, CustomerID: 2, CustomerName: "测试客户", Grade: "B", RiskLevel: "medium", CreditScore: 78,
			Tags: []string{"关注回款"}, Status: "active", UpdatedAt: "2026-06-18 12:05:00", Actor: "admin",
		}},
		CustomerComplaints: []CustomerComplaint{{
			ID: 4, ComplaintNo: "CP-1", CustomerID: 2, ProjectID: 3, Title: "等待时间过长", Level: "medium",
			Status: "open", Owner: "调度主管", CreatedAt: "2026-06-18 12:10:00",
		}},
		ContractAttachments: []ContractAttachment{{
			ID: 5, ContractID: 6, CustomerID: 2, FileName: "合同.pdf", FileType: "contract_pdf", URL: "vault://contracts/6.pdf",
			Checksum: "sha256:test", Status: "active", UploadedBy: "admin", UploadedAt: "2026-06-18 12:15:00",
		}},
		PricePolicies: []PricePolicy{{
			ID: 6, CustomerID: 2, ProjectID: 3, ProductID: 4, CustomerGrade: "B", FloorPrice: 450, SalePrice: 480,
			TaxRateID: 7, EffectiveFrom: "2026-06-01", EffectiveTo: "2027-05-31", Status: "active",
		}},
		TaxRates: []TaxRate{{
			ID: 7, Name: "建材销售 13%", Rate: 0.13, Scope: "sales", Status: "active",
		}},
		Orders: []SalesOrder{{
			ID: 1, OrderNo: "SO-1", CustomerID: 2, ProjectID: 3, ProductID: 4, SiteID: 5,
			ProductLine: "asphalt", PlanQuantity: 30, SignedQty: 10, UnitPrice: 480, Status: "dispatching", RiskFlag: "ok", PlanTime: "2026-06-18 10:00:00", CreatedAt: "2026-06-18 09:00:00",
		}},
		DispatchOrders: []DispatchOrder{{
			ID: 11, DispatchNo: "DO-1", OrderID: 1, VehicleID: 7, DriverID: 8, SiteID: 5, ProjectID: 3,
			PlanQuantity: 30, LoadedQty: 20, SignedQty: 10, Status: "loaded", Exception: "", CreatedAt: "2026-06-18 09:30:00", UpdatedAt: "2026-06-18 09:50:00",
		}},
		DispatchSchedules: []DispatchSchedule{{
			ID: 12, ScheduleNo: "DS-1", SiteID: 5, VehicleID: 7, DriverID: 8, CarrierID: 9, ShiftDate: "2026-06-18",
			Shift: "day", CapacityQty: 120, AssignedQty: 30, Status: "active", CreatedAt: "2026-06-18 08:00:00", UpdatedAt: "2026-06-18 09:30:00",
		}},
		TransportSettlements: []TransportSettlement{{
			ID: 13, SettlementNo: "TS-1", CarrierID: 9, Period: "2026-06", TripCount: 1, Amount: 480, Status: "draft",
		}},
		TransportSettlementItems: []TransportSettlementItem{{
			ID: 14, SettlementID: 13, DispatchID: 11, DispatchNo: "DO-1", CarrierID: 9, VehicleID: 7, DriverID: 8,
			Quantity: 10, Amount: 480, Status: "pending", CreatedAt: "2026-06-18 12:00:00",
		}},
		ProductionBatches: []ProductionBatch{{
			ID: 21, BatchNo: "PB-1", TaskID: 20, PlanID: 19, OrderID: 1, SiteID: 5, ProductID: 4,
			Quantity: 30, PlantCode: "PLANT-1", QualityStatus: "pending", Status: "completed", StartedAt: "2026-06-18 09:10:00", CompletedAt: "2026-06-18 09:25:00",
		}},
		MixDesigns: []MixDesign{{
			ID: 22, ProductID: 4, SiteID: 5, Code: "MD-1", Version: "v1", StrengthGrade: "AC-13", Slump: "油石比 5.1%",
			Scope: "测试配比", Status: "approved", IsCurrent: true, EffectiveFrom: "2026-01-01", EffectiveTo: "2026-12-31",
			ApprovedBy: "admin", ApprovedAt: "2026-06-18 08:00:00", CreatedBy: "admin", CreatedAt: "2026-06-18 07:50:00", UpdatedAt: "2026-06-18 08:00:00",
		}},
		MixDesignTrialRuns: []MixDesignTrialRun{{
			ID: 23, TrialNo: "MTR-1", MixDesignID: 22, ProductID: 4, SiteID: 5, TargetStrength: "AC-13", Slump: "油石比 5.1%",
			Water: 0, SandRate: 42, AdmixtureRate: 1.2, Strength7d: 8.2, Strength28d: 9.1, Result: "passed", Conclusion: "ok",
			Tester: "quality", TestedAt: "2026-06-18 08:30:00", CreatedAt: "2026-06-18 08:20:00",
		}},
		LaboratorySamples: []LaboratorySample{{
			ID: 24, SampleNo: "LS-1", SourceType: "manual", SiteID: 5, ProductID: 4, MixDesignID: 22,
			SampleType: "marshall_stability", Status: "completed", Result: "passed", PlannedTestAt: "2026-06-18", CollectedAt: "2026-06-18 08:00:00", CreatedBy: "quality",
		}},
		LaboratoryTests: []LaboratoryTestRecord{{
			ID: 25, TestNo: "LT-1", SampleID: 24, EquipmentID: 26, SiteID: 5, TestType: "marshall_stability",
			Metric: "stability", Value: 9.1, Unit: "kN", Result: "passed", Status: "reviewed", Tester: "quality", TestedAt: "2026-06-18 08:30:00", Reviewer: "admin", ReviewedAt: "2026-06-18 09:00:00",
		}},
		LaboratoryEquipment: []LaboratoryEquipment{{
			ID: 26, EquipmentNo: "EQ-1", Name: "马歇尔稳定度仪", SiteID: 5, Model: "LWD-3", SerialNo: "SER-1",
			Status: "active", CalibrationCycleDays: 180, LastCalibrationAt: "2026-06-01", NextCalibrationAt: "2026-11-28", CreatedAt: "2026-01-01 08:00:00",
		}},
		LaboratoryCalibrations: []LaboratoryCalibration{{
			ID: 27, CalibrationNo: "LC-1", EquipmentID: 26, SiteID: 5, Result: "passed", CalibratedAt: "2026-06-01",
			NextDueAt: "2026-11-28", CertificateNo: "CAL-1", Agency: "计量中心", Operator: "quality",
		}},
		QualityExceptions: []QualityException{{
			ID: 28, ExceptionNo: "QE-1", SourceType: "laboratory_test", SourceID: 25, SiteID: 5, Severity: "high",
			Title: "测试异常", Status: "open", Responsible: "quality", CreatedAt: "2026-06-18 09:10:00",
		}},
		ScaleTickets: []ScaleTicket{{
			ID: 31, TicketNo: "ST-1", TicketType: "outbound", DispatchID: 11, OrderID: 1, SiteID: 5, VehicleID: 7,
			PlateNo: "粤B12345", GrossWeight: 31.5, TareWeight: 12.5, NetWeight: 19, SignStatus: "pending", SettlementStatus: "pending", Status: "normal", CreatedAt: "2026-06-18 09:35:00",
		}},
		ScaleDeviceEvents: []ScaleDeviceEvent{{
			ID: 41, EventNo: "SDE-1", DeviceID: 9, DeviceCode: "SCALE-1", TicketID: 31, PlateNo: "粤B12345",
			Weight: 31.5, WeightType: "gross", Stable: true, CheatFlag: false, Status: "accepted", ReceivedAt: "2026-06-18 09:36:00",
		}},
		DeliverySignLinks: []DeliverySignLink{{
			ID: 50, LinkNo: "SL-1", DispatchID: 11, TicketID: 31, OrderID: 1, CustomerID: 2, ProjectID: 3,
			Channel: "sms", Phone: "13800010001", Status: "sent", SentAt: "2026-06-18 10:00:00", ExpiresAt: "2026-06-25 10:00:00", CreatedAt: "2026-06-18 10:00:00",
		}},
		DeliverySigns: []DeliverySign{{
			ID: 51, SignNo: "SIGN-1", DispatchID: 11, LinkID: 50, TicketID: 31, OrderID: 1, CustomerID: 2, ProjectID: 3,
			SignedQty: 10, SignedAt: "2026-06-18 10:30:00",
		}},
		DeliverySignAttachments: []DeliverySignAttachment{{
			ID: 52, SignID: 51, DispatchID: 11, TicketID: 31, FileName: "site.jpg", FileType: "photo",
			URL: "minio://delivery/site.jpg", Checksum: "sha256:site", UploadedBy: "王经理", UploadedAt: "2026-06-18 10:31:00",
		}},
		SalesInvoices: []SalesInvoice{{
			ID: 61, InvoiceNo: "INV-1", StatementID: 60, CustomerID: 2, Amount: 4800, TaxAmount: 624, TaxStatus: "submitted", Status: "issued", IssuedAt: "2026-06-18 11:00:00",
		}},
		RedLetterInfos: []RedLetterInfo{{
			ID: 63, InfoNo: "RLI-1", OriginalInvoiceID: 61, OriginalInvoiceNo: "INV-1", CustomerID: 2,
			Amount: -4800, TaxAmount: -624, Reason: "测试红字信息表", Applicant: "admin", Status: "approved",
			TaxControlNo: "RLI-TAX-1", RequestedAt: "2026-06-18 11:10:00", ApprovedBy: "admin", ApprovedAt: "2026-06-18 11:11:00",
		}},
		TaxGatewaySubmissions: []TaxGatewaySubmission{{
			ID: 62, SubmissionNo: "TGS-1", InvoiceID: 61, InvoiceNo: "INV-1", Provider: "external-tax-cn", Endpoint: "https://tax.example/api",
			RequestID: "tax-request-1", Status: "submitted", TaxControlNo: "TAX-1", FileURL: "https://tax.example/inv.pdf",
			Attempt: 1, DurationMs: 35, SubmittedAt: "2026-06-18 11:00:01", CompletedAt: "2026-06-18 11:00:02", Actor: "admin",
		}},
		Receivables: []Receivable{{
			ID: 71, BillNo: "AR-1", CustomerID: 2, StatementID: 60, InvoiceID: 61, Amount: 4800, ReceivedAmount: 1000, DueDate: "2026-07-18", Status: "partial", CreatedAt: "2026-06-18 11:05:00",
		}},
		PaymentPlans: []PaymentPlan{{
			ID: 72, PlanNo: "PP-1", ReceivableID: 71, CustomerID: 2, Amount: 1000, DueDate: "2026-07-01", Method: "bank", Status: "planned", CreatedAt: "2026-06-18 11:20:00",
		}},
		CollectionTasks: []CollectionTask{{
			ID: 73, TaskNo: "COL-1", ReceivableID: 71, CustomerID: 2, CustomerName: "测试客户", Amount: 3800,
			DueDate: "2026-07-18", OverdueDays: 0, Level: "pre_due", Channel: "sms", Status: "open", TemplateID: 74, SendCount: 1,
			LastSentAt: "2026-06-18 11:31:00", GeneratedAt: "2026-06-18 11:30:00",
		}},
		CollectionTemplates: []CollectionTemplate{{
			ID: 74, Code: "pre_due_sms", Name: "到期前短信提醒", Level: "pre_due", Channel: "sms",
			Content: "{{customerName}} 应收 {{amount}} 即将到期", Enabled: true, UpdatedAt: "2026-06-18 11:29:00",
		}},
		CollectionDispatches: []CollectionDispatch{{
			ID: 75, DispatchNo: "CD-1", TaskID: 73, TemplateID: 74, CustomerID: 2, Channel: "sms", Target: "13800010001",
			Content: "测试客户 应收 3800.00 即将到期", Endpoint: "collection://local-simulator/sms", Status: "delivered", SentAt: "2026-06-18 11:31:00", Actor: "admin",
		}},
		Inventory: []InventoryItem{{
			ID: 81, SiteID: 5, Warehouse: "WH-1", Silo: "SILO-1", MaterialID: 6, BatchNo: "MAT-1", RawReceiptID: 80,
			SupplierID: 13, Quantity: 100, Unit: "t", QualityStatus: "passed", AvailableStatus: "available", UpdatedAt: "2026-06-18 08:00:00",
		}},
		InventoryFlows: []InventoryFlow{{
			ID: 91, FlowNo: "IF-1", SiteID: 5, MaterialID: 6, SourceType: "production_batch", SourceID: 21,
			Direction: "out", Quantity: 8, BalanceQty: 92, CreatedAt: "2026-06-18 09:20:00",
		}},
		Locations: []VehicleLocationEvent{{
			ID: 101, VehicleID: 7, PlateNo: "粤B12345", DriverID: 8, DispatchID: 11, DeviceID: "GPS-1", SourceType: "mqtt_gps-json",
			Longitude: 113.9, Latitude: 22.5, Speed: 45, OnlineStatus: "online", IsAbnormal: false, AbnormalType: "", LocationTime: "2026-06-18 09:40:00", ReceiveTime: "2026-06-18 09:40:03",
		}},
		LatestLocations: []VehicleLatestLocation{{
			VehicleID: 7, PlateNo: "粤B12345", Longitude: 113.9, Latitude: 22.5, Speed: 45, Direction: 90, OnlineStatus: "online",
			TransportStatus: "in_transit", LastLocationTime: "2026-06-18 09:40:00", CurrentOrderID: 1, CurrentProjectID: 3, CurrentSiteID: 5, CurrentCustomerID: 2,
		}},
		DeviceProtocolFrames: []DeviceProtocolFrame{{
			ID: 111, FrameNo: "DPF-1", Channel: "mqtt", Protocol: "gps-json", DeviceNo: "GPS-1", Raw: `{"secret":"raw-frame"}`,
			ParsedResource: "vehicle_location", ParsedID: 101, Status: "accepted", Error: "", ReceivedAt: "2026-06-18 09:40:03", Actor: "device:GPS-1",
		}},
	}

	inserts := businessProjectionInserts(data)
	if len(inserts) != len(postgresProjectionTableNames) {
		t.Fatalf("expected one insert per projection table, got %d rows for %d tables", len(inserts), len(postgresProjectionTableNames))
	}
	seen := map[string]int{}
	for _, insert := range inserts {
		seen[insert.table]++
		if len(insert.columns) != len(insert.values) {
			t.Fatalf("columns and values mismatch for %s: %d != %d", insert.table, len(insert.columns), len(insert.values))
		}
		if insert.table == "cbmp_biz_device_protocol_frames" {
			for _, column := range insert.columns {
				if column == "raw" {
					t.Fatalf("device protocol projection must not persist raw frame column")
				}
			}
			for _, value := range insert.values {
				if value == `{"secret":"raw-frame"}` {
					t.Fatalf("device protocol projection leaked raw frame payload")
				}
			}
		}
	}
	for _, table := range postgresProjectionTableNames {
		if seen[table] != 1 {
			t.Fatalf("expected projection table %s once, got %d", table, seen[table])
		}
	}
}

func TestPostgresProjectionSchemaAndInsertQuery(t *testing.T) {
	schema := postgresProjectionSchemaSQL()
	for _, table := range append(append([]string{}, postgresProjectionTableNames...), "cbmp_biz_projection_status") {
		if !strings.Contains(schema, "create table if not exists "+table) {
			t.Fatalf("projection schema missing table %s", table)
		}
	}
	query := projectionInsertQuery("cbmp_biz_orders", []string{"id", "order_no", "status"})
	if query != "insert into cbmp_biz_orders (id, order_no, status) values ($1, $2, $3)" {
		t.Fatalf("unexpected projection insert query: %s", query)
	}
}
