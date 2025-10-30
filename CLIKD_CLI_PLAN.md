# Clikd CLI - Production-Ready Implementation

## Folder-Struktur (Rust 2018+ / CLI Best Practices 2025)

```
apps/cli/
├── Cargo.toml
├── src/
│   ├── main.rs              # 20-30 Zeilen: Setup + Execute
│   ├── cli.rs               # clap Definitionen
│   ├── error.rs             # thiserror Error-Typen
│   ├── config.rs            # Config-Loader
│   │
│   ├── cmd/                 # Command handlers (thin - wie Supabase)
│   │   ├── auth.rs          # Auth subcommands
│   │   ├── start.rs         # Start command
│   │   ├── stop.rs          # Stop command
│   │   ├── status.rs        # Status command
│   │   ├── logs.rs          # Logs command
│   │   ├── db.rs            # DB commands
│   │   └── completions.rs   # Shell completions
│   │
│   ├── core/                # Business logic (wie Supabase internal/)
│   │   ├── auth/
│   │   │   ├── github.rs    # OAuth Device Flow
│   │   │   ├── token.rs     # Keyring storage
│   │   │   └── org_check.rs # GitHub org membership
│   │   │
│   │   ├── docker/
│   │   │   ├── manager.rs   # Bollard container orchestration
│   │   │   ├── services.rs  # Service definitions (aus docker-compose.yml)
│   │   │   ├── health.rs    # Health check polling
│   │   │   └── network.rs   # Network management
│   │   │
│   │   ├── git/
│   │   │   └── branch.rs    # Branch detection
│   │   │
│   │   ├── start/
│   │   │   └── runner.rs    # Start orchestration
│   │   │
│   │   └── stop/
│   │       └── runner.rs    # Stop orchestration
│   │
│   └── utils/               # Shared utilities
│       ├── terminal.rs      # owo-colors styling
│       └── retry.rs         # Retry logic
│
├── config/
│   └── default.toml
│
└── tests/
    ├── integration/         # assert_cmd tests
    └── cmd/                 # trycmd snapshot tests
```

### Wichtig - Keine mod.rs Files!

```rust
// src/cmd.rs existiert NICHT
// Stattdessen in src/cli.rs direkt importieren:
use crate::cmd::auth;
use crate::cmd::start;
// etc.
```

---

## Cargo.toml (CLI Best Practices 2025)

```toml
[package]
name = "clikd-cli"
version = "0.1.0"
edition = "2021"

[[bin]]
name = "clikd"
path = "src/main.rs"

[dependencies]
# CLI Framework
clap = { version = "4.5", features = ["derive", "env", "wrap_help", "cargo"] }
clap_complete = "4.5"

# Error Handling (2025 Standard)
anyhow = "1.0"           # Application layer
thiserror = "2.0"        # Library code

# Logging (tracing, nicht env_logger!)
tracing = "0.4"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }

# Terminal UI (2025 Standard)
owo-colors = { version = "4.2", features = ["supports-colors"] }
indicatif = "0.17"
dialoguer = "0.11"
ratatui = "0.29"         # Interactive TUI for dashboard
crossterm = "0.28"       # Terminal handling for ratatui

# Configuration (config für layered support)
config = "0.14"
serde = { version = "1.0", features = ["derive"] }
toml = "0.8"

# Docker SDK (wie Supabase)
bollard = "0.17"

# Git
git2 = "0.19"

# Auth & Security
keyring = "3.8"
reqwest = { version = "0.12", features = ["json", "rustls-tls"] }
secrecy = "0.8"
zeroize = "1.8"

# Async Runtime (nur benötigte features!)
tokio = { version = "1.40", features = ["rt-multi-thread", "macros"] }
futures = "0.3"

# Utilities
chrono = "0.4"
uuid = { version = "1.0", features = ["v4"] }

[dev-dependencies]
trycmd = "0.15"          # Snapshot testing
assert_cmd = "2.0"       # CLI integration tests
assert_fs = "1.1"        # Filesystem tests
tempfile = "3.8"

[profile.release]
opt-level = "z"          # Size optimization
lto = true
codegen-units = 1
strip = true
panic = "abort"
```

---

## Docker Services (aus docker-compose.yml)

### Service-Kategorien & Reihenfolge

#### 1. Databases (First Priority)

