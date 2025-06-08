package initialize

import (
	"os"
	"path/filepath"

	"clikd/internal/config"
	"clikd/internal/ui/bubble"

	tea "github.com/charmbracelet/bubbletea"
)

// determineConfigPath determines the configuration path
func determineConfigPath(m InitModel) tea.Cmd {
	return func() tea.Msg {
		var configDir string

		if m.Global {
			// Determine home directory
			home, err := os.UserHomeDir()
			if err != nil {
				return tea.Quit
			}

			// Global configuration directory
			configDir = filepath.Join(home, ".clikd")
		} else {
			// Local configuration directory
			configDir = "clikd"
		}

		configPath := filepath.Join(configDir, "config.toml")

		return ConfigPathMsg{
			ConfigPath: configPath,
			ConfigDir:  configDir,
		}
	}
}

// checkConfigExists checks if the configuration file exists
func checkConfigExists(configPath string) tea.Cmd {
	return func() tea.Msg {
		_, err := os.Stat(configPath)
		exists := err == nil

		return ConfigExistsMsg{
			Exists: exists,
		}
	}
}

// initConfigManager initializes the configuration manager
func initConfigManager(configPath string, loadExisting bool) tea.Cmd {
	return func() tea.Msg {
		manager := config.NewManager()

		if loadExisting {
			if err := manager.InitConfig(configPath); err != nil {
				return tea.Quit
			}
		} else {
			if err := manager.InitConfig(""); err != nil {
				return tea.Quit
			}
		}

		return ConfigManagerMsg{
			Manager: manager,
		}
	}
}

// checkEnvFileExists returns a tea.Cmd that checks if a .env file exists
func checkEnvFileExists() tea.Cmd {
	return func() tea.Msg {
		_, err := os.Stat(".env")
		exists := err == nil
		return EnvFileExistsMsg{Exists: exists}
	}
}

// confirmOverwrite presents a confirmation dialog to overwrite an existing configuration
func confirmOverwrite(m *InitModel) bool {
	message := "Configuration file already exists at " + m.ConfigPath + ". Overwrite it?"
	return bubble.RunConfirm("Confirm Overwrite", message)
}

// setupChangelogDefaultOptions sets up default changelog options in the config manager
func setupChangelogDefaultOptions(manager *config.Manager) {
	// Set default commit types and patterns
	manager.SetConfigValue("changelog.types", "feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert")
	manager.SetConfigValue("changelog.types.feat", "Features")
	manager.SetConfigValue("changelog.types.fix", "Bug Fixes")
	manager.SetConfigValue("changelog.types.docs", "Documentation")
	manager.SetConfigValue("changelog.types.style", "Styles")
	manager.SetConfigValue("changelog.types.refactor", "Code Refactoring")
	manager.SetConfigValue("changelog.types.perf", "Performance Improvements")
	manager.SetConfigValue("changelog.types.test", "Tests")
	manager.SetConfigValue("changelog.types.build", "Builds")
	manager.SetConfigValue("changelog.types.ci", "Continuous Integration")
	manager.SetConfigValue("changelog.types.chore", "Chores")
	manager.SetConfigValue("changelog.types.revert", "Reverts")
}
