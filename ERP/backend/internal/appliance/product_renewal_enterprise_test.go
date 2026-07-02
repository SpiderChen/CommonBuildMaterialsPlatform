//go:build legacy_product_ops

package appliance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestProductRenewalInvoiceRequiresRealTaxRequestForAcceptedStatus(t *testing.T) {
	app := newTestHTTPApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/1/invoice", strings.NewReader(`{"contractId":1,"paymentId":1,"taxStatus":"accepted"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.productOpsRenewalInvoice(rec, req, renewalTestSession(), "1")

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "外部请求号") {
		t.Fatalf("expected missing external request rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProductRenewalInvoiceDoesNotFakeTaxAcceptanceWithoutIntegration(t *testing.T) {
	app := newTestHTTPApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/1/invoice", strings.NewReader(`{"contractId":1,"paymentId":1,"invoiceType":"blue_e_invoice","taxRate":0.06}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.productOpsRenewalInvoice(rec, req, renewalTestSession(), "1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var invoice ProductRenewalInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode renewal invoice: %v", err)
	}
	if invoice.TaxStatus != "failed" || invoice.ExternalRequest != "" || invoice.FileURL != "" {
		t.Fatalf("renewal invoice must not fake tax success without integration, got %+v", invoice)
	}

	data, err := app.store.Snapshot()
	if err != nil {
		t.Fatalf("snapshot store: %v", err)
	}
	var syncRecord ProductRenewalSyncRecord
	for _, item := range data.ProductRenewalSyncRecords {
		if item.ResourceType == "invoice" && item.ResourceID == invoice.ID {
			syncRecord = item
			break
		}
	}
	if syncRecord.ID == 0 || syncRecord.Status != "failed" || syncRecord.Error == "" || syncRecord.ExternalRequestID != "" {
		t.Fatalf("expected failed renewal tax sync without fake external request, got %+v", syncRecord)
	}
}

func TestProductRenewalTaxSyncRequiresExternalRequestID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	data := SeedData()
	for i := range data.ProductRenewalIntegrations {
		if data.ProductRenewalIntegrations[i].Code == "tax_gateway" {
			data.ProductRenewalIntegrations[i].Status = "active"
			data.ProductRenewalIntegrations[i].Endpoint = server.URL
			break
		}
	}

	record, err := enqueueProductRenewalSyncRecord(&data, productRenewalSyncRequest{
		Scenario:     "tax",
		ResourceType: "invoice",
		ResourceID:   1,
		ResourceNo:   "RI202606190001",
		Task:         data.ProductRenewalTasks[0],
		Action:       "issue",
		Payload:      map[string]string{"invoiceNo": "RI202606190001"},
	})
	if err != nil {
		t.Fatalf("enqueue renewal tax sync: %v", err)
	}
	if record.Status != "failed" || !strings.Contains(record.Error, "外部请求号") || record.ExternalRequestID != "" {
		t.Fatalf("tax sync without external request id must fail, got %+v", record)
	}
	if data.ProductRenewalInvoices[0].TaxStatus != "failed" || data.ProductRenewalInvoices[0].ExternalRequest != "" {
		t.Fatalf("invoice must not be accepted without external request id, got %+v", data.ProductRenewalInvoices[0])
	}
}

func TestProductRenewalESignSendDoesNotFabricateLinkWithoutIntegration(t *testing.T) {
	app := newTestHTTPApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/1/esign", strings.NewReader(`{"contractId":1,"signer":"张三","phone":"13800000000"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.productOpsRenewalESign(rec, req, renewalTestSession(), "1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("send renewal e-sign status %d: %s", rec.Code, rec.Body.String())
	}
	var sign ProductRenewalESign
	if err := json.Unmarshal(rec.Body.Bytes(), &sign); err != nil {
		t.Fatalf("decode renewal e-sign: %v", err)
	}
	if sign.Status != "failed" || sign.LinkURL != "" {
		t.Fatalf("renewal e-sign must not fabricate link without integration, got %+v", sign)
	}
}

