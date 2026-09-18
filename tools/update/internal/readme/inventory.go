package readme

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// Headings of the package inventories under "## Usage". Each is followed by a
// blank line and then a single paragraph listing that roster's packages.
const (
	CoreHeading = "### Available core packages"
	DepsHeading = "### Available dependency packages"
	SigsHeading = "### Available SIG packages"
)

// entryPattern matches one inventory entry: a backticked package name and the
// optional parenthesised note following it, such as `pause` (Linux only).
var entryPattern = regexp.MustCompile("`([^`]+)`(?: \\(([^)]*)\\))?")

// Inventory rewrites the three package inventories in doc, listing core from
// the roster in core/default.nix and the other two from f. Notes already
// written beside a package are carried over, so a hand-written remark about
// one package survives another being added.
//
// A flat list is a lookup aid, so it is sorted by name rather than kept in
// roster order: that scans better and stays stable when a package is appended
// to a roster. The version tables keep roster order, where each column pairs
// with a header.
func Inventory(doc string, core []string, f *schema.File) (string, error) {
	lists := []struct {
		heading string
		names   []string
	}{
		{CoreHeading, core},
		{DepsHeading, prefixed("deps.", f.DepNames())},
		{SigsHeading, prefixed("sigs.", f.SigNames())},
	}

	lines := strings.Split(doc, "\n")
	for _, list := range lists {
		i, err := findList(lines, list.heading)
		if err != nil {
			return "", err
		}
		lines[i] = renderList(list.names, lines[i])
	}
	return strings.Join(lines, "\n"), nil
}

// findList returns the index of the paragraph listing the packages under
// heading.
func findList(lines []string, heading string) (int, error) {
	for i, line := range lines {
		if strings.TrimSpace(line) != heading {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			switch {
			case strings.TrimSpace(lines[j]) == "":
				continue
			case strings.HasPrefix(lines[j], "#"):
				return 0, fmt.Errorf("readme: %q is followed by a heading, not a package list", heading)
			default:
				return j, nil
			}
		}
		return 0, fmt.Errorf("readme: %q has no package list", heading)
	}
	return 0, fmt.Errorf("readme: no %q heading", heading)
}

// renderList writes names as a comma-separated paragraph in sorted order,
// keeping the notes existing already carries for any name still listed.
func renderList(names []string, existing string) string {
	names = slices.Sorted(slices.Values(names))

	notes := map[string]string{}
	for _, entry := range entryPattern.FindAllStringSubmatch(existing, -1) {
		if entry[2] != "" {
			notes[entry[1]] = entry[2]
		}
	}

	entries := make([]string, len(names))
	for i, name := range names {
		entries[i] = "`" + name + "`"
		if note, ok := notes[name]; ok {
			entries[i] += " (" + note + ")"
		}
	}
	return strings.Join(entries, ", ")
}

func prefixed(prefix string, names []string) []string {
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = prefix + name
	}
	return out
}
