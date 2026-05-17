package domain

import "time"

type TopologySummary struct {
	RouterCount             int `json:"routerCount"`
	SubnetCount             int `json:"subnetCount"`
	DeviceCount             int `json:"deviceCount"`
	HiddenDeviceCount       int `json:"hiddenDeviceCount"`
	ServiceCount            int `json:"serviceCount"`
	UnsupportedSubnetCount  int `json:"unsupportedSubnetCount"`
	UnmappedDeviceCount     int `json:"unmappedDeviceCount"`
	DiscoveredAddressCount  int `json:"discoveredAddressCount"`
	DiscoveredOpenPortCount int `json:"discoveredOpenPortCount"`
}

type NetworkTopology struct {
	GeneratedAt         time.Time                    `json:"generatedAt"`
	Summary             TopologySummary              `json:"summary"`
	Routers             []TopologyRouter             `json:"routers"`
	Subnets             []TopologySubnet             `json:"subnets"`
	AddressGroups       []TopologyAddressGroup       `json:"addressGroups,omitempty"`
	InfrastructureNodes []TopologyInfrastructureNode `json:"infrastructureNodes,omitempty"`
	Devices             []TopologyDevice             `json:"devices"`
	Services            []TopologyService            `json:"services"`
	Edges               []TopologyEdge               `json:"edges"`
	Warnings            []string                     `json:"warnings,omitempty"`
}

type TopologyRouter struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Address         string `json:"address,omitempty"`
	SubnetID        string `json:"subnetId,omitempty"`
	GatewayInferred bool   `json:"gatewayInferred"`
}

type TopologySubnet struct {
	ID                     string   `json:"id"`
	ScanTargetID           string   `json:"scanTargetId,omitempty"`
	Name                   string   `json:"name"`
	CIDR                   string   `json:"cidr"`
	Family                 string   `json:"family"`
	Enabled                bool     `json:"enabled"`
	AutoDetected           bool     `json:"autoDetected"`
	ParentSubnetID         string   `json:"parentSubnetId,omitempty"`
	ChildSubnetIDs         []string `json:"childSubnetIds,omitempty"`
	NetworkAddress         string   `json:"networkAddress,omitempty"`
	BroadcastAddress       string   `json:"broadcastAddress,omitempty"`
	FirstUsableAddress     string   `json:"firstUsableAddress,omitempty"`
	LastUsableAddress      string   `json:"lastUsableAddress,omitempty"`
	AddressCount           uint64   `json:"addressCount"`
	UsableAddressCount     uint64   `json:"usableAddressCount"`
	DiscoveredDeviceCount  int      `json:"discoveredDeviceCount"`
	DiscoveredAddressCount int      `json:"discoveredAddressCount"`
	ServiceCount           int      `json:"serviceCount"`
	UtilizationPct         float64  `json:"utilizationPct"`
	GatewayAddress         string   `json:"gatewayAddress,omitempty"`
	GatewayInferred        bool     `json:"gatewayInferred"`
	Warnings               []string `json:"warnings,omitempty"`
}

type TopologyDevice struct {
	ID                 string             `json:"id"`
	SubnetID           string             `json:"subnetId"`
	Label              string             `json:"label"`
	Hostname           string             `json:"hostname,omitempty"`
	PrimaryMAC         string             `json:"primaryMac,omitempty"`
	PrimaryAddress     string             `json:"primaryAddress,omitempty"`
	Addresses          []string           `json:"addresses,omitempty"`
	OpenPorts          []int              `json:"openPorts,omitempty"`
	ServiceCount       int                `json:"serviceCount"`
	Hidden             bool               `json:"hidden"`
	IdentityConfidence IdentityConfidence `json:"identityConfidence"`
	LastSeenAt         time.Time          `json:"lastSeenAt"`
}

type TopologyAddressGroup struct {
	ID                     string  `json:"id"`
	SubnetID               string  `json:"subnetId"`
	ParentGroupID          string  `json:"parentGroupId,omitempty"`
	Name                   string  `json:"name"`
	CIDR                   string  `json:"cidr"`
	Family                 string  `json:"family"`
	Depth                  int     `json:"depth"`
	NetworkAddress         string  `json:"networkAddress"`
	BroadcastAddress       string  `json:"broadcastAddress"`
	FirstUsableAddress     string  `json:"firstUsableAddress"`
	LastUsableAddress      string  `json:"lastUsableAddress"`
	AddressCount           uint64  `json:"addressCount"`
	UsableAddressCount     uint64  `json:"usableAddressCount"`
	DiscoveredDeviceCount  int     `json:"discoveredDeviceCount"`
	DiscoveredAddressCount int     `json:"discoveredAddressCount"`
	ServiceCount           int     `json:"serviceCount"`
	UtilizationPct         float64 `json:"utilizationPct"`
}

