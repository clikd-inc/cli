package initialize

import (
	"fmt"
	"strings"

	"clikd/internal/ui/bubble"
	"clikd/internal/ui/styles"
)

// View renders the model
func View(m InitModel) string {
	if m.Error != nil {
		return styles.ErrorText(fmt.Sprintf("Error: %s", m.Error))
	}

	// Main content based on current step
	content := ""

	// Logo and title only once at the beginning
	content += styles.RenderLogo() + "\n"
	content += styles.H1.Render("Welcome to the clikd Configuration Assistant!") + "\n"
	content += styles.NormalText.Render("This assistant helps you set up clikd for your project.") + "\n\n"

	// If we have an active component, just render that
	if m.confirmModel != nil {
		return content + m.confirmModel.View()
	}

	if m.selectModel != nil {
		return content + m.selectModel.View()
	}

	if m.inputModel != nil {
		return content + m.inputModel.View()
	}

	if m.progressModel != nil {
		return content + m.progressModel.View()
	}

	// Main content depending on step - only show the current step
	switch m.CurrentStep {
	case StepStart:
		content += styles.InfoText("Initialization in progress...")

	case StepConfigType:
		if m.IsInGitRepo {
			content += styles.H2.Render("Choose Configuration Type") + "\n\n"
			content += styles.NormalText.Render("Do you want to create a local configuration for this repository?") + "\n\n"
		}

	case StepCreateDirs:
		content += styles.InfoText("Creating configuration directories...")

	case StepConfirmOverwrite:
		content += styles.H2.Render("Existing Configuration") + "\n\n"
		content += styles.NormalText.Render(fmt.Sprintf("Configuration file already exists at %s. Do you want to overwrite it?", m.ConfigPath)) + "\n\n"

	case StepGeneralConfig:
		content += styles.SectionTitle("General Configuration") + "\n\n"
		content += styles.H2.Render("Select Log Level") + "\n\n"
		content += styles.InfoText("Loading options...")

	case StepAIConfig:
		content += styles.SectionTitle("AI Configuration") + "\n\n"
		content += styles.NormalText.Render("Do you want to enable AI features?") + "\n\n"

	case StepProviderSelection:
		content += styles.SectionTitle("AI Provider") + "\n\n"
		content += styles.InfoText("We recommend using Mistral as your AI provider for the best balance of cost and performance.") + "\n"
		content += styles.InfoText("Mistral offers high-quality models with more reasonable pricing than other providers.") + "\n\n"
		content += styles.NormalText.Render("Select a provider:") + "\n\n"

	case StepModelSelection:
		content += styles.SectionTitle("AI Model") + "\n\n"
		content += styles.NormalText.Render(fmt.Sprintf("Select a model for %s:", m.AIProvider)) + "\n\n"

	case StepAdvancedAIOptions:
		content += styles.SectionTitle("Advanced AI Options") + "\n\n"
		content += styles.InfoText("Advanced AI options allow you to configure technical parameters such as:") + "\n"
		content += styles.InfoText("  - Maximum input tokens (default: 4096) - controls context size") + "\n"
		content += styles.InfoText("  - Maximum output tokens (default: 500) - controls response length") + "\n"
		content += styles.InfoText("  - Custom API endpoints (default: official provider endpoints)") + "\n"
		content += styles.InfoText("  - Custom HTTP headers (default: none, standard authentication)") + "\n\n"

	case StepAPIKeyConfig:
		content += styles.SectionTitle("API Key Configuration") + "\n\n"

		if m.Global {
			content += styles.InfoText(fmt.Sprintf("You can get API keys at the %s website", m.AIProvider)) + "\n\n"
		} else {
			content += styles.InfoText("For local projects, the API key is loaded from the .env file in the project directory:") + "\n"
			content += styles.HighlightStyle.Render("  CLIKD_API_KEY=YOUR_API_KEY") + "\n\n"
		}

	case StepChangelogConfig:
		content += styles.SectionTitle("Changelog Configuration") + "\n\n"
		content += styles.NormalText.Render("Do you want to configure changelog features?") + "\n\n"

	case StepChangelogStyle:
		content += styles.SectionTitle("Changelog Style") + "\n\n"
		content += styles.NormalText.Render("Select the style for your changelog:") + "\n\n"

	case StepChangelogJIRA:
		content += styles.SectionTitle("JIRA Integration") + "\n\n"
		content += styles.NormalText.Render("Do you want to enable JIRA integration for your changelog?") + "\n\n"

	case StepChangelogSort:
		content += styles.SectionTitle("Changelog Sort Order") + "\n\n"
		content += styles.NormalText.Render("How do you want to sort entries in your changelog?") + "\n\n"

	case StepChangelogAdvanced:
		content += styles.SectionTitle("Advanced Changelog Options") + "\n\n"
		content += styles.NormalText.Render("Do you want to configure advanced changelog options?") + "\n\n"

	case StepChangelogCase:
		content += styles.SectionTitle("Changelog Case Sensitivity") + "\n\n"
		content += styles.NormalText.Render("Make changelog generation case-insensitive?") + "\n\n"

	case StepProjectStructure:
		content += styles.SectionTitle("Creating Project Structure") + "\n\n"
		content += styles.InfoText("Setting up directories and templates...") + "\n\n"

		if m.ProgressPercent > 0 {
			progressView := bubble.NewProgressModel("", "", 50)
			progressView.Percent = m.ProgressPercent
			content += progressView.Progress.View() + "\n\n"
		}

	case StepSummary:
		content += styles.H1.Render("Configuration Completed") + "\n\n"
		content += styles.SectionTitle("Configuration Summary") + "\n\n"

		// General configuration
		content += styles.SuccessIcon + " " + styles.BoldText.Render("General:") + "\n"
		content += "   - Log Level: " + styles.NormalText.Render(m.Manager.GetConfig().General.LogLevel) + "\n"
		content += "   - Color: " + styles.NormalText.Render(BoolToString(m.Manager.GetConfig().General.Color)) + "\n\n"

		// AI Configuration
		if m.AIEnabled {
			content += styles.SuccessIcon + " " + styles.BoldText.Render("AI:") + "\n"
			content += "   - Provider: " + styles.NormalText.Render(m.AIProvider) + "\n"
			content += "   - Model: " + styles.NormalText.Render(m.AIModel) + "\n"

			if m.ApiKeyStatus == "check" {
				content += "   - API Key: " + styles.InfoText("Please add to .env file") + "\n\n"
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
			content += "\n" + styles.SuccessStyle.Render(m.Message)
		case "error":
			content += "\n" + styles.ErrorStyle.Render(m.Message)
		case "info":
			content += "\n" + styles.InfoStyle.Render(m.Message)
		case "warning":
			content += "\n" + styles.WarningStyle.Render(m.Message)
		}
	}

	return content
}

// renderSelectOptions renders the selection options
func renderSelectOptions(m InitModel) string {
	content := ""

	if m.selectModel != nil {
		return m.selectModel.View()
	}

	return content
}
