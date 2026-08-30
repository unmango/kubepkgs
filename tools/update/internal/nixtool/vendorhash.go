package nixtool

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

var gotHashRe = regexp.MustCompile(`got:\s*(sha256-[A-Za-z0-9+/=]+)`)

// ResolveVendorHash resolves the real Nix vendorHash for a Go module by
// building attr (the caller is responsible for having already written a
// fake vendorHash for it to packages.json) and parsing the real hash Nix
// reports out of the resulting hash-mismatch failure.
//
// Nix's buildGoModule intentionally fails fast when vendorHash doesn't
// match the actual vendored module tree, reporting the real hash it
// computed in the failure output ("got: sha256-..."). That failure is
// expected and is NOT treated as a Go error: a non-zero exit
// (*exec.ExitError) is inspected for a "got:" hash rather than returned
// as-is. found is false, with a nil error, when the build fails but no
// "got:" hash appears in its output (the actual failure case, which the
// caller should treat as "could not resolve, roll back"). Any other error
// (nix not found, context canceled, etc.) is fatal and returned as err.
func ResolveVendorHash(ctx context.Context, attr string) (hash string, found bool, err error) {
	output, runErr := runCombinedOutput(ctx, "nix", "build", "--no-link", attr)

	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		return "", false, fmt.Errorf("nixtool: running nix build %s: %w", attr, runErr)
	}

	match := gotHashRe.FindSubmatch(output)
	if match == nil {
		return "", false, nil
	}
	return string(match[1]), true, nil
}
