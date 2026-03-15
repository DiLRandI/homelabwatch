# homelabwatch

Homelabwatch is a lightweight self-hosted homelab dashboard built as a single
Go application with a React frontend. It discovers services, tracks devices,
runs health checks, and serves the UI and API from one container.

## Features

- Dashboard for services, devices, bookmarks, worker state, and recent events
- Service discovery from Docker endpoints and seeded LAN scan targets
- Device tracking keyed by MAC address when available, with fallback identity
- Health monitoring with HTTP, TCP, and ping checks
- Bookmark management for manual links and external services
- Single-container deployment with SQLite persistence
- SSE updates from the backend to the frontend

## Stack

- Backend: Go, REST API, SQLite, in-process background workers
- Frontend: React, Vite, Tailwind CSS
- Packaging: multi-stage Docker build

## Quick Start

### Run with Docker

Build the image:

```bash
docker build -t homelabwatch:local .
```

Run it:

```bash
docker run --rm \
  -p 8080:8080 \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  homelabwatch:local
```

Open `http://localhost:8080` and complete bootstrap in the UI.

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
cd web
npm install
npm run build
```

Start the backend from the repo root:

```bash
go run ./cmd/homelabwatch
```

Then open `http://localhost:8080`.

`npm run dev` is available for frontend-only work, but the current app expects
API requests on the same origin, so a dev proxy is needed if you want to use
the Vite server against the Go API.

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
- `components/bootstrap`: bootstrap screen
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

## API Notes

- Read endpoints are available without a token
- Write endpoints require `X-Admin-Token` after bootstrap
- Bootstrap is completed through `POST /api/v1/bootstrap/init`
- Live updates are streamed from `GET /api/v1/events`

## Verification

Useful checks:

```bash
go test ./...
cd web && npm run build
docker build -t homelabwatch:local .
```

## License

MIT. See [`LICENSE`](LICENSE).
