package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"golang.org/x/mod/semver"
)

// FakeVendorHash is the placeholder recorded for a SIG version's vendorHash
// before its real value has been resolved via nixtool.ResolveVendorHash.
const FakeVendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

// CoreEntry is the packages.json record for one Kubernetes minor's source.
type CoreEntry struct {
	Version string `json:"version"`
	SrcHash string `json:"srcHash"`
	Commit  string `json:"commit"`
}

// SigVersion is the packages.json record for one released version of a SIG.
// It is keyed by the SIG's own version rather than by Kubernetes minor, so a
// version shared by several minors is recorded once.
type SigVersion struct {
	Commit     string `json:"commit"`
	SrcHash    string `json:"srcHash"`
	VendorHash string `json:"vendorHash"`
}

// Sig is one tracked SIG project: where to fetch it from, which version each
// Kubernetes minor pins, and the hashes for each of those versions.
type Sig struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
	// Repo is the GitHub repository, when it differs from Name. cluster-autoscaler
	// lives in kubernetes/autoscaler, for instance.
	Repo string `json:"repo,omitempty"`
	// Path locates the package definition, relative to the repo's sigs/ dir.
	Path string `json:"path"`
	// TagPrefix is what a release tag puts in front of the version, when it is
	// not the usual "v". cluster-autoscaler tags "cluster-autoscaler-1.36.1",
	// kustomize tags "kustomize/v5.8.1".
	TagPrefix string `json:"tagPrefix,omitempty"`
	// Subdir is the directory within the repository the module lives in, for
	// projects that share a repository with their siblings.
	Subdir string `json:"subdir,omitempty"`
	// Minors maps a Kubernetes minor to the SIG version it pins.
	Minors map[string]string `json:"minors"`
	// Versions maps a SIG version to its hashes.
	Versions map[string]SigVersion `json:"versions"`
}

// File is the schema of packages.json, the single source of truth for tracked
// versions and pinned hashes.
type File struct {
	Supported  []string             `json:"supported"`
	Latest     string               `json:"latest"`
	Kubernetes map[string]CoreEntry `json:"kubernetes"`
	Sigs       []Sig                `json:"sigs"`
}

// Sig returns the tracked SIG with the given name.
func (f *File) Sig(name string) (*Sig, bool) {
	for i := range f.Sigs {
		if f.Sigs[i].Name == name {
			return &f.Sigs[i], true
		}
	}
	return nil, false
}

// GitHubRepo returns the repository to fetch the SIG from, which is its name
// unless the record says otherwise.
func (s *Sig) GitHubRepo() string {
	if s.Repo != "" {
		return s.Repo
	}
	return s.Name
}

// Tag returns the release tag naming the given version of the SIG. Most tag
// as "v" + version; the ones that don't say so in packages.json.
func (s *Sig) Tag(version string) string {
	if s.TagPrefix != "" {
		return s.TagPrefix + version
	}
	return "v" + version
}

// SigNames returns every tracked SIG name, in file order.
func (f *File) SigNames() []string {
	names := make([]string, len(f.Sigs))
	for i, s := range f.Sigs {
		names[i] = s.Name
	}
	return names
}

// Prune drops Versions records no longer pinned by any supported minor, so a
// version bump doesn't leave the hashes of the version it replaced behind.
// It returns the number of records removed.
func (f *File) Prune() int {
	removed := 0
	for i := range f.Sigs {
		sig := &f.Sigs[i]
		used := map[string]bool{}
		for _, minor := range f.Supported {
			used[sig.Minors[minor]] = true
		}
		for version := range sig.Versions {
			if !used[version] {
				delete(sig.Versions, version)
				removed++
			}
		}
	}
	return removed
}

// MinorOrder returns the supported minors in ascending semver order. It is
// the order Supported is kept in, and therefore the order every minor-keyed
// object in packages.json is emitted in.
func MinorOrder(minors []string) []string {
	ordered := append([]string(nil), minors...)
	sort.Slice(ordered, func(i, j int) bool {
		return semver.Compare("v"+ordered[i], "v"+ordered[j]) < 0
	})
	return ordered
}

