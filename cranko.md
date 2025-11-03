# Cranko Release-Automation - Vollständige Integration in Clikd CLI

## Executive Summary

**Ziel:** 1:1 Feature-Parität mit Cranko - ALLE 29 Commands mit vollständiger UI/UX

## Cranko Command-Inventar (100% Coverage)

### Developer Commands (6)

- ✅ **bootstrap** - Repository für Cranko initialisieren
- ✅ **status** - Commit-Historie seit letztem Release anzeigen
- ✅ **stage** - Projekte für Release vorbereiten (Changelog-Draft)
- ✅ **confirm** - Staged Releases auf RC-Branch committen
- ✅ **log** - Git-Log für spezifisches Projekt
- ✅ **diff** - Diff seit letztem Release

### CI/CD Commands (18)

#### Release Workflow:
- ✅ **release-workflow apply-versions** - Versionen aus RC-Commit anwenden
- ✅ **release-workflow commit** - Release-Commit erstellen
- ✅ **release-workflow tag** - Git-Tags erstellen

#### Cargo:
- ✅ **cargo foreach-released** - Command auf released Crates ausführen
- ✅ **cargo package-released-binaries** - Binaries für released Crates packen

#### NPM:
- ✅ **npm foreach-released** - Command auf released Packages ausführen
- ✅ **npm install-token** - NPM Auth-Token für CI installieren
- ✅ **npm lerna-workaround** - Lerna-Kompatibilität

#### Python:
- ✅ **python foreach-released** - Command auf released Packages ausführen
- ✅ **python install-token** - PyPI Token für CI installieren

#### GitHub:
- ✅ **github create-releases** - GitHub Releases erstellen
- ✅ **github create-custom-release** - Custom GitHub Release
- ✅ **github delete-release** - GitHub Release löschen
- ✅ **github install-credential-helper** - Git Credential Helper
- ✅ **github upload-artifacts** - Artifacts zu Release hochladen

#### Zenodo (Scientific Publishing):
- ✅ **zenodo preregister** - DOI vorregistrieren
- ✅ **zenodo upload-artifacts** - Artifacts hochladen
- ✅ **zenodo publish** - Zenodo-Deposition veröffentlichen

#### CI Utils:
- ✅ **ci-util env-to-file** - ENV-Variable in Datei schreiben (für Secrets)

### Utility Commands (5)

- ✅ **show** - Verschiedene Infos anzeigen (Subcommands):
  - `show version` - Projekt-Version anzeigen
  - `show if-released` - Check ob Projekt released wurde
  - `show toposort` - Projekte in Dependency-Order
  - `show tctag` - "thiscommit:"-Tag generieren
  - `show cranko-version-doi` / `cranko-concept-doi` - Cranko DOIs
- ✅ **git-util reboot-branch** - Branch neu aufsetzen
- ✅ **help** - Hilfe anzeigen
- ✅ **list-commands** - Alle Commands auflisten

---

## CLI-Struktur (Hybrid mit ALLEN Commands)

