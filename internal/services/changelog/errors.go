package changelog

import (
	"errors"
)

// Errors entsprechend dem original-package
var (
	// ErrNotSpecifiedCLIContext wird zurückgegeben, wenn der CLI-Kontext nicht angegeben ist
	ErrNotSpecifiedCLIContext = errors.New("cli context is not specified")

	// ErrNotFoundWorkingDirectory wird zurückgegeben, wenn das Arbeitsverzeichnis nicht gefunden werden kann
	ErrNotFoundWorkingDirectory = errors.New("failed to find working directory")

	// ErrNotFoundRepositoryFromWorkingDir wird zurückgegeben, wenn das Repository aus dem Arbeitsverzeichnis nicht gefunden werden kann
	ErrNotFoundRepositoryFromWorkingDir = errors.New("failed to find repository from working directory")

	// ErrNotSpecifiedOutputPath wird zurückgegeben, wenn der Ausgabepfad nicht angegeben ist
	ErrNotSpecifiedOutputPath = errors.New("output path must be specified")

	// ErrNotFoundOutputPath wird zurückgegeben, wenn der Ausgabepfad nicht gefunden werden kann
	ErrNotFoundOutputPath = errors.New("output path is not found")

	// ErrFailedOutputPath wird zurückgegeben, wenn der Ausgabepfad nicht erstellt werden kann
	ErrFailedOutputPath = errors.New("failed to create output path")

	// ErrFailedPushBacklog wird zurückgegeben, wenn das Backlog nicht aktualisiert werden kann
	ErrFailedPushBacklog = errors.New("failed to push backlog")

	// ErrNotSpecifiedCONFIGEnv wird zurückgegeben, wenn die CONFIG-Umgebungsvariable nicht angegeben ist
	ErrNotSpecifiedCONFIGEnv = errors.New("$CONFIG is not specified")

	// ErrNotSpecifiedJiraUsername wird zurückgegeben, wenn der Jira-Benutzername nicht angegeben ist
	ErrNotSpecifiedJiraUsername = errors.New("jira username not specified")

	// ErrNotSpecifiedJiraToken wird zurückgegeben, wenn das Jira-Token nicht angegeben ist
	ErrNotSpecifiedJiraToken = errors.New("jira token not specified")

	// ErrNotSpecifiedJiraURL wird zurückgegeben, wenn die Jira-URL nicht angegeben ist
	ErrNotSpecifiedJiraURL = errors.New("jira url not specified")

	// ErrInvalidJiraURL wird zurückgegeben, wenn die Jira-URL ungültig ist
	ErrInvalidJiraURL = errors.New("invalid jira url")
)