// Newest returns the highest supported minor, i.e. the one a newly added
// minor inherits its SIG pins from.
func (f *File) Newest() (string, bool) {
	ordered := MinorOrder(f.Supported)
	if len(ordered) == 0 {
		return "", false
	}
	return ordered[len(ordered)-1], true
}

// Oldest returns the lowest supported minor, i.e. the one AddMinor retires to
// keep the supported window a fixed width.
func (f *File) Oldest() (string, bool) {
	ordered := MinorOrder(f.Supported)
	if len(ordered) == 0 {
		return "", false
	}
	return ordered[0], true
}

// AddMinor starts tracking a Kubernetes minor at the given version. The new
// minor inherits its SIG pins verbatim from the newest minor already
// supported, and becomes Latest if it is now the newest. Its srcHash and
// commit are deliberately left empty: that is exactly CoreEntry.NeedsFetch,
// so generate-hashes picks the record up with no special casing. Moving a SIG
// to a new version stays a separate, deliberate edit.
func (f *File) AddMinor(minor, version string) error {
	if _, ok := f.Kubernetes[minor]; ok {
		return fmt.Errorf("schema: kubernetes %s is already tracked", minor)
	}

	// Resolve every inherited pin before touching the File, so a SIG that
	// cannot be inherited leaves the caller with the File it started with
	// rather than a half-added minor.
	inherit, hasInherit := f.Newest()
	if !hasInherit && len(f.Sigs) > 0 {
		return fmt.Errorf("schema: cannot add %s: no existing minor to inherit SIG pins from", minor)
	}
	inherited := make([]string, len(f.Sigs))
	for i, sig := range f.Sigs {
		pinned, ok := sig.Minors[inherit]
		if !ok {
			return fmt.Errorf("schema: cannot add %s: %s has no version for %s to inherit", minor, sig.Name, inherit)
		}
		inherited[i] = pinned
	}

	f.Supported = MinorOrder(append(f.Supported, minor))
	if f.Kubernetes == nil {
		f.Kubernetes = map[string]CoreEntry{}
	}
	f.Kubernetes[minor] = CoreEntry{Version: version}
	for i := range f.Sigs {
		f.Sigs[i].Minors[minor] = inherited[i]
	}

	if newest, ok := f.Newest(); ok && newest == minor {
		f.Latest = minor
	}
	return nil
}

// RetireMinor stops tracking a Kubernetes minor. Every trace of it has to go:
// Save only checks that supported minors have data, never that unsupported
// ones are absent, so a leftover kubernetes entry or SIG pin survives the
// write and fails later in nix/check-consistency.py. Call Prune afterwards to
// sweep the SIG version records the retired minor was the last to pin.
func (f *File) RetireMinor(minor string) error {
	if _, ok := f.Kubernetes[minor]; !ok {
		return fmt.Errorf("schema: kubernetes %s is not tracked", minor)
	}
	if minor == f.Latest {
		return fmt.Errorf("schema: refusing to retire %s, it is latest", minor)
	}

	remaining := make([]string, 0, len(f.Supported))
	for _, m := range f.Supported {
		if m != minor {
			remaining = append(remaining, m)
		}
	}
	f.Supported = remaining
	delete(f.Kubernetes, minor)
	for i := range f.Sigs {
		delete(f.Sigs[i].Minors, minor)
	}
	return nil
}

// NeedsFetch reports whether this Kubernetes record is missing hashes for the
// version it names. fetch-versions clears both fields when it bumps a
// version, so a populated record always describes its current version.
func (e CoreEntry) NeedsFetch() bool {
	return e.SrcHash == "" || e.Commit == ""
}

// NeedsFetch reports whether this SIG version record is missing its source
// hashes. Records are keyed by the version they describe, so a populated one
// is current by construction.
func (v SigVersion) NeedsFetch() bool {
	return v.SrcHash == "" || v.Commit == ""
}

// NeedsVendorHash reports whether this SIG version still needs its vendorHash
// resolved by a build.
func (v SigVersion) NeedsVendorHash() bool {
	return v.VendorHash == "" || v.VendorHash == FakeVendorHash
}

