package appliance

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Session struct {
	Token        string   `json:"token"`
	User         User     `json:"user"`
	DeviceScopes []string `json:"deviceScopes,omitempty"`
	CreatedAt    string   `json:"createdAt"`
	LastSeenAt   string   `json:"lastSeenAt"`
	ExpiresAt    string   `json:"expiresAt"`
	IP           string   `json:"ip"`
	UserAgent    string   `json:"userAgent"`
}

type App struct {
	store       DataStore
	runtime     *RuntimeServices
	backups     *BackupManager
	hub         *Hub
	gateway     *DeviceGateway
	frontendDir string
	mu          sync.RWMutex
	sessions    map[string]Session
	ssoStates   map[string]OIDCLoginState
}

func NewApp(store DataStore, frontendDir string) *App {
	app := &App{store: store, runtime: NewRuntimeServicesFromEnv(), backups: NewBackupManagerFromEnv(), hub: NewHub(), frontendDir: frontendDir, sessions: map[string]Session{}, ssoStates: map[string]OIDCLoginState{}}
	app.gateway = NewDeviceGatewayFromEnv(app)
	return app
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", a.eventsHandler)
	mux.HandleFunc("/api/", a.apiHandler)
	if a.frontendDir != "" {
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(a.frontendDir, "assets")))))
		mux.HandleFunc("/", a.clientHandler)
	}
	return withCORS(logRequests(mux))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
			"Authorization",
			"Content-Type",
			"X-Device-Key",
			"X-CBMP-Timestamp",
			"X-CBMP-Signature",
			"X-CBMP-Request-Id",
			"X-CBMP-Channel-Token",
		}, ", "))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if !strings.HasPrefix(r.URL.Path, "/assets/") {
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
		}
	})
}

func (a *App) clientHandler(w http.ResponseWriter, r *http.Request) {
	if a.frontendDir == "" {
		http.NotFound(w, r)
		return
	}
	root := filepath.Clean(a.frontendDir)
	if r.URL.Path != "/" {
		candidate := filepath.Clean(filepath.Join(root, r.URL.Path))
		if rel, err := filepath.Rel(root, candidate); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			if _, err := os.Stat(candidate); err == nil {
				http.ServeFile(w, r, candidate)
				return
			}
		}
	}
	http.ServeFile(w, r, filepath.Join(root, "index.html"))
}

func (a *App) apiHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	parts := splitPath(path)
	if len(parts) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{
			"name":        "建材 ERP API",
			"product":     "common-build-materials-erp",
			"description": "客户侧私有化建材 ERP，用于销售、生产、实验室、调度、地磅、签收、结算、采购、库存和财务管理",
		})
		return
	}

	if parts[0] == "health" {
		a.health(w, r)
		return
	}
	data := a.mustSnapshot()
	if ipWhitelistEnforced(data) && !clientIPAllowed(data, clientIP(r)) {
		writeError(w, http.StatusForbidden, "IP not in whitelist")
		return
	}
	if parts[0] == "auth" && len(parts) == 2 && parts[1] == "login" {
		a.login(w, r)
		return
	}
	if parts[0] == "auth" && len(parts) >= 2 && parts[1] == "sso" {
		a.authSSO(w, r, parts[2:])
		return
	}
	if parts[0] == "scim" && len(parts) >= 2 && parts[1] == "v2" {
		a.scim(w, r, parts[2:])
		return
	}
	if parts[0] == "product-ops" {
		writeError(w, http.StatusNotFound, "product operations APIs belong to OperationsPlatform, not ERP")
		return
	}
	if isTaxGatewayCallbackPath(parts) && r.Method == http.MethodPost {
		a.taxGatewayCallback(w, r)
		return
	}
	if isCollectionCallbackPath(parts) && r.Method == http.MethodPost {
		a.collectionCallback(w, r)
		return
	}
	if parts[0] == "public" && len(parts) >= 2 && parts[1] == "delivery-sign" {
		a.publicDeliverySign(w, r, parts[2:])
		return
	}
	if parts[0] == "system" && len(parts) == 4 && parts[1] == "updates" && parts[3] == "download" && r.Method == http.MethodGet && updaterTokenFromRequest(r, "") != "" {
		a.systemUpdateRuntimeDownload(w, r, parts[2])
		return
	}

	session, ok := a.sessionFromRequest(r)
	if !ok && (parts[0] == "iot" || isWeighbridgeDevicePath(parts) || isProductionDevicePath(parts)) {
		session, ok = a.deviceSessionFromRequest(r)
	}
	if !ok {
		unauthorized(w)
		return
	}
	if permission := requiredPermission(parts, r.Method); permission != "" {
		if strings.HasPrefix(session.User.Username, "device:") {
			if !permissionGranted(session.DeviceScopes, permission) {
				writeError(w, http.StatusForbidden, "permission denied: "+permission)
				return
			}
		} else if !canAccess(a.mustSnapshot(), session.User, permission) {
			writeError(w, http.StatusForbidden, "permission denied: "+permission)
			return
		}
	}

	switch parts[0] {
	case "me":
		writeJSON(w, http.StatusOK, map[string]interface{}{"user": publicUser(session.User), "watermark": a.mustSnapshot().License.Watermark})
	case "bootstrap":
		a.bootstrap(w, r, session)
	case "dashboard":
		a.dashboard(w, r, session)
	case "master":
		a.master(w, r, session, parts[1:])
	case "contracts":
		a.contracts(w, r, session, parts[1:])
	case "orders":
		a.orders(w, r, session, parts[1:])
	case "production-plans":
		a.productionPlans(w, r, session, parts[1:])
	case "quality":
		a.quality(w, r, session, parts[1:])
	case "laboratory":
		a.laboratory(w, r, session, parts[1:])
	case "dispatch-center":
		a.dispatchCenter(w, r, session, parts[1:])
	case "dispatch-orders":
		a.dispatchOrders(w, r, session, parts[1:])
	case "weighbridge":
		a.weighbridge(w, r, session, parts[1:])
	case "delivery":
		a.delivery(w, r, session, parts[1:])
	case "statements":
		a.statements(w, r, session, parts[1:])
	case "procurement":
		a.procurement(w, r, session, parts[1:])
	case "finance":
		a.finance(w, r, session, parts[1:])
	case "portal":
		a.portal(w, r, session, parts[1:])
	case "vehicle":
		a.vehicle(w, r, session, parts[1:])
	case "iot":
		a.iot(w, r, session, parts[1:])
	case "rules":
		a.rules(w, r, session, parts[1:])
	case "integrations":
		a.integrations(w, r, session, parts[1:])
	case "approvals":
		a.approvals(w, r, session, parts[1:])
	case "reports":
		a.reports(w, r, session)
	case "system":
		a.system(w, r, session, parts[1:])
	case "simulate":
		a.simulate(w, r, session, parts[1:])
	default:
		writeError(w, http.StatusNotFound, "unknown api")
	}
}

func isWeighbridgeDevicePath(parts []string) bool {
	return len(parts) >= 2 && parts[0] == "weighbridge" && (parts[1] == "device-events" || parts[1] == "protocols")
}

func isProductionDevicePath(parts []string) bool {
	return len(parts) >= 2 && parts[0] == "production-plans" && parts[1] == "protocols"
}

func isTaxGatewayCallbackPath(parts []string) bool {
	return len(parts) == 3 && parts[0] == "finance" && parts[1] == "tax" && parts[2] == "callback"
}

func isCollectionCallbackPath(parts []string) bool {
	return len(parts) == 3 && parts[0] == "finance" && parts[1] == "collections" && parts[2] == "callback"
}

