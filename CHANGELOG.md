# Changelog

All notable changes to the Clikd Development CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-10-29

### Added

#### Core CLI Application
- Complete Rust-based command-line tool for Clikd platform development
- 977 lines of production-ready Rust code
- Clap 4.5 with derive API for type-safe command definitions
- Tokio 1.0 async runtime with full features
- Anyhow error handling with user-friendly messages
- Structured logging with tracing and tracing-subscriber

#### Interactive Command Selector
- **Full-screen TUI with Ratatui 0.29 + Crossterm 0.29**
- Launched automatically when running `clikd` without arguments
- **9 Available Commands**:
  - `start` - Start development services
  - `stop` - Stop running services
  - `status` - Monitor service status
  - `switch` - Switch between environments
  - `logs` - View and filter service logs
  - `db` - Database management
  - `gen` - Generate client code
  - `deploy` - Deploy to environments
  - `tui` - Launch unified TUI interface

- **Keyboard Navigation**:
  - ↑↓ arrows or j/k vim-style navigation
  - Enter to execute selected command
  - q or Esc to quit
  - Visual selection with cyan highlighting
  - Help footer with key bindings

- **Professional UI**:
  - Boxed layout with clear sections
  - Command descriptions for clarity
  - Bold text for selected items
  - Consistent color scheme

#### Configuration Management System

##### Configuration File (`clikd.toml`)
Complete TOML-based configuration with 9 major sections:

**1. Project Configuration**
```toml
[project]
name = "clikd"
organization = "clikd-inc"
```

**2. API Gateway Configuration**
```toml
[api]
enabled = true
port = 9080
external_url = "http://localhost:9080"
```

**3. Service Definitions (5 services)**
- **Gate** (Authentication Service)
  - Image: `ghcr.io/clikd-inc/gate`
  - Ports: 8081 (HTTP), 9001 (gRPC)
  - Health check: `/health`
  - Dependencies: PostgreSQL Auth, KeyDB

- **Rig** (Backend API)
  - Image: `ghcr.io/clikd-inc/rig`
  - Ports: 8082 (HTTP), 9002 (gRPC)
  - Health check: `/health`
  - Dependencies: Gate, PostgreSQL Shared, ScyllaDB, NATS, MinIO

- **Media** (File Processing - Planned)
  - Image: `ghcr.io/clikd-inc/media`
  - Ports: 8083 (HTTP), 9004 (gRPC)
  - Dependencies: Rig, MinIO

- **Realtime** (WebSocket Service - Planned)
  - Image: `ghcr.io/clikd-inc/realtime`
  - Ports: 8084 (HTTP), 9003 (gRPC)
  - Dependencies: Rig, NATS

- **Studio** (Admin Dashboard)
  - Image: `ghcr.io/clikd-inc/studio`
  - Port: 3001
  - Dependencies: APISIX

**4. Database Configuration (3 databases)**
- **PostgreSQL Auth**
  - Port: 5433
  - Database: `clikd_auth_{branch}`
  - Image: `postgres:18.0`

- **PostgreSQL Shared**
  - Port: 5434
  - Database: `clikd_rig_{branch}`
  - Image: `postgres:18.0`

- **KeyDB** (Redis-compatible)
  - Port: 6380
  - Prefix: `clikd_{branch}:*`
  - Image: `eqalpha/keydb:x86_64_v6.3.4`

- **ScyllaDB** (Cassandra-compatible)
  - CQL Port: 9043
  - API Port: 10000
  - Keyspace: `clikd_{branch}`
  - Image: `scylladb/scylla:2025.1.9`

**5. Infrastructure Services**
- **MinIO** (Object Storage)
  - API Port: 9000
  - Console Port: 9001
  - Image: `minio/minio:RELEASE.2025-10-15T22-41-28Z`

- **NATS** (Message Streaming)
  - Client Port: 4222
  - HTTP Port: 8222
  - Image: `nats:2.12.1`

- **APISIX** (API Gateway)
  - HTTP Port: 9080
  - Admin Port: 9180
  - Image: `ghcr.io/clikd-inc/apisix:latest`

