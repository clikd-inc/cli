package initialize

import (
	"clikd/internal/config"
	"clikd/internal/ui/bubble"
	"fmt"

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

	// Handle results from components
	case bubble.ConfirmResultMsg:
		m.confirmModel = nil
		switch m.CurrentStep {
		case StepConfigType:
			m.Global = !msg.Result // If not using local config, use global
			m.CurrentStep = StepCreateDirs
			return m, determineConfigPath(m)

		case StepConfirmOverwrite:
			if !msg.Result {
				m.Message = "Aborted, existing configuration will not be overwritten."
				m.MessageType = "info"
				m.Done = true
				return m, tea.Quit
			}
			m.Force = true
			return m, initConfigManager(m.ConfigPath, m.ConfigExists && !m.Force)

		case StepAIConfig:
			m.AIEnabled = msg.Result
			m.Manager.SetConfigValue("ai.enable", BoolToString(msg.Result))

			if msg.Result {
				// Set up provider selection
				return m, setupProviderSelection(m)
			} else {
				// Skip AI config, go to changelog
				m.CurrentStep = StepChangelogConfig
				return m, setupChangelogConfig(m)
			}
		}

	case bubble.SelectResultMsg:
		m.selectModel = nil
		switch m.CurrentStep {
		case StepGeneralConfig:
			if value, ok := msg.Value.(string); ok {
				m.Manager.SetConfigValue("general.log_level", value)
				// Setup color confirmation
				return m, setupColorConfig(m)
			}

		case StepProviderSelection:
			if value, ok := msg.Value.(string); ok {
				m.AIProvider = value
				m.Manager.SetConfigValue("ai.provider", value)
				return m, setupModelSelection(m, value)
			}

		case StepModelSelection:
			if value, ok := msg.Value.(string); ok {
				m.AIModel = value
				m.Manager.SetConfigValue("ai.model", value)
				return m, setupAdvancedOptions(m)
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

		// Setup the confirm component for repository configuration
		confirmModel := bubble.NewConfirmModel(
			"Choose Configuration Type",
			fmt.Sprintf("Git repository detected: %s\n\nDo you want to create a local configuration for this repository?", m.RepoURL),
		)
		m.confirmModel = &confirmModel
		return m, nil

	case ConfigPathMsg:
		m.ConfigPath = msg.ConfigPath
		m.ConfigDir = msg.ConfigDir
		return m, checkConfigExists(m.ConfigPath)

	case ConfigExistsMsg:
		m.ConfigExists = msg.Exists

		// If configuration exists and no force flag
		if m.ConfigExists && !m.Force && !m.Yes {
			confirmModel := bubble.NewConfirmModel(
				"Existing Configuration",
				fmt.Sprintf("Configuration file already exists at %s. Do you want to overwrite it?", m.ConfigPath),
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
			m.Manager.SetConfigValue("general.color", "true")
			m.CurrentStep = StepAIConfig
			return m, setupAIConfig(m)
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
		return m, nil

	case ProjectStructureErrorMsg:
		// Project structure creation failed
		m.Error = msg.Error
		m.Message = fmt.Sprintf("Error creating project structure: %s", msg.Error)
		m.MessageType = "error"
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

// Helper functions to set up components for different steps

func setupColorConfig(m InitModel) tea.Cmd {
	confirmModel := bubble.NewConfirmModel(
		"Terminal Color",
		"Enable colored terminal output?",
	)
	m.confirmModel = &confirmModel
	return nil
}

func setupAIConfig(m InitModel) tea.Cmd {
	confirmModel := bubble.NewConfirmModel(
		"AI Configuration",
		"Do you want to enable AI features?",
	)
	m.confirmModel = &confirmModel
	return nil
}

func setupProviderSelection(m InitModel) tea.Cmd {
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
	return nil
}

func setupModelSelection(m InitModel, provider string) tea.Cmd {
	m.CurrentStep = StepModelSelection

	// For simplicity, just hardcode some models for each provider
	var models []string
	switch provider {
	case "mistral":
		models = []string{"mistral-tiny", "mistral-small", "mistral-medium", "mistral-large"}
	case "anthropic":
		models = []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}
	case "openai":
		models = []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
	default:
		// Use default model if provider not recognized
		defaultModel, _ := config.GetDefaultModelForProvider(provider)
		m.AIModel = defaultModel
		m.Manager.SetConfigValue("ai.model", defaultModel)
		return setupAdvancedOptions(m)
	}

	modelItems := make([]bubble.SelectItem, len(models))
	for i, model := range models {
		modelItems[i] = bubble.SelectItem{
			Title:       model,
			Description: "", // Add description if available
			Value:       model,
		}
	}

	selectModel := bubble.NewSelectModel(fmt.Sprintf("Select Model for %s", provider), modelItems)
	m.selectModel = &selectModel
	return nil
}

func setupAdvancedOptions(m InitModel) tea.Cmd {
	m.CurrentStep = StepAdvancedAIOptions
	// For simplicity, we'll skip this for now and move directly to API key config
	return setupAPIKeyConfig(m)
}

func setupAPIKeyConfig(m InitModel) tea.Cmd {
	m.CurrentStep = StepAPIKeyConfig
	// For simplicity, we'll skip this for now and move directly to changelog config
	return setupChangelogConfig(m)
}

func setupChangelogConfig(m InitModel) tea.Cmd {
	m.CurrentStep = StepChangelogConfig
	// For simplicity, we'll go directly to project structure
	return createProjectStructure(m)
}
