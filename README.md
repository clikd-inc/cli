# Clikd

Clikd is a powerful and modular CLI tool built with Go, Cobra, and Viper.

## Features

- Modular command architecture
- Configuration management with Viper
- Environment variable support
- Subcommand structure
- Logging system

## Installation

```bash
go install github.com/nyxb/clikd@latest
```

Or build from source:

```bash
git clone https://github.com/nyxb/clikd.git
cd clikd
go build -o clikd
```

## Usage

```bash
# Get help
clikd --help

# Run with a specific config file
clikd --config=/path/to/config.toml

# Set log level
clikd --log-level=debug

# Get version
clikd version

# Initialize configuration
clikd init

# Initialize global configuration
clikd init --global

# Say hello
clikd hello --name="World"

# Say hello to the world
clikd hello world
```

## Configuration

Clikd looks for configuration in the following locations:
1. Custom config file specified by `--config` flag
2. `$HOME/.clikd/config.toml`
3. `./config.toml` in the current directory

You can also use environment variables prefixed with `CLIKD_` to configure the application.

Example configuration file (TOML):

```toml
version = "1.0.0"

[general]
log_level = "debug"
color = true

[ai]
enable = true
default_model = "mistral-medium"
default_provider = "mistral"
```

To initialize a new configuration file:

```bash
# Initialize in current directory
clikd init

# Initialize global configuration
clikd init --global

# Override existing configuration
clikd init --force
```

## Project Structure

```
clikd/
├── cmd/                    # Entry points for executables
│   └── clikd/              # Main CLI application
│       └── main.go         # Main entry point
├── pkg/                    # Package code
│   ├── commands/           # CLI commands
│   │   ├── root/           # Root command
│   │   ├── version/        # Version command
│   │   ├── initialize/     # Initialize command
│   │   └── hello/          # Hello command with subcommands
│   ├── config/             # Configuration management
│   ├── models/             # Data models
│   └── utils/              # Utility functions
├── go.mod                  # Go module file
└── go.sum                  # Go module checksums
```

## Extending

To add a new command:

1. Create a new package in `pkg/commands/your-command/`
2. Implement your command using Cobra
3. Add your command to the root command in `cmd/clikd/main.go`

## License

MIT 
