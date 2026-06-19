package appliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type pluginRunRequest struct {
	Action     string                 `json:"action"`
	Permission string                 `json:"permission"`
	Input      map[string]interface{} `json:"input"`
}

func (a *App) systemPluginRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, a.mustSnapshot().PluginRuns)
}

func (a *App) runPlugin(w http.ResponseWriter, r *http.Request, session Session, pluginID string) {
	var req pluginRunRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid plugin run payload")
		return
	}
	var run PluginRun
	err := a.store.Mutate(func(data *AppData) error {
		idx := -1
		for i := range data.Plugins {
			if data.Plugins[i].ID == pluginID {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("插件不存在")
		}
		plugin := data.Plugins[idx]
		if plugin.Status != "enabled" {
			return fmt.Errorf("插件未启用")
		}
		if !checksumVerified(plugin.Checksum, plugin.Signature) {
			return fmt.Errorf("插件验签失败")
		}
		permission := fallback(req.Permission, firstPermission(plugin.Permissions))
		if !pluginHasPermission(plugin, permission) {
			return fmt.Errorf("插件缺少权限: %s", permission)
		}
		start := time.Now()
		inputJSON, _ := json.Marshal(req.Input)
		output, runErr := executeSandboxedPlugin(plugin, fallback(req.Action, permission), req.Input)
		outputJSON, _ := json.Marshal(output)
		status := "succeeded"
		errText := ""
		if runErr != nil {
			status = "failed"
			errText = runErr.Error()
		}
		run = PluginRun{
			ID: nextID(data, "pluginRun"), RunNo: number("PLG", data.Next["pluginRun"]),
			PluginID: plugin.ID, PluginName: plugin.Name, Runtime: fallback(plugin.Runtime, plugin.Sandbox.Runtime),
			Action: fallback(req.Action, permission), Permission: permission, Status: status,
			Input: string(inputJSON), Output: string(outputJSON), Error: errText, Actor: session.User.Username,
			StartedAt: start.Format("2006-01-02 15:04:05"), CompletedAt: nowString(), DurationMs: time.Since(start).Milliseconds(),
		}
		data.PluginRuns = append(data.PluginRuns, run)
		if len(data.PluginRuns) > 200 {
			data.PluginRuns = data.PluginRuns[len(data.PluginRuns)-200:]
		}
		data.Plugins[idx].LastRunAt = run.CompletedAt
		addAudit(data, session.User.Username, "run", "plugin", run.ID, plugin.ID+"/"+run.Action+"/"+status, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, run, "system.plugin.run")
}

func executeSandboxedPlugin(plugin Plugin, action string, input map[string]interface{}) (map[string]interface{}, error) {
	runtime := fallback(plugin.Runtime, plugin.Sandbox.Runtime)
	switch plugin.ID {
	case "settlement-tax-cn":
		amount := floatFromInput(input, "amount")
		rate := floatFromInput(input, "taxRate")
		if rate == 0 {
			rate = 0.13
		}
		tax := round(amount * rate)
		return map[string]interface{}{
			"runtime": runtime, "sandbox": plugin.Sandbox, "action": action,
			"amount": round(amount), "taxRate": rate, "taxAmount": tax, "totalWithTax": round(amount + tax),
		}, nil
	case "adapter-scale-standard":
		gross := floatFromInput(input, "grossWeight")
		tare := floatFromInput(input, "tareWeight")
		if gross <= 0 || tare < 0 || gross < tare {
			return nil, fmt.Errorf("地磅重量数据无效")
		}
		return map[string]interface{}{
			"runtime": runtime, "sandbox": plugin.Sandbox, "action": action,
			"netWeight": round(gross - tare), "stable": boolFromInput(input, "stable", true),
		}, nil
	case "adapter-gps-mqtt":
		return map[string]interface{}{
			"runtime": runtime, "sandbox": plugin.Sandbox, "action": action,
			"plateNo": stringFromInput(input, "plateNo"), "longitude": floatFromInput(input, "longitude"),
			"latitude": floatFromInput(input, "latitude"), "speed": floatFromInput(input, "speed"),
			"normalized": true,
		}, nil
	default:
		return nil, fmt.Errorf("插件运行时未注册")
	}
}

func pluginHasPermission(plugin Plugin, permission string) bool {
	for _, granted := range plugin.Permissions {
		if granted == permission || granted == "*" {
			return true
		}
		if strings.HasSuffix(granted, ".*") && strings.HasPrefix(permission, strings.TrimSuffix(granted, "*")) {
			return true
		}
	}
	return false
}

func firstPermission(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}

func floatFromInput(input map[string]interface{}, key string) float64 {
	value, ok := input[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case json.Number:
		v, _ := typed.Float64()
		return v
	default:
		return 0
	}
}

func boolFromInput(input map[string]interface{}, key string, fallback bool) bool {
	value, ok := input[key]
	if !ok {
		return fallback
	}
	typed, ok := value.(bool)
	if !ok {
		return fallback
	}
	return typed
}

func stringFromInput(input map[string]interface{}, key string) string {
	value, ok := input[key]
	if !ok {
		return ""
	}
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return typed
}
