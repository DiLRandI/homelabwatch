package sqlite

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct {
	db     *sql.DB
	readDB *sql.DB
}

func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	writeDB, err := openSQLiteDB(path, 1, 1)
	if err != nil {
		return nil, err
	}
	store := &Store{db: writeDB}
	if err := store.migrate(context.Background()); err != nil {
		_ = writeDB.Close()
		return nil, err
	}
	if err := store.seedBuiltInServiceDefinitions(context.Background()); err != nil {
		_ = writeDB.Close()
		return nil, err
	}
	readDB, err := openSQLiteDB(path, 4, 4)
	if err != nil {
		_ = writeDB.Close()
		return nil, err
	}
	store.readDB = readDB
	return store, nil
}

func (s *Store) Close() error {
	var readErr error
	if s.readDB != nil {
		readErr = s.readDB.Close()
	}
	return errors.Join(s.db.Close(), readErr)
}

func (s *Store) reader() *sql.DB {
	if s.readDB != nil {
		return s.readDB
	}
	return s.db
}

func openSQLiteDB(path string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	return db, nil
}

func sqliteDSN(path string) string {
	query := url.Values{}
	for _, pragma := range []string{
		"foreign_keys=on",
		"journal_mode=WAL",
		"synchronous=NORMAL",
		"busy_timeout=5000",
	} {
		query.Add("_pragma", pragma)
	}
	if strings.Contains(path, "?") {
		return path + "&" + query.Encode()
	}
	return path + "?" + query.Encode()
}

func (s *Store) migrate(ctx context.Context) error {
	migrationDir, err := findMigrationsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		var exists int
		if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM schema_migrations WHERE version = ?", entry.Name()).Scan(&exists); err != nil {
			if !strings.Contains(err.Error(), "no such table") {
				return err
			}
			exists = 0
		}
		if exists > 0 {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationDir, entry.Name()))
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		for statement := range strings.SplitSeq(string(content), ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)", entry.Name(), nowString()); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func findMigrationsDir() (string, error) {
	for _, start := range candidateRoots() {
		for current := start; current != "/" && current != "."; current = filepath.Dir(current) {
			candidate := filepath.Join(current, "migrations")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate, nil
			}
			next := filepath.Dir(current)
			if next == current {
				break
			}
		}
	}
	return "", errors.New("migrations directory not found")
}

func candidateRoots() []string {
	roots := make([]string, 0, 2)
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	if executable, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(executable))
	}
	return roots
}

func (s *Store) BootstrapStatus(ctx context.Context) (domain.BootstrapStatus, error) {
	var raw string
	if err := s.reader().QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = 'initialized'").Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BootstrapStatus{Initialized: false}, nil
		}
		return domain.BootstrapStatus{}, err
	}
	return domain.BootstrapStatus{Initialized: raw == "true"}, nil
}

func (s *Store) GetAppSettings(ctx context.Context) (domain.AppSettings, error) {
	rows, err := s.reader().QueryContext(ctx, "SELECT key, value, updated_at FROM app_settings")
	if err != nil {
		return domain.AppSettings{}, err
	}
	defer rows.Close()
	settings := domain.AppSettings{
		DefaultScanPorts:          []int{},
		BookmarkPolicy:            domain.BookmarkAutomationManual,
		AutoBookmarkSources:       []domain.ServiceSource{domain.ServiceSourceDocker, domain.ServiceSourceLAN, domain.ServiceSourceMDNS},
		AutoBookmarkMinConfidence: 90,
	}
	for rows.Next() {
		var key, value, updatedAt string
		if err := rows.Scan(&key, &value, &updatedAt); err != nil {
			return domain.AppSettings{}, err
		}
		switch key {
		case "initialized":
			settings.Initialized = value == "true"
		case "admin_token_hash":
			settings.AdminTokenHash = value
		case "appliance_name":
			settings.ApplianceName = value
		case "initialized_at":
			settings.InitializedAt = parseTime(value)
		case "last_bootstrap_at":
			settings.LastBootstrapAt = parseTime(value)
		case "auto_scan_enabled":
			settings.AutoScanEnabled = value == "true"
		case "default_scan_ports":
			_ = json.Unmarshal([]byte(value), &settings.DefaultScanPorts)
		case "bookmark_policy":
			settings.BookmarkPolicy = domain.BookmarkAutomationPolicy(value)
		case "auto_bookmark_sources":
			_ = json.Unmarshal([]byte(value), &settings.AutoBookmarkSources)
		case "auto_bookmark_min_confidence":
			if parsed, err := strconv.Atoi(value); err == nil {
				settings.AutoBookmarkMinConfidence = parsed
			}
		}
		settings.UpdatedAt = parseTime(updatedAt)
	}
	settings.LegacyTokenEnabled = settings.AdminTokenHash != ""
	return settings, rows.Err()
}

