//go:build legacy_product_ops

package appliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type productAlertGovernanceContext struct {
	Component string
	Metric    string
	EventNo   string
}

func normalizeProductAlertChannel(req ProductAlertChannel, actor string) (ProductAlertChannel, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.Token = strings.TrimSpace(req.Token)
	req.Secret = strings.TrimSpace(req.Secret)
	req.Status = fallback(strings.ToLower(strings.TrimSpace(req.Status)), "active")
	req.Remark = strings.TrimSpace(req.Remark)
	if err := validateNoMockEndpoint(req.Endpoint, "通知通道 endpoint"); err != nil {
		return req, err
	}
	if req.Name == "" {
		return req, fmt.Errorf("通知通道名称不能为空")
	}
	if req.Type == "" {
		req.Type = req.Code
	}
	switch req.Type {
	case "sse", "local", "webhook", "enterprise_wechat", "sms", "itsm":
	default:
		return req, fmt.Errorf("通知通道类型必须是 sse、local、webhook、enterprise_wechat、sms 或 itsm")
	}
	if req.Status == "active" && req.Type != "sse" && req.Type != "local" && req.Endpoint == "" {
		return req, fmt.Errorf("启用外部通知通道必须配置真实 endpoint")
	}
	if req.Code == "" {
		req.Code = req.Type
	}
	if req.RetryLimit <= 0 {
		req.RetryLimit = 3
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 3
	}
	if req.CreatedAt == "" {
		req.CreatedAt = nowString()
	}
	if req.CreatedBy == "" {
		req.CreatedBy = actor
	}
	return req, nil
}

func applyProductAlertGovernance(data *AppData, alert *SystemAlert, ctx productAlertGovernanceContext, created bool) {
	if alert == nil || alert.Status == "handled" || alert.Status == "closed" {
		return
	}
	now := fallback(alert.LastSeenAt, nowString())
	alert.Severity = fallback(normalizeTelemetrySeverity(alert.Severity), "warning")
	alert.Status = fallback(strings.TrimSpace(alert.Status), "open")
	alert.GroupKey = fallback(alert.GroupKey, productAlertGroupKey(*alert, ctx))
	if alert.EventCount <= 0 {
		alert.EventCount = 1
	} else if !created {
		alert.EventCount++
	}

	policy := matchProductAlertPolicy(*data, *alert, ctx)
	action := "created"
	if !created {
		action = "aggregated"
	}
	message := fmt.Sprintf("%s %s 已进入告警中心", alert.AlertNo, alert.Title)
	if !created {
		message = fmt.Sprintf("%s 已聚合第 %d 次重复事件", alert.AlertNo, alert.EventCount)
	}
	if policy != nil {
		alert.PolicyNo = policy.PolicyNo
		if policy.AggregateWindowMinutes > 0 && !created && alertWindowExpired(alert.FirstSeenAt, now, policy.AggregateWindowMinutes) {
			alert.FirstSeenAt = now
			alert.EventCount = 1
			alert.SuppressedUntil = ""
			action = "created"
			message = fmt.Sprintf("%s 超出聚合窗口后重新计入告警", alert.AlertNo)
		}
		if policy.SuppressMinutes > 0 && !created && alert.EventCount > 1 {
			alert.SuppressedUntil = addMinutesString(now, policy.SuppressMinutes)
			action = "suppressed"
			message = fmt.Sprintf("%s 重复事件已抑制至 %s", alert.AlertNo, alert.SuppressedUntil)
		}
		if policy.EscalateAfterMinutes > 0 && alert.EscalationLevel == "" && alertElapsedMinutes(alert.FirstSeenAt, now) >= policy.EscalateAfterMinutes {
			alert.EscalationLevel = fallback(policy.EscalateTo, "on_call")
			alert.EscalatedAt = now
			alert.Severity = "critical"
			action = "escalated"
			message = fmt.Sprintf("%s 已超过 %d 分钟未处理，升级给 %s", alert.AlertNo, policy.EscalateAfterMinutes, alert.EscalationLevel)
		}
	}
	appendProductAlertNotifications(data, *alert, policy, action, message)
}

