package lan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Provider struct {
	MaxHosts     int
	ProbeTimeout time.Duration
	PingTimeout  time.Duration
	WorkerCount  int
}

func NewProvider() *Provider {
	return &Provider{
		MaxHosts:     4096,
		ProbeTimeout: 500 * time.Millisecond,
		PingTimeout:  900 * time.Millisecond,
		WorkerCount:  64,
	}
}

func (p *Provider) SuggestedTargets(defaultPorts []int) ([]domain.ScanTargetSeed, error) {
	prefixes, err := localPrefixes()
	if err != nil {
		return nil, err
	}
	results := make([]domain.ScanTargetSeed, 0, len(prefixes))
	for _, prefix := range prefixes {
		results = append(results, domain.ScanTargetSeed{
			Name:                "Auto-detected subnet",
			CIDR:                prefix.String(),
			AutoDetected:        true,
			Enabled:             true,
			ScanIntervalSeconds: 300,
			CommonPorts:         append([]int(nil), defaultPorts...),
		})
	}
	return results, nil
}

func (p *Provider) Discover(ctx context.Context, targets []domain.ScanTarget) ([]domain.Observation, error) {
	if p.WorkerCount <= 0 {
		p.WorkerCount = 64
	}
	type job struct {
		target domain.ScanTarget
		addr   netip.Addr
	}
	jobs := make(chan job)
	results := make(chan domain.Observation, p.WorkerCount)
	var wg sync.WaitGroup

	for index := 0; index < p.WorkerCount; index++ {
		wg.Go(func() {
			for item := range jobs {
				observation, ok := p.probeHost(ctx, item.target, item.addr)
				if !ok {
					continue
				}
				select {
				case results <- observation:
				case <-ctx.Done():
					return
				}
			}
		})
	}

	go func() {
		defer close(jobs)
		for _, target := range targets {
			if !target.Enabled {
				continue
			}
			prefix, err := netip.ParsePrefix(target.CIDR)
			if err != nil {
				continue
			}
			for _, addr := range expandPrefix(prefix, p.MaxHosts) {
				select {
				case jobs <- job{target: target, addr: addr}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	observations := make([]domain.Observation, 0)
	for observation := range results {
		observations = append(observations, observation)
	}
	return observations, ctx.Err()
}

func (p *Provider) probeHost(ctx context.Context, target domain.ScanTarget, addr netip.Addr) (domain.Observation, bool) {
	ip := addr.String()
	now := time.Now().UTC()
	alive := p.ping(ip)
	openPorts := p.scanPorts(ctx, ip, target.CommonPorts)
	if !alive && len(openPorts) == 0 {
		return domain.Observation{}, false
	}
	hostname := reverseLookup(ip)
	mac, iface := lookupARP(ip)
	identityKey := identityKey(mac, hostname, ip, target.CIDR)
	confidence := domain.IdentityConfidenceLow
	if mac != "" {
		confidence = domain.IdentityConfidenceHigh
	}
	device := domain.DeviceObservation{
		IdentityKey: identityKey,
		PrimaryMAC:  mac,
		Hostname:    hostname,
		DisplayName: firstNonEmpty(hostname, ip, mac),
		IPAddress:   ip,
		Interface:   iface,
		Confidence:  confidence,
		LastSeenAt:  now,
	}
	services := make([]domain.ServiceObservation, 0, len(openPorts))
	for _, port := range openPorts {
		hint := portHint(port)
		serviceType := serviceTypeForPort(port)
		device.Ports = append(device.Ports, domain.PortObservation{Port: port, Protocol: "tcp", ServiceHint: hint})
		services = append(services, domain.ServiceObservation{
			Name:            serviceName(hostname, ip, port, hint),
			Source:          domain.ServiceSourceLAN,
			SourceRef:       fmt.Sprintf("%s:%d/tcp", identityKey, port),
			DeviceKey:       identityKey,
			ServiceTypeHint: serviceType,
			AddressSource:   domain.ServiceAddressDevicePrimary,
			HostValue:       ip,
			Icon:            serviceType,
			Scheme:          serviceScheme(port),
			Host:            ip,
			Port:            port,
			URL:             buildURL(serviceScheme(port), ip, port, ""),
			LastSeenAt:      now,
			Details: map[string]any{
				"hostname": hostname,
				"mac":      mac,
				"hint":     hint,
			},
		})
		if strings.HasSuffix(strings.ToLower(hostname), ".local") {
			services = append(services, domain.ServiceObservation{
				Name:            serviceName(hostname, hostname, port, hint),
				Source:          domain.ServiceSourceMDNS,
				SourceRef:       fmt.Sprintf("%s:%d/mdns", hostname, port),
				DeviceKey:       identityKey,
				ServiceTypeHint: serviceType,
				AddressSource:   domain.ServiceAddressMDNSHostname,
				HostValue:       hostname,
				Icon:            serviceType,
				Scheme:          serviceScheme(port),
				Host:            hostname,
				Port:            port,
				URL:             buildURL(serviceScheme(port), hostname, port, ""),
				LastSeenAt:      now,
				Details: map[string]any{
					"hostname":    hostname,
					"mac":         mac,
					"hint":        hint,
					"mdnsService": mdnsServiceForPort(port),
				},
			})
		}
	}
	return domain.Observation{Device: device, Services: services}, true
}

func (p *Provider) scanPorts(ctx context.Context, host string, ports []int) []int {
	openPorts := make([]int, 0, len(ports))
	for _, port := range ports {
		timeoutCtx, cancel := context.WithTimeout(ctx, p.ProbeTimeout)
		conn, err := (&net.Dialer{}).DialContext(timeoutCtx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
		cancel()
		if err != nil {
			continue
		}
		_ = conn.Close()
		openPorts = append(openPorts, port)
	}
	slices.Sort(openPorts)
	return openPorts
}

func (p *Provider) ping(host string) bool {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(p.PingTimeout))
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("hlw")},
	}
	payload, err := message.Marshal(nil)
	if err != nil {
		return false
	}
	if _, err := conn.WriteTo(payload, &net.IPAddr{IP: net.ParseIP(host)}); err != nil {
		return false
	}
	buffer := make([]byte, 1500)
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return false
	}
	reply, err := icmp.ParseMessage(1, buffer[:n])
	return err == nil && reply.Type == ipv4.ICMPTypeEchoReply
}

func localPrefixes() ([]netip.Prefix, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	results := make([]netip.Prefix, 0)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			ipNet, ok := address.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.To4() == nil {
				continue
			}
			prefix, err := netip.ParsePrefix(ipNet.String())
			if err != nil {
				continue
			}
			prefix = prefix.Masked()
			if _, ok := seen[prefix.String()]; ok {
				continue
			}
			seen[prefix.String()] = struct{}{}
			results = append(results, prefix)
		}
	}
	slices.SortFunc(results, func(a, b netip.Prefix) int { return strings.Compare(a.String(), b.String()) })
	return results, nil
}

