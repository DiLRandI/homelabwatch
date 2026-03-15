package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func TestInitializeSeedsBootstrapState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.Initialize(ctx, domain.BootstrapInput{
		AdminToken:       "secret",
		AutoScanEnabled:  true,
		DefaultScanPorts: []int{22, 80},
		ScanTargets: []domain.ScanTargetSeed{
			{Name: "lan", CIDR: "192.168.1.0/24", Enabled: true, ScanIntervalSeconds: 300, CommonPorts: []int{22, 80}},
		},
	})
	if err != nil {
		t.Fatalf("initialize store: %v", err)
	}

	status, err := store.BootstrapStatus(ctx)
	if err != nil {
		t.Fatalf("bootstrap status: %v", err)
	}
	if !status.Initialized {
		t.Fatalf("expected initialized status")
	}

	targets, err := store.ListScanTargets(ctx)
	if err != nil {
		t.Fatalf("list scan targets: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 scan target, got %d", len(targets))
	}
}

func TestDeviceObservationMACReuseKeepsSingleDevice(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	first := domain.DeviceObservation{
		IdentityKey: "mac:aa:bb:cc:dd:ee:ff",
		PrimaryMAC:  "aa:bb:cc:dd:ee:ff",
		Hostname:    "nas",
		IPAddress:   "192.168.1.10",
		Confidence:  domain.IdentityConfidenceHigh,
		LastSeenAt:  time.Now().UTC(),
	}
	second := domain.DeviceObservation{
		IdentityKey: "mac:aa:bb:cc:dd:ee:ff",
		PrimaryMAC:  "aa:bb:cc:dd:ee:ff",
		Hostname:    "nas",
		IPAddress:   "192.168.1.20",
		Confidence:  domain.IdentityConfidenceHigh,
		LastSeenAt:  time.Now().UTC().Add(time.Minute),
	}

	device, err := store.UpsertDeviceObservation(ctx, first)
	if err != nil {
		t.Fatalf("upsert first observation: %v", err)
	}
	device, err = store.UpsertDeviceObservation(ctx, second)
	if err != nil {
		t.Fatalf("upsert second observation: %v", err)
	}

	devices, err := store.ListDevices(ctx)
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if device.PrimaryMAC != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("unexpected mac %q", device.PrimaryMAC)
	}
	if len(device.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(device.Addresses))
	}
}

func TestCheckResultsUpdateServiceStatus(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	service, err := store.SaveManualService(ctx, domain.Service{
		Name:   "Test Service",
		Source: domain.ServiceSourceManual,
		Scheme: "http",
		Host:   "127.0.0.1",
		Port:   8080,
		URL:    "http://127.0.0.1:8080",
	})
	if err != nil {
		t.Fatalf("save service: %v", err)
	}

	checkA, err := store.SaveServiceCheck(ctx, domain.ServiceCheck{
		ServiceID:         service.ID,
		Name:              "http",
		Type:              domain.CheckTypeHTTP,
		Target:            service.URL,
		IntervalSeconds:   60,
		TimeoutSeconds:    5,
		ExpectedStatusMin: 200,
		ExpectedStatusMax: 399,
		Enabled:           true,
	})
	if err != nil {
		t.Fatalf("save check a: %v", err)
	}
	checkB, err := store.SaveServiceCheck(ctx, domain.ServiceCheck{
		ServiceID:       service.ID,
		Name:            "tcp",
		Type:            domain.CheckTypeTCP,
		Target:          "127.0.0.1:8080",
		IntervalSeconds: 60,
		TimeoutSeconds:  5,
		Enabled:         true,
	})
	if err != nil {
		t.Fatalf("save check b: %v", err)
	}

	now := time.Now().UTC()
	if err := store.SaveCheckResult(ctx, domain.CheckResult{CheckID: checkA.ID, ServiceID: service.ID, Status: domain.HealthStatusHealthy, CheckedAt: now}); err != nil {
		t.Fatalf("record first result: %v", err)
	}
	if err := store.SaveCheckResult(ctx, domain.CheckResult{CheckID: checkB.ID, ServiceID: service.ID, Status: domain.HealthStatusUnhealthy, CheckedAt: now}); err != nil {
		t.Fatalf("record second result: %v", err)
	}

	updated, err := store.GetService(ctx, service.ID)
	if err != nil {
		t.Fatalf("get service: %v", err)
	}
	if updated.Status != domain.HealthStatusDegraded {
		t.Fatalf("expected degraded status, got %s", updated.Status)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := New(filepath.Join(t.TempDir(), "homelabwatch.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func newBootstrappedStore(t *testing.T) *Store {
	t.Helper()
	store := newTestStore(t)
	err := store.Initialize(context.Background(), domain.BootstrapInput{
		AdminToken:       "secret",
		AutoScanEnabled:  true,
		DefaultScanPorts: []int{22, 80},
	})
	if err != nil {
		t.Fatalf("bootstrap store: %v", err)
	}
	return store
}
