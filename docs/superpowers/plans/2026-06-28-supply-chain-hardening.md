# Supply-Chain Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden HomelabWatch against Shai-Hulud-style dependency, CI/CD, credential-theft, and release-pipeline compromise.

**Architecture:** Replace npm installation paths with a pinned pnpm toolchain that denies unreviewed lifecycle scripts, then layer least-privilege immutable CI, dependency and source scanning, hardened container builds, and incident-response documentation around the existing Go/React application. Add bounded HTTP request handling because unbounded request bodies amplify a compromised or exposed runtime.

**Tech Stack:** Go 1.25, React 19, Node.js 24, pnpm 11.9.0, GitHub Actions, Docker/BuildKit, Dependabot, CodeQL, govulncheck, gitleaks, Trivy, Syft.

---

### Task 1: Migrate the frontend to a controlled pnpm install policy

**Files:**
- Modify: `web/package.json`
- Create: `pnpm-workspace.yaml`
- Create: `pnpm-lock.yaml`
- Delete: `web/package-lock.json`
- Modify: `Makefile`
- Modify: `.gitignore`

- [ ] Pin `packageManager` to `pnpm@11.9.0` and convert direct dependency ranges to exact versions already recorded in the trusted lockfile.
- [ ] Configure a seven-day release delay, strict missing-time behavior, exotic-subdependency blocking, dependency-state verification, strict build-script review, and an explicit `allowBuilds` policy.
- [ ] Generate the pnpm lockfile from the reviewed npm lockfile without running dependency scripts.
- [ ] Run `pnpm install --frozen-lockfile`, `pnpm test -- --run`, and `pnpm build`.
- [ ] Commit with `build: harden pnpm dependency installation`.

### Task 2: Harden CI and release workflows

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`
- Create: `.github/workflows/codeql.yml`
- Create: `.github/workflows/dependency-review.yml`
- Create: `.github/workflows/security.yml`
- Create: `.github/dependabot.yml`
- Create: `.github/CODEOWNERS`

- [ ] Pin every action to a verified full commit SHA with a trailing release comment.
- [ ] Set top-level `permissions: {}` and grant job-local read or write permissions only where required.
- [ ] Use Corepack with the repository-pinned pnpm version and frozen installs.
- [ ] Add Go tests, vet, govulncheck, pnpm audit, lockfile-drift checks, gitleaks, CodeQL, dependency review, and conservative Dependabot updates.
- [ ] Add release concurrency, a production environment gate, tag/ref validation, and explicit release-job permissions.
- [ ] Validate workflow YAML and action references.
- [ ] Commit with `ci: harden workflows and add security gates`.

### Task 3: Harden Docker and release artifacts

**Files:**
- Modify: `Dockerfile`
- Modify: `Dockerfile.release`
- Modify: `.dockerignore`
- Modify: `.goreleaser.yaml`
- Create: `.github/workflows/container-security.yml`

- [ ] Pin base images by digest after verifying architecture support.
- [ ] Use pnpm with frozen lockfile and approved scripts in the frontend build stage.
- [ ] Create and use an unprivileged runtime account and restrict copied build context.
- [ ] Generate SBOMs and scan the built container for high/critical vulnerabilities.
- [ ] Build the local container and verify the runtime user.
- [ ] Commit with `build: harden container and release artifacts`.

### Task 4: Bound HTTP request bodies

**Files:**
- Modify: `internal/api/http/router.go`
- Modify: `internal/api/http/bookmarks.go`
- Modify: `internal/api/http/router_test.go`

- [ ] Add failing tests proving oversized JSON and multipart asset requests return `413 Request Entity Too Large`.
- [ ] Run the focused tests and confirm the expected failures.
- [ ] Add a global bounded-body middleware plus a smaller bookmark-asset limit without changing normal API behavior.
- [ ] Run focused and full Go tests, then `go vet ./...`.
- [ ] Commit with `security: bound API request bodies`.

### Task 5: Document prevention and incident response

**Files:**
- Create: `docs/security/supply-chain-hardening.md`
- Modify: `SECURITY.md`
- Modify: `CONTRIBUTING.md`
- Modify: `README.md`

- [ ] Document current Shai-Hulud and Mini Shai-Hulud behavior with dated, linked sources and repository-specific mitigations.
- [ ] Document build-script approval, dependency review, local security commands, clean recovery, and frozen reinstall procedures.
- [ ] Add a non-secret credential/session/package-publishing rotation checklist and external GitHub settings checklist.
- [ ] Record residual risks and manual production actions.
- [ ] Commit with `docs: add supply-chain incident response guidance`.

### Task 6: Final verification and scan closure

**Files:**
- Modify only if verification finds a defect.

- [ ] Run `pnpm install --frozen-lockfile`, frontend tests/build, `pnpm audit`, `go mod verify`, `go list -m all`, `go test ./...`, `go vet ./...`, `govulncheck ./...`, gitleaks, workflow syntax checks, Docker build, SBOM generation, and container vulnerability scanning where tooling is available.
- [ ] Inspect `git diff --check`, the final branch diff, and commit history.
- [ ] Record exact pass/fail/tool-unavailable results and remaining manual actions in `docs/security/supply-chain-hardening.md`.
