package version

import (
	"context"
	"fmt"
	"strings"
	"time"

	"clikd/internal/services/update"
	"clikd/internal/ui/bubble"
	"clikd/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates a new version command
func NewVersionCmd(version string) *cobra.Command {
	var checkForUpdates bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number of clikd",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a stylish version display
			versionBox := renderVersionInfo(version)
			cmd.Println(versionBox)

			// Check for updates if flag is set
			if checkForUpdates {
				// Create a context with timeout for update check
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				cmd.Println(styles.InfoStyle.Render("\nChecking for updates..."))

				// Check for updates
				hasUpdate, latestVersion, releaseURL, err := update.CheckForUpdates(ctx, version)
				if err != nil {
					return fmt.Errorf("failed to check for updates: %w", err)
				}

				if hasUpdate {
					// Get terminal width (default to 80 if can't determine)
					width := 80

					// Render update notification
					notification := bubble.RenderUpdateNotification(version, latestVersion, releaseURL, width)
					cmd.Println(notification)
				} else {
					checkmark := styles.SuccessStyle.Render("✓")
					message := styles.SuccessStyle.Render("You're using the latest version!")
					cmd.Printf("\n%s %s\n", checkmark, message)
				}
			}

			return nil
		},
	}

	// Add flag to check for updates (using -u to avoid conflict with -c for --config)
	cmd.Flags().BoolVarP(&checkForUpdates, "check", "u", false, "Check for updates")

	return cmd
}

// renderVersionInfo creates a stylish box with the version information
func renderVersionInfo(version string) string {
	// Create the version header
	header := styles.H1.Render("clikd CLI")

	// Create the version line
	versionLine := fmt.Sprintf("Version: %s",
		styles.SuccessStyle.Render(version))

	// Create a line with build info (can be extended later)
	buildLine := fmt.Sprintf("Build: %s",
		styles.InfoStyle.Render(time.Now().Format("2006-01-02")))

	// Copyright and additional info
	copyrightLine := styles.Subtle.Render("© 2023-2024 CLIKD Inc.")

	// Join all parts
	content := strings.Join([]string{
		header,
		"",
		versionLine,
		buildLine,
		"",
		copyrightLine,
	}, "\n")

	// Create a stylish box
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 3).
		Width(50)

	return boxStyle.Render(content)
}
