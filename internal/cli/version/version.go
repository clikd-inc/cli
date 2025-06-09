package version

// Version enthält die aktuelle Version der CLI
// Wird zur Build-Zeit durch den Linker gesetzt
var Version = "dev"

// GetVersion gibt die aktuelle Version zurück
func GetVersion() string {
	return Version
}
