package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/nixtool"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

func newVendorHashesCmd() *cobra.Command {
	var sigs []string
	var versions []string
	var all bool

	cmd := &cobra.Command{
		Use:   "vendor-hashes",
		Short: "Resolve real Nix vendorHash values for tracked SIG versions",
		Long: "Resolve real Nix vendorHash values for tracked SIG versions.\n\n" +
			"Each SIG version is resolved once, no matter how many Kubernetes minors pin it. " +
			"By default only versions without a resolved hash are built; pass --all to redo every one.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runVendorHashes(cmd.Context(), packagesPath(root), sigs, versions, all, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringSliceVar(&sigs, "sig", nil, "restrict to one or more packages (default: all)")
	cmd.Flags().StringSliceVar(&versions, "version", nil, "restrict to one or more SIG versions (default: all)")
	cmd.Flags().BoolVar(&all, "all", false, "re-resolve versions that already have a vendorHash")
	return cmd
}

// target is one package version to resolve, and the minor whose attribute
// path evaluates to it.
type target struct {
	roster     string
	sig        string
	version    string
	buildMinor string
	pinnedBy   []string
}

func runVendorHashes(ctx context.Context, path string, sigFilter, versionFilter []string, all bool, stderr io.Writer) error {
	f, err := schema.Load(path)
	if err != nil {
		return err
	}

	targets, err := selectTargets(f, sigFilter, versionFilter, all)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Fprintln(stderr, "vendor-hashes: nothing to resolve")
		return nil
	}

	system, err := nixtool.CurrentSystem(ctx)
	if err != nil {
		return err
	}

	var failed []string
	for _, t := range targets {
		if err := resolve(ctx, path, system, t); err != nil {
			fmt.Fprintf(stderr, "FAILED %s %s: %v\n", t.sig, t.version, err)
			failed = append(failed, fmt.Sprintf("%s@%s", t.sig, t.version))
			continue
		}
		fmt.Fprintf(stderr, "%s %s: resolved via %s, pinned by %v\n", t.sig, t.version, t.buildMinor, t.pinnedBy)
	}

	if len(failed) > 0 {
		return fmt.Errorf("vendor-hashes: failed: %v", failed)
	}
	return nil
}

func selectTargets(f *schema.File, sigFilter, versionFilter []string, all bool) ([]target, error) {
	wanted := func(filter []string, value string) bool {
		if len(filter) == 0 {
			return true
		}
		for _, f := range filter {
			if f == value {
				return true
			}
		}
		return false
	}

	var targets []target
	for roster, sig := range f.Packages {
		if !wanted(sigFilter, sig.Name) {
			continue
		}
		for _, version := range sig.VersionOrder() {
			if !wanted(versionFilter, version) {
				continue
			}
			if !all && !sig.Versions[version].NeedsVendorHash() {
				continue
			}
			minor, ok := sig.BuildMinor(f.Supported, version)
			if !ok {
				return nil, fmt.Errorf("vendor-hashes: %s %s is pinned by no supported minor", sig.Name, version)
			}
			targets = append(targets, target{
				roster:     roster,
				sig:        sig.Name,
				version:    version,
				buildMinor: minor,
				pinnedBy:   sig.PinnedBy(f.Supported, version),
			})
		}
	}

	if len(targets) == 0 && (len(sigFilter) > 0 || len(versionFilter) > 0) && all {
		return nil, fmt.Errorf("vendor-hashes: no versions matched sig=%v version=%v", sigFilter, versionFilter)
	}
	return targets, nil
}

// resolve resolves the real vendorHash for one SIG version by writing a fake
// hash for it, building the attribute, and parsing the real hash out of the
// resulting failure. packages.json is backed up before mutation and restored
// if resolution doesn't succeed, including on SIGINT/SIGTERM mid-build.
func resolve(ctx context.Context, path, system string, t target) error {
	backup, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	restore := func() { _ = os.WriteFile(path, backup, 0o644) }

	f, err := schema.Load(path)
	if err != nil {
		return err
	}
	sig, ok := f.Sig(t.sig)
	if !ok {
		return fmt.Errorf("no such sig %q", t.sig)
	}

	entry := sig.Versions[t.version]
	entry.VendorHash = schema.FakeVendorHash
	sig.Versions[t.version] = entry
	if err := schema.Save(path, f); err != nil {
		restore()
		return err
	}

	buildCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	attr := fmt.Sprintf(".#legacyPackages.%s.kubernetes.%q.%s.%s", system, t.buildMinor, t.roster, t.sig)
	hash, found, err := nixtool.ResolveVendorHash(buildCtx, attr)

	switch {
	case err != nil:
		restore()
		return err
	case buildCtx.Err() != nil:
		restore()
		return fmt.Errorf("interrupted while building %s: %w", attr, buildCtx.Err())
	case !found:
		restore()
		return fmt.Errorf("no vendorHash reported by nix build for %s", attr)
	}

	entry.VendorHash = hash
	sig.Versions[t.version] = entry
	if err := schema.Save(path, f); err != nil {
		restore()
		return err
	}
	return nil
}
