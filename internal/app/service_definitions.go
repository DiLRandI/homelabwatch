package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/monitoring"
)

func (a *App) ListServiceDefinitions(ctx context.Context) ([]domain.ServiceDefinition, error) {
	return a.store.ListServiceDefinitions(ctx)
}

func (a *App) SaveServiceDefinition(ctx context.Context, input domain.ServiceDefinitionInput) (domain.ServiceDefinition, error) {
	item, err := a.store.SaveServiceDefinition(ctx, input)
	if err != nil {
		return domain.ServiceDefinition{}, err
	}
	if err := a.reapplyServiceDefinitions(ctx); err != nil {
		return domain.ServiceDefinition{}, err
	}
	a.publish("service-definition", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteServiceDefinition(ctx context.Context, id string) error {
	if err := a.store.DeleteServiceDefinition(ctx, id); err != nil {
		return err
	}
	if err := a.reapplyServiceDefinitions(ctx); err != nil {
		return err
	}
	a.publish("service-definition", id, "deleted", nil)
	return nil
}

func (a *App) ReapplyServiceDefinition(ctx context.Context, id string) error {
	if err := a.reapplyServiceDefinitions(ctx); err != nil {
		return err
	}
	a.publish("service-definition", id, "reapplied", map[string]string{"id": id})
	return nil
}

func (a *App) TestServiceCheck(ctx context.Context, serviceID string, input domain.EndpointTestInput) (domain.EndpointTestResult, error) {
	service, err := a.store.GetService(ctx, serviceID)
	if err != nil {
		return domain.EndpointTestResult{}, err
	}
	check := input.Check
	check.ServiceID = serviceID
	check.SubjectID = firstNonEmpty(strings.TrimSpace(check.SubjectID), serviceID)
	if check.SubjectType == "" {
		check.SubjectType = domain.HealthCheckSubjectService
	}
	check.Type = normalizedCheckType(check.Type)
	check.Protocol = normalizedProtocol(
		check.Type,
		firstNonEmpty(check.Protocol, service.HealthScheme, service.Scheme),
	)
	check.AddressSource = firstNonEmptyAddressSource(
		check.AddressSource,
		service.HealthAddressSource,
		service.AddressSource,
	)
	check.HostValue = firstNonEmpty(
		strings.TrimSpace(check.HostValue),
		strings.TrimSpace(check.Host),
		service.HealthHostValue,
		service.HostValue,
		service.HealthHost,
		service.Host,
	)
	check.Host = resolveAddressSourceHost(
		check.AddressSource,
		check.HostValue,
		firstNonEmpty(service.HealthHost, service.Host),
	)
	if check.Port == 0 {
		check.Port = nonZeroOr(service.HealthPort, service.Port)
	}
	check.Path = normalizePath(check.Path)
	check.Method = ensureMethod(check.Method)
	check.IntervalSeconds = nonZeroOr(check.IntervalSeconds, 60)
	check.TimeoutSeconds = nonZeroOr(check.TimeoutSeconds, 10)
	check.ExpectedStatusMin = defaultStatusMin(check.ExpectedStatusMin, check.Type)
	check.ExpectedStatusMax = defaultStatusMax(check.ExpectedStatusMax, check.Type)
	check.Target = resolvedCheckTarget(check)

	switch check.Type {
	case domain.CheckTypeHTTP:
		return a.testHTTPServiceCheck(ctx, service, check, input.DiscoverPaths || check.Path == "")
	default:
		result := monitoring.RunAdhocCheck(ctx, check)
		return domain.EndpointTestResult{
			Check:             check,
			ResolvedURL:       result.ResolvedTarget,
			Status:            result.Status,
			HTTPStatusCode:    result.HTTPStatusCode,
			LatencyMS:         result.LatencyMS,
			ResponseSizeBytes: result.ResponseSizeBytes,
			Message:           result.Message,
			CheckedAt:         result.CheckedAt,
		}, nil
	}
}

func (a *App) refreshDiscoveredServiceDefinitions(ctx context.Context) error {
	definitions, err := a.store.ListServiceDefinitions(ctx)
	if err != nil {
		return err
	}
	items, err := a.store.ListDiscoveredServices(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.State == domain.DiscoveryStateAccepted {
			continue
		}
		evidence := buildDefinitionEvidence(item)
		title, body, headers, _ := a.httpFingerprint(ctx, firstNonEmpty(item.URL, buildServiceURL(item.Scheme, item.Host, item.Port, item.Path)), 2*time.Second)
		evidence.PageTitle = title
		evidence.Body = body
		evidence.Headers = headers
		match := bestDefinitionMatch(definitions, evidence)
		if err := a.store.ApplyDiscoveredServiceDefinition(ctx, item.ID, match, max(item.ConfidenceScore, scoreForMatch(match)), time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) applyBestDefinitionToService(ctx context.Context, service domain.Service) (domain.Service, error) {
	if service.HealthConfigMode == domain.HealthConfigModeCustom {
		return service, nil
	}
	definitions, err := a.store.ListServiceDefinitions(ctx)
	if err != nil {
		return domain.Service{}, err
	}
	evidence := buildServiceDefinitionEvidence(service)
	title, body, headers, _ := a.httpFingerprint(ctx, service.URL, 2*time.Second)
	evidence.PageTitle = title
	evidence.Body = body
	evidence.Headers = headers
	match := bestDefinitionMatch(definitions, evidence)
	if match == nil {
		if err := a.store.SyncServiceHealthChecks(ctx, service.ID); err != nil {
			return domain.Service{}, err
		}
		return a.store.GetService(ctx, service.ID)
	}
	service.ServiceDefinitionID = match.Definition.ID
	service.ServiceType = firstNonEmpty(match.Definition.Key, service.ServiceType)
	service.Icon = firstNonEmpty(match.Definition.Icon, service.Icon)
	service.FingerprintedAt = time.Now().UTC()
	saved, err := a.store.SaveManualService(ctx, service)
	if err != nil {
		return domain.Service{}, err
	}
	if err := a.store.SyncServiceHealthChecks(ctx, saved.ID); err != nil {
		return domain.Service{}, err
	}
	return a.store.GetService(ctx, saved.ID)
}

func (a *App) reapplyServiceDefinitions(ctx context.Context) error {
	services, err := a.store.ListServices(ctx)
	if err != nil {
		return err
	}
	for _, service := range services {
		if service.HealthConfigMode == domain.HealthConfigModeCustom {
			continue
		}
		if _, err := a.applyBestDefinitionToService(ctx, service); err != nil {
			return err
		}
	}
	return a.refreshDiscoveredServiceDefinitions(ctx)
}

func (a *App) testHTTPServiceCheck(ctx context.Context, service domain.Service, check domain.ServiceCheck, discoverPaths bool) (domain.EndpointTestResult, error) {
	definitions, err := a.store.ListServiceDefinitions(ctx)
	if err != nil {
		return domain.EndpointTestResult{}, err
	}
	candidates := []string{}
	if check.Path != "" {
		candidates = append(candidates, normalizePath(check.Path))
	}
	if len(candidates) == 0 && discoverPaths {
		candidates = append(candidates, preferredDefinitionPaths(definitions, buildServiceDefinitionEvidence(service))...)
		candidates = append(candidates, "", "/admin", "/health", "/status", "/api/health", "/-/healthy", "/api/")
	}
	if len(candidates) == 0 {
		candidates = []string{normalizePath(firstNonEmpty(service.HealthPath, service.Path))}
	}

	seen := map[string]struct{}{}
	var lastResult domain.EndpointTestResult
	for _, path := range candidates {
		path = normalizePath(path)
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		candidate := check
		candidate.Path = path
		candidate.Target = resolvedCheckTarget(candidate)
		result, title, body, headers, err := a.runHTTPCheckProbe(ctx, candidate)
		if err != nil {
			lastResult = domain.EndpointTestResult{
				Check:     candidate,
				Message:   err.Error(),
				Status:    domain.HealthStatusUnhealthy,
				CheckedAt: time.Now().UTC(),
			}
			continue
		}
		evidence := buildServiceDefinitionEvidence(service)
		evidence.PageTitle = title
		evidence.Body = body
		evidence.Headers = headers
		match := bestDefinitionMatch(definitions, evidence)
		result.Check = candidate
		if match != nil {
			result.MatchedServiceDefinition = &match.Definition
		}
		if result.Status == domain.HealthStatusHealthy {
			result.Check.Path = path
			return result, nil
		}
		lastResult = result
	}
	if lastResult.CheckedAt.IsZero() {
		lastResult = domain.EndpointTestResult{
			Check:     check,
			Status:    domain.HealthStatusUnknown,
			Message:   "no endpoint candidates were tested",
			CheckedAt: time.Now().UTC(),
		}
	}
	return lastResult, nil
}

func (a *App) runHTTPCheckProbe(ctx context.Context, check domain.ServiceCheck) (domain.EndpointTestResult, string, string, http.Header, error) {
	timeout := time.Duration(nonZeroOr(check.TimeoutSeconds, 10)) * time.Second
	request, err := http.NewRequestWithContext(ctx, ensureMethod(check.Method), resolvedCheckTarget(check), nil)
	if err != nil {
		return domain.EndpointTestResult{}, "", "", nil, err
	}
	client := &http.Client{Timeout: timeout}
	start := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return domain.EndpointTestResult{}, "", "", nil, err
	}
	defer response.Body.Close()
	limited, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		return domain.EndpointTestResult{}, "", "", nil, err
	}
	latency := time.Since(start).Milliseconds()
	body := string(limited)
	result := domain.EndpointTestResult{
		ResolvedURL:       request.URL.String(),
		Status:            domain.HealthStatusUnhealthy,
		HTTPStatusCode:    response.StatusCode,
		LatencyMS:         latency,
		ResponseSizeBytes: int64(len(limited)),
		Message:           response.Status,
		CheckedAt:         time.Now().UTC(),
	}
	if response.StatusCode >= check.ExpectedStatusMin && response.StatusCode <= check.ExpectedStatusMax {
		result.Status = domain.HealthStatusHealthy
	}
	return result, extractTitle(body), body, response.Header.Clone(), nil
}

func (a *App) httpFingerprint(ctx context.Context, rawURL string, timeout time.Duration) (string, string, http.Header, error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", "", nil, errors.New("empty url")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", nil, err
	}
	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return "", "", nil, err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		return "", "", nil, err
	}
	body := string(payload)
	return extractTitle(body), body, response.Header.Clone(), nil
}

