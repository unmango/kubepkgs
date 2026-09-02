package readme_test

import (
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/kubepkgs/tools/update/internal/readme"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// rowPattern mirrors how nix/check-consistency.py finds the table, so these
// specs parse the output the same way the check does.
var rowPattern = regexp.MustCompile(`(?m)^\|.*\|$`)

func cells(row string) []string {
	GinkgoHelper()
	parts := strings.Split(strings.Trim(row, "|"), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func fixture() *schema.File {
	return &schema.File{
		Supported:  []string{"1.35", "1.36"},
		Latest:     "1.36",
		Kubernetes: map[string]schema.CoreEntry{"1.35": {}, "1.36": {}},
		Sigs: []schema.Sig{
			{Name: "cluster-api", Minors: map[string]string{"1.35": "1.9.11", "1.36": "1.10.10"}},
			{Name: "metrics-server", Minors: map[string]string{"1.35": "0.7.2", "1.36": "0.7.2"}},
		},
	}
}

var _ = Describe("Table", func() {
	It("names the SIG columns in packages.json order", func() {
		table, err := readme.Table(fixture())
		Expect(err).NotTo(HaveOccurred())

		Expect(cells(rowPattern.FindAllString(table, -1)[0])).To(
			Equal([]string{"Kubernetes", "cluster-api", "metrics-server"}))
	})

	It("lists minors newest first and marks exactly one latest", func() {
		table, err := readme.Table(fixture())
		Expect(err).NotTo(HaveOccurred())

		rows := rowPattern.FindAllString(table, -1)[2:]
		Expect(cells(rows[0])[0]).To(Equal("**1.36** (latest)"))
		Expect(cells(rows[1])[0]).To(Equal("**1.35**"))
	})

	It("shows the major.minor series rather than the pinned version", func() {
		table, err := readme.Table(fixture())
		Expect(err).NotTo(HaveOccurred())

		Expect(cells(rowPattern.FindAllString(table, -1)[2])).To(
			Equal([]string{"**1.36** (latest)", "1.10", "0.7"}))
	})

	It("moves the latest marker when latest moves", func() {
		f := fixture()
		Expect(f.AddMinor("1.37", "1.37.0")).To(Succeed())

		table, err := readme.Table(f)
		Expect(err).NotTo(HaveOccurred())

		rows := rowPattern.FindAllString(table, -1)[2:]
		Expect(cells(rows[0])[0]).To(Equal("**1.37** (latest)"))
		Expect(cells(rows[1])[0]).To(Equal("**1.36**"))
	})

	It("reports a SIG missing a pin for a supported minor", func() {
		f := fixture()
		delete(f.Sigs[0].Minors, "1.35")

		_, err := readme.Table(f)
		Expect(err).To(MatchError(ContainSubstring("cluster-api has no version")))
	})
})

var _ = Describe("Render", func() {
	const doc = "# kubepkgs\n\n## Supported versions\n\n" +
		"| Kubernetes | cluster-api |\n| ---------- | ----------- |\n| **1.34** (latest) | 1.9 |\n" +
		"Exact patch versions are pinned in `packages.json`.\n\n## Usage\n"

	It("replaces the table and preserves the prose around it", func() {
		out, err := readme.Render(doc, fixture())
		Expect(err).NotTo(HaveOccurred())

		Expect(out).To(HavePrefix("# kubepkgs\n\n## Supported versions\n\n"))
		Expect(out).To(HaveSuffix("Exact patch versions are pinned in `packages.json`.\n\n## Usage\n"))
		Expect(out).NotTo(ContainSubstring("**1.34**"))
		Expect(out).To(ContainSubstring("**1.36** (latest)"))
	})

	It("leaves a later unrelated table alone", func() {
		out, err := readme.Render(doc+"\n| Other | Table |\n| ----- | ----- |\n| a | b |\n", fixture())
		Expect(err).NotTo(HaveOccurred())

		Expect(out).To(ContainSubstring("| Other | Table |"))
	})

	It("reports a document with no table", func() {
		_, err := readme.Render("# kubepkgs\n\nNo table here.\n", fixture())
		Expect(err).To(MatchError(ContainSubstring("no supported-versions table")))
	})
})

var _ = Describe("Retarget", func() {
	moves := map[string]string{"1.33": "1.34", "1.36": "1.37"}

	It("rewrites attribute paths", func() {
		Expect(readme.Retarget(`kubernetes."1.33".sigs.cluster-api`, moves)).To(
			Equal(`kubernetes."1.34".sigs.cluster-api`))
	})

	It("rewrites flake check names", func() {
		Expect(readme.Retarget(`checks.x86_64-linux."core-1.36-kubectl"`, moves)).To(
			Equal(`checks.x86_64-linux."core-1.37-kubectl"`))
	})

	It("leaves minors it was not told to move alone", func() {
		Expect(readme.Retarget(`kubernetes."1.35".kubectl`, moves)).To(
			Equal(`kubernetes."1.35".kubectl`))
	})

	It("leaves versions that only look like minors alone", func() {
		Expect(readme.Retarget("cluster-api 1.36 is pinned", moves)).To(
			Equal("cluster-api 1.36 is pinned"))
	})
})
