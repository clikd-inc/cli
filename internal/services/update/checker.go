package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// GithubRelease represents a GitHub release
type GithubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	URL     string `json:"html_url"`
}

// CheckForUpdates checks if a newer version is available
func CheckForUpdates(ctx context.Context, currentVersion string) (bool, string, string, error) {
	// Get repository owner and name from environment or use defaults
	repoOwner := os.Getenv("CLIKD_REPO_OWNER")
	if repoOwner == "" {
		repoOwner = "clikd-inc"
	}

	repoName := os.Getenv("CLIKD_REPO_NAME")
	if repoName == "" {
		repoName = "cli"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Build GitHub API URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, "", "", err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "clikd-update-checker")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return false, "", "", fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}

	// Parse response
	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", "", err
	}

	// Normalize versions for semver comparison
	currentSemver := normalizeVersion(currentVersion)
	latestSemver := normalizeVersion(release.TagName)

	// Skip update check for development versions
	if currentVersion == "dev" || currentVersion == "development" {
		return false, strings.TrimPrefix(release.TagName, "v"), release.URL, nil
	}

	// Use semver to compare versions properly
	// semver.Compare returns:
	// -1 if current < latest (update available)
	//  0 if current == latest (no update)
	//  1 if current > latest (current is newer)
	hasUpdate := semver.Compare(currentSemver, latestSemver) < 0

	return hasUpdate, strings.TrimPrefix(release.TagName, "v"), release.URL, nil
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
