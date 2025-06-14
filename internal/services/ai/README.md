# CLIKD AI Service

The AI Service provides a clean, efficient, and unified interface for integrating AI capabilities into the CLIKD CLI. It follows a performance-optimized architecture designed for CLI environments where speed and resource efficiency are critical.

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
    └── changelog_test.go # Tests for changelog functionality
```

## Core Components

### Service Interface

The central `Service` interface in `service.go` provides high-level access to AI capabilities:

```go
// Service defines the interface for AI operations
type Service interface {
    // Commit-related operations - batch processing only
    EnhanceCommitMessagesBatch(commitMessages []string) (map[string][]string, error)
}
```

The service is designed for maximum performance, focusing on batch processing rather than individual operations to minimize API calls and latency.

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

### Multi-Provider Support

The service supports multiple AI providers through a unified interface:

- **OpenAI** - GPT models
- **Mistral AI** - Mistral models
- **Anthropic** - Claude models
- **Groq** - Fast inference models
- **OpenRouter** - Meta-provider with fallback capabilities
- **Local** - Ollama for local model execution

### Use Cases

Domain-specific AI functionality is organized in the `usecases/` directory:

- **Changelog** - Enhancing commit messages for better changelog generation
- Future use cases can be added without modifying existing code

## Performance Optimizations

The AI service is highly optimized for CLI environments:

1. **Batch Processing** - Processes multiple items in a single API call
2. **Low Temperature Settings** - Uses 0.1 temperature for faster, deterministic responses
3. **Optimized Token Selection** - Uses topP=0.7 for more focused output
4. **Minimal Logging** - Only logs essential information
5. **Error Resilience** - Gracefully handles API failures with fallbacks
6. **Memory Management** - Optimized for low memory footprint

## Usage Example

```go
import (
    "context"
    "clikd/internal/config"
    "clikd/internal/services/ai"
)

func main() {
    // Load global configuration
    cfg, err := config.EnsureInitialized()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Create AI service with parameters from global configuration
    ctx := context.Background()
    service, err := ai.NewService(
        ctx,
        cfg.AI.Provider,         // provider
        cfg.AI.Model,            // model
        cfg.AI.APIKey,           // apiKey
        cfg.AI.APIURL,           // endpoint
        cfg.AI.TokensMaxInput,   // tokensMaxInput
        cfg.AI.TokensMaxOutput,  // tokensMaxOutput
    )
    if err != nil {
        log.Fatalf("Failed to create AI service: %v", err)
    }
    
    // Process multiple commit messages in a single batch for optimal performance
    commitMessages := []string{
        "feat(auth): add login endpoint and fix validation bugs and update tests",
        "fix(api): resolve timeout issue in authentication service",
        "chore: update dependencies to latest versions"
    }
    
    enhancedMessages, err := service.EnhanceCommitMessagesBatch(commitMessages)
    if err != nil {
        log.Printf("Warning: AI enhancement failed: %v", err)
        // Service will return original messages on error - no need for additional fallback
    }
    
    // enhancedMessages is a map[string][]string where:
    // - key: original commit message
    // - value: array of enhanced/split messages
}
```

## Configuration

The AI service configuration is managed through the global config system:

```toml
# Example configuration in config.toml
[ai]
provider = "mistral"
model = "mistral-large-latest"
api_key = "your-api-key"  # For global config only
tokens_max_input = 4000
tokens_max_output = 1000
```

For local projects, API keys should be stored in a `.env` file:

```
CLIKD_API_KEY=your-api-key
```

## Implementation Details

### Gollm Integration

The service uses the [Gollm](https://github.com/teilomillet/gollm) library as a unified client for all AI providers, providing:

- Consistent interface across providers
- Automatic retries for API failures
- Proper error handling
- Memory management for context windows

### Error Handling

The service is designed for graceful degradation:

- Returns original content when AI enhancement fails
- Provides detailed error logging for debugging
- Implements timeouts to prevent hanging operations

## Testing

The AI service includes comprehensive tests:

- Unit tests for all components
- Mock clients for testing without API calls
- Integration tests for end-to-end verification

## Design Principles

1. **Performance First** - Optimized for CLI environments where speed is critical
2. **Batch Processing** - Minimizes API calls by processing multiple items at once
3. **Clean Architecture** - Clear separation between service, clients, and use cases
4. **Interface Segregation** - Domain-specific interfaces in use cases
5. **Dependency Inversion** - The service depends on abstractions, not concrete implementations
6. **Graceful Degradation** - Returns original content when AI enhancement fails
7. **Modularity** - New use cases can be added without modifying existing code
