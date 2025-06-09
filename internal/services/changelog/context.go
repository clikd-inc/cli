package changelog

import (
	"io"
)

// CLIContext enthält den Kontext für die CLI-Ausführung
type CLIContext struct {
	WorkingDir       string
	Stdout           io.Writer
	Stderr           io.Writer
	ConfigPath       string
	Template         string
	RepositoryURL    string
	OutputPath       string
	Silent           bool
	NoColor          bool
	NoEmoji          bool
	NoCaseSensitive  bool
	Query            string
	NextTag          string
	TagFilterPattern string
	JiraUsername     string
	JiraToken        string
	JiraURL          string
	Paths            []string
	Sort             string
}

// InitContext enthält den Kontext für die Initialisierung
type InitContext struct {
	WorkingDir string
	Stdout     io.Writer
	Stderr     io.Writer
	Style      string // Stil des Changelogs (z.B. "github", "gitlab")
	Template   string // Pfad zum Template
	ConfigDir  string // Verzeichnis für Konfigurationsdateien
}

// NewCLIContext erstellt einen neuen CLI-Kontext mit Standardwerten
func NewCLIContext(workingDir string, stdout, stderr io.Writer) *CLIContext {
	return &CLIContext{
		WorkingDir: workingDir,
		Stdout:     stdout,
		Stderr:     stderr,
		NoColor:    false,
		NoEmoji:    false,
	}
}
