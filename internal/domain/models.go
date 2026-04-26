package domain

import "time"

type HealthStatus string

const (
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type ServiceSource string

const (
	ServiceSourceManual ServiceSource = "manual"
	ServiceSourceDocker ServiceSource = "docker"
	ServiceSourceLAN    ServiceSource = "lan"
	ServiceSourceMDNS   ServiceSource = "mdns"
)

type ServiceAddressSource string

const (
	ServiceAddressLiteralHost   ServiceAddressSource = "literal_host"
	ServiceAddressDevicePrimary ServiceAddressSource = "device_primary_ip"
	ServiceAddressMDNSHostname  ServiceAddressSource = "mdns_hostname"
)

type DiscoveryState string

const (
	DiscoveryStatePending  DiscoveryState = "pending"
	DiscoveryStateAccepted DiscoveryState = "accepted"
	DiscoveryStateIgnored  DiscoveryState = "ignored"
)

type BookmarkAutomationPolicy string

const (
	BookmarkAutomationManual             BookmarkAutomationPolicy = "manual"
	BookmarkAutomationAutoHighConfidence BookmarkAutomationPolicy = "auto_high_confidence"
)

type CheckType string

const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
	CheckTypePing CheckType = "ping"
)

type HealthCheckSubjectType string

const (
	HealthCheckSubjectService           HealthCheckSubjectType = "service"
	HealthCheckSubjectDiscoveredService HealthCheckSubjectType = "discovered_service"
)

type HealthConfigMode string

const (
	HealthConfigModeAuto   HealthConfigMode = "auto"
	HealthConfigModeCustom HealthConfigMode = "custom"
)

type HealthCheckConfigSource string

const (
	HealthCheckConfigSourceDefinition HealthCheckConfigSource = "definition"
	HealthCheckConfigSourceUser       HealthCheckConfigSource = "user"
	HealthCheckConfigSourceMigrated   HealthCheckConfigSource = "migrated"
	HealthCheckConfigSourceFallback   HealthCheckConfigSource = "fallback"
)

type IdentityConfidence string

const (
	IdentityConfidenceLow  IdentityConfidence = "low"
	IdentityConfidenceHigh IdentityConfidence = "high"
)

type BootstrapStatus struct {
	Initialized bool `json:"initialized"`
}

type UIBootstrap struct {
	Initialized    bool   `json:"initialized"`
	TrustedNetwork bool   `json:"trustedNetwork"`
	CSRFToken      string `json:"csrfToken"`
}

type SetupInput struct {
	ApplianceName    string               `json:"applianceName"`
	AutoScanEnabled  bool                 `json:"autoScanEnabled"`
	DefaultScanPorts []int                `json:"defaultScanPorts"`
	DockerEndpoints  []DockerEndpointSeed `json:"dockerEndpoints"`
	ScanTargets      []ScanTargetSeed     `json:"scanTargets"`
	RunDiscovery     bool                 `json:"runDiscovery"`
}

type DockerEndpointSeed struct {
	Name                string `json:"name"`
	Kind                string `json:"kind"`
	Address             string `json:"address"`
	TLSCAPath           string `json:"tlsCaPath,omitempty"`
	TLSCertPath         string `json:"tlsCertPath,omitempty"`
	TLSKeyPath          string `json:"tlsKeyPath,omitempty"`
	Enabled             bool   `json:"enabled"`
	ScanIntervalSeconds int    `json:"scanIntervalSeconds"`
}

type ScanTargetSeed struct {
	Name                string `json:"name"`
	CIDR                string `json:"cidr"`
	AutoDetected        bool   `json:"autoDetected"`
	Enabled             bool   `json:"enabled"`
	ScanIntervalSeconds int    `json:"scanIntervalSeconds"`
	CommonPorts         []int  `json:"commonPorts"`
}

type AppSettings struct {
	Initialized               bool                     `json:"initialized"`
	AdminTokenHash            string                   `json:"-"`
	ApplianceName             string                   `json:"applianceName,omitempty"`
	InitializedAt             time.Time                `json:"initializedAt"`
	LastBootstrapAt           time.Time                `json:"lastBootstrapAt"`
	AutoScanEnabled           bool                     `json:"autoScanEnabled"`
	DefaultScanPorts          []int                    `json:"defaultScanPorts"`
	BookmarkPolicy            BookmarkAutomationPolicy `json:"bookmarkPolicy,omitempty"`
	AutoBookmarkSources       []ServiceSource          `json:"autoBookmarkSources,omitempty"`
	AutoBookmarkMinConfidence int                      `json:"autoBookmarkMinConfidence,omitempty"`
	UpdatedAt                 time.Time                `json:"updatedAt"`
	TrustedCIDRs              []string                 `json:"trustedCidrs,omitempty"`
	TrustedNetwork            bool                     `json:"trustedNetwork,omitempty"`
	LegacyTokenEnabled        bool                     `json:"legacyTokenEnabled,omitempty"`
}

