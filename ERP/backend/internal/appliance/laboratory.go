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
		case "mix-design-plant-profiles":
			writeJSON(w, http.StatusOK, data.MixDesignPlantProfiles)
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
	if len(parts) == 3 && parts[0] == "mix-designs" && parts[2] == "plant-profiles" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.createMixDesignPlantProfile(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "mix-design-plant-profiles" && parts[2] == "approve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.approveMixDesignPlantProfile(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "mix-design-plant-profiles" && parts[2] == "retire" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.retireMixDesignPlantProfile(w, r, session, id)
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
		"kpis":                   buildLaboratoryKPI(data),
		"mixDesigns":             listOrEmpty(data.MixDesigns),
		"mixDesignPlantProfiles": listOrEmpty(data.MixDesignPlantProfiles),
		"trialRuns":              listOrEmpty(data.MixDesignTrialRuns),
		"qualityInspections":     listOrEmpty(data.QualityInspections),
		"qualitySamples":         listOrEmpty(data.QualitySamples),
		"rawInspections":         listOrEmpty(data.RawMaterialInspections),
		"samples":                listOrEmpty(data.LaboratorySamples),
		"tests":                  listOrEmpty(data.LaboratoryTests),
		"equipment":              listOrEmpty(data.LaboratoryEquipment),
		"calibrations":           listOrEmpty(data.LaboratoryCalibrations),
		"exceptions":             listOrEmpty(data.QualityExceptions),
		"batches":                listOrEmpty(data.ProductionBatches),
		"receipts":               listOrEmpty(data.RawMaterialReceipts),
		"products":               listOrEmpty(data.Products),
		"materials":              listOrEmpty(data.Materials),
		"sites":                  listOrEmpty(data.Sites),
		"plants":                 listOrEmpty(data.Plants),
		"plantBufferLocations":   listOrEmpty(data.PlantBufferLocations),
		"dictionaries":           activeDataDictionaries(data.DataDictionaries),
	}
}

