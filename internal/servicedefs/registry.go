package servicedefs

import (
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"

	"github.com/deleema/homelabwatch/internal/domain"
)

type Candidate struct {
	Name        string
	ServiceType string
	Icon        string
	Scheme      string
	Host        string
	HostValue   string
	Port        int
	Path        string
	Details     map[string]any
}

func BuiltInDefinitions() []domain.ServiceDefinition {
	return []domain.ServiceDefinition{
		{
			ID:       "builtin_pihole",
			Key:      "pihole",
			Name:     "Pi-hole",
			Icon:     "pihole",
			Priority: 90,
			BuiltIn:  true,
			Enabled:  true,
			Matchers: []domain.ServiceDefinitionMatcher{
				{ID: "builtin_pihole_image", Type: "container_image", Operator: "contains", Value: "pihole/pihole", Weight: 95},
				{ID: "builtin_pihole_title", Type: "page_title", Operator: "contains", Value: "Pi-hole", Weight: 85},
				{ID: "builtin_pihole_body", Type: "body_substring", Operator: "contains", Value: "Pi-hole", Weight: 80},
			},
			CheckTemplates: []domain.ServiceDefinitionCheckTemplate{
				{Name: "Pi-hole admin", Type: domain.CheckTypeHTTP, Path: "/admin", Method: "GET", IntervalSeconds: 60, TimeoutSeconds: 10, ExpectedStatusMin: 200, ExpectedStatusMax: 399, Enabled: true, SortOrder: 0, ConfigSource: domain.HealthCheckConfigSourceDefinition},
			},
		},
		{
			ID:       "builtin_grafana",
			Key:      "grafana",
			Name:     "Grafana",
			Icon:     "grafana",
			Priority: 100,
			BuiltIn:  true,
			Enabled:  true,
			Matchers: []domain.ServiceDefinitionMatcher{
				{ID: "builtin_grafana_image", Type: "container_image", Operator: "contains", Value: "grafana/grafana", Weight: 95},
				{ID: "builtin_grafana_port", Type: "port", Operator: "exact", Value: "3000", Weight: 75},
				{ID: "builtin_grafana_title", Type: "page_title", Operator: "contains", Value: "Grafana", Weight: 85},
				{ID: "builtin_grafana_header", Type: "http_header", Operator: "exists", Extra: "x-grafana-user", Weight: 90},
			},
			CheckTemplates: []domain.ServiceDefinitionCheckTemplate{
				{Name: "Grafana health", Type: domain.CheckTypeHTTP, Path: "/api/health", Method: "GET", IntervalSeconds: 60, TimeoutSeconds: 10, ExpectedStatusMin: 200, ExpectedStatusMax: 399, Enabled: true, SortOrder: 0, ConfigSource: domain.HealthCheckConfigSourceDefinition},
			},
		},
		{
			ID:       "builtin_prometheus",
			Key:      "prometheus",
			Name:     "Prometheus",
			Icon:     "prometheus",
			Priority: 100,
			BuiltIn:  true,
			Enabled:  true,
			Matchers: []domain.ServiceDefinitionMatcher{
				{ID: "builtin_prom_image", Type: "container_image", Operator: "contains", Value: "prom/prometheus", Weight: 95},
				{ID: "builtin_prom_port", Type: "port", Operator: "exact", Value: "9090", Weight: 75},
				{ID: "builtin_prom_title", Type: "page_title", Operator: "contains", Value: "Prometheus", Weight: 85},
				{ID: "builtin_prom_body", Type: "body_substring", Operator: "contains", Value: "Prometheus", Weight: 80},
			},
			CheckTemplates: []domain.ServiceDefinitionCheckTemplate{
				{Name: "Prometheus health", Type: domain.CheckTypeHTTP, Path: "/-/healthy", Method: "GET", IntervalSeconds: 60, TimeoutSeconds: 10, ExpectedStatusMin: 200, ExpectedStatusMax: 399, Enabled: true, SortOrder: 0, ConfigSource: domain.HealthCheckConfigSourceDefinition},
			},
		},
		{
			ID:       "builtin_home_assistant",
			Key:      "home-assistant",
			Name:     "Home Assistant",
			Icon:     "home-assistant",
			Priority: 100,
			BuiltIn:  true,
			Enabled:  true,
			Matchers: []domain.ServiceDefinitionMatcher{
				{ID: "builtin_ha_image", Type: "container_image", Operator: "contains", Value: "homeassistant/home-assistant", Weight: 95},
				{ID: "builtin_ha_port", Type: "port", Operator: "exact", Value: "8123", Weight: 75},
				{ID: "builtin_ha_mdns", Type: "mdns_service", Operator: "exact", Value: "_home-assistant._tcp", Weight: 90},
				{ID: "builtin_ha_title", Type: "page_title", Operator: "contains", Value: "Home Assistant", Weight: 85},
			},
			CheckTemplates: []domain.ServiceDefinitionCheckTemplate{
				{Name: "Home Assistant API", Type: domain.CheckTypeHTTP, Path: "/api/", Method: "GET", IntervalSeconds: 60, TimeoutSeconds: 10, ExpectedStatusMin: 200, ExpectedStatusMax: 399, Enabled: true, SortOrder: 0, ConfigSource: domain.HealthCheckConfigSourceDefinition},
			},
		},
		{
			ID:       "builtin_plex",
			Key:      "plex",
			Name:     "Plex",
			Icon:     "plex",
			Priority: 80,
			BuiltIn:  true,
			Enabled:  true,
			Matchers: []domain.ServiceDefinitionMatcher{
				{ID: "builtin_plex_image", Type: "container_image", Operator: "contains", Value: "plex", Weight: 90},
				{ID: "builtin_plex_port", Type: "port", Operator: "exact", Value: "32400", Weight: 70},
				{ID: "builtin_plex_mdns", Type: "mdns_service", Operator: "exact", Value: "_plexmediasvr._tcp", Weight: 90},
			},
			CheckTemplates: []domain.ServiceDefinitionCheckTemplate{
				{Name: "Plex connectivity", Type: domain.CheckTypeTCP, IntervalSeconds: 60, TimeoutSeconds: 10, Enabled: true, SortOrder: 0, ConfigSource: domain.HealthCheckConfigSourceDefinition},
			},
		},
	}
}