type SettingsView struct {
	AppSettings        AppSettings         `json:"appSettings"`
	DockerEndpoints    []DockerEndpoint    `json:"dockerEndpoints"`
	ScanTargets        []ScanTarget        `json:"scanTargets"`
	JobState           []JobState          `json:"jobState"`
	APIAccess          APIAccessView       `json:"apiAccess"`
	Discovery          DiscoverySettings   `json:"discovery"`
	ServiceDefinitions []ServiceDefinition `json:"serviceDefinitions"`
}

type DashboardSummary struct {
	TotalServices      int `json:"totalServices"`
	HealthyServices    int `json:"healthyServices"`
	DegradedServices   int `json:"degradedServices"`
	UnhealthyServices  int `json:"unhealthyServices"`
	DevicesSeen        int `json:"devicesSeen"`
	Bookmarks          int `json:"bookmarks"`
	RunningContainers  int `json:"runningContainers"`
	DiscoveredServices int `json:"discoveredServices"`
}

type Dashboard struct {
	Summary            DashboardSummary    `json:"summary"`
	Services           []Service           `json:"services"`
	Containers         []Service           `json:"containers"`
	Devices            []Device            `json:"devices"`
	Bookmarks          []Bookmark          `json:"bookmarks"`
	DiscoveredServices []DiscoveredService `json:"discoveredServices"`
	RecentEvents       []ServiceEvent      `json:"recentEvents"`
}

type TokenScope string

const (
	TokenScopeRead  TokenScope = "read"
	TokenScopeWrite TokenScope = "write"
)

