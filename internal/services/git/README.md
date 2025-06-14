# CLIKD Git Service

The Git Service provides a robust, high-performance interface for interacting with Git repositories. It offers a clean architecture for extracting, parsing, and processing Git data with advanced filtering and formatting capabilities.

## Architecture

The service follows a modular architecture with clear separation of concerns:

```
internal/services/git/
├── service.go           # Main service interface and implementation
├── fields.go            # Core data structures and models
├── commit_parser.go     # Parsing Git commit data
├── commit_extractor.go  # Extracting and grouping commits
├── tag_reader.go        # Reading and processing Git tags
├── tag_selector.go      # Selecting tags based on queries
├── commit_filter.go     # Filtering commits based on criteria
├── errors.go            # Service-specific error types
```

## Core Components

### Service Interface

The central `Service` interface in `service.go` provides high-level access to Git functionality:

```go
// Service defines the main interface for Git functionality
type Service interface {
    // Core repository operations
    GetRepositoryRoot() (string, error)
    IsGitRepository() (bool, error)
    HasRemote() (bool, error)
    GetCurrentBranch() (string, error)

    // Commit operations
    GetCommits(revision string, paths []string) ([]*Commit, error)
    GetCommitsWithOptions(options CommitOptions) ([]*Commit, error)
    GetCommitsBetweenTags(fromTag, toTag string) ([]*Commit, error)
    GetStagedChanges() (string, error)

    // Tag operations
    GetLatestTag() (string, error)
    GetTags() ([]string, error)
    GetTagsWithPattern(pattern string) ([]string, error)
    GetAllTagsWithDetails() ([]*Tag, error)
    CreateTag(tag string, message string) error

    // High-level operations
    SelectTagRange(query string) (string, string, error)
    SelectTagsWithQuery(tags []*Tag, query string) ([]*Tag, string, error)
    ExtractCommits(commits []*Commit, opts *Options) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup)
    GetDiffBetweenTags(fromTag, toTag string) (string, error)
    
    // Analysis operations
    AnalyzeRepository(ctx context.Context) (*RepositoryInfo, error)
    GetChangelogData(ctx context.Context, options *ChangelogOptions) (*ChangelogData, error)
}
```

### Data Models

The service uses a rich set of data models defined in `fields.go`:

- **Commit**: Represents a Git commit with detailed metadata
- **Tag**: Represents a Git tag with related tag information
- **CommitGroup**: Groups commits by configurable criteria
- **NoteGroup**: Groups commit notes by title

### Commit Parser

The `commitParser` in `commit_parser.go` handles the complex task of parsing Git commit data:

- Extracts metadata like authors, committers, and timestamps
- Parses conventional commit format (type, scope, subject)
- Identifies merge and revert commits
- Extracts references, mentions, and notes
- Supports Jira issue ID extraction

### Tag Reader & Selector

The tag components provide powerful tag handling:

- **tagReader**: Reads tags with filtering and sorting options
- **tagSelector**: Selects tags based on complex queries like `v1.0.0..v2.0.0`

## Key Features

### Advanced Commit Parsing

The service provides sophisticated commit parsing capabilities:

- **Conventional Commits**: Full support for the conventional commit format
- **Co-authors & Signers**: Extracts co-author and signed-off-by information
- **References & Mentions**: Identifies issue references and @mentions
- **Notes & Breaking Changes**: Extracts structured notes from commit messages
- **Jira Integration**: Parses Jira issue IDs and can fetch additional information

### Flexible Tag Handling

Comprehensive tag management features:

- **Tag Filtering**: Filter tags using regular expressions
- **Tag Sorting**: Sort tags by date or semantic version
- **Tag Queries**: Select tags using range queries (`v1.0.0..v2.0.0`, `v1.0.0..`, `..v2.0.0`)
- **Tag Details**: Access tag metadata including dates and related tags

### Powerful Commit Grouping

Highly configurable commit grouping system:

- **Group By**: Group commits by any field (type, scope, author, etc.)
- **Sort By**: Sort commits and groups by any field
- **Title Maps**: Map raw group titles to formatted display titles
- **Custom Sorting**: Support for custom sort orders

### Repository Analysis

Comprehensive repository analysis capabilities:

- **Repository Info**: Extract key repository metadata
- **Changelog Data**: Generate structured data for changelog generation
- **Path Filtering**: Filter commits by file paths
- **Pattern Matching**: Use regular expressions for advanced filtering

## Usage Examples

### Basic Repository Operations

