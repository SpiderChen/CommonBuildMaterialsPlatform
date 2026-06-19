package ops

type AppData struct {
	SchemaVersion  int64                `json:"schemaVersion"`
	Next           map[string]int64      `json:"next"`
	Customers      []CustomerDeployment `json:"customers"`
	Renewals       []LicenseRenewal     `json:"renewals"`
	Alerts         []SystemAlert        `json:"alerts"`
	UpdatePackages []UpdatePackage      `json:"updatePackages"`
	Assignments    []UpdateAssignment   `json:"assignments"`
	AuditLogs      []AuditLog           `json:"auditLogs"`
}

type CustomerDeployment struct {
	ID                   int64    `json:"id"`
	CustomerName         string   `json:"customerName"`
	ProductName          string   `json:"productName"`
	LicenseID            string   `json:"licenseId"`
	Edition              string   `json:"edition"`
	DeploymentMode       string   `json:"deploymentMode"`
	Environment          string   `json:"environment"`
	ServerEndpoint        string   `json:"serverEndpoint"`
	ContactName          string   `json:"contactName"`
	ContactPhone         string   `json:"contactPhone"`
	ExpiresAt            string   `json:"expiresAt"`
	RenewalStatus         string   `json:"renewalStatus"`
	Modules              []string `json:"modules"`
	MaxSites             int      `json:"maxSites"`
	MaxVehicles          int      `json:"maxVehicles"`
	CurrentClientVersion string   `json:"currentClientVersion"`
	CurrentServerVersion string   `json:"currentServerVersion"`
	TargetClientVersion  string   `json:"targetClientVersion"`
	TargetServerVersion  string   `json:"targetServerVersion"`
	HealthStatus         string   `json:"healthStatus"`
	LastHeartbeatAt      string   `json:"lastHeartbeatAt"`
	Notes                string   `json:"notes"`
}

type LicenseRenewal struct {
	ID           int64  `json:"id"`
	RenewalNo    string `json:"renewalNo"`
	CustomerID   int64  `json:"customerId"`
	LicenseID    string `json:"licenseId"`
	OldExpiresAt string `json:"oldExpiresAt"`
	NewExpiresAt string `json:"newExpiresAt"`
	Edition      string `json:"edition"`
	MaxSites     int    `json:"maxSites"`
	MaxVehicles  int    `json:"maxVehicles"`
	Status       string `json:"status"`
	Operator     string `json:"operator"`
	Note         string `json:"note"`
	CreatedAt    string `json:"createdAt"`
}

type SystemAlert struct {
	ID             int64  `json:"id"`
	AlertNo        string `json:"alertNo"`
	CustomerID     int64  `json:"customerId"`
	Source         string `json:"source"`
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Status         string `json:"status"`
	FirstSeenAt    string `json:"firstSeenAt"`
	LastSeenAt     string `json:"lastSeenAt"`
	AcknowledgedAt string `json:"acknowledgedAt"`
	ResolvedAt     string `json:"resolvedAt"`
	Assignee       string `json:"assignee"`
	Resolution     string `json:"resolution"`
}

type UpdatePackage struct {
	ID           int64  `json:"id"`
	PackageNo    string `json:"packageNo"`
	Target       string `json:"target"`
	ProductName  string `json:"productName"`
	Version      string `json:"version"`
	Channel      string `json:"channel"`
	Status       string `json:"status"`
	FileName     string `json:"fileName"`
	Checksum     string `json:"checksum"`
	MinVersion   string `json:"minVersion"`
	RolloutPct   int    `json:"rolloutPct"`
	ReleaseNotes string `json:"releaseNotes"`
	UploadedAt   string `json:"uploadedAt"`
	PublishedAt  string `json:"publishedAt"`
}

type UpdateAssignment struct {
	ID          int64  `json:"id"`
	PackageID   int64  `json:"packageId"`
	CustomerID  int64  `json:"customerId"`
	Status      string `json:"status"`
	AssignedAt  string `json:"assignedAt"`
	DownloadedAt string `json:"downloadedAt"`
	AppliedAt   string `json:"appliedAt"`
	Error       string `json:"error"`
}

type AuditLog struct {
	ID        int64  `json:"id"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Detail    string `json:"detail"`
	CreatedAt string `json:"createdAt"`
}

type OperationsSummary struct {
	ActiveCustomers       int `json:"activeCustomers"`
	ExpiringLicenses      int `json:"expiringLicenses"`
	ExpiredLicenses       int `json:"expiredLicenses"`
	OpenAlerts            int `json:"openAlerts"`
	CriticalAlerts        int `json:"criticalAlerts"`
	ClientPackages        int `json:"clientPackages"`
	ServerPackages        int `json:"serverPackages"`
	PendingUpdateRollouts int `json:"pendingUpdateRollouts"`
}

type RenewLicenseRequest struct {
	ExpiresAt   string `json:"expiresAt"`
	Edition     string `json:"edition"`
	MaxSites    int    `json:"maxSites"`
	MaxVehicles int    `json:"maxVehicles"`
	Operator    string `json:"operator"`
	Note        string `json:"note"`
}

type CreateUpdatePackageRequest struct {
	Target       string `json:"target"`
	ProductName  string `json:"productName"`
	Version      string `json:"version"`
	Channel      string `json:"channel"`
	FileName     string `json:"fileName"`
	Checksum     string `json:"checksum"`
	MinVersion   string `json:"minVersion"`
	RolloutPct   int    `json:"rolloutPct"`
	ReleaseNotes string `json:"releaseNotes"`
}

type AssignUpdatePackageRequest struct {
	CustomerIDs []int64 `json:"customerIds"`
}

type ResolveAlertRequest struct {
	Operator   string `json:"operator"`
	Resolution string `json:"resolution"`
}
