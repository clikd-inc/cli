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
	// Standard logging methods with structured key-value pairs
	Debug(msg interface{}, keyvals ...interface{})
	Info(msg interface{}, keyvals ...interface{})
	Warn(msg interface{}, keyvals ...interface{})
	Error(msg interface{}, keyvals ...interface{})
	Fatal(msg interface{}, keyvals ...interface{})

	// Formatted logging methods
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	// Structured logging
	With(keyvals ...interface{}) Logger
	WithFields(fields map[string]interface{}) Logger

	// Configuration
	GetLevel() string
	SetOutput(w io.Writer)
	Helper()
}

// LoggerOptions contains configuration options for creating a new logger
type LoggerOptions struct {
	Level           string
	UseColors       bool
	ReportCaller    bool
	ReportTimestamp bool
	TimeFormat      string
	Prefix          string
	Formatter       string // "text", "json", "logfmt"
}

// CharmLogger implementiert die Logger-Schnittstelle mit dem Charm Log-Logger
type CharmLogger struct {
	logger *log.Logger
}

// NewLogger erstellt einen neuen Logger mit dem angegebenen Log-Level und Farboptionen
func NewLogger(level string, useColors bool) Logger {
	return NewLoggerWithOptions(LoggerOptions{
		Level:           level,
		UseColors:       useColors,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
		Formatter:       "text",
	})
}

// NewLoggerWithOptions erstellt einen neuen Logger mit erweiterten Optionen
func NewLoggerWithOptions(opts LoggerOptions) Logger {
	logger := log.New(os.Stdout)

	// Setze das Log-Level
	switch opts.Level {
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
	logger.SetReportTimestamp(opts.ReportTimestamp)
	if opts.TimeFormat != "" {
		logger.SetTimeFormat(opts.TimeFormat)
	}

	// Caller-Informationen
	logger.SetReportCaller(opts.ReportCaller)

	// Prefix setzen
	if opts.Prefix != "" {
		logger.SetPrefix(opts.Prefix)
	}

	// Formatter setzen
	switch opts.Formatter {
	case "json":
		logger.SetFormatter(log.JSONFormatter)
	case "logfmt":
		logger.SetFormatter(log.LogfmtFormatter)
	default:
		logger.SetFormatter(log.TextFormatter)
	}

	// Deaktiviere Farben, wenn gewünscht
	if !opts.UseColors {
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

// Debug loggt eine Debug-Nachricht mit strukturierten Key-Value-Paaren
func (l *CharmLogger) Debug(msg interface{}, keyvals ...interface{}) {
	l.logger.Debug(msg, keyvals...)
}

// Info loggt eine Info-Nachricht mit strukturierten Key-Value-Paaren
func (l *CharmLogger) Info(msg interface{}, keyvals ...interface{}) {
	l.logger.Info(msg, keyvals...)
}

// Warn loggt eine Warnungs-Nachricht mit strukturierten Key-Value-Paaren
func (l *CharmLogger) Warn(msg interface{}, keyvals ...interface{}) {
	l.logger.Warn(msg, keyvals...)
}

// Error loggt eine Fehler-Nachricht mit strukturierten Key-Value-Paaren
func (l *CharmLogger) Error(msg interface{}, keyvals ...interface{}) {
	l.logger.Error(msg, keyvals...)
}

// Fatal loggt eine fatale Fehler-Nachricht und beendet das Programm
func (l *CharmLogger) Fatal(msg interface{}, keyvals ...interface{}) {
	l.logger.Fatal(msg, keyvals...)
}

// Debugf loggt eine formatierte Debug-Nachricht
func (l *CharmLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Infof loggt eine formatierte Info-Nachricht
func (l *CharmLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf loggt eine formatierte Warnungs-Nachricht
func (l *CharmLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf loggt eine formatierte Fehler-Nachricht
func (l *CharmLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatalf loggt eine formatierte fatale Fehler-Nachricht und beendet das Programm
func (l *CharmLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// With gibt einen neuen Logger mit zusätzlichen Key-Value-Paaren zurück
func (l *CharmLogger) With(keyvals ...interface{}) Logger {
	return &CharmLogger{
		logger: l.logger.With(keyvals...),
	}
}

// WithFields gibt einen neuen Logger mit zusätzlichen Feldern zurück
func (l *CharmLogger) WithFields(fields map[string]interface{}) Logger {
	keyvals := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		keyvals = append(keyvals, k, v)
	}

	return &CharmLogger{
		logger: l.logger.With(keyvals...),
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