```rust
#[derive(Subcommand)]
pub enum Commands {
    // === BESTEHENDE CLIKD COMMANDS (unverändert) ===
    Login { no_browser: bool },
    Logout,
    #[command(subcommand)]
    Auth(AuthCommands),
    Init(InitArgs),
    Start(StartArgs),
    Stop(StopArgs),
    Status(StatusArgs),  // Docker-Status
    Logs(LogsArgs),
    #[command(subcommand)]
    Db(DbCommands),
    Completions { shell: Shell },

    // === NEUE RELEASE COMMANDS ===

    // Top-Level (häufig genutzt)
    /// Bootstrap Cranko in repository
    Bootstrap(BootstrapArgs),

    /// Stage projects for release
    Stage(StageArgs),

    /// Confirm staged releases
    Confirm(ConfirmArgs),

    /// Show project commit log
    Log(LogArgs),

    /// Show diff since last release
    Diff(DiffArgs),

    /// Show project release status
    ReleaseStatus(ReleaseStatusArgs),  // Namens-Kollision vermeiden

    // Namespace: release (CI/CD-Workflow)
    #[command(subcommand)]
    Release(ReleaseCommands),

    // Namespace: show (Utility)
    #[command(subcommand)]
    Show(ShowCommands),

    // Namespace: git-util
    #[command(subcommand)]
    GitUtil(GitUtilCommands),

    // Namespace: ci-util
    #[command(subcommand)]
    CiUtil(CiUtilCommands),
}

#[derive(Subcommand)]
pub enum ReleaseCommands {
    /// Apply versions from rc commit
    ApplyVersions(ApplyVersionsArgs),

    /// Create release commit
    Commit(ReleaseCommitArgs),

    /// Create git tags for releases
    Tag(ReleaseTagArgs),

    // Package-Manager Sub-Namespaces
    #[command(subcommand)]
    Cargo(CargoCommands),

    #[command(subcommand)]
    Npm(NpmCommands),

    #[command(subcommand)]
    Python(PythonCommands),

    // GitHub Sub-Namespace
    #[command(subcommand)]
    Github(GithubCommands),

    // Zenodo Sub-Namespace
    #[command(subcommand)]
    Zenodo(ZenodoCommands),
}

#[derive(Subcommand)]
pub enum CargoCommands {
    /// Execute command on released crates
    ForeachReleased(ForeachReleasedArgs),

    /// Package binaries for released crates
    PackageReleasedBinaries(PackageReleasedBinariesArgs),
}

#[derive(Subcommand)]
pub enum NpmCommands {
    /// Execute command on released packages
    ForeachReleased(ForeachReleasedArgs),

    /// Install NPM auth token for CI
    InstallToken(InstallTokenArgs),

    /// Lerna compatibility workaround
    LernaWorkaround,
}

#[derive(Subcommand)]
pub enum PythonCommands {
    /// Execute command on released packages
    ForeachReleased(ForeachReleasedArgs),

    /// Install PyPI token for CI
    InstallToken(InstallTokenArgs),
}

#[derive(Subcommand)]
pub enum GithubCommands {
    /// Create GitHub releases
    CreateReleases(CreateReleasesArgs),

    /// Create custom GitHub release
    CreateCustomRelease(CreateCustomReleaseArgs),

    /// Delete GitHub release
    DeleteRelease(DeleteReleaseArgs),

    /// Install Git credential helper
    InstallCredentialHelper,

    /// Upload artifacts to release
    UploadArtifacts(UploadArtifactsArgs),
}

#[derive(Subcommand)]
pub enum ZenodoCommands {
    /// Pre-register DOI for release
    Preregister(ZenodoPreregisterArgs),

    /// Upload artifacts to Zenodo
    UploadArtifacts(ZenodoUploadArtifactsArgs),

    /// Publish Zenodo deposition
    Publish(ZenodoPublishArgs),
}

#[derive(Subcommand)]
pub enum ShowCommands {
    /// Show project version
    Version(ShowVersionArgs),

    /// Check if project was released
    IfReleased(ShowIfReleasedArgs),

    /// Show projects in topological order
    Toposort,

    /// Generate thiscommit: tag
    TcTag,

    /// Show Cranko version DOI
    CrankoVersionDoi,

    /// Show Cranko concept DOI
    CrankoConceptDoi,
}

#[derive(Subcommand)]
pub enum GitUtilCommands {
    /// Reboot a branch from scratch
    RebootBranch(RebootBranchArgs),
}

#[derive(Subcommand)]
pub enum CiUtilCommands {
    /// Save environment variable to file
    EnvToFile(EnvToFileArgs),
}
```

---

## Vollständige Ordnerstruktur

