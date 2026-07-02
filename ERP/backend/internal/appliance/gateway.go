package appliance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type gatewayOverview struct {
	Routes      []GatewayRoute    `json:"routes"`
	Events      []GatewayEvent    `json:"events"`
	NginxConfig string            `json:"nginxConfig"`
	ReloadPlan  GatewayReloadPlan `json:"reloadPlan"`
}

type GatewayReloadPlan struct {
	ReloadRequired  bool                `json:"reloadRequired"`
	Valid           bool                `json:"valid"`
	ConfigHash      string              `json:"configHash"`
	ConfigPath      string              `json:"configPath"`
	ReloadCommand   string              `json:"reloadCommand"`
	GeneratedAt     string              `json:"generatedAt"`
	LastReloadAt    string              `json:"lastReloadAt"`
	ReloadedAt      string              `json:"reloadedAt,omitempty"`
	ActiveRoutes    int                 `json:"activeRoutes"`
	DrainingRoutes  int                 `json:"drainingRoutes"`
	AutomaticReload bool                `json:"automaticReload"`
	Steps           []string            `json:"steps"`
	Errors          []string            `json:"errors"`
	DrainChecks     []GatewayDrainCheck `json:"drainChecks"`
}

type GatewayDrainCheck struct {
	RouteID          int64  `json:"routeId"`
	RouteName        string `json:"routeName"`
	PathPrefix       string `json:"pathPrefix"`
	DrainEnabled     bool   `json:"drainEnabled"`
	DrainUntil       string `json:"drainUntil"`
	Active           bool   `json:"active"`
	RemainingSeconds int64  `json:"remainingSeconds"`
	ProbePath        string `json:"probePath"`
	ExpectedHeader   string `json:"expectedHeader"`
	Status           string `json:"status"`
}

func (a *App) systemGateway(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		if r.Method == http.MethodGet {
			data := a.mustSnapshot()
			writeJSON(w, http.StatusOK, gatewayOverview{
				Routes: data.GatewayRoutes, Events: data.GatewayEvents,
				NginxConfig: renderGatewayNginx(data.GatewayRoutes),
				ReloadPlan:  buildGatewayReloadPlan(data.GatewayRoutes, data.GatewayEvents),
			})
			return
		}
		if r.Method == http.MethodPost {
			a.upsertGatewayRoute(w, r, session)
			return
		}
	}
	if len(parts) == 1 && parts[0] == "nginx" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"nginxConfig": renderGatewayNginx(a.mustSnapshot().GatewayRoutes)})
		return
	}
	if len(parts) == 1 && parts[0] == "reload-plan" && r.Method == http.MethodGet {
		data := a.mustSnapshot()
		writeJSON(w, http.StatusOK, buildGatewayReloadPlan(data.GatewayRoutes, data.GatewayEvents))
		return
	}
	if len(parts) == 1 && parts[0] == "reload" && r.Method == http.MethodPost {
		a.recordGatewayReload(w, r, session)
		return
	}
	if len(parts) == 2 && parts[0] == "routes" && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.deleteGatewayRoute(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "routes" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		switch parts[2] {
		case "canary":
			a.updateGatewayCanary(w, r, session, id)
		case "drain":
			a.updateGatewayDrain(w, r, session, id)
		case "status":
			a.updateGatewayStatus(w, r, session, id)
		default:
			writeError(w, http.StatusNotFound, "unknown gateway route action")
		}
		return
	}
	writeError(w, http.StatusNotFound, "unknown gateway route")
}

func (a *App) recordGatewayReload(w http.ResponseWriter, r *http.Request, session Session) {
	var plan GatewayReloadPlan
	err := a.store.Mutate(func(data *AppData) error {
		plan = buildGatewayReloadPlan(data.GatewayRoutes, data.GatewayEvents)
		if !plan.Valid {
			return fmt.Errorf("网关配置校验失败: %s", strings.Join(plan.Errors, "; "))
		}
		plan.ReloadedAt = nowString()
		addGatewaySystemEvent(data, session.User.Username, "reload", fmt.Sprintf("配置 hash %s，%d 条排空检查", plan.ConfigHash, len(plan.DrainChecks)), clientIP(r))
		plan = buildGatewayReloadPlan(data.GatewayRoutes, data.GatewayEvents)
		plan.ReloadedAt = nowString()
		return nil
	})
	a.respondMutation(w, err, plan, "system.gateway.reload")
}