func expandPrefix(prefix netip.Prefix, maxHosts int) []netip.Addr {
	prefix = prefix.Masked()
	if !prefix.Addr().Is4() {
		return nil
	}
	hostBits := 32 - prefix.Bits()
	if hostBits > 12 {
		return nil
	}
	limit := min(1<<hostBits, maxHosts)
	results := make([]netip.Addr, 0, limit)
	base := prefix.Addr()
	for index := 1; index < limit-1; index++ {
		addr := addrNext(base, index)
		if !prefix.Contains(addr) {
			break
		}
		results = append(results, addr)
	}
	return results
}

func addrNext(addr netip.Addr, step int) netip.Addr {
	value := addr.As4()
	number := uint32(value[0])<<24 | uint32(value[1])<<16 | uint32(value[2])<<8 | uint32(value[3])
	number += uint32(step)
	return netip.AddrFrom4([4]byte{byte(number >> 24), byte(number >> 16), byte(number >> 8), byte(number)})
}

func reverseLookup(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func lookupARP(ip string) (string, string) {
	content, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(string(content), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 6 || fields[0] != ip {
			continue
		}
		return strings.ToLower(fields[3]), fields[5]
	}
	return "", ""
}

func identityKey(mac, hostname, ip, cidr string) string {
	if mac != "" {
		return "mac:" + strings.ToLower(mac)
	}
	hash := sha256.Sum256([]byte(strings.ToLower(firstNonEmpty(hostname, ip) + "|" + cidr)))
	return "fp:" + hex.EncodeToString(hash[:8])
}

func serviceScheme(port int) string {
	switch port {
	case 443, 8443, 9443:
		return "https"
	default:
		return "http"
	}
}

func portHint(port int) string {
	switch port {
	case 22:
		return "ssh"
	case 53:
		return "dns"
	case 80:
		return "http"
	case 443:
		return "https"
	case 3000, 8080, 8123, 9000, 9443, 32400:
		return "web"
	default:
		return ""
	}
}

func serviceTypeForPort(port int) string {
	switch port {
	case 3000:
		return "grafana"
	case 8123:
		return "home-assistant"
	case 9090:
		return "prometheus"
	case 32400:
		return "plex"
	case 9000, 9443:
		return "portainer"
	default:
		return ""
	}
}

func serviceName(hostname, ip string, port int, hint string) string {
	if displayName := displayNameForPort(port); displayName != "" {
		return displayName
	}
	return fmt.Sprintf("%s %s", firstNonEmpty(hostname, ip), strings.ToUpper(firstNonEmpty(hint, "service")))
}

func displayNameForPort(port int) string {
	switch port {
	case 3000:
		return "Grafana"
	case 8123:
		return "Home Assistant"
	case 9090:
		return "Prometheus"
	case 32400:
		return "Plex"
	case 9000, 9443:
		return "Portainer"
	default:
		return ""
	}
}

func mdnsServiceForPort(port int) string {
	switch port {
	case 8123:
		return "_home-assistant._tcp"
	case 32400:
		return "_plexmediasvr._tcp"
	case 443, 8443, 9443:
		return "_https._tcp"
	default:
		return "_http._tcp"
	}
}

func buildURL(scheme, host string, port int, path string) string {
	base := fmt.Sprintf("%s://%s", scheme, host)
	if port > 0 && port != 80 && port != 443 {
		base = fmt.Sprintf("%s:%d", base, port)
	}
	if path == "" {
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
