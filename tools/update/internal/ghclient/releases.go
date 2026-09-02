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

// LatestPatch returns the newest published, non-draft, non-prerelease version
// of owner/repo whose tag falls within the given minor series, or "" if none
// match. tagPrefix is what the project puts in front of the version ("v" for
// most, "cluster-autoscaler-" or "kustomize/v" for the ones that share a
// repository); minorPrefix "1.34" then matches a tag "v1.34.9". The returned
// string carries no prefix, matching packages.json's convention.
//
// A tag that does not carry tagPrefix is skipped rather than trimmed, so a
// repository publishing several projects' tags does not leak its siblings'
// releases into this one's series.
func (c *Client) LatestPatch(ctx context.Context, owner, repo, tagPrefix, minorPrefix string) (string, error) {
	releases, err := c.listReleases(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	var best string
	for _, r := range releases {
		if r.GetDraft() || r.GetPrerelease() {
			continue
		}
		tag := r.GetTagName()
		if !strings.HasPrefix(tag, tagPrefix) {
			continue
		}
		version := strings.TrimPrefix(tag, tagPrefix)
		if !strings.HasPrefix(version, minorPrefix+".") {
			continue
		}
		if best == "" || semver.Compare("v"+version, "v"+best) > 0 {
			best = version
		}
	}
	return best, nil
}

// LatestMinor returns the newest minor series owner/repo has a published,
// non-draft, non-prerelease release in, e.g. "1.37". The returned string has
// no "v" prefix, matching packages.json's convention.
//
// Kubernetes publishes a minor's release candidates (v1.37.0-rc.1) before its
// stable .0, and this shares LatestPatch's prerelease filter, so a new minor
// only appears here once its stable .0 ships. That is the right moment to
// start tracking it.
//
// This reads the same single page listReleases caches, so it answers "is
// there a minor newer than the one we track" reliably but is not a complete
// enumeration of every minor a repo has ever released.
func (c *Client) LatestMinor(ctx context.Context, owner, repo string) (string, error) {
	releases, err := c.listReleases(ctx, owner, repo)
	if err != nil {
		return "", err
	}

	var best string
	for _, r := range releases {
		if r.GetDraft() || r.GetPrerelease() {
			continue
		}
		minor := semver.MajorMinor(r.GetTagName())
		if minor == "" {
			continue
		}
		if best == "" || semver.Compare(minor, best) > 0 {
			best = minor
		}
	}
	return strings.TrimPrefix(best, "v"), nil
}
