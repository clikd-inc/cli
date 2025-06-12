package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOutputWriter(t *testing.T) {
	// Create a mock logger for testing
	logger := &MockLogger{}

	t.Run("Empty output path returns stdout", func(t *testing.T) {
		writer, err := CreateOutputWriter("", logger)
		assert.NoError(t, err)
		assert.Equal(t, os.Stdout, writer)
	})

	t.Run("Create file with valid path", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "test_changelog.md")

		writer, err := CreateOutputWriter(outputPath, logger)
		assert.NoError(t, err)
		assert.NotNil(t, writer)

		// Verify file was created
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)

		// Clean up
		if file, ok := writer.(*os.File); ok {
			file.Close()
		}
	})

	t.Run("Create file with nested directory", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "nested", "dir", "changelog.md")

		writer, err := CreateOutputWriter(outputPath, logger)
		assert.NoError(t, err)
		assert.NotNil(t, writer)

		// Verify directory and file were created
		_, err = os.Stat(filepath.Dir(outputPath))
		assert.NoError(t, err)
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)

		// Clean up
		if file, ok := writer.(*os.File); ok {
			file.Close()
		}
	})

	t.Run("Error creating directory with invalid path", func(t *testing.T) {
		// Try to create a file in a path that can't be created (e.g., under a file instead of directory)
		tempDir := t.TempDir()

		// Create a file first
		existingFile := filepath.Join(tempDir, "existing_file.txt")
		err := os.WriteFile(existingFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Try to create a directory under this file (should fail)
		invalidPath := filepath.Join(existingFile, "subdir", "changelog.md")

		writer, err := CreateOutputWriter(invalidPath, logger)
		assert.Error(t, err)
		assert.Nil(t, writer)
	})
}

func TestCloseOutputWriter(t *testing.T) {
	logger := &MockLogger{}

	t.Run("Close file writer", func(t *testing.T) {
		// Create temporary file
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "test.txt")

		file, err := os.Create(tempFile)
		require.NoError(t, err)

		// Close the writer
		CloseOutputWriter(file, logger)

		// Verify file is closed by trying to write (should fail)
		_, err = file.Write([]byte("test"))
		assert.Error(t, err)
	})

	t.Run("Don't close stdout", func(t *testing.T) {
		// This should not close stdout (no error should occur)
		CloseOutputWriter(os.Stdout, logger)

		// Stdout should still be writable
		_, err := os.Stdout.Write([]byte(""))
		assert.NoError(t, err)
	})

	t.Run("Don't close non-file writer", func(t *testing.T) {
		// Create a string writer (not a file)
		var buf strings.Builder

		// This should not panic or error
		CloseOutputWriter(&buf, logger)

		// Buffer should still be writable
		_, err := buf.Write([]byte("test"))
		assert.NoError(t, err)
	})
}

func TestResolveConfigPath(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	t.Run("Absolute path returns unchanged", func(t *testing.T) {
		absolutePath := "/absolute/path/to/config.yml"
		result := ResolveConfigPath(absolutePath)
		assert.Equal(t, absolutePath, result)
	})

	t.Run("Relative path gets resolved", func(t *testing.T) {
		// Create temporary directory and change to it
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		relativePath := "config/changelog.yml"
		result := ResolveConfigPath(relativePath)

		expected := filepath.Join(tempDir, relativePath)
		// Normalize paths to handle symlinks (e.g., /var -> /private/var on macOS)
		resultAbs, _ := filepath.EvalSymlinks(result)
		expectedAbs, _ := filepath.EvalSymlinks(expected)
		assert.Equal(t, expectedAbs, resultAbs)
	})

	t.Run("Current directory path", func(t *testing.T) {
		// Create temporary directory and change to it
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		relativePath := "./config.yml"
		result := ResolveConfigPath(relativePath)

		expected := filepath.Join(tempDir, relativePath)
		// Normalize paths to handle symlinks (e.g., /var -> /private/var on macOS)
		resultAbs, _ := filepath.EvalSymlinks(result)
		expectedAbs, _ := filepath.EvalSymlinks(expected)
		assert.Equal(t, expectedAbs, resultAbs)
	})

	t.Run("Parent directory path", func(t *testing.T) {
		// Create temporary directory and change to it
		tempDir := t.TempDir()
		subDir := filepath.Join(tempDir, "subdir")
		os.MkdirAll(subDir, 0755)
		os.Chdir(subDir)

		relativePath := "../config.yml"
		result := ResolveConfigPath(relativePath)

		expected := filepath.Join(subDir, relativePath)
		// Normalize paths to handle symlinks (e.g., /var -> /private/var on macOS)
		resultAbs, _ := filepath.EvalSymlinks(result)
		expectedAbs, _ := filepath.EvalSymlinks(expected)
		assert.Equal(t, expectedAbs, resultAbs)
	})

	t.Run("Simple filename", func(t *testing.T) {
		// Create temporary directory and change to it
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		relativePath := "config.yml"
		result := ResolveConfigPath(relativePath)

		expected := filepath.Join(tempDir, relativePath)
		// Normalize paths to handle symlinks (e.g., /var -> /private/var on macOS)
		resultAbs, _ := filepath.EvalSymlinks(result)
		expectedAbs, _ := filepath.EvalSymlinks(expected)
		assert.Equal(t, expectedAbs, resultAbs)
	})
}

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	DebugMessages []string
	ErrorMessages []string
}

func (m *MockLogger) Debug(msg interface{}, keyvals ...interface{}) {
	if s, ok := msg.(string); ok {
		m.DebugMessages = append(m.DebugMessages, s)
	}
}

func (m *MockLogger) Info(msg interface{}, keyvals ...interface{}) {}

func (m *MockLogger) Warn(msg interface{}, keyvals ...interface{}) {}

func (m *MockLogger) Error(msg interface{}, keyvals ...interface{}) {
	if s, ok := msg.(string); ok {
		m.ErrorMessages = append(m.ErrorMessages, s)
	}
}

func (m *MockLogger) Fatal(msg interface{}, keyvals ...interface{}) {}

func (m *MockLogger) Debugf(format string, args ...interface{}) {}

func (m *MockLogger) Infof(format string, args ...interface{}) {}

func (m *MockLogger) Warnf(format string, args ...interface{}) {}

func (m *MockLogger) Errorf(format string, args ...interface{}) {}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {}

func (m *MockLogger) With(keyvals ...interface{}) Logger {
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) Logger {
	return m
}

func (m *MockLogger) GetLevel() string {
	return "debug"
}

func (m *MockLogger) SetOutput(w io.Writer) {}

func (m *MockLogger) Helper() {}
