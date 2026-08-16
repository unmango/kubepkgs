package ghclient

import (
	"context"
	"fmt"
)

// ResolveCommit returns the commit SHA the given tag (e.g. "v1.34.9") points
// at, dereferencing annotated tag objects to their target commit. Mirrors
// generate-hashes.nix's fetch_commit: a lightweight tag's ref object points
// directly at a commit; an annotated tag's ref object points at a tag
// object, which itself points at the commit.
func (c *Client) ResolveCommit(ctx context.Context, owner, repo, tag string) (string, error) {
	ref, _, err := c.gh.Git.GetRef(ctx, owner, repo, "tags/"+tag)
	if err != nil {
		return "", fmt.Errorf("ghclient: resolving ref tags/%s for %s/%s: %w", tag, owner, repo, err)
	}

	obj := ref.GetObject()
	if obj.GetType() != "tag" {
		return obj.GetSHA(), nil
	}

	tagObj, _, err := c.gh.Git.GetTag(ctx, owner, repo, obj.GetSHA())
	if err != nil {
		return "", fmt.Errorf("ghclient: dereferencing annotated tag %s for %s/%s: %w", tag, owner, repo, err)
	}
	return tagObj.GetObject().GetSHA(), nil
}
