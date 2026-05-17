package app

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/domain"
)

func TestTopologyBuildsIPv4SubnetAndServices(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	target, err := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "LAN", CIDR: "192.168.1.0/24", Enabled: true})
	if err != nil {
		t.Fatalf("save target: %v", err)
	}
	device, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{
		IdentityKey: "mac:aa", PrimaryMAC: "aa", Hostname: "nas", IPAddress: "192.168.1.20",
		Confidence: domain.IdentityConfidenceHigh, Ports: []domain.PortObservation{{Port: 22, Protocol: "tcp"}}, LastSeenAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("save device: %v", err)
	}
	if _, err := application.SaveManualService(ctx, domain.Service{Name: "SSH", DeviceID: device.ID, Host: "192.168.1.20", Port: 22, Scheme: "tcp"}); err != nil {
		t.Fatalf("save service: %v", err)
	}

	topology, err := application.Topology(ctx)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	var subnet domain.TopologySubnet
	for _, candidate := range topology.Subnets {
		if candidate.ScanTargetID == target.ID {
			subnet = candidate
		}
	}
	if subnet.ID == "" {
		t.Fatalf("expected scan target subnet, got %#v", topology.Subnets)
	}
	if subnet.NetworkAddress != "192.168.1.0" || subnet.BroadcastAddress != "192.168.1.255" || subnet.FirstUsableAddress != "192.168.1.1" || subnet.LastUsableAddress != "192.168.1.254" {
		t.Fatalf("unexpected IPv4 range: %#v", subnet)
	}
	if subnet.AddressCount != 256 || subnet.UsableAddressCount != 254 || subnet.DiscoveredDeviceCount != 1 || subnet.ServiceCount != 1 {
		t.Fatalf("unexpected subnet counts: %#v", subnet)
	}
	var router domain.TopologyRouter
	for _, candidate := range topology.Routers {
		if candidate.SubnetID == subnet.ID {
			router = candidate
		}
	}
	if router.ID == "" || !router.GatewayInferred || router.Address != "192.168.1.1" {
		t.Fatalf("expected inferred gateway, got %#v", topology.Routers)
	}
	if len(topology.Devices) != 1 || topology.Devices[0].SubnetID != subnet.ID || topology.Devices[0].ServiceCount != 1 || len(topology.Services) != 1 {
		t.Fatalf("unexpected device/service topology: %#v %#v", topology.Devices, topology.Services)
	}
}

func TestTopologyAssignsNarrowestSubnetAndUnmapped(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	parent, _ := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "LAN", CIDR: "10.0.0.0/16", Enabled: true})
	child, _ := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "Rack", CIDR: "10.0.5.0/24", Enabled: true})
	if _, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{IdentityKey: "mac:bb", IPAddress: "10.0.5.20", Confidence: domain.IdentityConfidenceHigh, LastSeenAt: time.Now().UTC()}); err != nil {
		t.Fatalf("save child device: %v", err)
	}
	if _, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{IdentityKey: "mac:cc", IPAddress: "172.16.0.9", Confidence: domain.IdentityConfidenceLow, LastSeenAt: time.Now().UTC()}); err != nil {
		t.Fatalf("save unmapped device: %v", err)
	}

	topology, err := application.Topology(ctx)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	var parentSubnet, childSubnet domain.TopologySubnet
	for _, subnet := range topology.Subnets {
		if subnet.ScanTargetID == parent.ID {
			parentSubnet = subnet
		}
		if subnet.ScanTargetID == child.ID {
			childSubnet = subnet
		}
	}
	if childSubnet.ParentSubnetID != parentSubnet.ID {
		t.Fatalf("expected child parent %q, got %q", parentSubnet.ID, childSubnet.ParentSubnetID)
	}
	for _, device := range topology.Devices {
		if device.PrimaryAddress == "10.0.5.20" && device.SubnetID != childSubnet.ID {
			t.Fatalf("expected narrowest subnet assignment, got %q", device.SubnetID)
		}
	}
	if topology.Summary.UnmappedDeviceCount != 1 {
		t.Fatalf("expected one unmapped device, got %#v", topology.Summary)
	}
}