type definitionEvidence struct {
	Port        int
	Image       string
	MDNSService string
	Labels      map[string]string
	Headers     http.Header
	PageTitle   string
	Body        string
}

func buildDefinitionEvidence(item domain.DiscoveredService) definitionEvidence {
	evidence := definitionEvidence{
		Port:   item.Port,
		Labels: map[string]string{},
	}
	for _, probe := range item.Evidence {
		if image := stringMapString(probe.Details, "image"); image != "" && evidence.Image == "" {
			evidence.Image = image
		}
		if mdns := stringMapString(probe.Details, "mdnsService"); mdns != "" && evidence.MDNSService == "" {
			evidence.MDNSService = mdns
		}
		maps.Copy(evidence.Labels, mapStringValues(probe.Details["labels"]))
	}
	return evidence
}

func buildServiceDefinitionEvidence(service domain.Service) definitionEvidence {
	evidence := definitionEvidence{
		Port:   service.Port,
		Labels: map[string]string{},
	}
	if service.Details != nil {
		evidence.Image = stringMapString(service.Details, "image")
		maps.Copy(evidence.Labels, mapStringValues(service.Details["labels"]))
	}
	return evidence
}

func bestDefinitionMatch(definitions []domain.ServiceDefinition, evidence definitionEvidence) *domain.ServiceDefinitionMatch {
	matches := make([]domain.ServiceDefinitionMatch, 0)
	for _, definition := range definitions {
		if !definition.Enabled {
			continue
		}
		score := 0
		reasons := make([]string, 0)
		for _, matcher := range definition.Matchers {
			if matcherMatches(matcher, evidence) {
				score += nonZeroOr(matcher.Weight, 50)
				reasons = append(reasons, matcherDescription(matcher))
			}
		}
		if score == 0 {
			continue
		}
		matches = append(matches, domain.ServiceDefinitionMatch{
			Definition: definition,
			Score:      score,
			Reasons:    reasons,
		})
	}
	if len(matches) == 0 {
		return nil
	}
	sort.Slice(matches, func(left, right int) bool {
		if matches[left].Score == matches[right].Score {
			if matches[left].Definition.Priority == matches[right].Definition.Priority {
				if matches[left].Definition.BuiltIn == matches[right].Definition.BuiltIn {
					return strings.Compare(matches[left].Definition.Name, matches[right].Definition.Name) < 0
				}
				return !matches[left].Definition.BuiltIn
			}
			return matches[left].Definition.Priority > matches[right].Definition.Priority
		}
		return matches[left].Score > matches[right].Score
	})
	return &matches[0]
}

