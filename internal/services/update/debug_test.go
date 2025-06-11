package update

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/mod/semver"
)

func TestDebugNormalizeVersion(t *testing.T) {
	testCases := []string{"1.2", "1", "v1.2", "v1"}

	for _, tc := range testCases {
		fmt.Printf("Input: %q\n", tc)
		fmt.Printf("  semver.IsValid: %v\n", semver.IsValid(tc))

		// Add v prefix if missing
		withV := tc
		if !strings.HasPrefix(tc, "v") {
			withV = "v" + tc
		}
		fmt.Printf("  With v prefix: %q, semver.IsValid: %v\n", withV, semver.IsValid(withV))

		result := normalizeVersion(tc)
		fmt.Printf("  normalizeVersion result: %q\n", result)
		fmt.Printf("  Result is valid: %v\n", semver.IsValid(result))
		fmt.Println()
	}
}