func (s *Store) Initialize(ctx context.Context, input domain.SetupInput) error {
	status, err := s.BootstrapStatus(ctx)
	if err != nil {
		return err
	}
	if status.Initialized {
		return errors.New("bootstrap already completed")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := nowString()
	settings := map[string]string{
		"initialized":        "true",
		"appliance_name":     firstNonEmpty(strings.TrimSpace(input.ApplianceName), "HomelabWatch"),
		"initialized_at":     now,
		"last_bootstrap_at":  now,
		"auto_scan_enabled":  boolString(input.AutoScanEnabled),
		"default_scan_ports": string(mustJSON(input.DefaultScanPorts)),
		"bookmark_policy":    string(domain.BookmarkAutomationManual),
		"auto_bookmark_sources": string(mustJSON([]domain.ServiceSource{
			domain.ServiceSourceDocker,
			domain.ServiceSourceLAN,
			domain.ServiceSourceMDNS,
		})),
		"auto_bookmark_min_confidence": "90",
	}
	for key, value := range settings {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO app_settings(key, value, updated_at)
			VALUES(?, ?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
		`, key, value, now); err != nil {
			return err
		}
	}
	for _, endpoint := range input.DockerEndpoints {
		if _, err := s.upsertDockerEndpointTx(ctx, tx, domain.DockerEndpoint{
			Name:                endpoint.Name,
			Kind:                endpoint.Kind,
			Address:             endpoint.Address,
			TLSCAPath:           endpoint.TLSCAPath,
			TLSCertPath:         endpoint.TLSCertPath,
			TLSKeyPath:          endpoint.TLSKeyPath,
			Enabled:             endpoint.Enabled,
			ScanIntervalSeconds: endpoint.ScanIntervalSeconds,
		}); err != nil {
			return err
		}
	}
	for _, target := range input.ScanTargets {
		if _, err := s.upsertScanTargetTx(ctx, tx, domain.ScanTarget{
			Name:                target.Name,
			CIDR:                target.CIDR,
			AutoDetected:        target.AutoDetected,
			Enabled:             target.Enabled,
			ScanIntervalSeconds: target.ScanIntervalSeconds,
			CommonPorts:         append([]int(nil), target.CommonPorts...),
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ValidateAPIToken(ctx context.Context, token string, requiredScope domain.TokenScope) (bool, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	settings, err := s.GetAppSettings(ctx)
	if err != nil {
		return false, err
	}
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	if settings.AdminTokenHash != "" && hash == settings.AdminTokenHash {
		return true, nil
	}
	var (
		scope     string
		revokedAt string
	)
	err = s.db.QueryRowContext(ctx, `SELECT scope, COALESCE(revoked_at, '') FROM api_tokens WHERE token_hash = ?`, hash).Scan(&scope, &revokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if revokedAt != "" || !tokenScopeAllows(domain.TokenScope(scope), requiredScope) {
		return false, nil
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE api_tokens SET last_used_at = ?, updated_at = ? WHERE token_hash = ?`, nowString(), nowString(), hash); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) GetSettingsView(ctx context.Context) (domain.SettingsView, error) {
	appSettings, err := s.GetAppSettings(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	dockerEndpoints, err := s.ListDockerEndpoints(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	scanTargets, err := s.ListScanTargets(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	jobState, err := s.ListJobState(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	apiTokens, err := s.ListAPITokens(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	serviceDefinitions, err := s.ListServiceDefinitions(ctx)
	if err != nil {
		return domain.SettingsView{}, err
	}
	return domain.SettingsView{
		AppSettings:        appSettings,
		DockerEndpoints:    dockerEndpoints,
		ScanTargets:        scanTargets,
		JobState:           jobState,
		ServiceDefinitions: serviceDefinitions,
		APIAccess: domain.APIAccessView{
			Tokens:                apiTokens,
			LegacyAdminTokenAlive: appSettings.AdminTokenHash != "",
		},
		Discovery: domain.DiscoverySettings{
			BookmarkPolicy:            appSettings.BookmarkPolicy,
			AutoBookmarkSources:       append([]domain.ServiceSource(nil), appSettings.AutoBookmarkSources...),
			AutoBookmarkMinConfidence: appSettings.AutoBookmarkMinConfidence,
		},
	}, nil
}

func (s *Store) ListAPITokens(ctx context.Context) ([]domain.APIToken, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, name, scope, token_prefix, COALESCE(last_used_at, ''), created_at, updated_at, COALESCE(revoked_at, '') FROM api_tokens ORDER BY revoked_at IS NOT NULL, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.APIToken
	for rows.Next() {
		var item domain.APIToken
		var scope, lastUsedAt, createdAt, updatedAt, revokedAt string
		if err := rows.Scan(&item.ID, &item.Name, &scope, &item.Prefix, &lastUsedAt, &createdAt, &updatedAt, &revokedAt); err != nil {
			return nil, err
		}
		item.Scope = domain.TokenScope(scope)
		item.LastUsedAt = parseTime(lastUsedAt)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		item.RevokedAt = parseTime(revokedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateAPIToken(ctx context.Context, input domain.CreateAPITokenInput) (domain.CreatedAPIToken, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.CreatedAPIToken{}, errors.New("token name is required")
	}
	scope := input.Scope
	if scope == "" {
		scope = domain.TokenScopeWrite
	}
	if scope != domain.TokenScopeRead && scope != domain.TokenScopeWrite {
		return domain.CreatedAPIToken{}, errors.New("token scope must be read or write")
	}
	secret, err := generateAPITokenSecret()
	if err != nil {
		return domain.CreatedAPIToken{}, err
	}
	now := time.Now().UTC()
	item := domain.APIToken{
		ID:        newID("tok"),
		Name:      name,
		Scope:     scope,
		Prefix:    tokenPrefix(secret),
		CreatedAt: now,
		UpdatedAt: now,
	}
	sum := sha256.Sum256([]byte(secret))
	if _, err := s.db.ExecContext(ctx, `INSERT INTO api_tokens(id, name, scope, token_hash, token_prefix, last_used_at, revoked_at, created_at, updated_at) VALUES(?, ?, ?, ?, ?, NULL, NULL, ?, ?)`,
		item.ID,
		item.Name,
		item.Scope,
		hex.EncodeToString(sum[:]),
		item.Prefix,
		item.CreatedAt.Format(time.RFC3339Nano),
		item.UpdatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return domain.CreatedAPIToken{}, err
	}
	return domain.CreatedAPIToken{Token: item, Secret: secret}, nil
}

func (s *Store) RevokeAPIToken(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE api_tokens SET revoked_at = ?, updated_at = ? WHERE id = ? AND revoked_at IS NULL`, nowString(), nowString(), id)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) ListDockerEndpoints(ctx context.Context) ([]domain.DockerEndpoint, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, name, kind, address, COALESCE(tls_ca_path, ''), COALESCE(tls_cert_path, ''), COALESCE(tls_key_path, ''), enabled, scan_interval_seconds, COALESCE(last_success_at, ''), COALESCE(last_error, ''), created_at, updated_at FROM docker_endpoints ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.DockerEndpoint
	for rows.Next() {
		var item domain.DockerEndpoint
		var enabled int
		var lastSuccessAt, lastError, createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.Name, &item.Kind, &item.Address, &item.TLSCAPath, &item.TLSCertPath, &item.TLSKeyPath, &enabled, &item.ScanIntervalSeconds, &lastSuccessAt, &lastError, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		item.LastSuccessAt = parseTime(lastSuccessAt)
		item.LastError = lastError
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SaveDockerEndpoint(ctx context.Context, endpoint domain.DockerEndpoint) (domain.DockerEndpoint, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DockerEndpoint{}, err
	}
	defer tx.Rollback()
	item, err := s.upsertDockerEndpointTx(ctx, tx, endpoint)
	if err != nil {
		return domain.DockerEndpoint{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.DockerEndpoint{}, err
	}
	return item, nil
}

func (s *Store) upsertDockerEndpointTx(ctx context.Context, tx *sql.Tx, endpoint domain.DockerEndpoint) (domain.DockerEndpoint, error) {
	now := time.Now().UTC()
	if endpoint.ID == "" {
		endpoint.ID = newID("dep")
		endpoint.CreatedAt = now
	}
	if endpoint.ScanIntervalSeconds == 0 {
		endpoint.ScanIntervalSeconds = 30
	}
	endpoint.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO docker_endpoints(id, name, kind, address, tls_ca_path, tls_cert_path, tls_key_path, enabled, scan_interval_seconds, last_success_at, last_error, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, kind = excluded.kind, address = excluded.address, tls_ca_path = excluded.tls_ca_path, tls_cert_path = excluded.tls_cert_path, tls_key_path = excluded.tls_key_path, enabled = excluded.enabled, scan_interval_seconds = excluded.scan_interval_seconds, updated_at = excluded.updated_at
	`, endpoint.ID, endpoint.Name, endpoint.Kind, endpoint.Address, nullableString(endpoint.TLSCAPath), nullableString(endpoint.TLSCertPath), nullableString(endpoint.TLSKeyPath), boolInt(endpoint.Enabled), endpoint.ScanIntervalSeconds, nullableTime(endpoint.LastSuccessAt), nullableString(endpoint.LastError), endpoint.CreatedAt.Format(time.RFC3339Nano), endpoint.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.DockerEndpoint{}, err
	}
	return s.getDockerEndpointTx(ctx, tx, endpoint.ID)
}

func (s *Store) getDockerEndpointTx(ctx context.Context, tx *sql.Tx, id string) (domain.DockerEndpoint, error) {
	var item domain.DockerEndpoint
	var enabled int
	var lastSuccessAt, lastError, createdAt, updatedAt string
	err := tx.QueryRowContext(ctx, `SELECT id, name, kind, address, COALESCE(tls_ca_path, ''), COALESCE(tls_cert_path, ''), COALESCE(tls_key_path, ''), enabled, scan_interval_seconds, COALESCE(last_success_at, ''), COALESCE(last_error, ''), created_at, updated_at FROM docker_endpoints WHERE id = ?`, id).Scan(&item.ID, &item.Name, &item.Kind, &item.Address, &item.TLSCAPath, &item.TLSCertPath, &item.TLSKeyPath, &enabled, &item.ScanIntervalSeconds, &lastSuccessAt, &lastError, &createdAt, &updatedAt)
	if err != nil {
		return domain.DockerEndpoint{}, err
	}
	item.Enabled = enabled == 1
	item.LastSuccessAt = parseTime(lastSuccessAt)
	item.LastError = lastError
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, nil
}

func (s *Store) DeleteDockerEndpoint(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM docker_endpoints WHERE id = ?", id)
	return err
}

func (s *Store) UpdateDockerEndpointStatus(ctx context.Context, id string, succeededAt time.Time, lastError string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE docker_endpoints SET last_success_at = ?, last_error = ?, updated_at = ? WHERE id = ?`, nullableTime(succeededAt), nullableString(lastError), nowString(), id)
	return err
}

func (s *Store) ListScanTargets(ctx context.Context) ([]domain.ScanTarget, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, name, cidr, auto_detected, enabled, scan_interval_seconds, common_ports, created_at, updated_at FROM scan_targets ORDER BY auto_detected DESC, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.ScanTarget
	for rows.Next() {
		var item domain.ScanTarget
		var autoDetected, enabled int
		var commonPorts, createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.Name, &item.CIDR, &autoDetected, &enabled, &item.ScanIntervalSeconds, &commonPorts, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.AutoDetected = autoDetected == 1
		item.Enabled = enabled == 1
		_ = json.Unmarshal([]byte(commonPorts), &item.CommonPorts)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SaveScanTarget(ctx context.Context, target domain.ScanTarget) (domain.ScanTarget, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ScanTarget{}, err
	}
	defer tx.Rollback()
	item, err := s.upsertScanTargetTx(ctx, tx, target)
	if err != nil {
		return domain.ScanTarget{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ScanTarget{}, err
	}
	return item, nil
}

func (s *Store) upsertScanTargetTx(ctx context.Context, tx *sql.Tx, target domain.ScanTarget) (domain.ScanTarget, error) {
	now := time.Now().UTC()
	if target.ID == "" {
		target.ID = newID("stg")
		target.CreatedAt = now
	}
	if target.Name == "" {
		target.Name = target.CIDR
	}
	if target.ScanIntervalSeconds == 0 {
		target.ScanIntervalSeconds = 300
	}
	if len(target.CommonPorts) == 0 {
		target.CommonPorts = []int{22, 80, 443}
	}
	target.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scan_targets(id, name, cidr, auto_detected, enabled, scan_interval_seconds, common_ports, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, cidr = excluded.cidr, auto_detected = excluded.auto_detected, enabled = excluded.enabled, scan_interval_seconds = excluded.scan_interval_seconds, common_ports = excluded.common_ports, updated_at = excluded.updated_at
	`, target.ID, target.Name, target.CIDR, boolInt(target.AutoDetected), boolInt(target.Enabled), target.ScanIntervalSeconds, string(mustJSON(target.CommonPorts)), target.CreatedAt.Format(time.RFC3339Nano), target.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.ScanTarget{}, err
	}
	return s.getScanTargetTx(ctx, tx, target.ID)
}

func (s *Store) getScanTargetTx(ctx context.Context, tx *sql.Tx, id string) (domain.ScanTarget, error) {
	var item domain.ScanTarget
	var autoDetected, enabled int
	var commonPorts, createdAt, updatedAt string
	err := tx.QueryRowContext(ctx, `SELECT id, name, cidr, auto_detected, enabled, scan_interval_seconds, common_ports, created_at, updated_at FROM scan_targets WHERE id = ?`, id).Scan(&item.ID, &item.Name, &item.CIDR, &autoDetected, &enabled, &item.ScanIntervalSeconds, &commonPorts, &createdAt, &updatedAt)
	if err != nil {
		return domain.ScanTarget{}, err
	}
	item.AutoDetected = autoDetected == 1
	item.Enabled = enabled == 1
	_ = json.Unmarshal([]byte(commonPorts), &item.CommonPorts)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	return item, nil
}

func (s *Store) DeleteScanTarget(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM scan_targets WHERE id = ?", id)
	return err
}

func (s *Store) ListServices(ctx context.Context) ([]domain.Service, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT s.id, s.name, s.slug, s.source_type, s.source_ref, COALESCE(s.origin_discovered_service_id, ''), COALESCE(s.service_definition_id, ''), COALESCE(s.service_type, ''), COALESCE(s.health_config_mode, 'auto'), COALESCE(s.address_source, 'literal_host'), COALESCE(s.host_value, s.host, ''), COALESCE(s.device_id, ''), COALESCE(d.display_name, d.hostname, ''), COALESCE(s.icon, ''), COALESCE(s.scheme, ''), s.host, s.port, COALESCE(s.path, ''), s.url, s.hidden, s.status, COALESCE(s.last_seen_at, ''), COALESCE(s.last_checked_at, ''), COALESCE(s.fingerprinted_at, ''), s.details_json, s.created_at, s.updated_at FROM services s LEFT JOIN devices d ON d.id = s.device_id ORDER BY s.hidden ASC, s.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.Service
	for rows.Next() {
		item, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.resolveServices(ctx, items); err != nil {
		return nil, err
	}
	if err := s.attachChecks(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetService(ctx context.Context, id string) (domain.Service, error) {
	row := s.reader().QueryRowContext(ctx, `SELECT s.id, s.name, s.slug, s.source_type, s.source_ref, COALESCE(s.origin_discovered_service_id, ''), COALESCE(s.service_definition_id, ''), COALESCE(s.service_type, ''), COALESCE(s.health_config_mode, 'auto'), COALESCE(s.address_source, 'literal_host'), COALESCE(s.host_value, s.host, ''), COALESCE(s.device_id, ''), COALESCE(d.display_name, d.hostname, ''), COALESCE(s.icon, ''), COALESCE(s.scheme, ''), s.host, s.port, COALESCE(s.path, ''), s.url, s.hidden, s.status, COALESCE(s.last_seen_at, ''), COALESCE(s.last_checked_at, ''), COALESCE(s.fingerprinted_at, ''), s.details_json, s.created_at, s.updated_at FROM services s LEFT JOIN devices d ON d.id = s.device_id WHERE s.id = ?`, id)
	item, err := scanService(row)
	if err != nil {
		return domain.Service{}, err
	}
	resolved := []domain.Service{item}
	if err := s.resolveServices(ctx, resolved); err != nil {
		return domain.Service{}, err
	}
	item = resolved[0]
	checks, err := s.ListServiceChecks(ctx, id)
	if err != nil {
		return domain.Service{}, err
	}
	item.Checks = checks
	return item, nil
}

func (s *Store) SaveManualService(ctx context.Context, service domain.Service) (domain.Service, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Service{}, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if service.ID == "" {
		service.ID = newID("svc")
		service.CreatedAt = now
	}
	if service.Source == "" {
		service.Source = domain.ServiceSourceManual
	}
	if service.SourceRef == "" {
		service.SourceRef = service.ID
	}
	if service.Slug == "" {
		service.Slug = slugify(service.Name)
	}
	if service.Status == "" {
		service.Status = domain.HealthStatusUnknown
	}
	if service.HealthConfigMode == "" {
		service.HealthConfigMode = domain.HealthConfigModeAuto
	}
	if service.AddressSource == "" {
		service.AddressSource = domain.ServiceAddressLiteralHost
	}
	if service.AddressSource == domain.ServiceAddressLiteralHost && strings.TrimSpace(service.Host) != "" {
		service.HostValue = service.Host
	} else if service.HostValue == "" {
		service.HostValue = firstNonEmpty(service.Host, extractHost(service.URL))
	}
	service.Host = resolveAddressSourceHost(service.AddressSource, service.HostValue, "")
	if service.URL == "" {
		service.URL = buildServiceURL(service.Scheme, service.Host, service.Port, service.Path)
	}
	if service.Details == nil {
		service.Details = map[string]any{}
	}
	service.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `INSERT INTO services(id, name, slug, source_type, source_ref, origin_discovered_service_id, service_definition_id, service_type, health_config_mode, address_source, host_value, device_id, icon, scheme, host, port, path, url, hidden, status, last_seen_at, last_checked_at, fingerprinted_at, details_json, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET name = excluded.name, slug = excluded.slug, source_type = excluded.source_type, source_ref = excluded.source_ref, origin_discovered_service_id = excluded.origin_discovered_service_id, service_definition_id = excluded.service_definition_id, service_type = excluded.service_type, health_config_mode = excluded.health_config_mode, address_source = excluded.address_source, host_value = excluded.host_value, device_id = excluded.device_id, icon = excluded.icon, scheme = excluded.scheme, host = excluded.host, port = excluded.port, path = excluded.path, url = excluded.url, hidden = excluded.hidden, details_json = excluded.details_json, fingerprinted_at = excluded.fingerprinted_at, updated_at = excluded.updated_at`, service.ID, service.Name, service.Slug, service.Source, service.SourceRef, nullableString(service.OriginDiscoveredServiceID), nullableString(service.ServiceDefinitionID), service.ServiceType, service.HealthConfigMode, service.AddressSource, service.HostValue, nullableString(service.DeviceID), nullableString(service.Icon), nullableString(service.Scheme), service.Host, service.Port, nullableString(service.Path), service.URL, boolInt(service.Hidden), service.Status, nullableTime(service.LastSeenAt), nullableTime(service.LastCheckedAt), nullableTime(service.FingerprintedAt), string(mustJSON(service.Details)), service.CreatedAt.Format(time.RFC3339Nano), service.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.Service{}, err
	}
	if err := s.ensureDefaultCheckTx(ctx, tx, service); err != nil {
		return domain.Service{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Service{}, err
	}
	return s.GetService(ctx, service.ID)
}

func (s *Store) UpsertDiscoveredService(ctx context.Context, observation domain.ServiceObservation, deviceID string) (domain.Service, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Service{}, err
	}
	defer tx.Rollback()
	now := observation.LastSeenAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	service := domain.Service{ID: newID("svc"), Name: observation.Name, Slug: slugify(observation.Name), Source: observation.Source, SourceRef: observation.SourceRef, ServiceType: observation.ServiceTypeHint, HealthConfigMode: domain.HealthConfigModeAuto, AddressSource: firstNonEmptyAddressSource(observation.AddressSource, domain.ServiceAddressLiteralHost), HostValue: firstNonEmpty(observation.HostValue, observation.Host), DeviceID: deviceID, Icon: observation.Icon, Scheme: observation.Scheme, Host: resolveAddressSourceHost(firstNonEmptyAddressSource(observation.AddressSource, domain.ServiceAddressLiteralHost), firstNonEmpty(observation.HostValue, observation.Host), ""), Port: observation.Port, Path: observation.Path, URL: firstNonEmpty(observation.URL, buildServiceURL(observation.Scheme, firstNonEmpty(observation.HostValue, observation.Host), observation.Port, observation.Path)), Status: domain.HealthStatusUnknown, LastSeenAt: now, Details: observation.Details, CreatedAt: now, UpdatedAt: now}
	if _, err := tx.ExecContext(ctx, `INSERT INTO services(id, name, slug, source_type, source_ref, origin_discovered_service_id, service_definition_id, service_type, health_config_mode, address_source, host_value, device_id, icon, scheme, host, port, path, url, hidden, status, last_seen_at, last_checked_at, fingerprinted_at, details_json, created_at, updated_at) VALUES(?, ?, ?, ?, ?, NULL, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, NULL, NULL, ?, ?, ?) ON CONFLICT(source_type, source_ref) DO UPDATE SET name = excluded.name, service_type = excluded.service_type, health_config_mode = excluded.health_config_mode, address_source = excluded.address_source, host_value = excluded.host_value, device_id = excluded.device_id, icon = excluded.icon, scheme = excluded.scheme, host = excluded.host, port = excluded.port, path = excluded.path, url = excluded.url, last_seen_at = excluded.last_seen_at, details_json = excluded.details_json, updated_at = excluded.updated_at`, service.ID, service.Name, service.Slug, service.Source, service.SourceRef, service.ServiceType, service.HealthConfigMode, service.AddressSource, service.HostValue, nullableString(deviceID), nullableString(service.Icon), nullableString(service.Scheme), service.Host, service.Port, nullableString(service.Path), service.URL, service.Status, service.LastSeenAt.Format(time.RFC3339Nano), string(mustJSON(service.Details)), service.CreatedAt.Format(time.RFC3339Nano), service.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.Service{}, err
	}
	if err := tx.QueryRowContext(ctx, "SELECT id FROM services WHERE source_type = ? AND source_ref = ?", observation.Source, observation.SourceRef).Scan(&service.ID); err != nil {
		return domain.Service{}, err
	}
	if err := s.ensureDefaultCheckTx(ctx, tx, service); err != nil {
		return domain.Service{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Service{}, err
	}
	return s.GetService(ctx, service.ID)
}

func (s *Store) DeleteService(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM services WHERE id = ?", id)
	return err
}

func (s *Store) ListServiceEvents(ctx context.Context, serviceID string, limit int) ([]domain.ServiceEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.reader().QueryContext(ctx, `SELECT id, service_id, event_type, status, message, created_at FROM service_events WHERE service_id = ? ORDER BY created_at DESC LIMIT ?`, serviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.ServiceEvent
	for rows.Next() {
		var item domain.ServiceEvent
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.EventType, &item.Status, &item.Message, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) listLegacyServiceChecks(ctx context.Context, serviceID string) ([]domain.ServiceCheck, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT c.id, c.service_id, c.name, c.type, c.target, c.interval_seconds, c.timeout_seconds, c.expected_status_min, c.expected_status_max, c.enabled, c.created_at, c.updated_at, COALESCE(r.id, ''), COALESCE(r.status, ''), COALESCE(r.latency_ms, 0), COALESCE(r.message, ''), COALESCE(r.checked_at, '') FROM service_checks c LEFT JOIN (SELECT cr1.* FROM check_results cr1 JOIN (SELECT check_id, MAX(checked_at) AS checked_at FROM check_results GROUP BY check_id) latest ON latest.check_id = cr1.check_id AND latest.checked_at = cr1.checked_at) r ON r.check_id = c.id WHERE c.service_id = ? ORDER BY c.name`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.ServiceCheck
	for rows.Next() {
		var item domain.ServiceCheck
		var enabled int
		var createdAt, updatedAt string
		var resultID, resultStatus, resultMessage, resultCheckedAt string
		var resultLatency int64
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.Name, &item.Type, &item.Target, &item.IntervalSeconds, &item.TimeoutSeconds, &item.ExpectedStatusMin, &item.ExpectedStatusMax, &enabled, &createdAt, &updatedAt, &resultID, &resultStatus, &resultLatency, &resultMessage, &resultCheckedAt); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		if resultID != "" {
			item.LastResult = &domain.CheckResult{ID: resultID, CheckID: item.ID, ServiceID: item.ServiceID, Status: domain.HealthStatus(resultStatus), LatencyMS: resultLatency, Message: resultMessage, CheckedAt: parseTime(resultCheckedAt)}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) saveLegacyServiceCheck(ctx context.Context, check domain.ServiceCheck) (domain.ServiceCheck, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ServiceCheck{}, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if check.ID == "" {
		check.ID = newID("chk")
		check.CreatedAt = now
	}
	if check.Name == "" {
		check.Name = strings.ToUpper(string(check.Type)) + " check"
	}
	if check.IntervalSeconds == 0 {
		check.IntervalSeconds = 60
	}
	if check.TimeoutSeconds == 0 {
		check.TimeoutSeconds = 10
	}
	if check.Type == domain.CheckTypeHTTP {
		if check.ExpectedStatusMin == 0 {
			check.ExpectedStatusMin = 200
		}
		if check.ExpectedStatusMax == 0 {
			check.ExpectedStatusMax = 399
		}
	}
	check.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `INSERT INTO service_checks(id, service_id, name, type, target, interval_seconds, timeout_seconds, expected_status_min, expected_status_max, enabled, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET name = excluded.name, type = excluded.type, target = excluded.target, interval_seconds = excluded.interval_seconds, timeout_seconds = excluded.timeout_seconds, expected_status_min = excluded.expected_status_min, expected_status_max = excluded.expected_status_max, enabled = excluded.enabled, updated_at = excluded.updated_at`, check.ID, check.ServiceID, check.Name, check.Type, check.Target, check.IntervalSeconds, check.TimeoutSeconds, check.ExpectedStatusMin, check.ExpectedStatusMax, boolInt(check.Enabled), check.CreatedAt.Format(time.RFC3339Nano), check.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.ServiceCheck{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ServiceCheck{}, err
	}
	items, err := s.ListServiceChecks(ctx, check.ServiceID)
	if err != nil {
		return domain.ServiceCheck{}, err
	}
	for _, item := range items {
		if item.ID == check.ID {
			return item, nil
		}
	}
	return domain.ServiceCheck{}, sql.ErrNoRows
}

func (s *Store) deleteLegacyServiceCheck(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM service_checks WHERE id = ?", id)
	return err
}

func (s *Store) ListDevices(ctx context.Context) ([]domain.Device, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, identity_key, COALESCE(primary_mac, ''), COALESCE(hostname, ''), COALESCE(display_name, ''), identity_confidence, hidden, first_seen_at, last_seen_at, created_at, updated_at FROM devices ORDER BY hidden ASC, COALESCE(display_name, hostname, identity_key)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.Device
	for rows.Next() {
		var item domain.Device
		var hidden int
		var firstSeenAt, lastSeenAt, createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.IdentityKey, &item.PrimaryMAC, &item.Hostname, &item.DisplayName, &item.IdentityConfidence, &hidden, &firstSeenAt, &lastSeenAt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.Hidden = hidden == 1
		item.FirstSeenAt = parseTime(firstSeenAt)
		item.LastSeenAt = parseTime(lastSeenAt)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.attachDeviceDetails(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetDevice(ctx context.Context, id string) (domain.Device, error) {
	items, err := s.ListDevices(ctx)
	if err != nil {
		return domain.Device{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.Device{}, sql.ErrNoRows
}

func (s *Store) UpdateDevice(ctx context.Context, id string, displayName *string, hidden *bool) (domain.Device, error) {
	if _, err := s.db.ExecContext(ctx, `UPDATE devices SET display_name = COALESCE(?, display_name), hidden = COALESCE(?, hidden), updated_at = ? WHERE id = ?`, nullablePtrString(displayName), boolPtrInt(hidden), nowString(), id); err != nil {
		return domain.Device{}, err
	}
	return s.GetDevice(ctx, id)
}

func (s *Store) UpsertDeviceObservation(ctx context.Context, observation domain.DeviceObservation) (domain.Device, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Device{}, err
	}
	defer tx.Rollback()
	if observation.LastSeenAt.IsZero() {
		observation.LastSeenAt = time.Now().UTC()
	}
	device, err := s.findOrCreateDeviceTx(ctx, tx, observation)
	if err != nil {
		return domain.Device{}, err
	}
	if err := s.upsertDeviceAddressTx(ctx, tx, device.ID, observation); err != nil {
		return domain.Device{}, err
	}
	for _, port := range observation.Ports {
		if err := s.upsertDevicePortTx(ctx, tx, device.ID, port, observation.LastSeenAt); err != nil {
			return domain.Device{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Device{}, err
	}
	return s.GetDevice(ctx, device.ID)
}

func (s *Store) findOrCreateDeviceTx(ctx context.Context, tx *sql.Tx, observation domain.DeviceObservation) (domain.Device, error) {
	now := observation.LastSeenAt.Format(time.RFC3339Nano)
	var item domain.Device
	var hidden int
	var firstSeenAt, lastSeenAt, createdAt, updatedAt string
	err := tx.QueryRowContext(ctx, `SELECT id, identity_key, COALESCE(primary_mac, ''), COALESCE(hostname, ''), COALESCE(display_name, ''), identity_confidence, hidden, first_seen_at, last_seen_at, created_at, updated_at FROM devices WHERE identity_key = ? OR (? <> '' AND primary_mac = ?) LIMIT 1`, observation.IdentityKey, observation.PrimaryMAC, observation.PrimaryMAC).Scan(&item.ID, &item.IdentityKey, &item.PrimaryMAC, &item.Hostname, &item.DisplayName, &item.IdentityConfidence, &hidden, &firstSeenAt, &lastSeenAt, &createdAt, &updatedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Device{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		item = domain.Device{ID: newID("dev"), IdentityKey: observation.IdentityKey, PrimaryMAC: observation.PrimaryMAC, Hostname: observation.Hostname, DisplayName: firstNonEmpty(observation.DisplayName, observation.Hostname, observation.IPAddress, observation.PrimaryMAC, observation.IdentityKey), IdentityConfidence: observation.Confidence, FirstSeenAt: observation.LastSeenAt, LastSeenAt: observation.LastSeenAt, CreatedAt: observation.LastSeenAt, UpdatedAt: observation.LastSeenAt}
		if item.IdentityConfidence == "" {
			item.IdentityConfidence = domain.IdentityConfidenceLow
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO devices(id, identity_key, primary_mac, hostname, display_name, identity_confidence, hidden, first_seen_at, last_seen_at, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)`, item.ID, item.IdentityKey, nullableString(item.PrimaryMAC), nullableString(item.Hostname), nullableString(item.DisplayName), item.IdentityConfidence, item.FirstSeenAt.Format(time.RFC3339Nano), item.LastSeenAt.Format(time.RFC3339Nano), item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
			return domain.Device{}, err
		}
		return item, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE devices SET primary_mac = COALESCE(NULLIF(?, ''), primary_mac), hostname = COALESCE(NULLIF(?, ''), hostname), display_name = COALESCE(NULLIF(?, ''), display_name), identity_confidence = ?, last_seen_at = ?, updated_at = ? WHERE id = ?`, observation.PrimaryMAC, observation.Hostname, observation.DisplayName, firstNonEmpty(string(observation.Confidence), string(item.IdentityConfidence)), now, now, item.ID); err != nil {
		return domain.Device{}, err
	}
	return item, nil
}

func (s *Store) upsertDeviceAddressTx(ctx context.Context, tx *sql.Tx, deviceID string, observation domain.DeviceObservation) error {
	if observation.IPAddress == "" {
		return nil
	}
	now := observation.LastSeenAt.Format(time.RFC3339Nano)
	_, err := tx.ExecContext(ctx, `INSERT INTO device_addresses(id, device_id, ip_address, mac_address, interface_name, is_primary, first_seen_at, last_seen_at) VALUES(?, ?, ?, ?, ?, 1, ?, ?) ON CONFLICT(device_id, ip_address) DO UPDATE SET mac_address = COALESCE(NULLIF(excluded.mac_address, ''), device_addresses.mac_address), interface_name = COALESCE(NULLIF(excluded.interface_name, ''), device_addresses.interface_name), is_primary = 1, last_seen_at = excluded.last_seen_at`, newID("adr"), deviceID, observation.IPAddress, nullableString(observation.PrimaryMAC), nullableString(observation.Interface), now, now)
	return err
}

func (s *Store) upsertDevicePortTx(ctx context.Context, tx *sql.Tx, deviceID string, port domain.PortObservation, seenAt time.Time) error {
	now := seenAt.Format(time.RFC3339Nano)
	_, err := tx.ExecContext(ctx, `INSERT INTO device_ports(id, device_id, port, protocol, service_hint, open, first_seen_at, last_seen_at) VALUES(?, ?, ?, ?, ?, 1, ?, ?) ON CONFLICT(device_id, port, protocol) DO UPDATE SET service_hint = excluded.service_hint, open = 1, last_seen_at = excluded.last_seen_at`, newID("prt"), deviceID, port.Port, port.Protocol, nullableString(port.ServiceHint), now, now)
	return err
}

func (s *Store) GetDashboard(ctx context.Context) (domain.Dashboard, error) {
	services, err := s.ListServices(ctx)
	if err != nil {
		return domain.Dashboard{}, err
	}
	devices, err := s.ListDevices(ctx)
	if err != nil {
		return domain.Dashboard{}, err
	}
	bookmarks, err := s.ListBookmarks(ctx)
	if err != nil {
		return domain.Dashboard{}, err
	}
	discoveredServices, err := s.ListDiscoveredServices(ctx)
	if err != nil {
		return domain.Dashboard{}, err
	}
	events, err := s.ListRecentEvents(ctx, 20)
	if err != nil {
		return domain.Dashboard{}, err
	}
	summary := domain.DashboardSummary{TotalServices: len(services), DevicesSeen: len(devices), Bookmarks: len(bookmarks)}
	for _, item := range discoveredServices {
		if item.State == domain.DiscoveryStatePending {
			summary.DiscoveredServices++
		}
	}
	containers := make([]domain.Service, 0)
	for _, item := range services {
		if item.Source == domain.ServiceSourceDocker {
			containers = append(containers, item)
			summary.RunningContainers++
		}
		switch item.Status {
		case domain.HealthStatusHealthy:
			summary.HealthyServices++
		case domain.HealthStatusDegraded:
			summary.DegradedServices++
		case domain.HealthStatusUnhealthy:
			summary.UnhealthyServices++
		}
	}
	return domain.Dashboard{
		Summary:            summary,
		Services:           services,
		Containers:         containers,
		Devices:            devices,
		Bookmarks:          bookmarks,
		DiscoveredServices: discoveredServices,
		RecentEvents:       events,
	}, nil
}

func (s *Store) ListRecentEvents(ctx context.Context, limit int) ([]domain.ServiceEvent, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, service_id, event_type, status, message, created_at FROM service_events ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.ServiceEvent
	for rows.Next() {
		var item domain.ServiceEvent
		var createdAt string
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.EventType, &item.Status, &item.Message, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) saveLegacyCheckResult(ctx context.Context, result domain.CheckResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if result.ID == "" {
		result.ID = newID("res")
	}
	if result.CheckedAt.IsZero() {
		result.CheckedAt = time.Now().UTC()
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO check_results(id, check_id, service_id, status, latency_ms, message, checked_at) VALUES(?, ?, ?, ?, ?, ?, ?)`, result.ID, result.CheckID, result.ServiceID, result.Status, result.LatencyMS, nullableString(result.Message), result.CheckedAt.Format(time.RFC3339Nano)); err != nil {
		return err
	}
	var previousStatus string
	if err := tx.QueryRowContext(ctx, "SELECT status FROM services WHERE id = ?", result.ServiceID).Scan(&previousStatus); err != nil {
		return err
	}
	nextStatus, err := s.rollupServiceStatusTx(ctx, tx, result.ServiceID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE services SET status = ?, last_checked_at = ?, updated_at = ? WHERE id = ?`, nextStatus, result.CheckedAt.Format(time.RFC3339Nano), nowString(), result.ServiceID); err != nil {
		return err
	}
	if previousStatus != string(nextStatus) {
		if err := s.insertServiceEventTx(ctx, tx, result.ServiceID, "health_changed", nextStatus, firstNonEmpty(result.Message, fmt.Sprintf("service status changed to %s", nextStatus))); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) rollupLegacyServiceStatusTx(ctx context.Context, tx *sql.Tx, serviceID string) (domain.HealthStatus, error) {
	rows, err := tx.QueryContext(ctx, `SELECT COALESCE(r.status, 'unknown') FROM service_checks c LEFT JOIN (SELECT cr1.* FROM check_results cr1 JOIN (SELECT check_id, MAX(checked_at) AS checked_at FROM check_results GROUP BY check_id) latest ON latest.check_id = cr1.check_id AND latest.checked_at = cr1.checked_at) r ON r.check_id = c.id WHERE c.service_id = ? AND c.enabled = 1`, serviceID)
	if err != nil {
		return domain.HealthStatusUnknown, err
	}
	defer rows.Close()
	var statuses []domain.HealthStatus
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return domain.HealthStatusUnknown, err
		}
		statuses = append(statuses, domain.HealthStatus(status))
	}
	if len(statuses) == 0 {
		return domain.HealthStatusUnknown, nil
	}
	allHealthy := true
	allUnhealthy := true
	for _, status := range statuses {
		switch status {
		case domain.HealthStatusHealthy:
			allUnhealthy = false
		case domain.HealthStatusUnhealthy:
			allHealthy = false
		default:
			allHealthy = false
			allUnhealthy = false
		}
	}
	if allHealthy {
		return domain.HealthStatusHealthy, nil
	}
	if allUnhealthy {
		return domain.HealthStatusUnhealthy, nil
	}
	return domain.HealthStatusDegraded, nil
}

func (s *Store) getLegacyChecksDue(ctx context.Context) ([]domain.MonitorCheck, error) {
	primaryAddresses, err := s.loadPrimaryDeviceAddresses(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.reader().QueryContext(ctx, `SELECT
		c.id, c.service_id, c.name, c.type, c.target, c.interval_seconds, c.timeout_seconds, c.expected_status_min, c.expected_status_max, c.enabled, c.created_at, c.updated_at,
		COALESCE(r.id, ''), COALESCE(r.status, ''), COALESCE(r.latency_ms, 0), COALESCE(r.message, ''), COALESCE(r.checked_at, ''),
		s.id, s.name, s.slug, s.source_type, s.source_ref, COALESCE(s.origin_discovered_service_id, ''), COALESCE(s.service_type, ''), COALESCE(s.address_source, 'literal_host'), COALESCE(s.host_value, s.host, ''), COALESCE(s.device_id, ''), '', COALESCE(s.icon, ''), COALESCE(s.scheme, ''), s.host, s.port, COALESCE(s.path, ''), s.url, s.hidden, s.status, COALESCE(s.last_seen_at, ''), COALESCE(s.last_checked_at, ''), s.details_json, s.created_at, s.updated_at
	FROM service_checks c
	JOIN services s ON s.id = c.service_id
	LEFT JOIN (
		SELECT cr1.*
		FROM check_results cr1
		JOIN (
			SELECT check_id, MAX(checked_at) AS checked_at
			FROM check_results
			GROUP BY check_id
		) latest ON latest.check_id = cr1.check_id AND latest.checked_at = cr1.checked_at
	) r ON r.check_id = c.id
	WHERE c.enabled = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.MonitorCheck
	for rows.Next() {
		var check domain.ServiceCheck
		var service domain.Service
		var enabled, hidden int
		var createdAt, updatedAt, resultID, resultStatus, resultMessage, resultCheckedAt string
		var resultLatency int64
		var lastSeenAt, lastCheckedAt, detailsJSON, svcCreatedAt, svcUpdatedAt string
		if err := rows.Scan(&check.ID, &check.ServiceID, &check.Name, &check.Type, &check.Target, &check.IntervalSeconds, &check.TimeoutSeconds, &check.ExpectedStatusMin, &check.ExpectedStatusMax, &enabled, &createdAt, &updatedAt, &resultID, &resultStatus, &resultLatency, &resultMessage, &resultCheckedAt, &service.ID, &service.Name, &service.Slug, &service.Source, &service.SourceRef, &service.OriginDiscoveredServiceID, &service.ServiceType, &service.AddressSource, &service.HostValue, &service.DeviceID, &service.DeviceName, &service.Icon, &service.Scheme, &service.Host, &service.Port, &service.Path, &service.URL, &hidden, &service.Status, &lastSeenAt, &lastCheckedAt, &detailsJSON, &svcCreatedAt, &svcUpdatedAt); err != nil {
			return nil, err
		}
		check.Enabled = enabled == 1
		check.CreatedAt = parseTime(createdAt)
		check.UpdatedAt = parseTime(updatedAt)
		if resultID != "" {
			check.LastResult = &domain.CheckResult{
				ID:        resultID,
				CheckID:   check.ID,
				ServiceID: check.ServiceID,
				Status:    domain.HealthStatus(resultStatus),
				LatencyMS: resultLatency,
				Message:   resultMessage,
				CheckedAt: parseTime(resultCheckedAt),
			}
		}
		service.Hidden = hidden == 1
		service.LastSeenAt = parseTime(lastSeenAt)
		service.LastCheckedAt = parseTime(lastCheckedAt)
		service.CreatedAt = parseTime(svcCreatedAt)
		service.UpdatedAt = parseTime(svcUpdatedAt)
		_ = json.Unmarshal([]byte(detailsJSON), &service.Details)
		service.Host = resolveAddressSourceHost(service.AddressSource, service.HostValue, primaryAddresses[service.DeviceID])
		service.URL = buildServiceURL(service.Scheme, service.Host, service.Port, service.Path)
		switch check.Type {
		case domain.CheckTypeHTTP:
			check.Target = service.URL
		case domain.CheckTypeTCP:
			check.Target = fmt.Sprintf("%s:%d", service.Host, service.Port)
		case domain.CheckTypePing:
			check.Target = service.Host
		}
		items = append(items, domain.MonitorCheck{Check: check, Service: service})
	}
	return items, rows.Err()
}

func (s *Store) legacyCleanup(ctx context.Context, retain time.Duration) error {
	cutoff := time.Now().UTC().Add(-retain).Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, "DELETE FROM check_results WHERE checked_at < ?", cutoff); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "DELETE FROM service_events WHERE created_at < ?", cutoff); err != nil {
		return err
	}
	return nil
}

func (s *Store) RecordJobRun(ctx context.Context, jobName string, runErr error) error {
	now := nowString()
	lastSuccess := ""
	lastError := ""
	if runErr == nil {
		lastSuccess = now
	} else {
		lastError = runErr.Error()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO job_state(job_name, last_run_at, last_success_at, last_error, updated_at) VALUES(?, ?, ?, ?, ?) ON CONFLICT(job_name) DO UPDATE SET last_run_at = excluded.last_run_at, last_success_at = CASE WHEN excluded.last_success_at <> '' THEN excluded.last_success_at ELSE job_state.last_success_at END, last_error = excluded.last_error, updated_at = excluded.updated_at`, jobName, now, lastSuccess, lastError, now)
	return err
}

func (s *Store) ListJobState(ctx context.Context) ([]domain.JobState, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT job_name, COALESCE(last_run_at, ''), COALESCE(last_success_at, ''), COALESCE(last_error, ''), updated_at FROM job_state ORDER BY job_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.JobState
	for rows.Next() {
		var item domain.JobState
		var lastRunAt, lastSuccessAt, updatedAt string
		if err := rows.Scan(&item.JobName, &lastRunAt, &lastSuccessAt, &item.LastError, &updatedAt); err != nil {
			return nil, err
		}
		item.LastRunAt = parseTime(lastRunAt)
		item.LastSuccessAt = parseTime(lastSuccessAt)
		item.UpdatedAt = parseTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) attachChecks(ctx context.Context, services []domain.Service) error {
	for index := range services {
		checks, err := s.ListServiceChecks(ctx, services[index].ID)
		if err != nil {
			return err
		}
		services[index].Checks = checks
	}
	return nil
}

func (s *Store) resolveServices(ctx context.Context, services []domain.Service) error {
	if len(services) == 0 {
		return nil
	}
	primaryAddresses, err := s.loadPrimaryDeviceAddresses(ctx)
	if err != nil {
		return err
	}
	for index := range services {
		services[index].Host = resolveAddressSourceHost(services[index].AddressSource, services[index].HostValue, primaryAddresses[services[index].DeviceID])
		services[index].URL = buildServiceURL(services[index].Scheme, services[index].Host, services[index].Port, services[index].Path)
	}
	return nil
}

func (s *Store) attachDeviceDetails(ctx context.Context, devices []domain.Device) error {
	for index := range devices {
		addresses, err := s.listDeviceAddresses(ctx, devices[index].ID)
		if err != nil {
			return err
		}
		ports, err := s.listDevicePorts(ctx, devices[index].ID)
		if err != nil {
			return err
		}
		devices[index].Addresses = addresses
		devices[index].Ports = ports
	}
	return nil
}

func (s *Store) listDeviceAddresses(ctx context.Context, deviceID string) ([]domain.DeviceAddress, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, device_id, ip_address, COALESCE(mac_address, ''), COALESCE(interface_name, ''), is_primary, first_seen_at, last_seen_at FROM device_addresses WHERE device_id = ? ORDER BY is_primary DESC, ip_address`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.DeviceAddress
	for rows.Next() {
		var item domain.DeviceAddress
		var isPrimary int
		var firstSeenAt, lastSeenAt string
		if err := rows.Scan(&item.ID, &item.DeviceID, &item.IPAddress, &item.MACAddress, &item.InterfaceName, &isPrimary, &firstSeenAt, &lastSeenAt); err != nil {
			return nil, err
		}
		item.IsPrimary = isPrimary == 1
		item.FirstSeenAt = parseTime(firstSeenAt)
		item.LastSeenAt = parseTime(lastSeenAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) listDevicePorts(ctx context.Context, deviceID string) ([]domain.DevicePort, error) {
	rows, err := s.reader().QueryContext(ctx, `SELECT id, device_id, port, protocol, COALESCE(service_hint, ''), open, first_seen_at, last_seen_at FROM device_ports WHERE device_id = ? ORDER BY port`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.DevicePort
	for rows.Next() {
		var item domain.DevicePort
		var open int
		var firstSeenAt, lastSeenAt string
		if err := rows.Scan(&item.ID, &item.DeviceID, &item.Port, &item.Protocol, &item.ServiceHint, &open, &firstSeenAt, &lastSeenAt); err != nil {
			return nil, err
		}
		item.Open = open == 1
		item.FirstSeenAt = parseTime(firstSeenAt)
		item.LastSeenAt = parseTime(lastSeenAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ensureLegacyDefaultCheckTx(ctx context.Context, tx *sql.Tx, service domain.Service) error {
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM service_checks WHERE service_id = ?", service.ID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	check := domain.ServiceCheck{ID: newID("chk"), ServiceID: service.ID, IntervalSeconds: 60, TimeoutSeconds: 10, Enabled: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	switch {
	case strings.HasPrefix(service.URL, "https://"), service.Scheme == "https":
		check.Name = "HTTPS availability"
		check.Type = domain.CheckTypeHTTP
		check.Target = service.URL
		check.ExpectedStatusMin = 200
		check.ExpectedStatusMax = 399
	case strings.HasPrefix(service.URL, "http://"), service.Scheme == "http":
		check.Name = "HTTP availability"
		check.Type = domain.CheckTypeHTTP
		check.Target = service.URL
		check.ExpectedStatusMin = 200
		check.ExpectedStatusMax = 399
	default:
		check.Name = "TCP connectivity"
		check.Type = domain.CheckTypeTCP
		check.Target = fmt.Sprintf("%s:%d", service.Host, service.Port)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO service_checks(id, service_id, name, type, target, interval_seconds, timeout_seconds, expected_status_min, expected_status_max, enabled, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, check.ID, check.ServiceID, check.Name, check.Type, check.Target, check.IntervalSeconds, check.TimeoutSeconds, check.ExpectedStatusMin, check.ExpectedStatusMax, boolInt(check.Enabled), check.CreatedAt.Format(time.RFC3339Nano), check.UpdatedAt.Format(time.RFC3339Nano))
	return err
}

func (s *Store) insertServiceEventTx(ctx context.Context, tx *sql.Tx, serviceID, eventType string, status domain.HealthStatus, message string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO service_events(id, service_id, event_type, status, message, created_at) VALUES(?, ?, ?, ?, ?, ?)`, newID("evt"), serviceID, eventType, status, message, nowString())
	return err
}

func scanService(scanner interface{ Scan(dest ...any) error }) (domain.Service, error) {
	var item domain.Service
	var originDiscoveredServiceID, serviceDefinitionID, serviceType, healthConfigMode, addressSource, hostValue, deviceID, deviceName, icon, scheme, path, lastSeenAt, lastCheckedAt, fingerprintedAt, detailsJSON, createdAt, updatedAt string
	var hidden int
	if err := scanner.Scan(&item.ID, &item.Name, &item.Slug, &item.Source, &item.SourceRef, &originDiscoveredServiceID, &serviceDefinitionID, &serviceType, &healthConfigMode, &addressSource, &hostValue, &deviceID, &deviceName, &icon, &scheme, &item.Host, &item.Port, &path, &item.URL, &hidden, &item.Status, &lastSeenAt, &lastCheckedAt, &fingerprintedAt, &detailsJSON, &createdAt, &updatedAt); err != nil {
		return domain.Service{}, err
	}
	item.OriginDiscoveredServiceID = originDiscoveredServiceID
	item.ServiceDefinitionID = serviceDefinitionID
	item.ServiceType = serviceType
	item.HealthConfigMode = domain.HealthConfigMode(healthConfigMode)
	item.AddressSource = domain.ServiceAddressSource(addressSource)
	item.HostValue = hostValue
	item.DeviceID = deviceID
	item.DeviceName = deviceName
	item.Icon = icon
	item.Scheme = scheme
	item.Path = path
	item.Hidden = hidden == 1
	item.LastSeenAt = parseTime(lastSeenAt)
	item.LastCheckedAt = parseTime(lastCheckedAt)
	item.FingerprintedAt = parseTime(fingerprintedAt)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	_ = json.Unmarshal([]byte(detailsJSON), &item.Details)
	return item, nil
}

func nowString() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func boolPtrInt(value *bool) any {
	if value == nil {
		return nil
	}
	if *value {
		return 1
	}
	return 0
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullablePtrString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return *value
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return prefix + "_" + hex.EncodeToString(buf)
}

func generateAPITokenSecret() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return "hlw_" + base64.RawURLEncoding.EncodeToString(buffer), nil
}

func tokenPrefix(secret string) string {
	if len(secret) <= 12 {
		return secret
	}
	return secret[:12]
}

func tokenScopeAllows(actual, required domain.TokenScope) bool {
	if actual == domain.TokenScopeWrite {
		return true
	}
	return actual == required
}

func slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", ":", "-", ".", "-")
	input = replacer.Replace(input)
	input = strings.Trim(input, "-")
	if input == "" {
		return newID("svc")
	}
	return input
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

func extractHost(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
