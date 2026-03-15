# HomelabWatch

HomelabWatch is a self-hosted homelab operations dashboard for service
discovery, Docker workload visibility, device tracking, bookmarks, and health
monitoring. It runs as a single container with SQLite persistence.

## Highlights

- Discovers services from Docker and LAN scan targets
- Tracks devices with MAC-aware identity where available
- Monitors services with HTTP, TCP, and ping checks
- Supports custom HTTP health endpoints instead of assuming `/`
- Includes built-in service definitions for common apps such as Grafana,
  Prometheus, Pi-hole, Home Assistant, and Plex
- Lets you test endpoints from the UI before saving them
- Streams live updates into the dashboard
- Provides scoped bearer tokens for external automation

## Quick Start

```bash
docker run --rm \
  -p 8080:8080 \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  deleema1/homelabwatch:latest
```

Then open:

```text
http://localhost:8080
```

On an empty `/data` volume, HomelabWatch starts with a setup wizard. No admin
token paste step is required in the browser UI.

## Recommended Linux Run

For better LAN discovery and ping checks on Linux:

```bash
docker run --rm \
  --network host \
  --cap-add NET_RAW \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  deleema1/homelabwatch:latest
```

## Health Checks

Health checks are configurable per service from the dashboard.

HTTP checks can define:

- protocol
- host
- port
- path
- method
- expected status range
- timeout
- interval

You can also run an on-demand endpoint test to see:

- HTTP status
- latency
- response size
- resolved URL
- matched service definition

## Service Definitions

HomelabWatch ships with a built-in service-definition registry and supports
custom definitions stored in SQLite through the UI and API.

Definitions drive:

- automatic fingerprinting
- default health endpoints
- ports
- icons
- health-check templates

Services that users customize manually are not overwritten by later automatic
reapply runs.

## Volumes And Runtime Expectations

- Persist `/data` if you want state to survive restarts.
- Mount `/var/run/docker.sock` if you want local Docker discovery.
- Without a persistent `/data` volume, a fresh container starts the setup
  wizard again.
- Database migrations run automatically at startup.

## Security Model

- The built-in web UI is open by design for trusted local or LAN use.
- Browser writes are limited to trusted networks, same-origin requests, and a
  server-issued CSRF token.
- External scripts and integrations should use bearer tokens created from
  `Settings > API access`.
- If you expose the app outside your local network, put it behind a reverse
  proxy or VPN and tighten the trusted CIDR list.

## Important Environment Variables

- `HOMELABWATCH_LISTEN_ADDR`
- `HOMELABWATCH_DATA_DIR`
- `HOMELABWATCH_DB_PATH`
- `HOMELABWATCH_STATIC_DIR`
- `HOMELABWATCH_SEED_CIDRS`
- `HOMELABWATCH_DEFAULT_SCAN_PORTS`
- `HOMELABWATCH_SEED_DOCKER_SOCKET`
- `HOMELABWATCH_TRUSTED_CIDRS`

Example:

```bash
docker run --rm \
  -p 8080:8080 \
  -e HOMELABWATCH_TRUSTED_CIDRS=127.0.0.1/32,192.168.1.0/24 \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  deleema1/homelabwatch:latest
```

## External API

- UI routes: `/api/ui/v1/*`
- External token-auth API: `/api/external/v1/*`
- Legacy `/api/v1/*` token-auth routes remain for compatibility

Example:

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/external/v1/dashboard
```

## Source

- Repository: `https://github.com/deleema/homelabwatch`
- Full project documentation lives in the repository `README.md`
