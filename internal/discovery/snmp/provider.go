package snmp

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	g "github.com/gosnmp/gosnmp"

	"github.com/deleema/homelabwatch/internal/domain"
)

const (
	oidIfDescr          = ".1.3.6.1.2.1.2.2.1.2"
	oidIfType           = ".1.3.6.1.2.1.2.2.1.3"
	oidIfSpeed          = ".1.3.6.1.2.1.2.2.1.5"
	oidIfOperStatus     = ".1.3.6.1.2.1.2.2.1.8"
	oidIfName           = ".1.3.6.1.2.1.31.1.1.1.1"
	oidIfHighSpeed      = ".1.3.6.1.2.1.31.1.1.1.15"
	oidIfAlias          = ".1.3.6.1.2.1.31.1.1.1.18"
	oidLLDPLocChassisID = ".1.0.8802.1.1.2.1.3.2.0"
	oidLLDPLocSysName   = ".1.0.8802.1.1.2.1.3.3.0"
	oidLLDPLocPortID    = ".1.0.8802.1.1.2.1.3.7.1.3"
	oidLLDPLocPortDesc  = ".1.0.8802.1.1.2.1.3.7.1.4"
	oidLLDPRemChassisID = ".1.0.8802.1.1.2.1.4.1.1.5"
	oidLLDPRemPortID    = ".1.0.8802.1.1.2.1.4.1.1.7"
	oidLLDPRemPortDesc  = ".1.0.8802.1.1.2.1.4.1.1.8"
	oidLLDPRemSysName   = ".1.0.8802.1.1.2.1.4.1.1.9"
	oidBasePortIfIndex  = ".1.3.6.1.2.1.17.1.4.1.2"
	oidBridgeFDBPort    = ".1.3.6.1.2.1.17.4.3.1.2"
	oidBridgeFDBStatus  = ".1.3.6.1.2.1.17.4.3.1.3"
	oidQBridgeFDBPort   = ".1.3.6.1.2.1.17.7.1.2.2.1.2"
	oidQBridgeFDBStatus = ".1.3.6.1.2.1.17.7.1.2.2.1.3"
)

type SourceResult struct {
	Source      domain.TopologySource
	Observation domain.TopologySourceObservation
	Error       error
}

type Provider struct {
	open func(context.Context, domain.TopologySource) (session, error)
	now  func() time.Time
}

type session interface {
	WalkAll(oid string) ([]g.SnmpPDU, error)
	Close() error
}

type realSession struct {
	client *g.GoSNMP
	bulk   bool
}

func NewProvider() *Provider {
	return &Provider{
		open: openSession,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (p *Provider) Discover(ctx context.Context, sources []domain.TopologySource) []SourceResult {
	if p == nil {
		p = NewProvider()
	}
	results := make([]SourceResult, 0, len(sources))
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		observedAt := p.now()
		result := SourceResult{
			Source: source,
			Observation: domain.TopologySourceObservation{
				SourceID:   source.ID,
				Source:     source,
				ObservedAt: observedAt,
			},
		}
		sess, err := p.open(ctx, source)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}
		var walkErrs []error
		if interfaces, err := pollInterfaces(sess, source.ID, observedAt); err == nil {
			result.Observation.Interfaces = interfaces
		} else {
			walkErrs = append(walkErrs, err)
		}
		if links, err := pollLLDP(sess, source.ID, observedAt, result.Observation.Interfaces); err == nil {
			result.Observation.LLDPLinks = links
		} else {
			walkErrs = append(walkErrs, err)
		}
		if links, err := pollQBridge(sess, source.ID, observedAt, result.Observation.Interfaces); err == nil && len(links) > 0 {
			result.Observation.MACLinks = links
		} else {
			if err != nil {
				walkErrs = append(walkErrs, err)
			}
			if fallback, fallbackErr := pollBridge(sess, source.ID, observedAt, result.Observation.Interfaces); fallbackErr == nil {
				result.Observation.MACLinks = fallback
			} else {
				walkErrs = append(walkErrs, fallbackErr)
			}
		}
		_ = sess.Close()
		if len(result.Observation.Interfaces) == 0 && len(result.Observation.LLDPLinks) == 0 && len(result.Observation.MACLinks) == 0 && len(walkErrs) > 0 {
			result.Error = errors.Join(walkErrs...)
		}
		results = append(results, result)
	}
	return results
}

