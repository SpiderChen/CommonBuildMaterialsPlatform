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
	if !hasSiloQty(procurement.Silos, "SAND-01", 830) {
		t.Fatalf("expected source silo to follow transfer balance, got %+v", procurement.Silos)
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
	if !hasStocktake(procurement.Stocktakes, stocktake.ID, "completed") || !hasInventoryQty(procurement.Inventory, 1, 3, 825) || !hasSiloQty(procurement.Silos, "SAND-01", 825) {
		t.Fatalf("expected reviewed stocktake and adjusted inventory, got stocktakes=%+v inventory=%+v silos=%+v", procurement.Stocktakes, procurement.Inventory, procurement.Silos)
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

func fetchProcurementOverview(t *testing.T, app *App, token string) struct {
	Inventory  []InventoryItem      `json:"inventory"`
	Flows      []InventoryFlow      `json:"flows"`
	Transfers  []InventoryTransfer  `json:"transfers"`
	Stocktakes []InventoryStocktake `json:"stocktakes"`
	Silos      []Silo               `json:"silos"`
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
		Silos      []Silo               `json:"silos"`
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
	for _, item := range items {
		if item.SupplierStatementID == statementID {
			return item.Status == "confirmed"
		}
	}
	return false
}
