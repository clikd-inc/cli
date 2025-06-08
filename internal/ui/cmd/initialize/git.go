package initialize

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// checkGitRepo checks if the current directory is a Git repository
// and returns GitRepoMsg containing the repository status
func checkGitRepo() tea.Msg {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()

	isInGitRepo := err == nil
	repoURL := ""

	if isInGitRepo {
		// Determine repository URL
		remoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
		if remoteOutput, err := remoteCmd.Output(); err == nil {
			repoURL = strings.TrimSpace(string(remoteOutput))
		}
	}

	return GitRepoMsg{
		IsInGitRepo: isInGitRepo,
		RepoURL:     repoURL,
	}
}