type APIToken struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Scope      TokenScope `json:"scope"`
	Prefix     string     `json:"prefix"`
	LastUsedAt time.Time  `json:"lastUsedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	RevokedAt  time.Time  `json:"revokedAt"`
}

type APIAccessView struct {
	Tokens                []APIToken `json:"tokens"`
	LegacyAdminTokenAlive bool       `json:"legacyAdminTokenAlive"`
}

type CreateAPITokenInput struct {
	Name  string     `json:"name"`
	Scope TokenScope `json:"scope"`
}

type CreatedAPIToken struct {
	Token  APIToken `json:"token"`
	Secret string   `json:"secret"`
}

type Service struct {
	ID                        string               `json:"id"`
	Name                      string               `json:"name"`
	Slug                      string               `json:"slug"`
	Source                    ServiceSource        `json:"source"`
	SourceRef                 string               `json:"sourceRef"`
	OriginDiscoveredServiceID string               `json:"originDiscoveredServiceId,omitempty"`
	ServiceDefinitionID       string               `json:"serviceDefinitionId,omitempty"`
	ServiceType               string               `json:"serviceType,omitempty"`
	HealthConfigMode          HealthConfigMode     `json:"healthConfigMode,omitempty"`
	AddressSource             ServiceAddressSource `json:"addressSource,omitempty"`
	HostValue                 string               `json:"hostValue,omitempty"`
	HealthAddressSource       ServiceAddressSource `json:"healthAddressSource,omitempty"`
	HealthHostValue           string               `json:"healthHostValue,omitempty"`
	DeviceID                  string               `json:"deviceId,omitempty"`
	DeviceName                string               `json:"deviceName,omitempty"`
	Icon                      string               `json:"icon,omitempty"`
	Scheme                    string               `json:"scheme,omitempty"`
	HealthScheme              string               `json:"healthScheme,omitempty"`
	Host                      string               `json:"host"`
	Port                      int                  `json:"port"`
	Path                      string               `json:"path,omitempty"`
	URL                       string               `json:"url"`
	HealthHost                string               `json:"healthHost,omitempty"`
	HealthPort                int                  `json:"healthPort,omitempty"`
	HealthPath                string               `json:"healthPath,omitempty"`
	HealthURL                 string               `json:"healthUrl,omitempty"`
	Hidden                    bool                 `json:"hidden"`
	Status                    HealthStatus         `json:"status"`
	LastSeenAt                time.Time            `json:"lastSeenAt"`
	LastCheckedAt             time.Time            `json:"lastCheckedAt"`
	FingerprintedAt           time.Time            `json:"fingerprintedAt"`
	Details                   map[string]any       `json:"details,omitempty"`
	CreatedAt                 time.Time            `json:"createdAt"`
	UpdatedAt                 time.Time            `json:"updatedAt"`
	Checks                    []ServiceCheck       `json:"checks,omitempty"`
}

type ServiceCheck struct {
	ID                  string                  `json:"id"`
	SubjectType         HealthCheckSubjectType  `json:"subjectType,omitempty"`
	SubjectID           string                  `json:"subjectId,omitempty"`
	ServiceID           string                  `json:"serviceId,omitempty"`
	Name                string                  `json:"name"`
	Type                CheckType               `json:"type"`
	Protocol            string                  `json:"protocol,omitempty"`
	AddressSource       ServiceAddressSource    `json:"addressSource,omitempty"`
	HostValue           string                  `json:"hostValue,omitempty"`
	Host                string                  `json:"host,omitempty"`
	Port                int                     `json:"port,omitempty"`
	Path                string                  `json:"path,omitempty"`
	Method              string                  `json:"method,omitempty"`
	Target              string                  `json:"target,omitempty"`
	IntervalSeconds     int                     `json:"intervalSeconds"`
	TimeoutSeconds      int                     `json:"timeoutSeconds"`
	ExpectedStatusMin   int                     `json:"expectedStatusMin,omitempty"`
	ExpectedStatusMax   int                     `json:"expectedStatusMax,omitempty"`
	Enabled             bool                    `json:"enabled"`
	SortOrder           int                     `json:"sortOrder,omitempty"`
	ConfigSource        HealthCheckConfigSource `json:"configSource,omitempty"`
	ServiceDefinitionID string                  `json:"serviceDefinitionId,omitempty"`
	CreatedAt           time.Time               `json:"createdAt"`
	UpdatedAt           time.Time               `json:"updatedAt"`
	LastResult          *CheckResult            `json:"lastResult,omitempty"`
}

type CheckResult struct {
	ID                string                 `json:"id"`
	CheckID           string                 `json:"checkId"`
	ServiceID         string                 `json:"serviceId,omitempty"`
	SubjectType       HealthCheckSubjectType `json:"subjectType,omitempty"`
	SubjectID         string                 `json:"subjectId,omitempty"`
	Status            HealthStatus           `json:"status"`
	LatencyMS         int64                  `json:"latencyMs"`
	HTTPStatusCode    int                    `json:"httpStatusCode,omitempty"`
	ResponseSizeBytes int64                  `json:"responseSizeBytes,omitempty"`
	ResolvedTarget    string                 `json:"resolvedTarget,omitempty"`
	Message           string                 `json:"message,omitempty"`
	CheckedAt         time.Time              `json:"checkedAt"`
}

type ServiceEvent struct {
	ID        string       `json:"id"`
	ServiceID string       `json:"serviceId"`
	EventType string       `json:"eventType"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message"`
	CreatedAt time.Time    `json:"createdAt"`
}

type Device struct {
	ID                 string             `json:"id"`
	IdentityKey        string             `json:"identityKey"`
	PrimaryMAC         string             `json:"primaryMac,omitempty"`
	Hostname           string             `json:"hostname,omitempty"`
	DisplayName        string             `json:"displayName,omitempty"`
	IdentityConfidence IdentityConfidence `json:"identityConfidence"`
	Hidden             bool               `json:"hidden"`
	FirstSeenAt        time.Time          `json:"firstSeenAt"`
	LastSeenAt         time.Time          `json:"lastSeenAt"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
	Addresses          []DeviceAddress    `json:"addresses,omitempty"`
	Ports              []DevicePort       `json:"ports,omitempty"`
}

type DeviceAddress struct {
	ID            string    `json:"id"`
	DeviceID      string    `json:"deviceId"`
	IPAddress     string    `json:"ipAddress"`
	MACAddress    string    `json:"macAddress,omitempty"`
	InterfaceName string    `json:"interfaceName,omitempty"`
	IsPrimary     bool      `json:"isPrimary"`
	FirstSeenAt   time.Time `json:"firstSeenAt"`
	LastSeenAt    time.Time `json:"lastSeenAt"`
}

type DevicePort struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"deviceId"`
	Port        int       `json:"port"`
	Protocol    string    `json:"protocol"`
	ServiceHint string    `json:"serviceHint,omitempty"`
	Open        bool      `json:"open"`
	FirstSeenAt time.Time `json:"firstSeenAt"`
	LastSeenAt  time.Time `json:"lastSeenAt"`
}

