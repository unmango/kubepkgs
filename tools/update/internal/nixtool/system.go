package nixtool

import (
	"context"
	"fmt"
	"strings"
)

// CurrentSystem returns Nix's own system identifier for the current host
// (e.g. "x86_64-linux"), matching `nix eval --impure --raw --expr
// 'builtins.currentSystem'`. This intentionally shells out rather than
// deriving it from runtime.GOOS/GOARCH: the attribute path it's used to
// build (legacyPackages.<system>...) must match Nix's own system string
// exactly, which isn't guaranteed to follow Go's naming.
func CurrentSystem(ctx context.Context) (string, error) {
	output, err := runCombinedOutput(ctx, "nix", "eval", "--impure", "--raw", "--expr", "builtins.currentSystem")
	if err != nil {
		return "", fmt.Errorf("nixtool: nix eval builtins.currentSystem: %w: %s", err, output)
	}
	return strings.TrimSpace(string(output)), nil
}
