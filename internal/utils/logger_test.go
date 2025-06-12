package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewLogger tests logger creation
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		withColor bool
	}{
		{
			name:      "Debug logger with color",
			level:     "debug",
			withColor: true,
		},
		{
			name:      "Info logger without color",
			level:     "info",
			withColor: false,
		},
		{
			name:      "Error logger with color",
			level:     "error",
			withColor: true,
		},
		{
			name:      "Invalid level defaults to info",
			level:     "invalid",
			withColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level, tt.withColor)

			assert.NotNil(t, logger)
			assert.Implements(t, (*Logger)(nil), logger)

			// Verify it's a CharmLogger
			charmLogger, ok := logger.(*CharmLogger)
			assert.True(t, ok)
			assert.NotNil(t, charmLogger.logger)
		})
	}
}

// TestLogger_StructuredLogging tests structured logging with key-value pairs
func TestLogger_StructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", false) // No color for easier testing

	// Redirect output to buffer for testing
	charmLogger := logger.(*CharmLogger)
	charmLogger.logger.SetOutput(&buf)

	tests := []struct {
		name     string
		logFunc  func()
		expected []string // Strings that should appear in output
	}{
		{
			name: "Debug with key-value pairs",
			logFunc: func() {
				logger.Debug("Processing request", "userID", 12345, "action", "login")
			},
			expected: []string{"Processing request", "userID", "12345", "action", "login"},
		},
		{
			name: "Info with mixed types",
			logFunc: func() {
				logger.Info("Database connection", "host", "localhost", "port", 5432, "ssl", true)
			},
			expected: []string{"Database connection", "host", "localhost", "port", "5432", "ssl", "true"},
		},
		{
			name: "Error with error object",
			logFunc: func() {
				logger.Error("Failed to connect", "error", "connection timeout", "retries", 3)
			},
			expected: []string{"Failed to connect", "error", "connection timeout", "retries", "3"},
		},
		{
			name: "No key-value pairs",
			logFunc: func() {
				logger.Info("Simple message")
			},
			expected: []string{"Simple message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestLogger_FormattedLogging tests formatted logging methods
func TestLogger_FormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", false)

	charmLogger := logger.(*CharmLogger)
	charmLogger.logger.SetOutput(&buf)

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Debugf with formatting",
			logFunc: func() {
				logger.Debugf("Processing user %d with action %s", 12345, "login")
			},
			expected: "Processing user 12345 with action login",
		},
		{
			name: "Infof with formatting",
			logFunc: func() {
				logger.Infof("Connected to %s:%d", "localhost", 5432)
			},
			expected: "Connected to localhost:5432",
		},
		{
			name: "Errorf with formatting",
			logFunc: func() {
				logger.Errorf("Failed after %d retries: %s", 3, "timeout")
			},
			expected: "Failed after 3 retries: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}

// TestLogger_WithFields tests sub-logger creation with fields
func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", false)

	charmLogger := logger.(*CharmLogger)
	charmLogger.logger.SetOutput(&buf)

	// Create sub-logger with fields
	subLogger := logger.WithFields(map[string]interface{}{
		"module":  "auth",
		"version": "1.0.0",
	})

	assert.NotNil(t, subLogger)
	assert.Implements(t, (*Logger)(nil), subLogger)

	// Test that sub-logger includes the fields
	buf.Reset()
	subLogger.Info("User authenticated", "userID", 123)

	output := buf.String()
	assert.Contains(t, output, "User authenticated")
	assert.Contains(t, output, "module")
	assert.Contains(t, output, "auth")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "userID")
	assert.Contains(t, output, "123")
}

// TestLogger_With tests sub-logger creation with single key-value pair
func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", false)

	charmLogger := logger.(*CharmLogger)
	charmLogger.logger.SetOutput(&buf)

	// Create sub-logger with single field
	subLogger := logger.With("requestID", "req-123")

	assert.NotNil(t, subLogger)
	assert.Implements(t, (*Logger)(nil), subLogger)

	// Test that sub-logger includes the field
	buf.Reset()
	subLogger.Info("Request processed")

	output := buf.String()
	assert.Contains(t, output, "Request processed")
	assert.Contains(t, output, "requestID")
	assert.Contains(t, output, "req-123")
}

// TestLogger_Helper tests the helper method
func TestLogger_Helper(t *testing.T) {
	logger := NewLogger("debug", false)

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Helper()
	})
}

// TestLogger_SetOutput tests output redirection
func TestLogger_SetOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("info", false)

	// Redirect output
	logger.SetOutput(&buf)

	// Log something
	logger.Info("Test message")

	// Verify output was captured
	output := buf.String()
	assert.Contains(t, output, "Test message")
}

