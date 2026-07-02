package appliance

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type TaxGatewayConfig struct {
	Provider  string
	URL       string
	Token     string
	Secret    string
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

type taxGatewayResult struct {
	Provider     string
	Endpoint     string
	RequestID    string
	Status       string
	TaxControlNo string
	FileURL      string
	Error        string
	Attempt      int
	DurationMs   int64
}

type taxGatewayCallbackPayload struct {
	RequestID    string
	InvoiceNo    string
	Action       string
	Status       string
	TaxControlNo string
	FileURL      string
	Error        string
	Provider     string
}

func NewTaxGatewayConfigFromEnv() TaxGatewayConfig {
	timeoutMs := intFromEnv("CBMP_TAX_GATEWAY_TIMEOUT_MS", 2500)
	retries := intFromEnv("CBMP_TAX_GATEWAY_RETRIES", 2)
	if retries < 0 {
		retries = 0
	}
	endpoint := strings.TrimSpace(os.Getenv("CBMP_TAX_GATEWAY_URL"))
	provider := strings.TrimSpace(os.Getenv("CBMP_TAX_GATEWAY_PROVIDER"))
	if provider == "" {
		provider = defaultTaxGatewayProvider(endpoint)
	}
	return TaxGatewayConfig{
		Provider:  provider,
		URL:       endpoint,
		Token:     strings.TrimSpace(os.Getenv("CBMP_TAX_GATEWAY_TOKEN")),
		Secret:    strings.TrimSpace(os.Getenv("CBMP_TAX_GATEWAY_SECRET")),
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
		Retries:   retries,
		UserAgent: "cbmp-appliance/1.0",
	}
}

func (r *RuntimeServices) SubmitTaxInvoice(ctx context.Context, invoice SalesInvoice) taxGatewayResult {
	return r.submitTaxGateway(ctx, invoice, nil, "issue")
}

func (r *RuntimeServices) SubmitTaxRedInvoice(ctx context.Context, invoice SalesInvoice, original SalesInvoice) taxGatewayResult {
	return r.submitTaxGateway(ctx, invoice, &original, "red_offset")
}

func (r *RuntimeServices) submitTaxGateway(ctx context.Context, invoice SalesInvoice, original *SalesInvoice, action string) taxGatewayResult {
	cfg := TaxGatewayConfig{Timeout: 2500 * time.Millisecond, UserAgent: "cbmp-appliance/1.0"}
	if r != nil {
		cfg = r.taxGateway
		if cfg.Timeout <= 0 {
			cfg.Timeout = 2500 * time.Millisecond
		}
		if cfg.UserAgent == "" {
			cfg.UserAgent = "cbmp-appliance/1.0"
		}
	}
	if cfg.Provider == "" {
		cfg.Provider = defaultTaxGatewayProvider(cfg.URL)
	}
	if cfg.URL == "" {
		return taxGatewayResult{
			Provider:  cfg.Provider,
			Endpoint:  strings.TrimSpace(cfg.URL),
			RequestID: taxGatewayRequestID(invoice, action),
			Status:    "failed",
			Error:     "税控网关未配置真实 endpoint",
			Attempt:   1,
		}
	}
	if err := validateNoMockEndpoint(cfg.URL, "税控网关 endpoint"); err != nil {
		return taxGatewayResult{
			Provider:  cfg.Provider,
			Endpoint:  strings.TrimSpace(cfg.URL),
			RequestID: taxGatewayRequestID(invoice, action),
			Status:    "failed",
			Error:     err.Error(),
			Attempt:   1,
		}
	}
	return submitTaxInvoiceHTTP(ctx, cfg, invoice, original, action)
}

func submitTaxInvoiceHTTP(ctx context.Context, cfg TaxGatewayConfig, invoice SalesInvoice, original *SalesInvoice, action string) taxGatewayResult {
	start := time.Now()
	requestID := taxGatewayRequestID(invoice, action)
	endpoint := sanitizedTaxEndpoint(cfg.URL)
	payload := map[string]interface{}{
		"requestId":       requestID,
		"action":          action,
		"provider":        cfg.Provider,
		"invoiceNo":       invoice.InvoiceNo,
		"invoiceType":     fallback(invoice.InvoiceType, "blue"),
		"invoiceCategory": fallback(invoice.InvoiceCategory, "blue_vat_special"),
		"statementId":     invoice.StatementID,
		"customerId":      invoice.CustomerID,
		"amount":          invoice.Amount,
		"taxRate":         invoice.TaxRate,
		"taxAmount":       invoice.TaxAmount,
		"issuedAt":        invoice.IssuedAt,
	}
	if original != nil {
		payload["originalInvoiceId"] = original.ID
		payload["originalInvoiceNo"] = original.InvoiceNo
		payload["originalTaxControlNo"] = original.TaxControlNo
		payload["redLetterInfoId"] = invoice.RedLetterInfoID
		payload["redLetterInfoNo"] = invoice.RedLetterInfoNo
		payload["redReason"] = invoice.RedReason
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}
	}

	attempts := cfg.Retries + 1
	if attempts < 1 {
		attempts = 1
	}
	client := &http.Client{Timeout: cfg.Timeout}
	var lastErr string
	for attempt := 1; attempt <= attempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(body))
		if err != nil {
			return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: err.Error(), Attempt: attempt, DurationMs: time.Since(start).Milliseconds()}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", cfg.UserAgent)
		req.Header.Set("X-CBMP-Request-Id", requestID)
		if cfg.Token != "" {
			req.Header.Set("Authorization", "Bearer "+cfg.Token)
		}
		if cfg.Secret != "" {
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			req.Header.Set("X-CBMP-Timestamp", timestamp)
			req.Header.Set("X-CBMP-Signature", taxGatewaySignature(cfg.Secret, timestamp, body))
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err.Error()
			if attempt < attempts {
				time.Sleep(time.Duration(attempt) * 120 * time.Millisecond)
				continue
			}
			return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: lastErr, Attempt: attempt, DurationMs: time.Since(start).Milliseconds()}
		}
		result := parseTaxGatewayHTTPResponse(resp, cfg, endpoint, requestID, attempt, start)
		if result.Status == "failed" && resp.StatusCode >= 500 && attempt < attempts {
			lastErr = result.Error
			time.Sleep(time.Duration(attempt) * 120 * time.Millisecond)
			continue
		}
		return result
	}
	return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: lastErr, Attempt: attempts, DurationMs: time.Since(start).Milliseconds()}
}

