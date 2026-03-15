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

	err := store.Initialize(ctx, domain.SetupInput{
		ApplianceName:    "Lab",
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

func TestCreateAndValidateAPIToken(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	created, err := store.CreateAPIToken(ctx, domain.CreateAPITokenInput{
		Name:  "Automation",
		Scope: domain.TokenScopeWrite,
	})
	if err != nil {
		t.Fatalf("create api token: %v", err)
	}
	if created.Secret == "" {
		t.Fatalf("expected raw token secret")
	}

	ok, err := store.ValidateAPIToken(ctx, created.Secret, domain.TokenScopeWrite)
	if err != nil {
		t.Fatalf("validate api token: %v", err)
	}
	if !ok {
		t.Fatalf("expected api token to validate")
	}

	if err := store.RevokeAPIToken(ctx, created.Token.ID); err != nil {
		t.Fatalf("revoke api token: %v", err)
	}
	ok, err = store.ValidateAPIToken(ctx, created.Secret, domain.TokenScopeRead)
	if err != nil {
		t.Fatalf("validate revoked api token: %v", err)
	}
	if ok {
		t.Fatalf("expected revoked token to fail validation")
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

func TestBookmarkLinkedServiceTracksServiceURL(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	service, err := store.SaveManualService(ctx, domain.Service{
		Name:   "Grafana",
		Source: domain.ServiceSourceManual,
		Scheme: "http",
		Host:   "192.168.1.10",
		Port:   3000,
		URL:    "http://192.168.1.10:3000",
	})
	if err != nil {
		t.Fatalf("save service: %v", err)
	}

	bookmark, err := store.SaveBookmark(ctx, domain.BookmarkInput{
		Name:      "Operations Grafana",
		ServiceID: service.ID,
	})
	if err != nil {
		t.Fatalf("save bookmark: %v", err)
	}
	if bookmark.URL != "http://192.168.1.10:3000" {
		t.Fatalf("unexpected initial bookmark url %q", bookmark.URL)
	}

	service.Host = "192.168.1.45"
	service.URL = "http://192.168.1.45:3000"
	if _, err := store.SaveManualService(ctx, service); err != nil {
		t.Fatalf("update service: %v", err)
	}

	updated, err := store.GetBookmark(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("get bookmark: %v", err)
	}
	if updated.URL != "http://192.168.1.45:3000" {
		t.Fatalf("expected bookmark url to follow linked service, got %q", updated.URL)
	}
}

func TestDeleteFolderPromotesBookmarksToParent(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	parent, err := store.SaveFolder(ctx, domain.FolderInput{Name: "Infrastructure"})
	if err != nil {
		t.Fatalf("save parent folder: %v", err)
	}
	child, err := store.SaveFolder(ctx, domain.FolderInput{
		Name:     "Monitoring",
		ParentID: parent.ID,
	})
	if err != nil {
		t.Fatalf("save child folder: %v", err)
	}
	bookmark, err := store.SaveBookmark(ctx, domain.BookmarkInput{
		FolderID: child.ID,
		Name:     "Grafana",
		Tags:     []string{"monitoring"},
		URL:      "http://192.168.1.20:3000",
	})
	if err != nil {
		t.Fatalf("save bookmark: %v", err)
	}

	if err := store.DeleteFolder(ctx, child.ID); err != nil {
		t.Fatalf("delete child folder: %v", err)
	}

	updated, err := store.GetBookmark(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("get bookmark: %v", err)
	}
	if updated.FolderID != parent.ID {
		t.Fatalf("expected bookmark folder to be promoted to parent, got %q", updated.FolderID)
	}
}

func TestOpenBookmarkUpdatesUsage(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	bookmark, err := store.SaveBookmark(ctx, domain.BookmarkInput{
		Name: "Home Assistant",
		URL:  "http://192.168.1.20:8123",
	})
	if err != nil {
		t.Fatalf("save bookmark: %v", err)
	}

	opened, err := store.OpenBookmark(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("open bookmark: %v", err)
	}
	if opened.ClickCount != 1 {
		t.Fatalf("expected click count 1, got %d", opened.ClickCount)
	}
	if opened.LastOpenedAt.IsZero() {
		t.Fatalf("expected last opened at to be set")
	}
}

func TestBookmarkLinkedToDiscoveredServiceTracksURLChanges(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	device, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{
		IdentityKey: "mac:aa:bb:cc:dd:ee:11",
		PrimaryMAC:  "aa:bb:cc:dd:ee:11",
		Hostname:    "raspberrypi",
		IPAddress:   "192.168.1.20",
		Confidence:  domain.IdentityConfidenceHigh,
		LastSeenAt:  time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("save device: %v", err)
	}

	service, err := store.UpsertDiscoveredService(ctx, domain.ServiceObservation{
		Name:       "Home Assistant",
		Source:     domain.ServiceSourceLAN,
		SourceRef:  "raspberrypi:8123/tcp",
		DeviceKey:  device.IdentityKey,
		Scheme:     "http",
		Host:       "192.168.1.20",
		Port:       8123,
		URL:        "http://192.168.1.20:8123",
		LastSeenAt: time.Now().UTC(),
	}, device.ID)
	if err != nil {
		t.Fatalf("save discovered service: %v", err)
	}

	bookmark, err := store.SaveBookmark(ctx, domain.BookmarkInput{
		ServiceID: service.ID,
		Tags:      []string{"automation"},
	})
	if err != nil {
		t.Fatalf("save bookmark: %v", err)
	}
	if bookmark.URL != "http://192.168.1.20:8123" {
		t.Fatalf("expected initial service url, got %q", bookmark.URL)
	}

	_, err = store.UpsertDiscoveredService(ctx, domain.ServiceObservation{
		Name:       "Home Assistant",
		Source:     domain.ServiceSourceLAN,
		SourceRef:  "raspberrypi:8123/tcp",
		DeviceKey:  device.IdentityKey,
		Scheme:     "http",
		Host:       "192.168.1.45",
		Port:       8123,
		URL:        "http://192.168.1.45:8123",
		LastSeenAt: time.Now().UTC().Add(time.Minute),
	}, device.ID)
	if err != nil {
		t.Fatalf("update discovered service: %v", err)
	}

	updated, err := store.GetBookmark(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("get updated bookmark: %v", err)
	}
	if updated.URL != "http://192.168.1.45:8123" {
		t.Fatalf("expected updated service url, got %q", updated.URL)
	}
}

func TestBookmarkOpenTracksUsage(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	bookmark, err := store.SaveBookmark(ctx, domain.BookmarkInput{
		Name: "Router",
		URL:  "http://192.168.1.1",
	})
	if err != nil {
		t.Fatalf("save bookmark: %v", err)
	}

	opened, err := store.OpenBookmark(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("open bookmark: %v", err)
	}
	if opened.ClickCount != 1 {
		t.Fatalf("expected click count 1, got %d", opened.ClickCount)
	}
	if opened.LastOpenedAt.IsZero() {
		t.Fatalf("expected last opened time to be set")
	}
}

func TestBookmarksSupportFoldersAndTags(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx := context.Background()

	folder, err := store.SaveFolder(ctx, domain.FolderInput{
		Name:     "Monitoring",
		Position: 1,
	})
	if err != nil {
		t.Fatalf("save folder: %v", err)
	}

	bookmark, err := store.SaveBookmark(ctx, domain.BookmarkInput{
		FolderID: folder.ID,
		Name:     "Grafana",
		URL:      "http://192.168.1.10:3000",
		Tags:     []string{"monitoring", "dashboards"},
	})
	if err != nil {
		t.Fatalf("save bookmark: %v", err)
	}
	if bookmark.FolderID != folder.ID {
		t.Fatalf("expected bookmark folder id %q, got %q", folder.ID, bookmark.FolderID)
	}
	if len(bookmark.Tags) != 2 {
		t.Fatalf("expected two tags, got %d", len(bookmark.Tags))
	}

	folders, err := store.ListFolders(ctx)
	if err != nil {
		t.Fatalf("list folders: %v", err)
	}
	if len(folders) != 1 || folders[0].BookmarkCount != 1 {
		t.Fatalf("expected one folder with one bookmark, got %#v", folders)
	}

	tags, err := store.ListTags(ctx)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
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
	err := store.Initialize(context.Background(), domain.SetupInput{
		ApplianceName:    "Lab",
		AutoScanEnabled:  true,
		DefaultScanPorts: []int{22, 80},
	})
	if err != nil {
		t.Fatalf("bootstrap store: %v", err)
	}
	return store
}
