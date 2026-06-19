package appliance

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SessionPolicy struct {
	TimeoutMinutes     int      `json:"timeoutMinutes"`
	MaxSessionsPerUser int      `json:"maxSessionsPerUser"`
	IPWhitelistEnabled bool     `json:"ipWhitelistEnabled"`
	AllowedIPRanges    []string `json:"allowedIpRanges"`
}

type ActiveSessionSummary struct {
	UserID      int64  `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	RoleCode    string `json:"roleCode"`
	IP          string `json:"ip"`
	UserAgent   string `json:"userAgent"`
	CreatedAt   string `json:"createdAt"`
	LastSeenAt  string `json:"lastSeenAt"`
	ExpiresAt   string `json:"expiresAt"`
	AgeMinutes  int    `json:"ageMinutes"`
}

type SecurityReport struct {
	TotalUsers              int      `json:"totalUsers"`
	ActiveUsers             int      `json:"activeUsers"`
	MFAUsers                int      `json:"mfaUsers"`
	MFACoverage             float64  `json:"mfaCoverage"`
	SSOProviders            int      `json:"ssoProviders"`
	SCIMProviders           int      `json:"scimProviders"`
	SCIMEventsLast24h       int      `json:"scimEventsLast24h"`
	EnabledSecurityPolicies int      `json:"enabledSecurityPolicies"`
	DeviceCredentials       int      `json:"deviceCredentials"`
	ActiveSessions          int      `json:"activeSessions"`
	StaleSessions           int      `json:"staleSessions"`
	ExpiringSessions        int      `json:"expiringSessions"`
	LoginLast24h            int      `json:"loginLast24h"`
	FailedLoginLast24h      int      `json:"failedLoginLast24h"`
	AuditEventsLast24h      int      `json:"auditEventsLast24h"`
	IPWhitelistEnabled      bool     `json:"ipWhitelistEnabled"`
	RiskLevel               string   `json:"riskLevel"`
	Recommendations         []string `json:"recommendations"`
}

func buildSessionPolicy(data AppData) SessionPolicy {
	policy := SessionPolicy{
		TimeoutMinutes:     securityPolicyInt(data, "session_timeout_minutes", 480),
		MaxSessionsPerUser: securityPolicyInt(data, "session_max_per_user", 5),
		IPWhitelistEnabled: ipWhitelistEnforced(data),
	}
	for _, item := range data.SecurityPolicies {
		if item.Enabled && item.Type == "ip_whitelist" && strings.TrimSpace(item.Value) != "" {
			policy.AllowedIPRanges = append(policy.AllowedIPRanges, strings.TrimSpace(item.Value))
		}
	}
	return policy
}

func securityPolicyInt(data AppData, policyType string, fallbackValue int) int {
	for _, item := range data.SecurityPolicies {
		if !item.Enabled || item.Type != policyType {
			continue
		}
		value, err := strconv.Atoi(strings.TrimSpace(item.Value))
		if err == nil && value > 0 {
			return value
		}
	}
	return fallbackValue
}

func ipWhitelistEnforced(data AppData) bool {
	if os.Getenv("CBMP_ENFORCE_IP_WHITELIST") != "1" {
		return false
	}
	for _, item := range data.SecurityPolicies {
		if item.Enabled && item.Type == "ip_whitelist" && strings.TrimSpace(item.Value) != "" {
			return true
		}
	}
	return false
}

func sessionExpiresAt(createdAt string, timeoutMinutes int) string {
	created, ok := parseSecurityTime(createdAt)
	if !ok {
		created = time.Now()
	}
	if timeoutMinutes <= 0 {
		timeoutMinutes = 480
	}
	return created.Add(time.Duration(timeoutMinutes) * time.Minute).Format("2006-01-02 15:04:05")
}

func parseSecurityTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local); err == nil {
		return parsed, true
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, true
	}
	return time.Time{}, false
}

func (a *App) pruneSessionsForUserLocked(userID int64, policy SessionPolicy) {
	now := time.Now()
	for token, session := range a.sessions {
		if expiresAt, ok := parseSecurityTime(session.ExpiresAt); ok && now.After(expiresAt) {
			delete(a.sessions, token)
		}
	}
	if policy.MaxSessionsPerUser <= 0 {
		return
	}
	type pair struct {
		token   string
		session Session
	}
	sessions := []pair{}
	for token, session := range a.sessions {
		if session.User.ID == userID {
			sessions = append(sessions, pair{token: token, session: session})
		}
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].session.CreatedAt < sessions[j].session.CreatedAt })
	for len(sessions) >= policy.MaxSessionsPerUser {
		delete(a.sessions, sessions[0].token)
		sessions = sessions[1:]
	}
}

func (a *App) activeSessionSummaries(policy SessionPolicy) []ActiveSessionSummary {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []ActiveSessionSummary{}
	for token, session := range a.sessions {
		if session.ExpiresAt == "" {
			session.ExpiresAt = sessionExpiresAt(session.CreatedAt, policy.TimeoutMinutes)
			a.sessions[token] = session
		}
		if expiresAt, ok := parseSecurityTime(session.ExpiresAt); ok && now.After(expiresAt) {
			delete(a.sessions, token)
			continue
		}
		createdAt, ok := parseSecurityTime(session.CreatedAt)
		ageMinutes := 0
		if ok {
			ageMinutes = int(now.Sub(createdAt).Minutes())
		}
		out = append(out, ActiveSessionSummary{
			UserID: session.User.ID, Username: session.User.Username, DisplayName: session.User.DisplayName,
			RoleCode: session.User.RoleCode, IP: session.IP, UserAgent: session.UserAgent,
			CreatedAt: session.CreatedAt, LastSeenAt: session.LastSeenAt, ExpiresAt: session.ExpiresAt,
			AgeMinutes: ageMinutes,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeenAt > out[j].LastSeenAt })
	return out
}

func buildSecurityReport(data AppData, sessions []ActiveSessionSummary, policy SessionPolicy) SecurityReport {
	report := SecurityReport{
		TotalUsers:         len(data.Users),
		ActiveSessions:     len(sessions),
		SSOProviders:       len(data.OIDCProviders),
		SCIMProviders:      len(data.SCIMProviders),
		DeviceCredentials:  len(data.DeviceCredentials),
		IPWhitelistEnabled: policy.IPWhitelistEnabled,
	}
	for _, user := range data.Users {
		if user.Status == "active" {
			report.ActiveUsers++
		}
		if user.MFAEnabled {
			report.MFAUsers++
		}
	}
	if report.ActiveUsers > 0 {
		report.MFACoverage = percent(report.MFAUsers, report.ActiveUsers)
	}
	for _, item := range data.SecurityPolicies {
		if item.Enabled {
			report.EnabledSecurityPolicies++
		}
	}
	now := time.Now()
	for _, session := range sessions {
		if lastSeenAt, ok := parseSecurityTime(session.LastSeenAt); ok && now.Sub(lastSeenAt) > 30*time.Minute {
			report.StaleSessions++
		}
		if expiresAt, ok := parseSecurityTime(session.ExpiresAt); ok && expiresAt.Sub(now) <= 15*time.Minute {
			report.ExpiringSessions++
		}
	}
	cutoff := now.Add(-24 * time.Hour)
	for _, audit := range data.AuditLogs {
		createdAt, ok := parseSecurityTime(audit.CreatedAt)
		if !ok || createdAt.Before(cutoff) {
			continue
		}
		report.AuditEventsLast24h++
		switch audit.Action {
		case "login":
			report.LoginLast24h++
		case "failed_login":
			report.FailedLoginLast24h++
		}
	}
	for _, event := range data.SCIMEvents {
		createdAt, ok := parseSecurityTime(event.CreatedAt)
		if ok && !createdAt.Before(cutoff) {
			report.SCIMEventsLast24h++
		}
	}
	report.RiskLevel = "low"
	if report.MFACoverage < 60 {
		report.RiskLevel = "medium"
		report.Recommendations = append(report.Recommendations, "MFA 覆盖率低于 60%，建议优先启用调度、财务和系统管理员 MFA")
	}
	if report.FailedLoginLast24h >= 5 {
		report.RiskLevel = "high"
		report.Recommendations = append(report.Recommendations, "近 24 小时失败登录较多，建议检查账号爆破和 IP 白名单")
	}
	if !report.IPWhitelistEnabled {
		report.Recommendations = append(report.Recommendations, "生产环境建议设置 CBMP_ENFORCE_IP_WHITELIST=1 启用运维 IP 白名单")
	}
	if enabledSCIMProviders(data.SCIMProviders) == 0 {
		report.Recommendations = append(report.Recommendations, "建议启用 SCIM 企业目录同步，减少离职账号滞留")
	}
	if len(report.Recommendations) == 0 {
		report.Recommendations = append(report.Recommendations, "当前安全策略处于低风险状态，继续保持 MFA、SSO 和审计巡检")
	}
	return report
}

func enabledSCIMProviders(items []SCIMProvider) int {
	total := 0
	for _, item := range items {
		if item.Status == "enabled" {
			total++
		}
	}
	return total
}
