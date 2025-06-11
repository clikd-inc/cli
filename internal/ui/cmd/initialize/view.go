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
		if m.IsInGitRepo && m.RepoURL != "" {
			content += "\n\n" + styles.SuccessText(fmt.Sprintf("✓ Git repository detected: %s", m.RepoURL))
		}

	case StepConfigType:
		if m.IsInGitRepo {
			content += styles.H2.Render("Choose Configuration Type") + "\n\n"
			content += styles.InfoText(fmt.Sprintf("Git repository detected: %s", m.RepoURL)) + "\n\n"
			content += styles.NormalText.Render("Select your preferred configuration scope:") + "\n\n"
		}

	case StepCreateDirs:
		content += styles.InfoText("Creating configuration directories...")

	case StepConfirmOverwrite:
		content += styles.H2.Render("Existing Configuration") + "\n\n"
		configType := "Local"
		configLocation := "clikd/"
		if m.Global {
			configType = "Global"
			configLocation = "~/.clikd/"
		}
		content += styles.NormalText.Render(fmt.Sprintf("%s configuration already exists in %s. Do you want to overwrite it?", configType, configLocation)) + "\n\n"

	case StepGeneralConfig:
		content += styles.SectionTitle("General Configuration") + "\n\n"
		content += styles.H2.Render("Select Log Level") + "\n\n"
		content += styles.InfoText("Loading options...")

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

	case StepAITokensInput:
		content += styles.SectionTitle("Max Input Tokens") + "\n\n"
		content += styles.InfoText("Maximum number of input tokens (context size)") + "\n"
		content += styles.InfoText("Higher values allow for larger context but cost more") + "\n\n"

	case StepAITokensOutput:
		content += styles.SectionTitle("Max Output Tokens") + "\n\n"
		content += styles.InfoText("Maximum number of output tokens (response length)") + "\n"
		content += styles.InfoText("Higher values allow for longer responses but cost more") + "\n\n"

	case StepAICustomURL:
		content += styles.SectionTitle("Custom API URL") + "\n\n"
		content += styles.InfoText("Custom API endpoint URL (leave empty to use official API)") + "\n"
		content += styles.InfoText("Use this for proxy servers or alternative endpoints") + "\n\n"

	case StepAICustomHeaders:
		content += styles.SectionTitle("Custom API Headers") + "\n\n"
		content += styles.InfoText("Custom HTTP headers in JSON format") + "\n"
		content += styles.InfoText("Leave empty for standard authentication") + "\n\n"

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

	case StepChangelogURL:
		content += styles.SectionTitle("Repository URL") + "\n\n"
		content += styles.NormalText.Render("What is the URL of your repository?") + "\n\n"

	case StepChangelogStyle:
		content += styles.SectionTitle("Changelog Style") + "\n\n"
		content += styles.NormalText.Render("Select the style for your changelog:") + "\n\n"

	case StepChangelogFormat:
		content += styles.SectionTitle("Commit Message Format") + "\n\n"
		content += styles.NormalText.Render("Choose the format of your favorite commit message:") + "\n\n"

	case StepChangelogTemplate:
		content += styles.SectionTitle("Template Style") + "\n\n"
		content += styles.NormalText.Render("What is your favorite template style?") + "\n\n"

	case StepChangelogColor:
		content += styles.SectionTitle("Terminal Color") + "\n\n"
		content += styles.NormalText.Render("Enable colored terminal output for changelog?") + "\n\n"

	case StepChangelogMerges:
		content += styles.SectionTitle("Merge Commits") + "\n\n"
		content += styles.NormalText.Render("Do you include Merge Commit in CHANGELOG?") + "\n\n"

	case StepChangelogReverts:
		content += styles.SectionTitle("Revert Commits") + "\n\n"
		content += styles.NormalText.Render("Do you include Revert Commit in CHANGELOG?") + "\n\n"

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
		content += "\n"

		// AI Configuration
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

		// Changelog Configuration
		if m.ChangelogEnabled {
			content += styles.SuccessIcon + " " + styles.BoldText.Render("Changelog:") + "\n"
			content += "   - Style: " + styles.NormalText.Render(m.ChangelogStyle) + "\n"
			content += "   - Format: " + styles.NormalText.Render(m.ChangelogFormat) + "\n"
			content += "   - Template: " + styles.NormalText.Render(m.ChangelogTemplate) + "\n"
			content += "   - Color Output: " + styles.NormalText.Render(BoolToString(m.ChangelogColorEnabled)) + "\n"
			content += "   - Include Merges: " + styles.NormalText.Render(BoolToString(m.ChangelogIncludeMerges)) + "\n"
			content += "   - Include Reverts: " + styles.NormalText.Render(BoolToString(m.ChangelogIncludeReverts)) + "\n"
			content += "   - Config Dir: " + styles.NormalText.Render("clikd/changelog/") + "\n\n"
		} else {
			content += styles.WarningIcon + " " + styles.BoldText.Render("Changelog: Disabled") + "\n\n"
		}

		// Next steps
		content += styles.SectionTitle("Next Steps") + "\n\n"

		// API Key configuration (if needed)
		if m.ApiKeyStatus != "done" {
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
		if m.ChangelogEnabled {
			stepNum := "1"
			if m.ApiKeyStatus != "done" {
				stepNum = "2"
			}
			content += styles.ArrowIcon + " " + styles.BoldText.Render(stepNum+". Generate a changelog:") + "\n"
			content += "   " + styles.HighlightStyle.Render("clikd changelog -o CHANGELOG.md") + "\n\n"
		}

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