```go
import (
    "context"
    "clikd/internal/services/git"
)

func main() {
    // Create git service with current directory
    service, err := git.NewService()
    if err != nil {
        log.Fatalf("Failed to create git service: %v", err)
    }
    
    // Check if current directory is a git repository
    isRepo, err := service.IsGitRepository()
    if err != nil || !isRepo {
        log.Fatalf("Current directory is not a git repository")
    }
    
    // Get current branch
    branch, err := service.GetCurrentBranch()
    if err != nil {
        log.Fatalf("Failed to get current branch: %v", err)
    }
    fmt.Printf("Current branch: %s\n", branch)
    
    // Get repository root
    root, err := service.GetRepositoryRoot()
    if err != nil {
        log.Fatalf("Failed to get repository root: %v", err)
    }
    fmt.Printf("Repository root: %s\n", root)
}
```

### Working with Tags

```go
// Get all tags
tags, err := service.GetTags()
if err != nil {
    log.Fatalf("Failed to get tags: %v", err)
}
fmt.Printf("Found %d tags\n", len(tags))

// Get tags matching a pattern
versionTags, err := service.GetTagsWithPattern("^v[0-9]")
if err != nil {
    log.Fatalf("Failed to get version tags: %v", err)
}
fmt.Printf("Found %d version tags\n", len(versionTags))

// Get latest tag
latestTag, err := service.GetLatestTag()
if err != nil {
    log.Fatalf("Failed to get latest tag: %v", err)
}
fmt.Printf("Latest tag: %s\n", latestTag)

// Select tags with a query
tagDetails, err := service.GetAllTagsWithDetails()
if err != nil {
    log.Fatalf("Failed to get tag details: %v", err)
}
selectedTags, firstTag, err := service.SelectTagsWithQuery(tagDetails, "v1.0.0..v2.0.0")
if err != nil {
    log.Fatalf("Failed to select tags: %v", err)
}
```

### Analyzing Commits

```go
// Get commits for a specific revision
commits, err := service.GetCommits("HEAD~10..HEAD", []string{})
if err != nil {
    log.Fatalf("Failed to get commits: %v", err)
}

// Configure options for commit extraction
options := &git.Options{
    CommitGroupBy:        "Type",
    CommitGroupSortBy:    "Title",
    CommitGroupTitleMaps: map[string]string{
        "feat":     "Features",
        "fix":      "Bug Fixes",
        "perf":     "Performance Improvements",
        "refactor": "Code Refactoring",
    },
    HeaderPattern:     "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$",
    HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
    NoteKeywords:      []string{"BREAKING CHANGE"},
}

// Extract and group commits
commitGroups, mergeCommits, revertCommits, noteGroups := service.ExtractCommits(commits, options)

// Process commit groups
for _, group := range commitGroups {
    fmt.Printf("Group: %s (%d commits)\n", group.Title, len(group.Commits))
    for _, commit := range group.Commits {
        fmt.Printf("  - %s: %s\n", commit.Hash.Short, commit.Subject)
    }
}
```

### Repository Analysis

```go
// Analyze repository
ctx := context.Background()
repoInfo, err := service.AnalyzeRepository(ctx)
if err != nil {
    log.Fatalf("Failed to analyze repository: %v", err)
}
fmt.Printf("Repository analysis:\n")
fmt.Printf("  - Root: %s\n", repoInfo.Root)
fmt.Printf("  - Current branch: %s\n", repoInfo.CurrentBranch)
fmt.Printf("  - Latest tag: %s\n", repoInfo.LatestTag)
fmt.Printf("  - Total tags: %d\n", repoInfo.TotalTags)
fmt.Printf("  - Total commits: %d\n", repoInfo.TotalCommits)

// Get changelog data
options := &git.ChangelogOptions{
    Query:            "v1.0.0..v2.0.0",
    TagFilterPattern: "^v",
    TagSortBy:        "semver",
}
changelogData, err := service.GetChangelogData(ctx, options)
if err != nil {
    log.Fatalf("Failed to get changelog data: %v", err)
}
```

## Performance Considerations

The Git Service includes several performance optimizations:

1. **Efficient Parsing**: Uses optimized Git log format for faster parsing
2. **Caching**: Caches parsed commits and tags where appropriate
3. **Path Filtering**: Uses Git's native path filtering for better performance
4. **Batch Processing**: Processes commits in batches when possible
5. **Regex Optimization**: Uses compiled regular expressions for better performance

## Design Principles

1. **Modularity**: Clear separation between components
2. **Extensibility**: Easy to add new functionality
3. **Performance**: Optimized for large repositories
4. **Robustness**: Comprehensive error handling
5. **Flexibility**: Configurable to match different project needs
