package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	if plan.PlantID == 0 || plan.PlantCode != "NS-AMP240" {
		t.Fatalf("expected plan to bind production line, got %+v", plan)
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
	if task.PlantID != plan.PlantID || task.PlantCode != plan.PlantCode {
		t.Fatalf("expected task to inherit production line, got task=%+v plan=%+v", task, plan)
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
	if batch.PlantID != plan.PlantID || batch.PlantCode != plan.PlantCode {
		t.Fatalf("expected batch to inherit production line, got batch=%+v plan=%+v", batch, plan)
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

func TestProductionPlanRejectsLineWithoutRequiredBuffer(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	if err := app.store.Mutate(func(data *AppData) error {
		data.Plants = append(data.Plants, Plant{ID: 99, SiteID: 1, Name: "南山备用沥青线", Code: "NS-AMP160", Capacity: "160t/h", Status: "running"})
		return nil
	}); err != nil {
		t.Fatalf("seed backup production line: %v", err)
	}

	rec := testRequest(t, app, token, http.MethodPost, "/api/production-plans", `{"orderId":2,"plantId":99,"planQuantity":12,"planDate":"2026-06-20","shift":"白班"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected recipe mismatch rejection, got status %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "生产线与配比不匹配") {
		t.Fatalf("expected production line mismatch message, got %s", rec.Body.String())
	}
}

func TestProductionLineMixProfileLocksTaskAndBatchConsumption(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	qualityToken := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/2/plant-profiles", `{
		"plantId":1,
		"code":"MD-AC20-NS-AMP240-TEST",
		"version":"v2-line-test",
		"scope":"1号线机制砂微调",
		"effectiveFrom":"2026-06-01",
		"materials":[{"materialId":3,"adjustment":-12,"unit":"kg/t","bufferId":1}]
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant profile status %d: %s", rec.Code, rec.Body.String())
	}
	var profile MixDesignPlantProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode plant profile: %v", err)
	}
	if profile.MixDesignID != 2 || profile.PlantID != 1 || !profile.IsCurrent {
		t.Fatalf("unexpected plant profile: %+v", profile)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/production-plans", `{"orderId":2,"plantId":1,"planQuantity":10,"planDate":"2026-06-22","shift":"白班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan ProductionPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode production plan: %v", err)
	}
	if plan.MixDesignID != 2 || plan.MixProfileID != profile.ID {
		t.Fatalf("expected plan to lock plant profile %d, got %+v", profile.ID, plan)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/production-plans/"+strconv.FormatInt(plan.ID, 10)+"/tasks", `{"planQty":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production task status %d: %s", rec.Code, rec.Body.String())
	}
	var task ProductionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode production task: %v", err)
	}
	if task.MixDesignID != 2 || task.MixProfileID != profile.ID {
		t.Fatalf("expected task to inherit plant profile %d, got %+v", profile.ID, task)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/2/plant-profiles", `{
		"plantId":1,
		"code":"MD-AC20-NS-AMP240-NEW",
		"version":"v2-line-new",
		"effectiveFrom":"2026-06-01",
		"materials":[{"materialId":3,"adjustment":-60,"unit":"kg/t","bufferId":1}]
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create replacement plant profile status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/production-plans/tasks/"+strconv.FormatInt(task.ID, 10)+"/batches", `{"quantity":10,"qualityStatus":"passed","status":"released","completedAt":"2026-06-22 10:20:00"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production batch status %d: %s", rec.Code, rec.Body.String())
	}
	var batch ProductionBatch
	if err := json.Unmarshal(rec.Body.Bytes(), &batch); err != nil {
		t.Fatalf("decode production batch: %v", err)
	}
	if batch.MixDesignID != 2 || batch.MixProfileID != profile.ID {
		t.Fatalf("expected batch to keep original plant profile %d, got %+v", profile.ID, batch)
	}
	flow, ok := findPlantBufferFlow(app.mustSnapshot(), batch.ID, 3)
	if !ok || flow.Quantity != 3.48 {
		t.Fatalf("expected old profile buffer deduction 3.48, got found=%v flow=%+v", ok, flow)
	}
}

func TestMixDesignPlantProfileWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	qualityToken := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"mix_design_plant_profile_review","name":"生产线配比微调审批","category":"approval","resource":"mix_design_plant_profile","trigger":{"eventType":"mix_design_plant_profile.submitted","resource":"mix_design_plant_profile","conditions":[{"field":"plantId","operator":"equals","value":"1"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"生产线配比复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant profile workflow status %d: %s", rec.Code, rec.Body.String())
	}

	before := app.mustSnapshot()
	previousCurrentID := int64(0)
	for _, item := range before.MixDesignPlantProfiles {
		if item.MixDesignID == 2 && item.PlantID == 1 && item.Status == "approved" && item.IsCurrent {
			previousCurrentID = item.ID
			break
		}
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/2/plant-profiles", `{
		"plantId":1,
		"code":"MD-AC20-NS-AMP240-WF",
		"version":"v2-line-wf",
		"scope":"流程审批机制砂微调",
		"effectiveFrom":"2026-06-01",
		"materials":[{"materialId":3,"adjustment":-36,"unit":"kg/t","bufferId":1}]
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow plant profile status %d: %s", rec.Code, rec.Body.String())
	}
	var profile MixDesignPlantProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode workflow plant profile: %v", err)
	}
	if profile.Status != "pending_approval" || profile.IsCurrent {
		t.Fatalf("expected pending non-current plant profile before workflow approval, got %+v", profile)
	}

	snapshot := app.mustSnapshot()
	pending, ok := mixDesignPlantProfileByID(snapshot.MixDesignPlantProfiles, profile.ID)
	if !ok || pending.Status != "pending_approval" || pending.IsCurrent {
		t.Fatalf("expected persisted pending plant profile, got found=%v profile=%+v", ok, pending)
	}
	if previousCurrentID > 0 {
		previous, ok := mixDesignPlantProfileByID(snapshot.MixDesignPlantProfiles, previousCurrentID)
		if !ok || previous.Status != "approved" || !previous.IsCurrent {
			t.Fatalf("previous current profile should stay active before workflow approval, got found=%v profile=%+v", ok, previous)
		}
	}
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "mix_design_plant_profile" && task.ResourceID == profile.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending plant profile workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"生产线配比复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve plant profile workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	approved, ok := mixDesignPlantProfileByID(snapshot.MixDesignPlantProfiles, profile.ID)
	if !ok || approved.Status != "approved" || !approved.IsCurrent || approved.ApprovedAt == "" {
		t.Fatalf("expected workflow-approved current plant profile, got found=%v profile=%+v", ok, approved)
	}
	if previousCurrentID > 0 {
		previous, ok := mixDesignPlantProfileByID(snapshot.MixDesignPlantProfiles, previousCurrentID)
		if !ok || previous.Status != "retired" || previous.IsCurrent {
			t.Fatalf("expected previous profile retired after workflow approval, got found=%v profile=%+v", ok, previous)
		}
	}
}

func TestMixDesignPlantProfileRetireWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"mix_design_plant_profile_retire_review","name":"生产线配比退役审批","category":"approval","resource":"mix_design_plant_profile","trigger":{"eventType":"mix_design_plant_profile.retire_requested","resource":"mix_design_plant_profile","conditions":[{"field":"plantId","operator":"equals","value":"1"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"退役复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant profile retire workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot := app.mustSnapshot()
	profileID := int64(0)
	for _, item := range snapshot.MixDesignPlantProfiles {
		if item.PlantID == 1 && item.Status == "approved" && item.IsCurrent {
			profileID = item.ID
			break
		}
	}
	if profileID == 0 {
		t.Fatalf("expected current plant profile seed, got %+v", snapshot.MixDesignPlantProfiles)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/mix-design-plant-profiles/"+strconv.FormatInt(profileID, 10)+"/retire", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request plant profile retire workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending MixDesignPlantProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending plant profile retire: %v", err)
	}
	if pending.Status != "approved" || !pending.IsCurrent || pending.RetiredAt != "" {
		t.Fatalf("plant profile should stay current before retire workflow approval, got %+v", pending)
	}

	snapshot = app.mustSnapshot()
	current, ok := mixDesignPlantProfileByID(snapshot.MixDesignPlantProfiles, profileID)
	if !ok || current.Status != "approved" || !current.IsCurrent {
		t.Fatalf("expected current plant profile before approval, got found=%v profile=%+v", ok, current)
	}
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "mix_design_plant_profile" && task.ResourceID == profileID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending plant profile retire workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"退役复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve plant profile retire workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	retired, ok := mixDesignPlantProfileByID(snapshot.MixDesignPlantProfiles, profileID)
	if !ok || retired.Status != "retired" || retired.IsCurrent || retired.RetiredAt == "" {
		t.Fatalf("expected workflow-approved retired plant profile, got found=%v profile=%+v", ok, retired)
	}
}

func TestProductionPlanAdjustmentAutoTasksAndCancel(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/production-plans", `{"orderId":2,"planQuantity":121,"planDate":"2026-06-20","shift":"白班"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected over-plan rejection, got status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans", `{"orderId":2,"planQuantity":30,"planDate":"2026-06-20","shift":"白班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan ProductionPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan: %v", err)
	}
	if plan.CapacityStatus != "ok" || plan.InventoryStatus == "" || plan.RecipeStatus != "ok" {
		t.Fatalf("expected readiness status on plan, got %+v", plan)
	}

	planID := strconv.FormatInt(plan.ID, 10)
	rec = testRequest(t, app, token, http.MethodPatch, "/api/production-plans/"+planID, `{"planQuantity":40,"planDate":"2026-06-21","shift":"夜班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("update production plan status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode updated plan: %v", err)
	}
	if plan.PlanQuantity != 40 || plan.PlanDate != "2026-06-21" || plan.Shift != "夜班" {
		t.Fatalf("unexpected updated plan: %+v", plan)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+planID+"/tasks/auto", `{"taskQty":20}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("auto create tasks status %d: %s", rec.Code, rec.Body.String())
	}
	var tasks []ProductionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("decode auto tasks: %v", err)
	}
	if len(tasks) != 2 || tasks[0].PlanQty != 20 || tasks[1].PlanQty != 20 {
		t.Fatalf("unexpected auto tasks: %+v", tasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+planID+"/cancel", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("cancel production plan status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode cancelled plan: %v", err)
	}
	if plan.Status != "cancelled" {
		t.Fatalf("expected cancelled plan, got %+v", plan)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/production-plans/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("production overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Plans []ProductionPlan      `json:"plans"`
		Tasks []ProductionTask      `json:"tasks"`
		KPIs  ProductionOverviewKPI `json:"kpis"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if !hasProductionPlanStatus(overview.Plans, plan.ID, "cancelled") {
		t.Fatalf("expected cancelled plan in overview, got %+v", overview.Plans)
	}
	if !allProductionTasksCancelled(overview.Tasks, tasks) {
		t.Fatalf("expected auto tasks cancelled, got %+v", overview.Tasks)
	}
	if overview.KPIs.PlanCount == 0 || overview.KPIs.ActivePlanCount == 0 {
		t.Fatalf("expected production kpis, got %+v", overview.KPIs)
	}
}

func TestProductionPlanCancelWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"production_plan_cancel_review","name":"生产计划取消审批","category":"approval","resource":"production_plan","trigger":{"eventType":"production_plan.cancel_requested","resource":"production_plan","conditions":[{"field":"siteId","operator":"equals","value":"1"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"取消复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production cancel workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans", `{"orderId":2,"planQuantity":20,"planDate":"2026-06-23","shift":"白班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow production plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan ProductionPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode workflow production plan: %v", err)
	}
	planID := strconv.FormatInt(plan.ID, 10)
	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+planID+"/tasks/auto", `{"taskQty":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("auto create workflow plan tasks status %d: %s", rec.Code, rec.Body.String())
	}
	var tasks []ProductionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("decode workflow plan tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected two workflow plan tasks, got %+v", tasks)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+planID+"/cancel", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request production cancel workflow status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode pending cancel plan: %v", err)
	}
	if plan.Status != "pending_approval" {
		t.Fatalf("expected pending production plan before cancel approval, got %+v", plan)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/production-plans/"+planID+"/tasks", `{"planQty":1}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected pending cancel plan to block new tasks, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/production-plans/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("production overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Plans []ProductionPlan `json:"plans"`
		Tasks []ProductionTask `json:"tasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode pending cancel overview: %v", err)
	}
	if !hasProductionPlanStatus(overview.Plans, plan.ID, "pending_approval") || !allProductionTasksStatus(overview.Tasks, tasks, "pending") {
		t.Fatalf("expected plan pending and tasks unchanged before approval, got plans=%+v tasks=%+v", overview.Plans, overview.Tasks)
	}

	snapshot := app.mustSnapshot()
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "production_plan" && task.ResourceID == plan.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending production plan workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"取消复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve production cancel workflow status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/production-plans/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("production overview after approval status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode approved cancel overview: %v", err)
	}
	if !hasProductionPlanStatus(overview.Plans, plan.ID, "cancelled") || !allProductionTasksCancelled(overview.Tasks, tasks) {
		t.Fatalf("expected workflow-approved cancellation, got plans=%+v tasks=%+v", overview.Plans, overview.Tasks)
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

func hasProductionPlanStatus(items []ProductionPlan, id int64, status string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == status
		}
	}
	return false
}

func allProductionTasksCancelled(items []ProductionTask, created []ProductionTask) bool {
	return allProductionTasksStatus(items, created, "cancelled")
}

func allProductionTasksStatus(items []ProductionTask, created []ProductionTask, status string) bool {
	wanted := map[int64]bool{}
	for _, item := range created {
		wanted[item.ID] = false
	}
	for _, item := range items {
		if _, ok := wanted[item.ID]; ok {
			wanted[item.ID] = item.Status == status
		}
	}
	for _, ok := range wanted {
		if !ok {
			return false
		}
	}
	return true
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

func findPlantBufferFlow(data AppData, batchID, materialID int64) (PlantBufferFlow, bool) {
	for _, item := range data.PlantBufferFlows {
		if item.SourceType == "production_batch" && item.SourceID == batchID && item.MaterialID == materialID {
			return item, true
		}
	}
	return PlantBufferFlow{}, false
}

func mixDesignPlantProfileByID(items []MixDesignPlantProfile, id int64) (MixDesignPlantProfile, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return MixDesignPlantProfile{}, false
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
