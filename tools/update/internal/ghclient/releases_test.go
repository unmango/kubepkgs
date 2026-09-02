package ghclient_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/kubepkgs/tools/update/internal/ghclient"
)

const releasesJSON = `[
  {"tag_name": "v1.34.9", "draft": false, "prerelease": false},
  {"tag_name": "v1.34.8", "draft": false, "prerelease": false},
  {"tag_name": "v1.34.10", "draft": false, "prerelease": true},
  {"tag_name": "v1.34.11", "draft": true, "prerelease": false},
  {"tag_name": "v1.35.0", "draft": false, "prerelease": false},
  {"tag_name": "v1.36.0-rc.1", "draft": false, "prerelease": true},
  {"tag_name": "not-a-version", "draft": false, "prerelease": false}
]`

// sharedRepoJSON mirrors a repository publishing several projects' tags, like
// kubernetes/autoscaler.
const sharedRepoJSON = `[
  {"tag_name": "cluster-autoscaler-1.34.5", "draft": false, "prerelease": false},
  {"tag_name": "cluster-autoscaler-1.34.4", "draft": false, "prerelease": false},
  {"tag_name": "vertical-pod-autoscaler-1.34.9", "draft": false, "prerelease": false},
  {"tag_name": "addon-resizer-1.34.20", "draft": false, "prerelease": false}
]`

var _ = Describe("LatestPatch", func() {
	var (
		server *httptest.Server
		client *ghclient.Client
		hits   int
	)

	BeforeEach(func() {
		hits = 0
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/kubernetes/kubernetes/releases", func(w http.ResponseWriter, r *http.Request) {
			hits++
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, releasesJSON)
		})
		server = httptest.NewServer(mux)

		var err error
		client, err = ghclient.NewWithHTTPClient(server.Client(), server.URL+"/")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	It("picks the highest non-draft, non-prerelease patch in the minor series", func() {
		latest, err := client.LatestPatch(context.Background(), "kubernetes", "kubernetes", "v", "1.34")
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).To(Equal("1.34.9"))
	})

	It("excludes prereleases and drafts even if numerically higher", func() {
		latest, err := client.LatestPatch(context.Background(), "kubernetes", "kubernetes", "v", "1.34")
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).NotTo(Equal("1.34.10"))
		Expect(latest).NotTo(Equal("1.34.11"))
	})

	It("returns empty for a minor series with no matching releases", func() {
		latest, err := client.LatestPatch(context.Background(), "kubernetes", "kubernetes", "v", "1.99")
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).To(Equal(""))
	})

	It("caches the release list across repeated calls for the same repo", func() {
		_, err := client.LatestPatch(context.Background(), "kubernetes", "kubernetes", "v", "1.34")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.LatestPatch(context.Background(), "kubernetes", "kubernetes", "v", "1.35")
		Expect(err).NotTo(HaveOccurred())
		Expect(hits).To(Equal(1))
	})
})

var _ = Describe("LatestMinor", func() {
	var (
		server *httptest.Server
		client *ghclient.Client
	)

	BeforeEach(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/kubernetes/kubernetes/releases", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, releasesJSON)
		})
		server = httptest.NewServer(mux)

		var err error
		client, err = ghclient.NewWithHTTPClient(server.Client(), server.URL+"/")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	It("returns the newest minor with a stable release", func() {
		minor, err := client.LatestMinor(context.Background(), "kubernetes", "kubernetes")
		Expect(err).NotTo(HaveOccurred())
		Expect(minor).To(Equal("1.35"))
	})

	It("ignores a minor that has only shipped release candidates", func() {
		minor, err := client.LatestMinor(context.Background(), "kubernetes", "kubernetes")
		Expect(err).NotTo(HaveOccurred())
		Expect(minor).NotTo(Equal("1.36"))
	})
})

var _ = Describe("LatestPatch with a project-specific tag prefix", func() {
	var (
		server *httptest.Server
		client *ghclient.Client
	)

	BeforeEach(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/kubernetes/autoscaler/releases", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, sharedRepoJSON)
		})
		server = httptest.NewServer(mux)

		var err error
		client, err = ghclient.NewWithHTTPClient(server.Client(), server.URL+"/")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	It("matches the prefixed tags and ignores the repository's other projects", func() {
		latest, err := client.LatestPatch(context.Background(),
			"kubernetes", "autoscaler", "cluster-autoscaler-", "1.34")
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).To(Equal("1.34.5"))
	})

	It("finds nothing when the prefix does not match any tag", func() {
		latest, err := client.LatestPatch(context.Background(),
			"kubernetes", "autoscaler", "v", "1.34")
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).To(Equal(""))
	})
})
