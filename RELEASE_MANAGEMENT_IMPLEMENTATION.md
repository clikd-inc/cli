# Release Management Integration - Implementation Plan

> Integriert alle Cranko Features (außer Zenodo) + Go + Elixir Support in die clikd CLI

**Branch:** `feat/release-management`
**Start:** 2025-01-21
**Geschätzte Dauer:** 10-12 Tage

---

## 📋 Gesamtübersicht

### Ziel
Vollständige Integration von Cranko's Release-Management-System in die clikd CLI mit zusätzlichem Go und Elixir Support.

### Strategie
**Code-Reuse statt Neuschreiben:**
- Kopieren von funktionierendem Cranko-Code
- Anpassen an clikd Namenskonventionen
- Erweitern mit Go + Elixir Loaders
- Integration in bestehende CLI-Struktur

### Sprach-Support
- ✅ Rust (Cargo.toml)
- ✅ NPM/Node (package.json)
- ✅ Python (pyproject.toml, setup.py)
- ✅ C# (.csproj) [optional]
- 🆕 **Go (go.mod)**
- 🆕 **Elixir (mix.exs)**

---

## 📂 File Mapping: Cranko → CLI

### Core Infrastructure

| Cranko Source | CLI Destination | Änderungen |
|---------------|----------------|------------|
| `src/version.rs` | `src/core/release/version.rs` | `.cranko` → `.clikd` paths |
| `src/errors.rs` | `src/core/release/errors.rs` | Merge mit bestehendem error.rs |
| `src/config.rs` | `src/core/release/config.rs` | Config dir: `.cranko/` → `.clikd/` |
| `src/repository.rs` | `src/core/release/repository.rs` | Paths anpassen |
| `src/graph.rs` | `src/core/release/graph.rs` | Minimal changes |
| `src/project.rs` | `src/core/release/project.rs` | Minimal changes |
| `src/rewriters.rs` | `src/core/release/rewriters.rs` | Trait definition |

### Ecosystem Loaders

| Cranko Source | CLI Destination | Änderungen |
|---------------|----------------|------------|
| `src/cargo.rs` | `src/core/ecosystem/cargo.rs` | Config paths |
| `src/npm.rs` | `src/core/ecosystem/npm.rs` | Config paths |
| `src/pypa.rs` | `src/core/ecosystem/pypa.rs` | Config paths |
| `src/csproj.rs` | `src/core/ecosystem/csproj.rs` | Optional, config paths |
| `src/changelog.rs` | `src/core/release/changelog.rs` | Path adjustments |
| **NEU** | `src/core/ecosystem/go.rs` | Von Grund auf (basierend auf cargo.rs) |
| **NEU** | `src/core/ecosystem/elixir.rs` | Von Grund auf (basierend auf npm.rs) |

### Workflow & Commands

| Cranko Source | CLI Destination | Änderungen |
|---------------|----------------|------------|
| `src/app.rs` | `src/core/release/session.rs` | Session management |
| `src/bootstrap.rs` | `src/cmd/release/init.rs` | Bootstrap → init command |
| `src/main.rs` (stage) | `src/cmd/release/stage.rs` | Stage command |
| `src/main.rs` (confirm) | `src/cmd/release/confirm.rs` | Confirm command |
| `src/main.rs` (release-workflow) | `src/cmd/release/workflow.rs` | Apply/commit/tag |
| `src/main.rs` (show) | `src/cmd/release/show.rs` | Info commands |
| `src/main.rs` (status) | `src/cmd/release/status.rs` | Status command |

### GitHub Integration

| Cranko Source | CLI Destination | Änderungen |
|---------------|----------------|------------|
| `src/github.rs` | `src/core/github/release.rs` | API client |
| `src/gitutil.rs` | `src/core/git/utils.rs` | Merge mit bestehendem git/ |

### Utilities

| Cranko Source | CLI Destination | Notiz |
|---------------|----------------|-------|
| `src/logger.rs` | ❌ Skip | Nutzen bestehendes tracing |
| `src/env.rs` | `src/core/release/env.rs` | Kleine helpers |

