package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestStockYardReceiptTransferAdjustmentAndProtocolIngest(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/stock-yards", `{"siteId":1,"code":"NS-YARD-TEST","name":"测试骨料堆场","type":"aggregate_yard","area":"测试区","capacity":500,"unit":"t","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stock yard status %d: %s", rec.Code, rec.Body.String())
	}
	var yard StockYard
	if err := json.Unmarshal(rec.Body.Bytes(), &yard); err != nil {
		t.Fatalf("decode stock yard: %v", err)
	}
	if yard.ID == 0 || yard.Code != "NS-YARD-TEST" || yard.SiteID != 1 {
		t.Fatalf("unexpected stock yard: %+v", yard)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/stock-yards", `{"siteId":1,"code":"NS-YARD-DEL","name":"可删除堆场","type":"aggregate_yard","area":"测试区","capacity":500,"unit":"t","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create removable stock yard status %d: %s", rec.Code, rec.Body.String())
	}
	var removableYard StockYard
	if err := json.Unmarshal(rec.Body.Bytes(), &removableYard); err != nil {
		t.Fatalf("decode removable stock yard: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/stock-yards/"+strconv.FormatInt(removableYard.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete removable stock yard status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/stock-yard-piles", `{"siteId":1,"yardId":`+strconv.FormatInt(yard.ID, 10)+`,"code":"NS-YARD-PILE-T","name":"测试机制砂堆位","materialId":3,"supplierId":1,"capacity":200,"unit":"t","warningQty":20,"qualityStatus":"passed","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stock yard pile status %d: %s", rec.Code, rec.Body.String())
	}
	var pile StockYardPile
	if err := json.Unmarshal(rec.Body.Bytes(), &pile); err != nil {
		t.Fatalf("decode stock yard pile: %v", err)
	}
	if pile.ID == 0 || pile.YardCode != "NS-YARD-TEST" || pile.CurrentQty != 0 {
		t.Fatalf("unexpected stock yard pile: %+v", pile)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/stock-yards/"+strconv.FormatInt(yard.ID, 10), "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("delete stock yard with pile should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/master/stock-yard-piles", `{"siteId":1,"yardId":`+strconv.FormatInt(yard.ID, 10)+`,"code":"NS-YARD-PILE-DEL","name":"可删除堆位","materialId":3,"supplierId":1,"capacity":200,"unit":"t","warningQty":20,"qualityStatus":"passed","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create removable stock yard pile status %d: %s", rec.Code, rec.Body.String())
	}
	var removablePile StockYardPile
	if err := json.Unmarshal(rec.Body.Bytes(), &removablePile); err != nil {
		t.Fatalf("decode removable stock yard pile: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/stock-yard-piles/"+strconv.FormatInt(removablePile.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete removable stock yard pile status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/yard-receipts", `{"pileId":`+strconv.FormatInt(pile.ID, 10)+`,"materialId":3,"supplierId":1,"batchNo":"SAND-TEST-001","quantity":50,"unit":"t","moistureRate":4.4,"qualityStatus":"passed","remark":"测试入场"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("stock yard receipt status %d: %s", rec.Code, rec.Body.String())
	}
	var yardFlow StockYardFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &yardFlow); err != nil {
		t.Fatalf("decode stock yard receipt flow: %v", err)
	}
	if yardFlow.Direction != "in" || yardFlow.Quantity != 50 || yardFlow.BalanceQty != 50 {
		t.Fatalf("unexpected stock yard receipt flow: %+v", yardFlow)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/stock-yard-piles/"+strconv.FormatInt(pile.ID, 10), "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("delete stock yard pile with flow should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/buffer-transfers", `{"bufferId":1,"yardPileId":`+strconv.FormatInt(pile.ID, 10)+`,"materialId":3,"quantity":10,"unit":"t","remark":"堆位上料测试"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("buffer transfer from stock yard status %d: %s", rec.Code, rec.Body.String())
	}
	var bufferFlow PlantBufferFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &bufferFlow); err != nil {
		t.Fatalf("decode buffer transfer flow: %v", err)
	}
	if bufferFlow.SourceType != "stock_yard_pile" || bufferFlow.Direction != "in" || bufferFlow.Quantity != 10 {
		t.Fatalf("unexpected buffer transfer flow: %+v", bufferFlow)
	}
	if stockYardPileQty(app.mustSnapshot().StockYardPiles, "NS-YARD-PILE-T") != 40 {
		t.Fatalf("expected pile qty 40, got %+v", app.mustSnapshot().StockYardPiles)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/yard-adjustments", `{"pileCode":"NS-YARD-PILE-T","actualQty":38,"moistureRate":4.8,"qualityStatus":"passed","status":"active","remark":"测试盘点"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("stock yard adjustment status %d: %s", rec.Code, rec.Body.String())
	}
	if stockYardPileQty(app.mustSnapshot().StockYardPiles, "NS-YARD-PILE-T") != 38 {
		t.Fatalf("expected adjusted pile qty 38, got %+v", app.mustSnapshot().StockYardPiles)
	}

	payload := `{"channel":"industrial-control-gateway","protocol":"yard-json","payload":{"deviceNo":"YARD-NS-AGG","yardCode":"NS-YARD-TEST","pileCode":"NS-YARD-PILE-T","materialId":3,"quantity":41.5,"moistureRate":4.9,"qualityStatus":"passed","status":"active","reportedAt":"2026-06-20 13:00:00"}}`
	rec = testDeviceRequest(t, app, "plant-demo-key-1", http.MethodPost, "/api/production-plans/protocols/yard/ingest", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("stock yard protocol ingest status %d: %s", rec.Code, rec.Body.String())
	}
	var response protocolIngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode stock yard protocol response: %v", err)
	}
	if response.Frame.Status != "accepted" || response.Frame.ParsedResource != "stock_yard_pile" || response.StockYardFlow == nil || response.StockYardFlow.BalanceQty != 41.5 {
		t.Fatalf("unexpected stock yard protocol response: %+v", response)
	}
}

func TestStockYardAdjustmentWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"stock_yard_adjustment_review","name":"堆位盘点复核","category":"approval","resource":"stock_yard_adjustment","trigger":{"eventType":"stock_yard_adjustment.requested","resource":"stock_yard_adjustment"},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"盘点复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stock yard adjustment workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/stock-yards", `{"siteId":1,"code":"NS-YARD-WF","name":"流程骨料堆场","type":"aggregate_yard","area":"流程区","capacity":500,"unit":"t","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stock yard status %d: %s", rec.Code, rec.Body.String())
	}
	var yard StockYard
	if err := json.Unmarshal(rec.Body.Bytes(), &yard); err != nil {
		t.Fatalf("decode stock yard: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/stock-yard-piles", `{"siteId":1,"yardId":`+strconv.FormatInt(yard.ID, 10)+`,"code":"NS-YARD-PILE-WF","name":"流程机制砂堆位","materialId":3,"supplierId":1,"capacity":200,"unit":"t","warningQty":20,"qualityStatus":"passed","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create stock yard pile status %d: %s", rec.Code, rec.Body.String())
	}
	var pile StockYardPile
	if err := json.Unmarshal(rec.Body.Bytes(), &pile); err != nil {
		t.Fatalf("decode stock yard pile: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/yard-receipts", `{"pileId":`+strconv.FormatInt(pile.ID, 10)+`,"materialId":3,"supplierId":1,"batchNo":"SAND-WF-001","quantity":50,"unit":"t","moistureRate":4.4,"qualityStatus":"passed","remark":"流程入场"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("stock yard receipt status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/procurement/yard-adjustments", `{"pileId":`+strconv.FormatInt(pile.ID, 10)+`,"actualQty":38,"moistureRate":4.8,"qualityStatus":"passed","status":"active","remark":"流程盘点"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("stock yard adjustment workflow request status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode workflow instance: %v", err)
	}
	if instance.Resource != "stock_yard_adjustment" || instance.ResourceID != pile.ID || instance.Status != "pending" {
		t.Fatalf("unexpected workflow instance: %+v", instance)
	}
	snapshot := app.mustSnapshot()
	if stockYardPileQty(snapshot.StockYardPiles, "NS-YARD-PILE-WF") != 50 {
		t.Fatalf("expected pile qty unchanged before approval, got %+v", snapshot.StockYardPiles)
	}
	if len(snapshot.WorkflowTasks) == 0 || snapshot.WorkflowEvents[len(snapshot.WorkflowEvents)-1].Status != "handled" {
		t.Fatalf("expected handled yard adjustment workflow, got events %+v tasks %+v", snapshot.WorkflowEvents, snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[0].ID, 10)+"/act", `{"action":"approve","comment":"盘点通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve yard adjustment workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if stockYardPileQty(snapshot.StockYardPiles, "NS-YARD-PILE-WF") != 38 {
		t.Fatalf("expected approved pile adjustment qty 38, got %+v", snapshot.StockYardPiles)
	}
	if !hasStockYardFlow(snapshot.StockYardFlows, "NS-YARD-PILE-WF", "yard_adjustment", "adjustment_out") {
		t.Fatalf("expected workflow stock yard adjustment flow, got %+v", snapshot.StockYardFlows)
	}
}

func stockYardPileQty(items []StockYardPile, code string) float64 {
	for _, item := range items {
		if item.Code == code {
			return item.CurrentQty
		}
	}
	return 0
}

func hasStockYardFlow(items []StockYardFlow, pileCode string, sourceType string, direction string) bool {
	for _, item := range items {
		if item.PileCode == pileCode && item.SourceType == sourceType && item.Direction == direction {
			return true
		}
	}
	return false
}
