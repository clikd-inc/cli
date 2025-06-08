package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCmd creates a new version command
func NewVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number of clikd",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("clikd version %s\n", version)
		},
	}
}
