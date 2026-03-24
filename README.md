# HomelabWatch

HomelabWatch is a self-hosted homelab discovery and monitoring platform that
runs as a single Go service with a React frontend and SQLite persistence. It
discovers devices and services, tracks bookmarks, fingerprints common apps, and
monitors health from one container.

## Features

- Dashboard for services, containers, devices, bookmarks, worker state, and recent events
- Service discovery from Docker endpoints and seeded LAN scan targets
- Device tracking keyed by MAC address when available, with fallback identity
- Health monitoring with HTTP, TCP, and ping checks
- Custom HTTP health endpoints with editable path, method, status range, timeout, and interval
- Endpoint test workflow that returns status, latency, response size, and matched service definition
- Built-in service-definition registry for common apps such as Pi-hole, Grafana, Prometheus, Home Assistant, and Plex
- Automatic service fingerprinting from ports, container image hints, mDNS metadata, HTTP headers, page titles, and body signatures
- Bookmark management for manual links and external services
- First-run setup wizard instead of manual bootstrap secrets in the browser
- Managed external API tokens with revocation from the settings surface
- SSE updates from the backend to the frontend
- Single-container deployment with SQLite persistence

## Stack

- Backend: Go, REST API, SQLite, in-process background workers
- Frontend: React, Vite, Tailwind CSS
- Packaging: multi-stage Docker build

## Quick Start

### Common Make targets

```bash
make help
make web-install
make web-build
make test
make build
make run
make docker-build
make release-check
make release-snapshot
```

### Run with Docker

Build the image:

```bash
make docker-build
```

Run it:

```bash
docker run --rm \
  -p 8080:8080 \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  homelabwatch:local
```

On a fresh `/data` volume, HomelabWatch starts with a setup wizard in the
browser. The local UI stays open for trusted LAN clients, and external
automation tokens are created later from `Settings > API access`.

For LAN discovery and ping checks on Linux, host networking and raw socket
access are usually required:

```bash
docker run --rm \
  --network host \
  --cap-add NET_RAW \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  homelabwatch:local
```

### Run locally

Install frontend dependencies and build assets:

```bash
make web-install
make web-build
```

Start the backend:

```bash
make run
```

Then open `http://localhost:8080`.

`npm run dev` is available for frontend-only work, but the app expects
same-origin API requests, so a dev proxy is needed if you want to point Vite at
the Go API.

## Health Monitoring

HomelabWatch no longer assumes that every HTTP service is healthy at `/`.
Health checks are now modeled explicitly and can be edited per service.

Each HTTP check can define:

- `protocol`
- `host`
- `port`
- `path`
- `method`
- `expectedStatusMin` and `expectedStatusMax`
- `timeoutSeconds`
- `intervalSeconds`

The dashboard exposes this through `Edit health` on each service card. Users
can:

- create multiple checks per service
- choose HTTP, TCP, or ping checks
- test a candidate endpoint before saving
- switch a service from auto-managed checks to custom checks by editing it

The endpoint tester returns:

- resolved URL
- status
- HTTP status code
- latency
- response size
- matched service definition, when one is recognized

## Service Definitions And Fingerprinting

HomelabWatch ships with a built-in registry of known services and default
health-check templates. The current built-in set includes:

- Pi-hole
- Grafana
- Prometheus
- Home Assistant
- Plex

Definitions are used for:

- default ports
- default health paths
- icon selection
- automatic health-check provisioning
- fingerprint scoring

Fingerprinting uses a mix of:

- exposed port hints
- container image names and labels
- mDNS metadata
- HTTP response headers
- page titles
- body substrings

Unknown services do not get aggressive background path probing. Instead:

- known matches receive definition-driven HTTP checks
- unmatched services fall back to TCP or ping checks
- smart path discovery happens when the user runs `Test endpoint` with a blank HTTP path

Custom service definitions are supported today through the dashboard and API and
are stored in SQLite. YAML-backed custom definition loading is not part of the
current runtime yet.

## Discovery And Promotion Behavior

- Discovered services are fingerprinted in the background.
- Accepted discovered services preserve their managed health checks and recent
  health history when promoted into first-class services and bookmarks.
- Service-definition reapply only updates services still in `auto` mode.
- User-edited services stay in `custom` mode and are not overwritten by later
  definition refreshes.

## Configuration

Configuration can come from `config.yaml` / `config.yml` or environment
variables.

Example config: [`config.example.yaml`](config.example.yaml)

Important environment variables:

- `HOMELABWATCH_LISTEN_ADDR`
- `HOMELABWATCH_DATA_DIR`
- `HOMELABWATCH_DB_PATH`
- `HOMELABWATCH_STATIC_DIR`
- `HOMELABWATCH_CONFIG`
- `HOMELABWATCH_SEED_CIDRS`
- `HOMELABWATCH_DEFAULT_SCAN_PORTS`
- `HOMELABWATCH_SEED_DOCKER_SOCKET`
- `HOMELABWATCH_TRUSTED_CIDRS`