func DefaultCandidatePaths() []string {
	return []string{"/", "/admin", "/health", "/status", "/api/health", "/-/healthy", "/api/"}
}

func MergeDefinitions(custom []domain.ServiceDefinition) []domain.ServiceDefinition {
	items := append([]domain.ServiceDefinition{}, BuiltInDefinitions()...)
	items = append(items, custom...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority == items[j].Priority {
			if items[i].BuiltIn == items[j].BuiltIn {
				return items[i].Name < items[j].Name
			}
			return !items[i].BuiltIn && items[j].BuiltIn
		}
		return items[i].Priority > items[j].Priority
	})
	return items
}

func CandidateFromService(item domain.Service) Candidate {
	return Candidate{
		Name:        item.Name,
		ServiceType: item.ServiceType,
		Icon:        item.Icon,
		Scheme:      item.Scheme,
		Host:        item.Host,
		HostValue:   item.HostValue,
		Port:        item.Port,
		Path:        item.Path,
		Details:     item.Details,
	}
}

func CandidateFromDiscoveredService(item domain.DiscoveredService) Candidate {
	details := map[string]any{}
	maps.Copy(details, item.Details)
	for _, evidence := range item.Evidence {
		if image := stringMapValue(evidence.Details, "image"); image != "" && stringMapValue(details, "image") == "" {
			details["image"] = image
		}
		if title := stringMapValue(evidence.Details, "pageTitle"); title != "" && stringMapValue(details, "pageTitle") == "" {
			details["pageTitle"] = title
		}
		if body := stringMapValue(evidence.Details, "bodySnippet"); body != "" && stringMapValue(details, "bodySnippet") == "" {
			details["bodySnippet"] = body
		}
		if mdns := stringMapValue(evidence.Details, "mdnsService"); mdns != "" && stringMapValue(details, "mdnsService") == "" {
			details["mdnsService"] = mdns
		}
	}
	return Candidate{
		Name:        item.Name,
		ServiceType: item.ServiceType,
		Icon:        item.Icon,
		Scheme:      item.Scheme,
		Host:        item.Host,
		HostValue:   item.HostValue,
		Port:        item.Port,
		Path:        item.Path,
		Details:     details,
	}
}

