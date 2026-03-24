# Changelog

All notable changes to this project will be documented in this file.

The format follows Keep a Changelog and the project remains pre-`1.0`, so
breaking changes may still happen between minor releases.

## [Unreleased]

### Added

- Added a screen-based application shell with dedicated Dashboard, Bookmarks,
  Services, Health, Discovery, Devices, Definitions, and Settings views.
- Added targeted SSE refresh handling for discovery, services, definitions,
  health checks, and bootstrap changes.
- Added CI, issue templates, PR template, contributor guide, security policy,
  roadmap, and architecture/domain documentation.
- Added HTTP router tests covering bootstrap, trusted-console CSRF enforcement,
  and token-scope validation.

### Changed

- Split the browser bootstrap and data loading logic into narrower hooks instead
  of a single root coordinator.
- Split HTTP route registration from resource handlers so `router.go` is now a
  router/bootstrap module instead of a catch-all handler file.
- Reframed the dashboard around overview and activity while moving management
  surfaces into dedicated screens.

### Docs

- Updated the main README and Docker Hub documentation for the current
  control-plane structure and trust model.

## [0.1.0]

### Added

- Initial public pre-release with Docker and LAN discovery, bookmarks, device
  inventory, health monitoring, service definitions, setup wizard, and scoped
  external API tokens.
