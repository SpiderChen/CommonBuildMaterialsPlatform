package appliance

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newTestHTTPApp(t *testing.T) *App {
	t.Helper()
	store := NewStore(filepath.Join(t.TempDir(), "app.vault"), "test-key")
	if err := store.Load(); err != nil {
		t.Fatalf("load store: %v", err)
	}
	return NewApp(store, "")
}

func TestAppFrontendStaticServingIsExplicit(t *testing.T) {
	app := newTestHTTPApp(t)

	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("root should be disabled without frontend dir, got %d: %s", rec.Code, rec.Body.String())
	}

	frontendDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(frontendDir, "index.html"), []byte("client shell"), 0o600); err != nil {
		t.Fatalf("write test index: %v", err)
	}
	store := NewStore(filepath.Join(t.TempDir(), "app.vault"), "test-key")
	if err := store.Load(); err != nil {
		t.Fatalf("load store: %v", err)
	}
	withFrontend := NewApp(store, frontendDir)

	rec = httptest.NewRecorder()
	withFrontend.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "client shell") {
		t.Fatalf("root should serve configured frontend, status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIRootAndCORSAreERPOriented(t *testing.T) {
	app := newTestHTTPApp(t)

	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("api root status %d: %s", rec.Code, rec.Body.String())
	}
	var info map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode api root: %v", err)
	}
	if info["product"] != "common-build-materials-erp" || !strings.Contains(info["name"], "ERP") {
		t.Fatalf("api root should expose ERP identity: %+v", info)
	}

	req := httptest.NewRequest(http.MethodOptions, "/api/iot/vehicle/location/report", nil)
	req.Header.Set("Origin", "https://customer-app.example")
	req.Header.Set("Access-Control-Request-Headers", "X-Device-Key, X-CBMP-Signature")
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("cors preflight status %d: %s", rec.Code, rec.Body.String())
	}
	allowed := rec.Header().Get("Access-Control-Allow-Headers")
	for _, header := range []string{"X-Device-Key", "X-CBMP-Signature", "X-CBMP-Request-Id"} {
		if !strings.Contains(allowed, header) {
			t.Fatalf("cors allow headers missing %s: %s", header, allowed)
		}
	}
}

func TestERPRejectsProductOpsRoutes(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/product-ops/overview", ""},
		{http.MethodPost, "/api/product-ops/alerts", `{}`},
		{http.MethodPost, "/api/product-ops/probes/report", `{}`},
		{http.MethodPost, "/api/product-ops/renewals/sync-callback", `{}`},
	}
	for _, item := range cases {
		rec := testRequest(t, app, adminToken, item.method, item.path, item.body)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "OperationsPlatform") {
			t.Fatalf("ERP should reject %s %s, got %d: %s", item.method, item.path, rec.Code, rec.Body.String())
		}
	}
}

