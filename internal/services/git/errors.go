package git

import (
	"errors"
	"fmt"
)

// Vordefinierte Fehler
var (
	// ErrNotFoundTag wird zurückgegeben, wenn ein Tag nicht gefunden wird
	ErrNotFoundTag = errors.New("git: tag not found")

	// ErrInvalidTagFormat wird zurückgegeben, wenn ein ungültiges Tag-Format verwendet wird
	ErrInvalidTagFormat = errors.New("git: invalid tag format")
)

// TagNotFoundError repräsentiert einen Fehler, wenn ein bestimmtes Tag nicht gefunden werden kann
type TagNotFoundError struct {
	TagName string
}

func (e *TagNotFoundError) Error() string {
	return fmt.Sprintf("git: tag %q not found", e.TagName)
}

// NotImplementedError repräsentiert einen Fehler für noch nicht implementierte Funktionen
type NotImplementedError struct {
	Feature string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("git: %s not implemented", e.Feature)
}

// Fehler mit Format-String und Argumenten erzeugen
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