func (a *App) upsertGatewayRoute(w http.ResponseWriter, r *http.Request, session Session) {
	var req GatewayRoute
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid gateway route payload")
		return
	}
	var updated GatewayRoute
	err := a.store.Mutate(func(data *AppData) error {
		normalized, err := normalizeGatewayRoute(req)
		if err != nil {
			return err
		}
		for i := range data.GatewayRoutes {
			if data.GatewayRoutes[i].ID == normalized.ID || data.GatewayRoutes[i].PathPrefix == normalized.PathPrefix {
				normalized.ID = data.GatewayRoutes[i].ID
				normalized.UpdatedAt = nowString()
				data.GatewayRoutes[i] = normalized
				updated = normalized
				addGatewayEvent(data, session.User.Username, updated, "upsert", "更新网关路由", clientIP(r))
				return nil
			}
		}
		normalized.ID = nextID(data, "gatewayRoute")
		normalized.UpdatedAt = nowString()
		data.GatewayRoutes = append(data.GatewayRoutes, normalized)
		updated = normalized
		addGatewayEvent(data, session.User.Username, updated, "create", "新增网关路由", clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, "system.gateway.route")
}

func (a *App) updateGatewayCanary(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		CanaryPercent  int    `json:"canaryPercent"`
		CanaryUpstream string `json:"canaryUpstream"`
	}
	_ = readJSON(r, &req)
	var updated GatewayRoute
	err := a.store.Mutate(func(data *AppData) error {
		idx := gatewayRouteIndex(data.GatewayRoutes, id)
		if idx < 0 {
			return fmt.Errorf("网关路由不存在")
		}
		if req.CanaryPercent < 0 || req.CanaryPercent > 100 {
			return fmt.Errorf("灰度比例必须在 0-100")
		}
		data.GatewayRoutes[idx].CanaryPercent = req.CanaryPercent
		if strings.TrimSpace(req.CanaryUpstream) != "" {
			data.GatewayRoutes[idx].CanaryUpstream = strings.TrimSpace(req.CanaryUpstream)
		}
		data.GatewayRoutes[idx].UpdatedAt = nowString()
		updated = data.GatewayRoutes[idx]
		addGatewayEvent(data, session.User.Username, updated, "canary", fmt.Sprintf("灰度比例 %d%%", req.CanaryPercent), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, "system.gateway.canary")
}

func (a *App) updateGatewayDrain(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Enabled    bool   `json:"enabled"`
		DurationMs int    `json:"durationMs"`
		DrainUntil string `json:"drainUntil"`
	}
	_ = readJSON(r, &req)
	var updated GatewayRoute
	err := a.store.Mutate(func(data *AppData) error {
		idx := gatewayRouteIndex(data.GatewayRoutes, id)
		if idx < 0 {
			return fmt.Errorf("网关路由不存在")
		}
		data.GatewayRoutes[idx].DrainEnabled = req.Enabled
		if req.Enabled {
			if req.DrainUntil != "" {
				data.GatewayRoutes[idx].DrainUntil = req.DrainUntil
			} else {
				duration := time.Duration(nonZeroInt(int64(req.DurationMs), int64(300000))) * time.Millisecond
				data.GatewayRoutes[idx].DrainUntil = time.Now().Add(duration).Format("2006-01-02 15:04:05")
			}
		} else {
			data.GatewayRoutes[idx].DrainUntil = ""
		}
		data.GatewayRoutes[idx].UpdatedAt = nowString()
		updated = data.GatewayRoutes[idx]
		addGatewayEvent(data, session.User.Username, updated, "drain", fmt.Sprintf("连接排空 %v", req.Enabled), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, "system.gateway.drain")
}

func (a *App) updateGatewayStatus(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Status string `json:"status"`
	}
	_ = readJSON(r, &req)
	var updated GatewayRoute
	topic := "system.gateway.status"
	err := a.store.Mutate(func(data *AppData) error {
		idx := gatewayRouteIndex(data.GatewayRoutes, id)
		if idx < 0 {
			return fmt.Errorf("网关路由不存在")
		}
		status := fallback(strings.TrimSpace(req.Status), "active")
		if status != "active" && status != "disabled" {
			return fmt.Errorf("网关路由状态无效")
		}
		if hasPendingWorkflowForResource(*data, "gateway_route", id) {
			updated = data.GatewayRoutes[idx]
			topic = "system.gateway.status_requested"
			return nil
		}
		_, instances, err := publishGatewayRouteStatusWorkflow(data, data.GatewayRoutes[idx], status, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			updated = data.GatewayRoutes[idx]
			topic = "system.gateway.status_requested"
			addAudit(data, session.User.Username, "request_status", "gateway_route", id, updated.PathPrefix+"/"+status, clientIP(r))
			return nil
		}
		next, err := applyGatewayRouteStatusLocked(data, id, status, session.User.Username, clientIP(r))
		if err != nil {
			return err
		}
		updated = next
		return nil
	})
	a.respondMutation(w, err, updated, topic)
}

func (a *App) deleteGatewayRoute(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var deleted GatewayRoute
	err := a.store.Mutate(func(data *AppData) error {
		idx := gatewayRouteIndex(data.GatewayRoutes, id)
		if idx < 0 {
			return fmt.Errorf("网关路由不存在")
		}
		route := data.GatewayRoutes[idx]
		if fallback(strings.TrimSpace(route.Status), "active") == "active" {
			return fmt.Errorf("启用中的网关路由不能删除")
		}
		if hasPendingWorkflowForResource(*data, "gateway_route", id) {
			return fmt.Errorf("网关路由存在待审批流程，不能删除")
		}
		data.GatewayRoutes = append(data.GatewayRoutes[:idx], data.GatewayRoutes[idx+1:]...)
		deleted = route
		addGatewayEvent(data, session.User.Username, deleted, "delete", "删除网关路由", clientIP(r))
		return nil
	})
	a.respondMutation(w, err, deleted, "system.gateway.route.deleted")
}

func publishGatewayRouteStatusWorkflow(data *AppData, item GatewayRoute, targetStatus string, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "gateway_route.status_change_requested",
		Source:     "system",
		Resource:   "gateway_route",
		ResourceID: item.ID,
		ResourceNo: item.PathPrefix,
		Title:      "网关路由状态变更 " + item.PathPrefix,
		Actor:      actor,
		Reason:     "网关路由状态变更审批",
		Variables: map[string]string{
			"targetStatus":   targetStatus,
			"currentStatus":  item.Status,
			"name":           item.Name,
			"pathPrefix":     item.PathPrefix,
			"stableUpstream": item.StableUpstream,
			"canaryUpstream": item.CanaryUpstream,
			"canaryPercent":  strconv.Itoa(item.CanaryPercent),
			"drainEnabled":   strconv.FormatBool(item.DrainEnabled),
			"readTimeoutSec": strconv.Itoa(item.ReadTimeoutSec),
		},
	})
}

