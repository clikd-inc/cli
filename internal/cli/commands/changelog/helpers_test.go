package changelog

import (
	"clikd/internal/utils"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockLogger ist eine einfache Implementierung der utils.Logger-Schnittstelle für Tests
type MockLogger struct{}

func (m *MockLogger) Debug(format string, args ...interface{})              {}
func (m *MockLogger) Info(format string, args ...interface{})               {}
func (m *MockLogger) Warn(format string, args ...interface{})               {}
func (m *MockLogger) Error(format string, args ...interface{})              {}
func (m *MockLogger) Fatal(format string, args ...interface{})              {}
func (m *MockLogger) WithFields(fields map[string]interface{}) utils.Logger { return m }
func (m *MockLogger) GetLevel() string                                      { return "info" }
func (m *MockLogger) SetOutput(w io.Writer)                                 {}

func TestCreateOutputWriter(t *testing.T) {
	logger := &MockLogger{}

	// Test mit leerem Pfad (sollte stdout zurückgeben)
	writer, err := utils.CreateOutputWriter("", logger)
	assert.NoError(t, err)
	assert.Equal(t, os.Stdout, writer)

	// Test mit echtem Pfad
	tempDir, err := os.MkdirTemp("", "changelog-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "output.md")
	writer, err = utils.CreateOutputWriter(outputPath, logger)
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	// Prüfen, ob es sich um eine Datei handelt
	_, ok := writer.(*os.File)
	assert.True(t, ok)

	// Datei schließen
	utils.CloseOutputWriter(writer, logger)
}

func TestCloseOutputWriter(t *testing.T) {
	logger := &MockLogger{}

	// Temporäre Datei erstellen
	tempFile, err := os.CreateTemp("", "changelog-test-*.md")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Datei schließen
	_ = tempFile.Close() // Schließen wir die Datei vor dem Test, damit CloseOutputWriter keinen Fehler wirft

	// CloseOutputWriter sollte keine Fehler verursachen, auch wenn die Datei bereits geschlossen ist
	utils.CloseOutputWriter(tempFile, logger)

	// Stdout schließen sollte keinen Fehler verursachen (wird ignoriert)
	utils.CloseOutputWriter(os.Stdout, logger)
}

func TestResolveConfigPath(t *testing.T) {
	// Absoluter Pfad sollte unverändert zurückgegeben werden
	absPath := "/absolute/path/to/config.yml"
	result := utils.ResolveConfigPath(absPath)
	assert.Equal(t, absPath, result)

	// Relativer Pfad sollte in absoluten Pfad umgewandelt werden
	relPath := "relative/path/config.yml"
	result = utils.ResolveConfigPath(relPath)

	// Das aktuelle Arbeitsverzeichnis plus relativer Pfad
	wd, err := os.Getwd()
	assert.NoError(t, err)
	expected := filepath.Join(wd, relPath)

	assert.Equal(t, expected, result)
}
