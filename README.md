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

```bash
# Initialize a new Clikd project
clikd init

# Start local development environment
clikd start

# Stop all services
clikd stop

# View service status
clikd status

# View service logs
clikd logs [service]

# Authentication
clikd login
clikd logout

# Database management
clikd db push    # Apply migrations
clikd db reset   # Reset database
```

## Features

- **Docker-based Services**: Orchestrates PostgreSQL, Redis, MinIO, APISIX, and more
- **Automatic Health Checking**: Waits for services to be healthy before starting dependent services
- **Volume Management**: Labeled volumes for easy cleanup and isolation
- **IDE Integration**: Automatic VSCode and IntelliJ IDEA configuration
- **Branch Isolation**: Separate containers per branch for conflict-free development

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