func productAlertGroupKey(alert SystemAlert, ctx productAlertGovernanceContext) string {
	parts := []string{
		fmt.Sprintf("instance:%d", alert.InstanceID),
		"source:" + strings.ToLower(strings.TrimSpace(alert.Source)),
		"component:" + strings.ToLower(strings.TrimSpace(ctx.Component)),
		"metric:" + strings.ToLower(strings.TrimSpace(ctx.Metric)),
		"title:" + strings.ToLower(strings.TrimSpace(alert.Title)),
	}
	return strings.Join(parts, "|")
}

func matchProductAlertPolicy(data AppData, alert SystemAlert, ctx productAlertGovernanceContext) *ProductAlertPolicy {
	var selected *ProductAlertPolicy
	selectedScore := -1
	for i := range data.ProductAlertPolicies {
		policy := &data.ProductAlertPolicies[i]
		if policy.Status != "active" {
			continue
		}
		score := 0
		if policy.Source != "" && policy.Source != "all" {
			if policy.Source != alert.Source {
				continue
			}
			score += 4
		}
		if policy.Component != "" && policy.Component != "all" {
			if policy.Component != ctx.Component {
				continue
			}
			score += 3
		}
		if policy.Metric != "" && policy.Metric != "all" {
			if policy.Metric != ctx.Metric {
				continue
			}
			score += 2
		}
		if policy.Severity != "" && policy.Severity != "all" {
			if severityRank(alert.Severity) < severityRank(policy.Severity) {
				continue
			}
			score++
		}
		if score > selectedScore {
			selected = policy
			selectedScore = score
		}
	}
	return selected
}

func appendProductAlertNotifications(data *AppData, alert SystemAlert, policy *ProductAlertPolicy, action, message string) {
	channels := []string{"sse"}
	target := ""
	policyID := int64(0)
	policyNo := ""
	if policy != nil {
		policyID = policy.ID
		policyNo = policy.PolicyNo
		if len(policy.NotifyChannels) > 0 {
			channels = policy.NotifyChannels
		}
		target = policy.EscalateTo
	}
	if action == "suppressed" {
		channels = []string{"suppression"}
	}
	if action == "aggregated" && alert.SuppressedUntil != "" {
		channels = []string{"aggregation"}
	}
	now := fallback(alert.LastSeenAt, nowString())
	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		id := nextID(data, "alertNotification")
		notification := ProductAlertNotification{
			ID:             id,
			NotificationNo: number("AN", id),
			AlertID:        alert.ID,
			AlertNo:        alert.AlertNo,
			PolicyID:       policyID,
			PolicyNo:       policyNo,
			InstanceID:     alert.InstanceID,
			CustomerName:   alert.CustomerName,
			Action:         action,
			Severity:       alert.Severity,
			Channel:        channel,
			Target:         target,
			Status:         "pending",
			Message:        message,
			CreatedAt:      now,
		}
		applyProductAlertChannel(data, &notification)
		deliverProductAlertNotification(data, &notification)
		data.ProductAlertNotifications = append(data.ProductAlertNotifications, notification)
	}
	trimProductAlertNotifications(data)
}

func applyProductAlertChannel(data *AppData, notification *ProductAlertNotification) {
	channel := strings.TrimSpace(notification.Channel)
	for i := range data.ProductAlertChannels {
		item := data.ProductAlertChannels[i]
		if item.Status != "active" {
			continue
		}
		if item.Code == channel || item.Type == channel {
			notification.ChannelID = item.ID
			notification.ChannelNo = item.ChannelNo
			notification.Channel = fallback(item.Type, channel)
			notification.Target = fallback(notification.Target, item.Code)
			notification.Endpoint = item.Endpoint
			return
		}
	}
}

