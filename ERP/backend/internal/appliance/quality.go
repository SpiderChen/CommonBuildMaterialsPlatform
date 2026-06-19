package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (a *App) quality(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 || parts[0] == "overview" {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, qualityPayload(data))
		return
	}
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		switch parts[0] {
		case "inspections":
			writeJSON(w, http.StatusOK, data.QualityInspections)
		case "samples":
			writeJSON(w, http.StatusOK, data.QualitySamples)
		case "raw-inspections":
			writeJSON(w, http.StatusOK, data.RawMaterialInspections)
		default:
			writeError(w, http.StatusNotFound, "unknown quality resource")
		}
		return
	}
	if len(parts) == 1 && parts[0] == "inspections" && r.Method == http.MethodPost {
		a.createQualityInspection(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "raw-inspections" && parts[2] == "review" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.reviewRawMaterialInspection(w, r, session, id)
		return
	}
	if len(parts) == 1 && parts[0] == "raw-inspections" && r.Method == http.MethodPost {
		a.createRawMaterialInspection(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "samples" && parts[2] == "test" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.testQualitySample(w, r, session, id)
		return
	}
	writeError(w, http.StatusNotFound, "unknown quality route")
}

func qualityPayload(data AppData) map[string]interface{} {
	return map[string]interface{}{
		"inspections":       data.QualityInspections,
		"samples":           data.QualitySamples,
		"rawInspections":    data.RawMaterialInspections,
		"batches":           data.ProductionBatches,
		"receipts":          data.RawMaterialReceipts,
		"mixDesigns":        data.MixDesigns,
		"laboratorySamples": data.LaboratorySamples,
		"laboratoryTests":   data.LaboratoryTests,
		"equipment":         data.LaboratoryEquipment,
		"calibrations":      data.LaboratoryCalibrations,
		"qualityExceptions": data.QualityExceptions,
		"laboratoryKpis":    buildLaboratoryKPI(data),
	}
}

func (a *App) createQualityInspection(w http.ResponseWriter, r *http.Request, session Session) {
	var req QualityInspection
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid quality inspection")
		return
	}
	var item QualityInspection
	err := a.store.Mutate(func(data *AppData) error {
		batch, ok := findProductionBatch(*data, req.BatchID)
		if !ok {
			return fmt.Errorf("生产批次不存在")
		}
		if hasQualityInspection(*data, batch.ID) {
			return fmt.Errorf("生产批次已创建质检单")
		}
		item = req
		item.ID = nextID(data, "qualityInspection")
		item.InspectionNo = number("QI", item.ID)
		item.BatchID = batch.ID
		item.BatchNo = batch.BatchNo
		item.SiteID = batch.SiteID
		item.ProductID = batch.ProductID
		item.MixDesignID = batch.MixDesignID
		item.Inspector = fallback(req.Inspector, session.User.DisplayName)
		item.Slump = fallback(req.Slump, "180mm")
		item.SampleCount = 2
		item.Result = "pending"
		item.Status = "sampling"
		item.CreatedAt = nowString()
		data.QualityInspections = append(data.QualityInspections, item)
		samples := defaultQualitySamples(data, item, batch.CompletedAt)
		data.QualitySamples = append(data.QualitySamples, samples...)
		data.LaboratorySamples = append(data.LaboratorySamples, qualitySamplesToLaboratorySamples(data, item, samples)...)
		updateBatchQualityStatus(data, batch.ID, "sampling")
		addAudit(data, session.User.Username, "create", "quality_inspection", item.ID, item.InspectionNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "quality.inspection.created")
}

func (a *App) testQualitySample(w http.ResponseWriter, r *http.Request, session Session, sampleID int64) {
	var req QualitySample
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid quality sample")
		return
	}
	var item QualitySample
	err := a.store.Mutate(func(data *AppData) error {
		found := false
		for i := range data.QualitySamples {
			if data.QualitySamples[i].ID != sampleID {
				continue
			}
			data.QualitySamples[i].Strength = req.Strength
			data.QualitySamples[i].Result = fallback(req.Result, qualityResult(req.Strength))
			data.QualitySamples[i].Status = "completed"
			data.QualitySamples[i].TestedAt = fallback(req.TestedAt, nowString())
			data.QualitySamples[i].Remark = fallback(req.Remark, data.QualitySamples[i].Remark)
			item = data.QualitySamples[i]
			found = true
			break
		}
		if !found {
			return fmt.Errorf("试块不存在")
		}
		refreshQualityInspection(data, item.InspectionID)
		updateLaboratorySampleFromQualitySample(data, item)
		if item.Result == "failed" {
			appendQualityException(data, "quality_sample", item.ID, sampleSiteID(*data, item), "试块强度不合格", "试块 "+item.SampleNo+" 试验结果不合格", "high", session.User.DisplayName)
		}
		addAudit(data, session.User.Username, "test", "quality_sample", item.ID, item.SampleNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "quality.sample.tested")
}

func (a *App) createRawMaterialInspection(w http.ResponseWriter, r *http.Request, session Session) {
	var req RawMaterialInspection
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid raw material inspection")
		return
	}
	var item RawMaterialInspection
	err := a.store.Mutate(func(data *AppData) error {
		receipt, ok := findRawMaterialReceipt(*data, req.ReceiptID)
		if !ok {
			return fmt.Errorf("原料入库单不存在")
		}
		if hasRawMaterialInspection(*data, receipt.ID) {
			return fmt.Errorf("原料入库单已创建质检单")
		}
		item = req
		item.ID = nextID(data, "rawInspection")
		item.InspectionNo = number("RQI", item.ID)
		item.ReceiptID = receipt.ID
		item.ReceiptNo = receipt.ReceiptNo
		item.SiteID = receipt.SiteID
		item.SupplierID = receipt.SupplierID
		item.MaterialID = receipt.MaterialID
		item.Inspector = fallback(req.Inspector, session.User.DisplayName)
		item.SampleNo = fallback(req.SampleNo, number("RQS", item.ID))
		item.Result = "pending"
		item.Status = "pending_review"
		item.CreatedAt = nowString()
		data.RawMaterialInspections = append(data.RawMaterialInspections, item)
		data.LaboratorySamples = append(data.LaboratorySamples, rawInspectionToLaboratorySample(data, item))
		updateRawReceiptQualityStatus(data, receipt.ID, "testing", "")
		updateInventoryQualityStatus(data, receipt.SiteID, receipt.MaterialID, "testing", "blocked")
		addAudit(data, session.User.Username, "create", "raw_material_inspection", item.ID, item.InspectionNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "quality.raw_inspection.created")
}

func (a *App) reviewRawMaterialInspection(w http.ResponseWriter, r *http.Request, session Session, inspectionID int64) {
	var req RawMaterialInspection
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid raw material inspection review")
		return
	}
	var item RawMaterialInspection
	err := a.store.Mutate(func(data *AppData) error {
		found := false
		for i := range data.RawMaterialInspections {
			if data.RawMaterialInspections[i].ID != inspectionID {
				continue
			}
			data.RawMaterialInspections[i].Moisture = nonZero(req.Moisture, data.RawMaterialInspections[i].Moisture)
			data.RawMaterialInspections[i].MudContent = nonZero(req.MudContent, data.RawMaterialInspections[i].MudContent)
			data.RawMaterialInspections[i].Fineness = fallback(req.Fineness, data.RawMaterialInspections[i].Fineness)
			data.RawMaterialInspections[i].Result = fallback(req.Result, rawInspectionResult(data.RawMaterialInspections[i]))
			data.RawMaterialInspections[i].Status = "completed"
			data.RawMaterialInspections[i].CompletedAt = nowString()
			data.RawMaterialInspections[i].Remark = fallback(req.Remark, data.RawMaterialInspections[i].Remark)
			item = data.RawMaterialInspections[i]
			found = true
			break
		}
		if !found {
			return fmt.Errorf("原料质检单不存在")
		}
		availableStatus := "available"
		if item.Result == "failed" {
			availableStatus = "blocked"
		}
		updateRawReceiptQualityStatus(data, item.ReceiptID, item.Result, item.Result)
		updateInventoryQualityStatus(data, item.SiteID, item.MaterialID, item.Result, availableStatus)
		updateLaboratorySampleFromRawInspection(data, item)
		if item.Result == "failed" {
			appendQualityException(data, "raw_material_inspection", item.ID, item.SiteID, "原料检验不合格", "原料质检单 "+item.InspectionNo+" 复核为不合格", "high", session.User.DisplayName)
		}
		addAudit(data, session.User.Username, "review", "raw_material_inspection", item.ID, item.InspectionNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "quality.raw_inspection.reviewed")
}

func defaultQualitySamples(data *AppData, inspection QualityInspection, batchCompletedAt string) []QualitySample {
	ages := []int{7, 28}
	items := make([]QualitySample, 0, len(ages))
	for _, age := range ages {
		id := nextID(data, "qualitySample")
		items = append(items, QualitySample{
			ID: id, SampleNo: number("QS", id), InspectionID: inspection.ID, BatchID: inspection.BatchID,
			SampleType: "compressive_strength", AgeDays: age, PlannedTestAt: addDays(batchCompletedAt, age),
			Result: "pending", Status: "pending",
		})
	}
	return items
}

func qualitySamplesToLaboratorySamples(data *AppData, inspection QualityInspection, samples []QualitySample) []LaboratorySample {
	items := make([]LaboratorySample, 0, len(samples))
	for _, sample := range samples {
		id := nextID(data, "labSample")
		items = append(items, LaboratorySample{
			ID: id, SampleNo: number("LS", id), SourceType: "quality_sample", SourceID: sample.ID,
			SiteID: inspection.SiteID, ProductID: inspection.ProductID, MixDesignID: inspection.MixDesignID,
			BatchID: inspection.BatchID, InspectionID: inspection.ID, SampleType: sample.SampleType,
			Status: sample.Status, Result: sample.Result, PlannedTestAt: sample.PlannedTestAt,
			CollectedAt: inspection.CreatedAt, CreatedBy: inspection.Inspector, Remark: sample.SampleNo,
		})
	}
	return items
}

func rawInspectionToLaboratorySample(data *AppData, item RawMaterialInspection) LaboratorySample {
	id := nextID(data, "labSample")
	return LaboratorySample{
		ID: id, SampleNo: fallback(item.SampleNo, number("LS", id)), SourceType: "raw_material_inspection", SourceID: item.ID,
		SiteID: item.SiteID, MaterialID: item.MaterialID, RawInspectionID: item.ID, SampleType: "raw_material",
		Status: "testing", Result: item.Result, CollectedAt: item.CreatedAt, CreatedBy: item.Inspector, Remark: item.Remark,
	}
}

func updateLaboratorySampleFromQualitySample(data *AppData, sample QualitySample) {
	for i := range data.LaboratorySamples {
		if data.LaboratorySamples[i].SourceType == "quality_sample" && data.LaboratorySamples[i].SourceID == sample.ID {
			data.LaboratorySamples[i].Status = sample.Status
			data.LaboratorySamples[i].Result = sample.Result
			data.LaboratorySamples[i].Remark = fallback(sample.Remark, data.LaboratorySamples[i].Remark)
			return
		}
	}
}

func updateLaboratorySampleFromRawInspection(data *AppData, item RawMaterialInspection) {
	for i := range data.LaboratorySamples {
		if data.LaboratorySamples[i].SourceType == "raw_material_inspection" && data.LaboratorySamples[i].SourceID == item.ID {
			data.LaboratorySamples[i].Status = item.Status
			data.LaboratorySamples[i].Result = item.Result
			data.LaboratorySamples[i].Remark = fallback(item.Remark, data.LaboratorySamples[i].Remark)
			return
		}
	}
}

func sampleSiteID(data AppData, sample QualitySample) int64 {
	for _, inspection := range data.QualityInspections {
		if inspection.ID == sample.InspectionID {
			return inspection.SiteID
		}
	}
	return 0
}

func findProductionBatch(data AppData, id int64) (ProductionBatch, bool) {
	for _, item := range data.ProductionBatches {
		if item.ID == id {
			return item, true
		}
	}
	return ProductionBatch{}, false
}

func hasQualityInspection(data AppData, batchID int64) bool {
	for _, item := range data.QualityInspections {
		if item.BatchID == batchID {
			return true
		}
	}
	return false
}

func findRawMaterialReceipt(data AppData, id int64) (RawMaterialReceipt, bool) {
	for _, item := range data.RawMaterialReceipts {
		if item.ID == id {
			return item, true
		}
	}
	return RawMaterialReceipt{}, false
}

func hasRawMaterialInspection(data AppData, receiptID int64) bool {
	for _, item := range data.RawMaterialInspections {
		if item.ReceiptID == receiptID {
			return true
		}
	}
	return false
}

func rawInspectionResult(item RawMaterialInspection) string {
	if item.MudContent > 5 {
		return "failed"
	}
	return "passed"
}

func qualityResult(strength float64) string {
	if strength > 0 && strength < 30 {
		return "failed"
	}
	return "passed"
}

func refreshQualityInspection(data *AppData, inspectionID int64) {
	total := 0
	completed := 0
	failed := false
	for _, sample := range data.QualitySamples {
		if sample.InspectionID != inspectionID {
			continue
		}
		total++
		if sample.Status == "completed" {
			completed++
			if sample.Result == "failed" {
				failed = true
			}
		}
	}
	for i := range data.QualityInspections {
		if data.QualityInspections[i].ID != inspectionID {
			continue
		}
		if completed == 0 {
			data.QualityInspections[i].Status = "sampling"
			data.QualityInspections[i].Result = "pending"
			updateBatchQualityStatus(data, data.QualityInspections[i].BatchID, "sampling")
			return
		}
		if completed < total {
			data.QualityInspections[i].Status = "testing"
			data.QualityInspections[i].Result = "pending"
			updateBatchQualityStatus(data, data.QualityInspections[i].BatchID, "testing")
			return
		}
		data.QualityInspections[i].Status = "completed"
		data.QualityInspections[i].CompletedAt = nowString()
		if failed {
			data.QualityInspections[i].Result = "failed"
			updateBatchQualityStatus(data, data.QualityInspections[i].BatchID, "failed")
		} else {
			data.QualityInspections[i].Result = "passed"
			updateBatchQualityStatus(data, data.QualityInspections[i].BatchID, "passed")
		}
		return
	}
}

func updateBatchQualityStatus(data *AppData, batchID int64, status string) {
	for i := range data.ProductionBatches {
		if data.ProductionBatches[i].ID == batchID {
			data.ProductionBatches[i].QualityStatus = status
			return
		}
	}
}

func updateRawReceiptQualityStatus(data *AppData, receiptID int64, qualityStatus, receiptStatus string) {
	for i := range data.RawMaterialReceipts {
		if data.RawMaterialReceipts[i].ID != receiptID {
			continue
		}
		data.RawMaterialReceipts[i].QualityStatus = qualityStatus
		if receiptStatus == "failed" {
			data.RawMaterialReceipts[i].Status = "rejected"
		}
		if receiptStatus == "passed" && data.RawMaterialReceipts[i].Status == "rejected" {
			data.RawMaterialReceipts[i].Status = "stocked"
		}
		return
	}
}

func updateInventoryQualityStatus(data *AppData, siteID, materialID int64, qualityStatus, availableStatus string) {
	for i := range data.Inventory {
		if data.Inventory[i].SiteID == siteID && data.Inventory[i].MaterialID == materialID {
			data.Inventory[i].QualityStatus = qualityStatus
			data.Inventory[i].AvailableStatus = availableStatus
			data.Inventory[i].UpdatedAt = nowString()
		}
	}
}

func addDays(value string, days int) string {
	base, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		base = time.Now()
	}
	return base.AddDate(0, 0, days).Format("2006-01-02")
}
