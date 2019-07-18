package git

import (
	"fmt"
	"strings"

	"github.com/summerwind/terraform-controller/pkg/exec"
)

type Git struct {
	exec.Command
}

func New(dir string) *Git {
	return &Git{
		exec.Command{
			Name:       "git",
			Path:       "git",
			WorkingDir: dir,
			Debug:      false,
		},
	}
}

// Checkout checkouts repository with specified URL and revision.
func (git *Git) Checkout(url, rev string) (string, error) {
	var err error

	// Initialize current directory as a local repository
	_, err = git.Run("init")
	if err != nil {
		return "", fmt.Errorf("failed to initialize as a git repository: %v", err)
	}

	// Add specified URL as a remote repository.
	_, err = git.Run("remote", "add", "origin", url)
	if err != nil {
		return "", fmt.Errorf("failed to add remote repository: %v", err)
	}

	// Fetch most recent commit from origin.
	_, err = git.Run("fetch", "--depth=1", "--recurse-submodules=yes", "origin", rev)
	if err != nil {
		_, err = git.Run("pull", "--recurse-submodules=yes", "origin")
		if err != nil {
			return "", fmt.Errorf("failed to pull source: %v", err)
		}

		_, err = git.Run("checkout", rev)
		if err != nil {
			return "", fmt.Errorf("failed to checkout specified revision: %v", err)
		}
	} else {
		_, err = git.Run("reset", "--hard", "FETCH_HEAD")
		if err != nil {
			return "", fmt.Errorf("failed to reset repository: %v", err)
		}
	}

	// Get current commit hash
	result, err := git.Run("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current revision: %v", err)
	}

	return strings.TrimSpace(string(result.Stdout)), nil
}
