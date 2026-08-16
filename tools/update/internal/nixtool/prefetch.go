package nixtool

import (
	"context"
	"encoding/json"
	"fmt"
)

// PrefetchGithub shells to nix-prefetch-github to compute the
// fetchFromGitHub srcHash for owner/repo at rev (e.g. "v1.34.9").
func PrefetchGithub(ctx context.Context, owner, repo, rev string) (string, error) {
	output, err := runCombinedOutput(ctx, "nix-prefetch-github", "--json", "--rev", rev, owner, repo)
	if err != nil {
		return "", fmt.Errorf("nixtool: nix-prefetch-github %s/%s@%s: %w: %s", owner, repo, rev, err, output)
	}

	var result struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("nixtool: parsing nix-prefetch-github output for %s/%s@%s: %w", owner, repo, rev, err)
	}
	return result.Hash, nil
}