func MatchDefinitions(definitions []domain.ServiceDefinition, candidate Candidate) (domain.ServiceDefinitionMatch, bool) {
	best := domain.ServiceDefinitionMatch{}
	found := false
	for _, definition := range definitions {
		if !definition.Enabled {
			continue
		}
		score := 0
		reasons := []string{}
		for _, matcher := range definition.Matchers {
			if matched, reason := matchCandidate(matcher, candidate); matched {
				score += matcher.Weight
				reasons = append(reasons, reason)
			}
		}
		if score == 0 {
			continue
		}
		if !found || score > best.Score || (score == best.Score && definition.Priority > best.Definition.Priority) {
			best = domain.ServiceDefinitionMatch{
				Definition: definition,
				Score:      score,
				Reasons:    reasons,
			}
			found = true
		}
	}
	return best, found
}

func InstantiateChecks(subjectType domain.HealthCheckSubjectType, subjectID string, addressSource domain.ServiceAddressSource, hostValue, host, scheme string, port int, path string, definition domain.ServiceDefinition) []domain.ServiceCheck {
	checks := make([]domain.ServiceCheck, 0, len(definition.CheckTemplates))
	for index, template := range definition.CheckTemplates {
		checkHostValue := firstNonEmpty(template.HostValue, hostValue, host)
		checkHost := firstNonEmpty(host, checkHostValue)
		checkProtocol := firstNonEmpty(template.Protocol, scheme)
		checkPort := port
		if checkPort <= 0 {
			checkPort = template.Port
		}
		checkPath := firstNonEmpty(template.Path, path)
		if template.Type != domain.CheckTypeHTTP && template.Path != "" {
			checkPath = ""
		}
		check := domain.ServiceCheck{
			SubjectType:         subjectType,
			SubjectID:           subjectID,
			ServiceID:           subjectID,
			Name:                firstNonEmpty(template.Name, fmt.Sprintf("%s check", strings.ToUpper(string(template.Type)))),
			Type:                template.Type,
			Protocol:            checkProtocol,
			AddressSource:       firstAddressSource(template.AddressSource, addressSource, domain.ServiceAddressLiteralHost),
			HostValue:           checkHostValue,
			Host:                checkHost,
			Port:                checkPort,
			Path:                checkPath,
			Method:              firstNonEmpty(template.Method, "GET"),
			IntervalSeconds:     defaultInt(template.IntervalSeconds, 60),
			TimeoutSeconds:      defaultInt(template.TimeoutSeconds, 10),
			ExpectedStatusMin:   defaultStatusMin(template.ExpectedStatusMin, template.Type),
			ExpectedStatusMax:   defaultStatusMax(template.ExpectedStatusMax, template.Type),
			Enabled:             template.Enabled,
			SortOrder:           template.SortOrder,
			ConfigSource:        firstConfigSource(template.ConfigSource, domain.HealthCheckConfigSourceDefinition),
			ServiceDefinitionID: definition.ID,
		}
		if !check.Enabled {
			check.Enabled = true
		}
		if check.SortOrder == 0 {
			check.SortOrder = index
		}
		check.Target = ResolveCheckTarget(check)
		checks = append(checks, check)
	}
	return checks
}

func ResolveCheckTarget(check domain.ServiceCheck) string {
	host := firstNonEmpty(check.Host, check.HostValue)
	switch check.Type {
	case domain.CheckTypeHTTP:
		return buildURL(firstNonEmpty(check.Protocol, "http"), host, check.Port, check.Path)
	case domain.CheckTypeTCP:
		if host == "" || check.Port <= 0 {
			return ""
		}
		return fmt.Sprintf("%s:%d", host, check.Port)
	case domain.CheckTypePing:
		return host
	default:
		return ""
	}
}