---

## 🏗️ Neue Directory Struktur

```
src/
├── cli.rs                          # Erweitert um release commands
├── cmd/
│   ├── auth.rs                     # Existing
│   ├── start.rs                    # Existing
│   ├── stop.rs                     # Existing
│   ├── status.rs                   # Existing (container status)
│   └── release/                    # 🆕 NEW
│       ├── mod.rs                  # Release subcommand routing
│       ├── init.rs                 # From bootstrap.rs
│       ├── stage.rs                # From main.rs
│       ├── confirm.rs              # From main.rs
│       ├── status.rs               # From main.rs
│       ├── show.rs                 # From main.rs
│       └── workflow.rs             # From main.rs (apply/commit/tag)
├── core/
│   ├── auth/                       # Existing
│   ├── docker/                     # Existing
│   ├── git/                        # Existing
│   │   └── utils.rs                # Merge gitutil.rs hier
│   ├── release/                    # 🆕 NEW (Cranko core)
│   │   ├── mod.rs
│   │   ├── version.rs              # Version enum + bumping
│   │   ├── errors.rs               # Release errors
│   │   ├── config.rs               # .clikd/config.toml
│   │   ├── repository.rs           # Git operations
│   │   ├── graph.rs                # Dependency graph
│   │   ├── project.rs              # Project abstraction
│   │   ├── rewriters.rs            # Rewriter trait
│   │   ├── changelog.rs            # Changelog generation
│   │   ├── session.rs              # App session (from app.rs)
│   │   └── env.rs                  # Environment helpers
│   ├── ecosystem/                  # 🆕 NEW (Language loaders)
│   │   ├── mod.rs
│   │   ├── cargo.rs                # Rust/Cargo loader
│   │   ├── npm.rs                  # NPM/Node loader
│   │   ├── pypa.rs                 # Python loader
│   │   ├── csproj.rs               # C# loader (optional)
│   │   ├── go.rs                   # 🆕 Go loader
│   │   └── elixir.rs               # 🆕 Elixir loader
│   └── github/                     # 🆕 NEW
│       ├── mod.rs
│       └── release.rs              # GitHub API (from github.rs)
└── utils/                          # Existing
    ├── terminal.rs
    └── ...
```

---

## 📦 Cargo.toml Dependencies (zu ergänzen)

```toml
# Bestehende (behalten)
git2 = "0.20.2"                    # ✅ Bereits vorhanden
semver = "1.0"                     # ✅ Bereits vorhanden
serde = { version = "1.0", features = ["derive"] }  # ✅ Vorhanden
serde_json = "1.0"                 # ✅ Vorhanden
toml = "0.9.8"                     # ✅ Vorhanden (upgrade von 0.8)
anyhow = "1.0"                     # ✅ Vorhanden
thiserror = "2.0"                  # ✅ Vorhanden (upgrade von 1.0)
chrono = "0.4"                     # ✅ Vorhanden
reqwest = { version = "0.12", features = ["json", "rustls-tls"] }  # ✅ Vorhanden

# Neu hinzufügen
cargo_metadata = "0.18"            # 🆕 Für Cargo workspace analysis
toml_edit = "0.22"                 # 🆕 Format-preserving TOML edits
petgraph = "0.6"                   # 🆕 Dependency graph
atomicwrites = "0.4"               # 🆕 Atomic file writes
nom = "7"                          # 🆕 Parser combinators (für PEP440)
quick-xml = "0.36"                 # 🆕 Für .csproj parsing
json5 = "0.4"                      # 🆕 Für .npmrc / tsconfig.json
lru = "0.12"                       # 🆕 LRU cache
git-url-parse = "0.4"              # 🆕 Git URL parsing
percent-encoding = "2"             # 🆕 URL encoding
target-lexicon = "0.12"            # 🆕 Target platform detection
terminal_size = "0.3"              # ✅ Vorhanden (für textwrap)
textwrap = "0.16"                  # 🆕 Text wrapping
base64 = "0.22"                    # 🆕 Base64 encoding
flate2 = "1.0"                     # 🆕 Compression (für tar.gz)
tar = "0.4"                        # 🆕 Tar archive handling
zip = { version = "2.2", default-features = false, features = ["deflate", "time"] }  # 🆕 Zip handling

# Optional (C# support)
# quick-xml bereits oben erwähnt

# Development
trycmd = "0.15"                    # ✅ Vorhanden
assert_cmd = "2.0"                 # ✅ Vorhanden
assert_fs = "1.1"                  # ✅ Vorhanden
tempfile = "3.8"                   # ✅ Vorhanden
```

