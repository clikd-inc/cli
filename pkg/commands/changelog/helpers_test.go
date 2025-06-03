package changelog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOutputWriter(t *testing.T) {
	// Test mit leerem Pfad (sollte stdout zurückgeben)
	writer, err := createOutputWriter("")
	assert.NoError(t, err)
	assert.Equal(t, os.Stdout, writer)

	// Test mit echtem Pfad
	tempDir, err := os.MkdirTemp("", "changelog-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "output.md")
	writer, err = createOutputWriter(outputPath)
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	// Prüfen, ob es sich um eine Datei handelt
	_, ok := writer.(*os.File)
	assert.True(t, ok)

	// Datei schließen
	closeOutputWriter(writer)
}

func TestCloseOutputWriter(t *testing.T) {
	// Temporäre Datei erstellen
	tempFile, err := os.CreateTemp("", "changelog-test-*.md")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Datei schließen
	_ = tempFile.Close() // Schließen wir die Datei vor dem Test, damit closeOutputWriter keinen Fehler wirft

	// closeOutputWriter sollte keine Fehler verursachen, auch wenn die Datei bereits geschlossen ist
	closeOutputWriter(tempFile)

	// Stdout schließen sollte keinen Fehler verursachen (wird ignoriert)
	closeOutputWriter(os.Stdout)
}

func TestResolveConfigPath(t *testing.T) {
	// Absoluter Pfad sollte unverändert zurückgegeben werden
	absPath := "/absolute/path/to/config.yml"
	result := resolveConfigPath(absPath)
	assert.Equal(t, absPath, result)

	// Relativer Pfad sollte in absoluten Pfad umgewandelt werden
	relPath := "relative/path/config.yml"
	result = resolveConfigPath(relPath)

	// Das aktuelle Arbeitsverzeichnis plus relativer Pfad
	wd, err := os.Getwd()
	assert.NoError(t, err)
	expected := filepath.Join(wd, relPath)

	assert.Equal(t, expected, result)
}
