package bubble

import (
	"fmt"
	"strings"

	"clikd/internal/ui/styles"
)

// RenderUpdateNotification renders a notification about a newer version
func RenderUpdateNotification(currentVersion, latestVersion, releaseURL string, width int) string {
	// Determine real content width inside borders/padding
	contentWidth := width - 4 // Account for border and padding

	// Don't render if width is too narrow
	if contentWidth < 40 {
		return ""
	}

	// Create the update sparkle indicator
	updateIndicator := styles.UpdateIndicator.Render("✨ Update available!")

	// Create the version line with current and new version
	versionLine := fmt.Sprintf(
		"Current: %s → New: %s",
		currentVersion,
		styles.UpdateVersion.Render(latestVersion),
	)

	// Create the install command line
	installCmd := fmt.Sprintf(
		"Run: %s",
		styles.UpdateCommand.Render("brew upgrade clikd"),
	)

	// Create the URL line
	urlLine := fmt.Sprintf(
		"Details: %s",
		styles.UpdateURL.Render(releaseURL),
	)

	// Join all the parts
	content := strings.Join([]string{
		updateIndicator,
		versionLine,
		installCmd,
		urlLine,
	}, "\n")

	// Return the full boxed notification
	return styles.UpdateNotification.Width(contentWidth).Render(content)
}
