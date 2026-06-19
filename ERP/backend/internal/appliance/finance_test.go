package appliance

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestInvoiceTaxSubmissionAndCustomerDownload(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices", `{"statementId":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var invoice SalesInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode invoice: %v", err)
	}
	if invoice.TaxStatus != "pending" || invoice.FileURL != "" {
		t.Fatalf("expected pending tax invoice, got %+v", invoice)
	}

	customerToken := testLogin(t, app, "customer", "customer123")
	invoiceID := strconv.FormatInt(invoice.ID, 10)
	rec = testRequest(t, app, customerToken, http.MethodGet, "/api/finance/invoices/"+invoiceID+"/download", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected pending invoice download rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/"+invoiceID+"/submit-tax", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("submit tax status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode submitted invoice: %v", err)
	}
	if invoice.TaxStatus != "submitted" || invoice.TaxControlNo == "" || invoice.FileURL == "" {
		t.Fatalf("expected submitted tax invoice, got %+v", invoice)
	}

	rec = testRequest(t, app, customerToken, http.MethodGet, "/api/finance/invoices/"+invoiceID+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("download invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var download struct {
		Invoice      SalesInvoice `json:"invoice"`
		FileName     string       `json:"fileName"`
		URL          string       `json:"url"`
		DownloadedAt string       `json:"downloadedAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode download: %v", err)
	}
	if download.Invoice.DownloadedAt == "" || download.DownloadedAt == "" || download.URL == "" || download.FileName == "" {
		t.Fatalf("expected invoice download metadata, got %+v", download)
	}
}

func TestExternalTaxGatewaySubmissionRecordsAuditTrail(t *testing.T) {
	const secret = "tax-secret"
	var seen int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen++
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tax-token" {
			t.Errorf("unexpected authorization header %q", got)
		}
		raw, _ := io.ReadAll(r.Body)
		timestamp := r.Header.Get("X-CBMP-Timestamp")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(timestamp))
		mac.Write([]byte("."))
		mac.Write(raw)
		if got, want := r.Header.Get("X-CBMP-Signature"), "hmac-sha256="+hex.EncodeToString(mac.Sum(nil)); got != want {
			t.Errorf("unexpected signature %q, want %q", got, want)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Errorf("decode tax payload: %v", err)
		}
		if payload["invoiceNo"] == "" || payload["requestId"] == "" {
			t.Errorf("expected invoice payload with request id, got %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"status":"submitted","data":{"taxControlNo":"EXT-TAX-0001","fileUrl":"https://tax.example/invoices/EXT-TAX-0001.pdf"}}`))
	}))
	defer server.Close()

	t.Setenv("CBMP_TAX_GATEWAY_URL", server.URL)
	t.Setenv("CBMP_TAX_GATEWAY_PROVIDER", "external-tax-cn")
	t.Setenv("CBMP_TAX_GATEWAY_TOKEN", "tax-token")
	t.Setenv("CBMP_TAX_GATEWAY_SECRET", secret)
	t.Setenv("CBMP_TAX_GATEWAY_RETRIES", "0")

	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices", `{"statementId":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var invoice SalesInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode invoice: %v", err)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/"+strconv.FormatInt(invoice.ID, 10)+"/submit-tax", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("submit tax status %d: %s", rec.Code, rec.Body.String())
	}
	if seen != 1 {
		t.Fatalf("expected one external tax request, got %d", seen)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode submitted invoice: %v", err)
	}
	if invoice.TaxStatus != "submitted" || invoice.TaxControlNo != "EXT-TAX-0001" || invoice.FileURL == "" {
		t.Fatalf("expected external tax result, got %+v", invoice)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		TaxGatewaySubmissions []TaxGatewaySubmission `json:"taxGatewaySubmissions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode finance overview: %v", err)
	}
	var submission TaxGatewaySubmission
	for _, item := range overview.TaxGatewaySubmissions {
		if item.InvoiceID == invoice.ID {
			submission = item
		}
	}
	if submission.ID == 0 || submission.Status != "submitted" || submission.Provider != "external-tax-cn" || submission.RequestID == "" || submission.Attempt != 1 {
		t.Fatalf("expected submitted tax gateway trail, got %+v", submission)
	}
}

func TestExternalTaxGatewayFailureMarksInvoiceFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv("CBMP_TAX_GATEWAY_URL", server.URL)
	t.Setenv("CBMP_TAX_GATEWAY_PROVIDER", "faulty-tax-cn")
	t.Setenv("CBMP_TAX_GATEWAY_RETRIES", "0")

	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices", `{"statementId":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var invoice SalesInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode invoice: %v", err)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/"+strconv.FormatInt(invoice.ID, 10)+"/submit-tax", `{}`)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected bad gateway, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Invoices              []SalesInvoice         `json:"invoices"`
		TaxGatewaySubmissions []TaxGatewaySubmission `json:"taxGatewaySubmissions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode finance overview: %v", err)
	}
	var updated SalesInvoice
	for _, item := range overview.Invoices {
		if item.ID == invoice.ID {
			updated = item
		}
	}
	if updated.TaxStatus != "failed" {
		t.Fatalf("expected failed invoice tax status, got %+v", updated)
	}
	var submission TaxGatewaySubmission
	for _, item := range overview.TaxGatewaySubmissions {
		if item.InvoiceID == invoice.ID {
			submission = item
		}
	}
	if submission.Status != "failed" || submission.Error == "" || submission.Provider != "faulty-tax-cn" {
		t.Fatalf("expected failed tax gateway trail, got %+v", submission)
	}
}

func TestTaxGatewayAsyncCallbackCompletesAcceptedInvoice(t *testing.T) {
	const secret = "callback-secret"
	var requestID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode gateway request: %v", err)
		}
		requestID, _ = payload["requestId"].(string)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"status":"accepted","requestId":"` + requestID + `"}`))
	}))
	defer server.Close()

	t.Setenv("CBMP_TAX_GATEWAY_URL", server.URL)
	t.Setenv("CBMP_TAX_GATEWAY_PROVIDER", "async-tax-cn")
	t.Setenv("CBMP_TAX_GATEWAY_SECRET", secret)
	t.Setenv("CBMP_TAX_GATEWAY_RETRIES", "0")

	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices", `{"statementId":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var invoice SalesInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode invoice: %v", err)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/"+strconv.FormatInt(invoice.ID, 10)+"/submit-tax", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("submit tax status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode accepted invoice: %v", err)
	}
	if invoice.TaxStatus != "accepted" || requestID == "" {
		t.Fatalf("expected accepted invoice with request id, got invoice=%+v requestID=%q", invoice, requestID)
	}

	callback := []byte(`{"requestId":"` + requestID + `","status":"submitted","taxControlNo":"ASYNC-TAX-001","fileUrl":"https://tax.example/async.pdf"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/finance/tax/callback", bytes.NewReader(callback))
	req.Header.Set("Content-Type", "application/json")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("X-CBMP-Timestamp", timestamp)
	req.Header.Set("X-CBMP-Signature", signedTaxCallback(secret, timestamp, callback))
	resp := httptest.NewRecorder()
	app.Routes().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("callback status %d: %s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Invoice    SalesInvoice         `json:"invoice"`
		Submission TaxGatewaySubmission `json:"submission"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode callback response: %v", err)
	}
	if payload.Invoice.TaxStatus != "submitted" || payload.Invoice.TaxControlNo != "ASYNC-TAX-001" || payload.Submission.Status != "submitted" {
		t.Fatalf("expected submitted async invoice, got %+v", payload)
	}
}

