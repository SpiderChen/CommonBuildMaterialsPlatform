package appliance

type LicenseInfo struct {
	LicenseID             string   `json:"licenseId"`
	CustomerName          string   `json:"customerName"`
	Watermark             string   `json:"watermark"`
	ExpiresAt             string   `json:"expiresAt"`
	Edition               string   `json:"edition"`
	Modules               []string `json:"modules"`
	MaxSites              int      `json:"maxSites"`
	MaxVehicles           int      `json:"maxVehicles"`
	IssuedAt              string   `json:"issuedAt"`
	Issuer                string   `json:"issuer"`
	PublicKey             string   `json:"publicKey,omitempty"`
	PublicKeyFingerprint  string   `json:"publicKeyFingerprint,omitempty"`
	Signature             string   `json:"signature"`
	LastVerifiedAt        string   `json:"lastVerifiedAt"`
	LastVerificationState string   `json:"lastVerificationState"`
	LastVerificationError string   `json:"lastVerificationError"`
}

type LicensePackage struct {
	ID                    int64    `json:"id"`
	LicenseID             string   `json:"licenseId"`
	CustomerName          string   `json:"customerName"`
	Watermark             string   `json:"watermark"`
	ExpiresAt             string   `json:"expiresAt"`
	Edition               string   `json:"edition"`
	Modules               []string `json:"modules"`
	MaxSites              int      `json:"maxSites"`
	MaxVehicles           int      `json:"maxVehicles"`
	IssuedAt              string   `json:"issuedAt"`
	Issuer                string   `json:"issuer"`
	PublicKey             string   `json:"publicKey"`
	PublicKeyFingerprint  string   `json:"publicKeyFingerprint"`
	Signature             string   `json:"signature"`
	Status                string   `json:"status"`
	ActivatedAt           string   `json:"activatedAt"`
	LastVerificationState string   `json:"lastVerificationState"`
	LastVerificationError string   `json:"lastVerificationError"`
}

type LicenseIssueRecord struct {
	ID                   int64    `json:"id"`
	IssueNo              string   `json:"issueNo"`
	LicenseID            string   `json:"licenseId"`
	CustomerName         string   `json:"customerName"`
	Watermark            string   `json:"watermark"`
	ExpiresAt            string   `json:"expiresAt"`
	Edition              string   `json:"edition"`
	Modules              []string `json:"modules"`
	MaxSites             int      `json:"maxSites"`
	MaxVehicles          int      `json:"maxVehicles"`
	Issuer               string   `json:"issuer"`
	PublicKeyFingerprint string   `json:"publicKeyFingerprint"`
	Status               string   `json:"status"`
	IssuedAt             string   `json:"issuedAt"`
	Actor                string   `json:"actor"`
}

type LicenseRevocation struct {
	ID        int64  `json:"id"`
	RevokeNo  string `json:"revokeNo"`
	LicenseID string `json:"licenseId"`
	Reason    string `json:"reason"`
	Status    string `json:"status"`
	RevokedAt string `json:"revokedAt"`
	Actor     string `json:"actor"`
}

type LicensePortalOverview struct {
	KPIs            LicensePortalKPI        `json:"kpis"`
	Customers       []LicensePortalCustomer `json:"customers"`
	RecentEvents    []LicensePortalEvent    `json:"recentEvents"`
	RequiredModules []string                `json:"requiredModules"`
}

type LicensePortalKPI struct {
	TotalCustomers  int     `json:"totalCustomers"`
	TotalPackages   int     `json:"totalPackages"`
	ActivePackages  int     `json:"activePackages"`
	IssuedPackages  int     `json:"issuedPackages"`
	RevokedPackages int     `json:"revokedPackages"`
	Expiring30Days  int     `json:"expiring30Days"`
	ExpiredPackages int     `json:"expiredPackages"`
	DownloadCount   int     `json:"downloadCount"`
	ModuleCoverage  float64 `json:"moduleCoverage"`
	RiskLevel       string  `json:"riskLevel"`
}

type LicensePortalCustomer struct {
	CustomerName         string   `json:"customerName"`
	Watermark            string   `json:"watermark"`
	LicenseID            string   `json:"licenseId"`
	Edition              string   `json:"edition"`
	Status               string   `json:"status"`
	ExpiresAt            string   `json:"expiresAt"`
	DaysToExpire         int      `json:"daysToExpire"`
	Modules              []string `json:"modules"`
	ModuleCount          int      `json:"moduleCount"`
	ModuleCoverage       float64  `json:"moduleCoverage"`
	MaxSites             int      `json:"maxSites"`
	MaxVehicles          int      `json:"maxVehicles"`
	PackageCount         int      `json:"packageCount"`
	LatestPackageID      int64    `json:"latestPackageId"`
	LatestIssuedAt       string   `json:"latestIssuedAt"`
	LatestActivatedAt    string   `json:"latestActivatedAt"`
	LatestDownloadAt     string   `json:"latestDownloadAt"`
	RenewalAvailable     bool     `json:"renewalAvailable"`
	Revoked              bool     `json:"revoked"`
	VerificationState    string   `json:"verificationState"`
	VerificationError    string   `json:"verificationError"`
	RiskLevel            string   `json:"riskLevel"`
	LastOperation        string   `json:"lastOperation"`
	PublicKeyFingerprint string   `json:"publicKeyFingerprint"`
}

type LicensePortalEvent struct {
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	LicenseID string `json:"licenseId"`
	Detail    string `json:"detail"`
	Actor     string `json:"actor"`
	IP        string `json:"ip"`
	CreatedAt string `json:"createdAt"`
}

type ProductInstance struct {
	ID               int64  `json:"id"`
	CustomerName     string `json:"customerName"`
	LicenseID        string `json:"licenseId"`
	Watermark        string `json:"watermark"`
	Edition          string `json:"edition"`
	DeploymentMode   string `json:"deploymentMode"`
	ClientVersion    string `json:"clientVersion"`
	ServerVersion    string `json:"serverVersion"`
	Endpoint         string `json:"endpoint"`
	Status           string `json:"status"`
	ProbeToken       string `json:"probeToken"`
	ProbeEnabled     bool   `json:"probeEnabled"`
	HealthStatus     string `json:"healthStatus"`
	LastProbeAt      string `json:"lastProbeAt"`
	LicenseExpiresAt string `json:"licenseExpiresAt"`
	DaysToExpire     int    `json:"daysToExpire"`
	LicenseRisk      string `json:"licenseRisk"`
	RenewalAvailable bool   `json:"renewalAvailable"`
	RenewalOwner     string `json:"renewalOwner"`
	RenewalStage     string `json:"renewalStage"`
	AlertLevel       string `json:"alertLevel"`
	LastHeartbeatAt  string `json:"lastHeartbeatAt"`
	LatestPackageID  int64  `json:"latestPackageId"`
	CreatedAt        string `json:"createdAt"`
	Remark           string `json:"remark"`
}

type SystemAlert struct {
	ID              int64  `json:"id"`
	AlertNo         string `json:"alertNo"`
	InstanceID      int64  `json:"instanceId"`
	CustomerName    string `json:"customerName"`
	Severity        string `json:"severity"`
	Source          string `json:"source"`
	Title           string `json:"title"`
	Message         string `json:"message"`
	Status          string `json:"status"`
	GroupKey        string `json:"groupKey"`
	PolicyNo        string `json:"policyNo"`
	EventCount      int    `json:"eventCount"`
	SuppressedUntil string `json:"suppressedUntil"`
	EscalationLevel string `json:"escalationLevel"`
	EscalatedAt     string `json:"escalatedAt"`
	FirstSeenAt     string `json:"firstSeenAt"`
	LastSeenAt      string `json:"lastSeenAt"`
	HandledBy       string `json:"handledBy"`
	HandledAt       string `json:"handledAt"`
}

type ProductRenewalTask struct {
	ID            int64   `json:"id"`
	TaskNo        string  `json:"taskNo"`
	InstanceID    int64   `json:"instanceId"`
	CustomerName  string  `json:"customerName"`
	LicenseID     string  `json:"licenseId"`
	Stage         string  `json:"stage"`
	Status        string  `json:"status"`
	Owner         string  `json:"owner"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	DueDate       string  `json:"dueDate"`
	NextFollowAt  string  `json:"nextFollowAt"`
	RiskLevel     string  `json:"riskLevel"`
	LastContactAt string  `json:"lastContactAt"`
	ClosedAt      string  `json:"closedAt"`
	CreatedAt     string  `json:"createdAt"`
	Remark        string  `json:"remark"`
}

type ProductRenewalQuote struct {
	ID           int64    `json:"id"`
	QuoteNo      string   `json:"quoteNo"`
	TaskID       int64    `json:"taskId"`
	InstanceID   int64    `json:"instanceId"`
	CustomerName string   `json:"customerName"`
	LicenseID    string   `json:"licenseId"`
	Amount       float64  `json:"amount"`
	Currency     string   `json:"currency"`
	Modules      []string `json:"modules"`
	NewExpiresAt string   `json:"newExpiresAt"`
	Status       string   `json:"status"`
	PreparedBy   string   `json:"preparedBy"`
	PreparedAt   string   `json:"preparedAt"`
	ApprovedBy   string   `json:"approvedBy"`
	ApprovedAt   string   `json:"approvedAt"`
	Remark       string   `json:"remark"`
}

type ProductRenewalContract struct {
	ID           int64   `json:"id"`
	ContractNo   string  `json:"contractNo"`
	TaskID       int64   `json:"taskId"`
	QuoteID      int64   `json:"quoteId"`
	InstanceID   int64   `json:"instanceId"`
	CustomerName string  `json:"customerName"`
	LicenseID    string  `json:"licenseId"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
	SignedBy     string  `json:"signedBy"`
	SignedAt     string  `json:"signedAt"`
	CreatedBy    string  `json:"createdBy"`
	CreatedAt    string  `json:"createdAt"`
	Remark       string  `json:"remark"`
}

type ProductRenewalPayment struct {
	ID           int64   `json:"id"`
	PaymentNo    string  `json:"paymentNo"`
	TaskID       int64   `json:"taskId"`
	ContractID   int64   `json:"contractId"`
	InstanceID   int64   `json:"instanceId"`
	CustomerName string  `json:"customerName"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Method       string  `json:"method"`
	Status       string  `json:"status"`
	PaidAt       string  `json:"paidAt"`
	CreatedBy    string  `json:"createdBy"`
	CreatedAt    string  `json:"createdAt"`
	Remark       string  `json:"remark"`
}

type ProductRenewalApproval struct {
	ID           int64   `json:"id"`
	ApprovalNo   string  `json:"approvalNo"`
	TaskID       int64   `json:"taskId"`
	QuoteID      int64   `json:"quoteId"`
	ContractID   int64   `json:"contractId"`
	InstanceID   int64   `json:"instanceId"`
	CustomerName string  `json:"customerName"`
	LicenseID    string  `json:"licenseId"`
	ApprovalType string  `json:"approvalType"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
	CurrentRole  string  `json:"currentRole"`
	RequestedBy  string  `json:"requestedBy"`
	RequestedAt  string  `json:"requestedAt"`
	ApprovedBy   string  `json:"approvedBy"`
	ApprovedAt   string  `json:"approvedAt"`
	Comment      string  `json:"comment"`
}

type ProductRenewalInvoice struct {
	ID              int64   `json:"id"`
	InvoiceNo       string  `json:"invoiceNo"`
	TaskID          int64   `json:"taskId"`
	ContractID      int64   `json:"contractId"`
	PaymentID       int64   `json:"paymentId"`
	InstanceID      int64   `json:"instanceId"`
	CustomerName    string  `json:"customerName"`
	LicenseID       string  `json:"licenseId"`
	Amount          float64 `json:"amount"`
	TaxRate         float64 `json:"taxRate"`
	TaxAmount       float64 `json:"taxAmount"`
	InvoiceType     string  `json:"invoiceType"`
	Status          string  `json:"status"`
	TaxStatus       string  `json:"taxStatus"`
	FileURL         string  `json:"fileUrl"`
	CreatedBy       string  `json:"createdBy"`
	CreatedAt       string  `json:"createdAt"`
	IssuedAt        string  `json:"issuedAt"`
	ExternalRequest string  `json:"externalRequest"`
	Remark          string  `json:"remark"`
}