func TestStandaloneProductOpsIgnoresLegacyTenantBoundary(t *testing.T) {
	app := newTestHTTPApp(t)
	if err := app.store.Mutate(func(data *AppData) error {
		if len(data.Tenants) != 0 || len(data.TenantPolicies) != 0 {
			t.Fatalf("standalone defaults should not create tenant records: %+v / %+v", data.Tenants, data.TenantPolicies)
		}
		data.Companies = append(data.Companies, Company{ID: 2, TenantID: 2, Name: "华东材料有限公司", Code: "EAST", Status: "active"})
		data.Sites = append(data.Sites, Site{ID: 99, CompanyID: 2, Name: "华东站", Code: "EAST-SITE", Address: "杭州", Longitude: 120.1, Latitude: 30.2, Status: "running"})
		data.Customers = append(data.Customers, Customer{ID: 99, CompanyID: 2, Name: "华东客户", Contact: "赵总", Phone: "13899990000", CreditLimit: 100000, Status: "active"})
		data.Projects = append(data.Projects, Project{ID: 99, CustomerID: 99, Name: "华东项目", Address: "杭州项目", Contact: "赵总", Phone: "13899990000", Longitude: 120.1, Latitude: 30.2, Status: "active"})
		data.Orders = append(data.Orders, SalesOrder{
			ID: 99, OrderNo: "SO-LEGACY-ORG", CustomerID: 99, ProjectID: 99, ProductID: 1, SiteID: 99,
			ProductLine: "concrete", PlanQuantity: 10, Unit: "m3", UnitPrice: 500,
			PlanTime: "2026-06-20 09:00:00", ReceiveAddress: "杭州项目", Contact: "赵总", Phone: "13899990000",
			SettlementMode: "月结", TransportMode: "自有车队", StrengthGrade: "C30", Slump: "180mm", PouringPart: "基础",
			Status: "approved", CreatedAt: "2026-06-18 15:00:00",
		})
		data.Users = append(data.Users, User{ID: 99, TenantID: 2, CompanyID: 2, Username: "east-admin", DisplayName: "华东管理员", RoleCode: "boss", Status: "active"})
		return nil
	}); err != nil {
		t.Fatalf("seed internal boundary data: %v", err)
	}

	token := testLogin(t, app, "admin", "admin123")
	rec := testRequest(t, app, token, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenants"`) {
		t.Fatalf("bootstrap should not expose tenant management fields: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantId"`) || strings.Contains(rec.Body.String(), `"dataScope":"tenant"`) {
		t.Fatalf("bootstrap should not expose tenant boundary fields: %s", rec.Body.String())
	}
	var bootstrap struct {
		Sites     []Site     `json:"sites"`
		Customers []Customer `json:"customers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if !hasSite(bootstrap.Sites, 99) || !hasCustomer(bootstrap.Customers, 99) {
		t.Fatalf("standalone platform should show all operational records: %+v", bootstrap)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/orders", "")
	if !strings.Contains(rec.Body.String(), "SO-LEGACY-ORG") {
		t.Fatalf("standalone platform should show legacy org order: %s", rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/security", "")
	if !strings.Contains(rec.Body.String(), "east-admin") {
		t.Fatalf("standalone platform should show platform user: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantPolicies"`) {
		t.Fatalf("security response should not expose tenant policy controls: %s", rec.Body.String())
	}
}

func TestProductAlertEnterpriseChannelDeliveryUsesAuthSignatureAndPayload(t *testing.T) {
	t.Skip("product operations moved to OperationsPlatform; ERP rejects /api/product-ops/*")
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	var received struct {
		authorization string
		channelToken  string
		timestamp     string
		signature     string
		payload       map[string]interface{}
	}
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		received.authorization = r.Header.Get("Authorization")
		received.channelToken = r.Header.Get("X-CBMP-Channel-Token")
		received.timestamp = r.Header.Get("X-CBMP-Timestamp")
		received.signature = r.Header.Get("X-CBMP-Signature")
		if err := json.NewDecoder(r.Body).Decode(&received.payload); err != nil {
			t.Fatalf("decode notification payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer receiver.Close()

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/channels", `{"name":"企业微信值班群","code":"enterprise_wechat","type":"enterprise_wechat","endpoint":"`+receiver.URL+`","token":"wechat-token","secret":"notify-secret","status":"active","retryLimit":3,"timeoutSeconds":2,"remark":"真实企业微信机器人"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save enterprise wechat alert channel status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/policies", `{"name":"服务端严重告警企业微信通知","source":"server","component":"server","metric":"all","severity":"critical","aggregateWindowMinutes":30,"suppressMinutes":0,"escalateAfterMinutes":0,"escalateTo":"on_call_manager","notifyChannels":["enterprise_wechat"],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save alert policy status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts", `{"instanceId":1,"severity":"critical","source":"server","title":"服务端数据库连接池耗尽","message":"客户现场服务端数据库连接池耗尽"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create alert status %d: %s", rec.Code, rec.Body.String())
	}
	if received.authorization != "Bearer wechat-token" || received.channelToken != "wechat-token" {
		t.Fatalf("expected notification auth headers, got Authorization=%q token=%q", received.authorization, received.channelToken)
	}
	if received.timestamp == "" || received.signature == "" {
		t.Fatalf("expected signed notification headers, got timestamp=%q signature=%q", received.timestamp, received.signature)
	}
	raw, _ := json.Marshal(received.payload)
	if expected := taxGatewaySignature("notify-secret", received.timestamp, raw); !hmac.Equal([]byte(expected), []byte(received.signature)) {
		t.Fatalf("unexpected notification signature: got %s want %s", received.signature, expected)
	}
	if received.payload["msgtype"] != "markdown" {
		t.Fatalf("expected enterprise wechat markdown payload: %+v", received.payload)
	}
	cbmp, _ := received.payload["cbmp"].(map[string]interface{})
	if cbmp["schema"] != "cbmp.alert.v1" || cbmp["channel"] != "enterprise_wechat" || cbmp["channelCode"] != "enterprise_wechat" || cbmp["customerName"] != "湾区建材集团" {
		t.Fatalf("expected cbmp alert payload metadata, got %+v", received.payload)
	}
	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview ProductOpsOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	foundDelivered := false
	for _, item := range overview.AlertNotifications {
		if item.Channel == "enterprise_wechat" && item.Status == "delivered" && item.ChannelNo != "" && item.Endpoint == receiver.URL {
			foundDelivered = true
			break
		}
	}
	if !foundDelivered {
		t.Fatalf("expected delivered enterprise wechat notification in overview: %+v", overview.AlertNotifications)
	}
}

func TestCustomerContactsAndBlacklistBlockOrders(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/customer-contacts", `{"customerId":1,"name":"赵工","phone":"13800019999","role":"项目副联系人","isDefault":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer contact status %d: %s", rec.Code, rec.Body.String())
	}
	var contact CustomerContact
	if err := json.Unmarshal(rec.Body.Bytes(), &contact); err != nil {
		t.Fatalf("decode customer contact: %v", err)
	}
	if !contact.IsDefault || contact.ID == 0 {
		t.Fatalf("expected default contact, got %+v", contact)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	var bootstrap struct {
		Customers        []Customer        `json:"customers"`
		CustomerContacts []CustomerContact `json:"customerContacts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	var updatedCustomer Customer
	for _, item := range bootstrap.Customers {
		if item.ID == 1 {
			updatedCustomer = item
		}
	}
	if updatedCustomer.Contact != "赵工" || updatedCustomer.Phone != "13800019999" {
		t.Fatalf("expected customer primary contact updated, got %+v", updatedCustomer)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-profiles/evaluate", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("evaluate customer profiles status %d: %s", rec.Code, rec.Body.String())
	}
	var profiles []CustomerProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profiles); err != nil {
		t.Fatalf("decode customer profiles: %v", err)
	}
	if len(profiles) == 0 || profiles[0].Grade == "" || profiles[0].CreditScore == 0 {
		t.Fatalf("expected evaluated customer profiles, got %+v", profiles)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-profiles", `{"customerId":1,"grade":"A","riskLevel":"low","creditScore":96,"tags":["战略客户"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save customer profile status %d: %s", rec.Code, rec.Body.String())
	}
	var profile CustomerProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode customer profile: %v", err)
	}
	if profile.Grade != "A" || profile.RiskLevel != "low" {
		t.Fatalf("expected saved customer profile, got %+v", profile)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-complaints", `{"customerId":1,"projectId":1,"title":"等待时间过长","content":"现场等待时间偏长","level":"medium"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer complaint status %d: %s", rec.Code, rec.Body.String())
	}
	var complaint CustomerComplaint
	if err := json.Unmarshal(rec.Body.Bytes(), &complaint); err != nil {
		t.Fatalf("decode customer complaint: %v", err)
	}
	if complaint.Status != "open" || complaint.ComplaintNo == "" || complaint.SLAHours != 24 || complaint.DueAt == "" || complaint.SLAStatus != "on_track" {
		t.Fatalf("expected open customer complaint, got %+v", complaint)
	}
	if err := app.store.Mutate(func(data *AppData) error {
		index := customerComplaintIndex(*data, complaint.ID)
		if index < 0 {
			t.Fatalf("customer complaint not found")
		}
		data.CustomerComplaints[index].DueAt = "2020-01-01 00:00:00"
		data.CustomerComplaints[index].SLAStatus = "on_track"
		data.CustomerComplaints[index].OverdueHours = 0
		return nil
	}); err != nil {
		t.Fatalf("make complaint overdue: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/master/customer-complaints", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list customer complaints status %d: %s", rec.Code, rec.Body.String())
	}
	var complaints []CustomerComplaint
	if err := json.Unmarshal(rec.Body.Bytes(), &complaints); err != nil {
		t.Fatalf("decode customer complaints: %v", err)
	}
	var listedComplaint CustomerComplaint
	for _, item := range complaints {
		if item.ID == complaint.ID {
			listedComplaint = item
		}
	}
	if listedComplaint.SLAStatus != "overdue" || listedComplaint.OverdueHours == 0 {
		t.Fatalf("expected dynamic overdue customer complaint, got %+v", listedComplaint)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-complaints/"+strconv.FormatInt(complaint.ID, 10)+"/close", `{"resolution":"已优化派车窗口"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("close customer complaint status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &complaint); err != nil {
		t.Fatalf("decode closed customer complaint: %v", err)
	}
	if complaint.Status != "closed" || complaint.Resolution == "" || complaint.SLAStatus != "breached" || complaint.OverdueHours == 0 {
		t.Fatalf("expected closed customer complaint, got %+v", complaint)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/contracts/1/attachments", `{"fileName":"补充协议.pdf","fileType":"supplement","url":"vault://contracts/supplement.pdf","checksum":"sha256:supplement"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create contract attachment status %d: %s", rec.Code, rec.Body.String())
	}
	var attachment ContractAttachment
	if err := json.Unmarshal(rec.Body.Bytes(), &attachment); err != nil {
		t.Fatalf("decode contract attachment: %v", err)
	}
	if attachment.ContractID != 1 || attachment.CustomerID != 1 || attachment.FileName == "" {
		t.Fatalf("expected contract attachment, got %+v", attachment)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap after customer profile status %d: %s", rec.Code, rec.Body.String())
	}
	var enriched struct {
		CustomerProfiles    []CustomerProfile    `json:"customerProfiles"`
		CustomerComplaints  []CustomerComplaint  `json:"customerComplaints"`
		ContractAttachments []ContractAttachment `json:"contractAttachments"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &enriched); err != nil {
		t.Fatalf("decode enriched bootstrap: %v", err)
	}
	if len(enriched.CustomerProfiles) == 0 || len(enriched.CustomerComplaints) == 0 || len(enriched.ContractAttachments) == 0 {
		t.Fatalf("expected enriched customer bootstrap data, got %+v", enriched)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-blacklists", `{"customerId":1,"reason":"逾期回款停供","blockOrders":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer blacklist status %d: %s", rec.Code, rec.Body.String())
	}
	var blacklist CustomerBlacklist
	if err := json.Unmarshal(rec.Body.Bytes(), &blacklist); err != nil {
		t.Fatalf("decode customer blacklist: %v", err)
	}
	if blacklist.Status != "active" || !blacklist.BlockOrders {
		t.Fatalf("expected active order-blocking blacklist, got %+v", blacklist)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":1,"planTime":"2026-06-18 16:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "停供") {
		t.Fatalf("expected blacklisted customer order rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customer-blacklists/"+strconv.FormatInt(blacklist.ID, 10)+"/release", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("release customer blacklist status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":1,"planTime":"2026-06-18 16:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected order after blacklist release, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDispatchScheduleAndCarrierSettlement(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders/schedules", `{"siteId":2,"vehicleId":4,"driverId":4,"carrierId":2,"shiftDate":"2026-06-19","shift":"day","capacityQty":90}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dispatch schedule status %d: %s", rec.Code, rec.Body.String())
	}
	var schedule DispatchSchedule
	if err := json.Unmarshal(rec.Body.Bytes(), &schedule); err != nil {
		t.Fatalf("decode dispatch schedule: %v", err)
	}
	if schedule.ID == 0 || schedule.CarrierID != 2 || schedule.CapacityQty != 90 {
		t.Fatalf("expected dispatch schedule, got %+v", schedule)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders/schedules", `{"siteId":2,"vehicleId":4,"driverId":4,"carrierId":2,"shiftDate":"2026-06-19","shift":"day","capacityQty":90}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "已有同班次排班") {
		t.Fatalf("expected duplicate schedule rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders", `{"orderId":2,"vehicleId":4,"planQuantity":20}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dispatch with schedule status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatch DispatchOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatch); err != nil {
		t.Fatalf("decode dispatch: %v", err)
	}
	if dispatch.VehicleID != 4 || dispatch.PlanQuantity != 20 {
		t.Fatalf("expected scheduled vehicle dispatch, got %+v", dispatch)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/dispatch-orders/schedules", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list schedules status %d: %s", rec.Code, rec.Body.String())
	}
	var schedules []DispatchSchedule
	if err := json.Unmarshal(rec.Body.Bytes(), &schedules); err != nil {
		t.Fatalf("decode schedules: %v", err)
	}
	var updatedSchedule DispatchSchedule
	for _, item := range schedules {
		if item.ID == schedule.ID {
			updatedSchedule = item
		}
	}
	if updatedSchedule.AssignedQty != 20 {
		t.Fatalf("expected schedule assigned qty updated, got %+v", updatedSchedule)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders/"+strconv.FormatInt(dispatch.ID, 10)+"/status", `{"status":"completed"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("complete dispatch status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders/carrier-settlements/generate", `{"carrierId":2,"period":"2026-06","ratePerTrip":500}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("generate carrier settlement status %d: %s", rec.Code, rec.Body.String())
	}
	var result struct {
		Settlement TransportSettlement       `json:"settlement"`
		Items      []TransportSettlementItem `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode carrier settlement: %v", err)
	}
	if result.Settlement.CarrierID != 2 || result.Settlement.TripCount != 1 || result.Settlement.Amount != 500 || len(result.Items) != 1 {
		t.Fatalf("expected one carrier settlement item, got %+v", result)
	}
	if result.Items[0].DispatchID != dispatch.ID || result.Items[0].Amount != 500 {
		t.Fatalf("expected settlement item for dispatch, got %+v", result.Items[0])
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders/carrier-settlements/generate", `{"carrierId":2,"period":"2026-06","ratePerTrip":500}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "没有可生成对账") {
		t.Fatalf("expected duplicate carrier settlement rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeliverySignLinkPublicSignAndAttachments(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders", `{"orderId":2,"vehicleId":4,"planQuantity":16}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dispatch status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatch DispatchOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatch); err != nil {
		t.Fatalf("decode dispatch: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets", `{"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`,"grossWeight":34,"tareWeight":18}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create outbound ticket status %d: %s", rec.Code, rec.Body.String())
	}
	var ticket ScaleTicket
	if err := json.Unmarshal(rec.Body.Bytes(), &ticket); err != nil {
		t.Fatalf("decode ticket: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/sign-links", `{"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`,"channel":"sms","phone":"13900001111"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create sign link status %d: %s", rec.Code, rec.Body.String())
	}
	var link DeliverySignLink
	if err := json.Unmarshal(rec.Body.Bytes(), &link); err != nil {
		t.Fatalf("decode sign link: %v", err)
	}
	if link.Token == "" || link.URL == "" || link.TicketID != ticket.ID || link.Status != "sent" {
		t.Fatalf("expected sent sign link with ticket, got %+v", link)
	}

	rec = testRequest(t, app, "", http.MethodGet, "/api/public/delivery-sign/"+link.Token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("public sign detail status %d: %s", rec.Code, rec.Body.String())
	}
	var detail publicDeliverySignDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode public detail: %v", err)
	}
	if detail.Dispatch.ID != dispatch.ID || detail.Ticket.ID != ticket.ID || detail.Project == "" {
		t.Fatalf("expected public detail for dispatch, got %+v", detail)
	}

	rec = testRequest(t, app, "", http.MethodPost, "/api/public/delivery-sign/"+link.Token, `{"signer":"赵工","phone":"13900001111","signedQty":16,"photo":"minio://delivery/site.jpg","signature":"赵工电子签名","remark":"现场验收","attachments":[{"fileName":"site.jpg","fileType":"photo","url":"minio://delivery/site.jpg","checksum":"sha256:site"}]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("public sign status %d: %s", rec.Code, rec.Body.String())
	}
	var sign DeliverySign
	if err := json.Unmarshal(rec.Body.Bytes(), &sign); err != nil {
		t.Fatalf("decode sign: %v", err)
	}
	if sign.LinkID != link.ID || sign.TicketID != ticket.ID || sign.SignedQty != 16 {
		t.Fatalf("expected public sign bound to link and ticket, got %+v", sign)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/delivery/sign-links", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list sign links status %d: %s", rec.Code, rec.Body.String())
	}
	var links []DeliverySignLink
	if err := json.Unmarshal(rec.Body.Bytes(), &links); err != nil {
		t.Fatalf("decode links: %v", err)
	}
	var used DeliverySignLink
	for _, item := range links {
		if item.ID == link.ID {
			used = item
		}
	}
	if used.Status != "used" || used.UsedAt == "" {
		t.Fatalf("expected used sign link, got %+v", used)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/delivery/sign-attachments", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list sign attachments status %d: %s", rec.Code, rec.Body.String())
	}
	var attachments []DeliverySignAttachment
	if err := json.Unmarshal(rec.Body.Bytes(), &attachments); err != nil {
		t.Fatalf("decode attachments: %v", err)
	}
	var foundAttachment bool
	for _, item := range attachments {
		if item.SignID == sign.ID && item.URL == "minio://delivery/site.jpg" {
			foundAttachment = true
		}
	}
	if !foundAttachment {
		t.Fatalf("expected sign photo attachment, got %+v", attachments)
	}

	rec = testRequest(t, app, "", http.MethodPost, "/api/public/delivery-sign/"+link.Token, `{"signer":"重复签收"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "签收链接已失效") {
		t.Fatalf("expected used link rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPricingPolicyEvaluationAndBelowFloorApproval(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/tax-rates", `{"name":"建材销售 6%","rate":0.06,"scope":"sales"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create tax rate status %d: %s", rec.Code, rec.Body.String())
	}
	var tax TaxRate
	if err := json.Unmarshal(rec.Body.Bytes(), &tax); err != nil {
		t.Fatalf("decode tax rate: %v", err)
	}
	if tax.ID == 0 || tax.Rate != 0.06 {
		t.Fatalf("expected tax rate, got %+v", tax)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/price-policies", `{"customerId":1,"projectId":1,"productId":1,"customerGrade":"A","floorPrice":505,"salePrice":515,"taxRateId":`+strconv.FormatInt(tax.ID, 10)+`,"effectiveFrom":"2026-06-01","effectiveTo":"2027-05-31"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create price policy status %d: %s", rec.Code, rec.Body.String())
	}
	var policy PricePolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &policy); err != nil {
		t.Fatalf("decode price policy: %v", err)
	}
	if policy.ID == 0 || policy.CustomerGrade != "A" || policy.SalePrice != 515 {
		t.Fatalf("expected price policy, got %+v", policy)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/pricing/evaluate", `{"customerId":1,"projectId":1,"productId":1,"planTime":"2026-06-20 10:00:00"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("evaluate pricing status %d: %s", rec.Code, rec.Body.String())
	}
	var quote PricingQuote
	if err := json.Unmarshal(rec.Body.Bytes(), &quote); err != nil {
		t.Fatalf("decode quote: %v", err)
	}
	if quote.PolicyID != policy.ID || quote.UnitPrice != 515 || quote.FloorPrice != 505 || quote.TaxRate != 0.06 {
		t.Fatalf("expected policy quote, got %+v", quote)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/orders", `{"customerId":1,"projectId":1,"productId":1,"siteId":1,"planQuantity":1,"unitPrice":500,"planTime":"2026-06-20 10:00:00","settlementMode":"月结","transportMode":"自有车队"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create below floor order status %d: %s", rec.Code, rec.Body.String())
	}
	var order SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode order: %v", err)
	}
	if order.Status != "pending_approval" || order.RiskFlag != "price_below_floor" || order.UnitPrice != 500 {
		t.Fatalf("expected below floor approval order, got %+v", order)
	}
	tasks := fetchApprovalTasks(t, app, token)
	var found bool
	for _, task := range tasks {
		if task.Resource == "sales_order" && task.ResourceID == order.ID && task.FlowCode == "price_below_floor" {
			found = strings.Contains(task.Reason, "低于底价")
		}
	}
	if !found {
		t.Fatalf("expected price below floor approval task, got %+v", tasks)
	}
}

func TestMasterDataBulkImportExport(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/master/export?resource=customers", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("export customers status %d: %s", rec.Code, rec.Body.String())
	}
	var exported MasterDataExport
	if err := json.Unmarshal(rec.Body.Bytes(), &exported); err != nil {
		t.Fatalf("decode customer export: %v", err)
	}
	if exported.Resource != "customers" || exported.Count == 0 || !containsString(exported.Fields, "name") {
		t.Fatalf("expected customer export fields, got %+v", exported)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/import", `{"resource":"materials","mode":"create","rows":[{"name":"矿粉","spec":"S95","unit":"t","safeStock":120},{"spec":"缺少名称"}]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("import materials status %d: %s", rec.Code, rec.Body.String())
	}
	var result MasterDataImportResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode import result: %v", err)
	}
	if result.Created != 1 || len(result.Errors) != 1 {
		t.Fatalf("expected partial material import, got %+v", result)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/import", `{"resource":"products","mode":"upsert","rows":[{"id":1,"line":"concrete","name":"预拌混凝土","spec":"C30 P8 年度版","unit":"m3","basePrice":530,"costPrice":365,"requiresMix":true,"status":"active"}]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upsert products status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode upsert result: %v", err)
	}
	if result.Updated != 1 || result.Created != 0 || len(result.Errors) != 0 {
		t.Fatalf("expected product upsert, got %+v", result)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/master/export?resource=products", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("export products status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &exported); err != nil {
		t.Fatalf("decode product export: %v", err)
	}
	var foundUpdated bool
	for _, row := range exported.Rows {
		if row["id"] == float64(1) && row["spec"] == "C30 P8 年度版" {
			foundUpdated = true
		}
	}
	if !foundUpdated {
		t.Fatalf("expected updated product in export, got %+v", exported.Rows)
	}
}

func TestManagementReportsCoverOperatingAgingQualityAndEnergy(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	qualityToken := testLogin(t, app, "quality", "quality123")

	rec := testRequest(t, app, qualityToken, http.MethodPost, "/api/quality/inspections", `{"batchId":1,"slump":"180mm","temperature":28.5,"remark":"报表质检样本"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create report quality inspection status %d: %s", rec.Code, rec.Body.String())
	}
	qualityOverview := fetchQualityOverview(t, app, qualityToken)
	for _, sample := range qualityOverview.Samples {
		rec = testRequest(t, app, qualityToken, http.MethodPost, "/api/quality/samples/"+strconv.FormatInt(sample.ID, 10)+"/test", `{"strength":36.8,"result":"passed"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("test report quality sample status %d: %s", rec.Code, rec.Body.String())
		}
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/reports", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("reports status %d: %s", rec.Code, rec.Body.String())
	}
	var reports ManagementReports
	if err := json.Unmarshal(rec.Body.Bytes(), &reports); err != nil {
		t.Fatalf("decode reports: %v", err)
	}
	if reports.Operating.Revenue <= 0 || reports.Operating.GrossProfit == 0 || reports.Operating.InventoryWarningCount < 0 {
		t.Fatalf("expected operating analysis, got %+v", reports.Operating)
	}
	if len(reports.ProjectProfit) == 0 || reports.ProjectProfit[0].Margin <= 0 {
		t.Fatalf("expected project profit report, got %+v", reports.ProjectProfit)
	}
	if len(reports.CustomerAging) == 0 || reports.AgingBuckets == nil {
		t.Fatalf("expected customer aging reports, got aging=%+v buckets=%+v", reports.CustomerAging, reports.AgingBuckets)
	}
	if reports.Quality.Inspections == 0 || reports.Quality.PassRate <= 0 {
		t.Fatalf("expected quality analysis, got %+v", reports.Quality)
	}
	if len(reports.Energy) == 0 || reports.Energy[0].ProducedQty <= 0 || reports.Energy[0].UnitPowerKwh <= 0 {
		t.Fatalf("expected energy analysis, got %+v", reports.Energy)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/dashboard", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard status %d: %s", rec.Code, rec.Body.String())
	}
	var dashboard struct {
		KPIs    map[string]float64       `json:"kpis"`
		Quality QualityAnalysisReport    `json:"quality"`
		Energy  []ProductionEnergyReport `json:"energy"`
		Aging   []CustomerAgingReport    `json:"customerAging"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if dashboard.KPIs["grossMargin"] <= 0 || dashboard.KPIs["qualityPassRate"] <= 0 || dashboard.KPIs["unitPowerKwh"] <= 0 {
		t.Fatalf("expected dashboard management KPIs, got %+v", dashboard.KPIs)
	}
	if dashboard.Quality.Inspections == 0 || len(dashboard.Energy) == 0 || len(dashboard.Aging) == 0 {
		t.Fatalf("expected dashboard report payload, got %+v", dashboard)
	}
}

func TestCustomerPortalComplaintSelfService(t *testing.T) {
	app := newTestHTTPApp(t)
	customerToken := testLogin(t, app, "customer", "customer123")

	rec := testRequest(t, app, customerToken, http.MethodGet, "/api/portal/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("portal overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview PortalOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode portal overview: %v", err)
	}
	if len(overview.Orders) == 0 || len(overview.Dispatches) == 0 {
		t.Fatalf("expected scoped portal overview, got %+v", overview)
	}
	for _, order := range overview.Orders {
		if order.CustomerID != 1 {
			t.Fatalf("customer saw order for customer %d", order.CustomerID)
		}
	}

	rec = testRequest(t, app, customerToken, http.MethodPost, "/api/portal/complaints", `{"projectId":1,"title":"客户自助反馈","content":"工地等待时间过长","level":"high"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create portal complaint status %d: %s", rec.Code, rec.Body.String())
	}
	var complaint CustomerComplaint
	if err := json.Unmarshal(rec.Body.Bytes(), &complaint); err != nil {
		t.Fatalf("decode portal complaint: %v", err)
	}
	if complaint.CustomerID != 1 || complaint.ProjectID != 1 || complaint.ComplaintNo == "" || complaint.DueAt == "" || complaint.SLAStatus == "" {
		t.Fatalf("expected scoped SLA complaint, got %+v", complaint)
	}

	rec = testRequest(t, app, customerToken, http.MethodGet, "/api/portal/complaints", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("portal complaints status %d: %s", rec.Code, rec.Body.String())
	}
	var complaints []CustomerComplaint
	if err := json.Unmarshal(rec.Body.Bytes(), &complaints); err != nil {
		t.Fatalf("decode portal complaints: %v", err)
	}
	var found bool
	for _, item := range complaints {
		if item.ID == complaint.ID && item.CustomerID == 1 {
			found = true
		}
		if item.CustomerID != 1 {
			t.Fatalf("customer saw complaint for customer %d", item.CustomerID)
		}
	}
	if !found {
		t.Fatalf("expected created complaint in portal list, got %+v", complaints)
	}

	rec = testRequest(t, app, customerToken, http.MethodPost, "/api/portal/complaints", `{"projectId":2,"title":"越权项目","content":"不应成功","level":"high"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "项目不属于当前客户") {
		t.Fatalf("expected cross-customer project rejected, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDriverPortalExceptionReporting(t *testing.T) {
	app := newTestHTTPApp(t)
	driverToken := testLogin(t, app, "driver", "driver123")

	rec := testRequest(t, app, driverToken, http.MethodGet, "/api/portal/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("driver portal overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview PortalOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode driver portal overview: %v", err)
	}
	if len(overview.Dispatches) != 1 || overview.Dispatches[0].DriverID != 1 {
		t.Fatalf("expected driver scoped dispatches, got %+v", overview.Dispatches)
	}

	rec = testRequest(t, app, driverToken, http.MethodPost, "/api/portal/dispatches/1/exception", `{"exception":"现场道路拥堵，预计延误 30 分钟","level":"high"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("driver exception status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatch DispatchOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatch); err != nil {
		t.Fatalf("decode driver exception dispatch: %v", err)
	}
	if dispatch.ID != 1 || !strings.Contains(dispatch.Exception, "道路拥堵") {
		t.Fatalf("expected dispatch exception persisted, got %+v", dispatch)
	}

	rec = testRequest(t, app, driverToken, http.MethodGet, "/api/portal/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("driver portal overview after exception status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode driver portal overview after exception: %v", err)
	}
	if len(overview.Alarms) == 0 || overview.Alarms[len(overview.Alarms)-1].AlarmType != "driver_exception" {
		t.Fatalf("expected driver exception alarm in overview, got %+v", overview.Alarms)
	}

	rec = testRequest(t, app, driverToken, http.MethodPost, "/api/portal/dispatches/2/exception", `{"exception":"不属于我的派车单","level":"high"}`)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "无权上报该派车单异常") {
		t.Fatalf("expected forbidden cross-driver exception, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdvancedPricePolicyRegionTierPromotion(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/price-policies", `{
		"customerId":1,
		"projectId":1,
		"productId":1,
		"customerGrade":"A",
		"region":"南山",
		"minQuantity":50,
		"maxQuantity":100,
		"floorPrice":480,
		"salePrice":540,
		"promotionName":"南山大方量九折",
		"promotionType":"percent",
		"promotionValue":0.1,
		"priority":30,
		"taxRateId":1,
		"effectiveFrom":"2026-06-01",
		"effectiveTo":"2027-05-31"
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create advanced price policy status %d: %s", rec.Code, rec.Body.String())
	}
	var policy PricePolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &policy); err != nil {
		t.Fatalf("decode advanced price policy: %v", err)
	}
	if policy.Region != "南山" || policy.MinQuantity != 50 || policy.PromotionType != "percent" || policy.Priority != 30 {
		t.Fatalf("expected advanced price policy, got %+v", policy)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/pricing/evaluate", `{"customerId":1,"projectId":1,"productId":1,"planQuantity":60,"planTime":"2026-06-20 10:00:00"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("evaluate advanced pricing status %d: %s", rec.Code, rec.Body.String())
	}
	var quote PricingQuote
	if err := json.Unmarshal(rec.Body.Bytes(), &quote); err != nil {
		t.Fatalf("decode advanced quote: %v", err)
	}
	if quote.PolicyID != policy.ID || quote.ListPrice != 540 || quote.UnitPrice != 486 || quote.PromotionAmount != 54 || quote.FloorPrice != 480 {
		t.Fatalf("expected promotional tier quote, got %+v", quote)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/pricing/evaluate", `{"customerId":1,"projectId":1,"productId":1,"planQuantity":20,"planTime":"2026-06-20 10:00:00"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("evaluate small quantity pricing status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &quote); err != nil {
		t.Fatalf("decode small quantity quote: %v", err)
	}
	if quote.PolicyID == policy.ID || quote.UnitPrice != 520 || quote.PromotionAmount != 0 {
		t.Fatalf("expected small quantity to fall back to base customer policy, got %+v", quote)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/orders", `{
		"customerId":1,
		"projectId":1,
		"siteId":1,
		"planTime":"2026-06-20 10:00:00",
		"settlementMode":"月结",
		"transportMode":"自有车队",
		"lines":[
			{"productId":1,"quantity":60,"strengthGrade":"C30","slump":"180mm","pouringPart":"底板"}
		]
	}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create advanced price order status %d: %s", rec.Code, rec.Body.String())
	}
	var order SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode advanced price order: %v", err)
	}
	if order.Status != "submitted" || order.RiskFlag != "" || order.UnitPrice != 486 || order.TotalAmount != 29160 {
		t.Fatalf("expected order priced by advanced policy, got %+v", order)
	}
	if len(order.Lines) != 1 || order.Lines[0].UnitPrice != 486 || order.Lines[0].PriceSource != "price_policy" {
		t.Fatalf("expected line priced by advanced policy, got %+v", order.Lines)
	}
}

func TestSalesOrderSupportsMultiplePricedLines(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	payload := `{
		"customerId":1,
		"projectId":1,
		"siteId":1,
		"planTime":"2026-06-20 10:00:00",
		"settlementMode":"月结",
		"transportMode":"自有车队",
		"pumpMode":"车载泵",
		"lines":[
			{"productId":1,"quantity":10,"strengthGrade":"C30","slump":"180mm","pouringPart":"底板"},
			{"productId":2,"quantity":5,"strengthGrade":"C40","slump":"160mm","pouringPart":"剪力墙"}
		]
	}`
	rec := testRequest(t, app, token, http.MethodPost, "/api/orders", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create multi-line order status %d: %s", rec.Code, rec.Body.String())
	}
	var order SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode multi-line order: %v", err)
	}
	if order.Status != "submitted" || order.RiskFlag != "" {
		t.Fatalf("expected submitted multi-line order, got %+v", order)
	}
	if order.ProductID != 1 || order.PlanQuantity != 15 || order.UnitPrice != 550 || order.TotalAmount != 8250 {
		t.Fatalf("unexpected multi-line order totals: %+v", order)
	}
	if len(order.Lines) != 2 {
		t.Fatalf("expected two order lines, got %+v", order.Lines)
	}
	if order.Lines[0].ProductID != 1 || order.Lines[0].UnitPrice != 520 || order.Lines[0].Amount != 5200 || order.Lines[0].PriceSource != "price_policy" {
		t.Fatalf("unexpected first order line: %+v", order.Lines[0])
	}
	if order.Lines[1].ProductID != 2 || order.Lines[1].UnitPrice != 610 || order.Lines[1].Amount != 3050 || order.Lines[1].PriceSource != "price_policy" {
		t.Fatalf("unexpected second order line: %+v", order.Lines[1])
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/orders/"+strconv.FormatInt(order.ID, 10)+"/approve", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve multi-line order status %d: %s", rec.Code, rec.Body.String())
	}

	secondLine := order.Lines[1]
	rec = testRequest(t, app, token, http.MethodPost, "/api/dispatch-orders", `{"orderId":`+strconv.FormatInt(order.ID, 10)+`,"lineId":`+strconv.FormatInt(secondLine.ID, 10)+`,"vehicleId":4,"planQuantity":5}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create line dispatch status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatch DispatchOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatch); err != nil {
		t.Fatalf("decode line dispatch: %v", err)
	}
	if dispatch.LineID != secondLine.ID || dispatch.LineSeq != secondLine.Seq || dispatch.ProductID != secondLine.ProductID || dispatch.PlanQuantity != 5 {
		t.Fatalf("expected dispatch bound to second line, got %+v", dispatch)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/weighbridge/tickets", `{"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`,"grossWeight":31,"tareWeight":18}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create line ticket status %d: %s", rec.Code, rec.Body.String())
	}
	var ticket ScaleTicket
	if err := json.Unmarshal(rec.Body.Bytes(), &ticket); err != nil {
		t.Fatalf("decode line ticket: %v", err)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/delivery/sign", `{"dispatchId":`+strconv.FormatInt(dispatch.ID, 10)+`,"ticketId":`+strconv.FormatInt(ticket.ID, 10)+`,"signer":"赵工","signedQty":5}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("sign line dispatch status %d: %s", rec.Code, rec.Body.String())
	}
	var sign DeliverySign
	if err := json.Unmarshal(rec.Body.Bytes(), &sign); err != nil {
		t.Fatalf("decode line sign: %v", err)
	}
	if sign.LineID != secondLine.ID || sign.ProductID != secondLine.ProductID || sign.SignedQty != 5 {
		t.Fatalf("expected line-level sign, got %+v", sign)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/statements", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("statements status %d: %s", rec.Code, rec.Body.String())
	}
	var statements []Statement
	if err := json.Unmarshal(rec.Body.Bytes(), &statements); err != nil {
		t.Fatalf("decode statements: %v", err)
	}
	var foundStatementItem bool
	for _, statement := range statements {
		for _, item := range statement.Items {
			if item.SignID == sign.ID {
				foundStatementItem = item.LineID == secondLine.ID && item.ProductID == secondLine.ProductID && item.UnitPrice == secondLine.UnitPrice && item.Amount == 3050
			}
		}
	}
	if !foundStatementItem {
		t.Fatalf("expected statement item priced by signed order line, got %+v", statements)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/orders", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("orders list status %d: %s", rec.Code, rec.Body.String())
	}
	var orders []SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &orders); err != nil {
		t.Fatalf("decode orders list: %v", err)
	}
	var listed SalesOrder
	for _, item := range orders {
		if item.ID == order.ID {
			listed = item
		}
	}
	if len(listed.Lines) != 2 || listed.TotalAmount != 8250 {
		t.Fatalf("expected listed order with lines, got %+v", listed)
	}
}

func TestContractVersionApprovalFlow(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	revisionPayload := `{
		"changeReason":"年度价格条款调整",
		"totalAmount":5600000,
		"items":[
			{"productId":1,"unit":"m3","quantity":11000,"unitPrice":530},
			{"productId":2,"unit":"m3","quantity":3500,"unitPrice":620}
		]
	}`
	rec := testRequest(t, app, token, http.MethodPost, "/api/contracts/1/revise", revisionPayload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("revise contract status %d: %s", rec.Code, rec.Body.String())
	}
	var revision Contract
	if err := json.Unmarshal(rec.Body.Bytes(), &revision); err != nil {
		t.Fatalf("decode contract revision: %v", err)
	}
	if revision.ID == 0 || revision.ParentID != 1 || revision.Version != 2 || revision.Status != "draft" || revision.Items[0].UnitPrice != 530 {
		t.Fatalf("unexpected contract revision: %+v", revision)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/contracts/"+strconv.FormatInt(revision.ID, 10)+"/submit", `{"reason":"年度价格条款调整"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("submit contract status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &revision); err != nil {
		t.Fatalf("decode submitted contract: %v", err)
	}
	if revision.Status != "pending_approval" || revision.SubmittedAt == "" {
		t.Fatalf("expected pending approval contract, got %+v", revision)
	}

	tasks := fetchApprovalTasks(t, app, token)
	var task ApprovalTask
	for _, item := range tasks {
		if item.Resource == "contract" && item.ResourceID == revision.ID && item.FlowCode == "contract_version" {
			task = item
		}
	}
	if task.ID == 0 || task.Status != "pending" || task.CurrentRole == "" {
		t.Fatalf("expected contract version approval task, got %+v", tasks)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+strconv.FormatInt(task.ID, 10)+"/act", `{"action":"approve","comment":"调度确认合同明细"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve contract first step status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/approvals/"+strconv.FormatInt(task.ID, 10)+"/act", `{"action":"approve","comment":"高管确认生效"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve contract final step status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/contracts", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("contracts list status %d: %s", rec.Code, rec.Body.String())
	}
	var contracts []Contract
	if err := json.Unmarshal(rec.Body.Bytes(), &contracts); err != nil {
		t.Fatalf("decode contracts: %v", err)
	}
	var oldVersion, newVersion Contract
	for _, item := range contracts {
		if item.ID == 1 {
			oldVersion = item
		}
		if item.ID == revision.ID {
			newVersion = item
		}
	}
	if oldVersion.Status != "superseded" {
		t.Fatalf("expected old contract superseded, got %+v", oldVersion)
	}
	if newVersion.Status != "active" || newVersion.ApprovedAt == "" || newVersion.ApprovedBy == "" {
		t.Fatalf("expected new contract active, got %+v", newVersion)
	}
}

func hasSite(items []Site, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasCustomer(items []Customer, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func signLicenseForTest(item LicensePackage, privateKey ed25519.PrivateKey) (LicensePackage, error) {
	payload, err := licenseCanonicalPayload(item)
	if err != nil {
		return item, err
	}
	item.Signature = "ed25519:" + base64.RawStdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	item.PublicKey = "ed25519:" + base64.RawStdEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey))
	return item, nil
}

func testLogin(t *testing.T, app *App, username, password string) string {
	t.Helper()
	body := bytes.NewBufferString(`{"username":"` + username + `","password":"` + password + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s status %d: %s", username, rec.Code, rec.Body.String())
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("empty token")
	}
	return payload.Token
}

func testRequest(t *testing.T, app *App, token, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	return rec
}

func TestEnterpriseAPIsCoverProcurementFinanceRulesAndUpdates(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/procurement/receipts", `{"purchaseOrderId":1,"siteId":1,"grossWeight":55.4,"tareWeight":18.2,"qualityStatus":"passed","status":"stocked"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("raw receipt status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/finance/invoices", `{"statementId":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("invoice status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/finance/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finance overview status %d: %s", rec.Code, rec.Body.String())
	}
	var finance struct {
		Receivables []Receivable `json:"receivables"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &finance); err != nil {
		t.Fatalf("decode finance: %v", err)
	}
	if len(finance.Receivables) == 0 {
		t.Fatalf("expected receivables")
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/finance/receipts", `{"receivableId":1,"amount":1000,"method":"bank"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("receipt status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/rules/evaluate", `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("rule evaluate status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/updates/2/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("update download status %d: %s", rec.Code, rec.Body.String())
	}
	var download UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode update download: %v", err)
	}
	if !download.Verified || download.Package.ID != 2 || !strings.Contains(download.FileName, download.Package.Version) {
		t.Fatalf("unexpected update download: %+v", download)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/updates/2/apply", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("update apply status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdatePackageArtifactPublishDownloadAndValidation(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	artifact := []byte(`{"component":"client","version":"2.2.0","payload":"客户现场客户端补丁"}`)
	sum := sha256.Sum256(artifact)
	checksum := "sha256:" + hex.EncodeToString(sum[:])
	payload, err := json.Marshal(map[string]interface{}{
		"version":               "2.2.0",
		"component":             "client",
		"channel":               "stable",
		"status":                "available",
		"checksum":              checksum,
		"signature":             "sig:client-220",
		"fileName":              "cbmp-client-2.2.0.pkg",
		"artifactContentType":   "application/octet-stream",
		"artifactContentBase64": base64.StdEncoding.EncodeToString(artifact),
		"rollbackVersion":       "2.1.0",
		"remark":                "真实客户端更新包",
	})
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/updates", string(payload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish artifact update status %d: %s", rec.Code, rec.Body.String())
	}
	var saved UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatalf("decode saved update: %v", err)
	}
	if saved.ID == 0 || saved.ArtifactSHA256 != checksum || saved.ArtifactSizeBytes != int64(len(artifact)) {
		t.Fatalf("unexpected saved artifact metadata: %+v", saved)
	}
	if !strings.HasPrefix(saved.Signature, "hmac-sha256:") {
		t.Fatalf("expected server generated hmac signature: %+v", saved)
	}
	if saved.ArtifactContentBase64 != "" {
		t.Fatalf("publish response should not expose artifact content")
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/updates", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list updates status %d: %s", rec.Code, rec.Body.String())
	}
	var updates []UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &updates); err != nil {
		t.Fatalf("decode updates: %v", err)
	}
	found := false
	for _, item := range updates {
		if item.ID == saved.ID {
			found = true
			if item.ArtifactContentBase64 != "" || item.ArtifactSHA256 != checksum {
				t.Fatalf("list should expose only artifact metadata: %+v", item)
			}
		}
	}
	if !found {
		t.Fatalf("published update missing from list")
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/updates/"+strconv.FormatInt(saved.ID, 10)+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("download artifact update status %d: %s", rec.Code, rec.Body.String())
	}
	var download UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode artifact download: %v", err)
	}
	if !download.Verified || download.ArtifactSHA256 != checksum || download.ArtifactSizeBytes != int64(len(artifact)) {
		t.Fatalf("unexpected artifact download metadata: %+v", download)
	}
	decoded, err := base64.StdEncoding.DecodeString(download.ArtifactContentBase64)
	if err != nil {
		t.Fatalf("decode artifact content: %v", err)
	}
	if !bytes.Equal(decoded, artifact) {
		t.Fatalf("downloaded artifact content mismatch")
	}
	if download.Package.ArtifactContentBase64 != "" {
		t.Fatalf("download package metadata should not duplicate artifact content")
	}

	badPayload, err := json.Marshal(map[string]interface{}{
		"version":               "2.2.1",
		"component":             "client",
		"checksum":              "sha256:" + strings.Repeat("0", 64),
		"signature":             "sig:client-221",
		"fileName":              "cbmp-client-2.2.1.pkg",
		"artifactContentBase64": base64.StdEncoding.EncodeToString(artifact),
	})
	if err != nil {
		t.Fatalf("marshal bad update payload: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/updates", string(badPayload))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected checksum mismatch rejection, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdatePackageDeltaArtifactPublishDownload(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	baseArtifact := []byte("hello world")
	targetArtifact := []byte("hello updater")
	baseSum := sha256.Sum256(baseArtifact)
	targetSum := sha256.Sum256(targetArtifact)
	baseChecksum := "sha256:" + hex.EncodeToString(baseSum[:])
	targetChecksum := "sha256:" + hex.EncodeToString(targetSum[:])
	patch, err := json.Marshal(map[string]interface{}{
		"algorithm":      "cbmp-copy-v1",
		"baseSha256":     baseChecksum,
		"targetSha256":   targetChecksum,
		"targetFileName": "cbmp-server-5.1.0.tar.gz",
		"ops": []map[string]interface{}{
			{"copy": map[string]interface{}{"offset": 0, "length": 6}},
			{"data": base64.StdEncoding.EncodeToString([]byte("updater"))},
		},
	})
	if err != nil {
		t.Fatalf("marshal delta patch: %v", err)
	}
	patchSum := sha256.Sum256(patch)
	patchChecksum := "sha256:" + hex.EncodeToString(patchSum[:])
	payload, err := json.Marshal(map[string]interface{}{
		"version":               "5.1.0",
		"component":             "server",
		"channel":               "stable",
		"status":                "available",
		"packageType":           "delta",
		"baseVersion":           "5.0.0",
		"deltaAlgorithm":        "cbmp-copy-v1",
		"baseArtifactSha256":    baseChecksum,
		"targetArtifactSha256":  targetChecksum,
		"checksum":              patchChecksum,
		"signature":             "sig:server-delta-510",
		"fileName":              "cbmp-server-5.1.0.patch.json",
		"artifactContentType":   "application/vnd.cbmp.delta+json",
		"artifactContentBase64": base64.StdEncoding.EncodeToString(patch),
		"rollbackVersion":       "5.0.0",
		"remark":                "服务端差分更新包",
	})
	if err != nil {
		t.Fatalf("marshal delta update payload: %v", err)
	}

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/updates", string(payload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish delta update status %d: %s", rec.Code, rec.Body.String())
	}
	var saved UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatalf("decode saved delta update: %v", err)
	}
	if saved.PackageType != "delta" || saved.BaseVersion != "5.0.0" || saved.DeltaAlgorithm != "cbmp-copy-v1" {
		t.Fatalf("unexpected delta metadata: %+v", saved)
	}
	if saved.ArtifactSHA256 != patchChecksum || saved.TargetArtifactSHA256 != targetChecksum || saved.BaseArtifactSHA256 != baseChecksum {
		t.Fatalf("unexpected delta sha metadata: %+v", saved)
	}
	if !updatePackageVerified(saved) {
		t.Fatalf("saved delta update should verify: %+v", saved)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/updates/"+strconv.FormatInt(saved.ID, 10)+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("download delta update status %d: %s", rec.Code, rec.Body.String())
	}
	var download UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode delta update download: %v", err)
	}
	if !download.Verified || download.Package.PackageType != "delta" || download.Manifest["targetArtifactSha256"] != targetChecksum {
		t.Fatalf("download should carry delta verification metadata: %+v", download)
	}
	decodedPatch, err := base64.StdEncoding.DecodeString(download.ArtifactContentBase64)
	if err != nil {
		t.Fatalf("decode delta artifact content: %v", err)
	}
	if !bytes.Equal(decodedPatch, patch) {
		t.Fatalf("downloaded delta patch mismatch")
	}
}

func TestUpdatePackageArtifactCanUseEd25519Signature(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate update signing key: %v", err)
	}
	t.Setenv("CBMP_UPDATE_SIGNING_PRIVATE_KEY", "ed25519:"+base64.RawStdEncoding.EncodeToString(privateKey))

	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	artifact := []byte("offline-verifiable-server-package")
	sum := sha256.Sum256(artifact)
	checksum := "sha256:" + hex.EncodeToString(sum[:])
	payload, err := json.Marshal(map[string]interface{}{
		"version":               "4.0.0",
		"component":             "server",
		"channel":               "stable",
		"checksum":              checksum,
		"signature":             "sig:placeholder",
		"fileName":              "cbmp-server-4.0.0.tar.gz",
		"artifactContentType":   "application/gzip",
		"artifactContentBase64": base64.StdEncoding.EncodeToString(artifact),
		"rollbackVersion":       "3.9.0",
	})
	if err != nil {
		t.Fatalf("marshal ed25519 update payload: %v", err)
	}

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/updates", string(payload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish ed25519 update status %d: %s", rec.Code, rec.Body.String())
	}
	var saved UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatalf("decode saved ed25519 update: %v", err)
	}
	if !strings.HasPrefix(saved.Signature, "ed25519:") || saved.SignaturePublicKey == "" || saved.SignatureKeyFingerprint == "" {
		t.Fatalf("expected ed25519 signature metadata: %+v", saved)
	}
	if !updatePackageVerified(saved) {
		t.Fatalf("saved ed25519 update should verify: %+v", saved)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/updates/"+strconv.FormatInt(saved.ID, 10)+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("download ed25519 update status %d: %s", rec.Code, rec.Body.String())
	}
	var download UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode ed25519 update download: %v", err)
	}
	if !download.Verified || download.Package.SignaturePublicKey == "" || download.Manifest["signaturePublicKey"] == "" {
		t.Fatalf("download should carry ed25519 verification metadata: %+v", download)
	}
}

func TestOfflineLicensePackageImportAndVerify(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate license key: %v", err)
	}
	pkg, err := signLicenseForTest(LicensePackage{
		LicenseID: "LIC-TEST-2026", CustomerName: "测试建材集团", Watermark: "CBMP-TEST-LICENSE",
		ExpiresAt: "2027-12-31", Edition: "Enterprise Offline", Modules: []string{"erp", "dispatch", "weighbridge"},
		MaxSites: 20, MaxVehicles: 5000, IssuedAt: "2026-06-18", Issuer: "CBMP License Center",
	}, privateKey)
	if err != nil {
		t.Fatalf("sign license: %v", err)
	}
	payload, _ := json.Marshal(pkg)
	rec := testRequest(t, app, token, http.MethodPost, "/api/system/license/import", string(payload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("license import status %d: %s", rec.Code, rec.Body.String())
	}
	var imported LicensePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &imported); err != nil {
		t.Fatalf("decode imported license: %v", err)
	}
	if imported.ID == 0 || imported.Status != "active" || imported.PublicKeyFingerprint == "" {
		t.Fatalf("unexpected imported license: %+v", imported)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/verify", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license verify status %d: %s", rec.Code, rec.Body.String())
	}
	var verification LicenseVerification
	if err := json.Unmarshal(rec.Body.Bytes(), &verification); err != nil {
		t.Fatalf("decode license verification: %v", err)
	}
	if !verification.Valid || verification.License.CustomerName != "测试建材集团" || verification.Fingerprint == "" {
		t.Fatalf("expected valid imported license, got %+v", verification)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/packages", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license packages status %d: %s", rec.Code, rec.Body.String())
	}
	var packages []LicensePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &packages); err != nil {
		t.Fatalf("decode license packages: %v", err)
	}
	if len(packages) != 1 || packages[0].Status != "active" {
		t.Fatalf("expected active license package history, got %+v", packages)
	}

	tampered := pkg
	tampered.MaxVehicles = 1
	payload, _ = json.Marshal(tampered)
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/license/import", string(payload))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected tampered license rejected, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLicenseIssueAndRevocationCenter(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate license key: %v", err)
	}
	issuePayload, _ := json.Marshal(map[string]interface{}{
		"customerName": "华南授权客户", "watermark": "CBMP-HN", "expiresAt": "2028-12-31",
		"edition": "Enterprise", "modules": []string{"core", "dispatch", "finance", "license"},
		"maxSites": 5, "maxVehicles": 20, "issuer": "CBMP 测试签发中心",
		"privateKey": "ed25519:" + base64.RawStdEncoding.EncodeToString(privateKey),
	})
	rec := testRequest(t, app, token, http.MethodPost, "/api/system/license/issues", string(issuePayload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("license issue status %d: %s", rec.Code, rec.Body.String())
	}
	var issued LicensePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &issued); err != nil {
		t.Fatalf("decode issued license: %v", err)
	}
	if issued.LicenseID == "" || issued.Signature == "" || issued.Status != "issued" {
		t.Fatalf("unexpected issued license: %+v", issued)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/packages", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license packages status %d: %s", rec.Code, rec.Body.String())
	}
	var issuedPackages []LicensePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &issuedPackages); err != nil {
		t.Fatalf("decode issued packages: %v", err)
	}
	if len(issuedPackages) != 1 || issuedPackages[0].ID != issued.ID || issuedPackages[0].Status != "issued" {
		t.Fatalf("expected issued package persisted for download, got %+v", issuedPackages)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/packages/"+strconv.FormatInt(issued.ID, 10)+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license package download status %d: %s", rec.Code, rec.Body.String())
	}
	var download LicensePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode license package download: %v", err)
	}
	if !download.Valid || download.Package.LicenseID != issued.LicenseID || download.FileName == "" {
		t.Fatalf("unexpected license package download: %+v", download)
	}

	renewPayload, _ := json.Marshal(map[string]interface{}{
		"expiresAt": "2029-12-31", "maxSites": 8, "maxVehicles": 30,
		"modules":    []string{"core", "dispatch", "finance", "license", "report"},
		"issuer":     "CBMP 测试续期中心",
		"privateKey": "ed25519:" + base64.RawStdEncoding.EncodeToString(privateKey),
	})
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/license/packages/"+strconv.FormatInt(issued.ID, 10)+"/renew", string(renewPayload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("license renew status %d: %s", rec.Code, rec.Body.String())
	}
	var renewed LicensePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &renewed); err != nil {
		t.Fatalf("decode renewed license: %v", err)
	}
	if renewed.Status != "issued" || renewed.ExpiresAt != "2029-12-31" || renewed.MaxVehicles != 30 || renewed.Signature == "" {
		t.Fatalf("unexpected renewed license: %+v", renewed)
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/packages/"+strconv.FormatInt(renewed.ID, 10)+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("renewed license download status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/issues", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license issues status %d: %s", rec.Code, rec.Body.String())
	}
	var issues []LicenseIssueRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatalf("decode license issues: %v", err)
	}
	if len(issues) != 2 || issues[0].LicenseID != issued.LicenseID || issues[1].LicenseID != renewed.LicenseID {
		t.Fatalf("unexpected license issues: %+v", issues)
	}

	payload, _ := json.Marshal(issued)
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/license/import", string(payload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("issued license import status %d: %s", rec.Code, rec.Body.String())
	}

	revokePayload, _ := json.Marshal(map[string]string{"licenseId": issued.LicenseID, "reason": "客户终止合同"})
	rec = testRequest(t, app, token, http.MethodPost, "/api/system/license/revoke", string(revokePayload))
	if rec.Code != http.StatusCreated {
		t.Fatalf("license revoke status %d: %s", rec.Code, rec.Body.String())
	}
	var revoked LicenseRevocation
	if err := json.Unmarshal(rec.Body.Bytes(), &revoked); err != nil {
		t.Fatalf("decode revocation: %v", err)
	}
	if revoked.LicenseID != issued.LicenseID || revoked.Status != "active" {
		t.Fatalf("unexpected revocation: %+v", revoked)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/portal", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license portal status %d: %s", rec.Code, rec.Body.String())
	}
	var portal LicensePortalOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &portal); err != nil {
		t.Fatalf("decode license portal: %v", err)
	}
	if portal.KPIs.TotalCustomers != 2 || portal.KPIs.TotalPackages != 3 || portal.KPIs.DownloadCount != 2 {
		t.Fatalf("unexpected license portal KPIs: %+v", portal.KPIs)
	}
	if portal.KPIs.RevokedPackages != 2 || portal.KPIs.RiskLevel != "revoked" {
		t.Fatalf("expected revoked portal risk, got %+v", portal.KPIs)
	}
	if len(portal.Customers) != 2 || len(portal.RecentEvents) == 0 || len(portal.RequiredModules) == 0 {
		t.Fatalf("expected portal customers, events and module baseline, got %+v", portal)
	}
	foundRevoked := false
	for _, customer := range portal.Customers {
		if customer.LicenseID == issued.LicenseID && customer.RiskLevel == "revoked" && customer.LatestDownloadAt != "" {
			foundRevoked = true
		}
	}
	if !foundRevoked {
		t.Fatalf("expected revoked customer with download trail, got %+v", portal.Customers)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/license/verify", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("license verify status %d: %s", rec.Code, rec.Body.String())
	}
	var verification LicenseVerification
	if err := json.Unmarshal(rec.Body.Bytes(), &verification); err != nil {
		t.Fatalf("decode license verification: %v", err)
	}
	if verification.Valid || !strings.Contains(verification.Reason, "吊销") {
		t.Fatalf("expected revoked license invalid, got %+v", verification)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/license/import", string(payload))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected revoked license import rejected, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPluginSandboxRunAndAuditTrail(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/system/plugins/settlement-tax-cn/run", `{"action":"calculate_tax","permission":"statement.calculate","input":{"amount":1000,"taxRate":0.13}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("plugin run status %d: %s", rec.Code, rec.Body.String())
	}
	var run PluginRun
	if err := json.Unmarshal(rec.Body.Bytes(), &run); err != nil {
		t.Fatalf("decode plugin run: %v", err)
	}
	if run.Status != "succeeded" || run.Runtime != "wasm" || !strings.Contains(run.Output, `"taxAmount":130`) {
		t.Fatalf("unexpected plugin run: %+v", run)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/plugins/settlement-tax-cn/run", `{"permission":"finance.write","input":{"amount":1000}}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected plugin permission denied, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/plugins/adapter-scale-standard/run", `{"permission":"scale.ticket.create","input":{"grossWeight":10,"tareWeight":20}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed plugin run should still be recorded, got %d: %s", rec.Code, rec.Body.String())
	}
	var failed PluginRun
	if err := json.Unmarshal(rec.Body.Bytes(), &failed); err != nil {
		t.Fatalf("decode failed plugin run: %v", err)
	}
	if failed.Status != "failed" || failed.Error == "" {
		t.Fatalf("expected failed plugin run with error: %+v", failed)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/plugins/runs", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("plugin runs status %d: %s", rec.Code, rec.Body.String())
	}
	var runs []PluginRun
	if err := json.Unmarshal(rec.Body.Bytes(), &runs); err != nil {
		t.Fatalf("decode plugin runs: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected succeeded and failed plugin runs, got %+v", runs)
	}
}

func TestGatewayRouteCanaryDrainAndNginxRender(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/system/gateway", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("gateway overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview gatewayOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode gateway overview: %v", err)
	}
	if len(overview.Routes) < 2 || !strings.Contains(overview.NginxConfig, "split_clients") {
		t.Fatalf("unexpected gateway overview: %+v", overview)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/gateway/routes/1/canary", `{"canaryPercent":25,"canaryUpstream":"cbmp-appliance-v2:8088"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gateway canary status %d: %s", rec.Code, rec.Body.String())
	}
	var route GatewayRoute
	if err := json.Unmarshal(rec.Body.Bytes(), &route); err != nil {
		t.Fatalf("decode gateway route: %v", err)
	}
	if route.CanaryPercent != 25 || route.CanaryUpstream != "cbmp-appliance-v2:8088" {
		t.Fatalf("unexpected canary route: %+v", route)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/gateway/routes/1/drain", `{"enabled":true,"durationMs":600000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gateway drain status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &route); err != nil {
		t.Fatalf("decode drained route: %v", err)
	}
	if !route.DrainEnabled || route.DrainUntil == "" {
		t.Fatalf("expected drain route: %+v", route)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/gateway/reload-plan", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("gateway reload plan status %d: %s", rec.Code, rec.Body.String())
	}
	var plan GatewayReloadPlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode gateway reload plan: %v", err)
	}
	if !plan.Valid || !plan.ReloadRequired || plan.DrainingRoutes == 0 || !strings.HasPrefix(plan.ConfigHash, "sha256:") {
		t.Fatalf("unexpected gateway reload plan: %+v", plan)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/gateway/reload", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("gateway reload status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("decode gateway reload response: %v", err)
	}
	if !plan.Valid || plan.ReloadRequired || plan.ReloadedAt == "" || plan.LastReloadAt == "" {
		t.Fatalf("expected reload to mark plan synced, got %+v", plan)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/gateway/nginx", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("gateway nginx status %d: %s", rec.Code, rec.Body.String())
	}
	var rendered struct {
		NginxConfig string `json:"nginxConfig"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rendered); err != nil {
		t.Fatalf("decode gateway nginx: %v", err)
	}
	if !strings.Contains(rendered.NginxConfig, "25% cbmp_api_canary") || !strings.Contains(rendered.NginxConfig, "X-CBMP-Drain true") {
		t.Fatalf("rendered nginx missing canary/drain: %s", rendered.NginxConfig)
	}
}

func TestOIDCSSOStartCallbackAndProvision(t *testing.T) {
	app := newTestHTTPApp(t)

	rec := testRequest(t, app, "", http.MethodGet, "/api/auth/sso/providers", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("sso providers status %d: %s", rec.Code, rec.Body.String())
	}
	var providers []OIDCProvider
	if err := json.Unmarshal(rec.Body.Bytes(), &providers); err != nil {
		t.Fatalf("decode providers: %v", err)
	}
	if len(providers) == 0 || providers[0].ClientSecret != "" {
		t.Fatalf("expected public active provider without secret: %+v", providers)
	}

	rec = testRequest(t, app, "", http.MethodPost, "/api/auth/sso/enterprise-demo/start", `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("sso start status %d: %s", rec.Code, rec.Body.String())
	}
	var start oidcStartResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &start); err != nil {
		t.Fatalf("decode sso start: %v", err)
	}
	if start.State == "" || start.Nonce == "" || !strings.Contains(start.AuthorizationURL, "response_type=code") {
		t.Fatalf("unexpected sso start response: %+v", start)
	}

	idToken := signOIDCTestToken(t, "cbmp-oidc-demo-secret", map[string]interface{}{
		"iss": "https://idp.example.com/cbmp", "aud": "cbmp-desktop",
		"exp": time.Now().Add(5 * time.Minute).Unix(), "nonce": start.Nonce,
		"preferred_username": "sso.manager", "name": "SSO 经理",
	})
	body, _ := json.Marshal(map[string]string{"state": start.State, "idToken": idToken})
	rec = testRequest(t, app, "", http.MethodPost, "/api/auth/sso/enterprise-demo/callback", string(body))
	if rec.Code != http.StatusOK {
		t.Fatalf("sso callback status %d: %s", rec.Code, rec.Body.String())
	}
	var session Session
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode sso session: %v", err)
	}
	if session.Token == "" || session.User.Username != "sso.manager" || session.User.RoleCode != "boss" || session.User.PasswordHash != "" {
		t.Fatalf("unexpected sso session: %+v", session)
	}

	rec = testRequest(t, app, session.Token, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("security status %d: %s", rec.Code, rec.Body.String())
	}
	var security struct {
		Users        []User         `json:"users"`
		SSOProviders []OIDCProvider `json:"ssoProviders"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode security: %v", err)
	}
	foundUser := false
	for _, user := range security.Users {
		if user.Username == "sso.manager" && user.PasswordHash == "" {
			foundUser = true
		}
	}
	if !foundUser || len(security.SSOProviders) == 0 || security.SSOProviders[0].ClientSecret != "" || security.SSOProviders[0].LastLoginAt == "" {
		t.Fatalf("expected provisioned user and public provider state: %+v", security)
	}

	rec = testRequest(t, app, "", http.MethodPost, "/api/auth/sso/enterprise-demo/start", `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("second sso start status %d: %s", rec.Code, rec.Body.String())
	}
	var deniedStart oidcStartResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &deniedStart)
	tampered := signOIDCTestToken(t, "wrong-secret", map[string]interface{}{
		"iss": "https://idp.example.com/cbmp", "aud": "cbmp-desktop",
		"exp": time.Now().Add(5 * time.Minute).Unix(), "nonce": deniedStart.Nonce,
		"preferred_username": "sso.manager",
	})
	body, _ = json.Marshal(map[string]string{"state": deniedStart.State, "idToken": tampered})
	rec = testRequest(t, app, "", http.MethodPost, "/api/auth/sso/enterprise-demo/callback", string(body))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected tampered id_token rejected, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSCIMProviderProvisioningAndSecurityReport(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/scim/providers", `{"name":"测试 SCIM","code":"test-scim","bearerToken":"test-scim-token","companyId":1,"siteId":0,"defaultRoleCode":"dispatcher","status":"enabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create scim provider status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantId"`) {
		t.Fatalf("scim provider response should not expose tenant boundary: %s", rec.Body.String())
	}
	var provider SCIMProvider
	if err := json.Unmarshal(rec.Body.Bytes(), &provider); err != nil {
		t.Fatalf("decode scim provider: %v", err)
	}
	if provider.ID == 0 || provider.BearerToken != "" || provider.Code != "test-scim" {
		t.Fatalf("expected public scim provider without token: %+v", provider)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("security status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantId"`) {
		t.Fatalf("security response should not expose tenant boundary: %s", rec.Body.String())
	}
	var security struct {
		Users         []User                  `json:"users"`
		SCIMProviders []SCIMProvider          `json:"scimProviders"`
		SCIMEvents    []SCIMProvisioningEvent `json:"scimEvents"`
		Report        SecurityReport          `json:"report"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode security: %v", err)
	}
	foundPublicProvider := false
	for _, item := range security.SCIMProviders {
		if item.Code == "test-scim" {
			foundPublicProvider = item.BearerToken == ""
		}
	}
	if !foundPublicProvider {
		t.Fatalf("expected public scim provider in security payload: %+v", security.SCIMProviders)
	}

	scimBody := `{"userName":"scim.dispatcher","displayName":"SCIM 调度员","active":true,"roles":[{"value":"dispatcher"}]}`
	rec = testRequest(t, app, "", http.MethodPost, "/api/scim/v2/Users", scimBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("scim without token should be unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, "test-scim-token", http.MethodPost, "/api/scim/v2/Users", scimBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("scim create user status %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID       string `json:"id"`
		UserName string `json:"userName"`
		Active   bool   `json:"active"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode scim created user: %v", err)
	}
	if created.ID == "" || created.UserName != "scim.dispatcher" || !created.Active {
		t.Fatalf("unexpected scim created user: %+v", created)
	}

	rec = testRequest(t, app, "test-scim-token", http.MethodPatch, "/api/scim/v2/Users/"+created.ID, `{"Operations":[{"op":"replace","path":"active","value":false}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("scim patch user status %d: %s", rec.Code, rec.Body.String())
	}
	var patched struct {
		Active bool `json:"active"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("decode scim patched user: %v", err)
	}
	if patched.Active {
		t.Fatalf("expected scim user deactivated: %+v", patched)
	}

	rec = testRequest(t, app, "test-scim-token", http.MethodGet, "/api/scim/v2/Users", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "scim.dispatcher") {
		t.Fatalf("scim list status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("security after scim status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode security after scim: %v", err)
	}
	foundUser := false
	for _, user := range security.Users {
		if user.Username == "scim.dispatcher" {
			foundUser = user.Status == "disabled" && user.PasswordHash == ""
		}
	}
	if !foundUser || len(security.SCIMEvents) < 2 || security.Report.SCIMProviders == 0 || security.Report.SCIMEventsLast24h == 0 {
		t.Fatalf("expected scim user, events and report counters: %+v", security)
	}
}

func TestProductOpsOverviewInstancesAlertsAndUpdates(t *testing.T) {
	t.Skip("product operations moved to OperationsPlatform; ERP rejects /api/product-ops/*")
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview ProductOpsOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode product ops overview: %v", err)
	}
	if overview.KPIs.Customers == 0 || len(overview.Instances) == 0 || len(overview.Alerts) == 0 || len(overview.RenewalTasks) == 0 {
		t.Fatalf("expected seeded product ops instances and alerts: %+v", overview.KPIs)
	}
	if overview.KPIs.OpenRenewals == 0 {
		t.Fatalf("expected seeded renewal tasks: %+v", overview.KPIs)
	}
	if overview.KPIs.ClientUpdatePackages == 0 || overview.KPIs.ServerUpdatePackages == 0 {
		t.Fatalf("expected client/server update packages: %+v", overview.KPIs)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/updates", `{"version":"1.0.2","component":"client","channel":"stable","status":"available","checksum":"sha256:client-102","signature":"sig:client-102","rollbackVersion":"1.0.1","fileName":"cbmp-client-1.0.2.json","sizeBytes":2048,"remark":"客户端运营台修复包"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish client update status %d: %s", rec.Code, rec.Body.String())
	}
	var clientUpdate UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &clientUpdate); err != nil {
		t.Fatalf("decode published client update: %v", err)
	}
	if clientUpdate.ID == 0 || clientUpdate.Component != "client" || clientUpdate.PublishedBy == "" || clientUpdate.PublishedAt == "" {
		t.Fatalf("unexpected published client update: %+v", clientUpdate)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/updates/"+strconv.FormatInt(clientUpdate.ID, 10)+"/download", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("download published client update status %d: %s", rec.Code, rec.Body.String())
	}
	var clientDownload UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &clientDownload); err != nil {
		t.Fatalf("decode published client update download: %v", err)
	}
	if !clientDownload.Verified || clientDownload.Package.Component != "client" || clientDownload.Package.DownloadCount == 0 || clientDownload.Package.LastDownloadedAt == "" {
		t.Fatalf("unexpected client update download: %+v", clientDownload)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/updates/"+strconv.FormatInt(clientUpdate.ID, 10)+"/apply", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("apply published client update status %d: %s", rec.Code, rec.Body.String())
	}
	var appliedClient UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &appliedClient); err != nil {
		t.Fatalf("decode applied client update: %v", err)
	}
	if appliedClient.Status != "installed" || appliedClient.AppliedBy == "" || appliedClient.AppliedAt == "" {
		t.Fatalf("expected applied client update metadata: %+v", appliedClient)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/instances", `{"customerName":"西南建材客户","licenseId":"LIC-SW-2026","watermark":"CBMP-SW","edition":"Enterprise","deploymentMode":"private","clientVersion":"1.0.1","serverVersion":"1.0.1","endpoint":"https://cbmp.sw.example","status":"online","licenseExpiresAt":"2026-09-30","renewalOwner":"客户成功-周舟","renewalStage":"新签","alertLevel":"normal"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product instance status %d: %s", rec.Code, rec.Body.String())
	}
	var instance ProductInstance
	if err := json.Unmarshal(rec.Body.Bytes(), &instance); err != nil {
		t.Fatalf("decode product instance: %v", err)
	}
	if instance.ID == 0 || instance.CustomerName != "西南建材客户" {
		t.Fatalf("unexpected product instance: %+v", instance)
	}
	if instance.ProbeToken == "" || !instance.ProbeEnabled {
		t.Fatalf("expected product instance probe token: %+v", instance)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/rollouts", `{"updateId":`+strconv.FormatInt(clientUpdate.ID, 10)+`,"strategy":"gray","targetInstanceIds":[`+strconv.FormatInt(instance.ID, 10)+`],"remark":"西南客户客户端灰度批次"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create update rollout status %d: %s", rec.Code, rec.Body.String())
	}
	var rollout ProductUpdateRollout
	if err := json.Unmarshal(rec.Body.Bytes(), &rollout); err != nil {
		t.Fatalf("decode update rollout: %v", err)
	}
	if rollout.ID == 0 || rollout.UpdateID != clientUpdate.ID || rollout.TotalTargets != 1 || len(rollout.Items) != 1 || rollout.Status != "pending" {
		t.Fatalf("unexpected update rollout: %+v", rollout)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/rollouts/"+strconv.FormatInt(rollout.ID, 10)+"/execute", `{"action":"apply","instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"dryRun":true,"remark":"预检灰度执行"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dry-run update rollout execution status %d: %s", rec.Code, rec.Body.String())
	}
	var dryRunExecution ProductUpdateExecution
	if err := json.Unmarshal(rec.Body.Bytes(), &dryRunExecution); err != nil {
		t.Fatalf("decode dry-run update execution: %v", err)
	}
	if dryRunExecution.Status != "dry_run_passed" || !dryRunExecution.DryRun || len(dryRunExecution.Steps) == 0 || !dryRunExecution.ChecksumVerified {
		t.Fatalf("expected successful dry-run execution: %+v", dryRunExecution)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/rollouts/"+strconv.FormatInt(rollout.ID, 10)+"/execute", `{"action":"apply","instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"remark":"灰度执行成功"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("execute update rollout status %d: %s", rec.Code, rec.Body.String())
	}
	var execution ProductUpdateExecution
	if err := json.Unmarshal(rec.Body.Bytes(), &execution); err != nil {
		t.Fatalf("decode update execution: %v", err)
	}
	if execution.Status != "succeeded" || execution.Action != "apply" || execution.InstanceID != instance.ID || len(execution.Steps) < 5 || execution.Result == "" {
		t.Fatalf("expected completed update execution: %+v", execution)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview after rollout status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode product ops overview after rollout: %v", err)
	}
	foundRolledInstance := false
	for _, item := range overview.Instances {
		if item.ID == instance.ID {
			foundRolledInstance = item.ClientVersion == clientUpdate.Version
			break
		}
	}
	if !foundRolledInstance || len(overview.UpdateRollouts) == 0 {
		t.Fatalf("expected rollout overview and instance version update: %+v", overview)
	}
	if len(overview.UpdateExecutions) < 2 || overview.KPIs.UpdateExecutions < 2 {
		t.Fatalf("expected update execution overview and KPI: %+v", overview.KPIs)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/updates", `{"version":"1.0.3","component":"client","channel":"stable","status":"available","checksum":"sha256:client-103","signature":"sig:client-103","rollbackVersion":"1.0.2","fileName":"cbmp-client-1.0.3.json","sizeBytes":3072,"remark":"客户端端内更新器更新包"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish updater client update status %d: %s", rec.Code, rec.Body.String())
	}
	var updaterUpdate UpdatePackage
	if err := json.Unmarshal(rec.Body.Bytes(), &updaterUpdate); err != nil {
		t.Fatalf("decode updater client update: %v", err)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/rollouts", `{"updateId":`+strconv.FormatInt(updaterUpdate.ID, 10)+`,"strategy":"gray","targetInstanceIds":[`+strconv.FormatInt(instance.ID, 10)+`],"remark":"西南端内更新器更新批次"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create updater rollout status %d: %s", rec.Code, rec.Body.String())
	}
	var updaterRollout ProductUpdateRollout
	if err := json.Unmarshal(rec.Body.Bytes(), &updaterRollout); err != nil {
		t.Fatalf("decode updater rollout: %v", err)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/rollouts/"+strconv.FormatInt(updaterRollout.ID, 10)+"/system-update-tasks", `{"action":"apply","instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"remark":"下发端内更新器执行"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("queue system update task status %d: %s", rec.Code, rec.Body.String())
	}
	var updaterTask ProductSystemUpdateTask
	if err := json.Unmarshal(rec.Body.Bytes(), &updaterTask); err != nil {
		t.Fatalf("decode system update task: %v", err)
	}
	if updaterTask.ID == 0 || updaterTask.ExecutionID == 0 || updaterTask.Status != "queued" || updaterTask.DownloadURL == "" || updaterTask.UpdaterTokenHint == "" {
		t.Fatalf("unexpected queued system update task: %+v", updaterTask)
	}
	rec = testRequest(t, app, "", http.MethodGet, "/api/product-ops/system-updates/tasks", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected system update poll without token unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
	req := httptest.NewRequest(http.MethodGet, "/api/product-ops/system-updates/tasks", nil)
	req.Header.Set("X-CBMP-Updater-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("system update poll status %d: %s", rec.Code, rec.Body.String())
	}
	var poll productSystemUpdateTaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &poll); err != nil {
		t.Fatalf("decode system update poll: %v", err)
	}
	if !poll.Accepted || len(poll.Tasks) == 0 || poll.Tasks[0].TaskNo != updaterTask.TaskNo || poll.Tasks[0].Status != "assigned" {
		t.Fatalf("expected assigned system update task: %+v", poll)
	}
	req = httptest.NewRequest(http.MethodGet, updaterTask.DownloadURL, nil)
	req.Header.Set("X-CBMP-Updater-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("system update package download status %d: %s", rec.Code, rec.Body.String())
	}
	var updaterDownload UpdatePackageDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &updaterDownload); err != nil {
		t.Fatalf("decode system update download: %v", err)
	}
	if !updaterDownload.Verified || updaterDownload.Package.ID != updaterUpdate.ID || updaterDownload.Package.Checksum != updaterTask.Checksum {
		t.Fatalf("expected authorized update package download: %+v", updaterDownload)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/system-updates/tasks/"+updaterTask.TaskNo+"/report", bytes.NewBufferString(`{"status":"running","progress":45,"step":"install","message":"端内更新器正在安装"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Updater-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("system update running report status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &updaterTask); err != nil {
		t.Fatalf("decode running system update report: %v", err)
	}
	if updaterTask.Status != "running" || updaterTask.Progress < 45 || len(updaterTask.Logs) < 3 {
		t.Fatalf("expected running system update task: %+v", updaterTask)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/system-updates/tasks/"+updaterTask.TaskNo+"/report", bytes.NewBufferString(`{"status":"succeeded","progress":100,"step":"health_check","message":"端内更新器完成更新并通过健康检查","currentVersion":"1.0.3","updaterVersion":"updater-1.0.0"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Updater-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("system update success report status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &updaterTask); err != nil {
		t.Fatalf("decode succeeded system update report: %v", err)
	}
	if updaterTask.Status != "succeeded" || updaterTask.Progress != 100 || updaterTask.CompletedAt == "" {
		t.Fatalf("expected succeeded system update task: %+v", updaterTask)
	}
	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview after updater status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode product ops overview after updater: %v", err)
	}
	foundUpdaterUpdatedInstance := false
	for _, item := range overview.Instances {
		if item.ID == instance.ID {
			foundUpdaterUpdatedInstance = item.ClientVersion == updaterUpdate.Version
			break
		}
	}
	if !foundUpdaterUpdatedInstance || len(overview.SystemUpdateTasks) == 0 || overview.KPIs.SystemUpdateTasks == 0 {
		t.Fatalf("expected system update overview and version sync: %+v", overview.KPIs)
	}

	probeBody := `{"component":"server","clientVersion":"1.0.1","serverVersion":"1.0.1","status":"critical","cpuPercent":92,"memoryPercent":77,"diskPercent":91,"queueBacklog":1400,"errorCount":12,"message":"服务端队列积压"}`
	rec = testRequest(t, app, "", http.MethodPost, "/api/product-ops/probes/report", probeBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected probe without token unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/probes/report", bytes.NewBufferString(probeBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Probe-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("probe report status %d: %s", rec.Code, rec.Body.String())
	}
	var probe productProbeReportResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &probe); err != nil {
		t.Fatalf("decode probe report: %v", err)
	}
	if !probe.Accepted || probe.Report.ID == 0 || !probe.Report.AlertRaised || probe.Alert == nil || probe.Instance.Status != "degraded" {
		t.Fatalf("expected probe report to raise alert and degrade instance: %+v", probe)
	}

	recoverBody := `{"component":"server","clientVersion":"1.0.3","serverVersion":"1.0.1","status":"healthy","cpuPercent":42,"memoryPercent":55,"diskPercent":61,"queueBacklog":10,"errorCount":0,"message":"指标恢复正常"}`
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/probes/report", bytes.NewBufferString(recoverBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Probe-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("probe recovery status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &probe); err != nil {
		t.Fatalf("decode probe recovery: %v", err)
	}
	if probe.Report.AlertRaised || probe.Instance.Status != "online" || probe.Instance.HealthStatus != "healthy" {
		t.Fatalf("expected probe recovery to restore instance: %+v", probe)
	}

	telemetryBody := `{"source":"trace","component":"server","severity":"error","eventType":"http_500","traceId":"trace-sw-500","endpoint":"/api/dispatch","durationMs":4200,"statusCode":500,"errorMessage":"调度接口返回 500","message":"客户现场链路追踪异常"}`
	rec = testRequest(t, app, "", http.MethodPost, "/api/product-ops/telemetry/report", telemetryBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected telemetry without token unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/telemetry/report", bytes.NewBufferString(telemetryBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Probe-Token", instance.ProbeToken)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("telemetry report status %d: %s", rec.Code, rec.Body.String())
	}
	var telemetry productTelemetryReportResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &telemetry); err != nil {
		t.Fatalf("decode telemetry report: %v", err)
	}
	if !telemetry.Accepted || telemetry.Event.ID == 0 || !telemetry.Event.AlertRaised || telemetry.Alert == nil || telemetry.Alert.Source != "telemetry" || telemetry.Instance.Status != "degraded" {
		t.Fatalf("expected telemetry event to raise alert and degrade instance: %+v", telemetry)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/monitoring/integrations", `{"name":"测试 Prometheus 接入","code":"test-prometheus","provider":"prometheus-test","token":"mon-test-prometheus","status":"active","remark":"HTTP webhook"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save monitoring integration status %d: %s", rec.Code, rec.Body.String())
	}
	var monitoringIntegration ProductMonitoringIntegration
	if err := json.Unmarshal(rec.Body.Bytes(), &monitoringIntegration); err != nil {
		t.Fatalf("decode monitoring integration: %v", err)
	}
	if monitoringIntegration.ID == 0 || monitoringIntegration.Token == "" || monitoringIntegration.Status != "active" {
		t.Fatalf("unexpected monitoring integration: %+v", monitoringIntegration)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/monitoring/rules", `{"name":"测试服务端 CPU 告警","source":"prometheus-test","component":"server","metric":"cpu_percent","operator":">=","threshold":88,"severity":"critical","status":"active","notifyChannels":["sse","webhook"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save product alert rule status %d: %s", rec.Code, rec.Body.String())
	}
	var alertRule ProductAlertRule
	if err := json.Unmarshal(rec.Body.Bytes(), &alertRule); err != nil {
		t.Fatalf("decode product alert rule: %v", err)
	}
	if alertRule.ID == 0 || alertRule.RuleNo == "" || alertRule.Metric != "cpu_percent" {
		t.Fatalf("unexpected product alert rule: %+v", alertRule)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/policies", `{"name":"测试监控聚合抑制","source":"monitoring","component":"server","metric":"cpu_percent","severity":"warning","aggregateWindowMinutes":60,"suppressMinutes":20,"escalateAfterMinutes":0,"escalateTo":"on_call","notifyChannels":["sse","webhook"],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save alert policy status %d: %s", rec.Code, rec.Body.String())
	}
	var alertPolicy ProductAlertPolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &alertPolicy); err != nil {
		t.Fatalf("decode alert policy: %v", err)
	}
	if alertPolicy.ID == 0 || alertPolicy.PolicyNo == "" || alertPolicy.AggregateWindowMinutes != 60 {
		t.Fatalf("unexpected alert policy: %+v", alertPolicy)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/channels", `{"name":"测试短信通道","code":"sms","type":"sms","endpoint":"mock://fail","status":"active","retryLimit":3,"timeoutSeconds":1,"remark":"先模拟失败，后续重试成功"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save alert channel status %d: %s", rec.Code, rec.Body.String())
	}
	var alertChannel ProductAlertChannel
	if err := json.Unmarshal(rec.Body.Bytes(), &alertChannel); err != nil {
		t.Fatalf("decode alert channel: %v", err)
	}
	if alertChannel.ID == 0 || alertChannel.ChannelNo == "" || alertChannel.Type != "sms" {
		t.Fatalf("unexpected alert channel: %+v", alertChannel)
	}

	monitoringBody := `{"integrationCode":"test-prometheus","watermark":"` + instance.Watermark + `","provider":"prometheus-test","source":"prometheus-test","component":"server","metric":"cpu_percent","value":96.5,"severity":"warning","status":"firing","title":"Prometheus CPU 告警","message":"CPU 96.5%"}`
	rec = testRequest(t, app, "", http.MethodPost, "/api/product-ops/monitoring/report", monitoringBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected monitoring without token unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/monitoring/report", bytes.NewBufferString(monitoringBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Monitoring-Token", monitoringIntegration.Token)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("monitoring report status %d: %s", rec.Code, rec.Body.String())
	}
	var monitoring productMonitoringReportResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &monitoring); err != nil {
		t.Fatalf("decode monitoring report: %v", err)
	}
	if !monitoring.Accepted || monitoring.Event.ID == 0 || !monitoring.Event.AlertRaised || monitoring.Event.MatchedRuleNo != alertRule.RuleNo || monitoring.Alert == nil || monitoring.Alert.Source != "monitoring" {
		t.Fatalf("expected monitoring event to match rule and raise alert: %+v", monitoring)
	}
	firstMonitoringAlertID := monitoring.Alert.ID
	req = httptest.NewRequest(http.MethodPost, "/api/product-ops/monitoring/report", bytes.NewBufferString(monitoringBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Monitoring-Token", monitoringIntegration.Token)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("second monitoring report status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &monitoring); err != nil {
		t.Fatalf("decode second monitoring report: %v", err)
	}
	if monitoring.Alert == nil || monitoring.Alert.ID != firstMonitoringAlertID || monitoring.Alert.EventCount < 2 || monitoring.Alert.SuppressedUntil == "" || monitoring.Alert.PolicyNo != alertPolicy.PolicyNo {
		t.Fatalf("expected monitoring alert to aggregate and suppress: %+v", monitoring.Alert)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/policies", `{"name":"测试服务端升级策略","source":"server","component":"server","metric":"all","severity":"warning","aggregateWindowMinutes":60,"suppressMinutes":0,"escalateAfterMinutes":1,"escalateTo":"duty_manager","notifyChannels":["sse"],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save server alert policy status %d: %s", rec.Code, rec.Body.String())
	}
	var serverPolicy ProductAlertPolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &serverPolicy); err != nil {
		t.Fatalf("decode server alert policy: %v", err)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts", `{"instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"severity":"warning","source":"server","title":"服务端长时间未恢复","message":"客户服务端持续异常","firstSeenAt":"2026-01-01 00:00:00"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create escalated system alert status %d: %s", rec.Code, rec.Body.String())
	}
	var escalatedAlert SystemAlert
	if err := json.Unmarshal(rec.Body.Bytes(), &escalatedAlert); err != nil {
		t.Fatalf("decode escalated alert: %v", err)
	}
	if escalatedAlert.EscalationLevel != "duty_manager" || escalatedAlert.EscalatedAt == "" || escalatedAlert.PolicyNo != serverPolicy.PolicyNo {
		t.Fatalf("expected alert governance to escalate manual alert: %+v", escalatedAlert)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/policies", `{"name":"测试短信失败重试策略","source":"client","component":"client","metric":"all","severity":"warning","aggregateWindowMinutes":60,"suppressMinutes":0,"escalateAfterMinutes":0,"escalateTo":"sms_on_call","notifyChannels":["sms"],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save sms alert policy status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts", `{"instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"severity":"warning","source":"client","title":"客户端同步失败需短信通知","message":"客户客户端同步持续失败"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create sms alert status %d: %s", rec.Code, rec.Body.String())
	}
	var smsAlert SystemAlert
	if err := json.Unmarshal(rec.Body.Bytes(), &smsAlert); err != nil {
		t.Fatalf("decode sms alert: %v", err)
	}
	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview after sms alert status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode product ops overview after sms alert: %v", err)
	}
	var failedNotification ProductAlertNotification
	for _, item := range overview.AlertNotifications {
		if item.AlertNo == smsAlert.AlertNo && item.Channel == "sms" {
			failedNotification = item
			break
		}
	}
	if failedNotification.ID == 0 || failedNotification.Status != "failed" || failedNotification.Error == "" || failedNotification.NextRetryAt == "" {
		t.Fatalf("expected failed sms notification with retry: %+v", failedNotification)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/channels", `{"id":`+strconv.FormatInt(alertChannel.ID, 10)+`,"name":"测试短信通道","code":"sms","type":"sms","endpoint":"mock://success","status":"active","retryLimit":3,"timeoutSeconds":1,"remark":"修复后重试成功"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("fix alert channel status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/notifications/"+strconv.FormatInt(failedNotification.ID, 10)+"/retry", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("retry alert notification status %d: %s", rec.Code, rec.Body.String())
	}
	var retriedNotification ProductAlertNotification
	if err := json.Unmarshal(rec.Body.Bytes(), &retriedNotification); err != nil {
		t.Fatalf("decode retried notification: %v", err)
	}
	if retriedNotification.Status != "delivered" || retriedNotification.DeliveredAt == "" || retriedNotification.AttemptCount < 2 {
		t.Fatalf("expected delivered retried notification: %+v", retriedNotification)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals", `{"instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"stage":"合同确认","status":"open","owner":"客户成功-周舟","amount":66000,"currency":"CNY","dueDate":"2026-09-30","nextFollowAt":"2026-06-21 10:00:00","riskLevel":"warning","remark":"测试续费任务"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal task status %d: %s", rec.Code, rec.Body.String())
	}
	var renewal ProductRenewalTask
	if err := json.Unmarshal(rec.Body.Bytes(), &renewal); err != nil {
		t.Fatalf("decode renewal task: %v", err)
	}
	if renewal.ID == 0 || renewal.CustomerName != "西南建材客户" || renewal.TaskNo == "" || renewal.Status != "open" {
		t.Fatalf("unexpected renewal task: %+v", renewal)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/quote", `{"amount":66000,"currency":"CNY","modules":["license","update","support"],"newExpiresAt":"2027-09-30","remark":"续费报价单"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal quote status %d: %s", rec.Code, rec.Body.String())
	}
	var quote ProductRenewalQuote
	if err := json.Unmarshal(rec.Body.Bytes(), &quote); err != nil {
		t.Fatalf("decode renewal quote: %v", err)
	}
	if quote.ID == 0 || quote.TaskID != renewal.ID || quote.Status != "sent" || quote.Amount != 66000 {
		t.Fatalf("unexpected renewal quote: %+v", quote)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/approval", `{"action":"submit","quoteId":`+strconv.FormatInt(quote.ID, 10)+`,"approvalType":"quote","amount":66000,"currency":"CNY","currentRole":"boss","comment":"续费报价审批"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("submit renewal approval status %d: %s", rec.Code, rec.Body.String())
	}
	var renewalApproval ProductRenewalApproval
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalApproval); err != nil {
		t.Fatalf("decode renewal approval: %v", err)
	}
	if renewalApproval.ID == 0 || renewalApproval.Status != "pending" || renewalApproval.QuoteID != quote.ID {
		t.Fatalf("unexpected renewal approval: %+v", renewalApproval)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/approval", `{"action":"approve","approvalId":`+strconv.FormatInt(renewalApproval.ID, 10)+`,"comment":"审批通过"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("approve renewal approval status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalApproval); err != nil {
		t.Fatalf("decode approved renewal approval: %v", err)
	}
	if renewalApproval.Status != "approved" || renewalApproval.ApprovedBy == "" || renewalApproval.ApprovedAt == "" {
		t.Fatalf("expected approved renewal approval: %+v", renewalApproval)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/contract", `{"quoteId":`+strconv.FormatInt(quote.ID, 10)+`,"remark":"客户确认合同"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal contract status %d: %s", rec.Code, rec.Body.String())
	}
	var renewalContract ProductRenewalContract
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalContract); err != nil {
		t.Fatalf("decode renewal contract: %v", err)
	}
	if renewalContract.ID == 0 || renewalContract.QuoteID != quote.ID || renewalContract.Status != "signed" || renewalContract.Amount != 66000 {
		t.Fatalf("unexpected renewal contract: %+v", renewalContract)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/esign", `{"action":"send","contractId":`+strconv.FormatInt(renewalContract.ID, 10)+`,"signer":"客户授权代表","phone":"13800019999","channel":"local_esign","remark":"发送续费电子签"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("send renewal e-sign status %d: %s", rec.Code, rec.Body.String())
	}
	var renewalESign ProductRenewalESign
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalESign); err != nil {
		t.Fatalf("decode renewal e-sign: %v", err)
	}
	if renewalESign.ID == 0 || renewalESign.Status != "sent" || renewalESign.LinkURL == "" {
		t.Fatalf("unexpected renewal e-sign: %+v", renewalESign)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/esign", `{"action":"complete","signId":`+strconv.FormatInt(renewalESign.ID, 10)+`,"contractId":`+strconv.FormatInt(renewalContract.ID, 10)+`,"signature":"客户授权代表 电子签名","remark":"客户已完成电子签"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("complete renewal e-sign status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalESign); err != nil {
		t.Fatalf("decode completed renewal e-sign: %v", err)
	}
	if renewalESign.Status != "signed" || renewalESign.SignedAt == "" || renewalESign.Signature == "" {
		t.Fatalf("expected signed renewal e-sign: %+v", renewalESign)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/payment", `{"contractId":`+strconv.FormatInt(renewalContract.ID, 10)+`,"amount":66000,"currency":"CNY","method":"bank","remark":"续费全额回款"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal payment status %d: %s", rec.Code, rec.Body.String())
	}
	var renewalPayment ProductRenewalPayment
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalPayment); err != nil {
		t.Fatalf("decode renewal payment: %v", err)
	}
	if renewalPayment.ID == 0 || renewalPayment.ContractID != renewalContract.ID || renewalPayment.Status != "paid" || renewalPayment.Amount != 66000 {
		t.Fatalf("unexpected renewal payment: %+v", renewalPayment)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/integrations", `{"name":"测试税控网关","code":"tax_gateway","provider":"tax","scenario":"tax","endpoint":"mock://fail","status":"active","retryLimit":3,"timeoutSeconds":1,"remark":"先模拟税控失败"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save renewal tax integration status %d: %s", rec.Code, rec.Body.String())
	}
	var renewalIntegration ProductRenewalIntegration
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalIntegration); err != nil {
		t.Fatalf("decode renewal integration: %v", err)
	}
	if renewalIntegration.ID == 0 || renewalIntegration.Code != "tax_gateway" || renewalIntegration.Endpoint != "mock://fail" {
		t.Fatalf("unexpected renewal integration: %+v", renewalIntegration)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/invoice", `{"contractId":`+strconv.FormatInt(renewalContract.ID, 10)+`,"paymentId":`+strconv.FormatInt(renewalPayment.ID, 10)+`,"invoiceType":"blue_e_invoice","taxRate":0.06,"remark":"续费电子发票"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var renewalInvoice ProductRenewalInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &renewalInvoice); err != nil {
		t.Fatalf("decode renewal invoice: %v", err)
	}
	if renewalInvoice.ID == 0 || renewalInvoice.PaymentID != renewalPayment.ID || renewalInvoice.Status != "issued" || renewalInvoice.TaxStatus != "failed" || renewalInvoice.FileURL == "" {
		t.Fatalf("unexpected renewal invoice: %+v", renewalInvoice)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview after renewal commercial status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode product ops overview after renewal commercial: %v", err)
	}
	if len(overview.RenewalQuotes) == 0 || len(overview.RenewalContracts) == 0 || len(overview.RenewalPayments) == 0 || overview.KPIs.PaidRenewalAmount < 66000 {
		t.Fatalf("expected renewal commercial overview: %+v", overview.KPIs)
	}
	if len(overview.RenewalApprovals) == 0 || len(overview.RenewalESigns) == 0 || len(overview.RenewalInvoices) == 0 || overview.KPIs.IssuedRenewalInvoices == 0 {
		t.Fatalf("expected renewal enterprise overview: %+v", overview.KPIs)
	}
	var failedRenewalSync ProductRenewalSyncRecord
	for _, item := range overview.RenewalSyncRecords {
		if item.ResourceType == "invoice" && item.ResourceID == renewalInvoice.ID {
			failedRenewalSync = item
			break
		}
	}
	if failedRenewalSync.ID == 0 || failedRenewalSync.Status != "failed" || failedRenewalSync.Error == "" || failedRenewalSync.NextRetryAt == "" || overview.KPIs.FailedRenewalSyncRecords == 0 {
		t.Fatalf("expected failed renewal tax sync: record=%+v kpis=%+v", failedRenewalSync, overview.KPIs)
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/integrations", `{"id":`+strconv.FormatInt(renewalIntegration.ID, 10)+`,"name":"测试税控网关","code":"tax_gateway","provider":"tax","scenario":"tax","endpoint":"mock://success","status":"active","retryLimit":3,"timeoutSeconds":1,"remark":"修复后重试成功"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("fix renewal tax integration status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/sync-records/"+strconv.FormatInt(failedRenewalSync.ID, 10)+"/retry", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("retry renewal sync status %d: %s", rec.Code, rec.Body.String())
	}
	var retriedRenewalSync ProductRenewalSyncRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &retriedRenewalSync); err != nil {
		t.Fatalf("decode retried renewal sync: %v", err)
	}
	if retriedRenewalSync.Status != "succeeded" || retriedRenewalSync.CompletedAt == "" || retriedRenewalSync.AttemptCount < 2 {
		t.Fatalf("expected succeeded retried renewal sync: %+v", retriedRenewalSync)
	}
	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("product ops overview after renewal sync retry status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode product ops overview after renewal sync retry: %v", err)
	}
	if overview.KPIs.ActiveRenewalIntegrations == 0 || overview.KPIs.RenewalSyncRecords < 3 || overview.KPIs.FailedRenewalSyncRecords != 0 {
		t.Fatalf("expected renewal integration and sync KPI recovery: %+v", overview.KPIs)
	}
	var retriedInvoice ProductRenewalInvoice
	for _, item := range overview.RenewalInvoices {
		if item.ID == renewalInvoice.ID {
			retriedInvoice = item
			break
		}
	}
	if retriedInvoice.ID == 0 || retriedInvoice.TaxStatus != "accepted" || retriedInvoice.ExternalRequest == "" {
		t.Fatalf("expected renewal invoice tax status after retry: %+v", retriedInvoice)
	}
	if len(overview.MonitoringIntegrations) == 0 || len(overview.AlertRules) == 0 || len(overview.MonitoringEvents) == 0 || overview.KPIs.MonitoringAlerts == 0 {
		t.Fatalf("expected monitoring overview: %+v", overview.KPIs)
	}
	if len(overview.AlertPolicies) == 0 || len(overview.AlertNotifications) == 0 || overview.KPIs.ActiveAlertPolicies == 0 || overview.KPIs.AlertNotifications == 0 || overview.KPIs.SuppressedAlerts == 0 || overview.KPIs.EscalatedAlerts == 0 {
		t.Fatalf("expected alert governance overview: %+v", overview.KPIs)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/"+strconv.FormatInt(renewal.ID, 10)+"/close", `{"remark":"续费已完成"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("close renewal task status %d: %s", rec.Code, rec.Body.String())
	}
	var closedRenewal ProductRenewalTask
	if err := json.Unmarshal(rec.Body.Bytes(), &closedRenewal); err != nil {
		t.Fatalf("decode closed renewal task: %v", err)
	}
	if closedRenewal.Status != "closed" || closedRenewal.ClosedAt == "" || closedRenewal.Stage != "已续费" {
		t.Fatalf("expected closed renewal task: %+v", closedRenewal)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts", `{"instanceId":`+strconv.FormatInt(instance.ID, 10)+`,"severity":"critical","source":"server","title":"服务端心跳中断","message":"客户服务端 15 分钟未上报心跳"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create system alert status %d: %s", rec.Code, rec.Body.String())
	}
	var alert SystemAlert
	if err := json.Unmarshal(rec.Body.Bytes(), &alert); err != nil {
		t.Fatalf("decode system alert: %v", err)
	}
	if alert.ID == 0 || alert.CustomerName != "西南建材客户" || alert.Status != "open" {
		t.Fatalf("unexpected system alert: %+v", alert)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/alerts/"+strconv.FormatInt(alert.ID, 10)+"/handle", `{"remark":"已联系客户重启服务"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("handle system alert status %d: %s", rec.Code, rec.Body.String())
	}
	var handled SystemAlert
	if err := json.Unmarshal(rec.Body.Bytes(), &handled); err != nil {
		t.Fatalf("decode handled alert: %v", err)
	}
	if handled.Status != "handled" || handled.HandledBy == "" {
		t.Fatalf("expected handled alert with actor: %+v", handled)
	}
}

