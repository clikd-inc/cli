# CLIKD Changelog Service

The Changelog Service is a powerful, flexible system for generating beautiful changelogs from Git repositories. It provides a clean architecture for extracting, processing, and formatting commit history into standardized changelog documents with AI-enhanced descriptions.

## Architecture

The service follows a modular architecture with clear separation of concerns:

```
internal/services/changelog/
├── service.go             # Main service interface and implementation
├── chglog.go              # Core changelog generation engine
├── types.go               # Data structures for changelog generation
├── config.go              # Configuration models and normalization
├── config_loader.go       # Loading and parsing configuration
├── config_builder.go      # Building configuration from templates
├── processor.go           # Processors for different platforms (GitHub, GitLab, etc.)
├── processor_factory.go   # Factory for creating processors
├── template_builder.go    # Interface for template builders
├── kac_template_builder.go # Keep-a-Changelog style templates
├── custom_template_builder.go # Custom style templates
├── variables.go           # Constants and shared variables
├── jira.go                # Jira integration
├── validation.go          # Configuration validation
└── context.go             # Context for changelog generation
```

## Core Components

### Service Interface

The central `Service` interface in `service.go` provides high-level access to changelog functionality:

```go
// Service provides functions for changelog management
type Service struct {
    ConfigPath string
    factory    ServiceFactoryInterface // Optional factory for AI service injection
}

// Key methods:
// - GenerateChangelog(ctx context.Context, options *GenerationOptions) error
// - PrepareGeneration(ctx context.Context, options *GenerationOptions) (*GenerationResult, error)
// - InitializeTemplates(style string, configDir string) error
// - EnsureTemplateExists(templatePath, style string) error
// - EnsureConfigExists(configPath, style string) error
```

### Generator

The `Generator` in `chglog.go` is the core engine responsible for extracting commit information and rendering the changelog:

```go
// Generator of CHANGELOG
type Generator struct {
    client     gitcmd.Client
    config     *Config
    gitService git.Service
    aiService  ai.Service // Optional AI service for enhancement
    logger     utils.Logger
}

// Key methods:
// - Generate(w io.Writer, query string) error
// - SetAIService(aiService ai.Service)
```

### Template Builders

The service supports multiple template styles through the `TemplateBuilder` interface:

- **Keep-a-Changelog**: Standard format following [keepachangelog.com](https://keepachangelog.com) conventions
- **Custom**: Customizable templates with different styling options

### Processors

Processors handle platform-specific formatting and link generation:

- **GitHub**: Optimized for GitHub-style changelogs
- **GitLab**: Optimized for GitLab-style changelogs
- **Bitbucket**: Optimized for Bitbucket-style changelogs

## Key Features

### AI Enhancement

The service integrates with the AI service to enhance commit messages:

- Splits complex commits into individual changelog entries
- Improves clarity and readability
- Maintains conventional commit format
- Processes commits in efficient batches

### Multiple Template Styles

Supports multiple template styles out of the box:

- **Standard**: Clean, simple format with emoji support
- **Keep-a-Changelog**: Industry standard format following [keepachangelog.com](https://keepachangelog.com)
- **Cool**: Modern, visually appealing format

### Flexible Filtering

Comprehensive filtering options:

- Tag-based filtering (e.g., `v1.0.0..v2.0.0`)
- Path-based filtering
- Commit type filtering
- Case sensitivity options

### Integration Options

Built-in support for external integrations:

- **Jira**: Link issues to Jira tickets
- **GitHub/GitLab/Bitbucket**: Automatic link formatting
- **Custom Processors**: Extensible processing pipeline

## Usage Example

```go
import (
    "context"
    "clikd/internal/services/changelog"
)

func main() {
    // Create changelog service
    service := changelog.NewService("./changelog")
    
    // Generate changelog
    ctx := context.Background()
    err := service.GenerateChangelog(ctx, &changelog.GenerationOptions{
        ConfigPath:    "./changelog/config.yml",
        Template:      "CHANGELOG.tpl.md",
        RepositoryURL: "https://github.com/nyxb/clikd",
        OutputPath:    "CHANGELOG.md",
        Query:         "v1.0.0..v2.0.0",
        NoAI:          false,
    })
    
    if err != nil {
        log.Fatalf("Failed to generate changelog: %v", err)
    }
}
```

## Configuration

The service uses a YAML configuration format:

```yaml
bin: git
template: CHANGELOG.tpl.md
style: github
info:
  title: CHANGELOG
  repository_url: https://github.com/nyxb/clikd
options:
  tag_filter_pattern: "^v"
  sort: "semver"  # or "date"
  commits:
    filters:
      Type:
        - feat
        - fix
        - perf
        - refactor
    sort_by: Scope
  commit_groups:
    group_by: Type
    sort_by: Title
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
    pattern_maps:
      - Type
      - Scope
      - Subject
  notes:
    keywords:
      - BREAKING CHANGE
```

## Template System

The template system uses Go's text/template with additional helpers from sprig:

```
{{ range .Versions }}
## {{ .Tag.Name }} - {{ datetime "2006-01-02" .Tag.Date }}

{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}
{{ end -}}
```

## Performance Optimizations

The service includes several performance optimizations:

1. **Batch Processing**: Processes multiple commits in a single AI call
2. **Caching**: Caches parsed commits and templates
3. **Parallel Processing**: Processes commit groups in parallel
4. **Efficient Filtering**: Uses regex-based filtering for speed
5. **Minimal Dependencies**: Keeps external dependencies to a minimum

## Design Principles

1. **Modularity**: Clear separation between components
2. **Extensibility**: Easy to add new template styles and processors
3. **Performance**: Optimized for large repositories
4. **Compatibility**: Works with multiple Git hosting platforms
5. **Flexibility**: Configurable to match different project needs
6. **AI Integration**: Optional AI enhancement for better readability 
