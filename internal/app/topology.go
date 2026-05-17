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

type topologyDeviceBuild struct {
	item domain.TopologyDevice
}

type infrastructureBuild struct {
	nodes        []domain.TopologyInfrastructureNode
	edges        []domain.TopologyEdge
	sourceNodeID map[string]string
	sourceDepth  map[string]int
	lldpPorts    map[string]bool
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
	observations, err := a.store.ListTopologyObservations(ctx)
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

	deviceBuilds := make([]topologyDeviceBuild, 0, len(devices))
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
		deviceBuilds = append(deviceBuilds, topologyDeviceBuild{item: td})
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
	for _, built := range deviceBuilds {
		topology.Devices = append(topology.Devices, built.item)
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

	logicalEdges := make([]domain.TopologyEdge, 0, len(topology.Devices)+len(subnets))
	for i := range subnets {
		subnets[i].item.UtilizationPct = utilizationPct(subnets[i].item.DiscoveredAddressCount, subnets[i].item.UsableAddressCount)
		if subnets[i].item.ParentSubnetID == "" && subnets[i].valid && subnets[i].item.Enabled {
			router := domain.TopologyRouter{
				ID: fmt.Sprintf("router:%s", subnets[i].item.ID), Label: "Inferred gateway",
				Address: subnets[i].item.GatewayAddress, SubnetID: subnets[i].item.ID, GatewayInferred: true,
			}
			topology.Routers = append(topology.Routers, router)
			logicalEdges = append(logicalEdges, inferredAddressEdge(router.ID, subnets[i].item.ID, "router-subnet", "inferred gateway"))
		}
		if subnets[i].item.ParentSubnetID != "" {
			logicalEdges = append(logicalEdges, inferredAddressEdge(subnets[i].item.ParentSubnetID, subnets[i].item.ID, "subnet-subnet", "contains"))
		}
		topology.Subnets = append(topology.Subnets, subnets[i].item)
	}
	addressGroups, addressGroupEdges, deviceLogicalParents := buildTopologyAddressGroups(subnets, topology.Devices)
	topology.AddressGroups = addressGroups
	logicalEdges = append(logicalEdges, addressGroupEdges...)
	for _, device := range topology.Devices {
		parentID := firstNonEmpty(deviceLogicalParents[device.ID], device.SubnetID)
		kind := "subnet-device"
		if strings.HasPrefix(parentID, "addrgrp:") {
			kind = "address-group-device"
		}
		logicalEdges = append(logicalEdges, inferredAddressEdge(parentID, device.ID, kind, "inferred"))
	}

	infrastructure := buildTopologyInfrastructure(observations, topology.Subnets)
	topology.InfrastructureNodes = infrastructure.nodes
	observedDeviceEdges := buildInfrastructureDeviceEdges(topology.Devices, observations, infrastructure)
	topology.Edges = append(topology.Edges, infrastructure.edges...)
	topology.Edges = append(topology.Edges, observedDeviceEdges...)
	topology.Edges = append(topology.Edges, logicalEdges...)

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

func buildTopologyAddressGroups(subnets []topologySubnetBuild, devices []domain.TopologyDevice) ([]domain.TopologyAddressGroup, []domain.TopologyEdge, map[string]string) {
	devicesBySubnet := map[string][]domain.TopologyDevice{}
	for _, device := range devices {
		devicesBySubnet[device.SubnetID] = append(devicesBySubnet[device.SubnetID], device)
	}
	groups := []domain.TopologyAddressGroup{}
	edges := []domain.TopologyEdge{}
	deviceParents := map[string]string{}
	for _, subnet := range subnets {
		if !subnet.valid || !subnet.item.Enabled || !subnet.prefix.Addr().Is4() {
			continue
		}
		subnetDevices := devicesBySubnet[subnet.item.ID]
		if len(subnetDevices) == 0 {
			continue
		}
		bucketBits := addressGroupBucketBits(subnet.prefix, subnetDevices)
		if bucketBits == 0 {
			continue
		}
		bucketMembers := map[string][]domain.TopologyDevice{}
		bucketPrefixes := map[string]netip.Prefix{}
		for _, device := range subnetDevices {
			addr, ok := deviceAddressInPrefix(device, subnet.prefix)
			if !ok {
				continue
			}
			bucket := netip.PrefixFrom(addr, bucketBits).Masked()
			key := bucket.String()
			bucketPrefixes[key] = bucket
			bucketMembers[key] = append(bucketMembers[key], device)
		}
		keys := make([]string, 0, len(bucketMembers))
		for key := range bucketMembers {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return compareCIDRStrings(keys[i], keys[j]) })
		for _, key := range keys {
			group := buildTopologyAddressGroup(subnet.item.ID, bucketPrefixes[key], bucketMembers[key])
			groups = append(groups, group)
			edges = append(edges, inferredAddressEdge(subnet.item.ID, group.ID, "subnet-address-group", "contains"))
			for _, device := range bucketMembers[key] {
				deviceParents[device.ID] = group.ID
			}
		}
	}
	return groups, edges, deviceParents
}

func addressGroupBucketBits(prefix netip.Prefix, devices []domain.TopologyDevice) int {
	ones := prefix.Bits()
	switch {
	case ones < 24:
		return 24
	case ones == 24:
		if len(devices) > 12 && distinctDeviceBuckets(devices, prefix, 26) >= 2 {
			return 26
		}
	case ones >= 25 && ones <= 27:
		if len(devices) > 12 && distinctDeviceBuckets(devices, prefix, 28) >= 2 {
			return 28
		}
	}
	return 0
}

func distinctDeviceBuckets(devices []domain.TopologyDevice, parent netip.Prefix, bits int) int {
	buckets := map[string]struct{}{}
	for _, device := range devices {
		addr, ok := deviceAddressInPrefix(device, parent)
		if !ok {
			continue
		}
		buckets[netip.PrefixFrom(addr, bits).Masked().String()] = struct{}{}
	}
	return len(buckets)
}

func deviceAddressInPrefix(device domain.TopologyDevice, prefix netip.Prefix) (netip.Addr, bool) {
	addresses := append([]string{}, device.PrimaryAddress)
	addresses = append(addresses, device.Addresses...)
	for _, raw := range addresses {
		addr, err := netip.ParseAddr(raw)
		if err == nil && addr.Is4() && prefix.Contains(addr) {
			return addr, true
		}
	}
	return netip.Addr{}, false
}

func buildTopologyAddressGroup(subnetID string, prefix netip.Prefix, devices []domain.TopologyDevice) domain.TopologyAddressGroup {
	addressCount, usableCount, network, broadcast, first, last := ipv4PrefixRange(prefix)
	var addressCountSeen, serviceCount int
	for _, device := range devices {
		addressCountSeen += len(device.Addresses)
		serviceCount += device.ServiceCount
	}
	return domain.TopologyAddressGroup{
		ID:                     "addrgrp:" + subnetID + ":" + prefix.String(),
		SubnetID:               subnetID,
		Name:                   prefix.String(),
		CIDR:                   prefix.String(),
		Family:                 "ipv4",
		Depth:                  1,
		NetworkAddress:         network,
		BroadcastAddress:       broadcast,
		FirstUsableAddress:     first,
		LastUsableAddress:      last,
		AddressCount:           addressCount,
		UsableAddressCount:     usableCount,
		DiscoveredDeviceCount:  len(devices),
		DiscoveredAddressCount: addressCountSeen,
		ServiceCount:           serviceCount,
		UtilizationPct:         utilizationPct(addressCountSeen, usableCount),
	}
}

func ipv4PrefixRange(prefix netip.Prefix) (uint64, uint64, string, string, string, string) {
	ones := prefix.Bits()
	hostBits := 32 - ones
	addressCount := uint64(1) << hostBits
	usableCount := addressCount
	if ones <= 30 && addressCount >= 2 {
		usableCount = addressCount - 2
	}
	network := prefix.Masked().Addr()
	broadcast := addrFromUint32(addrToUint32(network) + uint32(addressCount) - 1)
	first := network
	last := broadcast
	if ones <= 30 {
		first = addrFromUint32(addrToUint32(network) + 1)
		last = addrFromUint32(addrToUint32(broadcast) - 1)
	}
	return addressCount, usableCount, network.String(), broadcast.String(), first.String(), last.String()
}

func inferredAddressEdge(sourceID, targetID, kind, label string) domain.TopologyEdge {
	return domain.TopologyEdge{
		ID:         edgeID(sourceID, targetID, kind+":"+label),
		SourceID:   sourceID,
		TargetID:   targetID,
		Kind:       kind,
		Label:      label,
		Source:     "address",
		Confidence: "low",
		Protocol:   "address",
		Inferred:   true,
	}
}

func buildTopologyInfrastructure(observations []domain.TopologySourceObservation, subnets []domain.TopologySubnet) infrastructureBuild {
	build := infrastructureBuild{
		sourceNodeID: map[string]string{},
		sourceDepth:  map[string]int{},
		lldpPorts:    map[string]bool{},
	}
	nodeByID := map[string]*domain.TopologyInfrastructureNode{}
	nodeByChassis := map[string]string{}
	nodeByAddress := map[string]string{}
	for _, obs := range observations {
		if len(obs.Interfaces) == 0 && len(obs.LLDPLinks) == 0 && len(obs.MACLinks) == 0 {
			continue
		}
		source := obs.Source
		chassisID := localChassisID(obs)
		systemName := localSystemName(obs)
		nodeID := infrastructureNodeID(firstNonEmpty(chassisID, source.ID, source.Address))
		node := &domain.TopologyInfrastructureNode{
			ID:                nodeID,
			SourceID:          source.ID,
			Kind:              firstNonEmpty(source.Role, "unknown"),
			Label:             firstNonEmpty(systemName, source.Name, source.Address, source.ID),
			ManagementAddress: source.Address,
			ChassisID:         chassisID,
			SystemName:        systemName,
			Role:              source.Role,
			Root:              source.Root,
			LastSeenAt:        obs.ObservedAt,
		}
		nodeByID[node.ID] = node
		build.sourceNodeID[source.ID] = node.ID
		if chassisID != "" {
			nodeByChassis[normalizeIdentity(chassisID)] = node.ID
		}
		if source.Address != "" {
			nodeByAddress[strings.ToLower(source.Address)] = node.ID
		}
	}

	type lldpEdge struct {
		a          string
		b          string
		aPort      string
		bPort      string
		label      string
		observedAt time.Time
	}
	edgeByKey := map[string]lldpEdge{}
	for _, obs := range observations {
		localID := build.sourceNodeID[obs.SourceID]
		if localID == "" {
			continue
		}
		for _, link := range obs.LLDPLinks {
			if link.LocalIfIndex > 0 {
				build.lldpPorts[portKey(obs.SourceID, link.LocalIfIndex, "")] = true
			}
			if link.LocalPortID != "" {
				build.lldpPorts[portKey(obs.SourceID, 0, link.LocalPortID)] = true
			}
			remoteID := ""
			if link.RemoteChassisID != "" {
				remoteID = nodeByChassis[normalizeIdentity(link.RemoteChassisID)]
			}
			if remoteID == "" && link.RemoteManagementAddress != "" {
				remoteID = nodeByAddress[strings.ToLower(link.RemoteManagementAddress)]
			}
			if remoteID == "" {
				remoteIdentity := firstNonEmpty(link.RemoteChassisID, link.RemoteManagementAddress, link.RemoteSystemName, link.RemotePortID)
				if remoteIdentity == "" {
					continue
				}
				remoteID = infrastructureNodeID(remoteIdentity)
				if nodeByID[remoteID] == nil {
					nodeByID[remoteID] = &domain.TopologyInfrastructureNode{
						ID:                remoteID,
						Kind:              "unknown",
						Label:             firstNonEmpty(link.RemoteSystemName, link.RemoteManagementAddress, link.RemoteChassisID, "Unknown infrastructure"),
						ManagementAddress: link.RemoteManagementAddress,
						ChassisID:         link.RemoteChassisID,
						SystemName:        link.RemoteSystemName,
						LastSeenAt:        link.LastSeenAt,
					}
				}
				if link.RemoteChassisID != "" {
					nodeByChassis[normalizeIdentity(link.RemoteChassisID)] = remoteID
				}
				if link.RemoteManagementAddress != "" {
					nodeByAddress[strings.ToLower(link.RemoteManagementAddress)] = remoteID
				}
			}
			if remoteID == localID {
				continue
			}
			key := lldpPairKey(link, localID, remoteID)
			if _, ok := edgeByKey[key]; ok {
				continue
			}
			edgeByKey[key] = lldpEdge{
				a:          localID,
				b:          remoteID,
				aPort:      firstNonEmpty(link.LocalPortName, link.LocalPortID, link.LocalPortDescription),
				bPort:      firstNonEmpty(link.RemotePortID, link.RemotePortDescription),
				label:      lldpEdgeLabel(link),
				observedAt: link.LastSeenAt,
			}
		}
	}

	roots := chooseInfrastructureRoots(observations, subnets, build.sourceNodeID)
	for _, rootID := range roots {
		if node := nodeByID[rootID]; node != nil {
			node.Root = true
		}
	}
	adjacency := map[string][]lldpEdge{}
	for _, edge := range edgeByKey {
		adjacency[edge.a] = append(adjacency[edge.a], edge)
		adjacency[edge.b] = append(adjacency[edge.b], lldpEdge{a: edge.b, b: edge.a, aPort: edge.bPort, bPort: edge.aPort, label: edge.label, observedAt: edge.observedAt})
	}
	for key := range adjacency {
		sort.Slice(adjacency[key], func(i, j int) bool {
			if adjacency[key][i].b == adjacency[key][j].b {
				return adjacency[key][i].label < adjacency[key][j].label
			}
			return adjacency[key][i].b < adjacency[key][j].b
		})
	}
	visited := map[string]bool{}
	parent := map[string]string{}
	depth := map[string]int{}
	crossSeen := map[string]bool{}
	bfs := func(root string) {
		if root == "" || visited[root] {
			return
		}
		queue := []string{root}
		visited[root] = true
		depth[root] = 0
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, edge := range adjacency[current] {
				if !visited[edge.b] {
					visited[edge.b] = true
					parent[edge.b] = current
					depth[edge.b] = depth[current] + 1
					queue = append(queue, edge.b)
					build.edges = append(build.edges, observedInfrastructureEdge(edge.a, edge.b, "infrastructure-link", edge.label, edge.observedAt))
					continue
				}
				if parent[current] == edge.b || parent[edge.b] == current {
					continue
				}
				key := sortedPair(edge.a, edge.b) + ":" + edge.label
				if crossSeen[key] {
					continue
				}
				crossSeen[key] = true
				build.edges = append(build.edges, observedInfrastructureEdge(edge.a, edge.b, "cross-link", edge.label, edge.observedAt))
			}
		}
	}
	for _, root := range roots {
		bfs(root)
	}
	ids := make([]string, 0, len(nodeByID))
	for id := range nodeByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		bfs(id)
	}
	for sourceID, nodeID := range build.sourceNodeID {
		build.sourceDepth[sourceID] = depth[nodeID]
	}
	for _, id := range ids {
		build.nodes = append(build.nodes, *nodeByID[id])
	}
	sort.Slice(build.edges, func(i, j int) bool {
		if build.edges[i].Kind == build.edges[j].Kind {
			return build.edges[i].ID < build.edges[j].ID
		}
		return build.edges[i].Kind < build.edges[j].Kind
	})
	return build
}

