package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestInventoryTransferStocktakeAndSupplierStatementApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/procurement/transfers", `{"fromSiteId":1,"toSiteId":2,"materialId":3,"quantity":10,"remark":"宝安站应急补料"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create transfer status %d: %s", rec.Code, rec.Body.String())
	}
	var transfer InventoryTransfer
	if err := json.Unmarshal(rec.Body.Bytes(), &transfer); err != nil {
		t.Fatalf("decode transfer: %v", err)
	}
	if transfer.Status != "pending_approval" || transfer.TransferNo == "" {
		t.Fatalf("unexpected transfer: %+v", transfer)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/transfers/"+strconv.FormatInt(transfer.ID, 10)+"/complete", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected transfer completion to wait for approval, got %d: %s", rec.Code, rec.Body.String())
	}

	tasks := fetchApprovalTasks(t, app, token)
	task := findApprovalTaskForResource(tasks, "inventory_transfer", transfer.ID)
	if task.ID == 0 || task.CurrentRole != "dispatcher" || task.Status != "pending" {
		t.Fatalf("expected pending transfer approval task, got %+v from %+v", task, tasks)
	}
	taskID := strconv.FormatInt(task.ID, 10)
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+taskID+"/act", `{"action":"approve","comment":"调度确认调拨"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first transfer approval status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+taskID+"/act", `{"action":"approve","comment":"高管确认调拨"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("final transfer approval status %d: %s", rec.Code, rec.Body.String())
	}
	var approvedTask ApprovalTask
	if err := json.Unmarshal(rec.Body.Bytes(), &approvedTask); err != nil {
		t.Fatalf("decode transfer approval: %v", err)
	}
	if approvedTask.Status != "approved" || len(approvedTask.Actions) != 2 {
		t.Fatalf("expected approved transfer task, got %+v", approvedTask)
	}

	procurement := fetchProcurementOverview(t, app, token)
	if !hasInventoryTransfer(procurement.Transfers, transfer.ID, "approved") {
		t.Fatalf("expected approved transfer before completion, got %+v", procurement.Transfers)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/transfers/"+strconv.FormatInt(transfer.ID, 10)+"/complete", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("complete transfer status %d: %s", rec.Code, rec.Body.String())
	}

	procurement = fetchProcurementOverview(t, app, token)
	if !hasInventoryTransfer(procurement.Transfers, transfer.ID, "completed") {
		t.Fatalf("expected completed transfer, got %+v", procurement.Transfers)
	}
	if !hasInventoryQty(procurement.Inventory, 1, 3, 830) || !hasInventoryQty(procurement.Inventory, 2, 3, 10) {
		t.Fatalf("expected transfer inventory balances, got %+v", procurement.Inventory)
	}
	if !hasInventoryLot(procurement.Inventory, 2, 3, "SAND-20260617-A", 1, 10) {
		t.Fatalf("expected transferred lot to keep batch and receipt source, got %+v", procurement.Inventory)
	}
	if !hasSiloQty(app.mustSnapshot().Silos, "SAND-01", 830) {
		t.Fatalf("expected source silo to follow transfer balance, got %+v", app.mustSnapshot().Silos)
	}
	if countInventoryFlows(procurement.Flows, "inventory_transfer", transfer.ID) != 2 {
		t.Fatalf("expected two transfer flows, got %+v", procurement.Flows)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/stocktakes", `{"siteId":1,"materialId":3,"actualQty":825,"remark":"月末盘点"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stocktake status %d: %s", rec.Code, rec.Body.String())
	}
	var stocktake InventoryStocktake
	if err := json.Unmarshal(rec.Body.Bytes(), &stocktake); err != nil {
		t.Fatalf("decode stocktake: %v", err)
	}
	if stocktake.BookQty != 830 || stocktake.DiffQty != -5 || stocktake.Status != "pending_review" {
		t.Fatalf("unexpected stocktake: %+v", stocktake)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/stocktakes/"+strconv.FormatInt(stocktake.ID, 10)+"/review", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("review stocktake status %d: %s", rec.Code, rec.Body.String())
	}
	procurement = fetchProcurementOverview(t, app, token)
	if !hasStocktake(procurement.Stocktakes, stocktake.ID, "completed") || !hasInventoryQty(procurement.Inventory, 1, 3, 825) || !hasSiloQty(app.mustSnapshot().Silos, "SAND-01", 825) {
		t.Fatalf("expected reviewed stocktake and adjusted inventory, got stocktakes=%+v inventory=%+v silos=%+v", procurement.Stocktakes, procurement.Inventory, app.mustSnapshot().Silos)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/finance/supplier-statements", `{"supplierId":1,"period":"2026-06"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create supplier statement status %d: %s", rec.Code, rec.Body.String())
	}
	var statement SupplierStatement
	if err := json.Unmarshal(rec.Body.Bytes(), &statement); err != nil {
		t.Fatalf("decode supplier statement: %v", err)
	}
	if statement.Amount <= 0 || statement.Status != "submitted" {
		t.Fatalf("unexpected supplier statement: %+v", statement)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/finance/supplier-statements/"+strconv.FormatInt(statement.ID, 10)+"/approve", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve supplier statement status %d: %s", rec.Code, rec.Body.String())
	}
	var approved SupplierStatement
	if err := json.Unmarshal(rec.Body.Bytes(), &approved); err != nil {
		t.Fatalf("decode approved supplier statement: %v", err)
	}
	if approved.Status != "approved" || approved.ApprovedBy == "" || approved.ApprovedAt == "" {
		t.Fatalf("expected approved supplier statement, got %+v", approved)
	}
	finance := fetchFinanceOverview(t, app, token)
	if !hasConfirmedPayableForStatement(finance.Payables, approved.ID) {
		t.Fatalf("expected confirmed payable for statement, got %+v", finance.Payables)
	}
}

func TestInventoryStocktakeWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"inventory_stocktake_review","name":"库存盘点复核","category":"approval","resource":"inventory_stocktake","trigger":{"eventType":"inventory_stocktake.review_requested","resource":"inventory_stocktake","conditions":[{"field":"materialId","operator":"equals","value":"3"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"盘点复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stocktake workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/stocktakes", `{"siteId":1,"materialId":3,"actualQty":835,"remark":"流程盘点"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow stocktake status %d: %s", rec.Code, rec.Body.String())
	}
	var stocktake InventoryStocktake
	if err := json.Unmarshal(rec.Body.Bytes(), &stocktake); err != nil {
		t.Fatalf("decode workflow stocktake: %v", err)
	}
	if stocktake.Status != "pending_review" || stocktake.DiffQty == 0 {
		t.Fatalf("unexpected workflow stocktake seed state: %+v", stocktake)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/stocktakes/"+strconv.FormatInt(stocktake.ID, 10)+"/review", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request stocktake workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending InventoryStocktake
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending stocktake: %v", err)
	}
	if pending.Status != "pending_approval" || pending.ReviewedAt != "" {
		t.Fatalf("expected pending stocktake before workflow approval, got %+v", pending)
	}
	procurement := fetchProcurementOverview(t, app, token)
	if !hasInventoryQty(procurement.Inventory, 1, 3, stocktake.BookQty) || !hasSiloQty(app.mustSnapshot().Silos, "SAND-01", stocktake.BookQty) {
		t.Fatalf("inventory should not change before stocktake approval, got inventory=%+v silos=%+v", procurement.Inventory, app.mustSnapshot().Silos)
	}
	if countInventoryFlows(procurement.Flows, "inventory_stocktake", stocktake.ID) != 0 {
		t.Fatalf("stocktake flow should not be created before approval, got %+v", procurement.Flows)
	}

	snapshot := app.mustSnapshot()
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "inventory_stocktake" && task.ResourceID == stocktake.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending stocktake workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"盘点复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve stocktake workflow status %d: %s", rec.Code, rec.Body.String())
	}
	procurement = fetchProcurementOverview(t, app, token)
	if !hasStocktake(procurement.Stocktakes, stocktake.ID, "completed") || !hasInventoryQty(procurement.Inventory, 1, 3, stocktake.ActualQty) || !hasSiloQty(app.mustSnapshot().Silos, "SAND-01", stocktake.ActualQty) {
		t.Fatalf("expected approved stocktake to adjust inventory, got stocktakes=%+v inventory=%+v silos=%+v", procurement.Stocktakes, procurement.Inventory, app.mustSnapshot().Silos)
	}
	if countInventoryFlows(procurement.Flows, "inventory_stocktake", stocktake.ID) != 1 {
		t.Fatalf("expected stocktake adjustment flow after approval, got %+v", procurement.Flows)
	}
}

func TestSupplierStatementWorkflowApprovalConfirmsPayable(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"supplier_statement_finance","name":"供应商对账审批","category":"approval","resource":"supplier_statement","trigger":{"eventType":"supplier_statement.submitted","resource":"supplier_statement","conditions":[{"field":"amount","operator":"greater_than","value":"0"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"财务确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create supplier statement workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/finance/supplier-statements", `{"supplierId":1,"period":"2026-07"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create supplier statement status %d: %s", rec.Code, rec.Body.String())
	}
	var statement SupplierStatement
	if err := json.Unmarshal(rec.Body.Bytes(), &statement); err != nil {
		t.Fatalf("decode supplier statement: %v", err)
	}
	if statement.Status != "pending_approval" {
		t.Fatalf("expected supplier statement pending workflow approval, got %+v", statement)
	}
	finance := fetchFinanceOverview(t, app, token)
	if !hasPayableForStatementStatus(finance.Payables, statement.ID, "open") {
		t.Fatalf("expected payable pre-linked but not confirmed before workflow approval, got %+v", finance.Payables)
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowEvents) != 1 || snapshot.WorkflowEvents[0].EventType != "supplier_statement.submitted" || snapshot.WorkflowEvents[0].Status != "handled" {
		t.Fatalf("expected handled supplier statement workflow event, got %+v", snapshot.WorkflowEvents)
	}
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].Resource != "supplier_statement" || snapshot.WorkflowTasks[0].Status != "pending" {
		t.Fatalf("expected pending supplier statement workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/finance/supplier-statements/"+strconv.FormatInt(statement.ID, 10)+"/approve", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("direct approve should not bypass pending workflow, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[0].ID, 10)+"/act", `{"action":"approve","comment":"财务确认"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("act supplier statement workflow status %d: %s", rec.Code, rec.Body.String())
	}
	finance = fetchFinanceOverview(t, app, token)
	if !hasPayableForStatementStatus(finance.Payables, statement.ID, "confirmed") {
		t.Fatalf("expected confirmed payable after workflow approval, got %+v", finance.Payables)
	}
	var approved SupplierStatement
	for _, item := range app.mustSnapshot().SupplierStatements {
		if item.ID == statement.ID {
			approved = item
			break
		}
	}
	if approved.Status != "approved" || approved.ApprovedBy == "" || approved.ApprovedAt == "" {
		t.Fatalf("expected approved supplier statement after workflow result, got %+v", approved)
	}
}

func fetchProcurementOverview(t *testing.T, app *App, token string) struct {
	Inventory  []InventoryItem      `json:"inventory"`
	Flows      []InventoryFlow      `json:"flows"`
	Transfers  []InventoryTransfer  `json:"transfers"`
	Stocktakes []InventoryStocktake `json:"stocktakes"`
} {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/procurement/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("procurement overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Inventory  []InventoryItem      `json:"inventory"`
		Flows      []InventoryFlow      `json:"flows"`
		Transfers  []InventoryTransfer  `json:"transfers"`
		Stocktakes []InventoryStocktake `json:"stocktakes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode procurement overview: %v", err)
	}
	return overview
}

func fetchFinanceOverview(t *testing.T, app *App, token string) struct {
	Payables []Payable `json:"payables"`
} {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Payables []Payable `json:"payables"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode finance overview: %v", err)
	}
	return overview
}

func hasInventoryTransfer(items []InventoryTransfer, id int64, status string) bool {
	for _, item := range items {
		if item.ID == id {
			if item.Status != status {
				return false
			}
			if status == "completed" {
				return item.CompletedAt != ""
			}
			return true
		}
	}
	return false
}

func findApprovalTaskForResource(items []ApprovalTask, resource string, resourceID int64) ApprovalTask {
	for _, item := range items {
		if item.Resource == resource && item.ResourceID == resourceID {
			return item
		}
	}
	return ApprovalTask{}
}

func hasInventoryQty(items []InventoryItem, siteID, materialID int64, qty float64) bool {
	for _, item := range items {
		if item.SiteID == siteID && item.MaterialID == materialID {
			return inventoryBalance(AppData{Inventory: items}, siteID, materialID) == qty
		}
	}
	return false
}

func hasInventoryLot(items []InventoryItem, siteID, materialID int64, batchNo string, rawReceiptID int64, qty float64) bool {
	for _, item := range items {
		if item.SiteID == siteID && item.MaterialID == materialID && item.BatchNo == batchNo && item.RawReceiptID == rawReceiptID && item.Quantity == qty {
			return true
		}
	}
	return false
}

func hasSiloQty(items []Silo, code string, qty float64) bool {
	for _, item := range items {
		if item.Code == code {
			return item.CurrentQty == qty
		}
	}
	return false
}

func countInventoryFlows(items []InventoryFlow, sourceType string, sourceID int64) int {
	count := 0
	for _, item := range items {
		if item.SourceType == sourceType && item.SourceID == sourceID {
			count++
		}
	}
	return count
}

func hasStocktake(items []InventoryStocktake, id int64, status string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == status && item.ReviewedAt != ""
		}
	}
	return false
}

func hasConfirmedPayableForStatement(items []Payable, statementID int64) bool {
	return hasPayableForStatementStatus(items, statementID, "confirmed")
}

func hasPayableForStatementStatus(items []Payable, statementID int64, status string) bool {
	for _, item := range items {
		if item.SupplierStatementID == statementID && item.Status == status {
			return true
		}
	}
	return false
}