```
apps/cli/src/
├── main.rs
├── cli.rs (ALLE Commands registriert!)
├── error.rs (erweitert um Release-Errors)
├── config.rs (erweitert um ReleaseConfig)
├── lib.rs
│
├── cmd/
│   ├── auth.rs ✅ (besteht)
│   ├── init.rs ✅ (besteht)
│   ├── start.rs ✅ (besteht)
│   ├── stop.rs ✅ (besteht)
│   ├── status.rs ✅ (besteht - Docker)
│   ├── logs.rs ✅ (besteht)
│   ├── db.rs ✅ (besteht)
│   ├── completions.rs ✅ (besteht)
│   │
│   ├── release/                    # NEU (18 Commands)
│   │   ├── bootstrap.rs           # TOP-LEVEL: clikd bootstrap
│   │   ├── stage.rs               # TOP-LEVEL: clikd stage
│   │   ├── confirm.rs             # TOP-LEVEL: clikd confirm
│   │   ├── release_status.rs      # TOP-LEVEL: clikd release-status
│   │   ├── log.rs                 # TOP-LEVEL: clikd log
│   │   ├── diff.rs                # TOP-LEVEL: clikd diff
│   │   ├── apply_versions.rs      # clikd release apply-versions
│   │   ├── commit.rs              # clikd release commit
│   │   ├── tag.rs                 # clikd release tag
│   │   ├── cargo.rs               # clikd release cargo {...}
│   │   ├── npm.rs                 # clikd release npm {...}
│   │   ├── python.rs              # clikd release python {...}
│   │   ├── github.rs              # clikd release github {...}
│   │   └── zenodo.rs              # clikd release zenodo {...}
│   │
│   ├── show.rs                     # NEU: clikd show {...}
│   ├── git_util.rs                 # NEU: clikd git-util {...}
│   └── ci_util.rs                  # NEU: clikd ci-util {...}
│
├── core/
│   ├── auth/ ✅ (besteht)
│   ├── docker/ ✅ (besteht)
│   ├── git/ ✅ (besteht - erweitern!)
│   │   ├── branch.rs ✅
│   │   ├── gitignore.rs ✅
│   │   ├── repository.rs          # NEU: Git-Operationen
│   │   ├── history.rs             # NEU: Commit-History
│   │   └── util.rs                # NEU: reboot-branch, etc.
│   ├── ide/ ✅ (besteht)
│   ├── start/ ✅ (besteht)
│   ├── stop/ ✅ (besteht)
│   ├── status/ ✅ (besteht)
│   ├── config/ ✅ (besteht)
│   ├── root.rs ✅ (besteht)
│   │
│   ├── release/                    # NEU: Release Business Logic
│   │   ├── project_service.rs     # Project discovery
│   │   ├── version_service.rs     # Version management
│   │   ├── graph_service.rs       # Dependency graph
│   │   ├── history_service.rs     # Git history analysis
│   │   ├── workflow_service.rs    # stage/confirm/apply orchestration
│   │   ├── changelog_service.rs   # Changelog generation
│   │   ├── tag_service.rs         # Git tags
│   │   ├── foreach_service.rs     # foreach-released logic
│   │   ├── models.rs              # Domain models
│   │   └── errors.rs              # Release errors
│   │
│   ├── package_managers/           # NEU: Package-Manager
│   │   ├── cargo.rs               # Cargo.toml
│   │   ├── npm.rs                 # package.json
│   │   ├── python.rs              # pyproject.toml
│   │   ├── csproj.rs              # .csproj
│   │   ├── detector.rs            # Auto-detection
│   │   ├── rewriter.rs            # Trait für Rewriters
│   │   └── models.rs
│   │
│   ├── github/                     # NEU: GitHub API
│   │   ├── client.rs              # API client
│   │   ├── releases.rs            # Release operations
│   │   ├── artifacts.rs           # Artifact uploads
│   │   └── credentials.rs         # Credential helper
│   │
│   └── zenodo/                     # NEU: Zenodo API
│       ├── client.rs              # API client
│       ├── deposition.rs          # Deposition operations
│       └── metadata.rs            # Metadata parsing
│
└── utils/
    ├── terminal.rs ✅ (besteht)
    ├── retry.rs ✅ (besteht)
    ├── theme.rs ✅ (besteht)
    ├── changelog.rs                # NEU: Changelog utilities
    └── semver.rs                   # NEU: Semver helpers
```

---

## Cargo.toml - Vollständige Dependencies

