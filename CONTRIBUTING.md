# Contributing to HomelabWatch

HomelabWatch is an incremental product, not a rewrite-friendly playground.
Contributions should preserve the current deployment model, API compatibility,
and trusted-LAN operating model unless the change explicitly targets one of
those areas.

## Development Environment

- Go `1.25.x`
- Node.js `24.x`
- Corepack with the repository-pinned pnpm `11.9.0`
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
pnpm --dir web test --run
pnpm --dir web build
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
- `docs`: architecture, domain, operations, and launch-readiness documentation

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
- Frontend work: run `pnpm --dir web test --run` and `pnpm --dir web build`
- User-facing feature work: run both
- If a change affects HTTP routing, security, or token behavior, add or update
  tests under `internal/api/http`

## Dependency and Build-Script Security

- Run `corepack enable`, then install only with
  `pnpm install --frozen-lockfile`.
- Do not use npm, regenerate `package-lock.json`, bypass the seven-day release
  delay, or run lifecycle scripts from an untrusted branch.
- Direct dependencies use exact versions. Explain every dependency addition and
  include its maintainer, source repository, publication history, and lifecycle
  scripts in the PR.
- A dependency build script is denied unless it appears under `allowBuilds` in
  `pnpm-workspace.yaml`. Run `pnpm ignored-builds`, inspect the package and its
  published tarball, then use `pnpm approve-builds <package>` only after review.
  Record denied packages with `pnpm approve-builds !<package>`.
- Never run `pnpm approve-builds --all` and never enable
  `dangerouslyAllowAllBuilds`.
- Package-manager, lockfile, build-script policy, workflow, and Docker changes
  require a maintainer review. They must not be auto-merged.

See `docs/security/supply-chain-hardening.md` for the complete approval and
incident-response process.

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