func openSession(ctx context.Context, source domain.TopologySource) (session, error) {
	port := source.Port
	if port == 0 {
		port = 161
	}
	timeout := source.TimeoutMS
	if timeout == 0 {
		timeout = 1500
	}
	retries := source.Retries
	if retries == 0 {
		retries = 1
	}
	client := &g.GoSNMP{
		Context:            ctx,
		Port:               uint16(port),
		Retries:            retries,
		Target:             source.Address,
		Timeout:            time.Duration(timeout) * time.Millisecond,
		Version:            g.Version2c,
		Community:          source.Community,
		MaxRepetitions:     25,
		ExponentialTimeout: true,
	}
	if strings.EqualFold(source.SNMPVersion, "v3") {
		authProtocol := authProtocol(source.AuthProtocol)
		privProtocol := privacyProtocol(source.PrivacyProtocol)
		flags := g.NoAuthNoPriv
		if authProtocol != g.NoAuth && privProtocol != g.NoPriv {
			flags = g.AuthPriv
		} else if authProtocol != g.NoAuth {
			flags = g.AuthNoPriv
		}
		client.Version = g.Version3
		client.SecurityModel = g.UserSecurityModel
		client.MsgFlags = flags
		client.SecurityParameters = &g.UsmSecurityParameters{
			UserName:                 source.Username,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: source.AuthPassphrase,
			PrivacyProtocol:          privProtocol,
			PrivacyPassphrase:        source.PrivacyPassphrase,
		}
	}
	if err := client.Connect(); err != nil {
		return nil, err
	}
	return realSession{client: client, bulk: client.Version != g.Version1}, nil
}

func (s realSession) WalkAll(oid string) ([]g.SnmpPDU, error) {
	if s.bulk {
		return s.client.BulkWalkAll(oid)
	}
	return s.client.WalkAll(oid)
}

func (s realSession) Close() error { return s.client.Close() }

func pollInterfaces(sess session, sourceID string, observedAt time.Time) ([]domain.TopologyInterfaceObservation, error) {
	interfaces := map[int]*domain.TopologyInterfaceObservation{}
	walks := []struct {
		oid string
		set func(*domain.TopologyInterfaceObservation, g.SnmpPDU)
	}{
		{oidIfDescr, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) { item.IfDescription = pduText(pdu) }},
		{oidIfType, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) { item.IfType = pduInt(pdu) }},
		{oidIfSpeed, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) { item.SpeedBPS = pduUint(pdu) }},
		{oidIfOperStatus, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) {
			item.OperStatus = operStatus(pduInt(pdu))
		}},
		{oidIfName, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) { item.IfName = pduText(pdu) }},
		{oidIfAlias, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) { item.IfAlias = pduText(pdu) }},
		{oidIfHighSpeed, func(item *domain.TopologyInterfaceObservation, pdu g.SnmpPDU) {
			if speedMbps := pduUint(pdu); speedMbps > 0 {
				item.SpeedBPS = speedMbps * 1_000_000
			}
		}},
	}
	var errs []error
	for _, walk := range walks {
		pdus, err := sess.WalkAll(walk.oid)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", walk.oid, err))
			continue
		}
		for _, pdu := range pdus {
			ifIndex := trailingInt(pdu.Name, walk.oid)
			if ifIndex == 0 {
				continue
			}
			item := interfaces[ifIndex]
			if item == nil {
				item = &domain.TopologyInterfaceObservation{SourceID: sourceID, IfIndex: ifIndex, LastSeenAt: observedAt, CreatedAt: observedAt, UpdatedAt: observedAt}
				interfaces[ifIndex] = item
			}
			walk.set(item, pdu)
		}
	}
	if len(interfaces) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	items := make([]domain.TopologyInterfaceObservation, 0, len(interfaces))
	for _, item := range interfaces {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].IfIndex < items[j].IfIndex })
	return items, nil
}

