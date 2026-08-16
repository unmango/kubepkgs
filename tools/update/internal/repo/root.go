// Package repo locates the root of the git repository this tool operates on.
package repo

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// FindRoot returns the absolute path to the top level of the current git
// repository, equivalent to `git rev-parse --show-toplevel`.
func FindRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