**6. Observability Stack (SigNoz)**
- **OpenTelemetry Collector**
  - gRPC Port: 4317
  - HTTP Port: 4318

- **SigNoz Frontend**
  - Port: 3301
  - Image: `signoz/frontend:0.97.0`

- **ClickHouse** (Analytics Database)
  - HTTP Port: 8123
  - TCP Port: 9000
  - Image: `clickhouse/clickhouse-server:25.8.11-lts`

- **ZooKeeper** (Coordination)
  - Port: 2181
  - Image: `bitnami/zookeeper:latest`

**7. Admin Tools (Optional)**
- **Adminer** (PostgreSQL Admin)
  - Port: 8090
  - Image: `adminer:latest`

- **Redis Commander** (KeyDB Admin)
  - Port: 8091
  - Image: `ghcr.io/joeferner/redis-commander:latest`

- **Cassandra Web** (ScyllaDB Admin)
  - Port: 8092
  - Image: `delermando/docker-cassandra-web:latest`

- **Swagger UI** (API Docs)
  - Port: 8093
  - Image: `swaggerapi/swagger-ui:latest`

**8. Code Generation Configuration**
```toml
[codegen]
openapi_url = "http://localhost:3002/api/openapi.json"

[codegen.swift]
output_dir = "../clients/ios"
package_name = "ClikdAPI"

[codegen.kotlin]
output_dir = "../clients/android"
package_name = "com.clikd.api"

[codegen.typescript]
output_dir = "../clients/web"
package_name = "@clikd/api"
```

**9. Deployment Configuration**
```toml
[deployment]
kubernetes_context = "clikd-cluster"
namespace_prefix = "clikd"

[deployment.staging]
auto_deploy_branches = ["develop", "staging/*"]

[deployment.production]
require_approval = true
rollback_enabled = true
```

**10. Development Settings**
```toml
[development]
branch_isolation = true      # Each git branch gets isolated environment
auto_migrate = true          # Auto-run migrations on start
auto_seed = true             # Auto-seed test data
hot_reload = true            # Auto-rebuild on code changes
network = "clikd-network"    # Docker network name
startup_timeout = "2m"       # Max time to wait for services
```

##### Configuration API
Complete Rust API for configuration management:

**Core Methods:**
- `load(path)` - Load configuration from specific path
- `load_or_default()` - Auto-discover config or use defaults
- `save(path)` - Persist configuration to disk
- `get_service_image(service, branch)` - Resolve Docker image with branch tagging
- `get_database_name(db_type, branch)` - Branch-isolated database names
- `get_keyspace_name(branch)` - ScyllaDB keyspace with branch isolation
- `get_keydb_prefix(branch)` - KeyDB key prefix with branch isolation

**Branch Isolation Logic:**
- Sanitize branch name: Replace `/\:*?"<>|` with `_`
- Docker network: `clikd-{sanitized_branch}`
- PostgreSQL Auth: `clikd_auth_{sanitized_branch}`
- PostgreSQL Shared: `clikd_rig_{sanitized_branch}`
- ScyllaDB Keyspace: `clikd_{sanitized_branch}`
- KeyDB Prefix: `clikd_{sanitized_branch}:*`
- Container labels: `clikd.branch={original_branch}`

#### Git Integration Module

##### Git Repository Detection
- Automatic repository root detection
- `.git` directory validation
- Repository path extraction

##### Branch Management
- **Current Branch Detection**: `git2::Repository::head()`
- **Branch Name Extraction**: From HEAD reference
- **Branch Sanitization**: Convert invalid characters to underscores
  - Valid: `feat/new-feature` → `feat_new_feature`
  - Valid: `release:v1.0.0` → `release_v1_0_0`

##### Repository Information
- **Commit Hash**: Short 8-character SHA
- **Commit Message**: Latest commit message
- **Working Tree Status**: Clean or dirty (uncommitted changes)
- **Main Branch Detection**: Automatically identify main/master branch

##### Git Operations
- Branch switching detection for environment switching
- Pre-commit hooks integration (planned)
- Git status integration with CLI output

#### Command Structure

##### Service Orchestration Commands

