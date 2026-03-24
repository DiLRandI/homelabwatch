# Contributing to HomelabWatch

HomelabWatch is an incremental product, not a rewrite-friendly playground.
Contributions should preserve the current deployment model, API compatibility,
and trusted-LAN operating model unless the change explicitly targets one of
those areas.

## Development Environment

- Go `1.25.x`
- Node.js `24.x`
- npm `10+`
- Docker is optional for local development and required for image validation

## Local Setup

1. Install frontend dependencies.

```bash
make web-install
```

2. Build the frontend bundle.

```bash
make web-build
```

3. Run the backend.

```bash
make run
```

4. Open `http://localhost:8080`.

If you want a complete verification pass before opening a PR:

```bash
go test ./...
cd web && npm run build
```

## Repo Guide

- `cmd/homelabwatch`: application entrypoint
- `internal/api/http`: HTTP routes, middleware, and resource handlers
- `internal/app`: orchestration and product behavior
- `internal/domain`: shared models and payload shapes
- `internal/discovery`: Docker and LAN discovery providers
- `internal/monitoring`: health-check execution
- `internal/store/sqlite`: SQLite persistence, queries, and migrations
- `web/src/app`: app shell and route-level screens
- `web/src/components`: shared UI primitives and screen sections
- `web/src/hooks`: screen/bootstrap data loading and SSE integration
- `docs`: architecture, domain, and operations documentation

## Change Expectations

- Prefer incremental refactors over broad rewrites.
- Keep backend changes flowing through `domain -> store -> app -> api`.
- Preserve compatibility for `/api/v1/*` and `/api/external/v1/*` unless a task
  explicitly allows a breaking change.
- Add a new migration for schema changes. Do not edit older migrations.
- Keep Go JSON tags and frontend payload names aligned.
- Reuse existing screen sections and health UI where possible before creating a
  parallel pattern.
- When service-check or service-definition behavior changes, update both
  `README.md` and `DOCKERHUB.md`.

## Testing Expectations

- Backend work: run `go test ./...`
- Frontend work: run `cd web && npm run build`
- User-facing feature work: run both
- If a change affects HTTP routing, security, or token behavior, add or update
  tests under `internal/api/http`

## Pull Requests

- Keep PRs scoped to one feature area or one cross-cutting cleanup.
- Explain behavioral impact, migration risk, and any API changes.
- Include screenshots or short screen recordings for visible UI changes when
  practical.
- Call out follow-up work explicitly instead of hiding it inside TODOs.

## Release Notes and Changelog

HomelabWatch keeps a human-maintained `CHANGELOG.md`.

- Add user-visible changes to the `Unreleased` section.
- Note migrations, operational changes, and compatibility risks.
- Keep entries short and factual.