// TestLogger_SetPrefix tests prefix setting via WithFields
func TestLogger_SetPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("info", false)
	logger.SetOutput(&buf)

	// Create logger with prefix-like field
	prefixedLogger := logger.WithFields(map[string]interface{}{"prefix": "CLIKD"})

	// Log something
	prefixedLogger.Info("Test message")

	// Verify prefix-like field appears
	output := buf.String()
	assert.Contains(t, output, "CLIKD")
	assert.Contains(t, output, "Test message")
}

// TestLogger_LogLevels tests different log levels
func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFunc   func(Logger)
		shouldLog bool
	}{
		{
			name:     "Debug level allows debug",
			logLevel: "debug",
			logFunc: func(l Logger) {
				l.Debug("debug message")
			},
			shouldLog: true,
		},
		{
			name:     "Info level blocks debug",
			logLevel: "info",
			logFunc: func(l Logger) {
				l.Debug("debug message")
			},
			shouldLog: false,
		},
		{
			name:     "Info level allows info",
			logLevel: "info",
			logFunc: func(l Logger) {
				l.Info("info message")
			},
			shouldLog: true,
		},
		{
			name:     "Error level blocks info",
			logLevel: "error",
			logFunc: func(l Logger) {
				l.Info("info message")
			},
			shouldLog: false,
		},
		{
			name:     "Error level allows error",
			logLevel: "error",
			logFunc: func(l Logger) {
				l.Error("error message")
			},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.logLevel, false)
			logger.SetOutput(&buf)

			tt.logFunc(logger)

			output := buf.String()
			if tt.shouldLog {
				assert.NotEmpty(t, output, "Should have logged something")
			} else {
				assert.Empty(t, output, "Should not have logged anything")
			}
		})
	}
}

// TestLogger_EdgeCases tests edge cases and error conditions
func TestLogger_EdgeCases(t *testing.T) {
	t.Run("Nil values in key-value pairs", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("debug", false)
		logger.SetOutput(&buf)

		// Should handle nil values gracefully
		assert.NotPanics(t, func() {
			logger.Debug("Test message", "key", nil, "other", "value")
		})

		output := buf.String()
		assert.Contains(t, output, "Test message")
	})

	t.Run("Odd number of key-value arguments", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("debug", false)
		logger.SetOutput(&buf)

		// Should handle odd number of arguments gracefully
		assert.NotPanics(t, func() {
			logger.Debug("Test message", "key1", "value1", "key2")
		})

		output := buf.String()
		assert.Contains(t, output, "Test message")
	})

	t.Run("Empty message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("debug", false)
		logger.SetOutput(&buf)

		// Should handle empty message
		assert.NotPanics(t, func() {
			logger.Debug("", "key", "value")
		})
	})

	t.Run("Very long message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("debug", false)
		logger.SetOutput(&buf)

		longMessage := strings.Repeat("a", 10000)

		// Should handle very long messages
		assert.NotPanics(t, func() {
			logger.Debug(longMessage)
		})

		output := buf.String()
		assert.Contains(t, output, longMessage)
	})
}

// TestDefaultLogger tests the default logger
func TestDefaultLogger(t *testing.T) {
	assert.NotNil(t, DefaultLogger)
	assert.Implements(t, (*Logger)(nil), DefaultLogger)

	// Should be able to use default logger
	assert.NotPanics(t, func() {
		DefaultLogger.Info("Test message")
	})
}

// TestLogger_ChainedWithFields tests chaining WithFields calls
func TestLogger_ChainedWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", false)
	logger.SetOutput(&buf)

	// Chain multiple WithFields calls
	subLogger := logger.
		WithFields(map[string]interface{}{"module": "auth"}).
		WithFields(map[string]interface{}{"version": "1.0.0"}).
		With("requestID", "req-123")

	buf.Reset()
	subLogger.Info("Chained logger test")

	output := buf.String()
	assert.Contains(t, output, "Chained logger test")
	assert.Contains(t, output, "module")
	assert.Contains(t, output, "auth")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "requestID")
	assert.Contains(t, output, "req-123")
}

// TestLogger_ConcurrentAccess tests concurrent access to logger
func TestLogger_ConcurrentAccess(t *testing.T) {
	logger := NewLogger("debug", false)

	// Should not panic with concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 100; j++ {
				logger.Info("Concurrent message", "goroutine", id, "iteration", j)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passed if no panic occurred
	assert.True(t, true)
}
