package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"clikd/internal/utils"

	"golang.org/x/mod/semver"
)

// Service defines the interface for update checking functionality
type Service interface {
	// CheckForUpdates checks if a newer version is available
	CheckForUpdates(ctx context.Context, currentVersion string) (bool, string, string, error)

	// GetLatestRelease gets the latest release information
	GetLatestRelease(ctx context.Context) (*ReleaseInfo, error)

	// CompareVersions compares two version strings using semantic versioning
	CompareVersions(current, latest string) (int, error)
}

// ServiceImpl is the concrete implementation of the Service interface
type ServiceImpl struct {
	repoOwner  string
	repoName   string
	baseURL    string
	httpClient *http.Client
	logger     utils.Logger
}

// UpdateOptions contains configuration for the update service
type UpdateOptions struct {
	RepoOwner string
	RepoName  string
	BaseURL   string // Optional base URL for testing
	Timeout   time.Duration
}

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

// NewService creates a new update service with default configuration
func NewService() Service {
	return NewServiceWithOptions(&UpdateOptions{
		RepoOwner: "clikd-inc",
		RepoName:  "cli",
		Timeout:   5 * time.Second,
	}, nil)
}

// NewServiceWithOptions creates a new update service with custom options
func NewServiceWithOptions(options *UpdateOptions, logger utils.Logger) Service {
	if options == nil {
		options = &UpdateOptions{
			RepoOwner: "clikd-inc",
			RepoName:  "cli",
			Timeout:   5 * time.Second,
		}
	}

	if logger == nil {
		logger = utils.DefaultLogger.WithFields(map[string]interface{}{"module": "update"})
	}

	return &ServiceImpl{
		repoOwner: options.RepoOwner,
		repoName:  options.RepoName,
		baseURL:   options.BaseURL,
		httpClient: &http.Client{
			Timeout: options.Timeout,
		},
		logger: logger,
	}
}

// CheckForUpdates implements the Service interface
func (s *ServiceImpl) CheckForUpdates(ctx context.Context, currentVersion string) (bool, string, string, error) {
	s.logger.Debug("Checking for updates", "currentVersion", currentVersion)

	release, err := s.GetLatestRelease(ctx)
	if err != nil {
		return false, "", "", err
	}

	// Normalize versions for semver comparison
	currentSemver := normalizeVersion(currentVersion)
	latestSemver := normalizeVersion(release.TagName)

	// Skip update check for development versions
	if currentVersion == "dev" || currentVersion == "development" {
		s.logger.Debug("Development version detected, skipping update check")
		return false, strings.TrimPrefix(release.TagName, "v"), release.URL, nil
	}

	// Use semver to compare versions properly
	// semver.Compare returns:
	// -1 if current < latest (update available)
	//  0 if current == latest (no update)
	//  1 if current > latest (current is newer)
	hasUpdate := semver.Compare(currentSemver, latestSemver) < 0

	latestVersionClean := strings.TrimPrefix(release.TagName, "v")

	s.logger.Debug("Update check complete", "hasUpdate", hasUpdate, "latest", latestVersionClean)
	return hasUpdate, latestVersionClean, release.URL, nil
}

// GetLatestRelease implements the Service interface
func (s *ServiceImpl) GetLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	s.logger.Debug("Fetching latest release from GitHub API")

	// Build GitHub API URL - use custom base URL if provided (for testing)
	var url string
	if s.baseURL != "" {
		url = fmt.Sprintf("%s/repos/%s/%s/releases/latest", s.baseURL, s.repoOwner, s.repoName)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", s.repoOwner, s.repoName)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "clikd-update-checker")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}

	// Parse response
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	s.logger.Debug("Latest release fetched", "tagName", release.TagName)
	return &release, nil
}

// CompareVersions implements the Service interface
func (s *ServiceImpl) CompareVersions(current, latest string) (int, error) {
	currentSemver := normalizeVersion(current)
	latestSemver := normalizeVersion(latest)

	if !semver.IsValid(currentSemver) {
		return 0, fmt.Errorf("invalid current version: %s", current)
	}

	if !semver.IsValid(latestSemver) {
		return 0, fmt.Errorf("invalid latest version: %s", latest)
	}

	return semver.Compare(currentSemver, latestSemver), nil
}

// Legacy function for backward compatibility
// Deprecated: Use Service.CheckForUpdates instead
func CheckForUpdates(ctx context.Context, currentVersion string) (bool, string, string, error) {
	service := NewService()
	return service.CheckForUpdates(ctx, currentVersion)
}

// normalizeVersion ensures the version string is valid for semver comparison
func normalizeVersion(version string) string {
	// Handle development versions first
	if version == "dev" || version == "development" {
		return "v0.0.0-dev"
	}

	// Ensure version starts with 'v'
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Handle development versions with v prefix
	if version == "vdev" || version == "vdevelopment" {
		return "v0.0.0-dev"
	}

	// If already valid, return as is
	if semver.IsValid(version) {
		return version
	}

	// If not valid semver, try to make it valid
	// Remove the 'v' prefix for processing
	cleaned := strings.TrimPrefix(version, "v")

	// Handle empty or invalid input
	if cleaned == "" || cleaned == "invalid" {
		return "v0.0.0"
	}

	// Split into parts
	parts := strings.Split(cleaned, ".")

	// Filter out non-numeric parts and ensure we have valid numbers
	validParts := make([]string, 0, 3)
	for i, part := range parts {
		// Only take the first 3 parts
		if i >= 3 {
			break
		}

		// Check if part is numeric (basic check)
		if part != "" && isNumeric(part) {
			validParts = append(validParts, part)
		}
	}

	// Ensure we have at least 3 parts (major.minor.patch)
	for len(validParts) < 3 {
		validParts = append(validParts, "0")
	}

	normalized := "v" + strings.Join(validParts, ".")

	// Final validation
	if semver.IsValid(normalized) {
		return normalized
	}

	// Fallback to v0.0.0 if we can't normalize
	return "v0.0.0"
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