Example trust boundary override:

```bash
HOMELABWATCH_TRUSTED_CIDRS=127.0.0.1/32,192.168.1.0/24
```

## API Notes

- The built-in browser UI uses `/api/ui/v1/*`
- External automation uses bearer tokens against `/api/external/v1/*`
- Legacy `/api/v1/*` token-auth endpoints remain for compatibility
- Live updates for the browser UI are streamed from `GET /api/ui/v1/events`

Health-related endpoints:

- `GET /api/ui/v1/services/{id}/checks`
- `POST /api/ui/v1/services/{id}/checks`
- `PATCH /api/ui/v1/checks/{id}`
- `DELETE /api/ui/v1/checks/{id}`
- `POST /api/ui/v1/services/{id}/checks/test`

Service-definition endpoints:

- `GET /api/ui/v1/service-definitions`
- `POST /api/ui/v1/service-definitions`
- `PATCH /api/ui/v1/service-definitions/{id}`
- `DELETE /api/ui/v1/service-definitions/{id}`
- `POST /api/ui/v1/service-definitions/{id}/reapply`

Common external API flow:

1. open the UI and finish setup
2. go to `Settings > API access`
3. create a read or write token
4. call `/api/external/v1/*` with `Authorization: Bearer <token>`

## Frontend Structure

```text
web/src/
  App.jsx
  components/
    bootstrap/
    dashboard/
    discovery/
    forms/
    health/
    ui/
  hooks/
  lib/
  main.jsx
```

Current responsibilities:

- `App.jsx`: root composition only
- `components/bootstrap`: first-run setup wizard
- `components/dashboard`: dashboard sections and layout
- `components/discovery`: discovery review flows
- `components/forms`: form-specific state and submit handling
- `components/health`: health modal, tester, and badges
- `components/ui`: shared presentation primitives
- `hooks/useHomelabwatchApp.js`: app state, actions, and data loading
- `hooks/useServerEvents.js`: SSE subscription lifecycle
- `lib/api.js`: API requests
- `lib/forms.js`: form defaults and request-shaping helpers
- `lib/format.js`: display helpers

## Backend Layout

Key directories:

- `cmd/homelabwatch`: application entrypoint
- `internal/api`: HTTP and SSE handlers
- `internal/app`: orchestration layer and high-level behavior
- `internal/discovery`: Docker and LAN discovery providers
- `internal/domain`: shared domain models and request/response types
- `internal/monitoring`: health-check execution
- `internal/servicedefs`: built-in service-definition registry and helpers
- `internal/store/sqlite`: persistence and migrations
- `migrations`: schema bootstrap and upgrades

## Security Model

- The browser UI has no sign-in screen by product design.
- The built-in UI is intended for trusted local or LAN use.
- UI reads are open, but UI writes require all of:
  - a client IP inside `HOMELABWATCH_TRUSTED_CIDRS`
  - same-origin browser requests
  - the console CSRF token issued by the server
- External clients should use managed bearer tokens against the external API.
- Legacy `admin_token_hash` installs remain compatible on the external API
  during migration, but the browser no longer asks for that token.

## Deployment Notes

- Mount `/data` if you want persistent SQLite state across container restarts.
- Mount `/var/run/docker.sock` if you want automatic local Docker discovery.
- Use `--network host --cap-add NET_RAW` on Linux if you want the best LAN
  discovery and ping behavior.
- If you expose the UI beyond your local network, put it behind a reverse
  proxy or VPN and tighten `HOMELABWATCH_TRUSTED_CIDRS`.
- Database migrations run automatically on startup.
- A Docker Hub specific README is available in [`DOCKERHUB.md`](DOCKERHUB.md).

## Verification

Useful checks:

```bash
make test
make web-build
make docker-build
```

For release validation:

```bash
make release-check
make release-snapshot
```

`make release-snapshot` expects Docker, Buildx, and GoReleaser to be installed
locally because it validates the multi-platform release packaging too.

## Release Automation

GitHub releases are automated with GitHub Actions and GoReleaser.

- Workflow trigger: publishing a GitHub release
- Binary assets: Linux `amd64`, `arm64`, `armv6`, and `armv7`
- Docker images: `linux/amd64`, `linux/arm64`, `linux/arm/v6`, and `linux/arm/v7`
- Docker registry: Docker Hub

When a release is published, the workflow:

1. builds the React frontend
2. runs `go test ./...`
3. builds release archives for each Linux target
4. uploads the archives and `checksums.txt` to the GitHub release
5. builds and pushes multi-platform Docker images to Docker Hub

Stable releases publish Docker tags:

- `vX.Y.Z`
- `X.Y`
- `X`
- `latest`

Prereleases only publish the exact version tag, for example `v0.2.0-rc1`.
