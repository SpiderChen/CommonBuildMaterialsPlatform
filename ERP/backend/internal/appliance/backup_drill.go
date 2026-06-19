package appliance

import (
	"fmt"
	"net/http"
	"time"
)

func (a *App) backupDrills(w http.ResponseWriter, r *http.Request, session Session) {
	switch r.Method {
	case http.MethodGet:
		drills := a.mustSnapshot().BackupDrills
		if drills == nil {
			drills = []BackupDrill{}
		}
		writeJSON(w, http.StatusOK, drills)
	case http.MethodPost:
		a.runBackupDrill(w, r, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) runBackupDrill(w http.ResponseWriter, r *http.Request, session Session) {
	start := time.Now()
	snapshot := a.mustSnapshot()
	beforeCounts := backupObjectCounts(snapshot)
	info, err := a.backups.Create(snapshot)
	checks := []string{}
	status := "passed"
	errText := ""
	var restored AppData
	if err != nil {
		status = "failed"
		errText = err.Error()
	} else {
		checks = append(checks, "encrypted_backup_created")
		restored, err = a.backups.Restore(info.Name)
		if err != nil {
			status = "failed"
			errText = err.Error()
		} else {
			checks = append(checks, "decrypt_restore_verified", "schema_defaults_verified")
			if !sameBackupCounts(beforeCounts, backupObjectCounts(restored)) {
				status = "failed"
				errText = "恢复对象数量校验失败"
			} else {
				checks = append(checks, "core_counts_match")
			}
		}
	}
	var drill BackupDrill
	mutationErr := a.store.Mutate(func(data *AppData) error {
		id := nextID(data, "backupDrill")
		drill = BackupDrill{
			ID: id, DrillNo: number("DR", id), BackupName: info.Name, Status: status,
			StartedAt: start.Format("2006-01-02 15:04:05"), CompletedAt: nowString(), DurationMs: time.Since(start).Milliseconds(),
			SnapshotSize: info.Size, SchemaVersion: snapshot.SchemaVersion, ObjectCounts: beforeCounts,
			Checks: checks, Error: errText, Actor: session.User.Username,
		}
		data.BackupDrills = append(data.BackupDrills, drill)
		if len(data.BackupDrills) > 100 {
			data.BackupDrills = data.BackupDrills[len(data.BackupDrills)-100:]
		}
		addAudit(data, session.User.Username, "drill", "backup", drill.ID, drill.DrillNo+"/"+drill.Status, clientIP(r))
		return nil
	})
	if mutationErr != nil {
		writeError(w, http.StatusInternalServerError, mutationErr.Error())
		return
	}
	if status == "failed" {
		a.respondMutation(w, nil, drill, "system.backup.drill.failed")
		return
	}
	a.respondMutation(w, nil, drill, "system.backup.drill.passed")
}

func backupObjectCounts(data AppData) map[string]int {
	return map[string]int{
		"users":                  len(data.Users),
		"roles":                  len(data.Roles),
		"customers":              len(data.Customers),
		"orders":                 len(data.Orders),
		"scaleTickets":           len(data.ScaleTickets),
		"inventory":              len(data.Inventory),
		"pluginRuns":             len(data.PluginRuns),
		"oidcProviders":          len(data.OIDCProviders),
		"scimProviders":          len(data.SCIMProviders),
		"scimEvents":             len(data.SCIMEvents),
		"productInstances":       len(data.ProductInstances),
		"systemAlerts":           len(data.SystemAlerts),
		"renewalTasks":           len(data.ProductRenewalTasks),
		"renewalQuotes":          len(data.ProductRenewalQuotes),
		"renewalContracts":       len(data.ProductRenewalContracts),
		"renewalPayments":        len(data.ProductRenewalPayments),
		"renewalApprovals":       len(data.ProductRenewalApprovals),
		"renewalInvoices":        len(data.ProductRenewalInvoices),
		"renewalESigns":          len(data.ProductRenewalESigns),
		"renewalIntegrations":    len(data.ProductRenewalIntegrations),
		"renewalSyncRecords":     len(data.ProductRenewalSyncRecords),
		"probeReports":           len(data.ProductProbeReports),
		"telemetryEvents":        len(data.ProductTelemetryEvents),
		"monitoringIntegrations": len(data.ProductMonitoringIntegrations),
		"productAlertRules":      len(data.ProductAlertRules),
		"monitoringEvents":       len(data.ProductMonitoringEvents),
		"alertPolicies":          len(data.ProductAlertPolicies),
		"alertChannels":          len(data.ProductAlertChannels),
		"alertNotifications":     len(data.ProductAlertNotifications),
		"updateRollouts":         len(data.ProductUpdateRollouts),
		"updateExecutions":       len(data.ProductUpdateExecutions),
		"systemUpdateTasks":      len(data.ProductSystemUpdateTasks),
	}
}

func sameBackupCounts(left, right map[string]int) bool {
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	for key := range right {
		if _, ok := left[key]; !ok {
			return false
		}
	}
	return true
}

func requirePassedBackupDrill(drill BackupDrill) error {
	if drill.Status != "passed" {
		return fmt.Errorf("backup drill failed: %s", drill.Error)
	}
	if drill.BackupName == "" || drill.SnapshotSize == 0 || len(drill.Checks) == 0 {
		return fmt.Errorf("backup drill evidence incomplete")
	}
	return nil
}
