package docker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

type EndpointResult struct {
	Endpoint     domain.DockerEndpoint
	Observations []domain.Observation
	Err          error
}

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Discover(ctx context.Context, endpoints []domain.DockerEndpoint) []EndpointResult {
	results := make([]EndpointResult, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if !endpoint.Enabled {
			continue
		}
		items, err := p.discoverEndpoint(ctx, endpoint)
		results = append(results, EndpointResult{
			Endpoint:     endpoint,
			Observations: items,
			Err:          err,
		})
	}
	return results
}

func (p *Provider) discoverEndpoint(ctx context.Context, endpoint domain.DockerEndpoint) ([]domain.Observation, error) {
	client, baseURL, err := buildClient(endpoint)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/containers/json", nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("docker endpoint %s returned %s", endpoint.Name, response.Status)
	}

	var containers []containerSummary
	if err := json.NewDecoder(response.Body).Decode(&containers); err != nil {
		return nil, err
	}

	host, ip := endpointHost(endpoint)
	now := time.Now().UTC()
	device := domain.DeviceObservation{
		IdentityKey: "docker-host:" + endpoint.ID,
		Hostname:    host,
		DisplayName: firstNonEmpty(endpoint.Name, host),
		IPAddress:   ip,
		Confidence:  domain.IdentityConfidenceLow,
		LastSeenAt:  now,
	}
	observations := []domain.Observation{{Device: device}}

	for _, container := range containers {
		if disabledByLabel(container.Labels) || strings.EqualFold(container.State, "exited") {
			continue
		}
		name := firstNonEmpty(container.Labels["homelabwatch.name"], firstContainerName(container.Names), container.Image)
		overrideURL := container.Labels["homelabwatch.url"]
		overrideHost := container.Labels["homelabwatch.host"]
		overridePath := container.Labels["homelabwatch.path"]
		icon := container.Labels["homelabwatch.icon"]

		services := make([]domain.ServiceObservation, 0)
		for _, port := range effectivePorts(container.Ports, overrideURL) {
			serviceHost := firstNonEmpty(overrideHost, ip, host, "localhost")
			scheme := schemeForPort(port, overrideURL)
			urlValue := overrideURL
			if urlValue == "" {
				urlValue = buildURL(scheme, serviceHost, port, overridePath)
			}
			serviceType := serviceTypeFromImage(container.Image)
			addressSource := domain.ServiceAddressLiteralHost
			hostValue := serviceHost
			if ip != "" {
				addressSource = domain.ServiceAddressDevicePrimary
				hostValue = ip
			}
			serviceName := name
			if len(container.Ports) > 1 && port > 0 {
				serviceName = fmt.Sprintf("%s %d", name, port)
			}
			services = append(services, domain.ServiceObservation{
				Name:            serviceName,
				Source:          domain.ServiceSourceDocker,
				SourceRef:       fmt.Sprintf("%s:%s:%d", endpoint.ID, container.ID, port),
				DeviceKey:       device.IdentityKey,
				ServiceTypeHint: serviceType,
				AddressSource:   addressSource,
				HostValue:       hostValue,
				Icon:            firstNonEmpty(icon, serviceType),
				Scheme:          scheme,
				Host:            serviceHost,
				Port:            port,
				Path:            overridePath,
				URL:             urlValue,
				LastSeenAt:      now,
				Details: map[string]any{
					"containerID":   container.ID,
					"containerName": firstContainerName(container.Names),
					"image":         container.Image,
					"state":         container.State,
					"labels":        container.Labels,
					"endpoint":      endpoint.Name,
				},
			})
		}
		if len(services) > 0 {
			observations = append(observations, domain.Observation{Device: device, Services: services})
		}
	}
	return observations, nil
}

func serviceTypeFromImage(image string) string {
	image = strings.ToLower(strings.TrimSpace(image))
	switch {
	case strings.Contains(image, "grafana/grafana"):
		return "grafana"
	case strings.Contains(image, "homeassistant/home-assistant"):
		return "home-assistant"
	case strings.Contains(image, "prom/prometheus"):
		return "prometheus"
	case strings.Contains(image, "plexinc/pms-docker"), strings.Contains(image, "linuxserver/plex"):
		return "plex"
	case strings.Contains(image, "portainer/portainer"):
		return "portainer"
	case strings.Contains(image, "nextcloud"):
		return "nextcloud"
	default:
		return ""
	}
}