```toml
[dependencies]
# === BESTEHENDE (behalten) ===
clap = { version = "4.5", features = ["derive", "env", "wrap_help", "cargo"] }
clap_complete = "4.5"
anyhow = "1.0"
thiserror = "2.0"
tracing = "0.1.41"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }
owo-colors = { version = "4.2", features = ["supports-colors"] }
spinoff = { version = "0.8", features = ["arc"] }
kdam = { version = "0.6", features = ["gradient", "rich", "spinner"] }
dialoguer = "0.12"
open = "5.3"
config = "0.15.18"
serde = { workspace = true }
toml = "0.9.8"
dirs = "6.0"
minijinja = "2.5"
bollard = "0.19.3"
git2 = "0.20.2"
keyring = { version = "3.6.3", features = ["apple-native", "windows-native", "linux-native"] }
reqwest = { version = "0.12", features = ["json", "rustls-tls"] }
ureq = { version = "3.1.2", features = ["json"] }
secrecy = "0.10.3"
zeroize = "1.8"
tokio = { version = "1.40", features = ["rt-multi-thread", "macros"] }
futures = "0.3"
chrono = "0.4"
uuid = { version = "1.0", features = ["v4"] }

# === NEU für Cranko-Funktionalität ===
# Graph & Dependencies
petgraph = "0.7"                    # Dependency Graph mit toposort

# Package Manager Parsing
cargo_metadata = "0.19"             # Cargo.toml + Workspaces
toml_edit = "0.22"                  # TOML Format-erhaltend
quick-xml = "0.37"                  # .csproj XML

# Semver
semver = "1.0"                      # Version parsing & comparison

# Text Processing
nom = "7.1"                         # Changelog parsing
textwrap = "0.16"                   # Text wrapping für Output

# HTTP (für Zenodo)
base64 = "0.22"                     # Base64 encoding
percent-encoding = "2.3"            # URL encoding

# Tar & Zip (für Artifacts)
tar = "0.4"                         # Tar archives
flate2 = "1.0"                      # Gzip compression
zip = { version = "2.2", default-features = false, features = ["deflate", "time"] }

# Random (für tctag)
rand = "0.8"                        # thiscommit: tag generation

# Optional: JSON5 für Zenodo metadata
json5 = "0.4"                       # JSON5 parsing
```

---

## Vollständiger Implementierungs-Plan

### Phase 1: Core Infrastructure (Woche 1)

**Deliverable:** Fundament + erste Commands testbar

1. **Domain Models** (`core/release/models.rs`)
   ```rust
   pub struct Project { /* siehe Cranko */ }
   pub struct Version { /* siehe Cranko */ }
   pub struct Dependency { /* siehe Cranko */ }
   pub struct ReleaseInfo { /* siehe Cranko */ }
   pub struct CommitInfo { /* siehe Cranko */ }
   ```

2. **Git Repository** (`core/git/repository.rs`)
   - Alle Git-Operationen aus Cranko portieren
   - History-Scanning, Branch-Management, Tag-Creation

3. **Project Detection** (`core/package_managers/detector.rs`)
   - Scannt Repo rekursiv
   - Erkennt Cargo, NPM, Python, C# Projekte

4. **CLI Structure** (alle Commands registrieren!)
   - `cli.rs` mit ALLEN 29 Commands
   - Erst als Stubs (`return unimplemented!()`)

**Test:** `clikd bootstrap`, `clikd stage`, `clikd status` kompilieren

---

### Phase 2: Developer Workflow (Woche 2)

**Deliverable:** Dev-Commands vollständig nutzbar

5. **Bootstrap** (`cmd/release/bootstrap.rs`)
   - Initialisiert `.clikd/release.toml`
   - Erstellt rc/release Branches
   - Seed-Versionen für existierende Projekte

6. **Status** (`cmd/release/release_status.rs`)
   - Zeigt Commits seit letztem Release
   - Nutzt `utils/theme.rs` für Output

7. **Stage** (`cmd/release/stage.rs`)
   - Changelog-Draft mit minijinja
   - Editor öffnen (dialoguer)
   - RC-Info in Git schreiben

8. **Confirm** (`cmd/release/confirm.rs`)
   - RC-Commit erstellen
   - Dependency-Validation
   - Working-Tree reset

9. **Log & Diff** (`cmd/release/log.rs`, `diff.rs`)
   - Git-Log für Projekt
   - Diff seit letztem Release

**Test:** Kompletter Dev-Workflow funktioniert

---

### Phase 3: Package-Manager-Integration (Woche 2-3)

**Deliverable:** Rewriters für alle Package-Manager

10. **Cargo** (`core/package_managers/cargo.rs`)
    - Workspace-Detection (cargo_metadata)
    - Version-Rewriting (toml_edit)
    - Dependency-Rewriting

11. **NPM** (`core/package_managers/npm.rs`)
    - package.json Parsing (serde_json)
    - Lerna-Support

