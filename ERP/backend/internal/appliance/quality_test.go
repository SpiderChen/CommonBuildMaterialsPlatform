package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestQualityInspectionSamplesUpdateBatchStatus(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/quality/inspections", `{"batchId":1,"slump":"180mm","temperature":28.5,"remark":"出厂抽检"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quality inspection status %d: %s", rec.Code, rec.Body.String())
	}
	var inspection QualityInspection
	if err := json.Unmarshal(rec.Body.Bytes(), &inspection); err != nil {
		t.Fatalf("decode inspection: %v", err)
	}
	if inspection.BatchID != 1 || inspection.Status != "sampling" || inspection.SampleCount != 2 {
		t.Fatalf("unexpected inspection: %+v", inspection)
	}

	overview := fetchQualityOverview(t, app, token)
	if len(overview.Samples) != 2 {
		t.Fatalf("expected two default samples, got %+v", overview.Samples)
	}
	for i, sample := range overview.Samples {
		body := `{"strength":35.6,"result":"passed"}`
		if i == 1 {
			body = `{"strength":42.8,"result":"passed"}`
		}
		rec = testRequest(t, app, token, http.MethodPost, "/api/quality/samples/"+strconv.FormatInt(sample.ID, 10)+"/test", body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("test quality sample status %d: %s", rec.Code, rec.Body.String())
		}
	}

	overview = fetchQualityOverview(t, app, token)
	if !hasCompletedInspection(overview.Inspections, inspection.ID, "passed") {
		t.Fatalf("expected passed completed inspection, got %+v", overview.Inspections)
	}
	if !hasBatchQualityStatus(overview.Batches, 1, "passed") {
		t.Fatalf("expected batch quality status passed, got %+v", overview.Batches)
	}
}

func TestRawMaterialInspectionReleasesInventory(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/quality/raw-inspections", `{"receiptId":1,"moisture":3.2,"mudContent":1.4,"fineness":"II区","remark":"入厂抽检"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create raw material inspection status %d: %s", rec.Code, rec.Body.String())
	}
	var inspection RawMaterialInspection
	if err := json.Unmarshal(rec.Body.Bytes(), &inspection); err != nil {
		t.Fatalf("decode raw inspection: %v", err)
	}
	if inspection.ReceiptID != 1 || inspection.Status != "pending_review" || inspection.Result != "pending" {
		t.Fatalf("unexpected raw inspection: %+v", inspection)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/quality/raw-inspections/"+strconv.FormatInt(inspection.ID, 10)+"/review", `{"result":"passed","remark":"指标合格"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("review raw material inspection status %d: %s", rec.Code, rec.Body.String())
	}

	overview := fetchQualityOverview(t, app, token)
	if !hasCompletedRawInspection(overview.RawInspections, inspection.ID, "passed") {
		t.Fatalf("expected passed raw inspection, got %+v", overview.RawInspections)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/procurement/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("procurement overview status %d: %s", rec.Code, rec.Body.String())
	}
	var procurement struct {
		Receipts  []RawMaterialReceipt `json:"receipts"`
		Inventory []InventoryItem      `json:"inventory"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &procurement); err != nil {
		t.Fatalf("decode procurement overview: %v", err)
	}
	if !hasReceiptQualityStatus(procurement.Receipts, 1, "passed") {
		t.Fatalf("expected receipt quality passed, got %+v", procurement.Receipts)
	}
	if !hasInventoryQualityStatus(procurement.Inventory, 1, 3, "passed", "available") {
		t.Fatalf("expected released inventory, got %+v", procurement.Inventory)
	}
}

func fetchQualityOverview(t *testing.T, app *App, token string) struct {
	Inspections    []QualityInspection     `json:"inspections"`
	Samples        []QualitySample         `json:"samples"`
	RawInspections []RawMaterialInspection `json:"rawInspections"`
	Batches        []ProductionBatch       `json:"batches"`
} {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/quality/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("quality overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Inspections    []QualityInspection     `json:"inspections"`
		Samples        []QualitySample         `json:"samples"`
		RawInspections []RawMaterialInspection `json:"rawInspections"`
		Batches        []ProductionBatch       `json:"batches"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode quality overview: %v", err)
	}
	return overview
}

func hasCompletedInspection(items []QualityInspection, id int64, result string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == "completed" && item.Result == result && item.CompletedAt != ""
		}
	}
	return false
}

func hasBatchQualityStatus(items []ProductionBatch, id int64, status string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.QualityStatus == status
		}
	}
	return false
}

func hasCompletedRawInspection(items []RawMaterialInspection, id int64, result string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.Status == "completed" && item.Result == result && item.CompletedAt != ""
		}
	}
	return false
}

func hasReceiptQualityStatus(items []RawMaterialReceipt, id int64, status string) bool {
	for _, item := range items {
		if item.ID == id {
			return item.QualityStatus == status
		}
	}
	return false
}

func hasInventoryQualityStatus(items []InventoryItem, siteID, materialID int64, qualityStatus, availableStatus string) bool {
	for _, item := range items {
		if item.SiteID == siteID && item.MaterialID == materialID {
			return item.QualityStatus == qualityStatus && item.AvailableStatus == availableStatus
		}
	}
	return false
}
