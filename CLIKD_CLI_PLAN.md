# Clikd Development CLI - Production Implementation Plan

_Complete implementation guide following Rust CLI Best Practices 2025_

---

## Overview

A professional, terminal-based development tool for the Clikd gaming social platform. Orchestrates 4 core microservices plus studio dashboard across branch-isolated environments with automated service management, database operations, and deployment workflows.

## Architecture Principles

### Design Philosophy
- **Hybrid execution model**: Direct commands for automation, interactive selectors for exploration
- **Professional tooling**: Technical precision over visual flourish  
- **Standards-compliant**: Follows Rust CLI Best Practices 2025
- **Production-ready**: Security, testing, and performance from day one

### Core Services (Monorepo)
```
clikd-monorepo/
├── services/
│   ├── auth/              # Rust (Axum) - :3001/:9001
│   ├── api/               # Rust (Axum) - :3002/:9002
│   ├── realtime/          # Elixir (Phoenix) - :3003/:9003
│   └── media/             # Rust (FFmpeg) - :3004/:9004
├── studio/                # Next.js Dashboard - :3000
├── cli/                   # This tool
├── clients/               # Generated SDK clients
└── k8s/                   # Kubernetes manifests
```

### Database Architecture (Per Branch)
- **PostgreSQL**: `clikd_auth_{branch}`, `clikd_rig_{branch}`
- **ScyllaDB**: Keyspace `clikd_{branch}`  
- **KeyDB**: Database 0 with `clikd_{branch}:*` prefixes

---

## Technical Stack

### Core Dependencies
```toml
[dependencies]
# CLI Framework
clap = { version = "4.5", features = ["derive", "env", "wrap_help"] }
clap_complete = "4.5"

# Error Handling
anyhow = "1.0"
thiserror = "1.0"

# Logging & Observability  
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }

# Terminal UI
owo-colors = { version = "4.2", features = ["supports-colors"] }
indicatif = "0.17"
dialoguer = "0.11"

# Optional: Full TUI mode
ratatui = { version = "0.29", optional = true }
crossterm = { version = "0.27", optional = true }

# Configuration - Layered support
config = "0.14"
serde = { version = "1.0", features = ["derive"] }

# Docker & Git
bollard = "0.17"
git2 = "0.19"

# Security
secrecy = "0.8"
zeroize = "1.8"

[profile.release]
opt-level = "z"
lto = true
codegen-units = 1
strip = true
panic = "abort"
```

---

## Configuration Management (config-rs)

### Layered Configuration Priority
1. **Environment Variables** (highest): `CLIKD_*`
2. **Local Overrides**: `config/local.toml` (gitignored)
3. **Environment-Specific**: `config/{env}.toml`
4. **Default** (lowest): `config/default.toml`

### Example Configuration

**config/default.toml**
```toml
[project]
name = "clikd"
monorepo_root = "../"

[registry]
url = "ghcr.io"
organization = "clikd-org"

[services.auth]
image = "ghcr.io/clikd-org/auth-service"
port = 3001
grpc_port = 9001
health_check_path = "/health"

[databases.postgresql]
host = "localhost"
port = 5432
user = "postgres"

[development]
auto_migrate = true
log_level = "debug"
```

**config/production.toml**
```toml
[databases.postgresql]
host = "prod-postgres.clikd.internal"
max_connections = 50

[development]
auto_migrate = false
log_level = "warn"
```

### Configuration Code

**src/config.rs**
```rust
use config::{Config as ConfigBuilder, Environment, File};

impl Config {
    pub fn load() -> Result<Self> {
        let env = std::env::var("CLIKD_ENV")
            .unwrap_or_else(|_| "development".into());
        
        ConfigBuilder::builder()
            .add_source(File::with_name("config/default"))
            .add_source(File::with_name(&format!("config/{}", env)).required(false))
            .add_source(File::with_name("config/local").required(false))
            .add_source(Environment::with_prefix("CLIKD").separator("__"))
            .build()?
            .try_deserialize()
    }
}
```

---

## Command Interface