**`clikd start [OPTIONS]`** (Planned)
- Start all or specific services
- Options:
  - `--service <name>` - Start specific service
  - `--pull` - Pull latest images before starting
  - `--rebuild` - Rebuild images before starting
  - `--no-deps` - Don't start dependencies
  - `--headless` - No interactive output (CI mode)

**`clikd stop [OPTIONS]`** (Planned)
- Stop all or specific services
- Options:
  - `--service <name>` - Stop specific service
  - `--force` - Force kill containers
  - `--cleanup` - Remove containers and volumes

**`clikd status [OPTIONS]`** (Planned)
- Monitor service health and status
- Options:
  - `--watch` - Continuous monitoring
  - `--json` - JSON output for scripting

**`clikd switch <environment>`** (Planned)
- Switch between development environments
- Environments: dev, staging, production
- Automatically stops old environment and starts new

##### Database Management Commands

**`clikd db migrate [OPTIONS]`** (Planned)
- Run database migrations
- Options:
  - `--target <name>` - Run specific migration
  - `--rollback` - Rollback last migration
  - `--dry-run` - Show changes without applying

**`clikd db diff [OPTIONS]`** (Planned)
- View schema differences
- Options:
  - `--branch <name>` - Compare with another branch
  - `--format <type>` - Output format (text, json, sql)

**`clikd db reset [OPTIONS]`** (Planned)
- Reset database to clean state
- Options:
  - `--force` - Skip confirmation
  - `--keep-data` - Reset schema but keep data

**`clikd db seed [OPTIONS]`** (Planned)
- Seed database with test data
- Options:
  - `--dataset <name>` - Use specific seed dataset
  - `--partial` - Seed only specific tables

**`clikd db dump [OPTIONS]`** (Planned)
- Create database backup
- Options:
  - `--output <path>` - Output file path
  - `--compress` - Compress dump file

##### Code Generation Commands

**`clikd gen swift [OPTIONS]`** (Planned)
- Generate Swift iOS client
- Options:
  - `--output <dir>` - Output directory
  - `--package <name>` - Swift package name

**`clikd gen kotlin [OPTIONS]`** (Planned)
- Generate Kotlin Android client
- Options:
  - `--output <dir>` - Output directory
  - `--package <name>` - Java package name

**`clikd gen typescript [OPTIONS]`** (Planned)
- Generate TypeScript web client
- Options:
  - `--output <dir>` - Output directory
  - `--package <name>` - NPM package name

**`clikd gen all [OPTIONS]`** (Planned)
- Generate all client libraries in parallel
- Options:
  - `--force` - Overwrite existing files

##### Logging Commands

**`clikd logs [OPTIONS] <service>`** (Planned)
- View service logs
- Options:
  - `--follow` - Follow log output
  - `--tail <n>` - Show last N lines
  - `--since <time>` - Show logs since timestamp
  - `--filter <pattern>` - Filter logs by pattern
  - `--json` - JSON output

##### Deployment Commands

**`clikd deploy <environment> [OPTIONS]`** (Planned)
- Deploy to Kubernetes
- Environments: staging, production
- Options:
  - `--version <tag>` - Deploy specific version
  - `--dry-run` - Show deployment plan
  - `--approve` - Approve production deployment

##### TUI Commands

**`clikd tui`** (Planned)
- Launch full-screen TUI dashboard
- Features:
  - Real-time service status
  - Live log streaming
  - Resource usage graphs
  - Quick actions toolbar

#### Docker Integration

##### Bollard Client (v0.19.3)
- Async Docker daemon client
- Full Docker API support
- HTTP/Unix socket communication
- TLS support for remote Docker hosts

##### Container Lifecycle Management
- **Container Operations**:
  - Create containers from images
  - Start/stop containers
  - Remove containers
  - Inspect container state
  - Execute commands in containers

- **Image Operations**:
  - Pull images from registries
  - Build images from Dockerfile
  - Tag images
  - Remove unused images

- **Network Operations**:
  - Create Docker networks
  - Connect containers to networks
  - Disconnect containers
  - Network inspection

- **Volume Operations**:
  - Create named volumes
  - Mount volumes to containers
  - Volume inspection
  - Remove unused volumes

