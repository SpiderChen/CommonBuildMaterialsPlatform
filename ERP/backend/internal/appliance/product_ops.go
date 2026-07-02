//go:build legacy_product_ops

package appliance

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type ProductOpsOverview struct {
	KPIs                   ProductOpsKPI                  `json:"kpis"`
	Instances              []ProductInstance              `json:"instances"`
	Alerts                 []SystemAlert                  `json:"alerts"`
	RenewalTasks           []ProductRenewalTask           `json:"renewalTasks"`
	RenewalQuotes          []ProductRenewalQuote          `json:"renewalQuotes"`
	RenewalContracts       []ProductRenewalContract       `json:"renewalContracts"`
	RenewalPayments        []ProductRenewalPayment        `json:"renewalPayments"`
	RenewalApprovals       []ProductRenewalApproval       `json:"renewalApprovals"`
	RenewalInvoices        []ProductRenewalInvoice        `json:"renewalInvoices"`
	RenewalESigns          []ProductRenewalESign          `json:"renewalESigns"`
	RenewalIntegrations    []ProductRenewalIntegration    `json:"renewalIntegrations"`
	RenewalSyncRecords     []ProductRenewalSyncRecord     `json:"renewalSyncRecords"`
	ProbeReports           []ProductProbeReport           `json:"probeReports"`
	TelemetryEvents        []ProductTelemetryEvent        `json:"telemetryEvents"`
	MonitoringIntegrations []ProductMonitoringIntegration `json:"monitoringIntegrations"`
	AlertRules             []ProductAlertRule             `json:"alertRules"`
	MonitoringEvents       []ProductMonitoringEvent       `json:"monitoringEvents"`
	AlertPolicies          []ProductAlertPolicy           `json:"alertPolicies"`
	AlertChannels          []ProductAlertChannel          `json:"alertChannels"`
	AlertNotifications     []ProductAlertNotification     `json:"alertNotifications"`
	UpdateRollouts         []ProductUpdateRollout         `json:"updateRollouts"`
	UpdateExecutions       []ProductUpdateExecution       `json:"updateExecutions"`
	SystemUpdateTasks      []ProductSystemUpdateTask      `json:"systemUpdateTasks"`
	Updates                []UpdatePackage                `json:"updates"`
	LicensePortal          LicensePortalOverview          `json:"licensePortal"`
	RecentEvents           []LicensePortalEvent           `json:"recentEvents"`
	Runtime                RuntimeStatus                  `json:"runtime"`
}

type ProductOpsKPI struct {
	Customers                 int     `json:"customers"`
	OnlineInstances           int     `json:"onlineInstances"`
	DegradedInstances         int     `json:"degradedInstances"`
	ExpiringLicenses          int     `json:"expiringLicenses"`
	OpenAlerts                int     `json:"openAlerts"`
	CriticalAlerts            int     `json:"criticalAlerts"`
	OpenRenewals              int     `json:"openRenewals"`
	OverdueRenewals           int     `json:"overdueRenewals"`
	PendingRenewalQuotes      int     `json:"pendingRenewalQuotes"`
	PendingRenewalContracts   int     `json:"pendingRenewalContracts"`
	PaidRenewalAmount         float64 `json:"paidRenewalAmount"`
	PendingRenewalApprovals   int     `json:"pendingRenewalApprovals"`
	IssuedRenewalInvoices     int     `json:"issuedRenewalInvoices"`
	PendingRenewalESigns      int     `json:"pendingRenewalESigns"`
	ActiveRenewalIntegrations int     `json:"activeRenewalIntegrations"`
	RenewalSyncRecords        int     `json:"renewalSyncRecords"`
	FailedRenewalSyncRecords  int     `json:"failedRenewalSyncRecords"`
	PendingRenewalSyncRecords int     `json:"pendingRenewalSyncRecords"`
	ProbeReports              int     `json:"probeReports"`
	UnhealthyProbes           int     `json:"unhealthyProbes"`
	TelemetryEvents           int     `json:"telemetryEvents"`
	CriticalTelemetryEvents   int     `json:"criticalTelemetryEvents"`
	MonitoringIntegrations    int     `json:"monitoringIntegrations"`
	ActiveAlertRules          int     `json:"activeAlertRules"`
	MonitoringEvents          int     `json:"monitoringEvents"`
	MonitoringAlerts          int     `json:"monitoringAlerts"`
	ActiveAlertPolicies       int     `json:"activeAlertPolicies"`
	SuppressedAlerts          int     `json:"suppressedAlerts"`
	EscalatedAlerts           int     `json:"escalatedAlerts"`
	ActiveAlertChannels       int     `json:"activeAlertChannels"`
	AlertNotifications        int     `json:"alertNotifications"`
	FailedAlertNotifications  int     `json:"failedAlertNotifications"`
	PendingAlertNotifications int     `json:"pendingAlertNotifications"`
	ActiveRollouts            int     `json:"activeRollouts"`
	FailedRolloutItems        int     `json:"failedRolloutItems"`
	UpdateExecutions          int     `json:"updateExecutions"`
	FailedUpdateExecutions    int     `json:"failedUpdateExecutions"`
	SystemUpdateTasks         int     `json:"systemUpdateTasks"`
	RunningSystemUpdateTasks  int     `json:"runningSystemUpdateTasks"`
	FailedSystemUpdateTasks   int     `json:"failedSystemUpdateTasks"`
	ClientUpdatePackages      int     `json:"clientUpdatePackages"`
	ServerUpdatePackages      int     `json:"serverUpdatePackages"`
	AvailableUpdates          int     `json:"availableUpdates"`
}

func (a *App) productOps(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "overview") {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, buildProductOpsOverview(a.mustSnapshot(), a.runtimeStatus()))
		return
	}
	switch parts[0] {
	case "instances":
		a.productOpsInstances(w, r, session, parts[1:])
	case "alerts":
		a.productOpsAlerts(w, r, session, parts[1:])
	case "renewals":
		a.productOpsRenewals(w, r, session, parts[1:])
	case "monitoring":
		a.productOpsMonitoring(w, r, session, parts[1:])
	case "rollouts":
		a.productOpsRollouts(w, r, session, parts[1:])
	default:
		writeError(w, http.StatusNotFound, "unknown product ops route")
	}
}

