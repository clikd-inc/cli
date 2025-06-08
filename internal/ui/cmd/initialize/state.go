package initialize

import (
	"fmt"

	"clikd/internal/config"
	"clikd/internal/ui/bubble"

	tea "github.com/charmbracelet/bubbletea"
)

// Steps for initialization
const (
	StepStart             = "start"
	StepCheckRepo         = "check_repo"
	StepConfigType        = "config_type"
	StepCreateDirs        = "create_dirs"
	StepConfirmOverwrite  = "confirm_overwrite"
	StepGeneralConfig     = "general_config"
	StepColorConfig       = "color_config"
	StepAIConfig          = "ai_config"
	StepProviderSelection = "provider_selection"
	StepModelSelection    = "model_selection"
	StepAdvancedAIOptions = "advanced_ai_options"
	StepAPIKeyConfig      = "api_key_config"
	StepChangelogConfig   = "changelog_config"
	StepChangelogStyle    = "changelog_style"
	StepChangelogJIRA     = "changelog_jira"
	StepChangelogSort     = "changelog_sort"
	StepChangelogAdvanced = "changelog_advanced"
	StepChangelogCase     = "changelog_case"
	StepProjectStructure  = "project_structure"
	StepSummary           = "summary"
	StepComplete          = "complete"
)

// Error types
var (
	ErrInitAborted = fmt.Errorf("initialization aborted by user")
	ErrNoHomeDir   = fmt.Errorf("could not determine home directory")
)

// Message types for the update function

// GitRepoMsg contains information about the Git repository
type GitRepoMsg struct {
	IsInGitRepo bool
	RepoURL     string
}

// ConfigPathMsg contains information about the configuration path
type ConfigPathMsg struct {
	ConfigPath string
	ConfigDir  string
}

// ConfigExistsMsg indicates whether the configuration file exists
type ConfigExistsMsg struct {
	Exists bool
}

// ConfigManagerMsg contains the configuration manager
type ConfigManagerMsg struct {
	Manager *config.Manager
}

// ProjectStructureErrorMsg contains an error when creating the project structure
type ProjectStructureErrorMsg struct {
	Error error
}

// ProjectStructureCompleteMsg indicates that the project structure was successfully created
type ProjectStructureCompleteMsg struct{}

// LogLevelSelectedMsg indicates that a log level was selected
type LogLevelSelectedMsg struct {
	LogLevel string
}

// ColorConfigSelectedMsg indicates that the color configuration was selected
type ColorConfigSelectedMsg struct {
	Enabled bool
}

// AIConfigSelectedMsg indicates that the AI configuration was selected
type AIConfigSelectedMsg struct {
	Enabled bool
}

// ProviderSelectedMsg indicates that an AI provider was selected
type ProviderSelectedMsg struct {
	Provider string
}

// ModelSelectedMsg indicates that an AI model was selected
type ModelSelectedMsg struct {
	Model string
}

// EnvFileExistsMsg indicates whether a .env file exists in the current directory
type EnvFileExistsMsg struct {
	Exists bool
}

// AIOptionsMsg signals that AI options configuration is complete
type AIOptionsMsg struct{}

// ConfirmResult is sent when a confirmation dialog is completed
type ConfirmResult struct {
	Result bool
	Step   string
}

// SelectResult is sent when a selection is made
type SelectResult struct {
	Value interface{}
	Step  string
}

// ForceStepChangeMsg ist eine spezielle Nachricht, die verwendet wird, um einen Schrittwechsel zu erzwingen
type ForceStepChangeMsg struct {
	NewStep string
}

// StepCompleteMsg is sent when a step is completed
type StepCompleteMsg struct {
	NextStep string
}

// InitModel represents the main model for the initialization process
type InitModel struct {
	// Configuration options
	Global bool
	Force  bool
	Yes    bool

	// State variables
	CurrentStep      string
	ConfigPath       string
	ConfigDir        string
	IsInGitRepo      bool
	RepoURL          string
	ApiKeyStatus     string
	ConfigExists     bool
	Manager          *config.Manager
	AIEnabled        bool
	AIProvider       string
	AIModel          string
	AICustomSettings bool

	// UI components (real Bubble Tea models)
	confirmModel  *bubble.ConfirmModel
	selectModel   *bubble.SelectModel
	progressModel *bubble.ProgressModel
	inputModel    *bubble.InputModel

	// UI state
	ProgressPercent float64
	Message         string
	MessageType     string // "success", "error", "info", "warning"

	// Results and errors
	Error error
	Done  bool
}

// NewInitModel creates a new initialization model
func NewInitModel(global, force, yes bool) InitModel {
	return InitModel{
		Global:          global,
		Force:           force,
		Yes:             yes,
		CurrentStep:     StepStart,
		ApiKeyStatus:    "pending",
		ProgressPercent: 0,
	}
}

// Init initializes the model
func (m InitModel) Init() tea.Cmd {
	return checkGitRepo
}

// View renders the model
func (m InitModel) View() string {
	return View(m)
}