func parseTaxGatewayHTTPResponse(resp *http.Response, cfg TaxGatewayConfig, endpoint, requestID string, attempt int, start time.Time) taxGatewayResult {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return taxGatewayResult{
			Provider:   cfg.Provider,
			Endpoint:   endpoint,
			RequestID:  requestID,
			Status:     "failed",
			Error:      fmt.Sprintf("tax gateway status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))),
			Attempt:    attempt,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: "invalid tax gateway response: " + err.Error(), Attempt: attempt, DurationMs: time.Since(start).Milliseconds()}
	}
	status := strings.ToLower(firstTaxGatewayString(decoded, "status", "state"))
	code := strings.ToLower(firstTaxGatewayString(decoded, "code", "resultCode"))
	success := taxGatewaySuccess(decoded, status, code)
	if !success {
		errMessage := firstTaxGatewayString(decoded, "error", "message", "msg")
		if errMessage == "" {
			errMessage = "tax gateway rejected invoice"
		}
		return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: errMessage, Attempt: attempt, DurationMs: time.Since(start).Milliseconds()}
	}
	if status == "pending" || status == "processing" {
		status = "accepted"
	}
	taxControlNo := firstTaxGatewayString(decoded, "taxControlNo", "tax_control_no", "taxSerialNo", "serialNo")
	fileURL := firstTaxGatewayString(decoded, "fileUrl", "fileURL", "downloadUrl", "pdfUrl")
	if taxControlNo == "" {
		if status == "accepted" {
			responseRequestID := firstTaxGatewayString(decoded, "requestId", "request_id")
			if responseRequestID == "" {
				responseRequestID = requestID
			}
			return taxGatewayResult{
				Provider:   cfg.Provider,
				Endpoint:   endpoint,
				RequestID:  responseRequestID,
				Status:     "accepted",
				Attempt:    attempt,
				DurationMs: time.Since(start).Milliseconds(),
			}
		}
		return taxGatewayResult{Provider: cfg.Provider, Endpoint: endpoint, RequestID: requestID, Status: "failed", Error: "tax gateway response missing taxControlNo", Attempt: attempt, DurationMs: time.Since(start).Milliseconds()}
	}
	if status == "" || status == "success" || status == "ok" {
		status = "submitted"
	}
	responseRequestID := firstTaxGatewayString(decoded, "requestId", "request_id")
	if responseRequestID == "" {
		responseRequestID = requestID
	}
	return taxGatewayResult{
		Provider:     cfg.Provider,
		Endpoint:     endpoint,
		RequestID:    responseRequestID,
		Status:       status,
		TaxControlNo: taxControlNo,
		FileURL:      fileURL,
		Attempt:      attempt,
		DurationMs:   time.Since(start).Milliseconds(),
	}
}

