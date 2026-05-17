package app

import (
	"context"
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
