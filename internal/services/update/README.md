# CLIKD Update Service

The Update Service provides a reliable, efficient mechanism for checking and notifying users about available updates to the CLIKD CLI. It follows semantic versioning principles and provides a clean API for version comparison and update notifications.

## Architecture

The service follows a clean architecture with clear separation of concerns:

```
internal/services/update/
├── checker.go       # Main service implementation for update checking
├── checker_test.go  # Comprehensive tests for the update service
```

## Core Components

### Service Interface

The central `Service` interface in `checker.go` provides high-level access to update functionality:

```go
// Service defines the interface for update checking functionality
type Service interface {
    // CheckForUpdates checks if a newer version is available
    CheckForUpdates(ctx context.Context, currentVersion string) (bool, string, string, error)

    // GetLatestRelease gets the latest release information
    GetLatestRelease(ctx context.Context) (*ReleaseInfo, error)

    // CompareVersions compares two version strings using semantic versioning
    CompareVersions(current, latest string) (int, error)
}
```

### Data Models

The service uses several key data structures:

```go
// ReleaseInfo contains information about a release
type ReleaseInfo struct {
    TagName string `json:"tag_name"`
    Name    string `json:"name"`
    Body    string `json:"body"`
    URL     string `json:"html_url"`
}

// UpdateResult contains the result of an update check
type UpdateResult struct {
    HasUpdate      bool
    CurrentVersion string
    LatestVersion  string
    ReleaseURL     string
    Release        *ReleaseInfo
}

// UpdateOptions contains configuration for the update service
type UpdateOptions struct {
    RepoOwner string
    RepoName  string
    BaseURL   string // Optional base URL for testing
    Timeout   time.Duration
}
```

## Key Features

### Semantic Version Handling

The service provides robust semantic versioning support:

- **Version Normalization**: Converts various version formats to valid semver
- **Proper Comparison**: Uses `golang.org/x/mod/semver` for accurate version comparison
- **Development Version Handling**: Special handling for development versions
- **Flexible Input**: Handles versions with or without 'v' prefix

### GitHub Integration

Built-in GitHub API integration for release information:

- **Latest Release Detection**: Fetches latest release information from GitHub
- **Release Details**: Provides access to release notes, version numbers, and URLs
- **HTTP Client Configuration**: Configurable timeouts and error handling
- **Custom Endpoints**: Support for custom API endpoints for testing

### Robust Error Handling

Comprehensive error handling for all operations:

- **Timeout Handling**: Configurable timeouts for API requests
- **HTTP Error Detection**: Proper handling of HTTP error codes
- **JSON Parsing Errors**: Graceful handling of malformed API responses
- **Context Support**: All operations respect context cancellation

## Usage Examples

### Basic Update Check

```go
import (
    "context"
    "fmt"
    "clikd/internal/services/update"
)

func checkForUpdates() {
    // Create a new update service with default configuration
    service := update.NewService()
    
    // Check for updates
    ctx := context.Background()
    currentVersion := "1.0.0"
    
    hasUpdate, latestVersion, releaseURL, err := service.CheckForUpdates(ctx, currentVersion)
    if err != nil {
        fmt.Printf("Error checking for updates: %v\n", err)
        return
    }
    
    if hasUpdate {
        fmt.Printf("Update available: %s -> %s\n", currentVersion, latestVersion)
        fmt.Printf("Download: %s\n", releaseURL)
    } else {
        fmt.Println("You are using the latest version.")
    }
}
```

### Custom Configuration

```go
// Create a service with custom configuration
options := &update.UpdateOptions{
    RepoOwner: "custom-org",
    RepoName:  "custom-repo",
    Timeout:   10 * time.Second,
}

service := update.NewServiceWithOptions(options, logger)

// Use the service
release, err := service.GetLatestRelease(ctx)
if err != nil {
    // Handle error
}

fmt.Printf("Latest release: %s\n", release.TagName)
fmt.Printf("Release notes: %s\n", release.Body)
```

### Version Comparison

```go
// Compare two versions
service := update.NewService()
result, err := service.CompareVersions("1.0.0", "1.1.0")
if err != nil {
    // Handle error
}

switch result {
case -1:
    fmt.Println("First version is older")
case 0:
    fmt.Println("Versions are equal")
case 1:
    fmt.Println("First version is newer")
}
```

## Version Normalization

The service includes sophisticated version normalization logic:

1. **Development Versions**: `dev` or `development` are treated as `v0.0.0-dev`
2. **V Prefix**: Versions without a 'v' prefix have it added automatically
3. **Incomplete Versions**: Versions with fewer than 3 parts (e.g., `1.2`) are padded with zeros
4. **Extra Parts**: Versions with more than 3 parts have the extra parts trimmed
5. **Invalid Versions**: Invalid versions are normalized to `v0.0.0`

## Testing

The service includes comprehensive tests:

- **Unit Tests**: Tests for all public methods and internal functions
- **Mock HTTP Server**: Uses `httptest` package to mock GitHub API responses
- **Edge Cases**: Tests for timeout handling, HTTP errors, and invalid JSON
- **Version Normalization**: Tests for various version formats and edge cases

## Performance Considerations

The update service is designed for efficiency:

1. **Configurable Timeouts**: Prevents hanging on slow connections
2. **Minimal Dependencies**: Uses only the standard library and `golang.org/x/mod/semver`
3. **Efficient HTTP**: Uses a single HTTP client instance
4. **Context Support**: All operations respect context cancellation for proper resource management

## Design Principles

1. **Clean Interface**: Simple, intuitive interface for update checking
2. **Robustness**: Comprehensive error handling and version normalization
3. **Testability**: Designed for easy testing with dependency injection
4. **Flexibility**: Configurable for different repositories and environments
5. **Backward Compatibility**: Maintains legacy functions for backward compatibility 
