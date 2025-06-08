package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"clikd/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// Update updates the model based on messages
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Verbesserte Handhabung von Texteingaben:
		// Wenn ein aktives Texteingabefeld vorhanden ist, leite den Tastendruck zuerst dorthin weiter
		if m.ActiveInput != "" {
			// Spezielle Handhabung der Enter-Taste
			if msg.String() == "enter" {
				// Enter-Taste verarbeiten basierend auf dem aktuellen Schritt
				switch m.CurrentStep {
				case StepAdvancedAIOptions:
					// Handle the different inputs for advanced options
					switch m.ActiveInput {
					case "Max Input Tokens":
						inputValue := m.TextInput.Value()
						if inputValue == "" {
							inputValue = "4096" // Default
						}
						m.Manager.SetConfigValue("ai.tokens_max_input", inputValue)

						// Move to max output tokens
						m.ActiveInput = "Max Output Tokens"
						m.TextInput.Placeholder = "500"
						m.TextInput.SetValue("")
						return m, nil

					case "Max Output Tokens":
						inputValue := m.TextInput.Value()
						if inputValue == "" {
							inputValue = "500" // Default
						}
						m.Manager.SetConfigValue("ai.tokens_max_output", inputValue)

						// Move to custom API URL
						m.ActiveInput = "Custom API URL"
						m.TextInput.Placeholder = "Leave empty to use official API"
						m.TextInput.SetValue("")
						return m, nil

					case "Custom API URL":
						inputValue := m.TextInput.Value()
						if inputValue != "" {
							m.Manager.SetConfigValue("ai.api_url", inputValue)
						}

						// Move to custom headers
						m.ActiveInput = "Custom API Headers"
						m.TextInput.Placeholder = "JSON format, leave empty for standard auth"
						m.TextInput.SetValue("")
						return m, nil

					case "Custom API Headers":
						inputValue := m.TextInput.Value()
						if inputValue != "" {
							m.Manager.SetConfigValue("ai.api_custom_headers", inputValue)
						}

						// Done with advanced options, return AIOptionsMsg to transition to API key config
						m.ActiveInput = ""
						return m, func() tea.Msg {
							return AIOptionsMsg{}
						}
					}

				case StepAPIKeyConfig:
					if m.Global && m.ActiveInput != "" {
						// Handle global API key input
						apiKey := m.TextInput.Value()
						if apiKey != "" {
							m.Manager.SetConfigValue("ai.api_key", apiKey)
							m.ApiKeyStatus = "done"
						}

						// Move to changelog configuration
						m.ActiveInput = ""
						m.CurrentStep = StepChangelogConfig
						m.SelectOptions = []SelectOption{
							{Title: "Yes", Description: "Configure changelog features", Value: true},
							{Title: "No", Description: "Skip changelog configuration", Value: false},
						}
						m.SelectCursor = 0
						return m, nil
					} else if !m.Global && m.ActiveInput == "API Key" {
						// Handle API key input for .env file
						apiKey := m.TextInput.Value()
						if apiKey != "" {
							// Create/Update .env file
							envContent := fmt.Sprintf("# API key for clikd\nCLIKD_API_KEY=%s\n", apiKey)
							err := os.WriteFile(".env", []byte(envContent), 0600)
							if err != nil {
								m.Message = fmt.Sprintf("Error creating/updating .env file: %s", err.Error())
								m.MessageType = "error"
							} else {
								m.Message = ".env file with API key has been created/updated."
								m.MessageType = "success"
							}
							m.ApiKeyStatus = "check"
						}

						// Move to changelog configuration
						m.ActiveInput = ""
						m.CurrentStep = StepChangelogConfig
						m.SelectOptions = []SelectOption{
							{Title: "Yes", Description: "Configure changelog features", Value: true},
							{Title: "No", Description: "Skip changelog configuration", Value: false},
						}
						m.SelectCursor = 0
						return m, nil
					}
				case "changelog_jira":
					if m.ActiveInput == "JIRA Prefix" {
						prefix := m.TextInput.Value()
						if prefix == "" {
							prefix = "PROJ" // Default
						}
						m.Manager.SetConfigValue("changelog.jira_prefix", prefix)

						// Move to changelog sort order
						m.ActiveInput = ""
						m.CurrentStep = "changelog_sort"
						m.SelectOptions = []SelectOption{
							{Title: "Newest first", Description: "Show newest entries at the top", Value: "desc"},
							{Title: "Oldest first", Description: "Show oldest entries at the top", Value: "asc"},
						}
						m.SelectCursor = 0
						return m, nil
					}
				case "changelog_advanced":
					if m.ActiveInput != "" {
						inputValue := m.TextInput.Value()
						switch m.ActiveInput {
						case "Custom Template":
							if inputValue != "" {
								m.Manager.SetConfigValue("changelog.template", inputValue)
							}
						case "Custom Sections":
							if inputValue != "" {
								m.Manager.SetConfigValue("changelog.sections", inputValue)
							}
						}

						// Move to case sensitivity
						m.ActiveInput = ""
						m.CurrentStep = "changelog_case"
						m.SelectOptions = []SelectOption{
							{Title: "Yes", Description: "Ignore case when generating changelog", Value: true},
							{Title: "No", Description: "Consider case when generating changelog", Value: false},
						}
						m.SelectCursor = 0
						return m, nil
					}
				}
			}

			// Aktualisiere das Texteingabefeld mit der eingegebenen Taste
			var cmd tea.Cmd
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}

		// Standard-Tastenbehandlung für nicht-Texteingabe-Modi
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.CurrentStep == StepConfigType || m.CurrentStep == "confirm_overwrite" ||
				m.CurrentStep == StepGeneralConfig || m.CurrentStep == "color_config" ||
				m.CurrentStep == StepAIConfig || m.CurrentStep == StepProviderSelection ||
				m.CurrentStep == StepModelSelection || m.CurrentStep == StepAdvancedAIOptions ||
				m.CurrentStep == StepAPIKeyConfig || m.CurrentStep == StepChangelogConfig ||
				m.CurrentStep == "changelog_style" || m.CurrentStep == "changelog_jira" ||
				m.CurrentStep == "changelog_sort" || m.CurrentStep == "changelog_advanced" ||
				m.CurrentStep == "changelog_case" {
				if m.SelectCursor > 0 {
					m.SelectCursor--
				}
			}

		case "down", "j":
			if m.CurrentStep == StepConfigType || m.CurrentStep == "confirm_overwrite" ||
				m.CurrentStep == StepGeneralConfig || m.CurrentStep == "color_config" ||
				m.CurrentStep == StepAIConfig || m.CurrentStep == StepProviderSelection ||
				m.CurrentStep == StepModelSelection || m.CurrentStep == StepAdvancedAIOptions ||
				m.CurrentStep == StepAPIKeyConfig || m.CurrentStep == StepChangelogConfig ||
				m.CurrentStep == "changelog_style" || m.CurrentStep == "changelog_jira" ||
				m.CurrentStep == "changelog_sort" || m.CurrentStep == "changelog_advanced" ||
				m.CurrentStep == "changelog_case" {
				if len(m.SelectOptions) > 0 && m.SelectCursor < len(m.SelectOptions)-1 {
					m.SelectCursor++
				}
			}

		case "enter":
			switch m.CurrentStep {
			case StepConfigType:
				if len(m.SelectOptions) > 0 {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)
					if !selectedValue {
						// User selected "No" -> global configuration
						m.Global = true
					}
					return m, determineConfigPath(m)
				}

			case "confirm_overwrite":
				if len(m.SelectOptions) > 0 {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)
					if selectedValue {
						// User selected "Yes" -> overwrite
						m.Force = true
						return m, initConfigManager(m.ConfigPath, false)
					} else {
						// User selected "No" -> don't overwrite
						m.Message = "Aborted, existing configuration will not be overwritten."
						m.MessageType = "info"
						m.Done = true
						return m, tea.Quit
					}
				}

			case StepGeneralConfig:
				// Log level selection
				if len(m.SelectOptions) > 0 {
					selectedOption := m.SelectOptions[m.SelectCursor]
					logLevel := selectedOption.Value.(string)

					// Set the log level in the configuration
					m.Manager.SetConfigValue("general.log_level", logLevel)

					// Move to color selection
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Enable colored terminal output", Value: true},
						{Title: "No", Description: "Disable colored terminal output", Value: false},
					}
					m.SelectCursor = 0
					m.CurrentStep = "color_config"
					return m, nil
				}

			case "color_config":
				if len(m.SelectOptions) > 0 {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)
					m.Manager.SetConfigValue("general.color", fmt.Sprintf("%t", selectedValue))

					// Move to AI configuration
					m.CurrentStep = StepAIConfig
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Enable AI features", Value: true},
						{Title: "No", Description: "Disable AI features", Value: false},
					}
					m.SelectCursor = 0
					return m, nil
				}

			case StepAIConfig:
				if len(m.SelectOptions) > 0 {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)
					m.AIEnabled = selectedValue
					m.Manager.SetConfigValue("ai.enable", fmt.Sprintf("%t", selectedValue))

					if selectedValue {
						// If AI is enabled, continue to provider selection
						m.CurrentStep = StepProviderSelection

						// Get supported providers
						providerOptions := config.SupportedProviders

						// Filter for gollm-supported providers
						supportedGollmProviders := []string{"mistral", "openai", "anthropic", "groq"}
						filteredProviderOptions := []string{}

						for _, provider := range providerOptions {
							for _, supported := range supportedGollmProviders {
								if provider == supported {
									filteredProviderOptions = append(filteredProviderOptions, provider)
									break
								}
							}
						}

						// Create select options for providers
						providerItems := make([]SelectOption, len(filteredProviderOptions))
						for i, provider := range filteredProviderOptions {
							defaultModel, _ := config.GetDefaultModelForProvider(provider)
							description := fmt.Sprintf("Default model: %s", defaultModel)
							if provider == "mistral" {
								description = fmt.Sprintf("RECOMMENDED - Default model: %s", defaultModel)
							}
							providerItems[i] = SelectOption{
								Title:       provider,
								Description: description,
								Value:       provider,
							}
						}

						m.SelectOptions = providerItems
						m.SelectCursor = 0
						return m, nil
					} else {
						// If AI is disabled, skip to changelog configuration
						m.CurrentStep = StepChangelogConfig
						m.SelectOptions = []SelectOption{
							{Title: "Yes", Description: "Configure changelog features", Value: true},
							{Title: "No", Description: "Skip changelog configuration", Value: false},
						}
						m.SelectCursor = 0
						return m, nil
					}
				}

			case StepProviderSelection:
				if len(m.SelectOptions) > 0 {
					// Save selected provider
					m.AIProvider = m.SelectOptions[m.SelectCursor].Value.(string)
					m.Manager.SetConfigValue("ai.provider", m.AIProvider)

					// Move to model selection
					m.CurrentStep = StepModelSelection

					// Get models for the selected provider
					supportedModels, _ := config.GetSupportedModelsForProvider(m.AIProvider)

					// Create model items
					modelItems := make([]SelectOption, len(supportedModels))

					// Define recommended models for each provider
					recommendedModels := map[string]string{
						"mistral":   "mistral-medium",
						"openai":    "gpt-4o",
						"anthropic": "claude-3-sonnet",
						"groq":      "llama3-70b-8192",
					}

					defaultModel, _ := config.GetDefaultModelForProvider(m.AIProvider)
					recommendedModel := defaultModel

					if recModel, exists := recommendedModels[m.AIProvider]; exists {
						recommendedModel = recModel
					}

					for i, model := range supportedModels {
						description := fmt.Sprintf("Model for %s", m.AIProvider)

						if model == defaultModel && model == recommendedModel {
							description = "DEFAULT & RECOMMENDED - Best balance of performance and cost"
						} else if model == defaultModel {
							description = "DEFAULT - Standard model for this provider"
						} else if model == recommendedModel {
							description = "RECOMMENDED - Best balance of performance and cost"
						}

						// Provider-specific descriptions
						if m.AIProvider == "mistral" {
							if model == "mistral-tiny" {
								description = "Fastest, most cost-effective option, but less capable"
							} else if model == "mistral-small" {
								description = "Good balance of speed and capability"
							} else if model == "mistral-medium" {
								description = "RECOMMENDED - Best overall value"
							} else if model == "mistral-large" {
								description = "Most capable Mistral model, but more expensive"
							}
						} else if m.AIProvider == "openai" {
							if model == "gpt-3.5-turbo" {
								description = "Faster and cheaper, but less capable"
							} else if model == "gpt-4o" {
								description = "RECOMMENDED - Latest model with best performance"
							} else if model == "gpt-4" {
								description = "Older model, still powerful but slower than gpt-4o"
							}
						} else if m.AIProvider == "anthropic" {
							if model == "claude-3-haiku" {
								description = "Faster and cheaper, good for simple tasks"
							} else if model == "claude-3-sonnet" {
								description = "RECOMMENDED - Good balance of capability and cost"
							} else if model == "claude-3-opus" {
								description = "Most capable Claude model, but more expensive"
							}
						}

						modelItems[i] = SelectOption{
							Title:       model,
							Description: description,
							Value:       model,
						}
					}

					m.SelectOptions = modelItems
					m.SelectCursor = 0
					return m, nil
				}

			case StepModelSelection:
				if len(m.SelectOptions) > 0 {
					// Save selected model
					m.AIModel = m.SelectOptions[m.SelectCursor].Value.(string)
					m.Manager.SetConfigValue("ai.model", m.AIModel)

					// Ask if user wants to configure advanced options
					m.CurrentStep = StepAdvancedAIOptions
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Configure advanced AI options", Value: true},
						{Title: "No", Description: "Use default settings", Value: false},
					}
					m.SelectCursor = 0
					return m, nil
				}

			case StepAdvancedAIOptions:
				if len(m.SelectOptions) > 0 && m.ActiveInput == "" {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)
					m.AICustomSettings = selectedValue

					if selectedValue {
						// If advanced options selected, start with max input tokens
						m.ActiveInput = "Max Input Tokens"
						m.TextInput.Placeholder = "4096"
						m.TextInput.Focus()
						m.TextInput.SetValue("")
						return m, nil
					} else {
						// If no advanced options, use defaults and continue to API key configuration
						m.Manager.SetConfigValue("ai.tokens_max_input", "4096")
						m.Manager.SetConfigValue("ai.tokens_max_output", "500")

						// Return AIOptionsMsg to transition to API key config
						return m, func() tea.Msg {
							return AIOptionsMsg{}
						}
					}
				} else if m.ActiveInput != "" {
					// Handle the different inputs for advanced options
					switch m.ActiveInput {
					case "Max Input Tokens":
						inputValue := m.TextInput.Value()
						if inputValue == "" {
							inputValue = "4096" // Default
						}
						m.Manager.SetConfigValue("ai.tokens_max_input", inputValue)

						// Move to max output tokens
						m.ActiveInput = "Max Output Tokens"
						m.TextInput.Placeholder = "500"
						m.TextInput.SetValue("")
						return m, nil

					case "Max Output Tokens":
						inputValue := m.TextInput.Value()
						if inputValue == "" {
							inputValue = "500" // Default
						}
						m.Manager.SetConfigValue("ai.tokens_max_output", inputValue)

						// Move to custom API URL
						m.ActiveInput = "Custom API URL"
						m.TextInput.Placeholder = "Leave empty to use official API"
						m.TextInput.SetValue("")
						return m, nil

					case "Custom API URL":
						inputValue := m.TextInput.Value()
						if inputValue != "" {
							m.Manager.SetConfigValue("ai.api_url", inputValue)
						}

						// Move to custom headers
						m.ActiveInput = "Custom API Headers"
						m.TextInput.Placeholder = "JSON format, leave empty for standard auth"
						m.TextInput.SetValue("")
						return m, nil

					case "Custom API Headers":
						inputValue := m.TextInput.Value()
						if inputValue != "" {
							m.Manager.SetConfigValue("ai.api_custom_headers", inputValue)
						}

						// Done with advanced options, return AIOptionsMsg to transition to API key config
						m.ActiveInput = ""
						return m, func() tea.Msg {
							return AIOptionsMsg{}
						}
					}
				}

			case StepAPIKeyConfig:
				if m.Global && m.ActiveInput != "" {
					// Handle global API key input
					apiKey := m.TextInput.Value()
					if apiKey != "" {
						m.Manager.SetConfigValue("ai.api_key", apiKey)
						m.ApiKeyStatus = "done"
					}

					// Move to changelog configuration
					m.ActiveInput = ""
					m.CurrentStep = StepChangelogConfig
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Configure changelog features", Value: true},
						{Title: "No", Description: "Skip changelog configuration", Value: false},
					}
					m.SelectCursor = 0
					return m, nil
				} else if !m.Global && len(m.SelectOptions) > 0 {
					// Handle local API key configuration (.env file)
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)

					if selectedValue {
						// If user wants to create/update .env file
						m.ActiveInput = "API Key"
						m.TextInput.Placeholder = "Enter your API key"
						m.TextInput.Focus()
						m.TextInput.SetValue("")
						return m, nil
					} else {
						// Skip API key configuration
						m.CurrentStep = StepChangelogConfig
						m.SelectOptions = []SelectOption{
							{Title: "Yes", Description: "Configure changelog features", Value: true},
							{Title: "No", Description: "Skip changelog configuration", Value: false},
						}
						m.SelectCursor = 0
						return m, nil
					}
				} else if !m.Global && m.ActiveInput == "API Key" {
					// Handle API key input for .env file
					apiKey := m.TextInput.Value()
					if apiKey != "" {
						// Create/Update .env file
						envContent := fmt.Sprintf("# API key for clikd\nCLIKD_API_KEY=%s\n", apiKey)
						err := os.WriteFile(".env", []byte(envContent), 0600)
						if err != nil {
							m.Message = fmt.Sprintf("Error creating/updating .env file: %s", err.Error())
							m.MessageType = "error"
						} else {
							m.Message = ".env file with API key has been created/updated."
							m.MessageType = "success"
							m.ApiKeyStatus = "done"
						}
					}

					// Move to changelog configuration
					m.ActiveInput = ""
					m.CurrentStep = StepChangelogConfig
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Configure changelog features", Value: true},
						{Title: "No", Description: "Skip changelog configuration", Value: false},
					}
					m.SelectCursor = 0
					return m, nil
				}

			case StepChangelogConfig:
				if len(m.SelectOptions) > 0 {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)

					if selectedValue {
						// Configure changelog features
						// First select style
						m.CurrentStep = "changelog_style"
						m.SelectOptions = []SelectOption{
							{Title: "github", Description: "RECOMMENDED - GitHub-style with Markdown (most widely used)", Value: "github"},
							{Title: "gitlab", Description: "GitLab-style with Markdown", Value: "gitlab"},
							{Title: "bitbucket", Description: "Bitbucket-style with Markdown", Value: "bitbucket"},
						}
						m.SelectCursor = 0
						return m, nil
					} else {
						// Skip changelog configuration
						m.CurrentStep = StepProjectStructure
						return m, createProjectStructure(m)
					}
				}

			case "changelog_style":
				if len(m.SelectOptions) > 0 {
					// Save selected style
					style := m.SelectOptions[m.SelectCursor].Value.(string)
					m.Manager.SetConfigValue("changelog.style", style)

					// If in a git repository, set the repository URL
					if m.IsInGitRepo && m.RepoURL != "" {
						m.Manager.SetConfigValue("changelog.repository_url", m.RepoURL)
					}

					// Ask about JIRA integration
					m.CurrentStep = "changelog_jira"
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Enable JIRA integration for changelog", Value: true},
						{Title: "No", Description: "Skip JIRA integration", Value: false},
					}
					m.SelectCursor = 0
					return m, nil
				}

			case "changelog_jira":
				if len(m.SelectOptions) > 0 {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)
					if selectedValue {
						// User wants to enable JIRA integration
						m.Manager.SetConfigValue("changelog.jira", "true")
						// Prompt for JIRA prefix
						m.ActiveInput = "JIRA Prefix"
						m.TextInput.Placeholder = "PROJ"
						m.TextInput.Focus()
						m.TextInput.SetValue("")
						m.SelectOptions = []SelectOption{} // Clear options
						return m, nil
					} else {
						// User doesn't want JIRA integration
						m.Manager.SetConfigValue("changelog.jira", "false")

						// Skip to sort order
						m.CurrentStep = "changelog_sort"
						m.SelectOptions = []SelectOption{
							{Title: "Newest first", Description: "Show newest entries at the top", Value: "desc"},
							{Title: "Oldest first", Description: "Show oldest entries at the top", Value: "asc"},
						}
						m.SelectCursor = 0
						return m, nil
					}
				} else if m.ActiveInput == "JIRA Prefix" {
					// Handle JIRA prefix input
					prefix := m.TextInput.Value()
					if prefix == "" {
						prefix = "PROJ" // Default
					}
					m.Manager.SetConfigValue("changelog.jira_prefix", prefix)

					// Move to changelog sort order
					m.ActiveInput = ""
					m.CurrentStep = "changelog_sort"
					m.SelectOptions = []SelectOption{
						{Title: "Newest first", Description: "Show newest entries at the top", Value: "desc"},
						{Title: "Oldest first", Description: "Show oldest entries at the top", Value: "asc"},
					}
					m.SelectCursor = 0
					return m, nil
				}

			case "changelog_sort":
				if len(m.SelectOptions) > 0 {
					// Save sort order
					sortOrder := m.SelectOptions[m.SelectCursor].Value.(string)
					m.Manager.SetConfigValue("changelog.sort", sortOrder)

					// Ask about advanced changelog options
					m.CurrentStep = "changelog_advanced"
					m.SelectOptions = []SelectOption{
						{Title: "Yes", Description: "Configure advanced changelog options", Value: true},
						{Title: "No", Description: "Use default settings", Value: false},
					}
					m.SelectCursor = 0
					return m, nil
				}

			case "changelog_advanced":
				if len(m.SelectOptions) > 0 && m.ActiveInput == "" {
					selectedValue := m.SelectOptions[m.SelectCursor].Value.(bool)

					if selectedValue {
						// Configure advanced options starting with tag filter
						m.ActiveInput = "Tag Filter Pattern"
						m.TextInput.Placeholder = "v*"
						m.TextInput.Focus()
						m.TextInput.SetValue("")
						return m, nil
					} else {
						// Use defaults and continue to project structure
						m.Manager.SetConfigValue("changelog.tag_filter_pattern", "v*")
						m.Manager.SetConfigValue("changelog.path", "CHANGELOG.md")
						m.Manager.SetConfigValue("changelog.no_case", "false")

						// Configure standard commit types and patterns
						setupChangelogDefaultOptions(m.Manager)

						m.CurrentStep = StepProjectStructure
						return m, createProjectStructure(m)
					}
				} else if m.ActiveInput != "" {
					// Handle different advanced inputs
					switch m.ActiveInput {
					case "Tag Filter Pattern":
						tagPattern := m.TextInput.Value()
						if tagPattern == "" {
							tagPattern = "v*" // Default
						}
						m.Manager.SetConfigValue("changelog.tag_filter_pattern", tagPattern)

						// Move to changelog path
						m.ActiveInput = "Changelog Path"
						m.TextInput.Placeholder = "CHANGELOG.md"
						m.TextInput.SetValue("")
						return m, nil

					case "Changelog Path":
						changelogPath := m.TextInput.Value()
						if changelogPath == "" {
							changelogPath = "CHANGELOG.md" // Default
						}
						m.Manager.SetConfigValue("changelog.path", changelogPath)

						// Ask about case sensitivity
						m.ActiveInput = ""
						m.CurrentStep = "changelog_case"
						m.SelectOptions = []SelectOption{
							{Title: "Yes", Description: "Make changelog generation case insensitive", Value: true},
							{Title: "No", Description: "Keep case sensitivity (default)", Value: false},
						}
						m.SelectCursor = 0
						return m, nil
					}
				}

			case "changelog_case":
				if len(m.SelectOptions) > 0 {
					noCaseOption := m.SelectOptions[m.SelectCursor].Value.(bool)
					m.Manager.SetConfigValue("changelog.no_case", fmt.Sprintf("%t", noCaseOption))

					// Configure standard commit types and patterns
					setupChangelogDefaultOptions(m.Manager)

					// Continue to project structure
					m.CurrentStep = StepProjectStructure
					return m, createProjectStructure(m)
				}
			}
		}

	case GitRepoMsg:
		m.IsInGitRepo = msg.IsInGitRepo
		m.RepoURL = msg.RepoURL
		m.CurrentStep = StepConfigType

		// In non-interactive mode or if global is set, continue directly
		if m.Yes || m.Global {
			return m, determineConfigPath(m)
		}

		// Set options for configuration type selection
		if m.IsInGitRepo {
			m.SelectOptions = []SelectOption{
				{Title: "Yes", Description: "Local configuration for this repository", Value: true},
				{Title: "No", Description: "Use global configuration", Value: false},
			}
			m.SelectCursor = 0
		} else {
			// If no Git repository, go directly to global configuration
			m.Global = true
			return m, determineConfigPath(m)
		}

		return m, nil

	case ConfigPathMsg:
		m.ConfigPath = msg.ConfigPath
		m.ConfigDir = msg.ConfigDir
		m.CurrentStep = StepCreateDirs
		return m, checkConfigExists(m.ConfigPath)

	case ConfigExistsMsg:
		m.ConfigExists = msg.Exists

		// If configuration exists and no force flag
		if m.ConfigExists && !m.Force && !m.Yes {
			m.CurrentStep = "confirm_overwrite"

			// Set options for overwrite confirmation
			m.SelectOptions = []SelectOption{
				{Title: "Yes", Description: "Overwrite existing configuration", Value: true},
				{Title: "No", Description: "Cancel and keep existing configuration", Value: false},
			}
			m.SelectCursor = 0

			return m, nil
		}

		// Force or non-interactive, continue directly
		return m, initConfigManager(m.ConfigPath, m.ConfigExists && !m.Force)

	case ConfigManagerMsg:
		m.Manager = msg.Manager
		m.CurrentStep = StepGeneralConfig

		// In non-interactive mode use default settings
		if m.Yes {
			m.Manager.SetConfigValue("general.log_level", "info")
			m.Manager.SetConfigValue("general.color", "true")
			m.CurrentStep = StepAIConfig
			return m, nil
		}

		// Set log level options
		m.SelectOptions = []SelectOption{
			{Title: "info", Description: "Standard log level (recommended)", Value: "info"},
			{Title: "debug", Description: "Verbose logging for troubleshooting", Value: "debug"},
			{Title: "warn", Description: "Only warnings and errors", Value: "warn"},
			{Title: "error", Description: "Only errors", Value: "error"},
		}
		m.SelectCursor = 0

		return m, nil

	case ProjectStructureCompleteMsg:
		// Project structure completed successfully
		m.CurrentStep = StepSummary
		return m, nil

	case ProjectStructureErrorMsg:
		// Project structure creation failed
		m.Error = msg.Error
		return m, tea.Quit

	case AIOptionsMsg:
		// Update AI provider and model in the configuration
		m.Manager.SetConfigValue("ai.provider", m.AIProvider)
		m.Manager.SetConfigValue("ai.model", m.AIModel)

		// Move to API key configuration
		m.ActiveInput = ""
		m.CurrentStep = StepAPIKeyConfig

		// Different behavior for global vs local configuration
		if m.Global {
			// For global config, set up text input for API key
			m.ActiveInput = "API Key"
			m.TextInput.Placeholder = "Enter your API key or leave empty"
			m.TextInput.Focus()
		} else {
			// Check if .env file exists
			return m, checkEnvFileExists()
		}
		return m, nil

	case EnvFileExistsMsg:
		envExists := msg.Exists

		// If .env file exists, show message and set options differently
		if envExists {
			m.Message = "Found existing .env file."
			m.MessageType = "info"
			m.SelectOptions = []SelectOption{
				{Title: "Yes", Description: "Update API key in existing .env file", Value: true},
				{Title: "No", Description: "Keep existing .env file and skip API key configuration", Value: false},
			}
		} else {
			// No .env file, ask about creating one
			m.SelectOptions = []SelectOption{
				{Title: "Yes", Description: "Create a .env file with API key", Value: true},
				{Title: "No", Description: "Skip API key configuration", Value: false},
			}
		}
		m.SelectCursor = 0
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// Command functions

