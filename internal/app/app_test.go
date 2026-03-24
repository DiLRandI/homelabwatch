package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/config"
	"github.com/deleema/homelabwatch/internal/domain"
	"github.com/deleema/homelabwatch/internal/events"
	"github.com/deleema/homelabwatch/internal/store/sqlite"
)

func TestSetupInitializesWorkspace(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})

	err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		AutoScanEnabled:  true,
		DefaultScanPorts: []int{22, 80},
		RunDiscovery:     false,
	})
	if err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	status, err := store.BootstrapStatus(context.Background())
	if err != nil {
		t.Fatalf("bootstrap status: %v", err)
	}
	if !status.Initialized {
		t.Fatalf("expected initialized store")
	}

	settings, err := application.Settings(context.Background())
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.AppSettings.ApplianceName != "Rack Alpha" {
		t.Fatalf("expected appliance name to persist, got %q", settings.AppSettings.ApplianceName)
	}
}

func TestSetupRejectsSecondRun(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})

	if err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		DefaultScanPorts: []int{22, 80},
	}); err != nil {
		t.Fatalf("first setup: %v", err)
	}

	err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Beta",
		DefaultScanPorts: []int{22, 80},
	})
	if err == nil {
		t.Fatalf("expected second setup attempt to fail")
	}
}

func TestCreateAPITokenSupportsScopedValidation(t *testing.T) {
	application, _, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})

	if err := application.Setup(context.Background(), domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		DefaultScanPorts: []int{22, 80},
	}); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	created, err := application.CreateAPIToken(context.Background(), domain.CreateAPITokenInput{
		Name:  "Read only",
		Scope: domain.TokenScopeRead,
	})
	if err != nil {
		t.Fatalf("create api token: %v", err)
	}
	if created.Secret == "" {
		t.Fatalf("expected raw api token secret")
	}

	ok, err := application.ValidateAPIToken(context.Background(), created.Secret, domain.TokenScopeRead)
	if err != nil {
		t.Fatalf("validate read token: %v", err)
	}
	if !ok {
		t.Fatalf("expected read token to validate for read scope")
	}

	ok, err = application.ValidateAPIToken(context.Background(), created.Secret, domain.TokenScopeWrite)
	if err != nil {
		t.Fatalf("validate write scope with read token: %v", err)
	}
	if ok {
		t.Fatalf("expected read token to fail write scope validation")
	}
}

func TestCreateBookmarkFromDiscoveredServicePromotesSuggestionAndTracksDeviceIP(t *testing.T) {
	application, store, _ := newTestApp(t, config.Config{
		DefaultScanPorts: []int{22, 80},
	})
	ctx := context.Background()

	if err := application.Setup(ctx, domain.SetupInput{
		ApplianceName:    "Rack Alpha",
		DefaultScanPorts: []int{22, 80},
	}); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	device, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{
		IdentityKey: "mac:aa:bb:cc:dd:ee:22",
		PrimaryMAC:  "aa:bb:cc:dd:ee:22",
		Hostname:    "raspberrypi.local",
		IPAddress:   "192.168.1.20",
		Confidence:  domain.IdentityConfidenceHigh,
		LastSeenAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("save device: %v", err)
	}

	if _, err := store.UpsertDiscoveredServiceObservation(ctx, domain.ServiceObservation{
		Name:            "Home Assistant",
		Source:          domain.ServiceSourceLAN,
		SourceRef:       "mac:aa:bb:cc:dd:ee:22:8123/tcp",
		ServiceTypeHint: "home-assistant",
		AddressSource:   domain.ServiceAddressDevicePrimary,
		HostValue:       "192.168.1.20",
		Host:            "192.168.1.20",
		Scheme:          "http",
		Port:            8123,
		URL:             "http://192.168.1.20:8123",
		LastSeenAt:      time.Now().UTC(),
	}, device.ID); err != nil {
		t.Fatalf("save lan discovery: %v", err)
	}

	if _, err := store.UpsertDiscoveredServiceObservation(ctx, domain.ServiceObservation{
		Name:            "Home Assistant",
		Source:          domain.ServiceSourceMDNS,
		SourceRef:       "raspberrypi.local:8123/mdns",
		ServiceTypeHint: "home-assistant",
		AddressSource:   domain.ServiceAddressMDNSHostname,
		HostValue:       "raspberrypi.local",
		Host:            "raspberrypi.local",
		Scheme:          "http",
		Port:            8123,
		URL:             "http://raspberrypi.local:8123",
		LastSeenAt:      time.Now().UTC(),
	}, device.ID); err != nil {
		t.Fatalf("save mdns discovery: %v", err)
	}

	discovered, err := application.ListDiscoveredServices(ctx)
	if err != nil {
		t.Fatalf("list discovered services: %v", err)
	}
	if len(discovered) != 1 {
		t.Fatalf("expected one merged discovery suggestion, got %d", len(discovered))
	}
	if len(discovered[0].SourceTypes) != 2 {
		t.Fatalf("expected merged source evidence, got %#v", discovered[0].SourceTypes)
	}

	bookmark, err := application.CreateBookmarkFromDiscoveredService(ctx, domain.CreateBookmarkFromDiscoveredServiceInput{
		DiscoveredServiceID: discovered[0].ID,
	})
	if err != nil {
		t.Fatalf("create bookmark from discovery: %v", err)
	}
	if bookmark.URL != "http://192.168.1.20:8123" {
		t.Fatalf("unexpected bookmark url %q", bookmark.URL)
	}

	if _, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{
		IdentityKey: "mac:aa:bb:cc:dd:ee:22",
		PrimaryMAC:  "aa:bb:cc:dd:ee:22",
		Hostname:    "raspberrypi.local",
		IPAddress:   "192.168.1.45",
		Confidence:  domain.IdentityConfidenceHigh,
		LastSeenAt:  time.Now().UTC().Add(time.Minute),
	}); err != nil {
		t.Fatalf("update device ip: %v", err)
	}

	updated, err := application.GetBookmark(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("reload bookmark: %v", err)
	}
	if updated.URL != "http://192.168.1.45:8123" {
		t.Fatalf("expected bookmark url to follow device ip, got %q", updated.URL)
	}
}

func newTestApp(t *testing.T, overrides config.Config) (*App, *sqlite.Store, config.Config) {
	t.Helper()

	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr:       ":0",
		DataDir:          dataDir,
		DBPath:           filepath.Join(dataDir, "homelabwatch.db"),
		StaticDir:        dataDir,
		SeedDockerSocket: false,
		DefaultScanPorts: []int{22, 80, 443},
		TrustedCIDRs:     []string{"127.0.0.1/32"},
	}
	if len(overrides.DefaultScanPorts) > 0 {
		cfg.DefaultScanPorts = append([]int(nil), overrides.DefaultScanPorts...)
	}
	if len(overrides.TrustedCIDRs) > 0 {
		cfg.TrustedCIDRs = append([]string(nil), overrides.TrustedCIDRs...)
	}

	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	return New(cfg, store, events.NewBus()), store, cfg
}
