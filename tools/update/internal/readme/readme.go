// Package readme renders the supported-versions table in README.md from
// packages.json, and retargets the version-pinned examples scattered through
// the docs when the supported window moves.
//
// The table format is dictated by nix/check-consistency.py's check_readme,
// which parses it back out and compares it cell by cell against
// packages.json. That check stays the independent verifier: this package
// generates, it does not validate.
package readme

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// minorColumn is the first column's header. check_readme ignores its
// contents, but every other column name is matched against the SIG roster.
const minorColumn = "Kubernetes"

// rowPattern matches a Markdown table row. check_readme finds the table the
// same way, over the whole file, so the supported-versions table has to be
// the first one and no other pipe-delimited line may precede it.
var rowPattern = regexp.MustCompile(`(?m)^\|.*\|$`)

// Render replaces the version tables in doc with ones generated from f,
// preserving everything around them. The first table lists the SIG pins; when
// f tracks dependencies, the second lists those. Documents are rewritten back
// to front so replacing one table does not shift the next one's offsets.
func Render(doc string, f *schema.File) (string, error) {
	wanted := []([]string){f.SigNames()}
	if len(f.Deps) > 0 {
		wanted = append(wanted, f.DepNames())
	}

	tables := findTables(doc)
	switch {
	case len(tables) == 0:
		return "", fmt.Errorf("readme: no supported-versions table found")
	case len(tables) < len(wanted):
		return "", fmt.Errorf("readme: found %d version table(s), need %d", len(tables), len(wanted))
	}

	for i := len(wanted) - 1; i >= 0; i-- {
		roster := f.Sigs
		if i == 1 {
			roster = f.Deps
		}
		table, err := Table(f, roster)
		if err != nil {
			return "", err
		}

		start, end := tables[i][0], tables[i][1]
		// GFM keeps consuming rows until a blank line, so prose sitting
		// directly under the last row renders as a final row of the table
		// rather than as a paragraph. Guarantee the separation instead of
		// trusting the document to already have it.
		rest := doc[end:]
		if trailer := strings.TrimLeft(rest, "\n"); trailer != rest && trailer != "" {
			rest = "\n\n" + trailer
		}
		doc = doc[:start] + table + rest
	}
	return doc, nil
}

// findTables returns the [start, end) offsets of each contiguous run of table
// rows in doc, so runs separated by prose are treated as separate tables
// rather than one swallowing the next.
func findTables(doc string) [][2]int {
	rows := rowPattern.FindAllStringIndex(doc, -1)

	var tables [][2]int
	for i := 0; i < len(rows); {
		j := i
		for j+1 < len(rows) && strings.TrimSpace(doc[rows[j][1]:rows[j+1][0]]) == "" {
			j++
		}
		// A header, a separator and at least one data row; anything shorter is
		// not one of the version tables.
		if j-i+1 >= 3 {
			tables = append(tables, [2]int{rows[i][0], rows[j][1]})
		}
		i = j + 1
	}
	return tables
}

// Table renders one roster's version table, newest minor first. Cells hold the
// major.minor series of each pin rather than its full version, which is what
// check_readme compares against and why a patch bump leaves the table
// untouched.
func Table(f *schema.File, roster []schema.Sig) (string, error) {
	header := append([]string{minorColumn}, names(roster)...)

	cells := [][]string{header}
	for _, minor := range slices.Backward(schema.MinorOrder(f.Supported)) {
		label := "**" + minor + "**"
		if minor == f.Latest {
			label += " (latest)"
		}
		row := []string{label}
		for _, sig := range roster {
			pinned, ok := sig.Minors[minor]
			if !ok {
				return "", fmt.Errorf("readme: %s has no version for supported minor %q", sig.Name, minor)
			}
			row = append(row, series(pinned))
		}
		cells = append(cells, row)
	}

	widths := make([]int, len(header))
	for _, row := range cells {
		for i, cell := range row {
			widths[i] = max(widths[i], len(cell))
		}
	}

	var b strings.Builder
	writeRow := func(row []string) {
		for i, cell := range row {
			fmt.Fprintf(&b, "| %-*s ", widths[i], cell)
		}
		b.WriteString("|\n")
	}
	writeRow(cells[0])
	separator := make([]string, len(header))
	for i, w := range widths {
		separator[i] = strings.Repeat("-", w)
	}
	writeRow(separator)
	for _, row := range cells[1:] {
		writeRow(row)
	}
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func names(roster []schema.Sig) []string {
	out := make([]string, len(roster))
	for i, sig := range roster {
		out[i] = sig.Name
	}
	return out
}

// series is the major.minor of a pinned version, as check_readme computes it:
// a literal split on dots, not a semver parse.
func series(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}

// examplePatterns match the two ways the docs name a Kubernetes minor: an
// attribute path (kubernetes."1.37".kubectl) and a flake check name
// (core-1.37-kubectl). Both stop resolving once the minor leaves the
// supported window.
var examplePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(kubernetes\.")(\d+\.\d+)(")`),
	regexp.MustCompile(`(core-)(\d+\.\d+)(-)`),
}

// Retarget rewrites version-pinned doc examples according to moves, a map
// from the minor an example currently names to the one it should name.
// Examples naming a minor absent from moves are left alone.
func Retarget(doc string, moves map[string]string) string {
	for _, pattern := range examplePatterns {
		doc = pattern.ReplaceAllStringFunc(doc, func(match string) string {
			groups := pattern.FindStringSubmatch(match)
			moved, ok := moves[groups[2]]
			if !ok {
				return match
			}
			return groups[1] + moved + groups[3]
		})
	}
	return doc
}
