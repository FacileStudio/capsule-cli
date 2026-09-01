# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html). While on
`0.x`, a breaking change bumps the minor.

Every entry below was reconstructed from git history on 2026-08-24, so they
record what shipped rather than what was written down at the time.

## [Unreleased]

## [0.2.0] — 2026-09-01

### Added

- `capsule keys` command group (`list`, `create`, `revoke`) to manage secret backend and public browser API keys.
- API client methods `ListKeys`, `CreateKey`, and `RevokeKey`.
- `--json` persistent flag for structured output.

## [0.1.2] — 2026-08-24

### Added

- Releases publish a Homebrew formula to `FacileStudio/homebrew-tap`.
- `integrations/SKILL.md`, so an AI agent can drive the CLI.

### Changed

- `install.sh` bootstraps `facile` from `get.facile.studio` and falls back to
  the raw GitHub URL, so one unreachable host cannot break the install command.
- The README installs via `facile` and drops the paths that write somewhere
  other than `~/.local/bin`.

## [0.1.1] — 2026-08-10

### Fixed

- `capsule --version` printed cobra's default `capsule version <v>`. It now
  prints `capsule <semver>`, which is the line the installer parses.

### Changed

- Command output goes through an `internal/ui` module instead of being printed
  ad hoc from each command.

## [0.1.0] — 2026-08-10

### Added

- First release. `seal`, `reveal` and `revoke` share a secret snippet through a
  Capsule instance, with encryption and decryption done in the CLI so the
  server never sees the plaintext or the key.
- `config` reads and writes the local settings; the default instance is
  `capsule.facile.studio`.
- The server URL is derived from the capsule URL passed to `reveal` and
  `revoke`, so a link from another instance works without reconfiguring.
- `--version` reports the tag, injected by ldflags at build time.

[Unreleased]: https://github.com/FacileStudio/capsule-cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/FacileStudio/capsule-cli/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/FacileStudio/capsule-cli/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/FacileStudio/capsule-cli/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/FacileStudio/capsule-cli/releases/tag/v0.1.0
