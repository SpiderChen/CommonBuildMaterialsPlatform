package appliance

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestPlantProtocolFrameIngestCreatesProductionBatchWithActualConsumption(t *testing.T) {
	app := newTestHTTPApp(t)
	task := createPlantTaskForTest(t, app, 12)
	payload := `{"channel":"opc","protocol":"plant-json","payload":{
		"deviceNo":"PLANT-NS-HZS180",
		"taskId":` + strconv.FormatInt(task.ID, 10) + `,
		"batchNo":"PLC-BATCH-001",
		"plantCode":"NS-HZS180",
		"quantity":6,
		"operator":"PLC-A",
		"qualityStatus":"pending",
		"status":"released",
		"startedAt":"2026-06-20 09:00:00",
		"completedAt":"2026-06-20 09:20:00",
		"materials":[
			{"materialId":1,"quantity":2.28,"unit":"t"},
			{"materialId":4,"quantity":6.12,"unit":"t"}
		]
	}}`
	rec := testDeviceRequest(t, app, "plant-demo-key-1", http.MethodPost, "/api/production-plans/protocols/plant/ingest", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("plant protocol ingest status %d: %s", rec.Code, rec.Body.String())
	}
	var response protocolIngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode plant protocol response: %v", err)
	}
	if response.Frame.Status != "accepted" || response.Frame.ParsedResource != "production_batch" || response.Frame.ParsedID == 0 {
		t.Fatalf("unexpected plant protocol frame: %+v", response.Frame)
	}
	if response.ProductionBatch == nil || response.ProductionBatch.BatchNo != "PLC-BATCH-001" || response.ProductionBatch.Quantity != 6 {
		t.Fatalf("unexpected plant production batch: %+v", response.ProductionBatch)
	}
	data := app.mustSnapshot()
	if !hasActualConsumptionFlow(data.InventoryFlows, response.ProductionBatch.ID, 2) {
		t.Fatalf("expected actual consumption inventory flows, got %+v", data.InventoryFlows)
	}
	if !hasProtocolFrame(data.DeviceProtocolFrames, "production_batch", response.ProductionBatch.ID, "accepted") {
		t.Fatalf("expected accepted production protocol frame, got %+v", data.DeviceProtocolFrames)
	}
}

func TestDeviceGatewaySerialFileIngestsPlantBatchFrame(t *testing.T) {
	app := newTestHTTPApp(t)
	task := createPlantTaskForTest(t, app, 8)
	dir := t.TempDir()
	serialFile := filepath.Join(dir, "plant-serial.log")
	line := "KEY=plant-demo-key-1|PLANT," + strconv.FormatInt(task.ID, 10) + ",NS-HZS180,4,2026-06-20 10:20:00,PLC-B,2026-06-20 10:00:00,released,pending,1:1.52:t|4:4.08:t\n"
	if err := os.WriteFile(serialFile, []byte(line), 0600); err != nil {
		t.Fatalf("write plant serial file: %v", err)
	}
	gateway := NewDeviceGateway(app, []DeviceGatewayEndpointConfig{
		{
			Name: "plant-file-test", Kind: "file", FilePath: serialFile, Parser: "plant",
			Channel: "serial", Protocol: "plant-csv", Permission: "plant:report",
		},
	}, 10*time.Millisecond)
	app.gateway = gateway
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := gateway.Start(ctx); err != nil {
		t.Fatalf("start plant file gateway: %v", err)
	}
	waitForCondition(t, 500*time.Millisecond, func() bool {
		data := app.mustSnapshot()
		for _, item := range data.ProductionBatches {
			if item.TaskID == task.ID && item.Quantity == 4 {
				return true
			}
		}
		return false
	})
	data := app.mustSnapshot()
	var batch ProductionBatch
	for _, item := range data.ProductionBatches {
		if item.TaskID == task.ID && item.Quantity == 4 {
			batch = item
			break
		}
	}
	if batch.ID == 0 || batch.PlantCode != "NS-HZS180" || batch.Status != "released" {
		t.Fatalf("expected plant gateway production batch, got %+v", data.ProductionBatches)
	}
	if !hasActualConsumptionFlow(data.InventoryFlows, batch.ID, 2) {
		t.Fatalf("expected gateway actual consumption flows, got %+v", data.InventoryFlows)
	}
	statuses := gateway.Status()
	if len(statuses) != 1 || statuses[0].AcceptedFrames != 1 || statuses[0].RejectedFrames != 0 {
		t.Fatalf("unexpected plant gateway status: %+v", statuses)
	}
}

func createPlantTaskForTest(t *testing.T, app *App, qty float64) ProductionTask {
	t.Helper()
	token := testLogin(t, app, "admin", "admin123")
	rec := testRequest(t, app, token, http.MethodPost, "/api/production-plans", `{"orderId":2,"planQuantity":`+strconv.FormatFloat(qty, 'f', -1, 64)+`,"planDate":"2026-06-20","shift":"白班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant test plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan ProductionPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode plant test plan: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+strconv.FormatInt(plan.ID, 10)+"/tasks", `{"planQty":`+strconv.FormatFloat(qty, 'f', -1, 64)+`}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant test task status %d: %s", rec.Code, rec.Body.String())
	}
	var task ProductionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode plant test task: %v", err)
	}
	return task
}

func hasActualConsumptionFlow(items []InventoryFlow, batchID int64, want int) bool {
	count := 0
	for _, item := range items {
		if item.SourceType == "production_batch" && item.SourceID == batchID && item.Remark == "拌合楼实际消耗" {
			count++
		}
	}
	return count == want
}

func hasProtocolFrame(items []DeviceProtocolFrame, resource string, parsedID int64, status string) bool {
	for _, item := range items {
		if item.ParsedResource == resource && item.ParsedID == parsedID && item.Status == status {
			return true
		}
	}
	return false
}