### Entry Point (20-30 lines)

**src/main.rs**
```rust
use anyhow::Result;
use clap::Parser;

#[tokio::main]
async fn main() -> Result<()> {
    setup_logging()?;
    
    let cli = cli::Cli::parse();
    let config = config::Config::load()?;
    
    commands::execute(cli, config).await
}
```

### CLI Definitions

**src/cli.rs**
```rust
#[derive(Parser)]
#[command(name = "clikd", about = "Development CLI for Clikd platform")]
pub struct Cli {
    #[arg(short, long, env = "CLIKD_CONFIG")]
    pub config: Option<PathBuf>,

    #[arg(short, long, env = "CLIKD_ENV", default_value = "development")]
    pub env: String,

    #[arg(short, long, action = ArgAction::Count, global = true)]
    pub verbose: u8,

    #[arg(long, global = true)]
    pub no_color: bool,

    #[arg(long, global = true)]
    pub no_interactive: bool,

    #[command(subcommand)]
    pub command: Option<Commands>,
}

#[derive(Subcommand)]
pub enum Commands {
    Start {
        #[arg(long, value_delimiter = ',')]
        exclude: Option<Vec<String>>,
        
        #[arg(long)]
        pull: bool,
    },
    
    Stop {
        #[arg(short, long)]
        force: bool,
    },
    
    Status {
        #[arg(short, long, value_enum, default_value = "text")]
        format: OutputFormat,
    },
    
    Db { #[command(subcommand)] command: DbCommands },
    Gen { #[command(subcommand)] command: GenCommands },
    Deploy { environment: Environment, #[arg(short, long)] yes: bool },
    
    #[cfg(feature = "tui")]
    Tui,
    
    Completions { shell: Shell },
    Config { #[arg(long)] show_files: bool },
}
```

---

## Core Implementation

### Service Orchestration

**src/core/docker.rs**
```rust
use indicatif::{ProgressBar, MultiProgress};

pub struct ServiceManager {
    docker: Docker,
    branch: String,
    config: Config,
}

impl ServiceManager {
    pub async fn start_all(&self, exclude: Option<Vec<String>>) -> Result<()> {
        let services = self.config.service_names()
            .into_iter()
            .filter(|s| !exclude.as_ref().map_or(false, |e| e.contains(s)))
            .collect::<Vec<_>>();
        
        let multi = MultiProgress::new();
        let overall = multi.add(ProgressBar::new(services.len() as u64));
        
        for service in services {
            let pb = multi.add(ProgressBar::new_spinner());
            pb.set_message(format!("Starting {}", service));
            
            self.start_service(&service).await?;
            
            if self.health_check(&service).await? {
                pb.finish_with_message(format!("✓ {} started", service));
            }
            
            overall.inc(1);
        }
        
        Ok(())
    }
}
```

### Git Integration

