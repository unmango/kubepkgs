// Package corepkgs reads the core package roster out of core/default.nix.
//
// Core binaries carry their build configuration in Nix rather than in
// packages.json, so that attribute set is the only place the roster is
// written down. Reading it keeps the README's core inventory derived from the
// same source the packages themselves come from.
package corepkgs

import (
	"fmt"
	"regexp"
)

// bodyPattern finds the `in` that opens the file's result attribute set. The
// `let` bindings above it sit at the same indentation as the attributes below
// it, so the roster can only be read from the part after it. Nested
// `let ... in` blocks are indented and do not match.
var bodyPattern = regexp.MustCompile(`(?m)^in$`)

// attrPattern matches a top-level attribute of the result set. Everything
// inside an attribute's own value is indented further.
var attrPattern = regexp.MustCompile(`(?m)^  ([A-Za-z][\w-]*) = `)

// Names returns the packages core/default.nix declares, in declaration order.
func Names(source string) ([]string, error) {
	body := bodyPattern.FindAllStringIndex(source, -1)
	if body == nil {
		return nil, fmt.Errorf("corepkgs: core/default.nix has no top-level `in`")
	}

	matches := attrPattern.FindAllStringSubmatch(source[body[len(body)-1][1]:], -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("corepkgs: core/default.nix declares no packages")
	}

	names := make([]string, len(matches))
	for i, match := range matches {
		names[i] = match[1]
	}
	return names, nil
}
