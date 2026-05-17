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
	GeneratedAt time.Time         `json:"generatedAt"`
	Summary     TopologySummary   `json:"summary"`
	Routers     []TopologyRouter  `json:"routers"`
	Subnets     []TopologySubnet  `json:"subnets"`
	Devices     []TopologyDevice  `json:"devices"`
	Services    []TopologyService `json:"services"`
	Edges       []TopologyEdge    `json:"edges"`
	Warnings    []string          `json:"warnings,omitempty"`
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
	ID       string `json:"id"`
	SourceID string `json:"sourceId"`
	TargetID string `json:"targetId"`
	Kind     string `json:"kind"`
}