func buildInfrastructureDeviceEdges(devices []domain.TopologyDevice, observations []domain.TopologySourceObservation, infra infrastructureBuild) []domain.TopologyEdge {
	type candidate struct {
		link  domain.TopologyMACLinkObservation
		node  string
		depth int
		fan   int
		trunk bool
	}
	fanout := map[string]map[string]struct{}{}
	linksByMAC := map[string][]domain.TopologyMACLinkObservation{}
	for _, obs := range observations {
		for _, link := range obs.MACLinks {
			if link.MACAddress == "" {
				continue
			}
			key := portKey(link.SourceID, link.IfIndex, "")
			if fanout[key] == nil {
				fanout[key] = map[string]struct{}{}
			}
			mac := normalizeMAC(link.MACAddress)
			fanout[key][mac] = struct{}{}
			linksByMAC[mac] = append(linksByMAC[mac], link)
		}
	}
	edges := []domain.TopologyEdge{}
	for _, device := range devices {
		mac := normalizeMAC(device.PrimaryMAC)
		if mac == "" {
			continue
		}
		rawCandidates := linksByMAC[mac]
		if len(rawCandidates) == 0 {
			continue
		}
		candidates := make([]candidate, 0, len(rawCandidates))
		for _, link := range rawCandidates {
			nodeID := infra.sourceNodeID[link.SourceID]
			if nodeID == "" {
				continue
			}
			key := portKey(link.SourceID, link.IfIndex, "")
			trunk := infra.lldpPorts[key]
			if !trunk && link.IfIndex == 0 {
				trunk = infra.lldpPorts[portKey(link.SourceID, 0, link.IfName)]
			}
			candidates = append(candidates, candidate{link: link, node: nodeID, depth: infra.sourceDepth[link.SourceID], fan: len(fanout[key]), trunk: trunk})
		}
		if len(candidates) == 0 {
			continue
		}
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].depth != candidates[j].depth {
				return candidates[i].depth > candidates[j].depth
			}
			if candidates[i].trunk != candidates[j].trunk {
				return !candidates[i].trunk
			}
			if candidates[i].fan != candidates[j].fan {
				return candidates[i].fan < candidates[j].fan
			}
			if candidates[i].link.SourceID == candidates[j].link.SourceID {
				return candidates[i].link.IfIndex < candidates[j].link.IfIndex
			}
			return candidates[i].link.SourceID < candidates[j].link.SourceID
		})
		selected := candidates[0]
		confidence := "medium"
		if selected.fan > 8 && !selected.trunk {
			confidence = "low"
		}
		label := firstNonEmpty(selected.link.IfName, selected.link.IfDescription, fmt.Sprintf("port %d", selected.link.BridgePort))
		edges = append(edges, domain.TopologyEdge{
			ID:         edgeID(selected.node, device.ID, "infrastructure-port-device"),
			SourceID:   selected.node,
			TargetID:   device.ID,
			Kind:       "infrastructure-port-device",
			Label:      label,
			Source:     "bridge",
			Protocol:   "snmp",
			Confidence: confidence,
			ObservedAt: selected.link.LastSeenAt,
		})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return edges
}

