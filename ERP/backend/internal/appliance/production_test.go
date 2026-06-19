package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestProductionTaskBatchAndDailyReport(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/production-plans", `{"orderId":2,"planQuantity":24,"planDate":"2026-06-20","shift":"白班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan ProductionPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan: %v", err)
	}
	if plan.Status != "scheduled" || plan.ProductID != 2 || plan.PlanQuantity != 24 {
		t.Fatalf("unexpected production plan: %+v", plan)
	}

	planID := strconv.FormatInt(plan.ID, 10)
	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+planID+"/tasks", `{"planQty":12}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production task status %d: %s", rec.Code, rec.Body.String())
	}
	var task ProductionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	if task.PlanID != plan.ID || task.MixDesignID == 0 || task.PlanQty != 12 || task.Status != "pending" {
		t.Fatalf("unexpected production task: %+v", task)
	}

	taskID := strconv.FormatInt(task.ID, 10)
	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/tasks/"+taskID+"/batches", `{"quantity":12,"qualityStatus":"passed","status":"released","completedAt":"2026-06-20 10:20:00"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production batch status %d: %s", rec.Code, rec.Body.String())
	}
	var batch ProductionBatch
	if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
		t.Fatalf("decode batch: %v", err)
	}
	if batch.TaskID != task.ID || batch.Quantity != 12 || batch.Status != "released" {
		t.Fatalf("unexpected production batch: %+v", batch)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/production-plans/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("production overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Plans   []ProductionPlan      `json:"plans"`
		Tasks   []ProductionTask      `json:"tasks"`
		Batches []ProductionBatch     `json:"batches"`
		Traces  []InventoryBatchTrace `json:"traces"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode production overview: %v", err)
	}
	if !hasCompletedProductionTask(overview.Tasks, task.ID) {
		t.Fatalf("expected completed task after batch, got %+v", overview.Tasks)
	}
	if !hasProducedPlanQty(overview.Plans, plan.ID, 12) {
		t.Fatalf("expected plan produced quantity update, got %+v", overview.Plans)
	}
	if !hasProductionBatch(overview.Batches, batch.ID) {
		t.Fatalf("expected batch in overview, got %+v", overview.Batches)
	}
	if countBatchTraces(overview.Traces, batch.ID) < 5 {
		t.Fatalf("expected material batch traces, got %+v", overview.Traces)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/procurement/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("procurement overview status %d: %s", rec.Code, rec.Body.String())
	}
	var procurement struct {
		Flows []InventoryFlow `json:"flows"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &procurement); err != nil {
		t.Fatalf("decode procurement overview: %v", err)
	}
	if countBatchInventoryFlows(procurement.Flows, batch.ID) < 5 {
		t.Fatalf("expected mix design inventory deductions, got %+v", procurement.Flows)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/reports/generate", `{"siteId":1,"reportDate":"2026-06-20"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("generate production report status %d: %s", rec.Code, rec.Body.String())
	}
	var report ProductionDailyReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if report.PlannedQty != 24 || report.ProducedQty != 12 || report.BatchCount != 1 || report.QualityPassed != 1 || report.MaterialCost != 5160 {
		t.Fatalf("unexpected production report: %+v", report)
	}
}

func hasCompletedProductionTask(items []ProductionTask, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == "completed" && item.ProducedQty == item.PlanQty
		}
	}
	return false
}

func hasProducedPlanQty(items []ProductionPlan, id int64, qty float64) bool {
	for _, item := range items {
		if item.ID == id {
			return item.ProducedQty == qty
		}
	}
	return false
}

func hasProductionBatch(items []ProductionBatch, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func countBatchInventoryFlows(items []InventoryFlow, batchID int64) int {
	count := 0
	for _, item := range items {
		if item.SourceType == "production_batch" && item.SourceID == batchID {
			count++
		}
	}
	return count
}

func countBatchTraces(items []InventoryBatchTrace, batchID int64) int {
	count := 0
	for _, item := range items {
		if item.ProductionBatchID == batchID {
			if item.BatchNo == "" || item.Warehouse == "" || item.Silo == "" || item.Quantity <= 0 {
				return 0
			}
			count++
		}
	}
	return count
}