func listOrEmpty[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
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
		item.Slump = fallback(req.Slump, "油石比 5.1%")
		item.Scope = fallback(req.Scope, "站内标准生产配比")
		item.Status = fallback(req.Status, "draft")
		item.IsCurrent = false
		item.CreatedBy = fallback(req.CreatedBy, session.User.DisplayName)
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		if err := validateMixMaterials(*data, &item); err != nil {
			return err
		}
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
		if err := validateMixMaterials(*data, &item); err != nil {
			return err
		}
		data.MixDesigns = append(data.MixDesigns, item)
		addAudit(data, session.User.Username, "revise", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.mix_design.revised")
}

type mixDesignApprovalRequest struct {
	EffectiveFrom string `json:"effectiveFrom"`
	EffectiveTo   string `json:"effectiveTo"`
	TrialRunID    int64  `json:"trialRunId"`
}

func (a *App) approveLaboratoryMixDesign(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req mixDesignApprovalRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design approval")
		return
	}
	var item MixDesign
	topic := "laboratory.mix_design.approved"
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
		if err := validateMixMaterials(*data, &data.MixDesigns[idx]); err != nil {
			return err
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
		event, instances, err := publishMixDesignApprovalWorkflow(data, data.MixDesigns[idx], req, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			data.MixDesigns[idx].Status = "pending_approval"
			data.MixDesigns[idx].UpdatedAt = nowString()
			item = data.MixDesigns[idx]
			topic = "laboratory.mix_design.workflow_requested"
			addAudit(data, session.User.Username, "request_approve", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
			_ = event
			return nil
		}
		next, err := applyMixDesignApprovalLocked(data, id, req, session.User.DisplayName)
		if err != nil {
			return err
		}
		item = next
		addAudit(data, session.User.Username, "approve", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishMixDesignApprovalWorkflow(data *AppData, item MixDesign, req mixDesignApprovalRequest, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "mix_design.submitted",
		Source:     "laboratory",
		Resource:   "mix_design",
		ResourceID: item.ID,
		ResourceNo: item.Code + " " + item.Version,
		Title:      "生产配比审批",
		Actor:      actor,
		Reason:     fallback(item.Scope, "生产配比审批"),
		Variables: map[string]string{
			"workflowAction": "approve",
			"productId":      fmt.Sprintf("%d", item.ProductID),
			"siteId":         fmt.Sprintf("%d", item.SiteID),
			"parentId":       fmt.Sprintf("%d", item.ParentID),
			"code":           item.Code,
			"version":        item.Version,
			"strengthGrade":  item.StrengthGrade,
			"trialRunId":     fmt.Sprintf("%d", req.TrialRunID),
			"effectiveFrom":  req.EffectiveFrom,
			"effectiveTo":    req.EffectiveTo,
		},
	})
}

func applyMixDesignApprovalLocked(data *AppData, id int64, req mixDesignApprovalRequest, actor string) (MixDesign, error) {
	idx := mixDesignIndex(*data, id)
	if idx < 0 {
		return MixDesign{}, fmt.Errorf("配比不存在")
	}
	if len(data.MixDesigns[idx].Materials) == 0 {
		return MixDesign{}, fmt.Errorf("配比材料不能为空")
	}
	if err := validateMixMaterials(*data, &data.MixDesigns[idx]); err != nil {
		return MixDesign{}, err
	}
	if req.TrialRunID > 0 {
		trial, ok := findMixDesignTrialRun(*data, req.TrialRunID)
		if !ok || trial.MixDesignID != id {
			return MixDesign{}, fmt.Errorf("试配记录不存在")
		}
		if trial.Result != "passed" {
			return MixDesign{}, fmt.Errorf("试配未合格不能审批配比")
		}
	}
	now := nowString()
	data.MixDesigns[idx].Status = "approved"
	data.MixDesigns[idx].IsCurrent = true
	data.MixDesigns[idx].EffectiveFrom = fallback(req.EffectiveFrom, todayString())
	data.MixDesigns[idx].EffectiveTo = fallback(req.EffectiveTo, data.MixDesigns[idx].EffectiveTo)
	data.MixDesigns[idx].ApprovedBy = actor
	data.MixDesigns[idx].ApprovedAt = now
	data.MixDesigns[idx].UpdatedAt = now
	item := data.MixDesigns[idx]
	retirePreviousCurrentMixDesigns(data, item.ID, item.ProductID, item.SiteID, now)
	return item, nil
}

func (a *App) retireLaboratoryMixDesign(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item MixDesign
	topic := "laboratory.mix_design.retired"
	err := a.store.Mutate(func(data *AppData) error {
		idx := mixDesignIndex(*data, id)
		if idx < 0 {
			return fmt.Errorf("配比不存在")
		}
		if session.User.SiteID > 0 && data.MixDesigns[idx].SiteID != 0 && data.MixDesigns[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权停用其他站点配比")
		}
		if data.MixDesigns[idx].Status == "retired" {
			item = data.MixDesigns[idx]
			return nil
		}
		if hasPendingWorkflowForResource(*data, "mix_design", id) {
			item = data.MixDesigns[idx]
			topic = "laboratory.mix_design.retire_requested"
			return nil
		}
		_, instances, err := publishMixDesignRetireWorkflow(data, data.MixDesigns[idx], session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			item = data.MixDesigns[idx]
			topic = "laboratory.mix_design.retire_requested"
			addAudit(data, session.User.Username, "request_retire", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
			return nil
		}
		next, err := applyMixDesignRetireLocked(data, id)
		if err != nil {
			return err
		}
		item = next
		addAudit(data, session.User.Username, "retire", "mix_design", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishMixDesignRetireWorkflow(data *AppData, item MixDesign, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "mix_design.retire_requested",
		Source:     "laboratory",
		Resource:   "mix_design",
		ResourceID: item.ID,
		ResourceNo: item.Code + " " + item.Version,
		Title:      "生产配比退役",
		Actor:      actor,
		Reason:     fallback(item.Scope, "生产配比退役审批"),
		Variables: map[string]string{
			"workflowAction": "retire",
			"productId":      fmt.Sprintf("%d", item.ProductID),
			"siteId":         fmt.Sprintf("%d", item.SiteID),
			"code":           item.Code,
			"version":        item.Version,
			"isCurrent":      fmt.Sprintf("%t", item.IsCurrent),
		},
	})
}

func applyMixDesignRetireLocked(data *AppData, id int64) (MixDesign, error) {
	idx := mixDesignIndex(*data, id)
	if idx < 0 {
		return MixDesign{}, fmt.Errorf("配比不存在")
	}
	data.MixDesigns[idx].Status = "retired"
	data.MixDesigns[idx].IsCurrent = false
	data.MixDesigns[idx].RetiredAt = nowString()
	data.MixDesigns[idx].UpdatedAt = data.MixDesigns[idx].RetiredAt
	return data.MixDesigns[idx], nil
}

func (a *App) createMixDesignPlantProfile(w http.ResponseWriter, r *http.Request, session Session, mixDesignID int64) {
	var req MixDesignPlantProfile
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design plant profile")
		return
	}
	var item MixDesignPlantProfile
	err := a.store.Mutate(func(data *AppData) error {
		base, ok := findMixDesign(*data, mixDesignID)
		if !ok {
			return fmt.Errorf("配比不存在")
		}
		if base.Status != "approved" {
			return fmt.Errorf("基础配比未审批不能配置生产线微调")
		}
		if session.User.SiteID > 0 && base.SiteID != 0 && base.SiteID != session.User.SiteID {
			return fmt.Errorf("无权操作其他站点配比")
		}
		plant, ok := findPlantByID(*data, req.PlantID)
		if !ok {
			return fmt.Errorf("生产线不存在")
		}
		if base.SiteID != 0 && plant.SiteID != base.SiteID {
			return fmt.Errorf("生产线必须属于基础配比站点")
		}
		if session.User.SiteID > 0 && plant.SiteID != session.User.SiteID {
			return fmt.Errorf("无权操作其他站点生产线")
		}
		item = req
		item.ID = nextID(data, "mixPlantProfile")
		item.MixDesignID = base.ID
		item.ProductID = base.ProductID
		item.SiteID = plant.SiteID
		item.PlantID = plant.ID
		item.PlantCode = plant.Code
		item.Code = fallback(req.Code, base.Code+"-"+plant.Code)
		item.Version = fallback(req.Version, base.Version+"-line")
		item.Scope = fallback(req.Scope, plant.Name+"生产线配比")
		item.Status = fallback(req.Status, "approved")
		item.IsCurrent = item.Status == "approved"
		item.EffectiveFrom = fallback(req.EffectiveFrom, todayString())
		item.EffectiveTo = req.EffectiveTo
		item.CreatedBy = fallback(req.CreatedBy, session.User.DisplayName)
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		if item.Status == "approved" {
			item.ApprovedBy = fallback(req.ApprovedBy, session.User.DisplayName)
			item.ApprovedAt = item.CreatedAt
		}
		if err := validateMixDesignPlantProfile(*data, base, &item); err != nil {
			return err
		}
		if item.Status == "approved" {
			_, instances, err := publishMixDesignPlantProfileWorkflow(data, item, mixDesignPlantProfileApprovalRequest{EffectiveFrom: item.EffectiveFrom, EffectiveTo: item.EffectiveTo}, session.User.Username)
			if err != nil {
				return err
			}
			if len(instances) > 0 {
				item.Status = "pending_approval"
				item.IsCurrent = false
				item.ApprovedBy = ""
				item.ApprovedAt = ""
			}
		}
		if item.Status == "approved" && item.IsCurrent {
			retirePreviousCurrentMixDesignPlantProfiles(data, item.ID, item.MixDesignID, item.PlantID, item.UpdatedAt)
		}
		data.MixDesignPlantProfiles = append(data.MixDesignPlantProfiles, item)
		addAudit(data, session.User.Username, "create", "mix_design_plant_profile", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "laboratory.mix_design_plant_profile.created")
}

type mixDesignPlantProfileApprovalRequest struct {
	EffectiveFrom string `json:"effectiveFrom"`
	EffectiveTo   string `json:"effectiveTo"`
}

func (a *App) approveMixDesignPlantProfile(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req mixDesignPlantProfileApprovalRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid mix design plant profile approval")
		return
	}
	var item MixDesignPlantProfile
	topic := "laboratory.mix_design_plant_profile.approved"
	err := a.store.Mutate(func(data *AppData) error {
		idx := mixDesignPlantProfileIndex(*data, id)
		if idx < 0 {
			return fmt.Errorf("生产线配比不存在")
		}
		base, ok := findMixDesign(*data, data.MixDesignPlantProfiles[idx].MixDesignID)
		if !ok {
			return fmt.Errorf("基础配比不存在")
		}
		if session.User.SiteID > 0 && data.MixDesignPlantProfiles[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权审批其他站点生产线配比")
		}
		if err := validateMixDesignPlantProfile(*data, base, &data.MixDesignPlantProfiles[idx]); err != nil {
			return err
		}
		_, instances, err := publishMixDesignPlantProfileWorkflow(data, data.MixDesignPlantProfiles[idx], req, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			data.MixDesignPlantProfiles[idx].Status = "pending_approval"
			data.MixDesignPlantProfiles[idx].IsCurrent = false
			data.MixDesignPlantProfiles[idx].UpdatedAt = nowString()
			item = data.MixDesignPlantProfiles[idx]
			topic = "laboratory.mix_design_plant_profile.workflow_requested"
			addAudit(data, session.User.Username, "request_approve", "mix_design_plant_profile", item.ID, item.Code+" "+item.Version, clientIP(r))
			return nil
		}
		next, err := applyMixDesignPlantProfileApprovalLocked(data, id, req, session.User.DisplayName)
		if err != nil {
			return err
		}
		item = next
		addAudit(data, session.User.Username, "approve", "mix_design_plant_profile", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishMixDesignPlantProfileWorkflow(data *AppData, item MixDesignPlantProfile, req mixDesignPlantProfileApprovalRequest, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "mix_design_plant_profile.submitted",
		Source:     "laboratory",
		Resource:   "mix_design_plant_profile",
		ResourceID: item.ID,
		ResourceNo: item.Code + " " + item.Version,
		Title:      "生产线配比审批",
		Actor:      actor,
		Reason:     fallback(item.Scope, "生产线配比微调审批"),
		Variables: map[string]string{
			"workflowAction": "approve",
			"mixDesignId":    fmt.Sprintf("%d", item.MixDesignID),
			"productId":      fmt.Sprintf("%d", item.ProductID),
			"siteId":         fmt.Sprintf("%d", item.SiteID),
			"plantId":        fmt.Sprintf("%d", item.PlantID),
			"plantCode":      item.PlantCode,
			"code":           item.Code,
			"version":        item.Version,
			"effectiveFrom":  fallback(req.EffectiveFrom, item.EffectiveFrom),
			"effectiveTo":    fallback(req.EffectiveTo, item.EffectiveTo),
		},
	})
}

func applyMixDesignPlantProfileApprovalLocked(data *AppData, id int64, req mixDesignPlantProfileApprovalRequest, actor string) (MixDesignPlantProfile, error) {
	idx := mixDesignPlantProfileIndex(*data, id)
	if idx < 0 {
		return MixDesignPlantProfile{}, fmt.Errorf("生产线配比不存在")
	}
	base, ok := findMixDesign(*data, data.MixDesignPlantProfiles[idx].MixDesignID)
	if !ok {
		return MixDesignPlantProfile{}, fmt.Errorf("基础配比不存在")
	}
	if err := validateMixDesignPlantProfile(*data, base, &data.MixDesignPlantProfiles[idx]); err != nil {
		return MixDesignPlantProfile{}, err
	}
	now := nowString()
	data.MixDesignPlantProfiles[idx].Status = "approved"
	data.MixDesignPlantProfiles[idx].IsCurrent = true
	data.MixDesignPlantProfiles[idx].EffectiveFrom = fallback(req.EffectiveFrom, todayString())
	data.MixDesignPlantProfiles[idx].EffectiveTo = fallback(req.EffectiveTo, data.MixDesignPlantProfiles[idx].EffectiveTo)
	data.MixDesignPlantProfiles[idx].ApprovedBy = actor
	data.MixDesignPlantProfiles[idx].ApprovedAt = now
	data.MixDesignPlantProfiles[idx].UpdatedAt = now
	item := data.MixDesignPlantProfiles[idx]
	retirePreviousCurrentMixDesignPlantProfiles(data, item.ID, item.MixDesignID, item.PlantID, now)
	return item, nil
}

func (a *App) retireMixDesignPlantProfile(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item MixDesignPlantProfile
	topic := "laboratory.mix_design_plant_profile.retired"
	err := a.store.Mutate(func(data *AppData) error {
		idx := mixDesignPlantProfileIndex(*data, id)
		if idx < 0 {
			return fmt.Errorf("生产线配比不存在")
		}
		if session.User.SiteID > 0 && data.MixDesignPlantProfiles[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权停用其他站点生产线配比")
		}
		if data.MixDesignPlantProfiles[idx].Status == "retired" {
			item = data.MixDesignPlantProfiles[idx]
			return nil
		}
		if hasPendingWorkflowForResource(*data, "mix_design_plant_profile", id) {
			item = data.MixDesignPlantProfiles[idx]
			topic = "laboratory.mix_design_plant_profile.retire_requested"
			return nil
		}
		_, instances, err := publishMixDesignPlantProfileRetireWorkflow(data, data.MixDesignPlantProfiles[idx], session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			item = data.MixDesignPlantProfiles[idx]
			topic = "laboratory.mix_design_plant_profile.retire_requested"
			addAudit(data, session.User.Username, "request_retire", "mix_design_plant_profile", item.ID, item.Code+" "+item.Version, clientIP(r))
			return nil
		}
		next, err := applyMixDesignPlantProfileRetireLocked(data, id)
		if err != nil {
			return err
		}
		item = next
		addAudit(data, session.User.Username, "retire", "mix_design_plant_profile", item.ID, item.Code+" "+item.Version, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishMixDesignPlantProfileRetireWorkflow(data *AppData, item MixDesignPlantProfile, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "mix_design_plant_profile.retire_requested",
		Source:     "laboratory",
		Resource:   "mix_design_plant_profile",
		ResourceID: item.ID,
		ResourceNo: item.Code + " " + item.Version,
		Title:      "生产线配比退役",
		Actor:      actor,
		Reason:     fallback(item.Scope, "生产线配比退役审批"),
		Variables: map[string]string{
			"workflowAction": "retire",
			"mixDesignId":    fmt.Sprintf("%d", item.MixDesignID),
			"productId":      fmt.Sprintf("%d", item.ProductID),
			"siteId":         fmt.Sprintf("%d", item.SiteID),
			"plantId":        fmt.Sprintf("%d", item.PlantID),
			"plantCode":      item.PlantCode,
			"code":           item.Code,
			"version":        item.Version,
			"isCurrent":      fmt.Sprintf("%t", item.IsCurrent),
		},
	})
}

func applyMixDesignPlantProfileRetireLocked(data *AppData, id int64) (MixDesignPlantProfile, error) {
	idx := mixDesignPlantProfileIndex(*data, id)
	if idx < 0 {
		return MixDesignPlantProfile{}, fmt.Errorf("生产线配比不存在")
	}
	data.MixDesignPlantProfiles[idx].Status = "retired"
	data.MixDesignPlantProfiles[idx].IsCurrent = false
	data.MixDesignPlantProfiles[idx].RetiredAt = nowString()
	data.MixDesignPlantProfiles[idx].UpdatedAt = data.MixDesignPlantProfiles[idx].RetiredAt
	return data.MixDesignPlantProfiles[idx], nil
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
	topic := "laboratory.test.reviewed"
	err := a.store.Mutate(func(data *AppData) error {
		idx := laboratoryTestIndex(*data, testID)
		if idx < 0 {
			return fmt.Errorf("试验记录不存在")
		}
		if session.User.SiteID > 0 && data.LaboratoryTests[idx].SiteID != session.User.SiteID {
			return fmt.Errorf("无权复核其他站点试验")
		}
		_, instances, err := publishLaboratoryTestReviewWorkflow(data, data.LaboratoryTests[idx], req, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			data.LaboratoryTests[idx].Status = "pending_approval"
			item = data.LaboratoryTests[idx]
			topic = "laboratory.test.workflow_requested"
			addAudit(data, session.User.Username, "request_review", "laboratory_test", item.ID, item.TestNo, clientIP(r))
			return nil
		}
		next, err := applyLaboratoryTestReviewLocked(data, testID, req, session.User.DisplayName)
		if err != nil {
			return err
		}
		item = next
		addAudit(data, session.User.Username, "review", "laboratory_test", item.ID, item.TestNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishLaboratoryTestReviewWorkflow(data *AppData, item LaboratoryTestRecord, req LaboratoryTestRecord, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	value := item.Value
	if req.Value > 0 {
		value = req.Value
	}
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "laboratory_test.review_requested",
		Source:     "laboratory",
		Resource:   "laboratory_test",
		ResourceID: item.ID,
		ResourceNo: item.TestNo,
		Title:      "试验复核 " + item.TestNo,
		Actor:      actor,
		Reason:     fallback(req.Remark, "实验室试验复核"),
		Variables: map[string]string{
			"sampleId":    fmt.Sprintf("%d", item.SampleID),
			"equipmentId": fmt.Sprintf("%d", item.EquipmentID),
			"siteId":      fmt.Sprintf("%d", item.SiteID),
			"testType":    item.TestType,
			"metric":      item.Metric,
			"value":       fmt.Sprintf("%.4f", value),
			"unit":        item.Unit,
			"result":      fallback(req.Result, item.Result),
			"reviewer":    req.Reviewer,
			"remark":      fallback(req.Remark, item.Remark),
		},
	})
}

func applyLaboratoryTestReviewLocked(data *AppData, testID int64, req LaboratoryTestRecord, actor string) (LaboratoryTestRecord, error) {
	idx := laboratoryTestIndex(*data, testID)
	if idx < 0 {
		return LaboratoryTestRecord{}, fmt.Errorf("试验记录不存在")
	}
	if req.Result != "" {
		data.LaboratoryTests[idx].Result = req.Result
	}
	if req.Value > 0 {
		data.LaboratoryTests[idx].Value = req.Value
	}
	data.LaboratoryTests[idx].Status = "reviewed"
	data.LaboratoryTests[idx].Reviewer = fallback(req.Reviewer, actor)
	data.LaboratoryTests[idx].ReviewedAt = nowString()
	data.LaboratoryTests[idx].Remark = fallback(req.Remark, data.LaboratoryTests[idx].Remark)
	item := data.LaboratoryTests[idx]
	updateLaboratorySample(data, item.SampleID, "completed", item.Result)
	if item.Result == "failed" {
		appendQualityException(data, "laboratory_test", item.ID, item.SiteID, "试验结果不合格", "试验 "+item.TestNo+" 复核为不合格", "high", actor)
	}
	return item, nil
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
		if err := publishQualityExceptionWorkflow(data, item, session.User.Username); err != nil {
			return err
		}
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
	topic := "laboratory.exception.handled"
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.QualityExceptions {
			if data.QualityExceptions[i].ID != id {
				continue
			}
			if session.User.SiteID > 0 && data.QualityExceptions[i].SiteID != session.User.SiteID {
				return fmt.Errorf("无权关闭其他站点异常")
			}
			if hasPendingWorkflowForEvent(*data, "quality_exception", id, "quality_exception.close_requested") {
				item = data.QualityExceptions[i]
				topic = "laboratory.exception.close_requested"
				return nil
			}
			_, instances, err := publishQualityExceptionCloseWorkflow(data, data.QualityExceptions[i], req, session.User.Username)
			if err != nil {
				return err
			}
			if len(instances) > 0 {
				item = data.QualityExceptions[i]
				topic = "laboratory.exception.close_requested"
				addAudit(data, session.User.Username, "request_close", "quality_exception", item.ID, item.ExceptionNo, clientIP(r))
				return nil
			}
			next, err := applyQualityExceptionCloseLocked(data, id, req, session.User.DisplayName)
			if err != nil {
				return err
			}
			item = next
			addAudit(data, session.User.Username, "handle", "quality_exception", item.ID, item.ExceptionNo, clientIP(r))
			return nil
		}
		return fmt.Errorf("质量异常不存在")
	})
	a.respondMutation(w, err, item, topic)
}

func publishQualityExceptionCloseWorkflow(data *AppData, item QualityException, req QualityException, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "quality_exception.close_requested",
		Source:     "laboratory",
		Resource:   "quality_exception",
		ResourceID: item.ID,
		ResourceNo: item.ExceptionNo,
		Title:      "质量异常关闭 " + item.ExceptionNo,
		Actor:      actor,
		Reason:     fallback(strings.TrimSpace(req.CorrectiveAction), fallback(strings.TrimSpace(item.Description), "质量异常关闭审批")),
		Variables: map[string]string{
			"severity":         item.Severity,
			"sourceType":       item.SourceType,
			"sourceId":         fmt.Sprintf("%d", item.SourceID),
			"siteId":           fmt.Sprintf("%d", item.SiteID),
			"responsible":      fallback(strings.TrimSpace(req.Responsible), item.Responsible),
			"rootCause":        fallback(strings.TrimSpace(req.RootCause), item.RootCause),
			"correctiveAction": fallback(strings.TrimSpace(req.CorrectiveAction), item.CorrectiveAction),
			"targetStatus":     "closed",
			"currentStatus":    item.Status,
		},
	})
}

func applyQualityExceptionCloseLocked(data *AppData, id int64, req QualityException, closedBy string) (QualityException, error) {
	for i := range data.QualityExceptions {
		if data.QualityExceptions[i].ID != id {
			continue
		}
		data.QualityExceptions[i].RootCause = fallback(req.RootCause, data.QualityExceptions[i].RootCause)
		data.QualityExceptions[i].CorrectiveAction = fallback(req.CorrectiveAction, data.QualityExceptions[i].CorrectiveAction)
		data.QualityExceptions[i].Responsible = fallback(req.Responsible, data.QualityExceptions[i].Responsible)
		data.QualityExceptions[i].Status = "closed"
		data.QualityExceptions[i].HandledAt = nowString()
		data.QualityExceptions[i].ClosedBy = closedBy
		return data.QualityExceptions[i], nil
	}
	return QualityException{}, fmt.Errorf("质量异常不存在")
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

func validateMixMaterials(data AppData, item *MixDesign) error {
	if len(item.Materials) == 0 {
		return fmt.Errorf("配比材料不能为空")
	}
	seen := map[int64]bool{}
	for i := range item.Materials {
		materialID := item.Materials[i].MaterialID
		if materialID <= 0 {
			return fmt.Errorf("配比材料必须选择物料")
		}
		if seen[materialID] {
			return fmt.Errorf("配比材料不能重复")
		}
		material, ok := findMaterial(data, materialID)
		if !ok {
			return fmt.Errorf("配比材料不存在")
		}
		if material.Status != "" && material.Status != "active" {
			return fmt.Errorf("配比材料已停用")
		}
		if item.Materials[i].Dosage <= 0 {
			return fmt.Errorf("%s 用量必须大于 0", material.Name)
		}
		if item.Materials[i].Unit == "" {
			item.Materials[i].Unit = "kg/t"
		}
		seen[materialID] = true
	}
	return nil
}

func validateMixDesignPlantProfile(data AppData, base MixDesign, item *MixDesignPlantProfile) error {
	if item.MixDesignID != base.ID {
		return fmt.Errorf("生产线配比必须关联基础配比")
	}
	if item.PlantID == 0 {
		return fmt.Errorf("生产线不能为空")
	}
	plant, ok := findPlantByID(data, item.PlantID)
	if !ok {
		return fmt.Errorf("生产线不存在")
	}
	if base.SiteID != 0 && plant.SiteID != base.SiteID {
		return fmt.Errorf("生产线必须属于基础配比站点")
	}
	baseMaterials := map[int64]MixDesignMaterial{}
	for _, material := range base.Materials {
		baseMaterials[material.MaterialID] = material
	}
	seen := map[int64]bool{}
	for i := range item.Materials {
		material := &item.Materials[i]
		if material.MaterialID <= 0 {
			return fmt.Errorf("微调物料不能为空")
		}
		baseMaterial, ok := baseMaterials[material.MaterialID]
		if !ok {
			return fmt.Errorf("微调物料必须存在于基础配比")
		}
		if _, ok := findMaterial(data, material.MaterialID); !ok {
			return fmt.Errorf("微调物料不存在")
		}
		if seen[material.MaterialID] {
			return fmt.Errorf("微调物料不能重复")
		}
		seen[material.MaterialID] = true
		material.Unit = fallback(material.Unit, baseMaterial.Unit)
		if material.Dosage < 0 {
			return fmt.Errorf("微调用量不能小于 0")
		}
		if material.Dosage == 0 && material.Adjustment == 0 && material.BufferID == 0 && strings.TrimSpace(material.BufferCode) == "" {
			return fmt.Errorf("微调项必须填写用量、增减量或筒仓")
		}
		if material.BufferID != 0 || strings.TrimSpace(material.BufferCode) != "" {
			buffer, ok := findPlantBufferForProfile(data, item.PlantID, material.BufferID, material.BufferCode)
			if !ok {
				return fmt.Errorf("微调筒仓不存在")
			}
			if !productionLineBufferCanCarryMaterial(buffer, material.MaterialID) {
				return fmt.Errorf("微调筒仓不能承接该物料")
			}
			material.BufferID = buffer.ID
			material.BufferCode = buffer.Code
		}
	}
	applied := applyMixDesignPlantProfile(base, *item)
	return validateMixMaterials(data, &applied)
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

func mixDesignPlantProfileIndex(data AppData, id int64) int {
	for i := range data.MixDesignPlantProfiles {
		if data.MixDesignPlantProfiles[i].ID == id {
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

func retirePreviousCurrentMixDesignPlantProfiles(data *AppData, currentID, mixDesignID, plantID int64, retiredAt string) {
	for i := range data.MixDesignPlantProfiles {
		if data.MixDesignPlantProfiles[i].ID == currentID ||
			data.MixDesignPlantProfiles[i].MixDesignID != mixDesignID ||
			data.MixDesignPlantProfiles[i].PlantID != plantID {
			continue
		}
		if data.MixDesignPlantProfiles[i].IsCurrent && data.MixDesignPlantProfiles[i].Status == "approved" {
			data.MixDesignPlantProfiles[i].IsCurrent = false
			data.MixDesignPlantProfiles[i].Status = "retired"
			data.MixDesignPlantProfiles[i].RetiredAt = retiredAt
			data.MixDesignPlantProfiles[i].UpdatedAt = retiredAt
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
	_ = publishQualityExceptionWorkflow(data, item, fallback(responsible, "system"))
	return item
}

func publishQualityExceptionWorkflow(data *AppData, item QualityException, actor string) error {
	eventKey := "quality_exception:" + item.ExceptionNo
	_, _, err := publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "quality_exception.submitted",
		Source:     "laboratory",
		EventKey:   eventKey,
		Resource:   "quality_exception",
		ResourceID: item.ID,
		ResourceNo: item.ExceptionNo,
		Title:      fallback(item.Title, "质量异常"),
		Actor:      actor,
		Reason:     item.Description,
		Variables: map[string]string{
			"severity":    item.Severity,
			"sourceType":  item.SourceType,
			"sourceId":    fmt.Sprintf("%d", item.SourceID),
			"siteId":      fmt.Sprintf("%d", item.SiteID),
			"responsible": item.Responsible,
		},
	})
	return err
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
