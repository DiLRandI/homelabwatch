# Feature: Status Pages

## Goal

Add read-only status pages that expose selected service health in a clean view
for household, team, or internal users without giving them access to the
management UI.

## User Value

- share current service availability without exposing settings
- quickly answer "is it down for everyone or just me?"
- group services by audience or environment
- publish maintenance or incident notes alongside live status

## Scope

Initial implementation should support:

- multiple status pages
- selected services per page
- public read-only page by slug
- optional page visibility toggle
- display:
  - overall page status
  - service status
  - latest check result
  - last checked time
  - recent incident/maintenance notes
- basic page branding:
  - title
  - description

Defer:

- custom domains
- anonymous subscriptions
- comments
- complex themes
- authenticated public users

## Domain Model

Add:

- `StatusPage`
  - `id`
  - `slug`
  - `title`
  - `description`
  - `enabled`
  - `createdAt`, `updatedAt`
- `StatusPageService`
  - `statusPageId`
  - `serviceId`
  - `sortOrder`
  - `displayName`
- `StatusPageAnnouncement`
  - `id`
  - `statusPageId`
  - `kind`: `info`, `maintenance`, `incident`, `resolved`
  - `title`
  - `message`
  - `startsAt`
  - `endsAt`
  - `createdAt`, `updatedAt`

Read models should include service status and latest check summary, but should
not expose management-only service details.

## Backend Work

- Add SQLite migrations for status pages, selected services, and announcements.
- Add app methods for page CRUD, service assignment, announcement CRUD, and
  public page read.
- Add a public read route that does not require trusted-console access:
  - `GET /status/{slug}`
  - `GET /api/public/v1/status-pages/{slug}`
- Add trusted UI routes:
  - `GET /api/ui/v1/status-pages`
  - `POST /api/ui/v1/status-pages`
  - `GET /api/ui/v1/status-pages/{id}`
  - `PATCH /api/ui/v1/status-pages/{id}`
  - `DELETE /api/ui/v1/status-pages/{id}`
  - `PUT /api/ui/v1/status-pages/{id}/services`
  - `POST /api/ui/v1/status-pages/{id}/announcements`
  - `PATCH /api/ui/v1/status-page-announcements/{id}`
  - `DELETE /api/ui/v1/status-page-announcements/{id}`
- Keep public API responses minimal and safe.
- When health checks update, publish a `status-page` SSE event for the admin UI.

## Frontend Work

- Add Status Pages route or Settings subsection for management.
- Add public status page screen rendered by the React app for `/status/{slug}`.
- Management UI should include:
  - page list
  - page editor
  - service picker and ordering
  - announcement editor
  - preview/open page action
- Public UI should include:
  - page title and description
  - overall status
  - service rows grouped by current status
  - announcements
  - last updated timestamp

## Routing Notes

The Go server serves the React app. Ensure `/status/{slug}` falls through to the
frontend while `/api/public/v1/status-pages/{slug}` returns JSON.

## Tests

Backend:

- public disabled page returns not found
- public enabled page returns only safe fields
- service assignment preserves order
- announcements filter by active time window
- status rolls up from selected services

Frontend:

- public status page renders healthy and unhealthy services
- disabled or missing page error state
- management CRUD flows

## Acceptance Criteria

- an operator can create a status page and select services
- a public read-only URL displays current health without management access
- disabling a page makes the public page unavailable
- announcements can be created and shown on the page
- service health changes are reflected after refresh or SSE-driven invalidation
