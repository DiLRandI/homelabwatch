# Feature: Notification Engine

## Goal

Add first-class notifications for important operational changes so operators do
not need to keep the HomelabWatch UI open to know when something needs
attention.

## User Value

- know when a service/check goes down or recovers
- know when an important device disappears or reappears
- know when discovery finds a new device or service
- know when a worker repeatedly fails
- reduce alert fatigue with per-rule enablement and test delivery

## Scope

Initial implementation should support:

- notification channels:
  - generic webhook
  - ntfy-compatible HTTP publish
  - SMTP email, if the implementation remains small
- notification rules:
  - service health changes: healthy, degraded, unhealthy
  - check failures and recoveries
  - new discovered service
  - new device
  - device offline/reconnected, when the device has alerting enabled
  - worker failure
- per-rule enabled state
- per-channel enabled state
- test notification action
- delivery history with success/failure message

Defer:

- escalation policies
- multi-user routing
- incident ownership
- complex templating language
- large integration catalog

## Domain Model

Add narrow entities instead of widening `SettingsView` indefinitely:

- `NotificationChannel`
  - `id`
  - `name`
  - `type`: `webhook`, `ntfy`, `smtp`
  - `enabled`
  - `config`
  - `createdAt`, `updatedAt`
- `NotificationRule`
  - `id`
  - `name`
  - `eventType`
  - `enabled`
  - `channelIds`
  - `filters`
  - `createdAt`, `updatedAt`
- `NotificationDelivery`
  - `id`
  - `ruleId`
  - `channelId`
  - `eventType`
  - `status`: `pending`, `sent`, `failed`
  - `message`
  - `attemptedAt`

Keep secrets out of API responses. Store sensitive channel config in SQLite,
but redact it in read models.

## Backend Work

- Add SQLite migrations for channels, rules, and delivery history.
- Add `internal/notifications` package for delivery clients.
- Add app-layer orchestration that subscribes to domain events and evaluates
  notification rules.
- Reuse existing `domain.EventEnvelope` where possible. Do not make SSE the
  source of truth; publish to notifications from the app/event bus side.
- Add retry-light behavior: one immediate attempt, record failure, no long
  background retry queue in the first version.
- Add explicit test-send command that sends a sample message through one
  channel.

Recommended API surface:

- `GET /api/ui/v1/notifications/channels`
- `POST /api/ui/v1/notifications/channels`
- `PATCH /api/ui/v1/notifications/channels/{id}`
- `DELETE /api/ui/v1/notifications/channels/{id}`
- `POST /api/ui/v1/notifications/channels/{id}/test`
- `GET /api/ui/v1/notifications/rules`
- `POST /api/ui/v1/notifications/rules`
- `PATCH /api/ui/v1/notifications/rules/{id}`
- `DELETE /api/ui/v1/notifications/rules/{id}`
- `GET /api/ui/v1/notifications/deliveries`

Mirror read/write routes under `/api/external/v1/*` with existing token scope
rules.

## Frontend Work

- Add a Notifications area under Settings or a dedicated route if the UI becomes
  too dense.
- Provide:
  - channel list
  - channel editor
  - rule list
  - rule editor
  - delivery history
  - test notification button
- Show redacted secrets as configured, not as readable values.
- Surface delivery errors clearly without blocking unrelated settings.

## Events

Add an SSE event type:

- `notification`

Use it to refresh notification lists and delivery history after rule/channel
changes or test sends.

## Tests

Backend:

- channel CRUD validation
- secret redaction
- rule matching
- webhook delivery success/failure
- ntfy delivery request shape
- delivery history persistence
- trusted-console and external-token authorization

Frontend:

- empty state
- create/edit channel
- create/edit rule
- test-send success and failure states

## Acceptance Criteria

- an operator can create a webhook or ntfy channel from the UI
- an operator can send a test notification
- a service health transition can trigger a notification
- delivery success/failure is visible in the UI
- notification secrets are never returned in plaintext from read endpoints
- existing setup, discovery, bookmarks, health, and token APIs remain stable