type DeviceObservation struct {
	IdentityKey string             `json:"identityKey"`
	PrimaryMAC  string             `json:"primaryMac,omitempty"`
	Hostname    string             `json:"hostname,omitempty"`
	DisplayName string             `json:"displayName,omitempty"`
	IPAddress   string             `json:"ipAddress,omitempty"`
	Interface   string             `json:"interface,omitempty"`
	Confidence  IdentityConfidence `json:"confidence"`
	Ports       []PortObservation  `json:"ports,omitempty"`
	LastSeenAt  time.Time          `json:"lastSeenAt"`
}

type PortObservation struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ServiceHint string `json:"serviceHint,omitempty"`
}

type ServiceObservation struct {
	Name            string               `json:"name"`
	Source          ServiceSource        `json:"source"`
	SourceRef       string               `json:"sourceRef"`
	DeviceKey       string               `json:"deviceKey,omitempty"`
	ServiceTypeHint string               `json:"serviceTypeHint,omitempty"`
	AddressSource   ServiceAddressSource `json:"addressSource,omitempty"`
	HostValue       string               `json:"hostValue,omitempty"`
	Icon            string               `json:"icon,omitempty"`
	Scheme          string               `json:"scheme,omitempty"`
	Host            string               `json:"host"`
	Port            int                  `json:"port"`
	Path            string               `json:"path,omitempty"`
	URL             string               `json:"url,omitempty"`
	LastSeenAt      time.Time            `json:"lastSeenAt"`
	Details         map[string]any       `json:"details,omitempty"`
}

type Observation struct {
	Device   DeviceObservation    `json:"device"`
	Services []ServiceObservation `json:"services,omitempty"`
}

type DiscoverySettings struct {
	BookmarkPolicy            BookmarkAutomationPolicy `json:"bookmarkPolicy"`
	AutoBookmarkSources       []ServiceSource          `json:"autoBookmarkSources"`
	AutoBookmarkMinConfidence int                      `json:"autoBookmarkMinConfidence"`
}

type DiscoveredService struct {
	ID                  string                   `json:"id"`
	DeviceID            string                   `json:"deviceId,omitempty"`
	DeviceName          string                   `json:"deviceName,omitempty"`
	MergeKey            string                   `json:"mergeKey"`
	Name                string                   `json:"name"`
	ServiceType         string                   `json:"serviceType,omitempty"`
	ConfidenceScore     int                      `json:"confidenceScore"`
	ServiceDefinitionID string                   `json:"serviceDefinitionId,omitempty"`
	AddressSource       ServiceAddressSource     `json:"addressSource,omitempty"`
	HostValue           string                   `json:"hostValue,omitempty"`
	Host                string                   `json:"host"`
	Scheme              string                   `json:"scheme,omitempty"`
	Port                int                      `json:"port"`
	Path                string                   `json:"path,omitempty"`
	URL                 string                   `json:"url"`
	Icon                string                   `json:"icon,omitempty"`
	State               DiscoveryState           `json:"state"`
	IgnoreFingerprint   string                   `json:"ignoreFingerprint,omitempty"`
	AutomationMode      BookmarkAutomationPolicy `json:"automationMode,omitempty"`
	HealthConfigMode    HealthConfigMode         `json:"healthConfigMode,omitempty"`
	Status              HealthStatus             `json:"status"`
	AcceptedServiceID   string                   `json:"acceptedServiceId,omitempty"`
	AcceptedBookmarkID  string                   `json:"acceptedBookmarkId,omitempty"`
	SourceTypes         []ServiceSource          `json:"sourceTypes,omitempty"`
	FirstSeenAt         time.Time                `json:"firstSeenAt"`
	LastSeenAt          time.Time                `json:"lastSeenAt"`
	LastCheckedAt       time.Time                `json:"lastCheckedAt"`
	LastFingerprintedAt time.Time                `json:"lastFingerprintedAt"`
	CreatedAt           time.Time                `json:"createdAt"`
	UpdatedAt           time.Time                `json:"updatedAt"`
	Details             map[string]any           `json:"details,omitempty"`
	Evidence            []DiscoveryEvidence      `json:"evidence,omitempty"`
}

