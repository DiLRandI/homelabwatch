# Changelog

All notable changes to this project will be documented in this file.

The format follows Keep a Changelog and the project remains pre-`1.0`, so
breaking changes may still happen between minor releases.

## [Unreleased]

### Changed

- Refreshed Go module dependencies.
- Refreshed frontend dependencies, including React, Vite, Tailwind CSS, and
  PostCSS.

### Docs

- Backfilled the changelog through the `v0.5.0` release.
- Updated project docs to match the current screen names, release state, and
  dependency-refresh status on `main`.

## [0.5.0] - 2026-04-03

### Added

- Added structured JSON logging through `slog`, controlled by `LOG_LEVEL`.
- Added separate service open URL and health-target defaults so health checks
  can monitor a different endpoint than the URL operators open.

### Changed

- Cleaned up service URL handling and health endpoint behavior.
- Refactored bookmark filtering logic and fixed bookmark page clipping.
- Updated README, Docker Hub, architecture, and operations documentation for
  the current control-plane structure and trust model.
- Replaced generated SVG architecture diagrams with Mermaid blocks in the
  GitHub-rendered docs.

## [0.4.0] - 2026-03-29

### Added

- Added dark-mode support.
- Added separate SQLite reader and writer connections.

### Changed

- Refactored backend internals while preserving the single-process SQLite
  deployment model.
- Simplified UI screens and fixed text/layout issues.
- Updated Go and frontend dependencies.
- Fixed the cleanup script.

## [0.3.0] - 2026-03-25

### Added

- Added a screen-based application shell with dedicated Dashboard, Bookmarks,
  Services, Health, Discovery, Devices, Definitions, and Settings views.
- Added a bookmarks-first navigation workspace.
- Added expanded discovery UI and network-discovery workflows.
- Added health-check management features.
- Added targeted SSE refresh handling for discovery, services, definitions,
  health checks, and bootstrap changes.
- Added CI, issue templates, PR template, contributor guide, security policy,
  roadmap, and architecture/domain documentation.
- Added HTTP router tests covering bootstrap, trusted-console CSRF enforcement,
  and token-scope validation.

### Changed

- Continued the backend and UI revamp toward dedicated product surfaces.
- Split the browser bootstrap and data loading logic into narrower hooks instead
  of a single root coordinator.
- Split HTTP route registration from resource handlers so `router.go` is now a
  router/bootstrap module instead of a catch-all handler file.
- Updated route structure and CI workflow behavior.
- Reframed the dashboard around overview and activity while moving management
  surfaces into dedicated screens.

## [0.2.0] - 2026-03-15

### Added

- Added redesigned layout components and UI primitives.
- Added dashboard sections for containers and API token management.
- Added auto-bootstrap and initialization-flow improvements.

### Changed

- Upgraded React/frontend dependencies and refactored Tailwind CSS
  configuration.

## [0.1.0] - 2026-03-15

### Added

- Initial public pre-release with Docker and LAN discovery, bookmarks, device
  inventory, health monitoring, service definitions, setup wizard, scoped
  external API tokens, and release automation.
