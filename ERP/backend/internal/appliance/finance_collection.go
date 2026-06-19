package appliance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (a *App) generateCollectionTasks(w http.ResponseWriter, r *http.Request, session Session) {
	var created []CollectionTask
	err := a.store.Mutate(func(data *AppData) error {
		today, _ := time.Parse("2006-01-02", todayString())
		for _, receivable := range data.Receivables {
			remaining := remainingReceivable(receivable)
			if remaining <= 0 || receivable.Status == "paid" || openCollectionTaskExists(*data, receivable.ID) {
				continue
			}
			customer, _ := findCustomer(*data, receivable.CustomerID)
			days := collectionOverdueDays(today, receivable.DueDate)
			taskID := nextID(data, "collectionTask")
			task := CollectionTask{
				ID:           taskID,
				TaskNo:       number("COL", taskID),
				ReceivableID: receivable.ID,
				CustomerID:   receivable.CustomerID,
				CustomerName: customer.Name,
				Amount:       remaining,
				DueDate:      receivable.DueDate,
				OverdueDays:  days,
				Level:        collectionLevel(days),
				Channel:      collectionChannel(days),
				Status:       "open",
				Message:      collectionMessage(customer.Name, remaining, days),
				GeneratedAt:  nowString(),
			}
			data.CollectionTasks = append(data.CollectionTasks, task)
			created = append(created, task)
			addAudit(data, session.User.Username, "generate", "collection_task", task.ID, task.TaskNo, clientIP(r))
		}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("finance.collection.generated", created)
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) handleCollectionTask(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Remark string `json:"remark"`
	}
	_ = readJSON(r, &req)
	var task CollectionTask
	err := a.store.Mutate(func(data *AppData) error {
		index := collectionTaskIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("催收任务不存在")
		}
		data.CollectionTasks[index].Status = "handled"
		data.CollectionTasks[index].HandledBy = fallback(session.User.DisplayName, session.User.Username)
		data.CollectionTasks[index].HandledAt = nowString()
		data.CollectionTasks[index].Remark = req.Remark
		task = data.CollectionTasks[index]
		addAudit(data, session.User.Username, "handle", "collection_task", task.ID, task.TaskNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, task, "finance.collection.handled")
}

func (a *App) createCollectionTemplate(w http.ResponseWriter, r *http.Request, session Session) {
	var item CollectionTemplate
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid collection template")
		return
	}
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.Level = strings.TrimSpace(item.Level)
	item.Channel = strings.TrimSpace(item.Channel)
	item.Content = strings.TrimSpace(item.Content)
	if item.Name == "" || item.Level == "" || item.Channel == "" || item.Content == "" {
		writeError(w, http.StatusBadRequest, "template name, level, channel and content are required")
		return
	}
	if item.Code == "" {
		item.Code = strings.ToLower(item.Level + "_" + item.Channel)
	}
	item.Enabled = true
	item.UpdatedAt = nowString()
	err := a.store.Mutate(func(data *AppData) error {
		item.ID = nextID(data, "collectionTemplate")
		data.CollectionTemplates = append(data.CollectionTemplates, item)
		addAudit(data, session.User.Username, "create", "collection_template", item.ID, item.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.collection.template.created")
}

func (a *App) sendCollectionTask(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		TemplateID int64  `json:"templateId"`
		Channel    string `json:"channel"`
		Target     string `json:"target"`
	}
	_ = readJSON(r, &req)
	var dispatch CollectionDispatch
	err := a.store.Mutate(func(data *AppData) error {
		index := collectionTaskIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("催收任务不存在")
		}
		task := data.CollectionTasks[index]
		if task.Status != "open" {
			return fmt.Errorf("只有 open 状态的催收任务可以发送")
		}
		receivable, _ := findReceivable(*data, task.ReceivableID)
		customer, _ := findCustomer(*data, task.CustomerID)
		channel := strings.TrimSpace(req.Channel)
		if channel == "" {
			channel = task.Channel
		}
		template, ok := selectCollectionTemplate(*data, task, channel, req.TemplateID)
		if !ok {
			return fmt.Errorf("未找到可用催收模板")
		}
		if strings.TrimSpace(req.Channel) == "" && template.Channel != "" {
			channel = template.Channel
		}
		target := strings.TrimSpace(req.Target)
		if target == "" {
			target = defaultCollectionTarget(*data, customer, channel)
		}
		if target == "" {
			return fmt.Errorf("客户缺少可用催收目标")
		}
		endpoint, endpointIndex := ensureCollectionEndpoint(data, channel)
		content := renderCollectionTemplate(template.Content, task, receivable, customer)
		status := "delivered"
		errText := ""
		if endpoint.Status == "offline" || endpoint.Status == "disabled" || endpoint.Status == "degraded" {
			status = "failed"
			errText = "collection endpoint " + endpoint.Status
		}
		if endpointIndex >= 0 {
			data.IntegrationEndpoints[endpointIndex].LastSyncAt = nowString()
			if status == "delivered" {
				data.IntegrationEndpoints[endpointIndex].Status = "online"
			}
		}
		dispatchID := nextID(data, "collectionDispatch")
		dispatchNo := number("CD", dispatchID)
		dispatch = CollectionDispatch{
			ID:                dispatchID,
			DispatchNo:        dispatchNo,
			TaskID:            task.ID,
			TemplateID:        template.ID,
			CustomerID:        task.CustomerID,
			Channel:           channel,
			Target:            target,
			Content:           content,
			Endpoint:          endpoint.URL,
			ProviderRequestID: collectionProviderRequestID(dispatchNo),
			Status:            status,
			Error:             errText,
			SentAt:            nowString(),
			Actor:             fallback(session.User.DisplayName, session.User.Username),
		}
		data.CollectionDispatches = append(data.CollectionDispatches, dispatch)
		data.CollectionTasks[index].TemplateID = template.ID
		data.CollectionTasks[index].SendCount++
		data.CollectionTasks[index].LastSentAt = dispatch.SentAt
		data.Notifications = append(data.Notifications, Notification{
			ID:         nextID(data, "notification"),
			TargetRole: "boss",
			Channel:    "system",
			Title:      "催收发送" + dispatch.DispatchNo,
			Content:    content,
			Status:     "unread",
			CreatedAt:  dispatch.SentAt,
		})
		addAudit(data, session.User.Username, "send", "collection_task", task.ID, dispatch.DispatchNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, dispatch, "finance.collection.sent")
}

type collectionCallbackPayload struct {
	RequestID         string `json:"requestId"`
	DispatchNo        string `json:"dispatchNo"`
	ProviderMessageID string `json:"providerMessageId"`
	Status            string `json:"status"`
	Error             string `json:"error"`
	Provider          string `json:"provider"`
}

func (a *App) collectionCallback(w http.ResponseWriter, r *http.Request) {
	payload, err := parseCollectionCallback(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var updated CollectionDispatch
	err = a.store.Mutate(func(data *AppData) error {
		idx := collectionDispatchIndex(*data, payload)
		if idx < 0 {
			return fmt.Errorf("催收发送流水不存在")
		}
		status := normalizeCollectionCallbackStatus(payload.Status)
		if status == "" {
			return fmt.Errorf("催收回执状态无效")
		}
		data.CollectionDispatches[idx].Status = status
		data.CollectionDispatches[idx].ProviderMessageID = fallback(payload.ProviderMessageID, data.CollectionDispatches[idx].ProviderMessageID)
		data.CollectionDispatches[idx].CallbackAt = nowString()
		data.CollectionDispatches[idx].Error = payload.Error
		updated = data.CollectionDispatches[idx]
		addAudit(data, fallback(payload.Provider, "collection-provider"), "callback", "collection_dispatch", updated.ID, status, clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("finance.collection.callback", updated)
	writeJSON(w, http.StatusOK, updated)
}

func parseCollectionCallback(r *http.Request) (collectionCallbackPayload, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return collectionCallbackPayload{}, err
	}
	if err := verifyCollectionCallbackSignature(os.Getenv("CBMP_COLLECTION_CALLBACK_SECRET"), r.Header.Get("X-CBMP-Timestamp"), r.Header.Get("X-CBMP-Signature"), raw); err != nil {
		return collectionCallbackPayload{}, err
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return collectionCallbackPayload{}, fmt.Errorf("invalid collection callback payload: %w", err)
	}
	return collectionCallbackPayload{
		RequestID:         firstTaxGatewayString(decoded, "requestId", "request_id"),
		DispatchNo:        firstTaxGatewayString(decoded, "dispatchNo", "dispatch_no"),
		ProviderMessageID: firstTaxGatewayString(decoded, "providerMessageId", "messageId", "msgId"),
		Status:            firstTaxGatewayString(decoded, "status", "state"),
		Error:             firstTaxGatewayString(decoded, "error", "message", "msg"),
		Provider:          firstTaxGatewayString(decoded, "provider", "channelProvider"),
	}, nil
}

func verifyCollectionCallbackSignature(secret, timestamp, signature string, body []byte) error {
	if secret == "" {
		return fmt.Errorf("未配置催收回调密钥")
	}
	if timestamp == "" || signature == "" {
		return fmt.Errorf("催收回调缺少签名头")
	}
	parsed, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("催收回调时间戳无效")
	}
	if delta := time.Since(time.Unix(parsed, 0)); delta > 10*time.Minute || delta < -10*time.Minute {
		return fmt.Errorf("催收回调时间戳超出窗口")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	expected := "hmac-sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("催收回调签名无效")
	}
	return nil
}

func collectionDispatchIndex(data AppData, payload collectionCallbackPayload) int {
	for i, item := range data.CollectionDispatches {
		if payload.RequestID != "" && item.ProviderRequestID == payload.RequestID {
			return i
		}
		if payload.DispatchNo != "" && item.DispatchNo == payload.DispatchNo {
			return i
		}
		if payload.ProviderMessageID != "" && item.ProviderMessageID == payload.ProviderMessageID {
			return i
		}
	}
	return -1
}

func normalizeCollectionCallbackStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "delivered", "success", "ok":
		return "delivered"
	case "sent", "accepted", "pending", "processing":
		return "sent"
	case "read", "opened":
		return "read"
	case "clicked":
		return "clicked"
	case "failed", "rejected", "error", "undelivered":
		return "failed"
	default:
		return ""
	}
}

func collectionProviderRequestID(dispatchNo string) string {
	return fmt.Sprintf("collection-%s-%d", dispatchNo, time.Now().UnixNano())
}

func openCollectionTaskExists(data AppData, receivableID int64) bool {
	for _, item := range data.CollectionTasks {
		if item.ReceivableID == receivableID && item.Status == "open" {
			return true
		}
	}
	return false
}

func collectionTaskIndex(data AppData, id int64) int {
	for i := range data.CollectionTasks {
		if data.CollectionTasks[i].ID == id {
			return i
		}
	}
	return -1
}

func collectionOverdueDays(today time.Time, dueDate string) int {
	due, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return 0
	}
	days := int(today.Sub(due).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func collectionLevel(days int) string {
	switch {
	case days == 0:
		return "pre_due"
	case days <= 30:
		return "gentle"
	case days <= 60:
		return "urgent"
	default:
		return "legal"
	}
}

func collectionChannel(days int) string {
	if days > 60 {
		return "wecom"
	}
	if days > 30 {
		return "phone"
	}
	return "sms"
}

func collectionMessage(customerName string, amount float64, days int) string {
	if customerName == "" {
		customerName = "客户"
	}
	if days <= 0 {
		return fmt.Sprintf("%s 应收 %.2f 即将到期，请提前确认回款计划", customerName, amount)
	}
	return fmt.Sprintf("%s 应收 %.2f 已逾期 %d 天，请跟进催收", customerName, amount, days)
}

func selectCollectionTemplate(data AppData, task CollectionTask, channel string, templateID int64) (CollectionTemplate, bool) {
	if templateID > 0 {
		for _, item := range data.CollectionTemplates {
			if item.ID == templateID && item.Enabled {
				return item, true
			}
		}
		return CollectionTemplate{}, false
	}
	for _, item := range data.CollectionTemplates {
		if item.Enabled && item.Level == task.Level && item.Channel == channel {
			return item, true
		}
	}
	for _, item := range data.CollectionTemplates {
		if item.Enabled && item.Channel == channel {
			return item, true
		}
	}
	for _, item := range data.CollectionTemplates {
		if item.Enabled {
			return item, true
		}
	}
	return CollectionTemplate{}, false
}

func renderCollectionTemplate(content string, task CollectionTask, receivable Receivable, customer Customer) string {
	customerName := fallback(task.CustomerName, customer.Name)
	if customerName == "" {
		customerName = "客户"
	}
	amount := task.Amount
	if amount == 0 {
		amount = remainingReceivable(receivable)
	}
	replacer := strings.NewReplacer(
		"{{customerName}}", customerName,
		"{{amount}}", fmt.Sprintf("%.2f", amount),
		"{{dueDate}}", task.DueDate,
		"{{overdueDays}}", fmt.Sprintf("%d", task.OverdueDays),
		"{{taskNo}}", task.TaskNo,
		"{{billNo}}", receivable.BillNo,
	)
	return replacer.Replace(content)
}

func defaultCollectionTarget(data AppData, customer Customer, channel string) string {
	for _, contact := range data.CustomerContacts {
		if contact.CustomerID == customer.ID && contact.Status == "active" && contact.IsDefault && contact.Phone != "" {
			return collectionTargetValue(channel, contact.Phone, customer.ID)
		}
	}
	for _, contact := range data.CustomerContacts {
		if contact.CustomerID == customer.ID && contact.Status == "active" && contact.Phone != "" {
			return collectionTargetValue(channel, contact.Phone, customer.ID)
		}
	}
	return collectionTargetValue(channel, customer.Phone, customer.ID)
}

func collectionTargetValue(channel, phone string, customerID int64) string {
	switch channel {
	case "wecom":
		if phone != "" {
			return "wecom://" + phone
		}
		return fmt.Sprintf("wecom://customer/%d", customerID)
	default:
		return phone
	}
}

func ensureCollectionEndpoint(data *AppData, channel string) (IntegrationEndpoint, int) {
	for i, item := range data.IntegrationEndpoints {
		if item.Type == "collection" && item.Protocol == channel {
			return item, i
		}
	}
	endpoint := IntegrationEndpoint{
		ID:         nextID(data, "integration"),
		Name:       collectionEndpointName(channel),
		Type:       "collection",
		Protocol:   channel,
		URL:        "collection://local-simulator/" + channel,
		Status:     "online",
		LastSyncAt: nowString(),
	}
	data.IntegrationEndpoints = append(data.IntegrationEndpoints, endpoint)
	return endpoint, len(data.IntegrationEndpoints) - 1
}

func collectionEndpointName(channel string) string {
	switch channel {
	case "wecom":
		return "催收企微通道"
	case "phone":
		return "催收电话任务"
	default:
		return "催收短信通道"
	}
}
