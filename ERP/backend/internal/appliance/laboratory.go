package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *App) laboratory(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 || parts[0] == "overview" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, laboratoryPayload(data))
		return
	}
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		switch parts[0] {
		case "mix-designs":
			writeJSON(w, http.StatusOK, data.MixDesigns)
		case "trial-runs":
			writeJSON(w, http.StatusOK, data.MixDesignTrialRuns)
		case "samples":
			writeJSON(w, http.StatusOK, data.LaboratorySamples)
		case "tests":
			writeJSON(w, http.StatusOK, data.LaboratoryTests)
		case "equipment":
			writeJSON(w, http.StatusOK, data.LaboratoryEquipment)
		case "calibrations":
			writeJSON(w, http.StatusOK, data.LaboratoryCalibrations)
		case "exceptions":
			writeJSON(w, http.StatusOK, data.QualityExceptions)
		default:
			writeError(w, http.StatusNotFound, "unknown laboratory resource")
		}
		return
	}
	if len(parts) == 1 && parts[0] == "mix-designs" && r.Method == http.MethodPost {
		a.createLaboratoryMixDesign(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "mix-designs" && parts[2] == "revise" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.reviseLaboratoryMixDesign(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "mix-designs" && parts[2] == "approve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.approveLaboratoryMixDesign(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "mix-designs" && parts[2] == "retire" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.retireLaboratoryMixDesign(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "mix-designs" && parts[2] == "trial-runs" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.createMixDesignTrialRun(w, r, session, id)
		return
	}
	if len(parts) == 1 && parts[0] == "samples" && r.Method == http.MethodPost {
		a.createLaboratorySample(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "samples" && parts[2] == "tests" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.createLaboratoryTest(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "tests" && parts[2] == "review" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.reviewLaboratoryTest(w, r, session, id)
		return
	}
	if len(parts) == 1 && parts[0] == "equipment" && r.Method == http.MethodPost {
		a.createLaboratoryEquipment(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "equipment" && parts[2] == "calibrations" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.createLaboratoryCalibration(w, r, session, id)
		return
	}
	if len(parts) == 1 && parts[0] == "exceptions" && r.Method == http.MethodPost {
		a.createLaboratoryException(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "exceptions" && parts[2] == "handle" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.handleLaboratoryException(w, r, session, id)
		return
	}
	writeError(w, http.StatusNotFound, "unknown laboratory route")
}

func laboratoryPayload(data AppData) map[string]interface{} {
	return map[string]interface{}{
		"kpis":               buildLaboratoryKPI(data),
		"mixDesigns":         data.MixDesigns,
		"trialRuns":          data.MixDesignTrialRuns,
		"qualityInspections": data.QualityInspections,
		"qualitySamples":     data.QualitySamples,
		"rawInspections":     data.RawMaterialInspections,
		"samples":            data.LaboratorySamples,
		"tests":              data.LaboratoryTests,
		"equipment":          data.LaboratoryEquipment,
		"calibrations":       data.LaboratoryCalibrations,
		"exceptions":         data.QualityExceptions,
		"batches":            data.ProductionBatches,
		"receipts":           data.RawMaterialReceipts,
		"products":           data.Products,
		"materials":          data.Materials,
		"sites":              data.Sites,
	}
}

func (a *App) createLaboratoryMixDesign(w http.ResponseWriter, r *http.Request, session Session) {
	var req MixDesign
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design")
		return
	}
	var item MixDesign
	err := a.store.Mutate(func(data *AppData) error {
		product, ok := findProduct(*data, req.ProductID)
		if !ok {
			return fmt.Errorf("产品不存在")
		}
		if len(req.Materials) == 0 {
			return fmt.Errorf("配比材料不能为空")
		}
		siteID, err := laboratorySiteID(*data, session, req.SiteID)
		if err != nil {
			return err
		}
		item = req
		item.ID = nextID(data, "mixDesign")
		item.SiteID = siteID
		item.Code = fallback(req.Code, number("MD", item.ID))
		item.Version = fallback(req.Version, "v1")
		item.StrengthGrade = fallback(req.StrengthGrade, product.Spec)
		item.Slump = fallback(req.Slump, "180mm")
		item.Scope = fallback(req.Scope, "站内标准生产配比")
		item.Status = fallback(req.Status, "draft")
		item.IsCurrent = false
		item.CreatedBy = fallback(req.CreatedBy, session.User.DisplayName)
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		normalizeMixMaterials(&item)
		data.MixDesigns = append(data.MixDesigns, item)
		addAudit(data, session.User.Username, "create", "mix_design", item.ID, item.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.mix_design.created")
}

func (a *App) reviseLaboratoryMixDesign(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req MixDesign
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design revision")
		return
	}
	var item MixDesign
	err := a.store.Mutate(func(data *AppData) error {
		base, ok := findMixDesign(*data, id)
		if !ok {
			return fmt.Errorf("配比不存在")
		}
		if session.User.SiteID > 0 && base.SiteID != 0 && base.SiteID != session.User.SiteID {
			return fmt.Errorf("无权修订其他站点配比")
		}
		item = base
		item.ID = nextID(data, "mixDesign")
		item.ParentID = base.ID
		item.Code = fallback(req.Code, base.Code)
		item.Version = fallback(req.Version, nextMixVersion(base.Version))
		item.StrengthGrade = fallback(req.StrengthGrade, base.StrengthGrade)
		item.Slump = fallback(req.Slump, base.Slump)
		item.Scope = fallback(req.Scope, base.Scope)
		item.Status = "draft"
		item.IsCurrent = false
		item.EffectiveFrom = req.EffectiveFrom
		item.EffectiveTo = req.EffectiveTo
		item.ApprovedBy = ""
		item.ApprovedAt = ""
		item.RetiredAt = ""
		item.CreatedBy = fallback(req.CreatedBy, session.User.DisplayName)
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		if len(req.Materials) > 0 {
			item.Materials = req.Materials
		}
		normalizeMixMaterials(&item)
		data.MixDesigns = append(data.MixDesigns, item)
		addAudit(data, session.User.Username, "revise", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.mix_design.revised")
}

func (a *App) approveLaboratoryMixDesign(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		EffectiveFrom string `json:"effectiveFrom"`
		EffectiveTo   string `json:"effectiveTo"`
		TrialRunID    int64  `json:"trialRunId"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design approval")
		return
	}
	var item MixDesign
	err := a.store.Mutate(func(data *AppData) error {
		idx := mixDesignIndex(*data, id)
		if idx < 0 {
			return fmt.Errorf("配比不存在")
		}
		if session.User.SiteID > 0 && data.MixDesigns[idx].SiteID != 0 && data.MixDesigns[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权审批其他站点配比")
		}
		if len(data.MixDesigns[idx].Materials) == 0 {
			return fmt.Errorf("配比材料不能为空")
		}
		if req.TrialRunID > 0 {
			trial, ok := findMixDesignTrialRun(*data, req.TrialRunID)
			if !ok || trial.MixDesignID != id {
				return fmt.Errorf("试配记录不存在")
			}
			if trial.Result != "passed" {
				return fmt.Errorf("试配未合格不能审批配比")
			}
		}
		now := nowString()
		data.MixDesigns[idx].Status = "approved"
		data.MixDesigns[idx].IsCurrent = true
		data.MixDesigns[idx].EffectiveFrom = fallback(req.EffectiveFrom, todayString())
		data.MixDesigns[idx].EffectiveTo = fallback(req.EffectiveTo, data.MixDesigns[idx].EffectiveTo)
		data.MixDesigns[idx].ApprovedBy = session.User.DisplayName
		data.MixDesigns[idx].ApprovedAt = now
		data.MixDesigns[idx].UpdatedAt = now
		item = data.MixDesigns[idx]
		retirePreviousCurrentMixDesigns(data, item.ID, item.ProductID, item.SiteID, now)
		addAudit(data, session.User.Username, "approve", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.mix_design.approved")
}

func (a *App) retireLaboratoryMixDesign(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item MixDesign
	err := a.store.Mutate(func(data *AppData) error {
		idx := mixDesignIndex(*data, id)
		if idx < 0 {
			return fmt.Errorf("配比不存在")
		}
		if session.User.SiteID > 0 && data.MixDesigns[idx].SiteID != 0 && data.MixDesigns[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权停用其他站点配比")
		}
		data.MixDesigns[idx].Status = "retired"
		data.MixDesigns[idx].IsCurrent = false
		data.MixDesigns[idx].RetiredAt = nowString()
		data.MixDesigns[idx].UpdatedAt = data.MixDesigns[idx].RetiredAt
		item = data.MixDesigns[idx]
		addAudit(data, session.User.Username, "retire", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.mix_design.retired")
}

func (a *App) createMixDesignTrialRun(w http.ResponseWriter, r *http.Request, session Session, mixDesignID int64) {
	var req MixDesignTrialRun
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design trial")
		return
	}
	var item MixDesignTrialRun
	err := a.store.Mutate(func(data *AppData) error {
		mix, ok := findMixDesign(*data, mixDesignID)
		if !ok {
			return fmt.Errorf("配比不存在")
		}
		if session.User.SiteID > 0 && mix.SiteID != 0 && mix.SiteID != session.User.SiteID {
			return fmt.Errorf("无权操作其他站点配比")
		}
		item = req
		item.ID = nextID(data, "mixTrial")
		item.TrialNo = number("MTR", item.ID)
		item.MixDesignID = mix.ID
		item.ProductID = mix.ProductID
		item.SiteID = mix.SiteID
		item.TargetStrength = fallback(req.TargetStrength, mix.StrengthGrade)
		item.Slump = fallback(req.Slump, mix.Slump)
		item.Result = fallback(req.Result, trialResult(req))
		item.Conclusion = fallback(req.Conclusion, trialConclusion(item))
		item.Tester = fallback(req.Tester, session.User.DisplayName)
		item.TestedAt = fallback(req.TestedAt, nowString())
		item.CreatedAt = nowString()
		data.MixDesignTrialRuns = append(data.MixDesignTrialRuns, item)
		addAudit(data, session.User.Username, "create", "mix_design_trial", item.ID, item.TrialNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.trial.created")
}

func (a *App) createLaboratorySample(w http.ResponseWriter, r *http.Request, session Session) {
	var req LaboratorySample
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid laboratory sample")
		return
	}
	var item LaboratorySample
	err := a.store.Mutate(func(data *AppData) error {
		siteID, err := laboratorySiteID(*data, session, req.SiteID)
		if err != nil {
			return err
		}
		item = req
		item.ID = nextID(data, "labSample")
		item.SampleNo = fallback(req.SampleNo, number("LS", item.ID))
		item.SourceType = fallback(req.SourceType, "manual")
		item.SiteID = siteID
		item.SampleType = fallback(req.SampleType, "manual")
		item.Status = fallback(req.Status, "pending")
		item.Result = fallback(req.Result, "pending")
		item.CollectedAt = fallback(req.CollectedAt, nowString())
		item.CreatedBy = fallback(req.CreatedBy, session.User.DisplayName)
		data.LaboratorySamples = append(data.LaboratorySamples, item)
		addAudit(data, session.User.Username, "create", "laboratory_sample", item.ID, item.SampleNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.sample.created")
}

func (a *App) createLaboratoryTest(w http.ResponseWriter, r *http.Request, session Session, sampleID int64) {
	var req LaboratoryTestRecord
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid laboratory test")
		return
	}
	var item LaboratoryTestRecord
	err := a.store.Mutate(func(data *AppData) error {
		sample, ok := findLaboratorySample(*data, sampleID)
		if !ok {
			return fmt.Errorf("样品不存在")
		}
		if session.User.SiteID > 0 && sample.SiteID != session.User.SiteID {
			return fmt.Errorf("无权试验其他站点样品")
		}
		equipment, ok := findLaboratoryEquipment(*data, req.EquipmentID)
		if !ok {
			return fmt.Errorf("试验仪器不存在")
		}
		if err := validateLaboratoryEquipment(equipment); err != nil {
			return err
		}
		item = req
		item.ID = nextID(data, "labTest")
		item.TestNo = fallback(req.TestNo, number("LT", item.ID))
		item.SampleID = sample.ID
		item.SiteID = sample.SiteID
		item.TestType = fallback(req.TestType, sample.SampleType)
		item.Metric = fallback(req.Metric, "strength")
		item.Unit = fallback(req.Unit, "MPa")
		item.Result = fallback(req.Result, laboratoryTestResult(req.Value))
		item.Status = "pending_review"
		item.Tester = fallback(req.Tester, session.User.DisplayName)
		item.TestedAt = fallback(req.TestedAt, nowString())
		data.LaboratoryTests = append(data.LaboratoryTests, item)
		updateLaboratorySample(data, sample.ID, "testing", "pending")
		addAudit(data, session.User.Username, "create", "laboratory_test", item.ID, item.TestNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.test.created")
}

func (a *App) reviewLaboratoryTest(w http.ResponseWriter, r *http.Request, session Session, testID int64) {
	var req LaboratoryTestRecord
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid laboratory test review")
		return
	}
	var item LaboratoryTestRecord
	err := a.store.Mutate(func(data *AppData) error {
		idx := laboratoryTestIndex(*data, testID)
		if idx < 0 {
			return fmt.Errorf("试验记录不存在")
		}
		if session.User.SiteID > 0 && data.LaboratoryTests[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权复核其他站点试验")
		}
		if req.Result != "" {
			data.LaboratoryTests[idx].Result = req.Result
		}
		if req.Value > 0 {
			data.LaboratoryTests[idx].Value = req.Value
		}
		data.LaboratoryTests[idx].Status = "reviewed"
		data.LaboratoryTests[idx].Reviewer = fallback(req.Reviewer, session.User.DisplayName)
		data.LaboratoryTests[idx].ReviewedAt = nowString()
		data.LaboratoryTests[idx].Remark = fallback(req.Remark, data.LaboratoryTests[idx].Remark)
		item = data.LaboratoryTests[idx]
		updateLaboratorySample(data, item.SampleID, "completed", item.Result)
		if item.Result == "failed" {
			appendQualityException(data, "laboratory_test", item.ID, item.SiteID, "试验结果不合格", "试验 "+item.TestNo+" 复核为不合格", "high", session.User.DisplayName)
		}
		addAudit(data, session.User.Username, "review", "laboratory_test", item.ID, item.TestNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.test.reviewed")
}

func (a *App) createLaboratoryEquipment(w http.ResponseWriter, r *http.Request, session Session) {
	var req LaboratoryEquipment
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid laboratory equipment")
		return
	}
	var item LaboratoryEquipment
	err := a.store.Mutate(func(data *AppData) error {
		siteID, err := laboratorySiteID(*data, session, req.SiteID)
		if err != nil {
			return err
		}
		item = req
		item.ID = nextID(data, "labEquipment")
		item.EquipmentNo = fallback(req.EquipmentNo, number("EQ", item.ID))
		item.Name = fallback(req.Name, "实验仪器")
		item.SiteID = siteID
		item.Status = fallback(req.Status, "active")
		if item.CalibrationCycleDays == 0 {
			item.CalibrationCycleDays = 180
		}
		item.CreatedAt = fallback(req.CreatedAt, nowString())
		if item.NextCalibrationAt == "" && item.LastCalibrationAt != "" {
			item.NextCalibrationAt = labAddDays(item.LastCalibrationAt, item.CalibrationCycleDays)
		}
		data.LaboratoryEquipment = append(data.LaboratoryEquipment, item)
		addAudit(data, session.User.Username, "create", "laboratory_equipment", item.ID, item.EquipmentNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.equipment.saved")
}

func (a *App) createLaboratoryCalibration(w http.ResponseWriter, r *http.Request, session Session, equipmentID int64) {
	var req LaboratoryCalibration
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid laboratory calibration")
		return
	}
	var item LaboratoryCalibration
	err := a.store.Mutate(func(data *AppData) error {
		idx := laboratoryEquipmentIndex(*data, equipmentID)
		if idx < 0 {
			return fmt.Errorf("仪器不存在")
		}
		if session.User.SiteID > 0 && data.LaboratoryEquipment[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权校准其他站点仪器")
		}
		item = req
		item.ID = nextID(data, "labCalibration")
		item.CalibrationNo = fallback(req.CalibrationNo, number("LC", item.ID))
		item.EquipmentID = data.LaboratoryEquipment[idx].ID
		item.SiteID = data.LaboratoryEquipment[idx].SiteID
		item.Result = fallback(req.Result, "passed")
		item.CalibratedAt = fallback(req.CalibratedAt, todayString())
		item.NextDueAt = fallback(req.NextDueAt, labAddDays(item.CalibratedAt, data.LaboratoryEquipment[idx].CalibrationCycleDays))
		item.Operator = fallback(req.Operator, session.User.DisplayName)
		data.LaboratoryCalibrations = append(data.LaboratoryCalibrations, item)
		data.LaboratoryEquipment[idx].LastCalibrationAt = item.CalibratedAt
		data.LaboratoryEquipment[idx].NextCalibrationAt = item.NextDueAt
		if item.Result == "passed" && data.LaboratoryEquipment[idx].Status == "calibration_due" {
			data.LaboratoryEquipment[idx].Status = "active"
		}
		if item.Result == "failed" {
			data.LaboratoryEquipment[idx].Status = "disabled"
		}
		addAudit(data, session.User.Username, "calibrate", "laboratory_equipment", item.EquipmentID, item.CalibrationNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.calibration.created")
}

func (a *App) createLaboratoryException(w http.ResponseWriter, r *http.Request, session Session) {
	var req QualityException
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid quality exception")
		return
	}
	var item QualityException
	err := a.store.Mutate(func(data *AppData) error {
		siteID, err := laboratorySiteID(*data, session, req.SiteID)
		if err != nil {
			return err
		}
		item = req
		item.ID = nextID(data, "qualityException")
		item.ExceptionNo = fallback(req.ExceptionNo, number("QE", item.ID))
		item.SiteID = siteID
		item.SourceType = fallback(req.SourceType, "manual")
		item.Severity = fallback(req.Severity, "medium")
		item.Status = fallback(req.Status, "open")
		item.Responsible = fallback(req.Responsible, session.User.DisplayName)
		item.CreatedAt = fallback(req.CreatedAt, nowString())
		data.QualityExceptions = append(data.QualityExceptions, item)
		addAudit(data, session.User.Username, "create", "quality_exception", item.ID, item.ExceptionNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.exception.created")
}

func (a *App) handleLaboratoryException(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req QualityException
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid quality exception handling")
		return
	}
	var item QualityException
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.QualityExceptions {
			if data.QualityExceptions[i].ID != id {
				continue
			}
			if session.User.SiteID > 0 && data.QualityExceptions[i].SiteID != session.User.SiteID {
				return fmt.Errorf("无权关闭其他站点异常")
			}
			data.QualityExceptions[i].RootCause = fallback(req.RootCause, data.QualityExceptions[i].RootCause)
			data.QualityExceptions[i].CorrectiveAction = fallback(req.CorrectiveAction, data.QualityExceptions[i].CorrectiveAction)
			data.QualityExceptions[i].Responsible = fallback(req.Responsible, data.QualityExceptions[i].Responsible)
			data.QualityExceptions[i].Status = "closed"
			data.QualityExceptions[i].HandledAt = nowString()
			data.QualityExceptions[i].ClosedBy = session.User.DisplayName
			item = data.QualityExceptions[i]
			addAudit(data, session.User.Username, "handle", "quality_exception", item.ID, item.ExceptionNo, clientIP(r))
			return nil
		}
		return fmt.Errorf("质量异常不存在")
	})
	a.respondMutation(w, err, item, "laboratory.exception.handled")
}

func buildLaboratoryKPI(data AppData) LaboratoryKPI {
	kpi := LaboratoryKPI{
		MixDesigns: len(data.MixDesigns), TrialRuns: len(data.MixDesignTrialRuns), Samples: len(data.LaboratorySamples),
		Tests: len(data.LaboratoryTests), Equipments: len(data.LaboratoryEquipment),
	}
	passedTests := 0
	for _, item := range data.MixDesigns {
		if item.IsCurrent && item.Status == "approved" {
			kpi.CurrentMixDesigns++
		}
		if item.Status == "draft" || item.Status == "pending_approval" {
			kpi.PendingMixDesigns++
		}
	}
	for _, item := range data.LaboratorySamples {
		if item.Status != "completed" {
			kpi.PendingSamples++
		}
	}
	for _, item := range data.LaboratoryTests {
		if item.Status == "pending_review" {
			kpi.PendingReviews++
		}
		if item.Result == "passed" {
			passedTests++
		}
	}
	for _, item := range data.LaboratoryEquipment {
		if item.Status == "disabled" || item.Status == "retired" {
			continue
		}
		days := daysUntil(item.NextCalibrationAt)
		if days < 0 {
			kpi.CalibrationOverdue++
		} else if days <= 30 {
			kpi.CalibrationDue++
		}
	}
	for _, item := range data.QualityExceptions {
		if item.Status != "closed" {
			kpi.OpenExceptions++
		}
	}
	kpi.PassRate = moneyPercent(float64(passedTests), float64(kpi.Tests))
	return kpi
}

func laboratorySiteID(data AppData, session Session, requested int64) (int64, error) {
	if session.User.SiteID > 0 {
		if requested > 0 && requested != session.User.SiteID {
			return 0, fmt.Errorf("无权操作其他站点数据")
		}
		return session.User.SiteID, nil
	}
	if requested > 0 {
		return requested, nil
	}
	if len(data.Sites) > 0 {
		return data.Sites[0].ID, nil
	}
	return 0, fmt.Errorf("站点不存在")
}

func normalizeMixMaterials(item *MixDesign) {
	for i := range item.Materials {
		if item.Materials[i].Unit == "" {
			item.Materials[i].Unit = "kg/m3"
		}
	}
}

func nextMixVersion(value string) string {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "v")
	n, err := strconv.Atoi(trimmed)
	if err != nil {
		return value + "-rev"
	}
	return fmt.Sprintf("v%d", n+1)
}

func mixDesignIndex(data AppData, id int64) int {
	for i := range data.MixDesigns {
		if data.MixDesigns[i].ID == id {
			return i
		}
	}
	return -1
}

func retirePreviousCurrentMixDesigns(data *AppData, currentID, productID, siteID int64, retiredAt string) {
	for i := range data.MixDesigns {
		if data.MixDesigns[i].ID == currentID || data.MixDesigns[i].ProductID != productID {
			continue
		}
		if !mixDesignMatchesSite(data.MixDesigns[i], siteID) {
			continue
		}
		if data.MixDesigns[i].IsCurrent && data.MixDesigns[i].Status == "approved" {
			data.MixDesigns[i].IsCurrent = false
			data.MixDesigns[i].Status = "retired"
			data.MixDesigns[i].RetiredAt = retiredAt
			data.MixDesigns[i].UpdatedAt = retiredAt
		}
	}
}

func findMixDesignTrialRun(data AppData, id int64) (MixDesignTrialRun, bool) {
	for _, item := range data.MixDesignTrialRuns {
		if item.ID == id {
			return item, true
		}
	}
	return MixDesignTrialRun{}, false
}

func trialResult(item MixDesignTrialRun) string {
	if item.Strength28d > 0 && item.Strength28d < 30 {
		return "failed"
	}
	if item.Strength7d > 0 || item.Strength28d > 0 {
		return "passed"
	}
	return "pending"
}

func trialConclusion(item MixDesignTrialRun) string {
	if item.Result == "failed" {
		return "试配指标未满足要求"
	}
	if item.Result == "passed" {
		return "试配指标满足生产放行要求"
	}
	return "等待试验结果"
}

func findLaboratorySample(data AppData, id int64) (LaboratorySample, bool) {
	for _, item := range data.LaboratorySamples {
		if item.ID == id {
			return item, true
		}
	}
	return LaboratorySample{}, false
}

func updateLaboratorySample(data *AppData, id int64, status, result string) {
	for i := range data.LaboratorySamples {
		if data.LaboratorySamples[i].ID == id {
			data.LaboratorySamples[i].Status = status
			data.LaboratorySamples[i].Result = result
			return
		}
	}
}

func laboratoryTestIndex(data AppData, id int64) int {
	for i := range data.LaboratoryTests {
		if data.LaboratoryTests[i].ID == id {
			return i
		}
	}
	return -1
}

func laboratoryEquipmentIndex(data AppData, id int64) int {
	for i := range data.LaboratoryEquipment {
		if data.LaboratoryEquipment[i].ID == id {
			return i
		}
	}
	return -1
}

func findLaboratoryEquipment(data AppData, id int64) (LaboratoryEquipment, bool) {
	for _, item := range data.LaboratoryEquipment {
		if item.ID == id {
			return item, true
		}
	}
	return LaboratoryEquipment{}, false
}

func validateLaboratoryEquipment(item LaboratoryEquipment) error {
	if item.Status != "active" {
		return fmt.Errorf("仪器不可用")
	}
	if item.NextCalibrationAt != "" && daysUntil(item.NextCalibrationAt) < 0 {
		return fmt.Errorf("仪器校准已过期")
	}
	return nil
}

func laboratoryTestResult(value float64) string {
	if value > 0 && value < 30 {
		return "failed"
	}
	return "passed"
}

func appendQualityException(data *AppData, sourceType string, sourceID, siteID int64, title, description, severity, responsible string) QualityException {
	for _, item := range data.QualityExceptions {
		if item.SourceType == sourceType && item.SourceID == sourceID && item.Status != "closed" {
			return item
		}
	}
	id := nextID(data, "qualityException")
	item := QualityException{
		ID: id, ExceptionNo: number("QE", id), SourceType: sourceType, SourceID: sourceID, SiteID: siteID,
		Severity: fallback(severity, "medium"), Title: title, Description: description, Status: "open",
		Responsible: responsible, CreatedAt: nowString(),
	}
	data.QualityExceptions = append(data.QualityExceptions, item)
	return item
}

func labAddDays(value string, days int) string {
	base, err := parseLabDate(value)
	if err != nil {
		base = time.Now()
	}
	return base.AddDate(0, 0, days).Format("2006-01-02")
}

func daysUntil(value string) int {
	if value == "" {
		return 9999
	}
	target, err := parseLabDate(value)
	if err != nil {
		return 9999
	}
	today, _ := time.ParseInLocation("2006-01-02", todayString(), time.Local)
	return int(target.Sub(today).Hours() / 24)
}

func parseLabDate(value string) (time.Time, error) {
	if len(value) >= len("2006-01-02 15:04:05") {
		if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value[:len("2006-01-02 15:04:05")], time.Local); err == nil {
			return parsed, nil
		}
	}
	if len(value) >= len("2006-01-02") {
		return time.ParseInLocation("2006-01-02", value[:len("2006-01-02")], time.Local)
	}
	return time.Time{}, fmt.Errorf("invalid date")
}
