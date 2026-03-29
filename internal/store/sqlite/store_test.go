package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func TestSQLiteConnectionsUseWALAndBusyTimeout(t *testing.T) {
	store := newTestStore(t)

	for name, handle := range map[string]*sql.DB{
		"reader": store.readDB,
		"writer": store.db,
	} {
		if handle == nil {
			t.Fatalf("%s handle is nil", name)
		}

		var journalMode string
		if err := handle.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
			t.Fatalf("%s journal mode: %v", name, err)
		}
		if !strings.EqualFold(journalMode, "wal") {
			t.Fatalf("%s journal mode = %q, want WAL", name, journalMode)
		}

		var synchronous int
		if err := handle.QueryRow(`PRAGMA synchronous`).Scan(&synchronous); err != nil {
			t.Fatalf("%s synchronous pragma: %v", name, err)
		}
		if synchronous != 1 {
			t.Fatalf("%s synchronous = %d, want 1 (NORMAL)", name, synchronous)
		}

		var foreignKeys int
		if err := handle.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
			t.Fatalf("%s foreign_keys pragma: %v", name, err)
		}
		if foreignKeys != 1 {
			t.Fatalf("%s foreign_keys = %d, want 1", name, foreignKeys)
		}

		var busyTimeout int
		if err := handle.QueryRow(`PRAGMA busy_timeout`).Scan(&busyTimeout); err != nil {
			t.Fatalf("%s busy_timeout pragma: %v", name, err)
		}
		if busyTimeout < 5000 {
			t.Fatalf("%s busy_timeout = %d, want at least 5000", name, busyTimeout)
		}
	}
}

func TestConcurrentDashboardAndSettingsReadsDuringWriteBursts(t *testing.T) {
	store := newBootstrappedStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reportErr := make(chan error, 1)
	fail := func(err error) {
		if err == nil {
			return
		}
		select {
		case reportErr <- err:
			cancel()
		default:
		}
	}

	var wg sync.WaitGroup
	for writerID := 0; writerID < 3; writerID++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for iteration := 0; iteration < 25; iteration++ {
				if ctx.Err() != nil {
					return
				}
				seenAt := time.Now().UTC().Add(time.Duration(iteration) * time.Second)
				device, err := store.UpsertDeviceObservation(ctx, domain.DeviceObservation{
					IdentityKey: fmt.Sprintf("writer-%d-device", writerID),
					PrimaryMAC:  fmt.Sprintf("aa:bb:cc:dd:ee:%02x", writerID),
					Hostname:    fmt.Sprintf("node-%d", writerID),
					IPAddress:   fmt.Sprintf("192.168.%d.%d", writerID+1, iteration+10),
					Confidence:  domain.IdentityConfidenceHigh,
					LastSeenAt:  seenAt,
				})
				if err != nil {
					fail(fmt.Errorf("writer %d upsert device: %w", writerID, err))
					return
				}
				_, err = store.UpsertDiscoveredServiceObservation(ctx, domain.ServiceObservation{
					Name:            fmt.Sprintf("Service %d-%d", writerID, iteration),
					Source:          domain.ServiceSourceLAN,
					SourceRef:       fmt.Sprintf("writer-%d-service-%d/tcp", writerID, iteration),
					DeviceKey:       device.IdentityKey,
					AddressSource:   domain.ServiceAddressDevicePrimary,
					ServiceTypeHint: "http",
					HostValue:       fmt.Sprintf("192.168.%d.%d", writerID+1, iteration+10),
					Scheme:          "http",
					Host:            fmt.Sprintf("192.168.%d.%d", writerID+1, iteration+10),
					Port:            8000 + writerID,
					URL:             fmt.Sprintf("http://192.168.%d.%d:%d", writerID+1, iteration+10, 8000+writerID),
					LastSeenAt:      seenAt,
				}, device.ID)
				if err != nil {
					fail(fmt.Errorf("writer %d upsert discovered service: %w", writerID, err))
					return
				}
				if err := store.RecordJobRun(ctx, fmt.Sprintf("writer-%d", writerID), nil); err != nil {
					fail(fmt.Errorf("writer %d record job run: %w", writerID, err))
					return
				}
			}
		}(writerID)
	}

	for readerID := 0; readerID < 4; readerID++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for iteration := 0; iteration < 40; iteration++ {
				if ctx.Err() != nil {
					return
				}
				readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
				if _, err := store.GetDashboard(readCtx); err != nil {
					readCancel()
					fail(fmt.Errorf("reader %d dashboard read: %w", readerID, err))
					return
				}
				readCancel()

				readCtx, readCancel = context.WithTimeout(ctx, 2*time.Second)
				if _, err := store.GetSettingsView(readCtx); err != nil {
					readCancel()
					fail(fmt.Errorf("reader %d settings read: %w", readerID, err))
					return
				}
				readCancel()
			}
		}(readerID)
	}

	wg.Wait()

	select {
	case err := <-reportErr:
		t.Fatal(err)
	default:
	}
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("concurrent read/write test timed out")
	}
}

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
