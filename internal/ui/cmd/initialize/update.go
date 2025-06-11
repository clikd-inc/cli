package initialize

import (
	"clikd/internal/config"
	"clikd/internal/ui/bubble"
	"clikd/internal/ui/styles"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model using our UI components
func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}

		// If we have an active component, let it handle the key press
		if m.confirmModel != nil {
			newConfirm, confirmCmd := m.confirmModel.Update(msg)
			confirmModel, ok := newConfirm.(bubble.ConfirmModel)
			if ok {
				m.confirmModel = &confirmModel
				return m, confirmCmd
			}
		}

		if m.selectModel != nil {
			newSelect, selectCmd := m.selectModel.Update(msg)
			selectModel, ok := newSelect.(bubble.SelectModel)
			if ok {
				m.selectModel = &selectModel
				return m, selectCmd
			}
		}

		if m.inputModel != nil {
			newInput, inputCmd := m.inputModel.Update(msg)
			inputModel, ok := newInput.(bubble.InputModel)
			if ok {
				m.inputModel = &inputModel
				return m, inputCmd
			}
		}

	case bubble.InputResultMsg:
		m.inputModel = nil
		switch m.CurrentStep {
		case StepChangelogURL:
			// Repository URL was entered
			m.ChangelogRepositoryURL = msg.Value
			// Continue to style selection
			m.CurrentStep = StepChangelogStyle
			styleItems := []bubble.SelectItem{
				{
					Title:       "github",
					Description: "RECOMMENDED - GitHub-style with Markdown (most widely used)",
					Value:       "github",
				},
				{
					Title:       "gitlab",
					Description: "GitLab-style with Markdown",
					Value:       "gitlab",
				},
				{
					Title:       "bitbucket",
					Description: "Bitbucket-style with Markdown",
					Value:       "bitbucket",
				},
				{
					Title:       "none",
					Description: "Simple format without special links",
					Value:       "none",
				},
			}
			selectModel := bubble.NewSelectModel("Select Changelog Style", styleItems)
			m.selectModel = &selectModel
			return m, nil

		case StepAITokensInput:
			// Max input tokens was entered
			m.AITokensMaxInput = msg.Value
			m.Manager.SetConfigValue("ai.tokens_max_input", msg.Value)

			// Continue to max output tokens
			m.CurrentStep = StepAITokensOutput
			inputModel := bubble.NewInputModel(
				"Max Output Tokens",
				"Maximum number of output tokens (response length)",
				"500",
			)
			m.inputModel = &inputModel
			return m, nil

		case StepAITokensOutput:
			// Max output tokens was entered
			m.AITokensMaxOutput = msg.Value
			m.Manager.SetConfigValue("ai.tokens_max_output", msg.Value)

			// Continue to custom API URL
			m.CurrentStep = StepAICustomURL
			inputModel := bubble.NewInputModel(
				"Custom API URL",
				"Custom API endpoint URL (leave empty to use official API)",
				"",
			)
			m.inputModel = &inputModel
			return m, nil

		case StepAICustomURL:
			// Custom API URL was entered
			m.AICustomURL = msg.Value
			m.Manager.SetConfigValue("ai.api_url", msg.Value)

			// Continue to custom headers
			m.CurrentStep = StepAICustomHeaders
			inputModel := bubble.NewInputModel(
				"Custom API Headers",
				"Custom HTTP headers in JSON format (leave empty for standard authentication)",
				"",
			)
			m.inputModel = &inputModel
			return m, nil

		case StepAICustomHeaders:
			// Custom headers was entered
			m.AICustomHeaders = msg.Value
			m.Manager.SetConfigValue("ai.api_custom_headers", msg.Value)

			// Continue to API key configuration
			m.CurrentStep = StepAPIKeyConfig
			confirmModel := bubble.NewConfirmModel(
				"API Key Configuration",
				fmt.Sprintf("Do you want to configure your %s API key now?", m.AIProvider),
			)
			m.confirmModel = &confirmModel
			return m, nil

			// StepChangelogConfigDir removed - we use fixed directory structure: clikd/changelog/
		}

	// Handle results from components
	case bubble.ConfirmResultMsg:
		m.confirmModel = nil
		switch m.CurrentStep {

		case StepConfirmOverwrite:
			if !msg.Result {
				m.Message = "Aborted, existing configuration will not be overwritten."
				m.MessageType = "info"
				m.Done = true
				return m, tea.Quit
			}
			m.Force = true
			return m, initConfigManager(m.ConfigPath, m.ConfigExists && !m.Force)

		case StepChangelogConfig:
			m.ChangelogEnabled = msg.Result
			if msg.Result {
				// Ask for repository URL first
				m.CurrentStep = StepChangelogURL
				defaultURL := m.RepoURL
				if defaultURL == "" {
					defaultURL = ""
				}
				inputModel := bubble.NewInputModel(
					"Repository URL",
					"What is the URL of your repository?",
					defaultURL,
				)
				m.inputModel = &inputModel
				return m, nil
			} else {
				// Skip changelog config, go to project structure
				m.CurrentStep = StepProjectStructure
				return m, createProjectStructure(m)
			}

		case StepChangelogMerges:
			m.ChangelogIncludeMerges = msg.Result
			// Ask about revert commits
			m.CurrentStep = StepChangelogReverts
			confirmModel := bubble.NewConfirmModel(
				"Revert Commits",
				"Do you include Revert Commit in CHANGELOG?",
			)
			m.confirmModel = &confirmModel
			return m, nil

		case StepChangelogReverts:
			m.ChangelogIncludeReverts = msg.Result
			// Go directly to project structure (fixed config dir: clikd/changelog/)
			m.CurrentStep = StepProjectStructure
			return m, createProjectStructure(m)

		case StepChangelogColor:
			m.ChangelogColorEnabled = msg.Result
			// Ask about merge commits
			m.CurrentStep = StepChangelogMerges
			confirmModel := bubble.NewConfirmModel(
				"Merge Commits",
				"Do you include Merge Commit in CHANGELOG?",
			)
			m.confirmModel = &confirmModel
			return m, nil

		case StepAdvancedAIOptions:
			m.AICustomSettings = msg.Result
			if msg.Result {
				// Start the multi-step advanced AI options configuration
				m.CurrentStep = StepAITokensInput
				inputModel := bubble.NewInputModel(
					"Max Input Tokens",
					"Maximum number of input tokens (context size)",
					"4096",
				)
				m.inputModel = &inputModel
				return m, nil
			} else {
				// Use default values for AI configuration
				m.Manager.SetConfigValue("ai.provider", "mistral")
				m.Manager.SetConfigValue("ai.model", "mistral-medium")
				m.Manager.SetConfigValue("ai.tokens_max_input", "4096")
				m.Manager.SetConfigValue("ai.tokens_max_output", "500")
				m.Manager.SetConfigValue("ai.api_url", "")
				m.Manager.SetConfigValue("ai.api_custom_headers", "")

				// Set the values in the model as well
				m.AIProvider = "mistral"
				m.AIModel = "mistral-medium"
				m.AITokensMaxInput = "4096"
				m.AITokensMaxOutput = "500"
				m.AICustomURL = ""
				m.AICustomHeaders = ""
			}

			// Go to API key configuration
			m.CurrentStep = StepAPIKeyConfig
			confirmModel := bubble.NewConfirmModel(
				"API Key Configuration",
				fmt.Sprintf("Do you want to configure your %s API key now?", m.AIProvider),
			)
			m.confirmModel = &confirmModel
			return m, nil

		case StepAPIKeyConfig:
			if msg.Result {
				// Configure API key
				apiKey := configureAPIKey(m.AIProvider, m.Global)
				if apiKey != "" {
					if m.Global {
						// For global config, save API key in config.toml
						m.Manager.SetConfigValue("ai.api_key", apiKey)
						m.ApiKeyStatus = "done"
					} else {
						// For local config, create/update .env file
						if err := createOrUpdateEnvFile(apiKey); err != nil {
							m.Error = err
							m.Message = fmt.Sprintf("Error creating .env file: %s", err)
							m.MessageType = "error"
							return m, tea.Quit
						}
						m.ApiKeyStatus = "done"
					}
				}
			} else {
				// User chose not to configure API key
				if m.Global {
					m.ApiKeyStatus = "pending"
				} else {
					m.ApiKeyStatus = "check"
				}
			}

			// Continue to changelog config (only for local configurations)
			if !m.Global {
				m.CurrentStep = StepChangelogConfig
				confirmModel := bubble.NewConfirmModel(
					"Changelog Configuration",
					"Do you want to configure changelog features?",
				)
				m.confirmModel = &confirmModel
				return m, nil
			} else {
				// For global config, skip changelog and go to project structure
				m.CurrentStep = StepProjectStructure
				return m, createProjectStructure(m)
			}
		}

	case bubble.SelectResultMsg:
		m.selectModel = nil
		switch m.CurrentStep {
		case StepConfigType:
			if value, ok := msg.Value.(string); ok {
				m.Global = (value == "global")
				m.CurrentStep = StepCreateDirs
				return m, determineConfigPath(m)
			}

		case StepGeneralConfig:
			if value, ok := msg.Value.(string); ok {
				m.Manager.SetConfigValue("general.log_level", value)
				// AI is now mandatory, go directly to provider selection
				m.CurrentStep = StepProviderSelection

				// Get available providers
				providers := []string{"mistral", "anthropic", "openai"}
				providerItems := make([]bubble.SelectItem, len(providers))

				for i, provider := range providers {
					defaultModel, _ := config.GetDefaultModelForProvider(provider)
					providerItems[i] = bubble.SelectItem{
						Title:       provider,
						Description: fmt.Sprintf("Default model: %s", defaultModel),
						Value:       provider,
					}
				}

				selectModel := bubble.NewSelectModel("Select AI Provider", providerItems)
				m.selectModel = &selectModel
				return m, nil
			}

		case StepProviderSelection:
			if value, ok := msg.Value.(string); ok {
				m.AIProvider = value
				m.Manager.SetConfigValue("ai.provider", value)

				// Initialize model selection directly
				m.CurrentStep = StepModelSelection

				// For simplicity, just hardcode some models for each provider
				var models []string
				switch value {
				case "mistral":
					models = []string{"mistral-tiny", "mistral-small", "mistral-medium", "mistral-large"}
				case "anthropic":
					models = []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}
				case "openai":
					models = []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
				default:
					// Use default model if provider not recognized
					defaultModel, _ := config.GetDefaultModelForProvider(value)
					m.AIModel = defaultModel
					m.Manager.SetConfigValue("ai.model", defaultModel)

					// Skip to changelog config
					m.CurrentStep = StepChangelogConfig
					return m, createProjectStructure(m)
				}

				modelItems := make([]bubble.SelectItem, len(models))
				for i, model := range models {
					modelItems[i] = bubble.SelectItem{
						Title:       model,
						Description: "", // Add description if available
						Value:       model,
					}
				}

				selectModel := bubble.NewSelectModel(fmt.Sprintf("Select Model for %s", value), modelItems)
				m.selectModel = &selectModel
				return m, nil
			}

		case StepModelSelection:
			if value, ok := msg.Value.(string); ok {
				m.AIModel = value
				m.Manager.SetConfigValue("ai.model", value)

				// Go to advanced AI options configuration
				m.CurrentStep = StepAdvancedAIOptions
				confirmModel := bubble.NewConfirmModel(
					"Advanced AI Options",
					"Do you want to configure advanced AI options (token limits, custom endpoints, etc.)?",
				)
				m.confirmModel = &confirmModel
				return m, nil
			}

		case StepChangelogStyle:
			if value, ok := msg.Value.(string); ok {
				m.ChangelogStyle = value

				// Ask about commit message format
				m.CurrentStep = StepChangelogFormat
				formatItems := []bubble.SelectItem{
					{
						Title:       "<type>(<scope>): <subject>",
						Description: "feat(core): Add new feature",
						Value:       "<type>(<scope>): <subject>",
					},
					{
						Title:       "<type>: <subject>",
						Description: "feat: Add new feature",
						Value:       "<type>: <subject>",
					},
					{
						Title:       "<<type> subject>",
						Description: "Add new feature",
						Value:       "<<type> subject>",
					},
					{
						Title:       "<subject>",
						Description: "Add new feature (Not detect `type` field)",
						Value:       "<subject>",
					},
					{
						Title:       ":<type>: <subject>",
						Description: ":sparkles: Add new feature (Commit message with emoji format)",
						Value:       ":<type>: <subject>",
					},
				}
				selectModel := bubble.NewSelectModel("Choose Commit Message Format", formatItems)
				m.selectModel = &selectModel
				return m, nil
			}

		case StepChangelogFormat:
			if value, ok := msg.Value.(string); ok {
				m.ChangelogFormat = value

				// Ask for template style
				m.CurrentStep = StepChangelogTemplate
				templateItems := []bubble.SelectItem{
					{
						Title:       "standard",
						Description: "Standard changelog template",
						Value:       "standard",
						Preview:     GetTemplatePreview("standard"),
					},
					{
						Title:       "keep-a-changelog",
						Description: "Keep a Changelog format",
						Value:       "keep-a-changelog",
						Preview:     GetTemplatePreview("keep-a-changelog"),
					},
					{
						Title:       "cool",
						Description: "Cool template with emojis",
						Value:       "cool",
						Preview:     GetTemplatePreview("cool"),
					},
				}
				selectModel := bubble.NewSelectModel("Select Template Style", templateItems)
				m.selectModel = &selectModel
				return m, nil
			}

		case StepChangelogTemplate:
			if value, ok := msg.Value.(string); ok {
				m.ChangelogTemplate = value

				// Ask about terminal color for changelog
				m.CurrentStep = StepChangelogColor
				confirmModel := bubble.NewConfirmModel(
					"Terminal Color",
					"Enable colored terminal output for changelog?",
				)
				m.confirmModel = &confirmModel
				return m, nil
			}
		}

	// Original message types
	case GitRepoMsg:
		m.IsInGitRepo = msg.IsInGitRepo
		m.RepoURL = msg.RepoURL
		m.CurrentStep = StepConfigType

		// In non-interactive mode or if global is set, continue directly
		if m.Yes || m.Global {
			m.CurrentStep = StepCreateDirs
			return m, determineConfigPath(m)
		}

		// Setup the select component for configuration type
		title := "Choose Configuration Type"
		if m.IsInGitRepo && m.RepoURL != "" {
			title = fmt.Sprintf("Git repository detected: %s\n\nChoose Configuration Type", styles.SuccessText(m.RepoURL))
		}

		configTypeItems := []bubble.SelectItem{
			{
				Title:       "Local",
				Description: "Project-specific configuration (recommended for teams)",
				Value:       "local",
			},
			{
				Title:       "Global",
				Description: "System-wide configuration (good for personal use)",
				Value:       "global",
			},
		}
		selectModel := bubble.NewSelectModel(title, configTypeItems)
		m.selectModel = &selectModel
		return m, nil

	case ConfigPathMsg:
		m.ConfigPath = msg.ConfigPath
		m.ConfigDir = msg.ConfigDir
		return m, checkConfigExists(m.ConfigPath)

	case ConfigExistsMsg:
		m.ConfigExists = msg.Exists

		// If configuration exists and no force flag
		if m.ConfigExists && !m.Force && !m.Yes {
			configType := "Local"
			configLocation := "clikd/"
			if m.Global {
				configType = "Global"
				configLocation = "~/.clikd/"
			}
			confirmModel := bubble.NewConfirmModel(
				"Existing Configuration",
				fmt.Sprintf("%s configuration already exists in %s. Do you want to overwrite it?", configType, configLocation),
			)
			m.confirmModel = &confirmModel
			m.CurrentStep = StepConfirmOverwrite
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

			// Configure AI with defaults in non-interactive mode
			m.AIProvider = "mistral"
			m.AIModel = "mistral-medium"
			m.Manager.SetConfigValue("ai.provider", "mistral")
			m.Manager.SetConfigValue("ai.model", "mistral-medium")

			// Set default values for AI configuration
			m.Manager.SetConfigValue("ai.provider", "mistral")
			m.Manager.SetConfigValue("ai.model", "mistral-medium")
			m.Manager.SetConfigValue("ai.tokens_max_input", "4096")
			m.Manager.SetConfigValue("ai.tokens_max_output", "500")
			m.Manager.SetConfigValue("ai.api_url", "")
			m.Manager.SetConfigValue("ai.api_custom_headers", "")

			// Set the values in the model as well
			m.AIProvider = "mistral"
			m.AIModel = "mistral-medium"
			m.AITokensMaxInput = "4096"
			m.AITokensMaxOutput = "500"
			m.AICustomURL = ""
			m.AICustomHeaders = ""

			// Configure changelog with defaults (only for local configuration)
			if !m.Global {
				m.ChangelogEnabled = true
				m.ChangelogRepositoryURL = m.RepoURL
				m.ChangelogStyle = "github"
				m.ChangelogFormat = "<type>(<scope>): <subject>"
				m.ChangelogTemplate = "standard"
				m.ChangelogColorEnabled = true
				m.ChangelogIncludeMerges = true
				m.ChangelogIncludeReverts = true
			}

			// Skip to project structure creation
			m.CurrentStep = StepProjectStructure
			return m, createProjectStructure(m)
		}

		// Setup log level selection
		logLevelItems := []bubble.SelectItem{
			{Title: "info", Description: "Standard log level (recommended)", Value: "info"},
			{Title: "debug", Description: "Verbose logging for troubleshooting", Value: "debug"},
			{Title: "warn", Description: "Only warnings and errors", Value: "warn"},
			{Title: "error", Description: "Only errors", Value: "error"},
		}

		selectModel := bubble.NewSelectModel("Select Log Level", logLevelItems)
		m.selectModel = &selectModel
		return m, nil

	case ProjectStructureCompleteMsg:
		// Project structure completed successfully
		m.CurrentStep = StepSummary
		m.Message = "Configuration completed successfully."
		m.MessageType = "success"
		m.Done = true
		return m, tea.Quit

	case ProjectStructureErrorMsg:
		// Project structure creation failed
		m.Error = msg.Error
		m.Message = fmt.Sprintf("Error creating project structure: %s", msg.Error)
		m.MessageType = "error"
		return m, tea.Quit

	case ForceStepChangeMsg:
		m.CurrentStep = msg.NewStep

		// Initialize appropriate component based on new step
		switch msg.NewStep {
		default:
			// No special initialization needed for other steps
		}

		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// Helper functions to set up components for different steps

// setupChangelogConfig prepares the changelog configuration UI
func setupChangelogConfig(m InitModel) tea.Cmd {
	m.CurrentStep = StepChangelogConfig
	confirmModel := bubble.NewConfirmModel(
		"Changelog Configuration",
		"Do you want to configure changelog features?",
	)
	m.confirmModel = &confirmModel
	return nil
}

// createOrUpdateEnvFile creates or updates the .env file with the API key
func createOrUpdateEnvFile(apiKey string) error {
	envPath := ".env"

	// Check if .env file exists
	var content string
	if data, err := os.ReadFile(envPath); err == nil {
		content = string(data)
	}

	// Add or update CLIKD_API_KEY
	envLine := fmt.Sprintf("CLIKD_API_KEY=%s\n", apiKey)

	// If file doesn't exist or is empty, just write the API key
	if content == "" {
		return os.WriteFile(envPath, []byte(envLine), 0644)
	}

	// For simplicity, just append the API key (in a real implementation,
	// you might want to check if CLIKD_API_KEY already exists and replace it)
	content += envLine

	return os.WriteFile(envPath, []byte(content), 0644)
}