// determineConfigPath determines the configuration path
func determineConfigPath(m InitModel) tea.Cmd {
	return func() tea.Msg {
		var configDir string

		if m.Global {
			// Determine home directory
			home, err := os.UserHomeDir()
			if err != nil {
				return tea.Quit
			}

			// Global configuration directory
			configDir = filepath.Join(home, ".clikd")
		} else {
			// Local configuration directory
			configDir = "clikd"
		}

		configPath := filepath.Join(configDir, "config.toml")

		return ConfigPathMsg{
			ConfigPath: configPath,
			ConfigDir:  configDir,
		}
	}
}

// checkConfigExists checks if the configuration file exists
func checkConfigExists(configPath string) tea.Cmd {
	return func() tea.Msg {
		_, err := os.Stat(configPath)
		exists := err == nil

		return ConfigExistsMsg{
			Exists: exists,
		}
	}
}

// initConfigManager initializes the configuration manager
func initConfigManager(configPath string, loadExisting bool) tea.Cmd {
	return func() tea.Msg {
		manager := config.NewManager()

		if loadExisting {
			if err := manager.InitConfig(configPath); err != nil {
				return tea.Quit
			}
		} else {
			if err := manager.InitConfig(""); err != nil {
				return tea.Quit
			}
		}

		return ConfigManagerMsg{
			Manager: manager,
		}
	}
}

// checkGitRepo checks if we are in a Git repository
func checkGitRepo() tea.Msg {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()

	isInGitRepo := err == nil
	repoURL := ""

	if isInGitRepo {
		// Determine repository URL
		remoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
		if remoteOutput, err := remoteCmd.Output(); err == nil {
			repoURL = strings.TrimSpace(string(remoteOutput))
		}
	}

	return GitRepoMsg{
		IsInGitRepo: isInGitRepo,
		RepoURL:     repoURL,
	}
}

// checkEnvFileExists returns a tea.Cmd that checks if a .env file exists
func checkEnvFileExists() tea.Cmd {
	return func() tea.Msg {
		_, err := os.Stat(".env")
		exists := err == nil
		return EnvFileExistsMsg{Exists: exists}
	}
}
