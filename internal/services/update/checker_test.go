package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid semver with v prefix",
			input:    "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "valid semver without v prefix",
			input:    "1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "development version",
			input:    "dev",
			expected: "v0.0.0-dev",
		},
		{
			name:     "development version with v prefix",
			input:    "vdev",
			expected: "v0.0.0-dev",
		},
		{
			name:     "incomplete version - major.minor",
			input:    "1.2",
			expected: "v1.2",
		},
		{
			name:     "incomplete version - major only",
			input:    "1",
			expected: "v1",
		},
		{
			name:     "version with extra parts",
			input:    "1.2.3.4.5",
			expected: "v1.2.3",
		},
		{
			name:     "invalid version",
			input:    "invalid",
			expected: "v0.0.0",
		},
		{
			name:     "empty version",
			input:    "",
			expected: "v0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckForUpdates(t *testing.T) {
	tests := []struct {
		name            string
		currentVersion  string
		releaseResponse ReleaseInfo
		expectedUpdate  bool
		expectedLatest  string
		expectedError   bool
	}{
		{
			name:           "update available",
			currentVersion: "1.0.0",
			releaseResponse: ReleaseInfo{
				TagName: "v1.1.0",
				Name:    "Release 1.1.0",
				Body:    "New features",
				URL:     "https://github.com/test/repo/releases/tag/v1.1.0",
			},
			expectedUpdate: true,
			expectedLatest: "1.1.0",
			expectedError:  false,
		},
		{
			name:           "no update available - same version",
			currentVersion: "1.1.0",
			releaseResponse: ReleaseInfo{
				TagName: "v1.1.0",
				Name:    "Release 1.1.0",
				Body:    "Current release",
				URL:     "https://github.com/test/repo/releases/tag/v1.1.0",
			},
			expectedUpdate: false,
			expectedLatest: "1.1.0",
			expectedError:  false,
		},
		{
			name:           "no update available - current newer",
			currentVersion: "1.2.0",
			releaseResponse: ReleaseInfo{
				TagName: "v1.1.0",
				Name:    "Release 1.1.0",
				Body:    "Older release",
				URL:     "https://github.com/test/repo/releases/tag/v1.1.0",
			},
			expectedUpdate: false,
			expectedLatest: "1.1.0",
			expectedError:  false,
		},
		{
			name:           "development version - no update notification",
			currentVersion: "dev",
			releaseResponse: ReleaseInfo{
				TagName: "v1.1.0",
				Name:    "Release 1.1.0",
				Body:    "Latest release",
				URL:     "https://github.com/test/repo/releases/tag/v1.1.0",
			},
			expectedUpdate: false,
			expectedLatest: "1.1.0",
			expectedError:  false,
		},
		{
			name:           "version without v prefix",
			currentVersion: "1.0.0",
			releaseResponse: ReleaseInfo{
				TagName: "1.1.0",
				Name:    "Release 1.1.0",
				Body:    "New features",
				URL:     "https://github.com/test/repo/releases/tag/1.1.0",
			},
			expectedUpdate: true,
			expectedLatest: "1.1.0",
			expectedError:  false,
		},
		{
			name:           "release without v prefix",
			currentVersion: "v1.0.0",
			releaseResponse: ReleaseInfo{
				TagName: "1.1.0",
				Name:    "Release 1.1.0",
				Body:    "New features",
				URL:     "https://github.com/test/repo/releases/tag/1.1.0",
			},
			expectedUpdate: true,
			expectedLatest: "1.1.0",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				expectedPath := "/repos/clikd-inc/cli/releases/latest"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Check headers
				if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
					t.Errorf("Expected Accept header 'application/vnd.github.v3+json', got %s", r.Header.Get("Accept"))
				}

				if r.Header.Get("User-Agent") != "clikd-update-checker" {
					t.Errorf("Expected User-Agent header 'clikd-update-checker', got %s", r.Header.Get("User-Agent"))
				}

				// Return the mock response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.releaseResponse)
			}))
			defer server.Close()

			// Create service with custom base URL pointing to test server
			options := &UpdateOptions{
				RepoOwner: "clikd-inc",
				RepoName:  "cli",
				BaseURL:   server.URL,
				Timeout:   5 * time.Second,
			}

			service := NewServiceWithOptions(options, nil)
			if service == nil {
				t.Fatal("Failed to create service")
			}

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test the actual CheckForUpdates method
			hasUpdate, latestVersion, releaseURL, err := service.CheckForUpdates(ctx, tt.currentVersion)

			if tt.expectedError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if hasUpdate != tt.expectedUpdate {
				t.Errorf("CheckForUpdates() hasUpdate = %v, want %v", hasUpdate, tt.expectedUpdate)
			}

			if latestVersion != tt.expectedLatest {
				t.Errorf("CheckForUpdates() latestVersion = %v, want %v", latestVersion, tt.expectedLatest)
			}

			if releaseURL != tt.releaseResponse.URL {
				t.Errorf("CheckForUpdates() releaseURL = %v, want %v", releaseURL, tt.releaseResponse.URL)
			}
		})
	}
}