---

## 🎯 Phase-by-Phase Implementation

### **Phase 1: Foundation Setup** ⏱️ 0.5 Tag

**Ziel:** Verzeichnisstruktur anlegen und Dateien kopieren

**Tasks:**
1. ✅ Create feature branch `feat/release-management`
2. 📁 Create directory structure:
   ```bash
   mkdir -p src/cmd/release
   mkdir -p src/core/release
   mkdir -p src/core/ecosystem
   mkdir -p src/core/github
   ```
3. 📋 Copy Cranko source files:
   ```bash
   # Core files
   cp sources-original/cranko-master/src/version.rs src/core/release/
   cp sources-original/cranko-master/src/errors.rs src/core/release/
   cp sources-original/cranko-master/src/config.rs src/core/release/
   cp sources-original/cranko-master/src/repository.rs src/core/release/
   cp sources-original/cranko-master/src/graph.rs src/core/release/
   cp sources-original/cranko-master/src/project.rs src/core/release/
   cp sources-original/cranko-master/src/rewriters.rs src/core/release/
   cp sources-original/cranko-master/src/changelog.rs src/core/release/
   cp sources-original/cranko-master/src/app.rs src/core/release/session.rs
   cp sources-original/cranko-master/src/env.rs src/core/release/

   # Loaders
   cp sources-original/cranko-master/src/cargo.rs src/core/ecosystem/
   cp sources-original/cranko-master/src/npm.rs src/core/ecosystem/
   cp sources-original/cranko-master/src/pypa.rs src/core/ecosystem/
   cp sources-original/cranko-master/src/csproj.rs src/core/ecosystem/

   # GitHub
   cp sources-original/cranko-master/src/github.rs src/core/github/release.rs
   cp sources-original/cranko-master/src/gitutil.rs src/core/git/utils.rs

   # Bootstrap
   cp sources-original/cranko-master/src/bootstrap.rs src/cmd/release/init.rs
   ```
4. 📦 Update `Cargo.toml` mit neuen dependencies
5. ✅ Commit: `chore: copy cranko source files for release management`

**Erfolg:** Alle Files kopiert, Projekt kompiliert NICHT (expected)

---

### **Phase 2: Core Infrastructure** ⏱️ 2 Tage

**Ziel:** Basis-Layer zum Laufen bringen

#### 2.1 Version Management (0.5 Tag)

**File:** `src/core/release/version.rs`

**Änderungen:**
- Import paths anpassen
- Tests aktivieren
- PEP440 parser testen

**Tests:**
```rust
#[test]
fn test_semver_bump() {
    let mut v = Version::Semver(semver::Version::new(1, 2, 3));
    v.bump(BumpScheme::Minor).unwrap();
    assert_eq!(v.to_string(), "1.3.0");
}
```

**Commit:** `feat(release): add version management with semver/pep440 support`

#### 2.2 Error Handling (0.25 Tag)

**File:** `src/core/release/errors.rs`

**Änderungen:**
- Merge mit bestehendem `src/error.rs` wenn nötig
- Oder als separate `ReleaseError` types

**Commit:** `feat(release): add release-specific error types`

#### 2.3 Configuration (0.5 Tag)

**File:** `src/core/release/config.rs`

**Änderungen:**
- `.cranko/` → `.clikd/`
- Config file path resolution
- Default values anpassen