func deliverProductAlertNotification(data *AppData, notification *ProductAlertNotification) {
	now := nowString()
	notification.AttemptCount++
	channel := productAlertChannelForNotification(*data, *notification)
	channelType := strings.ToLower(strings.TrimSpace(notification.Channel))
	if channel != nil {
		channelType = fallback(strings.ToLower(strings.TrimSpace(channel.Type)), channelType)
		notification.Endpoint = fallback(notification.Endpoint, channel.Endpoint)
	}
	switch channelType {
	case "", "sse", "local", "suppression", "aggregation":
		notification.Status = "delivered"
		notification.DeliveredAt = now
		notification.Error = ""
	case "webhook", "enterprise_wechat", "sms", "itsm":
		if notification.Endpoint == "" {
			notification.Status = "failed"
			notification.Error = "通知通道未配置 endpoint"
			notification.NextRetryAt = addMinutesString(now, 5)
			break
		}
		if err := postProductAlertWebhook(notification, channel); err != nil {
			notification.Status = "failed"
			notification.Error = err.Error()
			if channel == nil || channel.RetryLimit <= 0 || notification.AttemptCount < channel.RetryLimit {
				notification.NextRetryAt = addMinutesString(now, 5*notification.AttemptCount)
			} else {
				notification.NextRetryAt = ""
			}
			break
		}
		notification.Status = "delivered"
		notification.DeliveredAt = now
		notification.Error = ""
		notification.NextRetryAt = ""
	default:
		notification.Status = "failed"
		notification.Error = "未知通知通道 " + notification.Channel
		notification.NextRetryAt = addMinutesString(now, 5)
	}
	if notification.ChannelID != 0 {
		for i := range data.ProductAlertChannels {
			if data.ProductAlertChannels[i].ID != notification.ChannelID {
				continue
			}
			if notification.Status == "delivered" {
				data.ProductAlertChannels[i].LastDeliveredAt = notification.DeliveredAt
				data.ProductAlertChannels[i].LastError = ""
			} else if notification.Error != "" {
				data.ProductAlertChannels[i].LastError = notification.Error
			}
			break
		}
	}
}

func productAlertChannelForNotification(data AppData, notification ProductAlertNotification) *ProductAlertChannel {
	for i := range data.ProductAlertChannels {
		item := &data.ProductAlertChannels[i]
		if item.Status != "active" {
			continue
		}
		if notification.ChannelID != 0 && item.ID == notification.ChannelID {
			return item
		}
		if notification.ChannelNo != "" && item.ChannelNo == notification.ChannelNo {
			return item
		}
		if notification.Channel != "" && (item.Code == notification.Channel || item.Type == notification.Channel) {
			return item
		}
	}
	return nil
}

func postProductAlertWebhook(notification *ProductAlertNotification, channel *ProductAlertChannel) error {
	endpoint := strings.TrimSpace(notification.Endpoint)
	if err := validateNoMockEndpoint(endpoint, "通知通道 endpoint"); err != nil {
		return err
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("不支持的通知 endpoint scheme: %s", parsed.Scheme)
	}
	payload := productAlertNotificationPayload(*notification, channel)
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	timeout := 3
	if channel != nil && channel.TimeoutSeconds > 0 {
		timeout = channel.TimeoutSeconds
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "cbmp-alert-notifier/1.0")
	req.Header.Set("X-CBMP-Notification-No", notification.NotificationNo)
	req.Header.Set("X-CBMP-Alert-No", notification.AlertNo)
	if channel != nil && channel.Token != "" {
		req.Header.Set("Authorization", "Bearer "+channel.Token)
		req.Header.Set("X-CBMP-Channel-Token", channel.Token)
	}
	if channel != nil && channel.Secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		req.Header.Set("X-CBMP-Timestamp", timestamp)
		req.Header.Set("X-CBMP-Signature", taxGatewaySignature(channel.Secret, timestamp, raw))
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		summary := strings.TrimSpace(string(body))
		if summary != "" {
			return fmt.Errorf("通知 endpoint 返回 %s: %s", resp.Status, summary)
		}
		return fmt.Errorf("通知 endpoint 返回 %s", resp.Status)
	}
	return nil
}

