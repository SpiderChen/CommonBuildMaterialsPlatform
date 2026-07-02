package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestPlantBufferTransferAdjustmentAndProtocolIngest(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/plant-buffer-locations", `{"siteId":1,"plantId":1,"code":"NS-AMP240-AGG-T","name":"测试骨料仓","type":"aggregate_bin","materialId":3,"allowedMaterialIds":[3],"capacity":100,"unit":"t","warningQty":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant buffer status %d: %s", rec.Code, rec.Body.String())
	}
	var buffer PlantBufferLocation
	if err := json.Unmarshal(rec.Body.Bytes(), &buffer); err != nil {
		t.Fatalf("decode buffer: %v", err)
	}
	if buffer.ID == 0 || buffer.PlantCode != "NS-AMP240" || buffer.CurrentQty != 0 {
		t.Fatalf("unexpected buffer: %+v", buffer)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/plant-buffer-locations", `{"siteId":1,"plantId":1,"code":"ns-amp240-agg-t","name":"重复编码骨料仓","type":"aggregate_bin","materialId":3,"capacity":100,"unit":"t","warningQty":10}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate buffer code to fail, got status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPut, "/api/master/plant-buffer-locations/"+strconv.FormatInt(buffer.ID, 10), `{"siteId":1,"plantId":1,"code":"NS-AMP240-AGG-01","name":"撞码骨料仓","type":"aggregate_bin","materialId":3,"capacity":100,"unit":"t","warningQty":10}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate buffer code update to fail, got status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/plant-buffer-locations", `{"siteId":1,"plantId":1,"code":"NS-AMP240-AGG-DEL","name":"可删除骨料仓","type":"aggregate_bin","materialId":3,"capacity":100,"unit":"t","warningQty":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create removable plant buffer status %d: %s", rec.Code, rec.Body.String())
	}
	var removable PlantBufferLocation
	if err := json.Unmarshal(rec.Body.Bytes(), &removable); err != nil {
		t.Fatalf("decode removable buffer: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/plant-buffer-locations/"+strconv.FormatInt(removable.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete removable plant buffer status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/buffer-transfers", `{"bufferId":`+strconv.FormatInt(buffer.ID, 10)+`,"materialId":3,"quantity":10,"remark":"测试上料"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("buffer transfer status %d: %s", rec.Code, rec.Body.String())
	}
	var flow PlantBufferFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &flow); err != nil {
		t.Fatalf("decode transfer flow: %v", err)
	}
	if flow.Direction != "in" || flow.Quantity != 10 || flow.BalanceQty != 10 {
		t.Fatalf("unexpected transfer flow: %+v", flow)
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/plant-buffer-locations/"+strconv.FormatInt(buffer.ID, 10), "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("delete plant buffer with flow should be rejected, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/buffer-adjustments", `{"bufferCode":"NS-AMP240-AGG-T","actualQty":8,"moistureRate":4.5,"qualityStatus":"passed","remark":"盘点"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("buffer adjustment status %d: %s", rec.Code, rec.Body.String())
	}
	if !plantBufferHasQty(app.mustSnapshot().PlantBufferLocations, "NS-AMP240-AGG-T", 8) {
		t.Fatalf("expected adjusted buffer qty, got %+v", app.mustSnapshot().PlantBufferLocations)
	}

	payload := `{"channel":"industrial-control-gateway","protocol":"buffer-json","payload":{"deviceNo":"PLANT-NS-AMP240","plantCode":"NS-AMP240","bufferCode":"NS-AMP240-AGG-T","materialId":3,"quantity":7,"moistureRate":4.8,"qualityStatus":"passed","status":"active","reportedAt":"2026-06-20 11:00:00"}}`
	rec = testDeviceRequest(t, app, "plant-demo-key-1", http.MethodPost, "/api/production-plans/protocols/buffer/ingest", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("buffer protocol ingest status %d: %s", rec.Code, rec.Body.String())
	}
	var response protocolIngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode buffer protocol response: %v", err)
	}
	if response.Frame.Status != "accepted" || response.Frame.ParsedResource != "plant_buffer" || response.PlantBufferFlow == nil || response.PlantBufferFlow.BalanceQty != 7 {
		t.Fatalf("unexpected buffer protocol response: %+v", response)
	}
}

func TestPlantBufferAdjustmentWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"plant_buffer_adjustment_review","name":"筒仓盘点复核","category":"approval","resource":"plant_buffer_adjustment","trigger":{"eventType":"plant_buffer_adjustment.requested","resource":"plant_buffer_adjustment"},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"盘点复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant buffer adjustment workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/plant-buffer-locations", `{"siteId":1,"plantId":1,"code":"NS-AMP240-AGG-WF","name":"流程骨料仓","type":"aggregate_bin","materialId":3,"allowedMaterialIds":[3],"capacity":100,"unit":"t","warningQty":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant buffer status %d: %s", rec.Code, rec.Body.String())
	}
	var buffer PlantBufferLocation
	if err := json.Unmarshal(rec.Body.Bytes(), &buffer); err != nil {
		t.Fatalf("decode buffer: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/buffer-transfers", `{"bufferId":`+strconv.FormatInt(buffer.ID, 10)+`,"materialId":3,"quantity":10,"remark":"流程测试上料"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("buffer transfer status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/buffer-adjustments", `{"bufferId":`+strconv.FormatInt(buffer.ID, 10)+`,"actualQty":8,"moistureRate":4.5,"qualityStatus":"passed","remark":"流程盘点"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("buffer adjustment workflow request status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode workflow instance: %v", err)
	}
	if instance.Resource != "plant_buffer_adjustment" || instance.ResourceID != buffer.ID || instance.Status != "pending" {
		t.Fatalf("unexpected workflow instance: %+v", instance)
	}
	snapshot := app.mustSnapshot()
	if plantBufferQty(snapshot.PlantBufferLocations, "NS-AMP240-AGG-WF") != 10 {
		t.Fatalf("expected buffer qty unchanged before approval, got %+v", snapshot.PlantBufferLocations)
	}
	if len(snapshot.WorkflowTasks) == 0 || snapshot.WorkflowEvents[len(snapshot.WorkflowEvents)-1].Status != "handled" {
		t.Fatalf("expected handled buffer adjustment workflow, got events %+v tasks %+v", snapshot.WorkflowEvents, snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[0].ID, 10)+"/act", `{"action":"approve","comment":"盘点通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve buffer adjustment workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	if plantBufferQty(snapshot.PlantBufferLocations, "NS-AMP240-AGG-WF") != 8 {
		t.Fatalf("expected approved buffer adjustment qty 8, got %+v", snapshot.PlantBufferLocations)
	}
	if !hasPlantBufferFlow(snapshot.PlantBufferFlows, "NS-AMP240-AGG-WF", "buffer_adjustment", "adjustment_out") {
		t.Fatalf("expected workflow buffer adjustment flow, got %+v", snapshot.PlantBufferFlows)
	}
}

func TestPlantBatchCanConsumeSpecifiedBuffer(t *testing.T) {
	app := newTestHTTPApp(t)
	task := createPlantTaskForTest(t, app, 6)
	before := plantBufferQty(app.mustSnapshot().PlantBufferLocations, "NS-AMP240-AGG-02")
	payload := `{"channel":"opc","protocol":"plant-json","payload":{
		"deviceNo":"PLANT-NS-AMP240",
		"taskId":` + strconv.FormatInt(task.ID, 10) + `,
		"batchNo":"PLC-BUFFER-001",
		"plantCode":"NS-AMP240",
		"quantity":3,
		"completedAt":"2026-06-20 12:20:00",
		"materials":[{"materialId":4,"quantity":1.5,"unit":"t","bufferCode":"NS-AMP240-AGG-02"}]
	}}`
	rec := testDeviceRequest(t, app, "plant-demo-key-1", http.MethodPost, "/api/production-plans/protocols/plant/ingest", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("plant protocol ingest status %d: %s", rec.Code, rec.Body.String())
	}
	data := app.mustSnapshot()
	after := plantBufferQty(data.PlantBufferLocations, "NS-AMP240-AGG-02")
	if round(before-after) != 1.5 {
		t.Fatalf("expected buffer consumption 1.5, before %.2f after %.2f", before, after)
	}
	if !hasPlantBufferFlow(data.PlantBufferFlows, "NS-AMP240-AGG-02", "production_batch", "out") {
		t.Fatalf("expected production batch buffer flow, got %+v", data.PlantBufferFlows)
	}
}

func plantBufferQty(items []PlantBufferLocation, code string) float64 {
	for _, item := range items {
		if item.Code == code {
			return item.CurrentQty
		}
	}
	return 0
}

func plantBufferHasQty(items []PlantBufferLocation, code string, qty float64) bool {
	return plantBufferQty(items, code) == qty
}

func hasPlantBufferFlow(items []PlantBufferFlow, bufferCode string, sourceType string, direction string) bool {
	for _, item := range items {
		if item.BufferCode == bufferCode && item.SourceType == sourceType && item.Direction == direction {
			return true
		}
	}
	return false
}