**Config Format:**
```toml
# .clikd/config.toml
[repo]
upstream_urls = ["git@github.com:clikd-inc/clikd.git"]
rc_name = "rc"
release_name = "release"
release_tag_name_format = "{project_slug}@{version}"

[rust]
workspace_mode = "independent"

[npm]
internal_dep_protocol = "workspace"

[go]
tag_prefix = "v"

[elixir]
hex_organization = "clikd"

[projects.some-project]
ignore = true
```

**Tests:**
- Config parsing
- Default values
- Missing file handling

**Commit:** `feat(release): add .clikd/config.toml support`

#### 2.4 Module Organization (0.25 Tag)

**Files:**
- `src/core/release/mod.rs`
- `src/core/ecosystem/mod.rs`
- `src/core/github/mod.rs`

**Expose Public API:**
```rust
// src/core/release/mod.rs
pub mod version;
pub mod config;
pub mod errors;
pub mod repository;
pub mod graph;
pub mod project;
pub mod rewriters;
pub mod changelog;
pub mod session;

pub use version::{Version, VersionBumpScheme};
pub use config::ConfigurationFile;
pub use errors::ReleaseError;
```

**Commit:** `chore(release): organize module structure`

#### 2.5 Integration Check (0.5 Tag)

**Ziel:** Core modules kompilieren

```bash
cargo check
cargo test --package clikd --lib core::release
```

**Fix:** Alle Compile errors beheben

**Commit:** `fix(release): resolve compilation errors in core modules`

---

### **Phase 3: Repository Layer** ⏱️ 1.5 Tage

**Ziel:** Git operations und Dependency Graph

#### 3.1 Repository Module (0.75 Tag)

**File:** `src/core/release/repository.rs`

**Änderungen:**
- Paths: `.cranko/` → `.clikd/`
- Integrate mit bestehendem `src/core/git/`
- Git operations testen

**Key Functions:**
```rust
impl Repository {
    pub fn open_from_env() -> Result<Self>;
    pub fn scan_paths(&self, callback: F) -> Result<()>;
    pub fn get_latest_release_info() -> Result<ReleaseCommitInfo>;
    pub fn create_tag(&self, name: &str) -> Result<()>;
}
```

**Tests:**
- Open repository
- Scan git index
- Tag creation

**Commit:** `feat(release): add git repository operations layer`

#### 3.2 Dependency Graph (0.5 Tag)

**File:** `src/core/release/graph.rs`

**Änderungen:**
- Minimal (petgraph dependency)

**Tests:**
- Add projects
- Detect cycles
- Topological sort

**Commit:** `feat(release): add project dependency graph`

#### 3.3 Project Abstraction (0.25 Tag)

**File:** `src/core/release/project.rs`

**Änderungen:**
- Path handling anpassen

**Commit:** `feat(release): add project abstraction layer`

---

### **Phase 4: Ecosystem Loaders** ⏱️ 2 Tage

**Ziel:** Rust, NPM, Python Support

#### 4.1 Cargo Loader (0.5 Tag)

**File:** `src/core/ecosystem/cargo.rs`

**Änderungen:**
- Config paths
- Tests mit Fixture Cargo.toml

**Tests:**
```rust
#[test]
fn test_cargo_detection() {
    // Create temp Cargo.toml
    // Run loader
    // Assert project detected
}
```

**Commit:** `feat(ecosystem): add Cargo/Rust loader`

#### 4.2 NPM Loader (0.5 Tag)

**File:** `src/core/ecosystem/npm.rs`

**Änderungen:**
- Config paths
- Workspace protocol handling

**Tests:**
- package.json detection
- Version extraction
- Dependency tracking

**Commit:** `feat(ecosystem): add NPM/JavaScript loader`

#### 4.3 Python Loader (0.5 Tag)

**File:** `src/core/ecosystem/pypa.rs`

**Änderungen:**
- Config paths
- Comment marker detection

**Tests:**
- pyproject.toml detection
- setup.py detection
- Version extraction

**Commit:** `feat(ecosystem): add Python/PyPA loader`