func TestProductRenewalESignRequiresRealSignature(t *testing.T) {
	signServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"externalStatus":"sent","linkUrl":"https://sign.example.test/renewal/RS-REAL"}`))
	}))
	defer signServer.Close()

	app := newTestHTTPApp(t)
	if err := app.store.Mutate(func(data *AppData) error {
		integrationID := nextID(data, "renewalIntegration")
		data.ProductRenewalIntegrations = append(data.ProductRenewalIntegrations, ProductRenewalIntegration{
			ID:             integrationID,
			IntegrationNo:  number("RI", integrationID),
			Name:           "测试电子签网关",
			Code:           "test_esign",
			Provider:       "test_esign",
			Scenario:       "esign",
			Status:         "active",
			Endpoint:       signServer.URL,
			RetryLimit:     1,
			TimeoutSeconds: 3,
			CreatedBy:      "test",
			CreatedAt:      nowString(),
		})
		return nil
	}); err != nil {
		t.Fatalf("activate renewal e-sign integration: %v", err)
	}

	sendReq := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/1/esign", strings.NewReader(`{"contractId":1,"channel":"test_esign","signer":"张三","phone":"13800000000"}`))
	sendReq.Header.Set("Content-Type", "application/json")
	sendRec := httptest.NewRecorder()

	app.productOpsRenewalESign(sendRec, sendReq, renewalTestSession(), "1")

	if sendRec.Code != http.StatusCreated {
		t.Fatalf("send renewal e-sign status %d: %s", sendRec.Code, sendRec.Body.String())
	}
	var sign ProductRenewalESign
	if err := json.Unmarshal(sendRec.Body.Bytes(), &sign); err != nil {
		t.Fatalf("decode renewal e-sign: %v", err)
	}
	if sign.Status != "sent" || sign.LinkURL == "" {
		t.Fatalf("expected provider-backed sent e-sign, got %+v", sign)
	}

	missingReq := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/1/esign", strings.NewReader(`{"action":"complete","contractId":1,"signId":`+strconv.FormatInt(sign.ID, 10)+`}`))
	missingReq.Header.Set("Content-Type", "application/json")
	missingRec := httptest.NewRecorder()

	app.productOpsRenewalESign(missingRec, missingReq, renewalTestSession(), "1")

	if missingRec.Code != http.StatusBadRequest || !strings.Contains(missingRec.Body.String(), "真实签名") {
		t.Fatalf("expected missing real signature rejection, got %d: %s", missingRec.Code, missingRec.Body.String())
	}

	completeReq := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/1/esign", strings.NewReader(`{"action":"complete","contractId":1,"signId":`+strconv.FormatInt(sign.ID, 10)+`,"signature":"sig-real"}`))
	completeReq.Header.Set("Content-Type", "application/json")
	completeRec := httptest.NewRecorder()

	app.productOpsRenewalESign(completeRec, completeReq, renewalTestSession(), "1")

	if completeRec.Code != http.StatusCreated {
		t.Fatalf("complete renewal e-sign status %d: %s", completeRec.Code, completeRec.Body.String())
	}
	var completed ProductRenewalESign
	if err := json.Unmarshal(completeRec.Body.Bytes(), &completed); err != nil {
		t.Fatalf("decode completed renewal e-sign: %v", err)
	}
	if completed.Status != "signed" || completed.Signature != "sig-real" {
		t.Fatalf("expected real signature to be stored, got %+v", completed)
	}
}

func TestProductRenewalSideEffectsDoNotFabricateExternalArtifacts(t *testing.T) {
	data := AppData{
		ProductRenewalInvoices: []ProductRenewalInvoice{{
			ID: 1, InvoiceNo: "RI-1", TaxStatus: "pending",
		}},
		ProductRenewalESigns: []ProductRenewalESign{{
			ID: 2, SignNo: "RS-2", Signer: "张三", Status: "sent",
		}},
	}

	applyProductRenewalSyncSideEffects(&data, ProductRenewalSyncRecord{
		ResourceType:      "invoice",
		ResourceID:        1,
		Status:            "succeeded",
		ExternalRequestID: "tax-request-1",
		ResponsePayload:   `{"externalRequestId":"tax-request-1","externalStatus":"accepted"}`,
	})
	if data.ProductRenewalInvoices[0].FileURL != "" {
		t.Fatalf("renewal invoice sync must not fabricate file url, got %+v", data.ProductRenewalInvoices[0])
	}

	applyProductRenewalSyncSideEffects(&data, ProductRenewalSyncRecord{
		ResourceType:    "esign",
		ResourceID:      2,
		Status:          "succeeded",
		ExternalStatus:  "signed",
		ResponsePayload: `{"externalStatus":"signed"}`,
	})
	if data.ProductRenewalESigns[0].Signature != "" || data.ProductRenewalESigns[0].LinkURL != "" {
		t.Fatalf("esign sync must not fabricate signature or link, got %+v", data.ProductRenewalESigns[0])
	}
}

func renewalTestSession() Session {
	return Session{User: User{Username: "admin", DisplayName: "平台管理员", RoleCode: "boss"}}
}