func (a *App) mustSnapshot() AppData {
	data, err := a.store.Snapshot()
	if err != nil {
		panic(err)
	}
	return data
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func unauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "unauthorized")
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func tokenString() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func (a *App) sessionByToken(token string) (Session, bool) {
	if token == "" {
		return Session{}, false
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	session, ok := a.sessions[token]
	return session, ok
}

func (a *App) sessionFromRequest(r *http.Request) (Session, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		if c, err := r.Cookie("cbmp_session"); err == nil {
			token = c.Value
		}
	}
	if token == "" {
		return Session{}, false
	}
	policy := buildSessionPolicy(a.mustSnapshot())
	now := time.Now()
	a.mu.Lock()
	session, ok := a.sessions[token]
	if !ok {
		a.mu.Unlock()
		return Session{}, false
	}
	if session.ExpiresAt == "" {
		session.ExpiresAt = sessionExpiresAt(session.CreatedAt, policy.TimeoutMinutes)
	}
	if expiresAt, ok := parseSecurityTime(session.ExpiresAt); ok && now.After(expiresAt) {
		delete(a.sessions, token)
		a.mu.Unlock()
		_ = a.store.Mutate(func(data *AppData) error {
			addAudit(data, session.User.Username, "session_expired", "session", session.User.ID, session.User.Username, clientIP(r))
			return nil
		})
		return Session{}, false
	}
	session.LastSeenAt = now.Format("2006-01-02 15:04:05")
	session.IP = clientIP(r)
	session.UserAgent = r.UserAgent()
	a.sessions[token] = session
	a.mu.Unlock()
	return session, true
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	data, err := a.store.Snapshot()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "ok",
		"service":       "cbmp-golang-appliance",
		"schemaVersion": data.SchemaVersion,
		"modules":       len(data.Modules),
		"runtime":       a.runtimeStatus(),
		"time":          nowString(),
	})
}