func productAlertNotificationPayload(notification ProductAlertNotification, channel *ProductAlertChannel) map[string]interface{} {
	channelType := strings.ToLower(strings.TrimSpace(notification.Channel))
	channelCode := channelType
	if channel != nil {
		channelType = fallback(strings.ToLower(strings.TrimSpace(channel.Type)), channelType)
		channelCode = fallback(channel.Code, channelType)
	}
	common := map[string]interface{}{
		"schema":         "cbmp.alert.v1",
		"notificationNo": notification.NotificationNo,
		"alertNo":        notification.AlertNo,
		"policyNo":       notification.PolicyNo,
		"instanceId":     notification.InstanceID,
		"customerName":   notification.CustomerName,
		"action":         notification.Action,
		"severity":       notification.Severity,
		"channel":        channelType,
		"channelCode":    channelCode,
		"target":         notification.Target,
		"message":        notification.Message,
		"createdAt":      notification.CreatedAt,
	}
	switch channelType {
	case "enterprise_wechat":
		return map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"content": fmt.Sprintf("**[%s] %s**\n> 客户：%s\n> 告警：%s\n> 动作：%s\n> 通知：%s",
					strings.ToUpper(notification.Severity), notification.Message, notification.CustomerName, notification.AlertNo, notification.Action, notification.NotificationNo),
			},
			"cbmp": common,
		}
	case "sms":
		return map[string]interface{}{
			"requestId":    notification.NotificationNo,
			"templateCode": "CBMP_ALERT",
			"phone":        notification.Target,
			"content":      notification.Message,
			"params": map[string]interface{}{
				"alertNo":      notification.AlertNo,
				"customerName": notification.CustomerName,
				"severity":     notification.Severity,
				"action":       notification.Action,
			},
			"cbmp": common,
		}
	case "itsm":
		return map[string]interface{}{
			"requestId":   notification.NotificationNo,
			"eventType":   "incident",
			"externalId":  notification.AlertNo,
			"priority":    productAlertITSMImpact(notification.Severity),
			"summary":     notification.Message,
			"description": notification.Message,
			"customer":    notification.CustomerName,
			"source":      "cbmp",
			"cbmp":        common,
		}
	default:
		return common
	}
}

func productAlertITSMImpact(severity string) string {
	switch normalizeTelemetrySeverity(severity) {
	case "critical":
		return "P1"
	case "warning":
		return "P2"
	default:
		return "P3"
	}
}

func retryProductAlertNotification(data *AppData, id int64) (ProductAlertNotification, error) {
	for i := range data.ProductAlertNotifications {
		if data.ProductAlertNotifications[i].ID != id {
			continue
		}
		data.ProductAlertNotifications[i].Status = "pending"
		data.ProductAlertNotifications[i].NextRetryAt = ""
		applyProductAlertChannel(data, &data.ProductAlertNotifications[i])
		deliverProductAlertNotification(data, &data.ProductAlertNotifications[i])
		return data.ProductAlertNotifications[i], nil
	}
	return ProductAlertNotification{}, fmt.Errorf("通知记录不存在")
}

func trimProductAlertNotifications(data *AppData) {
	const maxItems = 300
	if len(data.ProductAlertNotifications) <= maxItems {
		return
	}
	data.ProductAlertNotifications = append([]ProductAlertNotification{}, data.ProductAlertNotifications[len(data.ProductAlertNotifications)-maxItems:]...)
}

func alertWindowExpired(firstSeenAt, nowText string, minutes int) bool {
	if minutes <= 0 {
		return false
	}
	return alertElapsedMinutes(firstSeenAt, nowText) > minutes
}

func alertElapsedMinutes(firstSeenAt, nowText string) int {
	first, ok := parseLocalDateTime(firstSeenAt)
	if !ok {
		return 0
	}
	now, ok := parseLocalDateTime(nowText)
	if !ok {
		now = time.Now()
	}
	if now.Before(first) {
		return 0
	}
	return int(now.Sub(first).Minutes())
}

func addMinutesString(value string, minutes int) string {
	base, ok := parseLocalDateTime(value)
	if !ok {
		base = time.Now()
	}
	return base.Add(time.Duration(minutes) * time.Minute).Format("2006-01-02 15:04:05")
}

func parseLocalDateTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local); err == nil {
		return parsed, true
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local(), true
	}
	return time.Time{}, false
}

func severityRank(value string) int {
	switch normalizeTelemetrySeverity(value) {
	case "critical":
		return 3
	case "warning":
		return 2
	case "normal":
		return 1
	default:
		return 0
	}
}
