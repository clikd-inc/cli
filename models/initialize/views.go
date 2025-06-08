package initialize

import (
	"fmt"
	"strings"

	"clikd/internal/styles"
)

// View renders the model
func (m InitModel) View() string {
	if m.Error != nil {
		return styles.ErrorText(fmt.Sprintf("Error: %s", m.Error))
	}

	// Main content based on current step
	content := ""

	// Logo and heading
	content += styles.RenderLogo() + "\n"
	content += styles.H1.Render("Welcome to the clikd Configuration Assistant!") + "\n"
	content += styles.NormalText.Render("This assistant helps you set up clikd for your project.") + "\n\n"

	// Main content depending on step
	switch m.CurrentStep {
	case StepStart:
		content += styles.InfoText("Initialization in progress...")

	case StepConfigType:
		if m.IsInGitRepo {
			content += styles.InfoText("Git repository detected: "+m.RepoURL) + "\n\n"
			content += styles.H2.Render("Choose Configuration Type") + "\n\n"
			content += styles.NormalText.Render("Do you want to create a local configuration for this repository?") + "\n\n"

			// Display options
			content += m.renderSelectOptions()
		}

	case "confirm_overwrite":
		content += styles.WarningText("Configuration file already exists at "+m.ConfigPath) + "\n\n"
		content += styles.H2.Render("Existing Configuration") + "\n\n"
		content += styles.NormalText.Render("Do you want to overwrite the existing configuration?") + "\n\n"

		// Display options
		content += m.renderSelectOptions()

	case StepGeneralConfig:
		content += styles.SectionTitle("General Configuration") + "\n\n"
		content += styles.NormalText.Render("Select log level:") + "\n\n"

		// Display options if we have them
		if len(m.SelectOptions) > 0 {
			content += m.renderSelectOptions()
		} else {
			content += styles.InfoText("Loading options...")
		}

	case "color_config":
		content += styles.SectionTitle("Terminal Color") + "\n\n"
		content += styles.NormalText.Render("Enable colored terminal output?") + "\n\n"
		content += m.renderSelectOptions()

	case StepAIConfig:
		content += styles.SectionTitle("AI Configuration") + "\n\n"
		content += styles.NormalText.Render("Do you want to enable AI features?") + "\n\n"

		// Display options if we have them
		if len(m.SelectOptions) > 0 {
			content += m.renderSelectOptions()
		}

	case StepProviderSelection:
		content += styles.SectionTitle("AI Provider") + "\n\n"
		content += styles.InfoText("We recommend using Mistral as your AI provider for the best balance of cost and performance.") + "\n"
		content += styles.InfoText("Mistral offers high-quality models with more reasonable pricing than other providers.") + "\n\n"
		content += styles.NormalText.Render("Select a provider:") + "\n\n"

		// Display options if we have them
		if len(m.SelectOptions) > 0 {
			content += m.renderSelectOptions()
		}

	case StepModelSelection:
		content += styles.SectionTitle("AI Model") + "\n\n"
		content += styles.NormalText.Render(fmt.Sprintf("Select a model for %s:", m.AIProvider)) + "\n\n"

		// Display options if we have them
		if len(m.SelectOptions) > 0 {
			content += m.renderSelectOptions()
		}

	case StepAdvancedAIOptions:
		content += styles.SectionTitle("Advanced AI Options") + "\n\n"
		content += styles.InfoText("Advanced AI options allow you to configure technical parameters such as:") + "\n"
		content += styles.InfoText("  - Maximum input tokens (default: 4096) - controls context size") + "\n"
		content += styles.InfoText("  - Maximum output tokens (default: 500) - controls response length") + "\n"
		content += styles.InfoText("  - Custom API endpoints (default: official provider endpoints)") + "\n"
		content += styles.InfoText("  - Custom HTTP headers (default: none, standard authentication)") + "\n\n"

		// If we're in input mode, show the current input
		if m.ActiveInput != "" {
			content += styles.H2.Render(m.ActiveInput) + "\n"
			content += m.TextInput.View() + "\n\n"
		} else if len(m.SelectOptions) > 0 {
			content += m.renderSelectOptions()
		}

	case StepAPIKeyConfig:
		content += styles.SectionTitle("API Key Configuration") + "\n\n"

		// Show different content based on whether it's global config or local project
		if m.Global {
			content += styles.InfoText(fmt.Sprintf("You can get API keys at the %s website", m.AIProvider)) + "\n\n"

			if m.ActiveInput != "" {
				content += styles.H2.Render("API Key") + "\n"
				content += styles.InfoText("Enter your API key (or press Enter to do this later):") + "\n\n"
				content += m.TextInput.View() + "\n"
			}
		} else {
			content += styles.InfoText("For local projects, the API key is loaded from the .env file in the project directory:") + "\n"
			content += styles.HighlightStyle.Render("  CLIKD_API_KEY=YOUR_API_KEY") + "\n\n"

			if len(m.SelectOptions) > 0 {
				content += m.renderSelectOptions()
			} else if m.ActiveInput == "API Key" {
				content += styles.H2.Render("API Key") + "\n"
				content += styles.InfoText("Enter your API key for the .env file:") + "\n\n"
				content += m.TextInput.View() + "\n"
			}
		}

	case StepChangelogConfig:
		content += styles.SectionTitle("Changelog Configuration") + "\n\n"
		content += styles.NormalText.Render("Do you want to configure changelog features?") + "\n\n"

		if len(m.SelectOptions) > 0 {
			content += m.renderSelectOptions()
		}

	case "changelog_style":
		content += styles.SectionTitle("Changelog Style") + "\n\n"
		content += styles.NormalText.Render("Select the style for your changelog:") + "\n\n"
		content += m.renderSelectOptions()

	case "changelog_jira":
		content += styles.SectionTitle("JIRA Integration") + "\n\n"
		content += styles.NormalText.Render("Do you want to enable JIRA integration for your changelog?") + "\n\n"

		if m.ActiveInput == "JIRA Prefix" {
			content += styles.H2.Render("JIRA Prefix") + "\n"
			content += styles.InfoText("Enter your JIRA project prefix (e.g., PROJ):") + "\n\n"
			content += m.TextInput.View() + "\n\n"
			content += styles.InfoText("Press Enter to confirm or use default ('PROJ')") + "\n"
		} else {
			content += m.renderSelectOptions()
		}

	case "changelog_sort":
		content += styles.SectionTitle("Changelog Sort Order") + "\n\n"
		content += styles.NormalText.Render("How do you want to sort entries in your changelog?") + "\n\n"
		content += m.renderSelectOptions()

	case "changelog_advanced":
		content += styles.SectionTitle("Advanced Changelog Options") + "\n\n"
		content += styles.NormalText.Render("Do you want to configure advanced changelog options?") + "\n\n"

		if m.ActiveInput != "" {
			content += styles.H2.Render(m.ActiveInput) + "\n"
			content += m.TextInput.View() + "\n\n"
		} else {
			content += m.renderSelectOptions()
		}

	case "changelog_case":
		content += styles.SectionTitle("Changelog Case Sensitivity") + "\n\n"
		content += styles.NormalText.Render("Make changelog generation case-insensitive?") + "\n\n"
		content += m.renderSelectOptions()

	case StepProjectStructure:
		content += styles.SectionTitle("Creating Project Structure") + "\n\n"
		content += styles.InfoText("Setting up directories and templates...") + "\n\n"

		// Display progress if available
		if m.Progress.Percent() > 0 {
			content += m.Progress.View() + "\n\n"
		}

	case StepSummary:
		content += styles.H1.Render("Configuration Completed") + "\n\n"
		content += styles.SectionTitle("Configuration Summary") + "\n\n"

		// General configuration
		content += styles.SuccessIcon + " " + styles.BoldText.Render("General:") + "\n"
		content += "   - Log Level: " + styles.NormalText.Render(m.Manager.GetConfig().General.LogLevel) + "\n"
		content += "   - Config File: " + styles.NormalText.Render(m.ConfigPath) + "\n\n"

		// AI configuration (if enabled)
		if m.AIEnabled {
			content += styles.SuccessIcon + " " + styles.BoldText.Render("AI Configuration:") + "\n"
			content += "   - Provider: " + styles.NormalText.Render(m.AIProvider) + "\n"
			content += "   - Model: " + styles.NormalText.Render(m.AIModel) + "\n"

			// API Key status
			if m.ApiKeyStatus == "check" {
				content += "   - API Key: " + styles.HighlightStyle.Render("Make sure your .env file contains CLIKD_API_KEY=YOUR_API_KEY") + "\n\n"
			} else if m.ApiKeyStatus == "done" {
				content += "   - API Key: " + styles.SuccessText("Configured") + "\n\n"
			} else {
				content += "   - API Key: " + styles.WarningText("Not yet configured") + "\n\n"
			}
		} else {
			content += styles.WarningIcon + " " + styles.BoldText.Render("AI Features: Disabled") + "\n\n"
		}

		// Next steps
		content += styles.SectionTitle("Next Steps") + "\n\n"

		// API Key configuration (if needed)
		if m.AIEnabled && m.ApiKeyStatus != "done" {
			if m.ApiKeyStatus == "check" {
				content += styles.SuccessIcon + " " + styles.BoldText.Render("1. Check API Key in .env file:") + "\n"
				content += "   " + styles.HighlightStyle.Render(fmt.Sprintf("CLIKD_API_KEY=YOUR_%s_API_KEY", strings.ToUpper(m.AIProvider))) + "\n\n"
			} else if m.Global {
				content += styles.ArrowIcon + " " + styles.BoldText.Render("1. Configure API Key:") + "\n"
				content += "   " + styles.HighlightStyle.Render("clikd init config set ai.api_key=YOUR_API_KEY") + "\n\n"
			} else {
				content += styles.ArrowIcon + " " + styles.BoldText.Render("1. Create .env file with API Key:") + "\n"
				content += "   " + styles.HighlightStyle.Render(fmt.Sprintf("CLIKD_API_KEY=YOUR_%s_API_KEY", strings.ToUpper(m.AIProvider))) + "\n\n"
			}
		}

		// Changelog generation
		if m.AIEnabled {
			content += styles.ArrowIcon + " " + styles.BoldText.Render("2. Generate a changelog:") + "\n"
		} else {
			content += styles.ArrowIcon + " " + styles.BoldText.Render("1. Generate a changelog:") + "\n"
		}
		content += "   " + styles.HighlightStyle.Render("clikd changelog -o CHANGELOG.md") + "\n\n"

	case StepComplete:
		content += styles.SuccessText("clikd is now ready to use! Enjoy!")
	}

	// Display messages
	if m.Message != "" {
		switch m.MessageType {
		case "success":
			content += "\n\n" + styles.SuccessText(m.Message)
		case "error":
			content += "\n\n" + styles.ErrorText(m.Message)
		case "warning":
			content += "\n\n" + styles.WarningText(m.Message)
		default:
			content += "\n\n" + styles.InfoText(m.Message)
		}
	}

	// Wrap in a box
	return styles.BoxedContent(content)
}

// renderSelectOptions renders the selection options
func (m InitModel) renderSelectOptions() string {
	content := ""

	for i, option := range m.SelectOptions {
		// Cursor and styling for selected option
		if i == m.SelectCursor {
			content += styles.SelectedStyle.Render(fmt.Sprintf("> [%s]", option.Title))
		} else {
			content += styles.UnselectedStyle.Render(fmt.Sprintf("  [%s]", option.Title))
		}

		if option.Description != "" {
			content += " - " + styles.Subtle.Render(option.Description)
		}

		content += "\n"
	}

	content += "\n" + styles.Subtle.Render("(Use arrow keys and Enter to select)")

	return content
}