func TestTopologyWarnsForUnsupportedTargets(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "Broken", CIDR: "not-a-cidr", Enabled: false}); err != nil {
		t.Fatalf("save invalid target: %v", err)
	}
	if _, err := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "IPv6", CIDR: "fd00::/64", Enabled: true}); err != nil {
		t.Fatalf("save ipv6 target: %v", err)
	}
	topology, err := application.Topology(ctx)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	if topology.Summary.UnsupportedSubnetCount != 2 || len(topology.Warnings) != 2 {
		t.Fatalf("expected warnings for invalid and IPv6 targets, got %#v %#v", topology.Summary, topology.Warnings)
	}
}

func TestTopologyBuildsLogicalAddressGroups(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "Corp", CIDR: "10.0.0.0/16", Enabled: true}); err != nil {
		t.Fatalf("save target: %v", err)
	}
	for _, address := range []string{"10.0.5.20", "10.0.9.30"} {
		if _, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{IdentityKey: "ip:" + address, IPAddress: address, Confidence: domain.IdentityConfidenceHigh, LastSeenAt: time.Now().UTC()}); err != nil {
			t.Fatalf("save device %s: %v", address, err)
		}
	}
	topology, err := application.Topology(ctx)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	var cidrs []string
	for _, group := range topology.AddressGroups {
		cidrs = append(cidrs, group.CIDR)
	}
	if len(cidrs) != 2 || cidrs[0] != "10.0.5.0/24" || cidrs[1] != "10.0.9.0/24" {
		t.Fatalf("expected occupied /24 groups, got %#v", topology.AddressGroups)
	}
}

func TestTopologySplitsBusy24IntoOccupied26Groups(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "LAN", CIDR: "192.168.1.0/24", Enabled: true}); err != nil {
		t.Fatalf("save target: %v", err)
	}
	for i := 0; i < 13; i++ {
		host := 10 + i
		if i > 6 {
			host = 80 + i
		}
		address := "192.168.1." + strconv.Itoa(host)
		if _, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{IdentityKey: "ip:" + address, IPAddress: address, Confidence: domain.IdentityConfidenceHigh, LastSeenAt: time.Now().UTC()}); err != nil {
			t.Fatalf("save device %s: %v", address, err)
		}
	}
	topology, err := application.Topology(ctx)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	var cidrs []string
	for _, group := range topology.AddressGroups {
		cidrs = append(cidrs, group.CIDR)
	}
	if len(cidrs) != 2 || cidrs[0] != "192.168.1.0/26" || cidrs[1] != "192.168.1.64/26" {
		t.Fatalf("expected occupied /26 groups, got %#v", topology.AddressGroups)
	}
}

