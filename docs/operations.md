# Operations Guide

## Container Runtime Expectations

- Persist `/data` if you want state to survive restarts.
- Mount `/var/run/docker.sock` only if you want local Docker discovery.
- Linux LAN discovery and ping checks usually work best with `--network host`
  and `--cap-add NET_RAW`.
- Database migrations run automatically at startup.

## Docker Compose Example

The repository includes `docker-compose.example.yml` as a starting point.

Typical flow:

```bash
cp docker-compose.example.yml docker-compose.yml
docker compose up -d
```

## Backups

HomelabWatch stores persistent state under `/data`.

At minimum, back up:

- `/data/homelabwatch.db`
- `/data/bookmark-assets/`
- any custom `config.yaml` you use outside container env vars

For a simple filesystem-level backup:

1. Stop the container or pause writes.
2. Copy the `/data` directory.
3. Restart the container.

## Restore

1. Stop HomelabWatch.
2. Restore the `/data` directory.
3. Start the same or newer version of HomelabWatch.
4. Confirm setup state, bookmarks, services, and health history in the UI.

## Upgrades

1. Read the latest release notes and `CHANGELOG.md`.
2. Back up `/data`.
3. Pull or build the new image.
4. Restart the container with the same `/data` volume.
5. Verify `GET /healthz`, the dashboard, and worker status after startup.

## Post-Upgrade Checks

- setup is still marked initialized
- bookmarks open correctly
- discovery endpoints and scan targets remain configured
- health checks still show recent results
- external automation tokens still validate

## Troubleshooting

### Discovery finds nothing

- verify the Docker socket mount if you expect Docker discovery
- verify `seedCidrs` or configured scan targets for LAN discovery
- verify host networking and `NET_RAW` on Linux for ping and network scanning

### Browser UI is read-only

- confirm the client IP falls inside `HOMELABWATCH_TRUSTED_CIDRS`
- confirm you are accessing the app from the same origin
- reload the page to refresh the console CSRF token

### Health checks fail unexpectedly

- verify path, scheme, host, port, and expected status range
- use the built-in endpoint tester before changing the saved check
- confirm the service has not switched from definition-managed to custom mode