func applyGatewayRouteStatusLocked(data *AppData, id int64, status string, actor string, ip string) (GatewayRoute, error) {
	status = fallback(strings.TrimSpace(status), "active")
	if status != "active" && status != "disabled" {
		return GatewayRoute{}, fmt.Errorf("网关路由状态无效")
	}
	idx := gatewayRouteIndex(data.GatewayRoutes, id)
	if idx < 0 {
		return GatewayRoute{}, fmt.Errorf("网关路由不存在")
	}
	data.GatewayRoutes[idx].Status = status
	data.GatewayRoutes[idx].UpdatedAt = nowString()
	updated := data.GatewayRoutes[idx]
	addGatewayEvent(data, actor, updated, "status", status, ip)
	return updated, nil
}

func normalizeGatewayRoute(route GatewayRoute) (GatewayRoute, error) {
	route.Name = strings.TrimSpace(route.Name)
	route.PathPrefix = strings.TrimSpace(route.PathPrefix)
	route.StableUpstream = strings.TrimSpace(route.StableUpstream)
	route.CanaryUpstream = strings.TrimSpace(route.CanaryUpstream)
	if route.Name == "" || route.PathPrefix == "" || route.StableUpstream == "" {
		return route, fmt.Errorf("网关路由名称、路径和稳定 upstream 必填")
	}
	if !strings.HasPrefix(route.PathPrefix, "/") {
		return route, fmt.Errorf("网关路径必须以 / 开头")
	}
	if route.CanaryPercent < 0 || route.CanaryPercent > 100 {
		return route, fmt.Errorf("灰度比例必须在 0-100")
	}
	route.Status = fallback(route.Status, "active")
	if route.ReadTimeoutSec <= 0 {
		route.ReadTimeoutSec = 120
	}
	return route, nil
}

