package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/discovery/docker"
	"github.com/deleema/homelabwatch/internal/discovery/lan"
	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/monitoring"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
	"github.com/deleema/homelabwatch/internal/worker"
)

const eventsRetention = 14 * 24 * time.Hour

type App struct {
	config    config.Config
	store     *sqlite.Store
	bus       *events.Bus
	docker    *docker.Provider
	lan       *lan.Provider
	monitor   *monitoring.Runner
	scheduler *worker.Scheduler
}

type BootstrapResult struct {
	Performed      bool
	Generated      bool
	AdminToken     string
	AdminTokenFile string
}

func New(cfg config.Config, store *sqlite.Store, bus *events.Bus) *App {
	instance := &App{
		config:  cfg,
		store:   store,
		bus:     bus,
		docker:  docker.NewProvider(),
		lan:     lan.NewProvider(),
		monitor: monitoring.NewRunner(store),
	}
	instance.scheduler = worker.NewScheduler(
		worker.Job{Name: "docker-sync", Interval: 30 * time.Second, Run: instance.runDockerDiscovery},
		worker.Job{Name: "lan-scan", Interval: 5 * time.Minute, Run: instance.runLANDiscovery},
		worker.Job{Name: "health-checks", Interval: 30 * time.Second, Run: instance.runMonitoring},
		worker.Job{Name: "cleanup", Interval: 24 * time.Hour, Run: instance.runCleanup},
	)
	return instance
}

func (a *App) Start(ctx context.Context) {
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

func (a *App) EnsureBootstrap(ctx context.Context) (BootstrapResult, error) {
	status, err := a.store.BootstrapStatus(ctx)
	if err != nil {
		return BootstrapResult{}, err
	}
	if status.Initialized || !a.config.AutoBootstrap {
		return BootstrapResult{}, nil
	}

	token := strings.TrimSpace(a.config.AdminToken)
	generated := false
	if token == "" {
		token, err = generateAdminToken()
		if err != nil {
			return BootstrapResult{}, err
		}
		generated = true
	}

	if err := a.Initialize(ctx, domain.BootstrapInput{
		AdminToken:       token,
		AutoScanEnabled:  true,
		DefaultScanPorts: append([]int(nil), a.config.DefaultScanPorts...),
	}); err != nil {
		return BootstrapResult{}, err
	}

	result := BootstrapResult{
		Performed:  true,
		Generated:  generated,
		AdminToken: token,
	}
	tokenFile, fileErr := a.writeAdminTokenFile(token)
	result.AdminTokenFile = tokenFile
	if fileErr != nil {
		return result, fileErr
	}
	return result, nil
}

func (a *App) Initialize(ctx context.Context, input domain.BootstrapInput) error {
	status, err := a.store.BootstrapStatus(ctx)
	if err != nil {
		return err
	}
	if status.Initialized {
		return errors.New("bootstrap already completed")
	}
	if strings.TrimSpace(input.AdminToken) == "" {
		return errors.New("admin token is required")
	}
	if len(input.DefaultScanPorts) == 0 {
		input.DefaultScanPorts = append([]int(nil), a.config.DefaultScanPorts...)
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
		"autoScanEnabled":  input.AutoScanEnabled,
		"defaultScanPorts": input.DefaultScanPorts,
	})
	return nil
}

func (a *App) ValidateAdminToken(ctx context.Context, token string) (bool, error) {
	return a.store.ValidateAdminToken(ctx, token)
}

func (a *App) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	return a.store.GetDashboard(ctx)
}