func (a *App) runtimeStatus() RuntimeStatus {
	status := a.runtime.Status()
	if _, ok := a.store.(*PostgresStore); ok {
		status.Storage = "postgres"
		status.BusinessTables = "enabled"
		status.BusinessTableCount = len(postgresProjectionTableNames)
		status.DomainTables = "enabled"
		status.DomainResourceCount = len(appDataDomainResources())
		if data, err := a.store.Snapshot(); err == nil {
			status.BusinessProjectionRows = businessProjectionRowCount(data)
			status.DomainRowCount = domainRowCount(data)
		}
	}
	if a.gateway != nil {
		status.DeviceGateways = a.gateway.Status()
	}
	return status
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		MFACode  string `json:"mfaCode"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid login payload")
		return
	}
	data := a.mustSnapshot()
	for _, user := range data.Users {
		if user.Username == req.Username && user.Status == "active" && verifyPassword(req.Password, user) {
			if user.MFAEnabled {
				step, ok := verifyTOTP(user.MFASecret, req.MFACode, time.Now(), user.MFALastUsedStep)
				if !ok {
					if strings.TrimSpace(req.MFACode) != "" {
						_ = a.store.Mutate(func(data *AppData) error {
							addAudit(data, user.Username, "failed_login", "sys_user", user.ID, "MFA 动态码无效", clientIP(r))
							return nil
						})
						writeError(w, http.StatusUnauthorized, "MFA 动态码无效")
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"mfaRequired": true,
						"username":    user.Username,
						"displayName": user.DisplayName,
					})
					return
				}
				_ = a.store.Mutate(func(data *AppData) error {
					for i := range data.Users {
						if data.Users[i].ID == user.ID {
							data.Users[i].MFALastUsedStep = step
							user = data.Users[i]
							return nil
						}
					}
					return nil
				})
			}
			session := a.issueSession(w, r, user)
			writeJSON(w, http.StatusOK, session)
			return
		}
	}
	_ = a.store.Mutate(func(data *AppData) error {
		addAudit(data, req.Username, "failed_login", "sys_user", 0, "用户名或密码错误", clientIP(r))
		return nil
	})
	writeError(w, http.StatusUnauthorized, "用户名或密码错误")
}

func (a *App) issueSession(w http.ResponseWriter, r *http.Request, user User) Session {
	return a.issueSessionWithDetail(w, r, user, "")
}

func (a *App) issueSessionWithDetail(w http.ResponseWriter, r *http.Request, user User, detail string) Session {
	token := tokenString()
	policy := buildSessionPolicy(a.mustSnapshot())
	createdAt := nowString()
	session := Session{
		Token: token, User: user, CreatedAt: createdAt, LastSeenAt: createdAt,
		ExpiresAt: sessionExpiresAt(createdAt, policy.TimeoutMinutes), IP: clientIP(r), UserAgent: r.UserAgent(),
	}
	a.mu.Lock()
	a.pruneSessionsForUserLocked(user.ID, policy)
	a.sessions[token] = session
	a.mu.Unlock()
	_ = a.store.Mutate(func(data *AppData) error {
		auditDetail := fallback(detail, "用户登录")
		if user.MFAEnabled && detail == "" {
			auditDetail = "用户 MFA 登录"
		}
		addAudit(data, user.Username, "login", "sys_user", user.ID, auditDetail, clientIP(r))
		return nil
	})
	http.SetCookie(w, &http.Cookie{Name: "cbmp_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return publicSession(session)
}

func publicSession(session Session) Session {
	session.User = publicUser(session.User)
	return session
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return strings.Split(v, ",")[0]
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx > -1 {
		return host[:idx]
	}
	return host
}

func clientIPAllowed(data AppData, ipText string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipText))
	if ip == nil {
		return false
	}
	for _, policy := range data.SecurityPolicies {
		if !policy.Enabled || policy.Type != "ip_whitelist" {
			continue
		}
		value := strings.TrimSpace(policy.Value)
		if value == "" {
			continue
		}
		if strings.Contains(value, "/") {
			if _, network, err := net.ParseCIDR(value); err == nil && network.Contains(ip) {
				return true
			}
			continue
		}
		if allowed := net.ParseIP(value); allowed != nil && allowed.Equal(ip) {
			return true
		}
	}
	return false
}

func (a *App) bootstrap(w http.ResponseWriter, r *http.Request, session Session) {
	data := scopedData(a.mustSnapshot(), session.User)
	for i := range data.Users {
		data.Users[i] = publicUser(data.Users[i])
	}
	data.Companies = publicCompanies(data.Companies)
	for i := range data.Roles {
		if data.Roles[i].DataScope == "tenant" {
			data.Roles[i].DataScope = "platform"
		}
	}
	data.CustomerComplaints = complaintsWithSLAStatus(data.CustomerComplaints, time.Now())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":                   publicUser(session.User),
		"license":                data.License,
		"modules":                data.Modules,
		"roles":                  data.Roles,
		"companies":              data.Companies,
		"departments":            data.Departments,
		"sites":                  data.Sites,
		"plants":                 data.Plants,
		"customers":              data.Customers,
		"customerContacts":       data.CustomerContacts,
		"customerBlacklists":     data.CustomerBlacklists,
		"customerProfiles":       data.CustomerProfiles,
		"customerComplaints":     data.CustomerComplaints,
		"pricePolicies":          data.PricePolicies,
		"taxRates":               data.TaxRates,
		"projects":               data.Projects,
		"products":               data.Products,
		"materials":              data.Materials,
		"carriers":               data.Carriers,
		"vehicles":               data.Vehicles,
		"drivers":                data.Drivers,
		"contracts":              data.Contracts,
		"contractAttachments":    data.ContractAttachments,
		"dispatchSchedules":      data.DispatchSchedules,
		"productionPlans":        data.ProductionPlans,
		"mixDesigns":             data.MixDesigns,
		"mixDesignTrialRuns":     data.MixDesignTrialRuns,
		"productionTasks":        data.ProductionTasks,
		"productionBatches":      data.ProductionBatches,
		"productionReports":      data.ProductionReports,
		"qualityInspections":     data.QualityInspections,
		"qualitySamples":         data.QualitySamples,
		"laboratorySamples":      data.LaboratorySamples,
		"laboratoryTests":        data.LaboratoryTests,
		"laboratoryEquipment":    data.LaboratoryEquipment,
		"laboratoryCalibrations": data.LaboratoryCalibrations,
		"qualityExceptions":      data.QualityExceptions,
		"inventory":              data.Inventory,
	})
}

func (a *App) dashboard(w http.ResponseWriter, r *http.Request, session Session) {
	data := scopedData(a.mustSnapshot(), session.User)
	reports := buildManagementReports(data)
	today := todayString()
	var todayQty, todayAmount, signedQty, signedAmount float64
	var todayOrders, dispatching, inTransit, completedDispatch, openAlarms int
	customerDebt := map[int64]float64{}
	siteProduction := map[int64]float64{}
	for _, order := range data.Orders {
		if strings.HasPrefix(order.CreatedAt, today) {
			todayOrders++
			todayQty += order.PlanQuantity
			todayAmount += orderTotalAmount(order)
		}
		if order.Status == "delivering" || order.Status == "dispatching" {
			dispatching++
		}
		customerDebt[order.CustomerID] += (order.PlanQuantity - order.SignedQty) * order.UnitPrice
	}
	for _, sign := range data.DeliverySigns {
		if strings.HasPrefix(sign.SignedAt, today) {
			signedQty += sign.SignedQty
			if order, ok := findOrder(data, sign.OrderID); ok {
				signedAmount += sign.SignedQty * order.UnitPrice
			}
		}
	}
	for _, plan := range data.ProductionPlans {
		if plan.PlanDate == today {
			siteProduction[plan.SiteID] += plan.ProducedQty
		}
	}
	for _, dispatch := range data.DispatchOrders {
		switch dispatch.Status {
		case "in_transit", "departed":
			inTransit++
		case "completed":
			completedDispatch++
		}
	}
	for _, alarm := range data.VehicleAlarms {
		if alarm.Status == "open" {
			openAlarms++
		}
	}
	online := 0
	for _, loc := range data.LatestLocations {
		if loc.OnlineStatus == "online" {
			online++
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kpis": map[string]interface{}{
			"todayOrders":       todayOrders,
			"todayPlannedQty":   round(todayQty),
			"todaySalesAmount":  round(todayAmount),
			"todaySignedQty":    round(signedQty),
			"todaySignedAmount": round(signedAmount),
			"dispatchingOrders": dispatching,
			"inTransitVehicles": inTransit,
			"completedDispatch": completedDispatch,
			"vehicleOnlineRate": percent(online, len(data.LatestLocations)),
			"openAlarms":        openAlarms,
			"grossMargin":       reports.Operating.GrossMargin,
			"overdueReceivable": reports.Operating.OverdueReceivable,
			"qualityPassRate":   reports.Quality.PassRate,
			"unitPowerKwh":      firstUnitPowerKwh(reports.Energy),
		},
		"siteProduction": siteProduction,
		"customerDebt":   customerDebt,
		"recentOrders":   lastOrders(data.Orders, 8),
		"alarms":         data.VehicleAlarms,
		"operating":      reports.Operating,
		"customerAging":  reports.CustomerAging,
		"quality":        reports.Quality,
		"energy":         reports.Energy,
	})
}

func (a *App) master(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		writeError(w, http.StatusBadRequest, "missing master resource")
		return
	}
	resource := parts[0]
	if len(parts) == 1 && resource == "export" && r.Method == http.MethodGet {
		a.exportMasterData(w, r, session)
		return
	}
	if len(parts) == 1 && resource == "import" && r.Method == http.MethodPost {
		a.importMasterData(w, r, session)
		return
	}
	if len(parts) == 3 && resource == "customer-contacts" && parts[2] == "default" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.setDefaultCustomerContact(w, r, session, id)
		return
	}
	if len(parts) == 3 && resource == "customer-blacklists" && parts[2] == "release" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.releaseCustomerBlacklist(w, r, session, id)
		return
	}
	if len(parts) == 2 && resource == "customer-profiles" && parts[1] == "evaluate" && r.Method == http.MethodPost {
		a.evaluateCustomerProfiles(w, r, session)
		return
	}
	if len(parts) == 3 && resource == "customer-complaints" && parts[2] == "close" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.closeCustomerComplaint(w, r, session, id)
		return
	}
	if len(parts) == 2 && resource == "pricing" && parts[1] == "evaluate" && r.Method == http.MethodPost {
		a.evaluatePricing(w, r, session)
		return
	}
	if len(parts) == 2 && (r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete) {
		id, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "invalid master resource id")
			return
		}
		if r.Method == http.MethodDelete {
			a.deleteMasterResource(w, r, session, resource, id)
			return
		}
		a.updateMasterResource(w, r, session, resource, id)
		return
	}
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		switch resource {
		case "customers":
			writeJSON(w, http.StatusOK, data.Customers)
		case "customer-contacts":
			writeJSON(w, http.StatusOK, data.CustomerContacts)
		case "customer-blacklists":
			writeJSON(w, http.StatusOK, data.CustomerBlacklists)
		case "customer-profiles":
			writeJSON(w, http.StatusOK, data.CustomerProfiles)
		case "customer-complaints":
			writeJSON(w, http.StatusOK, complaintsWithSLAStatus(data.CustomerComplaints, time.Now()))
		case "projects":
			writeJSON(w, http.StatusOK, data.Projects)
		case "products":
			writeJSON(w, http.StatusOK, data.Products)
		case "price-policies":
			writeJSON(w, http.StatusOK, data.PricePolicies)
		case "tax-rates":
			writeJSON(w, http.StatusOK, data.TaxRates)
		case "vehicles":
			writeJSON(w, http.StatusOK, data.Vehicles)
		case "drivers":
			writeJSON(w, http.StatusOK, data.Drivers)
		case "materials":
			writeJSON(w, http.StatusOK, data.Materials)
		case "inventory":
			writeJSON(w, http.StatusOK, data.Inventory)
		case "sites":
			writeJSON(w, http.StatusOK, data.Sites)
		case "carriers":
			writeJSON(w, http.StatusOK, data.Carriers)
		default:
			writeError(w, http.StatusNotFound, "unknown master resource")
		}
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	switch resource {
	case "customers":
		var item Customer
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid customer")
			return
		}
		err := a.store.Mutate(func(data *AppData) error {
			item.ID = nextID(data, "customer")
			if item.Status == "" {
				item.Status = "active"
			}
			data.Customers = append(data.Customers, item)
			addAudit(data, session.User.Username, "create", "customer", item.ID, item.Name, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, item, "master.customer.created")
	case "customer-contacts":
		a.createCustomerContact(w, r, session)
	case "customer-blacklists":
		a.createCustomerBlacklist(w, r, session)
	case "customer-profiles":
		a.createCustomerProfile(w, r, session)
	case "customer-complaints":
		a.createCustomerComplaint(w, r, session)
	case "projects":
		a.createProject(w, r, session)
	case "products":
		a.createProduct(w, r, session)
	case "materials":
		a.createMaterial(w, r, session)
	case "drivers":
		a.createDriver(w, r, session)
	case "sites":
		a.createSite(w, r, session)
	case "inventory":
		a.createInventoryItem(w, r, session)
	case "carriers":
		a.createCarrier(w, r, session)
	case "price-policies":
		a.createPricePolicy(w, r, session)
	case "tax-rates":
		a.createTaxRate(w, r, session)
	case "vehicles":
		var item Vehicle
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid vehicle")
			return
		}
		err := a.store.Mutate(func(data *AppData) error {
			item.ID = nextID(data, "vehicle")
			if item.Status == "" {
				item.Status = "active"
			}
			if item.OnlineStatus == "" {
				item.OnlineStatus = "offline"
			}
			if item.BusinessStatus == "" {
				item.BusinessStatus = "idle"
			}
			data.Vehicles = append(data.Vehicles, item)
			addAudit(data, session.User.Username, "create", "vehicle", item.ID, item.PlateNo, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, item, "master.vehicle.created")
	default:
		writeError(w, http.StatusNotFound, "unknown master resource")
	}
}

func (a *App) respondMutation(w http.ResponseWriter, err error, payload interface{}, topic string) {
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit(topic, payload)
	writeJSON(w, http.StatusCreated, payload)
}

func (a *App) emit(topic string, payload interface{}) {
	a.hub.Broadcast(topic, payload)
	if a.runtime != nil {
		a.runtime.Publish(topic, payload)
	}
}

func (a *App) contracts(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 2 && parts[1] == "submit" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.submitContract(w, r, session, id)
		return
	}
	if len(parts) == 2 && parts[1] == "revise" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.reviseContract(w, r, session, id)
		return
	}
	if len(parts) == 2 && parts[1] == "attachments" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		if r.Method == http.MethodGet {
			data := scopedData(a.mustSnapshot(), session.User)
			writeJSON(w, http.StatusOK, filter(data.ContractAttachments, func(item ContractAttachment) bool { return item.ContractID == id }))
			return
		}
		if r.Method == http.MethodPost {
			a.createContractAttachment(w, r, session, id)
			return
		}
	}
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).Contracts)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	a.createContract(w, r, session)
}

func (a *App) orders(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).Orders)
			return
		}
		if r.Method == http.MethodPost {
			a.createOrder(w, r, session)
			return
		}
	}
	if len(parts) == 2 && parts[1] == "approve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.setOrderStatus(w, r, session, id, "approved")
		return
	}
	if len(parts) == 2 && parts[1] == "close" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.setOrderStatus(w, r, session, id, "closed")
		return
	}
	writeError(w, http.StatusNotFound, "unknown order route")
}

func (a *App) createOrder(w http.ResponseWriter, r *http.Request, session Session) {
	var item SalesOrder
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid order")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		customer, ok := findCustomer(*data, item.CustomerID)
		if !ok {
			return fmt.Errorf("客户不存在")
		}
		if risk, blocked := orderBlockedByCustomerRisk(*data, customer); blocked {
			addAudit(data, session.User.Username, "block", "sales_order", customer.ID, risk.Reason, clientIP(r))
			return fmt.Errorf("客户 %s 已被风控停供：%s", customer.Name, risk.Reason)
		}
		project, ok := findProject(*data, item.ProjectID)
		if !ok {
			return fmt.Errorf("项目不存在")
		}
		riskFlags, riskReasons, err := prepareSalesOrderLines(data, &item, customer, project)
		if err != nil {
			return err
		}
		item.ID = nextID(data, "order")
		item.OrderNo = number("SO", item.ID)
		item.ReceiveAddress = fallback(item.ReceiveAddress, project.Address)
		item.Contact = fallback(item.Contact, project.Contact)
		item.Phone = fallback(item.Phone, project.Phone)
		item.CreatedAt = nowString()
		if customer.Receivable+orderTotalAmount(item) > customer.CreditLimit {
			riskFlags = appendOrderRisk(riskFlags, "credit_limit")
			riskReasons = append(riskReasons, fmt.Sprintf("客户应收 %.2f + 本单 %.2f 超过信用额度 %.2f", customer.Receivable, orderTotalAmount(item), customer.CreditLimit))
		}
		if len(riskFlags) > 0 {
			item.RiskFlag = strings.Join(riskFlags, ",")
			item.Status = "pending_approval"
		} else {
			item.Status = "submitted"
		}
		data.Orders = append(data.Orders, item)
		if len(riskFlags) > 0 {
			flowCode := "order_credit_risk"
			title := "销售订单风险审批"
			if strings.Contains(item.RiskFlag, "price_below_floor") && !strings.Contains(item.RiskFlag, "credit_limit") {
				flowCode = "price_below_floor"
				title = "低于底价销售订单审批"
			}
			if _, err := submitApprovalTask(
				data,
				flowCode,
				"sales_order",
				item.ID,
				item.OrderNo,
				title,
				session.User.Username,
				strings.Join(riskReasons, "；"),
			); err != nil {
				return err
			}
		}
		addAudit(data, session.User.Username, "create", "sales_order", item.ID, item.OrderNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "sales.order.created")
}

func (a *App) setOrderStatus(w http.ResponseWriter, r *http.Request, session Session, id int64, status string) {
	var updated SalesOrder
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Orders {
			if data.Orders[i].ID == id {
				if status == "approved" && data.Orders[i].Status == "pending_approval" {
					return fmt.Errorf("订单存在待审批任务，需先完成通用审批")
				}
				data.Orders[i].Status = status
				updated = data.Orders[i]
				addAudit(data, session.User.Username, "status", "sales_order", id, status, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("订单不存在")
	})
	a.respondMutation(w, err, updated, "sales.order.update")
}

func (a *App) productionPlans(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	a.production(w, r, session, parts)
}

func (a *App) dispatchOrders(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "schedules" {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).DispatchSchedules)
			return
		}
		if r.Method == http.MethodPost {
			a.createDispatchSchedule(w, r, session)
			return
		}
	}
	if len(parts) == 1 && parts[0] == "carrier-settlements" && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, map[string]interface{}{"settlements": data.TransportSettlements, "items": data.TransportSettlementItems})
		return
	}
	if len(parts) == 2 && parts[0] == "carrier-settlements" && parts[1] == "generate" && r.Method == http.MethodPost {
		a.generateCarrierSettlement(w, r, session)
		return
	}
	if len(parts) == 0 {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).DispatchOrders)
			return
		}
		if r.Method == http.MethodPost {
			a.createDispatch(w, r, session)
			return
		}
	}
	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		a.advanceDispatch(w, r, session, id)
		return
	}
	writeError(w, http.StatusNotFound, "unknown dispatch route")
}

func (a *App) createDispatch(w http.ResponseWriter, r *http.Request, session Session) {
	var item DispatchOrder
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid dispatch")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		order, ok := findOrder(*data, item.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		if order.Status != "approved" && order.Status != "scheduled" && order.Status != "dispatching" && order.Status != "delivering" {
			return fmt.Errorf("订单状态不允许派车")
		}
		vehicle, ok := findVehicle(*data, item.VehicleID)
		if !ok {
			return fmt.Errorf("车辆不存在")
		}
		if vehicle.Status != "active" {
			return fmt.Errorf("车辆不可用")
		}
		remainingQty := order.PlanQuantity - order.DispatchedQty
		if item.LineID != 0 {
			line, ok := findOrderLine(order, item.LineID)
			if !ok {
				return fmt.Errorf("订单明细不存在")
			}
			remainingQty = line.Quantity - dispatchedQtyForOrderLine(*data, order.ID, line.ID)
			item.LineSeq = line.Seq
			item.ProductID = line.ProductID
			item.ProductName = line.ProductName
			if item.ProductName == "" {
				if product, ok := findProduct(*data, line.ProductID); ok {
					item.ProductName = productName(product)
				}
			}
		} else {
			lines := orderLines(order)
			if len(lines) == 1 {
				item.LineID = lines[0].ID
				item.LineSeq = lines[0].Seq
				item.ProductID = lines[0].ProductID
				item.ProductName = lines[0].ProductName
			} else {
				item.ProductID = order.ProductID
			}
			if item.ProductName == "" && item.ProductID != 0 {
				if product, ok := findProduct(*data, item.ProductID); ok {
					item.ProductName = productName(product)
				}
			}
		}
		if remainingQty <= 0 {
			return fmt.Errorf("订单已无剩余可派量")
		}
		item.ID = nextID(data, "dispatch")
		item.DispatchNo = number("DO", item.ID)
		item.DriverID = nonZeroInt(item.DriverID, vehicle.DriverID)
		item.SiteID = order.SiteID
		item.ProjectID = order.ProjectID
		item.PlanQuantity = nonZero(item.PlanQuantity, math.Min(36, remainingQty))
		if item.PlanQuantity <= 0 {
			return fmt.Errorf("派车数量必须大于 0")
		}
		if item.PlanQuantity > remainingQty {
			return fmt.Errorf("派车数量超过剩余可派量")
		}
		item.QueueNo = fmt.Sprintf("A%03d", item.ID+20)
		item.ETA = time.Now().Add(80 * time.Minute).Format("2006-01-02 15:04:05")
		item.Status = "assigned"
		item.CreatedAt = nowString()
		item.UpdatedAt = item.CreatedAt
		if err := reserveDispatchSchedule(data, item); err != nil {
			return err
		}
		data.DispatchOrders = append(data.DispatchOrders, item)
		for i := range data.Orders {
			if data.Orders[i].ID == order.ID {
				data.Orders[i].Status = "dispatching"
				data.Orders[i].DispatchedQty += item.PlanQuantity
			}
		}
		for i := range data.Vehicles {
			if data.Vehicles[i].ID == vehicle.ID {
				data.Vehicles[i].BusinessStatus = "assigned"
			}
		}
		addAudit(data, session.User.Username, "create", "dispatch_order", item.ID, item.DispatchNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "dispatch.order.update")
}

func (a *App) advanceDispatch(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Status    string `json:"status"`
		Exception string `json:"exception"`
	}
	_ = readJSON(r, &req)
	var updated DispatchOrder
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.DispatchOrders {
			if data.DispatchOrders[i].ID == id {
				next := req.Status
				if next == "" {
					next = nextDispatchStatus(data.DispatchOrders[i].Status)
				}
				data.DispatchOrders[i].Status = next
				data.DispatchOrders[i].Exception = req.Exception
				data.DispatchOrders[i].UpdatedAt = nowString()
				updated = data.DispatchOrders[i]
				updateVehicleStatus(data, updated.VehicleID, vehicleStatusForDispatch(next))
				addAudit(data, session.User.Username, "status", "dispatch_order", id, next, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("派车单不存在")
	})
	a.respondMutation(w, err, updated, "dispatch.order.update")
}

func (a *App) weighbridge(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "tickets" {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).ScaleTickets)
			return
		}
		if r.Method == http.MethodPost {
			a.createTicket(w, r, session)
			return
		}
	}
	if len(parts) == 2 && parts[0] == "tickets" && parts[1] == "transfer" && r.Method == http.MethodPost {
		a.createTransferTicket(w, r, session)
		return
	}
	if len(parts) == 2 && parts[0] == "tickets" && parts[1] == "return" && r.Method == http.MethodPost {
		a.createReturnTicket(w, r, session)
		return
	}
	if len(parts) == 2 && parts[0] == "tickets" && parts[1] == "waste" && r.Method == http.MethodPost {
		a.createWasteTicket(w, r, session)
		return
	}
	if len(parts) == 1 && parts[0] == "ticket-prints" && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, visibleTicketPrintLogs(data))
		return
	}
	if len(parts) == 1 && parts[0] == "ticket-voids" && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, visibleTicketVoidLogs(data))
		return
	}
	if len(parts) == 1 && parts[0] == "weight-records" && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, data.ScaleWeightRecords)
		return
	}
	if len(parts) == 1 && parts[0] == "device-events" {
		a.scaleDeviceEvents(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "protocols" && parts[1] == "scale" && parts[2] == "ingest" && r.Method == http.MethodPost {
		a.ingestScaleProtocolFrame(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "tickets" && parts[2] == "reprint" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.reprintTicket(w, r, session, id)
		return
	}
	if len(parts) == 4 && parts[0] == "tickets" && parts[2] == "void" && parts[3] == "request" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.requestVoidTicket(w, r, session, id)
		return
	}
	if len(parts) == 4 && parts[0] == "tickets" && parts[2] == "void" && parts[3] == "approve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.approveVoidTicket(w, r, session, id)
		return
	}
	writeError(w, http.StatusNotFound, "unknown weighbridge route")
}

func (a *App) createTicket(w http.ResponseWriter, r *http.Request, session Session) {
	var item ScaleTicket
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		dispatch, ok := findDispatch(*data, item.DispatchID)
		if !ok {
			return fmt.Errorf("派车单不存在")
		}
		order, ok := findOrder(*data, dispatch.OrderID)
		if !ok {
			return fmt.Errorf("订单不存在")
		}
		vehicle, ok := findVehicle(*data, dispatch.VehicleID)
		if !ok {
			return fmt.Errorf("车辆不存在")
		}
		item.ID = nextID(data, "ticket")
		item.TicketNo = number("ST", item.ID)
		item.TicketType = "product_out"
		item.OrderID = order.ID
		item.SiteID = dispatch.SiteID
		item.VehicleID = vehicle.ID
		item.PlateNo = vehicle.PlateNo
		item.NetWeight = round(item.GrossWeight - item.TareWeight)
		if item.NetWeight <= 0 {
			item.NetWeight = dispatch.PlanQuantity
		}
		item.Unit = order.Unit
		if item.SnapshotURL == "" {
			device := defaultScaleDeviceForSite(*data, dispatch.SiteID)
			item.SnapshotURL = scaleCaptureURL(device.Code, item.TicketNo, "gross")
		}
		item.PrintCount = 1
		item.SignStatus = "pending"
		item.SettlementStatus = "pending"
		item.Status = "valid"
		item.CreatedAt = nowString()
		data.ScaleTickets = append(data.ScaleTickets, item)
		device := defaultScaleDeviceForSite(*data, dispatch.SiteID)
		appendWeightRecord(data, device.ID, item.ID, item.PlateNo, item.TareWeight, "tare", scaleCaptureURL(device.Code, item.TicketNo, "tare"), item.CreatedAt)
		appendWeightRecord(data, device.ID, item.ID, item.PlateNo, item.GrossWeight, "gross", item.SnapshotURL, item.CreatedAt)
		for i := range data.DispatchOrders {
			if data.DispatchOrders[i].ID == dispatch.ID {
				data.DispatchOrders[i].LoadedQty = dispatch.PlanQuantity
				data.DispatchOrders[i].Status = "departed"
				data.DispatchOrders[i].UpdatedAt = nowString()
			}
		}
		for i := range data.Orders {
			if data.Orders[i].ID == order.ID {
				data.Orders[i].Status = "delivering"
			}
		}
		updateVehicleStatus(data, vehicle.ID, "in_transit")
		addAudit(data, session.User.Username, "create", "scale_ticket", item.ID, item.TicketNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "ticket.created")
}

func (a *App) reprintTicket(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var logItem TicketPrintLog
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.ScaleTickets {
			if data.ScaleTickets[i].ID != id {
				continue
			}
			if !userCanAccessTicket(*data, session.User, data.ScaleTickets[i]) {
				return fmt.Errorf("无权操作该票据")
			}
			if data.ScaleTickets[i].Status != "valid" {
				return fmt.Errorf("非有效票据不能补打")
			}
			data.ScaleTickets[i].PrintCount++
			logItem = TicketPrintLog{
				ID:        nextID(data, "printLog"),
				TicketID:  id,
				PrintedBy: session.User.Username,
				PrintedAt: nowString(),
			}
			data.TicketPrintLogs = append(data.TicketPrintLogs, logItem)
			addAudit(data, session.User.Username, "reprint", "scale_ticket", id, data.ScaleTickets[i].TicketNo, clientIP(r))
			return nil
		}
		return fmt.Errorf("票据不存在")
	})
	a.respondMutation(w, err, logItem, "ticket.reprinted")
}

func (a *App) requestVoidTicket(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid void request")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "作废原因不能为空")
		return
	}
	var logItem TicketVoidLog
	err := a.store.Mutate(func(data *AppData) error {
		var ticket ScaleTicket
		found := false
		for _, item := range data.ScaleTickets {
			if item.ID == id {
				ticket = item
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("票据不存在")
		}
		if !userCanAccessTicket(*data, session.User, ticket) {
			return fmt.Errorf("无权操作该票据")
		}
		if ticket.Status != "valid" {
			return fmt.Errorf("非有效票据不能申请作废")
		}
		for _, item := range data.TicketVoidLogs {
			if item.TicketID == id && item.Status == "pending" {
				return fmt.Errorf("该票据已有待审批作废申请")
			}
		}
		logItem = TicketVoidLog{
			ID:        nextID(data, "voidLog"),
			TicketID:  id,
			Reason:    req.Reason,
			Status:    "pending",
			CreatedAt: nowString(),
		}
		data.TicketVoidLogs = append(data.TicketVoidLogs, logItem)
		addAudit(data, session.User.Username, "request_void", "scale_ticket", id, ticket.TicketNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, logItem, "ticket.void.requested")
}

func (a *App) approveVoidTicket(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Approved bool `json:"approved"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid void approval")
		return
	}
	var logItem TicketVoidLog
	topic := "ticket.void.rejected"
	err := a.store.Mutate(func(data *AppData) error {
		ticketIndex := -1
		for i := range data.ScaleTickets {
			if data.ScaleTickets[i].ID == id {
				ticketIndex = i
				break
			}
		}
		if ticketIndex < 0 {
			return fmt.Errorf("票据不存在")
		}
		if !userCanAccessTicket(*data, session.User, data.ScaleTickets[ticketIndex]) {
			return fmt.Errorf("无权操作该票据")
		}
		voidIndex := -1
		for i := range data.TicketVoidLogs {
			if data.TicketVoidLogs[i].TicketID == id && data.TicketVoidLogs[i].Status == "pending" {
				voidIndex = i
				break
			}
		}
		if voidIndex < 0 {
			return fmt.Errorf("没有待审批作废申请")
		}
		data.TicketVoidLogs[voidIndex].ApprovedBy = session.User.Username
		if req.Approved {
			data.TicketVoidLogs[voidIndex].Status = "approved"
			data.ScaleTickets[ticketIndex].Status = "void"
			data.ScaleTickets[ticketIndex].SettlementStatus = "void"
			data.ScaleTickets[ticketIndex].SignStatus = "void"
			topic = "ticket.void.approved"
			addAudit(data, session.User.Username, "approve_void", "scale_ticket", id, data.ScaleTickets[ticketIndex].TicketNo, clientIP(r))
		} else {
			data.TicketVoidLogs[voidIndex].Status = "rejected"
			addAudit(data, session.User.Username, "reject_void", "scale_ticket", id, data.ScaleTickets[ticketIndex].TicketNo, clientIP(r))
		}
		logItem = data.TicketVoidLogs[voidIndex]
		return nil
	})
	a.respondMutation(w, err, logItem, topic)
}

