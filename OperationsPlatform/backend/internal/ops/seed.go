package ops

func SeedData() AppData {
	return AppData{
		SchemaVersion: 1,
		Next: map[string]int64{
			"customer":   3,
			"renewal":    1,
			"alert":      4,
			"package":    2,
			"assignment": 2,
			"audit":      3,
		},
		Customers: []CustomerDeployment{
			{
				ID: 1, CustomerName: "湾区建材集团", ProductName: "CommonBuildMaterialsPlatform", LicenseID: "CBMP-WQ-2026-001",
				UpdaterToken: "ops-updater-demo-1",
				Edition:      "Enterprise Appliance", DeploymentMode: "private_server", Environment: "production", ServerEndpoint: "https://erp.wanqu.example.com",
				ContactName: "陈经理", ContactPhone: "13800010001", ExpiresAt: "2026-09-30", RenewalStatus: "expiring",
				Modules: []string{"erp", "production", "dispatch", "weighbridge", "finance", "update"}, MaxSites: 20, MaxVehicles: 5000,
				CurrentClientVersion: "1.4.1", CurrentServerVersion: "1.4.2", TargetClientVersion: "1.4.3", TargetServerVersion: "1.4.3",
				HealthStatus: "degraded", LastHeartbeatAt: "2026-06-19 09:18:00", Notes: "客户服务器私有化部署，需提前 45 天跟进续费。",
			},
			{
				ID: 2, CustomerName: "华东骨料供应链", ProductName: "CommonBuildMaterialsPlatform", LicenseID: "CBMP-HD-2026-006",
				UpdaterToken: "ops-updater-demo-2",
				Edition:      "Standard Appliance", DeploymentMode: "private_server", Environment: "production", ServerEndpoint: "https://materials.east.example.com",
				ContactName: "李总", ContactPhone: "13900020002", ExpiresAt: "2027-03-31", RenewalStatus: "active",
				Modules: []string{"erp", "inventory", "dispatch", "weighbridge"}, MaxSites: 5, MaxVehicles: 600,
				CurrentClientVersion: "1.4.3", CurrentServerVersion: "1.4.3", TargetClientVersion: "1.4.3", TargetServerVersion: "1.4.3",
				HealthStatus: "healthy", LastHeartbeatAt: "2026-06-19 09:20:10", Notes: "标准私有化部署，自动更新通道已打开。",
			},
			{
				ID: 3, CustomerName: "西南沥青联合", ProductName: "CommonBuildMaterialsPlatform", LicenseID: "CBMP-XN-2025-018",
				UpdaterToken: "ops-updater-demo-3",
				Edition:      "Enterprise Appliance", DeploymentMode: "private_server", Environment: "production", ServerEndpoint: "https://cbmp.southwest.example.com",
				ContactName: "周工", ContactPhone: "13700030003", ExpiresAt: "2026-06-10", RenewalStatus: "expired",
				Modules: []string{"erp", "production", "quality", "dispatch"}, MaxSites: 12, MaxVehicles: 1800,
				CurrentClientVersion: "1.3.9", CurrentServerVersion: "1.4.0", TargetClientVersion: "1.4.3", TargetServerVersion: "1.4.3",
				HealthStatus: "critical", LastHeartbeatAt: "2026-06-19 08:02:11", Notes: "授权已过期，需要运营介入处理。",
			},
		},
		Alerts: []SystemAlert{
			{
				ID: 1, AlertNo: "AL202606190001", CustomerID: 1, Source: "server", Severity: "critical",
				Title: "服务端 API 错误率升高", Message: "过去 10 分钟 /api/orders 5xx 错误率超过 8%。",
				Status: "open", FirstSeenAt: "2026-06-19 08:55:00", LastSeenAt: "2026-06-19 09:15:00", Assignee: "交付运维",
			},
			{
				ID: 2, AlertNo: "AL202606190002", CustomerID: 3, Source: "license", Severity: "critical",
				Title: "授权已过期", Message: "客户授权到期日早于当前日期，业务系统进入续费风险状态。",
				Status: "open", FirstSeenAt: "2026-06-19 08:10:00", LastSeenAt: "2026-06-19 08:10:00", Assignee: "客户成功",
			},
			{
				ID: 3, AlertNo: "AL202606190003", CustomerID: 1, Source: "client", Severity: "warning",
				Title: "客户端版本落后", Message: "仍有桌面客户端停留在 1.4.1，低于目标版本 1.4.3。",
				Status: "acknowledged", FirstSeenAt: "2026-06-18 16:00:00", LastSeenAt: "2026-06-19 09:00:00", AcknowledgedAt: "2026-06-19 09:05:00", Assignee: "产品运营",
			},
			{
				ID: 4, AlertNo: "AL202606190004", CustomerID: 2, Source: "backup", Severity: "info",
				Title: "备份演练完成", Message: "客户侧备份恢复演练通过，核心对象计数一致。",
				Status: "resolved", FirstSeenAt: "2026-06-18 22:30:00", LastSeenAt: "2026-06-18 22:35:00", ResolvedAt: "2026-06-18 22:40:00", Assignee: "交付运维", Resolution: "自动演练通过",
			},
		},
		UpdatePackages: []UpdatePackage{
			{
				ID: 1, PackageNo: "UP202606190001", Target: "client", ProductName: "CommonBuildMaterialsPlatform",
				Version: "1.4.3", Channel: "stable", Status: "published", FileName: "cbmp-desktop-1.4.3.dmg",
				Checksum: "sha256:client-143-release", MinVersion: "1.3.0", RolloutPct: 100,
				ReleaseNotes: "修复调度监控刷新、优化签收附件列表。", UploadedAt: "2026-06-19 07:30:00", PublishedAt: "2026-06-19 08:00:00",
			},
			{
				ID: 2, PackageNo: "UP202606190002", Target: "server", ProductName: "CommonBuildMaterialsPlatform",
				Version: "1.4.3", Channel: "stable", Status: "staged", FileName: "cbmp-appliance-1.4.3-linux-amd64.tar.gz",
				Checksum: "sha256:server-143-release", MinVersion: "1.3.0", RolloutPct: 25,
				ReleaseNotes: "补齐授权验签日志、更新设备协议兼容层。", UploadedAt: "2026-06-19 07:40:00",
			},
		},
		Assignments: []UpdateAssignment{
			{ID: 1, PackageID: 1, CustomerID: 1, Status: "assigned", AssignedAt: "2026-06-19 08:05:00"},
			{ID: 2, PackageID: 1, CustomerID: 2, Status: "applied", AssignedAt: "2026-06-19 08:05:00", DownloadedAt: "2026-06-19 08:20:00", AppliedAt: "2026-06-19 08:40:00"},
		},
		AuditLogs: []AuditLog{
			{ID: 1, Actor: "system", Action: "alert.created", Target: "AL202606190001", Detail: "服务端错误率超过阈值", CreatedAt: "2026-06-19 08:55:00"},
			{ID: 2, Actor: "system", Action: "package.published", Target: "UP202606190001", Detail: "客户端 1.4.3 发布到 stable", CreatedAt: "2026-06-19 08:00:00"},
			{ID: 3, Actor: "ops", Action: "alert.acknowledged", Target: "AL202606190003", Detail: "产品运营已确认客户端版本落后", CreatedAt: "2026-06-19 09:05:00"},
		},
	}
}
