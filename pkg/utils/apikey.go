package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ProviderKeyInfo holds information about an API key for a specific provider
type ProviderKeyInfo struct {
	Name            string // Name of the provider (e.g., "OpenAI", "Mistral")
	ConfigKey       string // Key in config file (e.g., "ai.models.openai.api_key")
	EnvVarName      string // Environment variable name (e.g., "CLIKD_OPENAI_API_KEY")
	EnvVarNameShort string // Short env var for error messages (e.g., "OPENAI_API_KEY")
	Required        bool   // Whether this key is required for the current operation
}

// GetAPIKey retrieves the appropriate API key based on the following logic:
// 1. If a local configuration exists, look for the key in .env file
// 2. If no key in .env and a global key exists, ask the user if they want to use it
// 3. If user agrees or no local config exists, use the global key
// 4. If no key is found anywhere, provide clear instructions for the user
// Note: Environment variables from the shell are NOT considered, only .env file and global config
func GetAPIKey(provider ProviderKeyInfo, localConfigExists bool) (string, error) {
	// Check if we have a local config
	if localConfigExists {
		// Try to get key from .env file
		envKey := getEnvKeyFromDotEnv(provider.EnvVarName)

		// If we found a key in .env, return it
		if envKey != "" {
			return envKey, nil
		}

		// No key in .env, check if we have a global key
		globalKey := viper.GetString(provider.ConfigKey)
		if globalKey != "" {
			// Ask user if they want to use the global key
			fmt.Printf("No %s found in .env file. A global API key is available.\n", provider.EnvVarNameShort)
			fmt.Print("Do you want to use the global API key? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("error reading user input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response == "y" || response == "yes" {
				return globalKey, nil
			}

			// User declined to use global key
			if provider.Required {
				errorMsg := fmt.Sprintf("API-Schlüssel für %s wird benötigt, wenn KI-Funktionen aktiviert sind.", provider.Name)
				errorMsg += fmt.Sprintf("\nBitte fügen Sie den Schlüssel zur .env-Datei hinzu: %s=ihr_api_schlüssel", provider.EnvVarName)
				errorMsg += fmt.Sprintf("\nOder deaktivieren Sie die KI-Funktionen durch Weglassen des --ai Flags.")
				return "", fmt.Errorf("%s", errorMsg)
			}
			return "", nil
		}

		// No key in .env and no global key
		if provider.Required {
			// Klare Anweisungen für das Hinzufügen eines API-Schlüssels
			errMsg := fmt.Sprintf("API-Schlüssel für %s nicht gefunden. ", provider.Name)

			if os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED") == "true" {
				errMsg += "KI-Funktionen wurden explizit aktiviert (--ai), aber kein API-Schlüssel gefunden.\n\n"
			}

			errMsg += "Sie können den Schlüssel auf folgende Weise hinzufügen:\n\n"
			errMsg += fmt.Sprintf("1. Erstellen Sie eine .env-Datei im Projektverzeichnis und fügen Sie hinzu:\n   %s=ihr_api_schlüssel\n\n", provider.EnvVarName)
			errMsg += fmt.Sprintf("2. Oder fügen Sie den Schlüssel zu Ihrer globalen Konfiguration hinzu:\n   clikd config set %s ihr_api_schlüssel\n\n", provider.ConfigKey)
			errMsg += fmt.Sprintf("Um einen API-Schlüssel zu erhalten, besuchen Sie die Website des Anbieters: %s", getProviderURL(provider.Name))

			if os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED") == "true" {
				errMsg += fmt.Sprintf("\n\nAlternativ können Sie die KI-Funktionen deaktivieren mit:\n  clikd config set ai.enable false")
			}

			return "", fmt.Errorf("%s", errMsg)
		}
		return "", nil
	}

	// No local config, try to get from global config
	globalKey := viper.GetString(provider.ConfigKey)
	if globalKey != "" {
		return globalKey, nil
	}

	// No global key either
	if provider.Required {
		// Klare Anweisungen für das Hinzufügen eines API-Schlüssels zur globalen Konfiguration
		errMsg := fmt.Sprintf("API-Schlüssel für %s nicht gefunden. ", provider.Name)

		if os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED") == "true" {
			errMsg += "KI-Funktionen wurden explizit aktiviert (--ai), aber kein API-Schlüssel gefunden.\n\n"
		}

		errMsg += fmt.Sprintf("Sie können den Schlüssel zu Ihrer globalen Konfiguration hinzufügen:\n\n")
		errMsg += fmt.Sprintf("  clikd config set %s ihr_api_schlüssel\n\n", provider.ConfigKey)
		errMsg += fmt.Sprintf("Um einen API-Schlüssel zu erhalten, besuchen Sie die Website des Anbieters: %s", getProviderURL(provider.Name))

		if os.Getenv("CLIKD_AI_EXPLICITLY_ENABLED") == "true" {
			errMsg += fmt.Sprintf("\n\nAlternativ können Sie die KI-Funktionen deaktivieren mit:\n  clikd config set ai.enable false")
		}

		return "", fmt.Errorf("%s", errMsg)
	}

	return "", nil
}

// getProviderURL returns the URL where users can obtain an API key for the given provider
func getProviderURL(providerName string) string {
	switch providerName {
	case "OpenAI":
		return "https://platform.openai.com/api-keys"
	case "Mistral":
		return "https://console.mistral.ai/api-keys/"
	case "Azure OpenAI":
		return "https://portal.azure.com/#create/Microsoft.CognitiveServicesOpenAI"
	default:
		return "die entsprechende Anbieter-Website"
	}
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

// Examples of provider configurations
var (
	OpenAIProvider = ProviderKeyInfo{
		Name:            "OpenAI",
		ConfigKey:       "ai.models.openai.api_key",
		EnvVarName:      "CLIKD_OPENAI_API_KEY",
		EnvVarNameShort: "OPENAI_API_KEY",
		Required:        false,
	}

	MistralProvider = ProviderKeyInfo{
		Name:            "Mistral",
		ConfigKey:       "ai.models.mistral.api_key",
		EnvVarName:      "CLIKD_MISTRAL_API_KEY",
		EnvVarNameShort: "MISTRAL_API_KEY",
		Required:        false,
	}

	// Add more providers as needed
)