func pollLLDP(sess session, sourceID string, observedAt time.Time, interfaces []domain.TopologyInterfaceObservation) ([]domain.TopologyLLDPLinkObservation, error) {
	localChassisID := ""
	if pdus, err := sess.WalkAll(oidLLDPLocChassisID); err == nil && len(pdus) > 0 {
		localChassisID = pduID(pdus[0])
	}
	localSystemName := ""
	if pdus, err := sess.WalkAll(oidLLDPLocSysName); err == nil && len(pdus) > 0 {
		localSystemName = pduText(pdus[0])
	}
	localPorts := map[int]domain.TopologyLLDPLinkObservation{}
	for _, oid := range []string{oidLLDPLocPortID, oidLLDPLocPortDesc} {
		pdus, err := sess.WalkAll(oid)
		if err != nil {
			continue
		}
		for _, pdu := range pdus {
			portNum := trailingInt(pdu.Name, oid)
			if portNum == 0 {
				continue
			}
			item := localPorts[portNum]
			item.LocalChassisID = localChassisID
			item.LocalSystemName = localSystemName
			if oid == oidLLDPLocPortID {
				item.LocalPortID = pduID(pdu)
				item.LocalPortName = pduText(pdu)
			} else {
				item.LocalPortDescription = pduText(pdu)
			}
			item.LocalIfIndex = matchLLDPPortToInterface(portNum, item.LocalPortID, item.LocalPortDescription, interfaces)
			localPorts[portNum] = item
		}
	}
	remote := map[string]*domain.TopologyLLDPLinkObservation{}
	walks := []struct {
		oid string
		set func(*domain.TopologyLLDPLinkObservation, g.SnmpPDU)
	}{
		{oidLLDPRemChassisID, func(item *domain.TopologyLLDPLinkObservation, pdu g.SnmpPDU) { item.RemoteChassisID = pduID(pdu) }},
		{oidLLDPRemPortID, func(item *domain.TopologyLLDPLinkObservation, pdu g.SnmpPDU) { item.RemotePortID = pduID(pdu) }},
		{oidLLDPRemPortDesc, func(item *domain.TopologyLLDPLinkObservation, pdu g.SnmpPDU) {
			item.RemotePortDescription = pduText(pdu)
		}},
		{oidLLDPRemSysName, func(item *domain.TopologyLLDPLinkObservation, pdu g.SnmpPDU) { item.RemoteSystemName = pduText(pdu) }},
	}
	var errs []error
	for _, walk := range walks {
		pdus, err := sess.WalkAll(walk.oid)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", walk.oid, err))
			continue
		}
		for _, pdu := range pdus {
			key := strings.TrimPrefix(strings.TrimPrefix(pdu.Name, walk.oid), ".")
			if key == "" {
				continue
			}
			item := remote[key]
			if item == nil {
				item = &domain.TopologyLLDPLinkObservation{SourceID: sourceID, LocalChassisID: localChassisID, LocalSystemName: localSystemName, LastSeenAt: observedAt, CreatedAt: observedAt, UpdatedAt: observedAt}
				if portNum := lldpRemoteLocalPort(key); portNum > 0 {
					if local, ok := localPorts[portNum]; ok {
						item.LocalPortID = local.LocalPortID
						item.LocalPortName = local.LocalPortName
						item.LocalPortDescription = local.LocalPortDescription
						item.LocalIfIndex = local.LocalIfIndex
					}
				}
				remote[key] = item
			}
			walk.set(item, pdu)
		}
	}
	if len(remote) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	items := make([]domain.TopologyLLDPLinkObservation, 0, len(remote))
	for _, item := range remote {
		if item.RemoteChassisID == "" && item.RemotePortID == "" && item.RemoteSystemName == "" {
			continue
		}
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].LocalPortID == items[j].LocalPortID {
			return items[i].RemoteChassisID < items[j].RemoteChassisID
		}
		return items[i].LocalPortID < items[j].LocalPortID
	})
	return items, nil
}

func pollQBridge(sess session, sourceID string, observedAt time.Time, interfaces []domain.TopologyInterfaceObservation) ([]domain.TopologyMACLinkObservation, error) {
	basePorts, _ := pollBridgePortIfIndex(sess)
	ports, err := sess.WalkAll(oidQBridgeFDBPort)
	if err != nil {
		return nil, err
	}
	statuses := map[string]string{}
	if pdus, err := sess.WalkAll(oidQBridgeFDBStatus); err == nil {
		for _, pdu := range pdus {
			statuses[qBridgeMACKey(pdu.Name, oidQBridgeFDBStatus)] = fdbStatus(pduInt(pdu))
		}
	}
	items := make([]domain.TopologyMACLinkObservation, 0, len(ports))
	for _, pdu := range ports {
		vlan, mac := qBridgeVLANMAC(pdu.Name, oidQBridgeFDBPort)
		if mac == "" {
			continue
		}
		bridgePort := pduInt(pdu)
		item := buildMACObservation(sourceID, mac, vlan, bridgePort, basePorts[bridgePort], statuses[qBridgeMACKey(pdu.Name, oidQBridgeFDBPort)], observedAt, interfaces)
		items = append(items, item)
	}
	return items, nil
}

