package ops

type AppData struct {
	SchemaVersion  int64                `json:"schemaVersion"`
	Next           map[string]int64     `json:"next"`
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
	UpdaterToken         string   `json:"updaterToken"`
	Edition              string   `json:"edition"`
	DeploymentMode       string   `json:"deploymentMode"`
	Environment          string   `json:"environment"`
	ServerEndpoint       string   `json:"serverEndpoint"`
	ContactName          string   `json:"contactName"`
	ContactPhone         string   `json:"contactPhone"`
	ExpiresAt            string   `json:"expiresAt"`
	RenewalStatus        string   `json:"renewalStatus"`
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
	ID                   int64    `json:"id"`
	RenewalNo            string   `json:"renewalNo"`
	CustomerID           int64    `json:"customerId"`
	LicenseID            string   `json:"licenseId"`
	OldExpiresAt         string   `json:"oldExpiresAt"`
	NewExpiresAt         string   `json:"newExpiresAt"`
	Edition              string   `json:"edition"`
	Modules              []string `json:"modules"`
	MaxSites             int      `json:"maxSites"`
	MaxVehicles          int      `json:"maxVehicles"`
	LicensePackageNo     string   `json:"licensePackageNo"`
	Watermark            string   `json:"watermark"`
	Issuer               string   `json:"issuer"`
	IssuedAt             string   `json:"issuedAt"`
	PublicKey            string   `json:"publicKey,omitempty"`
	PublicKeyFingerprint string   `json:"publicKeyFingerprint,omitempty"`
	Signature            string   `json:"signature,omitempty"`
	DownloadCount        int64    `json:"downloadCount"`
	LastDownloadedAt     string   `json:"lastDownloadedAt"`
	Status               string   `json:"status"`
	Operator             string   `json:"operator"`
	Note                 string   `json:"note"`
	CreatedAt            string   `json:"createdAt"`
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
	ID                    int64  `json:"id"`
	PackageNo             string `json:"packageNo"`
	Target                string `json:"target"`
	Component             string `json:"component"`
	ProductName           string `json:"productName"`
	Version               string `json:"version"`
	Channel               string `json:"channel"`
	Status                string `json:"status"`
	PackageType           string `json:"packageType"`
	FileName              string `json:"fileName"`
	Checksum              string `json:"checksum"`
	Signature             string `json:"signature"`
	ArtifactFileName      string `json:"artifactFileName"`
	ArtifactContentType   string `json:"artifactContentType"`
	ArtifactSizeBytes     int64  `json:"artifactSizeBytes"`
	ArtifactSHA256        string `json:"artifactSha256"`
	ArtifactContentBase64 string `json:"artifactContentBase64,omitempty"`
	MinVersion            string `json:"minVersion"`
	RolloutPct            int    `json:"rolloutPct"`
	ReleaseNotes          string `json:"releaseNotes"`
	UploadedAt            string `json:"uploadedAt"`
	PublishedAt           string `json:"publishedAt"`
	DownloadCount         int64  `json:"downloadCount"`
	LastDownloadedAt      string `json:"lastDownloadedAt"`
}