func matcherMatches(matcher domain.ServiceDefinitionMatcher, evidence definitionEvidence) bool {
	operator := strings.ToLower(strings.TrimSpace(firstNonEmpty(matcher.Operator, "equals")))
	switch matcher.Type {
	case "port":
		return compareStrings(operator, strconv.Itoa(evidence.Port), matcher.Value)
	case "container_image":
		return compareStrings(operator, evidence.Image, matcher.Value)
	case "mdns_service":
		return compareStrings(operator, evidence.MDNSService, matcher.Value)
	case "page_title":
		return compareStrings(operator, evidence.PageTitle, matcher.Value)
	case "body_substring":
		return compareStrings(operator, evidence.Body, matcher.Value)
	case "http_header":
		headerName := textKey(matcher.Extra)
		if headerName == "" {
			return false
		}
		return compareStrings(operator, evidence.Headers.Get(headerName), matcher.Value)
	case "docker_label":
		key := strings.TrimSpace(matcher.Extra)
		if key == "" {
			return false
		}
		return compareStrings(operator, evidence.Labels[key], matcher.Value)
	default:
		return false
	}
}

func preferredDefinitionPaths(definitions []domain.ServiceDefinition, evidence definitionEvidence) []string {
	match := bestDefinitionMatch(definitions, evidence)
	if match == nil {
		return nil
	}
	paths := make([]string, 0, len(match.Definition.CheckTemplates))
	for _, template := range match.Definition.CheckTemplates {
		if template.Type != domain.CheckTypeHTTP {
			continue
		}
		paths = append(paths, normalizePath(template.Path))
	}
	return paths
}

