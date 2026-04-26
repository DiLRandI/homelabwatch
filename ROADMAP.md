# Roadmap

HomelabWatch is moving from "feature-rich MVP" to "polished open-source control
plane." The roadmap favors structural cleanup, calmer UX, and release
confidence over net-new surface area.

## Current State

- Latest tagged release: `v0.5.0`.
- `main` is past `v0.5.0` with Go module and frontend package refreshes.
- Core product surfaces are implemented; pending work is mostly launch
  readiness, release validation, and incremental domain/read-model hardening.

## Active Direction

### Cohesive Control Plane

This milestone is about making the product feel intentionally structured.

- dashboard narrowed to overview, favorites, status, and recent activity
- major management areas split into dedicated screens
- SSE refresh behavior targeted by feature instead of full-page reloads
- HTTP router split into route registration plus resource-specific handlers
- contributor docs, CI, issue templates, and release scaffolding added

## Next Milestones

### Launch Readiness

- browser smoke coverage for setup, bookmarks, discovery, and health flows
- published screenshots and short demo assets based on seeded example data
- more focused backend package boundaries inside `internal/app` and
  `internal/store/sqlite`
- upgrade and backup validation as part of release prep

### Domain Hardening

- reduce `Dashboard` and `SettingsView` as catch-all frontend contracts
- introduce narrower read models for screen-specific data
- continue moving screen-local UI sections closer to their owning feature areas
- make activity/event history a first-class product surface instead of only a
  dashboard subsection

## Non-Goals

- multi-container deployment
- replacing SQLite
- adding a heavy client state library without a concrete need
- public-internet first authentication for the built-in UI
- rewriting the product around YAML-defined runtime service definitions
