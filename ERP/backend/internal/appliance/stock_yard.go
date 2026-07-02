package appliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type stockYardReceiptRequest struct {
	PileID        int64   `json:"pileId"`
	PileCode      string  `json:"pileCode"`
	MaterialID    int64   `json:"materialId"`
	SupplierID    int64   `json:"supplierId"`
	BatchNo       string  `json:"batchNo"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	MoistureRate  float64 `json:"moistureRate"`
	QualityStatus string  `json:"qualityStatus"`
	Remark        string  `json:"remark"`
}

type stockYardAdjustmentRequest struct {
	PileID        int64   `json:"pileId"`
	PileCode      string  `json:"pileCode"`
	ActualQty     float64 `json:"actualQty"`
	MoistureRate  float64 `json:"moistureRate"`
	QualityStatus string  `json:"qualityStatus"`
	Status        string  `json:"status"`
	Remark        string  `json:"remark"`
}

type yardLevelPayload struct {
	DeviceNo      string  `json:"deviceNo"`
	YardCode      string  `json:"yardCode"`
	PileCode      string  `json:"pileCode"`
	MaterialID    int64   `json:"materialId"`
	Quantity      float64 `json:"quantity"`
	MoistureRate  float64 `json:"moistureRate"`
	QualityStatus string  `json:"qualityStatus"`
	Status        string  `json:"status"`
	ReportedAt    string  `json:"reportedAt"`
}

func (a *App) createStockYard(w http.ResponseWriter, r *http.Request, session Session) {
	var item StockYard
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid stock yard")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		normalized, err := normalizeStockYard(*data, session.User, item, StockYard{})
		if err != nil {
			return err
		}
		normalized.ID = nextID(data, "stockYard")
		normalized.CreatedAt = nowString()
		normalized.UpdatedAt = normalized.CreatedAt
		data.StockYards = append(data.StockYards, normalized)
		item = normalized
		addAudit(data, session.User.Username, "create", "stock_yard", item.ID, item.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.stock_yard.created")
}

func (a *App) updateStockYard(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item StockYard
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid stock yard")
		return
	}
	var updated StockYard
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.StockYards {
			if data.StockYards[i].ID != id {
				continue
			}
			normalized, err := normalizeStockYard(*data, session.User, item, data.StockYards[i])
			if err != nil {
				return err
			}
			normalized.ID = id
			normalized.CreatedAt = data.StockYards[i].CreatedAt
			normalized.UpdatedAt = nowString()
			data.StockYards[i] = normalized
			updated = normalized
			addAudit(data, session.User.Username, "update", "stock_yard", id, normalized.Code, clientIP(r))
			return nil
		}
		return fmt.Errorf("堆场不存在")
	})
	a.respondUpdate(w, err, updated, "master.stock_yard.updated")
}

func (a *App) createStockYardPile(w http.ResponseWriter, r *http.Request, session Session) {
	var item StockYardPile
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid stock yard pile")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		normalized, err := normalizeStockYardPile(*data, session.User, item, StockYardPile{})
		if err != nil {
			return err
		}
		normalized.ID = nextID(data, "stockYardPile")
		normalized.CreatedAt = nowString()
		normalized.UpdatedAt = normalized.CreatedAt
		data.StockYardPiles = append(data.StockYardPiles, normalized)
		item = normalized
		addAudit(data, session.User.Username, "create", "stock_yard_pile", item.ID, item.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.stock_yard_pile.created")
}

func (a *App) updateStockYardPile(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item StockYardPile
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid stock yard pile")
		return
	}
	var updated StockYardPile
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.StockYardPiles {
			if data.StockYardPiles[i].ID != id {
				continue
			}
			normalized, err := normalizeStockYardPile(*data, session.User, item, data.StockYardPiles[i])
			if err != nil {
				return err
			}
			normalized.ID = id
			normalized.CurrentQty = data.StockYardPiles[i].CurrentQty
			normalized.CreatedAt = data.StockYardPiles[i].CreatedAt
			normalized.UpdatedAt = nowString()
			data.StockYardPiles[i] = normalized
			updated = normalized
			addAudit(data, session.User.Username, "update", "stock_yard_pile", id, normalized.Code, clientIP(r))
			return nil
		}
		return fmt.Errorf("堆位不存在")
	})
	a.respondUpdate(w, err, updated, "master.stock_yard_pile.updated")
}

func normalizeStockYard(data AppData, user User, item StockYard, current StockYard) (StockYard, error) {
	siteID := nonZeroInt(item.SiteID, current.SiteID)
	if user.Username == "import" {
		if _, ok := findSite(data, siteID); !ok {
			return item, fmt.Errorf("站点不存在")
		}
	} else {
		var err error
		siteID, err = writableSiteID(data, user, siteID)
		if err != nil {
			return item, err
		}
	}
	item.SiteID = siteID
	item.Code = strings.TrimSpace(fallback(item.Code, current.Code))
	item.Name = strings.TrimSpace(fallback(item.Name, current.Name))
	if item.Code == "" || item.Name == "" {
		return item, fmt.Errorf("堆场名称和编码不能为空")
	}
	if stockYardCodeExists(data.StockYards, item.Code, current.ID) {
		return item, fmt.Errorf("堆场编码已存在")
	}
	item.Type = fallback(strings.TrimSpace(item.Type), fallback(current.Type, "aggregate_yard"))
	item.Area = fallback(strings.TrimSpace(item.Area), current.Area)
	item.Capacity = nonZero(item.Capacity, current.Capacity)
	if item.Capacity <= 0 {
		return item, fmt.Errorf("堆场容量必须大于 0")
	}
	item.Unit = fallback(strings.TrimSpace(item.Unit), fallback(current.Unit, "t"))
	item.Status = fallback(strings.TrimSpace(item.Status), fallback(current.Status, "active"))
	item.GatewayDeviceNo = fallback(strings.TrimSpace(item.GatewayDeviceNo), current.GatewayDeviceNo)
	item.LastReportedAt = current.LastReportedAt
	return item, nil
}

func normalizeStockYardPile(data AppData, user User, item StockYardPile, current StockYardPile) (StockYardPile, error) {
	if item.YardID == 0 {
		item.YardID = current.YardID
	}
	yard, ok := findStockYardByID(data, item.YardID)
	if !ok {
		return item, fmt.Errorf("堆场不存在")
	}
	siteID := nonZeroInt(item.SiteID, yard.SiteID)
	if user.Username == "import" {
		if _, ok := findSite(data, siteID); !ok {
			return item, fmt.Errorf("站点不存在")
		}
	} else {
		var err error
		siteID, err = writableSiteID(data, user, siteID)
		if err != nil {
			return item, err
		}
	}
	if siteID != yard.SiteID {
		return item, fmt.Errorf("堆位站点必须与堆场一致")
	}
	item.SiteID = siteID
	item.YardCode = yard.Code
	item.Code = strings.TrimSpace(fallback(item.Code, current.Code))
	item.Name = strings.TrimSpace(fallback(item.Name, current.Name))
	if item.Code == "" || item.Name == "" {
		return item, fmt.Errorf("堆位名称和编码不能为空")
	}
	if stockYardPileCodeExists(data.StockYardPiles, item.Code, current.ID) {
		return item, fmt.Errorf("堆位编码已存在")
	}
	item.MaterialID = nonZeroInt(item.MaterialID, current.MaterialID)
	if item.MaterialID != 0 {
		if _, ok := findMaterial(data, item.MaterialID); !ok {
			return item, fmt.Errorf("物料不存在")
		}
	}
	item.SupplierID = nonZeroInt(item.SupplierID, current.SupplierID)
	item.BatchNo = fallback(strings.TrimSpace(item.BatchNo), current.BatchNo)
	item.Capacity = nonZero(item.Capacity, current.Capacity)
	if item.Capacity <= 0 {
		return item, fmt.Errorf("堆位容量必须大于 0")
	}
	item.Unit = fallback(strings.TrimSpace(item.Unit), fallback(current.Unit, "t"))
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

func (a *App) createStockYardReceipt(w http.ResponseWriter, r *http.Request, session Session) {
	var req stockYardReceiptRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid stock yard receipt")
		return
	}
	var flow StockYardFlow
	err := a.store.Mutate(func(data *AppData) error {
		pile, index, err := stockYardPileFromRequest(*data, session.User, req.PileID, req.PileCode)
		if err != nil {
			return err
		}
		materialID := nonZeroInt(req.MaterialID, pile.MaterialID)
		if materialID == 0 {
			return fmt.Errorf("请选择入场物料")
		}
		if _, ok := findMaterial(*data, materialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		quantity := round(req.Quantity)
		if quantity <= 0 {
			return fmt.Errorf("入场数量必须大于 0")
		}
		if pile.CurrentQty+quantity > pile.Capacity {
			return fmt.Errorf("入场后超过堆位容量")
		}
		pile.MaterialID = materialID
		pile.SupplierID = nonZeroInt(req.SupplierID, pile.SupplierID)
		pile.BatchNo = fallback(strings.TrimSpace(req.BatchNo), pile.BatchNo)
		pile.Unit = fallback(req.Unit, pile.Unit)
		pile.CurrentQty = round(pile.CurrentQty + quantity)
		pile.MoistureRate = nonZero(req.MoistureRate, pile.MoistureRate)
		pile.QualityStatus = fallback(req.QualityStatus, pile.QualityStatus)
		pile.UpdatedAt = nowString()
		data.StockYardPiles[index] = pile
		flow = appendStockYardFlow(data, pile, "yard_receipt", 0, "in", quantity, session.User.Username, fallback(req.Remark, "堆场入场"), nowString())
		addAudit(data, session.User.Username, "receipt", "stock_yard_pile", flow.ID, flow.PileCode, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, flow, "procurement.stock_yard.receipt.created")
}

func (a *App) createStockYardAdjustment(w http.ResponseWriter, r *http.Request, session Session) {
	var req stockYardAdjustmentRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid stock yard adjustment")
		return
	}
	var response interface{}
	topic := "procurement.stock_yard.adjusted"
	err := a.store.Mutate(func(data *AppData) error {
		pile, index, err := stockYardPileFromRequest(*data, session.User, req.PileID, req.PileCode)
		if err != nil {
			return err
		}
		req.PileID = pile.ID
		req.PileCode = pile.Code
		_, instances, err := publishStockYardAdjustmentWorkflow(data, pile, req, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			response = instances[0]
			topic = "procurement.stock_yard.adjustment.workflow_requested"
			addAudit(data, session.User.Username, "request_adjustment", "stock_yard_pile", pile.ID, pile.Code, clientIP(r))
			return nil
		}
		flow, err := applyStockYardAdjustmentLocked(data, data.StockYardPiles[index].ID, req, 0, session.User.Username)
		if err != nil {
			return err
		}
		response = flow
		addAudit(data, session.User.Username, "adjust", "stock_yard_pile", flow.ID, flow.PileCode, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, response, topic)
}

func publishStockYardAdjustmentWorkflow(data *AppData, pile StockYardPile, req stockYardAdjustmentRequest, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	actual := round(req.ActualQty)
	delta := round(actual - pile.CurrentQty)
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "stock_yard_adjustment.requested",
		Source:     "procurement",
		Resource:   "stock_yard_adjustment",
		ResourceID: pile.ID,
		ResourceNo: pile.Code,
		Title:      "堆位盘点校正",
		Actor:      actor,
		Reason:     fallback(strings.TrimSpace(req.Remark), fmt.Sprintf("堆位 %s 盘点 %.2f%s", pile.Code, actual, fallback(pile.Unit, "t"))),
		Variables: map[string]string{
			"pileId":        fmt.Sprintf("%d", pile.ID),
			"pileCode":      pile.Code,
			"yardId":        fmt.Sprintf("%d", pile.YardID),
			"actualQty":     fmt.Sprintf("%.4f", actual),
			"currentQty":    fmt.Sprintf("%.4f", pile.CurrentQty),
			"deltaQty":      fmt.Sprintf("%.4f", delta),
			"unit":          fallback(pile.Unit, "t"),
			"moistureRate":  fmt.Sprintf("%.4f", req.MoistureRate),
			"qualityStatus": strings.TrimSpace(req.QualityStatus),
			"status":        strings.TrimSpace(req.Status),
			"remark":        strings.TrimSpace(req.Remark),
		},
	})
}

func applyStockYardAdjustmentLocked(data *AppData, pileID int64, req stockYardAdjustmentRequest, sourceID int64, actor string) (StockYardFlow, error) {
	pile, index, ok := findStockYardPileByID(*data, pileID)
	if !ok {
		return StockYardFlow{}, fmt.Errorf("堆位不存在")
	}
	actual := round(req.ActualQty)
	if actual < 0 {
		return StockYardFlow{}, fmt.Errorf("盘点数量不能小于 0")
	}
	if actual > pile.Capacity {
		return StockYardFlow{}, fmt.Errorf("盘点数量超过堆位容量")
	}
	delta := round(actual - pile.CurrentQty)
	direction := "adjustment"
	if delta > 0 {
		direction = "adjustment_in"
	} else if delta < 0 {
		direction = "adjustment_out"
	}
	pile.CurrentQty = actual
	pile.MoistureRate = nonZero(req.MoistureRate, pile.MoistureRate)
	pile.QualityStatus = fallback(req.QualityStatus, pile.QualityStatus)
	pile.Status = fallback(req.Status, pile.Status)
	pile.UpdatedAt = nowString()
	data.StockYardPiles[index] = pile
	return appendStockYardFlow(data, pile, "yard_adjustment", sourceID, direction, abs(delta), actor, fallback(req.Remark, "堆场盘点校正"), nowString()), nil
}

func consumeStockYardPile(data *AppData, pileID int64, pileCode string, quantity float64, actor string, remark string) (StockYardPile, StockYardFlow, error) {
	var pile StockYardPile
	var index int
	var ok bool
	if pileID != 0 {
		pile, index, ok = findStockYardPileByID(*data, pileID)
	} else {
		pile, index, ok = findStockYardPileByCode(*data, pileCode)
	}
	if !ok {
		return pile, StockYardFlow{}, fmt.Errorf("堆位不存在")
	}
	quantity = round(quantity)
	if quantity <= 0 {
		return pile, StockYardFlow{}, fmt.Errorf("堆位出料数量必须大于 0")
	}
	if pile.QualityStatus != "" && pile.QualityStatus != "passed" {
		return pile, StockYardFlow{}, fmt.Errorf("堆位质量状态不可用")
	}
	if pile.Status != "active" && pile.Status != "running" {
		return pile, StockYardFlow{}, fmt.Errorf("堆位状态不可用")
	}
	if pile.CurrentQty < quantity {
		return pile, StockYardFlow{}, fmt.Errorf("堆位余额不足")
	}
	pile.CurrentQty = round(pile.CurrentQty - quantity)
	pile.UpdatedAt = nowString()
	data.StockYardPiles[index] = pile
	flow := appendStockYardFlow(data, pile, "plant_buffer_transfer", 0, "out", quantity, actor, remark, nowString())
	return pile, flow, nil
}

func ingestYardLevelReport(data *AppData, session Session, req protocolFrameIngestRequest, payload yardLevelPayload) (StockYardFlow, error) {
	if payload.PileCode == "" {
		return StockYardFlow{}, fmt.Errorf("堆位编码不能为空")
	}
	pile, index, ok := findStockYardPileByCode(*data, payload.PileCode)
	if !ok {
		return StockYardFlow{}, fmt.Errorf("堆位不存在")
	}
	if _, err := writableSiteID(*data, session.User, pile.SiteID); err != nil {
		return StockYardFlow{}, err
	}
	if payload.YardCode != "" && !strings.EqualFold(payload.YardCode, pile.YardCode) {
		return StockYardFlow{}, fmt.Errorf("上报堆场与堆位不匹配")
	}
	if payload.MaterialID != 0 {
		if _, ok := findMaterial(*data, payload.MaterialID); !ok {
			return StockYardFlow{}, fmt.Errorf("物料不存在")
		}
		pile.MaterialID = payload.MaterialID
	}
	if pile.MaterialID == 0 {
		return StockYardFlow{}, fmt.Errorf("堆位未绑定物料")
	}
	if payload.Quantity < 0 {
		return StockYardFlow{}, fmt.Errorf("堆位数量不能小于 0")
	}
	if payload.Quantity > pile.Capacity {
		return StockYardFlow{}, fmt.Errorf("堆位数量超过容量")
	}
	reportedAt := fallback(payload.ReportedAt, nowString())
	delta := round(payload.Quantity - pile.CurrentQty)
	pile.CurrentQty = round(payload.Quantity)
	pile.MoistureRate = nonZero(payload.MoistureRate, pile.MoistureRate)
	pile.QualityStatus = fallback(payload.QualityStatus, pile.QualityStatus)
	pile.Status = fallback(payload.Status, pile.Status)
	pile.GatewayDeviceNo = fallback(payload.DeviceNo, pile.GatewayDeviceNo)
	pile.GatewayChannel = req.Channel
	pile.GatewayProtocol = req.Protocol
	pile.LastReportedAt = reportedAt
	pile.LastFrameStatus = "accepted"
	pile.GatewayError = ""
	pile.UpdatedAt = nowString()
	data.StockYardPiles[index] = pile
	for i := range data.StockYards {
		if data.StockYards[i].ID == pile.YardID {
			data.StockYards[i].GatewayDeviceNo = fallback(payload.DeviceNo, data.StockYards[i].GatewayDeviceNo)
			data.StockYards[i].LastReportedAt = reportedAt
			data.StockYards[i].UpdatedAt = nowString()
			break
		}
	}
	direction := "level_report"
	if delta > 0 {
		direction = "level_in"
	} else if delta < 0 {
		direction = "level_out"
	}
	return appendStockYardFlow(data, pile, "yard_level_report", 0, direction, abs(delta), session.User.Username, "网关堆场料位上报", reportedAt), nil
}

func (a *App) ingestYardProtocolFrame(w http.ResponseWriter, r *http.Request, session Session) {
	req, err := readProtocolFrameIngestRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Channel = fallback(strings.TrimSpace(req.Channel), "industrial-control-gateway")
	req.Protocol = fallback(strings.TrimSpace(req.Protocol), "yard-json")
	response, err := a.processYardProtocolFrame(r, session, req)
	if err != nil {
		writeProtocolFrameError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) processYardProtocolFrame(r *http.Request, session Session, req protocolFrameIngestRequest) (protocolIngestResponse, error) {
	payload, err := parseYardProtocolFrame(req)
	frame := baseProtocolFrame(session, req, "stock_yard_pile")
	frame.DeviceNo = fallback(payload.DeviceNo, fallback(payload.YardCode, deviceNoFromSession(session)))
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	var flow StockYardFlow
	err = a.store.Mutate(func(data *AppData) error {
		var err error
		flow, err = ingestYardLevelReport(data, session, req, payload)
		return err
	})
	if err != nil {
		return protocolIngestResponse{}, a.rejectProtocolFrameData(r, frame, err)
	}
	frame.ParsedID = flow.PileID
	frame.Status = "accepted"
	saved, err := a.recordDeviceProtocolFrame(r, frame)
	if err != nil {
		return protocolIngestResponse{}, err
	}
	a.emit("procurement.stock_yard.reported", flow)
	a.emit("device.protocol.frame", saved)
	return protocolIngestResponse{Frame: saved, StockYardFlow: &flow}, nil
}

func appendStockYardFlow(data *AppData, pile StockYardPile, sourceType string, sourceID int64, direction string, quantity float64, actor string, remark string, createdAt string) StockYardFlow {
	id := nextID(data, "stockYardFlow")
	flow := StockYardFlow{
		ID: id, FlowNo: number("SYF", id), SiteID: pile.SiteID, YardID: pile.YardID, PileID: pile.ID,
		PileCode: pile.Code, MaterialID: pile.MaterialID, SourceType: sourceType, SourceID: sourceID,
		Direction: direction, Quantity: round(quantity), BalanceQty: pile.CurrentQty, Unit: fallback(pile.Unit, "t"),
		MoistureRate: pile.MoistureRate, QualityStatus: pile.QualityStatus, Actor: fallback(actor, "system"),
		Remark: remark, CreatedAt: fallback(createdAt, nowString()),
	}
	data.StockYardFlows = append(data.StockYardFlows, flow)
	return flow
}

func stockYardPileFromRequest(data AppData, user User, id int64, code string) (StockYardPile, int, error) {
	var pile StockYardPile
	var index int
	var ok bool
	if id != 0 {
		pile, index, ok = findStockYardPileByID(data, id)
	} else if strings.TrimSpace(code) != "" {
		pile, index, ok = findStockYardPileByCode(data, code)
	}
	if !ok {
		return pile, 0, fmt.Errorf("堆位不存在")
	}
	if _, err := writableSiteID(data, user, pile.SiteID); err != nil {
		return pile, 0, err
	}
	return pile, index, nil
}

func findStockYardByID(data AppData, id int64) (StockYard, bool) {
	for _, item := range data.StockYards {
		if item.ID == id {
			return item, true
		}
	}
	return StockYard{}, false
}

func findStockYardByCode(data AppData, code string) (StockYard, int, bool) {
	for i, item := range data.StockYards {
		if strings.EqualFold(item.Code, strings.TrimSpace(code)) {
			return item, i, true
		}
	}
	return StockYard{}, 0, false
}

func findStockYardPileByID(data AppData, id int64) (StockYardPile, int, bool) {
	for i, item := range data.StockYardPiles {
		if item.ID == id {
			return item, i, true
		}
	}
	return StockYardPile{}, 0, false
}

func findStockYardPileByCode(data AppData, code string) (StockYardPile, int, bool) {
	for i, item := range data.StockYardPiles {
		if strings.EqualFold(item.Code, strings.TrimSpace(code)) {
			return item, i, true
		}
	}
	return StockYardPile{}, 0, false
}

func stockYardCodeExists(items []StockYard, code string, exceptID int64) bool {
	for _, item := range items {
		if item.ID != exceptID && strings.EqualFold(item.Code, code) {
			return true
		}
	}
	return false
}

func stockYardPileCodeExists(items []StockYardPile, code string, exceptID int64) bool {
	for _, item := range items {
		if item.ID != exceptID && strings.EqualFold(item.Code, code) {
			return true
		}
	}
	return false
}

func parseYardProtocolFrame(req protocolFrameIngestRequest) (yardLevelPayload, error) {
	var payload yardLevelPayload
	if len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			return payload, fmt.Errorf("invalid yard json payload")
		}
		return normalizeYardPayload(payload), nil
	}
	raw := strings.TrimSpace(req.Raw)
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return payload, fmt.Errorf("invalid yard json frame")
		}
		return normalizeYardPayload(payload), nil
	}
	return payload, fmt.Errorf("yard frame requires json payload")
}

func normalizeYardPayload(payload yardLevelPayload) yardLevelPayload {
	payload.DeviceNo = strings.TrimSpace(payload.DeviceNo)
	payload.YardCode = strings.TrimSpace(payload.YardCode)
	payload.PileCode = strings.TrimSpace(payload.PileCode)
	payload.QualityStatus = strings.TrimSpace(payload.QualityStatus)
	payload.Status = strings.TrimSpace(payload.Status)
	payload.ReportedAt = strings.TrimSpace(payload.ReportedAt)
	return payload
}
