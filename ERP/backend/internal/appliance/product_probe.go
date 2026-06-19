package appliance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type productProbeReportRequest struct {
	ProbeToken    string  `json:"probeToken"`
	Watermark     string  `json:"watermark"`
	Component     string  `json:"component"`
	ClientVersion string  `json:"clientVersion"`
	ServerVersion string  `json:"serverVersion"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float64 `json:"memoryPercent"`
	DiskPercent   float64 `json:"diskPercent"`
	QueueBacklog  int     `json:"queueBacklog"`
	ErrorCount    int     `json:"errorCount"`
	Message       string  `json:"message"`
	ReportedAt    string  `json:"reportedAt"`
}

type productProbeReportResponse struct {
	Accepted bool               `json:"accepted"`
	Instance ProductInstance    `json:"instance"`
	Report   ProductProbeReport `json:"report"`
	Alert    *SystemAlert       `json:"alert,omitempty"`
}

func (a *App) productOpsProbeReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req productProbeReportRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid probe report payload")
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
	var response productProbeReportResponse
	err := a.store.Mutate(func(data *AppData) error {
		index := -1
		for i := range data.ProductInstances {
			if data.ProductInstances[i].ProbeEnabled && data.ProductInstances[i].ProbeToken == token {
				index = i
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("probe token invalid")
		}
		instance := &data.ProductInstances[index]
		now := nowString()
		status := normalizeProbeStatus(req.Status)
		report := ProductProbeReport{
			ID: nextID(data, "probeReport"), ReportNo: number("PR", data.Next["probeReport"]),
			InstanceID: instance.ID, CustomerName: instance.CustomerName, Watermark: fallback(req.Watermark, instance.Watermark),
			Component:     fallback(strings.ToLower(strings.TrimSpace(req.Component)), "server"),
			ClientVersion: fallback(strings.TrimSpace(req.ClientVersion), instance.ClientVersion),
			ServerVersion: fallback(strings.TrimSpace(req.ServerVersion), instance.ServerVersion),
			Status:        status, CPUPercent: req.CPUPercent, MemoryPercent: req.MemoryPercent, DiskPercent: req.DiskPercent,
			QueueBacklog: req.QueueBacklog, ErrorCount: req.ErrorCount, Message: strings.TrimSpace(req.Message),
			ReportedAt: fallback(req.ReportedAt, now), ReceivedAt: now, SourceIP: clientIP(r),
		}
		severity, issue := evaluateProbeReport(report, latestUpdateVersions(data.Updates))
		report.AlertRaised = issue != ""
		data.ProductProbeReports = append(data.ProductProbeReports, report)
		trimProductProbeReports(data)

		instance.ClientVersion = fallback(report.ClientVersion, instance.ClientVersion)
		instance.ServerVersion = fallback(report.ServerVersion, instance.ServerVersion)
		instance.LastProbeAt = report.ReceivedAt
		instance.LastHeartbeatAt = report.ReceivedAt
		instance.HealthStatus = status
		if issue == "" {
			instance.Status = "online"
			if instance.AlertLevel == "" || instance.AlertLevel == "critical" || instance.AlertLevel == "warning" {
				instance.AlertLevel = fallback(instance.LicenseRisk, "normal")
			}
		} else {
			instance.Status = "degraded"
			instance.AlertLevel = severity
		}

		alert := syncProbeAlert(data, *instance, report, severity, issue)
		addAudit(data, "probe:"+instance.Watermark, "report", "product_probe_report", report.ID, instance.CustomerName+" "+status, clientIP(r))
		response = productProbeReportResponse{Accepted: true, Instance: *instance, Report: report, Alert: alert}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	a.emit("product_ops.probe.reported", response)
	if response.Alert != nil && response.Alert.Status != "handled" {
		a.emit("product_ops.alert.created", response.Alert)
	}
	writeJSON(w, http.StatusCreated, response)
}

func productProbeToken(instance ProductInstance) string {
	seed := strings.Join([]string{"probe", instance.Watermark, instance.LicenseID, instance.CustomerName}, "|")
	sum := sha256.Sum256([]byte(seed))
	return "probe-" + hex.EncodeToString(sum[:])[:24]
}

func normalizeProbeStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "ok", "online", "healthy", "normal":
		return "healthy"
	case "critical", "down", "offline":
		return "critical"
	case "warning", "degraded":
		return "degraded"
	default:
		if status == "" {
			return "healthy"
		}
		return status
	}
}

func evaluateProbeReport(report ProductProbeReport, latest map[string]string) (string, string) {
	issues := []string{}
	severity := "warning"
	if report.Status == "critical" || report.Status == "offline" {
		severity = "critical"
		issues = append(issues, "探针状态为 "+report.Status)
	} else if report.Status != "healthy" {
		issues = append(issues, "探针状态为 "+report.Status)
	}
	if report.CPUPercent >= 90 || report.MemoryPercent >= 90 || report.DiskPercent >= 90 {
		severity = "critical"
		issues = append(issues, fmt.Sprintf("资源使用率过高 CPU %.1f%% / Memory %.1f%% / Disk %.1f%%", report.CPUPercent, report.MemoryPercent, report.DiskPercent))
	} else if report.CPUPercent >= 80 || report.MemoryPercent >= 80 || report.DiskPercent >= 80 {
		issues = append(issues, fmt.Sprintf("资源使用率偏高 CPU %.1f%% / Memory %.1f%% / Disk %.1f%%", report.CPUPercent, report.MemoryPercent, report.DiskPercent))
	}
	if report.ErrorCount >= 10 || report.QueueBacklog >= 1000 {
		severity = "critical"
		issues = append(issues, fmt.Sprintf("错误数或队列积压过高 errors=%d backlog=%d", report.ErrorCount, report.QueueBacklog))
	} else if report.ErrorCount > 0 || report.QueueBacklog >= 100 {
		issues = append(issues, fmt.Sprintf("错误数或队列积压升高 errors=%d backlog=%d", report.ErrorCount, report.QueueBacklog))
	}
	if latest["client"] != "" && report.ClientVersion != "" && report.ClientVersion < latest["client"] {
		issues = append(issues, "客户端版本低于最新包 "+latest["client"])
	}
	if latest["server"] != "" && report.ServerVersion != "" && report.ServerVersion < latest["server"] {
		issues = append(issues, "服务端版本低于最新包 "+latest["server"])
	}
	if len(issues) == 0 {
		return "normal", ""
	}
	return severity, strings.Join(issues, "；")
}

func latestUpdateVersions(updates []UpdatePackage) map[string]string {
	out := map[string]string{}
	for _, update := range updates {
		if update.Status != "available" && update.Status != "gray" && update.Status != "installed" {
			continue
		}
		component := update.Component
		if component == "" {
			component = "server"
		}
		if component == "all" {
			if update.Version > out["client"] {
				out["client"] = update.Version
			}
			if update.Version > out["server"] {
				out["server"] = update.Version
			}
			continue
		}
		if update.Version > out[component] {
			out[component] = update.Version
		}
	}
	return out
}

func syncProbeAlert(data *AppData, instance ProductInstance, report ProductProbeReport, severity, issue string) *SystemAlert {
	for i := range data.SystemAlerts {
		alert := &data.SystemAlerts[i]
		if alert.InstanceID != instance.ID || alert.Source != "probe" || (alert.Status == "handled" || alert.Status == "closed") {
			continue
		}
		if issue == "" {
			alert.Status = "handled"
			alert.HandledBy = "system-probe"
			alert.HandledAt = report.ReceivedAt
			alert.LastSeenAt = report.ReceivedAt
			alert.Message = alert.Message + "；探针恢复：" + fallback(report.Message, "指标恢复正常")
			return alert
		}
		alert.Severity = severity
		alert.Message = issue
		alert.LastSeenAt = report.ReceivedAt
		applyProductAlertGovernance(data, alert, productAlertGovernanceContext{Component: report.Component, Metric: "probe_health", EventNo: report.ReportNo}, false)
		return alert
	}
	if issue == "" {
		return nil
	}
	id := nextID(data, "systemAlert")
	alert := SystemAlert{
		ID: id, AlertNo: number("AL", id), InstanceID: instance.ID, CustomerName: instance.CustomerName,
		Severity: severity, Source: "probe", Title: "客户现场探针异常", Message: issue,
		Status: "open", FirstSeenAt: report.ReceivedAt, LastSeenAt: report.ReceivedAt,
	}
	data.SystemAlerts = append(data.SystemAlerts, alert)
	applyProductAlertGovernance(data, &data.SystemAlerts[len(data.SystemAlerts)-1], productAlertGovernanceContext{Component: report.Component, Metric: "probe_health", EventNo: report.ReportNo}, true)
	return &data.SystemAlerts[len(data.SystemAlerts)-1]
}

func trimProductProbeReports(data *AppData) {
	const maxReports = 200
	if len(data.ProductProbeReports) <= maxReports {
		return
	}
	data.ProductProbeReports = append([]ProductProbeReport{}, data.ProductProbeReports[len(data.ProductProbeReports)-maxReports:]...)
}