func matchCandidate(matcher domain.ServiceDefinitionMatcher, candidate Candidate) (bool, string) {
	switch matcher.Type {
	case "port":
		if candidate.Port <= 0 {
			return false, ""
		}
		return compareString(matcher.Operator, strconv.Itoa(candidate.Port), matcher.Value), fmt.Sprintf("port=%d", candidate.Port)
	case "container_image":
		image := strings.ToLower(strings.TrimSpace(stringMapValue(candidate.Details, "image")))
		if image == "" {
			return false, ""
		}
		return compareString(matcher.Operator, image, matcher.Value), "image"
	case "mdns_service":
		mdnsService := strings.ToLower(strings.TrimSpace(stringMapValue(candidate.Details, "mdnsService")))
		if mdnsService == "" {
			return false, ""
		}
		return compareString(matcher.Operator, mdnsService, matcher.Value), "mdns"
	case "page_title":
		title := stringMapValue(candidate.Details, "pageTitle")
		if title == "" {
			return false, ""
		}
		return compareString(matcher.Operator, title, matcher.Value), "pageTitle"
	case "body_substring":
		body := stringMapValue(candidate.Details, "bodySnippet")
		if body == "" {
			return false, ""
		}
		return compareString("contains", body, matcher.Value), "bodySnippet"
	case "http_header":
		headers := nestedStringMap(candidate.Details, "httpHeaders")
		if len(headers) == 0 {
			return false, ""
		}
		headerValue := headers[strings.ToLower(strings.TrimSpace(matcher.Extra))]
		if matcher.Operator == "exists" {
			return headerValue != "", matcher.Extra
		}
		if headerValue == "" {
			return false, ""
		}
		return compareString(matcher.Operator, headerValue, matcher.Value), matcher.Extra
	default:
		return false, ""
	}
}

func compareString(operator, actual, expected string) bool {
	actual = strings.ToLower(strings.TrimSpace(actual))
	expected = strings.ToLower(strings.TrimSpace(expected))
	switch operator {
	case "", "exact":
		return actual == expected
	case "contains":
		return strings.Contains(actual, expected)
	case "exists":
		return actual != ""
	default:
		return actual == expected
	}
}

func nestedStringMap(values map[string]any, key string) map[string]string {
	if values == nil {
		return nil
	}
	raw, ok := values[key]
	if !ok {
		return nil
	}
	switch typed := raw.(type) {
	case map[string]string:
		out := map[string]string{}
		for nestedKey, nestedValue := range typed {
			out[strings.ToLower(strings.TrimSpace(nestedKey))] = nestedValue
		}
		return out
	case map[string]any:
		out := map[string]string{}
		for nestedKey, nestedValue := range typed {
			stringValue, ok := nestedValue.(string)
			if !ok {
				continue
			}
			out[strings.ToLower(strings.TrimSpace(nestedKey))] = stringValue
		}
		return out
	default:
		return nil
	}
}

func stringMapValue(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	raw, ok := values[key]
	if !ok {
		return ""
	}
	stringValue, ok := raw.(string)
	if !ok {
		return ""
	}
	return stringValue
}

func buildURL(scheme, host string, port int, path string) string {
	if host == "" {
		return ""
	}
	base := fmt.Sprintf("%s://%s", firstNonEmpty(scheme, "http"), host)
	if port > 0 && port != 80 && port != 443 {
		base = fmt.Sprintf("%s:%d", base, port)
	}
	if strings.TrimSpace(path) == "" {
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

func defaultInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func defaultStatusMin(value int, checkType domain.CheckType) int {
	if value > 0 {
		return value
	}
	if checkType == domain.CheckTypeHTTP {
		return 200
	}
	return 0
}

func defaultStatusMax(value int, checkType domain.CheckType) int {
	if value > 0 {
		return value
	}
	if checkType == domain.CheckTypeHTTP {
		return 399
	}
	return 0
}

func firstAddressSource(values ...domain.ServiceAddressSource) domain.ServiceAddressSource {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstConfigSource(values ...domain.HealthCheckConfigSource) domain.HealthCheckConfigSource {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
