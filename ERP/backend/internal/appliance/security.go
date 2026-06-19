package appliance

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

func requiredPermission(parts []string, method string) string {
	if len(parts) == 0 {
		return ""
	}
	write := method != http.MethodGet
	switch parts[0] {
	case "me":
		return ""
	case "bootstrap":
		return "bootstrap:read"
	case "dashboard":
		return "dashboard:read"
	case "product-ops":
		if write {
			return "system:write"
		}
		return "system:read"
	case "master":
		if write {
			return "master:write"
		}
		return "master:read"
	case "contracts":
		if write {
			return "contract:write"
		}
		return "contract:read"
	case "orders":
		if write {
			return "order:write"
		}
		return "order:read"
	case "production-plans":
		if len(parts) > 1 && parts[1] == "protocols" && write {
			return "plant:report"
		}
		if write {
			return "production:write"
		}
		return "production:read"
	case "quality":
		if write {
			return "quality:write"
		}
		return "quality:read"
	case "laboratory":
		if write {
			return "quality:write"
		}
		return "quality:read"
	case "dispatch-center":
		if write {
			return "dispatch:write"
		}
		return "dispatch:read"
	case "dispatch-orders":
		if write {
			return "dispatch:write"
		}
		return "dispatch:read"
	case "weighbridge":
		if len(parts) > 1 && (parts[1] == "device-events" || parts[1] == "protocols") && write {
			return "scale:report"
		}
		if write {
			return "ticket:write"
		}
		return "ticket:read"
	case "delivery":
		if write {
			return "sign:create"
		}
		return "delivery:read"
	case "statements":
		if write {
			return "statement:confirm"
		}
		return "statement:read"
	case "procurement":
		if write {
			return "procurement:write"
		}
		return "procurement:read"
	case "finance":
		if write {
			return "finance:write"
		}
		return "finance:read"
	case "portal":
		if !write && len(parts) > 1 && parts[1] == "overview" {
			return "bootstrap:read"
		}
		if write && len(parts) > 3 && parts[1] == "dispatches" && parts[3] == "exception" {
			return "sign:create"
		}
		if write {
			return "customer:write"
		}
		return "customer:read"
	case "vehicle":
		if write {
			return "vehicle:write"
		}
		return "vehicle:read"
	case "iot":
		return "location:report"
	case "rules":
		if write {
			return "rule:write"
		}
		return "rule:read"
	case "integrations":
		if write {
			return "integration:write"
		}
		return "integration:read"
	case "approvals":
		if write {
			return "approval:write"
		}
		return "approval:read"
	case "reports":
		return "report:read"
	case "system":
		if write {
			return "system:write"
		}
		return "system:read"
	case "simulate":
		return "system:write"
	default:
		return ""
	}
}

func canAccess(data AppData, user User, permission string) bool {
	for _, role := range data.Roles {
		if role.Code != user.RoleCode {
			continue
		}
		return permissionGranted(role.Permissions, permission)
	}
	return false
}

func permissionGranted(grants []string, permission string) bool {
	for _, granted := range grants {
		if granted == "*" || granted == permission {
			return true
		}
		if strings.HasSuffix(granted, ":*") && strings.HasPrefix(permission, strings.TrimSuffix(granted, "*")) {
			return true
		}
	}
	return false
}

func (a *App) deviceSessionFromRequest(r *http.Request) (Session, bool) {
	key := strings.TrimSpace(r.Header.Get("X-Device-Key"))
	return a.deviceSessionFromKey(key, clientIP(r))
}

func (a *App) deviceSessionFromKey(key string, ip string) (Session, bool) {
	if key == "" {
		return Session{}, false
	}
	keyHash := sha256Hex(key)
	var device DeviceCredential
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.DeviceCredentials {
			if data.DeviceCredentials[i].Status != "active" || data.DeviceCredentials[i].KeyHash != keyHash {
				continue
			}
			data.DeviceCredentials[i].LastUsedAt = nowString()
			device = data.DeviceCredentials[i]
			addAudit(data, "device:"+device.DeviceNo, "auth", "device_credential", device.ID, "设备密钥认证", ip)
			return nil
		}
		return http.ErrNoCookie
	})
	if err != nil {
		return Session{}, false
	}
	return Session{
		Token:        "device:" + device.DeviceNo,
		DeviceScopes: append([]string(nil), device.Scopes...),
		User: User{
			Username:    "device:" + device.DeviceNo,
			DisplayName: "设备 " + device.DeviceNo,
			RoleCode:    "device",
			Status:      "active",
		},
		CreatedAt: nowString(),
	}, true
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func publicDeviceCredentials(items []DeviceCredential) []DeviceCredential {
	out := append([]DeviceCredential(nil), items...)
	for i := range out {
		out[i].KeyHash = ""
	}
	return out
}