#### 4.4 Changelog Generator (0.5 Tag)

**File:** `src/core/release/changelog.rs`

**Änderungen:**
- Template system
- Format customization

**Tests:**
- Changelog creation
- RC format
- Finalization

**Commit:** `feat(release): add changelog generation`

---

### **Phase 5: Release Workflow** ⏱️ 2 Tage

**Ziel:** stage → confirm → apply → commit → tag workflow

#### 5.1 Session Management (0.5 Tag)

**File:** `src/core/release/session.rs` (from app.rs)

**Änderungen:**
- Remove Cranko-specific CLI dependencies
- Clean API for commands

**Commit:** `feat(release): add session management`

#### 5.2 Init Command (0.5 Tag)

**File:** `src/cmd/release/init.rs` (from bootstrap.rs)

**Änderungen:**
- Command structure für clap
- Create `.clikd/config.toml`
- Detect projects

**Usage:**
```bash
clikd release init
```

**Commit:** `feat(release): add release init command`

#### 5.3 Status Command (0.25 Tag)

**File:** `src/cmd/release/status.rs`

**Commit:** `feat(release): add release status command`

#### 5.4 Stage Command (0.25 Tag)

**File:** `src/cmd/release/stage.rs`

**Usage:**
```bash
clikd release stage my-project
clikd release stage --all
```

**Commit:** `feat(release): add release stage command`

#### 5.5 Confirm Command (0.25 Tag)

**File:** `src/cmd/release/confirm.rs`

**Commit:** `feat(release): add release confirm command`

#### 5.6 Workflow Commands (0.25 Tag)

**File:** `src/cmd/release/workflow.rs`

**Commands:**
- `clikd release apply` (CI only)
- `clikd release commit` (CI only)
- `clikd release tag` (CI only)

**Commit:** `feat(release): add workflow commands for CI`

---

### **Phase 6: GitHub Integration** ⏱️ 1 Tag

**Ziel:** GitHub Releases + Artifact Upload

#### 6.1 GitHub API Client (0.5 Tag)

**File:** `src/core/github/release.rs`

**Änderungen:**
- Reqwest client setup
- Token handling (GITHUB_TOKEN env var)

**Commit:** `feat(github): add GitHub API client`

#### 6.2 Release Creation (0.25 Tag)

**Command:** `clikd github create-releases`

**Commit:** `feat(github): add GitHub release creation`

#### 6.3 Artifact Upload (0.25 Tag)

**Command:** `clikd github upload <tag> <files...>`

**Commit:** `feat(github): add artifact upload to releases`

---

### **Phase 7: New Ecosystem Loaders** ⏱️ 2 Tage

**Ziel:** Go + Elixir Support

#### 7.1 Go Loader (1 Tag)

**File:** `src/core/ecosystem/go.rs`

**Basierend auf:** `cargo.rs` (ähnliche Struktur)

**Detection:**
```rust
fn detection_file(&self) -> &str {
    "go.mod"
}
```

**Version Extraction:**
```rust
// go.mod hat keine Version field
// Version kommt von Git tags: v1.2.3
fn extract_version(&self, repo: &Repository) -> Result<Version> {
    // Find latest vX.Y.Z tag
    // Parse as semver
}
```

**Rewriter:**
```rust
// go.mod selbst hat keine version
// Aber go.mod dependencies müssen ggf. updated werden
impl Rewriter for GoRewriter {
    fn rewrite(&self, session: &Session, changes: &mut ChangeList) -> Result<()> {
        // Update replace directives für internal modules
        // Minimal intervention
    }
}
```

**Tests:**
- go.mod detection
- Tag-based versioning
- Multi-module repos

**Commit:** `feat(ecosystem): add Go module loader`

#### 7.2 Elixir Loader (1 Tag)

**File:** `src/core/ecosystem/elixir.rs`

**Basierend auf:** `npm.rs` (ähnliche Version in File)

**Detection:**
```rust
fn detection_file(&self) -> &str {
    "mix.exs"
}
```

