# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-08-16

### Added
- In-place updates for `ametnes_service` (`name`, `description`, `capacity`,
  `nodes`, `config`, `alias`) and `ametnes_network` (`name`, `description`,
  `config`) — these no longer force recreation.
- `alias` attribute on `ametnes_service`.
- Configurable `timeouts` blocks on `ametnes_service` (create `60m`, update
  `45m`, delete `10m`) and `ametnes_network` (create `15m`, update `15m`,
  delete `10m`).
- `config` map on `ametnes_network` with a `public` key to select public vs
  private load balancer provisioning.
- `ametnes_service.network` is now optional; a network is auto-created when
  omitted.
- `ametnes_service.nodes` now defaults to `1`.
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
  and `architecture` presets; removed `memory = 1` / `cpu = 1` from example
  capacity blocks.

### Removed
- Redundant `network` (Number) field from the `ametnes_network` schema.

### Fixed
- Documentation typos (`lcoation` → `location`, `creat` → `create`,
  `loadbalance` → `load balancer`).