##### Health Check Framework (Planned)
- Periodic health endpoint polling
- Exponential backoff retry logic
- Configurable timeouts per service
- Dependency-aware health checks
- Real-time status updates

##### Service Orchestration Strategy

**Startup Sequence:**
```
Phase 1: Infrastructure (Parallel)
├── PostgreSQL Auth
├── PostgreSQL Shared
├── KeyDB
├── ScyllaDB
├── MinIO
└── NATS
  ↓ (Wait for health)

Phase 2: Backend Services (Dependency-aware)
├── Gate → depends: PostgreSQL Auth, KeyDB
├── Rig  → depends: Gate, PostgreSQL Shared, ScyllaDB, NATS, MinIO
├── Media → depends: Rig, MinIO
└── Realtime → depends: Rig, NATS
  ↓ (Wait for health)

Phase 3: Gateway
└── APISIX → depends: Gate, Rig, Media, Realtime
  ↓ (Wait for health)

Phase 4: Frontend
└── Studio → depends: APISIX
```

**Shutdown Sequence:**
```
Reverse dependency order
Frontend → Gateway → Backend Services → Infrastructure
```

#### Technology Stack

##### Core Dependencies
```toml
# Async Runtime
tokio = { version = "1.0", features = ["full"] }

# CLI Framework
clap = { version = "4.5", features = ["derive", "env"] }

# Error Handling
anyhow = "1.0"

# Logging
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }
```

##### Terminal UI
```toml
# TUI Framework
ratatui = "0.29"
crossterm = "0.29"
```

##### Configuration & Serialization
```toml
# Config Management
serde = { version = "1.0", features = ["derive"] }
toml = "0.9.8"

# JSON Processing
serde_json = "1.0"
```

##### Infrastructure Integration
```toml
# Docker Client
bollard = "0.19.3"

# Git Integration
git2 = "0.20.2"
```

##### HTTP & Code Generation
```toml
# HTTP Client
reqwest = { version = "0.12", features = ["json"] }

# OpenAPI Parsing
openapiv3 = "2.0"

# Template Engine
handlebars = "6.0"
```

##### Optional Features
```toml
# Kubernetes Client (feature-gated)
kube = { version = "2.0.1", features = ["client", "derive"], optional = true }
k8s-openapi = { version = "0.26", features = ["latest"], optional = true }
```

##### Development Tools
```toml
[dev-dependencies]
tokio-test = "0.4"
```

#### Branch-Isolated Development

##### Isolation Guarantees
- **Network Isolation**: Each branch gets dedicated Docker network
- **Database Isolation**: Branch-specific database names
- **Cache Isolation**: Branch-prefixed KeyDB keys
- **Container Isolation**: Branch labels on all containers
- **No Conflicts**: Multiple developers can work on different branches simultaneously

##### Branch Naming Strategy
- **Feature Branches**: `feat/user-authentication` → `feat_user_authentication`
- **Release Branches**: `release/v1.0.0` → `release_v1_0_0`
- **Hotfix Branches**: `hotfix/critical-bug` → `hotfix_critical_bug`

##### Environment Variables per Branch
- Automatically generated `.env.{branch}` files
- Service URLs with branch-specific ports
- Database connection strings with branch isolation
- API keys scoped to branch environment

#### Development Workflow Support

##### One-Command Setup (Planned)
```bash
clikd init
# Detects git repository
# Generates clikd.toml
# Pulls Docker images
# Starts all services with health checks
# Displays connection URLs
# Generates .env.local for frontend
```

##### Branch Switching Workflow (Planned)
```bash
git checkout feat/new-feature
# CLI detects branch change
# Stops old branch services
# Starts new branch services
# Updates environment variables
```

##### Hot Reload Support (Planned)
- Watch monorepo for code changes
- Auto-rebuild affected service containers
- Rolling restart with zero downtime
- Preserve database state during reload

##### CI/CD Integration
```bash
# Headless mode for CI pipelines
export CLIKD_ENV=ci
clikd start --headless --pull
clikd db migrate
clikd test
clikd stop --force
```

#### Infrastructure Context