12. **Python** (`core/package_managers/python.rs`)
    - pyproject.toml (toml_edit)
    - setup.py Support

13. **C#** (`core/package_managers/csproj.rs`)
    - XML-Parsing (quick-xml)
    - AssemblyVersion + PackageVersion

**Test:** Version-Rewriting für alle Package-Manager

---

### Phase 4: CI/CD Workflow (Woche 3)

**Deliverable:** Release-Workflow vollständig

14. **Apply Versions** (`cmd/release/apply_versions.rs`)
    - RC-Commit lesen
    - Rewriters aufrufen
    - Changelog-Updates

15. **Commit** (`cmd/release/commit.rs`)
    - Release-Commit erstellen
    - Multi-Project-Support

16. **Tag** (`cmd/release/tag.rs`)
    - Annotated Tags
    - Format: `{project}@{version}`

17. **Foreach-Released** (`core/release/foreach_service.rs`)
    - Cargo foreach-released
    - NPM foreach-released
    - Python foreach-released

**Test:** CI/CD Pipeline simulieren

---

### Phase 5: GitHub Integration (Woche 3-4)

**Deliverable:** GitHub-Commands vollständig

18. **GitHub Client** (`core/github/client.rs`)
    - REST API via reqwest
    - Token aus keyring (bestehender Auth!)

19. **Create Releases** (`cmd/release/github.rs`)
    - Releases für alle released Projects
    - Release-Notes aus Changelog

20. **Upload Artifacts** (`cmd/release/github.rs`)
    - Binary/Asset-Upload
    - Checksums

21. **Credential Helper** (`cmd/release/github.rs`)
    - Git Credential Helper installieren

**Test:** GitHub Release erstellen

---

### Phase 6: Zenodo Integration (Optional, Woche 4)

**Deliverable:** Scientific Publishing

22. **Zenodo Client** (`core/zenodo/client.rs`)
    - Zenodo REST API
    - DOI-Management

23. **Preregister** (`cmd/release/zenodo.rs`)
    - DOI vorregistrieren

24. **Upload & Publish** (`cmd/release/zenodo.rs`)
    - Artifacts hochladen
    - Deposition veröffentlichen

**Test:** Zenodo-Workflow (mit Sandbox)

---

### Phase 7: Utility Commands (Woche 4)

**Deliverable:** Alle Utility-Commands

25. **Show Commands** (`cmd/show.rs`)
    - version, if-released, toposort, tctag
    - cranko-version-doi, cranko-concept-doi

26. **Git-Util** (`cmd/git_util.rs`)
    - reboot-branch

27. **CI-Util** (`cmd/ci_util.rs`)
    - env-to-file (Secrets für CI)

**Test:** Alle Commands manuell testen

---

### Phase 8: Polish & Documentation (Woche 4)

**Deliverable:** Production-Ready

28. **Error-Handling**
    - Alle Error-Cases abdecken
    - Hilfreiche Error-Messages

29. **Progress-Indicators**
    - Spinners für langläufige Ops
    - Progress-Bars für Downloads

30. **Help-Texte**
    - Detaillierte `--help` für alle Commands
    - Examples in Help-Text

31. **Integration-Tests**
    - End-to-End Tests
    - CI/CD Pipeline

---

## UI/UX - Production-Ready von Anfang an

### Terminal-Output (nutzt bestehende `utils/theme.rs`)

```rust
// ALLE Commands nutzen konsistente UI:
println!("{}", header("Bootstrapping Release Workflow"));
println!("{}", step_message("Detecting projects..."));

// Projekt-Liste mit Highlighting
for proj in projects {
    println!("  {} {}",
        highlight(&proj.name),
        dimmed(&format!("({})", proj.version))
    );
}

// Spinners für Operations
let mut sp = create_spinner("Analyzing git history...");
// ... work
sp.success("Found 12 commits since last release");

// Erfolgs-Messages
println!("\n{}", success_message("3 projects staged for release"));

// Warnungen
if has_uncommitted_changes {
    println!("{}", warning_message("Uncommitted changes detected"));
}

// Errors
if dependency_cycle {
    println!("{}", error_message("Dependency cycle detected!"));
}
```

