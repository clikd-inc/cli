package changelog

import (
	"fmt"
	"os"
	"path/filepath"

	"clikd/internal/services/changelog/configs"
	tbuilders "clikd/internal/services/changelog/template_builders"
	tpls "clikd/internal/services/changelog/templates"
)

// Service stellt Funktionen für die Changelog-Verwaltung bereit
type Service struct {
	ConfigPath string
}

// NewService erstellt einen neuen Changelog-Service
func NewService(configPath string) *Service {
	return &Service{
		ConfigPath: configPath,
	}
}

// InitializeTemplates erstellt die Template- und Konfigurationsdateien
func (s *Service) InitializeTemplates(style string, configDir string) error {
	// Erstelle die Verzeichnisse
	templateDir := filepath.Join(configDir, "templates")
	configDir = filepath.Join(configDir, "config")

	dirs := []string{templateDir, configDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("fehler beim Erstellen des Verzeichnisses %s: %w", dir, err)
		}
	}

	// Template- und Konfigurationsdateien schreiben
	templatePath := filepath.Join(templateDir, style+".tpl.md")
	configPath := filepath.Join(configDir, style+".yml")

	// Hole den passenden Template-Builder
	builder := tbuilders.GetTemplateBuilder(style)

	// Erstelle ein Answer-Objekt
	answer := &Answer{
		Style:               style,
		Template:            style,
		CommitMessageFormat: "type(scope): subject",
		IncludeMerges:       true,
		IncludeReverts:      true,
		ConfigDir:           configDir,
	}

	// Generiere das Template
	templateContent, err := builder.Build(answer)
	if err != nil {
		return fmt.Errorf("fehler beim Generieren des Templates: %w", err)
	}

	// Schreibe das Template
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("fehler beim Schreiben der Template-Datei: %w", err)
	}

	// Schreibe die Konfiguration
	if err := os.WriteFile(configPath, []byte(configs.GetConfig(style)), 0644); err != nil {
		return fmt.Errorf("fehler beim Schreiben der Konfigurations-Datei: %w", err)
	}

	return nil
}

// EnsureTemplateExists stellt sicher, dass die Template-Datei existiert
// Falls nicht, wird sie aus dem eingebetteten Template wiederhergestellt
func (s *Service) EnsureTemplateExists(templatePath, style string) error {
	if templatePath == "" {
		return nil // Keine Template-Datei konfiguriert
	}

	// Prüfe, ob die Datei existiert
	if _, err := os.Stat(templatePath); err == nil {
		return nil // Datei existiert
	}

	// Stelle sicher, dass das Verzeichnis existiert
	dir := filepath.Dir(templatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Template-Verzeichnisses: %w", err)
	}

	// Schreibe das Template
	if err := os.WriteFile(templatePath, []byte(tpls.GetTemplate(style)), 0644); err != nil {
		return fmt.Errorf("fehler beim Wiederherstellen des Templates: %w", err)
	}

	fmt.Printf("Template-Datei wurde wiederhergestellt: %s\n", templatePath)
	return nil
}

// EnsureConfigExists stellt sicher, dass die Konfigurations-Datei existiert
// Falls nicht, wird sie aus der eingebetteten Konfiguration wiederhergestellt
func (s *Service) EnsureConfigExists(configPath, style string) error {
	if configPath == "" {
		return nil // Keine Konfigurations-Datei konfiguriert
	}

	// Prüfe, ob die Datei existiert
	if _, err := os.Stat(configPath); err == nil {
		return nil // Datei existiert
	}

	// Stelle sicher, dass das Verzeichnis existiert
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Konfigurations-Verzeichnisses: %w", err)
	}

	// Schreibe die Konfiguration
	if err := os.WriteFile(configPath, []byte(configs.GetConfig(style)), 0644); err != nil {
		return fmt.Errorf("fehler beim Wiederherstellen der Konfiguration: %w", err)
	}

	fmt.Printf("Konfigurations-Datei wurde wiederhergestellt: %s\n", configPath)
	return nil
}
