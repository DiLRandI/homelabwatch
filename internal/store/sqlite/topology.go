package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (s *Store) ListTopologySources(ctx context.Context) ([]domain.TopologySource, error) {
	return s.listTopologySources(ctx, false)
}

func (s *Store) ListTopologySourcesForDiscovery(ctx context.Context) ([]domain.TopologySource, error) {
	return s.listTopologySources(ctx, true)
}

func (s *Store) listTopologySources(ctx context.Context, includeSecrets bool) ([]domain.TopologySource, error) {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, name, address, port, enabled, poll_interval_seconds, timeout_ms, retries, snmp_version,
			COALESCE(community, ''), COALESCE(username, ''), COALESCE(auth_protocol, ''), COALESCE(auth_passphrase, ''),
			COALESCE(privacy_protocol, ''), COALESCE(privacy_passphrase, ''), role, root,
			COALESCE(last_success_at, ''), COALESCE(last_error, ''), created_at, updated_at
		FROM topology_sources
		ORDER BY root DESC, name, address
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.TopologySource{}
	for rows.Next() {
		item, err := scanTopologySource(rows, includeSecrets)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetTopologySourceForDiscovery(ctx context.Context, id string) (domain.TopologySource, error) {
	row := s.reader().QueryRowContext(ctx, `
		SELECT id, name, address, port, enabled, poll_interval_seconds, timeout_ms, retries, snmp_version,
			COALESCE(community, ''), COALESCE(username, ''), COALESCE(auth_protocol, ''), COALESCE(auth_passphrase, ''),
			COALESCE(privacy_protocol, ''), COALESCE(privacy_passphrase, ''), role, root,
			COALESCE(last_success_at, ''), COALESCE(last_error, ''), created_at, updated_at
		FROM topology_sources
		WHERE id = ?
	`, id)
	return scanTopologySource(row, true)
}

func (s *Store) SaveTopologySource(ctx context.Context, source domain.TopologySource) (domain.TopologySource, error) {
	source = normalizeTopologySource(source)
	if strings.TrimSpace(source.Name) == "" {
		return domain.TopologySource{}, errors.New("topology source name is required")
	}
	if strings.TrimSpace(source.Address) == "" {
		return domain.TopologySource{}, errors.New("topology source address is required")
	}
	if source.Port <= 0 || source.Port > 65535 {
		return domain.TopologySource{}, errors.New("topology source port must be between 1 and 65535")
	}
	if source.SNMPVersion != "v2c" && source.SNMPVersion != "v3" {
		return domain.TopologySource{}, errors.New("snmp version must be v2c or v3")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.TopologySource{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if source.ID == "" {
		source.ID = newID("tsrc")
		source.CreatedAt = now
	} else if source.CreatedAt.IsZero() {
		existing, err := s.getTopologySourceTx(ctx, tx, source.ID, true)
		if err == nil {
			source.CreatedAt = existing.CreatedAt
			source.LastSuccessAt = existing.LastSuccessAt
			source.LastError = existing.LastError
		} else if !errors.Is(err, sql.ErrNoRows) {
			return domain.TopologySource{}, err
		} else {
			source.CreatedAt = now
		}
	}
	source.UpdatedAt = now

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO topology_sources(
			id, name, address, port, enabled, poll_interval_seconds, timeout_ms, retries, snmp_version,
			community, username, auth_protocol, auth_passphrase, privacy_protocol, privacy_passphrase,
			role, root, last_success_at, last_error, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			address = excluded.address,
			port = excluded.port,
			enabled = excluded.enabled,
			poll_interval_seconds = excluded.poll_interval_seconds,
			timeout_ms = excluded.timeout_ms,
			retries = excluded.retries,
			snmp_version = excluded.snmp_version,
			community = excluded.community,
			username = excluded.username,
			auth_protocol = excluded.auth_protocol,
			auth_passphrase = excluded.auth_passphrase,
			privacy_protocol = excluded.privacy_protocol,
			privacy_passphrase = excluded.privacy_passphrase,
			role = excluded.role,
			root = excluded.root,
			updated_at = excluded.updated_at
	`, source.ID, source.Name, source.Address, source.Port, boolInt(source.Enabled), source.PollIntervalSeconds, source.TimeoutMS, source.Retries,
		source.SNMPVersion, nullableString(source.Community), nullableString(source.Username), nullableString(source.AuthProtocol),
		nullableString(source.AuthPassphrase), nullableString(source.PrivacyProtocol), nullableString(source.PrivacyPassphrase),
		source.Role, boolInt(source.Root), nullableTime(source.LastSuccessAt), nullableString(source.LastError),
		source.CreatedAt.Format(time.RFC3339Nano), source.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return domain.TopologySource{}, err
	}
	item, err := s.getTopologySourceTx(ctx, tx, source.ID, false)
	if err != nil {
		return domain.TopologySource{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.TopologySource{}, err
	}
	return item, nil
}

func (s *Store) DeleteTopologySource(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM topology_sources WHERE id = ?", id)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) UpdateTopologySourceStatus(ctx context.Context, id string, succeededAt time.Time, lastError string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE topology_sources SET last_success_at = CASE WHEN ? IS NOT NULL THEN ? ELSE last_success_at END, last_error = ?, updated_at = ? WHERE id = ?`, nullableTime(succeededAt), nullableTime(succeededAt), nullableString(lastError), nowString(), id)
	return err
}

func (s *Store) ReplaceTopologyObservations(ctx context.Context, sourceID string, obs domain.TopologySourceObservation) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, table := range []string{"topology_interfaces", "topology_lldp_links", "topology_mac_links"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE source_id = ?", sourceID); err != nil {
			return err
		}
	}
	now := time.Now().UTC()
	seenAt := obs.ObservedAt
	if seenAt.IsZero() {
		seenAt = now
	}
	for _, item := range obs.Interfaces {
		if item.ID == "" {
			item.ID = newID("tif")
		}
		if item.LastSeenAt.IsZero() {
			item.LastSeenAt = seenAt
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = now
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO topology_interfaces(id, source_id, if_index, if_name, if_description, if_alias, if_type, oper_status, speed_bps, last_seen_at, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, item.ID, sourceID, item.IfIndex, nullableString(item.IfName), nullableString(item.IfDescription), nullableString(item.IfAlias),
			item.IfType, nullableString(item.OperStatus), int64(item.SpeedBPS), item.LastSeenAt.Format(time.RFC3339Nano),
			item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	for _, item := range obs.LLDPLinks {
		if item.ID == "" {
			item.ID = newID("tll")
		}
		if item.LastSeenAt.IsZero() {
			item.LastSeenAt = seenAt
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = now
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO topology_lldp_links(
				id, source_id, local_chassis_id, local_system_name, local_port_id, local_port_name, local_port_description, local_if_index,
				remote_chassis_id, remote_system_name, remote_port_id, remote_port_description, remote_management_address,
				last_seen_at, created_at, updated_at
			) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, item.ID, sourceID, nullableString(item.LocalChassisID), nullableString(item.LocalSystemName), nullableString(item.LocalPortID),
			nullableString(item.LocalPortName), nullableString(item.LocalPortDescription), nullableInt(item.LocalIfIndex),
			nullableString(item.RemoteChassisID), nullableString(item.RemoteSystemName), nullableString(item.RemotePortID),
			nullableString(item.RemotePortDescription), nullableString(item.RemoteManagementAddress), item.LastSeenAt.Format(time.RFC3339Nano),
			item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	for _, item := range obs.MACLinks {
		if item.ID == "" {
			item.ID = newID("tml")
		}
		if item.LastSeenAt.IsZero() {
			item.LastSeenAt = seenAt
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = now
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO topology_mac_links(id, source_id, mac_address, vlan, bridge_port, if_index, if_name, if_description, status, last_seen_at, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, item.ID, sourceID, item.MACAddress, nullableInt(item.VLAN), nullableInt(item.BridgePort), nullableInt(item.IfIndex),
			nullableString(item.IfName), nullableString(item.IfDescription), nullableString(item.Status), item.LastSeenAt.Format(time.RFC3339Nano),
			item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListTopologyObservations(ctx context.Context) ([]domain.TopologySourceObservation, error) {
	sources, err := s.ListTopologySources(ctx)
	if err != nil {
		return nil, err
	}
	observations := make([]domain.TopologySourceObservation, 0, len(sources))
	index := map[string]int{}
	for _, source := range sources {
		index[source.ID] = len(observations)
		observations = append(observations, domain.TopologySourceObservation{SourceID: source.ID, Source: source})
	}
	if err := s.attachTopologyInterfaces(ctx, observations, index); err != nil {
		return nil, err
	}
	if err := s.attachTopologyLLDPLinks(ctx, observations, index); err != nil {
		return nil, err
	}
	if err := s.attachTopologyMACLinks(ctx, observations, index); err != nil {
		return nil, err
	}
	for i := range observations {
		observations[i].ObservedAt = maxObservationTime(observations[i])
	}
	return observations, nil
}

func (s *Store) getTopologySourceTx(ctx context.Context, tx *sql.Tx, id string, includeSecrets bool) (domain.TopologySource, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, name, address, port, enabled, poll_interval_seconds, timeout_ms, retries, snmp_version,
			COALESCE(community, ''), COALESCE(username, ''), COALESCE(auth_protocol, ''), COALESCE(auth_passphrase, ''),
			COALESCE(privacy_protocol, ''), COALESCE(privacy_passphrase, ''), role, root,
			COALESCE(last_success_at, ''), COALESCE(last_error, ''), created_at, updated_at
		FROM topology_sources
		WHERE id = ?
	`, id)
	return scanTopologySource(row, includeSecrets)
}

func scanTopologySource(scanner interface{ Scan(dest ...any) error }, includeSecrets bool) (domain.TopologySource, error) {
	var item domain.TopologySource
	var enabled, root int
	var community, authPassphrase, privacyPassphrase, lastSuccessAt, createdAt, updatedAt string
	err := scanner.Scan(&item.ID, &item.Name, &item.Address, &item.Port, &enabled, &item.PollIntervalSeconds, &item.TimeoutMS, &item.Retries,
		&item.SNMPVersion, &community, &item.Username, &item.AuthProtocol, &authPassphrase, &item.PrivacyProtocol, &privacyPassphrase,
		&item.Role, &root, &lastSuccessAt, &item.LastError, &createdAt, &updatedAt)
	if err != nil {
		return domain.TopologySource{}, err
	}
	item.Enabled = enabled == 1
	item.Root = root == 1
	item.LastSuccessAt = parseTime(lastSuccessAt)
	item.CreatedAt = parseTime(createdAt)
	item.UpdatedAt = parseTime(updatedAt)
	item.HasCommunity = strings.TrimSpace(community) != ""
	item.HasAuthPassphrase = strings.TrimSpace(authPassphrase) != ""
	item.HasPrivacyPassphrase = strings.TrimSpace(privacyPassphrase) != ""
	if includeSecrets {
		item.Community = community
		item.AuthPassphrase = authPassphrase
		item.PrivacyPassphrase = privacyPassphrase
	}
	return item, nil
}

func normalizeTopologySource(source domain.TopologySource) domain.TopologySource {
	source.Name = strings.TrimSpace(source.Name)
	source.Address = strings.TrimSpace(source.Address)
	source.SNMPVersion = strings.ToLower(strings.TrimSpace(source.SNMPVersion))
	if source.SNMPVersion == "" {
		source.SNMPVersion = "v2c"
	}
	if source.Port == 0 {
		source.Port = 161
	}
	if source.PollIntervalSeconds == 0 {
		source.PollIntervalSeconds = 300
	}
	if source.TimeoutMS == 0 {
		source.TimeoutMS = 1500
	}
	if source.Retries == 0 {
		source.Retries = 1
	}
	source.AuthProtocol = strings.ToLower(firstNonEmpty(strings.TrimSpace(source.AuthProtocol), "none"))
	source.PrivacyProtocol = strings.ToLower(firstNonEmpty(strings.TrimSpace(source.PrivacyProtocol), "none"))
	source.Role = strings.ToLower(firstNonEmpty(strings.TrimSpace(source.Role), "unknown"))
	source.Community = strings.TrimSpace(source.Community)
	source.Username = strings.TrimSpace(source.Username)
	source.AuthPassphrase = strings.TrimSpace(source.AuthPassphrase)
	source.PrivacyPassphrase = strings.TrimSpace(source.PrivacyPassphrase)
	return source
}

func (s *Store) attachTopologyInterfaces(ctx context.Context, observations []domain.TopologySourceObservation, index map[string]int) error {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, source_id, if_index, COALESCE(if_name, ''), COALESCE(if_description, ''), COALESCE(if_alias, ''), COALESCE(if_type, 0),
			COALESCE(oper_status, ''), COALESCE(speed_bps, 0), last_seen_at, created_at, updated_at
		FROM topology_interfaces
		ORDER BY source_id, if_index
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.TopologyInterfaceObservation
		var lastSeenAt, createdAt, updatedAt string
		var speed int64
		if err := rows.Scan(&item.ID, &item.SourceID, &item.IfIndex, &item.IfName, &item.IfDescription, &item.IfAlias, &item.IfType, &item.OperStatus, &speed, &lastSeenAt, &createdAt, &updatedAt); err != nil {
			return err
		}
		item.SpeedBPS = uint64(speed)
		item.LastSeenAt = parseTime(lastSeenAt)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		if obsIndex, ok := index[item.SourceID]; ok {
			observations[obsIndex].Interfaces = append(observations[obsIndex].Interfaces, item)
		}
	}
	return rows.Err()
}

func (s *Store) attachTopologyLLDPLinks(ctx context.Context, observations []domain.TopologySourceObservation, index map[string]int) error {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, source_id, COALESCE(local_chassis_id, ''), COALESCE(local_system_name, ''), COALESCE(local_port_id, ''), COALESCE(local_port_name, ''),
			COALESCE(local_port_description, ''), COALESCE(local_if_index, 0), COALESCE(remote_chassis_id, ''), COALESCE(remote_system_name, ''),
			COALESCE(remote_port_id, ''), COALESCE(remote_port_description, ''), COALESCE(remote_management_address, ''),
			last_seen_at, created_at, updated_at
		FROM topology_lldp_links
		ORDER BY source_id, local_port_id, remote_chassis_id, remote_port_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.TopologyLLDPLinkObservation
		var lastSeenAt, createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.SourceID, &item.LocalChassisID, &item.LocalSystemName, &item.LocalPortID, &item.LocalPortName,
			&item.LocalPortDescription, &item.LocalIfIndex, &item.RemoteChassisID, &item.RemoteSystemName, &item.RemotePortID,
			&item.RemotePortDescription, &item.RemoteManagementAddress, &lastSeenAt, &createdAt, &updatedAt); err != nil {
			return err
		}
		item.LastSeenAt = parseTime(lastSeenAt)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		if obsIndex, ok := index[item.SourceID]; ok {
			observations[obsIndex].LLDPLinks = append(observations[obsIndex].LLDPLinks, item)
		}
	}
	return rows.Err()
}

func (s *Store) attachTopologyMACLinks(ctx context.Context, observations []domain.TopologySourceObservation, index map[string]int) error {
	rows, err := s.reader().QueryContext(ctx, `
		SELECT id, source_id, mac_address, COALESCE(vlan, 0), COALESCE(bridge_port, 0), COALESCE(if_index, 0),
			COALESCE(if_name, ''), COALESCE(if_description, ''), COALESCE(status, ''), last_seen_at, created_at, updated_at
		FROM topology_mac_links
		ORDER BY source_id, if_index, mac_address, vlan
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.TopologyMACLinkObservation
		var lastSeenAt, createdAt, updatedAt string
		if err := rows.Scan(&item.ID, &item.SourceID, &item.MACAddress, &item.VLAN, &item.BridgePort, &item.IfIndex, &item.IfName, &item.IfDescription, &item.Status, &lastSeenAt, &createdAt, &updatedAt); err != nil {
			return err
		}
		item.LastSeenAt = parseTime(lastSeenAt)
		item.CreatedAt = parseTime(createdAt)
		item.UpdatedAt = parseTime(updatedAt)
		if obsIndex, ok := index[item.SourceID]; ok {
			observations[obsIndex].MACLinks = append(observations[obsIndex].MACLinks, item)
		}
	}
	return rows.Err()
}

func maxObservationTime(obs domain.TopologySourceObservation) time.Time {
	var max time.Time
	for _, item := range obs.Interfaces {
		if item.LastSeenAt.After(max) {
			max = item.LastSeenAt
		}
	}
	for _, item := range obs.LLDPLinks {
		if item.LastSeenAt.After(max) {
			max = item.LastSeenAt
		}
	}
	for _, item := range obs.MACLinks {
		if item.LastSeenAt.After(max) {
			max = item.LastSeenAt
		}
	}
	return max
}
