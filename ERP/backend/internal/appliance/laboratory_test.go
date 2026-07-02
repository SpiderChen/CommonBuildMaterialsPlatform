package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLaboratoryPayloadEncodesEmptyCollectionsAsArrays(t *testing.T) {
	raw, err := json.Marshal(laboratoryPayload(AppData{}))
	if err != nil {
		t.Fatalf("marshal laboratory payload: %v", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode laboratory payload: %v", err)
	}

	fields := []string{
		"mixDesigns",
		"trialRuns",
		"qualityInspections",
		"qualitySamples",
		"rawInspections",
		"samples",
		"tests",
		"equipment",
		"calibrations",
		"exceptions",
		"batches",
		"receipts",
		"products",
		"materials",
		"sites",
	}
	for _, field := range fields {
		var items []json.RawMessage
		if err := json.Unmarshal(payload[field], &items); err != nil {
			t.Fatalf("expected %s to be a JSON array, got %s: %v", field, string(payload[field]), err)
		}
		if items == nil {
			t.Fatalf("expected %s to decode as a non-nil slice", field)
		}
	}
}

func TestLaboratoryExceptionCreationPublishesWorkflowEvent(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	qualityToken := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_lab","name":"实验室质量异常审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.submitted","resource":"quality_exception","conditions":[{"field":"severity","operator":"equals","value":"high"}]},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow definition status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/exceptions", `{"siteId":1,"severity":"high","title":"试验异常自动审批","description":"稳定度试验异常","responsible":"quality"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quality exception status %d: %s", rec.Code, rec.Body.String())
	}
	var item QualityException
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode quality exception: %v", err)
	}
	snapshot := app.mustSnapshot()
	if len(snapshot.WorkflowEvents) != 1 || snapshot.WorkflowEvents[0].EventType != "quality_exception.submitted" || snapshot.WorkflowEvents[0].ResourceID != item.ID || snapshot.WorkflowEvents[0].Status != "handled" {
		t.Fatalf("expected handled quality exception workflow event, got %+v", snapshot.WorkflowEvents)
	}
	if len(snapshot.WorkflowInstances) != 1 || snapshot.WorkflowInstances[0].Resource != "quality_exception" || snapshot.WorkflowInstances[0].ResourceID != item.ID || snapshot.WorkflowInstances[0].CurrentRole != "quality" {
		t.Fatalf("expected quality exception workflow instance, got %+v", snapshot.WorkflowInstances)
	}
	if len(snapshot.WorkflowTasks) != 1 || snapshot.WorkflowTasks[0].Status != "pending" || snapshot.WorkflowTasks[0].RoleCode != "quality" {
		t.Fatalf("expected quality exception workflow task, got %+v", snapshot.WorkflowTasks)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(snapshot.WorkflowTasks[0].ID, 10)+"/act", `{"action":"approve","comment":"质检确认"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("act quality exception workflow task status %d: %s", rec.Code, rec.Body.String())
	}
	var instance WorkflowInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode quality exception workflow instance: %v", err)
	}
	if instance.Status != "approved" || instance.Resource != "quality_exception" || instance.ResourceID != item.ID {
		t.Fatalf("expected approved quality exception workflow instance, got %+v", instance)
	}
}

func TestLaboratoryExceptionCloseWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	qualityToken := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/exceptions", `{"siteId":1,"severity":"medium","title":"关闭审批异常","description":"等待闭环","responsible":"quality"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quality exception status %d: %s", rec.Code, rec.Body.String())
	}
	var item QualityException
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode quality exception: %v", err)
	}
	if item.Status != "open" {
		t.Fatalf("expected open quality exception, got %+v", item)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"quality_exception_close_review","name":"质量异常关闭审批","category":"approval","resource":"quality_exception","trigger":{"eventType":"quality_exception.close_requested","resource":"quality_exception","conditions":[{"field":"targetStatus","operator":"equals","value":"closed"}]},"steps":[{"seq":1,"roleCode":"quality","action":"approve","name":"质检闭环确认"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quality exception close workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/exceptions/"+strconv.FormatInt(item.ID, 10)+"/handle", `{"rootCause":"原材料含水波动","correctiveAction":"调整含水率并复检"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request quality exception close workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending QualityException
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending quality exception: %v", err)
	}
	if pending.Status != "open" || pending.RootCause != "" || pending.CorrectiveAction != "" {
		t.Fatalf("quality exception should remain unchanged before workflow approval, got %+v", pending)
	}

	snapshot := app.mustSnapshot()
	current, ok := qualityExceptionByID(snapshot.QualityExceptions, item.ID)
	if !ok || current.Status != "open" || current.RootCause != "" || current.CorrectiveAction != "" {
		t.Fatalf("expected persisted open quality exception before approval, got found=%v item=%+v", ok, current)
	}
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "quality_exception" && task.ResourceID == item.ID && task.Status == "pending" {
			if workflowTaskEventType(snapshot, task) == "quality_exception.close_requested" {
				taskID = task.ID
				break
			}
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending quality exception close workflow task, got tasks=%+v events=%+v", snapshot.WorkflowTasks, snapshot.WorkflowEvents)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"闭环确认"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve quality exception close workflow status %d: %s", rec.Code, rec.Body.String())
	}
	updated, ok := qualityExceptionByID(app.mustSnapshot().QualityExceptions, item.ID)
	if !ok || updated.Status != "closed" || updated.RootCause != "原材料含水波动" || updated.CorrectiveAction != "调整含水率并复检" {
		t.Fatalf("expected closed quality exception after workflow approval, got found=%v item=%+v", ok, updated)
	}
}

func TestLaboratoryMixDesignLifecycleAndProductionDefault(t *testing.T) {
	app := newTestHTTPApp(t)
	qualityToken := testLogin(t, app, "quality", "quality123")
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs", `{
		"productId":2,"siteId":1,"code":"MD-BAD-MATERIAL","version":"v1","strengthGrade":"AC-20","slump":"油石比 4.8%",
		"materials":[{"materialId":999,"dosage":390,"unit":"kg/t"}]
	}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "配比材料不存在") {
		t.Fatalf("expected unknown material rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs", `{
		"productId":2,"siteId":1,"code":"MD-DUP-MATERIAL","version":"v1","strengthGrade":"AC-20","slump":"油石比 4.8%",
		"materials":[{"materialId":1,"dosage":390,"unit":"kg/t"},{"materialId":1,"dosage":20,"unit":"kg/t"}]
	}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "配比材料不能重复") {
		t.Fatalf("expected duplicate material rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs", `{
		"productId":2,"siteId":1,"code":"MD-AC20-PAVE","version":"v3","strengthGrade":"AC-20","slump":"油石比 4.8%",
		"materials":[{"materialId":1,"dosage":52,"unit":"kg/t"},{"materialId":3,"dosage":380,"unit":"kg/t"},{"materialId":4,"dosage":560,"unit":"kg/t"},{"materialId":5,"dosage":8.8,"unit":"kg/t"}]
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create mix design status %d: %s", rec.Code, rec.Body.String())
	}
	var mix MixDesign
	if err := json.Unmarshal(rec.Body.Bytes(), &mix); err != nil {
		t.Fatalf("decode mix: %v", err)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/"+strconv.FormatInt(mix.ID, 10)+"/trial-runs", `{"strength28d":25,"result":"failed"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create failed trial status %d: %s", rec.Code, rec.Body.String())
	}
	var failedTrial MixDesignTrialRun
	if err := json.Unmarshal(rec.Body.Bytes(), &failedTrial); err != nil {
		t.Fatalf("decode failed trial: %v", err)
	}
	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/"+strconv.FormatInt(mix.ID, 10)+"/approve", `{"trialRunId":`+strconv.FormatInt(failedTrial.ID, 10)+`}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected failed trial blocks approval, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/"+strconv.FormatInt(mix.ID, 10)+"/trial-runs", `{"strength7d":33,"strength28d":45,"result":"passed"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create passed trial status %d: %s", rec.Code, rec.Body.String())
	}
	var passedTrial MixDesignTrialRun
	if err := json.Unmarshal(rec.Body.Bytes(), &passedTrial); err != nil {
		t.Fatalf("decode passed trial: %v", err)
	}

	ch := app.hub.Subscribe()
	defer app.hub.Unsubscribe(ch)
	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/"+strconv.FormatInt(mix.ID, 10)+"/approve", `{"trialRunId":`+strconv.FormatInt(passedTrial.ID, 10)+`,"effectiveFrom":"2026-06-19"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve mix status %d: %s", rec.Code, rec.Body.String())
	}
	expectEvent(t, ch, "laboratory.mix_design.approved")

	overview := fetchLaboratoryOverview(t, app, qualityToken)
	if !hasCurrentMixDesign(overview.MixDesigns, mix.ID, 2, 1) || !hasRetiredMixDesign(overview.MixDesigns, 2) {
		t.Fatalf("expected new current and old retired mix designs, got %+v", overview.MixDesigns)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/production-plans", `{"orderId":2,"planQuantity":10,"planDate":"2026-06-21","shift":"白班"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan ProductionPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode plan: %v", err)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/production-plans/"+strconv.FormatInt(plan.ID, 10)+"/tasks", `{"planQty":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create production task status %d: %s", rec.Code, rec.Body.String())
	}
	var task ProductionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	if task.MixDesignID != mix.ID {
		t.Fatalf("expected production task to use approved current mix %d, got %+v", mix.ID, task)
	}
	if !hasHistoricalBatchMixID(fetchLaboratoryOverview(t, app, adminToken).Batches, 1, 1) {
		t.Fatalf("expected existing production batch to keep historical mix id")
	}
}

func TestMixDesignApprovalWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	qualityToken := testLogin(t, app, "quality", "quality123")
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"mix_design_approval_review","name":"生产配比审批","category":"approval","resource":"mix_design","trigger":{"eventType":"mix_design.submitted","resource":"mix_design","conditions":[{"field":"siteId","operator":"equals","value":"1"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"实验室复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create mix design workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs", `{
		"productId":2,"siteId":1,"code":"MD-AC20-WF","version":"v9","strengthGrade":"AC-20","slump":"油石比 4.9%",
		"materials":[{"materialId":1,"dosage":50,"unit":"kg/t"},{"materialId":3,"dosage":370,"unit":"kg/t"},{"materialId":4,"dosage":570,"unit":"kg/t"},{"materialId":5,"dosage":9.1,"unit":"kg/t"}]
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow mix design status %d: %s", rec.Code, rec.Body.String())
	}
	var mix MixDesign
	if err := json.Unmarshal(rec.Body.Bytes(), &mix); err != nil {
		t.Fatalf("decode workflow mix: %v", err)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/"+strconv.FormatInt(mix.ID, 10)+"/trial-runs", `{"strength7d":32,"strength28d":44,"result":"passed"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create workflow trial status %d: %s", rec.Code, rec.Body.String())
	}
	var trial MixDesignTrialRun
	if err := json.Unmarshal(rec.Body.Bytes(), &trial); err != nil {
		t.Fatalf("decode workflow trial: %v", err)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs/"+strconv.FormatInt(mix.ID, 10)+"/approve", `{"trialRunId":`+strconv.FormatInt(trial.ID, 10)+`,"effectiveFrom":"2026-06-25"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request mix design workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending MixDesign
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending mix: %v", err)
	}
	if pending.Status != "pending_approval" || pending.IsCurrent {
		t.Fatalf("expected pending mix before workflow approval, got %+v", pending)
	}
	overview := fetchLaboratoryOverview(t, app, qualityToken)
	if hasCurrentMixDesign(overview.MixDesigns, mix.ID, 2, 1) || hasRetiredMixDesign(overview.MixDesigns, 2) {
		t.Fatalf("mix should not be current before workflow approval, got %+v", overview.MixDesigns)
	}

	snapshot := app.mustSnapshot()
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "mix_design" && task.ResourceID == mix.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending mix design workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"实验室复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve mix design workflow status %d: %s", rec.Code, rec.Body.String())
	}
	overview = fetchLaboratoryOverview(t, app, qualityToken)
	if !hasCurrentMixDesign(overview.MixDesigns, mix.ID, 2, 1) || !hasRetiredMixDesign(overview.MixDesigns, 2) {
		t.Fatalf("expected workflow-approved current mix and retired old mix, got %+v", overview.MixDesigns)
	}
}

func TestMixDesignRetireWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/workflows/definitions", `{"code":"mix_design_retire_review","name":"生产配比退役审批","category":"approval","resource":"mix_design","trigger":{"eventType":"mix_design.retire_requested","resource":"mix_design","conditions":[{"field":"siteId","operator":"equals","value":"1"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"退役复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create mix retire workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/mix-designs/2/retire", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request mix retire workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending MixDesign
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending retire mix: %v", err)
	}
	if pending.Status != "approved" || !pending.IsCurrent || pending.RetiredAt != "" {
		t.Fatalf("mix should stay current before retire workflow approval, got %+v", pending)
	}
	overview := fetchLaboratoryOverview(t, app, token)
	if !hasCurrentMixDesign(overview.MixDesigns, 2, 2, 1) {
		t.Fatalf("mix should remain current before retire approval, got %+v", overview.MixDesigns)
	}

	snapshot := app.mustSnapshot()
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "mix_design" && task.ResourceID == 2 && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending mix retire workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"退役复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve mix retire workflow status %d: %s", rec.Code, rec.Body.String())
	}
	overview = fetchLaboratoryOverview(t, app, token)
	if !hasRetiredMixDesign(overview.MixDesigns, 2) {
		t.Fatalf("expected workflow-approved retired mix, got %+v", overview.MixDesigns)
	}
}

