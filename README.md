# Clikd CLI

Local development environment management for Clikd.

## Installation

### macOS (Homebrew)

```bash
brew install clikd-inc/tap/clikd
```

### Linux / macOS (Shell script)

```bash
curl --proto '=https' --tlsv1.2 -LsSf https://github.com/clikd-inc/cli/releases/latest/download/clikd-installer.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://github.com/clikd-inc/cli/releases/latest/download/clikd-installer.ps1 | iex
```

### Windows (Scoop)

```powershell
scoop bucket add clikd https://github.com/clikd-inc/scoop-bucket
scoop install clikd
```

### Windows (MSI Installer)

Download the latest `.msi` installer from the [releases page](https://github.com/clikd-inc/cli/releases).

### From Source

```bash
cargo install --git https://github.com/clikd-inc/cli
```

## Usage

### Project Management

```bash
# Initialize a new Clikd project
clikd init

# Start local development environment
clikd start

# Stop all services
clikd stop

# View service status (interactive TUI)
clikd status
```

### Authentication

```bash
# Login to Clikd platform
clikd login

# Logout from Clikd platform
clikd logout

# Check authentication status
clikd auth status
```

### Release Management

The CLI includes powerful release management for monorepos and multi-language projects:

```bash
# Initialize release management in your repository
clikd release init

# Check which projects have changes and preview release
clikd release status

# Prepare a new release (bump versions, update changelogs)
clikd release prepare patch   # Patch version bump (0.1.0 → 0.1.1)
clikd release prepare minor   # Minor version bump (0.1.0 → 0.2.0)
clikd release prepare major   # Major version bump (0.1.0 → 1.0.0)
```

#### Release Management Features

- **Multi-Language Support**: Automatically detects and manages versions for:
  - Rust (Cargo.toml)
  - Node.js (package.json)
  - Python (setup.py, pyproject.toml)
  - Go (go.mod)
  - Elixir (mix.exs)
  - C# (.csproj)

- **Dependency Resolution**: Analyzes project dependencies and determines correct release order

- **Automatic Changelog Generation**: Creates and updates CHANGELOG.md files based on Git commits

- **Monorepo-Aware**: Handles complex dependency graphs in monorepos with multiple interconnected projects

#### Example Workflow

```bash
# 1. Initialize release management
cd /path/to/your/repo
clikd release init

# 2. Make your changes and commit them
git add .
git commit -m "feat: add new feature"

# 3. Check what will be released
clikd release status

# 4. Prepare the release (updates versions and changelogs)
clikd release prepare minor

# 5. Review and commit the changes
git add .
git commit -m "chore(release): prepare 0.2.0"
git tag v0.2.0
git push origin main --tags
```

#### Configuration

Release management is configured via `.clikd/release.toml` in your repository:

```toml
[repository]
upstream_name = "origin"

[[projects]]
name = "my-rust-crate"
qualifier = "cargo"
changelog_path = "CHANGELOG.md"
release_branch = "main"

[[projects]]
name = "my-frontend"
qualifier = "npm"
changelog_path = "packages/frontend/CHANGELOG.md"
```

Configuration is automatically created when you run `clikd release init`.

#### Advanced Usage

```bash
# Force initialization even with uncommitted changes
clikd release init --force

# Use a different upstream remote
clikd release init --upstream upstream

# Manual version specification (bypasses automatic bump detection)
clikd release prepare manual
```

## Features

### Service Management
- **Docker-based Services**: Orchestrates PostgreSQL, Redis, MinIO, APISIX, and more
- **Automatic Health Checking**: Waits for services to be healthy before starting dependent services
- **Volume Management**: Labeled volumes for easy cleanup and isolation
- **IDE Integration**: Automatic VSCode and IntelliJ IDEA configuration
- **Branch Isolation**: Separate containers per branch for conflict-free development
- **Interactive TUI**: Real-time container monitoring with live metrics and log viewer

### Release Management
- **Monorepo Support**: Handles complex multi-project repositories with ease
- **Multi-Language**: Supports Rust, Node.js, Python, Go, Elixir, and C# projects
- **Dependency Graph**: Automatically determines release order based on project dependencies
- **Version Bumping**: Semantic versioning with major, minor, and patch bumps
- **Changelog Generation**: Automatic CHANGELOG.md creation from Git commits
- **Git Integration**: Works seamlessly with Git tags and commit history

## Requirements

- Docker or OrbStack
- macOS, Linux, or Windows

## Development

```bash
# Clone the repository
git clone https://github.com/clikd-inc/cli.git
cd cli

# Build
cargo build --release

# Run
cargo run -- --help

# Test
cargo test
```

## License

MIT
