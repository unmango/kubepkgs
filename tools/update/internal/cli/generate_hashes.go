package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/ghclient"
	"github.com/unmango/kubepkgs/tools/update/internal/nixtool"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// kubernetesTarget is the generate-hashes target name for the Kubernetes
// core source, alongside one target per tracked SIG.
const kubernetesTarget = "kubernetes"

func newGenerateHashesCmd() *cobra.Command {
	var targets []string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "generate-hashes",
		Short: "Populate packages.json srcHash/commit (and vendorHash placeholders) from the tracked versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runGenerateHashes(cmd.Context(), ghclient.NewFromEnv(),
				packagesPath(root), targets, dryRun, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringSliceVar(&targets, "target", nil,
		"targets to refresh: "+kubernetesTarget+" or a SIG name (default: all)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing packages.json")
	return cmd
}

type hashFailure struct {
	target  string
	version string
	err     error
}

func runGenerateHashes(ctx context.Context, gh *ghclient.Client, path string, targets []string, dryRun bool, stderr io.Writer) error {
	f, err := schema.Load(path)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		targets = append([]string{kubernetesTarget}, f.SigNames()...)
	}
	for _, target := range targets {
		if target == kubernetesTarget {
			continue
		}
		if _, ok := f.Sig(target); !ok {
			return fmt.Errorf("generate-hashes: unknown target %q (known: %s)",
				target, strings.Join(append([]string{kubernetesTarget}, f.SigNames()...), ", "))
		}
	}

	// Stale records for versions no longer pinned would otherwise be
	// refetched here and never used.
	f.Prune()

	var failures []hashFailure

	for _, target := range targets {
		if target == kubernetesTarget {
			for _, minor := range f.Supported {
				entry, ok := f.Kubernetes[minor]
				if !ok {
					failures = append(failures, hashFailure{target, minor, fmt.Errorf("minor %q not found in packages.json", minor)})
					continue
				}
				srcHash, commit, err := fetchSource(ctx, gh, "kubernetes", "kubernetes", entry.Version)
				if err != nil {
					failures = append(failures, hashFailure{target, entry.Version, err})
					continue
				}
				entry.SrcHash, entry.Commit = srcHash, commit
				f.Kubernetes[minor] = entry
				fmt.Fprintf(stderr, "kubernetes %s: %s\n", minor, entry.Version)
			}
			continue
		}

		sig, _ := f.Sig(target)
		// Keyed by SIG version, so a version shared by several minors is
		// fetched once rather than once per minor.
		for _, version := range sigVersionsInUse(f, sig) {
			srcHash, commit, err := fetchSource(ctx, gh, sig.Owner, sig.Name, version)
			if err != nil {
				failures = append(failures, hashFailure{sig.Name, version, err})
				continue
			}
			vendorHash := schema.FakeVendorHash
			if existing, ok := sig.Versions[version]; ok && existing.VendorHash != "" {
				vendorHash = existing.VendorHash
			}
			sig.Versions[version] = schema.SigVersion{Commit: commit, SrcHash: srcHash, VendorHash: vendorHash}
			fmt.Fprintf(stderr, "%s %s (pinned by %v)\n", sig.Name, version, sig.PinnedBy(f.Supported, version))
		}
	}

	for _, fail := range failures {
		fmt.Fprintf(stderr, "FAILED %s %s: %v\n", fail.target, fail.version, fail.err)
	}

	if dryRun {
		fmt.Fprintln(stderr, "packages.json: dry run, not writing")
	} else if err := schema.Save(path, f); err != nil {
		return err
	} else {
		fmt.Fprintf(stderr, "Wrote %s\n", path)
	}

	if len(failures) > 0 {
		return fmt.Errorf("generate-hashes: %d target(s) failed", len(failures))
	}
	return nil
}

// sigVersionsInUse returns every distinct version the supported minors pin,
// in ascending semver order. Versions absent from sig.Versions (a fresh bump
// that has no record yet) are included.
func sigVersionsInUse(f *schema.File, sig *schema.Sig) []string {
	for _, minor := range f.Supported {
		version := sig.Minors[minor]
		if _, ok := sig.Versions[version]; !ok {
			sig.Versions[version] = schema.SigVersion{}
		}
	}
	return sig.VersionOrder()
}

func fetchSource(ctx context.Context, gh *ghclient.Client, owner, repo, version string) (srcHash, commit string, err error) {
	tag := "v" + version
	if srcHash, err = nixtool.PrefetchGithub(ctx, owner, repo, tag); err != nil {
		return "", "", err
	}
	if commit, err = gh.ResolveCommit(ctx, owner, repo, tag); err != nil {
		return "", "", err
	}
	return srcHash, commit, nil
}