- **postgres-auth** - Port 5433 - Health: pg_isready
- **postgres-rig** - Port 5434 - Health: pg_isready
- **keydb** - Port 6380 - Health: keydb-cli ping
- **scylladb** - Port 9043 - Health: cqlsh (60s startup!)

#### 2. Infrastructure

- **minio** - Port 9000/9901 - Health: /minio/health/live
- **nats** - Port 4222/8222 - No health check

#### 3. Observability (Dependency Chain!)

- **zookeeper-1** - Port 2181/8094 - Health: zkServer.sh status
- **clickhouse** - Port 8123/9100 - Health: SELECT 1 query
  - depends_on: zookeeper-1
- **schema-migrator** - (one-shot) - depends_on: clickhouse
- **signoz** - Port 3301 - depends_on: schema-migrator
- **otel-collector** - Port 4317/4318 - depends_on: clickhouse

#### 4. Backend Services (DEV MODE!)

- **gate** - Port 8081/9001 - APP_ENV=development, RUST_LOG=debug
  - depends_on: postgres-auth, keydb
- **rig** - Port 8082/9002 - APP_ENV=development, RUST_LOG=debug
  - depends_on: postgres-rig, keydb, scylladb, nats, minio
- **apisix** - Port 9080 - API Gateway
  - depends_on: gate, rig
- **studio** - Port 3001 - NODE_ENV=development, APP_ENV=development
  - depends_on: apisix

#### 5. Admin UIs (Optional - exclude by default)

- **postgres-admin** - Port 8090
- **keydb-admin** - Port 8091
- **scylla-admin** - Port 8092
- **swagger-ui** - Port 8093

---

## Core Implementation

### 1. CLI Definition (src/cli.rs)

```rust
use clap::{Parser, Subcommand, Args, ValueEnum};

#[derive(Parser)]
#[command(name = "clikd", version, about = "Development CLI for Clikd platform")]
pub struct Cli {
    /// Increase verbosity (-v, -vv, -vvv)
    #[arg(short, long, action = clap::ArgAction::Count, global = true)]
    pub verbose: u8,

    /// Disable colored output
    #[arg(long, global = true)]
    pub no_color: bool,

    /// Environment (development/production)
    #[arg(short, long, global = true, env = "CLIKD_ENV")]
    pub env: Option<String>,

    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand)]
pub enum Commands {
    /// Authentication commands
    #[command(subcommand)]
    Auth(AuthCommands),

    /// Start local development environment
    Start(StartArgs),

    /// Stop local development environment
    Stop(StopArgs),

    /// Show service status
    Status(StatusArgs),

    /// View service logs
    Logs(LogsArgs),

    /// Database commands
    #[command(subcommand)]
    Db(DbCommands),

    /// Generate shell completions
    Completions {
        #[arg(value_enum)]
        shell: Shell,
    },
}

#[derive(Subcommand)]
pub enum AuthCommands {
    /// Login via GitHub OAuth Device Flow
    Login {
        #[arg(long)]
        no_browser: bool,
    },

    /// Logout (clear stored token)
    Logout,

    /// Check authentication status
    Status,
}

#[derive(Args)]
pub struct StartArgs {
    /// Exclude services (comma-separated)
    /// Examples: postgres-admin,keydb-admin,scylla-admin,swagger-ui
    #[arg(long, value_delimiter = ',')]
    pub exclude: Option<Vec<String>>,

    /// Pull latest images before starting
    #[arg(long)]
    pub pull: bool,

    /// Ignore health check failures
    #[arg(long)]
    pub ignore_health_check: bool,
}

#[derive(Args)]
pub struct StopArgs {
    /// Force stop without confirmation
    #[arg(short, long)]
    pub force: bool,

    /// Delete all volumes
    #[arg(long)]
    pub purge: bool,
}

#[derive(Args)]
pub struct StatusArgs {
    /// Output format
    #[arg(short, long, value_enum, default_value = "table")]
    pub format: OutputFormat,
}

#[derive(Clone, ValueEnum)]
pub enum OutputFormat {
    Table,
    Json,
    Env,
}
```

### 2. Main Entry Point (src/main.rs - MAX 30 Zeilen!)