func taxGatewaySuccess(decoded map[string]interface{}, status, code string) bool {
	if success, ok := decoded["success"].(bool); ok {
		return success
	}
	switch status {
	case "", "submitted", "success", "accepted", "processing", "pending", "ok":
		return true
	case "failed", "rejected", "error":
		return false
	}
	switch code {
	case "", "0", "200", "success", "ok":
		return true
	default:
		return false
	}
}

func firstTaxGatewayString(decoded map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := stringFromMap(decoded, key); value != "" {
			return value
		}
	}
	if nested, ok := decoded["data"].(map[string]interface{}); ok {
		for _, key := range keys {
			if value := stringFromMap(nested, key); value != "" {
				return value
			}
		}
	}
	if nested, ok := decoded["result"].(map[string]interface{}); ok {
		for _, key := range keys {
			if value := stringFromMap(nested, key); value != "" {
				return value
			}
		}
	}
	return ""
}

func stringFromMap(values map[string]interface{}, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func taxGatewayRequestID(invoice SalesInvoice, action string) string {
	prefix := "tax"
	if action == "red_offset" {
		prefix = "tax-red"
	}
	return fmt.Sprintf("%s-%s-%d", prefix, invoice.InvoiceNo, time.Now().UnixNano())
}

func taxGatewaySignature(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return "hmac-sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func parseTaxGatewayCallback(r *http.Request, cfg TaxGatewayConfig) (taxGatewayCallbackPayload, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return taxGatewayCallbackPayload{}, err
	}
	if err := verifyTaxGatewayCallbackSignature(cfg.Secret, r.Header.Get("X-CBMP-Timestamp"), r.Header.Get("X-CBMP-Signature"), raw); err != nil {
		return taxGatewayCallbackPayload{}, err
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return taxGatewayCallbackPayload{}, fmt.Errorf("invalid tax callback payload: %w", err)
	}
	status := strings.ToLower(firstTaxGatewayString(decoded, "status", "state"))
	if status == "success" || status == "ok" {
		status = "submitted"
	}
	if status == "pending" || status == "processing" {
		status = "accepted"
	}
	return taxGatewayCallbackPayload{
		RequestID:    firstTaxGatewayString(decoded, "requestId", "request_id"),
		InvoiceNo:    firstTaxGatewayString(decoded, "invoiceNo", "invoice_no"),
		Action:       firstTaxGatewayString(decoded, "action"),
		Status:       status,
		TaxControlNo: firstTaxGatewayString(decoded, "taxControlNo", "tax_control_no", "taxSerialNo", "serialNo"),
		FileURL:      firstTaxGatewayString(decoded, "fileUrl", "fileURL", "downloadUrl", "pdfUrl"),
		Error:        firstTaxGatewayString(decoded, "error", "message", "msg"),
		Provider:     firstTaxGatewayString(decoded, "provider"),
	}, nil
}

func verifyTaxGatewayCallbackSignature(secret, timestamp, signature string, body []byte) error {
	if secret == "" {
		return fmt.Errorf("未配置税控回调密钥")
	}
	if timestamp == "" || signature == "" {
		return fmt.Errorf("税控回调缺少签名头")
	}
	parsed, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("税控回调时间戳无效")
	}
	if delta := time.Since(time.Unix(parsed, 0)); delta > 10*time.Minute || delta < -10*time.Minute {
		return fmt.Errorf("税控回调时间戳超出窗口")
	}
	expected := taxGatewaySignature(secret, timestamp, body)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("税控回调签名无效")
	}
	return nil
}

func sanitizedTaxEndpoint(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "<configured>"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func updateTaxIntegrationEndpoint(data *AppData, result taxGatewayResult, status string) {
	endpoint := result.Endpoint
	for i := range data.IntegrationEndpoints {
		item := data.IntegrationEndpoints[i]
		if item.Type == "finance" && (strings.Contains(item.Name, "税控") || strings.Contains(item.URL, "tax")) {
			data.IntegrationEndpoints[i].Protocol = "rest/http"
			data.IntegrationEndpoints[i].URL = endpoint
			data.IntegrationEndpoints[i].Status = status
			data.IntegrationEndpoints[i].LastSyncAt = nowString()
			return
		}
	}
	data.IntegrationEndpoints = append(data.IntegrationEndpoints, IntegrationEndpoint{
		ID:         nextID(data, "integration"),
		Name:       "财务税控接口",
		Type:       "finance",
		Protocol:   "rest/http",
		URL:        endpoint,
		Status:     status,
		LastSyncAt: nowString(),
	})
}

func defaultTaxGatewayProvider(endpoint string) string {
	normalized := strings.ToLower(strings.TrimSpace(endpoint))
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "mock://") || strings.HasPrefix(normalized, "tax://") {
		return ""
	}
	return "external-tax"
}

func intFromEnv(name string, fallbackValue int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallbackValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallbackValue
	}
	return parsed
}
