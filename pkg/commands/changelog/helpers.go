package changelog

import (
	"io"
	"os"
	"path/filepath"

	"clikd/pkg/utils"
)

// createOutputWriter erstellt einen Writer für die Ausgabe
func createOutputWriter(outputPath string) (io.Writer, error) {
	logger := utils.NewLogger("info", true)

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

// closeOutputWriter schließt den Writer, wenn es sich um eine Datei handelt
func closeOutputWriter(writer io.Writer) {
	logger := utils.NewLogger("info", true)

	if file, ok := writer.(*os.File); ok && file != os.Stdout {
		logger.Debug("Closing output file")
		if err := file.Close(); err != nil {
			logger.Error("Failed to close file: %v", err)
		}
	}
}

// resolveConfigPath löst den Pfad zur Konfigurationsdatei auf
func resolveConfigPath(configPath string) string {
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
