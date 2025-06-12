package initialize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewInitCmd(t *testing.T) {
	cmd := NewInitCmd()

	// Test basic command properties
	if cmd.Use != "init" {
		t.Errorf("Expected Use to be 'init', got %s", cmd.Use)
	}

	if cmd.Short != "Initialize the clikd configuration" {
		t.Errorf("Expected Short description, got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Initialize the clikd configuration either globally or for the current project") {
		t.Errorf("Expected Long description to contain expected text, got %s", cmd.Long)
	}

	// Test that flags are set
	globalFlag := cmd.Flags().Lookup("global")
	if globalFlag == nil {
		t.Error("Expected --global flag to be present")
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag to be present")
	}

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("Expected --yes flag to be present")
	}

	// Test that subcommands are added
	configCmd := cmd.Commands()
	if len(configCmd) == 0 {
		t.Error("Expected subcommands to be present")
	}

	// Find config subcommand
	var foundConfigCmd *cobra.Command
	for _, subCmd := range configCmd {
		if subCmd.Use == "config" {
			foundConfigCmd = subCmd
			break
		}
	}

	if foundConfigCmd == nil {
		t.Error("Expected 'config' subcommand to be present")
	}
}

func TestNewConfigCmd(t *testing.T) {
	cmd := newConfigCmd()

	// Test basic command properties
	if cmd.Use != "config" {
		t.Errorf("Expected Use to be 'config', got %s", cmd.Use)
	}

	if cmd.Short != "Manage clikd configuration" {
		t.Errorf("Expected Short description, got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Manage clikd configuration settings") {
		t.Errorf("Expected Long description to contain expected text, got %s", cmd.Long)
	}

	// Test that subcommands are added
	subCommands := cmd.Commands()
	if len(subCommands) != 3 {
		t.Errorf("Expected 3 subcommands, got %d", len(subCommands))
	}

	// Check for specific subcommands (extract command name from Use field)
	expectedSubCommands := []string{"get", "set", "list"}
	foundSubCommands := make(map[string]bool)

	for _, subCmd := range subCommands {
		// Extract command name (before any space or bracket)
		parts := strings.Fields(subCmd.Use)
		if len(parts) > 0 {
			foundSubCommands[parts[0]] = true
		}
	}

	for _, expected := range expectedSubCommands {
		if !foundSubCommands[expected] {
			t.Errorf("Expected subcommand '%s' not found. Found: %v", expected, foundSubCommands)
		}
	}
}

func TestNewConfigGetCmd(t *testing.T) {
	cmd := newConfigGetCmd()

	// Test basic command properties
	if cmd.Use != "get [key]" {
		t.Errorf("Expected Use to be 'get [key]', got %s", cmd.Use)
	}

	if cmd.Short != "Get a configuration value" {
		t.Errorf("Expected Short description, got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Get a configuration value by key") {
		t.Errorf("Expected Long description to contain expected text, got %s", cmd.Long)
	}

	// Test args validation
	if cmd.Args == nil {
		t.Error("Expected Args validation to be set")
	}

	// Test that it requires exactly one argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error for no arguments")
	}

	err = cmd.Args(cmd, []string{"key"})
	if err != nil {
		t.Errorf("Expected no error for one argument, got %v", err)
	}

	err = cmd.Args(cmd, []string{"key1", "key2"})
	if err == nil {
		t.Error("Expected error for too many arguments")
	}
}

func TestNewConfigSetCmd(t *testing.T) {
	cmd := newConfigSetCmd()

	// Test basic command properties
	if cmd.Use != "set [key=value]" {
		t.Errorf("Expected Use to be 'set [key=value]', got %s", cmd.Use)
	}

	if cmd.Short != "Set a configuration value" {
		t.Errorf("Expected Short description, got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Set a configuration value by key") {
		t.Errorf("Expected Long description to contain expected text, got %s", cmd.Long)
	}

	// Test args validation
	if cmd.Args == nil {
		t.Error("Expected Args validation to be set")
	}

	// Test that it requires exactly one argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error for no arguments")
	}

	err = cmd.Args(cmd, []string{"key=value"})
	if err != nil {
		t.Errorf("Expected no error for one argument, got %v", err)
	}

	err = cmd.Args(cmd, []string{"key1=value1", "key2=value2"})
	if err == nil {
		t.Error("Expected error for too many arguments")
	}
}

func TestNewConfigListCmd(t *testing.T) {
	cmd := newConfigListCmd()

	// Test basic command properties
	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %s", cmd.Use)
	}

	if cmd.Short != "List all configuration values" {
		t.Errorf("Expected Short description, got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "List all configuration values") {
		t.Errorf("Expected Long description to contain expected text, got %s", cmd.Long)
	}

	// Test that it doesn't require arguments (Args should be nil or allow zero args)
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Expected no error for zero arguments, got %v", err)
		}
	}
}