func localChassisID(obs domain.TopologySourceObservation) string {
	for _, link := range obs.LLDPLinks {
		if link.LocalChassisID != "" {
			return link.LocalChassisID
		}
	}
	return ""
}

func localSystemName(obs domain.TopologySourceObservation) string {
	for _, link := range obs.LLDPLinks {
		if link.LocalSystemName != "" {
			return link.LocalSystemName
		}
	}
	return ""
}

func chooseInfrastructureRoots(observations []domain.TopologySourceObservation, subnets []domain.TopologySubnet, sourceNodeID map[string]string) []string {
	var roots []string
	for _, obs := range observations {
		if obs.Source.Root && sourceNodeID[obs.SourceID] != "" {
			roots = append(roots, sourceNodeID[obs.SourceID])
		}
	}
	if len(roots) > 0 {
		return uniqueStrings(roots)
	}
	gateways := map[string]struct{}{}
	for _, subnet := range subnets {
		if subnet.GatewayAddress != "" {
			gateways[subnet.GatewayAddress] = struct{}{}
		}
	}
	for _, obs := range observations {
		if _, ok := gateways[obs.Source.Address]; ok && sourceNodeID[obs.SourceID] != "" {
			roots = append(roots, sourceNodeID[obs.SourceID])
		}
	}
	if len(roots) > 0 {
		return uniqueStrings(roots)
	}
	type sourceRoot struct {
		address string
		nodeID  string
	}
	candidates := []sourceRoot{}
	for _, obs := range observations {
		if !obs.Source.Enabled || sourceNodeID[obs.SourceID] == "" {
			continue
		}
		candidates = append(candidates, sourceRoot{address: obs.Source.Address, nodeID: sourceNodeID[obs.SourceID]})
	}
	sort.Slice(candidates, func(i, j int) bool { return compareIPStrings(candidates[i].address, candidates[j].address) })
	if len(candidates) > 0 {
		return []string{candidates[0].nodeID}
	}
	return nil
}

