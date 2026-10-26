# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `visibility` and `scope` attributes on `ametnes_network`. Load balancers are
  now **private by default**; `scope` selects a cluster-wide (`location`, the
  default) or project-scoped (`project`) load balancer. When omitted, the
  location default applies and the resolved value is recorded. These replace
  the previously documented `config["public"]` key.
- A `network` block on `ametnes_location` carrying the load balancer policy:
  `public`/`private` capabilities and `visibility`/`scope` defaults, plus a
  `telemetry` block with `resources` and `system` scopes (each with
  `logs`/`metrics`). These can be updated in place; the location identity
  (`name`, `code`, `description`) stays immutable. The values mirror what the
  Ametnes Cloud console edits through `PUT metadata/locations`.
- New `UpdateLocation` client method that issues `PUT metadata/locations/{id}`.
- The `ametnes_location` data source exposes the location's load balancer
  capabilities/defaults and telemetry as read-only attributes.
- Documentation explaining automatic load balancer management (distribution
  across available load balancers, creation/deletion, and port allocation with
  the standard-port-then-next-free-port fallback) and the location load balancer
  policy, including that load balancer visibility (`public`/`private`) is
  irrelevant on-premises because every load balancer is provisioned inside the
  customer's own environment.

### Changed
- `ametnes_service` now defaults the auto-managed load balancer to **private**
  (`config["public.visible"]` defaults to `"false"` instead of `"true"`). Set
  `public.visible = "true"` to request a public load balancer, subject to the
  location's capabilities.
- `ametnes_service.kind` no longer forces replacement. Changing `kind` to a
  version listed in the current kind's catalog upgrade path upgrades the
  service in place, matching the control plane API. In-place updates are now
  also sent for `kind` changes, and the applied kind is read back into state.

## [0.4.0] - 2026-08-16

### Added
- In-place updates for `ametnes_service` (`name`, `description`, `capacity`,
  `config`, `alias`) and `ametnes_network` (`name`, `description`,
  `config`) - these no longer force recreation.
- `alias` attribute on `ametnes_service`.
- Configurable `timeouts` blocks on `ametnes_service` (create `60m`, update
  `45m`, delete `10m`) and `ametnes_network` (create `15m`, update `15m`,
  delete `10m`).
- `config` map on `ametnes_network` with a `public` key to select public vs
  private load balancer provisioning.
- `ametnes_service.network` is now optional; a network is auto-created when
  omitted.
- `ametnes_service.config` defaults `public.visible` to `"true"` when unset.
- Unit and mock tests for the client, status polling, resource create/update,
  and timeouts, plus a `make testunit` target.
- GitHub Actions workflow running `go vet` and unit tests on push and pull
  requests.

### Changed
- Status polling now fails immediately on `ERROR` and succeeds on
  `READY`/`INITIALIZED`, instead of waiting out the full timeout.
- Default API host updated to `https://cloud.ametnes.com/api/c/v1`.
- `ametnes_kinds` and `ametnes_network` data sources return clearer error
  messages and handle empty results.
- Examples and docs switched to `for_each` maps with `random_string` aliases
  and `architecture` presets; removed `cpu` and `memory` from capacity blocks
  (sizing is now driven entirely by the architecture preset).
- `storage` in the `capacity` block is optional - if omitted, the backend
  assigns a default value - and the configured value is distributed across
  the service's components in predetermined proportions.

### Removed
- Redundant `network` (Number) field from the `ametnes_network` schema.
- The `nodes` attribute from `ametnes_service`; the node count is now fixed
  internally to `1` and driven by the architecture preset.
- The `cpu` and `memory` attributes from the `ametnes_service` `capacity` block;
  compute sizing is now driven by the architecture preset.

### Fixed
- Documentation typos (`lcoation` → `location`, `creat` → `create`,
  `loadbalance` → `load balancer`).
