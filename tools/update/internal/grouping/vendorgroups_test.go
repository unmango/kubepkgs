package grouping_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/kubepkgs/tools/update/internal/grouping"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// Golden expectations transcribed from the current Makefile's hand-maintained
// VENDOR_HASH_ARGS_* entries, which DeriveGroups replaces. Any change here
// should be cross-checked against the Makefile (pre-migration) or
// `vendor-hashes --print-groups` (post-migration).
var expected = []grouping.Group{
	{Sig: "cluster-api", BuildMinor: "1.33", UpdateMinors: []string{"1.33"}},
	{Sig: "cluster-api", BuildMinor: "1.34", UpdateMinors: []string{"1.34", "1.35"}},
	{Sig: "cluster-api", BuildMinor: "1.36", UpdateMinors: []string{"1.36"}},
	{Sig: "kube-state-metrics", BuildMinor: "1.33", UpdateMinors: []string{"1.33", "1.34"}},
	{Sig: "kube-state-metrics", BuildMinor: "1.35", UpdateMinors: []string{"1.35", "1.36"}},
	{Sig: "metrics-server", BuildMinor: "1.33", UpdateMinors: []string{"1.33", "1.34", "1.35", "1.36"}},
	{Sig: "external-dns", BuildMinor: "1.33", UpdateMinors: []string{"1.33"}},
	{Sig: "external-dns", BuildMinor: "1.34", UpdateMinors: []string{"1.34", "1.35", "1.36"}},
}

var _ = Describe("DeriveGroups", func() {
	It("matches the current hand-maintained Makefile grouping", func() {
		versions, err := schema.LoadVersions("testdata/versions.json")
		Expect(err).NotTo(HaveOccurred())

		Expect(grouping.DeriveGroups(versions)).To(Equal(expected))
	})

	It("puts a SIG whose version never repeats into one group per minor", func() {
		versions := &schema.VersionsFile{
			Supported: []string{"1.10", "1.11"},
			Kubernetes: map[string]schema.MinorEntry{
				"1.10": {Sigs: schema.SigSet{ClusterAPI: "1.0.0"}},
				"1.11": {Sigs: schema.SigSet{ClusterAPI: "2.0.0"}},
			},
		}

		groups := grouping.DeriveGroups(versions)
		var clusterAPIGroups []grouping.Group
		for _, g := range groups {
			if g.Sig == "cluster-api" {
				clusterAPIGroups = append(clusterAPIGroups, g)
			}
		}

		Expect(clusterAPIGroups).To(Equal([]grouping.Group{
			{Sig: "cluster-api", BuildMinor: "1.10", UpdateMinors: []string{"1.10"}},
			{Sig: "cluster-api", BuildMinor: "1.11", UpdateMinors: []string{"1.11"}},
		}))
	})

	It("puts every minor with a shared version into a single group", func() {
		versions := &schema.VersionsFile{
			Supported: []string{"1.10", "1.11", "1.12"},
			Kubernetes: map[string]schema.MinorEntry{
				"1.10": {Sigs: schema.SigSet{MetricsServer: "0.7.2"}},
				"1.11": {Sigs: schema.SigSet{MetricsServer: "0.7.2"}},
				"1.12": {Sigs: schema.SigSet{MetricsServer: "0.7.2"}},
			},
		}

		groups := grouping.DeriveGroups(versions)
		var metricsServerGroups []grouping.Group
		for _, g := range groups {
			if g.Sig == "metrics-server" {
				metricsServerGroups = append(metricsServerGroups, g)
			}
		}

		Expect(metricsServerGroups).To(Equal([]grouping.Group{
			{Sig: "metrics-server", BuildMinor: "1.10", UpdateMinors: []string{"1.10", "1.11", "1.12"}},
		}))
	})
})