func compareStrings(operator, actual, expected string) bool {
	actual = strings.TrimSpace(strings.ToLower(actual))
	expected = strings.TrimSpace(strings.ToLower(expected))
	switch operator {
	case "contains":
		return actual != "" && expected != "" && strings.Contains(actual, expected)
	default:
		return actual != "" && actual == expected
	}
}

func extractTitle(body string) string {
	lower := strings.ToLower(body)
	start := strings.Index(lower, "<title>")
	end := strings.Index(lower, "</title>")
	if start < 0 || end <= start {
		return ""
	}
	return strings.TrimSpace(body[start+len("<title>") : end])
}

func textKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.Split(value, "-")
	for index := range parts {
		if parts[index] == "" {
			continue
		}
		parts[index] = strings.ToUpper(parts[index][:1]) + strings.ToLower(parts[index][1:])
	}
	return strings.Join(parts, "-")
}

func stringMapString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if raw, ok := values[key]; ok {
		switch typed := raw.(type) {
		case string:
			return typed
		}
	}
	return ""
}

func mapStringValues(raw any) map[string]string {
	values := map[string]string{}
	switch typed := raw.(type) {
	case map[string]string:
		maps.Copy(values, typed)
	case map[string]any:
		for key, value := range typed {
			if stringValue, ok := value.(string); ok {
				values[key] = stringValue
			}
		}
	}
	return values
}

