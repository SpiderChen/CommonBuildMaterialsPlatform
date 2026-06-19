package serverupdater

type Config struct {
	BaseURL               string                     `json:"baseUrl"`
	UpdaterToken          string                     `json:"updaterToken"`
	Watermark             string                     `json:"watermark"`
	RootDir               string                     `json:"rootDir"`
	TargetComponent       string                     `json:"targetComponent"`
	PollIntervalSeconds   int                        `json:"pollIntervalSeconds"`
	StrictChecksum        bool                       `json:"strictChecksum"`
	ResumeDownloads       bool                       `json:"resumeDownloads"`
	AutoRollbackOnFailure bool                       `json:"autoRollbackOnFailure"`
	Components            map[string]ComponentConfig `json:"components"`
}

type ComponentConfig struct {
	StopCommand     []string `json:"stopCommand"`
	InstallCommand  []string `json:"installCommand"`
	StartCommand    []string `json:"startCommand"`
	HealthCommand   []string `json:"healthCommand"`
	TimeoutSeconds  int      `json:"timeoutSeconds"`
	ExpectedService string   `json:"expectedService"`
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

type PollResponse struct {
	Accepted bool                      `json:"accepted"`
	Instance ProductInstance           `json:"instance"`
	Tasks    []ProductSystemUpdateTask `json:"tasks"`
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

type UpdatePackage struct {
	ID                      int64  `json:"id"`
	Version                 string `json:"version"`
	Component               string `json:"component"`
	Channel                 string `json:"channel"`
	Status                  string `json:"status"`
	PackageType             string `json:"packageType"`
	BaseVersion             string `json:"baseVersion"`
	DeltaAlgorithm          string `json:"deltaAlgorithm"`
	Checksum                string `json:"checksum"`
	Signature               string `json:"signature"`
	SignaturePublicKey      string `json:"signaturePublicKey"`
	SignatureKeyFingerprint string `json:"signatureKeyFingerprint"`
	FileName                string `json:"fileName"`
	SizeBytes               int64  `json:"sizeBytes"`
	ArtifactFileName        string `json:"artifactFileName"`
	ArtifactContentType     string `json:"artifactContentType"`
	ArtifactSHA256          string `json:"artifactSha256"`
	ArtifactSizeBytes       int64  `json:"artifactSizeBytes"`
	BaseArtifactSHA256      string `json:"baseArtifactSha256"`
	TargetArtifactSHA256    string `json:"targetArtifactSha256"`
	RollbackVersion         string `json:"rollbackVersion"`
	PublishedBy             string `json:"publishedBy"`
	PublishedAt             string `json:"publishedAt"`
	DownloadCount           int64  `json:"downloadCount"`
	LastDownloadedAt        string `json:"lastDownloadedAt"`
	Remark                  string `json:"remark"`
}

type ReportRequest struct {
	UpdaterToken   string `json:"updaterToken"`
	Status         string `json:"status"`
	Progress       int    `json:"progress"`
	Step           string `json:"step"`
	Message        string `json:"message"`
	Error          string `json:"error,omitempty"`
	CurrentVersion string `json:"currentVersion,omitempty"`
	UpdaterVersion string `json:"updaterVersion"`
}

type State struct {
	Components map[string]ComponentState `json:"components"`
}

type ComponentState struct {
	CurrentVersion  string `json:"currentVersion"`
	PreviousVersion string `json:"previousVersion"`
	ActiveSlot      string `json:"activeSlot"`
	CurrentRelease  string `json:"currentRelease"`
	PreviousRelease string `json:"previousRelease"`
	LastTaskNo      string `json:"lastTaskNo"`
	UpdatedAt       string `json:"updatedAt"`
}