type TopologyInfrastructureNode struct {
	ID                string    `json:"id"`
	SourceID          string    `json:"sourceId,omitempty"`
	Kind              string    `json:"kind"`
	Label             string    `json:"label"`
	ManagementAddress string    `json:"managementAddress,omitempty"`
	ChassisID         string    `json:"chassisId,omitempty"`
	SystemName        string    `json:"systemName,omitempty"`
	SystemDescription string    `json:"systemDescription,omitempty"`
	Role              string    `json:"role,omitempty"`
	Root              bool      `json:"root,omitempty"`
	LastSeenAt        time.Time `json:"lastSeenAt,omitempty"`
}

type TopologyService struct {
	ID       string        `json:"id"`
	DeviceID string        `json:"deviceId"`
	Name     string        `json:"name"`
	URL      string        `json:"url,omitempty"`
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Status   HealthStatus  `json:"status"`
	Source   ServiceSource `json:"source"`
	Icon     string        `json:"icon,omitempty"`
}

type TopologyEdge struct {
	ID         string    `json:"id"`
	SourceID   string    `json:"sourceId"`
	TargetID   string    `json:"targetId"`
	Kind       string    `json:"kind"`
	Label      string    `json:"label,omitempty"`
	Source     string    `json:"source,omitempty"`
	Confidence string    `json:"confidence,omitempty"`
	Protocol   string    `json:"protocol,omitempty"`
	ObservedAt time.Time `json:"observedAt,omitempty"`
	Inferred   bool      `json:"inferred,omitempty"`
}

type TopologySource struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Address              string    `json:"address"`
	Port                 int       `json:"port"`
	Enabled              bool      `json:"enabled"`
	PollIntervalSeconds  int       `json:"pollIntervalSeconds"`
	TimeoutMS            int       `json:"timeoutMs"`
	Retries              int       `json:"retries"`
	SNMPVersion          string    `json:"snmpVersion"`
	Community            string    `json:"community,omitempty"`
	HasCommunity         bool      `json:"hasCommunity,omitempty"`
	Username             string    `json:"username,omitempty"`
	AuthProtocol         string    `json:"authProtocol,omitempty"`
	AuthPassphrase       string    `json:"authPassphrase,omitempty"`
	HasAuthPassphrase    bool      `json:"hasAuthPassphrase,omitempty"`
	PrivacyProtocol      string    `json:"privacyProtocol,omitempty"`
	PrivacyPassphrase    string    `json:"privacyPassphrase,omitempty"`
	HasPrivacyPassphrase bool      `json:"hasPrivacyPassphrase,omitempty"`
	Role                 string    `json:"role"`
	Root                 bool      `json:"root"`
	LastSuccessAt        time.Time `json:"lastSuccessAt,omitempty"`
	LastError            string    `json:"lastError,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type TopologyAutoDiscoverInput struct {
	Community string `json:"community,omitempty"`
}

type TopologyAutoDiscoverResult struct {
	CandidateCount int              `json:"candidateCount"`
	TestedCount    int              `json:"testedCount"`
	Added          []TopologySource `json:"added"`
	Existing       []TopologySource `json:"existing,omitempty"`
	FailedCount    int              `json:"failedCount"`
}

type TopologySourceObservation struct {
	SourceID   string
	Source     TopologySource
	Interfaces []TopologyInterfaceObservation
	LLDPLinks  []TopologyLLDPLinkObservation
	MACLinks   []TopologyMACLinkObservation
	ObservedAt time.Time
}

type TopologyInterfaceObservation struct {
	ID            string
	SourceID      string
	IfIndex       int
	IfName        string
	IfDescription string
	IfAlias       string
	IfType        int
	OperStatus    string
	SpeedBPS      uint64
	LastSeenAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TopologyLLDPLinkObservation struct {
	ID                      string
	SourceID                string
	LocalChassisID          string
	LocalSystemName         string
	LocalPortID             string
	LocalPortName           string
	LocalPortDescription    string
	LocalIfIndex            int
	RemoteChassisID         string
	RemoteSystemName        string
	RemotePortID            string
	RemotePortDescription   string
	RemoteManagementAddress string
	LastSeenAt              time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type TopologyMACLinkObservation struct {
	ID            string
	SourceID      string
	MACAddress    string
	VLAN          int
	BridgePort    int
	IfIndex       int
	IfName        string
	IfDescription string
	Status        string
	LastSeenAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
