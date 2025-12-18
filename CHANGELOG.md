## [0.7.1](https://github.com/clikd-inc/cli/compare/v0.7.0...v0.7.1) (2025-12-18)


### Bug Fixes

* **ecosystem:** resolve Swift/Go versions from git tags ([#33](https://github.com/clikd-inc/cli/issues/33)) ([cf662cc](https://github.com/clikd-inc/cli/commit/cf662cc6899baf4fa38d21100c2ebfb3670894ac))

# [0.7.1](https://github.com/clikd-inc/cli/compare/v0.7.0...v0.7.1) (2025-12-18)


### Bug Fixes

* **ecosystem:** resolve Swift/Go versions from git tags ([#33](https://github.com/clikd-inc/cli/pull/33)) - Swift and Go ecosystems don't store version information in their manifest files. Previously showed 0.0.0, now correctly detects version from existing git tags (e.g., `v1.2.3` or `project-v1.2.3`)

# [0.7.0](https://github.com/clikd-inc/cli/compare/v0.6.0...v0.7.0) (2025-12-18)


### Features

* **ecosystem:** add Swift Package Manager support ([e08dcc4](https://github.com/clikd-inc/cli/commit/e08dcc4b2bcd592baf237722e23c4f0e9b71b8c0))

# [0.6.0](https://github.com/clikd-inc/cli/compare/v0.5.1...v0.6.0) (2025-12-06)


### Features

* **release:** implement PR-based release workflow ([dd35fbd](https://github.com/clikd-inc/cli/commit/dd35fbdc04249b43923403dd29bbba9251fe49ce))

## [0.5.1](https://github.com/clikd-inc/cli/compare/v0.5.0...v0.5.1) (2025-12-04)


### Bug Fixes

* **release:** exclude clikd/ directory from dirty check during init ([#23](https://github.com/clikd-inc/cli/issues/23)) ([a797679](https://github.com/clikd-inc/cli/commit/a797679e7f6249e82c7ba40a57f4fdf34e964746))

# [0.5.0](https://github.com/clikd-inc/cli/compare/v0.4.0...v0.5.0) (2025-12-03)


### Features

* **release:** add complete release management system with monorepo support ([86ef553](https://github.com/clikd-inc/cli/commit/86ef5533e859211af4d641d4b5f28cc91fdf4b64))

# [0.4.0](https://github.com/clikd-inc/cli/compare/v0.3.0...v0.4.0) (2025-11-09)


### Features

* add interactive TUI for container monitoring ([#7](https://github.com/clikd-inc/cli/issues/7)) ([92799c1](https://github.com/clikd-inc/cli/commit/92799c1c3017dc4cceacf83a3c9b72d2e1ebfbcc)), closes [#0d1117](https://github.com/clikd-inc/cli/issues/0d1117)


### BREAKING CHANGES

* db and logs commands have been removed

# [0.3.0](https://github.com/clikd-inc/cli/compare/v0.2.7...v0.3.0) (2025-11-08)


### Features

* **ci:** add CHANGELOG.md generation for releases ([e29a715](https://github.com/clikd-inc/cli/commit/e29a7152ff130e3dee868c1e50488f1b9400c04e))

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.6] - 2025-11-08

### Added

- **CLI Self-Update Notifications**
  - Automatic update check after every command execution
  - Notifies users when newer CLI version is available on GitHub
  - `--version` flag always fetches latest release information
  - Regular commands use 10-hour cache to minimize API calls
  - Non-blocking notifications with upgrade instructions

### Changed

- Custom `--version` implementation (replaces clap's default handler)
  - Provides consistent update checking behavior
  - Ensures notifications appear regardless of command success/failure

## [0.2.5] - 2025-11-08

### Changed

- Updated Docker image versions:
  - gate: 1.0.0 → 1.0.1
  - studio: 0.5.0 → 0.6.0

### Fixed

- Auto-create version pin files on first `clikd start` without requiring `clikd init`
- Users can now run `clikd start` directly (like Supabase) and still benefit from version tracking
- Version update warnings now work correctly for projects started without explicit initialization

## [0.2.4] - 2025-11-07

### Added

- **Per-Project Docker Image Version Pinning**
  - Automatically pins Docker image versions at project initialization
  - Stores versions in `clikd/.temp/*-version` files
  - Prevents breaking changes when updating CLI
  - Each project maintains its own image versions independently

- **Automatic Version Update Detection**
  - Warning displayed at startup when newer service versions are available
  - Compares local pinned versions with latest CLI defaults
  - Non-intrusive notifications for outdated services

- **Version Update Command** (`clikd update`)
  - Interactive upgrade command for service image versions
  - Shows detailed comparison of current vs. latest versions
  - Confirmation prompt before applying updates (skip with `--yes`)
  - Safe upgrade path without breaking existing projects

- **Automated Dependency Management**
  - Dependabot configuration for Docker image updates
  - Daily checks for Docker and Cargo dependency updates
  - Grouped minor/patch updates for easier review
  - Weekly GitHub Actions workflow updates

- **Single Source of Truth for Docker Images**
  - Centralized image version management in `config/images.Dockerfile`
  - Automatic parsing at runtime with zero overhead
  - Fallback to hardcoded defaults for safety
  - Simplifies version updates and maintenance

### Changed

- Refactored image configuration system for better maintainability
- Config loader now supports version override from project-specific files

## [0.2.3] - 2025-11-04

### Added

- APISIX route configuration for `/docs/*` endpoint
  - Routes documentation requests to Gate service (port 8081)
  - Matches backend API gateway configuration

## [0.2.2] - 2025-11-04

### Fixed

- Corrected authentication command references in output messages
  - Changed `clikd auth login` to `clikd login` in error and status messages
  - Updated authentication error message to show correct command syntax
  - Applied code formatting improvements for consistency

## [0.2.1] - 2025-11-03

### Changed

- **Package renamed from `clikd-cli` to `clikd`**
  - Simplified package name for standalone repository
  - Homebrew formula now published as `clikd` instead of `clikd-cli`
  - Installation: `brew install clikd-inc/tap/clikd`

## [0.2.0] - 2025-11-03

### Added

#### Core Commands
- **Project Initialization** (`clikd init`)
  - Automatic project setup with configuration generation
  - IDE integration support for VS Code and IntelliJ IDEA
  - Git integration with automatic `.gitignore` management
  - Branch-aware project isolation

- **Authentication System** (`clikd login`, `clikd logout`, `clikd auth status`)
  - GitHub OAuth device flow authentication
  - Organization membership verification (clikd-inc)
  - Secure token storage using system keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
  - Automatic browser opening for authorization (with `--no-browser` fallback)

- **Environment Management**
  - Complete Docker-based local development environment with 10 pre-configured services:
    - Gate (authentication service)
    - Rig (core API service with GraphQL)
    - Studio (frontend application)
    - PostgreSQL (auth and main databases)
    - KeyDB (Redis-compatible cache)
    - ScyllaDB (NoSQL database)
    - MinIO (object storage)
    - NATS with JetStream (message broker)
    - APISIX (API gateway with automated routing)
  - Intelligent dependency resolution and ordered service startup
  - Automatic health checking with configurable timeouts
  - Project and branch-based container isolation

- **Service Management**
  - Enhanced `start` command with:
    - Progress tracking for Docker image pulls
    - Platform-specific image support (ARM64, x86_64)
    - Service exclusion via `--exclude` flag
    - Health check bypass option (`--ignore-health-check`)
    - Automatic GHCR authentication for private registries
  - Improved `stop` command with:
    - `--purge` flag to remove all volumes
    - Graceful shutdown of all project containers

#### IDE Integration
- **VS Code Integration**
  - Auto-generated `.vscode/settings.json` with optimized formatter settings
  - Recommended extensions list (`.vscode/extensions.json`)
  - Biome, TOML, Tailwind CSS, TypeScript, Swift, and Kotlin support

- **IntelliJ IDEA / Android Studio Integration**
  - Auto-generated `.idea/clikd.xml` configuration
  - Config path auto-detection

#### Developer Experience
- **Rich Terminal UI**
  - Custom color theme with brand colors
  - Progress bars for long-running operations
  - Spinners for async tasks
  - Formatted success/error/warning/info messages
  - `--no-color` flag for accessibility

- **Configuration System**
  - Hierarchical TOML-based configuration
  - Environment variable overrides (`CLIKD_*` prefix)
  - Per-project and per-environment config support
  - Service enable/disable toggles
  - Custom port mappings
  - Image version overrides

- **Template System**
  - APISIX route configuration templates
  - Project configuration templates
  - IDE settings templates

### Changed

- Migrated from monorepo to standalone repository structure
- Refactored Docker manager for better image pulling and platform support
- Simplified network and health check modules
- Updated service image versions to latest stable releases
- Improved error handling with detailed error types

### Fixed

- Docker image authentication for private GHCR registries
- Service dependency resolution order
- Health check reliability and timeout handling
- Git branch detection and sanitization
- Configuration file merging across different sources

### Technical

- Rust 1.91 MSRV (Minimum Supported Rust Version)
- Cross-platform support: macOS (Intel & Apple Silicon), Linux (x86_64 & ARM64), Windows (x64)
- Automated multi-platform releases with cargo-dist
- MSI installer for Windows
- Homebrew tap for macOS

## [0.1.0] - 2024-XX-XX

### Added

- Initial release of Clikd CLI
- Basic command structure
- Docker service orchestration foundation

[0.2.0]: https://github.com/clikd-inc/cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/clikd-inc/cli/releases/tag/v0.1.0
