package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/grouping"
	"github.com/unmango/kubepkgs/tools/update/internal/nixtool"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

func newVendorHashesCmd() *cobra.Command {
	var sigs []string
	var minor string
	var printGroups bool

	cmd := &cobra.Command{
		Use:   "vendor-hashes",
		Short: "Resolve real Nix vendorHash values for SIG packages, grouped by shared SIG version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runVendorHashes(cmd.Context(), root, sigs, minor, printGroups, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringSliceVar(&sigs, "sig", nil, "restrict to one or more SIGs (default: all)")
	cmd.Flags().StringVar(&minor, "minor", "", "restrict to the group whose build minor matches")
	cmd.Flags().BoolVar(&printGroups, "print-groups", false,
		"print the derived vendor-hash groups and exit, without building or mutating hashes.json")
	return cmd
}

func runVendorHashes(ctx context.Context, root string, sigFilter []string, minorFilter string, printGroups bool, stdout, stderr io.Writer) error {
	versionsPath := filepath.Join(root, "versions.json")
	hashesPath := filepath.Join(root, "hashes.json")

	versions, err := schema.LoadVersions(versionsPath)
	if err != nil {
		return err
	}

	groups := filterGroups(grouping.DeriveGroups(versions), sigFilter, minorFilter)
	if len(groups) == 0 && (len(sigFilter) > 0 || minorFilter != "") {
		return fmt.Errorf("vendor-hashes: no groups matched sig=%v minor=%q", sigFilter, minorFilter)
	}

	if printGroups {
		for _, g := range groups {
			fmt.Fprintln(stdout, g.String())
		}
		return nil
	}

	system, err := nixtool.CurrentSystem(ctx)
	if err != nil {
		return err
	}

	var failed []string
	for _, g := range groups {
		if err := resolveGroup(ctx, hashesPath, versions.Supported, system, g); err != nil {
			fmt.Fprintf(stderr, "FAILED %s (build=%s): %v\n", g.Sig, g.BuildMinor, err)
			failed = append(failed, fmt.Sprintf("%s@%s", g.Sig, g.BuildMinor))
			continue
		}
		fmt.Fprintf(stderr, "%s: resolved via %s, applied to %v\n", g.Sig, g.BuildMinor, g.UpdateMinors)
	}

	if len(failed) > 0 {
		return fmt.Errorf("vendor-hashes: failed groups: %v", failed)
	}
	return nil
}

func filterGroups(groups []grouping.Group, sigs []string, minor string) []grouping.Group {
	if len(sigs) == 0 && minor == "" {
		return groups
	}

	sigSet := map[string]bool{}
	for _, s := range sigs {
		sigSet[s] = true
	}

	var filtered []grouping.Group
	for _, g := range groups {
		if len(sigs) > 0 && !sigSet[g.Sig] {
			continue
		}
		if minor != "" && g.BuildMinor != minor {
			continue
		}
		filtered = append(filtered, g)
	}
	return filtered
}

// resolveGroup resolves the real vendorHash for one group by writing a fake
// hash for g.BuildMinor, building it, and parsing the real hash out of the
// resulting failure. hashes.json is backed up before mutation and restored
// if resolution doesn't succeed, mirroring update-vendor-hash.nix's
// cp/trap/mv dance (including on SIGINT/SIGTERM mid-build).
func resolveGroup(ctx context.Context, hashesPath string, allMinors []string, system string, g grouping.Group) error {
	backup, err := os.ReadFile(hashesPath)
	if err != nil {
		return err
	}
	restore := func() { _ = os.WriteFile(hashesPath, backup, 0o644) }

	hashes, err := schema.LoadHashes(hashesPath)
	if err != nil {
		return err
	}

	buildEntry := hashes.Sigs[g.Sig][g.BuildMinor]
	buildEntry.VendorHash = schema.FakeVendorHash
	hashes.Sigs[g.Sig][g.BuildMinor] = buildEntry
	if err := schema.SaveHashes(hashesPath, hashes, allMinors); err != nil {
		restore()
		return err
	}

	buildCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	attr := fmt.Sprintf(".#legacyPackages.%s.kubernetes.%q.sigs.%s", system, g.BuildMinor, g.Sig)
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

	for _, m := range g.UpdateMinors {
		entry := hashes.Sigs[g.Sig][m]
		entry.VendorHash = hash
		hashes.Sigs[g.Sig][m] = entry
	}
	if err := schema.SaveHashes(hashesPath, hashes, allMinors); err != nil {
		restore()
		return err
	}
	return nil
}