**src/core/git.rs**
```rust
pub struct GitManager {
    repo: Repository,
}

impl GitManager {
    pub fn current_branch(&self) -> Result<String> {
        let head = self.repo.head()?;
        let branch = head.shorthand()
            .ok_or_else(|| git2::Error::from_str("No branch"))?;
        Ok(branch.to_string())
    }

    pub fn sanitize_branch_name(branch: &str) -> String {
        branch.replace('/', "_")
            .replace('-', "_")
            .to_lowercase()
    }
}
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [ ] Cargo project setup with all dependencies
- [ ] Layered config system (config-rs)
- [ ] Error handling (anyhow + thiserror)
- [ ] Logging (tracing)
- [ ] main.rs (20-30 lines)
- [ ] CLI definitions (clap)
- [ ] Testing infrastructure (trycmd, assert_cmd)
- [ ] GitHub Actions CI

### Phase 2: Core Commands (Week 2)
- [ ] Git integration
- [ ] Docker integration (Bollard)
- [ ] `start` command with progress bars
- [ ] `stop` command with confirmation
- [ ] `status` command
- [ ] Health checks

### Phase 3: Database Operations (Week 3)
- [ ] Database manager
- [ ] PostgreSQL/ScyllaDB/KeyDB operations
- [ ] `db migrate/reset/seed/diff/backup`

### Phase 4: Interactive Features (Week 4)
- [ ] Command selector (dialoguer)
- [ ] Color control (owo-colors)
- [ ] `--no-interactive` support

### Phase 5: Registry (Week 5)
- [ ] GitHub Container Registry auth
- [ ] Image pulling

### Phase 6: Logging (Week 6)
- [ ] `logs` command
- [ ] Multi-service aggregation
- [ ] Follow mode

### Phase 7: Code Generation (Week 7)
- [ ] OpenAPI fetching
- [ ] Swift/Kotlin/TypeScript generators
- [ ] `gen` commands

### Phase 8: Deployment (Week 8)
- [ ] Kubernetes integration
- [ ] `deploy` command
- [ ] Pre-flight checks

### Phase 9: Optional TUI (Week 9)
- [ ] Full TUI dashboard (ratatui)
- [ ] Feature flag: `--features tui`

### Phase 10: Polish (Week 10)
- [ ] Binary optimization
- [ ] Security audit
- [ ] Documentation
- [ ] Cross-compilation
- [ ] Distribution

---

## Testing Strategy

### Integration Tests
```rust
// tests/cli/start_tests.rs
#[test]
fn test_start_help() {
    Command::cargo_bin("clikd").unwrap()
        .arg("start").arg("--help")
        .assert().success();
}
```

### Snapshot Tests  
```toml
# tests/cmd/help.toml
bin.name = "clikd"
args = ["--help"]
status.code = 0
```

---

## Usage Examples

### Basic Workflow
```bash
# Start all services for current branch
clikd start

# Check status
clikd status

# View logs
clikd logs --service=api --follow

# Database operations
clikd db migrate

# Generate clients
clikd gen all

# Deploy to staging
clikd deploy staging

# Stop services
clikd stop
```

### Interactive Mode
```bash
# No arguments shows menu
clikd
```

### CI/CD Mode
```bash
export CLIKD_ENV=staging
clikd start --no-interactive
clikd deploy staging --yes
```

### Configuration
```bash
# Show current config
clikd config

# Show config sources
clikd config --show-files

# Override via environment
export CLIKD_ENV=production
export CLIKD_DATABASE__POSTGRESQL__PASSWORD=secret
clikd status
```

---

## Go-Live Checklist

### Code Quality
- [ ] All tests pass
- [ ] No clippy warnings  
- [ ] Code formatted
- [ ] Documentation complete
- [ ] No `todo!()`

### Security
- [ ] `cargo audit` passes
- [ ] No secrets in code
- [ ] Input validation
- [ ] Secrets use `secrecy`

### Distribution
- [ ] Cross-compilation working
- [ ] Binary size < 5 MB
- [ ] Shell completions
- [ ] CI/CD pipeline

---

## Success Metrics

### Performance
- CLI startup: < 100ms
- Service start: < 60s (all)
- Binary size: < 5 MB
- Memory usage: < 50 MB

### Reliability
- Branch detection: 100%
- Service start: 99%+
- Test coverage: > 80%

### Adoption
- Daily usage: 100% of developers
- Onboarding time: < 1 hour
- Manual operations: 0/week

---

## Key Differentiators

**Standards 2025:**
- config-rs for layered configuration
- anyhow/thiserror for errors
- tracing for logging
- owo-colors for terminal
- indicatif for progress
- dialoguer for prompts

**Architecture:**
- 20-30 line main.rs
- Command trait pattern
- No mod.rs files
- Hybrid execution model

**Production-Ready:**
- Security from day one
- Comprehensive testing
- Cross-platform distribution
- Professional error handling

---

## Next Steps

1. **Review** this plan with team
2. **Initialize** cli/ directory in monorepo
3. **Set up** GitHub Actions workflows
4. **Begin** Phase 1 implementation
5. **Review** after Phase 2
6. **Production** rollout after Phase 10

---

_This CLI will transform Clikd development: 15 minutes → 2 minutes environment startup_
