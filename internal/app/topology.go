package app

import (
	"context"
	"fmt"
	"math"
	"net/netip"
	"sort"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

const unmappedSubnetID = "subnet:unmapped"

type topologySubnetBuild struct {
	item   domain.TopologySubnet
	prefix netip.Prefix
	valid  bool
}

func (a *App) Topology(ctx context.Context) (domain.NetworkTopology, error) {
	targets, err := a.store.ListScanTargets(ctx)
	if err != nil {
		return domain.NetworkTopology{}, err
	}
	devices, err := a.store.ListDevices(ctx)
	if err != nil {
		return domain.NetworkTopology{}, err
	}
	services, err := a.store.ListServices(ctx)
	if err != nil {
		return domain.NetworkTopology{}, err
	}

	topology := domain.NetworkTopology{GeneratedAt: time.Now().UTC()}
	subnets := make([]topologySubnetBuild, 0, len(targets)+1)
	for _, target := range targets {
		built := buildTopologySubnet(target)
		if len(built.item.Warnings) > 0 {
			topology.Warnings = append(topology.Warnings, built.item.Warnings...)
		}
		subnets = append(subnets, built)
	}
	sort.Slice(subnets, func(i, j int) bool {
		if subnets[i].item.CIDR == subnets[j].item.CIDR {
			return subnets[i].item.Name < subnets[j].item.Name
		}
		return subnets[i].item.CIDR < subnets[j].item.CIDR
	})
	assignSubnetParents(subnets)

	serviceCounts := map[string]int{}
	for _, service := range services {
		if service.DeviceID != "" {
			serviceCounts[service.DeviceID]++
			topology.Services = append(topology.Services, domain.TopologyService{
				ID: service.ID, DeviceID: "device:" + service.DeviceID, Name: service.Name, URL: service.URL,
				Host: service.Host, Port: service.Port, Status: service.Status, Source: service.Source, Icon: service.Icon,
			})
		}
	}

	subnetIndex := map[string]int{}
	for i := range subnets {
		subnetIndex[subnets[i].item.ID] = i
	}
	unmappedNeeded := false
	for _, device := range devices {
		subnetID := matchDeviceSubnet(device, subnets)
		if subnetID == "" {
			subnetID = unmappedSubnetID
			unmappedNeeded = true
			topology.Summary.UnmappedDeviceCount++
		}
		td := buildTopologyDevice(device, subnetID, serviceCounts[device.ID])
		topology.Devices = append(topology.Devices, td)
		if device.Hidden {
			topology.Summary.HiddenDeviceCount++
		}
		if idx, ok := subnetIndex[subnetID]; ok {
			subnets[idx].item.DiscoveredDeviceCount++
			subnets[idx].item.DiscoveredAddressCount += len(td.Addresses)
			subnets[idx].item.ServiceCount += td.ServiceCount
		}
		topology.Summary.DiscoveredAddressCount += len(td.Addresses)
		topology.Summary.DiscoveredOpenPortCount += len(td.OpenPorts)
	}
	sort.Slice(topology.Devices, func(i, j int) bool {
		return compareTopologyDevices(topology.Devices[i], topology.Devices[j])
	})

	if unmappedNeeded {
		var unmappedDevices, unmappedAddresses, unmappedServices int
		for _, device := range topology.Devices {
			if device.SubnetID == unmappedSubnetID {
				unmappedDevices++
				unmappedAddresses += len(device.Addresses)
				unmappedServices += device.ServiceCount
			}
		}
		subnets = append(subnets, topologySubnetBuild{
			item: domain.TopologySubnet{
				ID: unmappedSubnetID, Name: "Unmapped devices", Family: "mixed", Enabled: true,
				DiscoveredDeviceCount:  unmappedDevices,
				DiscoveredAddressCount: unmappedAddresses,
				ServiceCount:           unmappedServices,
				Warnings:               []string{"Devices in this group did not match any configured scan target."},
			},
		})
	}

	for i := range subnets {
		subnets[i].item.UtilizationPct = utilizationPct(subnets[i].item.DiscoveredAddressCount, subnets[i].item.UsableAddressCount)
		if subnets[i].item.ParentSubnetID == "" && subnets[i].valid && subnets[i].item.Enabled {
			router := domain.TopologyRouter{
				ID: fmt.Sprintf("router:%s", subnets[i].item.ID), Label: "Inferred gateway",
				Address: subnets[i].item.GatewayAddress, SubnetID: subnets[i].item.ID, GatewayInferred: true,
			}
			topology.Routers = append(topology.Routers, router)
			topology.Edges = append(topology.Edges, domain.TopologyEdge{ID: router.ID + "->" + subnets[i].item.ID, SourceID: router.ID, TargetID: subnets[i].item.ID, Kind: "router-subnet"})
		}
		if subnets[i].item.ParentSubnetID != "" {
			topology.Edges = append(topology.Edges, domain.TopologyEdge{ID: subnets[i].item.ParentSubnetID + "->" + subnets[i].item.ID, SourceID: subnets[i].item.ParentSubnetID, TargetID: subnets[i].item.ID, Kind: "subnet-subnet"})
		}
		topology.Subnets = append(topology.Subnets, subnets[i].item)
	}
	for _, device := range topology.Devices {
		topology.Edges = append(topology.Edges, domain.TopologyEdge{ID: device.SubnetID + "->" + device.ID, SourceID: device.SubnetID, TargetID: device.ID, Kind: "subnet-device"})
	}

	topology.Summary.RouterCount = len(topology.Routers)
	topology.Summary.SubnetCount = len(topology.Subnets)
	topology.Summary.DeviceCount = len(topology.Devices)
	topology.Summary.ServiceCount = len(topology.Services)
	for _, subnet := range topology.Subnets {
		if subnet.Family != "ipv4" && subnet.ID != unmappedSubnetID {
			topology.Summary.UnsupportedSubnetCount++
		}
	}
	return topology, nil
}

func buildTopologySubnet(target domain.ScanTarget) topologySubnetBuild {
	name := strings.TrimSpace(target.Name)
	if name == "" {
		name = target.CIDR
	}
	item := domain.TopologySubnet{ID: "subnet:" + target.ID, ScanTargetID: target.ID, Name: name, CIDR: target.CIDR, Enabled: target.Enabled, AutoDetected: target.AutoDetected}
	prefix, err := netip.ParsePrefix(strings.TrimSpace(target.CIDR))
	if err != nil {
		item.Family = "unsupported"
		item.Warnings = append(item.Warnings, fmt.Sprintf("Scan target %q has invalid CIDR %q.", name, target.CIDR))
		return topologySubnetBuild{item: item}
	}
	prefix = prefix.Masked()
	item.CIDR = prefix.String()
	if !prefix.Addr().Is4() {
		item.Family = "ipv6"
		item.Warnings = append(item.Warnings, fmt.Sprintf("Scan target %q uses IPv6; topology math is limited to IPv4.", name))
		return topologySubnetBuild{item: item, prefix: prefix}
	}
	item.Family = "ipv4"
	ones := prefix.Bits()
	hostBits := 32 - ones
	item.AddressCount = uint64(1) << hostBits
	item.UsableAddressCount = item.AddressCount
	if ones <= 30 && item.AddressCount >= 2 {
		item.UsableAddressCount = item.AddressCount - 2
	}
	network := prefix.Addr()
	broadcast := addrFromUint32(addrToUint32(network) + uint32(item.AddressCount) - 1)
	item.NetworkAddress = network.String()
	item.BroadcastAddress = broadcast.String()
	if ones <= 30 {
		item.FirstUsableAddress = addrFromUint32(addrToUint32(network) + 1).String()
		item.LastUsableAddress = addrFromUint32(addrToUint32(broadcast) - 1).String()
		item.GatewayAddress = item.FirstUsableAddress
		item.GatewayInferred = true
	} else {
		item.FirstUsableAddress = network.String()
		item.LastUsableAddress = broadcast.String()
		item.GatewayAddress = network.String()
		item.GatewayInferred = true
	}
	return topologySubnetBuild{item: item, prefix: prefix, valid: true}
}

func assignSubnetParents(subnets []topologySubnetBuild) {
	for i := range subnets {
		if !subnets[i].valid {
			continue
		}
		best := -1
		for j := range subnets {
			if i == j || !subnets[j].valid || !subnets[j].prefix.Addr().Is4() || subnets[j].prefix.Bits() >= subnets[i].prefix.Bits() {
				continue
			}
			if subnets[j].prefix.Contains(subnets[i].prefix.Addr()) && (best == -1 || subnets[j].prefix.Bits() > subnets[best].prefix.Bits()) {
				best = j
			}
		}
		if best >= 0 {
			subnets[i].item.ParentSubnetID = subnets[best].item.ID
			subnets[best].item.ChildSubnetIDs = append(subnets[best].item.ChildSubnetIDs, subnets[i].item.ID)
		}
	}
}

func matchDeviceSubnet(device domain.Device, subnets []topologySubnetBuild) string {
	best := -1
	for _, address := range device.Addresses {
		addr, err := netip.ParseAddr(address.IPAddress)
		if err != nil || !addr.Is4() {
			continue
		}
		for i := range subnets {
			if !subnets[i].valid || !subnets[i].item.Enabled || !subnets[i].prefix.Contains(addr) {
				continue
			}
			if best == -1 || subnets[i].prefix.Bits() > subnets[best].prefix.Bits() {
				best = i
			}
		}
	}
	if best == -1 {
		return ""
	}
	return subnets[best].item.ID
}

func buildTopologyDevice(device domain.Device, subnetID string, serviceCount int) domain.TopologyDevice {
	addresses := make([]string, 0, len(device.Addresses))
	primary := ""
	for _, address := range device.Addresses {
		addresses = append(addresses, address.IPAddress)
		if address.IsPrimary {
			primary = address.IPAddress
		}
	}
	sort.Slice(addresses, func(i, j int) bool { return compareIPStrings(addresses[i], addresses[j]) })
	if primary == "" && len(addresses) > 0 {
		primary = addresses[0]
	}
	ports := make([]int, 0, len(device.Ports))
	for _, port := range device.Ports {
		if port.Open {
			ports = append(ports, port.Port)
		}
	}
	sort.Ints(ports)
	label := firstNonEmpty(strings.TrimSpace(device.DisplayName), strings.TrimSpace(device.Hostname), primary, device.PrimaryMAC, device.ID)
	return domain.TopologyDevice{ID: "device:" + device.ID, SubnetID: subnetID, Label: label, Hostname: device.Hostname, PrimaryMAC: device.PrimaryMAC, PrimaryAddress: primary, Addresses: addresses, OpenPorts: ports, ServiceCount: serviceCount, Hidden: device.Hidden, IdentityConfidence: device.IdentityConfidence, LastSeenAt: device.LastSeenAt}
}

func compareTopologyDevices(a, b domain.TopologyDevice) bool {
	if a.SubnetID != b.SubnetID {
		return a.SubnetID < b.SubnetID
	}
	if a.PrimaryAddress != "" && b.PrimaryAddress != "" && a.PrimaryAddress != b.PrimaryAddress {
		return compareIPStrings(a.PrimaryAddress, b.PrimaryAddress)
	}
	return a.Label < b.Label
}

func compareIPStrings(a, b string) bool {
	aa, aerr := netip.ParseAddr(a)
	bb, berr := netip.ParseAddr(b)
	if aerr == nil && berr == nil && aa.Is4() && bb.Is4() {
		return addrToUint32(aa) < addrToUint32(bb)
	}
	return a < b
}

func utilizationPct(discovered int, usable uint64) float64 {
	if usable == 0 {
		return 0
	}
	return math.Round((float64(discovered)/float64(usable))*10000) / 100
}

func addrToUint32(addr netip.Addr) uint32 {
	bytes := addr.As4()
	return uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
}

func addrFromUint32(value uint32) netip.Addr {
	return netip.AddrFrom4([4]byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)})
}