func TestTaxGatewayCallbackRejectsBadSignature(t *testing.T) {
	t.Setenv("CBMP_TAX_GATEWAY_SECRET", "callback-secret")
	app := newTestHTTPApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/finance/tax/callback", bytes.NewBufferString(`{"requestId":"bad"}`))
	req.Header.Set("X-CBMP-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("X-CBMP-Signature", "hmac-sha256=bad")
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected callback signature rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRedOffsetInvoiceCreatesNegativeRedInvoice(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/1/red-offset", `{"reason":"客户退货红冲"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("red offset status %d: %s", rec.Code, rec.Body.String())
	}
	var red SalesInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &red); err != nil {
		t.Fatalf("decode red invoice: %v", err)
	}
	if red.InvoiceType != "red" || red.OriginalInvoiceID != 1 || red.Amount >= 0 || red.TaxAmount >= 0 || red.TaxStatus != "submitted" || red.TaxControlNo == "" {
		t.Fatalf("expected submitted red invoice, got %+v", red)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Invoices              []SalesInvoice         `json:"invoices"`
		RedLetterInfos        []RedLetterInfo        `json:"redLetterInfos"`
		Receivables           []Receivable           `json:"receivables"`
		TaxGatewaySubmissions []TaxGatewaySubmission `json:"taxGatewaySubmissions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode finance overview: %v", err)
	}
	var original SalesInvoice
	for _, item := range overview.Invoices {
		if item.ID == 1 {
			original = item
		}
	}
	if original.TaxStatus != "red_offset" || original.Status != "red_offset" {
		t.Fatalf("expected original invoice red offset status, got %+v", original)
	}
	var adjustment Receivable
	for _, item := range overview.Receivables {
		if item.InvoiceID == red.ID {
			adjustment = item
		}
	}
	if adjustment.Amount >= 0 || adjustment.Status != "credited" {
		t.Fatalf("expected negative credited receivable adjustment, got %+v", adjustment)
	}
	var submission TaxGatewaySubmission
	for _, item := range overview.TaxGatewaySubmissions {
		if item.InvoiceID == red.ID {
			submission = item
		}
	}
	if submission.Action != "red_offset" || submission.Status != "submitted" {
		t.Fatalf("expected red offset tax submission, got %+v", submission)
	}
	var redInfo RedLetterInfo
	for _, item := range overview.RedLetterInfos {
		if item.RedInvoiceID == red.ID {
			redInfo = item
		}
	}
	if redInfo.ID == 0 || redInfo.Status != "used" || redInfo.InfoNo != red.RedLetterInfoNo || red.RedLetterInfoID == 0 {
		t.Fatalf("expected auto-used red letter info, got info=%+v red=%+v", redInfo, red)
	}
}

func TestRedLetterInfoApprovalAndUsage(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/red-letters", `{"originalInvoiceId":1,"reason":"项目退货红字申请"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create red letter info status %d: %s", rec.Code, rec.Body.String())
	}
	var info RedLetterInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode red letter info: %v", err)
	}
	if info.Status != "requested" || info.OriginalInvoiceID != 1 || info.Amount >= 0 || info.TaxAmount >= 0 {
		t.Fatalf("expected requested negative red letter info, got %+v", info)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/1/red-offset", `{"reason":"项目退货红字申请","redLetterInfoId":`+strconv.FormatInt(info.ID, 10)+`}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected unapproved red letter rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/red-letters/"+strconv.FormatInt(info.ID, 10)+"/approve", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve red letter info status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode approved red letter info: %v", err)
	}
	if info.Status != "approved" || info.TaxControlNo == "" || info.ApprovedAt == "" {
		t.Fatalf("expected approved red letter info, got %+v", info)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/invoices/1/red-offset", `{"reason":"项目退货红字申请","redLetterInfoId":`+strconv.FormatInt(info.ID, 10)+`}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("red offset with red letter info status %d: %s", rec.Code, rec.Body.String())
	}
	var red SalesInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &red); err != nil {
		t.Fatalf("decode red invoice: %v", err)
	}
	if red.RedLetterInfoID != info.ID || red.RedLetterInfoNo != info.InfoNo || red.InvoiceCategory != "red_vat_special" {
		t.Fatalf("expected red invoice linked to red letter info, got red=%+v info=%+v", red, info)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		RedLetterInfos []RedLetterInfo `json:"redLetterInfos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode finance overview: %v", err)
	}
	var used RedLetterInfo
	for _, item := range overview.RedLetterInfos {
		if item.ID == info.ID {
			used = item
		}
	}
	if used.Status != "used" || used.RedInvoiceID != red.ID || used.UsedAt == "" {
		t.Fatalf("expected used red letter info after red invoice submission, got %+v", used)
	}
}

func TestPaymentPlanSettlementUpdatesReceivableAndAging(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/payment-plans", `{"receivableId":1,"amount":3000,"dueDate":"2026-07-20","method":"bank","remark":"二期回款计划"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create payment plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan PaymentPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode payment plan: %v", err)
	}
	if plan.Status != "planned" || plan.PlanNo == "" || plan.Amount != 3000 {
		t.Fatalf("expected planned payment plan, got %+v", plan)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/payment-plans/"+strconv.FormatInt(plan.ID, 10)+"/settle", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("settle payment plan status %d: %s", rec.Code, rec.Body.String())
	}
	var settlement struct {
		PaymentPlan PaymentPlan `json:"paymentPlan"`
		Receipt     Receipt     `json:"receipt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &settlement); err != nil {
		t.Fatalf("decode payment settlement: %v", err)
	}
	if settlement.PaymentPlan.Status != "settled" || settlement.PaymentPlan.SettledAt == "" || settlement.Receipt.Amount != 3000 {
		t.Fatalf("expected settled payment plan and receipt, got %+v", settlement)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Receivables  []Receivable            `json:"receivables"`
		PaymentPlans []PaymentPlan           `json:"paymentPlans"`
		AgingBuckets []ReceivableAgingBucket `json:"agingBuckets"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode finance overview: %v", err)
	}
	var receivable Receivable
	for _, item := range overview.Receivables {
		if item.ID == 1 {
			receivable = item
		}
	}
	if receivable.ReceivedAmount != 11000 || receivable.Status != "partial" {
		t.Fatalf("expected receivable partially settled, got %+v", receivable)
	}
	var settled PaymentPlan
	for _, item := range overview.PaymentPlans {
		if item.ID == plan.ID {
			settled = item
		}
	}
	if settled.Status != "settled" {
		t.Fatalf("expected settled plan in overview, got %+v", settled)
	}
	if len(overview.AgingBuckets) == 0 || overview.AgingBuckets[0].Bucket != "current" {
		t.Fatalf("expected aging buckets in overview, got %+v", overview.AgingBuckets)
	}
}

