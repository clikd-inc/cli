# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
