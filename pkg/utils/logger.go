package utils

import (
	"fmt"
	"os"
	"time"
)

// LogLevel represents the severity level of log messages
type LogLevel int

const (
	// DEBUG level for detailed troubleshooting information
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for potentially harmful situations
	WARN
	// ERROR level for error events that might still allow the application to continue
	ERROR
	// FATAL level for severe error events that will lead the application to abort
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger provides basic logging functionality
type Logger struct {
	level     LogLevel
	useColors bool
}

// NewLogger creates a new logger with the specified log level
func NewLogger(level string, useColors bool) *Logger {
	var logLevel LogLevel
	switch level {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warn":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	case "fatal":
		logLevel = FATAL
	default:
		logLevel = INFO
	}

	return &Logger{
		level:     logLevel,
		useColors: useColors,
	}
}

// formatMessage formats a log message with timestamp and level
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]

	if l.useColors {
		var colorCode string
		switch level {
		case DEBUG:
			colorCode = "\033[36m" // Cyan
		case INFO:
			colorCode = "\033[32m" // Green
		case WARN:
			colorCode = "\033[33m" // Yellow
		case ERROR:
			colorCode = "\033[31m" // Red
		case FATAL:
			colorCode = "\033[35m" // Magenta
		}
		resetCode := "\033[0m"
		return fmt.Sprintf("%s [%s%s%s] %s", timestamp, colorCode, levelName, resetCode, message)
	}

	return fmt.Sprintf("%s [%s] %s", timestamp, levelName, message)
}

// log logs a message at the specified level if it's higher than or equal to the logger's level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	formattedMessage := l.formatMessage(level, message)

	if level >= ERROR {
		fmt.Fprintln(os.Stderr, formattedMessage)
	} else {
		fmt.Fprintln(os.Stdout, formattedMessage)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug level message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info level message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning level message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error level message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal level message and exits the application
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}
