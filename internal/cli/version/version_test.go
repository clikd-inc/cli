package version

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name           string
		setupVersion   string
		expectedResult string
	}{
		{
			name:           "default development version",
			setupVersion:   "dev",
			expectedResult: "dev",
		},
		{
			name:           "release version",
			setupVersion:   "1.2.3",
			expectedResult: "1.2.3",
		},
		{
			name:           "version with v prefix",
			setupVersion:   "v2.0.0",
			expectedResult: "v2.0.0",
		},
		{
			name:           "empty version",
			setupVersion:   "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original version
			originalVersion := Version

			// Set test version
			Version = tt.setupVersion

			// Test GetVersion
			result := GetVersion()

			// Verify result
			if result != tt.expectedResult {
				t.Errorf("GetVersion() = %v, want %v", result, tt.expectedResult)
			}

			// Restore original version
			Version = originalVersion
		})
	}
}

func TestVersionVariable(t *testing.T) {
	// Test that Version variable exists and has expected default
	if Version == "" {
		t.Error("Version variable should not be empty by default")
	}

	// Test that we can modify the Version variable
	originalVersion := Version
	testVersion := "test-version-123"

	Version = testVersion
	if Version != testVersion {
		t.Errorf("Version variable assignment failed, got %v, want %v", Version, testVersion)
	}

	// Restore original
	Version = originalVersion
}

func TestGetVersionConsistency(t *testing.T) {
	// Test that GetVersion() returns the same as direct access to Version
	if GetVersion() != Version {
		t.Errorf("GetVersion() = %v, but Version = %v, they should be equal", GetVersion(), Version)
	}

	// Test multiple calls return consistent results
	first := GetVersion()
	second := GetVersion()

	if first != second {
		t.Errorf("GetVersion() returned inconsistent results: first=%v, second=%v", first, second)
	}
}

// TestVersionLinkerIntegration tests that the version can be set via linker flags
// This test verifies the mechanism works, but the actual linker flag setting
// happens during the build process
func TestVersionLinkerIntegration(t *testing.T) {
	// Save original version
	originalVersion := Version

	// Simulate what the linker would do
	linkerSetVersion := "1.0.0-linker-test"
	Version = linkerSetVersion

	// Verify the version was set correctly
	if GetVersion() != linkerSetVersion {
		t.Errorf("Linker version setting failed, got %v, want %v", GetVersion(), linkerSetVersion)
	}

	// Restore original version
	Version = originalVersion
}