type ProductRenewalESign struct {
	ID           int64  `json:"id"`
	SignNo       string `json:"signNo"`
	TaskID       int64  `json:"taskId"`
	ContractID   int64  `json:"contractId"`
	InstanceID   int64  `json:"instanceId"`
	CustomerName string `json:"customerName"`
	LicenseID    string `json:"licenseId"`
	Signer       string `json:"signer"`
	Phone        string `json:"phone"`
	Channel      string `json:"channel"`
	Status       string `json:"status"`
	LinkURL      string `json:"linkUrl"`
	SentBy       string `json:"sentBy"`
	SentAt       string `json:"sentAt"`
	SignedAt     string `json:"signedAt"`
	Signature    string `json:"signature"`
	Remark       string `json:"remark"`
}

type ProductRenewalIntegration struct {
	ID             int64  `json:"id"`
	IntegrationNo  string `json:"integrationNo"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Provider       string `json:"provider"`
	Scenario       string `json:"scenario"`
	Endpoint       string `json:"endpoint"`
	Token          string `json:"token"`
	Secret         string `json:"secret"`
	Status         string `json:"status"`
	RetryLimit     int    `json:"retryLimit"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	LastSyncAt     string `json:"lastSyncAt"`
	LastError      string `json:"lastError"`
	CreatedBy      string `json:"createdBy"`
	CreatedAt      string `json:"createdAt"`
	Remark         string `json:"remark"`
}

type ProductRenewalSyncRecord struct {
	ID                int64  `json:"id"`
	SyncNo            string `json:"syncNo"`
	IntegrationID     int64  `json:"integrationId"`
	IntegrationNo     string `json:"integrationNo"`
	IntegrationCode   string `json:"integrationCode"`
	Provider          string `json:"provider"`
	Scenario          string `json:"scenario"`
	ResourceType      string `json:"resourceType"`
	ResourceID        int64  `json:"resourceId"`
	ResourceNo        string `json:"resourceNo"`
	TaskID            int64  `json:"taskId"`
	CustomerName      string `json:"customerName"`
	Action            string `json:"action"`
	Status            string `json:"status"`
	AttemptCount      int    `json:"attemptCount"`
	NextRetryAt       string `json:"nextRetryAt"`
	ExternalRequestID string `json:"externalRequestId"`
	ExternalStatus    string `json:"externalStatus"`
	RequestPayload    string `json:"requestPayload"`
	ResponsePayload   string `json:"responsePayload"`
	Error             string `json:"error"`
	CreatedAt         string `json:"createdAt"`
	LastAttemptAt     string `json:"lastAttemptAt"`
	CompletedAt       string `json:"completedAt"`
}

type ProductProbeReport struct {
	ID            int64   `json:"id"`
	ReportNo      string  `json:"reportNo"`
	InstanceID    int64   `json:"instanceId"`
	CustomerName  string  `json:"customerName"`
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
	ReceivedAt    string  `json:"receivedAt"`
	SourceIP      string  `json:"sourceIp"`
	AlertRaised   bool    `json:"alertRaised"`
}

type ProductTelemetryEvent struct {
	ID           int64  `json:"id"`
	EventNo      string `json:"eventNo"`
	InstanceID   int64  `json:"instanceId"`
	CustomerName string `json:"customerName"`
	Watermark    string `json:"watermark"`
	Source       string `json:"source"`
	Component    string `json:"component"`
	Severity     string `json:"severity"`
	EventType    string `json:"eventType"`
	TraceID      string `json:"traceId"`
	SpanID       string `json:"spanId"`
	Endpoint     string `json:"endpoint"`
	DurationMs   int    `json:"durationMs"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
	Message      string `json:"message"`
	OccurredAt   string `json:"occurredAt"`
	ReceivedAt   string `json:"receivedAt"`
	SourceIP     string `json:"sourceIp"`
	AlertRaised  bool   `json:"alertRaised"`
}

type ProductMonitoringIntegration struct {
	ID            int64  `json:"id"`
	IntegrationNo string `json:"integrationNo"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Provider      string `json:"provider"`
	Endpoint      string `json:"endpoint"`
	Token         string `json:"token"`
	Status        string `json:"status"`
	LastEventAt   string `json:"lastEventAt"`
	CreatedBy     string `json:"createdBy"`
	CreatedAt     string `json:"createdAt"`
	Remark        string `json:"remark"`
}

type ProductAlertRule struct {
	ID             int64    `json:"id"`
	RuleNo         string   `json:"ruleNo"`
	Name           string   `json:"name"`
	Source         string   `json:"source"`
	Component      string   `json:"component"`
	Metric         string   `json:"metric"`
	Operator       string   `json:"operator"`
	Threshold      float64  `json:"threshold"`
	Severity       string   `json:"severity"`
	Status         string   `json:"status"`
	NotifyChannels []string `json:"notifyChannels"`
	CreatedBy      string   `json:"createdBy"`
	CreatedAt      string   `json:"createdAt"`
	Remark         string   `json:"remark"`
}

type ProductAlertPolicy struct {
	ID                     int64    `json:"id"`
	PolicyNo               string   `json:"policyNo"`
	Name                   string   `json:"name"`
	Source                 string   `json:"source"`
	Component              string   `json:"component"`
	Metric                 string   `json:"metric"`
	Severity               string   `json:"severity"`
	AggregateWindowMinutes int      `json:"aggregateWindowMinutes"`
	SuppressMinutes        int      `json:"suppressMinutes"`
	EscalateAfterMinutes   int      `json:"escalateAfterMinutes"`
	EscalateTo             string   `json:"escalateTo"`
	NotifyChannels         []string `json:"notifyChannels"`
	Status                 string   `json:"status"`
	CreatedBy              string   `json:"createdBy"`
	CreatedAt              string   `json:"createdAt"`
	Remark                 string   `json:"remark"`
}

type ProductAlertChannel struct {
	ID              int64  `json:"id"`
	ChannelNo       string `json:"channelNo"`
	Name            string `json:"name"`
	Code            string `json:"code"`
	Type            string `json:"type"`
	Endpoint        string `json:"endpoint"`
	Token           string `json:"token"`
	Secret          string `json:"secret"`
	Status          string `json:"status"`
	RetryLimit      int    `json:"retryLimit"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
	LastDeliveredAt string `json:"lastDeliveredAt"`
	LastError       string `json:"lastError"`
	CreatedBy       string `json:"createdBy"`
	CreatedAt       string `json:"createdAt"`
	Remark          string `json:"remark"`
}

type ProductAlertNotification struct {
	ID             int64  `json:"id"`
	NotificationNo string `json:"notificationNo"`
	AlertID        int64  `json:"alertId"`
	AlertNo        string `json:"alertNo"`
	PolicyID       int64  `json:"policyId"`
	PolicyNo       string `json:"policyNo"`
	InstanceID     int64  `json:"instanceId"`
	CustomerName   string `json:"customerName"`
	Action         string `json:"action"`
	Severity       string `json:"severity"`
	ChannelID      int64  `json:"channelId"`
	ChannelNo      string `json:"channelNo"`
	Channel        string `json:"channel"`
	Target         string `json:"target"`
	Endpoint       string `json:"endpoint"`
	Status         string `json:"status"`
	AttemptCount   int    `json:"attemptCount"`
	NextRetryAt    string `json:"nextRetryAt"`
	Message        string `json:"message"`
	Error          string `json:"error"`
	CreatedAt      string `json:"createdAt"`
	DeliveredAt    string `json:"deliveredAt"`
}

type ProductMonitoringEvent struct {
	ID              int64             `json:"id"`
	EventNo         string            `json:"eventNo"`
	IntegrationID   int64             `json:"integrationId"`
	IntegrationName string            `json:"integrationName"`
	Provider        string            `json:"provider"`
	InstanceID      int64             `json:"instanceId"`
	CustomerName    string            `json:"customerName"`
	Watermark       string            `json:"watermark"`
	Source          string            `json:"source"`
	Component       string            `json:"component"`
	Metric          string            `json:"metric"`
	Value           float64           `json:"value"`
	Severity        string            `json:"severity"`
	Status          string            `json:"status"`
	Title           string            `json:"title"`
	Message         string            `json:"message"`
	Labels          map[string]string `json:"labels"`
	OccurredAt      string            `json:"occurredAt"`
	ReceivedAt      string            `json:"receivedAt"`
	SourceIP        string            `json:"sourceIp"`
	AlertRaised     bool              `json:"alertRaised"`
	MatchedRuleNo   string            `json:"matchedRuleNo"`
}

type Module struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Area        string `json:"area"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	HotPlug     bool   `json:"hotPlug"`
	Version     string `json:"version"`
}

type Plugin struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Status      string              `json:"status"`
	Version     string              `json:"version"`
	Checksum    string              `json:"checksum"`
	Signature   string              `json:"signature"`
	Permissions []string            `json:"permissions"`
	Runtime     string              `json:"runtime"`
	Entrypoint  string              `json:"entrypoint"`
	Sandbox     PluginSandboxPolicy `json:"sandbox"`
	LastRunAt   string              `json:"lastRunAt"`
}

type PluginSandboxPolicy struct {
	Runtime     string `json:"runtime"`
	TimeoutMs   int    `json:"timeoutMs"`
	Network     bool   `json:"network"`
	Filesystem  string `json:"filesystem"`
	MaxMemoryMB int    `json:"maxMemoryMb"`
}

type PluginRun struct {
	ID          int64  `json:"id"`
	RunNo       string `json:"runNo"`
	PluginID    string `json:"pluginId"`
	PluginName  string `json:"pluginName"`
	Runtime     string `json:"runtime"`
	Action      string `json:"action"`
	Permission  string `json:"permission"`
	Status      string `json:"status"`
	Input       string `json:"input"`
	Output      string `json:"output"`
	Error       string `json:"error"`
	Actor       string `json:"actor"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
	DurationMs  int64  `json:"durationMs"`
}

type Role struct {
	ID          int64    `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	DataScope   string   `json:"dataScope"`
}

type User struct {
	ID              int64  `json:"id"`
	TenantID        int64  `json:"tenantId,omitempty"`
	CompanyID       int64  `json:"companyId"`
	SiteID          int64  `json:"siteId"`
	CustomerID      int64  `json:"customerId"`
	DriverID        int64  `json:"driverId"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	RoleCode        string `json:"roleCode"`
	PasswordHash    string `json:"passwordHash,omitempty"`
	PasswordSalt    string `json:"passwordSalt,omitempty"`
	Status          string `json:"status"`
	MFAEnabled      bool   `json:"mfaEnabled"`
	MFASecret       string `json:"mfaSecret,omitempty"`
	MFALastUsedStep int64  `json:"mfaLastUsedStep,omitempty"`
}

type OIDCProvider struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	Code             string   `json:"code"`
	Issuer           string   `json:"issuer"`
	ClientID         string   `json:"clientId"`
	ClientSecret     string   `json:"clientSecret,omitempty"`
	AuthURL          string   `json:"authUrl"`
	TokenURL         string   `json:"tokenUrl"`
	UserInfoURL      string   `json:"userInfoUrl"`
	JWKSURL          string   `json:"jwksUrl"`
	RedirectURI      string   `json:"redirectUri"`
	Scopes           []string `json:"scopes"`
	UsernameClaim    string   `json:"usernameClaim"`
	DisplayNameClaim string   `json:"displayNameClaim"`
	RoleCode         string   `json:"roleCode"`
	TenantID         int64    `json:"tenantId,omitempty"`
	CompanyID        int64    `json:"companyId"`
	SiteID           int64    `json:"siteId"`
	AutoProvision    bool     `json:"autoProvision"`
	Status           string   `json:"status"`
	LastLoginAt      string   `json:"lastLoginAt"`
}

type SCIMProvider struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Code            string `json:"code"`
	BearerToken     string `json:"bearerToken,omitempty"`
	TenantID        int64  `json:"tenantId,omitempty"`
	CompanyID       int64  `json:"companyId"`
	SiteID          int64  `json:"siteId"`
	DefaultRoleCode string `json:"defaultRoleCode"`
	Status          string `json:"status"`
	LastSyncAt      string `json:"lastSyncAt"`
	CreatedAt       string `json:"createdAt"`
}

type SCIMProvisioningEvent struct {
	ID           int64  `json:"id"`
	EventNo      string `json:"eventNo"`
	ProviderID   int64  `json:"providerId"`
	ProviderCode string `json:"providerCode"`
	UserID       int64  `json:"userId"`
	Username     string `json:"username"`
	Action       string `json:"action"`
	Status       string `json:"status"`
	Detail       string `json:"detail"`
	CreatedAt    string `json:"createdAt"`
	Actor        string `json:"actor"`
	IP           string `json:"ip"`
}

type Tenant struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status string `json:"status"`
}

type Company struct {
	ID       int64  `json:"id"`
	TenantID int64  `json:"tenantId,omitempty"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Status   string `json:"status"`
}