// VersionOrder returns the SIG's tracked versions in ascending semver order.
func (s *Sig) VersionOrder() []string {
	versions := make([]string, 0, len(s.Versions))
	for v := range s.Versions {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool {
		return semver.Compare("v"+versions[i], "v"+versions[j]) < 0
	})
	return versions
}

// BuildMinor returns the first minor in supported that pins version, i.e. the
// minor whose attribute path resolves to that version of the SIG. Resolving a
// vendorHash means building one such attribute; which one doesn't matter,
// since they all evaluate to the same derivation.
func (s *Sig) BuildMinor(supported []string, version string) (string, bool) {
	for _, minor := range supported {
		if s.Minors[minor] == version {
			return minor, true
		}
	}
	return "", false
}

// PinnedBy returns every minor in supported that pins version, in order.
func (s *Sig) PinnedBy(supported []string, version string) []string {
	var minors []string
	for _, minor := range supported {
		if s.Minors[minor] == version {
			minors = append(minors, minor)
		}
	}
	return minors
}

// Load reads and parses packages.json from path.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("schema: parsing %s: %w", path, err)
	}
	for i := range f.Sigs {
		if f.Sigs[i].Minors == nil {
			f.Sigs[i].Minors = map[string]string{}
		}
		if f.Sigs[i].Versions == nil {
			f.Sigs[i].Versions = map[string]SigVersion{}
		}
	}
	return &f, nil
}

// Save writes f to path as 2-space indented JSON with a trailing newline.
// Object keys are emitted in a deterministic order (minors per Supported,
// SIG versions per semver) rather than the alphabetical order Go's map
// marshaling would otherwise produce, so a regenerated file diffs cleanly.
func Save(path string, f *File) error {
	supportedJSON, err := json.Marshal(f.Supported)
	if err != nil {
		return err
	}
	latestJSON, err := json.Marshal(f.Latest)
	if err != nil {
		return err
	}
	kubernetesJSON, err := marshalOrderedObject(f.Supported, func(minor string) (any, error) {
		entry, ok := f.Kubernetes[minor]
		if !ok {
			return nil, fmt.Errorf("schema: packages.json missing kubernetes entry for supported minor %q", minor)
		}
		return entry, nil
	})
	if err != nil {
		return err
	}
	sigsJSON, err := marshalSigs(f)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`{"supported":`)
	buf.Write(supportedJSON)
	buf.WriteString(`,"latest":`)
	buf.Write(latestJSON)
	buf.WriteString(`,"kubernetes":`)
	buf.Write(kubernetesJSON)
	buf.WriteString(`,"sigs":`)
	buf.Write(sigsJSON)
	buf.WriteByte('}')

	return writeIndentedJSON(path, buf.Bytes())
}

func marshalSigs(f *File) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := range f.Sigs {
		if i > 0 {
			buf.WriteByte(',')
		}
		sig := &f.Sigs[i]

		minorsJSON, err := marshalOrderedObject(f.Supported, func(minor string) (any, error) {
			version, ok := sig.Minors[minor]
			if !ok {
				return nil, fmt.Errorf("schema: packages.json: %s has no version for supported minor %q", sig.Name, minor)
			}
			return version, nil
		})
		if err != nil {
			return nil, err
		}
		versionsJSON, err := marshalOrderedObject(sig.VersionOrder(), func(version string) (any, error) {
			return sig.Versions[version], nil
		})
		if err != nil {
			return nil, err
		}

		head, err := json.Marshal(struct {
			Name      string `json:"name"`
			Owner     string `json:"owner"`
			Repo      string `json:"repo,omitempty"`
			Path      string `json:"path"`
			TagPrefix string `json:"tagPrefix,omitempty"`
			Subdir    string `json:"subdir,omitempty"`
		}{sig.Name, sig.Owner, sig.Repo, sig.Path, sig.TagPrefix, sig.Subdir})
		if err != nil {
			return nil, err
		}

		buf.Write(head[:len(head)-1]) // drop the closing brace, keep appending
		buf.WriteString(`,"minors":`)
		buf.Write(minorsJSON)
		buf.WriteString(`,"versions":`)
		buf.Write(versionsJSON)
		buf.WriteByte('}')
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}