```rust
use anyhow::Result;
use clap::Parser;
use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() -> Result<()> {
    // Setup logging basierend auf verbosity
    let cli = clikd_cli::cli::Cli::parse();
    init_logging(cli.verbose);

    // Color detection
    if cli.no_color {
        owo_colors::set_override(false);
    }

    // Load config
    let config = clikd_cli::config::load(cli.env.as_deref())?;

    // Execute command
    clikd_cli::execute(cli, config).await
}

fn init_logging(verbosity: u8) {
    let level = match verbosity {
        0 => "warn",
        1 => "info",
        2 => "debug",
        _ => "trace",
    };

    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::new(level))
        .init();
}
```

### 3. Service Definitions (src/core/docker/services.rs)

```rust
use bollard::service::{HealthConfig, Mount, PortBinding};
use std::collections::HashMap;
use std::time::Duration;

pub struct ServiceDefinition {
    pub name: &'static str,
    pub image: String,
    pub ports: Vec<(u16, u16)>,
    pub env: HashMap<String, String>,
    pub volumes: Vec<String>,
    pub health_check: Option<HealthCheck>,
    pub depends_on: Vec<&'static str>,
    pub command: Option<Vec<String>>,
}

pub struct HealthCheck {
    pub test: Vec<String>,
    pub interval: Duration,
    pub timeout: Duration,
    pub retries: u32,
    pub start_period: Option<Duration>,
}

/// Get all services in dependency order
pub fn all_services(branch: &str) -> Vec<ServiceDefinition> {
    vec![
        // 1. Databases first
        postgres_auth_service(branch),
        postgres_rig_service(branch),
        keydb_service(branch),
        scylladb_service(branch),

        // 2. Infrastructure
        minio_service(branch),
        nats_service(branch),

        // 3. Observability stack
        zookeeper_service(branch),
        clickhouse_service(branch),
        schema_migrator_service(branch),
        signoz_service(branch),
        otel_collector_service(branch),

        // 4. Backend services (DEV MODE!)
        gate_service(branch),
        rig_service(branch),
        apisix_service(branch),
        studio_service(branch),
    ]
}

fn gate_service(branch: &str) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("APP_ENV".into(), "development".into());  // ⭐ DEV MODE
    env.insert("RUST_LOG".into(), "debug".into());
    env.insert("HOST".into(), "0.0.0.0".into());
    env.insert("PORT".into(), "8081".into());
    env.insert("DATABASE_URL".into(),
        "postgresql://postgres:development@postgres-auth:5432/clikd_auth".into());
    env.insert("KEYDB_URL".into(), "redis://keydb:6379".into());
    env.insert("NATS_URL".into(), "nats://nats:4222".into());
    env.insert("OTEL_EXPORTER_OTLP_ENDPOINT".into(), "http://otel-collector:4317".into());
    env.insert("JWT_SECRET".into(), "dev-jwt-secret-32-bytes-long-enough-for-testing-abc123".into());
    env.insert("ENC_KEY_ACTIVE".into(), "gate1".into());
    env.insert("ENC_KEYS".into(), "gate1/MUKfFPL1zfhKfffX7usQbeWKd5L9iH65K4kCi7B3/KU=".into());
    env.insert("COOKIE_SECRET".into(), "dev-cookie-secret-32-bytes-long-enough-for-testing-def456".into());
    env.insert("INTERNAL_API_SECRET".into(), "dev-internal-api-secret-change-this".into());
    env.insert("PUBLIC_URL".into(), "http://localhost:8081".into());
    env.insert("ISSUER".into(), "http://localhost:8081".into());
    env.insert("RIG_INTERNAL_URL".into(), "http://rig:8082".into());
    env.insert("BACKEND_API_KEY".into(), "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into());
    env.insert("GATE_ANON_KEY".into(), "gt_publishable_dev_aNonymOusK3yForPubl1cAccess".into());
    env.insert("GATE_SECRET_KEY".into(), "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into());

    ServiceDefinition {
        name: "gate",
        image: "ghcr.io/clikd-inc/gate:0.1.0".into(),
        ports: vec![(8081, 8081), (9001, 9001)],
        env,
        volumes: vec![format!("clikd_gate_config_{}:/config", branch)],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "curl".into(), "-f".into(), "http://localhost:8081/health".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["postgres-auth", "keydb"],
        command: None,
    }
}

fn rig_service(branch: &str) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("APP_ENV".into(), "development".into());  // ⭐ DEV MODE
    env.insert("RUST_LOG".into(), "debug".into());
    env.insert("PORT".into(), "8082".into());
    env.insert("GRPC_PORT".into(), "9002".into());
    env.insert("DATABASE_URL".into(),
        "postgresql://postgres:development@postgres-rig:5432/clikd_rig".into());
    env.insert("KEYDB_URL".into(), "redis://keydb:6379".into());
    env.insert("SCYLLADB_HOSTS".into(), "scylladb:9042".into());
    env.insert("NATS_URL".into(), "nats://nats:4222".into());
    env.insert("MINIO_ENDPOINT".into(), "http://minio:9000".into());
    env.insert("MINIO_ROOT_USER".into(), "minioadmin".into());
    env.insert("MINIO_ROOT_PASSWORD".into(), "minioadmin".into());
    env.insert("OTEL_EXPORTER_OTLP_ENDPOINT".into(), "http://otel-collector:4317".into());

    ServiceDefinition {
        name: "rig",
        image: "ghcr.io/clikd-inc/rig:0.1.0".into(),
        ports: vec![(8082, 8082), (9002, 9002)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec!["CMD".into(), "curl".into(), "-f".into(), "http://localhost:8082/health".into()],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["postgres-rig", "keydb", "scylladb", "nats", "minio"],
        command: None,
    }
}

fn studio_service(branch: &str) -> ServiceDefinition {
    let mut env = HashMap::new();
    env.insert("NODE_ENV".into(), "development".into());  // ⭐ DEV MODE
    env.insert("APP_ENV".into(), "development".into());
    env.insert("CLIKD_URL".into(), "http://apisix:9080".into());
    env.insert("CLIKD_KEY".into(), "gt_secret_dev_S3rv1c3R0l3K3yForAdm1nAccess".into());
    env.insert("NEXT_PUBLIC_STUDIO_URL".into(), "http://localhost:3001".into());
    env.insert("NEXT_PUBLIC_APP_ENV".into(), "development".into());

    ServiceDefinition {
        name: "studio",
        image: "ghcr.io/clikd-inc/studio:0.1.0".into(),
        ports: vec![(3001, 3001)],
        env,
        volumes: vec![],
        health_check: Some(HealthCheck {
            test: vec![
                "CMD".into(),
                "bun".into(),
                "--eval".into(),
                "fetch('http://localhost:3001/api/health').then(r => process.exit(r.ok ? 0 : 1))".into()
            ],
            interval: Duration::from_secs(30),
            timeout: Duration::from_secs(10),
            retries: 3,
            start_period: None,
        }),
        depends_on: vec!["apisix"],
        command: None,
    }
}

// ... weitere Service-Definitionen für alle anderen Services
```