func TestInitCmd_Flags(t *testing.T) {
	cmd := NewInitCmd()

	tests := []struct {
		flagName     string
		flagType     string
		defaultValue string
		shorthand    string
	}{
		{"global", "bool", "false", "g"},
		{"force", "bool", "false", "f"},
		{"yes", "bool", "false", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}

			if flag.Value.Type() != tt.flagType {
				t.Errorf("Flag %s expected type %s, got %s", tt.flagName, tt.flagType, flag.Value.Type())
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s expected default value %s, got %s", tt.flagName, tt.defaultValue, flag.DefValue)
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s expected shorthand %s, got %s", tt.flagName, tt.shorthand, flag.Shorthand)
			}
		})
	}
}

func TestRunConfigGet(t *testing.T) {
	// Test the runConfigGet function
	err := runConfigGet("test.key")

	// Since the function currently returns nil, we expect no error
	if err != nil {
		t.Errorf("runConfigGet() returned unexpected error: %v", err)
	}
}

func TestRunConfigSet(t *testing.T) {
	// Test the runConfigSet function
	err := runConfigSet("test.key=test.value")

	// Since the function currently returns nil, we expect no error
	if err != nil {
		t.Errorf("runConfigSet() returned unexpected error: %v", err)
	}
}

func TestRunConfigList(t *testing.T) {
	// Test the runConfigList function
	err := runConfigList()

	// Since the function currently returns nil, we expect no error
	if err != nil {
		t.Errorf("runConfigList() returned unexpected error: %v", err)
	}
}

func TestInitCmd_Integration(t *testing.T) {
	// Integration test for the complete command structure
	cmd := NewInitCmd()

	// Test that the command can be executed (though it will call the UI)
	// We'll test the command structure rather than actual execution

	// Verify main command
	if cmd.Use != "init" {
		t.Error("Main command should be 'init'")
	}

	// Verify config subcommand exists and has proper structure
	configCmd := findSubCommand(cmd, "config")
	if configCmd == nil {
		t.Fatal("Config subcommand not found")
	}

	// Verify config subcommands
	getCmd := findSubCommand(configCmd, "get")
	if getCmd == nil {
		t.Error("Config get subcommand not found")
	}

	setCmd := findSubCommand(configCmd, "set")
	if setCmd == nil {
		t.Error("Config set subcommand not found")
	}

	listCmd := findSubCommand(configCmd, "list")
	if listCmd == nil {
		t.Error("Config list subcommand not found")
	}

	// Test flag setting
	err := cmd.Flags().Set("global", "true")
	if err != nil {
		t.Errorf("Failed to set global flag: %v", err)
	}

	err = cmd.Flags().Set("force", "true")
	if err != nil {
		t.Errorf("Failed to set force flag: %v", err)
	}

	err = cmd.Flags().Set("yes", "true")
	if err != nil {
		t.Errorf("Failed to set yes flag: %v", err)
	}

	// Verify flags were set
	globalFlag := cmd.Flags().Lookup("global")
	if globalFlag.Value.String() != "true" {
		t.Error("Global flag was not set correctly")
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag.Value.String() != "true" {
		t.Error("Force flag was not set correctly")
	}

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag.Value.String() != "true" {
		t.Error("Yes flag was not set correctly")
	}
}

func TestConfigCmd_Help(t *testing.T) {
	// Test that help output contains expected information
	cmd := newConfigCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute help
	cmd.Help()

	output := buf.String()
	if !strings.Contains(output, "config") {
		t.Error("Help output should contain 'config'")
	}

	if !strings.Contains(output, "Manage clikd configuration") {
		t.Error("Help output should contain description")
	}
}

// Helper function to find a subcommand by name
func findSubCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == name || strings.HasPrefix(subCmd.Use, name+" ") {
			return subCmd
		}
	}
	return nil
}

func TestInitCmd_CommandStructure(t *testing.T) {
	// Test the complete command structure
	cmd := NewInitCmd()

	// Test command hierarchy - we'll verify the structure directly

	// Check init command
	if len(cmd.Commands()) != 1 {
		t.Errorf("Expected 1 subcommand for init, got %d", len(cmd.Commands()))
	}

	configCmd := cmd.Commands()[0]
	if configCmd.Use != "config" {
		t.Errorf("Expected first subcommand to be 'config', got %s", configCmd.Use)
	}

	// Check config subcommands
	configSubCommands := configCmd.Commands()
	if len(configSubCommands) != 3 {
		t.Errorf("Expected 3 subcommands for config, got %d", len(configSubCommands))
	}

	// Verify all expected subcommands exist
	subCommandNames := make([]string, len(configSubCommands))
	for i, subCmd := range configSubCommands {
		// Extract just the command name (before any space)
		parts := strings.Split(subCmd.Use, " ")
		subCommandNames[i] = parts[0]
	}

	expectedSubCommands := []string{"get", "set", "list"}
	for _, expected := range expectedSubCommands {
		found := false
		for _, actual := range subCommandNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found in %v", expected, subCommandNames)
		}
	}
}