func scoreForMatch(match *domain.ServiceDefinitionMatch) int {
	if match == nil {
		return 0
	}
	return match.Score
}

func normalizedCheckType(value domain.CheckType) domain.CheckType {
	switch value {
	case domain.CheckTypeTCP, domain.CheckTypePing:
		return value
	default:
		return domain.CheckTypeHTTP
	}
}

func normalizedProtocol(checkType domain.CheckType, value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value != "" {
		return value
	}
	switch checkType {
	case domain.CheckTypeTCP:
		return "tcp"
	case domain.CheckTypePing:
		return "ping"
	default:
		return "http"
	}
}

func firstNonEmptyAddressSource(values ...domain.ServiceAddressSource) domain.ServiceAddressSource {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func ensureMethod(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return http.MethodGet
	}
	return method
}

func nonZeroOr(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func defaultStatusMin(value int, kind domain.CheckType) int {
	if value > 0 || kind != domain.CheckTypeHTTP {
		return value
	}
	return 200
}

func defaultStatusMax(value int, kind domain.CheckType) int {
	if value > 0 || kind != domain.CheckTypeHTTP {
		return value
	}
	return 399
}

func resolvedCheckTarget(check domain.ServiceCheck) string {
	switch check.Type {
	case domain.CheckTypeTCP:
		host := firstNonEmpty(check.Host, check.HostValue)
		if host == "" || check.Port == 0 {
			return ""
		}
		return fmt.Sprintf("%s:%d", host, check.Port)
	case domain.CheckTypePing:
		return firstNonEmpty(check.Host, check.HostValue)
	default:
		host := firstNonEmpty(check.Host, check.HostValue)
		if host == "" {
			return ""
		}
		base := fmt.Sprintf("%s://%s", normalizedProtocol(check.Type, check.Protocol), host)
		if check.Port > 0 && check.Port != 80 && check.Port != 443 {
			base = fmt.Sprintf("%s:%d", base, check.Port)
		}
		if check.Path == "" {
			return base
		}
		path := normalizePath(check.Path)
		if path == "" {
			return base
		}
		return base + path
	}
}

func matcherDescription(matcher domain.ServiceDefinitionMatcher) string {
	if matcher.Extra != "" {
		return fmt.Sprintf("%s:%s=%s", matcher.Type, matcher.Extra, matcher.Value)
	}
	return fmt.Sprintf("%s=%s", matcher.Type, matcher.Value)
}

func resolveAddressSourceHost(addressSource domain.ServiceAddressSource, hostValue, devicePrimaryAddress string) string {
	switch addressSource {
	case domain.ServiceAddressDevicePrimary:
		return firstNonEmpty(devicePrimaryAddress, hostValue)
	case domain.ServiceAddressMDNSHostname:
		return firstNonEmpty(hostValue, devicePrimaryAddress)
	default:
		return firstNonEmpty(hostValue, devicePrimaryAddress)
	}
}

func buildServiceURL(scheme, host string, port int, path string) string {
	if scheme == "" {
		if port == 443 || port == 8443 || port == 9443 {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	if host == "" {
		return ""
	}
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
