package schema_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

const testdata = "testdata/packages.json"

func load() *schema.File {
	GinkgoHelper()
	f, err := schema.Load(testdata)
	Expect(err).NotTo(HaveOccurred())
	return f
}

func save(f *schema.File) string {
	GinkgoHelper()
	out := filepath.Join(GinkgoT().TempDir(), "packages.json")
	Expect(schema.Save(out, f)).To(Succeed())
	data, err := os.ReadFile(out)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

var _ = Describe("File", func() {
	It("round-trips packages.json byte-for-byte", func() {
		original, err := os.ReadFile(testdata)
		Expect(err).NotTo(HaveOccurred())

		Expect(save(load())).To(Equal(string(original)))
	})

	It("orders minor keys per supported, not alphabetically", func() {
		f := &schema.File{
			Supported: []string{"1.40", "1.9"},
			Latest:    "1.40",
			Kubernetes: map[string]schema.CoreEntry{
				"1.9":  {Version: "1.9.0"},
				"1.40": {Version: "1.40.0"},
			},
		}

		Expect(save(f)).To(ContainSubstring(
			`"kubernetes": {
    "1.40": {`))
	})

	It("orders sig versions by semver, not lexically", func() {
		f := &schema.File{
			Supported:  []string{"1.99"},
			Latest:     "1.99",
			Kubernetes: map[string]schema.CoreEntry{"1.99": {Version: "1.99.0"}},
			Sigs: []schema.Sig{{
				Name:   "example",
				Owner:  "kubernetes-sigs",
				Path:   "network/example",
				Minors: map[string]string{"1.99": "1.10.0"},
				Versions: map[string]schema.SigVersion{
					"1.9.0":  {Commit: "b"},
					"1.10.0": {Commit: "c"},
					"1.2.0":  {Commit: "a"},
				},
			}},
		}

		Expect(save(f)).To(ContainSubstring(
			`"versions": {
        "1.2.0": {
          "commit": "a",`))
		Expect(save(f)).To(MatchRegexp(`(?s)"1\.2\.0".*"1\.9\.0".*"1\.10\.0"`))
	})

	It("fails rather than emitting a supported minor it has no entry for", func() {
		f := load()
		f.Supported = append(f.Supported, "1.99")

		out := filepath.Join(GinkgoT().TempDir(), "packages.json")
		Expect(schema.Save(out, f)).To(MatchError(ContainSubstring("1.99")))
	})
})

var _ = Describe("Sig lookup", func() {
	It("finds a tracked sig by name and reports every name in file order", func() {
		f := load()

		Expect(f.SigNames()).To(Equal([]string{
			"cluster-api", "kube-state-metrics", "metrics-server", "external-dns",
		}))

		sig, ok := f.Sig("metrics-server")
		Expect(ok).To(BeTrue())
		Expect(sig.Owner).To(Equal("kubernetes-sigs"))
		Expect(sig.Path).To(Equal("instrumentation/metrics-server"))

		_, ok = f.Sig("nope")
		Expect(ok).To(BeFalse())
	})

	It("records a version shared by several minors exactly once", func() {
		f := load()

		sig, ok := f.Sig("metrics-server")
		Expect(ok).To(BeTrue())
		Expect(sig.Minors).To(HaveLen(4))
		Expect(sig.Versions).To(HaveLen(1))
		Expect(sig.PinnedBy(f.Supported, "0.7.2")).To(Equal(f.Supported))
	})

	It("resolves the build minor to the first supported minor pinning a version", func() {
		f := load()

		sig, ok := f.Sig("cluster-api")
		Expect(ok).To(BeTrue())

		minor, ok := sig.BuildMinor(f.Supported, "1.9.11")
		Expect(ok).To(BeTrue())
		Expect(minor).To(Equal("1.34"))
		Expect(sig.PinnedBy(f.Supported, "1.9.11")).To(Equal([]string{"1.34", "1.35"}))

		_, ok = sig.BuildMinor(f.Supported, "9.9.9")
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("Work detection", func() {
	It("treats a record as needing a fetch until both hashes are present", func() {
		Expect(schema.CoreEntry{Version: "1.0.0"}.NeedsFetch()).To(BeTrue())
		Expect(schema.CoreEntry{Version: "1.0.0", SrcHash: "sha256-x"}.NeedsFetch()).To(BeTrue())
		Expect(schema.CoreEntry{Version: "1.0.0", Commit: "abc"}.NeedsFetch()).To(BeTrue())
		Expect(schema.CoreEntry{Version: "1.0.0", SrcHash: "sha256-x", Commit: "abc"}.NeedsFetch()).To(BeFalse())

		Expect(schema.SigVersion{}.NeedsFetch()).To(BeTrue())
		Expect(schema.SigVersion{SrcHash: "sha256-x", Commit: "abc"}.NeedsFetch()).To(BeFalse())
	})

	It("treats the fake vendorHash as unresolved", func() {
		Expect(schema.SigVersion{}.NeedsVendorHash()).To(BeTrue())
		Expect(schema.SigVersion{VendorHash: schema.FakeVendorHash}.NeedsVendorHash()).To(BeTrue())
		Expect(schema.SigVersion{VendorHash: "sha256-real"}.NeedsVendorHash()).To(BeFalse())
	})

	It("reports every tracked record as already fetched", func() {
		f := load()

		for _, minor := range f.Supported {
			Expect(f.Kubernetes[minor].NeedsFetch()).To(BeFalse(), "kubernetes %s", minor)
		}
		for _, sig := range f.Sigs {
			for version, record := range sig.Versions {
				Expect(record.NeedsFetch()).To(BeFalse(), "%s %s", sig.Name, version)
				Expect(record.NeedsVendorHash()).To(BeFalse(), "%s %s", sig.Name, version)
			}
		}
	})
})

var _ = Describe("Prune", func() {
	It("drops version records no supported minor pins any more", func() {
		f := load()
		sig, ok := f.Sig("cluster-api")
		Expect(ok).To(BeTrue())
		Expect(sig.Versions).To(HaveLen(3))

		// Move every minor still on 1.9.11 up to 1.10.10.
		for _, minor := range f.Supported {
			if sig.Minors[minor] == "1.9.11" {
				sig.Minors[minor] = "1.10.10"
			}
		}
		Expect(f.Prune()).To(Equal(1))

		Expect(sig.Versions).To(HaveKey("1.10.10"))
		Expect(sig.Versions).NotTo(HaveKey("1.9.11"))
		Expect(sig.Versions).To(HaveKey("1.8.12"))
	})

	It("removes nothing when every tracked version is still pinned", func() {
		Expect(load().Prune()).To(BeZero())
	})
})