**Version Extraction:**
```rust
// Parse Elixir code
fn extract_version(&self, path: &Path) -> Result<Version> {
    let content = fs::read_to_string(path)?;

    // Regex: version: "1.2.3"
    let re = Regex::new(r#"version:\s*"([^"]+)""#)?;
    let caps = re.captures(&content)?;
    let version_str = &caps[1];

    Version::parse_semver(version_str)
}
```

**Rewriter:**
```rust
impl Rewriter for ElixirRewriter {
    fn rewrite(&self, session: &Session, changes: &mut ChangeList) -> Result<()> {
        // Read mix.exs
        // Replace version: "old" with version: "new"
        // Update internal deps
    }
}
```

**Tests:**
- mix.exs detection
- Version extraction
- Hex package simulation
- Umbrella app support

**Commit:** `feat(ecosystem): add Elixir/Mix loader`

---

### **Phase 8: CLI Integration** ⏱️ 1.5 Tage

**Ziel:** Commands in CLI integrieren, Tests, Docs

#### 8.1 Command Integration (0.5 Tag)

**File:** `src/cli.rs`

**Änderungen:**
```rust
#[derive(Subcommand)]
pub enum Commands {
    // Existing commands
    Login { ... },
    Start(StartArgs),
    Stop(StopArgs),
    Status(StatusArgs),

    // NEW: Release management
    #[command(subcommand, about = "Release management commands")]
    Release(ReleaseCommands),

    // NEW: GitHub integration
    #[command(subcommand, about = "GitHub integration")]
    Github(GithubCommands),

    // NEW: Ecosystem-specific
    #[command(subcommand, about = "Cargo/Rust commands")]
    Cargo(CargoCommands),

    #[command(subcommand, about = "NPM/JavaScript commands")]
    Npm(NpmCommands),

    #[command(subcommand, about = "Go commands")]
    Go(GoCommands),

    #[command(subcommand, about = "Elixir commands")]
    Elixir(ElixirCommands),
}

#[derive(Subcommand)]
pub enum ReleaseCommands {
    Init(release::InitArgs),
    Status(release::StatusArgs),
    Stage(release::StageArgs),
    Confirm(release::ConfirmArgs),
    Apply(release::ApplyArgs),
    Commit(release::CommitArgs),
    Tag(release::TagArgs),
}
```

**Commit:** `feat(cli): integrate release commands into CLI structure`

#### 8.2 Foreach-Released Commands (0.5 Tag)

**Commands:**
```bash
clikd cargo foreach-released -- cargo publish
clikd npm foreach-released -- npm publish
clikd go foreach-released -- go build
clikd elixir foreach-released -- mix hex.publish
```

**Implementation:**
```rust
// src/cmd/ecosystem/cargo.rs
pub fn foreach_released(session: &Session, args: &[String]) -> Result<()> {
    for project in session.released_projects() {
        if project.is_cargo_project() {
            let workdir = project.path();
            Command::new(&args[0])
                .args(&args[1..])
                .current_dir(workdir)
                .status()?;
        }
    }
    Ok(())
}
```

**Commit:** `feat(ecosystem): add foreach-released commands`

#### 8.3 Integration Tests (0.25 Tag)

**File:** `tests/release_workflow.rs`

**Tests:**
```rust
#[test]
fn test_full_release_workflow() {
    let temp_repo = setup_test_repo_with_cargo_project();

    // Init
    cmd("clikd release init").assert().success();

    // Stage
    cmd("clikd release stage my-project").assert().success();

    // Confirm
    cmd("clikd release confirm").assert().success();

    // Check tags created
    assert!(tag_exists("my-project@1.0.0"));
}
```

**Commit:** `test(release): add integration tests for release workflow`

#### 8.4 Documentation (0.25 Tag)

**Files:**
- Update `README.md`
- Create `docs/RELEASE_MANAGEMENT.md`
- Help texts für alle Commands

**Commit:** `docs(release): add release management documentation`

---

## 🔄 Workflow Example (End Result)