func (a *App) delivery(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "sign" {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).DeliverySigns)
			return
		}
		if r.Method == http.MethodPost {
			a.signDelivery(w, r, session)
			return
		}
	}
	if len(parts) == 1 && parts[0] == "sign-links" {
		if r.Method == http.MethodGet {
			a.listDeliverySignLinks(w, r, session)
			return
		}
		if r.Method == http.MethodPost {
			a.createDeliverySignLink(w, r, session)
			return
		}
	}
	if len(parts) == 1 && parts[0] == "sign-attachments" && r.Method == http.MethodGet {
		a.listDeliverySignAttachments(w, r, session)
		return
	}
	if signID, ok := parseSignID(parts); ok && r.Method == http.MethodPost {
		a.addDeliverySignAttachment(w, r, session, signID)
		return
	}
	writeError(w, http.StatusNotFound, "unknown delivery route")
}

func (a *App) signDelivery(w http.ResponseWriter, r *http.Request, session Session) {
	var req deliverySignRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sign")
		return
	}
	var item DeliverySign
	err := a.store.Mutate(func(data *AppData) error {
		sign, err := completeDeliverySign(data, req.DeliverySign, req.Attachments, session.User.Username, clientIP(r))
		item = sign
		return err
	})
	a.respondMutation(w, err, item, "delivery.signed")
}