##### Managed Services (19 containers)
- **Databases**: PostgreSQL (2), ScyllaDB, KeyDB
- **Infrastructure**: MinIO, NATS, APISIX
- **Observability**: SigNoz, ClickHouse, ZooKeeper, OTEL Collector
- **Admin Tools**: Adminer, Redis Commander, Cassandra Web, Swagger UI
- **Backend**: Gate, Rig (Media, Realtime planned)
- **Frontend**: Studio

##### Port Allocation
- `3xxx` - Frontend applications
- `8xxx` - Backend services
- `9xxx` - gRPC APIs + Infrastructure
- `5xxx` - PostgreSQL databases
- `6xxx` - KeyDB cluster
- `4xxx` - NATS messaging
- `809x` - Development tools

##### Version Management
All infrastructure versions tracked in `infrastructure/versions.yml`:
- PostgreSQL 18.0
- KeyDB 6.3.4
- ScyllaDB 2025.1.9 LTS
- NATS 2.12.1
- MinIO 2025.10.15
- ClickHouse 25.8.11 LTS
- SigNoz 0.97.0
- APISIX 3.11.0

#### Documentation

##### Planning Documents
- **CLIKD_CLI_PLAN.md**: Complete implementation roadmap
- **CLIKD_INIT_ANALYSIS.md**: Init command design analysis
- Inline Rust doc comments for all public APIs
- Configuration examples in `clikd.toml`

##### Project Structure
```
apps/cli/
├── Cargo.toml              # Dependencies (977 lines total)
├── clikd.toml              # Service configuration
├── moon.yml                # Moon task runner config
└── src/
    ├── main.rs             # Entry point (167 lines)
    ├── lib.rs              # Public API (4 lines)
    ├── commands/
    │   ├── mod.rs          # Command handlers (167 lines)
    │   └── selector.rs     # TUI selector (241 lines)
    ├── config/
    │   └── mod.rs          # Configuration system (292 lines)
    └── git/
        └── mod.rs          # Git integration (106 lines)
```

#### Design Philosophy

##### Supabase CLI-Inspired
- One-command setup abstracts complexity
- Branch isolation prevents developer conflicts
- Health-driven orchestration ensures reliability
- Developer-first experience

##### Rust CLI Best Practices 2025
- Minimal `main.rs` (under 200 lines)
- Clap derive API for type safety
- Structured errors (anyhow for CLI, thiserror for libraries)
- Async-first with Tokio
- Configuration layers (env, file, defaults)
- No `mod.rs` files (Rust 2018+ style)
- Explicit imports, no pub use re-exports
- Trait-based abstractions for testability

#### Future Roadmap

##### Phase 1: Core Orchestration (Planned)
- Complete Docker integration with Bollard
- Implement start/stop commands with health checks
- Service dependency resolution
- Real-time status monitoring

##### Phase 2: Database Management (Planned)
- PostgreSQL migration runner with Refinery
- ScyllaDB keyspace management
- KeyDB operations with redis crate
- Schema diff visualization

##### Phase 3: Code Generation (Planned)
- OpenAPI fetcher with reqwest
- Swift/Kotlin/TypeScript generators
- Handlebars template system
- Parallel generation for performance

##### Phase 4: Advanced TUI (Planned)
- Full dashboard with ratatui
- Live log viewer with filtering
- Service control interface
- Resource usage graphs

##### Phase 5: Deployment (Planned)
- Kubernetes integration with kube-rs
- Environment-specific configurations
- Pre-flight validation
- Rollback capabilities

##### Phase 6: Testing & Distribution (Planned)
- Comprehensive unit tests
- Integration tests with testcontainers
- Performance optimization
- Binary distribution for multiple platforms

### Changed
- N/A (Initial release)

### Deprecated
- N/A (Initial release)

### Removed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- Git repository validation prevents directory traversal
- Branch name sanitization prevents injection attacks
- Docker socket access requires proper permissions
- Configuration file validation with Serde
- No hardcoded credentials or secrets
- Environment-based secret management

[0.1.0]: https://github.com/clikd-inc/clikd/releases/tag/cli-v0.1.0
