package changelog

import (
	"clikd/pkg/utils"
)

// LoggerWrapper ist ein Wrapper für den clikd-Logger
// der das Logger-Interface des Changelog-Projekts implementiert
type LoggerWrapper struct {
	logger *utils.Logger
}

// NewLoggerWrapper erstellt einen neuen Logger-Wrapper
func NewLoggerWrapper() *LoggerWrapper {
	// Verwende den NewLogger-Konstruktor mit Standard-Loglevel "info" und farbigem Output
	return &LoggerWrapper{
		logger: utils.NewLogger("info", true),
	}
}

// Log schreibt eine Log-Nachricht
func (w *LoggerWrapper) Log(format string, args ...interface{}) {
	w.logger.Info(format, args...)
}

// Error schreibt eine Fehlermeldung
func (w *LoggerWrapper) Error(format string, args ...interface{}) {
	w.logger.Error(format, args...)
}

// Warn schreibt eine Warnungsmeldung
func (w *LoggerWrapper) Warn(format string, args ...interface{}) {
	w.logger.Warn(format, args...)
}

// Debug schreibt eine Debug-Nachricht
func (w *LoggerWrapper) Debug(format string, args ...interface{}) {
	w.logger.Debug(format, args...)
}

// Info schreibt eine Info-Nachricht
func (w *LoggerWrapper) Info(format string, args ...interface{}) {
	w.logger.Info(format, args...)
}