func TestLaboratoryEquipmentSampleTestExceptionAndPermissions(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/laboratory/samples", `{"siteId":2,"sampleType":"manual"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected site scoped quality user blocked, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/equipment", `{"name":"万能材料试验机","siteId":1,"status":"active","lastCalibrationAt":"2025-01-01","nextCalibrationAt":"2025-06-01","calibrationCycleDays":180}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create expired equipment status %d: %s", rec.Code, rec.Body.String())
	}
	var equipment LaboratoryEquipment
	if err := json.Unmarshal(rec.Body.Bytes(), &equipment); err != nil {
		t.Fatalf("decode equipment: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/samples", `{"siteId":1,"productId":1,"mixDesignId":1,"sampleType":"compressive_strength","plannedTestAt":"2026-06-20"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create sample status %d: %s", rec.Code, rec.Body.String())
	}
	var sample LaboratorySample
	if err := json.Unmarshal(rec.Body.Bytes(), &sample); err != nil {
		t.Fatalf("decode sample: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/samples/"+strconv.FormatInt(sample.ID, 10)+"/tests", `{"equipmentId":`+strconv.FormatInt(equipment.ID, 10)+`,"value":32}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected expired equipment blocked, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/equipment/"+strconv.FormatInt(equipment.ID, 10)+"/calibrations", `{"result":"passed","calibratedAt":"2026-06-19","nextDueAt":"2027-01-01","certificateNo":"CAL-TEST-1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("calibrate equipment status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/samples/"+strconv.FormatInt(sample.ID, 10)+"/tests", `{"equipmentId":`+strconv.FormatInt(equipment.ID, 10)+`,"value":28,"metric":"28d_strength","unit":"MPa"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create lab test status %d: %s", rec.Code, rec.Body.String())
	}
	var testRecord LaboratoryTestRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &testRecord); err != nil {
		t.Fatalf("decode test: %v", err)
	}

	ch := app.hub.Subscribe()
	defer app.hub.Unsubscribe(ch)
	rec = testRequest(t, app, token, http.MethodPost, "/api/laboratory/tests/"+strconv.FormatInt(testRecord.ID, 10)+"/review", `{"result":"failed","remark":"强度不足"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("review lab test status %d: %s", rec.Code, rec.Body.String())
	}
	expectEvent(t, ch, "laboratory.test.reviewed")

	overview := fetchLaboratoryOverview(t, app, token)
	if !hasClosedSampleResult(overview.Samples, sample.ID, "failed") || !hasOpenExceptionForSource(overview.Exceptions, "laboratory_test", testRecord.ID) {
		t.Fatalf("expected failed sample and open exception, got samples=%+v exceptions=%+v", overview.Samples, overview.Exceptions)
	}
	if !hasAuditForResource(t, app, "laboratory_test", testRecord.ID) {
		t.Fatalf("expected audit log for laboratory test review")
	}
}

func TestLaboratoryTestReviewWorkflowAppliesAfterApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	qualityToken := testLogin(t, app, "quality", "quality123")
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/definitions", `{"code":"laboratory_test_review","name":"试验复核审批","category":"approval","resource":"laboratory_test","trigger":{"eventType":"laboratory_test.review_requested","resource":"laboratory_test","conditions":[{"field":"result","operator":"equals","value":"failed"}]},"steps":[{"seq":1,"roleCode":"boss","action":"approve","name":"试验复核"}],"status":"active","version":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create laboratory test workflow status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/equipment", `{"name":"压力试验机","siteId":1,"status":"active","lastCalibrationAt":"2026-01-01","nextCalibrationAt":"2027-01-01","calibrationCycleDays":365}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create equipment status %d: %s", rec.Code, rec.Body.String())
	}
	var equipment LaboratoryEquipment
	if err := json.Unmarshal(rec.Body.Bytes(), &equipment); err != nil {
		t.Fatalf("decode equipment: %v", err)
	}
	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/samples", `{"siteId":1,"productId":1,"mixDesignId":1,"sampleType":"compressive_strength","plannedTestAt":"2026-06-22"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create sample status %d: %s", rec.Code, rec.Body.String())
	}
	var sample LaboratorySample
	if err := json.Unmarshal(rec.Body.Bytes(), &sample); err != nil {
		t.Fatalf("decode sample: %v", err)
	}
	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/samples/"+strconv.FormatInt(sample.ID, 10)+"/tests", `{"equipmentId":`+strconv.FormatInt(equipment.ID, 10)+`,"value":32,"metric":"28d_strength","unit":"MPa"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create laboratory test status %d: %s", rec.Code, rec.Body.String())
	}
	var testRecord LaboratoryTestRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &testRecord); err != nil {
		t.Fatalf("decode laboratory test: %v", err)
	}

	rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/tests/"+strconv.FormatInt(testRecord.ID, 10)+"/review", `{"result":"failed","remark":"流程复核强度不足"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("request laboratory test workflow status %d: %s", rec.Code, rec.Body.String())
	}
	var pending LaboratoryTestRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &pending); err != nil {
		t.Fatalf("decode pending laboratory test: %v", err)
	}
	if pending.Status != "pending_approval" || pending.ReviewedAt != "" || pending.Result == "failed" {
		t.Fatalf("expected pending laboratory test before approval, got %+v", pending)
	}
	overview := fetchLaboratoryOverview(t, app, qualityToken)
	if hasClosedSampleResult(overview.Samples, sample.ID, "failed") || hasOpenExceptionForSource(overview.Exceptions, "laboratory_test", testRecord.ID) {
		t.Fatalf("sample and exception should not update before workflow approval, got samples=%+v exceptions=%+v", overview.Samples, overview.Exceptions)
	}

	snapshot := app.mustSnapshot()
	taskID := int64(0)
	for _, task := range snapshot.WorkflowTasks {
		if task.Resource == "laboratory_test" && task.ResourceID == testRecord.ID && task.Status == "pending" {
			taskID = task.ID
			break
		}
	}
	if taskID == 0 {
		t.Fatalf("expected pending laboratory test workflow task, got %+v", snapshot.WorkflowTasks)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/workflows/tasks/"+strconv.FormatInt(taskID, 10)+"/act", `{"action":"approve","comment":"试验复核通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve laboratory test workflow status %d: %s", rec.Code, rec.Body.String())
	}
	snapshot = app.mustSnapshot()
	approved, ok := laboratoryTestByID(snapshot.LaboratoryTests, testRecord.ID)
	if !ok || approved.Status != "reviewed" || approved.Result != "failed" || approved.ReviewedAt == "" {
		t.Fatalf("expected reviewed laboratory test after workflow approval, got found=%v test=%+v", ok, approved)
	}
	overview = fetchLaboratoryOverview(t, app, qualityToken)
	if !hasClosedSampleResult(overview.Samples, sample.ID, "failed") || !hasOpenExceptionForSource(overview.Exceptions, "laboratory_test", testRecord.ID) {
		t.Fatalf("expected approved workflow to close sample and create exception, got samples=%+v exceptions=%+v", overview.Samples, overview.Exceptions)
	}
}

