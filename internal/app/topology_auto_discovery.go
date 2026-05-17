package app

import (
	"context"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

const maxTopologyAutoProbeCandidates = 96

type topologyProbeCandidate struct {
	address string
	root    bool
}

func (a *App) AutoDiscoverTopologySources(ctx context.Context, input domain.TopologyAutoDiscoverInput) (domain.TopologyAutoDiscoverResult, error) {
	if !a.isBootstrapped(ctx) {
		return domain.TopologyAutoDiscoverResult{}, nil
	}
	targets, err := a.store.ListScanTargets(ctx)
	if err != nil {
		return domain.TopologyAutoDiscoverResult{}, err
	}
	devices, err := a.store.ListDevices(ctx)
	if err != nil {
		return domain.TopologyAutoDiscoverResult{}, err
	}
	existing, err := a.store.ListTopologySourcesForDiscovery(ctx)
	if err != nil {
		return domain.TopologyAutoDiscoverResult{}, err
	}

	candidates := buildTopologyProbeCandidates(targets, devices)
	result := domain.TopologyAutoDiscoverResult{CandidateCount: len(candidates)}
	existingByAddress := map[string]domain.TopologySource{}
	for _, source := range existing {
		key := topologySourceAddressKey(source.Address, source.Port)
		existingByAddress[key] = source
	}
	communities := topologyProbeCommunities(input, existing)
	for _, candidate := range candidates {
		if source, ok := existingByAddress[topologySourceAddressKey(candidate.address, 161)]; ok {
			result.Existing = appendUniqueTopologySource(result.Existing, redactTopologySource(source))
			continue
		}
		var lastErr error
		var probeSource domain.TopologySource
		var probeOK bool
		for _, community := range communities {
			source := domain.TopologySource{
				ID:                  "probe:" + safeID(candidate.address),
				Name:                "Auto " + candidate.address,
				Address:             candidate.address,
				Port:                161,
				Enabled:             true,
				PollIntervalSeconds: 300,
				TimeoutMS:           700,
				Retries:             1,
				SNMPVersion:         "v2c",
				Community:           community,
				Role:                "unknown",
				Root:                candidate.root,
			}
			probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			probe, err := a.topologySNMP.Probe(probeCtx, source)
			cancel()
			result.TestedCount++
			if err != nil {
				lastErr = err
				continue
			}
			source.Name = firstNonEmpty(probe.SystemName, "Network device "+candidate.address)
			source.Role = probe.Role
			if source.Role == "unknown" && candidate.root {
				source.Role = "router"
			}
			probeSource = source
			probeOK = true
			break
		}
		if !probeOK {
			result.FailedCount++
			if lastErr != nil {
				a.logger.Debug("topology auto-discovery probe failed", "address", candidate.address, "err", lastErr)
			}
			continue
		}
		saved, err := a.store.SaveTopologySource(ctx, probeSource)
		if err != nil {
			return domain.TopologyAutoDiscoverResult{}, err
		}
		result.Added = append(result.Added, saved)
		existingByAddress[topologySourceAddressKey(saved.Address, saved.Port)] = saved
		a.publish("topology-source", saved.ID, "created", saved)
	}
	a.publish("topology", "all", "updated", nil)
	return result, nil
}

func (a *App) AutoDiscoverAndRunTopology(ctx context.Context, input domain.TopologyAutoDiscoverInput) (domain.TopologyAutoDiscoverResult, error) {
	result, err := a.AutoDiscoverTopologySources(ctx, input)
	if err != nil {
		return domain.TopologyAutoDiscoverResult{}, err
	}
	if err := a.runTopologyDiscovery(ctx); err != nil {
		return result, err
	}
	return result, nil
}

func buildTopologyProbeCandidates(targets []domain.ScanTarget, devices []domain.Device) []topologyProbeCandidate {
	type candidateState struct {
		candidate topologyProbeCandidate
		order     int
	}
	seen := map[string]candidateState{}
	add := func(address string, root bool) {
		addr, err := netip.ParseAddr(strings.TrimSpace(address))
		if err != nil || !addr.Is4() {
			return
		}
		key := addr.String()
		state, ok := seen[key]
		if ok {
			state.candidate.root = state.candidate.root || root
			seen[key] = state
			return
		}
		seen[key] = candidateState{candidate: topologyProbeCandidate{address: key, root: root}, order: len(seen)}
	}
	for _, target := range targets {
		if !target.Enabled {
			continue
		}
		prefix, err := netip.ParsePrefix(strings.TrimSpace(target.CIDR))
		if err != nil || !prefix.Addr().Is4() {
			continue
		}
		prefix = prefix.Masked()
		add(inferredGatewayAddress(prefix), true)
		addPrefixOffset(prefix, 1, false, add)
		addPrefixOffset(prefix, 2, false, add)
		addPrefixOffset(prefix, -2, false, add)
		addPrefixOffset(prefix, -3, false, add)
	}
	deviceAddresses := make([]string, 0, len(devices))
	for _, device := range devices {
		for _, address := range device.Addresses {
			deviceAddresses = append(deviceAddresses, address.IPAddress)
		}
	}
	sort.Slice(deviceAddresses, func(i, j int) bool { return compareIPStrings(deviceAddresses[i], deviceAddresses[j]) })
	for _, address := range deviceAddresses {
		add(address, false)
	}
	items := make([]candidateState, 0, len(seen))
	for _, state := range seen {
		items = append(items, state)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].candidate.root != items[j].candidate.root {
			return items[i].candidate.root
		}
		if items[i].order != items[j].order {
			return items[i].order < items[j].order
		}
		return compareIPStrings(items[i].candidate.address, items[j].candidate.address)
	})
	limit := len(items)
	if limit > maxTopologyAutoProbeCandidates {
		limit = maxTopologyAutoProbeCandidates
	}
	candidates := make([]topologyProbeCandidate, 0, limit)
	for i := 0; i < limit; i++ {
		candidates = append(candidates, items[i].candidate)
	}
	return candidates
}