func (a *App) statements(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).Statements)
		return
	}
	if len(parts) == 2 && parts[1] == "confirm" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var updated Statement
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Statements {
				if data.Statements[i].ID == id {
					data.Statements[i].Status = "confirmed"
					data.Statements[i].ConfirmedBy = session.User.DisplayName
					data.Statements[i].ConfirmedAt = nowString()
					updated = data.Statements[i]
					addAudit(data, session.User.Username, "confirm", "customer_statement", id, updated.StatementNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("对账单不存在")
		})
		a.respondMutation(w, err, updated, "statement.confirmed")
		return
	}
	writeError(w, http.StatusNotFound, "unknown statement route")
}

func (a *App) vehicle(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) >= 1 && parts[0] == "fences" {
		a.geoFences(w, r, session, parts[1:])
		return
	}
	if len(parts) == 1 && parts[0] == "fence-events" {
		a.geoFenceEvents(w, r, session)
		return
	}
	if len(parts) == 2 && parts[0] == "location" && parts[1] == "latest" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).LatestLocations)
		return
	}
	if len(parts) == 2 && parts[0] == "track" && parts[1] == "replay" && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		vehicleID, _ := strconv.ParseInt(r.URL.Query().Get("vehicleId"), 10, 64)
		replay := buildTrackReplay(data, vehicleID, r.URL.Query().Get("startTime"), r.URL.Query().Get("endTime"))
		writeJSON(w, http.StatusOK, replay)
		return
	}
	if len(parts) == 1 && parts[0] == "track" && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		vehicleID, _ := strconv.ParseInt(r.URL.Query().Get("vehicleId"), 10, 64)
		startTime := r.URL.Query().Get("startTime")
		endTime := r.URL.Query().Get("endTime")
		out := []VehicleLocationEvent{}
		for _, loc := range data.Locations {
			if (vehicleID == 0 || loc.VehicleID == vehicleID) && betweenTime(loc.LocationTime, startTime, endTime) {
				out = append(out, loc)
			}
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	if len(parts) == 1 && parts[0] == "alarms" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).VehicleAlarms)
		return
	}
	writeError(w, http.StatusNotFound, "unknown vehicle route")
}

