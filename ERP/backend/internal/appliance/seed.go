package appliance

import "fmt"

func SeedData() AppData {
	adminSalt, adminHash := makePassword("admin123")
	dispatchSalt, dispatchHash := makePassword("dispatch123")
	driverSalt, driverHash := makePassword("driver123")
	customerSalt, customerHash := makePassword("customer123")
	qualitySalt, qualityHash := makePassword("quality123")

	data := AppData{
		SchemaVersion: 2,
		Next: map[string]int64{
			"user": 5, "company": 1, "site": 2, "customer": 2, "project": 2,
			"productInstance": 2, "systemAlert": 2, "renewalTask": 2, "renewalQuote": 1, "renewalContract": 1, "renewalPayment": 1, "renewalApproval": 1, "renewalInvoice": 1, "renewalESign": 1, "renewalIntegration": 3, "renewalSyncRecord": 1, "probeReport": 2, "telemetryEvent": 2, "monitoringIntegration": 1, "productAlertRule": 2, "monitoringEvent": 1, "alertPolicy": 2, "alertChannel": 5, "alertNotification": 1, "updateRollout": 1, "updateRolloutItem": 2, "updateExecution": 1, "updateExecutionStep": 7, "systemUpdateTask": 1, "systemUpdateTaskLog": 2,
			"product": 6, "material": 5, "supplier": 2, "vehicle": 4, "driver": 4,
			"contract": 1, "contractAttachment": 1, "order": 2, "orderLine": 2, "productionPlan": 1, "mixDesign": 2, "mixTrial": 1,
			"productionTask": 1, "productionBatch": 1, "productionReport": 1,
			"qualityInspection": 0, "qualitySample": 0, "rawInspection": 0, "labSample": 1, "labTest": 1, "labEquipment": 2, "labCalibration": 1, "qualityException": 1,
			"inventory": 6, "dispatch": 2, "dispatchSchedule": 2, "ticket": 2, "sign": 1, "statement": 1,
			"location": 4, "fence": 4, "alarm": 1, "update": 3, "audit": 1,
			"department": 3, "plant": 2, "warehouse": 2, "silo": 6, "contact": 2, "customerBlacklist": 0, "customerProfile": 2, "customerComplaint": 1,
			"pricePolicy": 5, "taxRate": 2, "carrier": 2, "device": 5,
			"purchaseRequest": 1, "purchaseOrder": 1, "receipt": 1, "inventoryFlow": 3, "inventoryTrace": 2,
			"inventoryTransfer": 0, "stocktake": 0,
			"scaleDevice": 2, "weightRecord": 4, "scaleDeviceEvent": 0, "deliveryNote": 1, "signLink": 1, "signAttachment": 1, "printLog": 1, "voidLog": 0,
			"invoice": 1, "redLetter": 0, "taxSubmission": 1, "receivable": 1, "moneyReceipt": 1, "paymentPlan": 1, "collectionTask": 0, "collectionTemplate": 3, "collectionDispatch": 0, "supplierStatement": 1,
			"payable": 1, "payment": 1, "transportSettlement": 1, "transportSettlementItem": 1, "costCalc": 1,
			"projectProfit": 1, "rule": 4, "notification": 2, "integration": 7,
			"approvalFlow": 5, "approvalTask": 0, "dict": 14, "deviceCredential": 6, "securityPolicy": 6, "fieldPolicy": 8,
			"licensePackage": 0, "licenseIssue": 0, "licenseRevocation": 0, "oidcProvider": 1, "scimProvider": 1, "scimEvent": 0, "backupDrill": 0,
			"gatewayRoute": 2, "gatewayEvent": 0,
			"pluginRun":     0,
			"protocolFrame": 0,
		},
		License: LicenseInfo{
			LicenseID:    "local-demo",
			CustomerName: "湾区建材集团",
			Watermark:    "CBMP-DEMO-WANQU-2026",
			ExpiresAt:    "2027-12-31",
			Edition:      "ERP Appliance",
			Modules:      []string{"erp", "production", "quality", "dispatch", "gps", "weighbridge", "settlement", "report", "plugin", "update"},
			MaxSites:     20,
			MaxVehicles:  5000,
			IssuedAt:     "2026-06-18",
			Issuer:       "local-demo",
			Signature:    "local-demo-signature",
		},
		Modules: []Module{
			{"dashboard", "首页驾驶舱", "运营中心", "经营指标、订单发货、车辆和预警总览", true, false, "1.0.0"},
			{"master-data", "基础资料中心", "运营中心", "公司、站点、客户、产品、车辆、司机、物料", true, true, "1.0.0"},
			{"contract", "客户合同", "销售到收款", "合同价格、信用额度、账期和项目控制", true, true, "1.0.0"},
			{"sales-order", "销售订单", "销售到收款", "多产品线统一下单与审批", true, true, "1.0.0"},
			{"production", "生产计划", "生产到发货", "排产、任务单、生产批次、日报、库存扣减", true, true, "1.0.0"},
			{"quality", "实验室管理", "生产到发货", "配比版本、试配、原料检验、试块、仪器校准和异常闭环", true, true, "1.0.0"},
			{"inventory", "采购库存", "采购到付款", "库存台账、料仓、批次与预警", true, true, "1.0.0"},
			{"weighbridge", "地磅票据", "生产到发货", "过磅、票据、抓拍、作废和结算锁定", true, true, "1.0.0"},
			{"dispatch", "调度运输", "调度到签收", "派车、排队、装料、在途、到场、卸料", true, true, "1.0.0"},
			{"gps", "车辆监控", "车辆监控中心", "实时定位、轨迹、围栏、ETA 和异常", true, true, "1.0.0"},
			{"sign", "工地签收", "调度到签收", "扫码签收、定位校验、照片和电子签名", true, true, "1.0.0"},
			{"settlement", "客户对账", "销售到收款", "签收自动汇总到对账和应收", true, true, "1.0.0"},
			{"report", "报表分析", "经营分析", "项目毛利、客户欠款、车辆效率、库存预警", true, true, "1.0.0"},
			{"license", "授权中心", "交付底座", "授权、模块开通、客户水印", true, false, "1.0.0"},
			{"plugin", "插件管理", "交付底座", "安装、启停、卸载、权限隔离", true, false, "1.0.0"},
			{"update", "更新中心", "交付底座", "更新包验签、回滚、灰度", true, false, "1.0.0"},
		},
		Plugins: []Plugin{
			{ID: "adapter-gps-mqtt", Name: "GPS MQTT 设备适配器", Type: "device_adapter", Status: "enabled", Version: "1.0.0", Checksum: "sha256:gps-mqtt-100", Signature: "sig:gps-mqtt-100", Permissions: []string{"vehicle.location.report"}, Runtime: "rpc", Entrypoint: "rpc://gps-mqtt/normalize", Sandbox: PluginSandboxPolicy{Runtime: "rpc", TimeoutMs: 1200, Network: false, Filesystem: "none", MaxMemoryMB: 64}},
			{ID: "adapter-scale-standard", Name: "标准地磅串口适配器", Type: "device_adapter", Status: "enabled", Version: "1.0.0", Checksum: "sha256:scale-standard-100", Signature: "sig:scale-standard-100", Permissions: []string{"scale.ticket.create"}, Runtime: "rpc", Entrypoint: "rpc://scale-standard/parse", Sandbox: PluginSandboxPolicy{Runtime: "rpc", TimeoutMs: 1200, Network: false, Filesystem: "none", MaxMemoryMB: 64}},
			{ID: "settlement-tax-cn", Name: "中国税率结算插件", Type: "business_extension", Status: "enabled", Version: "1.0.0", Checksum: "sha256:settlement-tax-cn-100", Signature: "sig:settlement-tax-cn-100", Permissions: []string{"statement.calculate"}, Runtime: "wasm", Entrypoint: "wasm://settlement-tax-cn/calculate", Sandbox: PluginSandboxPolicy{Runtime: "wasm", TimeoutMs: 800, Network: false, Filesystem: "none", MaxMemoryMB: 32}},
		},
		GatewayRoutes: []GatewayRoute{
			{ID: 1, Name: "核心 API", PathPrefix: "/api/", StableUpstream: "cbmp-appliance:8088", CanaryUpstream: "cbmp-appliance-canary:8088", CanaryPercent: 0, ReadTimeoutSec: 120, Status: "active", UpdatedAt: nowString()},
			{ID: 2, Name: "实时事件流", PathPrefix: "/api/events", StableUpstream: "cbmp-appliance:8088", CanaryUpstream: "cbmp-appliance-canary:8088", CanaryPercent: 0, ReadTimeoutSec: 3600, Status: "active", UpdatedAt: nowString()},
		},
		Roles: []Role{
			{1, "boss", "运营管理员", []string{"*"}, "platform"},
			{2, "dispatcher", "调度员", []string{"bootstrap:read", "dashboard:read", "master:read", "order:read", "dispatch:*", "vehicle:*", "ticket:*", "delivery:*", "approval:*", "rule:read"}, "site"},
			{3, "driver", "司机", []string{"bootstrap:read", "dashboard:read", "driver:*", "location:report", "sign:create"}, "driver"},
			{4, "customer", "客户用户", []string{"bootstrap:read", "dashboard:read", "customer:*", "order:read", "approval:read", "delivery:read", "sign:create", "statement:read", "statement:confirm", "finance:read"}, "customer"},
			{5, "device", "设备接入", []string{"location:report", "scale:report", "plant:report"}, "device"},
			{6, "quality", "实验室质检员", []string{"bootstrap:read", "dashboard:read", "production:read", "quality:*", "procurement:read"}, "site"},
		},
		Users: []User{
			{ID: 1, CompanyID: 1, Username: "admin", DisplayName: "平台管理员", RoleCode: "boss", PasswordHash: adminHash, PasswordSalt: adminSalt, Status: "active"},
			{ID: 2, CompanyID: 1, SiteID: 1, Username: "dispatcher", DisplayName: "南山站调度", RoleCode: "dispatcher", PasswordHash: dispatchHash, PasswordSalt: dispatchSalt, Status: "active"},
			{ID: 3, CompanyID: 1, SiteID: 1, DriverID: 1, Username: "driver", DisplayName: "司机李师傅", RoleCode: "driver", PasswordHash: driverHash, PasswordSalt: driverSalt, Status: "active"},
			{ID: 4, CompanyID: 1, CustomerID: 1, Username: "customer", DisplayName: "客户王经理", RoleCode: "customer", PasswordHash: customerHash, PasswordSalt: customerSalt, Status: "active"},
			{ID: 5, CompanyID: 1, SiteID: 1, Username: "quality", DisplayName: "实验室质检员", RoleCode: "quality", PasswordHash: qualityHash, PasswordSalt: qualitySalt, Status: "active"},
		},
		OIDCProviders: []OIDCProvider{
			{
				ID: 1, Name: "企业 OIDC 演示", Code: "enterprise-demo", Issuer: "https://idp.example.com/cbmp",
				ClientID: "cbmp-desktop", ClientSecret: "cbmp-oidc-demo-secret",
				AuthURL: "https://idp.example.com/oauth2/v1/authorize", TokenURL: "mock://oidc/token", UserInfoURL: "",
				RedirectURI: "http://127.0.0.1:8088/api/auth/sso/enterprise-demo/callback",
				Scopes:      []string{"openid", "profile", "email"}, UsernameClaim: "preferred_username", DisplayNameClaim: "name",
				RoleCode: "boss", CompanyID: 1, AutoProvision: true, Status: "enabled",
			},
		},
		SCIMProviders: []SCIMProvider{
			{
				ID: 1, Name: "企业 SCIM 演示", Code: "enterprise-scim", BearerToken: "demo-scim-token",
				CompanyID: 1, SiteID: 0, DefaultRoleCode: "customer", Status: "enabled",
				CreatedAt: "2026-06-18 09:00:00",
			},
		},
		ProductInstances: []ProductInstance{
			{
				ID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", Watermark: "CBMP-WQ-2026",
				Edition: "Operations", DeploymentMode: "private", ClientVersion: "1.0.1", ServerVersion: "1.0.1",
				Endpoint: "https://cbmp.wanqu.example", Status: "online", ProbeToken: "probe-wq-demo-token", ProbeEnabled: true,
				HealthStatus: "healthy", LastProbeAt: "2026-06-19 08:30:00", LicenseExpiresAt: "2026-07-20",
				DaysToExpire: 31, RenewalOwner: "客户成功-陈晨", RenewalStage: "报价确认", AlertLevel: "warning",
				LastHeartbeatAt: "2026-06-19 08:30:00", CreatedAt: "2026-06-18 09:00:00", Remark: "年度续费跟进中",
			},
			{
				ID: 2, CustomerName: "华南骨料运营中心", LicenseID: "LIC-HN-2026", Watermark: "CBMP-HN-OPS",
				Edition: "Professional", DeploymentMode: "private", ClientVersion: "1.0.0", ServerVersion: "1.0.1",
				Endpoint: "https://cbmp.hn.example", Status: "degraded", ProbeToken: "probe-hn-demo-token", ProbeEnabled: true,
				HealthStatus: "degraded", LastProbeAt: "2026-06-19 08:10:00", LicenseExpiresAt: "2026-06-30",
				DaysToExpire: 11, RenewalOwner: "客户成功-林岚", RenewalStage: "待客户盖章", AlertLevel: "critical",
				LastHeartbeatAt: "2026-06-19 08:10:00", CreatedAt: "2026-06-18 09:30:00", Remark: "服务端正常，客户端版本待升级",
			},
		},
		Companies: []Company{{ID: 1, Name: "湾区建材集团有限公司", Code: "WQBM", Status: "active"}},
		Sites: []Site{
			{1, 1, "南山混凝土站", "NS-CON", "深圳市南山区科苑路 88 号", 113.9345, 22.5431, "running"},
			{2, 1, "宝安综合材料站", "BA-MIX", "深圳市宝安区机场南路 21 号", 113.8293, 22.6392, "running"},
		},
		Departments: []Department{
			{1, 1, "集团运营中心", "OPS", 0, "active"},
			{2, 1, "南山站生产部", "NS-PROD", 1, "active"},
			{3, 1, "财务结算部", "FIN", 1, "active"},
		},
		Plants: []Plant{
			{1, 1, "南山 180 生产线", "NS-HZS180", "180m3/h", "opc-da/demo", "running"},
			{2, 2, "宝安综合材料线", "BA-MIX-01", "240t/h", "rest/demo", "running"},
		},
		Warehouses: []Warehouse{
			{1, 1, "南山原料仓", "NS-MAT", "raw_material", "active"},
			{2, 2, "宝安原料仓", "BA-MAT", "raw_material", "active"},
		},
		Silos: []Silo{
			{1, 1, 1, "水泥罐 1", "SILO-CEMENT-01", 1800, 1260, "active"},
			{2, 1, 3, "机制砂仓 1", "SAND-01", 2500, 840, "warning"},
			{3, 1, 4, "碎石仓 1", "STONE-01", 3000, 2380, "active"},
			{4, 2, 5, "外加剂罐 1", "ADMIX-01", 50, 18, "warning"},
			{5, 1, 2, "粉煤灰罐 1", "FLYASH-01", 900, 420, "active"},
			{6, 1, 5, "外加剂罐 2", "ADMIX-NS-01", 60, 36, "active"},
		},
		Customers: []Customer{
			{1, 1, "鹏城城市更新建设有限公司", "王经理", "13800010001", 3000000, 860000, 45, "active"},
			{2, 1, "前海市政工程有限公司", "陈工", "13800010002", 1800000, 320000, 30, "active"},
		},
		CustomerContacts: []CustomerContact{
			{1, 1, "王经理", "13800010001", "项目负责人", true, "active"},
			{2, 1, "刘工", "13800010003", "工地收货人", false, "active"},
		},
		CustomerProfiles: []CustomerProfile{
			{1, 1, "鹏城城市更新建设有限公司", "A", "low", 92, []string{"重点客户", "信用良好"}, "active", "2026-06-18 09:00:00", "system"},
			{2, 2, "前海市政工程有限公司", "B", "medium", 76, []string{"市政客户", "关注回款"}, "active", "2026-06-18 09:00:00", "system"},
		},
		CustomerComplaints: []CustomerComplaint{
			{
				ID: 1, ComplaintNo: "CP202606180001", CustomerID: 1, ProjectID: 1,
				Title: "到场等待时间过长", Content: "客户反馈午高峰到场排队 25 分钟", Level: "medium",
				Status: "open", Owner: "调度主管", SLAHours: 24, DueAt: "2026-06-19 12:10:00", SLAStatus: "on_track",
				CreatedAt: "2026-06-18 12:10:00",
			},
		},
		TaxRates: []TaxRate{
			{1, "建材销售 13%", 0.13, "sales", "active"},
			{2, "运输服务 9%", 0.09, "transport", "active"},
		},
		PricePolicies: []PricePolicy{
			{ID: 1, CustomerID: 1, ProjectID: 1, ProductID: 1, CustomerGrade: "A", FloorPrice: 500, SalePrice: 520, TaxRateID: 1, EffectiveFrom: "2026-06-01", EffectiveTo: "2027-05-31", Status: "active"},
			{ID: 2, CustomerID: 1, ProjectID: 1, ProductID: 2, CustomerGrade: "A", FloorPrice: 580, SalePrice: 610, TaxRateID: 1, EffectiveFrom: "2026-06-01", EffectiveTo: "2027-05-31", Status: "active"},
			{ID: 3, ProductID: 1, CustomerGrade: "B", FloorPrice: 510, SalePrice: 535, TaxRateID: 1, EffectiveFrom: "2026-06-01", EffectiveTo: "2027-05-31", Status: "active"},
			{ID: 4, CustomerID: 1, ProjectID: 1, ProductID: 1, CustomerGrade: "A", Region: "南山", MinQuantity: 50, MaxQuantity: 120, FloorPrice: 500, SalePrice: 540, PromotionName: "南山重点项目阶梯优惠", PromotionType: "percent", PromotionValue: 0.1, Priority: 20, TaxRateID: 1, EffectiveFrom: "2026-06-01", EffectiveTo: "2027-05-31", Status: "active"},
			{ID: 5, CustomerID: 1, ProjectID: 1, ProductID: 2, CustomerGrade: "A", Region: "南山", MinQuantity: 30, FloorPrice: 575, SalePrice: 625, PromotionName: "C40 大方量立减", PromotionType: "fixed", PromotionValue: 20, Priority: 15, TaxRateID: 1, EffectiveFrom: "2026-06-01", EffectiveTo: "2027-05-31", Status: "active"},
		},
		Suppliers: []Supplier{
			{1, "粤港砂石供应链", "黄总", "13700020001", "active"},
			{2, "华南水泥集团", "刘经理", "13700020002", "active"},
		},
		Carriers: []Carrier{
			{1, "自有车队", "车队长张工", "13900039999", "按趟结算", "active"},
			{2, "外协车队", "吴总", "13900038888", "月结", "active"},
		},
		Projects: []Project{
			{1, 1, "科技园二期地下室", "深圳市南山区高新南一道", "王经理", "13800010001", 113.9452, 22.5358, "active"},
			{2, 2, "前海滨海道路改造", "深圳市前海合作区临海大道", "陈工", "13800010002", 113.8891, 22.5148, "active"},
		},
		Products: []Product{
			{1, "concrete", "预拌混凝土", "C30 P8", "m3", 520, 365, true, "active"},
			{2, "concrete", "预拌混凝土", "C40", "m3", 610, 430, true, "active"},
			{3, "mortar_wet", "湿拌砂浆", "M10", "t", 280, 205, true, "active"},
			{4, "asphalt", "沥青混合料", "AC-13", "t", 720, 540, true, "active"},
			{5, "stabilized_soil", "水稳混合料", "5% 水泥剂量", "t", 165, 118, true, "active"},
			{6, "aggregate", "机制砂", "0-5mm", "t", 96, 66, false, "active"},
		},
		Materials: []Material{
			{1, "水泥", "P.O 42.5", "t", 800, "active"},
			{2, "粉煤灰", "II 级", "t", 300, "active"},
			{3, "机制砂", "0-5mm", "t", 1200, "active"},
			{4, "碎石", "5-25mm", "t", 1600, "active"},
			{5, "外加剂", "聚羧酸", "t", 25, "active"},
		},
		Drivers: []Driver{
			{1, "李师傅", "13900030001", "A2-440301", "2028-08-20", "active"},
			{2, "赵师傅", "13900030002", "A2-440302", "2027-03-12", "active"},
			{3, "周师傅", "13900030003", "A2-440303", "2026-12-01", "active"},
			{4, "吴师傅", "13900030004", "B2-440304", "2028-11-15", "active"},
		},
		Vehicles: []Vehicle{
			{1, "粤B12345", "搅拌车", "12m3", "自有车队", 1, 1, "online", "in_transit", "2027-09-12", "active"},
			{2, "粤B22336", "搅拌车", "12m3", "自有车队", 1, 2, "online", "waiting_load", "2027-10-02", "active"},
			{3, "粤B77889", "粉料罐车", "32t", "外协车队", 2, 3, "offline", "idle", "2026-10-20", "active"},
			{4, "粤B66778", "自卸车", "28t", "外协车队", 2, 4, "online", "idle", "2028-01-18", "active"},
		},
		VehicleDevices: []VehicleDevice{
			{1, 1, "GPS1000001", "http", "北斗终端", "online", "2026-06-18 11:02:00"},
			{2, 2, "GPS1000002", "mqtt", "北斗终端", "online", "2026-06-18 11:03:00"},
			{3, 3, "GPS1000003", "tcp", "第三方平台", "offline", "2026-06-18 08:02:00"},
			{4, 4, "APP1000004", "driver_app", "司机 App", "online", "2026-06-18 11:03:20"},
			{5, 0, "PLANT-NS-HZS180", "opc", "拌合楼控制系统", "online", "2026-06-18 10:20:00"},
		},
		DeviceCredentials: []DeviceCredential{
			{1, "GPS1000001", sha256Hex("device-demo-key-1"), []string{"location:report"}, "active", ""},
			{2, "GPS1000002", sha256Hex("device-demo-key-2"), []string{"location:report"}, "active", ""},
			{3, "GPS1000003", sha256Hex("device-demo-key-3"), []string{"location:report"}, "active", ""},
			{4, "APP1000004", sha256Hex("driver-app-demo-key"), []string{"location:report"}, "active", ""},
			{5, "NS-SCALE-01", sha256Hex("scale-demo-key-1"), []string{"scale:report"}, "active", ""},
			{6, "PLANT-NS-HZS180", sha256Hex("plant-demo-key-1"), []string{"plant:report"}, "active", ""},
		},
		SecurityPolicies: []SecurityPolicy{
			{1, "设备接口必须使用密钥", "device_key_required", "iot", true, "GPS/司机 App/第三方平台上报接口"},
			{2, "接口签名头", "api_signature_header", "X-Device-Key", true, "私有化交付默认设备密钥头"},
			{3, "本机运维白名单", "ip_whitelist", "127.0.0.1", true, "设置 CBMP_ENFORCE_IP_WHITELIST=1 后启用"},
			{4, "内网运维白名单", "ip_whitelist", "192.168.0.0/16", true, "客户内网运维段"},
			{5, "会话超时分钟", "session_timeout_minutes", "480", true, "超过该时间未重新登录会话失效"},
			{6, "单用户最大会话数", "session_max_per_user", "5", true, "同一用户超过上限会淘汰最早会话"},
		},
		FieldPolicies: []FieldPolicy{
			{1, "customer", "drivers", "phone", "phone", true, "客户侧查看司机手机号脱敏"},
			{2, "customer", "drivers", "licenseNo", "code", true, "客户侧查看司机证件号脱敏"},
			{3, "driver", "customers", "phone", "phone", true, "司机侧客户手机号脱敏"},
			{4, "driver", "customerContacts", "phone", "phone", true, "司机侧客户联系人手机号脱敏"},
			{5, "driver", "projects", "phone", "phone", true, "司机侧工地手机号脱敏"},
			{6, "driver", "orders", "phone", "phone", true, "司机侧订单收货手机号脱敏"},
			{7, "driver", "deliverySigns", "phone", "phone", true, "司机侧签收手机号脱敏"},
			{8, "device", "*", "*", "redact", true, "设备角色默认隐藏业务敏感字段"},
		},
		Contracts: []Contract{
			{
				ID: 1, CustomerID: 1, ProjectID: 1, ContractNo: "CT202606180001", Version: 1,
				Name: "科技园二期混凝土供应合同", ValidFrom: "2026-06-01", ValidTo: "2027-05-31",
				CreditPolicy: "信用额度内自动放行", TotalAmount: 5200000, UsedAmount: 624000,
				ApprovedAt: "2026-06-18 09:00:00", ApprovedBy: "system", Status: "active",
				Items: []ContractItem{{1, "m3", 10000, 520}, {2, "m3", 3000, 610}},
			},
		},
		ContractAttachments: []ContractAttachment{
			{1, 1, 1, "科技园二期供应合同.pdf", "contract_pdf", "vault://contracts/CT202606180001.pdf", "sha256:demo-contract", "active", "system", "2026-06-18 09:10:00"},
		},
		Orders: []SalesOrder{
			{
				ID: 1, OrderNo: "SO202606180001", CustomerID: 1, ProjectID: 1, ProductID: 1, SiteID: 1,
				ProductLine: "concrete", PlanQuantity: 180, Unit: "m3", UnitPrice: 520, TotalAmount: 93600,
				Lines:    []SalesOrderLine{{ID: 1, Seq: 1, ProductID: 1, ProductLine: "concrete", ProductName: "C30 商品混凝土", StrengthGrade: "C30", Slump: "180mm", PouringPart: "地下室底板", Quantity: 180, Unit: "m3", UnitPrice: 520, FloorPrice: 500, TaxRate: 0.13, Amount: 93600, PriceSource: "seed"}},
				PlanTime: "2026-06-18 14:00:00", ReceiveAddress: "深圳市南山区高新南一道", Contact: "王经理", Phone: "13800010001",
				SettlementMode: "月结", TransportMode: "自有车队", StrengthGrade: "C30", Slump: "180mm", PouringPart: "地下室底板", PumpMode: "车载泵",
				DispatchedQty: 72, SignedQty: 36, Status: "delivering", CreatedAt: "2026-06-18 09:10:00",
			},
			{
				ID: 2, OrderNo: "SO202606180002", CustomerID: 1, ProjectID: 1, ProductID: 2, SiteID: 1,
				ProductLine: "concrete", PlanQuantity: 120, Unit: "m3", UnitPrice: 610, TotalAmount: 73200,
				Lines:    []SalesOrderLine{{ID: 2, Seq: 1, ProductID: 2, ProductLine: "concrete", ProductName: "C40 商品混凝土", StrengthGrade: "C40", Slump: "160mm", PouringPart: "剪力墙", Quantity: 120, Unit: "m3", UnitPrice: 610, FloorPrice: 590, TaxRate: 0.13, Amount: 73200, PriceSource: "seed"}},
				PlanTime: "2026-06-19 08:30:00", ReceiveAddress: "深圳市南山区高新南一道", Contact: "王经理", Phone: "13800010001",
				SettlementMode: "月结", TransportMode: "自有车队", StrengthGrade: "C40", Slump: "160mm", PouringPart: "剪力墙", PumpMode: "泵车",
				Status: "approved", CreatedAt: "2026-06-18 09:25:00",
			},
		},
		ProductionPlans: []ProductionPlan{
			{1, "PP202606180001", 1, 1, 1, 180, 72, "2026-06-18", "白班", "ok", "ok", "ok", "producing"},
		},
		MixDesigns: []MixDesign{
			{
				ID: 1, ProductID: 1, SiteID: 1, Code: "MD-C30-P8", Version: "v3", StrengthGrade: "C30", Slump: "180mm",
				Scope: "地下室底板抗渗混凝土", Status: "approved", IsCurrent: true, EffectiveFrom: "2026-01-01", EffectiveTo: "2026-12-31",
				ApprovedBy: "实验室主任", ApprovedAt: "2026-01-01 09:00:00", CreatedBy: "system", CreatedAt: "2026-01-01 08:30:00", UpdatedAt: "2026-01-01 09:00:00",
				Materials: []MixDesignMaterial{{1, 320, "kg/m3"}, {2, 70, "kg/m3"}, {3, 790, "kg/m3"}, {4, 1040, "kg/m3"}, {5, 7.5, "kg/m3"}},
			},
			{
				ID: 2, ProductID: 2, SiteID: 1, Code: "MD-C40", Version: "v2", StrengthGrade: "C40", Slump: "160mm",
				Scope: "剪力墙泵送混凝土", Status: "approved", IsCurrent: true, EffectiveFrom: "2026-01-01", EffectiveTo: "2026-12-31",
				ApprovedBy: "实验室主任", ApprovedAt: "2026-01-01 09:20:00", CreatedBy: "system", CreatedAt: "2026-01-01 08:45:00", UpdatedAt: "2026-01-01 09:20:00",
				Materials: []MixDesignMaterial{{1, 380, "kg/m3"}, {2, 60, "kg/m3"}, {3, 760, "kg/m3"}, {4, 1020, "kg/m3"}, {5, 8.2, "kg/m3"}},
			},
		},
		MixDesignTrialRuns: []MixDesignTrialRun{
			{ID: 1, TrialNo: "MTR202606180001", MixDesignID: 1, ProductID: 1, SiteID: 1, TargetStrength: "C30", Slump: "180mm", Water: 165, SandRate: 42, AdmixtureRate: 1.2, Strength7d: 31.8, Strength28d: 42.5, Result: "passed", Conclusion: "工作性和强度满足生产放行", Tester: "实验室质检员", TestedAt: "2026-06-18 08:30:00", CreatedAt: "2026-06-18 08:20:00", Remark: "配比 v3 月度复核"},
		},
		ProductionTasks: []ProductionTask{
			{1, "PT202606180001", 1, 1, 1, 1, 1, 180, 72, "running", "2026-06-18 10:00:00", "", "2026-06-18 09:50:00", "2026-06-18 10:20:00"},
		},
		ProductionBatches: []ProductionBatch{
			{1, "PB202606180001", 1, 1, 1, 1, 1, 1, 72, "NS-HZS180", "生产员-陈工", "passed", "released", "2026-06-18 10:00:00", "2026-06-18 10:20:00"},
		},
		ProductionReports: []ProductionDailyReport{
			{1, "PDR202606180001", 1, "2026-06-18", 180, 72, 1, 26280, 1, 0, "generated", "2026-06-18 18:00:00"},
		},
		Inventory: []InventoryItem{
			{1, 1, "南山原料仓", "SILO-CEMENT-01", 1, "CEM-20260618-A", 0, 2, 1260, "t", "passed", "available", "2026-06-18 08:00:00"},
			{2, 1, "南山原料仓", "SAND-01", 3, "SAND-20260617-A", 1, 1, 840, "t", "passed", "warning", "2026-06-18 08:00:00"},
			{3, 1, "南山原料仓", "STONE-01", 4, "STONE-20260616-A", 0, 1, 2380, "t", "passed", "available", "2026-06-18 08:00:00"},
			{4, 2, "宝安原料仓", "ADMIX-01", 5, "ADM-20260618-A", 0, 2, 18, "t", "passed", "warning", "2026-06-18 08:00:00"},
			{5, 1, "南山原料仓", "FLYASH-01", 2, "FLY-20260618-A", 0, 2, 420, "t", "passed", "available", "2026-06-18 08:00:00"},
			{6, 1, "南山原料仓", "ADMIX-NS-01", 5, "ADM-NS-20260618-A", 0, 2, 36, "t", "passed", "available", "2026-06-18 08:00:00"},
		},
		InventoryBatchTraces: []InventoryBatchTrace{
			{1, "IBT202606180001", 1, "PB202606180001", 0, 1, 1, 1, 2, "CEM-20260618-A", "南山原料仓", "SILO-CEMENT-01", 23.04, "t", "2026-06-18 10:20:00"},
			{2, "IBT202606180002", 1, "PB202606180001", 0, 3, 1, 4, 1, "STONE-20260616-A", "南山原料仓", "STONE-01", 74.88, "t", "2026-06-18 10:20:00"},
		},
		PurchaseRequests: []PurchaseRequest{
			{1, "PR202606180001", 1, 3, 1200, "t", "2026-06-19 08:00:00", "approved", "2026-06-18 08:20:00"},
		},
		PurchaseOrders: []PurchaseOrder{
			{1, "PO202606180001", 1, 1, 3, 1200, 78, "t", "inbound", "2026-06-18 08:30:00"},
		},
		RawMaterialReceipts: []RawMaterialReceipt{
			{1, "RI202606180001", 1, 1, 2, 1, 3, "粤B原料01", 49.8, 18.4, 31.4, "passed", "stocked", "2026-06-18 09:00:00"},
		},
		LaboratorySamples: []LaboratorySample{
			{ID: 1, SampleNo: "LS202606180001", SourceType: "manual", SiteID: 1, ProductID: 1, MixDesignID: 1, SampleType: "mix_design_trial", Status: "completed", Result: "passed", PlannedTestAt: "2026-06-18", CollectedAt: "2026-06-18 08:15:00", CreatedBy: "实验室质检员", Remark: "C30 配比月度复核样"},
		},
		LaboratoryTests: []LaboratoryTestRecord{
			{ID: 1, TestNo: "LT202606180001", SampleID: 1, EquipmentID: 1, SiteID: 1, TestType: "compressive_strength", Metric: "28d_strength", Value: 42.5, Unit: "MPa", Result: "passed", Status: "reviewed", Tester: "实验室质检员", TestedAt: "2026-06-18 08:30:00", Reviewer: "实验室主任", ReviewedAt: "2026-06-18 08:45:00", Remark: "满足 C30 配比复核要求"},
		},
		LaboratoryEquipment: []LaboratoryEquipment{
			{ID: 1, EquipmentNo: "EQ-LAB-001", Name: "压力试验机", SiteID: 1, Model: "YES-2000", SerialNo: "NS-LAB-2000-01", Status: "active", CalibrationCycleDays: 180, LastCalibrationAt: "2026-06-01", NextCalibrationAt: "2026-11-28", CreatedAt: "2026-01-01 08:00:00", Remark: "混凝土试块抗压"},
			{ID: 2, EquipmentNo: "EQ-LAB-002", Name: "坍落度筒", SiteID: 1, Model: "STD-300", SerialNo: "NS-LAB-SLUMP-01", Status: "active", CalibrationCycleDays: 365, LastCalibrationAt: "2026-03-01", NextCalibrationAt: "2027-03-01", CreatedAt: "2026-01-01 08:10:00", Remark: "出厂工作性检测"},
		},
		LaboratoryCalibrations: []LaboratoryCalibration{
			{ID: 1, CalibrationNo: "LC202606010001", EquipmentID: 1, SiteID: 1, Result: "passed", CalibratedAt: "2026-06-01", NextDueAt: "2026-11-28", CertificateNo: "CAL-NS-20260601", Agency: "深圳计量检测中心", Operator: "实验室主任", Remark: "半年校准"},
		},
		QualityExceptions: []QualityException{
			{ID: 1, ExceptionNo: "QE202606180001", SourceType: "transport", SourceID: 2, SiteID: 1, Severity: "medium", Title: "运输时长接近质量风险阈值", Description: "DO202606180002 距离签收超时预警阈值较近，实验室持续关注坍落度复测。", Status: "open", Responsible: "南山站调度", CreatedAt: "2026-06-18 11:40:00"},
		},
		InventoryFlows: []InventoryFlow{
			{1, "IF202606180001", 1, 3, "raw_material_receipt", 1, "in", 31.4, 840, "原料入厂过磅入库", "2026-06-18 09:00:00"},
			{2, "IF202606180002", 1, 1, "production_batch", 1, "out", 23.04, 1260, "C30 理论扣减", "2026-06-18 10:20:00"},
			{3, "IF202606180003", 1, 4, "production_batch", 1, "out", 74.88, 2380, "C30 理论扣减", "2026-06-18 10:20:00"},
		},
		DispatchOrders: []DispatchOrder{
			{1, "DO202606180001", 1, 1, 1, 1, 1, 36, 36, 36, "A013", "2026-06-18 11:35:00", "completed", "", "2026-06-18 09:40:00", "2026-06-18 11:45:00", 1, 1, 1, "C30 商品混凝土"},
			{2, "DO202606180002", 1, 2, 2, 1, 1, 36, 36, 0, "A014", "2026-06-18 12:10:00", "in_transit", "", "2026-06-18 10:15:00", "2026-06-18 10:55:00", 1, 1, 1, "C30 商品混凝土"},
		},
		DispatchSchedules: []DispatchSchedule{
			{1, "DS202606180001", 1, 1, 1, 1, "2026-06-18", "day", 120, 36, "active", "2026-06-18 08:00:00", "2026-06-18 09:40:00"},
			{2, "DS202606180002", 1, 2, 2, 1, "2026-06-18", "day", 120, 36, "active", "2026-06-18 08:00:00", "2026-06-18 10:15:00"},
		},
		ScaleDevices: []ScaleDevice{
			{1, 1, "南山 1 号地磅", "NS-SCALE-01", "serial", "192.168.10.21", "online"},
			{2, 2, "宝安 1 号地磅", "BA-SCALE-01", "tcp", "192.168.20.21", "online"},
		},
		ScaleTickets: []ScaleTicket{
			{1, "ST202606180001", "product_out", 1, 1, 1, 1, "粤B12345", 30.8, 18.7, 12.1, "t", "capture://ns-scale-01/product-gross.jpg", 1, "signed", "pending", "valid", "2026-06-18 10:05:00", 0, 0, 0, 0, 0, ""},
			{2, "ST202606180002", "raw_material_in", 0, 0, 1, 0, "粤B原料01", 49.8, 18.4, 31.4, "t", "capture://ns-scale-01/raw-ri202606180001-gross.jpg", 1, "not_required", "pending", "valid", "2026-06-18 09:00:00", 1, 1, 3, 0, 0, "原料入厂"},
		},
		ScaleWeightRecords: []ScaleWeightRecord{
			{1, 1, 1, "粤B12345", 18.7, "tare", "capture://ns-scale-01/tare.jpg", "2026-06-18 09:58:00"},
			{2, 1, 1, "粤B12345", 30.8, "gross", "capture://ns-scale-01/gross.jpg", "2026-06-18 10:05:00"},
			{3, 1, 2, "粤B原料01", 18.4, "tare", "capture://ns-scale-01/raw-ri202606180001-tare.jpg", "2026-06-18 08:55:00"},
			{4, 1, 2, "粤B原料01", 49.8, "gross", "capture://ns-scale-01/raw-ri202606180001-gross.jpg", "2026-06-18 09:00:00"},
		},
		DeliveryNotes: []DeliveryNote{
			{1, "DN202606180001", 1, 1, 1, "qr://DN202606180001", "signed", "2026-06-18 10:05:10"},
		},
		DeliverySignLinks: []DeliverySignLink{
			{ID: 1, LinkNo: "SL202606180001", DispatchID: 1, TicketID: 1, OrderID: 1, LineID: 1, LineSeq: 1, ProductID: 1, ProductName: "C30 商品混凝土", CustomerID: 1, ProjectID: 1, Channel: "qr", Phone: "13800010001", Token: "seed-sign-token-used", URL: "/public/sign/seed-sign-token-used", QRCode: "qr://SL202606180001", Status: "used", SentAt: "2026-06-18 10:06:00", ExpiresAt: "2026-06-25 10:06:00", UsedAt: "2026-06-18 11:45:00", CreatedBy: "dispatcher", CreatedAt: "2026-06-18 10:06:00"},
		},
		TicketPrintLogs: []TicketPrintLog{
			{1, 1, "dispatcher", "2026-06-18 10:05:20"},
		},
		DeliverySigns: []DeliverySign{
			{ID: 1, SignNo: "DS202606180001", DispatchID: 1, LinkID: 1, TicketID: 1, OrderID: 1, LineID: 1, LineSeq: 1, ProductID: 1, ProductName: "C30 商品混凝土", CustomerID: 1, ProjectID: 1, Signer: "王经理", Phone: "13800010001", SignedQty: 36, Longitude: 113.9452, Latitude: 22.5358, Photo: "现场照片已归档", Signature: "电子签名已归档", Remark: "数量无异议", SignedAt: "2026-06-18 11:45:00"},
		},
		DeliverySignAttachments: []DeliverySignAttachment{
			{ID: 1, SignID: 1, DispatchID: 1, TicketID: 1, FileName: "site-delivery-20260618.jpg", FileType: "photo", URL: "photo://delivery/site-delivery-20260618.jpg", Checksum: "sha256:seed-site-photo", UploadedBy: "王经理", UploadedAt: "2026-06-18 11:45:10"},
		},
		Statements: []Statement{
			{1, "CS2026060001", 1, 1, "2026-06", 36, 18720, "", "", "draft", []StatementItem{{1, 1, 1, 1, 1, 36, 520, 18720}}},
		},
		SalesInvoices: []SalesInvoice{
			{1, "INV202606180001", 1, 1, 18720, 0.13, 2433.6, "TAX202606180001", "submitted", "invoice://INV202606180001.pdf", "", "issued", "2026-06-18 13:30:00", "blue", "blue_vat_special", 0, 0, "", "", ""},
		},
		TaxGatewaySubmissions: []TaxGatewaySubmission{
			{1, "TGS202606180001", 1, "INV202606180001", "issue", "local-simulator", "tax://local-simulator", "seed-tax-202606180001", "submitted", "TAX202606180001", "invoice://INV202606180001.pdf", "", 1, 12, "2026-06-18 13:30:00", "2026-06-18 13:30:01", "system"},
		},
		Receivables: []Receivable{
			{1, "AR202606180001", 1, 1, 1, 18720, 8000, "2026-07-31", "partial", "2026-06-18 13:31:00"},
		},
		Receipts: []Receipt{
			{1, "RC202606180001", 1, 1, 8000, "bank", "confirmed", "2026-06-18 14:00:00"},
		},
		PaymentPlans: []PaymentPlan{
			{1, "PP202606180001", 1, 1, 5000, "2026-07-15", "bank", "planned", "2026-06-18 14:10:00", "", "客户承诺首期回款"},
		},
		SupplierStatements: []SupplierStatement{
			{1, "SS2026060001", 1, "2026-06", 2449.2, "draft", "", ""},
		},
		Payables: []Payable{
			{1, "AP202606180001", 1, 0, 2449.2, 0, "2026-07-15", "open"},
		},
		TransportSettlements: []TransportSettlement{
			{1, "TS2026060001", 1, "2026-06", 2, 960, "draft"},
		},
		TransportSettlementItems: []TransportSettlementItem{
			{1, 1, 1, "DO202606180001", 1, 1, 1, 36, 480, "pending", "2026-06-18 12:00:00"},
		},
		CostCalcs: []CostCalc{
			{1, 1, 1, 13140, 480, 260, 180, 14060, "2026-06-18 12:00:00"},
		},
		ProjectProfits: []ProjectProfit{
			{1, 1, 18720, 14060, 4660, 24.89, "2026-06"},
		},
		Locations: []VehicleLocationEvent{
			{1, 1, "粤B12345", 1, 2, "GPS1000001", "gps_device", 113.9380, 22.5410, 42.5, 180, 123456.7, 1, "online", "南山区科苑路", false, "", "2026-06-18 10:58:00", "2026-06-18 10:58:01"},
			{2, 1, "粤B12345", 1, 2, "GPS1000001", "gps_device", 113.9410, 22.5392, 38.2, 180, 123459.1, 1, "online", "高新南一道", false, "", "2026-06-18 11:02:00", "2026-06-18 11:02:01"},
			{3, 2, "粤B22336", 2, 0, "GPS1000002", "gps_device", 113.9346, 22.5433, 0, 0, 8561.3, 1, "online", "南山混凝土站", false, "", "2026-06-18 11:03:00", "2026-06-18 11:03:01"},
			{4, 4, "粤B66778", 4, 0, "GPS1000004", "driver_app", 113.8296, 22.6389, 16.5, 90, 20410.1, 1, "online", "宝安机场南路", false, "", "2026-06-18 11:03:20", "2026-06-18 11:03:21"},
		},
		LatestLocations: []VehicleLatestLocation{
			{1, "粤B12345", 113.9410, 22.5392, 38.2, 180, "online", "in_transit", "2026-06-18 11:02:00", 1, 1, 1, 1},
			{2, "粤B22336", 113.9346, 22.5433, 0, 0, "online", "waiting_load", "2026-06-18 11:03:00", 0, 0, 1, 0},
			{3, "粤B77889", 113.8293, 22.6392, 0, 0, "offline", "idle", "2026-06-18 08:02:00", 0, 0, 2, 0},
			{4, "粤B66778", 113.8296, 22.6389, 16.5, 90, "online", "idle", "2026-06-18 11:03:20", 0, 0, 2, 0},
		},
		GeoFences: []GeoFence{
			{ID: 1, Name: "南山混凝土站围栏", Type: "site", SiteID: 1, Longitude: 113.9345, Latitude: 22.5431, Radius: 450, Shape: "circle", Status: "active"},
			{ID: 2, Name: "宝安综合材料站围栏", Type: "site", SiteID: 2, Longitude: 113.8293, Latitude: 22.6392, Radius: 500, Shape: "circle", Status: "active"},
			{ID: 3, Name: "科技园二期工地围栏", Type: "project", ProjectID: 1, Longitude: 113.9452, Latitude: 22.5358, Radius: 350, Shape: "circle", Status: "active"},
			{ID: 4, Name: "前海道路工地围栏", Type: "project", ProjectID: 2, Longitude: 113.8891, Latitude: 22.5148, Radius: 600, Shape: "circle", Status: "active"},
		},
		VehicleAlarms: []VehicleAlarm{
			{1, 3, 0, "vehicle_offline", "medium", "粤B77889 超过 3 小时无定位", "open", "2026-06-18 11:00:00", "", ""},
		},
		RuleDefinitions: []RuleDefinition{
			{1, "vehicle_offline_5m", "车辆离线超过 5 分钟", "vehicle", "offline_minutes", ">", 5, "medium", true, []string{"dispatcher"}, "离线车辆提醒调度员"},
			{2, "speed_over_80", "车辆超速 80km/h", "vehicle", "speed", ">", 80, "high", true, []string{"dispatcher", "boss"}, "超速实时预警"},
			{3, "concrete_timeout_90m", "混凝土 90 分钟未签收", "quality", "transport_minutes", ">", 90, "high", true, []string{"dispatcher", "quality"}, "混凝土质量风险"},
			{4, "site_arrived_unsigned_15m", "到场 15 分钟未签收", "delivery", "arrived_unsigned_minutes", ">", 15, "medium", true, []string{"dispatcher", "customer"}, "提醒工地签收"},
		},
		Notifications: []Notification{
			{1, "dispatcher", "system", "车辆离线预警", "粤B77889 超过 3 小时无定位", "unread", "2026-06-18 11:00:00"},
			{2, "boss", "system", "库存预警", "机制砂库存低于安全线", "unread", "2026-06-18 08:00:00"},
		},
		IntegrationEndpoints: []IntegrationEndpoint{
			{1, "API Gateway", "gateway", "http", "http://127.0.0.1:8088/api", "online", "2026-06-18 11:00:00"},
			{2, "GPS IoT Gateway", "iot", "mqtt/http/tcp", "iot://vehicle-location", "online", "2026-06-18 11:03:00"},
			{3, "地磅设备 Gateway", "weighbridge", "serial/tcp", "scale://standard", "online", "2026-06-18 10:05:00"},
			{4, "拌合楼接口 Gateway", "plant", "opc/rest", "plant://ns-hzs180", "online", "2026-06-18 10:20:00"},
			{5, "财务税控接口", "finance", "rest", "tax://demo", "standby", "2026-06-18 13:30:00"},
			{6, "催收短信通道", "collection", "sms", "collection://local-simulator/sms", "online", "2026-06-18 13:35:00"},
			{7, "催收企微通道", "collection", "wecom", "collection://local-simulator/wecom", "online", "2026-06-18 13:35:00"},
		},
		CollectionTemplates: []CollectionTemplate{
			{1, "pre_due_sms", "到期前短信提醒", "pre_due", "sms", "{{customerName}}：贵司应收 {{amount}} 元将于 {{dueDate}} 到期，请提前确认回款计划。", true, "2026-06-18 13:35:00"},
			{2, "urgent_phone", "逾期电话跟进话术", "urgent", "phone", "{{customerName}} 应收 {{amount}} 元已逾期 {{overdueDays}} 天，请财务负责人当日确认付款安排。", true, "2026-06-18 13:35:00"},
			{3, "legal_wecom", "法务企微通知", "legal", "wecom", "{{customerName}} 应收 {{amount}} 元已逾期 {{overdueDays}} 天，系统将同步法务催收流程，请尽快回款。", true, "2026-06-18 13:35:00"},
		},
		ApprovalFlows: []ApprovalFlow{
			{1, "order_credit_risk", "超信用订单审批", "sales_order", []ApprovalStep{{1, "dispatcher", "approve"}, {2, "boss", "approve"}}, "active"},
			{2, "ticket_void", "票据作废审批", "scale_ticket", []ApprovalStep{{1, "dispatcher", "approve"}, {2, "boss", "approve"}}, "active"},
			{3, "price_below_floor", "低于底价审批", "price_policy", []ApprovalStep{{1, "dispatcher", "approve"}, {2, "boss", "approve"}}, "active"},
			{4, "inventory_transfer", "库存调拨审批", "inventory_transfer", []ApprovalStep{{1, "dispatcher", "approve"}, {2, "boss", "approve"}}, "active"},
			{5, "contract_version", "合同版本审批", "contract", []ApprovalStep{{1, "dispatcher", "approve"}, {2, "boss", "approve"}}, "active"},
		},
		DataDictionaries: []DataDictionary{
			{1, "product_line", "concrete", "混凝土", 1, "active"},
			{2, "product_line", "mortar_wet", "湿拌砂浆", 2, "active"},
			{3, "product_line", "asphalt", "沥青混合料", 3, "active"},
			{4, "product_line", "stabilized_soil", "水稳", 4, "active"},
			{5, "dispatch_status", "assigned", "已派车", 1, "active"},
			{6, "dispatch_status", "in_transit", "运输中", 2, "active"},
			{7, "ticket_type", "product_out", "成品出厂", 1, "active"},
			{8, "ticket_type", "raw_material_in", "原料入厂", 2, "active"},
			{9, "invoice_type", "blue_vat_special", "蓝字增值税专用发票", 1, "active"},
			{10, "invoice_type", "blue_vat_normal", "蓝字增值税普通发票", 2, "active"},
			{11, "invoice_type", "blue_e_invoice", "蓝字电子普通发票", 3, "active"},
			{12, "invoice_type", "red_vat_special", "红字增值税专用发票", 4, "active"},
			{13, "invoice_type", "red_vat_normal", "红字增值税普通发票", 5, "active"},
			{14, "invoice_type", "red_e_invoice", "红字电子普通发票", 6, "active"},
		},
		Updates: []UpdatePackage{
			{ID: 1, Version: "1.0.0", Component: "server", Channel: "stable", Status: "installed", Checksum: "sha256:server-100", Signature: "sig:server-100", CreatedAt: "2026-06-18 08:00:00", Remark: "服务端初始交付版本"},
			{ID: 2, Version: "1.0.1", Component: "server", Channel: "gray", Status: "available", Checksum: "sha256:server-101", Signature: "sig:server-101", RollbackVersion: "1.0.0", CreatedAt: "2026-06-18 09:30:00", Remark: "服务端异常告警与更新策略灰度包"},
			{ID: 3, Version: "1.0.1", Component: "client", Channel: "stable", Status: "available", Checksum: "sha256:client-101", Signature: "sig:client-101", RollbackVersion: "1.0.0", CreatedAt: "2026-06-18 10:00:00", Remark: "Wails 客户端业务工作台更新包"},
		},
		SystemAlerts: []SystemAlert{
			{ID: 1, AlertNo: "AL202606190001", InstanceID: 2, CustomerName: "华南骨料运营中心", Severity: "critical", Source: "client", Title: "客户端版本落后", Message: "客户现场仍有 1.0.0 客户端，建议推送 1.0.1 更新包", Status: "open", GroupKey: "instance:2|source:client|component:client|metric:update_version|title:客户端版本落后", PolicyNo: "AP202606190001", EventCount: 2, SuppressedUntil: "2026-06-19 08:40:00", FirstSeenAt: "2026-06-19 08:10:00", LastSeenAt: "2026-06-19 08:10:00"},
			{ID: 2, AlertNo: "AL202606190002", InstanceID: 1, CustomerName: "湾区建材集团", Severity: "warning", Source: "license", Title: "授权即将到期", Message: "授权剩余约 31 天，请推进续费确认", Status: "open", GroupKey: "instance:1|source:license|component:license|metric:license_expire|title:授权即将到期", PolicyNo: "AP202606190002", EventCount: 1, FirstSeenAt: "2026-06-19 08:30:00", LastSeenAt: "2026-06-19 08:30:00"},
		},
		ProductRenewalTasks: []ProductRenewalTask{
			{ID: 1, TaskNo: "RN202606190001", InstanceID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", Stage: "报价确认", Status: "open", Owner: "客户成功-陈晨", Amount: 128000, Currency: "CNY", DueDate: "2026-07-20", NextFollowAt: "2026-06-21 10:00:00", RiskLevel: "warning", LastContactAt: "2026-06-18 16:30:00", CreatedAt: "2026-06-18 09:20:00", Remark: "客户已收到续费报价，等待采购确认"},
			{ID: 2, TaskNo: "RN202606190002", InstanceID: 2, CustomerName: "华南骨料运营中心", LicenseID: "LIC-HN-2026", Stage: "待客户盖章", Status: "open", Owner: "客户成功-林岚", Amount: 86000, Currency: "CNY", DueDate: "2026-06-30", NextFollowAt: "2026-06-20 15:00:00", RiskLevel: "critical", LastContactAt: "2026-06-19 08:40:00", CreatedAt: "2026-06-18 09:45:00", Remark: "续费合同已发出，需当天跟进盖章"},
		},
		ProductRenewalQuotes: []ProductRenewalQuote{
			{ID: 1, QuoteNo: "RQ202606190001", TaskID: 1, InstanceID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", Amount: 128000, Currency: "CNY", Modules: []string{"dashboard", "license", "update", "report"}, NewExpiresAt: "2027-07-20", Status: "approved", PreparedBy: "客户成功-陈晨", PreparedAt: "2026-06-18 10:00:00", ApprovedBy: "平台管理员", ApprovedAt: "2026-06-18 11:20:00", Remark: "年度授权续费报价"},
		},
		ProductRenewalContracts: []ProductRenewalContract{
			{ID: 1, ContractNo: "RC202606190001", TaskID: 1, QuoteID: 1, InstanceID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", Amount: 128000, Currency: "CNY", Status: "signed", SignedBy: "客户成功-陈晨", SignedAt: "2026-06-18 15:10:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-18 15:00:00", Remark: "客户已确认年度续费合同"},
		},
		ProductRenewalPayments: []ProductRenewalPayment{
			{ID: 1, PaymentNo: "RP202606190001", TaskID: 1, ContractID: 1, InstanceID: 1, CustomerName: "湾区建材集团", Amount: 64000, Currency: "CNY", Method: "bank", Status: "paid", PaidAt: "2026-06-19 09:10:00", CreatedBy: "财务-王维", CreatedAt: "2026-06-19 09:10:00", Remark: "首期回款 50%"},
		},
		ProductRenewalApprovals: []ProductRenewalApproval{
			{ID: 1, ApprovalNo: "RA202606190001", TaskID: 1, QuoteID: 1, ContractID: 1, InstanceID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", ApprovalType: "contract", Amount: 128000, Currency: "CNY", Status: "approved", CurrentRole: "boss", RequestedBy: "客户成功-陈晨", RequestedAt: "2026-06-18 14:30:00", ApprovedBy: "平台管理员", ApprovedAt: "2026-06-18 14:45:00", Comment: "年度续费合同金额和模块通过审批"},
		},
		ProductRenewalInvoices: []ProductRenewalInvoice{
			{ID: 1, InvoiceNo: "RI202606190001", TaskID: 1, ContractID: 1, PaymentID: 1, InstanceID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", Amount: 64000, TaxRate: 0.06, TaxAmount: 3840, InvoiceType: "blue_e_invoice", Status: "issued", TaxStatus: "accepted", FileURL: "renewal-invoice://RI202606190001.pdf", CreatedBy: "财务-王维", CreatedAt: "2026-06-19 09:20:00", IssuedAt: "2026-06-19 09:21:00", ExternalRequest: "local-tax-renewal-202606190001", Remark: "首期回款电子发票"},
		},
		ProductRenewalESigns: []ProductRenewalESign{
			{ID: 1, SignNo: "RS202606190001", TaskID: 1, ContractID: 1, InstanceID: 1, CustomerName: "湾区建材集团", LicenseID: "LIC-WQ-2026", Signer: "采购负责人-赵总", Phone: "13800019999", Channel: "local_esign", Status: "signed", LinkURL: "/public/renewal-sign/RS202606190001", SentBy: "客户成功-陈晨", SentAt: "2026-06-18 15:00:00", SignedAt: "2026-06-18 15:10:00", Signature: "赵总 电子签名", Remark: "年度续费合同电子签完成"},
		},
		ProductRenewalIntegrations: []ProductRenewalIntegration{
			{ID: 1, IntegrationNo: "RI202606190001", Name: "本地电子签网关", Code: "local_esign", Provider: "local", Scenario: "esign", Endpoint: "mock://success", Status: "active", RetryLimit: 3, TimeoutSeconds: 3, LastSyncAt: "2026-06-18 15:00:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-18 14:55:00", Remark: "交付演示默认电子签通道，可替换为法大大/DocuSign"},
			{ID: 2, IntegrationNo: "RI202606190002", Name: "财务回款确认网关", Code: "finance_bank", Provider: "finance", Scenario: "payment", Endpoint: "mock://success", Status: "active", RetryLimit: 3, TimeoutSeconds: 3, LastSyncAt: "2026-06-19 09:10:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-18 14:56:00", Remark: "对接银行回单、财务系统或 ERP 收款确认"},
			{ID: 3, IntegrationNo: "RI202606190003", Name: "税控电子发票网关", Code: "tax_gateway", Provider: "tax", Scenario: "tax", Endpoint: "mock://success", Status: "active", RetryLimit: 3, TimeoutSeconds: 3, LastSyncAt: "2026-06-19 09:20:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-18 14:57:00", Remark: "对接税控/电子发票平台"},
		},
		ProductRenewalSyncRecords: []ProductRenewalSyncRecord{
			{ID: 1, SyncNo: "RSY202606190001", IntegrationID: 3, IntegrationNo: "RI202606190003", IntegrationCode: "tax_gateway", Provider: "tax", Scenario: "tax", ResourceType: "invoice", ResourceID: 1, ResourceNo: "RI202606190001", TaskID: 1, CustomerName: "湾区建材集团", Action: "issue", Status: "succeeded", AttemptCount: 1, ExternalRequestID: "tax_gateway-RSY202606190001", ExternalStatus: "accepted", RequestPayload: `{"invoiceNo":"RI202606190001"}`, ResponsePayload: `{"externalStatus":"accepted"}`, CreatedAt: "2026-06-19 09:20:00", LastAttemptAt: "2026-06-19 09:20:00", CompletedAt: "2026-06-19 09:20:01"},
		},
		ProductProbeReports: []ProductProbeReport{
			{ID: 1, ReportNo: "PR202606190001", InstanceID: 1, CustomerName: "湾区建材集团", Watermark: "CBMP-WQ-2026", Component: "server", ClientVersion: "1.0.1", ServerVersion: "1.0.1", Status: "healthy", CPUPercent: 38.2, MemoryPercent: 55.1, DiskPercent: 61.4, QueueBacklog: 12, ErrorCount: 0, Message: "现场运行正常", ReportedAt: "2026-06-19 08:30:00", ReceivedAt: "2026-06-19 08:30:02", SourceIP: "127.0.0.1"},
			{ID: 2, ReportNo: "PR202606190002", InstanceID: 2, CustomerName: "华南骨料运营中心", Watermark: "CBMP-HN-OPS", Component: "client", ClientVersion: "1.0.0", ServerVersion: "1.0.1", Status: "degraded", CPUPercent: 72.5, MemoryPercent: 68.2, DiskPercent: 84.9, QueueBacklog: 320, ErrorCount: 5, Message: "客户端版本落后且错误数升高", ReportedAt: "2026-06-19 08:10:00", ReceivedAt: "2026-06-19 08:10:03", SourceIP: "127.0.0.1", AlertRaised: true},
		},
		ProductTelemetryEvents: []ProductTelemetryEvent{
			{ID: 1, EventNo: "TE202606190001", InstanceID: 1, CustomerName: "湾区建材集团", Watermark: "CBMP-WQ-2026", Source: "apm", Component: "server", Severity: "warning", EventType: "slow_request", TraceID: "trace-wq-001", Endpoint: "/api/orders", DurationMs: 1850, StatusCode: 200, Message: "客户现场订单接口耗时升高", OccurredAt: "2026-06-19 08:32:00", ReceivedAt: "2026-06-19 08:32:02", SourceIP: "127.0.0.1", AlertRaised: true},
			{ID: 2, EventNo: "TE202606190002", InstanceID: 2, CustomerName: "华南骨料运营中心", Watermark: "CBMP-HN-OPS", Source: "log", Component: "client", Severity: "critical", EventType: "runtime_error", TraceID: "trace-hn-002", ErrorMessage: "客户端同步任务连续失败", Message: "客户端同步任务连续失败", OccurredAt: "2026-06-19 08:12:00", ReceivedAt: "2026-06-19 08:12:04", SourceIP: "127.0.0.1", AlertRaised: true},
		},
		ProductMonitoringIntegrations: []ProductMonitoringIntegration{
			{ID: 1, IntegrationNo: "MI202606190001", Name: "Prometheus 客户现场告警", Code: "prometheus-site", Provider: "prometheus", Endpoint: "mock://operations-platform-monitoring", Token: "mon-prometheus-site", Status: "active", LastEventAt: "2026-06-19 08:50:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:00:00", Remark: "接收 Alertmanager webhook"},
		},
		ProductAlertRules: []ProductAlertRule{
			{ID: 1, RuleNo: "AR202606190001", Name: "CPU 使用率过高", Source: "prometheus", Component: "server", Metric: "cpu_percent", Operator: ">=", Threshold: 85, Severity: "critical", Status: "active", NotifyChannels: []string{"sse", "webhook"}, CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:05:00", Remark: "客户现场服务端 CPU 持续高于阈值"},
			{ID: 2, RuleNo: "AR202606190002", Name: "客户端错误数升高", Source: "sentry", Component: "client", Metric: "error_count", Operator: ">=", Threshold: 5, Severity: "warning", Status: "active", NotifyChannels: []string{"sse"}, CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:06:00", Remark: "客户端异常监控"},
		},
		ProductMonitoringEvents: []ProductMonitoringEvent{
			{ID: 1, EventNo: "ME202606190001", IntegrationID: 1, IntegrationName: "Prometheus 客户现场告警", Provider: "prometheus", InstanceID: 2, CustomerName: "华南骨料运营中心", Watermark: "CBMP-HN-OPS", Source: "prometheus", Component: "server", Metric: "cpu_percent", Value: 91.2, Severity: "critical", Status: "firing", Title: "CPU 使用率过高", Message: "Alertmanager 推送 CPU 使用率 91.2%", Labels: map[string]string{"job": "cbmp-server", "site": "hn-ops"}, OccurredAt: "2026-06-19 08:50:00", ReceivedAt: "2026-06-19 08:50:01", SourceIP: "127.0.0.1", AlertRaised: true, MatchedRuleNo: "AR202606190001"},
		},
		ProductAlertPolicies: []ProductAlertPolicy{
			{ID: 1, PolicyNo: "AP202606190001", Name: "严重告警 30 分钟聚合并升级", Source: "all", Component: "all", Metric: "all", Severity: "critical", AggregateWindowMinutes: 30, SuppressMinutes: 10, EscalateAfterMinutes: 15, EscalateTo: "on_call_manager", NotifyChannels: []string{"sse", "webhook"}, Status: "active", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:00:00", Remark: "严重异常先聚合降噪，超过 15 分钟未处理自动升级"},
			{ID: 2, PolicyNo: "AP202606190002", Name: "授权风险每日提醒", Source: "license", Component: "license", Metric: "license_expire", Severity: "warning", AggregateWindowMinutes: 1440, SuppressMinutes: 1440, EscalateAfterMinutes: 0, EscalateTo: "renewal_owner", NotifyChannels: []string{"sse"}, Status: "active", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:05:00", Remark: "授权到期风险按天聚合，避免反复提醒"},
		},
		ProductAlertChannels: []ProductAlertChannel{
			{ID: 1, ChannelNo: "AC202606190001", Name: "内部告警实时事件", Code: "sse", Type: "sse", Status: "active", RetryLimit: 1, TimeoutSeconds: 1, CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:00:00", Remark: "站内实时通知"},
			{ID: 2, ChannelNo: "AC202606190002", Name: "值班 Webhook", Code: "webhook", Type: "webhook", Endpoint: "mock://success", Status: "active", RetryLimit: 3, TimeoutSeconds: 3, LastDeliveredAt: "2026-06-19 08:10:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:01:00", Remark: "可替换为企业微信/飞书/ITSM webhook"},
			{ID: 3, ChannelNo: "AC202606190003", Name: "短信备用通道", Code: "sms", Type: "sms", Status: "active", RetryLimit: 3, TimeoutSeconds: 3, LastError: "通知通道未配置 endpoint", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:02:00", Remark: "生产部署需配置短信网关 endpoint"},
			{ID: 4, ChannelNo: "AC202606190004", Name: "企业微信值班群", Code: "enterprise_wechat", Type: "enterprise_wechat", Endpoint: "mock://success", Token: "wecom-demo-token", Secret: "wecom-demo-secret", Status: "active", RetryLimit: 3, TimeoutSeconds: 3, LastDeliveredAt: "2026-06-19 08:12:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:03:00", Remark: "按企业微信机器人 markdown payload 投递，生产替换为真实机器人 endpoint"},
			{ID: 5, ChannelNo: "AC202606190005", Name: "ITSM 工单通道", Code: "itsm", Type: "itsm", Endpoint: "mock://success", Token: "itsm-demo-token", Secret: "itsm-demo-secret", Status: "active", RetryLimit: 3, TimeoutSeconds: 5, LastDeliveredAt: "2026-06-19 08:13:00", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:04:00", Remark: "按 incident payload 创建外部工单，生产替换为 ServiceNow/Jira/Zendesk endpoint"},
		},
		ProductAlertNotifications: []ProductAlertNotification{
			{ID: 1, NotificationNo: "AN202606190001", AlertID: 1, AlertNo: "AL202606190001", PolicyID: 1, PolicyNo: "AP202606190001", InstanceID: 2, CustomerName: "华南骨料运营中心", Action: "suppressed", Severity: "critical", Channel: "suppression", Target: "on_call_manager", Status: "delivered", AttemptCount: 1, Message: "客户端版本落后重复事件已抑制至 2026-06-19 08:40:00", CreatedAt: "2026-06-19 08:10:00", DeliveredAt: "2026-06-19 08:10:00"},
		},
		ProductUpdateRollouts: []ProductUpdateRollout{
			{ID: 1, RolloutNo: "UR202606190001", UpdateID: 3, Version: "1.0.1", Component: "client", Strategy: "gray", Status: "running", TotalTargets: 2, AppliedTargets: 1, CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:35:00", StartedAt: "2026-06-19 08:40:00", Remark: "客户端 1.0.1 灰度发布", Items: []ProductUpdateRolloutItem{
				{ID: 1, InstanceID: 1, CustomerName: "湾区建材集团", FromVersion: "1.0.1", ToVersion: "1.0.1", Status: "applied", Message: "已确认客户端版本", StartedAt: "2026-06-19 08:40:00", AppliedAt: "2026-06-19 08:45:00"},
				{ID: 2, InstanceID: 2, CustomerName: "华南骨料运营中心", FromVersion: "1.0.0", ToVersion: "1.0.1", Status: "pending", Message: "等待客户窗口"},
			}},
		},
		ProductUpdateExecutions: []ProductUpdateExecution{
			{ID: 1, ExecutionNo: "UE202606190001", RolloutID: 1, RolloutNo: "UR202606190001", UpdateID: 3, InstanceID: 1, CustomerName: "湾区建材集团", Component: "client", Version: "1.0.1", Action: "apply", Status: "succeeded", ArtifactFileName: "cbmp-client-update-1.0.1.json", ChecksumVerified: true, StartedBy: "平台管理员", StartedAt: "2026-06-19 08:40:00", CompletedAt: "2026-06-19 08:45:00", DurationMs: 300000, PrecheckSummary: "验签通过，客户现场版本满足升级前置条件", Result: "客户端 1.0.1 已确认应用", Steps: []ProductUpdateExecutionStep{
				{ID: 1, Name: "验签更新包", Status: "succeeded", Message: "checksum/signature verified", StartedAt: "2026-06-19 08:40:00", CompletedAt: "2026-06-19 08:40:05", DurationMs: 5000},
				{ID: 2, Name: "创建现场快照", Status: "succeeded", Message: "snapshot before client update", StartedAt: "2026-06-19 08:40:05", CompletedAt: "2026-06-19 08:41:00", DurationMs: 55000},
				{ID: 3, Name: "分发更新包", Status: "succeeded", Message: "artifact delivered", StartedAt: "2026-06-19 08:41:00", CompletedAt: "2026-06-19 08:42:00", DurationMs: 60000},
				{ID: 4, Name: "停止客户端组件", Status: "succeeded", Message: "component stopped", StartedAt: "2026-06-19 08:42:00", CompletedAt: "2026-06-19 08:42:30", DurationMs: 30000},
				{ID: 5, Name: "安装目标版本", Status: "succeeded", Message: "installed 1.0.1", StartedAt: "2026-06-19 08:42:30", CompletedAt: "2026-06-19 08:44:00", DurationMs: 90000},
				{ID: 6, Name: "健康检查", Status: "succeeded", Message: "probe healthy", StartedAt: "2026-06-19 08:44:00", CompletedAt: "2026-06-19 08:45:00", DurationMs: 60000},
			}},
		},
		ProductSystemUpdateTasks: []ProductSystemUpdateTask{
			{ID: 1, TaskNo: "SU202606190001", ExecutionID: 1, ExecutionNo: "UE202606190001", RolloutID: 1, RolloutNo: "UR202606190001", RolloutItemID: 1, UpdateID: 3, InstanceID: 1, CustomerName: "湾区建材集团", Watermark: "CBMP-WQ-2026", Component: "client", Version: "1.0.1", FromVersion: "1.0.1", Action: "apply", Status: "succeeded", Progress: 100, ArtifactFileName: "cbmp-client-update-1.0.1.json", Checksum: "sha256:client-101", Signature: "sig:client-101", DownloadURL: "/api/system/updates/3/download", UpdaterTokenHint: "probe-...oken", CreatedBy: "平台管理员", CreatedAt: "2026-06-19 08:40:00", ClaimedAt: "2026-06-19 08:40:10", StartedAt: "2026-06-19 08:40:15", CompletedAt: "2026-06-19 08:45:00", LastHeartbeatAt: "2026-06-19 08:45:00", Result: "端内更新器已完成客户端更新", Logs: []ProductSystemUpdateTaskLog{
				{ID: 1, Status: "assigned", Progress: 10, Step: "claimed", Message: "端内更新器已拉取任务", CreatedAt: "2026-06-19 08:40:10"},
				{ID: 2, Status: "succeeded", Progress: 100, Step: "health_check", Message: "客户端版本 1.0.1 健康检查通过", CreatedAt: "2026-06-19 08:45:00"},
			}},
		},
		AuditLogs: []AuditLog{{1, "system", "seed", "platform", 1, "初始化演示业务数据", "127.0.0.1", "2026-06-18 08:00:00"}},
	}

	for _, vehicle := range data.Vehicles {
		if vehicle.CertExpiresAt < "2026-12-31" {
			data.VehicleAlarms = append(data.VehicleAlarms, VehicleAlarm{
				ID:        nextID(&data, "alarm"),
				VehicleID: vehicle.ID,
				AlarmType: "vehicle_cert_expiring",
				Level:     "low",
				Message:   fmt.Sprintf("%s 证件将在 %s 到期", vehicle.PlateNo, vehicle.CertExpiresAt),
				Status:    "open",
				CreatedAt: "2026-06-18 08:10:00",
			})
		}
	}
	return data
}