func pollBridge(sess session, sourceID string, observedAt time.Time, interfaces []domain.TopologyInterfaceObservation) ([]domain.TopologyMACLinkObservation, error) {
	basePorts, _ := pollBridgePortIfIndex(sess)
	ports, err := sess.WalkAll(oidBridgeFDBPort)
	if err != nil {
		return nil, err
	}
	statuses := map[string]string{}
	if pdus, err := sess.WalkAll(oidBridgeFDBStatus); err == nil {
		for _, pdu := range pdus {
			statuses[bridgeMACKey(pdu.Name, oidBridgeFDBStatus)] = fdbStatus(pduInt(pdu))
		}
	}
	items := make([]domain.TopologyMACLinkObservation, 0, len(ports))
	for _, pdu := range ports {
		mac := bridgeMACKey(pdu.Name, oidBridgeFDBPort)
		if mac == "" {
			continue
		}
		bridgePort := pduInt(pdu)
		item := buildMACObservation(sourceID, mac, 0, bridgePort, basePorts[bridgePort], statuses[mac], observedAt, interfaces)
		items = append(items, item)
	}
	return items, nil
}

func pollBridgePortIfIndex(sess session) (map[int]int, error) {
	pdus, err := sess.WalkAll(oidBasePortIfIndex)
	if err != nil {
		return nil, err
	}
	items := map[int]int{}
	for _, pdu := range pdus {
		bridgePort := trailingInt(pdu.Name, oidBasePortIfIndex)
		if bridgePort > 0 {
			items[bridgePort] = pduInt(pdu)
		}
	}
	return items, nil
}

func buildMACObservation(sourceID, mac string, vlan, bridgePort, ifIndex int, status string, observedAt time.Time, interfaces []domain.TopologyInterfaceObservation) domain.TopologyMACLinkObservation {
	item := domain.TopologyMACLinkObservation{
		SourceID:   sourceID,
		MACAddress: normalizeMAC(mac),
		VLAN:       vlan,
		BridgePort: bridgePort,
		IfIndex:    ifIndex,
		Status:     status,
		LastSeenAt: observedAt,
		CreatedAt:  observedAt,
		UpdatedAt:  observedAt,
	}
	for _, iface := range interfaces {
		if iface.IfIndex == ifIndex {
			item.IfName = iface.IfName
			item.IfDescription = iface.IfDescription
			break
		}
	}
	return item
}

func matchLLDPPortToInterface(portNum int, portID, portDescription string, interfaces []domain.TopologyInterfaceObservation) int {
	for _, iface := range interfaces {
		if iface.IfIndex == portNum {
			return iface.IfIndex
		}
	}
	needle := strings.ToLower(firstNonEmpty(portID, portDescription))
	if needle == "" {
		return 0
	}
	for _, iface := range interfaces {
		if strings.ToLower(iface.IfName) == needle || strings.ToLower(iface.IfDescription) == needle {
			return iface.IfIndex
		}
	}
	return 0
}

func trailingInt(oid, base string) int {
	parts := oidSuffixInts(oid, base)
	if len(parts) == 0 {
		return 0
	}
	return parts[len(parts)-1]
}

func lldpRemoteLocalPort(suffix string) int {
	parts := splitOIDInts(suffix)
	if len(parts) < 3 {
		return 0
	}
	return parts[1]
}

func qBridgeVLANMAC(oid, base string) (int, string) {
	parts := oidSuffixInts(oid, base)
	if len(parts) < 7 {
		return 0, ""
	}
	return parts[len(parts)-7], macFromInts(parts[len(parts)-6:])
}

func qBridgeMACKey(oid, base string) string {
	_, mac := qBridgeVLANMAC(oid, base)
	return mac
}

