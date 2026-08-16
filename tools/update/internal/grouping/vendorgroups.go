// Package grouping derives which Kubernetes minors share an identical SIG
// version, so a single resolved vendorHash can be reused across all of them
// instead of being rebuilt per minor. This replaces the hand-maintained
// VENDOR_HASH_ARGS_*/VENDOR_HASH_PKGS mapping that used to live in the
// Makefile.
package grouping

import (
	"fmt"
	"strings"

	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// Group is one SIG version's build-once, propagate-to-many unit: resolve
// vendorHash once by building BuildMinor, then apply the resolved hash to
// every minor in UpdateMinors (which always includes BuildMinor itself).
type Group struct {
	Sig          string
	BuildMinor   string
	UpdateMinors []string
}

func (g Group) String() string {
	return fmt.Sprintf("%s: build=%s update=%s", g.Sig, g.BuildMinor, strings.Join(g.UpdateMinors, " "))
}

// DeriveGroups groups versions.Supported minors by identical SIG version,
// per SIG, in the fixed SIG order (schema.Sigs) and first-seen minor order.
func DeriveGroups(versions *schema.VersionsFile) []Group {
	var groups []Group

	for _, sig := range schema.Sigs {
		indexByVersion := map[string]int{}

		for _, minor := range versions.Supported {
			version := versions.Kubernetes[minor].Sigs.Get(sig)

			if idx, ok := indexByVersion[version]; ok {
				groups[idx].UpdateMinors = append(groups[idx].UpdateMinors, minor)
				continue
			}

			indexByVersion[version] = len(groups)
			groups = append(groups, Group{
				Sig:          sig,
				BuildMinor:   minor,
				UpdateMinors: []string{minor},
			})
		}
	}

	return groups
}
