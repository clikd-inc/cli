package utils

import (
	"io"
	"os"
	"path/filepath"
)

// DotGet ruft einen Wert aus einem Objekt über einen Punktnotation-Pfad ab
// Beispiel: DotGet(commit, "Author.Name") würde commit.Author.Name zurückgeben
// Diese Funktion ist jetzt in gitutils.go definiert und wird hier nicht mehr benötigt

// AssignDynamicValues weist dynamisch Werte zu Struct-Feldern zu
// Diese Funktion ist jetzt in gitutils.go definiert und wird hier nicht mehr benötigt

// CompareValues vergleicht zwei Werte mit einem bestimmten Operator
// Diese Funktion ist jetzt in gitutils.go definiert und wird hier nicht mehr benötigt

// ConvNewline konvertiert verschiedene Zeilenumbruchformate in das angegebene Format
// Diese Funktion ist jetzt in gitutils.go definiert und wird hier nicht mehr benötigt

// CreateOutputWriter erstellt einen Writer für die Ausgabe von Changelog-Daten
func CreateOutputWriter(outputPath string, logger Logger) (io.Writer, error) {
	if outputPath == "" {
		logger.Debug("Using stdout as output writer")
		return os.Stdout, nil
	}

	logger.Debug("Creating output file: %s", outputPath)
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create directory: %s, error: %v", dir, err)
		return nil, err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		logger.Error("Failed to create file: %s, error: %v", outputPath, err)
		return nil, err
	}

	return file, nil
}

// CloseOutputWriter schließt den Writer, wenn es sich um eine Datei handelt
func CloseOutputWriter(writer io.Writer, logger Logger) {
	if file, ok := writer.(*os.File); ok && file != os.Stdout {
		logger.Debug("Closing output file")
		if err := file.Close(); err != nil {
			logger.Error("Failed to close file: %v", err)
		}
	}
}

// ResolveConfigPath löst den Pfad zur Konfigurationsdatei auf
func ResolveConfigPath(configPath string) string {
	if filepath.IsAbs(configPath) {
		return configPath
	}

	// Relativen Pfad auflösen
	wd, err := os.Getwd()
	if err != nil {
		return configPath // Im Fehlerfall den ursprünglichen Pfad zurückgeben
	}

	return filepath.Join(wd, configPath)
}
