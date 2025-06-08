# AI Service for clikd CLI

This module provides a centralized, service-based approach to integrating AI capabilities into the clikd CLI.

## Architecture

The AI service follows a clean architecture with clear separation of concerns:

```
internal/services/ai/
├── service.go         # Main service interface and implementation
├── client.go          # Client interface for different AI providers
├── config.go          # Configuration for AI providers and models
├── gollm.go           # Gollm client implementation (supports multiple providers)
└── usecases/          # Domain-specific AI use cases
    ├── changelog.go   # Changelog-related AI functionality
    └── commit.go      # Commit-related AI functionality
```

## Core Components

### Service Interface

The central `Service` interface in `service.go` provides high-level access to all AI capabilities:

```go
// Service provides a centralized way to access AI functionality
type Service interface {
    IsEnabled() bool
    GetConfig() *Config
    EnhanceChangelog(ctx context.Context, changelog string) (string, error)
    GenerateCommitMessage(ctx context.Context, diff string) (*usecases.CommitDetails, error)
    SuggestCommitType(ctx context.Context, diff string) (usecases.CommitType, error)
    GetClient() Client
}
```

### Client Interface

The `Client` interface in `client.go` defines the low-level operations for interacting with AI models:

```go
// Client defines the interface for AI providers
type Client interface {
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    Chat(ctx context.Context, messages []Message, options ...ChatOption) (*CompletionResponse, error)
    GetProvider() Provider
    GetModelName() string
    GetCapabilities() ModelCapabilities
}
```

### Use Cases

The use cases in the `usecases/` directory contain domain-specific AI functionality:

- **Changelog**: Enhancing changelogs, categorizing commits, etc.
- **Commit**: Generating commit messages, suggesting commit types, etc.

## Usage Example

```go
import (
    "context"
    "clikd/internal/config"
    "clikd/internal/services/ai"
)

func main() {
    // Load AI configuration
    cfg, err := config.EnsureInitialized()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Create AI service
    ctx := context.Background()
    aiCfg, err := ai.LoadConfig(cfg.Viper)
    if err != nil {
        log.Fatalf("Failed to load AI configuration: %v", err)
    }
    
    service, err := ai.NewService(ctx, aiCfg)
    if err != nil {
        log.Fatalf("Failed to create AI service: %v", err)
    }
    
    // Use the service
    if service.IsEnabled() {
        enhancedChangelog, err := service.EnhanceChangelog(ctx, "## [1.0.0] - 2023-01-01\n- Fixed bugs\n- Added features")
        if err != nil {
            log.Fatalf("Failed to enhance changelog: %v", err)
        }
        fmt.Println(enhancedChangelog)
    }
}
```

## Design Principles

1. **Clean Architecture**: Clear separation between service interface, clients, and use cases.
2. **Interface Segregation**: Domain-specific interfaces in use cases.
3. **Dependency Inversion**: The service depends on abstractions, not concrete implementations.
4. **Graceful Degradation**: All methods check if AI is enabled before making API calls.
5. **Modularity**: New use cases can be added without modifying existing code. 