func observedInfrastructureEdge(sourceID, targetID, kind, label string, observedAt time.Time) domain.TopologyEdge {
	return domain.TopologyEdge{
		ID:         edgeID(sourceID, targetID, kind+":"+label),
		SourceID:   sourceID,
		TargetID:   targetID,
		Kind:       kind,
		Label:      label,
		Source:     "lldp",
		Confidence: "high",
		Protocol:   "lldp",
		ObservedAt: observedAt,
	}
}

func lldpEdgeLabel(link domain.TopologyLLDPLinkObservation) string {
	local := firstNonEmpty(link.LocalPortName, link.LocalPortID, link.LocalPortDescription)
	remote := firstNonEmpty(link.RemotePortID, link.RemotePortDescription)
	if local != "" && remote != "" {
		return local + " <-> " + remote
	}
	return firstNonEmpty(local, remote, "LLDP")
}

func lldpPairKey(link domain.TopologyLLDPLinkObservation, localNodeID, remoteNodeID string) string {
	local := normalizeIdentity(firstNonEmpty(link.LocalChassisID, localNodeID)) + "|" + normalizeIdentity(link.LocalPortID)
	remote := normalizeIdentity(firstNonEmpty(link.RemoteChassisID, remoteNodeID)) + "|" + normalizeIdentity(link.RemotePortID)
	if local > remote {
		local, remote = remote, local
	}
	return local + "||" + remote
}