### 4. Start Command (src/cmd/start.rs)

```rust
use crate::cli::StartArgs;
use crate::core;
use anyhow::Result;
use bollard::Docker;
use owo_colors::OwoColorize;

pub async fn run(args: StartArgs) -> Result<()> {
    println!("{}", "Starting Clikd development environment...".cyan());

    // Connect to Docker
    let docker = Docker::connect_with_local_defaults()?;

    // Detect branch
    let branch = core::git::branch::current()?;
    println!("{} Detected branch: {}", "→".dimmed(), branch.yellow());

    // Pull images if requested
    if args.pull {
        println!("{}", "Pulling latest images...".cyan());
        // TODO: pull_images(&docker).await?;
    }

    // Start services
    core::start::runner::run(
        &docker,
        &branch,
        args.exclude.unwrap_or_default(),
    ).await?;

    println!("\n{}", "✓ All services started!".green().bold());
    println!("{} Run {} to see status", "→".dimmed(), "clikd status".yellow());

    Ok(())
}
```

---

## Testing Strategy (CLI Best Practices 2025)

### 1. Snapshot Tests (trycmd)

```toml
# tests/cmd/help.toml
bin.name = "clikd"
args = ["--help"]
status.code = 0
```

### 2. Integration Tests (assert_cmd)

```rust
// tests/integration/start_tests.rs
use assert_cmd::Command;

#[test]
fn test_start_help() {
    Command::cargo_bin("clikd")
        .unwrap()
        .arg("start")
        .arg("--help")
        .assert()
        .success();
}
```

---

## Success Criteria

### Code Quality

- ✅ Keine mod.rs files (Rust 2018+)
- ✅ CLI Standards 2025 (clap, owo-colors, indicatif, tracing)
- ✅ Error handling: anyhow (main) + thiserror (lib)
- ✅ Binary < 10 MB (release)
- ✅ Startup < 100ms

