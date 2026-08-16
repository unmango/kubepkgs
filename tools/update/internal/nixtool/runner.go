// Package nixtool shells out to the Nix-specific tools this port
// deliberately keeps as subprocesses rather than reimplementing in Go:
// nix-prefetch-github (fetchFromGitHub srcHash), nix build (the
// fake-hash/parse trick for resolving a Go module's vendorHash), and nix
// eval (the current Nix system identifier).
package nixtool

import (
	"context"
	"os/exec"
)

// runCombinedOutput runs name with args and returns its combined
// stdout+stderr. Overridable in tests.
var runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
