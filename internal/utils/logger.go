package utils

import (
	"io"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

// Logger ist eine Schnittstelle für einheitliches Logging in allen Modulen
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	WithFields(fields map[string]interface{}) Logger
	GetLevel() string
	SetOutput(w io.Writer)
}

// CharmLogger implementiert die Logger-Schnittstelle mit dem Charm Log-Logger
type CharmLogger struct {
	logger *log.Logger
}

// NewLogger erstellt einen neuen Logger mit dem angegebenen Log-Level und Farboptionen
func NewLogger(level string, useColors bool) Logger {
	logger := log.New(os.Stdout)

	// Setze das Log-Level
	switch level {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	case "fatal":
		logger.SetLevel(log.FatalLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Konfiguriere Ausgabeoptionen
	logger.SetReportTimestamp(true)
	logger.SetTimeFormat(time.DateTime)

	// Deaktiviere Farben, wenn gewünscht
	if !useColors {
		// Deaktiviere farbige Ausgabe mit dem NoColor-Profil
		logger.SetColorProfile(termenv.Ascii)
	}

	return &CharmLogger{
		logger: logger,
	}
}

// SetOutput setzt den Ausgabestream des Loggers
func (l *CharmLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// Debug loggt eine Debug-Nachricht
func (l *CharmLogger) Debug(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Info loggt eine Info-Nachricht
func (l *CharmLogger) Info(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warn loggt eine Warnungs-Nachricht
func (l *CharmLogger) Warn(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Error loggt eine Fehler-Nachricht
func (l *CharmLogger) Error(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatal loggt eine fatale Fehler-Nachricht und beendet das Programm
func (l *CharmLogger) Fatal(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// WithFields gibt einen neuen Logger mit zusätzlichen Feldern zurück
func (l *CharmLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := l.logger.With()
	for k, v := range fields {
		newLogger = newLogger.With(k, v)
	}

	return &CharmLogger{
		logger: newLogger,
	}
}

// Helper markiert die aufrufende Funktion als Helper-Funktion
func (l *CharmLogger) Helper() {
	l.logger.Helper()
}

// GetLevel gibt das aktuelle Log-Level zurück
func (l *CharmLogger) GetLevel() string {
	switch l.logger.GetLevel() {
	case log.DebugLevel:
		return "debug"
	case log.InfoLevel:
		return "info"
	case log.WarnLevel:
		return "warn"
	case log.ErrorLevel:
		return "error"
	case log.FatalLevel:
		return "fatal"
	default:
		return "info"
	}
}

// DefaultLogger ist der Standard-Logger für die Anwendung
var DefaultLogger = NewLogger("info", true)

// Log ist ein Hilfsmethode für Kompatibilität mit Bibliotheken, die eine Log-Methode erwarten
func (l *CharmLogger) Log(format string, args ...interface{}) {
	l.Info(format, args...)
}
