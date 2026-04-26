# Feature: Device Change Workflows

## Goal

Turn network inventory changes into actionable workflows so HomelabWatch can
help operators understand what changed on the LAN, not only what exists now.

## User Value

- see new devices immediately after scans
- detect device disappearance/reconnection
- notice IP, MAC, hostname, and port changes
- acknowledge known changes to reduce repeated noise
- mark critical devices for stronger visibility and notifications

## Scope

Initial implementation should support:

- device-level alert settings:
  - alert on new device
  - alert on offline/reconnected
  - alert on IP change
  - alert on port change
  - critical device flag
- durable device change events
- acknowledge/ignore actions for change events
- filters for unresolved, acknowledged, ignored, critical, and event type
- integration point for the notification engine

Defer:

- automated remediation
- network topology inference
- vendor fingerprint database
- full rule builder

## Domain Model

Add:

- `DeviceAlertPolicy`
  - `deviceId`
  - `alertNew`
  - `alertOffline`
  - `alertIPChange`
  - `alertPortChange`
  - `critical`
  - `updatedAt`
- `DeviceChangeEvent`
  - `id`
  - `deviceId`
  - `eventType`: `new_device`, `offline`, `reconnected`, `ip_changed`,
    `hostname_changed`, `port_opened`, `port_closed`
  - `severity`: `info`, `warning`, `critical`
  - `summary`
  - `details`
  - `state`: `open`, `acknowledged`, `ignored`
  - `createdAt`
  - `resolvedAt`
  - `updatedAt`

Reuse the existing activity/event log where practical, but create a specific
device-change read model so the UI can filter and manage events without parsing
generic messages.

## Backend Work

- Add migrations for alert policies and device change events.
- During discovery persistence, compare previous and new device observations:
  - first seen device
  - address added/removed
  - primary IP change
  - hostname change
  - port opened/closed
  - last seen gap crossing offline threshold
- Add an offline detection job or integrate it into cleanup/scan jobs.
- Add app methods to acknowledge, ignore, and restore events.
- Publish domain events for device changes so notifications can consume them.

Recommended API:

- `GET /api/ui/v1/device-changes`
- `PATCH /api/ui/v1/device-changes/{id}`
- `POST /api/ui/v1/device-changes/{id}/acknowledge`
- `POST /api/ui/v1/device-changes/{id}/ignore`
- `POST /api/ui/v1/device-changes/{id}/restore`
- `GET /api/ui/v1/devices/{id}/alert-policy`
- `PATCH /api/ui/v1/devices/{id}/alert-policy`

Mirror appropriate routes under `/api/external/v1/*`.

## Frontend Work

- Add a Device Changes section under Devices or a dedicated screen if volume
  warrants it.
- Add per-device alert policy controls to the device detail view.
- Provide:
  - event table/list
  - severity badges
  - filters
  - acknowledge/ignore actions
  - link to affected device and related services
- Surface critical unresolved events on the dashboard without overloading it.

## Interaction With Notifications

The notification engine should be able to subscribe to `device-change` events.
Do not hard-code notification delivery into discovery providers.

## Tests

Backend:

- new device event creation
- IP change event creation
- port opened/closed event creation
- duplicate suppression across repeated scans
- acknowledge/ignore state transitions
- alert policy persistence

Frontend:

- event filters
- acknowledge/ignore actions
- device alert policy editing

## Acceptance Criteria

- scanning a previously unknown device creates an open change event
- changing observed IP or ports creates a clear event once
- operators can acknowledge or ignore events
- critical devices are visually distinct
- device-change events are available to notification rules