func (a *App) iot(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 3 && parts[0] == "vehicle" && parts[1] == "location" && parts[2] == "report" && r.Method == http.MethodPost {
		a.reportLocation(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "vehicle" && parts[1] == "location" && parts[2] == "batch" && r.Method == http.MethodPost {
		a.reportLocationBatch(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "protocols" && parts[1] == "gps" && parts[2] == "ingest" && r.Method == http.MethodPost {
		a.ingestGPSProtocolFrame(w, r, session)
		return
	}
	writeError(w, http.StatusNotFound, "unknown iot route")
}

func (a *App) reportLocation(w http.ResponseWriter, r *http.Request, session Session) {
	var req locationReportPayload
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid location")
		return
	}
	if strings.HasPrefix(session.User.Username, "device:") {
		deviceNo := strings.TrimPrefix(session.User.Username, "device:")
		if req.DeviceNo != "" && req.DeviceNo != deviceNo {
			writeError(w, http.StatusForbidden, "device key does not match payload")
			return
		}
	}
	event, latest, err := a.recordLocationReport(r, session, req)
	if err == nil {
		a.runtime.CacheLatestLocation(latest)
		a.runtime.StoreTrackPoint(event)
	}
	a.respondMutation(w, err, event, "vehicle.location.update")
}

func (a *App) recordLocationReport(r *http.Request, session Session, req locationReportPayload) (VehicleLocationEvent, VehicleLatestLocation, error) {
	if strings.HasPrefix(session.User.Username, "device:") {
		deviceNo := strings.TrimPrefix(session.User.Username, "device:")
		if req.DeviceNo != "" && req.DeviceNo != deviceNo {
			return VehicleLocationEvent{}, VehicleLatestLocation{}, fmt.Errorf("device key does not match payload")
		}
		req.DeviceNo = deviceNo
	}
	var event VehicleLocationEvent
	var latest VehicleLatestLocation
	err := a.store.Mutate(func(data *AppData) error {
		if req.PlateNo == "" && req.DeviceNo != "" {
			if vehicle, ok := findVehicleByDeviceNo(*data, req.DeviceNo); ok {
				req.PlateNo = vehicle.PlateNo
			}
		}
		vehicle, ok := findVehicleByPlate(*data, req.PlateNo)
		if !ok {
			return fmt.Errorf("车辆不存在")
		}
		if session.User.RoleCode == "driver" && (session.User.DriverID == 0 || vehicle.DriverID != session.User.DriverID) {
			return fmt.Errorf("司机无权上报该车辆位置")
		}
		dispatchID, orderID, projectID, customerID := currentTripContext(*data, vehicle.ID)
		abnormal := req.Speed > 80
		abnormalType := ""
		if abnormal {
			abnormalType = "speeding"
		}
		event = VehicleLocationEvent{
			ID: nextID(data, "location"), VehicleID: vehicle.ID, PlateNo: vehicle.PlateNo,
			DriverID: vehicle.DriverID, DispatchID: dispatchID, DeviceID: req.DeviceNo,
			SourceType: fallback(req.SourceType, "gps_device"), Longitude: req.Longitude,
			Latitude: req.Latitude, Speed: req.Speed, Direction: req.Direction,
			Mileage: req.Mileage, AccStatus: req.AccStatus, OnlineStatus: "online",
			Address:    inferAddress(*data, req.Longitude, req.Latitude),
			IsAbnormal: abnormal, AbnormalType: abnormalType,
			LocationTime: fallback(req.LocationTime, nowString()), ReceiveTime: nowString(),
		}
		data.Locations = append(data.Locations, event)
		if len(data.Locations) > 2000 {
			data.Locations = data.Locations[len(data.Locations)-2000:]
		}
		latest = VehicleLatestLocation{
			VehicleID: vehicle.ID, PlateNo: vehicle.PlateNo, Longitude: req.Longitude, Latitude: req.Latitude,
			Speed: req.Speed, Direction: req.Direction, OnlineStatus: "online",
			TransportStatus: vehicle.BusinessStatus, LastLocationTime: event.LocationTime,
			CurrentOrderID: orderID, CurrentProjectID: projectID, CurrentSiteID: vehicle.SiteID, CurrentCustomerID: customerID,
		}
		upsertLatestLocation(data, latest)
		for i := range data.Vehicles {
			if data.Vehicles[i].ID == vehicle.ID {
				data.Vehicles[i].OnlineStatus = "online"
			}
		}
		if abnormal {
			data.VehicleAlarms = append(data.VehicleAlarms, VehicleAlarm{
				ID: nextID(data, "alarm"), VehicleID: vehicle.ID, DispatchID: dispatchID,
				AlarmType: "speeding", Level: "high", Message: vehicle.PlateNo + " 速度超过 80km/h",
				Status: "open", CreatedAt: nowString(),
			})
		}
		createFenceEvents(data, event)
		addAudit(data, session.User.Username, "report", "vehicle_location", event.ID, vehicle.PlateNo, clientIP(r))
		return nil
	})
	return event, latest, err
}

func (a *App) reports(w http.ResponseWriter, r *http.Request, session Session) {
	data := scopedData(a.mustSnapshot(), session.User)
	writeJSON(w, http.StatusOK, buildManagementReports(data))
}

func firstUnitPowerKwh(items []ProductionEnergyReport) float64 {
	if len(items) == 0 {
		return 0
	}
	return items[0].UnitPowerKwh
}

func (a *App) system(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		writeError(w, http.StatusBadRequest, "missing system resource")
		return
	}
	data := a.mustSnapshot()
	switch parts[0] {
	case "license":
		a.systemLicense(w, r, session, parts[1:])
	case "modules":
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, data.Modules)
			return
		}
		if len(parts) == 2 && r.Method == http.MethodPost {
			code := parts[1]
			var req struct {
				Enabled bool `json:"enabled"`
			}
			_ = readJSON(r, &req)
			var updated Module
			err := a.store.Mutate(func(data *AppData) error {
				for i := range data.Modules {
					if data.Modules[i].Code == code {
						data.Modules[i].Enabled = req.Enabled
						updated = data.Modules[i]
						addAudit(data, session.User.Username, "toggle", "module", 0, code, clientIP(r))
						return nil
					}
				}
				return fmt.Errorf("模块不存在")
			})
			a.respondMutation(w, err, updated, "system.module.update")
			return
		}
	case "plugins":
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, data.Plugins)
			return
		}
		if len(parts) == 2 && parts[1] == "install" && r.Method == http.MethodPost {
			var plugin Plugin
			if err := readJSON(r, &plugin); err != nil {
				writeError(w, http.StatusBadRequest, "invalid plugin")
				return
			}
			err := a.store.Mutate(func(data *AppData) error {
				if !checksumVerified(plugin.Checksum, plugin.Signature) {
					return fmt.Errorf("插件验签失败")
				}
				if plugin.ID == "" {
					return fmt.Errorf("插件 ID 不能为空")
				}
				plugin.Status = fallback(plugin.Status, "installed")
				for i := range data.Plugins {
					if data.Plugins[i].ID == plugin.ID {
						data.Plugins[i] = plugin
						addAudit(data, session.User.Username, "upgrade", "plugin", 0, plugin.ID, clientIP(r))
						return nil
					}
				}
				data.Plugins = append(data.Plugins, plugin)
				addAudit(data, session.User.Username, "install", "plugin", 0, plugin.ID, clientIP(r))
				return nil
			})
			a.respondMutation(w, err, plugin, "system.plugin.installed")
			return
		}
		if len(parts) == 2 && parts[1] == "runs" {
			a.systemPluginRuns(w, r)
			return
		}
		if len(parts) == 3 && parts[2] == "verify" && r.Method == http.MethodPost {
			id := parts[1]
			for _, plugin := range data.Plugins {
				if plugin.ID == id {
					writeJSON(w, http.StatusOK, map[string]interface{}{"plugin": plugin, "valid": checksumVerified(plugin.Checksum, plugin.Signature)})
					return
				}
			}
			writeError(w, http.StatusNotFound, "插件不存在")
			return
		}
		if len(parts) == 3 && parts[2] == "run" && r.Method == http.MethodPost {
			a.runPlugin(w, r, session, parts[1])
			return
		}
		if len(parts) == 2 && r.Method == http.MethodPost {
			id := parts[1]
			var req struct {
				Status string `json:"status"`
			}
			_ = readJSON(r, &req)
			var updated Plugin
			err := a.store.Mutate(func(data *AppData) error {
				for i := range data.Plugins {
					if data.Plugins[i].ID == id {
						data.Plugins[i].Status = fallback(req.Status, "enabled")
						updated = data.Plugins[i]
						addAudit(data, session.User.Username, "toggle", "plugin", 0, id, clientIP(r))
						return nil
					}
				}
				return fmt.Errorf("插件不存在")
			})
			a.respondMutation(w, err, updated, "system.plugin.update")
			return
		}
	case "updates":
		if len(parts) == 1 && r.Method == http.MethodPost {
			var req UpdatePackage
			if err := readJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid update package payload")
				return
			}
			var saved UpdatePackage
			err := a.store.Mutate(func(data *AppData) error {
				rawCreatedAt := req.CreatedAt
				normalized, err := normalizeUpdatePackage(req, session.User.DisplayName)
				if err != nil {
					return err
				}
				for i := range data.Updates {
					if data.Updates[i].ID == normalized.ID || sameUpdatePackage(data.Updates[i], normalized) {
						normalized.ID = data.Updates[i].ID
						normalized = mergeUpdatePackageArtifact(normalized, data.Updates[i])
						if rawCreatedAt == "" {
							normalized.CreatedAt = data.Updates[i].CreatedAt
						}
						normalized.DownloadCount = data.Updates[i].DownloadCount
						normalized.LastDownloadedAt = data.Updates[i].LastDownloadedAt
						normalized.AppliedBy = data.Updates[i].AppliedBy
						normalized.AppliedAt = data.Updates[i].AppliedAt
						data.Updates[i] = normalized
						saved = normalized
						addAudit(data, session.User.Username, "update", "update_package", normalized.ID, normalized.Component+" "+normalized.Version, clientIP(r))
						return nil
					}
				}
				normalized.ID = nextID(data, "update")
				data.Updates = append(data.Updates, normalized)
				saved = normalized
				addAudit(data, session.User.Username, "publish", "update_package", normalized.ID, normalized.Component+" "+normalized.Version, clientIP(r))
				return nil
			})
			a.respondMutation(w, err, sanitizeUpdatePackageForResponse(saved), "system.update.published")
			return
		}
		if len(parts) == 3 && parts[2] == "download" && r.Method == http.MethodGet {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			var download UpdatePackageDownload
			err := a.store.Mutate(func(data *AppData) error {
				for i := range data.Updates {
					if data.Updates[i].ID != id {
						continue
					}
					download = buildUpdatePackageDownload(data.Updates[i])
					if !download.Verified {
						return fmt.Errorf("更新包验签失败")
					}
					data.Updates[i].DownloadCount++
					data.Updates[i].LastDownloadedAt = nowString()
					download.Package = data.Updates[i]
					download.Package = sanitizeUpdatePackageForResponse(download.Package)
					addAudit(data, session.User.Username, "download", "update_package", id, data.Updates[i].Component+" "+data.Updates[i].Version, clientIP(r))
					return nil
				}
				return fmt.Errorf("更新包不存在")
			})
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, download)
			return
		}
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, sanitizeUpdatePackagesForResponse(data.Updates))
			return
		}
		if len(parts) == 3 && (parts[2] == "apply" || parts[2] == "rollback") && r.Method == http.MethodPost {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			var updated UpdatePackage
			err := a.store.Mutate(func(data *AppData) error {
				for i := range data.Updates {
					if data.Updates[i].ID != id {
						continue
					}
					if !updatePackageVerified(data.Updates[i]) {
						return fmt.Errorf("更新包验签失败")
					}
					if parts[2] == "apply" {
						for j := range data.Updates {
							if data.Updates[j].ID != id && data.Updates[j].Component == data.Updates[i].Component && data.Updates[j].Status == "installed" {
								data.Updates[j].Status = "superseded"
							}
						}
						data.Updates[i].Status = "installed"
						data.Updates[i].AppliedBy = session.User.DisplayName
						data.Updates[i].AppliedAt = nowString()
						updated = data.Updates[i]
						addAudit(data, session.User.Username, "apply", "update_package", id, updated.Version, clientIP(r))
						return nil
					}
					data.Updates[i].Status = "rolled_back"
					data.Updates[i].AppliedBy = session.User.DisplayName
					data.Updates[i].AppliedAt = nowString()
					updated = data.Updates[i]
					addAudit(data, session.User.Username, "rollback", "update_package", id, updated.RollbackVersion, clientIP(r))
					return nil
				}
				return fmt.Errorf("更新包不存在")
			})
			a.respondMutation(w, err, updated, "system.update.changed")
			return
		}
		writeError(w, http.StatusNotFound, "unknown update route")
	case "audit":
		writeJSON(w, http.StatusOK, data.AuditLogs)
	case "org":
		a.systemOrg(w, r, session, parts[1:])
	case "mfa":
		a.systemMFA(w, r, session, parts[1:])
	case "sso":
		a.systemSSO(w, r, session, parts[1:])
	case "scim":
		a.systemSCIM(w, r, session, parts[1:])
	case "field-policies":
		a.systemFieldPolicies(w, r, session, parts[1:])
	case "approval-flows":
		a.systemApprovalFlows(w, r, session, parts[1:])
	case "dictionaries":
		a.systemDictionaries(w, r, session, parts[1:])
	case "security":
		securityData := scopedData(data, session.User)
		sessionPolicy := buildSessionPolicy(securityData)
		activeSessions := a.activeSessionSummaries(sessionPolicy)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"policies":          securityData.SecurityPolicies,
			"fieldPolicies":     securityData.FieldPolicies,
			"deviceCredentials": publicDeviceCredentials(securityData.DeviceCredentials),
			"users":             publicUsers(securityData.Users),
			"ssoProviders":      publicOIDCProviders(securityData.OIDCProviders),
			"scimProviders":     publicSCIMProviders(securityData.SCIMProviders),
			"scimEvents":        securityData.SCIMEvents,
			"sessionPolicy":     sessionPolicy,
			"sessions":          activeSessions,
			"report":            buildSecurityReport(securityData, activeSessions, sessionPolicy),
		})
	case "runtime":
		writeJSON(w, http.StatusOK, a.runtimeStatus())
	case "map-config":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, a.runtime.MapConfig())
	case "gateway":
		a.systemGateway(w, r, session, parts[1:])
	case "backups":
		a.systemBackups(w, r, session, parts[1:])
	default:
		writeError(w, http.StatusNotFound, "unknown system resource")
	}
}