### Progress-Bars (wie bestehende Docker-Commands)

```rust
// Beim Artifact-Upload
let pb = ProgressBar::new(total_bytes);
pb.set_style(/* ... */);
// Update während Upload
pb.inc(chunk_size);
pb.finish_with_message("Upload complete!");
```

### Interactive Prompts (dialoguer bereits vorhanden)

```rust
// Stage-Command: User wählt Bump-Type
let bump_type = Select::new()
    .with_prompt("Select version bump for 'gate'")
    .items(&["major", "minor", "patch", "skip"])
    .default(2)  // patch
    .interact()?;
```

---

## Beispiel-Workflows (Production-Ready!)

### Setup (einmalig)

```bash
cd /Users/nyxb/Projects/clikd-project/clikd

clikd bootstrap
# ✨ Bootstrapping Release Workflow
# → Detecting projects...
#   • gate (0.1.0)
#   • rig (0.1.0)
#   • studio (0.1.0)
#   • cli (0.1.0)
# → Creating branches...
#   ✓ rc branch created
#   ✓ release branch created
# → Creating CHANGELOG.md files...
#   ✓ apps/services/gate/CHANGELOG.md
#   ✓ apps/services/rig/CHANGELOG.md
#   ✓ apps/studio/CHANGELOG.md
#   ✓ apps/cli/CHANGELOG.md
# ✅ Bootstrap complete! Run 'clikd stage' to get started.
```

### Entwicklung

```bash
# Status checken
clikd release-status
# gate: 5 commits since 0.1.0
# rig: 3 commits since 0.1.0
# studio: 0 commits since 0.1.0
# cli: 12 commits since 0.1.0

# Stage gate + rig für Release
clikd stage gate rig
# → Analyzing changes...
#   gate: 5 relevant commits
#   rig: 3 relevant commits
# → Generating changelog drafts...
#   ✓ apps/services/gate/CHANGELOG.md updated
#   ✓ apps/services/rig/CHANGELOG.md updated
#
# 📝 Please edit the changelogs and set version bumps:
#    apps/services/gate/CHANGELOG.md
#    apps/services/rig/CHANGELOG.md

# (User editiert Changelogs, setzt bump: minor)

clikd confirm
# → Reading changelog metadata...
#   gate: minor bump (0.1.0 => 0.2.0)
#   rig: patch bump (0.1.0 => 0.1.1)
# → Validating dependencies...
#   ✓ No dependency conflicts
# → Creating RC commit...
#   ✓ Commit created on 'rc' branch
# → Resetting working tree...
# ✅ Release staged! Push with: git push origin rc

git push origin rc
```

### CI/CD (GitHub Actions)

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    branches: [rc]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history

      - name: Install Clikd CLI
        run: cargo install --path apps/cli

      - name: Apply versions
        run: clikd release apply-versions

      - name: Build & Test
        run: cargo test --workspace

      - name: Create release commit
        run: clikd release commit

      - name: Create tags
        run: clikd release tag

      - name: Push release
        run: |
          git push origin release
          git push --tags

      - name: Build binaries
        run: clikd release cargo package-released-binaries

      - name: Create GitHub releases
        run: clikd release github create-releases
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Upload artifacts
        run: clikd release github upload-artifacts target/releases/*.tar.gz
```

---

## Testing-Strategie (Production-Grade)

### Unit-Tests (in source files)

```rust
// core/release/version_service.rs
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version_parsing() { /* ... */ }

    #[test]
    fn test_version_bumping() { /* ... */ }

    #[test]
    fn test_semver_comparison() { /* ... */ }
}
```

### Integration-Tests (tests/)

```rust
// tests/integration/release_workflow.rs
#[test]
fn test_full_release_workflow() {
    // Setup temp git repo
    let temp = TempDir::new().unwrap();
    setup_git_repo(&temp);

    // Bootstrap
    Command::cargo_bin("clikd").unwrap()
        .arg("bootstrap")
        .current_dir(&temp)
        .assert()
        .success();

    // Stage
    Command::cargo_bin("clikd").unwrap()
        .arg("stage")
        .arg("test-project")
        .current_dir(&temp)
        .assert()
        .success();

    // Confirm
    Command::cargo_bin("clikd").unwrap()
        .arg("confirm")
        .current_dir(&temp)
        .assert()
        .success();

    // Verify RC commit exists
    assert_rc_commit_created(&temp);
}
```

### Property-Based Tests (proptest)

```rust
// core/release/graph_service.rs
#[cfg(test)]
mod tests {
    use proptest::prelude::*;

