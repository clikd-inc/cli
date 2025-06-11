package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// GetAPIKey retrieves the CLIKD_API_KEY from the following sources in order:
// 1. .env file (if local config exists)
// 2. Environment variable CLIKD_API_KEY
// 3. Global configuration ai.api_key
func GetAPIKey() (string, error) {
	// 1. Check .env file first (if local config exists)
	if IsLocalConfigPresent() {
		if key := getEnvKeyFromDotEnv("CLIKD_API_KEY"); key != "" {
			return key, nil
		}
	}

	// 2. Check environment variable
	if key := os.Getenv("CLIKD_API_KEY"); key != "" {
		return key, nil
	}

	// 3. Check global configuration
	if key := viper.GetString("ai.api_key"); key != "" {
		return key, nil
	}

	// No API key found anywhere
	errMsg := "CLIKD_API_KEY not found. "

	if os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED") == "true" {
		errMsg += "AI features were explicitly enabled (--ai), but no API key found.\n\n"
	}

	errMsg += "You can add the API key in the following ways:\n\n"

	if IsLocalConfigPresent() {
		errMsg += "1. Create a .env file in the project directory and add:\n   CLIKD_API_KEY=your_api_key\n\n"
		errMsg += "2. Or add it to your global configuration:\n   clikd config set ai.api_key your_api_key\n\n"
	} else {
		errMsg += "Add it to your global configuration:\n   clikd config set ai.api_key your_api_key\n\n"
	}

	errMsg += "To configure your API key, you can:\n"
	errMsg += "1. Create a .env file in your project directory with:\n"
	errMsg += "   CLIKD_API_KEY=your_api_key\n"
	errMsg += "2. Or set it globally with:\n"
	errMsg += "   clikd config set ai.api_key your_api_key"

	if os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED") == "true" {
		errMsg += "\n\nNote: AI features were explicitly enabled via command line flag."
	}

	return "", fmt.Errorf("%s", errMsg)
}

// IsLocalConfigPresent checks if a local clikd configuration directory exists
func IsLocalConfigPresent() bool {
	// Check for ./clikd directory
	_, err := os.Stat("./clikd")
	return err == nil
}

// getEnvKeyFromDotEnv tries to read the specified environment variable from .env file
func getEnvKeyFromDotEnv(envVarName string) string {
	// Try to open .env file
	envFile, err := os.Open(".env")
	if err != nil {
		return "" // .env file doesn't exist or can't be opened
	}
	defer envFile.Close()

	// Read .env file line by line
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by = or export VAR=
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		// Handle export VAR=value syntax
		if strings.HasPrefix(key, "export ") {
			key = strings.TrimSpace(strings.TrimPrefix(key, "export"))
		}

		if key == envVarName {
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			return value
		}
	}

	return ""
}
