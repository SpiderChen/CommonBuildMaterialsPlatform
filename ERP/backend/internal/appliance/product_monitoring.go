//go:build legacy_product_ops

package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

type productMonitoringReportRequest struct {
	IntegrationToken string            `json:"integrationToken"`
	IntegrationCode  string            `json:"integrationCode"`
	Provider         string            `json:"provider"`
	InstanceID       int64             `json:"instanceId"`
	Watermark        string            `json:"watermark"`
	Source           string            `json:"source"`
	Component        string            `json:"component"`
	Metric           string            `json:"metric"`
	Value            float64           `json:"value"`
	Severity         string            `json:"severity"`
	Status           string            `json:"status"`
	Title            string            `json:"title"`
	Message          string            `json:"message"`
	Labels           map[string]string `json:"labels"`
	OccurredAt       string            `json:"occurredAt"`
}

type productMonitoringReportResponse struct {
	Accepted    bool                         `json:"accepted"`
	Integration ProductMonitoringIntegration `json:"integration"`
	Instance    ProductInstance              `json:"instance"`
	Event       ProductMonitoringEvent       `json:"event"`
	Alert       *SystemAlert                 `json:"alert,omitempty"`
}

func (a *App) productOpsMonitoring(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "integrations" && r.Method == http.MethodPost {
		var req ProductMonitoringIntegration
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid monitoring integration payload")
			return
		}
		var saved ProductMonitoringIntegration
		err := a.store.Mutate(func(data *AppData) error {
			req.Name = strings.TrimSpace(req.Name)
			req.Code = strings.TrimSpace(req.Code)
			if req.Name == "" || req.Code == "" {
				return fmt.Errorf("监控接入名称和编码不能为空")
			}
			now := nowString()
			req.Provider = fallback(strings.ToLower(strings.TrimSpace(req.Provider)), "custom")
			req.Endpoint = fallback(strings.TrimSpace(req.Endpoint), "/api/product-ops/monitoring/report")
			req.Status = fallback(strings.TrimSpace(req.Status), "active")
			req.Token = fallback(strings.TrimSpace(req.Token), "mon-"+tokenString())
			for i := range data.ProductMonitoringIntegrations {
				if data.ProductMonitoringIntegrations[i].ID == req.ID || data.ProductMonitoringIntegrations[i].Code == req.Code {
					req.ID = data.ProductMonitoringIntegrations[i].ID
					req.IntegrationNo = fallback(req.IntegrationNo, data.ProductMonitoringIntegrations[i].IntegrationNo)
					req.CreatedBy = fallback(req.CreatedBy, data.ProductMonitoringIntegrations[i].CreatedBy)
					req.CreatedAt = fallback(req.CreatedAt, data.ProductMonitoringIntegrations[i].CreatedAt)
					req.LastEventAt = fallback(req.LastEventAt, data.ProductMonitoringIntegrations[i].LastEventAt)
					data.ProductMonitoringIntegrations[i] = req
					saved = req
					addAudit(data, session.User.Username, "update", "monitoring_integration", req.ID, req.Code+" "+req.Name, clientIP(r))
					return nil
				}
			}
			req.ID = nextID(data, "monitoringIntegration")
			req.IntegrationNo = number("MI", req.ID)
			req.CreatedBy = session.User.DisplayName
			req.CreatedAt = now
			data.ProductMonitoringIntegrations = append(data.ProductMonitoringIntegrations, req)
			saved = req
			addAudit(data, session.User.Username, "create", "monitoring_integration", req.ID, req.Code+" "+req.Name, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "product_ops.monitoring.integration.saved")
		return
	}
	if len(parts) == 1 && parts[0] == "rules" && r.Method == http.MethodPost {
		var req ProductAlertRule
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid product alert rule payload")
			return
		}
		var saved ProductAlertRule
		err := a.store.Mutate(func(data *AppData) error {
			req.Name = strings.TrimSpace(req.Name)
			req.Metric = strings.TrimSpace(req.Metric)
			if req.Name == "" || req.Metric == "" {
				return fmt.Errorf("告警规则名称和指标不能为空")
			}
			req.Source = fallback(strings.ToLower(strings.TrimSpace(req.Source)), "custom")
			req.Component = fallback(strings.ToLower(strings.TrimSpace(req.Component)), "server")
			req.Operator = fallback(strings.TrimSpace(req.Operator), ">=")
			req.Severity = fallback(normalizeTelemetrySeverity(req.Severity), "warning")
			req.Status = fallback(strings.TrimSpace(req.Status), "active")
			if len(req.NotifyChannels) == 0 {
				req.NotifyChannels = []string{"sse"}
			}
			for i := range data.ProductAlertRules {
				if data.ProductAlertRules[i].ID == req.ID || (req.RuleNo != "" && data.ProductAlertRules[i].RuleNo == req.RuleNo) {
					req.ID = data.ProductAlertRules[i].ID
					req.RuleNo = fallback(req.RuleNo, data.ProductAlertRules[i].RuleNo)
					req.CreatedBy = fallback(req.CreatedBy, data.ProductAlertRules[i].CreatedBy)
					req.CreatedAt = fallback(req.CreatedAt, data.ProductAlertRules[i].CreatedAt)
					data.ProductAlertRules[i] = req
					saved = req
					addAudit(data, session.User.Username, "update", "product_alert_rule", req.ID, req.RuleNo+" "+req.Name, clientIP(r))
					return nil
				}
			}
			req.ID = nextID(data, "productAlertRule")
			req.RuleNo = number("AR", req.ID)
			req.CreatedBy = session.User.DisplayName
			req.CreatedAt = nowString()
			data.ProductAlertRules = append(data.ProductAlertRules, req)
			saved = req
			addAudit(data, session.User.Username, "create", "product_alert_rule", req.ID, req.RuleNo+" "+req.Name, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "product_ops.monitoring.rule.saved")
		return
	}
	writeError(w, http.StatusNotFound, "unknown product monitoring route")
}

func (a *App) productOpsMonitoringReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req productMonitoringReportRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid monitoring report payload")
		return
	}
	token := strings.TrimSpace(r.Header.Get("X-CBMP-Monitoring-Token"))
	if token == "" {
		token = strings.TrimSpace(req.IntegrationToken)
	}
	if token == "" {
		unauthorized(w)
		return
	}
	var response productMonitoringReportResponse
	err := a.store.Mutate(func(data *AppData) error {
		integrationIndex := productMonitoringIntegrationIndex(*data, token, req.IntegrationCode)
		if integrationIndex < 0 {
			return fmt.Errorf("monitoring integration invalid")
		}
		integration := &data.ProductMonitoringIntegrations[integrationIndex]
		if strings.TrimSpace(req.IntegrationCode) != "" && integration.Code != strings.TrimSpace(req.IntegrationCode) {
			return fmt.Errorf("monitoring integration code mismatch")
		}
		if integration.Status != "active" {
			return fmt.Errorf("monitoring integration disabled")
		}
		instanceIndex := productInstanceIndexForMonitoring(*data, req.InstanceID, req.Watermark, req.Labels)
		if instanceIndex < 0 {
			return fmt.Errorf("客户实例不存在")
		}
		instance := &data.ProductInstances[instanceIndex]
		now := nowString()
		source := fallback(strings.ToLower(strings.TrimSpace(req.Source)), fallback(integration.Provider, req.Provider))
		if source == "" {
			source = "custom"
		}
		eventID := nextID(data, "monitoringEvent")
		event := ProductMonitoringEvent{
			ID:              eventID,
			EventNo:         number("ME", eventID),
			IntegrationID:   integration.ID,
			IntegrationName: integration.Name,
			Provider:        fallback(strings.ToLower(strings.TrimSpace(req.Provider)), integration.Provider),
			InstanceID:      instance.ID,
			CustomerName:    instance.CustomerName,
			Watermark:       fallback(strings.TrimSpace(req.Watermark), instance.Watermark),
			Source:          source,
			Component:       fallback(strings.ToLower(strings.TrimSpace(req.Component)), "server"),
			Metric:          strings.TrimSpace(req.Metric),
			Value:           req.Value,
			Severity:        normalizeTelemetrySeverity(req.Severity),
			Status:          fallback(strings.ToLower(strings.TrimSpace(req.Status)), "firing"),
			Title:           strings.TrimSpace(req.Title),
			Message:         strings.TrimSpace(req.Message),
			Labels:          req.Labels,
			OccurredAt:      fallback(strings.TrimSpace(req.OccurredAt), now),
			ReceivedAt:      now,
			SourceIP:        clientIP(r),
		}
		severity, issue, ruleNo := evaluateMonitoringEvent(*data, event)
		event.AlertRaised = issue != ""
		event.MatchedRuleNo = ruleNo
		if issue != "" {
			event.Severity = severity
		}
		data.ProductMonitoringEvents = append(data.ProductMonitoringEvents, event)
		trimProductMonitoringEvents(data)
		integration.LastEventAt = now
		instance.LastHeartbeatAt = now
		instance.LastProbeAt = now
		if issue != "" {
			instance.Status = "degraded"
			instance.HealthStatus = "degraded"
			instance.AlertLevel = severity
		}
		alert := syncMonitoringAlert(data, *instance, event, severity, issue)
		addAudit(data, "monitoring:"+integration.Code, "report", "product_monitoring_event", event.ID, event.EventNo+" "+instance.CustomerName, clientIP(r))
		response = productMonitoringReportResponse{Accepted: true, Integration: *integration, Instance: *instance, Event: event, Alert: alert}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	a.emit("product_ops.monitoring.reported", response)
	if response.Alert != nil && response.Alert.Status != "handled" {
		a.emit("product_ops.alert.created", response.Alert)
	}
	writeJSON(w, http.StatusCreated, response)
}

func productMonitoringIntegrationIndex(data AppData, token, code string) int {
	token = strings.TrimSpace(token)
	code = strings.TrimSpace(code)
	for i, integration := range data.ProductMonitoringIntegrations {
		if token != "" && integration.Token == token {
			return i
		}
		if code != "" && integration.Code == code {
			return i
		}
	}
	return -1
}

func productInstanceIndexForMonitoring(data AppData, instanceID int64, watermark string, labels map[string]string) int {
	if instanceID != 0 {
		return productInstanceIndexByID(data, instanceID)
	}
	watermark = fallback(strings.TrimSpace(watermark), strings.TrimSpace(labels["watermark"]))
	if watermark == "" {
		watermark = strings.TrimSpace(labels["instance"])
	}
	for i := range data.ProductInstances {
		if data.ProductInstances[i].Watermark == watermark || data.ProductInstances[i].LicenseID == watermark {
			return i
		}
	}
	return -1
}

func evaluateMonitoringEvent(data AppData, event ProductMonitoringEvent) (string, string, string) {
	for _, rule := range data.ProductAlertRules {
		if rule.Status != "active" || !productAlertRuleMatches(rule, event) {
			continue
		}
		if compareMetric(event.Value, rule.Operator, rule.Threshold) {
			severity := normalizeTelemetrySeverity(rule.Severity)
			return severity, fmt.Sprintf("%s：%s %.2f %s %.2f", rule.Name, event.Metric, event.Value, rule.Operator, rule.Threshold), rule.RuleNo
		}
	}
	severity := normalizeTelemetrySeverity(event.Severity)
	if severity == "critical" || severity == "warning" || event.Status == "firing" || event.Status == "triggered" {
		if severity == "normal" {
			severity = "warning"
		}
		return severity, fallback(event.Message, fallback(event.Title, event.Metric+" 外部监控告警")), ""
	}
	return "normal", "", ""
}

func productAlertRuleMatches(rule ProductAlertRule, event ProductMonitoringEvent) bool {
	if rule.Source != "" && rule.Source != "all" && rule.Source != event.Source && rule.Source != event.Provider {
		return false
	}
	if rule.Component != "" && rule.Component != "all" && rule.Component != event.Component {
		return false
	}
	return rule.Metric == event.Metric
}

func compareMetric(value float64, operator string, threshold float64) bool {
	switch strings.TrimSpace(operator) {
	case ">", "gt":
		return value > threshold
	case ">=", "gte", "":
		return value >= threshold
	case "<", "lt":
		return value < threshold
	case "<=", "lte":
		return value <= threshold
	case "=", "==", "eq":
		return value == threshold
	case "!=", "ne":
		return value != threshold
	default:
		return value >= threshold
	}
}

func syncMonitoringAlert(data *AppData, instance ProductInstance, event ProductMonitoringEvent, severity, issue string) *SystemAlert {
	if issue == "" {
		return nil
	}
	for i := range data.SystemAlerts {
		alert := &data.SystemAlerts[i]
		if alert.InstanceID != instance.ID || alert.Source != "monitoring" || alert.Status == "handled" || alert.Status == "closed" {
			continue
		}
		alert.Severity = severity
		alert.Title = fallback(event.Title, alert.Title)
		alert.Message = issue
		alert.LastSeenAt = event.ReceivedAt
		applyProductAlertGovernance(data, alert, productAlertGovernanceContext{Component: event.Component, Metric: event.Metric, EventNo: event.EventNo}, false)
		return alert
	}
	id := nextID(data, "systemAlert")
	alert := SystemAlert{
		ID: id, AlertNo: number("AL", id), InstanceID: instance.ID, CustomerName: instance.CustomerName,
		Severity: severity, Source: "monitoring", Title: fallback(event.Title, "第三方监控告警"), Message: issue,
		Status: "open", FirstSeenAt: event.ReceivedAt, LastSeenAt: event.ReceivedAt,
	}
	data.SystemAlerts = append(data.SystemAlerts, alert)
	applyProductAlertGovernance(data, &data.SystemAlerts[len(data.SystemAlerts)-1], productAlertGovernanceContext{Component: event.Component, Metric: event.Metric, EventNo: event.EventNo}, true)
	return &data.SystemAlerts[len(data.SystemAlerts)-1]
}

func trimProductMonitoringEvents(data *AppData) {
	const maxEvents = 200
	if len(data.ProductMonitoringEvents) <= maxEvents {
		return
	}
	data.ProductMonitoringEvents = append([]ProductMonitoringEvent{}, data.ProductMonitoringEvents[len(data.ProductMonitoringEvents)-maxEvents:]...)
}
