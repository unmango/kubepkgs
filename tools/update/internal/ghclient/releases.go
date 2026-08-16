package ghclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
	"golang.org/x/mod/semver"
)

// listReleases returns owner/repo's most recent 100 releases, caching the
// result so repeated LatestPatch calls for the same repo (e.g. once per
// tracked Kubernetes minor) don't re-fetch.
func (c *Client) listReleases(ctx context.Context, owner, repo string) ([]*github.RepositoryRelease, error) {
	key := owner + "/" + repo
	if cached, ok := c.cache[key]; ok {
		return cached, nil
	}

	releases, _, err := c.gh.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("ghclient: listing releases for %s: %w", key, err)
	}

	c.cache[key] = releases
	return releases, nil
}

// LatestPatch returns the newest published, non-draft, non-prerelease
// version of owner/repo whose tag falls within the given minor series (e.g.
// minorPrefix "1.34" matches a tag "v1.34.9"), or "" if none match. The
// returned string has no "v" prefix, matching versions.json's convention.
func (c *Client) LatestPatch(ctx context.Context, owner, repo, minorPrefix string) (string, error) {
	releases, err := c.listReleases(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	var best string
	for _, r := range releases {
		if r.GetDraft() || r.GetPrerelease() {
			continue
		}
		version := strings.TrimPrefix(r.GetTagName(), "v")
		if !strings.HasPrefix(version, minorPrefix+".") {
			continue
		}
		if best == "" || semver.Compare("v"+version, "v"+best) > 0 {
			best = version
		}
	}
	return best, nil
}
