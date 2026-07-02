package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestSalesProductionDeliveryFinanceFlow(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	var order SalesOrder
	postJSON(t, app, token, "/api/orders", `{
		"customerId":1,
		"projectId":1,
		"productId":1,
		"siteId":1,
		"planQuantity":12,
		"unitPrice":520,
		"unit":"t",
		"planTime":"2026-07-01 09:00:00",
		"settlementMode":"月结",
		"transportMode":"自有车队",
		"lines":[{"productId":1,"quantity":12,"unitPrice":520,"unit":"t"}]
	}`, &order)
	if order.ID == 0 || order.Status != "submitted" {
		t.Fatalf("expected submitted sales order, got %+v", order)
	}

	postJSON(t, app, token, "/api/orders/"+strconv.FormatInt(order.ID, 10)+"/approve", `{}`, &order)
	if order.Status != "approved" {
		t.Fatalf("expected approved sales order, got %+v", order)
	}

	var plan ProductionPlan
	postJSON(t, app, token, "/api/production-plans", `{"orderId":`+strconv.FormatInt(order.ID, 10)+`,"plantId":1,"planQuantity":12,"planDate":"2026-07-01","shift":"day"}`, &plan)
	if plan.ID == 0 || plan.OrderID != order.ID || plan.Status != "scheduled" {
		t.Fatalf("expected scheduled production plan linked to order %d, got %+v", order.ID, plan)
	}

	var task ProductionTask
	postJSON(t, app, token, "/api/production-plans/"+strconv.FormatInt(plan.ID, 10)+"/tasks", `{"planQty":12}`, &task)
	if task.ID == 0 || task.PlanID != plan.ID || task.OrderID != order.ID {
		t.Fatalf("expected production task linked to plan %d/order %d, got %+v", plan.ID, order.ID, task)
	}

	var batch ProductionBatch
	postJSON(t, app, token, "/api/production-plans/tasks/"+strconv.FormatInt(task.ID, 10)+"/batches", `{"quantity":12,"qualityStatus":"passed","status":"released"}`, &batch)
	if batch.ID == 0 || batch.TaskID != task.ID || batch.OrderID != order.ID || batch.Quantity != 12 {
		t.Fatalf("expected production batch linked to task %d/order %d, got %+v", task.ID, order.ID, batch)
	}

	var dispatch DispatchOrder
	postJSON(t, app, token, "/api/dispatch-orders", `{"orderId":`+strconv.FormatInt(order.ID, 10)+`,"vehicleId":4,"planQuantity":12}`, &dispatch)
	if dispatch.ID == 0 || dispatch.OrderID != order.ID || dispatch.PlanQuantity != 12 || dispatch.Status != "assigned" {
		t.Fatalf("expected assigned dispatch linked to order %d, got %+v", order.ID, dispatch)
	}

	var note DeliveryNote
	postJSON(t, app, token, "/api/delivery/notes", `{"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`}`, &note)
	if note.ID == 0 || note.DispatchID != dispatch.ID || note.OrderID != order.ID {
		t.Fatalf("expected delivery note linked to dispatch %d/order %d, got %+v", dispatch.ID, order.ID, note)
	}

	var sign DeliverySign
	postJSON(t, app, token, "/api/delivery/sign", `{
		"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`,
		"signer":"王经理",
		"phone":"13800010001",
		"signedQty":12,
		"photo":"data:image/jpeg;base64,flow-sign-photo",
		"signature":"data:image/png;base64,flow-signature"
	}`, &sign)
	if sign.ID == 0 || sign.DispatchID != dispatch.ID || sign.OrderID != order.ID || sign.SignedQty != 12 {
		t.Fatalf("expected delivery sign linked to dispatch %d/order %d, got %+v", dispatch.ID, order.ID, sign)
	}

	statement, ok := statementForSign(app.mustSnapshot(), sign.ID)
	if !ok {
		t.Fatalf("expected delivery sign %d to create a customer statement", sign.ID)
	}
	if statement.Status != "draft" || statement.TotalQty != 12 || statement.TotalAmount != 6240 {
		t.Fatalf("expected draft statement for signed quantity, got %+v", statement)
	}

	postJSON(t, app, token, "/api/statements/"+strconv.FormatInt(statement.ID, 10)+"/confirm", `{}`, &statement)
	if statement.Status != "confirmed" {
		t.Fatalf("expected confirmed statement, got %+v", statement)
	}

	var invoice SalesInvoice
	postJSON(t, app, token, "/api/finance/invoices", `{"statementId":`+strconv.FormatInt(statement.ID, 10)+`,"invoiceCategory":"blue_vat_special"}`, &invoice)
	if invoice.ID == 0 || invoice.StatementID != statement.ID || invoice.Amount != statement.TotalAmount {
		t.Fatalf("expected invoice linked to statement %d, got %+v", statement.ID, invoice)
	}

	receivable, ok := receivableForInvoice(app.mustSnapshot(), invoice.ID)
	if !ok {
		t.Fatalf("expected invoice %d to create receivable", invoice.ID)
	}
	if receivable.Status != "open" || receivable.Amount != invoice.Amount {
		t.Fatalf("expected open receivable for invoice, got %+v", receivable)
	}

	var receipt Receipt
	postJSON(t, app, token, "/api/finance/receipts", `{"receivableId":`+strconv.FormatInt(receivable.ID, 10)+`,"amount":`+strconv.FormatFloat(receivable.Amount, 'f', -1, 64)+`,"method":"bank"}`, &receipt)
	if receipt.ID == 0 || receipt.ReceivableID != receivable.ID || receipt.Amount != receivable.Amount {
		t.Fatalf("expected receipt linked to receivable %d, got %+v", receivable.ID, receipt)
	}

	snapshot := app.mustSnapshot()
	assertOrderStatus(t, snapshot, order.ID, "completed")
	assertDispatchStatus(t, snapshot, dispatch.ID, "completed")
	assertDeliveryNoteStatus(t, snapshot, note.ID, "signed")
	assertStatementStatus(t, snapshot, statement.ID, "invoiced")
	assertReceivablePaid(t, snapshot, receivable.ID, receivable.Amount)
}

func TestPortalCarrierSettlementFlowUsesPersistedData(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	customerToken := testLogin(t, app, "customer", "customer123")
	driverToken := testLogin(t, app, "driver", "driver123")

	var order SalesOrder
	postJSON(t, app, adminToken, "/api/orders", `{
		"customerId":1,
		"projectId":1,
		"productId":1,
		"siteId":1,
		"planQuantity":9,
		"unitPrice":520,
		"unit":"t",
		"planTime":"2026-07-02 09:00:00",
		"settlementMode":"月结",
		"transportMode":"自有车队",
		"lines":[{"productId":1,"quantity":9,"unitPrice":520,"unit":"t"}]
	}`, &order)
	postJSON(t, app, adminToken, "/api/orders/"+strconv.FormatInt(order.ID, 10)+"/approve", `{}`, &order)
	if order.CustomerID != 1 || order.ProjectID != 1 || order.Status != "approved" {
		t.Fatalf("expected approved customer order, got %+v", order)
	}

	var dispatch DispatchOrder
	postJSON(t, app, adminToken, "/api/dispatch-orders", `{"orderId":`+strconv.FormatInt(order.ID, 10)+`,"vehicleId":1,"planQuantity":9}`, &dispatch)
	if dispatch.OrderID != order.ID || dispatch.DriverID != 1 || dispatch.VehicleID != 1 || dispatch.Status != "assigned" {
		t.Fatalf("expected dispatch for driver 1 on order %d, got %+v", order.ID, dispatch)
	}

	var driverPortal PortalOverview
	getJSON(t, app, driverToken, "/api/portal/overview", &driverPortal)
	if !portalHasDispatch(driverPortal.Dispatches, dispatch.ID) {
		t.Fatalf("driver portal did not expose created dispatch %d: %+v", dispatch.ID, driverPortal.Dispatches)
	}
	for _, item := range driverPortal.Dispatches {
		if item.DriverID != 1 {
			t.Fatalf("driver portal leaked dispatch for driver %d: %+v", item.DriverID, driverPortal.Dispatches)
		}
	}

	postJSON(t, app, driverToken, "/api/portal/dispatches/"+strconv.FormatInt(dispatch.ID, 10)+"/exception", `{"exception":"跨角色流程司机上报异常","level":"high"}`, &dispatch)
	if dispatch.Exception != "跨角色流程司机上报异常" {
		t.Fatalf("expected driver exception persisted on dispatch, got %+v", dispatch)
	}
	getJSON(t, app, driverToken, "/api/portal/overview", &driverPortal)
	if !portalHasAlarm(driverPortal.Alarms, dispatch.ID, "driver_exception") {
		t.Fatalf("driver portal did not expose exception alarm for dispatch %d: %+v", dispatch.ID, driverPortal.Alarms)
	}

	var note DeliveryNote
	postJSON(t, app, adminToken, "/api/delivery/notes", `{"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`}`, &note)
	if note.DispatchID != dispatch.ID || note.OrderID != order.ID {
		t.Fatalf("expected delivery note linked to created dispatch/order, got %+v", note)
	}

	var sign DeliverySign
	postJSON(t, app, adminToken, "/api/delivery/sign", `{
		"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`,
		"signer":"王经理",
		"phone":"13800010001",
		"signedQty":9,
		"photo":"data:image/jpeg;base64,portal-flow-sign-photo",
		"signature":"data:image/png;base64,portal-flow-signature"
	}`, &sign)
	if sign.DispatchID != dispatch.ID || sign.OrderID != order.ID || sign.CustomerID != 1 || sign.SignedQty != 9 {
		t.Fatalf("expected signed delivery linked to customer order, got %+v", sign)
	}

	statement, ok := statementForSign(app.mustSnapshot(), sign.ID)
	if !ok {
		t.Fatalf("expected sign %d to create customer statement", sign.ID)
	}
	postJSON(t, app, adminToken, "/api/statements/"+strconv.FormatInt(statement.ID, 10)+"/confirm", `{}`, &statement)
	if statement.Status != "confirmed" {
		t.Fatalf("expected confirmed statement, got %+v", statement)
	}

	var invoice SalesInvoice
	postJSON(t, app, adminToken, "/api/finance/invoices", `{"statementId":`+strconv.FormatInt(statement.ID, 10)+`,"invoiceCategory":"blue_vat_special"}`, &invoice)
	if invoice.StatementID != statement.ID || invoice.CustomerID != 1 || invoice.Amount != statement.TotalAmount {
		t.Fatalf("expected invoice linked to confirmed statement, got %+v", invoice)
	}

	var customerPortal PortalOverview
	getJSON(t, app, customerToken, "/api/portal/overview", &customerPortal)
	if !portalHasOrder(customerPortal.Orders, order.ID) || !portalHasDispatch(customerPortal.Dispatches, dispatch.ID) || !portalHasSign(customerPortal.Signs, sign.ID) || !portalHasStatement(customerPortal.Statements, statement.ID) || !portalHasInvoice(customerPortal.Invoices, invoice.ID) {
		t.Fatalf("customer portal did not expose committed order/delivery/finance chain: %+v", customerPortal)
	}
	for _, item := range customerPortal.Orders {
		if item.CustomerID != 1 {
			t.Fatalf("customer portal leaked order for customer %d: %+v", item.CustomerID, customerPortal.Orders)
		}
	}

	var complaint CustomerComplaint
	postJSON(t, app, customerToken, "/api/portal/complaints", `{"projectId":1,"title":"门户链路投诉","content":"签收后对账需要复核","level":"medium"}`, &complaint)
	if complaint.CustomerID != 1 || complaint.ProjectID != 1 || complaint.Status != "open" {
		t.Fatalf("expected persisted customer-scoped complaint, got %+v", complaint)
	}
	var complaints []CustomerComplaint
	getJSON(t, app, customerToken, "/api/portal/complaints", &complaints)
	if !portalHasComplaint(complaints, complaint.ID) {
		t.Fatalf("customer portal complaint list did not include created complaint %d: %+v", complaint.ID, complaints)
	}

	var settlementResult struct {
		Settlement TransportSettlement       `json:"settlement"`
		Items      []TransportSettlementItem `json:"items"`
	}
	postJSON(t, app, adminToken, "/api/dispatch-orders/carrier-settlements/generate", `{"carrierId":1,"period":"2026-07","ratePerTrip":500}`, &settlementResult)
	item, ok := settlementItemForDispatch(settlementResult.Items, dispatch.ID)
	if !ok || settlementResult.Settlement.CarrierID != 1 || settlementResult.Settlement.TripCount != 1 || item.Quantity != 9 || item.Amount != 500 {
		t.Fatalf("expected carrier settlement item for committed dispatch %d, got %+v", dispatch.ID, settlementResult)
	}

	var settlementBundle struct {
		Settlements []TransportSettlement     `json:"settlements"`
		Items       []TransportSettlementItem `json:"items"`
	}
	getJSON(t, app, adminToken, "/api/dispatch-orders/carrier-settlements", &settlementBundle)
	if item, ok = settlementItemForDispatch(settlementBundle.Items, dispatch.ID); !ok || item.SettlementID != settlementResult.Settlement.ID {
		t.Fatalf("carrier settlement list did not expose generated item for dispatch %d: %+v", dispatch.ID, settlementBundle.Items)
	}

	snapshot := app.mustSnapshot()
	assertOrderStatus(t, snapshot, order.ID, "completed")
	assertDispatchStatus(t, snapshot, dispatch.ID, "completed")
	assertDeliveryNoteStatus(t, snapshot, note.ID, "signed")
	assertStatementStatus(t, snapshot, statement.ID, "invoiced")
}

func TestProcurementInventorySupplierFinanceFlow(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	var receipt RawMaterialReceipt
	postJSON(t, app, token, "/api/procurement/receipts", `{
		"purchaseOrderId":1,
		"siteId":1,
		"grossWeight":42,
		"tareWeight":22,
		"qualityStatus":"pending",
		"status":"received",
		"plateNo":"粤B采购流"
	}`, &receipt)
	if receipt.ID == 0 || receipt.NetWeight != 20 || receipt.TicketID == 0 || receipt.QualityStatus != "pending" {
		t.Fatalf("expected raw material receipt with ticket and pending quality, got %+v", receipt)
	}
	assertReceiptInventory(t, app.mustSnapshot(), receipt.ID, receipt.SiteID, receipt.MaterialID, receipt.NetWeight)

	var inspection RawMaterialInspection
	postJSON(t, app, token, "/api/quality/raw-inspections", `{"receiptId":`+strconv.FormatInt(receipt.ID, 10)+`,"moisture":3.1,"mudContent":1.2,"fineness":"II区","remark":"采购流入厂抽检"}`, &inspection)
	if inspection.ID == 0 || inspection.ReceiptID != receipt.ID || inspection.Status != "pending_review" {
		t.Fatalf("expected pending raw inspection linked to receipt %d, got %+v", receipt.ID, inspection)
	}

	postJSON(t, app, token, "/api/quality/raw-inspections/"+strconv.FormatInt(inspection.ID, 10)+"/review", `{"result":"passed","remark":"指标合格"}`, &inspection)
	if inspection.Status != "completed" || inspection.Result != "passed" {
		t.Fatalf("expected completed raw inspection, got %+v", inspection)
	}
	assertReceiptQuality(t, app.mustSnapshot(), receipt.ID, "passed")

	var transfer InventoryTransfer
	postJSON(t, app, token, "/api/procurement/transfers", `{"fromSiteId":1,"toSiteId":2,"materialId":3,"quantity":5,"remark":"采购流跨站调拨"}`, &transfer)
	if transfer.ID == 0 || transfer.Status != "pending_approval" {
		t.Fatalf("expected transfer pending approval, got %+v", transfer)
	}
	approveGenericApprovalTask(t, app, token, "inventory_transfer", transfer.ID, "调拨审批")

	postJSON(t, app, token, "/api/procurement/transfers/"+strconv.FormatInt(transfer.ID, 10)+"/complete", `{}`, &transfer)
	if transfer.Status != "completed" {
		t.Fatalf("expected completed transfer, got %+v", transfer)
	}
	assertInventoryBalance(t, app.mustSnapshot(), 2, 3, 5)

	var stocktake InventoryStocktake
	postJSON(t, app, token, "/api/procurement/stocktakes", `{"siteId":2,"materialId":3,"actualQty":4.5,"remark":"采购流到站盘点"}`, &stocktake)
	if stocktake.ID == 0 || stocktake.BookQty != 5 || stocktake.DiffQty != -0.5 || stocktake.Status != "pending_review" {
		t.Fatalf("expected stocktake against transferred inventory, got %+v", stocktake)
	}
	postJSON(t, app, token, "/api/procurement/stocktakes/"+strconv.FormatInt(stocktake.ID, 10)+"/review", `{}`, &stocktake)
	if stocktake.Status != "completed" {
		t.Fatalf("expected completed stocktake, got %+v", stocktake)
	}
	assertInventoryBalance(t, app.mustSnapshot(), 2, 3, 4.5)

	var statement SupplierStatement
	postJSON(t, app, token, "/api/finance/supplier-statements", `{"supplierId":1,"period":"2026-07"}`, &statement)
	if statement.ID == 0 || statement.Amount <= 0 || statement.Status != "submitted" {
		t.Fatalf("expected supplier statement, got %+v", statement)
	}
	postJSON(t, app, token, "/api/finance/supplier-statements/"+strconv.FormatInt(statement.ID, 10)+"/approve", `{}`, &statement)
	if statement.Status != "approved" || statement.ApprovedBy == "" {
		t.Fatalf("expected approved supplier statement, got %+v", statement)
	}

	payables := payablesForSupplierStatement(app.mustSnapshot(), statement.ID)
	if len(payables) == 0 {
		t.Fatalf("expected payables linked to supplier statement %d", statement.ID)
	}
	for _, payable := range payables {
		var payment Payment
		postJSON(t, app, token, "/api/finance/payments", `{"payableId":`+strconv.FormatInt(payable.ID, 10)+`,"amount":`+strconv.FormatFloat(payable.Amount-payable.PaidAmount, 'f', -1, 64)+`,"method":"bank"}`, &payment)
		if payment.ID == 0 || payment.PayableID != payable.ID {
			t.Fatalf("expected payment linked to payable %d, got %+v", payable.ID, payment)
		}
	}
	assertSupplierStatementPaid(t, app.mustSnapshot(), statement.ID)
}

func postJSON(t *testing.T, app *App, token, path, body string, out interface{}) {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodPost, path, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST %s status %d: %s", path, rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("decode POST %s response: %v", path, err)
	}
}

func getJSON(t *testing.T, app *App, token, path string, out interface{}) {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, path, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status %d: %s", path, rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("decode GET %s response: %v", path, err)
	}
}

func approveGenericApprovalTask(t *testing.T, app *App, token, resource string, resourceID int64, comment string) {
	t.Helper()
	task := findApprovalTaskForResource(fetchApprovalTasks(t, app, token), resource, resourceID)
	if task.ID == 0 || task.Status != "pending" {
		t.Fatalf("expected pending approval task for %s %d, got %+v", resource, resourceID, task)
	}
	path := "/api/approvals/" + strconv.FormatInt(task.ID, 10) + "/act"
	var approved ApprovalTask
	postJSON(t, app, token, path, `{"action":"approve","comment":"`+comment+`一审"}`, &approved)
	postJSON(t, app, token, path, `{"action":"approve","comment":"`+comment+`终审"}`, &approved)
	if approved.Status != "approved" {
		t.Fatalf("expected approved task for %s %d, got %+v", resource, resourceID, approved)
	}
}

func statementForSign(data AppData, signID int64) (Statement, bool) {
	for _, statement := range data.Statements {
		for _, item := range statement.Items {
			if item.SignID == signID {
				return statement, true
			}
		}
	}
	return Statement{}, false
}

func payablesForSupplierStatement(data AppData, statementID int64) []Payable {
	items := []Payable{}
	for _, item := range data.Payables {
		if item.SupplierStatementID == statementID {
			items = append(items, item)
		}
	}
	return items
}

func portalHasOrder(items []SalesOrder, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func portalHasDispatch(items []DispatchOrder, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func portalHasSign(items []DeliverySign, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func portalHasStatement(items []Statement, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func portalHasInvoice(items []SalesInvoice, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func portalHasAlarm(items []VehicleAlarm, dispatchID int64, alarmType string) bool {
	for _, item := range items {
		if item.DispatchID == dispatchID && item.AlarmType == alarmType {
			return true
		}
	}
	return false
}

func portalHasComplaint(items []CustomerComplaint, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func settlementItemForDispatch(items []TransportSettlementItem, dispatchID int64) (TransportSettlementItem, bool) {
	for _, item := range items {
		if item.DispatchID == dispatchID {
			return item, true
		}
	}
	return TransportSettlementItem{}, false
}

func assertReceiptInventory(t *testing.T, data AppData, receiptID, siteID, materialID int64, quantity float64) {
	t.Helper()
	for _, item := range data.Inventory {
		if item.SiteID == siteID && item.MaterialID == materialID && item.RawReceiptID == receiptID && item.Quantity == quantity {
			return
		}
	}
	t.Fatalf("inventory lot for receipt %d site %d material %d quantity %.2f not found: %+v", receiptID, siteID, materialID, quantity, data.Inventory)
}

func assertReceiptQuality(t *testing.T, data AppData, receiptID int64, qualityStatus string) {
	t.Helper()
	for _, item := range data.RawMaterialReceipts {
		if item.ID == receiptID {
			if item.QualityStatus != qualityStatus {
				t.Fatalf("receipt %d quality = %s, want %s", receiptID, item.QualityStatus, qualityStatus)
			}
			return
		}
	}
	t.Fatalf("receipt %d not found", receiptID)
}

func assertInventoryBalance(t *testing.T, data AppData, siteID, materialID int64, quantity float64) {
	t.Helper()
	total := 0.0
	for _, item := range data.Inventory {
		if item.SiteID == siteID && item.MaterialID == materialID {
			total = round(total + item.Quantity)
		}
	}
	if total != quantity {
		t.Fatalf("inventory balance site %d material %d = %.2f, want %.2f", siteID, materialID, total, quantity)
	}
}

func assertSupplierStatementPaid(t *testing.T, data AppData, statementID int64) {
	t.Helper()
	for _, item := range data.Payables {
		if item.SupplierStatementID == statementID && (item.Status != "paid" || item.PaidAmount != item.Amount) {
			t.Fatalf("payable for supplier statement %d not paid: %+v", statementID, item)
		}
	}
}

func receivableForInvoice(data AppData, invoiceID int64) (Receivable, bool) {
	for _, item := range data.Receivables {
		if item.InvoiceID == invoiceID {
			return item, true
		}
	}
	return Receivable{}, false
}

func assertOrderStatus(t *testing.T, data AppData, id int64, status string) {
	t.Helper()
	for _, item := range data.Orders {
		if item.ID == id {
			if item.Status != status {
				t.Fatalf("order %d status = %s, want %s", id, item.Status, status)
			}
			return
		}
	}
	t.Fatalf("order %d not found", id)
}

func assertDispatchStatus(t *testing.T, data AppData, id int64, status string) {
	t.Helper()
	for _, item := range data.DispatchOrders {
		if item.ID == id {
			if item.Status != status {
				t.Fatalf("dispatch %d status = %s, want %s", id, item.Status, status)
			}
			return
		}
	}
	t.Fatalf("dispatch %d not found", id)
}

func assertDeliveryNoteStatus(t *testing.T, data AppData, id int64, status string) {
	t.Helper()
	for _, item := range data.DeliveryNotes {
		if item.ID == id {
			if item.Status != status {
				t.Fatalf("delivery note %d status = %s, want %s", id, item.Status, status)
			}
			return
		}
	}
	t.Fatalf("delivery note %d not found", id)
}

func assertStatementStatus(t *testing.T, data AppData, id int64, status string) {
	t.Helper()
	for _, item := range data.Statements {
		if item.ID == id {
			if item.Status != status {
				t.Fatalf("statement %d status = %s, want %s", id, item.Status, status)
			}
			return
		}
	}
	t.Fatalf("statement %d not found", id)
}

func assertReceivablePaid(t *testing.T, data AppData, id int64, amount float64) {
	t.Helper()
	for _, item := range data.Receivables {
		if item.ID == id {
			if item.Status != "paid" || item.ReceivedAmount != amount {
				t.Fatalf("receivable %d = %+v, want paid amount %.2f", id, item, amount)
			}
			return
		}
	}
	t.Fatalf("receivable %d not found", id)
}