func bridgeMACKey(oid, base string) string {
	parts := oidSuffixInts(oid, base)
	if len(parts) < 6 {
		return ""
	}
	return macFromInts(parts[len(parts)-6:])
}

func oidSuffixInts(oid, base string) []int {
	return splitOIDInts(strings.TrimPrefix(strings.TrimPrefix(oid, base), "."))
}

func splitOIDInts(suffix string) []int {
	suffix = strings.Trim(suffix, ".")
	if suffix == "" {
		return nil
	}
	raw := strings.Split(suffix, ".")
	items := make([]int, 0, len(raw))
	for _, part := range raw {
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil
		}
		items = append(items, value)
	}
	return items
}

func macFromInts(parts []int) string {
	if len(parts) != 6 {
		return ""
	}
	bytes := make([]byte, 6)
	for i, part := range parts {
		if part < 0 || part > 255 {
			return ""
		}
		bytes[i] = byte(part)
	}
	return normalizeMACBytes(bytes)
}

func pduText(pdu g.SnmpPDU) string {
	switch value := pdu.Value.(type) {
	case string:
		return strings.TrimSpace(value)
	case []byte:
		return strings.TrimSpace(string(value))
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func pduID(pdu g.SnmpPDU) string {
	switch value := pdu.Value.(type) {
	case []byte:
		if len(value) == 6 {
			return normalizeMACBytes(value)
		}
		if printableBytes(value) {
			return strings.TrimSpace(string(value))
		}
		return normalizeHexBytes(value)
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func pduInt(pdu g.SnmpPDU) int {
	value := g.ToBigInt(pdu.Value)
	if value == nil {
		return 0
	}
	return int(value.Int64())
}

func pduUint(pdu g.SnmpPDU) uint64 {
	value := g.ToBigInt(pdu.Value)
	if value == nil || value.Sign() < 0 {
		return 0
	}
	return value.Uint64()
}

func printableBytes(value []byte) bool {
	if len(value) == 0 {
		return false
	}
	for _, r := range string(value) {
		if r == unicode.ReplacementChar || (!unicode.IsPrint(r) && !unicode.IsSpace(r)) {
			return false
		}
	}
	return true
}

func normalizeHexBytes(value []byte) string {
	parts := make([]string, 0, len(value))
	for _, b := range value {
		parts = append(parts, fmt.Sprintf("%02x", b))
	}
	return strings.Join(parts, ":")
}

func normalizeMACBytes(value []byte) string {
	if len(value) != 6 {
		return normalizeHexBytes(value)
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", value[0], value[1], value[2], value[3], value[4], value[5])
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
	for i := 0; i < len(value); i += 2 {
		parts = append(parts, value[i:i+2])
	}
	return strings.Join(parts, ":")
}

func operStatus(value int) string {
	switch value {
	case 1:
		return "up"
	case 2:
		return "down"
	case 3:
		return "testing"
	case 5:
		return "dormant"
	case 6:
		return "not-present"
	case 7:
		return "lower-layer-down"
	default:
		return "unknown"
	}
}

func fdbStatus(value int) string {
	switch value {
	case 1:
		return "other"
	case 2:
		return "invalid"
	case 3:
		return "learned"
	case 4:
		return "self"
	case 5:
		return "management"
	default:
		return "unknown"
	}
}

func authProtocol(value string) g.SnmpV3AuthProtocol {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "md5":
		return g.MD5
	case "sha":
		return g.SHA
	case "sha224":
		return g.SHA224
	case "sha256":
		return g.SHA256
	case "sha384":
		return g.SHA384
	case "sha512":
		return g.SHA512
	default:
		return g.NoAuth
	}
}

func privacyProtocol(value string) g.SnmpV3PrivProtocol {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "des":
		return g.DES
	case "aes":
		return g.AES
	case "aes192":
		return g.AES192
	case "aes256":
		return g.AES256
	default:
		return g.NoPriv
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func lowestIP(values []string) string {
	sort.Slice(values, func(i, j int) bool {
		ai, aerr := netip.ParseAddr(values[i])
		bi, berr := netip.ParseAddr(values[j])
		if aerr == nil && berr == nil {
			return ai.Less(bi)
		}
		return values[i] < values[j]
	})
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
