package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestLaboratoryMixDesignLifecycleAndProductionDefault(t *testing.T) {
	app := newTestHTTPApp(t)
	qualityToken := testLogin(t, app, "quality", "quality123")
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, qualityToken, http.MethodPost, "/api/laboratory/mix-designs", `{
		"productId":2,"siteId":1,"code":"MD-C40-PUMP","version":"v3","strengthGrade":"C40","slump":"170mm",
		"materials":[{"materialId":1,"dosage":390,"unit":"kg/m3"},{"materialId":3,"dosage":750,"unit":"kg/m3"},{"materialId":4,"dosage":1010,"unit":"kg/m3"},{"materialId":5,"dosage":8.8,"unit":"kg/m3"}]
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
