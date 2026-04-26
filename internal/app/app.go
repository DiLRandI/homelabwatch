package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/discovery/docker"
	"github.com/deleema/homelabwatch/internal/discovery/lan"
	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/monitoring"
	"github.com/deleema/homelabwatch/internal/notifications"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
	"github.com/deleema/homelabwatch/internal/worker"
)

const eventsRetention = 14 * 24 * time.Hour

type App struct {
	config        config.Config
	logger        *slog.Logger
	store         *sqlite.Store
	bus           *events.Bus
	docker        *docker.Provider
	lan           *lan.Provider
	monitor       *monitoring.Runner
	notifications *notifications.Engine
	scheduler     *worker.Scheduler
}

func New(cfg config.Config, store *sqlite.Store, bus *events.Bus, logger *slog.Logger) *App {
	if logger == nil {
		logger = slog.Default()
	}
	instance := &App{
		config:        cfg,
		logger:        logger.With("component", "app"),
		store:         store,
		bus:           bus,
		docker:        docker.NewProvider(),
		lan:           lan.NewProvider(),
		monitor:       monitoring.NewRunner(store),
		notifications: notifications.NewEngine(store, bus, logger),
	}
	instance.scheduler = worker.NewScheduler(
		logger.With("component", "worker"),
		worker.Job{Name: "docker-sync", Interval: 30 * time.Second, Run: instance.runDockerDiscovery},
		worker.Job{Name: "lan-scan", Interval: 5 * time.Minute, Run: instance.runLANDiscovery},
		worker.Job{Name: "service-fingerprinting", Interval: 45 * time.Second, Run: instance.runFingerprinting},
		worker.Job{Name: "health-checks", Interval: 30 * time.Second, Run: instance.runMonitoring},
		worker.Job{Name: "cleanup", Interval: 24 * time.Hour, Run: instance.runCleanup},
	)
	return instance
}

func (a *App) Start(ctx context.Context) {
	a.logger.Info("starting application services")
	a.notifications.Start(ctx)
	a.scheduler.Start(ctx)
}

func (a *App) SubscribeEvents(buffer int) chan domain.EventEnvelope {
	return a.bus.Subscribe(buffer)
}

func (a *App) UnsubscribeEvents(ch chan domain.EventEnvelope) {
	a.bus.Unsubscribe(ch)
}

func (a *App) BootstrapStatus(ctx context.Context) (domain.BootstrapStatus, error) {
	return a.store.BootstrapStatus(ctx)
}

func (a *App) Setup(ctx context.Context, input domain.SetupInput) error {
	status, err := a.store.BootstrapStatus(ctx)
	if err != nil {
		return err
	}
	if status.Initialized {
		return errors.New("bootstrap already completed")
	}
	if len(input.DefaultScanPorts) == 0 {
		input.DefaultScanPorts = append([]int(nil), a.config.DefaultScanPorts...)
	}
	if strings.TrimSpace(input.ApplianceName) == "" {
		if hostname, err := os.Hostname(); err == nil && strings.TrimSpace(hostname) != "" {
			input.ApplianceName = hostname
		} else {
			input.ApplianceName = "HomelabWatch"
		}
	}
	input.DockerEndpoints = a.seedDockerEndpoints(input.DockerEndpoints)
	input.ScanTargets, err = a.seedScanTargets(input.ScanTargets, input.DefaultScanPorts)
	if err != nil {
		return err
	}
	if err := a.store.Initialize(ctx, input); err != nil {
		return err
	}
	a.publish("bootstrap", "app", "initialized", map[string]any{
		"applianceName":    input.ApplianceName,
		"autoScanEnabled":  input.AutoScanEnabled,
		"defaultScanPorts": input.DefaultScanPorts,
	})
	if input.RunDiscovery {
		go func() {
			_ = a.TriggerDiscovery(context.Background())
		}()
	}
	return nil
}

func (a *App) ValidateAPIToken(ctx context.Context, token string, requiredScope domain.TokenScope) (bool, error) {
	return a.store.ValidateAPIToken(ctx, token, requiredScope)
}

func (a *App) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	return a.store.GetDashboard(ctx)
}

func (a *App) Settings(ctx context.Context) (domain.SettingsView, error) {
	return a.store.GetSettingsView(ctx)
}

