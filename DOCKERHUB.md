# HomelabWatch

HomelabWatch is a self-hosted homelab discovery and monitoring control plane.
It runs as a single container with SQLite persistence and gives you a focused
dashboard plus dedicated screens for bookmarks, services, health, discovery,
devices, service definitions, and settings.

## Highlights

- calm dashboard for favorites, status, and recent activity
- Docker and LAN discovery with promotion into managed services and bookmarks
- device tracking with MAC-aware identity where available
- HTTP, TCP, and ping health checks with editable HTTP paths
- separate open URL and health URL targeting for services that need a
  different monitoring endpoint
- endpoint testing before saving a check
- built-in and custom service definitions for fingerprinting and managed checks
- scoped bearer tokens for external automation
- setup wizard on a fresh data volume
- live UI updates over SSE

## Quick Start

```bash
docker run --rm \
  -p 8080:8080 \
  -v "$(pwd)/data:/data" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  deleema1/homelabwatch:latest
```

Open `http://localhost:8080`.

On an empty `/data` volume, HomelabWatch starts with a setup wizard. No admin
token paste step is required in the browser UI.

The repository also includes a `docker-compose.example.yml` file for a simple
compose-based install.

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

## Runtime Expectations

- Persist `/data` if you want state to survive restarts.
- Mount `/var/run/docker.sock` if you want local Docker discovery.
- Mount a config file and set `HOMELABWATCH_CONFIG` if you want the container
  to load YAML configuration instead of relying only on env vars.
- Without persistent `/data`, a fresh container starts the setup wizard again.
- Database migrations run automatically at startup.

## Security Model

- The built-in web UI is open by design for trusted local or LAN use.
- Browser writes require trusted-network access, same-origin requests, and a
  server-issued CSRF token.
- External scripts and integrations should use bearer tokens created from
  `Settings > API access`.
- If you expose the app outside your local network, put it behind a reverse
  proxy or VPN and tighten the trusted CIDR list.
- Mounting `/var/run/docker.sock` grants privileged host visibility and should
  be treated accordingly.

## Health Checks

Health checks are configurable per service.

- Services can keep a user-facing open URL and a separate health target for
  monitoring.
- HTTP checks support protocol, host, port, path, method, timeout, interval,
  and expected status range.
- TCP and ping checks are supported for non-HTTP services or conservative
  fallbacks.
- The UI includes an endpoint tester that returns status, latency, response
  size, resolved URL, and matched service definition when available.

## Service Definitions

HomelabWatch ships with built-in service definitions and supports custom
definitions stored in SQLite through the UI and API.

Definitions drive:

- automatic fingerprinting
- default health endpoints
- ports
- icons
- managed health-check templates

Services that operators customize manually are not overwritten by later
automatic reapply runs.

## Important Environment Variables

- `HOMELABWATCH_LISTEN_ADDR`
- `HOMELABWATCH_DATA_DIR`
- `HOMELABWATCH_DB_PATH`
- `HOMELABWATCH_STATIC_DIR`
- `HOMELABWATCH_CONFIG`
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

If you do not set `HOMELABWATCH_TRUSTED_CIDRS`, the default trusted set covers
localhost, RFC1918 private IPv4 ranges, IPv4 link-local, IPv6 loopback,
IPv6 ULA, and IPv6 link-local ranges.

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
- Full documentation: `README.md`, `CONTRIBUTING.md`, `SECURITY.md`,
  `ROADMAP.md`, and `docs/*` in the repository