func (a *App) Settings(ctx context.Context) (domain.SettingsView, error) {
	item, err := a.store.GetSettingsView(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	item.AppSettings.AdminTokenFile = a.config.AdminTokenFile
	return item, nil
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

func (a *App) ListBookmarks(ctx context.Context) ([]domain.Bookmark, error) {
	return a.store.ListBookmarks(ctx)
}

func (a *App) SaveBookmark(ctx context.Context, bookmark domain.Bookmark) (domain.Bookmark, error) {
	item, err := a.store.SaveBookmark(ctx, bookmark)
	if err != nil {
		return domain.Bookmark{}, err
	}
	a.publish("bookmark", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteBookmark(ctx context.Context, id string) error {
	if err := a.store.DeleteBookmark(ctx, id); err != nil {
		return err
	}
	a.publish("bookmark", id, "deleted", nil)
	return nil
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
		return nil
	}
	endpoints, err := a.store.ListDockerEndpoints(ctx)
	if err != nil {
		_ = a.store.RecordJobRun(ctx, "docker-sync", err)
		return err
	}
	results := a.docker.Discover(ctx, endpoints)
	var runErr error
	for _, result := range results {
		if result.Err != nil && runErr == nil {
			runErr = result.Err
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
	_ = a.store.RecordJobRun(ctx, "docker-sync", runErr)
	return runErr
}

func (a *App) runLANDiscovery(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		return nil
	}
	settings, err := a.store.GetAppSettings(ctx)
	if err != nil {
		_ = a.store.RecordJobRun(ctx, "lan-scan", err)
		return err
	}
	if !settings.AutoScanEnabled {
		_ = a.store.RecordJobRun(ctx, "lan-scan", nil)
		return nil
	}
	targets, err := a.store.ListScanTargets(ctx)
	if err != nil {
		_ = a.store.RecordJobRun(ctx, "lan-scan", err)
		return err
	}
	observations, err := a.lan.Discover(ctx, targets)
	if err != nil && !errors.Is(err, context.Canceled) {
		_ = a.store.RecordJobRun(ctx, "lan-scan", err)
		return err
	}
	processErr := a.processObservations(ctx, observations)
	if processErr != nil {
		_ = a.store.RecordJobRun(ctx, "lan-scan", processErr)
		return processErr
	}
	_ = a.store.RecordJobRun(ctx, "lan-scan", nil)
	return nil
}

func (a *App) runMonitoring(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		return nil
	}
	results, err := a.monitor.RunDueChecks(ctx)
	if err != nil {
		_ = a.store.RecordJobRun(ctx, "health-checks", err)
		return err
	}
	for _, item := range results {
		a.publish("check", item.CheckID, "recorded", item)
	}
	_ = a.store.RecordJobRun(ctx, "health-checks", nil)
	return nil
}

func (a *App) runCleanup(ctx context.Context) error {
	if !a.isBootstrapped(ctx) {
		return nil
	}
	err := a.store.Cleanup(ctx, eventsRetention)
	_ = a.store.RecordJobRun(ctx, "cleanup", err)
	return err
}

func (a *App) processObservations(ctx context.Context, observations []domain.Observation) error {
	for _, observation := range observations {
		deviceID := ""
		if hasDevice(observation.Device) {
			device, err := a.store.UpsertDeviceObservation(ctx, observation.Device)
			if err != nil {
				return err
			}
			deviceID = device.ID
			a.publish("device", device.ID, "seen", device)
		}
		for _, serviceObservation := range observation.Services {
			service, err := a.store.UpsertDiscoveredService(ctx, serviceObservation, deviceID)
			if err != nil {
				return err
			}
			a.publish("service", service.ID, "seen", service)
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

func generateAdminToken() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (a *App) writeAdminTokenFile(token string) (string, error) {
	path := strings.TrimSpace(a.config.AdminTokenFile)
	if path == "" {
		return "", nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func normalizeService(service domain.Service) domain.Service {
	if service.Details == nil {
		service.Details = map[string]any{}
	}
	if service.Source == "" {
		service.Source = domain.ServiceSourceManual
	}
	if service.URL != "" && (service.Host == "" || service.Scheme == "") {
		if parsed, err := url.Parse(service.URL); err == nil {
			service.Scheme = firstNonEmpty(service.Scheme, parsed.Scheme)
			service.Host = firstNonEmpty(service.Host, parsed.Hostname())
			service.Path = firstNonEmpty(service.Path, parsed.EscapedPath())
			if port, err := strconv.Atoi(parsed.Port()); err == nil {
				service.Port = port
			}
		}
	}
	if service.URL == "" && service.Host != "" {
		service.URL = buildURL(service.Scheme, service.Host, service.Port, service.Path)
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
