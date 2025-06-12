package initialize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInitCmd(t *testing.T) {
	cmd := NewInitCmd()

	// Test basic command properties
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize the clikd configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "Initialize the clikd configuration either globally or for the current project")

	// Test flags
	globalFlag := cmd.Flags().Lookup("global")
	require.NotNil(t, globalFlag)
	assert.Equal(t, "g", globalFlag.Shorthand)
	assert.Equal(t, "false", globalFlag.DefValue)

	forceFlag := cmd.Flags().Lookup("force")
	require.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)

	yesFlag := cmd.Flags().Lookup("yes")
	require.NotNil(t, yesFlag)
	assert.Equal(t, "y", yesFlag.Shorthand)
	assert.Equal(t, "false", yesFlag.DefValue)

	// Test subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1)
	assert.Equal(t, "config", subcommands[0].Use)
}

func TestNewConfigCmd(t *testing.T) {
	cmd := newConfigCmd()

	// Test basic command properties
	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, "Manage clikd configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "Manage clikd configuration settings")

	// Test subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	// Check each subcommand
	subcommandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		subcommandNames[i] = subcmd.Use
	}
	assert.Contains(t, subcommandNames, "get [key]")
	assert.Contains(t, subcommandNames, "set [key=value]")
	assert.Contains(t, subcommandNames, "list")
}

func TestNewConfigGetCmd(t *testing.T) {
	cmd := newConfigGetCmd()

	// Test basic command properties
	assert.Equal(t, "get [key]", cmd.Use)
	assert.Equal(t, "Get a configuration value", cmd.Short)
	assert.Contains(t, cmd.Long, "Get a configuration value by key")

	// Test args validation by testing the function behavior
	assert.NotNil(t, cmd.Args)

	// Test that it requires exactly one argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "Should require at least one argument")

	err = cmd.Args(cmd, []string{"key"})
	assert.NoError(t, err, "Should accept exactly one argument")

	err = cmd.Args(cmd, []string{"key1", "key2"})
	assert.Error(t, err, "Should reject more than one argument")

	// Test that RunE is set
	assert.NotNil(t, cmd.RunE)
}

func TestNewConfigSetCmd(t *testing.T) {
	cmd := newConfigSetCmd()

	// Test basic command properties
	assert.Equal(t, "set [key=value]", cmd.Use)
	assert.Equal(t, "Set a configuration value", cmd.Short)
	assert.Contains(t, cmd.Long, "Set a configuration value by key")

	// Test args validation by testing the function behavior
	assert.NotNil(t, cmd.Args)

	// Test that it requires exactly one argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "Should require at least one argument")

	err = cmd.Args(cmd, []string{"key=value"})
	assert.NoError(t, err, "Should accept exactly one argument")

	err = cmd.Args(cmd, []string{"key1=value1", "key2=value2"})
	assert.Error(t, err, "Should reject more than one argument")

	// Test that RunE is set
	assert.NotNil(t, cmd.RunE)
}

func TestNewConfigListCmd(t *testing.T) {
	cmd := newConfigListCmd()

	// Test basic command properties
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List all configuration values", cmd.Short)
	assert.Contains(t, cmd.Long, "List all configuration values")

	// Test that RunE is set
	assert.NotNil(t, cmd.RunE)
}

func TestRunConfigGet(t *testing.T) {
	// Test the runConfigGet function
	err := runConfigGet("test.key")
	assert.NoError(t, err)

	// Test with different key formats
	err = runConfigGet("ai.provider")
	assert.NoError(t, err)

	err = runConfigGet("general.log_level")
	assert.NoError(t, err)
}

func TestRunConfigSet(t *testing.T) {
	// Test the runConfigSet function
	err := runConfigSet("test.key=value")
	assert.NoError(t, err)

	// Test with different key-value formats
	err = runConfigSet("ai.provider=openai")
	assert.NoError(t, err)

	err = runConfigSet("general.log_level=debug")
	assert.NoError(t, err)
}

func TestRunConfigList(t *testing.T) {
	// Test the runConfigList function
	err := runConfigList()
	assert.NoError(t, err)
}

func TestConfigGetCmdExecution(t *testing.T) {
	cmd := newConfigGetCmd()

	// Test command execution with valid args
	cmd.SetArgs([]string{"test.key"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestConfigSetCmdExecution(t *testing.T) {
	cmd := newConfigSetCmd()

	// Test command execution with valid args
	cmd.SetArgs([]string{"test.key=value"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestConfigListCmdExecution(t *testing.T) {
	cmd := newConfigListCmd()

	// Test command execution
	err := cmd.Execute()
	assert.NoError(t, err)
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