### Docker Integration

- ✅ Alle Services aus docker-compose.yml
- ✅ Development Mode (APP_ENV=development) für gate, rig, studio
- ✅ Bollard direkt (wie Supabase)
- ✅ Branch-isolierte Container & Volumes
- ✅ Dependency-aware startup order
- ✅ Health checks mit indicatif Progress bars

### Auth

- ✅ GitHub OAuth Device Flow
- ✅ Token in Keyring (cross-platform)
- ✅ Organization membership check

### CLI UX

- ✅ owo-colors für styling
- ✅ Progress bars (indicatif)
- ✅ Shell completions (clap_complete)
- ✅ Verbosity levels (-v, -vv, -vvv)
- ✅ --no-color support

---

## Implementierungs-Reihenfolge

### Phase 1: Core Foundation
1. **Grundgerüst** (Folder, Cargo.toml, main.rs, cli.rs, error.rs)
2. **Config-System** (Layered loading mit config crate)
3. **Utils** (terminal.rs mit owo-colors, retry.rs)
4. **Git-Integration** (Branch detection)
5. **Service-Definitionen** (Alle Services aus docker-compose.yml)

### Phase 2: Docker & Orchestration
6. **Auth** (Device Flow, Keyring, Org check) ✅ DONE
7. **Docker Registry Auth** (GHCR authentication)
8. **Docker Manager** (Bollard container creation, image pull)
9. **Start Orchestration** (Dependency-aware, health checks)
10. **Commands** (start, stop, status, logs)

### Phase 3: Polish & Advanced Features
11. **Shell Completions** (clap_complete)
12. **Testing** (trycmd + assert_cmd)
13. **ratatui Dashboard** (`clikd dashboard` oder `clikd status --watch`)

---

## 📊 ratatui Dashboard Feature

### Command: `clikd dashboard` oder `clikd status --watch`

**Interactive TUI Dashboard** für Live-Monitoring aller Services.

#### Features

**Core:**
- ✅ Real-time service status (Echtzeit-Updates alle 2s)
- ✅ Per-container CPU & Memory metrics
- ✅ Health status mit Color-Coding (green/yellow/red)
- ✅ Dependency tree visualization
- ✅ Live log preview für selected service

**Keyboard Navigation:**
- `↑/↓` - Service selection
- `Enter` - Show detailed logs
- `r` - Restart selected service
- `s` - Stop selected service
- `l` - Toggle live logs
- `q` - Quit dashboard

**Layout (Split View):**
```
┌─ Services ──────────────────────────────┬─ Logs (gate) ────────────────┐
│ ✓ gate          [healthy]  CPU: 2%     │ [INFO] Server started        │
│ ✓ rig           [healthy]  CPU: 5%     │ [DEBUG] Database connected   │
│ ⚠ postgres-auth [starting] CPU: 1%     │ [INFO] Listening on :8081    │
│ ✗ studio        [error]    CPU: 0%     │                              │
│ • minio         [stopped]              │                              │
├────────────────────────────────────────┴──────────────────────────────┤
│ [↑↓] Navigate | [Enter] Logs | [r] Restart | [s] Stop | [q] Quit     │
└──────────────────────────────────────────────────────────────────────┘
```

#### Implementation Structure

```rust
// src/cmd/dashboard.rs
pub async fn run() -> Result<()> {
    let mut terminal = setup_terminal()?;
    let mut app = DashboardApp::new().await?;

    loop {
        terminal.draw(|f| ui::render(f, &mut app))?;

        if let Some(event) = poll_event()? {
            if handle_event(&mut app, event).await? {
                break; // quit
            }
        }

        app.update().await?; // Poll Docker API
    }

    cleanup_terminal(terminal)?;
    Ok(())
}

// src/core/dashboard/
//   ├── app.rs          # Application state
//   ├── ui.rs           # ratatui rendering
//   ├── events.rs       # Keyboard event handling
//   └── docker_stats.rs # Docker metrics polling
```

**Tech Stack:**
- `ratatui` - TUI framework
- `crossterm` - Terminal handling
- `bollard` - Docker API (stats streaming)
- `tokio::sync::mpsc` - Event channels

**Inspiration:**
- `k9s` (Kubernetes dashboard)
- `lazydocker` (Docker TUI)
- `bottom` (System monitoring)
