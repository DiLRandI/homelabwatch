# Feature: Expanded Service Definition Catalog

## Goal

Grow HomelabWatch's built-in service recognition so common homelab apps are
identified accurately with useful icons and health checks immediately after
Docker or LAN discovery.

## User Value

- fewer manual edits after discovery
- better names, icons, and URLs for common services
- more reliable default health checks
- easier contribution path for community service definitions

## Scope

Expand built-in definitions beyond the current small set. Target services:

- AdGuard Home
- Jellyfin
- Sonarr
- Radarr
- Lidarr
- Prowlarr
- qBittorrent
- Transmission
- Portainer
- Nginx Proxy Manager
- Traefik
- Vaultwarden
- Nextcloud
- Immich
- Paperless-ngx
- Gitea
- Forgejo
- Homebox
- Uptime Kuma
- Syncthing
- Unifi Network Application
- Proxmox, where LAN evidence is enough to identify it
- TrueNAS, where LAN evidence is enough to identify it

Defer:

- remote downloadable catalogs
- user-submitted catalog sync
- YAML runtime definitions

## Domain Model

Use the existing `ServiceDefinition`, `ServiceDefinitionMatcher`, and
`ServiceDefinitionCheckTemplate` entities.

Avoid schema changes unless a missing matcher type is clearly needed.

Potential additive matcher types:

- `http_header_value`
- `favicon_hash`
- `openapi_path`

Only add matcher types that are implemented in fingerprinting and tested.

## Backend Work

- Extend `internal/servicedefs/registry.go`.
- Prefer high-confidence matchers:
  - Docker image contains
  - Docker labels
  - known port plus HTTP title
  - mDNS service
  - stable HTTP header
- Avoid overmatching on port alone unless the service has a very distinctive
  port and low-risk fallback behavior.
- Add health check templates for known health endpoints:
  - API health endpoint when stable
  - web UI path when no health endpoint exists
  - TCP fallback where HTTP is unreliable
- Add tests for matching confidence and tie-breaking.
- Add tests that built-in definitions instantiate valid checks.

## Frontend Work

- Ensure the Definitions screen handles a larger built-in list comfortably.
- Add search/filter if the list becomes difficult to scan.
- Built-in definitions should remain read-only except for enabled/disabled
  state if that behavior already exists or is easy to add safely.

## Documentation Work

- Add a table to README or a dedicated docs page listing built-in definitions:
  - service name
  - match signals
  - default health check
- Add contribution guidance:
  - required matchers
  - expected health check behavior
  - test expectations

## Quality Rules

- Every built-in must have at least one high-confidence matcher.
- Every built-in must have a useful icon key or safe fallback.
- Every built-in must have at least one check template.
- Definitions must not convert unknown services into misleading names based on
  weak evidence.

## Tests

Backend:

- one positive match test per new service
- at least a few negative/ambiguous match tests
- check template instantiation for each definition
- sorting priority remains deterministic

Frontend:

- definitions list remains usable with the larger catalog
- search/filter behavior, if added

## Acceptance Criteria

- at least 20 common services are recognized by built-in definitions
- recognized services get better names, icons, and checks
- false-positive-prone definitions use multiple signals or lower confidence
- tests cover matching and check template validity