func inferredGatewayAddress(prefix netip.Prefix) string {
	if prefix.Bits() <= 30 {
		return addrFromUint32(addrToUint32(prefix.Addr()) + 1).String()
	}
	return prefix.Addr().String()
}

func addPrefixOffset(prefix netip.Prefix, offset int64, root bool, add func(string, bool)) {
	ones := prefix.Bits()
	if ones < 0 || ones > 32 {
		return
	}
	size := uint64(1) << (32 - ones)
	if size == 0 {
		return
	}
	var index uint64
	if offset >= 0 {
		index = uint64(offset)
	} else {
		back := uint64(-offset)
		if back > size {
			return
		}
		index = size - back
	}
	if index >= size {
		return
	}
	addr := addrFromUint32(addrToUint32(prefix.Addr()) + uint32(index))
	if prefix.Contains(addr) {
		add(addr.String(), root)
	}
}

func topologyProbeCommunities(input domain.TopologyAutoDiscoverInput, existing []domain.TopologySource) []string {
	seen := map[string]struct{}{}
	var communities []string
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		communities = append(communities, value)
	}
	add(input.Community)
	for _, source := range existing {
		if source.Enabled && strings.EqualFold(source.SNMPVersion, "v2c") {
			add(source.Community)
		}
	}
	add("public")
	return communities
}

func topologySourceAddressKey(address string, port int) string {
	if port == 0 {
		port = 161
	}
	return strings.ToLower(strings.TrimSpace(address)) + ":" + strconv.Itoa(port)
}

func appendUniqueTopologySource(items []domain.TopologySource, source domain.TopologySource) []domain.TopologySource {
	for _, item := range items {
		if item.ID == source.ID {
			return items
		}
	}
	return append(items, source)
}

func redactTopologySource(source domain.TopologySource) domain.TopologySource {
	source.Community = ""
	source.AuthPassphrase = ""
	source.PrivacyPassphrase = ""
	return source
}