func (a *App) CreateAPIToken(ctx context.Context, input domain.CreateAPITokenInput) (domain.CreatedAPIToken, error) {
	return a.store.CreateAPIToken(ctx, input)
}

func (a *App) RevokeAPIToken(ctx context.Context, id string) error {
	return a.store.RevokeAPIToken(ctx, id)
}

func (a *App) ListServices(ctx context.Context) ([]domain.Service, error) {
	return a.store.ListServices(ctx)
}

func (a *App) GetService(ctx context.Context, id string) (domain.Service, error) {
	return a.store.GetService(ctx, id)
}

func (a *App) SaveManualService(ctx context.Context, service domain.Service) (domain.Service, error) {
	service = normalizeService(service)
	item, err := a.store.SaveManualService(ctx, service)
	if err != nil {
		return domain.Service{}, err
	}
	item, err = a.applyBestDefinitionToService(ctx, item)
	if err != nil {
		return domain.Service{}, err
	}
	a.publish("service", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteService(ctx context.Context, id string) error {
	if err := a.store.DeleteService(ctx, id); err != nil {
		return err
	}
	a.publish("service", id, "deleted", nil)
	return nil
}

func (a *App) ListServiceEvents(ctx context.Context, serviceID string) ([]domain.ServiceEvent, error) {
	return a.store.ListServiceEvents(ctx, serviceID, 50)
}

func (a *App) SaveServiceCheck(ctx context.Context, check domain.ServiceCheck) (domain.ServiceCheck, error) {
	item, err := a.store.SaveServiceCheck(ctx, check)
	if err != nil {
		return domain.ServiceCheck{}, err
	}
	a.publish("check", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteServiceCheck(ctx context.Context, id string) error {
	if err := a.store.DeleteServiceCheck(ctx, id); err != nil {
		return err
	}
	a.publish("check", id, "deleted", nil)
	return nil
}

func (a *App) ListDevices(ctx context.Context) ([]domain.Device, error) {
	return a.store.ListDevices(ctx)
}

func (a *App) GetDevice(ctx context.Context, id string) (domain.Device, error) {
	return a.store.GetDevice(ctx, id)
}

func (a *App) UpdateDevice(ctx context.Context, id string, displayName *string, hidden *bool) (domain.Device, error) {
	item, err := a.store.UpdateDevice(ctx, id, displayName, hidden)
	if err != nil {
		return domain.Device{}, err
	}
	a.publish("device", item.ID, "updated", item)
	return item, nil
}

func (a *App) ListDockerEndpoints(ctx context.Context) ([]domain.DockerEndpoint, error) {
	return a.store.ListDockerEndpoints(ctx)
}

func (a *App) SaveDockerEndpoint(ctx context.Context, endpoint domain.DockerEndpoint) (domain.DockerEndpoint, error) {
	item, err := a.store.SaveDockerEndpoint(ctx, endpoint)
	if err != nil {
		return domain.DockerEndpoint{}, err
	}
	a.publish("docker-endpoint", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteDockerEndpoint(ctx context.Context, id string) error {
	if err := a.store.DeleteDockerEndpoint(ctx, id); err != nil {
		return err
	}
	a.publish("docker-endpoint", id, "deleted", nil)
	return nil
}

func (a *App) ListScanTargets(ctx context.Context) ([]domain.ScanTarget, error) {
	return a.store.ListScanTargets(ctx)
}

func (a *App) SaveScanTarget(ctx context.Context, target domain.ScanTarget) (domain.ScanTarget, error) {
	item, err := a.store.SaveScanTarget(ctx, target)
	if err != nil {
		return domain.ScanTarget{}, err
	}
	a.publish("scan-target", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteScanTarget(ctx context.Context, id string) error {
	if err := a.store.DeleteScanTarget(ctx, id); err != nil {
		return err
	}
	a.publish("scan-target", id, "deleted", nil)
	return nil
}

func (a *App) TriggerDiscovery(ctx context.Context) error {
	if err := a.runDockerDiscovery(ctx); err != nil {
		return err
	}
	return a.runLANDiscovery(ctx)
}

func (a *App) TriggerMonitoring(ctx context.Context) error {
	return a.runMonitoring(ctx)
}

func (a *App) runDockerDiscovery(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		a.logger.Debug("docker discovery skipped", "reason", "not_bootstrapped")
		return nil
	}
	endpoints, err := a.store.ListDockerEndpoints(ctx)
	if err != nil {
		a.recordJobRun(ctx, "docker-sync", err)
		return err
	}
	results := a.docker.Discover(ctx, endpoints)
	var runErr error
	processedEndpoints := 0
	failedEndpoints := 0
	serviceObservations := 0
	for _, result := range results {
		processedEndpoints++
		for _, observation := range result.Observations {
			serviceObservations += len(observation.Services)
		}
		if result.Err != nil && runErr == nil {
			runErr = result.Err
		}
		if result.Err != nil {
			failedEndpoints++
			a.logger.Warn("docker endpoint discovery failed",
				"endpoint_id", result.Endpoint.ID,
				"endpoint_name", result.Endpoint.Name,
				"err", result.Err,
			)
		}
		successAt := time.Time{}
		lastError := ""
		if result.Err == nil {
			successAt = time.Now().UTC()
			if err := a.processObservations(ctx, result.Observations); err != nil && runErr == nil {
				runErr = err
			}
		} else {
			lastError = result.Err.Error()
		}
		_ = a.store.UpdateDockerEndpointStatus(ctx, result.Endpoint.ID, successAt, lastError)
	}
	a.logger.Debug("docker discovery completed",
		"endpoints", processedEndpoints,
		"failed_endpoints", failedEndpoints,
		"service_observations", serviceObservations,
	)
	a.recordJobRun(ctx, "docker-sync", runErr)
	return runErr
}

func (a *App) runLANDiscovery(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		a.logger.Debug("lan scan skipped", "reason", "not_bootstrapped")
		return nil
	}
	settings, err := a.store.GetAppSettings(ctx)
	if err != nil {
		a.recordJobRun(ctx, "lan-scan", err)
		return err
	}
	if !settings.AutoScanEnabled {
		a.logger.Debug("lan scan skipped", "reason", "auto_scan_disabled")
		a.recordJobRun(ctx, "lan-scan", nil)
		return nil
	}
	targets, err := a.store.ListScanTargets(ctx)
	if err != nil {
		a.recordJobRun(ctx, "lan-scan", err)
		return err
	}
	observations, err := a.lan.Discover(ctx, targets)
	if err != nil && !errors.Is(err, context.Canceled) {
		a.recordJobRun(ctx, "lan-scan", err)
		return err
	}
	processErr := a.processObservations(ctx, observations)
	if processErr != nil {
		a.recordJobRun(ctx, "lan-scan", processErr)
		return processErr
	}
	a.logger.Debug("lan scan completed", "targets", len(targets), "observations", len(observations))
	a.recordJobRun(ctx, "lan-scan", nil)
	return nil
}

func (a *App) runMonitoring(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		return nil
	}
	results, err := a.monitor.RunDueChecks(ctx)
	if err != nil {
		a.recordJobRun(ctx, "health-checks", err)
		return err
	}
	for _, item := range results {
		a.publishCheckOutcome(item)
	}
	discoveredChecks, err := a.runDiscoveredMonitoring(ctx)
	if err != nil {
		a.recordJobRun(ctx, "health-checks", err)
		return err
	}
	a.logger.Debug("health checks completed", "checks", len(results), "discovered_checks", discoveredChecks)
	a.recordJobRun(ctx, "health-checks", nil)
	return nil
}

func (a *App) runFingerprinting(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		return nil
	}
	err := a.store.RefingerprintDiscoveredServices(ctx)
	if err == nil {
		err = a.refreshDiscoveredServiceDefinitions(ctx)
	}
	if err == nil {
		a.logger.Debug("service fingerprinting completed")
	}
	a.recordJobRun(ctx, "service-fingerprinting", err)
	return err
}

func (a *App) runCleanup(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		return nil
	}
	err := a.store.Cleanup(ctx, eventsRetention)
	if err == nil {
		a.logger.Debug("cleanup completed")
	}
	a.recordJobRun(ctx, "cleanup", err)
	return err
}

func (a *App) processObservations(ctx context.Context, observations []domain.Observation) error {
	settings, err := a.store.GetDiscoverySettings(ctx)
	if err != nil {
		return err
	}
	for _, observation := range observations {
		deviceID := ""
		if hasDevice(observation.Device) {
			outcome, err := a.store.UpsertDeviceObservationWithOutcome(ctx, observation.Device)
			if err != nil {
				return err
			}
			device := outcome.Device
			deviceID = device.ID
			if outcome.Created {
				a.publish("device", device.ID, "created", device)
			} else {
				a.publish("device", device.ID, "seen", device)
			}
		}
		for _, serviceObservation := range observation.Services {
			outcome, err := a.store.UpsertDiscoveredServiceObservationWithOutcome(ctx, serviceObservation, deviceID)
			if err != nil {
				return err
			}
			discovered := outcome.DiscoveredService
			if outcome.Created {
				a.publish("discovered-service", discovered.ID, "created", discovered)
			} else {
				a.publish("discovered-service", discovered.ID, "seen", discovered)
			}
			if a.shouldAutoBookmark(discovered, settings) {
				bookmark, createErr := a.CreateBookmarkFromDiscoveredService(ctx, domain.CreateBookmarkFromDiscoveredServiceInput{
					DiscoveredServiceID: discovered.ID,
				})
				if createErr == nil {
					a.publish("bookmark", bookmark.ID, "upserted", bookmark)
				}
			}
		}
	}
	return nil
}

func (a *App) publish(resource, id, action string, payload any) {
	a.bus.Publish(domain.EventEnvelope{
		Type:       resource,
		Resource:   resource,
		ID:         id,
		Action:     action,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
	})
}

func (a *App) publishCheckOutcome(outcome domain.CheckResultOutcome) {
	a.publish("check", outcome.Result.CheckID, "recorded", outcome.Result)
	if outcome.Result.SubjectType == domain.HealthCheckSubjectService || outcome.Result.ServiceID != "" {
		serviceID := firstNonEmpty(outcome.Result.ServiceID, outcome.Result.SubjectID, outcome.Check.ServiceID, outcome.Check.SubjectID)
		a.publish("status-page", "all", "health_updated", map[string]any{
			"serviceId": serviceID,
			"status":    outcome.CurrentServiceStatus,
		})
	}
	if outcome.ServiceStatusChanged {
		a.publish("service", outcome.Service.ID, "health_changed", outcome)
	}
	if outcome.CheckFailedTransition {
		a.publish("check", outcome.Check.ID, "failed", outcome)
	}
	if outcome.CheckRecoveredTransition {
		a.publish("check", outcome.Check.ID, "recovered", outcome)
	}
}

func (a *App) recordJobRun(ctx context.Context, name string, runErr error) {
	outcome, err := a.store.RecordJobRunWithOutcome(ctx, name, runErr)
	if err != nil {
		a.logger.Warn("failed to record job run", "job", name, "err", err)
		return
	}
	if outcome.Failed {
		a.publish("worker", name, "failed", outcome)
	}
	if outcome.Recovered {
		a.publish("worker", name, "recovered", outcome)
	}
}

func (a *App) seedDockerEndpoints(existing []domain.DockerEndpointSeed) []domain.DockerEndpointSeed {
	if !a.config.SeedDockerSocket {
		return existing
	}
	for _, item := range existing {
		if strings.TrimSpace(item.Address) == "unix:///var/run/docker.sock" {
			return existing
		}
	}
	return append(existing, domain.DockerEndpointSeed{
		Name:                "Local Docker",
		Kind:                "local",
		Address:             "unix:///var/run/docker.sock",
		Enabled:             true,
		ScanIntervalSeconds: 30,
	})
}

func (a *App) seedScanTargets(existing []domain.ScanTargetSeed, ports []int) ([]domain.ScanTargetSeed, error) {
	if len(existing) > 0 {
		return existing, nil
	}
	if len(a.config.SeedCIDRs) > 0 {
		targets := make([]domain.ScanTargetSeed, 0, len(a.config.SeedCIDRs))
		for _, cidr := range a.config.SeedCIDRs {
			targets = append(targets, domain.ScanTargetSeed{
				Name:                cidr,
				CIDR:                cidr,
				Enabled:             true,
				ScanIntervalSeconds: 300,
				CommonPorts:         append([]int(nil), ports...),
			})
		}
		return targets, nil
	}
	targets, err := a.lan.SuggestedTargets(ports)
	if err != nil {
		return nil, err
	}
	return targets, nil
}

func (a *App) isBootstrapped(ctx context.Context) bool {
	status, err := a.store.BootstrapStatus(ctx)
	return err == nil && status.Initialized
}

func hasDevice(device domain.DeviceObservation) bool {
	return strings.TrimSpace(device.IdentityKey) != "" ||
		strings.TrimSpace(device.IPAddress) != "" ||
		strings.TrimSpace(device.PrimaryMAC) != ""
}

func normalizeService(service domain.Service) domain.Service {
	if service.Details == nil {
		service.Details = map[string]any{}
	}
	if service.Source == "" {
		service.Source = domain.ServiceSourceManual
	}
	if service.URL != "" && (service.Host == "" || service.Scheme == "" || service.Path == "") {
		scheme, host, port, path := parseServiceURLFields(service.URL)
		service.Scheme = firstNonEmpty(service.Scheme, scheme)
		service.Host = firstNonEmpty(service.Host, host)
		service.Path = firstNonEmpty(service.Path, path)
		if service.Port == 0 {
			service.Port = port
		}
	}
	if service.HealthURL != "" {
		scheme, host, port, path := parseServiceURLFields(service.HealthURL)
		service.HealthScheme = firstNonEmpty(service.HealthScheme, scheme)
		service.HealthHostValue = firstNonEmpty(service.HealthHostValue, host)
		service.HealthHost = firstNonEmpty(service.HealthHost, host)
		service.HealthPath = firstNonEmpty(service.HealthPath, path)
		if service.HealthPort == 0 {
			service.HealthPort = port
		}
		if service.HealthAddressSource == "" {
			service.HealthAddressSource = domain.ServiceAddressLiteralHost
		}
	}
	if service.HealthAddressSource == "" {
		service.HealthAddressSource = firstNonEmptyAddressSource(
			service.AddressSource,
			domain.ServiceAddressLiteralHost,
		)
	}
	if service.HealthHostValue == "" {
		service.HealthHostValue = firstNonEmpty(service.HostValue, service.Host)
	}
	if service.HealthHost == "" {
		service.HealthHost = firstNonEmpty(service.HealthHostValue, service.HostValue, service.Host)
	}
	if service.HealthScheme == "" {
		service.HealthScheme = firstNonEmpty(service.Scheme, "http")
	}
	if service.HealthPort == 0 {
		service.HealthPort = service.Port
	}
	if service.HealthPath == "" {
		service.HealthPath = service.Path
	}
	if service.URL == "" && service.Host != "" {
		service.URL = buildURL(service.Scheme, service.Host, service.Port, service.Path)
	}
	if service.HealthURL == "" {
		healthHost := firstNonEmpty(service.HealthHost, service.HealthHostValue)
		service.HealthURL = buildURL(
			service.HealthScheme,
			healthHost,
			service.HealthPort,
			service.HealthPath,
		)
	}
	if service.HealthConfigMode == "" && serviceHasExplicitHealthTarget(service) {
		service.HealthConfigMode = domain.HealthConfigModeCustom
	}
	return service
}

func buildURL(scheme, host string, port int, path string) string {
	if scheme == "" {
		scheme = "http"
	}
	base := fmt.Sprintf("%s://%s", scheme, host)
	if port > 0 && port != 80 && port != 443 {
		base = fmt.Sprintf("%s:%d", base, port)
	}
	if path == "" || path == "/" {
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

func parseServiceURLFields(raw string) (string, string, int, string) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", 0, ""
	}
	path := parsed.EscapedPath()
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	port, _ := strconv.Atoi(parsed.Port())
	return parsed.Scheme, parsed.Hostname(), port, path
}

func serviceHasExplicitHealthTarget(service domain.Service) bool {
	if service.HealthAddressSource != "" && service.HealthAddressSource != service.AddressSource {
		return true
	}
	if strings.TrimSpace(service.HealthHostValue) != "" &&
		strings.TrimSpace(service.HealthHostValue) != strings.TrimSpace(service.HostValue) &&
		strings.TrimSpace(service.HealthHostValue) != strings.TrimSpace(service.Host) {
		return true
	}
	if strings.TrimSpace(service.HealthScheme) != "" &&
		strings.TrimSpace(service.HealthScheme) != strings.TrimSpace(service.Scheme) {
		return true
	}
	if service.HealthPort > 0 && service.HealthPort != service.Port {
		return true
	}
	if normalizePath(service.HealthPath) != normalizePath(service.Path) {
		return true
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