func TestCollectionTasksGenerateSendAndHandle(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/finance/collections/generate", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("generate collection tasks status %d: %s", rec.Code, rec.Body.String())
	}
	var tasks []CollectionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("decode collection tasks: %v", err)
	}
	if len(tasks) == 0 || tasks[0].Status != "open" || tasks[0].Amount <= 0 || tasks[0].Channel == "" {
		t.Fatalf("expected open collection task, got %+v", tasks)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/collections/generate", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("generate duplicate collection tasks status %d: %s", rec.Code, rec.Body.String())
	}
	var duplicate []CollectionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &duplicate); err != nil {
		t.Fatalf("decode duplicate collection tasks: %v", err)
	}
	if len(duplicate) != 0 {
		t.Fatalf("expected duplicate generation to skip open tasks, got %+v", duplicate)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/collection-templates", `{"code":"custom_sms","name":"自定义催收短信","level":"pre_due","channel":"sms","content":"{{customerName}} 应收 {{amount}} 将于 {{dueDate}} 到期"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create collection template status %d: %s", rec.Code, rec.Body.String())
	}
	var template CollectionTemplate
	if err := json.Unmarshal(rec.Body.Bytes(), &template); err != nil {
		t.Fatalf("decode collection template: %v", err)
	}
	if template.ID == 0 || !template.Enabled {
		t.Fatalf("expected enabled collection template, got %+v", template)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/collections/"+strconv.FormatInt(tasks[0].ID, 10)+"/send", `{"templateId":`+strconv.FormatInt(template.ID, 10)+`}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("send collection task status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatch CollectionDispatch
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatch); err != nil {
		t.Fatalf("decode collection dispatch: %v", err)
	}
	if dispatch.Status != "delivered" || dispatch.TemplateID != template.ID || dispatch.Target == "" || dispatch.Endpoint == "" {
		t.Fatalf("expected delivered collection dispatch, got %+v", dispatch)
	}
	if dispatch.ProviderRequestID == "" {
		t.Fatalf("expected provider request id, got %+v", dispatch)
	}

	t.Setenv("CBMP_COLLECTION_CALLBACK_SECRET", "collection-secret")
	callback := []byte(`{"requestId":"` + dispatch.ProviderRequestID + `","status":"read","providerMessageId":"MSG-COL-001","provider":"sms-provider"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/finance/collections/callback", bytes.NewReader(callback))
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("X-CBMP-Timestamp", timestamp)
	req.Header.Set("X-CBMP-Signature", signedTaxCallback("collection-secret", timestamp, callback))
	resp := httptest.NewRecorder()
	app.Routes().ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("collection callback status %d: %s", resp.Code, resp.Body.String())
	}
	var callbackDispatch CollectionDispatch
	if err := json.Unmarshal(resp.Body.Bytes(), &callbackDispatch); err != nil {
		t.Fatalf("decode collection callback: %v", err)
	}
	if callbackDispatch.Status != "read" || callbackDispatch.ProviderMessageID != "MSG-COL-001" || callbackDispatch.CallbackAt == "" {
		t.Fatalf("expected collection callback metadata, got %+v", callbackDispatch)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/api/finance/collections/callback", bytes.NewBufferString(`{"requestId":"bad"}`))
	badReq.Header.Set("X-CBMP-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	badReq.Header.Set("X-CBMP-Signature", "hmac-sha256=bad")
	badResp := httptest.NewRecorder()
	app.Routes().ServeHTTP(badResp, badReq)
	if badResp.Code != http.StatusBadRequest {
		t.Fatalf("expected collection callback signature rejection, got %d: %s", badResp.Code, badResp.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		CollectionTasks      []CollectionTask     `json:"collectionTasks"`
		CollectionTemplates  []CollectionTemplate `json:"collectionTemplates"`
		CollectionDispatches []CollectionDispatch `json:"collectionDispatches"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode collection overview: %v", err)
	}
	if len(overview.CollectionTemplates) == 0 || len(overview.CollectionDispatches) == 0 {
		t.Fatalf("expected collection templates and dispatches in overview, got %+v", overview)
	}
	if overview.CollectionDispatches[0].Status != "read" || overview.CollectionDispatches[0].CallbackAt == "" {
		t.Fatalf("expected callback status in collection overview, got %+v", overview.CollectionDispatches)
	}
	var sentTask CollectionTask
	for _, item := range overview.CollectionTasks {
		if item.ID == tasks[0].ID {
			sentTask = item
		}
	}
	if sentTask.SendCount != 1 || sentTask.LastSentAt == "" || sentTask.TemplateID != template.ID {
		t.Fatalf("expected sent task metadata, got %+v", sentTask)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/finance/collections/"+strconv.FormatInt(tasks[0].ID, 10)+"/handle", `{"remark":"已电话确认回款"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("handle collection task status %d: %s", rec.Code, rec.Body.String())
	}
	var handled CollectionTask
	if err := json.Unmarshal(rec.Body.Bytes(), &handled); err != nil {
		t.Fatalf("decode handled collection task: %v", err)
	}
	if handled.Status != "handled" || handled.HandledAt == "" || handled.Remark == "" {
		t.Fatalf("expected handled collection task, got %+v", handled)
	}
}

func signedTaxCallback(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return "hmac-sha256=" + hex.EncodeToString(mac.Sum(nil))
}
