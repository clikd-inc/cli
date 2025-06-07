# Clikd

Clikd is a powerful and modular CLI tool built with Go, Cobra, and Viper.

## Features

- Modular command architecture
- Configuration management with Viper
- Environment variable support
- Subcommand structure
- Logging system

## KI-Unterstützung

Die clikd CLI enthält leistungsstarke KI-Funktionen, die über die `gollm`-Bibliothek implementiert sind. Die folgenden LLM-Provider werden unterstützt:

- **Mistral AI** (Standard): Mistral Medium, Small und Large
- **OpenAI**: GPT-4o, GPT-4o-mini, GPT-3.5-Turbo, GPT-4-Turbo
- **Anthropic**: Claude 3 (Opus, Sonnet, Haiku), Claude 3.5 Sonnet
- **Groq**: Llama 3 (8B, 70B), Mixtral 8x7B
- **OpenRouter**: Auto-Routing und Fallback-Funktionen
- **Lokale Modelle**: Ollama-Integration für Llama, Mistral und andere

### Konfiguration

Die KI-Funktionalität kann in der globalen Konfigurationsdatei (`$HOME/.clikd/config.toml`) oder in lokalen Projektkonfigurationen (`./clikd/config.toml`) konfiguriert werden.

API-Schlüssel sollten in einer `.env`-Datei im Projektverzeichnis gespeichert werden:

```
CLIKD_MISTRAL_API_KEY=sk-...
CLIKD_OPENAI_API_KEY=sk-...
CLIKD_ANTHROPIC_API_KEY=sk-...
CLIKD_GROQ_API_KEY=sk-...
CLIKD_OPENROUTER_API_KEY=sk-...
```

Alternativ können sie in der globalen Konfigurationsdatei gespeichert werden.

### Testen der KI-Funktionalität

Um die KI-Funktionalität zu testen, verwenden Sie den Befehl `ai-test`:

```bash
clikd ai-test "Wie kann ich einen Changelog erstellen?" --model=mistral-medium
```

Sie können das zu verwendende Modell mit dem `--model`-Flag angeben.

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
