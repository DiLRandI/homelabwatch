# HomelabWatch

HomelabWatch is a self-hosted homelab operations dashboard for service
discovery, Docker workload visibility, health monitoring, device tracking, and
bookmarks. It runs as a single container with SQLite persistence.

## What It Does

- Discovers services from Docker and LAN scan targets
- Shows running containers as first-class dashboard inventory
- Tracks devices using MAC-aware identity where available
- Runs HTTP, TCP, and ping health checks
- Streams live updates into the dashboard
- Lets you create scoped bearer tokens for external automation

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

On an empty `/data` volume, HomelabWatch starts with a 3-step setup wizard:

1. name the appliance
2. configure discovery defaults
3. launch the initial discovery run

No admin token paste step is required in the browser.

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

## Volumes And Runtime Expectations

- Persist `/data` if you want state to survive restarts.
- Mount `/var/run/docker.sock` if you want local Docker discovery.
- Without a persistent `/data` volume, each fresh container will start the
  setup wizard again.

## Security Model

- The built-in web UI is open by design for trusted local/LAN use.
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