func (a *App) systemFieldPolicies(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, a.mustSnapshot().FieldPolicies)
		return
	}
	if len(parts) == 0 && r.Method == http.MethodPost {
		var item FieldPolicy
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid field policy")
			return
		}
		err := a.store.Mutate(func(data *AppData) error {
			if item.RoleCode == "" || item.Resource == "" || item.Field == "" {
				return fmt.Errorf("字段策略必须包含角色、资源和字段")
			}
			item.ID = nextID(data, "fieldPolicy")
			item.Mask = fallback(item.Mask, "phone")
			item.Enabled = true
			data.FieldPolicies = append(data.FieldPolicies, item)
			addAudit(data, session.User.Username, "create", "field_policy", item.ID, item.RoleCode+"/"+item.Resource+"/"+item.Field, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, item, "system.field_policy.created")
		return
	}
	if len(parts) == 2 && parts[1] == "toggle" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Enabled bool `json:"enabled"`
		}
		_ = readJSON(r, &req)
		var updated FieldPolicy
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.FieldPolicies {
				if data.FieldPolicies[i].ID != id {
					continue
				}
				data.FieldPolicies[i].Enabled = req.Enabled
				updated = data.FieldPolicies[i]
				addAudit(data, session.User.Username, "toggle", "field_policy", id, updated.RoleCode+"/"+updated.Resource+"/"+updated.Field, clientIP(r))
				return nil
			}
			return fmt.Errorf("字段策略不存在")
		})
		a.respondMutation(w, err, updated, "system.field_policy.updated")
		return
	}
	writeError(w, http.StatusNotFound, "unknown field policy route")
}

