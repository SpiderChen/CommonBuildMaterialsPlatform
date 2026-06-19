package appliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type plantBatchPayload struct {
	DeviceNo      string                          `json:"deviceNo"`
	TaskID        int64                           `json:"taskId"`
	TaskNo        string                          `json:"taskNo"`
	BatchNo       string                          `json:"batchNo"`
	PlantCode     string                          `json:"plantCode"`
	Quantity      float64                         `json:"quantity"`
	Operator      string                          `json:"operator"`
	QualityStatus string                          `json:"qualityStatus"`
	Status        string                          `json:"status"`
	StartedAt     string                          `json:"startedAt"`
	CompletedAt   string                          `json:"completedAt"`
	Materials     []productionMaterialConsumption `json:"materials"`
}

type productionMaterialConsumption struct {
	MaterialID int64   `json:"materialId"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
}

func (a *App) ingestPlantProtocolFrame(w http.ResponseWriter, r *http.Request, session Session) {
	req, err := readProtocolFrameIngestRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Channel = fallback(strings.TrimSpace(req.Channel), "opc")
	req.Protocol = fallback(strings.TrimSpace(req.Protocol), "plant-json")
	response, err := a.processPlantProtocolFrame(r, session, req)
	if err != nil {
		writeProtocolFrameError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) processPlantProtocolFrame(r *http.Request, session Session, req protocolFrameIngestRequest) (protocolIngestResponse, error) {
	payload, err := parsePlantProtocolFrame(req)
	frame := baseProtocolFrame(session, req, "production_batch")
	frame.DeviceNo = fallback(payload.DeviceNo, fallback(payload.PlantCode, deviceNoFromSession(session)))
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	batch, err := a.recordPlantProductionBatch(r, session, payload)
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	frame.DeviceNo = fallback(payload.DeviceNo, fallback(batch.PlantCode, frame.DeviceNo))
	frame.ParsedID = batch.ID
	frame.Status = "accepted"
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		return protocolIngestResponse{}, err
	}
	a.emit("production.batch.created", batch)
	a.emit("device.protocol.frame", saved)
	return protocolIngestResponse{Frame: saved, ProductionBatch: &batch}, nil
}

func parsePlantProtocolFrame(req protocolFrameIngestRequest) (plantBatchPayload, error) {
	raw := strings.TrimSpace(req.Raw)
	var payload plantBatchPayload
	if len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return payload, fmt.Errorf("invalid plant json payload")
		}
		return normalizePlantBatchPayload(payload), nil
	}
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return payload, fmt.Errorf("invalid plant json frame")
		}
		return normalizePlantBatchPayload(payload), nil
	}
	fields, err := parseProtocolCSV(raw)
	if err != nil {
		return payload, fmt.Errorf("invalid plant csv frame: %w", err)
	}
	offset := 0
	if len(fields) > 0 && (strings.EqualFold(fields[0], "PLANT") || strings.EqualFold(fields[0], "BATCH")) {
		offset = 1
	}
	if len(fields)-offset < 4 {
		return payload, fmt.Errorf("plant frame requires task, plant, quantity and completedAt")
	}
	taskID, err := strconv.ParseInt(fields[offset], 10, 64)
	if err == nil {
		payload.TaskID = taskID
	} else {
		payload.TaskNo = fields[offset]
	}
	payload.PlantCode = fields[offset+1]
	payload.Quantity, err = parseFloatField(fields[offset+2], "quantity")
	if err != nil {
		return payload, err
	}
	payload.CompletedAt = fields[offset+3]
	if len(fields)-offset > 4 {
		payload.Operator = fields[offset+4]
	}
	if len(fields)-offset > 5 {
		payload.StartedAt = fields[offset+5]
	}
	if len(fields)-offset > 6 {
		payload.Status = fields[offset+6]
	}
	if len(fields)-offset > 7 {
		payload.QualityStatus = fields[offset+7]
	}
	if len(fields)-offset > 8 {
		materials, err := parsePlantMaterialCSV(fields[offset+8])
		if err != nil {
			return payload, err
		}
		payload.Materials = materials
	}
	return normalizePlantBatchPayload(payload), nil
}

func normalizePlantBatchPayload(payload plantBatchPayload) plantBatchPayload {
	payload.DeviceNo = strings.TrimSpace(payload.DeviceNo)
	payload.TaskNo = strings.TrimSpace(payload.TaskNo)
	payload.BatchNo = strings.TrimSpace(payload.BatchNo)
	payload.PlantCode = strings.TrimSpace(payload.PlantCode)
	payload.Operator = strings.TrimSpace(payload.Operator)
	payload.QualityStatus = strings.TrimSpace(payload.QualityStatus)
	payload.Status = strings.TrimSpace(payload.Status)
	payload.StartedAt = strings.TrimSpace(payload.StartedAt)
	payload.CompletedAt = strings.TrimSpace(payload.CompletedAt)
	for i := range payload.Materials {
		payload.Materials[i].Unit = strings.TrimSpace(payload.Materials[i].Unit)
	}
	return payload
}

func parsePlantMaterialCSV(raw string) ([]productionMaterialConsumption, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	segments := strings.Split(raw, "|")
	materials := make([]productionMaterialConsumption, 0, len(segments))
	for _, segment := range segments {
		parts := strings.Split(segment, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid plant material segment")
		}
		materialID, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid materialId")
		}
		quantity, err := parseFloatField(parts[1], "material quantity")
		if err != nil {
			return nil, err
		}
		unit := "t"
		if len(parts) > 2 {
			unit = strings.TrimSpace(parts[2])
		}
		materials = append(materials, productionMaterialConsumption{MaterialID: materialID, Quantity: quantity, Unit: unit})
	}
	return materials, nil
}

func (a *App) recordPlantProductionBatch(r *http.Request, session Session, payload plantBatchPayload) (ProductionBatch, error) {
	var item ProductionBatch
	err := a.store.Mutate(func(data *AppData) error {
		task, ok := findProductionTaskByProtocol(*data, payload)
		if !ok {
			return fmt.Errorf("生产任务不存在")
		}
		if task.Status == "completed" {
			return fmt.Errorf("生产任务已完成")
		}
		plan, ok := findProductionPlan(*data, task.PlanID)
		if !ok {
			return fmt.Errorf("生产计划不存在")
		}
		mix, ok := findMixDesign(*data, task.MixDesignID)
		if !ok {
			return fmt.Errorf("配方不存在")
		}
		remaining := round(task.PlanQty - task.ProducedQty)
		quantity := round(payload.Quantity)
		if quantity <= 0 {
			return fmt.Errorf("批次数量必须大于 0")
		}
		if quantity > remaining {
			return fmt.Errorf("批次数量超过任务剩余量")
		}
		item.ID = nextID(data, "productionBatch")
		item.BatchNo = fallback(payload.BatchNo, number("PB", item.ID))
		item.TaskID = task.ID
		item.PlanID = task.PlanID
		item.OrderID = task.OrderID
		item.SiteID = task.SiteID
		item.ProductID = task.ProductID
		item.MixDesignID = task.MixDesignID
		item.Quantity = quantity
		item.PlantCode = fallback(payload.PlantCode, defaultPlantCode(*data, task.SiteID))
		item.Operator = fallback(payload.Operator, session.User.DisplayName)
		item.QualityStatus = fallback(payload.QualityStatus, "pending")
		item.Status = fallback(payload.Status, "produced")
		item.StartedAt = fallback(payload.StartedAt, nowString())
		item.CompletedAt = fallback(payload.CompletedAt, nowString())
		if len(payload.Materials) > 0 {
			if err := consumeBatchInventoryMaterials(data, item, payload.Materials, "拌合楼实际消耗"); err != nil {
				return err
			}
		} else if err := consumeBatchInventory(data, item, mix); err != nil {
			return err
		}
		data.ProductionBatches = append(data.ProductionBatches, item)
		updateProductionProgress(data, task.ID, plan.ID, item.Quantity)
		addAudit(data, session.User.Username, "protocol_ingest", "production_batch", item.ID, item.BatchNo, clientIP(r))
		return nil
	})
	return item, err
}

func findProductionTaskByProtocol(data AppData, payload plantBatchPayload) (ProductionTask, bool) {
	if payload.TaskID > 0 {
		return findProductionTask(data, payload.TaskID)
	}
	for _, item := range data.ProductionTasks {
		if item.TaskNo == payload.TaskNo {
			return item, true
		}
	}
	return ProductionTask{}, false
}
