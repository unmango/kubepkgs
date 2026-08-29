package schema_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

var _ = Describe("VersionsFile", func() {
	It("round-trips versions.json byte-for-byte", func() {
		original, err := os.ReadFile("testdata/versions.json")
		Expect(err).NotTo(HaveOccurred())

		v, err := schema.LoadVersions("testdata/versions.json")
		Expect(err).NotTo(HaveOccurred())

		out := filepath.Join(GinkgoT().TempDir(), "versions.json")
		Expect(schema.SaveVersions(out, v)).To(Succeed())

		roundTripped, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(roundTripped)).To(Equal(string(original)))
	})

	It("preserves the fixed sigs field order regardless of map iteration", func() {
		v := &schema.VersionsFile{
			Supported: []string{"1.99"},
			Latest:    "1.99",
			Kubernetes: map[string]schema.MinorEntry{
				"1.99": {
					Version: "1.99.0",
					Sigs: schema.SigSet{
						ClusterAPI:       "1.0.0",
						KubeStateMetrics: "2.0.0",
						MetricsServer:    "3.0.0",
						ExternalDNS:      "4.0.0",
					},
				},
			},
		}

		out := filepath.Join(GinkgoT().TempDir(), "versions.json")
		Expect(schema.SaveVersions(out, v)).To(Succeed())

		data, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(
			`"sigs": {
        "cluster-api": "1.0.0",
        "kube-state-metrics": "2.0.0",
        "metrics-server": "3.0.0",
        "external-dns": "4.0.0"
      }`))
	})
})

var _ = Describe("SigSet", func() {
	It("gets and sets by SIG name", func() {
		var s schema.SigSet
		for _, sig := range schema.Sigs {
			s.Set(sig, sig+"-version")
		}
		for _, sig := range schema.Sigs {
			Expect(s.Get(sig)).To(Equal(sig + "-version"))
		}
	})

	It("no-ops for unknown SIG names", func() {
		var s schema.SigSet
		s.Set("unknown", "1.0.0")
		Expect(s.Get("unknown")).To(Equal(""))
	})
})

var _ = Describe("HashesFile", func() {
	It("round-trips hashes.json byte-for-byte", func() {
		original, err := os.ReadFile("testdata/hashes.json")
		Expect(err).NotTo(HaveOccurred())

		versions, err := schema.LoadVersions("testdata/versions.json")
		Expect(err).NotTo(HaveOccurred())

		h, err := schema.LoadHashes("testdata/hashes.json")
		Expect(err).NotTo(HaveOccurred())

		out := filepath.Join(GinkgoT().TempDir(), "hashes.json")
		Expect(schema.SaveHashes(out, h, versions.Supported)).To(Succeed())

		roundTripped, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(roundTripped)).To(Equal(string(original)))
	})
})
