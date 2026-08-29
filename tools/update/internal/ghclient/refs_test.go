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

var _ = Describe("ResolveCommit", func() {
	var (
		server *httptest.Server
		client *ghclient.Client
	)

	newClient := func(mux *http.ServeMux) {
		server = httptest.NewServer(mux)
		var err error
		client, err = ghclient.NewWithHTTPClient(server.Client(), server.URL+"/")
		Expect(err).NotTo(HaveOccurred())
	}

	AfterEach(func() {
		server.Close()
	})

	It("uses the ref's commit SHA directly for a lightweight tag", func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/kubernetes-sigs/cluster-api/git/ref/tags/v1.9.11", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"ref":"refs/tags/v1.9.11","object":{"type":"commit","sha":"deadbeef"}}`)
		})
		newClient(mux)

		sha, err := client.ResolveCommit(context.Background(), "kubernetes-sigs", "cluster-api", "v1.9.11")
		Expect(err).NotTo(HaveOccurred())
		Expect(sha).To(Equal("deadbeef"))
	})

	It("dereferences an annotated tag object to its target commit", func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/kubernetes/kubernetes/git/ref/tags/v1.34.9", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"ref":"refs/tags/v1.34.9","object":{"type":"tag","sha":"tagobjectsha"}}`)
		})
		mux.HandleFunc("/repos/kubernetes/kubernetes/git/tags/tagobjectsha", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"sha":"tagobjectsha","object":{"type":"commit","sha":"realcommitsha"}}`)
		})
		newClient(mux)

		sha, err := client.ResolveCommit(context.Background(), "kubernetes", "kubernetes", "v1.34.9")
		Expect(err).NotTo(HaveOccurred())
		Expect(sha).To(Equal("realcommitsha"))
	})
})
