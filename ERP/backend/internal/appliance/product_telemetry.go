package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

type productTelemetryReportRequest struct {
	ProbeToken   string `json:"probeToken"`
	Watermark    string `json:"watermark"`
	Source       string `json:"source"`
	Component    string `json:"component"`
	Severity     string `json:"severity"`
	EventType    string `json:"eventType"`
	TraceID      string `json:"traceId"`
	SpanID       string `json:"spanId"`
	Endpoint     string `json:"endpoint"`
	DurationMs   int    `json:"durationMs"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
	Message      string `json:"message"`
	OccurredAt   string `json:"occurredAt"`
}

type productTelemetryReportResponse struct {
	Accepted bool                  `json:"accepted"`
	Instance ProductInstance       `json:"instance"`
	Event    ProductTelemetryEvent `json:"event"`
	Alert    *SystemAlert          `json:"alert,omitempty"`
}

func (a *App) productOpsTelemetryReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req productTelemetryReportRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid telemetry payload")
		return
	}
	token := strings.TrimSpace(r.Header.Get("X-CBMP-Probe-Token"))
	if token == "" {
		token = strings.TrimSpace(req.ProbeToken)
	}
	if token == "" {
		unauthorized(w)
		return
	}
	var response productTelemetryReportResponse
	err := a.store.Mutate(func(data *AppData) error {
		index := productInstanceIndexByProbeToken(*data, token)
		if index < 0 {
			return fmt.Errorf("probe token invalid")
		}
		instance := &data.ProductInstances[index]
		now := nowString()
		event := ProductTelemetryEvent{
			ID:           nextID(data, "telemetryEvent"),
			InstanceID:   instance.ID,
			CustomerName: instance.CustomerName,
			Watermark:    fallback(strings.TrimSpace(req.Watermark), instance.Watermark),
			Source:       normalizeTelemetrySource(req.Source),
			Component:    fallback(strings.ToLower(strings.TrimSpace(req.Component)), "server"),
			Severity:     normalizeTelemetrySeverity(req.Severity),
			EventType:    fallback(strings.TrimSpace(req.EventType), "log"),
			TraceID:      strings.TrimSpace(req.TraceID),
			SpanID:       strings.TrimSpace(req.SpanID),
			Endpoint:     strings.TrimSpace(req.Endpoint),
			DurationMs:   req.DurationMs,
			StatusCode:   req.StatusCode,
			ErrorMessage: strings.TrimSpace(req.ErrorMessage),
			Message:      strings.TrimSpace(req.Message),
			OccurredAt:   fallback(strings.TrimSpace(req.OccurredAt), now),
			ReceivedAt:   now,
			SourceIP:     clientIP(r),
		}
		event.EventNo = number("TE", event.ID)
		severity, issue := evaluateTelemetryEvent(event)
		event.AlertRaised = issue != ""
		if issue != "" {
			event.Severity = severity
		}
		data.ProductTelemetryEvents = append(data.ProductTelemetryEvents, event)
		trimProductTelemetryEvents(data)

		instance.LastProbeAt = event.ReceivedAt
		instance.LastHeartbeatAt = event.ReceivedAt
		if issue != "" {
			instance.Status = "degraded"
			instance.HealthStatus = "degraded"
			instance.AlertLevel = severity
		}

		alert := syncTelemetryAlert(data, *instance, event, severity, issue)
		addAudit(data, "telemetry:"+instance.Watermark, "report", "product_telemetry_event", event.ID, instance.CustomerName+" "+event.Source+" "+event.Severity, clientIP(r))
		response = productTelemetryReportResponse{Accepted: true, Instance: *instance, Event: event, Alert: alert}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	a.emit("product_ops.telemetry.reported", response)
	if response.Alert != nil && response.Alert.Status != "handled" {
		a.emit("product_ops.alert.created", response.Alert)
	}
	writeJSON(w, http.StatusCreated, response)
}

func productInstanceIndexByProbeToken(data AppData, token string) int {
	for i := range data.ProductInstances {
		if data.ProductInstances[i].ProbeEnabled && data.ProductInstances[i].ProbeToken == token {
			return i
		}
	}
	return -1
}

func normalizeTelemetrySource(source string) string {
	source = strings.ToLower(strings.TrimSpace(source))
	switch source {
	case "log", "apm", "trace", "metric":
		return source
	default:
		if source == "" {
			return "log"
		}
		return source
	}
}

func normalizeTelemetrySeverity(severity string) string {
	severity = strings.ToLower(strings.TrimSpace(severity))
	switch severity {
	case "fatal", "panic", "critical":
		return "critical"
	case "error", "warning", "warn":
		return "warning"
	case "info", "normal", "ok", "healthy":
		return "normal"
	default:
		if severity == "" {
			return "normal"
		}
		return severity
	}
}

func evaluateTelemetryEvent(event ProductTelemetryEvent) (string, string) {
	issues := []string{}
	severity := normalizeTelemetrySeverity(event.Severity)
	if severity == "critical" {
		issues = append(issues, "严重级别为 critical")
	} else if severity == "warning" {
		issues = append(issues, "严重级别为 warning")
	}
	if strings.Contains(strings.ToLower(event.EventType), "error") || strings.TrimSpace(event.ErrorMessage) != "" {
		if severity != "critical" {
			severity = "warning"
		}
		issues = append(issues, fallback(event.ErrorMessage, "错误事件 "+event.EventType))
	}
	if event.StatusCode >= 500 {
		severity = "critical"
		issues = append(issues, fmt.Sprintf("接口 %s 返回 %d", fallback(event.Endpoint, "-"), event.StatusCode))
	} else if event.StatusCode >= 400 {
		if severity != "critical" {
			severity = "warning"
		}
		issues = append(issues, fmt.Sprintf("接口 %s 返回 %d", fallback(event.Endpoint, "-"), event.StatusCode))
	}
	if event.DurationMs >= 3000 {
		severity = "critical"
		issues = append(issues, fmt.Sprintf("链路耗时过高 %dms", event.DurationMs))
	} else if event.DurationMs >= 1000 {
		if severity != "critical" {
			severity = "warning"
		}
		issues = append(issues, fmt.Sprintf("链路耗时升高 %dms", event.DurationMs))
	}
	if len(issues) == 0 {
		return "normal", ""
	}
	return severity, strings.Join(issues, "；")
}

func syncTelemetryAlert(data *AppData, instance ProductInstance, event ProductTelemetryEvent, severity, issue string) *SystemAlert {
	if issue == "" {
		return nil
	}
	for i := range data.SystemAlerts {
		alert := &data.SystemAlerts[i]
		if alert.InstanceID != instance.ID || alert.Source != "telemetry" || alert.Status == "handled" || alert.Status == "closed" {
			continue
		}
		alert.Severity = severity
		alert.Message = issue
		alert.LastSeenAt = event.ReceivedAt
		applyProductAlertGovernance(data, alert, productAlertGovernanceContext{Component: event.Component, Metric: event.EventType, EventNo: event.EventNo}, false)
		return alert
	}
	id := nextID(data, "systemAlert")
	title := "客户现场可观测事件异常"
	if event.Source == "apm" {
		title = "客户现场 APM 性能异常"
	}
	if event.Source == "trace" {
		title = "客户现场链路追踪异常"
	}
	alert := SystemAlert{
		ID: id, AlertNo: number("AL", id), InstanceID: instance.ID, CustomerName: instance.CustomerName,
		Severity: severity, Source: "telemetry", Title: title, Message: issue,
		Status: "open", FirstSeenAt: event.ReceivedAt, LastSeenAt: event.ReceivedAt,
	}
	data.SystemAlerts = append(data.SystemAlerts, alert)
	applyProductAlertGovernance(data, &data.SystemAlerts[len(data.SystemAlerts)-1], productAlertGovernanceContext{Component: event.Component, Metric: event.EventType, EventNo: event.EventNo}, true)
	return &data.SystemAlerts[len(data.SystemAlerts)-1]
}

func trimProductTelemetryEvents(data *AppData) {
	const maxEvents = 200
	if len(data.ProductTelemetryEvents) <= maxEvents {
		return
	}
	data.ProductTelemetryEvents = append([]ProductTelemetryEvent{}, data.ProductTelemetryEvents[len(data.ProductTelemetryEvents)-maxEvents:]...)
}
