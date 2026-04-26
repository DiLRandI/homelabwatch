# Security Policy

## Supported Versions

HomelabWatch is pre-`1.0`.

- Latest tagged release, currently `v0.5.0`: supported
- `main`: supported for active contributors
- Older pre-release builds: best effort only

## Reporting a Vulnerability

Do not open a public issue for a security report.

Use one of these channels instead:

1. Open a private GitHub security advisory for this repository.
2. If that is not available, contact the maintainer through GitHub and request
   a private disclosure path.

Include:

- affected version or commit
- deployment mode
- reproduction steps
- impact assessment
- any logs, requests, or screenshots that help confirm the issue

## Security Model

HomelabWatch is designed for trusted local and LAN environments.

- The built-in browser UI is intentionally open for reading.
- Browser writes require a trusted client network, same-origin requests, and a
  server-issued CSRF token.
- External automation should use bearer tokens against `/api/external/v1/*`.
- Legacy `/api/v1/*` token routes remain for compatibility and should be
  treated as external automation surfaces, not as the primary browser API.

## Deployment Trust Boundaries

- Do not expose the built-in UI directly to the public internet.
- Put public access behind a reverse proxy, VPN, or other access-control layer.
- Tighten `HOMELABWATCH_TRUSTED_CIDRS` to match your environment.
- Mounting `/var/run/docker.sock` grants significant host visibility. Treat it
  as a privileged capability.
- Linux LAN discovery and ping checks may require `--network host` and
  `--cap-add NET_RAW`. Only grant those where needed.

## What To Expect After Reporting

- Confirmation that the report was received
- Follow-up questions if reproduction details are incomplete
- A fix and coordinated disclosure timeline when the issue is confirmed
- Release notes that describe the risk and upgrade path once the fix is public