func (a *App) systemBackups(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "drills" {
		a.backupDrills(w, r, session)
		return
	}
	if len(parts) == 0 {
		if r.Method == http.MethodGet {
			backups, err := a.backups.List()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, backups)
			return
		}
		if r.Method == http.MethodPost {
			data := a.mustSnapshot()
			info, err := a.backups.Create(data)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			_ = a.store.Mutate(func(data *AppData) error {
				addAudit(data, session.User.Username, "create", "backup", 0, info.Name, clientIP(r))
				return nil
			})
			a.respondMutation(w, nil, info, "system.backup.created")
			return
		}
	}
	if len(parts) == 2 && parts[1] == "restore" && r.Method == http.MethodPost {
		backupName := parts[0]
		restored, err := a.backups.Restore(backupName)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		err = a.store.Mutate(func(data *AppData) error {
			*data = restored
			addAudit(data, session.User.Username, "restore", "backup", 0, backupName, clientIP(r))
			return nil
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.respondMutation(w, nil, map[string]string{"restored": backupName}, "system.backup.restored")
		return
	}
	writeError(w, http.StatusNotFound, "unknown backup route")
}

func (a *App) simulate(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "tick" && r.Method == http.MethodPost {
		var events []VehicleLocationEvent
		err := a.store.Mutate(func(data *AppData) error {
			for _, latest := range data.LatestLocations {
				if latest.OnlineStatus != "online" {
					continue
				}
				latest.Longitude += 0.0012
				latest.Latitude -= 0.0008
				latest.Speed = 28 + float64(latest.VehicleID*3)
				event := VehicleLocationEvent{
					ID: nextID(data, "location"), VehicleID: latest.VehicleID, PlateNo: latest.PlateNo,
					Longitude: latest.Longitude, Latitude: latest.Latitude, Speed: latest.Speed,
					Direction: 150, OnlineStatus: "online", Address: inferAddress(*data, latest.Longitude, latest.Latitude),
					LocationTime: nowString(), ReceiveTime: nowString(), SourceType: "simulator",
				}
				data.Locations = append(data.Locations, event)
				upsertLatestLocation(data, latest)
				events = append(events, event)
			}
			addAudit(data, session.User.Username, "simulate", "vehicle_location", 0, "tick", clientIP(r))
			return nil
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("vehicle.location.update", events)
		writeJSON(w, http.StatusOK, events)
		return
	}
	writeError(w, http.StatusNotFound, "unknown simulate route")
}

func number(prefix string, id int64) string {
	return fmt.Sprintf("%s%s%04d", prefix, time.Now().Format("20060102"), id)
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func nonZero(value, fallbackValue float64) float64 {
	if value == 0 {
		return fallbackValue
	}
	return value
}

func nonZeroInt(value, fallbackValue int64) int64 {
	if value == 0 {
		return fallbackValue
	}
	return value
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return round(float64(part) / float64(total) * 100)
}