func TestProductRenewalSyncCallbackUpdatesInvoice(t *testing.T) {
	t.Skip("product operations moved to OperationsPlatform; ERP rejects /api/product-ops/*")
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/integrations", `{"name":"回调税控网关","code":"tax_gateway","provider":"tax","scenario":"tax","endpoint":"mock://fail","secret":"renewal-callback-secret","status":"active","retryLimit":3,"timeoutSeconds":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("save renewal callback integration status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/product-ops/renewals/1/invoice", `{"contractId":1,"paymentId":1,"invoiceType":"blue_e_invoice","taxRate":0.06,"remark":"回调测试发票"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create renewal callback invoice status %d: %s", rec.Code, rec.Body.String())
	}
	var invoice ProductRenewalInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &invoice); err != nil {
		t.Fatalf("decode renewal callback invoice: %v", err)
	}
	if invoice.TaxStatus != "failed" {
		t.Fatalf("expected tax sync failure before callback: %+v", invoice)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("overview before renewal callback status %d: %s", rec.Code, rec.Body.String())
	}
	var overview ProductOpsOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode overview before renewal callback: %v", err)
	}
	var failedSync ProductRenewalSyncRecord
	for _, item := range overview.RenewalSyncRecords {
		if item.ResourceType == "invoice" && item.ResourceID == invoice.ID {
			failedSync = item
			break
		}
	}
	if failedSync.ID == 0 || failedSync.Status != "failed" {
		t.Fatalf("expected failed sync before callback: %+v", failedSync)
	}

	callback := []byte(`{"syncNo":"` + failedSync.SyncNo + `","status":"accepted","externalStatus":"accepted","externalRequestId":"tax-callback-001","fileUrl":"https://tax.example/renewal.pdf"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req := httptest.NewRequest(http.MethodPost, "/api/product-ops/renewals/sync-callback", bytes.NewReader(callback))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CBMP-Timestamp", timestamp)
	req.Header.Set("X-CBMP-Signature", taxGatewaySignature("renewal-callback-secret", timestamp, callback))
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("renewal sync callback status %d: %s", rec.Code, rec.Body.String())
	}
	var callbackSync ProductRenewalSyncRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &callbackSync); err != nil {
		t.Fatalf("decode renewal callback sync: %v", err)
	}
	if callbackSync.Status != "succeeded" || callbackSync.ExternalRequestID != "tax-callback-001" || callbackSync.CompletedAt == "" {
		t.Fatalf("expected callback to succeed sync record: %+v", callbackSync)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/product-ops/overview", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("overview after renewal callback status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode overview after renewal callback: %v", err)
	}
	var updatedInvoice ProductRenewalInvoice
	for _, item := range overview.RenewalInvoices {
		if item.ID == invoice.ID {
			updatedInvoice = item
			break
		}
	}
	if updatedInvoice.ID == 0 || updatedInvoice.TaxStatus != "accepted" || updatedInvoice.ExternalRequest != "tax-callback-001" {
		t.Fatalf("expected callback to update invoice: %+v", updatedInvoice)
	}
}

func signOIDCTestToken(t *testing.T, secret string, claims map[string]interface{}) string {
	t.Helper()
	encode := func(payload interface{}) string {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal jwt payload: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	header := encode(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload := encode(claims)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestRBACAndCustomerDataScope(t *testing.T) {
	app := newTestHTTPApp(t)
	customerToken := testLogin(t, app, "customer", "customer123")

	rec := testRequest(t, app, customerToken, http.MethodGet, "/api/orders", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("customer orders status %d: %s", rec.Code, rec.Body.String())
	}
	var orders []SalesOrder
	if err := json.Unmarshal(rec.Body.Bytes(), &orders); err != nil {
		t.Fatalf("decode orders: %v", err)
	}
	if len(orders) == 0 {
		t.Fatalf("expected scoped customer orders")
	}
	for _, order := range orders {
		if order.CustomerID != 1 {
			t.Fatalf("customer saw order for customer %d", order.CustomerID)
		}
	}

	rec = testRequest(t, app, customerToken, http.MethodGet, "/api/system/updates", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden system access, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, customerToken, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("customer bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	var customerBootstrap struct {
		Customers []Customer      `json:"customers"`
		Sites     []Site          `json:"sites"`
		Vehicles  []Vehicle       `json:"vehicles"`
		Drivers   []Driver        `json:"drivers"`
		Inventory []InventoryItem `json:"inventory"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &customerBootstrap); err != nil {
		t.Fatalf("decode customer bootstrap: %v", err)
	}
	if len(customerBootstrap.Customers) != 1 || customerBootstrap.Customers[0].ID != 1 {
		t.Fatalf("customer bootstrap leaked customers: %+v", customerBootstrap.Customers)
	}
	if customerBootstrap.Customers[0].Phone != "13800010001" {
		t.Fatalf("customer own phone should not be masked, got %s", customerBootstrap.Customers[0].Phone)
	}
	for _, vehicle := range customerBootstrap.Vehicles {
		if vehicle.SiteID != 1 {
			t.Fatalf("customer bootstrap leaked unrelated vehicle: %+v", vehicle)
		}
	}
	for _, driver := range customerBootstrap.Drivers {
		if driver.Phone == "13900030001" || driver.LicenseNo == "A2-440301" {
			t.Fatalf("customer bootstrap did not mask driver sensitive fields: %+v", driver)
		}
	}
	for _, item := range customerBootstrap.Inventory {
		if item.SiteID != 1 {
			t.Fatalf("customer bootstrap leaked unrelated inventory: %+v", item)
		}
	}
}

