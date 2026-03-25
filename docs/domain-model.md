# Domain Model

HomelabWatch is easiest to maintain when the core concepts stay narrow. This
document records the intended ownership boundaries for the main entities.

## Device

Owns:

- identity and confidence
- hostnames, MAC address, primary IP, and observed addresses
- open ports and inventory metadata
- operator-facing rename and hidden state

References:

- related services and discovered services by `deviceId`

Does not own:

- bookmark presentation
- health-check definitions
- discovery scheduling policy

## Service

Owns:

- an accepted, operator-visible endpoint
- network address, scheme, port, path, URL, and status
- health configuration mode and attached checks
- optional service-definition match

References:

- `deviceId`
- `serviceDefinitionId`
- `originDiscoveredServiceId`

Does not own:

- bookmark folders or tags
- discovery evidence history
- worker/job execution state

## DiscoveredService

Owns:

- candidate identity and endpoint information
- confidence, evidence, and discovery state
- optional service-definition match prior to promotion

References:

- `deviceId`
- optional accepted bookmark or service derived from it

Does not own:

- long-lived curated bookmark data
- manual health policy unrelated to discovery

## Bookmark

Owns:

- curated link presentation
- folder placement, tags, sort order, and favorite state
- icon selection and manual notes

References:

- optional `serviceId`
- optional `deviceId`

Does not own:

- service monitoring behavior
- discovery policy

## Folder

Owns:

- bookmark hierarchy and ordering

References:

- parent folder relationships

Does not own:

- service inventory
- device grouping outside bookmark organization

## Tag

Owns:

- bookmark labels and filtering vocabulary

Does not own:

- global taxonomy for every entity in the system

## HealthCheck

Owns:

- executable check configuration
- schedule, timeout, expected response rules, and target resolution
- most recent result summary

References:

- `serviceId` or another explicit subject identifier
- optional `serviceDefinitionId` if it was definition-managed

Does not own:

- service metadata beyond what is required to execute the check

## ServiceDefinition

Owns:

- fingerprinting matchers
- default ports, icons, and health-check templates
- built-in versus custom definition identity

References:

- applied services by ID only

Does not own:

- runtime health results
- discovery scheduling

## DockerEndpoint

Owns:

- connection information for a Docker source
- polling cadence and enabled state

Does not own:

- discovered services themselves
- service presentation state

## ScanTarget

Owns:

- CIDR, port seed list, schedule, and enabled state for LAN discovery

Does not own:

- actual device or service inventory records

## APIAccessToken

Owns:

- token metadata, prefix, scope, usage timestamps, and revocation state

Does not own:

- browser-session security
- trusted-network policy

## JobState / WorkerState

Owns:

- scheduler execution metadata
- last run, last error, and health of background jobs

Does not own:

- user-facing operational narrative beyond execution state

## Event / Activity Log

Owns:

- human-readable operational history across services, discovery, checks, and
  setup

Does not own:

- SSE transport details
- the canonical state of the entities it describes

## AppSettings / DiscoverySettings

Owns:

- appliance identity
- trusted-network configuration exposure
- default discovery behavior and bookmark automation policy

Does not own:

- entire feature collections as embedded state forever
- screen-specific view composition beyond what is needed for the current API

## Practical Guidance

- Add new fields to the entity that truly owns the behavior.
- Keep cross-entity relationships as identifiers, not deeply nested ownership.
- Prefer new read models over broadening `Dashboard` or `SettingsView` without
  limit.
- When in doubt, put orchestration in `internal/app`, not in HTTP handlers or
  store helpers.