type Site struct {
	ID        int64   `json:"id"`
	CompanyID int64   `json:"companyId"`
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	Address   string  `json:"address"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Status    string  `json:"status"`
}

type Department struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"companyId"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	ParentID  int64  `json:"parentId"`
	Status    string `json:"status"`
}

type Plant struct {
	ID        int64  `json:"id"`
	SiteID    int64  `json:"siteId"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Capacity  string `json:"capacity"`
	Interface string `json:"interface"`
	Status    string `json:"status"`
}

type Warehouse struct {
	ID     int64  `json:"id"`
	SiteID int64  `json:"siteId"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type Silo struct {
	ID          int64   `json:"id"`
	WarehouseID int64   `json:"warehouseId"`
	MaterialID  int64   `json:"materialId"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Capacity    float64 `json:"capacity"`
	CurrentQty  float64 `json:"currentQty"`
	Status      string  `json:"status"`
}

type Customer struct {
	ID          int64   `json:"id"`
	CompanyID   int64   `json:"companyId"`
	Name        string  `json:"name"`
	Contact     string  `json:"contact"`
	Phone       string  `json:"phone"`
	CreditLimit float64 `json:"creditLimit"`
	Receivable  float64 `json:"receivable"`
	PaymentTerm int     `json:"paymentTerm"`
	Status      string  `json:"status"`
}

type CustomerContact struct {
	ID         int64  `json:"id"`
	CustomerID int64  `json:"customerId"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	IsDefault  bool   `json:"isDefault"`
	Status     string `json:"status"`
}

type CustomerBlacklist struct {
	ID            int64  `json:"id"`
	CustomerID    int64  `json:"customerId"`
	CustomerName  string `json:"customerName"`
	Reason        string `json:"reason"`
	Scope         string `json:"scope"`
	Severity      string `json:"severity"`
	BlockOrders   bool   `json:"blockOrders"`
	BlockDispatch bool   `json:"blockDispatch"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
	ReleasedAt    string `json:"releasedAt"`
	Actor         string `json:"actor"`
}

type CustomerProfile struct {
	ID           int64    `json:"id"`
	CustomerID   int64    `json:"customerId"`
	CustomerName string   `json:"customerName"`
	Grade        string   `json:"grade"`
	RiskLevel    string   `json:"riskLevel"`
	CreditScore  int      `json:"creditScore"`
	Tags         []string `json:"tags"`
	Status       string   `json:"status"`
	UpdatedAt    string   `json:"updatedAt"`
	Actor        string   `json:"actor"`
}