func TestDriverAndDispatcherBootstrapScopesRelatedMasterData(t *testing.T) {
	app := newTestHTTPApp(t)
	driverToken := testLogin(t, app, "driver", "driver123")

	rec := testRequest(t, app, driverToken, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("driver bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	var driverBootstrap struct {
		Customers []Customer      `json:"customers"`
		Projects  []Project       `json:"projects"`
		Vehicles  []Vehicle       `json:"vehicles"`
		Drivers   []Driver        `json:"drivers"`
		Inventory []InventoryItem `json:"inventory"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &driverBootstrap); err != nil {
		t.Fatalf("decode driver bootstrap: %v", err)
	}
	if len(driverBootstrap.Customers) != 1 || driverBootstrap.Customers[0].ID != 1 {
		t.Fatalf("driver bootstrap leaked customers: %+v", driverBootstrap.Customers)
	}
	if driverBootstrap.Customers[0].Phone != "138****0001" {
		t.Fatalf("driver bootstrap did not mask customer phone: %+v", driverBootstrap.Customers)
	}
	if len(driverBootstrap.Projects) != 1 || driverBootstrap.Projects[0].Phone != "138****0001" {
		t.Fatalf("driver bootstrap did not mask project phone: %+v", driverBootstrap.Projects)
	}
	for _, vehicle := range driverBootstrap.Vehicles {
		if vehicle.DriverID != 1 {
			t.Fatalf("driver bootstrap leaked unrelated vehicle: %+v", vehicle)
		}
	}
	if len(driverBootstrap.Drivers) != 1 || driverBootstrap.Drivers[0].ID != 1 {
		t.Fatalf("driver bootstrap leaked drivers: %+v", driverBootstrap.Drivers)
	}
	for _, item := range driverBootstrap.Inventory {
		if item.SiteID != 1 {
			t.Fatalf("driver bootstrap leaked unrelated inventory: %+v", item)
		}
	}

	dispatcherToken := testLogin(t, app, "dispatcher", "dispatch123")
	rec = testRequest(t, app, dispatcherToken, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dispatcher bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	var dispatcherBootstrap struct {
		Customers []Customer `json:"customers"`
		Sites     []Site     `json:"sites"`
		Vehicles  []Vehicle  `json:"vehicles"`
		Drivers   []Driver   `json:"drivers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatcherBootstrap); err != nil {
		t.Fatalf("decode dispatcher bootstrap: %v", err)
	}
	if len(dispatcherBootstrap.Sites) != 1 || dispatcherBootstrap.Sites[0].ID != 1 {
		t.Fatalf("dispatcher bootstrap leaked sites: %+v", dispatcherBootstrap.Sites)
	}
	for _, vehicle := range dispatcherBootstrap.Vehicles {
		if vehicle.SiteID != 1 {
			t.Fatalf("dispatcher bootstrap leaked unrelated vehicle: %+v", vehicle)
		}
	}
	if len(dispatcherBootstrap.Customers) != 1 || dispatcherBootstrap.Customers[0].ID != 1 {
		t.Fatalf("dispatcher bootstrap leaked customers: %+v", dispatcherBootstrap.Customers)
	}
	if dispatcherBootstrap.Customers[0].Phone != "13800010001" {
		t.Fatalf("dispatcher internal phone should not be masked, got %s", dispatcherBootstrap.Customers[0].Phone)
	}
	allowedDrivers := map[int64]bool{1: true, 2: true}
	for _, driver := range dispatcherBootstrap.Drivers {
		if !allowedDrivers[driver.ID] {
			t.Fatalf("dispatcher bootstrap leaked unrelated driver: %+v", driver)
		}
	}
}

func TestFieldPoliciesAreConfigurable(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")
	dispatcherToken := testLogin(t, app, "dispatcher", "dispatch123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/field-policies", `{"roleCode":"dispatcher","resource":"customers","field":"phone","mask":"phone","remark":"调度演示脱敏"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create field policy status %d: %s", rec.Code, rec.Body.String())
	}
	var policy FieldPolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &policy); err != nil {
		t.Fatalf("decode field policy: %v", err)
	}
	if policy.ID == 0 || !policy.Enabled {
		t.Fatalf("unexpected field policy: %+v", policy)
	}

	bootstrap := fetchBootstrapForMasking(t, app, dispatcherToken)
	if len(bootstrap.Customers) == 0 || bootstrap.Customers[0].Phone != "138****0001" {
		t.Fatalf("expected configurable dispatcher customer phone masking, got %+v", bootstrap.Customers)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/field-policies/"+strconv.FormatInt(policy.ID, 10)+"/toggle", `{"enabled":false}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("toggle field policy status %d: %s", rec.Code, rec.Body.String())
	}
	bootstrap = fetchBootstrapForMasking(t, app, dispatcherToken)
	if len(bootstrap.Customers) == 0 || bootstrap.Customers[0].Phone != "13800010001" {
		t.Fatalf("expected unmasked dispatcher customer phone after disabling policy, got %+v", bootstrap.Customers)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("system security status %d: %s", rec.Code, rec.Body.String())
	}
	var security struct {
		FieldPolicies []FieldPolicy `json:"fieldPolicies"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode system security: %v", err)
	}
	if !hasFieldPolicy(security.FieldPolicies, policy.ID) {
		t.Fatalf("expected system security to expose field policy, got %+v", security.FieldPolicies)
	}
}

func TestApprovalFlowsAndDictionariesAreConfigurable(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodGet, "/api/system/approval-flows", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("approval flows status %d: %s", rec.Code, rec.Body.String())
	}
	var flows []ApprovalFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &flows); err != nil {
		t.Fatalf("decode approval flows: %v", err)
	}
	if len(flows) < 4 {
		t.Fatalf("expected seeded approval flows, got %+v", flows)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/approval-flows", `{"code":"quality_exception","name":"质量异常审批","resource":"quality_inspection","steps":[{"roleCode":"quality","action":"approve"},{"roleCode":"boss","action":"approve"}],"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create approval flow status %d: %s", rec.Code, rec.Body.String())
	}
	var flow ApprovalFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &flow); err != nil {
		t.Fatalf("decode approval flow: %v", err)
	}
	if flow.ID == 0 || len(flow.Steps) != 2 || flow.Steps[0].Seq != 1 || flow.Steps[1].RoleCode != "boss" {
		t.Fatalf("unexpected approval flow: %+v", flow)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/approval-flows", `{"code":"quality_exception","name":"质量异常二级审批","resource":"quality_inspection","steps":[{"seq":1,"roleCode":"quality","action":"approve"},{"seq":2,"roleCode":"boss","action":"approve"}],"status":"draft"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("update approval flow status %d: %s", rec.Code, rec.Body.String())
	}
	var updatedFlow ApprovalFlow
	if err := json.Unmarshal(rec.Body.Bytes(), &updatedFlow); err != nil {
		t.Fatalf("decode updated approval flow: %v", err)
	}
	if updatedFlow.ID != flow.ID || updatedFlow.Name != "质量异常二级审批" || updatedFlow.Status != "draft" {
		t.Fatalf("expected approval flow upsert by code, got %+v", updatedFlow)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/approval-flows/"+strconv.FormatInt(flow.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("disable approval flow status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &updatedFlow); err != nil {
		t.Fatalf("decode disabled approval flow: %v", err)
	}
	if updatedFlow.Status != "disabled" {
		t.Fatalf("expected disabled approval flow, got %+v", updatedFlow)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/dictionaries", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dictionaries status %d: %s", rec.Code, rec.Body.String())
	}
	var dictionaries []DataDictionary
	if err := json.Unmarshal(rec.Body.Bytes(), &dictionaries); err != nil {
		t.Fatalf("decode dictionaries: %v", err)
	}
	if len(dictionaries) < 10 {
		t.Fatalf("expected seeded dictionaries, got %+v", dictionaries)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/dictionaries", `{"type":"product_line","code":"aggregate","label":"骨料","sort":6,"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dictionary status %d: %s", rec.Code, rec.Body.String())
	}
	var dictionary DataDictionary
	if err := json.Unmarshal(rec.Body.Bytes(), &dictionary); err != nil {
		t.Fatalf("decode dictionary: %v", err)
	}
	if dictionary.ID == 0 || dictionary.Type != "product_line" || dictionary.Code != "aggregate" {
		t.Fatalf("unexpected dictionary: %+v", dictionary)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/dictionaries", `{"type":"product_line","code":"aggregate","label":"砂石骨料","sort":6,"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("update dictionary status %d: %s", rec.Code, rec.Body.String())
	}
	var updatedDictionary DataDictionary
	if err := json.Unmarshal(rec.Body.Bytes(), &updatedDictionary); err != nil {
		t.Fatalf("decode updated dictionary: %v", err)
	}
	if updatedDictionary.ID != dictionary.ID || updatedDictionary.Label != "砂石骨料" {
		t.Fatalf("expected dictionary upsert by type/code, got %+v", updatedDictionary)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/dictionaries/"+strconv.FormatInt(dictionary.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("disable dictionary status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &updatedDictionary); err != nil {
		t.Fatalf("decode disabled dictionary: %v", err)
	}
	if updatedDictionary.Status != "disabled" {
		t.Fatalf("expected disabled dictionary, got %+v", updatedDictionary)
	}
}

