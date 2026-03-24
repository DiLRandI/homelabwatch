# Launch Readiness

This document is the release gate for turning HomelabWatch from a capable
pre-release into a trustworthy open-source control plane.

## Product Quality Bar

- The dashboard stays focused on favorites, fleet status, and recent activity.
- Major management tasks live in dedicated screens, not as dashboard overload.
- Setup, bookmark management, discovery review, health-check editing, and token
  creation all work without guesswork.
- Empty, loading, and error states are present for the main product surfaces.

## Code Quality Bar

- No single frontend file acts as a god-object for routing, data, and mutation
  logic.
- The HTTP layer is split into route registration, security middleware, and
  resource-specific handlers.
- New features can be added without widening `Dashboard` or `SettingsView`
  indefinitely.
- Incremental refactors preserve current deployment and API contracts unless a
  change is explicitly breaking.

## Documentation Bar

- `README.md` explains what the product is, how to run it, and how to trust it.
- `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`, and `CHANGELOG.md` exist and
  stay current.
- Architecture, domain, operations, and launch-readiness docs are present under
  `docs/`.
- Docker-focused docs stay aligned with runtime behavior.

## Operational Bar

- `go test ./...` passes.
- `cd web && npm run build` passes.
- Release config remains valid through `goreleaser check`.
- The Docker image still supports the single-container install path.
- Upgrades preserve setup state, bookmarks, discovery configuration, and health
  history on a persistent `/data` volume.

## OSS Readiness Bar

- CI runs on pushes and pull requests.
- Issue templates and a PR template are present.
- The code of conduct and security reporting path are documented.
- The roadmap explains active direction and non-goals.
- Release notes and the changelog call out migration or trust-boundary changes.

## Migration Guardrails

- Keep `/api/ui/v1/*`, `/api/v1/*`, and `/api/external/v1/*` stable unless a
  specific release is declared breaking.
- Treat SSE as an invalidation channel; do not break the event stream shape
  casually.
- Preserve setup semantics and bootstrap payloads.
- Preserve service-definition and health-check behavior unless the change is
  deliberate and documented.
- Refactor one layer at a time: UI ownership, HTTP routing, app orchestration,
  then store/read-model cleanup.

## Suggested Release Checklist

1. Run Go tests.
2. Build the frontend.
3. Validate GoReleaser config.
4. Review `CHANGELOG.md` and release notes.
5. Confirm docs for setup, security, and operations are up to date.
6. Smoke-check setup, dashboard, discovery, bookmarks, and health flows on a
   fresh `/data` volume.