type CustomerComplaint struct {
	ID           int64  `json:"id"`
	ComplaintNo  string `json:"complaintNo"`
	CustomerID   int64  `json:"customerId"`
	ProjectID    int64  `json:"projectId"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Level        string `json:"level"`
	Status       string `json:"status"`
	Owner        string `json:"owner"`
	SLAHours     int    `json:"slaHours"`
	DueAt        string `json:"dueAt"`
	SLAStatus    string `json:"slaStatus"`
	OverdueHours int    `json:"overdueHours"`
	CreatedAt    string `json:"createdAt"`
	ClosedAt     string `json:"closedAt"`
	Resolution   string `json:"resolution"`
}

type PricePolicy struct {
	ID             int64   `json:"id"`
	CustomerID     int64   `json:"customerId"`
	ProjectID      int64   `json:"projectId"`
	ProductID      int64   `json:"productId"`
	CustomerGrade  string  `json:"customerGrade"`
	Region         string  `json:"region"`
	MinQuantity    float64 `json:"minQuantity"`
	MaxQuantity    float64 `json:"maxQuantity"`
	FloorPrice     float64 `json:"floorPrice"`
	SalePrice      float64 `json:"salePrice"`
	PromotionName  string  `json:"promotionName"`
	PromotionType  string  `json:"promotionType"`
	PromotionValue float64 `json:"promotionValue"`
	Priority       int     `json:"priority"`
	TaxRateID      int64   `json:"taxRateId"`
	EffectiveFrom  string  `json:"effectiveFrom"`
	EffectiveTo    string  `json:"effectiveTo"`
	Status         string  `json:"status"`
}

type TaxRate struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Rate   float64 `json:"rate"`
	Scope  string  `json:"scope"`
	Status string  `json:"status"`
}

type Supplier struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Status  string `json:"status"`
}

type Carrier struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	Phone      string `json:"phone"`
	SettleMode string `json:"settleMode"`
	Status     string `json:"status"`
}

type Project struct {
	ID         int64   `json:"id"`
	CustomerID int64   `json:"customerId"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	Contact    string  `json:"contact"`
	Phone      string  `json:"phone"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	Status     string  `json:"status"`
}

type Product struct {
	ID          int64   `json:"id"`
	Line        string  `json:"line"`
	Name        string  `json:"name"`
	Spec        string  `json:"spec"`
	Unit        string  `json:"unit"`
	BasePrice   float64 `json:"basePrice"`
	CostPrice   float64 `json:"costPrice"`
	RequiresMix bool    `json:"requiresMix"`
	Status      string  `json:"status"`
}

type Material struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Spec      string  `json:"spec"`
	Unit      string  `json:"unit"`
	SafeStock float64 `json:"safeStock"`
	Status    string  `json:"status"`
}

type Vehicle struct {
	ID             int64  `json:"id"`
	PlateNo        string `json:"plateNo"`
	VehicleType    string `json:"vehicleType"`
	Capacity       string `json:"capacity"`
	Carrier        string `json:"carrier"`
	SiteID         int64  `json:"siteId"`
	DriverID       int64  `json:"driverId"`
	OnlineStatus   string `json:"onlineStatus"`
	BusinessStatus string `json:"businessStatus"`
	CertExpiresAt  string `json:"certExpiresAt"`
	Status         string `json:"status"`
}

type Driver struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	LicenseNo     string `json:"licenseNo"`
	LicenseExpire string `json:"licenseExpire"`
	Status        string `json:"status"`
}

type Contract struct {
	ID           int64          `json:"id"`
	CustomerID   int64          `json:"customerId"`
	ProjectID    int64          `json:"projectId"`
	ParentID     int64          `json:"parentId"`
	ContractNo   string         `json:"contractNo"`
	Version      int            `json:"version"`
	Name         string         `json:"name"`
	ValidFrom    string         `json:"validFrom"`
	ValidTo      string         `json:"validTo"`
	CreditPolicy string         `json:"creditPolicy"`
	TotalAmount  float64        `json:"totalAmount"`
	UsedAmount   float64        `json:"usedAmount"`
	ChangeReason string         `json:"changeReason"`
	SubmittedAt  string         `json:"submittedAt"`
	ApprovedAt   string         `json:"approvedAt"`
	ApprovedBy   string         `json:"approvedBy"`
	Status       string         `json:"status"`
	Items        []ContractItem `json:"items"`
}

type ContractAttachment struct {
	ID         int64  `json:"id"`
	ContractID int64  `json:"contractId"`
	CustomerID int64  `json:"customerId"`
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	URL        string `json:"url"`
	Checksum   string `json:"checksum"`
	Status     string `json:"status"`
	UploadedBy string `json:"uploadedBy"`
	UploadedAt string `json:"uploadedAt"`
}

type ContractItem struct {
	ProductID int64   `json:"productId"`
	Unit      string  `json:"unit"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type SalesOrderLine struct {
	ID            int64   `json:"id"`
	Seq           int     `json:"seq"`
	ProductID     int64   `json:"productId"`
	ProductLine   string  `json:"productLine"`
	ProductName   string  `json:"productName"`
	StrengthGrade string  `json:"strengthGrade"`
	Slump         string  `json:"slump"`
	PouringPart   string  `json:"pouringPart"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	UnitPrice     float64 `json:"unitPrice"`
	FloorPrice    float64 `json:"floorPrice"`
	TaxRate       float64 `json:"taxRate"`
	Amount        float64 `json:"amount"`
	PriceSource   string  `json:"priceSource"`
	RiskFlag      string  `json:"riskFlag"`
}

type SalesOrder struct {
	ID             int64            `json:"id"`
	OrderNo        string           `json:"orderNo"`
	CustomerID     int64            `json:"customerId"`
	ProjectID      int64            `json:"projectId"`
	ProductID      int64            `json:"productId"`
	SiteID         int64            `json:"siteId"`
	ProductLine    string           `json:"productLine"`
	PlanQuantity   float64          `json:"planQuantity"`
	Unit           string           `json:"unit"`
	UnitPrice      float64          `json:"unitPrice"`
	TotalAmount    float64          `json:"totalAmount"`
	Lines          []SalesOrderLine `json:"lines,omitempty"`
	PlanTime       string           `json:"planTime"`
	ReceiveAddress string           `json:"receiveAddress"`
	Contact        string           `json:"contact"`
	Phone          string           `json:"phone"`
	SettlementMode string           `json:"settlementMode"`
	TransportMode  string           `json:"transportMode"`
	StrengthGrade  string           `json:"strengthGrade"`
	Slump          string           `json:"slump"`
	PouringPart    string           `json:"pouringPart"`
	PumpMode       string           `json:"pumpMode"`
	DispatchedQty  float64          `json:"dispatchedQty"`
	SignedQty      float64          `json:"signedQty"`
	Status         string           `json:"status"`
	RiskFlag       string           `json:"riskFlag"`
	CreatedAt      string           `json:"createdAt"`
}

type ProductionPlan struct {
	ID              int64   `json:"id"`
	PlanNo          string  `json:"planNo"`
	OrderID         int64   `json:"orderId"`
	SiteID          int64   `json:"siteId"`
	ProductID       int64   `json:"productId"`
	PlanQuantity    float64 `json:"planQuantity"`
	ProducedQty     float64 `json:"producedQty"`
	PlanDate        string  `json:"planDate"`
	Shift           string  `json:"shift"`
	CapacityStatus  string  `json:"capacityStatus"`
	InventoryStatus string  `json:"inventoryStatus"`
	RecipeStatus    string  `json:"recipeStatus"`
	Status          string  `json:"status"`
}

type MixDesign struct {
	ID            int64               `json:"id"`
	ProductID     int64               `json:"productId"`
	SiteID        int64               `json:"siteId"`
	ParentID      int64               `json:"parentId"`
	Code          string              `json:"code"`
	Version       string              `json:"version"`
	StrengthGrade string              `json:"strengthGrade"`
	Slump         string              `json:"slump"`
	Scope         string              `json:"scope"`
	Status        string              `json:"status"`
	IsCurrent     bool                `json:"isCurrent"`
	EffectiveFrom string              `json:"effectiveFrom"`
	EffectiveTo   string              `json:"effectiveTo"`
	ApprovedBy    string              `json:"approvedBy"`
	ApprovedAt    string              `json:"approvedAt"`
	RetiredAt     string              `json:"retiredAt"`
	CreatedBy     string              `json:"createdBy"`
	CreatedAt     string              `json:"createdAt"`
	UpdatedAt     string              `json:"updatedAt"`
	Materials     []MixDesignMaterial `json:"materials"`
}

type MixDesignMaterial struct {
	MaterialID int64   `json:"materialId"`
	Dosage     float64 `json:"dosage"`
	Unit       string  `json:"unit"`
}

type MixDesignTrialRun struct {
	ID             int64   `json:"id"`
	TrialNo        string  `json:"trialNo"`
	MixDesignID    int64   `json:"mixDesignId"`
	ProductID      int64   `json:"productId"`
	SiteID         int64   `json:"siteId"`
	TargetStrength string  `json:"targetStrength"`
	Slump          string  `json:"slump"`
	Water          float64 `json:"water"`
	SandRate       float64 `json:"sandRate"`
	AdmixtureRate  float64 `json:"admixtureRate"`
	Strength7d     float64 `json:"strength7d"`
	Strength28d    float64 `json:"strength28d"`
	Result         string  `json:"result"`
	Conclusion     string  `json:"conclusion"`
	Tester         string  `json:"tester"`
	TestedAt       string  `json:"testedAt"`
	CreatedAt      string  `json:"createdAt"`
	Remark         string  `json:"remark"`
}

type LaboratorySample struct {
	ID              int64  `json:"id"`
	SampleNo        string `json:"sampleNo"`
	SourceType      string `json:"sourceType"`
	SourceID        int64  `json:"sourceId"`
	SiteID          int64  `json:"siteId"`
	ProductID       int64  `json:"productId"`
	MaterialID      int64  `json:"materialId"`
	MixDesignID     int64  `json:"mixDesignId"`
	BatchID         int64  `json:"batchId"`
	InspectionID    int64  `json:"inspectionId"`
	RawInspectionID int64  `json:"rawInspectionId"`
	SampleType      string `json:"sampleType"`
	Status          string `json:"status"`
	Result          string `json:"result"`
	PlannedTestAt   string `json:"plannedTestAt"`
	CollectedAt     string `json:"collectedAt"`
	CreatedBy       string `json:"createdBy"`
	Remark          string `json:"remark"`
}

type LaboratoryTestRecord struct {
	ID          int64   `json:"id"`
	TestNo      string  `json:"testNo"`
	SampleID    int64   `json:"sampleId"`
	EquipmentID int64   `json:"equipmentId"`
	SiteID      int64   `json:"siteId"`
	TestType    string  `json:"testType"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Result      string  `json:"result"`
	Status      string  `json:"status"`
	Tester      string  `json:"tester"`
	TestedAt    string  `json:"testedAt"`
	Reviewer    string  `json:"reviewer"`
	ReviewedAt  string  `json:"reviewedAt"`
	Remark      string  `json:"remark"`
}

type LaboratoryEquipment struct {
	ID                   int64  `json:"id"`
	EquipmentNo          string `json:"equipmentNo"`
	Name                 string `json:"name"`
	SiteID               int64  `json:"siteId"`
	Model                string `json:"model"`
	SerialNo             string `json:"serialNo"`
	Status               string `json:"status"`
	CalibrationCycleDays int    `json:"calibrationCycleDays"`
	LastCalibrationAt    string `json:"lastCalibrationAt"`
	NextCalibrationAt    string `json:"nextCalibrationAt"`
	CreatedAt            string `json:"createdAt"`
	Remark               string `json:"remark"`
}

type LaboratoryCalibration struct {
	ID            int64  `json:"id"`
	CalibrationNo string `json:"calibrationNo"`
	EquipmentID   int64  `json:"equipmentId"`
	SiteID        int64  `json:"siteId"`
	Result        string `json:"result"`
	CalibratedAt  string `json:"calibratedAt"`
	NextDueAt     string `json:"nextDueAt"`
	CertificateNo string `json:"certificateNo"`
	Agency        string `json:"agency"`
	Operator      string `json:"operator"`
	Remark        string `json:"remark"`
}

type QualityException struct {
	ID               int64  `json:"id"`
	ExceptionNo      string `json:"exceptionNo"`
	SourceType       string `json:"sourceType"`
	SourceID         int64  `json:"sourceId"`
	SiteID           int64  `json:"siteId"`
	Severity         string `json:"severity"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Status           string `json:"status"`
	Responsible      string `json:"responsible"`
	RootCause        string `json:"rootCause"`
	CorrectiveAction string `json:"correctiveAction"`
	CreatedAt        string `json:"createdAt"`
	HandledAt        string `json:"handledAt"`
	ClosedBy         string `json:"closedBy"`
}

type LaboratoryKPI struct {
	MixDesigns         int     `json:"mixDesigns"`
	CurrentMixDesigns  int     `json:"currentMixDesigns"`
	PendingMixDesigns  int     `json:"pendingMixDesigns"`
	TrialRuns          int     `json:"trialRuns"`
	Samples            int     `json:"samples"`
	PendingSamples     int     `json:"pendingSamples"`
	Tests              int     `json:"tests"`
	PendingReviews     int     `json:"pendingReviews"`
	Equipments         int     `json:"equipments"`
	CalibrationDue     int     `json:"calibrationDue"`
	CalibrationOverdue int     `json:"calibrationOverdue"`
	OpenExceptions     int     `json:"openExceptions"`
	PassRate           float64 `json:"passRate"`
}

type ProductionTask struct {
	ID          int64   `json:"id"`
	TaskNo      string  `json:"taskNo"`
	PlanID      int64   `json:"planId"`
	OrderID     int64   `json:"orderId"`
	SiteID      int64   `json:"siteId"`
	ProductID   int64   `json:"productId"`
	MixDesignID int64   `json:"mixDesignId"`
	PlanQty     float64 `json:"planQty"`
	ProducedQty float64 `json:"producedQty"`
	Status      string  `json:"status"`
	StartedAt   string  `json:"startedAt"`
	CompletedAt string  `json:"completedAt"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type ProductionBatch struct {
	ID            int64   `json:"id"`
	BatchNo       string  `json:"batchNo"`
	TaskID        int64   `json:"taskId"`
	PlanID        int64   `json:"planId"`
	OrderID       int64   `json:"orderId"`
	SiteID        int64   `json:"siteId"`
	ProductID     int64   `json:"productId"`
	MixDesignID   int64   `json:"mixDesignId"`
	Quantity      float64 `json:"quantity"`
	PlantCode     string  `json:"plantCode"`
	Operator      string  `json:"operator"`
	QualityStatus string  `json:"qualityStatus"`
	Status        string  `json:"status"`
	StartedAt     string  `json:"startedAt"`
	CompletedAt   string  `json:"completedAt"`
}

type ProductionDailyReport struct {
	ID             int64   `json:"id"`
	ReportNo       string  `json:"reportNo"`
	SiteID         int64   `json:"siteId"`
	ReportDate     string  `json:"reportDate"`
	PlannedQty     float64 `json:"plannedQty"`
	ProducedQty    float64 `json:"producedQty"`
	BatchCount     int     `json:"batchCount"`
	MaterialCost   float64 `json:"materialCost"`
	QualityPassed  int     `json:"qualityPassed"`
	QualityPending int     `json:"qualityPending"`
	Status         string  `json:"status"`
	GeneratedAt    string  `json:"generatedAt"`
}

type QualityInspection struct {
	ID           int64   `json:"id"`
	InspectionNo string  `json:"inspectionNo"`
	BatchID      int64   `json:"batchId"`
	BatchNo      string  `json:"batchNo"`
	SiteID       int64   `json:"siteId"`
	ProductID    int64   `json:"productId"`
	MixDesignID  int64   `json:"mixDesignId"`
	Inspector    string  `json:"inspector"`
	Slump        string  `json:"slump"`
	Temperature  float64 `json:"temperature"`
	SampleCount  int     `json:"sampleCount"`
	Result       string  `json:"result"`
	Status       string  `json:"status"`
	Remark       string  `json:"remark"`
	CreatedAt    string  `json:"createdAt"`
	CompletedAt  string  `json:"completedAt"`
}

type QualitySample struct {
	ID            int64   `json:"id"`
	SampleNo      string  `json:"sampleNo"`
	InspectionID  int64   `json:"inspectionId"`
	BatchID       int64   `json:"batchId"`
	SampleType    string  `json:"sampleType"`
	AgeDays       int     `json:"ageDays"`
	PlannedTestAt string  `json:"plannedTestAt"`
	TestedAt      string  `json:"testedAt"`
	Strength      float64 `json:"strength"`
	Result        string  `json:"result"`
	Status        string  `json:"status"`
	Remark        string  `json:"remark"`
}

type RawMaterialInspection struct {
	ID           int64   `json:"id"`
	InspectionNo string  `json:"inspectionNo"`
	ReceiptID    int64   `json:"receiptId"`
	ReceiptNo    string  `json:"receiptNo"`
	SiteID       int64   `json:"siteId"`
	SupplierID   int64   `json:"supplierId"`
	MaterialID   int64   `json:"materialId"`
	Inspector    string  `json:"inspector"`
	SampleNo     string  `json:"sampleNo"`
	Moisture     float64 `json:"moisture"`
	MudContent   float64 `json:"mudContent"`
	Fineness     string  `json:"fineness"`
	Result       string  `json:"result"`
	Status       string  `json:"status"`
	Remark       string  `json:"remark"`
	CreatedAt    string  `json:"createdAt"`
	CompletedAt  string  `json:"completedAt"`
}

type InventoryItem struct {
	ID              int64   `json:"id"`
	SiteID          int64   `json:"siteId"`
	Warehouse       string  `json:"warehouse"`
	Silo            string  `json:"silo"`
	MaterialID      int64   `json:"materialId"`
	BatchNo         string  `json:"batchNo"`
	RawReceiptID    int64   `json:"rawReceiptId"`
	SupplierID      int64   `json:"supplierId"`
	Quantity        float64 `json:"quantity"`
	Unit            string  `json:"unit"`
	QualityStatus   string  `json:"qualityStatus"`
	AvailableStatus string  `json:"availableStatus"`
	UpdatedAt       string  `json:"updatedAt"`
}

type PurchaseRequest struct {
	ID         int64   `json:"id"`
	RequestNo  string  `json:"requestNo"`
	SiteID     int64   `json:"siteId"`
	MaterialID int64   `json:"materialId"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
	RequiredAt string  `json:"requiredAt"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
}

type PurchaseOrder struct {
	ID         int64   `json:"id"`
	OrderNo    string  `json:"orderNo"`
	RequestID  int64   `json:"requestId"`
	SupplierID int64   `json:"supplierId"`
	MaterialID int64   `json:"materialId"`
	Quantity   float64 `json:"quantity"`
	UnitPrice  float64 `json:"unitPrice"`
	Unit       string  `json:"unit"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
}

type RawMaterialReceipt struct {
	ID              int64   `json:"id"`
	ReceiptNo       string  `json:"receiptNo"`
	PurchaseOrderID int64   `json:"purchaseOrderId"`
	SupplierID      int64   `json:"supplierId"`
	TicketID        int64   `json:"ticketId"`
	SiteID          int64   `json:"siteId"`
	MaterialID      int64   `json:"materialId"`
	PlateNo         string  `json:"plateNo"`
	GrossWeight     float64 `json:"grossWeight"`
	TareWeight      float64 `json:"tareWeight"`
	NetWeight       float64 `json:"netWeight"`
	QualityStatus   string  `json:"qualityStatus"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"createdAt"`
}

type InventoryFlow struct {
	ID         int64   `json:"id"`
	FlowNo     string  `json:"flowNo"`
	SiteID     int64   `json:"siteId"`
	MaterialID int64   `json:"materialId"`
	SourceType string  `json:"sourceType"`
	SourceID   int64   `json:"sourceId"`
	Direction  string  `json:"direction"`
	Quantity   float64 `json:"quantity"`
	BalanceQty float64 `json:"balanceQty"`
	Remark     string  `json:"remark"`
	CreatedAt  string  `json:"createdAt"`
}

type InventoryTransfer struct {
	ID          int64   `json:"id"`
	TransferNo  string  `json:"transferNo"`
	FromSiteID  int64   `json:"fromSiteId"`
	ToSiteID    int64   `json:"toSiteId"`
	MaterialID  int64   `json:"materialId"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	Status      string  `json:"status"`
	Remark      string  `json:"remark"`
	CreatedAt   string  `json:"createdAt"`
	CompletedAt string  `json:"completedAt"`
}

type InventoryStocktake struct {
	ID          int64   `json:"id"`
	StocktakeNo string  `json:"stocktakeNo"`
	SiteID      int64   `json:"siteId"`
	MaterialID  int64   `json:"materialId"`
	BookQty     float64 `json:"bookQty"`
	ActualQty   float64 `json:"actualQty"`
	DiffQty     float64 `json:"diffQty"`
	Unit        string  `json:"unit"`
	Status      string  `json:"status"`
	Remark      string  `json:"remark"`
	CreatedAt   string  `json:"createdAt"`
	ReviewedAt  string  `json:"reviewedAt"`
}

type InventoryBatchTrace struct {
	ID                int64   `json:"id"`
	TraceNo           string  `json:"traceNo"`
	ProductionBatchID int64   `json:"productionBatchId"`
	ProductionBatchNo string  `json:"productionBatchNo"`
	RawReceiptID      int64   `json:"rawReceiptId"`
	InventoryItemID   int64   `json:"inventoryItemId"`
	SiteID            int64   `json:"siteId"`
	MaterialID        int64   `json:"materialId"`
	SupplierID        int64   `json:"supplierId"`
	BatchNo           string  `json:"batchNo"`
	Warehouse         string  `json:"warehouse"`
	Silo              string  `json:"silo"`
	Quantity          float64 `json:"quantity"`
	Unit              string  `json:"unit"`
	CreatedAt         string  `json:"createdAt"`
}

type DispatchOrder struct {
	ID           int64   `json:"id"`
	DispatchNo   string  `json:"dispatchNo"`
	OrderID      int64   `json:"orderId"`
	VehicleID    int64   `json:"vehicleId"`
	DriverID     int64   `json:"driverId"`
	SiteID       int64   `json:"siteId"`
	ProjectID    int64   `json:"projectId"`
	PlanQuantity float64 `json:"planQuantity"`
	LoadedQty    float64 `json:"loadedQty"`
	SignedQty    float64 `json:"signedQty"`
	QueueNo      string  `json:"queueNo"`
	ETA          string  `json:"eta"`
	Status       string  `json:"status"`
	Exception    string  `json:"exception"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	LineID       int64   `json:"lineId"`
	LineSeq      int     `json:"lineSeq"`
	ProductID    int64   `json:"productId"`
	ProductName  string  `json:"productName"`
}

type DispatchSchedule struct {
	ID          int64   `json:"id"`
	ScheduleNo  string  `json:"scheduleNo"`
	SiteID      int64   `json:"siteId"`
	VehicleID   int64   `json:"vehicleId"`
	DriverID    int64   `json:"driverId"`
	CarrierID   int64   `json:"carrierId"`
	ShiftDate   string  `json:"shiftDate"`
	Shift       string  `json:"shift"`
	CapacityQty float64 `json:"capacityQty"`
	AssignedQty float64 `json:"assignedQty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type DispatchCenterOverview struct {
	KPIs              DispatchCenterKPI              `json:"kpis"`
	SiteProgress      []DispatchCenterSiteProgress   `json:"siteProgress"`
	VehicleQueue      []DispatchCenterQueueItem      `json:"vehicleQueue"`
	ProductionTasks   []DispatchCenterProductionTask `json:"productionTasks"`
	AvailableVehicles []DispatchCenterVehicle        `json:"availableVehicles"`
	LatestLocations   []VehicleLatestLocation        `json:"latestLocations"`
}

type DispatchCenterKPI struct {
	TotalVehicles         int     `json:"totalVehicles"`
	OnlineVehicles        int     `json:"onlineVehicles"`
	IdleVehicles          int     `json:"idleVehicles"`
	QueueVehicles         int     `json:"queueVehicles"`
	LoadingVehicles       int     `json:"loadingVehicles"`
	InTransitVehicles     int     `json:"inTransitVehicles"`
	ArrivedVehicles       int     `json:"arrivedVehicles"`
	ActiveDispatches      int     `json:"activeDispatches"`
	OpenSupplyOrders      int     `json:"openSupplyOrders"`
	ActiveProductionTasks int     `json:"activeProductionTasks"`
	VehicleOnlineRate     float64 `json:"vehicleOnlineRate"`
}

type DispatchCenterSiteProgress struct {
	OrderID           int64   `json:"orderId"`
	OrderNo           string  `json:"orderNo"`
	CustomerID        int64   `json:"customerId"`
	CustomerName      string  `json:"customerName"`
	ProjectID         int64   `json:"projectId"`
	ProjectName       string  `json:"projectName"`
	SiteID            int64   `json:"siteId"`
	SiteName          string  `json:"siteName"`
	ProductID         int64   `json:"productId"`
	ProductName       string  `json:"productName"`
	Unit              string  `json:"unit"`
	PlanQuantity      float64 `json:"planQuantity"`
	ProducedQty       float64 `json:"producedQty"`
	DispatchedQty     float64 `json:"dispatchedQty"`
	LoadedQty         float64 `json:"loadedQty"`
	SignedQty         float64 `json:"signedQty"`
	RemainingQty      float64 `json:"remainingQty"`
	ProducedPercent   float64 `json:"producedPercent"`
	DispatchedPercent float64 `json:"dispatchedPercent"`
	LoadedPercent     float64 `json:"loadedPercent"`
	SignedPercent     float64 `json:"signedPercent"`
	ActiveDispatches  int     `json:"activeDispatches"`
	QueueVehicles     int     `json:"queueVehicles"`
	InTransitVehicles int     `json:"inTransitVehicles"`
	NextETA           string  `json:"nextEta"`
	Status            string  `json:"status"`
}

type DispatchCenterQueueItem struct {
	DispatchID     int64   `json:"dispatchId"`
	DispatchNo     string  `json:"dispatchNo"`
	OrderID        int64   `json:"orderId"`
	OrderNo        string  `json:"orderNo"`
	CustomerName   string  `json:"customerName"`
	ProjectID      int64   `json:"projectId"`
	ProjectName    string  `json:"projectName"`
	SiteID         int64   `json:"siteId"`
	SiteName       string  `json:"siteName"`
	ProductName    string  `json:"productName"`
	VehicleID      int64   `json:"vehicleId"`
	PlateNo        string  `json:"plateNo"`
	DriverID       int64   `json:"driverId"`
	DriverName     string  `json:"driverName"`
	QueueNo        string  `json:"queueNo"`
	ETA            string  `json:"eta"`
	PlannedETA     string  `json:"plannedEta"`
	ETASource      string  `json:"etaSource"`
	ETAMinutes     float64 `json:"etaMinutes"`
	ETADistanceKm  float64 `json:"etaDistanceKm"`
	ETAConfidence  string  `json:"etaConfidence"`
	ETATarget      string  `json:"etaTarget"`
	ETASpeedKPH    float64 `json:"etaSpeedKph"`
	Status         string  `json:"status"`
	PlanQuantity   float64 `json:"planQuantity"`
	LoadedQty      float64 `json:"loadedQty"`
	SignedQty      float64 `json:"signedQty"`
	OnlineStatus   string  `json:"onlineStatus"`
	BusinessStatus string  `json:"businessStatus"`
	LastLocationAt string  `json:"lastLocationAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type DispatchCenterProductionTask struct {
	TaskID        int64   `json:"taskId"`
	TaskNo        string  `json:"taskNo"`
	PlanID        int64   `json:"planId"`
	PlanNo        string  `json:"planNo"`
	OrderID       int64   `json:"orderId"`
	OrderNo       string  `json:"orderNo"`
	CustomerName  string  `json:"customerName"`
	ProjectID     int64   `json:"projectId"`
	ProjectName   string  `json:"projectName"`
	SiteID        int64   `json:"siteId"`
	SiteName      string  `json:"siteName"`
	ProductID     int64   `json:"productId"`
	ProductName   string  `json:"productName"`
	MixDesignCode string  `json:"mixDesignCode"`
	PlanQty       float64 `json:"planQty"`
	ProducedQty   float64 `json:"producedQty"`
	RemainingQty  float64 `json:"remainingQty"`
	Progress      float64 `json:"progress"`
	Status        string  `json:"status"`
	StartedAt     string  `json:"startedAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type DispatchCenterVehicle struct {
	VehicleID         int64   `json:"vehicleId"`
	PlateNo           string  `json:"plateNo"`
	VehicleType       string  `json:"vehicleType"`
	Capacity          string  `json:"capacity"`
	Carrier           string  `json:"carrier"`
	SiteID            int64   `json:"siteId"`
	SiteName          string  `json:"siteName"`
	DriverID          int64   `json:"driverId"`
	DriverName        string  `json:"driverName"`
	OnlineStatus      string  `json:"onlineStatus"`
	BusinessStatus    string  `json:"businessStatus"`
	ScheduleNo        string  `json:"scheduleNo"`
	ScheduleRemaining float64 `json:"scheduleRemaining"`
	LastLocationAt    string  `json:"lastLocationAt"`
}

type ScaleDevice struct {
	ID       int64  `json:"id"`
	SiteID   int64  `json:"siteId"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Protocol string `json:"protocol"`
	IP       string `json:"ip"`
	Status   string `json:"status"`
}

type ScaleTicket struct {
	ID               int64   `json:"id"`
	TicketNo         string  `json:"ticketNo"`
	TicketType       string  `json:"ticketType"`
	DispatchID       int64   `json:"dispatchId"`
	OrderID          int64   `json:"orderId"`
	SiteID           int64   `json:"siteId"`
	VehicleID        int64   `json:"vehicleId"`
	PlateNo          string  `json:"plateNo"`
	GrossWeight      float64 `json:"grossWeight"`
	TareWeight       float64 `json:"tareWeight"`
	NetWeight        float64 `json:"netWeight"`
	Unit             string  `json:"unit"`
	SnapshotURL      string  `json:"snapshotUrl"`
	PrintCount       int     `json:"printCount"`
	SignStatus       string  `json:"signStatus"`
	SettlementStatus string  `json:"settlementStatus"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"createdAt"`
	ReceiptID        int64   `json:"receiptId"`
	SupplierID       int64   `json:"supplierId"`
	MaterialID       int64   `json:"materialId"`
	TransferID       int64   `json:"transferId"`
	RelatedTicketID  int64   `json:"relatedTicketId"`
	Remark           string  `json:"remark"`
}

type ScaleWeightRecord struct {
	ID          int64   `json:"id"`
	DeviceID    int64   `json:"deviceId"`
	TicketID    int64   `json:"ticketId"`
	PlateNo     string  `json:"plateNo"`
	Weight      float64 `json:"weight"`
	WeightType  string  `json:"weightType"`
	SnapshotURL string  `json:"snapshotUrl"`
	CreatedAt   string  `json:"createdAt"`
}

type ScaleDeviceEvent struct {
	ID                int64   `json:"id"`
	EventNo           string  `json:"eventNo"`
	DeviceID          int64   `json:"deviceId"`
	DeviceCode        string  `json:"deviceCode"`
	TicketID          int64   `json:"ticketId"`
	PlateNo           string  `json:"plateNo"`
	RecognizedPlateNo string  `json:"recognizedPlateNo"`
	Weight            float64 `json:"weight"`
	WeightType        string  `json:"weightType"`
	Stable            bool    `json:"stable"`
	SnapshotURL       string  `json:"snapshotUrl"`
	CheatFlag         bool    `json:"cheatFlag"`
	CheatReason       string  `json:"cheatReason"`
	Status            string  `json:"status"`
	ReceivedAt        string  `json:"receivedAt"`
}

type DeliveryNote struct {
	ID         int64  `json:"id"`
	NoteNo     string `json:"noteNo"`
	TicketID   int64  `json:"ticketId"`
	OrderID    int64  `json:"orderId"`
	DispatchID int64  `json:"dispatchId"`
	QRCode     string `json:"qrCode"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

type DeliverySignLink struct {
	ID          int64  `json:"id"`
	LinkNo      string `json:"linkNo"`
	DispatchID  int64  `json:"dispatchId"`
	TicketID    int64  `json:"ticketId"`
	OrderID     int64  `json:"orderId"`
	LineID      int64  `json:"lineId"`
	LineSeq     int    `json:"lineSeq"`
	ProductID   int64  `json:"productId"`
	ProductName string `json:"productName"`
	CustomerID  int64  `json:"customerId"`
	ProjectID   int64  `json:"projectId"`
	Channel     string `json:"channel"`
	Phone       string `json:"phone"`
	Token       string `json:"token"`
	URL         string `json:"url"`
	QRCode      string `json:"qrCode"`
	Status      string `json:"status"`
	SentAt      string `json:"sentAt"`
	ExpiresAt   string `json:"expiresAt"`
	UsedAt      string `json:"usedAt"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
}

type TicketPrintLog struct {
	ID        int64  `json:"id"`
	TicketID  int64  `json:"ticketId"`
	PrintedBy string `json:"printedBy"`
	PrintedAt string `json:"printedAt"`
}

type TicketVoidLog struct {
	ID         int64  `json:"id"`
	TicketID   int64  `json:"ticketId"`
	Reason     string `json:"reason"`
	ApprovedBy string `json:"approvedBy"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

type DeliverySign struct {
	ID          int64   `json:"id"`
	SignNo      string  `json:"signNo"`
	DispatchID  int64   `json:"dispatchId"`
	LinkID      int64   `json:"linkId"`
	TicketID    int64   `json:"ticketId"`
	OrderID     int64   `json:"orderId"`
	LineID      int64   `json:"lineId"`
	LineSeq     int     `json:"lineSeq"`
	ProductID   int64   `json:"productId"`
	ProductName string  `json:"productName"`
	CustomerID  int64   `json:"customerId"`
	ProjectID   int64   `json:"projectId"`
	Signer      string  `json:"signer"`
	Phone       string  `json:"phone"`
	SignedQty   float64 `json:"signedQty"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Photo       string  `json:"photo"`
	Signature   string  `json:"signature"`
	Remark      string  `json:"remark"`
	SignedAt    string  `json:"signedAt"`
}

type DeliverySignAttachment struct {
	ID         int64  `json:"id"`
	SignID     int64  `json:"signId"`
	DispatchID int64  `json:"dispatchId"`
	TicketID   int64  `json:"ticketId"`
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	URL        string `json:"url"`
	Checksum   string `json:"checksum"`
	UploadedBy string `json:"uploadedBy"`
	UploadedAt string `json:"uploadedAt"`
}

type Statement struct {
	ID          int64           `json:"id"`
	StatementNo string          `json:"statementNo"`
	CustomerID  int64           `json:"customerId"`
	ProjectID   int64           `json:"projectId"`
	Period      string          `json:"period"`
	TotalQty    float64         `json:"totalQty"`
	TotalAmount float64         `json:"totalAmount"`
	ConfirmedBy string          `json:"confirmedBy"`
	ConfirmedAt string          `json:"confirmedAt"`
	Status      string          `json:"status"`
	Items       []StatementItem `json:"items"`
}

type StatementItem struct {
	SignID    int64   `json:"signId"`
	OrderID   int64   `json:"orderId"`
	LineID    int64   `json:"lineId"`
	TicketID  int64   `json:"ticketId"`
	ProductID int64   `json:"productId"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Amount    float64 `json:"amount"`
}

type SalesInvoice struct {
	ID                int64   `json:"id"`
	InvoiceNo         string  `json:"invoiceNo"`
	StatementID       int64   `json:"statementId"`
	CustomerID        int64   `json:"customerId"`
	Amount            float64 `json:"amount"`
	TaxRate           float64 `json:"taxRate"`
	TaxAmount         float64 `json:"taxAmount"`
	TaxControlNo      string  `json:"taxControlNo"`
	TaxStatus         string  `json:"taxStatus"`
	FileURL           string  `json:"fileUrl"`
	DownloadedAt      string  `json:"downloadedAt"`
	Status            string  `json:"status"`
	IssuedAt          string  `json:"issuedAt"`
	InvoiceType       string  `json:"invoiceType"`
	InvoiceCategory   string  `json:"invoiceCategory"`
	OriginalInvoiceID int64   `json:"originalInvoiceId"`
	RedLetterInfoID   int64   `json:"redLetterInfoId"`
	RedLetterInfoNo   string  `json:"redLetterInfoNo"`
	RedReason         string  `json:"redReason"`
	RedAt             string  `json:"redAt"`
}

type RedLetterInfo struct {
	ID                int64   `json:"id"`
	InfoNo            string  `json:"infoNo"`
	OriginalInvoiceID int64   `json:"originalInvoiceId"`
	OriginalInvoiceNo string  `json:"originalInvoiceNo"`
	RedInvoiceID      int64   `json:"redInvoiceId"`
	CustomerID        int64   `json:"customerId"`
	Amount            float64 `json:"amount"`
	TaxAmount         float64 `json:"taxAmount"`
	Reason            string  `json:"reason"`
	Applicant         string  `json:"applicant"`
	Status            string  `json:"status"`
	TaxControlNo      string  `json:"taxControlNo"`
	RequestedAt       string  `json:"requestedAt"`
	ApprovedBy        string  `json:"approvedBy"`
	ApprovedAt        string  `json:"approvedAt"`
	UsedAt            string  `json:"usedAt"`
	Remark            string  `json:"remark"`
}

type TaxGatewaySubmission struct {
	ID           int64  `json:"id"`
	SubmissionNo string `json:"submissionNo"`
	InvoiceID    int64  `json:"invoiceId"`
	InvoiceNo    string `json:"invoiceNo"`
	Action       string `json:"action"`
	Provider     string `json:"provider"`
	Endpoint     string `json:"endpoint"`
	RequestID    string `json:"requestId"`
	Status       string `json:"status"`
	TaxControlNo string `json:"taxControlNo"`
	FileURL      string `json:"fileUrl"`
	Error        string `json:"error"`
	Attempt      int    `json:"attempt"`
	DurationMs   int64  `json:"durationMs"`
	SubmittedAt  string `json:"submittedAt"`
	CompletedAt  string `json:"completedAt"`
	Actor        string `json:"actor"`
}

type Receivable struct {
	ID             int64   `json:"id"`
	BillNo         string  `json:"billNo"`
	CustomerID     int64   `json:"customerId"`
	StatementID    int64   `json:"statementId"`
	InvoiceID      int64   `json:"invoiceId"`
	Amount         float64 `json:"amount"`
	ReceivedAmount float64 `json:"receivedAmount"`
	DueDate        string  `json:"dueDate"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"createdAt"`
}

type Receipt struct {
	ID           int64   `json:"id"`
	ReceiptNo    string  `json:"receiptNo"`
	ReceivableID int64   `json:"receivableId"`
	CustomerID   int64   `json:"customerId"`
	Amount       float64 `json:"amount"`
	Method       string  `json:"method"`
	Status       string  `json:"status"`
	ReceivedAt   string  `json:"receivedAt"`
}

type PaymentPlan struct {
	ID           int64   `json:"id"`
	PlanNo       string  `json:"planNo"`
	ReceivableID int64   `json:"receivableId"`
	CustomerID   int64   `json:"customerId"`
	Amount       float64 `json:"amount"`
	DueDate      string  `json:"dueDate"`
	Method       string  `json:"method"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
	SettledAt    string  `json:"settledAt"`
	Remark       string  `json:"remark"`
}

type ReceivableAgingBucket struct {
	Bucket        string  `json:"bucket"`
	Label         string  `json:"label"`
	Count         int     `json:"count"`
	Amount        float64 `json:"amount"`
	OverdueAmount float64 `json:"overdueAmount"`
}

type CollectionTask struct {
	ID           int64   `json:"id"`
	TaskNo       string  `json:"taskNo"`
	ReceivableID int64   `json:"receivableId"`
	CustomerID   int64   `json:"customerId"`
	CustomerName string  `json:"customerName"`
	Amount       float64 `json:"amount"`
	DueDate      string  `json:"dueDate"`
	OverdueDays  int     `json:"overdueDays"`
	Level        string  `json:"level"`
	Channel      string  `json:"channel"`
	Status       string  `json:"status"`
	Message      string  `json:"message"`
	TemplateID   int64   `json:"templateId"`
	SendCount    int     `json:"sendCount"`
	LastSentAt   string  `json:"lastSentAt"`
	GeneratedAt  string  `json:"generatedAt"`
	HandledBy    string  `json:"handledBy"`
	HandledAt    string  `json:"handledAt"`
	Remark       string  `json:"remark"`
}

type CollectionTemplate struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Level     string `json:"level"`
	Channel   string `json:"channel"`
	Content   string `json:"content"`
	Enabled   bool   `json:"enabled"`
	UpdatedAt string `json:"updatedAt"`
}

type CollectionDispatch struct {
	ID                int64  `json:"id"`
	DispatchNo        string `json:"dispatchNo"`
	TaskID            int64  `json:"taskId"`
	TemplateID        int64  `json:"templateId"`
	CustomerID        int64  `json:"customerId"`
	Channel           string `json:"channel"`
	Target            string `json:"target"`
	Content           string `json:"content"`
	Endpoint          string `json:"endpoint"`
	ProviderRequestID string `json:"providerRequestId"`
	ProviderMessageID string `json:"providerMessageId"`
	Status            string `json:"status"`
	Error             string `json:"error"`
	SentAt            string `json:"sentAt"`
	CallbackAt        string `json:"callbackAt"`
	Actor             string `json:"actor"`
}

type SupplierStatement struct {
	ID          int64   `json:"id"`
	StatementNo string  `json:"statementNo"`
	SupplierID  int64   `json:"supplierId"`
	Period      string  `json:"period"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	ApprovedBy  string  `json:"approvedBy"`
	ApprovedAt  string  `json:"approvedAt"`
}

type Payable struct {
	ID                  int64   `json:"id"`
	BillNo              string  `json:"billNo"`
	SupplierID          int64   `json:"supplierId"`
	SupplierStatementID int64   `json:"supplierStatementId"`
	Amount              float64 `json:"amount"`
	PaidAmount          float64 `json:"paidAmount"`
	DueDate             string  `json:"dueDate"`
	Status              string  `json:"status"`
}

type Payment struct {
	ID         int64   `json:"id"`
	PaymentNo  string  `json:"paymentNo"`
	PayableID  int64   `json:"payableId"`
	SupplierID int64   `json:"supplierId"`
	Amount     float64 `json:"amount"`
	Method     string  `json:"method"`
	PaidAt     string  `json:"paidAt"`
}

type TransportSettlement struct {
	ID           int64   `json:"id"`
	SettlementNo string  `json:"settlementNo"`
	CarrierID    int64   `json:"carrierId"`
	Period       string  `json:"period"`
	TripCount    int     `json:"tripCount"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
}

type TransportSettlementItem struct {
	ID           int64   `json:"id"`
	SettlementID int64   `json:"settlementId"`
	DispatchID   int64   `json:"dispatchId"`
	DispatchNo   string  `json:"dispatchNo"`
	CarrierID    int64   `json:"carrierId"`
	VehicleID    int64   `json:"vehicleId"`
	DriverID     int64   `json:"driverId"`
	Quantity     float64 `json:"quantity"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
}

type CostCalc struct {
	ID            int64   `json:"id"`
	ProjectID     int64   `json:"projectId"`
	OrderID       int64   `json:"orderId"`
	MaterialCost  float64 `json:"materialCost"`
	TransportCost float64 `json:"transportCost"`
	LaborCost     float64 `json:"laborCost"`
	OtherCost     float64 `json:"otherCost"`
	TotalCost     float64 `json:"totalCost"`
	CreatedAt     string  `json:"createdAt"`
}

type ProjectProfit struct {
	ID        int64   `json:"id"`
	ProjectID int64   `json:"projectId"`
	Revenue   float64 `json:"revenue"`
	Cost      float64 `json:"cost"`
	Profit    float64 `json:"profit"`
	Margin    float64 `json:"margin"`
	Period    string  `json:"period"`
}

type VehicleLocationEvent struct {
	ID           int64   `json:"id"`
	VehicleID    int64   `json:"vehicleId"`
	PlateNo      string  `json:"plateNo"`
	DriverID     int64   `json:"driverId"`
	DispatchID   int64   `json:"dispatchId"`
	DeviceID     string  `json:"deviceId"`
	SourceType   string  `json:"sourceType"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	Speed        float64 `json:"speed"`
	Direction    float64 `json:"direction"`
	Mileage      float64 `json:"mileage"`
	AccStatus    int     `json:"accStatus"`
	OnlineStatus string  `json:"onlineStatus"`
	Address      string  `json:"address"`
	IsAbnormal   bool    `json:"isAbnormal"`
	AbnormalType string  `json:"abnormalType"`
	LocationTime string  `json:"locationTime"`
	ReceiveTime  string  `json:"receiveTime"`
}

type TrackStopPoint struct {
	Longitude       float64 `json:"longitude"`
	Latitude        float64 `json:"latitude"`
	StartTime       string  `json:"startTime"`
	EndTime         string  `json:"endTime"`
	DurationMinutes float64 `json:"durationMinutes"`
	Address         string  `json:"address"`
}

type TrackCompressionSummary struct {
	Algorithm               string  `json:"algorithm"`
	ToleranceMeters         float64 `json:"toleranceMeters"`
	RawPointCount           int     `json:"rawPointCount"`
	CompressedPointCount    int     `json:"compressedPointCount"`
	CompressionRatio        float64 `json:"compressionRatio"`
	ReductionPercent        float64 `json:"reductionPercent"`
	PreservedStops          int     `json:"preservedStops"`
	PreservedAbnormalPoints int     `json:"preservedAbnormalPoints"`
}

type TrackReplay struct {
	VehicleID        int64                   `json:"vehicleId"`
	PlateNo          string                  `json:"plateNo"`
	StartTime        string                  `json:"startTime"`
	EndTime          string                  `json:"endTime"`
	DistanceKm       float64                 `json:"distanceKm"`
	DurationMinutes  float64                 `json:"durationMinutes"`
	AverageSpeed     float64                 `json:"averageSpeed"`
	MaxSpeed         float64                 `json:"maxSpeed"`
	StopCount        int                     `json:"stopCount"`
	Points           []VehicleLocationEvent  `json:"points"`
	CompressedPoints []VehicleLocationEvent  `json:"compressedPoints"`
	Compression      TrackCompressionSummary `json:"compression"`
	Stops            []TrackStopPoint        `json:"stops"`
	FenceEvents      []GeoFenceEvent         `json:"fenceEvents"`
	Tickets          []ScaleTicket           `json:"tickets"`
	Signs            []DeliverySign          `json:"signs"`
	ExportName       string                  `json:"exportName"`
}

type VehicleDevice struct {
	ID         int64  `json:"id"`
	VehicleID  int64  `json:"vehicleId"`
	DeviceNo   string `json:"deviceNo"`
	Protocol   string `json:"protocol"`
	Vendor     string `json:"vendor"`
	Status     string `json:"status"`
	LastSeenAt string `json:"lastSeenAt"`
}

type DeviceCredential struct {
	ID         int64    `json:"id"`
	DeviceNo   string   `json:"deviceNo"`
	KeyHash    string   `json:"keyHash,omitempty"`
	Scopes     []string `json:"scopes"`
	Status     string   `json:"status"`
	LastUsedAt string   `json:"lastUsedAt"`
}

type DeviceProtocolFrame struct {
	ID             int64  `json:"id"`
	FrameNo        string `json:"frameNo"`
	Channel        string `json:"channel"`
	Protocol       string `json:"protocol"`
	DeviceNo       string `json:"deviceNo"`
	Raw            string `json:"raw"`
	ParsedResource string `json:"parsedResource"`
	ParsedID       int64  `json:"parsedId"`
	Status         string `json:"status"`
	Error          string `json:"error"`
	ReceivedAt     string `json:"receivedAt"`
	Actor          string `json:"actor"`
}

type SecurityPolicy struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
	Remark  string `json:"remark"`
}

type FieldPolicy struct {
	ID       int64  `json:"id"`
	RoleCode string `json:"roleCode"`
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Mask     string `json:"mask"`
	Enabled  bool   `json:"enabled"`
	Remark   string `json:"remark"`
}

type TenantPolicy struct {
	ID                    int64    `json:"id"`
	TenantID              int64    `json:"tenantId"`
	Name                  string   `json:"name"`
	Code                  string   `json:"code"`
	Mode                  string   `json:"mode"`
	EnforceTenantBoundary bool     `json:"enforceTenantBoundary"`
	IsolateSystemSettings bool     `json:"isolateSystemSettings"`
	DataDomains           []string `json:"dataDomains"`
	DataSourceRef         string   `json:"dataSourceRef"`
	Status                string   `json:"status"`
	Remark                string   `json:"remark"`
	UpdatedAt             string   `json:"updatedAt"`
}

type VehicleLatestLocation struct {
	VehicleID         int64   `json:"vehicleId"`
	PlateNo           string  `json:"plateNo"`
	Longitude         float64 `json:"longitude"`
	Latitude          float64 `json:"latitude"`
	Speed             float64 `json:"speed"`
	Direction         float64 `json:"direction"`
	OnlineStatus      string  `json:"onlineStatus"`
	TransportStatus   string  `json:"transportStatus"`
	LastLocationTime  string  `json:"lastLocationTime"`
	CurrentOrderID    int64   `json:"currentOrderId"`
	CurrentProjectID  int64   `json:"currentProjectId"`
	CurrentSiteID     int64   `json:"currentSiteId"`
	CurrentCustomerID int64   `json:"currentCustomerId"`
}

type GeoPoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type GeoFence struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	SiteID    int64      `json:"siteId"`
	ProjectID int64      `json:"projectId"`
	Longitude float64    `json:"longitude"`
	Latitude  float64    `json:"latitude"`
	Radius    float64    `json:"radius"`
	Shape     string     `json:"shape"`
	Polygon   []GeoPoint `json:"polygon"`
	Status    string     `json:"status"`
}

type GeoFenceEvent struct {
	ID         int64  `json:"id"`
	VehicleID  int64  `json:"vehicleId"`
	FenceID    int64  `json:"fenceId"`
	EventType  string `json:"eventType"`
	DispatchID int64  `json:"dispatchId"`
	EventTime  string `json:"eventTime"`
}

type VehicleAlarm struct {
	ID         int64  `json:"id"`
	VehicleID  int64  `json:"vehicleId"`
	DispatchID int64  `json:"dispatchId"`
	AlarmType  string `json:"alarmType"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	HandledBy  string `json:"handledBy"`
	HandledAt  string `json:"handledAt"`
}

type RuleDefinition struct {
	ID          int64    `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Metric      string   `json:"metric"`
	Operator    string   `json:"operator"`
	Threshold   float64  `json:"threshold"`
	Level       string   `json:"level"`
	Enabled     bool     `json:"enabled"`
	NotifyRoles []string `json:"notifyRoles"`
	Description string   `json:"description"`
}

type Notification struct {
	ID         int64  `json:"id"`
	TargetRole string `json:"targetRole"`
	Channel    string `json:"channel"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

type IntegrationEndpoint struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Protocol   string `json:"protocol"`
	URL        string `json:"url"`
	Status     string `json:"status"`
	LastSyncAt string `json:"lastSyncAt"`
}

type ApprovalFlow struct {
	ID       int64          `json:"id"`
	Code     string         `json:"code"`
	Name     string         `json:"name"`
	Resource string         `json:"resource"`
	Steps    []ApprovalStep `json:"steps"`
	Status   string         `json:"status"`
}

type ApprovalStep struct {
	Seq      int    `json:"seq"`
	RoleCode string `json:"roleCode"`
	Action   string `json:"action"`
}

type ApprovalTask struct {
	ID          int64                `json:"id"`
	TaskNo      string               `json:"taskNo"`
	FlowCode    string               `json:"flowCode"`
	FlowName    string               `json:"flowName"`
	Resource    string               `json:"resource"`
	ResourceID  int64                `json:"resourceId"`
	ResourceNo  string               `json:"resourceNo"`
	Title       string               `json:"title"`
	Applicant   string               `json:"applicant"`
	CurrentStep int                  `json:"currentStep"`
	CurrentRole string               `json:"currentRole"`
	Status      string               `json:"status"`
	Reason      string               `json:"reason"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
	Actions     []ApprovalTaskAction `json:"actions"`
}

type ApprovalTaskAction struct {
	Seq      int    `json:"seq"`
	Step     int    `json:"step"`
	RoleCode string `json:"roleCode"`
	Action   string `json:"action"`
	Actor    string `json:"actor"`
	Comment  string `json:"comment"`
	ActedAt  string `json:"actedAt"`
}

type DataDictionary struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	Code   string `json:"code"`
	Label  string `json:"label"`
	Sort   int    `json:"sort"`
	Status string `json:"status"`
}

type UpdatePackage struct {
	ID                      int64  `json:"id"`
	Version                 string `json:"version"`
	Component               string `json:"component"`
	Channel                 string `json:"channel"`
	Status                  string `json:"status"`
	PackageType             string `json:"packageType,omitempty"`
	BaseVersion             string `json:"baseVersion,omitempty"`
	DeltaAlgorithm          string `json:"deltaAlgorithm,omitempty"`
	Checksum                string `json:"checksum"`
	Signature               string `json:"signature"`
	SignaturePublicKey      string `json:"signaturePublicKey,omitempty"`
	SignatureKeyFingerprint string `json:"signatureKeyFingerprint,omitempty"`
	FileName                string `json:"fileName"`
	SizeBytes               int64  `json:"sizeBytes"`
	ArtifactFileName        string `json:"artifactFileName,omitempty"`
	ArtifactContentType     string `json:"artifactContentType,omitempty"`
	ArtifactContentBase64   string `json:"artifactContentBase64,omitempty"`
	ArtifactSHA256          string `json:"artifactSha256,omitempty"`
	ArtifactSizeBytes       int64  `json:"artifactSizeBytes,omitempty"`
	BaseArtifactSHA256      string `json:"baseArtifactSha256,omitempty"`
	TargetArtifactSHA256    string `json:"targetArtifactSha256,omitempty"`
	RollbackVersion         string `json:"rollbackVersion"`
	PublishedBy             string `json:"publishedBy"`
	PublishedAt             string `json:"publishedAt"`
	DownloadCount           int    `json:"downloadCount"`
	LastDownloadedAt        string `json:"lastDownloadedAt"`
	AppliedBy               string `json:"appliedBy"`
	AppliedAt               string `json:"appliedAt"`
	CreatedAt               string `json:"createdAt"`
	Remark                  string `json:"remark"`
}

type ProductUpdateRollout struct {
	ID             int64                      `json:"id"`
	RolloutNo      string                     `json:"rolloutNo"`
	UpdateID       int64                      `json:"updateId"`
	Version        string                     `json:"version"`
	Component      string                     `json:"component"`
	Strategy       string                     `json:"strategy"`
	Status         string                     `json:"status"`
	TotalTargets   int                        `json:"totalTargets"`
	AppliedTargets int                        `json:"appliedTargets"`
	FailedTargets  int                        `json:"failedTargets"`
	CreatedBy      string                     `json:"createdBy"`
	CreatedAt      string                     `json:"createdAt"`
	StartedAt      string                     `json:"startedAt"`
	CompletedAt    string                     `json:"completedAt"`
	Remark         string                     `json:"remark"`
	Items          []ProductUpdateRolloutItem `json:"items"`
}

type ProductUpdateRolloutItem struct {
	ID           int64  `json:"id"`
	InstanceID   int64  `json:"instanceId"`
	CustomerName string `json:"customerName"`
	FromVersion  string `json:"fromVersion"`
	ToVersion    string `json:"toVersion"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	StartedAt    string `json:"startedAt"`
	AppliedAt    string `json:"appliedAt"`
	RolledBackAt string `json:"rolledBackAt"`
}

type ProductUpdateExecution struct {
	ID               int64                        `json:"id"`
	ExecutionNo      string                       `json:"executionNo"`
	RolloutID        int64                        `json:"rolloutId"`
	RolloutNo        string                       `json:"rolloutNo"`
	UpdateID         int64                        `json:"updateId"`
	InstanceID       int64                        `json:"instanceId"`
	CustomerName     string                       `json:"customerName"`
	Component        string                       `json:"component"`
	Version          string                       `json:"version"`
	Action           string                       `json:"action"`
	Status           string                       `json:"status"`
	ArtifactFileName string                       `json:"artifactFileName"`
	ChecksumVerified bool                         `json:"checksumVerified"`
	DryRun           bool                         `json:"dryRun"`
	StartedBy        string                       `json:"startedBy"`
	StartedAt        string                       `json:"startedAt"`
	CompletedAt      string                       `json:"completedAt"`
	DurationMs       int64                        `json:"durationMs"`
	PrecheckSummary  string                       `json:"precheckSummary"`
	Result           string                       `json:"result"`
	Error            string                       `json:"error"`
	Steps            []ProductUpdateExecutionStep `json:"steps"`
}

type ProductSystemUpdateTask struct {
	ID               int64                        `json:"id"`
	TaskNo           string                       `json:"taskNo"`
	ExecutionID      int64                        `json:"executionId"`
	ExecutionNo      string                       `json:"executionNo"`
	RolloutID        int64                        `json:"rolloutId"`
	RolloutNo        string                       `json:"rolloutNo"`
	RolloutItemID    int64                        `json:"rolloutItemId"`
	UpdateID         int64                        `json:"updateId"`
	InstanceID       int64                        `json:"instanceId"`
	CustomerName     string                       `json:"customerName"`
	Watermark        string                       `json:"watermark"`
	Component        string                       `json:"component"`
	Version          string                       `json:"version"`
	FromVersion      string                       `json:"fromVersion"`
	Action           string                       `json:"action"`
	Status           string                       `json:"status"`
	Progress         int                          `json:"progress"`
	ArtifactFileName string                       `json:"artifactFileName"`
	Checksum         string                       `json:"checksum"`
	Signature        string                       `json:"signature"`
	DownloadURL      string                       `json:"downloadUrl"`
	UpdaterTokenHint string                       `json:"updaterTokenHint"`
	CreatedBy        string                       `json:"createdBy"`
	CreatedAt        string                       `json:"createdAt"`
	ClaimedAt        string                       `json:"claimedAt"`
	StartedAt        string                       `json:"startedAt"`
	CompletedAt      string                       `json:"completedAt"`
	LastHeartbeatAt  string                       `json:"lastHeartbeatAt"`
	Result           string                       `json:"result"`
	Error            string                       `json:"error"`
	Remark           string                       `json:"remark"`
	Logs             []ProductSystemUpdateTaskLog `json:"logs"`
}

type ProductSystemUpdateTaskLog struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	Step      string `json:"step"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type ProductUpdateExecutionStep struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
	DurationMs  int64  `json:"durationMs"`
}

type AuditLog struct {
	ID         int64  `json:"id"`
	User       string `json:"user"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID int64  `json:"resourceId"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"createdAt"`
}

type BackupDrill struct {
	ID            int64          `json:"id"`
	DrillNo       string         `json:"drillNo"`
	BackupName    string         `json:"backupName"`
	Status        string         `json:"status"`
	StartedAt     string         `json:"startedAt"`
	CompletedAt   string         `json:"completedAt"`
	DurationMs    int64          `json:"durationMs"`
	SnapshotSize  int64          `json:"snapshotSize"`
	SchemaVersion int64          `json:"schemaVersion"`
	ObjectCounts  map[string]int `json:"objectCounts"`
	Checks        []string       `json:"checks"`
	Error         string         `json:"error"`
	Actor         string         `json:"actor"`
}

type GatewayRoute struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PathPrefix     string `json:"pathPrefix"`
	StableUpstream string `json:"stableUpstream"`
	CanaryUpstream string `json:"canaryUpstream"`
	CanaryPercent  int    `json:"canaryPercent"`
	DrainEnabled   bool   `json:"drainEnabled"`
	DrainUntil     string `json:"drainUntil"`
	ReadTimeoutSec int    `json:"readTimeoutSec"`
	Status         string `json:"status"`
	UpdatedAt      string `json:"updatedAt"`
}

type GatewayEvent struct {
	ID        int64  `json:"id"`
	EventNo   string `json:"eventNo"`
	RouteID   int64  `json:"routeId"`
	RouteName string `json:"routeName"`
	Action    string `json:"action"`
	Detail    string `json:"detail"`
	Actor     string `json:"actor"`
	CreatedAt string `json:"createdAt"`
}

type AppData struct {
	SchemaVersion                 int64                          `json:"schemaVersion"`
	License                       LicenseInfo                    `json:"license"`
	LicensePackages               []LicensePackage               `json:"licensePackages"`
	LicenseIssues                 []LicenseIssueRecord           `json:"licenseIssues"`
	LicenseRevocations            []LicenseRevocation            `json:"licenseRevocations"`
	ProductInstances              []ProductInstance              `json:"productInstances"`
	SystemAlerts                  []SystemAlert                  `json:"systemAlerts"`
	ProductRenewalTasks           []ProductRenewalTask           `json:"productRenewalTasks"`
	ProductRenewalQuotes          []ProductRenewalQuote          `json:"productRenewalQuotes"`
	ProductRenewalContracts       []ProductRenewalContract       `json:"productRenewalContracts"`
	ProductRenewalPayments        []ProductRenewalPayment        `json:"productRenewalPayments"`
	ProductRenewalApprovals       []ProductRenewalApproval       `json:"productRenewalApprovals"`
	ProductRenewalInvoices        []ProductRenewalInvoice        `json:"productRenewalInvoices"`
	ProductRenewalESigns          []ProductRenewalESign          `json:"productRenewalESigns"`
	ProductRenewalIntegrations    []ProductRenewalIntegration    `json:"productRenewalIntegrations"`
	ProductRenewalSyncRecords     []ProductRenewalSyncRecord     `json:"productRenewalSyncRecords"`
	ProductProbeReports           []ProductProbeReport           `json:"productProbeReports"`
	ProductTelemetryEvents        []ProductTelemetryEvent        `json:"productTelemetryEvents"`
	ProductMonitoringIntegrations []ProductMonitoringIntegration `json:"productMonitoringIntegrations"`
	ProductAlertRules             []ProductAlertRule             `json:"productAlertRules"`
	ProductMonitoringEvents       []ProductMonitoringEvent       `json:"productMonitoringEvents"`
	ProductAlertPolicies          []ProductAlertPolicy           `json:"productAlertPolicies"`
	ProductAlertChannels          []ProductAlertChannel          `json:"productAlertChannels"`
	ProductAlertNotifications     []ProductAlertNotification     `json:"productAlertNotifications"`
	ProductUpdateRollouts         []ProductUpdateRollout         `json:"productUpdateRollouts"`
	ProductUpdateExecutions       []ProductUpdateExecution       `json:"productUpdateExecutions"`
	ProductSystemUpdateTasks      []ProductSystemUpdateTask      `json:"productSystemUpdateTasks"`
	Modules                       []Module                       `json:"modules"`
	Plugins                       []Plugin                       `json:"plugins"`
	PluginRuns                    []PluginRun                    `json:"pluginRuns"`
	Roles                         []Role                         `json:"roles"`
	Users                         []User                         `json:"users"`
	OIDCProviders                 []OIDCProvider                 `json:"oidcProviders"`
	SCIMProviders                 []SCIMProvider                 `json:"scimProviders"`
	SCIMEvents                    []SCIMProvisioningEvent        `json:"scimEvents"`
	Tenants                       []Tenant                       `json:"tenants"`
	Companies                     []Company                      `json:"companies"`
	Sites                         []Site                         `json:"sites"`
	Departments                   []Department                   `json:"departments"`
	Plants                        []Plant                        `json:"plants"`
	Warehouses                    []Warehouse                    `json:"warehouses"`
	Silos                         []Silo                         `json:"silos"`
	Customers                     []Customer                     `json:"customers"`
	CustomerContacts              []CustomerContact              `json:"customerContacts"`
	CustomerBlacklists            []CustomerBlacklist            `json:"customerBlacklists"`
	CustomerProfiles              []CustomerProfile              `json:"customerProfiles"`
	CustomerComplaints            []CustomerComplaint            `json:"customerComplaints"`
	PricePolicies                 []PricePolicy                  `json:"pricePolicies"`
	TaxRates                      []TaxRate                      `json:"taxRates"`
	Suppliers                     []Supplier                     `json:"suppliers"`
	Carriers                      []Carrier                      `json:"carriers"`
	Projects                      []Project                      `json:"projects"`
	Products                      []Product                      `json:"products"`
	Materials                     []Material                     `json:"materials"`
	Vehicles                      []Vehicle                      `json:"vehicles"`
	Drivers                       []Driver                       `json:"drivers"`
	VehicleDevices                []VehicleDevice                `json:"vehicleDevices"`
	DeviceCredentials             []DeviceCredential             `json:"deviceCredentials"`
	DeviceProtocolFrames          []DeviceProtocolFrame          `json:"deviceProtocolFrames"`
	SecurityPolicies              []SecurityPolicy               `json:"securityPolicies"`
	FieldPolicies                 []FieldPolicy                  `json:"fieldPolicies"`
	TenantPolicies                []TenantPolicy                 `json:"tenantPolicies"`
	Contracts                     []Contract                     `json:"contracts"`
	ContractAttachments           []ContractAttachment           `json:"contractAttachments"`
	Orders                        []SalesOrder                   `json:"orders"`
	ProductionPlans               []ProductionPlan               `json:"productionPlans"`
	MixDesigns                    []MixDesign                    `json:"mixDesigns"`
	MixDesignTrialRuns            []MixDesignTrialRun            `json:"mixDesignTrialRuns"`
	ProductionTasks               []ProductionTask               `json:"productionTasks"`
	ProductionBatches             []ProductionBatch              `json:"productionBatches"`
	ProductionReports             []ProductionDailyReport        `json:"productionReports"`
	QualityInspections            []QualityInspection            `json:"qualityInspections"`
	QualitySamples                []QualitySample                `json:"qualitySamples"`
	RawMaterialInspections        []RawMaterialInspection        `json:"rawMaterialInspections"`
	LaboratorySamples             []LaboratorySample             `json:"laboratorySamples"`
	LaboratoryTests               []LaboratoryTestRecord         `json:"laboratoryTests"`
	LaboratoryEquipment           []LaboratoryEquipment          `json:"laboratoryEquipment"`
	LaboratoryCalibrations        []LaboratoryCalibration        `json:"laboratoryCalibrations"`
	QualityExceptions             []QualityException             `json:"qualityExceptions"`
	Inventory                     []InventoryItem                `json:"inventory"`
	InventoryTransfers            []InventoryTransfer            `json:"inventoryTransfers"`
	InventoryStocktakes           []InventoryStocktake           `json:"inventoryStocktakes"`
	InventoryBatchTraces          []InventoryBatchTrace          `json:"inventoryBatchTraces"`
	PurchaseRequests              []PurchaseRequest              `json:"purchaseRequests"`
	PurchaseOrders                []PurchaseOrder                `json:"purchaseOrders"`
	RawMaterialReceipts           []RawMaterialReceipt           `json:"rawMaterialReceipts"`
	InventoryFlows                []InventoryFlow                `json:"inventoryFlows"`
	DispatchOrders                []DispatchOrder                `json:"dispatchOrders"`
	DispatchSchedules             []DispatchSchedule             `json:"dispatchSchedules"`
	ScaleDevices                  []ScaleDevice                  `json:"scaleDevices"`
	ScaleTickets                  []ScaleTicket                  `json:"scaleTickets"`
	ScaleWeightRecords            []ScaleWeightRecord            `json:"scaleWeightRecords"`
	ScaleDeviceEvents             []ScaleDeviceEvent             `json:"scaleDeviceEvents"`
	DeliveryNotes                 []DeliveryNote                 `json:"deliveryNotes"`
	DeliverySignLinks             []DeliverySignLink             `json:"deliverySignLinks"`
	TicketPrintLogs               []TicketPrintLog               `json:"ticketPrintLogs"`
	TicketVoidLogs                []TicketVoidLog                `json:"ticketVoidLogs"`
	DeliverySigns                 []DeliverySign                 `json:"deliverySigns"`
	DeliverySignAttachments       []DeliverySignAttachment       `json:"deliverySignAttachments"`
	Statements                    []Statement                    `json:"statements"`
	SalesInvoices                 []SalesInvoice                 `json:"salesInvoices"`
	RedLetterInfos                []RedLetterInfo                `json:"redLetterInfos"`
	TaxGatewaySubmissions         []TaxGatewaySubmission         `json:"taxGatewaySubmissions"`
	Receivables                   []Receivable                   `json:"receivables"`
	Receipts                      []Receipt                      `json:"receipts"`
	PaymentPlans                  []PaymentPlan                  `json:"paymentPlans"`
	CollectionTasks               []CollectionTask               `json:"collectionTasks"`
	CollectionTemplates           []CollectionTemplate           `json:"collectionTemplates"`
	CollectionDispatches          []CollectionDispatch           `json:"collectionDispatches"`
	SupplierStatements            []SupplierStatement            `json:"supplierStatements"`
	Payables                      []Payable                      `json:"payables"`
	Payments                      []Payment                      `json:"payments"`
	TransportSettlements          []TransportSettlement          `json:"transportSettlements"`
	TransportSettlementItems      []TransportSettlementItem      `json:"transportSettlementItems"`
	CostCalcs                     []CostCalc                     `json:"costCalcs"`
	ProjectProfits                []ProjectProfit                `json:"projectProfits"`
	Locations                     []VehicleLocationEvent         `json:"locations"`
	LatestLocations               []VehicleLatestLocation        `json:"latestLocations"`
	GeoFences                     []GeoFence                     `json:"geoFences"`
	GeoFenceEvents                []GeoFenceEvent                `json:"geoFenceEvents"`
	VehicleAlarms                 []VehicleAlarm                 `json:"vehicleAlarms"`
	RuleDefinitions               []RuleDefinition               `json:"ruleDefinitions"`
	Notifications                 []Notification                 `json:"notifications"`
	IntegrationEndpoints          []IntegrationEndpoint          `json:"integrationEndpoints"`
	ApprovalFlows                 []ApprovalFlow                 `json:"approvalFlows"`
	ApprovalTasks                 []ApprovalTask                 `json:"approvalTasks"`
	DataDictionaries              []DataDictionary               `json:"dataDictionaries"`
	Updates                       []UpdatePackage                `json:"updates"`
	BackupDrills                  []BackupDrill                  `json:"backupDrills"`
	GatewayRoutes                 []GatewayRoute                 `json:"gatewayRoutes"`
	GatewayEvents                 []GatewayEvent                 `json:"gatewayEvents"`
	AuditLogs                     []AuditLog                     `json:"auditLogs"`
	Next                          map[string]int64               `json:"next"`
}