type DiscoveryEvidence struct {
	ID                  string         `json:"id"`
	DiscoveredServiceID string         `json:"discoveredServiceId"`
	Source              ServiceSource  `json:"source"`
	SourceRef           string         `json:"sourceRef"`
	ServiceTypeHint     string         `json:"serviceTypeHint,omitempty"`
	Name                string         `json:"name,omitempty"`
	Host                string         `json:"host,omitempty"`
	Port                int            `json:"port,omitempty"`
	Path                string         `json:"path,omitempty"`
	URL                 string         `json:"url,omitempty"`
	FingerprintHash     string         `json:"fingerprintHash,omitempty"`
	Details             map[string]any `json:"details,omitempty"`
	FirstSeenAt         time.Time      `json:"firstSeenAt"`
	LastSeenAt          time.Time      `json:"lastSeenAt"`
}

type Bookmark struct {
	ID                      string        `json:"id"`
	Name                    string        `json:"name"`
	URL                     string        `json:"url"`
	Description             string        `json:"description,omitempty"`
	Icon                    string        `json:"icon,omitempty"`
	Tags                    []string      `json:"tags,omitempty"`
	FolderID                string        `json:"folderId,omitempty"`
	FolderName              string        `json:"folderName,omitempty"`
	ServiceID               string        `json:"serviceId,omitempty"`
	ServiceName             string        `json:"serviceName,omitempty"`
	ServiceSource           ServiceSource `json:"serviceSource,omitempty"`
	ServiceSourceRef        string        `json:"serviceSourceRef,omitempty"`
	ServiceHidden           bool          `json:"serviceHidden"`
	DeviceID                string        `json:"deviceId,omitempty"`
	DeviceName              string        `json:"deviceName,omitempty"`
	HealthStatus            HealthStatus  `json:"healthStatus"`
	IsFavorite              bool          `json:"isFavorite"`
	FavoritePosition        int           `json:"favoritePosition,omitempty"`
	Position                int           `json:"position"`
	ClickCount              int           `json:"clickCount"`
	LastOpenedAt            time.Time     `json:"lastOpenedAt"`
	ManualName              string        `json:"manualName,omitempty"`
	ManualURL               string        `json:"manualUrl,omitempty"`
	IconMode                string        `json:"iconMode,omitempty"`
	IconValue               string        `json:"iconValue,omitempty"`
	UseDevicePrimaryAddress bool          `json:"useDevicePrimaryAddress"`
	Scheme                  string        `json:"scheme,omitempty"`
	Host                    string        `json:"host,omitempty"`
	Port                    int           `json:"port,omitempty"`
	Path                    string        `json:"path,omitempty"`
	CreatedAt               time.Time     `json:"createdAt"`
	UpdatedAt               time.Time     `json:"updatedAt"`
}

type DockerEndpoint struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Kind                string    `json:"kind"`
	Address             string    `json:"address"`
	TLSCAPath           string    `json:"tlsCaPath,omitempty"`
	TLSCertPath         string    `json:"tlsCertPath,omitempty"`
	TLSKeyPath          string    `json:"tlsKeyPath,omitempty"`
	Enabled             bool      `json:"enabled"`
	ScanIntervalSeconds int       `json:"scanIntervalSeconds"`
	LastSuccessAt       time.Time `json:"lastSuccessAt"`
	LastError           string    `json:"lastError,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type ScanTarget struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	CIDR                string    `json:"cidr"`
	AutoDetected        bool      `json:"autoDetected"`
	Enabled             bool      `json:"enabled"`
	ScanIntervalSeconds int       `json:"scanIntervalSeconds"`
	CommonPorts         []int     `json:"commonPorts"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type JobState struct {
	JobName             string    `json:"jobName"`
	LastRunAt           time.Time `json:"lastRunAt"`
	LastSuccessAt       time.Time `json:"lastSuccessAt"`
	LastError           string    `json:"lastError,omitempty"`
	ConsecutiveFailures int       `json:"consecutiveFailures"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type MonitorCheck struct {
	Check   ServiceCheck `json:"check"`
	Service Service      `json:"service"`
}

type EventEnvelope struct {
	Type       string    `json:"type"`
	Resource   string    `json:"resource"`
	ID         string    `json:"id"`
	Action     string    `json:"action"`
	Payload    any       `json:"payload,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}