func TestOrganizationManagementAndScopedBootstrap(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/org/tenants", `{"name":"华东建材集团","code":"east","status":"active"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("tenant management should not be exposed, status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/tenant-policies/1/toggle", `{"enabled":false}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("boundary policy management should not be exposed, status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/org/companies", `{"name":"华东骨料有限公司","code":"EAST-AGG"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create company status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantId"`) {
		t.Fatalf("company response should not expose tenant boundary: %s", rec.Body.String())
	}
	var company Company
	if err := json.Unmarshal(rec.Body.Bytes(), &company); err != nil {
		t.Fatalf("decode company: %v", err)
	}
	if company.ID == 0 || company.Status != "active" {
		t.Fatalf("unexpected company: %+v", company)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/org/departments", `{"companyId":`+strconv.FormatInt(company.ID, 10)+`,"name":"华东运营部","code":"EAST-OPS"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create department status %d: %s", rec.Code, rec.Body.String())
	}
	var department Department
	if err := json.Unmarshal(rec.Body.Bytes(), &department); err != nil {
		t.Fatalf("decode department: %v", err)
	}
	if department.CompanyID != company.ID || department.Status != "active" {
		t.Fatalf("unexpected department: %+v", department)
	}

	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/org/departments/"+strconv.FormatInt(department.ID, 10)+"/status", `{"status":"disabled"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("toggle department status %d: %s", rec.Code, rec.Body.String())
	}
	var disabled Department
	if err := json.Unmarshal(rec.Body.Bytes(), &disabled); err != nil {
		t.Fatalf("decode disabled department: %v", err)
	}
	if disabled.Status != "disabled" {
		t.Fatalf("department status not updated: %+v", disabled)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/org", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("org overview status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenants"`) {
		t.Fatalf("org overview should not expose tenant management fields: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantId"`) {
		t.Fatalf("org overview should not expose tenant boundary: %s", rec.Body.String())
	}
	var overview struct {
		Companies   []Company    `json:"companies"`
		Departments []Department `json:"departments"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode org overview: %v", err)
	}
	if len(overview.Companies) < 2 || len(overview.Departments) < 4 {
		t.Fatalf("org overview missing records: %+v", overview)
	}

	dispatcherToken := testLogin(t, app, "dispatcher", "dispatch123")
	rec = testRequest(t, app, dispatcherToken, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("dispatcher bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenants"`) {
		t.Fatalf("dispatcher bootstrap should not expose tenant management fields: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"tenantId"`) {
		t.Fatalf("dispatcher bootstrap should not expose tenant boundary: %s", rec.Body.String())
	}
	var dispatcherBootstrap struct {
		Companies   []Company    `json:"companies"`
		Departments []Department `json:"departments"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dispatcherBootstrap); err != nil {
		t.Fatalf("decode dispatcher org bootstrap: %v", err)
	}
	for _, item := range dispatcherBootstrap.Companies {
		if item.ID == company.ID {
			t.Fatalf("dispatcher saw unrelated company: %+v", dispatcherBootstrap.Companies)
		}
	}
	for _, item := range dispatcherBootstrap.Departments {
		if item.ID == department.ID {
			t.Fatalf("dispatcher saw unrelated department: %+v", dispatcherBootstrap.Departments)
		}
	}
	if len(dispatcherBootstrap.Companies) != 1 || dispatcherBootstrap.Companies[0].ID != 1 {
		t.Fatalf("dispatcher should see only site company: %+v", dispatcherBootstrap.Companies)
	}
}

func TestMFAEnrollmentAndLoginChallenge(t *testing.T) {
	app := newTestHTTPApp(t)
	adminToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, adminToken, http.MethodPost, "/api/system/mfa/users/2/enroll", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("mfa enroll status %d: %s", rec.Code, rec.Body.String())
	}
	var enrollment struct {
		User       User   `json:"user"`
		Secret     string `json:"secret"`
		OTPAuthURL string `json:"otpauthUrl"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &enrollment); err != nil {
		t.Fatalf("decode mfa enrollment: %v", err)
	}
	if enrollment.Secret == "" || enrollment.OTPAuthURL == "" || enrollment.User.MFASecret != "" {
		t.Fatalf("unexpected enrollment payload: %+v", enrollment)
	}

	code := totpCodeAt(enrollment.Secret, time.Now().Unix()/totpStepSeconds)
	rec = testRequest(t, app, adminToken, http.MethodPost, "/api/system/mfa/users/2/enable", `{"code":"`+code+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("mfa enable status %d: %s", rec.Code, rec.Body.String())
	}
	var enabled User
	if err := json.Unmarshal(rec.Body.Bytes(), &enabled); err != nil {
		t.Fatalf("decode mfa enabled user: %v", err)
	}
	if !enabled.MFAEnabled || enabled.MFASecret != "" {
		t.Fatalf("mfa enabled response leaked secret or not enabled: %+v", enabled)
	}

	rec = testRequest(t, app, "", http.MethodPost, "/api/auth/login", `{"username":"dispatcher","password":"dispatch123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mfa challenge status %d: %s", rec.Code, rec.Body.String())
	}
	var challenge struct {
		Token       string `json:"token"`
		MFARequired bool   `json:"mfaRequired"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &challenge); err != nil {
		t.Fatalf("decode mfa challenge: %v", err)
	}
	if !challenge.MFARequired || challenge.Token != "" {
		t.Fatalf("expected mfa challenge without token: %+v", challenge)
	}

	rec = testRequest(t, app, "", http.MethodPost, "/api/auth/login", `{"username":"dispatcher","password":"dispatch123","mfaCode":"`+code+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mfa login status %d: %s", rec.Code, rec.Body.String())
	}
	var session Session
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode mfa session: %v", err)
	}
	if session.Token == "" || !session.User.MFAEnabled || session.User.MFASecret != "" {
		t.Fatalf("expected mfa session without secret: %+v", session)
	}

	rec = testRequest(t, app, adminToken, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("system security status %d: %s", rec.Code, rec.Body.String())
	}
	var security struct {
		Users []User `json:"users"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode security users: %v", err)
	}
	for _, user := range security.Users {
		if user.ID == 2 {
			if !user.MFAEnabled || user.MFASecret != "" {
				t.Fatalf("security user should show mfa status without secret: %+v", user)
			}
			return
		}
	}
	t.Fatalf("dispatcher missing from security users: %+v", security.Users)
}

func TestSecurityReportSessionPolicyAndIPWhitelist(t *testing.T) {
	app := newTestHTTPApp(t)
	err := app.store.Mutate(func(data *AppData) error {
		for i := range data.SecurityPolicies {
			if data.SecurityPolicies[i].Type == "session_max_per_user" {
				data.SecurityPolicies[i].Value = "1"
			}
			if data.SecurityPolicies[i].Type == "session_timeout_minutes" {
				data.SecurityPolicies[i].Value = "60"
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("set session policy: %v", err)
	}
	firstToken := testLogin(t, app, "admin", "admin123")
	secondToken := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, firstToken, http.MethodGet, "/api/me", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected oldest session pruned, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, secondToken, http.MethodGet, "/api/system/security", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("system security status %d: %s", rec.Code, rec.Body.String())
	}
	var security struct {
		SessionPolicy SessionPolicy          `json:"sessionPolicy"`
		Sessions      []ActiveSessionSummary `json:"sessions"`
		Report        SecurityReport         `json:"report"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &security); err != nil {
		t.Fatalf("decode system security: %v", err)
	}
	if security.SessionPolicy.TimeoutMinutes != 60 || security.SessionPolicy.MaxSessionsPerUser != 1 {
		t.Fatalf("unexpected session policy: %+v", security.SessionPolicy)
	}
	if len(security.Sessions) != 1 || security.Sessions[0].Username != "admin" || security.Sessions[0].ExpiresAt == "" {
		t.Fatalf("unexpected active sessions: %+v", security.Sessions)
	}
	if security.Report.ActiveSessions != 1 || security.Report.TotalUsers == 0 || security.Report.EnabledSecurityPolicies == 0 {
		t.Fatalf("unexpected security report: %+v", security.Report)
	}

	t.Setenv("CBMP_ENFORCE_IP_WHITELIST", "1")
	allowed := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"admin123"}`))
	allowed.RemoteAddr = "127.0.0.1:1234"
	allowed.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, allowed)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected whitelist IP login allowed, got %d: %s", rec.Code, rec.Body.String())
	}
	denied := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"admin123"}`))
	denied.RemoteAddr = "10.9.8.7:1234"
	denied.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, denied)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected non-whitelist IP blocked, got %d: %s", rec.Code, rec.Body.String())
	}
}