func infrastructureNodeID(identity string) string {
	return "infra:" + safeID(identity)
}

func normalizeIdentity(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeMAC(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", ":")
	value = strings.ReplaceAll(value, ".", "")
	value = strings.ReplaceAll(value, " ", "")
	if strings.Count(value, ":") == 5 {
		parts := strings.Split(value, ":")
		for i, part := range parts {
			if len(part) == 1 {
				parts[i] = "0" + part
			}
		}
		return strings.Join(parts, ":")
	}
	value = strings.ReplaceAll(value, ":", "")
	if len(value) != 12 {
		return value
	}
	parts := make([]string, 0, 6)
	for i := 0; i < 12; i += 2 {
		parts = append(parts, value[i:i+2])
	}
	return strings.Join(parts, ":")
}

func portKey(sourceID string, ifIndex int, portID string) string {
	if ifIndex > 0 {
		return fmt.Sprintf("%s:%d", sourceID, ifIndex)
	}
	return sourceID + ":" + strings.ToLower(strings.TrimSpace(portID))
}

func edgeID(sourceID, targetID, kind string) string {
	return safeID(kind + ":" + sourceID + "->" + targetID)
}

func safeID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", ".", "-", ">", "-", "<", "-", "|", "-", "@", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "unknown"
	}
	return value
}

func sortedPair(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	var items []string
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	sort.Strings(items)
	return items
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

func compareCIDRStrings(a, b string) bool {
	aa, aerr := netip.ParsePrefix(a)
	bb, berr := netip.ParsePrefix(b)
	if aerr == nil && berr == nil && aa.Addr().Is4() && bb.Addr().Is4() {
		if addrToUint32(aa.Addr()) == addrToUint32(bb.Addr()) {
			return aa.Bits() < bb.Bits()
		}
		return addrToUint32(aa.Addr()) < addrToUint32(bb.Addr())
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