func gatewayRouteIndex(routes []GatewayRoute, id int64) int {
	for i := range routes {
		if routes[i].ID == id {
			return i
		}
	}
	return -1
}

func addGatewayEvent(data *AppData, actor string, route GatewayRoute, action, detail, ip string) {
	id := nextID(data, "gatewayEvent")
	event := GatewayEvent{
		ID: id, EventNo: number("GW", id), RouteID: route.ID, RouteName: route.Name,
		Action: action, Detail: detail, Actor: actor, CreatedAt: nowString(),
	}
	data.GatewayEvents = append(data.GatewayEvents, event)
	if len(data.GatewayEvents) > 200 {
		data.GatewayEvents = data.GatewayEvents[len(data.GatewayEvents)-200:]
	}
	addAudit(data, actor, action, "gateway_route", route.ID, route.PathPrefix+"/"+detail, ip)
}

func addGatewaySystemEvent(data *AppData, actor, action, detail, ip string) {
	id := nextID(data, "gatewayEvent")
	event := GatewayEvent{
		ID: id, EventNo: number("GW", id), RouteID: 0, RouteName: "API Gateway",
		Action: action, Detail: detail, Actor: actor, CreatedAt: nowString(),
	}
	data.GatewayEvents = append(data.GatewayEvents, event)
	if len(data.GatewayEvents) > 200 {
		data.GatewayEvents = data.GatewayEvents[len(data.GatewayEvents)-200:]
	}
	addAudit(data, actor, action, "gateway", 0, detail, ip)
}

func buildGatewayReloadPlan(routes []GatewayRoute, events []GatewayEvent) GatewayReloadPlan {
	config := renderGatewayNginx(routes)
	hash := sha256.Sum256([]byte(config))
	now := time.Now()
	lastReloadAt := latestGatewayReloadAt(events)
	plan := GatewayReloadPlan{
		ReloadRequired:  gatewayReloadRequired(routes, lastReloadAt),
		Valid:           true,
		ConfigHash:      "sha256:" + hex.EncodeToString(hash[:]),
		ConfigPath:      fallback(os.Getenv("CBMP_NGINX_CONFIG_PATH"), "/etc/nginx/conf.d/cbmp-gateway.conf"),
		ReloadCommand:   fallback(os.Getenv("CBMP_NGINX_RELOAD_COMMAND"), "nginx -t && nginx -s reload"),
		GeneratedAt:     now.Format("2006-01-02 15:04:05"),
		LastReloadAt:    lastReloadAt,
		AutomaticReload: os.Getenv("CBMP_GATEWAY_AUTO_RELOAD") == "1",
		Steps: []string{
			"写入生成的 Nginx 配置片段",
			"执行配置语法校验",
			"触发平滑 reload",
			"检查 X-CBMP-Drain 和 X-CBMP-Drain-Until 响应头",
		},
	}
	for _, route := range routes {
		if route.Status != "active" {
			continue
		}
		plan.ActiveRoutes++
		if strings.TrimSpace(route.StableUpstream) == "" {
			plan.Errors = append(plan.Errors, route.Name+" 缺少稳定 upstream")
		}
		if route.CanaryPercent > 0 && strings.TrimSpace(route.CanaryUpstream) == "" {
			plan.Errors = append(plan.Errors, route.Name+" 灰度比例大于 0 但缺少灰度 upstream")
		}
		check := gatewayDrainCheck(route, now)
		if check.DrainEnabled {
			plan.DrainingRoutes++
		}
		plan.DrainChecks = append(plan.DrainChecks, check)
		if check.Status == "invalid" {
			plan.Errors = append(plan.Errors, route.Name+" 排空截止时间无效")
		}
	}
	if plan.ActiveRoutes == 0 {
		plan.Errors = append(plan.Errors, "没有 active 网关路由")
	}
	plan.Valid = len(plan.Errors) == 0
	return plan
}