type containerSummary struct {
	ID     string            `json:"Id"`
	Image  string            `json:"Image"`
	State  string            `json:"State"`
	Names  []string          `json:"Names"`
	Ports  []containerPort   `json:"Ports"`
	Labels map[string]string `json:"Labels"`
}

type containerPort struct {
	PrivatePort uint16 `json:"PrivatePort"`
	PublicPort  uint16 `json:"PublicPort"`
	Type        string `json:"Type"`
}

func buildClient(endpoint domain.DockerEndpoint) (*http.Client, string, error) {
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	address := strings.TrimSpace(endpoint.Address)
	baseURL := address

	switch {
	case strings.HasPrefix(address, "unix://"):
		socketPath := strings.TrimPrefix(address, "unix://")
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 4 * time.Second}).DialContext(ctx, "unix", socketPath)
		}
		baseURL = "http://docker"
	case strings.HasPrefix(address, "tcp://"):
		baseURL = "http://" + strings.TrimPrefix(address, "tcp://")
	case strings.HasPrefix(address, "https://"), strings.HasPrefix(address, "http://"):
	default:
		return nil, "", fmt.Errorf("unsupported docker endpoint %q", address)
	}

	if endpoint.TLSCAPath != "" || endpoint.TLSCertPath != "" || endpoint.TLSKeyPath != "" {
		config, err := tlsConfig(endpoint)
		if err != nil {
			return nil, "", err
		}
		transport.TLSClientConfig = config
		if after, ok := strings.CutPrefix(baseURL, "http://"); ok {
			baseURL = "https://" + after
		}
	}
	return &http.Client{Timeout: 8 * time.Second, Transport: transport}, strings.TrimRight(baseURL, "/"), nil
}

func tlsConfig(endpoint domain.DockerEndpoint) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12}
	if endpoint.TLSCAPath != "" {
		pool := x509.NewCertPool()
		caPEM, err := os.ReadFile(endpoint.TLSCAPath)
		if err != nil {
			return nil, err
		}
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("invalid docker ca bundle")
		}
		config.RootCAs = pool
	}
	if endpoint.TLSCertPath != "" && endpoint.TLSKeyPath != "" {
		certificate, err := tls.LoadX509KeyPair(endpoint.TLSCertPath, endpoint.TLSKeyPath)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{certificate}
	}
	return config, nil
}

func endpointHost(endpoint domain.DockerEndpoint) (string, string) {
	address := strings.TrimSpace(endpoint.Address)
	if strings.HasPrefix(address, "unix://") {
		hostname, _ := os.Hostname()
		return firstNonEmpty(endpoint.Name, hostname, "localhost"), ""
	}
	parsed, err := url.Parse(address)
	if err != nil {
		return endpoint.Name, ""
	}
	host := parsed.Hostname()
	ip := ""
	if parsedIP := net.ParseIP(host); parsedIP != nil {
		ip = parsedIP.String()
	}
	return firstNonEmpty(endpoint.Name, host), ip
}

func disabledByLabel(labels map[string]string) bool {
	value := strings.TrimSpace(strings.ToLower(labels["homelabwatch.enable"]))
	return value == "false" || value == "0" || value == "no"
}

func effectivePorts(ports []containerPort, overrideURL string) []int {
	if overrideURL != "" {
		parsed, err := url.Parse(overrideURL)
		if err == nil && parsed.Port() != "" {
			if port, err := net.LookupPort("tcp", parsed.Port()); err == nil {
				return []int{port}
			}
		}
	}
	values := make([]int, 0, len(ports))
	for _, port := range ports {
		if port.PublicPort > 0 {
			values = append(values, int(port.PublicPort))
			continue
		}
		if port.PrivatePort > 0 {
			values = append(values, int(port.PrivatePort))
		}
	}
	return values
}

func firstContainerName(names []string) string {
	for _, name := range names {
		name = strings.TrimPrefix(name, "/")
		if name != "" {
			return name
		}
	}
	return ""
}

func schemeForPort(port int, overrideURL string) string {
	if overrideURL != "" {
		if strings.HasPrefix(overrideURL, "https://") {
			return "https"
		}
		if strings.HasPrefix(overrideURL, "http://") {
			return "http"
		}
	}
	switch port {
	case 443, 8443, 9443:
		return "https"
	default:
		return "http"
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
