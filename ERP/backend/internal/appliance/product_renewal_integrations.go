package appliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type productRenewalSyncRequest struct {
	Scenario     string
	ResourceType string
	ResourceID   int64
	ResourceNo   string
	Task         ProductRenewalTask
	Action       string
	Code         string
	Payload      interface{}
}

type productRenewalSyncCallbackPayload struct {
	SyncNo            string `json:"syncNo"`
	ExternalRequestID string `json:"externalRequestId"`
	ResourceType      string `json:"resourceType"`
	ResourceNo        string `json:"resourceNo"`
	Status            string `json:"status"`
	ExternalStatus    string `json:"externalStatus"`
	Error             string `json:"error"`
	FileURL           string `json:"fileUrl"`
	Signature         string `json:"signature"`
}

func (a *App) productRenewalSyncCallback(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid renewal sync callback body")
		return
	}
	var payload productRenewalSyncCallbackPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid renewal sync callback payload")
		return
	}
	var updated ProductRenewalSyncRecord
	err = a.store.Mutate(func(data *AppData) error {
		recordIndex := productRenewalSyncRecordIndex(*data, payload)
		if recordIndex < 0 {
			return fmt.Errorf("续费同步记录不存在")
		}
		record := &data.ProductRenewalSyncRecords[recordIndex]
		integration := currentProductRenewalIntegration(*data, *record)
		if integration != nil && integration.Secret != "" {
			if err := verifyTaxGatewayCallbackSignature(integration.Secret, r.Header.Get("X-CBMP-Timestamp"), r.Header.Get("X-CBMP-Signature"), raw); err != nil {
				return err
			}
		}
		now := nowString()
		status := normalizeProductRenewalCallbackStatus(payload.Status, payload.ExternalStatus, payload.Error)
		record.Status = status
		record.ExternalRequestID = fallback(strings.TrimSpace(payload.ExternalRequestID), record.ExternalRequestID)
		record.ExternalStatus = fallback(strings.ToLower(strings.TrimSpace(payload.ExternalStatus)), strings.ToLower(strings.TrimSpace(payload.Status)))
		record.ResponsePayload = string(raw)
		record.LastAttemptAt = now
		if status == "succeeded" {
			record.Error = ""
			record.NextRetryAt = ""
			record.CompletedAt = now
		} else if status == "failed" {
			record.Error = fallback(strings.TrimSpace(payload.Error), "外部系统回调失败")
			record.NextRetryAt = addMinutesString(now, 5*int(nonZeroInt(int64(record.AttemptCount), 1)))
			record.CompletedAt = ""
		}
		applyProductRenewalSyncSideEffects(data, *record)
		updated = *record
		addAudit(data, fallback(record.Provider, "renewal-provider"), "callback", "renewal_sync_record", record.ID, record.SyncNo+" "+record.Status, clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("product_ops.renewal.sync.callback", updated)
	writeJSON(w, http.StatusCreated, updated)
}

func normalizeProductRenewalIntegration(req ProductRenewalIntegration, actor string) (ProductRenewalIntegration, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	req.Scenario = strings.ToLower(strings.TrimSpace(req.Scenario))
	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.Token = strings.TrimSpace(req.Token)
	req.Secret = strings.TrimSpace(req.Secret)
	req.Status = fallback(strings.ToLower(strings.TrimSpace(req.Status)), "active")
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Name == "" {
		return req, fmt.Errorf("续费外部集成名称不能为空")
	}
	if req.Scenario == "" {
		req.Scenario = "all"
	}
	switch req.Scenario {
	case "all", "esign", "payment", "finance", "tax":
	default:
		return req, fmt.Errorf("续费集成场景必须是 all、esign、payment、finance 或 tax")
	}
	if req.Code == "" {
		req.Code = strings.ReplaceAll(strings.ToLower(req.Name), " ", "_")
	}
	if req.Provider == "" {
		req.Provider = req.Code
	}
	if req.RetryLimit <= 0 {
		req.RetryLimit = 3
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 3
	}
	if req.CreatedAt == "" {
		req.CreatedAt = nowString()
	}
	if req.CreatedBy == "" {
		req.CreatedBy = actor
	}
	return req, nil
}

func enqueueProductRenewalSyncRecord(data *AppData, req productRenewalSyncRequest) (ProductRenewalSyncRecord, error) {
	req.Scenario = strings.ToLower(strings.TrimSpace(req.Scenario))
	req.ResourceType = strings.ToLower(strings.TrimSpace(req.ResourceType))
	req.Action = fallback(strings.ToLower(strings.TrimSpace(req.Action)), "sync")
	id := nextID(data, "renewalSyncRecord")
	record := ProductRenewalSyncRecord{
		ID:           id,
		SyncNo:       number("RSY", id),
		Scenario:     fallback(req.Scenario, "all"),
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		ResourceNo:   req.ResourceNo,
		TaskID:       req.Task.ID,
		CustomerName: req.Task.CustomerName,
		Action:       req.Action,
		Status:       "pending",
		CreatedAt:    nowString(),
	}
	if raw, err := json.Marshal(req.Payload); err == nil {
		record.RequestPayload = string(raw)
	}
	if integration := selectProductRenewalIntegration(*data, record.Scenario, strings.ToLower(strings.TrimSpace(req.Code))); integration != nil {
		applyProductRenewalIntegrationToRecord(*integration, &record)
	}
	deliverProductRenewalSyncRecord(data, &record)
	applyProductRenewalSyncSideEffects(data, record)
	data.ProductRenewalSyncRecords = append(data.ProductRenewalSyncRecords, record)
	trimProductRenewalSyncRecords(data)
	return record, nil
}

func selectProductRenewalIntegration(data AppData, scenario, code string) *ProductRenewalIntegration {
	scenario = strings.ToLower(strings.TrimSpace(scenario))
	code = strings.ToLower(strings.TrimSpace(code))
	var fallbackMatch *ProductRenewalIntegration
	for i := range data.ProductRenewalIntegrations {
		item := &data.ProductRenewalIntegrations[i]
		if item.Status != "active" {
			continue
		}
		if !productRenewalIntegrationScenarioMatches(item.Scenario, scenario) {
			continue
		}
		if code != "" && (item.Code == code || item.Provider == code) {
			return item
		}
		if fallbackMatch == nil {
			fallbackMatch = item
		}
	}
	return fallbackMatch
}

func productRenewalIntegrationScenarioMatches(integrationScenario, scenario string) bool {
	integrationScenario = strings.ToLower(strings.TrimSpace(integrationScenario))
	scenario = strings.ToLower(strings.TrimSpace(scenario))
	if integrationScenario == "all" || integrationScenario == scenario {
		return true
	}
	return scenario == "payment" && integrationScenario == "finance"
}

func applyProductRenewalIntegrationToRecord(integration ProductRenewalIntegration, record *ProductRenewalSyncRecord) {
	record.IntegrationID = integration.ID
	record.IntegrationNo = integration.IntegrationNo
	record.IntegrationCode = integration.Code
	record.Provider = integration.Provider
}

func deliverProductRenewalSyncRecord(data *AppData, record *ProductRenewalSyncRecord) {
	now := nowString()
	record.AttemptCount++
	record.LastAttemptAt = now
	integration := currentProductRenewalIntegration(*data, *record)
	if integration == nil {
		record.Status = "failed"
		record.Error = "未配置可用的续费外部集成"
		record.NextRetryAt = addMinutesString(now, 5)
		return
	}
	applyProductRenewalIntegrationToRecord(*integration, record)
	endpoint := strings.TrimSpace(integration.Endpoint)
	if endpoint == "" {
		record.Status = "failed"
		record.Error = "续费外部集成未配置 endpoint"
		record.NextRetryAt = addMinutesString(now, 5)
		updateProductRenewalIntegrationState(data, record.IntegrationID, "", record.Error)
		return
	}
	status, response, err := postProductRenewalIntegrationPayload(*integration, *record)
	if err != nil {
		record.Status = "failed"
		record.Error = err.Error()
		record.ExternalStatus = status
		record.ResponsePayload = response
		record.NextRetryAt = addMinutesString(now, 5*record.AttemptCount)
		updateProductRenewalIntegrationState(data, record.IntegrationID, "", record.Error)
		return
	}
	record.Status = "succeeded"
	record.ExternalStatus = fallback(status, productRenewalExternalStatus(record.Scenario, record.Action))
	record.ExternalRequestID = fallback(record.ExternalRequestID, strings.ToLower(integration.Code)+"-"+record.SyncNo)
	record.ResponsePayload = response
	record.Error = ""
	record.NextRetryAt = ""
	record.CompletedAt = now
	updateProductRenewalIntegrationState(data, record.IntegrationID, now, "")
}

func currentProductRenewalIntegration(data AppData, record ProductRenewalSyncRecord) *ProductRenewalIntegration {
	for i := range data.ProductRenewalIntegrations {
		item := &data.ProductRenewalIntegrations[i]
		if item.Status != "active" {
			continue
		}
		if record.IntegrationID != 0 && item.ID == record.IntegrationID {
			return item
		}
		if record.IntegrationCode != "" && item.Code == record.IntegrationCode {
			return item
		}
	}
	return selectProductRenewalIntegration(data, record.Scenario, record.IntegrationCode)
}

func postProductRenewalIntegrationPayload(integration ProductRenewalIntegration, record ProductRenewalSyncRecord) (string, string, error) {
	endpoint := strings.TrimSpace(integration.Endpoint)
	switch endpoint {
	case "mock://success":
		status := productRenewalExternalStatus(record.Scenario, record.Action)
		response := map[string]string{
			"externalRequestId": strings.ToLower(integration.Code) + "-" + record.SyncNo,
			"externalStatus":    status,
			"provider":          integration.Provider,
		}
		raw, _ := json.Marshal(response)
		return status, string(raw), nil
	case "mock://fail":
		return "rejected", `{"error":"mock renewal integration failed"}`, fmt.Errorf("mock renewal integration failed")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", fmt.Errorf("不支持的续费集成 endpoint scheme: %s", parsed.Scheme)
	}
	payload := map[string]interface{}{
		"syncNo":            record.SyncNo,
		"scenario":          record.Scenario,
		"resourceType":      record.ResourceType,
		"resourceId":        record.ResourceID,
		"resourceNo":        record.ResourceNo,
		"taskId":            record.TaskID,
		"customerName":      record.CustomerName,
		"action":            record.Action,
		"requestPayload":    record.RequestPayload,
		"externalRequestId": record.ExternalRequestID,
		"createdAt":         record.CreatedAt,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}
	timeout := integration.TimeoutSeconds
	if timeout <= 0 {
		timeout = 3
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if integration.Token != "" {
		req.Header.Set("Authorization", "Bearer "+integration.Token)
	}
	if integration.Secret != "" {
		req.Header.Set("X-CBMP-Integration-Secret", integration.Secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	response := strings.TrimSpace(string(body))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.Status, response, fmt.Errorf("续费集成 endpoint 返回 %s", resp.Status)
	}
	return "accepted", response, nil
}

func productRenewalExternalStatus(scenario, action string) string {
	switch strings.ToLower(strings.TrimSpace(scenario)) {
	case "esign":
		if action == "complete" {
			return "signed"
		}
		return "sent"
	case "payment", "finance":
		return "confirmed"
	case "tax":
		return "accepted"
	default:
		return "accepted"
	}
}

func normalizeProductRenewalCallbackStatus(status, externalStatus, callbackError string) string {
	value := strings.ToLower(strings.TrimSpace(fallback(status, externalStatus)))
	if strings.TrimSpace(callbackError) != "" {
		return "failed"
	}
	switch value {
	case "success", "succeeded", "submitted", "accepted", "signed", "paid", "confirmed", "ok":
		return "succeeded"
	case "failed", "fail", "failure", "rejected", "error":
		return "failed"
	case "pending", "processing", "running":
		return "pending"
	default:
		return "succeeded"
	}
}

func updateProductRenewalIntegrationState(data *AppData, id int64, syncAt, lastError string) {
	if id == 0 {
		return
	}
	for i := range data.ProductRenewalIntegrations {
		if data.ProductRenewalIntegrations[i].ID != id {
			continue
		}
		if syncAt != "" {
			data.ProductRenewalIntegrations[i].LastSyncAt = syncAt
		}
		data.ProductRenewalIntegrations[i].LastError = lastError
		return
	}
}

func applyProductRenewalSyncSideEffects(data *AppData, record ProductRenewalSyncRecord) {
	switch record.ResourceType {
	case "esign":
		for i := range data.ProductRenewalESigns {
			if data.ProductRenewalESigns[i].ID != record.ResourceID {
				continue
			}
			if record.Status == "succeeded" {
				if record.ExternalStatus == "signed" {
					data.ProductRenewalESigns[i].Status = "signed"
					data.ProductRenewalESigns[i].SignedAt = fallback(data.ProductRenewalESigns[i].SignedAt, nowString())
					data.ProductRenewalESigns[i].Signature = fallback(data.ProductRenewalESigns[i].Signature, productRenewalCallbackString(record.ResponsePayload, "signature"))
					data.ProductRenewalESigns[i].Signature = fallback(data.ProductRenewalESigns[i].Signature, data.ProductRenewalESigns[i].Signer+" 电子签名")
				} else if data.ProductRenewalESigns[i].Status == "" || data.ProductRenewalESigns[i].Status == "failed" {
					data.ProductRenewalESigns[i].Status = "sent"
				}
				data.ProductRenewalESigns[i].LinkURL = fallback(data.ProductRenewalESigns[i].LinkURL, "/public/renewal-sign/"+data.ProductRenewalESigns[i].SignNo)
			} else if record.Status == "failed" {
				data.ProductRenewalESigns[i].Status = "failed"
			}
			return
		}
	case "invoice":
		for i := range data.ProductRenewalInvoices {
			if data.ProductRenewalInvoices[i].ID != record.ResourceID {
				continue
			}
			if record.Status == "succeeded" {
				data.ProductRenewalInvoices[i].Status = fallback(data.ProductRenewalInvoices[i].Status, "issued")
				data.ProductRenewalInvoices[i].TaxStatus = "accepted"
				data.ProductRenewalInvoices[i].ExternalRequest = fallback(record.ExternalRequestID, data.ProductRenewalInvoices[i].ExternalRequest)
				data.ProductRenewalInvoices[i].FileURL = fallback(productRenewalCallbackString(record.ResponsePayload, "fileUrl", "fileURL", "downloadUrl", "pdfUrl"), data.ProductRenewalInvoices[i].FileURL)
				data.ProductRenewalInvoices[i].FileURL = fallback(data.ProductRenewalInvoices[i].FileURL, "renewal-invoice://"+data.ProductRenewalInvoices[i].InvoiceNo+".pdf")
			} else if record.Status == "failed" {
				data.ProductRenewalInvoices[i].TaxStatus = "failed"
				data.ProductRenewalInvoices[i].ExternalRequest = fallback(record.ExternalRequestID, data.ProductRenewalInvoices[i].ExternalRequest)
			}
			return
		}
	}
}

func retryProductRenewalSyncRecord(data *AppData, id int64) (ProductRenewalSyncRecord, error) {
	for i := range data.ProductRenewalSyncRecords {
		if data.ProductRenewalSyncRecords[i].ID != id {
			continue
		}
		data.ProductRenewalSyncRecords[i].Status = "pending"
		data.ProductRenewalSyncRecords[i].NextRetryAt = ""
		data.ProductRenewalSyncRecords[i].Error = ""
		deliverProductRenewalSyncRecord(data, &data.ProductRenewalSyncRecords[i])
		applyProductRenewalSyncSideEffects(data, data.ProductRenewalSyncRecords[i])
		return data.ProductRenewalSyncRecords[i], nil
	}
	return ProductRenewalSyncRecord{}, fmt.Errorf("续费同步记录不存在")
}

func productRenewalSyncRecordIndex(data AppData, payload productRenewalSyncCallbackPayload) int {
	syncNo := strings.TrimSpace(payload.SyncNo)
	requestID := strings.TrimSpace(payload.ExternalRequestID)
	resourceType := strings.ToLower(strings.TrimSpace(payload.ResourceType))
	resourceNo := strings.TrimSpace(payload.ResourceNo)
	for i := range data.ProductRenewalSyncRecords {
		record := data.ProductRenewalSyncRecords[i]
		if syncNo != "" && record.SyncNo == syncNo {
			return i
		}
		if requestID != "" && record.ExternalRequestID == requestID {
			return i
		}
		if resourceType != "" && resourceNo != "" && record.ResourceType == resourceType && record.ResourceNo == resourceNo {
			return i
		}
	}
	return -1
}

func productRenewalCallbackString(raw string, keys ...string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := decoded[key]; ok {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func trimProductRenewalSyncRecords(data *AppData) {
	const maxItems = 500
	if len(data.ProductRenewalSyncRecords) <= maxItems {
		return
	}
	data.ProductRenewalSyncRecords = append([]ProductRenewalSyncRecord{}, data.ProductRenewalSyncRecords[len(data.ProductRenewalSyncRecords)-maxItems:]...)
}