func latestGatewayReloadAt(events []GatewayEvent) string {
	latest := ""
	for _, event := range events {
		if event.Action == "reload" && event.CreatedAt > latest {
			latest = event.CreatedAt
		}
	}
	return latest
}

func gatewayReloadRequired(routes []GatewayRoute, lastReloadAt string) bool {
	if lastReloadAt == "" {
		return len(routes) > 0
	}
	for _, route := range routes {
		if route.UpdatedAt > lastReloadAt {
			return true
		}
	}
	return false
}

func gatewayDrainCheck(route GatewayRoute, now time.Time) GatewayDrainCheck {
	check := GatewayDrainCheck{
		RouteID:        route.ID,
		RouteName:      route.Name,
		PathPrefix:     route.PathPrefix,
		DrainEnabled:   route.DrainEnabled,
		DrainUntil:     route.DrainUntil,
		ProbePath:      route.PathPrefix,
		ExpectedHeader: "X-CBMP-Drain",
		Status:         "inactive",
	}
	if !route.DrainEnabled {
		return check
	}
	check.Status = "active"
	check.Active = true
	if route.DrainUntil == "" {
		return check
	}
	until, err := time.Parse("2006-01-02 15:04:05", route.DrainUntil)
	if err != nil {
		check.Status = "invalid"
		check.Active = false
		return check
	}
	remaining := int64(until.Sub(now).Seconds())
	if remaining <= 0 {
		check.Status = "expired"
		check.Active = false
		return check
	}
	check.RemainingSeconds = remaining
	return check
}

func renderGatewayNginx(routes []GatewayRoute) string {
	var builder strings.Builder
	builder.WriteString("# Generated by CBMP gateway center. Review and reload Nginx after applying.\n")
	for _, route := range routes {
		if route.Status != "active" {
			continue
		}
		name := gatewayRouteName(route)
		builder.WriteString(fmt.Sprintf("upstream %s_stable { server %s; keepalive 32; }\n", name, route.StableUpstream))
		if route.CanaryUpstream != "" {
			builder.WriteString(fmt.Sprintf("upstream %s_canary { server %s; keepalive 16; }\n", name, route.CanaryUpstream))
			builder.WriteString(fmt.Sprintf("split_clients \"$request_id\" $%s_pool { %d%% %s_canary; * %s_stable; }\n", name, route.CanaryPercent, name, name))
		} else {
			builder.WriteString(fmt.Sprintf("map $request_id $%s_pool { default %s_stable; }\n", name, name))
		}
	}
	builder.WriteString("server {\n  listen 80;\n  server_name _;\n  client_max_body_size 50m;\n")
	for _, route := range routes {
		if route.Status != "active" {
			continue
		}
		name := gatewayRouteName(route)
		builder.WriteString(fmt.Sprintf("  location %s {\n", route.PathPrefix))
		if route.DrainEnabled {
			builder.WriteString("    add_header X-CBMP-Drain true always;\n")
			if route.DrainUntil != "" {
				builder.WriteString(fmt.Sprintf("    add_header X-CBMP-Drain-Until \"%s\" always;\n", route.DrainUntil))
			}
		}
		builder.WriteString(fmt.Sprintf("    proxy_pass http://$%s_pool;\n", name))
		builder.WriteString("    proxy_http_version 1.1;\n    proxy_set_header Connection \"\";\n    proxy_set_header Host $host;\n    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		builder.WriteString(fmt.Sprintf("    proxy_read_timeout %ds;\n", route.ReadTimeoutSec))
		if route.PathPrefix == "/api/events" {
			builder.WriteString("    proxy_buffering off;\n    proxy_cache off;\n")
		}
		builder.WriteString("  }\n")
	}
	builder.WriteString("  location / { try_files $uri $uri/ /index.html; }\n}\n")
	return builder.String()
}

func gatewayRouteName(route GatewayRoute) string {
	name := strings.Trim(route.PathPrefix, "/")
	if name == "" {
		name = "root"
	}
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return "cbmp_" + name
}