### Single Project Release

```bash
# Initialize
cd my-rust-project/
clikd release init

# Check status
clikd release status
# Output: my-lib: 5 commits since v0.1.0

# Stage for release
clikd release stage my-lib
# Creates CHANGELOG.md draft: "# rc: minor"

# Confirm
clikd release confirm
# Calculates: v0.1.0 → v0.2.0
# Creates RC commit on rc branch

# CI Pipeline (automatically)
clikd release apply     # Rewrites Cargo.toml with v0.2.0
clikd release commit    # Creates release commit
clikd release tag       # Tags: my-lib@0.2.0
clikd github create-releases
clikd cargo foreach-released -- cargo publish
```

### Monorepo Multi-Project Release

```bash
cd clikd/

# Check what needs release
clikd release status
# gate: 12 commits since v0.6.0
# mondo: 3 commits since v0.5.0
# clikd-events: 1 commit since v0.1.0

# Stage multiple
clikd release stage gate mondo clikd-events

# Edit changelogs
# gate/CHANGELOG.md: "# rc: minor"
# mondo/CHANGELOG.md: "# rc: patch"
# packages/events/CHANGELOG.md: "# rc: patch"

# Confirm
clikd release confirm
# Resolves dependencies
# gate v0.6.0 → v0.7.0 (depends on events v0.1.1)
# mondo v0.5.0 → v0.5.1
# clikd-events v0.1.0 → v0.1.1

# CI does the rest...
```

---

## ✅ Definition of Done

### Phase 1-2
- [ ] All Cranko core files copied
- [ ] Dependencies added to Cargo.toml
- [ ] Project compiles without errors
- [ ] Version management tests pass

### Phase 3-4
- [ ] Repository layer functional
- [ ] Dependency graph working
- [ ] Cargo loader detects projects
- [ ] NPM loader detects projects
- [ ] Python loader detects projects

### Phase 5
- [ ] `clikd release init` creates .clikd/config.toml
- [ ] `clikd release status` shows project states
- [ ] `clikd release stage` updates changelogs
- [ ] `clikd release confirm` creates RC commits
- [ ] Workflow commands (apply/commit/tag) work

### Phase 6
- [ ] GitHub release creation works
- [ ] Artifact upload works
- [ ] Token handling secure

### Phase 7
- [ ] Go projects detected
- [ ] Go versioning from tags works
- [ ] Elixir projects detected
- [ ] Elixir version rewriting works

### Phase 8
- [ ] All commands integrated in CLI
- [ ] Help texts complete
- [ ] Integration tests pass
- [ ] Documentation updated

---

## 📊 Progress Tracking

Use TodoWrite tool to track progress through each phase.

**Current Status:** Phase 1 - Setup

---

## 🐛 Known Issues & Edge Cases

### To Address During Implementation:

1. **Nested Workspaces:** Rust projects wie clikd haben root workspace + app workspaces
   - Solution: Loaders müssen beide erkennen

2. **Go Major Versions:** Go v2+ braucht /v2 in import path
   - Solution: Dokumentieren, nicht automatisch ändern

3. **Elixir Umbrella Apps:** Multiple apps in einem Repo
   - Solution: Jede app/ als separates Project

4. **Changelog Merge Conflicts:** Wenn mehrere Leute gleichzeitig stagen
   - Solution: Dokumentieren, Manual resolution

5. **CI Environment Detection:** Wie erkennen wir ob wir in CI sind?
   - Solution: Nutze ci_info crate (wie Cranko)

---

## 🎯 Success Metrics

Nach Completion:
- [ ] `clikd release init` in neuem Repo funktioniert
- [ ] Full workflow (stage → confirm → publish) funktioniert
- [ ] Alle 6 Sprachen supported (Rust, NPM, Python, C#, Go, Elixir)
- [ ] Monorepo wie clikd kann released werden
- [ ] GitHub integration funktioniert
- [ ] Zero breaking changes für bestehende CLI features

---

**Next Step:** Start Phase 1 - Foundation Setup 🚀
