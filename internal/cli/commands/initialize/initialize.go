package initialize

import (
	"clikd/internal/ui/cmd/initialize"

	"github.com/spf13/cobra"
)

// NewInitCmd creates a new init command
func NewInitCmd() *cobra.Command {
	var global bool
	var force bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the clikd configuration",
		Long:  `Initialize the clikd configuration either globally or for the current project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initialize.RunInitializationUI(global, force, yes)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Create a global configuration (default is local)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite of existing configuration")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Accept all defaults without prompting")

	// Add subcommands
	cmd.AddCommand(newConfigCmd())

	return cmd
}

// newConfigCmd creates a new config subcommand
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage clikd configuration",
		Long:  `Manage clikd configuration settings.`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigListCmd())

	return cmd
}

// newConfigGetCmd creates a config get subcommand
func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long:  `Get a configuration value by key.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			return runConfigGet(key)
		},
	}
}

// newConfigSetCmd creates a config set subcommand
func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key=value]",
		Short: "Set a configuration value",
		Long:  `Set a configuration value by key.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyValue := args[0]
			return runConfigSet(keyValue)
		},
	}
}

// newConfigListCmd creates a config list subcommand
func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  `List all configuration values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList()
		},
	}
}

// runConfigGet gets a configuration value
func runConfigGet(key string) error {
	// Implementation details...
	return nil
}

// runConfigSet sets a configuration value
func runConfigSet(keyValue string) error {
	// Implementation details...
	return nil
}

// runConfigList lists all configuration values
func runConfigList() error {
	// Implementation details...
	return nil
}