type UpdateAssignment struct {
	ID             int64  `json:"id"`
	PackageID      int64  `json:"packageId"`
	CustomerID     int64  `json:"customerId"`
	Status         string `json:"status"`
	AssignedAt     string `json:"assignedAt"`
	DownloadedAt   string `json:"downloadedAt"`
	AppliedAt      string `json:"appliedAt"`
	Error          string `json:"error"`
	Progress       int    `json:"progress"`
	Step           string `json:"step"`
	Message        string `json:"message"`
	UpdatedAt      string `json:"updatedAt"`
	UpdaterVersion string `json:"updaterVersion"`
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

type CreateCustomerDeploymentRequest struct {
	CustomerName         string   `json:"customerName"`
	ProductName          string   `json:"productName"`
	LicenseID            string   `json:"licenseId"`
	UpdaterToken         string   `json:"updaterToken"`
	Edition              string   `json:"edition"`
	DeploymentMode       string   `json:"deploymentMode"`
	Environment          string   `json:"environment"`
	ServerEndpoint       string   `json:"serverEndpoint"`
	ContactName          string   `json:"contactName"`
	ContactPhone         string   `json:"contactPhone"`
	ExpiresAt            string   `json:"expiresAt"`
	Modules              []string `json:"modules"`
	MaxSites             int      `json:"maxSites"`
	MaxVehicles          int      `json:"maxVehicles"`
	CurrentClientVersion string   `json:"currentClientVersion"`
	CurrentServerVersion string   `json:"currentServerVersion"`
	TargetClientVersion  string   `json:"targetClientVersion"`
	TargetServerVersion  string   `json:"targetServerVersion"`
	Notes                string   `json:"notes"`
	Operator             string   `json:"operator"`
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
	Target                string `json:"target"`
	ProductName           string `json:"productName"`
	Version               string `json:"version"`
	Channel               string `json:"channel"`
	FileName              string `json:"fileName"`
	Checksum              string `json:"checksum"`
	Signature             string `json:"signature"`
	ArtifactContentType   string `json:"artifactContentType"`
	ArtifactContentBase64 string `json:"artifactContentBase64"`
	MinVersion            string `json:"minVersion"`
	RolloutPct            int    `json:"rolloutPct"`
	ReleaseNotes          string `json:"releaseNotes"`
}

type AssignUpdatePackageRequest struct {
	CustomerIDs []int64 `json:"customerIds"`
}

type ResolveAlertRequest struct {
	Operator   string `json:"operator"`
	Resolution string `json:"resolution"`
}

type CreateAlertRequest struct {
	CustomerID   int64  `json:"customerId"`
	UpdaterToken string `json:"updaterToken"`
	Source       string `json:"source"`
	Severity     string `json:"severity"`
	Title        string `json:"title"`
	Message      string `json:"message"`
	Assignee     string `json:"assignee"`
	Operator     string `json:"operator"`
	OccurredAt   string `json:"occurredAt"`
	AutoResolve  bool   `json:"autoResolve"`
	Resolution   string `json:"resolution"`
}

type UpdateTaskPollRequest struct {
	UpdaterToken string `json:"updaterToken"`
	Watermark    string `json:"watermark"`
}

type ProductInstance struct {
	ID              int64  `json:"id"`
	CustomerName    string `json:"customerName"`
	Watermark       string `json:"watermark"`
	ClientVersion   string `json:"clientVersion"`
	ServerVersion   string `json:"serverVersion"`
	Endpoint        string `json:"endpoint"`
	ProbeToken      string `json:"probeToken"`
	HealthStatus    string `json:"healthStatus"`
	LastHeartbeatAt string `json:"lastHeartbeatAt"`
}

type ProductSystemUpdateTask struct {
	ID               int64  `json:"id"`
	TaskNo           string `json:"taskNo"`
	ExecutionID      int64  `json:"executionId"`
	ExecutionNo      string `json:"executionNo"`
	RolloutID        int64  `json:"rolloutId"`
	RolloutNo        string `json:"rolloutNo"`
	UpdateID         int64  `json:"updateId"`
	InstanceID       int64  `json:"instanceId"`
	CustomerName     string `json:"customerName"`
	Watermark        string `json:"watermark"`
	Component        string `json:"component"`
	Version          string `json:"version"`
	FromVersion      string `json:"fromVersion"`
	Action           string `json:"action"`
	Status           string `json:"status"`
	Progress         int    `json:"progress"`
	ArtifactFileName string `json:"artifactFileName"`
	Checksum         string `json:"checksum"`
	Signature        string `json:"signature"`
	DownloadURL      string `json:"downloadUrl"`
	CreatedAt        string `json:"createdAt"`
	Remark           string `json:"remark"`
}

type UpdateTaskPollResponse struct {
	Accepted bool                      `json:"accepted"`
	Instance ProductInstance           `json:"instance"`
	Tasks    []ProductSystemUpdateTask `json:"tasks"`
}

type UpdateTaskReportRequest struct {
	UpdaterToken   string `json:"updaterToken"`
	Status         string `json:"status"`
	Progress       int    `json:"progress"`
	Step           string `json:"step"`
	Message        string `json:"message"`
	Error          string `json:"error,omitempty"`
	CurrentVersion string `json:"currentVersion,omitempty"`
	UpdaterVersion string `json:"updaterVersion"`
}

type UpdatePackageDownload struct {
	FileName              string            `json:"fileName"`
	ContentType           string            `json:"contentType"`
	Verified              bool              `json:"verified"`
	GeneratedAt           string            `json:"generatedAt"`
	ArtifactFileName      string            `json:"artifactFileName"`
	ArtifactContentType   string            `json:"artifactContentType"`
	ArtifactSizeBytes     int64             `json:"artifactSizeBytes"`
	ArtifactSHA256        string            `json:"artifactSha256"`
	ArtifactContentBase64 string            `json:"artifactContentBase64"`
	Manifest              map[string]string `json:"manifest"`
	Package               UpdatePackage     `json:"package"`
}

type LicensePackage struct {
	ID                   int64    `json:"id"`
	LicenseID            string   `json:"licenseId"`
	CustomerName         string   `json:"customerName"`
	Watermark            string   `json:"watermark"`
	ExpiresAt            string   `json:"expiresAt"`
	Edition              string   `json:"edition"`
	Modules              []string `json:"modules"`
	MaxSites             int      `json:"maxSites"`
	MaxVehicles          int      `json:"maxVehicles"`
	IssuedAt             string   `json:"issuedAt"`
	Issuer               string   `json:"issuer"`
	PublicKey            string   `json:"publicKey"`
	PublicKeyFingerprint string   `json:"publicKeyFingerprint"`
	Signature            string   `json:"signature"`
	Status               string   `json:"status"`
}

type LicensePackageDownload struct {
	FileName      string             `json:"fileName"`
	ContentType   string             `json:"contentType"`
	GeneratedAt   string             `json:"generatedAt"`
	ContentBase64 string             `json:"contentBase64"`
	Customer      CustomerDeployment `json:"customer"`
	Renewal       LicenseRenewal     `json:"renewal"`
	Package       LicensePackage     `json:"package"`
}