func TestTopologyBuildsObservedInfrastructureAndMapsDeviceToDeepestSwitch(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{DefaultScanPorts: []int{22, 80}})
	ctx := context.Background()
	if err := application.Setup(ctx, domain.SetupInput{ApplianceName: "Lab", DefaultScanPorts: []int{22, 80}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := application.SaveScanTarget(ctx, domain.ScanTarget{Name: "LAN", CIDR: "192.168.1.0/24", Enabled: true}); err != nil {
		t.Fatalf("save target: %v", err)
	}
	device, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{IdentityKey: "mac:aa:bb:cc:dd:ee:ff", PrimaryMAC: "aa-bb-cc-dd-ee-ff", IPAddress: "192.168.1.20", Confidence: domain.IdentityConfidenceHigh, LastSeenAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("save device: %v", err)
	}
	core, err := store.SaveTopologySource(ctx, domain.TopologySource{Name: "Core", Address: "192.168.1.1", Enabled: true, SNMPVersion: "v2c", Community: "public", Role: "switch", Root: true})
	if err != nil {
		t.Fatalf("save core: %v", err)
	}
	access, err := store.SaveTopologySource(ctx, domain.TopologySource{Name: "Access", Address: "192.168.1.2", Enabled: true, SNMPVersion: "v2c", Community: "public", Role: "switch"})
	if err != nil {
		t.Fatalf("save access: %v", err)
	}
	dist, err := store.SaveTopologySource(ctx, domain.TopologySource{Name: "Distribution", Address: "192.168.1.3", Enabled: true, SNMPVersion: "v2c", Community: "public", Role: "switch"})
	if err != nil {
		t.Fatalf("save distribution: %v", err)
	}
	seen := time.Now().UTC()
	if err := store.ReplaceTopologyObservations(ctx, core.ID, domain.TopologySourceObservation{
		ObservedAt: seen,
		LLDPLinks: []domain.TopologyLLDPLinkObservation{
			{LocalChassisID: "core", LocalSystemName: "core", LocalPortID: "gi1", LocalIfIndex: 1, RemoteChassisID: "access", RemoteSystemName: "access", RemotePortID: "gi48"},
		},
		MACLinks: []domain.TopologyMACLinkObservation{{MACAddress: "aa:bb:cc:dd:ee:ff", IfIndex: 1, IfName: "gi1"}},
	}); err != nil {
		t.Fatalf("replace core observations: %v", err)
	}
	if err := store.ReplaceTopologyObservations(ctx, access.ID, domain.TopologySourceObservation{
		ObservedAt: seen,
		LLDPLinks: []domain.TopologyLLDPLinkObservation{
			{LocalChassisID: "access", LocalSystemName: "access", LocalPortID: "gi48", LocalIfIndex: 48, RemoteChassisID: "core", RemoteSystemName: "core", RemotePortID: "gi1"},
			{LocalChassisID: "access", LocalSystemName: "access", LocalPortID: "gi47", LocalIfIndex: 47, RemoteChassisID: "dist", RemoteSystemName: "dist", RemotePortID: "gi1"},
		},
		MACLinks: []domain.TopologyMACLinkObservation{{MACAddress: "aa:bb:cc:dd:ee:ff", IfIndex: 5, IfName: "gi5"}},
	}); err != nil {
		t.Fatalf("replace access observations: %v", err)
	}
	if err := store.ReplaceTopologyObservations(ctx, dist.ID, domain.TopologySourceObservation{
		ObservedAt: seen,
		LLDPLinks: []domain.TopologyLLDPLinkObservation{
			{LocalChassisID: "dist", LocalSystemName: "dist", LocalPortID: "gi1", LocalIfIndex: 1, RemoteChassisID: "core", RemoteSystemName: "core", RemotePortID: "gi2"},
		},
	}); err != nil {
		t.Fatalf("replace dist observations: %v", err)
	}

	topology, err := application.Topology(ctx)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	if len(topology.InfrastructureNodes) != 3 {
		t.Fatalf("expected three infrastructure nodes, got %#v", topology.InfrastructureNodes)
	}
	var infraLinks, crossLinks int
	var deviceEdge domain.TopologyEdge
	for _, edge := range topology.Edges {
		switch edge.Kind {
		case "infrastructure-link":
			infraLinks++
		case "cross-link":
			crossLinks++
		case "infrastructure-port-device":
			if edge.TargetID == "device:"+device.ID {
				deviceEdge = edge
			}
		}
	}
	if infraLinks != 2 || crossLinks != 1 {
		t.Fatalf("expected deduped LLDP tree with one cross-link, got infra=%d cross=%d edges=%#v", infraLinks, crossLinks, topology.Edges)
	}
	accessNodeID := ""
	for _, node := range topology.InfrastructureNodes {
		if node.SourceID == access.ID {
			accessNodeID = node.ID
		}
	}
	if deviceEdge.SourceID != accessNodeID || deviceEdge.Source != "bridge" || deviceEdge.Protocol != "snmp" {
		t.Fatalf("expected device attached to deepest access switch, got %#v accessNode=%q", deviceEdge, accessNodeID)
	}
}