    proptest! {
        #[test]
        fn test_no_cycles_in_dag(projects in project_vec_strategy()) {
            let graph = GraphService::build(projects)?;
            assert!(graph.toposort().is_ok());
        }
    }
}
```

---

## Migration-Checkliste (von Cranko)

### Module-Mapping (Port 1:1)

| Cranko Source     | Clikd Target                      | Status |
|-------------------|-----------------------------------|--------|
| src/version.rs    | core/release/models.rs            | Port   |
| src/project.rs    | core/release/models.rs            | Port   |
| src/graph.rs      | core/release/graph_service.rs     | Port   |
| src/repository.rs | core/git/repository.rs            | Port   |
| src/cargo.rs      | core/package_managers/cargo.rs    | Port   |
| src/npm.rs        | core/package_managers/npm.rs      | Port   |
| src/pypa.rs       | core/package_managers/python.rs   | Port   |
| src/csproj.rs     | core/package_managers/csproj.rs   | Port   |
| src/changelog.rs  | utils/changelog.rs                | Port   |
| src/github.rs     | core/github/client.rs             | Port   |
| src/zenodo.rs     | core/zenodo/client.rs             | Port   |
| src/rewriters.rs  | core/package_managers/rewriter.rs | Port   |
| src/bootstrap.rs  | cmd/release/bootstrap.rs          | Port   |
| src/gitutil.rs    | core/git/util.rs                  | Port   |
| src/logger.rs     | ❌ SKIP (use tracing)              | -      |
| src/env.rs        | ❌ SKIP (use std::env)             | -      |
| src/errors.rs     | error.rs (integrate)              | Adapt  |
| src/config.rs     | config.rs (integrate)             | Adapt  |
| src/app.rs        | ❌ SKIP (CLI-Session)              | -      |

### Modernisierung

- ✅ structopt → clap 4.5
- ✅ log → tracing
- ✅ termcolor → owo-colors
- ✅ Keine mod.rs Dateien
- ✅ Async für I/O (GitHub, Zenodo, File-Ops)
- ✅ Integration in bestehende Config/Error-Systeme

---

## Success Criteria (100% Feature-Parity)

### Functional

- ✅ ALLE 29 Cranko-Commands implementiert
- ✅ Cargo, NPM, Python, C# Support
- ✅ JIT-Versioning Workflow
- ✅ Dependency-Graph-Resolution
- ✅ Changelog-Generation
- ✅ GitHub + Zenodo Integration

### Non-Functional

- ✅ Production-Ready Code (keine TODOs)
- ✅ Error-Handling für alle Edge-Cases
- ✅ Progress-Indicators für langläufige Ops
- ✅ Integration-Tests (>80% Coverage)
- ✅ Performance: <100ms Startup, <1s für status
- ✅ Memory: <50MB für normale Operations

### UX

- ✅ Konsistente UI (nutzt utils/theme.rs)
- ✅ Hilfreiche Error-Messages
- ✅ Detaillierte `--help` für alle Commands
- ✅ Shell-Completions (clap_complete)

---

## Timeline & Milestones

### Milestone 1 (Woche 1): Core + Bootstrap
- ✅ 6 Developer Commands funktionieren

### Milestone 2 (Woche 2): Package-Manager + Workflow
- ✅ 18 CI/CD Commands funktionieren
- ✅ Version-Rewriting für alle Package-Manager

### Milestone 3 (Woche 3): GitHub + Polish
- ✅ GitHub-Integration vollständig
- ✅ Alle Commands getestet

### Milestone 4 (Woche 4): Zenodo + Production
- ✅ Zenodo-Integration (optional)
- ✅ Utility-Commands
- ✅ Integration-Tests
- ✅ Documentation
- ✅ Production-Ready!

---

## Geschätzter Aufwand

**4 Wochen für 100% Feature-Parity**

- **Risiko:** Low (Port bewährter Code, keine Breaking Changes)
- **Impact:** High (Professionelle Release-Automation für Monorepo)
