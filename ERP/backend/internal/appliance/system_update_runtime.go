package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (a *App) systemUpdateRuntimeDownload(w http.ResponseWriter, r *http.Request, idText string) {
	token := updaterTokenFromRequest(r, "")
	if token == "" {
		unauthorized(w)
		return
	}
	updateID, _ := strconv.ParseInt(idText, 10, 64)
	var download UpdatePackageDownload
	err := a.store.Mutate(func(data *AppData) error {
		instanceIndex := productInstanceIndexByUpdaterToken(*data, token, "")
		if instanceIndex < 0 {
			return fmt.Errorf("updater token invalid")
		}
		instance := data.ProductInstances[instanceIndex]
		authorized := false
		for _, task := range data.ProductSystemUpdateTasks {
			if task.InstanceID == instance.ID && task.UpdateID == updateID && activeSystemUpdateTaskStatus(task.Status) {
				authorized = true
				break
			}
		}
		if !authorized {
			return fmt.Errorf("updater task not authorized for this update package")
		}
		for i := range data.Updates {
			if data.Updates[i].ID != updateID {
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
			addAudit(data, "updater:"+instance.Watermark, "download", "update_package", updateID, data.Updates[i].Component+" "+data.Updates[i].Version, clientIP(r))
			return nil
		}
		return fmt.Errorf("更新包不存在")
	})
	if err != nil {
		if strings.Contains(err.Error(), "token") {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, download)
}

func updaterTokenFromRequest(r *http.Request, bodyToken string) string {
	token := strings.TrimSpace(r.Header.Get("X-CBMP-Updater-Token"))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-CBMP-Probe-Token"))
	}
	if token == "" {
		token = strings.TrimSpace(bodyToken)
	}
	return token
}

func activeSystemUpdateTaskStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "assigned", "running":
		return true
	default:
		return false
	}
}

func productInstanceIndexByUpdaterToken(data AppData, token, watermark string) int {
	for i := range data.ProductInstances {
		instance := data.ProductInstances[i]
		if instance.ProbeEnabled && instance.ProbeToken == token {
			if watermark == "" || instance.Watermark == watermark || instance.LicenseID == watermark {
				return i
			}
		}
	}
	return -1
}
