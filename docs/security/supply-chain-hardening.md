# Supply-Chain Hardening and Incident Response

Last updated: 2026-06-28

## Executive Summary

This repository was reviewed as though a Shai-Hulud-style npm compromise may
have occurred. The current tracked dependency graph did not contain the named
Shai-Hulud, Mini Shai-Hulud, or Miasma package families and no direct
Git, tarball, or plain-HTTP dependency was identified. That is not proof that a
developer endpoint, package-manager cache, GitHub account, CI runner, Docker Hub
account, or previously installed dependency was never compromised.

The main repository weaknesses were mutable GitHub Action tags, npm installs
that allowed dependency lifecycle scripts, permissive dependency ranges,
over-broad release permissions, a root container runtime, missing dependency
and secret scanning, no SBOM/container scan, and unbounded HTTP request bodies.
The hardening branch addresses those repository-controlled risks.

## Current Threat Findings

Microsoft's May 2026 update reports that Mini Shai-Hulud used routine package
installation as its entry point, optional Git dependencies that failed
silently after executing, obfuscated Bun payloads, broad credential discovery,
and persistence in Claude Code configuration. The earlier Shai-Hulud 2.0 wave
also used npm `preinstall` execution to deploy a credential scanner and rogue
GitHub runner. See [Microsoft's Shai-Hulud 2.0 and Mini Shai-Hulud
analysis](https://www.microsoft.com/en-us/security/blog/2025/12/09/shai-hulud-2-0-guidance-for-detecting-investigating-and-defending-against-the-supply-chain-attack/).

Unit 42 documented a CI attack that combined `pull_request_target`, checkout of
fork code, a poisoned pnpm cache, and extraction of an OIDC token from runner
memory. Later May 2026 waves returned to compromised maintainer accounts,
malicious `preinstall` hooks, bundled Bun, Git optional dependencies, and
version bumps designed to enter permissive semver ranges. The publicly released
worm code also makes copycat activity likely. See [Unit 42's June 2026 npm
threat landscape](https://unit42.paloaltonetworks.com/monitoring-npm-supply-chain-attacks/).

The defensive implications for this repository are:

- treat dependency installation and build caches as code-execution boundaries;
- reject unreviewed lifecycle scripts and exotic transitive sources;
- delay newly published versions and use exact direct versions;
- never execute fork code in `pull_request_target` or a privileged release job;
- keep OIDC and write tokens out of ordinary build/test jobs;
- pin Actions and base images to immutable identities;
- generate SBOMs and scan dependencies, source history, and images;
- assume any credential present on an infected host or runner was stolen.

The pnpm controls follow the official documentation for
[`minimumReleaseAge`, strict mode, exotic-subdependency blocking, and
`allowBuilds`](https://pnpm.io/settings) and
[`pnpm approve-builds`](https://pnpm.io/cli/approve-builds). Action SHA pinning,
least-privilege tokens, CODEOWNERS, and OIDC guidance follow
[GitHub's secure-use reference](https://docs.github.com/en/actions/reference/security/secure-use).
CISA's [software supply-chain guidance](https://www.cisa.gov/resources-tools/resources/securing-software-supply-chain-recommended-practices-developers)
supports reproducible builds, dependency integrity, and SBOM use.

## Repository Changes

### pnpm and React

- Replaced `web/package-lock.json` and every `npm ci` path with a root pnpm
  workspace and frozen `pnpm-lock.yaml`.
- Pinned pnpm to `11.9.0` and direct frontend dependencies to exact versions.
- Enforced a seven-day release delay, strict publication-time checks,
  `blockExoticSubdeps`, lockfile policy revalidation, and strict dependency
  builds.
- Denied the only identified lifecycle-script package, optional `fsevents`.
- Upgraded Vite to the patched `8.0.16` and overrode transitive `undici` to
  `7.28.0`; `pnpm audit` reports no known vulnerabilities.

### GitHub Actions and Updates

- Set `permissions: {}` at workflow scope and granted minimal job permissions.
- Pinned every action to a full commit SHA and pinned GoReleaser to `v2.16.0`.
- Added frozen-install drift checking, pnpm audit, Go tests, Go vet,
  govulncheck, CodeQL, dependency review, gitleaks, OpenSSF Scorecard, SBOM
  generation, and Trivy image scanning.
- Added release concurrency, semantic tag and main-ancestry validation, and a
  `production` environment boundary.
- Added CODEOWNERS for workflow, dependency, Docker, and release controls.
- Added conservative Dependabot groups and cooldowns. No auto-merge is
  configured; package-manager and build-script changes require human review.

### Containers and Runtime

- Pinned Node, Go, and Alpine base-image indexes by digest.
- Made both local and release images run as UID/GID `10001`.
- Added CA certificates to the release image and ensured `/data` is writable by
  the runtime user.
- Excluded environment files, npm credentials, local configuration, keys,
  editor/agent state, GitHub metadata, databases, and build output from Docker
  context.
- Enabled release image attestations/SBOMs and added scheduled container scans.

### Application Defense

- Limited ordinary request bodies to 2 MiB and bookmark asset uploads to 5 MiB.
- Return HTTP 413 for oversized JSON and multipart bodies, with regression
  tests.
- Create application data and asset directories with mode `0700`, and database
  and newly uploaded asset files with mode `0600`.
- Updated `golang.org/x/net` to `v0.55.0` after Trivy identified six fixed
  high-severity advisories in `v0.54.0`; govulncheck found no reachable
  vulnerable symbols before or after the update.

## Dependency Build-Script Approval

1. Review the package name for typosquatting, its owners, source repository,
   release history, download anomaly, and open security advisories.
2. Inspect `package.json` from the published tarball and every
   `preinstall`, `install`, and `postinstall` target. Trace downloaded or
   executed binaries and network destinations.
3. Add or update the dependency using exact versions on a clean machine or
   disposable container. Do not expose repository, cloud, npm, SSH, Docker, or
   CI credentials.
4. Run `pnpm install --ignore-scripts`, `pnpm ignored-builds`, tests, build, and
   `pnpm audit`.
5. If execution is necessary and the code is understood, run
   `pnpm approve-builds <package>`. To record a denial, run
   `pnpm approve-builds !<package>`.
6. Review the resulting `allowBuilds` and lockfile diff. Never use
   `pnpm approve-builds --all` or `dangerouslyAllowAllBuilds`.
7. Require a maintainer to approve the PR. Do not auto-merge changes involving
   `package.json`, `pnpm-lock.yaml`, `pnpm-workspace.yaml`, workflows, build
   tools, or Dockerfiles.

## After Suspected Compromise

Do not recover on the suspected host. Unit 42 reported a destructive switch in
one May 2026 variant after a stolen GitHub token was revoked, so first isolate
the host from networks, preserve volatile evidence and relevant logs, and move
response activity to a known-clean machine.

1. Isolate affected developer machines, self-hosted runners, build containers,
   and shared package caches. Disable affected workflows and publishing jobs.
2. Preserve process, network, package-manager, shell, endpoint-protection,
   GitHub Actions, registry, and cloud audit evidence. Look for unknown runners,
   repositories, workflow runs, releases, sessions, and persistence in
   `.claude/settings.json`, `.claude/setup.mjs`, `.vscode/tasks.json`, shell
   startup files, and scheduled services.
3. From a clean machine, revoke unknown sessions and rotate:
   - GitHub PATs, OAuth grants, GitHub CLI credentials, deploy keys, and app
     tokens;
   - npm tokens and package publishing/trusted-publisher identities;
   - Docker Hub or other registry tokens;
   - cloud access keys, OIDC trust bindings, Vault tokens, Kubernetes service
     account tokens, and CI/CD secrets;
   - SSH/Git credentials, database credentials, notification webhooks, and any
     application API tokens accessible to the host.
4. Review npm organization/package maintainers, 2FA/WebAuthn, provenance,
   publishing history, dist-tags, and unexpected versions. This repository does
   not currently publish npm packages, but compromised npm identities must
   still be revoked.
5. Review GitHub audit logs, Actions history, caches, artifacts, environments,
   runner registrations, release assets, branch/ruleset changes, deploy keys,
   webhooks, and repositories created or modified by the affected identity.
6. Review Docker registry push history and deployment logs. Rebuild and
   redeploy from a known-good commit on clean infrastructure; do not trust
   artifacts built during the exposure window.
7. Remove `node_modules`, run `pnpm store path`, manually delete that store
   directory, and reinstall with `pnpm install --frozen-lockfile` only after
   containment. Prefer a clean OS/container if credentials were exposed.

Never place credential values in an incident ticket or commit. Record only the
credential type, account/key name, storage location, owner, rotation status,
and timestamps.

## Required GitHub and Registry Settings

These settings cannot be enforced fully by repository files:

- configure the `production` environment with required reviewers and prevent
  self-review;
- require pull requests, CODEOWNERS approval, resolved conversations, signed
  commits/tags where practical, and all security checks on `main`;
- set the default `GITHUB_TOKEN` to read-only and prevent Actions from creating
  or approving pull requests;
- allow only required Actions and require full-SHA pins at organization level;
- enable Dependabot alerts/security updates, dependency graph, CodeQL, secret
  scanning, validity checks, and push protection;
- remove unused self-hosted runners and restrict runner groups;
- prefer short-lived OIDC or trusted publishing over long-lived cloud/npm
  credentials;
- use a scoped, expiring Docker Hub token until that registry supports an
  acceptable short-lived federation flow;
- review and delete untrusted Actions caches after any suspicious PR or runner
  execution.

## Local Security Checks

```bash
corepack enable
pnpm install --frozen-lockfile
pnpm audit --audit-level=high
pnpm --dir web test --run
pnpm --dir web build

go mod verify
go list -m all
go test ./...
go vet ./...
go install golang.org/x/vuln/cmd/govulncheck@v1.4.0
govulncheck ./...

gitleaks detect --source . --redact
docker build -t homelabwatch:security-check .
syft homelabwatch:security-check -o spdx-json
trivy image --ignore-unfixed --severity HIGH,CRITICAL \
  homelabwatch:security-check
```

Run `actionlint` against `.github/workflows` when changing CI. OpenSSF
Scorecard can also be run locally against the public repository; private
repositories without GitHub Advanced Security should use the Scorecard CLI
instead of publishing SARIF.

## Validation Results

The final branch was validated on 2026-06-28:

- frozen pnpm install, audit, 9 frontend tests, and production build passed;
- `go mod verify`, `go test ./...`, `go vet ./...`, and
  `govulncheck ./...` passed with no reported vulnerabilities;
- a redacted gitleaks scan of 50 commits found no candidate secrets;
- actionlint, workflow YAML parsing, mutable-action checks, and
  `git diff --check` passed;
- the Docker image built successfully, ran as `10001:10001`, could write
  `/data`, and could not write `/app`;
- Syft generated an SPDX JSON SBOM containing 33 packages;
- Trivy reported zero fixed high or critical vulnerabilities in the final
  image; and
- GoReleaser `v2.16.0` validated `.goreleaser.yaml`.

These results establish the state of this branch and tested image only. They do
not clear developer endpoints, external accounts, old artifacts, caches, or
deployments from the suspected exposure window.

## Remaining Risk

- Repository controls cannot prove whether historical endpoint or account
  credentials were stolen. Complete the external audit and rotation checklist.
- pnpm scripts for the application build still execute trusted, reviewed
  dependencies at build time; the allowlist primarily blocks dependency
  lifecycle hooks.
- The release job still requires a long-lived Docker Hub token. It is scoped to
  the login step, but should be replaced with short-lived federation when
  available.
- A Docker socket mount remains a host-privileged boundary. The non-root image
  needs the socket's host group added explicitly for discovery; do not weaken
  socket permissions.
- SNMP and notification-channel credentials are stored in the local SQLite
  database for runtime use. Protect and encrypt the `/data` volume and backups;
  application-level envelope encryption is not yet implemented.
- Base-image digest pins need routine reviewed updates so security patches are
  not missed.