func fetchLaboratoryOverview(t *testing.T, app *App, token string) struct {
	MixDesigns []MixDesign        `json:"mixDesigns"`
	Batches    []ProductionBatch  `json:"batches"`
	Samples    []LaboratorySample `json:"samples"`
	Exceptions []QualityException `json:"exceptions"`
} {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/laboratory/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("laboratory overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		MixDesigns []MixDesign        `json:"mixDesigns"`
		Batches    []ProductionBatch  `json:"batches"`
		Samples    []LaboratorySample `json:"samples"`
		Exceptions []QualityException `json:"exceptions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode laboratory overview: %v", err)
	}
	return overview
}

func hasCurrentMixDesign(items []MixDesign, id, productID, siteID int64) bool {
	for _, item := range items {
		if item.ID == id && item.ProductID == productID && item.SiteID == siteID {
			return item.Status == "approved" && item.IsCurrent
		}
	}
	return false
}

func hasRetiredMixDesign(items []MixDesign, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == "retired" && !item.IsCurrent
		}
	}
	return false
}

func hasHistoricalBatchMixID(items []ProductionBatch, batchID, mixID int64) bool {
	for _, item := range items {
		if item.ID == batchID {
			return item.MixDesignID == mixID
		}
	}
	return false
}

func hasClosedSampleResult(items []LaboratorySample, id int64, result string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == "completed" && item.Result == result
		}
	}
	return false
}

func hasOpenExceptionForSource(items []QualityException, sourceType string, sourceID int64) bool {
	for _, item := range items {
		if item.SourceType == sourceType && item.SourceID == sourceID {
			return item.Status == "open"
		}
	}
	return false
}

func qualityExceptionByID(items []QualityException, id int64) (QualityException, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return QualityException{}, false
}

func workflowTaskEventType(data AppData, task WorkflowTask) string {
	for _, instance := range data.WorkflowInstances {
		if instance.ID != task.InstanceID {
			continue
		}
		for _, event := range data.WorkflowEvents {
			if event.ID == instance.TriggerEventID {
				return event.EventType
			}
		}
	}
	return ""
}

func laboratoryTestByID(items []LaboratoryTestRecord, id int64) (LaboratoryTestRecord, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return LaboratoryTestRecord{}, false
}

func hasAuditForResource(t *testing.T, app *App, resource string, id int64) bool {
	t.Helper()
	data, err := app.store.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	for _, item := range data.AuditLogs {
		if item.Resource == resource && item.ResourceID == id {
			return true
		}
	}
	return false
}

func expectEvent(t *testing.T, ch chan Event, topic string) {
	t.Helper()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case event := <-ch:
			if event.Topic == topic {
				return
			}
		case <-timer.C:
			t.Fatalf("expected event %s", topic)
		}
	}
}
