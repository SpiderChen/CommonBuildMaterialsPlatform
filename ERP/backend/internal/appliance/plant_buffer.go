package appliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type plantBufferTransferRequest struct {
	BufferID     int64   `json:"bufferId"`
	BufferCode   string  `json:"bufferCode"`
	YardPileID   int64   `json:"yardPileId"`
	YardPileCode string  `json:"yardPileCode"`
	MaterialID   int64   `json:"materialId"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Remark       string  `json:"remark"`
}

type plantBufferAdjustmentRequest struct {
	BufferID      int64   `json:"bufferId"`
	BufferCode    string  `json:"bufferCode"`
	ActualQty     float64 `json:"actualQty"`
	MoistureRate  float64 `json:"moistureRate"`
	QualityStatus string  `json:"qualityStatus"`
	Status        string  `json:"status"`
	Remark        string  `json:"remark"`
}

type bufferLevelPayload struct {
	DeviceNo      string  `json:"deviceNo"`
	PlantCode     string  `json:"plantCode"`
	BufferCode    string  `json:"bufferCode"`
	BinCode       string  `json:"binCode"`
	MaterialID    int64   `json:"materialId"`
	Quantity      float64 `json:"quantity"`
	MoistureRate  float64 `json:"moistureRate"`
	QualityStatus string  `json:"qualityStatus"`
	Status        string  `json:"status"`
	ReportedAt    string  `json:"reportedAt"`
}

func (a *App) createPlantBufferLocation(w http.ResponseWriter, r *http.Request, session Session) {
	var item PlantBufferLocation
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid plant buffer location")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		normalized, err := normalizePlantBufferLocation(*data, session.User, item, PlantBufferLocation{})
		if err != nil {
			return err
		}
		normalized.ID = nextID(data, "plantBuffer")
		normalized.CreatedAt = nowString()
		normalized.UpdatedAt = normalized.CreatedAt
		data.PlantBufferLocations = append(data.PlantBufferLocations, normalized)
		item = normalized
		addAudit(data, session.User.Username, "create", "plant_buffer", item.ID, item.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.plant_buffer.created")
}

func (a *App) updatePlantBufferLocation(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item PlantBufferLocation
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid plant buffer location")
		return
	}
	var updated PlantBufferLocation
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.PlantBufferLocations {
			if data.PlantBufferLocations[i].ID != id {
				continue
			}
			normalized, err := normalizePlantBufferLocation(*data, session.User, item, data.PlantBufferLocations[i])
			if err != nil {
				return err
			}
			normalized.ID = id
			normalized.CurrentQty = data.PlantBufferLocations[i].CurrentQty
			normalized.CreatedAt = data.PlantBufferLocations[i].CreatedAt
			normalized.UpdatedAt = nowString()
			data.PlantBufferLocations[i] = normalized
			updated = normalized
			addAudit(data, session.User.Username, "update", "plant_buffer", id, normalized.Code, clientIP(r))
			return nil
		}
		return fmt.Errorf("暂存仓位不存在")
	})
	a.respondUpdate(w, err, updated, "master.plant_buffer.updated")
}

func (a *App) deletePlantBufferLocation(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var deleted PlantBufferLocation
	err := a.store.Mutate(func(data *AppData) error {
		for _, flow := range data.PlantBufferFlows {
			if flow.BufferID == id {
				return fmt.Errorf("暂存仓位已有流水，不能删除")
			}
		}
		for i, item := range data.PlantBufferLocations {
			if item.ID != id {
				continue
			}
			if item.CurrentQty != 0 {
				return fmt.Errorf("暂存仓位有余额，不能删除")
			}
			deleted = item
			data.PlantBufferLocations = append(data.PlantBufferLocations[:i], data.PlantBufferLocations[i+1:]...)
			addAudit(data, session.User.Username, "delete", "plant_buffer", id, item.Code, clientIP(r))
			return nil
		}
		return fmt.Errorf("暂存仓位不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("master.plant_buffer.deleted", deleted)
	writeJSON(w, http.StatusOK, deleted)
}

func normalizePlantBufferLocation(data AppData, user User, item PlantBufferLocation, current PlantBufferLocation) (PlantBufferLocation, error) {
	if item.PlantID == 0 {
		item.PlantID = current.PlantID
	}
	plant, ok := findPlantByID(data, item.PlantID)
	if !ok {
		return item, fmt.Errorf("生产线不存在")
	}
	requestedSiteID := nonZeroInt(item.SiteID, plant.SiteID)
	siteID := requestedSiteID
	if user.Username == "import" {
		if _, ok := findSite(data, requestedSiteID); !ok {
			return item, fmt.Errorf("站点不存在")
		}
	} else {
		var err error
		siteID, err = writableSiteID(data, user, requestedSiteID)
		if err != nil {
			return item, err
		}
	}
	if siteID != plant.SiteID {
		return item, fmt.Errorf("暂存仓位站点必须与生产线一致")
	}
	item.SiteID = siteID
	item.PlantCode = plant.Code
	item.Code = strings.TrimSpace(fallback(item.Code, current.Code))
	item.Name = strings.TrimSpace(fallback(item.Name, current.Name))
	if item.Code == "" || item.Name == "" {
		return item, fmt.Errorf("仓位名称和编码不能为空")
	}
	if plantBufferCodeExists(data.PlantBufferLocations, item.Code, current.ID) {
		return item, fmt.Errorf("仓位编码已存在")
	}
	item.Type = fallback(strings.TrimSpace(item.Type), fallback(current.Type, "aggregate_bin"))
	item.MaterialID = nonZeroInt(item.MaterialID, current.MaterialID)
	if item.MaterialID != 0 {
		if _, ok := findMaterial(data, item.MaterialID); !ok {
			return item, fmt.Errorf("物料不存在")
		}
	}
	item.AllowedMaterialIDs = normalizeAllowedMaterialIDs(data, item.AllowedMaterialIDs, current.AllowedMaterialIDs)
	item.Unit = fallback(strings.TrimSpace(item.Unit), fallback(current.Unit, "t"))
	item.Capacity = nonZero(item.Capacity, current.Capacity)
	if item.Capacity <= 0 {
		return item, fmt.Errorf("仓位容量必须大于 0")
	}
	item.WarningQty = nonZero(item.WarningQty, current.WarningQty)
	item.MoistureRate = nonZero(item.MoistureRate, current.MoistureRate)
	item.QualityStatus = fallback(strings.TrimSpace(item.QualityStatus), fallback(current.QualityStatus, "passed"))
	item.Status = fallback(strings.TrimSpace(item.Status), fallback(current.Status, "active"))
	item.GatewayDeviceNo = fallback(strings.TrimSpace(item.GatewayDeviceNo), current.GatewayDeviceNo)
	item.GatewayChannel = fallback(strings.TrimSpace(item.GatewayChannel), current.GatewayChannel)
	item.GatewayProtocol = fallback(strings.TrimSpace(item.GatewayProtocol), current.GatewayProtocol)
	item.LastReportedAt = current.LastReportedAt
	item.LastFrameStatus = current.LastFrameStatus
	item.GatewayError = current.GatewayError
	return item, nil
}

func normalizeAllowedMaterialIDs(data AppData, ids []int64, fallbackIDs []int64) []int64 {
	if len(ids) == 0 {
		ids = fallbackIDs
	}
	seen := map[int64]bool{}
	out := []int64{}
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		if _, ok := findMaterial(data, id); ok {
			out = append(out, id)
			seen[id] = true
		}
	}
	return out
}

func (a *App) createPlantBufferTransfer(w http.ResponseWriter, r *http.Request, session Session) {
	var req plantBufferTransferRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid plant buffer transfer")
		return
	}
	var flow PlantBufferFlow
	err := a.store.Mutate(func(data *AppData) error {
		buffer, index, err := plantBufferFromRequest(*data, session.User, req.BufferID, req.BufferCode)
		if err != nil {
			return err
		}
		var yardPile StockYardPile
		useYardPile := req.YardPileID != 0 || strings.TrimSpace(req.YardPileCode) != ""
		if useYardPile {
			yardPile, _, err = stockYardPileFromRequest(*data, session.User, req.YardPileID, req.YardPileCode)
			if err != nil {
				return err
			}
			if yardPile.SiteID != buffer.SiteID {
				return fmt.Errorf("堆位和筒仓必须属于同一站点")
			}
			if yardPile.MaterialID == 0 {
				return fmt.Errorf("堆位未绑定物料")
			}
		}
		materialID := nonZeroInt(req.MaterialID, buffer.MaterialID)
		if useYardPile && materialID == 0 {
			materialID = yardPile.MaterialID
		}
		if materialID == 0 {
			return fmt.Errorf("请选择上料物料")
		}
		if useYardPile && yardPile.MaterialID != materialID {
			return fmt.Errorf("堆位物料与上料物料不一致")
		}
		if !plantBufferMaterialAllowed(buffer, materialID) {
			return fmt.Errorf("物料不允许进入该仓位")
		}
		if _, ok := findMaterial(*data, materialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		quantity := round(req.Quantity)
		if quantity <= 0 {
			return fmt.Errorf("上料数量必须大于 0")
		}
		if buffer.CurrentQty+quantity > buffer.Capacity {
			return fmt.Errorf("上料后超过仓位容量")
		}
		var balance float64
		var yardFlow StockYardFlow
		if useYardPile {
			_, yardFlow, err = consumeStockYardPile(data, yardPile.ID, "", quantity, session.User.Username, fallback(req.Remark, "堆场堆位上料至筒仓"))
			if err != nil {
				return err
			}
		} else {
			var ok bool
			balance, _, ok = decreaseInventory(data, buffer.SiteID, materialID, quantity)
			if !ok {
				return fmt.Errorf("正式库存不足")
			}
		}
		buffer.MaterialID = materialID
		buffer.CurrentQty = round(buffer.CurrentQty + quantity)
		buffer.Unit = fallback(req.Unit, buffer.Unit)
		buffer.UpdatedAt = nowString()
		data.PlantBufferLocations[index] = buffer
		if useYardPile {
			flow = appendPlantBufferFlow(data, buffer, "stock_yard_pile", yardFlow.ID, "in", quantity, session.User.Username, fallback(req.Remark, "堆场堆位上料至筒仓"), nowString())
		} else {
			flow = appendPlantBufferFlow(data, buffer, "buffer_transfer", 0, "in", quantity, session.User.Username, fallback(req.Remark, "正式库存上料至筒仓"), nowString())
			data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
				ID: nextID(data, "inventoryFlow"), FlowNo: number("IF", data.Next["inventoryFlow"]),
				SiteID: buffer.SiteID, MaterialID: materialID, SourceType: "plant_buffer_transfer",
				SourceID: flow.ID, Direction: "out", Quantity: quantity, BalanceQty: balance,
				Remark: "筒仓上料", CreatedAt: flow.CreatedAt,
			})
		}
		addAudit(data, session.User.Username, "transfer", "plant_buffer", flow.ID, flow.BufferCode, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, flow, "production.buffer.transfer.created")
}

func (a *App) createPlantBufferAdjustment(w http.ResponseWriter, r *http.Request, session Session) {
	var req plantBufferAdjustmentRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid plant buffer adjustment")
		return
	}
	var response interface{}
	topic := "production.buffer.adjusted"
	err := a.store.Mutate(func(data *AppData) error {
		buffer, index, err := plantBufferFromRequest(*data, session.User, req.BufferID, req.BufferCode)
		if err != nil {
			return err
		}
		req.BufferID = buffer.ID
		req.BufferCode = buffer.Code
		_, instances, err := publishPlantBufferAdjustmentWorkflow(data, buffer, req, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			response = instances[0]
			topic = "production.buffer.adjustment.workflow_requested"
			addAudit(data, session.User.Username, "request_adjustment", "plant_buffer", buffer.ID, buffer.Code, clientIP(r))
			return nil
		}
		flow, err := applyPlantBufferAdjustmentLocked(data, data.PlantBufferLocations[index].ID, req, 0, session.User.Username)
		if err != nil {
			return err
		}
		response = flow
		addAudit(data, session.User.Username, "adjust", "plant_buffer", flow.ID, flow.BufferCode, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, response, topic)
}

func publishPlantBufferAdjustmentWorkflow(data *AppData, buffer PlantBufferLocation, req plantBufferAdjustmentRequest, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	actual := round(req.ActualQty)
	delta := round(actual - buffer.CurrentQty)
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "plant_buffer_adjustment.requested",
		Source:     "production",
		Resource:   "plant_buffer_adjustment",
		ResourceID: buffer.ID,
		ResourceNo: buffer.Code,
		Title:      "筒仓盘点校正",
		Actor:      actor,
		Reason:     fallback(strings.TrimSpace(req.Remark), fmt.Sprintf("筒仓 %s 盘点 %.2f%s", buffer.Code, actual, fallback(buffer.Unit, "t"))),
		Variables: map[string]string{
			"bufferId":      fmt.Sprintf("%d", buffer.ID),
			"bufferCode":    buffer.Code,
			"plantId":       fmt.Sprintf("%d", buffer.PlantID),
			"actualQty":     fmt.Sprintf("%.4f", actual),
			"currentQty":    fmt.Sprintf("%.4f", buffer.CurrentQty),
			"deltaQty":      fmt.Sprintf("%.4f", delta),
			"unit":          fallback(buffer.Unit, "t"),
			"moistureRate":  fmt.Sprintf("%.4f", req.MoistureRate),
			"qualityStatus": strings.TrimSpace(req.QualityStatus),
			"status":        strings.TrimSpace(req.Status),
			"remark":        strings.TrimSpace(req.Remark),
		},
	})
}

func applyPlantBufferAdjustmentLocked(data *AppData, bufferID int64, req plantBufferAdjustmentRequest, sourceID int64, actor string) (PlantBufferFlow, error) {
	buffer, index, ok := findPlantBufferByID(*data, bufferID)
	if !ok {
		return PlantBufferFlow{}, fmt.Errorf("暂存仓位不存在")
	}
	actual := round(req.ActualQty)
	if actual < 0 {
		return PlantBufferFlow{}, fmt.Errorf("盘点数量不能小于 0")
	}
	if actual > buffer.Capacity {
		return PlantBufferFlow{}, fmt.Errorf("盘点数量超过仓位容量")
	}
	delta := round(actual - buffer.CurrentQty)
	direction := "adjustment"
	if delta > 0 {
		direction = "adjustment_in"
	} else if delta < 0 {
		direction = "adjustment_out"
	}
	buffer.CurrentQty = actual
	buffer.MoistureRate = nonZero(req.MoistureRate, buffer.MoistureRate)
	buffer.QualityStatus = fallback(req.QualityStatus, buffer.QualityStatus)
	buffer.Status = fallback(req.Status, buffer.Status)
	buffer.UpdatedAt = nowString()
	data.PlantBufferLocations[index] = buffer
	return appendPlantBufferFlow(data, buffer, "buffer_adjustment", sourceID, direction, abs(delta), actor, fallback(req.Remark, "筒仓盘点校正"), nowString()), nil
}

func (a *App) ingestBufferProtocolFrame(w http.ResponseWriter, r *http.Request, session Session) {
	req, err := readProtocolFrameIngestRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Channel = fallback(strings.TrimSpace(req.Channel), "industrial-control-gateway")
	req.Protocol = fallback(strings.TrimSpace(req.Protocol), "buffer-json")
	response, err := a.processBufferProtocolFrame(r, session, req)
	if err != nil {
		writeProtocolFrameError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) processBufferProtocolFrame(r *http.Request, session Session, req protocolFrameIngestRequest) (protocolIngestResponse, error) {
	payload, err := parseBufferProtocolFrame(req)
	frame := baseProtocolFrame(session, req, "plant_buffer")
	frame.DeviceNo = fallback(payload.DeviceNo, fallback(payload.PlantCode, deviceNoFromSession(session)))
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	flow, err := a.recordBufferLevelReport(r, session, req, payload)
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	frame.ParsedID = flow.BufferID
	frame.Status = "accepted"
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		return protocolIngestResponse{}, err
	}
	a.emit("production.buffer.reported", flow)
	a.emit("device.protocol.frame", saved)
	return protocolIngestResponse{Frame: saved, PlantBufferFlow: &flow}, nil
}

func parseBufferProtocolFrame(req protocolFrameIngestRequest) (bufferLevelPayload, error) {
	var payload bufferLevelPayload
	if len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return payload, fmt.Errorf("invalid buffer json payload")
		}
		return normalizeBufferPayload(payload), nil
	}
	raw := strings.TrimSpace(req.Raw)
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return payload, fmt.Errorf("invalid buffer json frame")
		}
		return normalizeBufferPayload(payload), nil
	}
	return payload, fmt.Errorf("buffer frame requires json payload")
}

func normalizeBufferPayload(payload bufferLevelPayload) bufferLevelPayload {
	payload.DeviceNo = strings.TrimSpace(payload.DeviceNo)
	payload.PlantCode = strings.TrimSpace(payload.PlantCode)
	payload.BufferCode = strings.TrimSpace(fallback(payload.BufferCode, payload.BinCode))
	payload.QualityStatus = strings.TrimSpace(payload.QualityStatus)
	payload.Status = strings.TrimSpace(payload.Status)
	payload.ReportedAt = strings.TrimSpace(payload.ReportedAt)
	return payload
}

func (a *App) recordBufferLevelReport(r *http.Request, session Session, req protocolFrameIngestRequest, payload bufferLevelPayload) (PlantBufferFlow, error) {
	var flow PlantBufferFlow
	err := a.store.Mutate(func(data *AppData) error {
		if payload.BufferCode == "" {
			return fmt.Errorf("仓位编码不能为空")
		}
		buffer, index, ok := findPlantBufferByCode(*data, payload.BufferCode)
		if !ok {
			return fmt.Errorf("暂存仓位不存在")
		}
		if _, err := writableSiteID(*data, session.User, buffer.SiteID); err != nil {
			return err
		}
		if payload.PlantCode != "" && !strings.EqualFold(payload.PlantCode, buffer.PlantCode) {
			return fmt.Errorf("上报生产线与仓位不匹配")
		}
		if payload.MaterialID != 0 {
			if !plantBufferMaterialAllowed(buffer, payload.MaterialID) {
				return fmt.Errorf("上报物料不允许进入该仓位")
			}
			buffer.MaterialID = payload.MaterialID
		}
		if buffer.MaterialID == 0 {
			return fmt.Errorf("仓位未绑定物料")
		}
		if payload.Quantity < 0 {
			return fmt.Errorf("仓位数量不能小于 0")
		}
		if payload.Quantity > buffer.Capacity {
			return fmt.Errorf("仓位数量超过容量")
		}
		reportedAt := fallback(payload.ReportedAt, nowString())
		delta := round(payload.Quantity - buffer.CurrentQty)
		buffer.CurrentQty = round(payload.Quantity)
		buffer.MoistureRate = nonZero(payload.MoistureRate, buffer.MoistureRate)
		buffer.QualityStatus = fallback(payload.QualityStatus, buffer.QualityStatus)
		buffer.Status = fallback(payload.Status, buffer.Status)
		buffer.GatewayDeviceNo = fallback(payload.DeviceNo, buffer.GatewayDeviceNo)
		buffer.GatewayChannel = req.Channel
		buffer.GatewayProtocol = req.Protocol
		buffer.LastReportedAt = reportedAt
		buffer.LastFrameStatus = "accepted"
		buffer.GatewayError = ""
		buffer.UpdatedAt = nowString()
		data.PlantBufferLocations[index] = buffer
		direction := "level_report"
		if delta > 0 {
			direction = "level_in"
		} else if delta < 0 {
			direction = "level_out"
		}
		flow = appendPlantBufferFlow(data, buffer, "buffer_level_report", 0, direction, abs(delta), session.User.Username, "网关料位上报", reportedAt)
		addAudit(data, session.User.Username, "protocol_ingest", "plant_buffer", buffer.ID, buffer.Code, clientIP(r))
		return nil
	})
	return flow, err
}

func appendPlantBufferFlow(data *AppData, buffer PlantBufferLocation, sourceType string, sourceID int64, direction string, quantity float64, actor string, remark string, createdAt string) PlantBufferFlow {
	id := nextID(data, "plantBufferFlow")
	flow := PlantBufferFlow{
		ID: id, FlowNo: number("PBF", id), SiteID: buffer.SiteID, PlantID: buffer.PlantID,
		BufferID: buffer.ID, BufferCode: buffer.Code, MaterialID: buffer.MaterialID,
		SourceType: sourceType, SourceID: sourceID, Direction: direction, Quantity: round(quantity),
		BalanceQty: buffer.CurrentQty, Unit: fallback(buffer.Unit, "t"), MoistureRate: buffer.MoistureRate,
		QualityStatus: buffer.QualityStatus, Actor: fallback(actor, "system"), Remark: remark, CreatedAt: fallback(createdAt, nowString()),
	}
	data.PlantBufferFlows = append(data.PlantBufferFlows, flow)
	return flow
}

func consumeBatchBufferMaterial(data *AppData, batch ProductionBatch, material productionMaterialConsumption) (bool, error) {
	bufferCode := strings.TrimSpace(material.BufferCode)
	if bufferCode == "" {
		return false, nil
	}
	buffer, index, ok := findPlantBufferByCode(*data, bufferCode)
	if !ok {
		return true, fmt.Errorf("暂存仓位不存在")
	}
	if !strings.EqualFold(buffer.PlantCode, batch.PlantCode) {
		return true, fmt.Errorf("暂存仓位不属于生产批次生产线")
	}
	required := round(material.Quantity)
	if required <= 0 {
		return true, fmt.Errorf("物料消耗量必须大于 0")
	}
	if buffer.MaterialID != material.MaterialID {
		return true, fmt.Errorf("暂存仓位物料不匹配")
	}
	if buffer.QualityStatus != "" && buffer.QualityStatus != "passed" {
		return true, fmt.Errorf("暂存仓位质量状态不可用")
	}
	if buffer.Status != "active" && buffer.Status != "running" {
		return true, fmt.Errorf("暂存仓位状态不可用")
	}
	if buffer.CurrentQty < required {
		return true, fmt.Errorf("暂存仓位余额不足")
	}
	buffer.CurrentQty = round(buffer.CurrentQty - required)
	buffer.UpdatedAt = nowString()
	data.PlantBufferLocations[index] = buffer
	appendPlantBufferFlow(data, buffer, "production_batch", batch.ID, "out", required, "system", "生产批次从筒仓扣减", batch.CompletedAt)
	return true, nil
}

func plantBufferFromRequest(data AppData, user User, id int64, code string) (PlantBufferLocation, int, error) {
	var buffer PlantBufferLocation
	var index int
	var ok bool
	if id != 0 {
		buffer, index, ok = findPlantBufferByID(data, id)
	} else if strings.TrimSpace(code) != "" {
		buffer, index, ok = findPlantBufferByCode(data, code)
	}
	if !ok {
		return buffer, 0, fmt.Errorf("暂存仓位不存在")
	}
	if _, err := writableSiteID(data, user, buffer.SiteID); err != nil {
		return buffer, 0, err
	}
	return buffer, index, nil
}

func findPlantByID(data AppData, id int64) (Plant, bool) {
	for _, item := range data.Plants {
		if item.ID == id {
			return item, true
		}
	}
	return Plant{}, false
}

func findPlantBufferByID(data AppData, id int64) (PlantBufferLocation, int, bool) {
	for i, item := range data.PlantBufferLocations {
		if item.ID == id {
			return item, i, true
		}
	}
	return PlantBufferLocation{}, 0, false
}

func findPlantBufferByCode(data AppData, code string) (PlantBufferLocation, int, bool) {
	code = strings.TrimSpace(code)
	for i, item := range data.PlantBufferLocations {
		if strings.EqualFold(item.Code, code) {
			return item, i, true
		}
	}
	return PlantBufferLocation{}, 0, false
}

func findPlantBufferForProfile(data AppData, plantID, bufferID int64, bufferCode string) (PlantBufferLocation, bool) {
	var buffer PlantBufferLocation
	var ok bool
	if bufferID != 0 {
		buffer, _, ok = findPlantBufferByID(data, bufferID)
	} else {
		buffer, _, ok = findPlantBufferByCode(data, bufferCode)
	}
	if !ok || buffer.PlantID != plantID {
		return PlantBufferLocation{}, false
	}
	return buffer, true
}

func plantBufferCodeExists(items []PlantBufferLocation, code string, exceptID int64) bool {
	for _, item := range items {
		if item.ID != exceptID && strings.EqualFold(item.Code, code) {
			return true
		}
	}
	return false
}

func plantBufferMaterialAllowed(buffer PlantBufferLocation, materialID int64) bool {
	if len(buffer.AllowedMaterialIDs) == 0 {
		return true
	}
	for _, id := range buffer.AllowedMaterialIDs {
		if id == materialID {
			return true
		}
	}
	return false
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
