# homelabwatch

Homelabwatch is a lightweight self-hosted homelab dashboard built as a single
Go application with a React frontend. It discovers services, tracks devices,
runs health checks, and serves the UI and API from one container.

## Features

- Dashboard for services, containers, devices, bookmarks, worker state, and recent events
- Service discovery from Docker endpoints and seeded LAN scan targets
- Device tracking keyed by MAC address when available, with fallback identity
- Health monitoring with HTTP, TCP, and ping checks
- Bookmark management for manual links and external services
- First-run setup wizard instead of manual bootstrap secrets in the browser
- Managed external API tokens with revocation from the settings surface
- Single-container deployment with SQLite persistence
- SSE updates from the backend to the frontend

## Stack

- Backend: Go, REST API, SQLite, in-process background workers
- Frontend: React, Vite, Tailwind CSS
- Packaging: multi-stage Docker build

## Quick Start

### Common Make targets

```bash
make help
make test
make web-build
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

On a fresh `/data` volume, homelabwatch starts with a 3-step setup wizard in
the browser. The local web UI stays open for trusted LAN clients, and external
automation tokens are created later from `Settings > API access` instead of
being pasted into the dashboard.

First-run flow:

1. open `http://localhost:8080`
2. name the appliance and confirm you are on a trusted local/LAN client
3. choose discovery defaults and launch the initial discovery run
4. create external bearer tokens later from `Settings > API access` if you
   need automation or third-party integrations

For LAN discovery and ping checks on Linux, host networking and raw socket
access are typically required:

```bash
docker run --rm \
  --network host \
  --cap-add NET_RAW \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  homelabwatch:local
```

### Run locally

Build the frontend:

```bash
make web-install
make web-build
```

Start the backend from the repo root:

```bash
make run
```

Then open `http://localhost:8080`.

`npm run dev` is available for frontend-only work, but the current app expects
API requests on the same origin, so a dev proxy is needed if you want to use
the Vite server against the Go API.

## Deployment Notes

- Mount `/data` if you want persistent SQLite state across container restarts.
- Mount `/var/run/docker.sock` if you want automatic local Docker discovery.
- Use `--network host --cap-add NET_RAW` on Linux if you want the best LAN
  discovery and ping behavior.
- If you expose the UI beyond your local network, put it behind a reverse
  proxy or VPN and tighten `HOMELABWATCH_TRUSTED_CIDRS`.
- A Docker Hub specific README is available in [`DOCKERHUB.md`](DOCKERHUB.md).

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

## Frontend Structure

The frontend is intentionally split into small modules instead of a single
large app file:

```text
web/src/
  App.jsx
  components/
    bootstrap/
    dashboard/
    forms/
    ui/
  hooks/
  lib/
  main.jsx
```

Current responsibilities:

- `App.jsx`: root composition only
- `components/bootstrap`: first-run setup wizard
- `components/dashboard`: dashboard sections and layout
- `components/forms`: form-specific state and submit handling
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
- `internal/app`: orchestration layer
- `internal/discovery`: Docker and LAN discovery providers
- `internal/monitoring`: health-check execution
- `internal/store/sqlite`: persistence and migrations
- `migrations`: schema bootstrap and upgrades

## Security Model

- The browser UI has no sign-in screen by product design.
- The built-in UI is intended for trusted local or LAN use.
- UI reads are open, but UI writes require all of:
  - a client IP inside `HOMELABWATCH_TRUSTED_CIDRS`
  - same-origin browser requests
  - the console CSRF token issued by the server
- External clients should not use the UI endpoints. They should use managed
  bearer tokens against the external API.
- Legacy `admin_token_hash` installs remain compatible on the external API
  during migration, but the browser no longer asks for that token.

## API Notes

- The built-in browser UI uses `/api/ui/v1/*`
- UI writes are restricted to trusted networks and same-origin browser requests
- External automation uses bearer tokens against `/api/external/v1/*`
- Legacy `/api/v1/*` token-auth endpoints remain for compatibility
- Live updates for the browser UI are streamed from `GET /api/ui/v1/events`

Common API flow:

1. open the UI and finish setup
2. go to `Settings > API access`
3. create a read or write token
4. call `/api/external/v1/*` with `Authorization: Bearer <token>`

## Verification

Useful checks:

```bash
make test
make web-build
make docker-build
```

For release config validation:

```bash
make release-check
make release-snapshot
```

`make release-snapshot` expects Docker, Buildx, and GoReleaser to be installed
locally because it also validates the multi-platform release packaging.

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

Prereleases only publish the exact version tag, for example
`v0.2.0-rc1`.

### Required GitHub configuration

Set these on the GitHub repository before publishing releases:

- repository variable: `DOCKERHUB_USERNAME`
- repository secret: `DOCKERHUB_TOKEN`

The workflow file is [`release.yml`](.github/workflows/release.yml) and the
release definition is [`.goreleaser.yaml`](.goreleaser.yaml).

### Release assets

Each release uploads:

- `homelabwatch_<version>_linux_amd64.tar.gz`
- `homelabwatch_<version>_linux_arm64.tar.gz`
- `homelabwatch_<version>_linux_armv6.tar.gz`
- `homelabwatch_<version>_linux_armv7.tar.gz`
- `checksums.txt`

Each archive contains the `homelabwatch` binary plus `README.md`, `LICENSE`,
and `config.example.yaml`.

### Tag and release format

Use semantic version tags with a leading `v`, for example:

```text
v0.1.0
v0.1.1
v0.2.0-rc1
```

The workflow runs on the GitHub `release.published` event, so the usual
release flow is:

1. create a Git tag like `v0.1.0`
2. create or publish the matching GitHub release
3. let the workflow build the binaries and Docker images

The release image uses [`Dockerfile.release`](Dockerfile.release), while the
existing [`Dockerfile`](Dockerfile) remains the local and development
container build.

## License

MIT. See [`LICENSE`](LICENSE).
