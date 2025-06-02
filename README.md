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
clikd --config=/path/to/config.yaml

# Set log level
clikd --log-level=debug

# Get version
clikd version

# Say hello
clikd hello --name="World"

# Say hello to the world
clikd hello world
```

## Configuration

Clikd looks for configuration in the following locations:
1. Custom config file specified by `--config` flag
2. `$HOME/.clikd/config.yaml`
3. `./config.yaml` in the current directory

You can also use environment variables prefixed with `CLIKD_` to configure the application.

Example configuration file (YAML):

```yaml
log_level: debug
log_format: text

api:
  endpoint: https://api.example.com
  token: your-api-token
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
