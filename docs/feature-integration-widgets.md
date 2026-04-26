# Feature: Integration Widgets

## Goal

Add small, useful widgets for popular homelab services so HomelabWatch becomes
not only a discovery and monitoring surface, but a daily operational dashboard.

## User Value

- see key service stats without opening every app
- turn bookmarks into richer operational cards
- make discovered and accepted services feel more useful after promotion
- provide a clear extension point for future integrations

## Scope

Initial implementation should support:

- widget definitions tied to service definitions
- per-service widget enablement
- manual credential/API token config per service
- safe test connection action
- refresh interval per widget
- cached widget payloads stored in SQLite
- widgets rendered on service detail and optionally bookmark cards/dashboard

First widgets:

- Pi-hole or AdGuard Home:
  - queries today
  - blocked percentage
  - upstream status
- qBittorrent or Transmission:
  - active downloads
  - download/upload rate
  - paused/error count
- Jellyfin or Plex:
  - active sessions
  - library count if available
- Uptime Kuma:
  - monitor count by status if API access is configured

Defer:

- plugin marketplace
- arbitrary user-authored JavaScript
- write/control actions such as pause download or restart service
- iframe widgets

## Domain Model

Add:

- `IntegrationDefinition`
  - `id`
  - `serviceDefinitionKey`
  - `name`
  - `kind`
  - `enabled`
  - `defaultRefreshSeconds`
- `ServiceIntegration`
  - `id`
  - `serviceId`
  - `integrationDefinitionId`
  - `enabled`
  - `config`
  - `lastSuccessAt`
  - `lastError`
  - `createdAt`, `updatedAt`
- `WidgetSnapshot`
  - `id`
  - `serviceIntegrationId`
  - `status`
  - `payload`
  - `collectedAt`

Secrets in `config` must be redacted from read responses.

## Backend Work

- Add SQLite migrations for integration definitions, service integrations, and
  widget snapshots.
- Add `internal/integrations` package:
  - registry of integration clients
  - shared HTTP client helpers with timeout
  - per-integration fetchers
- Add app methods:
  - list available integrations for service
  - enable/configure integration
  - test integration
  - fetch cached widget snapshot
- Add scheduler job to refresh enabled widgets.
- Publish `integration-widget` SSE event when snapshots update.

Recommended API:

- `GET /api/ui/v1/services/{id}/integrations`
- `POST /api/ui/v1/services/{id}/integrations`
- `PATCH /api/ui/v1/service-integrations/{id}`
- `DELETE /api/ui/v1/service-integrations/{id}`
- `POST /api/ui/v1/service-integrations/{id}/test`
- `GET /api/ui/v1/service-integrations/{id}/snapshot`

## Frontend Work

- Add integration panel to service detail or Services screen.
- Provide:
  - available integration cards
  - credential/config form
  - test connection action
  - enabled/disabled state
  - last refresh and last error
- Render widget snapshots as compact cards:
  - fixed dimensions
  - clear loading/error/empty states
  - no layout shift when values update
- Optionally show enabled widgets on dashboard/bookmark cards after service
  detail behavior is stable.

## Security

- Never expose API tokens or passwords in read endpoints.
- Do not log secrets.
- Treat integration requests as outbound network calls controlled by trusted
  users only.
- Use conservative timeouts and avoid blocking the main scheduler on slow
  integrations.

## Tests

Backend:

- config redaction
- integration CRUD
- test connection success/failure with fake HTTP servers
- snapshot persistence
- scheduler refresh behavior
- timeout behavior

Frontend:

- configure integration
- test success/failure states
- widget snapshot rendering
- redacted config display

## Acceptance Criteria

- an operator can enable at least one integration for a recognized service
- credentials are stored and redacted correctly
- widget data refreshes in the background and appears in the UI
- failed widget refreshes show a useful error without breaking service health
- implementation provides a clear registry pattern for adding more widgets
