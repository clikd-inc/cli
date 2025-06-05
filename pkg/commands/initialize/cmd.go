package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"clikd/pkg/config"

	"github.com/spf13/cobra"
)

// NewInitCmd erstellt einen neuen Init-Befehl
func NewInitCmd() *cobra.Command {
	var forceFlag bool
	var globalFlag bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new clikd configuration",
		Long: `Initialize a new clikd configuration file.
This will create a new configuration file in the current directory or globally in your home directory.
If a configuration file already exists, it will not be overwritten unless the --force flag is used.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Bestimmen des Zielverzeichnisses
			var basePath string
			var clikdDirPath string

			if globalFlag {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("could not determine home directory: %w", err)
				}
				basePath = filepath.Join(homeDir, ".clikd")
				clikdDirPath = basePath
			} else {
				// Aktuelles Arbeitsverzeichnis verwenden
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("could not determine working directory: %w", err)
				}

				// Bei lokaler Konfiguration einen "clikd" Ordner erstellen
				clikdDirPath = filepath.Join(wd, "clikd")

				// Prüfe, ob ein Element mit dem Namen "clikd" bereits existiert
				info, err := os.Stat(clikdDirPath)
				if err == nil {
					// Element existiert
					if !info.IsDir() {
						// Es ist eine Datei, kein Verzeichnis
						return fmt.Errorf("cannot create directory %s: a file with this name already exists", clikdDirPath)
					}
				}

				basePath = clikdDirPath
			}

			// Überprüfen, ob die Konfigurationsdatei bereits existiert
			configFile := filepath.Join(basePath, "config.toml")
			_, err := os.Stat(configFile)
			if err == nil && !forceFlag {
				return fmt.Errorf("configuration file already exists at %s. Use --force to overwrite", configFile)
			}

			// Erstellen der Verzeichnisstruktur
			dirs := []string{
				basePath,
				filepath.Join(basePath, "templates"),
				filepath.Join(basePath, "cache"),
			}

			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("could not create directory %s: %w", dir, err)
				}
			}

			// Erstellen der Template-Datei für den Changelog
			changelogTemplate := filepath.Join(basePath, "templates", "changelog.md")
			if globalFlag || !fileExists(changelogTemplate) || forceFlag {
				defaultTemplate := `# {{ .Info.Title }}

All notable changes to this project will be documented in this file.

{{ if .Versions -}}
{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }} ({{ datetime "2006-01-02" .Tag.Date }})

{{ range .CommitGroups -}}
### {{ .Title }}

{{ range .Commits -}}
* {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .RevertCommits -}}
### Reverts

{{ range .RevertCommits -}}
* {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
### Merges

{{ range .MergeCommits -}}
* {{ .Header }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}

{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}`

				if err := os.WriteFile(changelogTemplate, []byte(defaultTemplate), 0644); err != nil {
					return fmt.Errorf("could not create changelog template: %w", err)
				}
			}

			// Einen neuen Konfigurationsmanager erstellen
			manager := config.NewManager()

			// Initialisieren mit Standardwerten
			manager.InitConfig("")

			// Changelog-Template-Pfad in der Konfiguration anpassen
			if !globalFlag {
				// Relativen Pfad zum Template setzen für Projektkonfiguration
				if err := manager.SetConfigValue("changelog.template", "templates/changelog.md"); err != nil {
					return fmt.Errorf("could not update template path: %w", err)
				}
			}

			// Konfiguration speichern
			if err := manager.SaveConfig(configFile); err != nil {
				return fmt.Errorf("could not save configuration: %w", err)
			}

			cmd.Printf("Configuration initialized at %s\n", configFile)
			if !globalFlag {
				cmd.Println("Project structure created:")
				cmd.Println("  ├── clikd/")
				cmd.Println("  │   ├── config.toml")
				cmd.Println("  │   ├── templates/")
				cmd.Println("  │   │   └── changelog.md")
				cmd.Println("  │   └── cache/")
			}
			return nil
		},
	}

	// Flags hinzufügen
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVarP(&globalFlag, "global", "g", false, "Create a global configuration in ~/.clikd")

	return cmd
}

// fileExists prüft, ob eine Datei existiert
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
