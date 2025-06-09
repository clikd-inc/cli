package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
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

	// Compare versions (simple string comparison for now)
	// Remove v prefix if present
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Check if versions are different - this is a simple implementation
	// In a real application, you'd want to do proper semver comparison
	hasUpdate := latestVersion != currentVersion

	return hasUpdate, latestVersion, release.URL, nil
}
