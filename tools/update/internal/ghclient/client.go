// Package ghclient wraps the small slice of the GitHub API this tool needs:
// listing releases (to find the latest patch in a minor series) and
// resolving a tag to the commit it points at.
package ghclient

import (
	"net/http"
	"os"

	"github.com/google/go-github/v93/github"
)

// Client wraps a GitHub API client with the release-listing cache LatestPatch
// relies on.
type Client struct {
	gh    *github.Client
	cache map[string][]*github.RepositoryRelease
}

// NewFromEnv builds a Client authenticated via the GITHUB_TOKEN environment
// variable (already exported via `gh auth token` in .envrc for local dev).
func NewFromEnv() (*Client, error) {
	return New(os.Getenv("GITHUB_TOKEN"))
}

// New builds a Client authenticated with the given token. An empty token
// falls back to unauthenticated (rate-limited) API access.
func New(token string) (*Client, error) {
	var opts []github.ClientOptionsFunc
	if token != "" {
		opts = append(opts, github.WithAuthToken(token))
	}
	return newClient(opts...)
}

// NewWithHTTPClient builds a Client that sends requests to baseURL (which
// must end in "/") via httpClient instead of the public GitHub API. Used by
// tests to point at an httptest server.
func NewWithHTTPClient(httpClient *http.Client, baseURL string) (*Client, error) {
	return newClient(github.WithHTTPClient(httpClient), github.WithURLs(&baseURL, nil))
}

func newClient(opts ...github.ClientOptionsFunc) (*Client, error) {
	gh, err := github.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{gh: gh, cache: map[string][]*github.RepositoryRelease{}}, nil
}
