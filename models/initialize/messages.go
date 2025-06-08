package initialize

import (
	"clikd/internal/config"
	"fmt"
)

// Nachrichtentypen für die Update-Funktion

// GitRepoMsg enthält Informationen über das Git-Repository
type GitRepoMsg struct {
	IsInGitRepo bool
	RepoURL     string
}

// ConfigPathMsg enthält Informationen über den Konfigurationspfad
type ConfigPathMsg struct {
	ConfigPath string
	ConfigDir  string
}

// ConfigExistsMsg zeigt an, ob die Konfigurationsdatei existiert
type ConfigExistsMsg struct {
	Exists bool
}

// ConfigManagerMsg enthält den Konfigurationsmanager
type ConfigManagerMsg struct {
	Manager *config.Manager
}

// ProjectStructureErrorMsg enthält einen Fehler bei der Erstellung der Projektstruktur
type ProjectStructureErrorMsg struct {
	Error error
}

// ProjectStructureCompleteMsg zeigt an, dass die Projektstruktur erfolgreich erstellt wurde
type ProjectStructureCompleteMsg struct{}

// LogLevelSelectedMsg zeigt an, dass ein Log-Level ausgewählt wurde
type LogLevelSelectedMsg struct {
	LogLevel string
}

// ColorConfigSelectedMsg zeigt an, dass die Farbkonfiguration ausgewählt wurde
type ColorConfigSelectedMsg struct {
	Enabled bool
}

// AIConfigSelectedMsg zeigt an, dass die KI-Konfiguration ausgewählt wurde
type AIConfigSelectedMsg struct {
	Enabled bool
}

// ProviderSelectedMsg zeigt an, dass ein KI-Provider ausgewählt wurde
type ProviderSelectedMsg struct {
	Provider string
}

// ModelSelectedMsg zeigt an, dass ein KI-Modell ausgewählt wurde
type ModelSelectedMsg struct {
	Model string
}

// EnvFileExistsMsg indicates whether a .env file exists in the current directory
type EnvFileExistsMsg struct {
	Exists bool
}

// AIOptionsMsg signals that AI options configuration is complete
type AIOptionsMsg struct{}

// Fehlertypen
var (
	ErrInitAborted = fmt.Errorf("initialization aborted by user")
	ErrNoHomeDir   = fmt.Errorf("could not determine home directory")
)

// Schritte für die Initialisierung
const (
	StepStart             = "start"
	StepCheckRepo         = "check_repo"
	StepConfigType        = "config_type"
	StepCreateDirs        = "create_dirs"
	StepGeneralConfig     = "general_config"
	StepAIConfig          = "ai_config"
	StepProviderSelection = "provider_selection"
	StepModelSelection    = "model_selection"
	StepAdvancedAIOptions = "advanced_ai_options"
	StepAPIKeyConfig      = "api_key_config"
	StepChangelogConfig   = "changelog_config"
	StepProjectStructure  = "project_structure"
	StepSummary           = "summary"
	StepComplete          = "complete"
)
