# Feature: Expanded Health Check Types

## Goal

Broaden HomelabWatch health monitoring beyond HTTP, TCP, and ping so it can
cover common homelab failure modes without requiring a separate uptime tool.

## User Value

- catch TLS certificates before expiry
- verify DNS records and local DNS servers
- check response content, not only HTTP status
- validate JSON API health responses
- monitor Docker container health state directly

## Scope

Add these check types:

- `dns`
- `tls`
- `http_keyword`
- `http_json`
- `docker_container`

Keep existing `http`, `tcp`, and `ping` behavior stable.

Defer:

- database checks
- browser-based checks
- gRPC checks
- distributed probe locations

## Domain Model

Extend `CheckType` with:

- `dns`
- `tls`
- `http_keyword`
- `http_json`
- `docker_container`

Extend `ServiceCheck` carefully with optional fields:

- `DNSRecordType`
- `ExpectedDNSValue`
- `TLSMinDaysRemaining`
- `Keyword`
- `KeywordMode`: `contains`, `not_contains`
- `JSONPath`
- `ExpectedJSONValue`
- `DockerContainerID`
- `DockerContainerName`

If the number of type-specific fields becomes unwieldy, add a structured
`Config map[string]any` column for future check-specific options. Prefer typed
fields for the first implementation only if the migration remains clear.

Extend `CheckResult` with optional fields:

- `ObservedValue`
- `ExpiresAt`
- `DaysRemaining`

## Backend Work

- Add migration for new check fields or a JSON config column.
- Update create/patch validation per check type.
- Update `internal/monitoring/runner.go`:
  - `dns`: resolve requested record and compare optional expected value
  - `tls`: connect with TLS and check certificate expiry
  - `http_keyword`: fetch HTTP response and apply keyword rule
  - `http_json`: fetch JSON response and evaluate a simple dot-path
  - `docker_container`: inspect known Docker source data or Docker endpoint
    state to confirm running/healthy status
- Keep timeout handling consistent with existing checks.
- Add endpoint tester support for all new check types.
- Update service definition templates so built-ins can use new check types when
  useful.

## API Work

Existing health endpoints should remain:

- `GET /api/ui/v1/services/{id}/checks`
- `POST /api/ui/v1/services/{id}/checks`
- `PATCH /api/ui/v1/checks/{id}`
- `DELETE /api/ui/v1/checks/{id}`
- `POST /api/ui/v1/services/{id}/checks/test`

Update request/response shapes additively. Do not remove existing fields.

## Frontend Work

- Update health check forms with type-specific fields.
- Keep fixed-size form sections so switching types does not produce layout
  instability.
- Show result details:
  - DNS observed value
  - TLS expiry and days remaining
  - keyword match result
  - JSON observed value
  - Docker state
- Update endpoint tester display for new result fields.

## Validation Rules

- `dns` requires host and record type.
- `tls` requires host and port, default port 443.
- `http_keyword` requires URL target or HTTP components plus keyword.
- `http_json` requires URL target or HTTP components plus JSON path.
- `docker_container` requires a source reference, container ID, or container
  name.
- Intervals and timeouts follow existing minimum/default behavior.

## Tests

Backend:

- validation for each type
- DNS success/failure with test resolver abstraction
- TLS expiry success/failure with test server
- keyword contains and not_contains
- JSON dot-path matching
- Docker state check against fake provider/store data
- migration preserves existing checks

Frontend:

- type selector renders correct fields
- tester displays type-specific output
- existing HTTP/TCP/ping forms still work

## Acceptance Criteria

- operators can create and test DNS, TLS, keyword, JSON, and Docker checks
- existing checks continue to run after migration
- service status aggregation includes new check results
- result messages are specific enough to troubleshoot failures