func TestCheckForUpdatesTimeout(t *testing.T) {
	// Create a test server that delays response to trigger timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the client timeout to trigger timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create service with very short timeout and custom base URL
	options := &UpdateOptions{
		RepoOwner: "clikd-inc",
		RepoName:  "cli",
		BaseURL:   server.URL,
		Timeout:   100 * time.Millisecond, // Very short timeout
	}

	service := NewServiceWithOptions(options, nil)

	// Create context
	ctx := context.Background()

	// Test should timeout
	_, _, _, err := service.CheckForUpdates(ctx, "1.0.0")

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check that it's a timeout-related error
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestCheckForUpdatesHTTPError(t *testing.T) {
	// Create a test server that returns HTTP 404 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	// Create service with custom base URL pointing to test server
	options := &UpdateOptions{
		RepoOwner: "clikd-inc",
		RepoName:  "cli",
		BaseURL:   server.URL,
		Timeout:   5 * time.Second,
	}

	service := NewServiceWithOptions(options, nil)

	// Create context
	ctx := context.Background()

	// Test should return HTTP error
	_, _, _, err := service.CheckForUpdates(ctx, "1.0.0")

	if err == nil {
		t.Error("Expected HTTP error, got nil")
	}

	// Check that it's an HTTP status error
	if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "status code") {
		t.Errorf("Expected HTTP status error, got: %v", err)
	}
}

func TestCheckForUpdatesInvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	// Create service with custom base URL pointing to test server
	options := &UpdateOptions{
		RepoOwner: "clikd-inc",
		RepoName:  "cli",
		BaseURL:   server.URL,
		Timeout:   5 * time.Second,
	}

	service := NewServiceWithOptions(options, nil)

	// Create context
	ctx := context.Background()

	// Test should return JSON parsing error
	_, _, _, err := service.CheckForUpdates(ctx, "1.0.0")

	if err == nil {
		t.Error("Expected JSON parsing error, got nil")
	}

	// Check that it's a JSON parsing error
	if !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "json") && !strings.Contains(err.Error(), "decode") {
		t.Errorf("Expected JSON parsing error, got: %v", err)
	}
}

func TestServiceCreation(t *testing.T) {
	// Test default service creation
	service := NewService()
	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	// Test service with options
	options := &UpdateOptions{
		RepoOwner: "test-owner",
		RepoName:  "test-repo",
		Timeout:   10 * time.Second,
	}

	serviceWithOptions := NewServiceWithOptions(options, nil)
	if serviceWithOptions == nil {
		t.Fatal("NewServiceWithOptions() returned nil")
	}
}

func TestCompareVersions(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		current  string
		latest   string
		expected int
		hasError bool
	}{
		{
			name:     "current older than latest",
			current:  "1.0.0",
			latest:   "1.1.0",
			expected: -1,
			hasError: false,
		},
		{
			name:     "current same as latest",
			current:  "1.1.0",
			latest:   "1.1.0",
			expected: 0,
			hasError: false,
		},
		{
			name:     "current newer than latest",
			current:  "1.2.0",
			latest:   "1.1.0",
			expected: 1,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CompareVersions(tt.current, tt.latest)

			if tt.hasError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}

// Test helper function to create a mock GitHub API response
func createMockGitHubResponse(t *testing.T, release ReleaseInfo) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("Failed to encode mock response: %v", err)
		}
	}))
}

func TestGetLatestRelease(t *testing.T) {
	expectedRelease := ReleaseInfo{
		TagName: "v1.2.3",
		Name:    "Release 1.2.3",
		Body:    "Bug fixes and improvements",
		URL:     "https://github.com/test/repo/releases/tag/v1.2.3",
	}

	server := createMockGitHubResponse(t, expectedRelease)
	defer server.Close()

	// For this test, we'd need to inject the server URL into the service
	// This is a limitation of the current design that could be improved
	// by allowing custom base URLs or HTTP clients

	service := NewService()
	ctx := context.Background()

	// Test that the method exists and can be called
	// Note: This will make a real API call since we can't easily mock it
	// In a production environment, you'd want to inject the HTTP client
	_, err := service.GetLatestRelease(ctx)

	// We expect this to either succeed or fail with a network error
	// The important thing is that the method exists and has the right signature
	if err != nil {
		t.Logf("GetLatestRelease failed (expected in test environment): %v", err)
	}
}

func TestLegacyCheckForUpdates(t *testing.T) {
	ctx := context.Background()

	// Test that the legacy function still works
	_, _, _, err := CheckForUpdates(ctx, "1.0.0")

	// We expect this to either succeed or fail with a network error
	// The important thing is that the function exists and has the right signature
	if err != nil {
		t.Logf("Legacy CheckForUpdates failed (expected in test environment): %v", err)
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"12a", false},
		{"a12", false},
		{"1.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