func (a *App) productOpsInstances(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodPost {
		var req ProductInstance
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid product instance payload")
			return
		}
		var updated ProductInstance
		err := a.store.Mutate(func(data *AppData) error {
			req.CustomerName = strings.TrimSpace(req.CustomerName)
			if req.CustomerName == "" {
				return fmt.Errorf("客户名称不能为空")
			}
			req.Status = fallback(req.Status, "online")
			req.DeploymentMode = fallback(req.DeploymentMode, "private")
			req.RenewalStage = fallback(req.RenewalStage, "待跟进")
			req.AlertLevel = fallback(req.AlertLevel, "normal")
			req.HealthStatus = fallback(req.HealthStatus, req.Status)
			if req.CreatedAt == "" {
				req.CreatedAt = nowString()
			}
			for i := range data.ProductInstances {
				if data.ProductInstances[i].ID == req.ID || (req.Watermark != "" && data.ProductInstances[i].Watermark == req.Watermark) {
					req.ID = data.ProductInstances[i].ID
					req.ProbeToken = fallback(req.ProbeToken, data.ProductInstances[i].ProbeToken)
					if req.ProbeToken == "" {
						req.ProbeToken = "probe-" + tokenString()
					}
					if !req.ProbeEnabled && data.ProductInstances[i].ProbeEnabled {
						req.ProbeEnabled = true
					}
					req.LastProbeAt = fallback(req.LastProbeAt, data.ProductInstances[i].LastProbeAt)
					data.ProductInstances[i] = req
					updated = req
					addAudit(data, session.User.Username, "update", "product_instance", req.ID, req.CustomerName, clientIP(r))
					return nil
				}
			}
			if req.ProbeToken == "" {
				req.ProbeToken = "probe-" + tokenString()
			}
			req.ProbeEnabled = true
			req.ID = nextID(data, "productInstance")
			data.ProductInstances = append(data.ProductInstances, req)
			updated = req
			addAudit(data, session.User.Username, "create", "product_instance", req.ID, req.CustomerName, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, updated, "product_ops.instance.saved")
		return
	}
	if len(parts) == 2 && parts[1] == "heartbeat" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var updated ProductInstance
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.ProductInstances {
				if data.ProductInstances[i].ID == id {
					data.ProductInstances[i].LastHeartbeatAt = nowString()
					data.ProductInstances[i].Status = "online"
					updated = data.ProductInstances[i]
					addAudit(data, session.User.Username, "heartbeat", "product_instance", id, updated.CustomerName, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("客户实例不存在")
		})
		a.respondMutation(w, err, updated, "product_ops.instance.heartbeat")
		return
	}
	writeError(w, http.StatusNotFound, "unknown product instance route")
}

func (a *App) productOpsAlerts(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "channels" && r.Method == http.MethodPost {
		var req ProductAlertChannel
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid alert channel payload")
			return
		}
		var saved ProductAlertChannel
		err := a.store.Mutate(func(data *AppData) error {
			normalized, err := normalizeProductAlertChannel(req, session.User.DisplayName)
			if err != nil {
				return err
			}
			for i := range data.ProductAlertChannels {
				if data.ProductAlertChannels[i].ID == normalized.ID || (normalized.Code != "" && data.ProductAlertChannels[i].Code == normalized.Code) {
					normalized.ID = data.ProductAlertChannels[i].ID
					normalized.ChannelNo = fallback(normalized.ChannelNo, data.ProductAlertChannels[i].ChannelNo)
					normalized.CreatedBy = fallback(normalized.CreatedBy, data.ProductAlertChannels[i].CreatedBy)
					normalized.CreatedAt = fallback(normalized.CreatedAt, data.ProductAlertChannels[i].CreatedAt)
					normalized.LastDeliveredAt = data.ProductAlertChannels[i].LastDeliveredAt
					normalized.LastError = data.ProductAlertChannels[i].LastError
					data.ProductAlertChannels[i] = normalized
					saved = normalized
					addAudit(data, session.User.Username, "update", "product_alert_channel", saved.ID, saved.Code+" "+saved.Name, clientIP(r))
					return nil
				}
			}
			normalized.ID = nextID(data, "alertChannel")
			normalized.ChannelNo = number("AC", normalized.ID)
			data.ProductAlertChannels = append(data.ProductAlertChannels, normalized)
			saved = normalized
			addAudit(data, session.User.Username, "create", "product_alert_channel", saved.ID, saved.Code+" "+saved.Name, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "product_ops.alert.channel.saved")
		return
	}
	if len(parts) == 3 && parts[0] == "notifications" && parts[2] == "retry" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var updated ProductAlertNotification
		err := a.store.Mutate(func(data *AppData) error {
			result, err := retryProductAlertNotification(data, id)
			updated = result
			if err == nil {
				addAudit(data, session.User.Username, "retry", "product_alert_notification", id, updated.NotificationNo+" "+updated.Status, clientIP(r))
			}
			return err
		})
		a.respondMutation(w, err, updated, "product_ops.alert.notification.retried")
		return
	}
	if len(parts) == 1 && parts[0] == "policies" && r.Method == http.MethodPost {
		var req ProductAlertPolicy
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid alert policy payload")
			return
		}
		var saved ProductAlertPolicy
		err := a.store.Mutate(func(data *AppData) error {
			req.Name = strings.TrimSpace(req.Name)
			if req.Name == "" {
				return fmt.Errorf("告警策略名称不能为空")
			}
			req.Source = fallback(strings.ToLower(strings.TrimSpace(req.Source)), "all")
			req.Component = fallback(strings.ToLower(strings.TrimSpace(req.Component)), "all")
			req.Metric = fallback(strings.ToLower(strings.TrimSpace(req.Metric)), "all")
			req.Severity = fallback(normalizeTelemetrySeverity(req.Severity), "warning")
			req.Status = fallback(strings.TrimSpace(req.Status), "active")
			if req.AggregateWindowMinutes <= 0 {
				req.AggregateWindowMinutes = 30
			}
			if req.SuppressMinutes < 0 {
				req.SuppressMinutes = 0
			}
			if req.EscalateAfterMinutes < 0 {
				req.EscalateAfterMinutes = 0
			}
			if len(req.NotifyChannels) == 0 {
				req.NotifyChannels = []string{"sse"}
			}
			for i := range data.ProductAlertPolicies {
				if data.ProductAlertPolicies[i].ID == req.ID || (req.PolicyNo != "" && data.ProductAlertPolicies[i].PolicyNo == req.PolicyNo) {
					req.ID = data.ProductAlertPolicies[i].ID
					req.PolicyNo = fallback(req.PolicyNo, data.ProductAlertPolicies[i].PolicyNo)
					req.CreatedBy = fallback(req.CreatedBy, data.ProductAlertPolicies[i].CreatedBy)
					req.CreatedAt = fallback(req.CreatedAt, data.ProductAlertPolicies[i].CreatedAt)
					data.ProductAlertPolicies[i] = req
					saved = req
					addAudit(data, session.User.Username, "update", "product_alert_policy", req.ID, req.PolicyNo+" "+req.Name, clientIP(r))
					return nil
				}
			}
			req.ID = nextID(data, "alertPolicy")
			req.PolicyNo = number("AP", req.ID)
			req.CreatedBy = session.User.DisplayName
			req.CreatedAt = nowString()
			data.ProductAlertPolicies = append(data.ProductAlertPolicies, req)
			saved = req
			addAudit(data, session.User.Username, "create", "product_alert_policy", req.ID, req.PolicyNo+" "+req.Name, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "product_ops.alert.policy.saved")
		return
	}
	if len(parts) == 0 && r.Method == http.MethodPost {
		var req SystemAlert
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid alert payload")
			return
		}
		var created SystemAlert
		err := a.store.Mutate(func(data *AppData) error {
			req.CustomerName = strings.TrimSpace(req.CustomerName)
			if req.CustomerName == "" && req.InstanceID != 0 {
				for _, instance := range data.ProductInstances {
					if instance.ID == req.InstanceID {
						req.CustomerName = instance.CustomerName
						break
					}
				}
			}
			if req.CustomerName == "" || strings.TrimSpace(req.Title) == "" {
				return fmt.Errorf("告警必须包含客户和标题")
			}
			id := nextID(data, "systemAlert")
			now := nowString()
			req.ID = id
			req.AlertNo = number("AL", id)
			req.Severity = fallback(req.Severity, "warning")
			req.Source = fallback(req.Source, "server")
			req.Status = fallback(req.Status, "open")
			req.FirstSeenAt = fallback(req.FirstSeenAt, now)
			req.LastSeenAt = fallback(req.LastSeenAt, now)
			ctx := productAlertGovernanceContext{Component: req.Source}
			req.GroupKey = fallback(req.GroupKey, productAlertGroupKey(req, ctx))
			for i := range data.SystemAlerts {
				if data.SystemAlerts[i].GroupKey == req.GroupKey && data.SystemAlerts[i].Status != "handled" && data.SystemAlerts[i].Status != "closed" {
					data.SystemAlerts[i].Severity = req.Severity
					data.SystemAlerts[i].Title = fallback(req.Title, data.SystemAlerts[i].Title)
					data.SystemAlerts[i].Message = req.Message
					data.SystemAlerts[i].LastSeenAt = req.LastSeenAt
					applyProductAlertGovernance(data, &data.SystemAlerts[i], ctx, false)
					created = data.SystemAlerts[i]
					addAudit(data, session.User.Username, "aggregate", "system_alert", created.ID, created.AlertNo+" "+created.Title, clientIP(r))
					return nil
				}
			}
			data.SystemAlerts = append(data.SystemAlerts, req)
			applyProductAlertGovernance(data, &data.SystemAlerts[len(data.SystemAlerts)-1], ctx, true)
			created = data.SystemAlerts[len(data.SystemAlerts)-1]
			addAudit(data, session.User.Username, "create", "system_alert", created.ID, created.AlertNo+" "+created.Title, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, created, "product_ops.alert.created")
		return
	}
	if len(parts) == 2 && parts[1] == "escalate" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Level  string `json:"level"`
			Remark string `json:"remark"`
		}
		_ = readJSON(r, &req)
		var updated SystemAlert
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.SystemAlerts {
				if data.SystemAlerts[i].ID == id {
					data.SystemAlerts[i].EscalationLevel = fallback(strings.TrimSpace(req.Level), "manual_on_call")
					data.SystemAlerts[i].EscalatedAt = nowString()
					data.SystemAlerts[i].Severity = "critical"
					if strings.TrimSpace(req.Remark) != "" {
						data.SystemAlerts[i].Message = data.SystemAlerts[i].Message + "；升级：" + strings.TrimSpace(req.Remark)
					}
					appendProductAlertNotifications(data, data.SystemAlerts[i], nil, "escalated", data.SystemAlerts[i].AlertNo+" 已人工升级给 "+data.SystemAlerts[i].EscalationLevel)
					updated = data.SystemAlerts[i]
					addAudit(data, session.User.Username, "escalate", "system_alert", id, updated.AlertNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("系统告警不存在")
		})
		a.respondMutation(w, err, updated, "product_ops.alert.escalated")
		return
	}
	if len(parts) == 2 && parts[1] == "handle" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Remark string `json:"remark"`
		}
		_ = readJSON(r, &req)
		var updated SystemAlert
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.SystemAlerts {
				if data.SystemAlerts[i].ID == id {
					data.SystemAlerts[i].Status = "handled"
					data.SystemAlerts[i].HandledBy = session.User.DisplayName
					data.SystemAlerts[i].HandledAt = nowString()
					if strings.TrimSpace(req.Remark) != "" {
						data.SystemAlerts[i].Message = data.SystemAlerts[i].Message + "；处理：" + strings.TrimSpace(req.Remark)
					}
					updated = data.SystemAlerts[i]
					addAudit(data, session.User.Username, "handle", "system_alert", id, updated.AlertNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("系统告警不存在")
		})
		a.respondMutation(w, err, updated, "product_ops.alert.handled")
		return
	}
	writeError(w, http.StatusNotFound, "unknown product alert route")
}

func (a *App) productOpsRenewals(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "integrations" && r.Method == http.MethodPost {
		var req ProductRenewalIntegration
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid renewal integration payload")
			return
		}
		var saved ProductRenewalIntegration
		err := a.store.Mutate(func(data *AppData) error {
			normalized, err := normalizeProductRenewalIntegration(req, session.User.DisplayName)
			if err != nil {
				return err
			}
			for i := range data.ProductRenewalIntegrations {
				if data.ProductRenewalIntegrations[i].ID == normalized.ID || (normalized.Code != "" && data.ProductRenewalIntegrations[i].Code == normalized.Code) {
					normalized.ID = data.ProductRenewalIntegrations[i].ID
					normalized.IntegrationNo = fallback(normalized.IntegrationNo, data.ProductRenewalIntegrations[i].IntegrationNo)
					normalized.CreatedBy = fallback(normalized.CreatedBy, data.ProductRenewalIntegrations[i].CreatedBy)
					normalized.CreatedAt = fallback(normalized.CreatedAt, data.ProductRenewalIntegrations[i].CreatedAt)
					normalized.LastSyncAt = data.ProductRenewalIntegrations[i].LastSyncAt
					normalized.LastError = data.ProductRenewalIntegrations[i].LastError
					data.ProductRenewalIntegrations[i] = normalized
					saved = normalized
					addAudit(data, session.User.Username, "update", "renewal_integration", saved.ID, saved.Code+" "+saved.Name, clientIP(r))
					return nil
				}
			}
			normalized.ID = nextID(data, "renewalIntegration")
			normalized.IntegrationNo = number("RI", normalized.ID)
			data.ProductRenewalIntegrations = append(data.ProductRenewalIntegrations, normalized)
			saved = normalized
			addAudit(data, session.User.Username, "create", "renewal_integration", saved.ID, saved.Code+" "+saved.Name, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "product_ops.renewal.integration.saved")
		return
	}
	if len(parts) == 3 && parts[0] == "sync-records" && parts[2] == "retry" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var updated ProductRenewalSyncRecord
		err := a.store.Mutate(func(data *AppData) error {
			result, err := retryProductRenewalSyncRecord(data, id)
			updated = result
			if err == nil {
				addAudit(data, session.User.Username, "retry", "renewal_sync_record", id, updated.SyncNo+" "+updated.Status, clientIP(r))
			}
			return err
		})
		a.respondMutation(w, err, updated, "product_ops.renewal.sync.retried")
		return
	}
	if len(parts) == 0 && r.Method == http.MethodPost {
		var req ProductRenewalTask
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid renewal task payload")
			return
		}
		var saved ProductRenewalTask
		err := a.store.Mutate(func(data *AppData) error {
			req.CustomerName = strings.TrimSpace(req.CustomerName)
			req.LicenseID = strings.TrimSpace(req.LicenseID)
			if req.InstanceID != 0 {
				for i := range data.ProductInstances {
					if data.ProductInstances[i].ID == req.InstanceID {
						req.CustomerName = fallback(req.CustomerName, data.ProductInstances[i].CustomerName)
						req.LicenseID = fallback(req.LicenseID, data.ProductInstances[i].LicenseID)
						data.ProductInstances[i].RenewalStage = fallback(req.Stage, data.ProductInstances[i].RenewalStage)
						data.ProductInstances[i].RenewalOwner = fallback(req.Owner, data.ProductInstances[i].RenewalOwner)
						break
					}
				}
			}
			if req.CustomerName == "" {
				return fmt.Errorf("续费任务必须包含客户名称")
			}
			req.Stage = fallback(strings.TrimSpace(req.Stage), "待跟进")
			req.Status = fallback(strings.TrimSpace(req.Status), "open")
			req.Owner = fallback(strings.TrimSpace(req.Owner), session.User.DisplayName)
			req.Currency = fallback(strings.TrimSpace(req.Currency), "CNY")
			req.RiskLevel = fallback(strings.TrimSpace(req.RiskLevel), productRenewalRisk(req.DueDate))
			if req.CreatedAt == "" {
				req.CreatedAt = nowString()
			}
			for i := range data.ProductRenewalTasks {
				if data.ProductRenewalTasks[i].ID == req.ID || (req.TaskNo != "" && data.ProductRenewalTasks[i].TaskNo == req.TaskNo) {
					req.ID = data.ProductRenewalTasks[i].ID
					req.TaskNo = fallback(req.TaskNo, data.ProductRenewalTasks[i].TaskNo)
					data.ProductRenewalTasks[i] = req
					saved = req
					addAudit(data, session.User.Username, "update", "renewal_task", req.ID, req.CustomerName+" "+req.Stage, clientIP(r))
					return nil
				}
			}
			req.ID = nextID(data, "renewalTask")
			req.TaskNo = fallback(req.TaskNo, number("RN", req.ID))
			data.ProductRenewalTasks = append(data.ProductRenewalTasks, req)
			saved = req
			addAudit(data, session.User.Username, "create", "renewal_task", req.ID, req.CustomerName+" "+req.Stage, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "product_ops.renewal.saved")
		return
	}
	if len(parts) == 2 && parts[1] == "quote" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req ProductRenewalQuote
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid renewal quote payload")
			return
		}
		var created ProductRenewalQuote
		err := a.store.Mutate(func(data *AppData) error {
			taskIndex := productRenewalTaskIndex(*data, id)
			if taskIndex < 0 {
				return fmt.Errorf("续费任务不存在")
			}
			task := &data.ProductRenewalTasks[taskIndex]
			now := nowString()
			quoteID := nextID(data, "renewalQuote")
			created = ProductRenewalQuote{
				ID:           quoteID,
				QuoteNo:      number("RQ", quoteID),
				TaskID:       task.ID,
				InstanceID:   task.InstanceID,
				CustomerName: fallback(strings.TrimSpace(req.CustomerName), task.CustomerName),
				LicenseID:    fallback(strings.TrimSpace(req.LicenseID), task.LicenseID),
				Amount:       req.Amount,
				Currency:     fallback(strings.TrimSpace(req.Currency), fallback(task.Currency, "CNY")),
				Modules:      req.Modules,
				NewExpiresAt: fallback(strings.TrimSpace(req.NewExpiresAt), task.DueDate),
				Status:       fallback(strings.TrimSpace(req.Status), "sent"),
				PreparedBy:   session.User.DisplayName,
				PreparedAt:   now,
				Remark:       strings.TrimSpace(req.Remark),
			}
			if created.Amount <= 0 {
				created.Amount = task.Amount
			}
			if created.Amount <= 0 {
				return fmt.Errorf("续费报价金额必须大于 0")
			}
			if len(created.Modules) == 0 {
				created.Modules = []string{"license", "update", "support"}
			}
			task.Stage = "报价已发送"
			task.Amount = created.Amount
			task.Currency = created.Currency
			task.LastContactAt = now
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			data.ProductRenewalQuotes = append(data.ProductRenewalQuotes, created)
			addAudit(data, session.User.Username, "quote", "renewal_task", task.ID, created.QuoteNo+" "+created.CustomerName, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, created, "product_ops.renewal.quote.created")
		return
	}
	if len(parts) == 2 && parts[1] == "contract" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req ProductRenewalContract
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid renewal contract payload")
			return
		}
		var created ProductRenewalContract
		err := a.store.Mutate(func(data *AppData) error {
			taskIndex := productRenewalTaskIndex(*data, id)
			if taskIndex < 0 {
				return fmt.Errorf("续费任务不存在")
			}
			task := &data.ProductRenewalTasks[taskIndex]
			quoteIndex := latestProductRenewalQuoteIndex(*data, task.ID, req.QuoteID)
			if quoteIndex < 0 {
				return fmt.Errorf("请先生成续费报价")
			}
			quote := &data.ProductRenewalQuotes[quoteIndex]
			now := nowString()
			contractID := nextID(data, "renewalContract")
			created = ProductRenewalContract{
				ID:           contractID,
				ContractNo:   fallback(strings.TrimSpace(req.ContractNo), number("RC", contractID)),
				TaskID:       task.ID,
				QuoteID:      quote.ID,
				InstanceID:   task.InstanceID,
				CustomerName: task.CustomerName,
				LicenseID:    task.LicenseID,
				Amount:       nonZero(req.Amount, quote.Amount),
				Currency:     fallback(strings.TrimSpace(req.Currency), quote.Currency),
				Status:       fallback(strings.TrimSpace(req.Status), "signed"),
				SignedBy:     fallback(strings.TrimSpace(req.SignedBy), session.User.DisplayName),
				SignedAt:     fallback(strings.TrimSpace(req.SignedAt), now),
				CreatedBy:    session.User.DisplayName,
				CreatedAt:    now,
				Remark:       strings.TrimSpace(req.Remark),
			}
			quote.Status = "approved"
			quote.ApprovedBy = session.User.DisplayName
			quote.ApprovedAt = now
			task.Stage = "合同已签"
			task.Amount = created.Amount
			task.Currency = created.Currency
			task.LastContactAt = now
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			data.ProductRenewalContracts = append(data.ProductRenewalContracts, created)
			addAudit(data, session.User.Username, "contract", "renewal_task", task.ID, created.ContractNo+" "+created.CustomerName, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, created, "product_ops.renewal.contract.created")
		return
	}
	if len(parts) == 2 && parts[1] == "payment" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req ProductRenewalPayment
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid renewal payment payload")
			return
		}
		var created ProductRenewalPayment
		err := a.store.Mutate(func(data *AppData) error {
			taskIndex := productRenewalTaskIndex(*data, id)
			if taskIndex < 0 {
				return fmt.Errorf("续费任务不存在")
			}
			task := &data.ProductRenewalTasks[taskIndex]
			contractIndex := latestProductRenewalContractIndex(*data, task.ID, req.ContractID)
			if contractIndex < 0 {
				return fmt.Errorf("请先确认续费合同")
			}
			contract := &data.ProductRenewalContracts[contractIndex]
			now := nowString()
			amount := req.Amount
			if amount <= 0 {
				amount = contract.Amount - productRenewalPaidAmount(*data, contract.ID)
			}
			if amount <= 0 {
				return fmt.Errorf("续费回款金额必须大于 0")
			}
			paymentID := nextID(data, "renewalPayment")
			created = ProductRenewalPayment{
				ID:           paymentID,
				PaymentNo:    number("RP", paymentID),
				TaskID:       task.ID,
				ContractID:   contract.ID,
				InstanceID:   task.InstanceID,
				CustomerName: task.CustomerName,
				Amount:       amount,
				Currency:     fallback(strings.TrimSpace(req.Currency), contract.Currency),
				Method:       fallback(strings.TrimSpace(req.Method), "bank"),
				Status:       fallback(strings.TrimSpace(req.Status), "paid"),
				PaidAt:       fallback(strings.TrimSpace(req.PaidAt), now),
				CreatedBy:    session.User.DisplayName,
				CreatedAt:    now,
				Remark:       strings.TrimSpace(req.Remark),
			}
			data.ProductRenewalPayments = append(data.ProductRenewalPayments, created)
			paid := productRenewalPaidAmount(*data, contract.ID)
			if paid >= contract.Amount {
				contract.Status = "paid"
				task.Status = "closed"
				task.Stage = "已回款"
				task.ClosedAt = now
			} else {
				contract.Status = "partial_paid"
				task.Stage = "部分回款"
			}
			task.LastContactAt = now
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			sync, _ := enqueueProductRenewalSyncRecord(data, productRenewalSyncRequest{
				Scenario:     "payment",
				ResourceType: "payment",
				ResourceID:   created.ID,
				ResourceNo:   created.PaymentNo,
				Task:         *task,
				Action:       "confirm",
				Code:         created.Method,
				Payload: map[string]interface{}{
					"payment":  created,
					"contract": *contract,
					"task":     *task,
				},
			})
			if sync.ID != 0 {
				created.Remark = fallback(created.Remark, "外部财务同步 "+sync.Status)
			}
			addAudit(data, session.User.Username, "payment", "renewal_task", task.ID, created.PaymentNo+" "+created.CustomerName, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, created, "product_ops.renewal.payment.created")
		return
	}
	if len(parts) == 2 && parts[1] == "approval" && r.Method == http.MethodPost {
		a.productOpsRenewalApproval(w, r, session, parts[0])
		return
	}
	if len(parts) == 2 && parts[1] == "invoice" && r.Method == http.MethodPost {
		a.productOpsRenewalInvoice(w, r, session, parts[0])
		return
	}
	if len(parts) == 2 && parts[1] == "esign" && r.Method == http.MethodPost {
		a.productOpsRenewalESign(w, r, session, parts[0])
		return
	}
	if len(parts) == 2 && parts[1] == "close" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Remark string `json:"remark"`
		}
		_ = readJSON(r, &req)
		var updated ProductRenewalTask
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.ProductRenewalTasks {
				if data.ProductRenewalTasks[i].ID == id {
					data.ProductRenewalTasks[i].Status = "closed"
					data.ProductRenewalTasks[i].ClosedAt = nowString()
					data.ProductRenewalTasks[i].Stage = "已续费"
					if strings.TrimSpace(req.Remark) != "" {
						data.ProductRenewalTasks[i].Remark = strings.TrimSpace(req.Remark)
					}
					productRenewalSyncInstanceStage(data, data.ProductRenewalTasks[i].InstanceID, data.ProductRenewalTasks[i].Stage)
					updated = data.ProductRenewalTasks[i]
					addAudit(data, session.User.Username, "close", "renewal_task", id, updated.TaskNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("续费任务不存在")
		})
		a.respondMutation(w, err, updated, "product_ops.renewal.closed")
		return
	}
	writeError(w, http.StatusNotFound, "unknown renewal task route")
}

func productRenewalTaskIndex(data AppData, id int64) int {
	for i, task := range data.ProductRenewalTasks {
		if task.ID == id {
			return i
		}
	}
	return -1
}

func latestProductRenewalQuoteIndex(data AppData, taskID, quoteID int64) int {
	index := -1
	for i, quote := range data.ProductRenewalQuotes {
		if quoteID > 0 {
			if quote.ID == quoteID && quote.TaskID == taskID {
				return i
			}
			continue
		}
		if quote.TaskID == taskID && (index < 0 || quote.PreparedAt > data.ProductRenewalQuotes[index].PreparedAt) {
			index = i
		}
	}
	return index
}

func latestProductRenewalContractIndex(data AppData, taskID, contractID int64) int {
	index := -1
	for i, contract := range data.ProductRenewalContracts {
		if contractID > 0 {
			if contract.ID == contractID && contract.TaskID == taskID {
				return i
			}
			continue
		}
		if contract.TaskID == taskID && (index < 0 || contract.CreatedAt > data.ProductRenewalContracts[index].CreatedAt) {
			index = i
		}
	}
	return index
}

func productRenewalPaidAmount(data AppData, contractID int64) float64 {
	var total float64
	for _, payment := range data.ProductRenewalPayments {
		if payment.ContractID == contractID && payment.Status == "paid" {
			total += payment.Amount
		}
	}
	return total
}

func productRenewalSyncInstanceStage(data *AppData, instanceID int64, stage string) {
	if instanceID == 0 {
		return
	}
	for i := range data.ProductInstances {
		if data.ProductInstances[i].ID == instanceID {
			data.ProductInstances[i].RenewalStage = stage
			return
		}
	}
}

func (a *App) productOpsRollouts(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodPost {
		var req struct {
			UpdateID          int64   `json:"updateId"`
			Strategy          string  `json:"strategy"`
			TargetInstanceIDs []int64 `json:"targetInstanceIds"`
			Remark            string  `json:"remark"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid rollout payload")
			return
		}
		var created ProductUpdateRollout
		err := a.store.Mutate(func(data *AppData) error {
			var update UpdatePackage
			for _, item := range data.Updates {
				if item.ID == req.UpdateID {
					update = item
					break
				}
			}
			if update.ID == 0 {
				return fmt.Errorf("更新包不存在")
			}
			if !updatePackageVerified(update) {
				return fmt.Errorf("更新包验签失败")
			}
			targets, err := productRolloutTargets(data.ProductInstances, req.TargetInstanceIDs)
			if err != nil {
				return err
			}
			now := nowString()
			id := nextID(data, "updateRollout")
			created = ProductUpdateRollout{
				ID:           id,
				RolloutNo:    number("UR", id),
				UpdateID:     update.ID,
				Version:      update.Version,
				Component:    fallback(update.Component, "server"),
				Strategy:     fallback(strings.ToLower(strings.TrimSpace(req.Strategy)), "gray"),
				Status:       "pending",
				TotalTargets: len(targets),
				CreatedBy:    session.User.DisplayName,
				CreatedAt:    now,
				Remark:       strings.TrimSpace(req.Remark),
			}
			for _, instance := range targets {
				itemID := nextID(data, "updateRolloutItem")
				created.Items = append(created.Items, ProductUpdateRolloutItem{
					ID:           itemID,
					InstanceID:   instance.ID,
					CustomerName: instance.CustomerName,
					FromVersion:  productRolloutInstanceVersion(instance, created.Component),
					ToVersion:    update.Version,
					Status:       "pending",
				})
			}
			data.ProductUpdateRollouts = append(data.ProductUpdateRollouts, created)
			addAudit(data, session.User.Username, "create", "update_rollout", created.ID, created.RolloutNo+" "+created.Version, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, created, "product_ops.rollout.created")
		return
	}
	if len(parts) == 2 && parts[1] == "advance" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Action     string `json:"action"`
			InstanceID int64  `json:"instanceId"`
			Message    string `json:"message"`
		}
		_ = readJSON(r, &req)
		var updated ProductUpdateRollout
		err := a.store.Mutate(func(data *AppData) error {
			action := fallback(strings.ToLower(strings.TrimSpace(req.Action)), "apply")
			if action != "apply" && action != "fail" && action != "rollback" {
				return fmt.Errorf("批次推进动作必须是 apply、fail 或 rollback")
			}
			for i := range data.ProductUpdateRollouts {
				if data.ProductUpdateRollouts[i].ID != id {
					continue
				}
				rollout := &data.ProductUpdateRollouts[i]
				itemIndex := productRolloutItemIndex(*rollout, req.InstanceID)
				if itemIndex < 0 {
					return fmt.Errorf("批次中没有可推进的客户实例")
				}
				now := nowString()
				if rollout.Status == "pending" {
					rollout.Status = "running"
					rollout.StartedAt = now
				}
				item := &rollout.Items[itemIndex]
				if item.StartedAt == "" {
					item.StartedAt = now
				}
				item.Message = fallback(strings.TrimSpace(req.Message), productRolloutDefaultMessage(action))
				switch action {
				case "apply":
					item.Status = "applied"
					item.AppliedAt = now
					if err := productRolloutSyncInstance(data, item.InstanceID, rollout.Component, rollout.Version); err != nil {
						return err
					}
				case "fail":
					item.Status = "failed"
				case "rollback":
					item.Status = "rolled_back"
					item.RolledBackAt = now
					if err := productRolloutSyncInstance(data, item.InstanceID, rollout.Component, item.FromVersion); err != nil {
						return err
					}
				}
				recomputeProductRollout(rollout)
				updated = *rollout
				addAudit(data, session.User.Username, action, "update_rollout", rollout.ID, rollout.RolloutNo+" "+item.CustomerName, clientIP(r))
				return nil
			}
			return fmt.Errorf("更新批次不存在")
		})
		a.respondMutation(w, err, updated, "product_ops.rollout.changed")
		return
	}
	if len(parts) == 2 && parts[1] == "execute" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req productUpdateExecutionRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid update execution payload")
			return
		}
		var execution ProductUpdateExecution
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.ProductUpdateRollouts {
				if data.ProductUpdateRollouts[i].ID != id {
					continue
				}
				rollout := &data.ProductUpdateRollouts[i]
				var update UpdatePackage
				for _, item := range data.Updates {
					if item.ID == rollout.UpdateID {
						update = item
						break
					}
				}
				if update.ID == 0 {
					return fmt.Errorf("更新包不存在")
				}
				result, err := executeProductUpdate(data, rollout, update, req, session.User.DisplayName)
				execution = result
				if result.ID != 0 {
					addAudit(data, session.User.Username, result.Action, "update_execution", result.ID, result.ExecutionNo+" "+result.CustomerName, clientIP(r))
				}
				return err
			}
			return fmt.Errorf("更新批次不存在")
		})
		a.respondMutation(w, err, execution, "product_ops.rollout.executed")
		return
	}
	if len(parts) == 2 && parts[1] == "system-update-tasks" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req productSystemUpdateTaskRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid system update task payload")
			return
		}
		var task ProductSystemUpdateTask
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.ProductUpdateRollouts {
				if data.ProductUpdateRollouts[i].ID != id {
					continue
				}
				rollout := &data.ProductUpdateRollouts[i]
				var update UpdatePackage
				for _, item := range data.Updates {
					if item.ID == rollout.UpdateID {
						update = item
						break
					}
				}
				if update.ID == 0 {
					return fmt.Errorf("更新包不存在")
				}
				result, err := queueProductSystemUpdateTask(data, rollout, update, req, session.User.DisplayName)
				task = result
				if result.ID != 0 {
					addAudit(data, session.User.Username, result.Action, "system_update_task", result.ID, result.TaskNo+" "+result.CustomerName, clientIP(r))
				}
				return err
			}
			return fmt.Errorf("更新批次不存在")
		})
		a.respondMutation(w, err, task, "product_ops.system_update.task.created")
		return
	}
	writeError(w, http.StatusNotFound, "unknown update rollout route")
}

func productRolloutTargets(instances []ProductInstance, ids []int64) ([]ProductInstance, error) {
	if len(ids) == 0 {
		if len(instances) == 0 {
			return nil, fmt.Errorf("没有可发布的客户实例")
		}
		return append([]ProductInstance{}, instances...), nil
	}
	wanted := map[int64]bool{}
	for _, id := range ids {
		if id > 0 {
			wanted[id] = true
		}
	}
	if len(wanted) == 0 {
		return nil, fmt.Errorf("目标客户实例不能为空")
	}
	targets := make([]ProductInstance, 0, len(wanted))
	for _, instance := range instances {
		if wanted[instance.ID] {
			targets = append(targets, instance)
			delete(wanted, instance.ID)
		}
	}
	if len(wanted) > 0 {
		return nil, fmt.Errorf("目标客户实例不存在")
	}
	return targets, nil
}

func productRolloutItemIndex(rollout ProductUpdateRollout, instanceID int64) int {
	if instanceID > 0 {
		for i, item := range rollout.Items {
			if item.InstanceID == instanceID {
				return i
			}
		}
		return -1
	}
	for i, item := range rollout.Items {
		if item.Status == "pending" || item.Status == "running" {
			return i
		}
	}
	return -1
}

func productRolloutSyncInstance(data *AppData, instanceID int64, component, version string) error {
	for i := range data.ProductInstances {
		if data.ProductInstances[i].ID != instanceID {
			continue
		}
		switch component {
		case "client":
			data.ProductInstances[i].ClientVersion = version
		case "server":
			data.ProductInstances[i].ServerVersion = version
		case "all":
			clientVersion, serverVersion := productRolloutSplitCombinedVersion(version)
			data.ProductInstances[i].ClientVersion = fallback(clientVersion, version)
			data.ProductInstances[i].ServerVersion = fallback(serverVersion, version)
		default:
			data.ProductInstances[i].ServerVersion = version
		}
		data.ProductInstances[i].LastHeartbeatAt = nowString()
		return nil
	}
	return fmt.Errorf("客户实例不存在")
}

func productRolloutInstanceVersion(instance ProductInstance, component string) string {
	switch component {
	case "client":
		return instance.ClientVersion
	case "server":
		return instance.ServerVersion
	case "all":
		if instance.ClientVersion == instance.ServerVersion {
			return instance.ClientVersion
		}
		return "client:" + instance.ClientVersion + " server:" + instance.ServerVersion
	default:
		return instance.ServerVersion
	}
}

func productRolloutSplitCombinedVersion(value string) (string, string) {
	if !strings.Contains(value, "client:") || !strings.Contains(value, "server:") {
		return value, value
	}
	var clientVersion, serverVersion string
	for _, field := range strings.Fields(value) {
		if strings.HasPrefix(field, "client:") {
			clientVersion = strings.TrimPrefix(field, "client:")
		}
		if strings.HasPrefix(field, "server:") {
			serverVersion = strings.TrimPrefix(field, "server:")
		}
	}
	return clientVersion, serverVersion
}

func recomputeProductRollout(rollout *ProductUpdateRollout) {
	rollout.TotalTargets = len(rollout.Items)
	rollout.AppliedTargets = 0
	rollout.FailedTargets = 0
	pending := 0
	for _, item := range rollout.Items {
		switch item.Status {
		case "applied":
			rollout.AppliedTargets++
		case "failed":
			rollout.FailedTargets++
		case "pending", "running":
			pending++
		}
	}
	if rollout.AppliedTargets == rollout.TotalTargets && rollout.TotalTargets > 0 {
		rollout.Status = "completed"
		rollout.CompletedAt = fallback(rollout.CompletedAt, nowString())
		return
	}
	if pending == 0 && rollout.FailedTargets > 0 {
		rollout.Status = "partial_failed"
		rollout.CompletedAt = fallback(rollout.CompletedAt, nowString())
		return
	}
	if rollout.StartedAt != "" {
		rollout.Status = "running"
	}
}

func productRolloutDefaultMessage(action string) string {
	switch action {
	case "apply":
		return "已应用更新包"
	case "fail":
		return "更新推进失败"
	case "rollback":
		return "已回滚到原版本"
	default:
		return "批次状态已更新"
	}
}

func buildProductOpsOverview(data AppData, runtime RuntimeStatus) ProductOpsOverview {
	licensePortal := buildLicensePortalOverview(data)
	instances := enrichProductInstances(data.ProductInstances, licensePortal)
	alerts := append([]SystemAlert{}, data.SystemAlerts...)
	renewals := append([]ProductRenewalTask{}, data.ProductRenewalTasks...)
	quotes := append([]ProductRenewalQuote{}, data.ProductRenewalQuotes...)
	contracts := append([]ProductRenewalContract{}, data.ProductRenewalContracts...)
	payments := append([]ProductRenewalPayment{}, data.ProductRenewalPayments...)
	approvals := append([]ProductRenewalApproval{}, data.ProductRenewalApprovals...)
	invoices := append([]ProductRenewalInvoice{}, data.ProductRenewalInvoices...)
	esigns := append([]ProductRenewalESign{}, data.ProductRenewalESigns...)
	renewalIntegrations := append([]ProductRenewalIntegration{}, data.ProductRenewalIntegrations...)
	renewalSyncRecords := append([]ProductRenewalSyncRecord{}, data.ProductRenewalSyncRecords...)
	probeReports := append([]ProductProbeReport{}, data.ProductProbeReports...)
	telemetryEvents := append([]ProductTelemetryEvent{}, data.ProductTelemetryEvents...)
	monitoringIntegrations := append([]ProductMonitoringIntegration{}, data.ProductMonitoringIntegrations...)
	alertRules := append([]ProductAlertRule{}, data.ProductAlertRules...)
	monitoringEvents := append([]ProductMonitoringEvent{}, data.ProductMonitoringEvents...)
	alertPolicies := append([]ProductAlertPolicy{}, data.ProductAlertPolicies...)
	alertChannels := append([]ProductAlertChannel{}, data.ProductAlertChannels...)
	alertNotifications := append([]ProductAlertNotification{}, data.ProductAlertNotifications...)
	rollouts := append([]ProductUpdateRollout{}, data.ProductUpdateRollouts...)
	executions := append([]ProductUpdateExecution{}, data.ProductUpdateExecutions...)
	systemTasks := append([]ProductSystemUpdateTask{}, data.ProductSystemUpdateTasks...)
	updates := sanitizeUpdatePackagesForResponse(data.Updates)
	sort.Slice(alerts, func(i, j int) bool { return alerts[i].LastSeenAt > alerts[j].LastSeenAt })
	sort.Slice(renewals, func(i, j int) bool {
		if renewals[i].Status != renewals[j].Status {
			return renewals[i].Status < renewals[j].Status
		}
		return renewals[i].DueDate < renewals[j].DueDate
	})
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].PreparedAt > quotes[j].PreparedAt })
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].CreatedAt > contracts[j].CreatedAt })
	sort.Slice(payments, func(i, j int) bool { return payments[i].CreatedAt > payments[j].CreatedAt })
	sort.Slice(approvals, func(i, j int) bool { return approvals[i].RequestedAt > approvals[j].RequestedAt })
	sort.Slice(invoices, func(i, j int) bool { return invoices[i].CreatedAt > invoices[j].CreatedAt })
	sort.Slice(esigns, func(i, j int) bool { return esigns[i].SentAt > esigns[j].SentAt })
	sort.Slice(renewalIntegrations, func(i, j int) bool { return renewalIntegrations[i].CreatedAt > renewalIntegrations[j].CreatedAt })
	sort.Slice(renewalSyncRecords, func(i, j int) bool { return renewalSyncRecords[i].CreatedAt > renewalSyncRecords[j].CreatedAt })
	sort.Slice(updates, func(i, j int) bool { return updates[i].CreatedAt > updates[j].CreatedAt })
	sort.Slice(probeReports, func(i, j int) bool { return probeReports[i].ReceivedAt > probeReports[j].ReceivedAt })
	sort.Slice(telemetryEvents, func(i, j int) bool { return telemetryEvents[i].ReceivedAt > telemetryEvents[j].ReceivedAt })
	sort.Slice(monitoringIntegrations, func(i, j int) bool { return monitoringIntegrations[i].CreatedAt > monitoringIntegrations[j].CreatedAt })
	sort.Slice(alertRules, func(i, j int) bool { return alertRules[i].CreatedAt > alertRules[j].CreatedAt })
	sort.Slice(monitoringEvents, func(i, j int) bool { return monitoringEvents[i].ReceivedAt > monitoringEvents[j].ReceivedAt })
	sort.Slice(alertPolicies, func(i, j int) bool { return alertPolicies[i].CreatedAt > alertPolicies[j].CreatedAt })
	sort.Slice(alertChannels, func(i, j int) bool { return alertChannels[i].CreatedAt > alertChannels[j].CreatedAt })
	sort.Slice(alertNotifications, func(i, j int) bool { return alertNotifications[i].CreatedAt > alertNotifications[j].CreatedAt })
	sort.Slice(executions, func(i, j int) bool { return executions[i].StartedAt > executions[j].StartedAt })
	sort.Slice(systemTasks, func(i, j int) bool { return systemTasks[i].CreatedAt > systemTasks[j].CreatedAt })
	if len(quotes) > 20 {
		quotes = quotes[:20]
	}
	if len(contracts) > 20 {
		contracts = contracts[:20]
	}
	if len(payments) > 20 {
		payments = payments[:20]
	}
	if len(approvals) > 20 {
		approvals = approvals[:20]
	}
	if len(invoices) > 20 {
		invoices = invoices[:20]
	}
	if len(esigns) > 20 {
		esigns = esigns[:20]
	}
	if len(renewalIntegrations) > 20 {
		renewalIntegrations = renewalIntegrations[:20]
	}
	if len(renewalSyncRecords) > 50 {
		renewalSyncRecords = renewalSyncRecords[:50]
	}
	if len(probeReports) > 20 {
		probeReports = probeReports[:20]
	}
	if len(telemetryEvents) > 20 {
		telemetryEvents = telemetryEvents[:20]
	}
	if len(monitoringIntegrations) > 20 {
		monitoringIntegrations = monitoringIntegrations[:20]
	}
	if len(alertRules) > 20 {
		alertRules = alertRules[:20]
	}
	if len(monitoringEvents) > 20 {
		monitoringEvents = monitoringEvents[:20]
	}
	if len(alertPolicies) > 20 {
		alertPolicies = alertPolicies[:20]
	}
	if len(alertChannels) > 20 {
		alertChannels = alertChannels[:20]
	}
	if len(alertNotifications) > 30 {
		alertNotifications = alertNotifications[:30]
	}
	if len(executions) > 20 {
		executions = executions[:20]
	}
	if len(systemTasks) > 20 {
		systemTasks = systemTasks[:20]
	}
	sort.Slice(rollouts, func(i, j int) bool { return rollouts[i].CreatedAt > rollouts[j].CreatedAt })
	if len(rollouts) > 12 {
		rollouts = rollouts[:12]
	}
	overview := ProductOpsOverview{
		Instances: instances, Alerts: alerts, RenewalTasks: renewals, RenewalQuotes: quotes, RenewalContracts: contracts, RenewalPayments: payments, RenewalApprovals: approvals, RenewalInvoices: invoices, RenewalESigns: esigns, RenewalIntegrations: renewalIntegrations, RenewalSyncRecords: renewalSyncRecords, ProbeReports: probeReports, TelemetryEvents: telemetryEvents, MonitoringIntegrations: monitoringIntegrations, AlertRules: alertRules, MonitoringEvents: monitoringEvents, AlertPolicies: alertPolicies, AlertChannels: alertChannels, AlertNotifications: alertNotifications, UpdateRollouts: rollouts, UpdateExecutions: executions, SystemUpdateTasks: systemTasks, Updates: updates, LicensePortal: licensePortal,
		RecentEvents: licensePortal.RecentEvents, Runtime: runtime,
	}
	overview.KPIs.Customers = len(instances)
	for _, instance := range instances {
		switch instance.Status {
		case "online":
			overview.KPIs.OnlineInstances++
		case "degraded", "offline":
			overview.KPIs.DegradedInstances++
		}
		if instance.DaysToExpire <= 45 || instance.LicenseRisk == "warning" || instance.LicenseRisk == "expired" {
			overview.KPIs.ExpiringLicenses++
		}
	}
	for _, alert := range alerts {
		if alert.Status != "handled" && alert.Status != "closed" {
			overview.KPIs.OpenAlerts++
			if alert.Severity == "critical" {
				overview.KPIs.CriticalAlerts++
			}
			if alert.SuppressedUntil != "" {
				overview.KPIs.SuppressedAlerts++
			}
			if alert.EscalationLevel != "" {
				overview.KPIs.EscalatedAlerts++
			}
		}
	}
	for _, task := range renewals {
		if task.Status != "closed" && task.Status != "cancelled" {
			overview.KPIs.OpenRenewals++
			if productRenewalRisk(task.DueDate) == "critical" || task.RiskLevel == "critical" {
				overview.KPIs.OverdueRenewals++
			}
		}
	}
	for _, quote := range data.ProductRenewalQuotes {
		if quote.Status != "approved" && quote.Status != "rejected" && quote.Status != "cancelled" {
			overview.KPIs.PendingRenewalQuotes++
		}
	}
	for _, contract := range data.ProductRenewalContracts {
		if contract.Status != "paid" && contract.Status != "cancelled" {
			overview.KPIs.PendingRenewalContracts++
		}
	}
	for _, payment := range data.ProductRenewalPayments {
		if payment.Status == "paid" {
			overview.KPIs.PaidRenewalAmount += payment.Amount
		}
	}
	for _, approval := range data.ProductRenewalApprovals {
		if approval.Status == "pending" {
			overview.KPIs.PendingRenewalApprovals++
		}
	}
	for _, invoice := range data.ProductRenewalInvoices {
		if invoice.Status == "issued" {
			overview.KPIs.IssuedRenewalInvoices++
		}
	}
	for _, sign := range data.ProductRenewalESigns {
		if sign.Status == "sent" || sign.Status == "pending" {
			overview.KPIs.PendingRenewalESigns++
		}
	}
	for _, integration := range data.ProductRenewalIntegrations {
		if integration.Status == "active" {
			overview.KPIs.ActiveRenewalIntegrations++
		}
	}
	overview.KPIs.RenewalSyncRecords = len(data.ProductRenewalSyncRecords)
	for _, record := range data.ProductRenewalSyncRecords {
		switch record.Status {
		case "failed":
			overview.KPIs.FailedRenewalSyncRecords++
		case "pending", "running":
			overview.KPIs.PendingRenewalSyncRecords++
		}
	}
	overview.KPIs.ProbeReports = len(data.ProductProbeReports)
	for _, report := range data.ProductProbeReports {
		if report.AlertRaised || report.Status != "healthy" {
			overview.KPIs.UnhealthyProbes++
		}
	}
	overview.KPIs.TelemetryEvents = len(data.ProductTelemetryEvents)
	for _, event := range data.ProductTelemetryEvents {
		if event.AlertRaised || event.Severity == "critical" {
			overview.KPIs.CriticalTelemetryEvents++
		}
	}
	for _, integration := range data.ProductMonitoringIntegrations {
		if integration.Status == "active" {
			overview.KPIs.MonitoringIntegrations++
		}
	}
	for _, rule := range data.ProductAlertRules {
		if rule.Status == "active" {
			overview.KPIs.ActiveAlertRules++
		}
	}
	overview.KPIs.MonitoringEvents = len(data.ProductMonitoringEvents)
	for _, event := range data.ProductMonitoringEvents {
		if event.AlertRaised || event.Severity == "critical" {
			overview.KPIs.MonitoringAlerts++
		}
	}
	for _, policy := range data.ProductAlertPolicies {
		if policy.Status == "active" {
			overview.KPIs.ActiveAlertPolicies++
		}
	}
	for _, channel := range data.ProductAlertChannels {
		if channel.Status == "active" {
			overview.KPIs.ActiveAlertChannels++
		}
	}
	overview.KPIs.AlertNotifications = len(data.ProductAlertNotifications)
	for _, notification := range data.ProductAlertNotifications {
		switch notification.Status {
		case "failed":
			overview.KPIs.FailedAlertNotifications++
		case "pending":
			overview.KPIs.PendingAlertNotifications++
		}
	}
	for _, rollout := range data.ProductUpdateRollouts {
		if rollout.Status == "pending" || rollout.Status == "running" {
			overview.KPIs.ActiveRollouts++
		}
		for _, item := range rollout.Items {
			if item.Status == "failed" {
				overview.KPIs.FailedRolloutItems++
			}
		}
	}
	overview.KPIs.UpdateExecutions = len(data.ProductUpdateExecutions)
	for _, execution := range data.ProductUpdateExecutions {
		if execution.Status == "failed" {
			overview.KPIs.FailedUpdateExecutions++
		}
	}
	overview.KPIs.SystemUpdateTasks = len(data.ProductSystemUpdateTasks)
	for _, task := range data.ProductSystemUpdateTasks {
		switch task.Status {
		case "queued", "assigned", "running":
			overview.KPIs.RunningSystemUpdateTasks++
		case "failed":
			overview.KPIs.FailedSystemUpdateTasks++
		}
	}
	for _, update := range updates {
		switch update.Component {
		case "client":
			overview.KPIs.ClientUpdatePackages++
		case "server":
			overview.KPIs.ServerUpdatePackages++
		default:
			overview.KPIs.ClientUpdatePackages++
			overview.KPIs.ServerUpdatePackages++
		}
		if update.Status == "available" || update.Status == "gray" {
			overview.KPIs.AvailableUpdates++
		}
	}
	return overview
}

func productRenewalRisk(dueDate string) string {
	days := licenseDaysToExpire(dueDate)
	if days < 0 {
		return "critical"
	}
	if days <= 15 {
		return "critical"
	}
	if days <= 45 {
		return "warning"
	}
	return "normal"
}

func enrichProductInstances(instances []ProductInstance, portal LicensePortalOverview) []ProductInstance {
	byLicense := map[string]LicensePortalCustomer{}
	byWatermark := map[string]LicensePortalCustomer{}
	byCustomer := map[string]LicensePortalCustomer{}
	for _, customer := range portal.Customers {
		byLicense[customer.LicenseID] = customer
		byWatermark[customer.Watermark] = customer
		byCustomer[customer.CustomerName] = customer
	}
	out := make([]ProductInstance, 0, len(instances))
	for _, instance := range instances {
		if customer, ok := matchLicenseCustomer(instance, byLicense, byWatermark, byCustomer); ok {
			instance.LicenseID = fallback(instance.LicenseID, customer.LicenseID)
			instance.Watermark = fallback(instance.Watermark, customer.Watermark)
			instance.Edition = fallback(instance.Edition, customer.Edition)
			instance.LicenseExpiresAt = fallback(instance.LicenseExpiresAt, customer.ExpiresAt)
			instance.DaysToExpire = customer.DaysToExpire
			instance.LicenseRisk = customer.RiskLevel
			instance.RenewalAvailable = customer.RenewalAvailable
			instance.LatestPackageID = customer.LatestPackageID
		} else {
			instance.DaysToExpire = licenseDaysToExpire(instance.LicenseExpiresAt)
			instance.LicenseRisk = productInstanceLicenseRisk(instance.DaysToExpire)
			instance.RenewalAvailable = instance.DaysToExpire >= 0
		}
		if instance.AlertLevel == "" || instance.AlertLevel == "normal" {
			instance.AlertLevel = instance.LicenseRisk
		}
		out = append(out, instance)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AlertLevel != out[j].AlertLevel {
			return productOpsRiskRank(out[i].AlertLevel) > productOpsRiskRank(out[j].AlertLevel)
		}
		return out[i].DaysToExpire < out[j].DaysToExpire
	})
	return out
}

func matchLicenseCustomer(instance ProductInstance, byLicense, byWatermark, byCustomer map[string]LicensePortalCustomer) (LicensePortalCustomer, bool) {
	if item, ok := byLicense[instance.LicenseID]; ok {
		return item, true
	}
	if item, ok := byWatermark[instance.Watermark]; ok {
		return item, true
	}
	if item, ok := byCustomer[instance.CustomerName]; ok {
		return item, true
	}
	return LicensePortalCustomer{}, false
}

func productInstanceLicenseRisk(days int) string {
	if days < 0 {
		return "expired"
	}
	if days <= 45 {
		return "warning"
	}
	return "normal"
}

func productOpsRiskRank(value string) int {
	switch value {
	case "critical", "expired":
		return 4
	case "warning", "revoked":
		return 3
	case "degraded":
		return 2
	default:
		return 1
	}
}