func fetchBootstrapForMasking(t *testing.T, app *App, token string) struct {
	Customers []Customer `json:"customers"`
} {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/bootstrap", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("bootstrap status %d: %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Customers []Customer `json:"customers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	return payload
}

func hasFieldPolicy(items []FieldPolicy, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func TestDeviceKeyCanReportLocationWithDeviceScope(t *testing.T) {
	app := newTestHTTPApp(t)
	body := `{"deviceNo":"GPS1000001","plateNo":"粤B12345","longitude":113.95,"latitude":22.53,"speed":35,"direction":120,"mileage":123500,"accStatus":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/iot/vehicle/location/report", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-Key", "device-demo-key-1")
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("device location status %d: %s", rec.Code, rec.Body.String())
	}

	bad := `{"deviceNo":"GPS1000002","plateNo":"粤B12345","longitude":113.95,"latitude":22.53}`
	req = httptest.NewRequest(http.MethodPost, "/api/iot/vehicle/location/report", bytes.NewBufferString(bad))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-Key", "device-demo-key-1")
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected device mismatch forbidden, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDriverOfflineLocationBatchUpload(t *testing.T) {
	app := newTestHTTPApp(t)
	driverToken := testLogin(t, app, "driver", "driver123")

	rec := testRequest(t, app, driverToken, http.MethodPost, "/api/iot/vehicle/location/batch", `{"reports":[
		{"plateNo":"粤B12345","longitude":113.9368,"latitude":22.5420,"speed":18,"direction":135,"mileage":120001,"accStatus":1,"locationTime":"2026-06-18 12:01:00","sourceType":"driver_app_offline"},
		{"plateNo":"粤B12345","longitude":113.9380,"latitude":22.5428,"speed":20,"direction":135,"mileage":120002,"accStatus":1,"locationTime":"2026-06-18 12:02:00","sourceType":"driver_app_offline"}
	]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("driver location batch status %d: %s", rec.Code, rec.Body.String())
	}
	var result locationBatchReportResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode driver location batch: %v", err)
	}
	if result.Total != 2 || result.Accepted != 2 || result.Rejected != 0 {
		t.Fatalf("expected accepted batch, got %+v", result)
	}
	data := app.mustSnapshot()
	var foundLatest bool
	for _, latest := range data.LatestLocations {
		if latest.VehicleID == 1 && latest.LastLocationTime == "2026-06-18 12:02:00" {
			foundLatest = true
		}
	}
	if !foundLatest {
		t.Fatalf("expected latest location updated from offline batch, got %+v", data.LatestLocations)
	}

	rec = testRequest(t, app, driverToken, http.MethodPost, "/api/iot/vehicle/location/batch", `{"reports":[
		{"plateNo":"粤B22336","longitude":113.94,"latitude":22.55,"speed":18,"direction":135,"mileage":120003,"accStatus":1,"locationTime":"2026-06-18 12:03:00","sourceType":"driver_app_offline"}
	]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("driver forbidden location batch status %d: %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode forbidden driver location batch: %v", err)
	}
	if result.Accepted != 0 || result.Rejected != 1 || !strings.Contains(result.Results[0].Error, "司机无权上报该车辆位置") {
		t.Fatalf("expected cross-vehicle location rejected, got %+v", result)
	}
}

func TestSystemBackupCreateListAndRestore(t *testing.T) {
	t.Setenv("CBMP_BACKUP_DIR", t.TempDir())
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/system/backups/drills", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("empty backup drills list status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("empty backup drills should serialize as [], got %s", rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/backups/drills", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("backup drill status %d: %s", rec.Code, rec.Body.String())
	}
	var drill BackupDrill
	if err := json.Unmarshal(rec.Body.Bytes(), &drill); err != nil {
		t.Fatalf("decode backup drill: %v", err)
	}
	if err := requirePassedBackupDrill(drill); err != nil {
		t.Fatalf("invalid backup drill: %v %+v", err, drill)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/backups/drills", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("backup drills list status %d: %s", rec.Code, rec.Body.String())
	}
	var drills []BackupDrill
	if err := json.Unmarshal(rec.Body.Bytes(), &drills); err != nil {
		t.Fatalf("decode backup drills: %v", err)
	}
	if len(drills) != 1 || drills[0].Status != "passed" {
		t.Fatalf("unexpected backup drills: %+v", drills)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/backups", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("backup create status %d: %s", rec.Code, rec.Body.String())
	}
	var backup BackupInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &backup); err != nil {
		t.Fatalf("decode backup: %v", err)
	}
	if backup.Name == "" {
		t.Fatalf("backup name is empty")
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/customers", `{"companyId":1,"name":"临时客户","contact":"临时","phone":"13000000000","creditLimit":1,"status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/system/backups/"+backup.Name+"/restore", `{}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("backup restore status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/master/customers", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("customers status %d: %s", rec.Code, rec.Body.String())
	}
	var customers []Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &customers); err != nil {
		t.Fatalf("decode customers: %v", err)
	}
	for _, customer := range customers {
		if customer.Name == "临时客户" {
			t.Fatalf("restore did not remove temporary customer")
		}
	}
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
