package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/ghclient"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

func newFetchVersionsCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "fetch-versions",
		Short: "Bump versions.json patch versions from upstream GitHub releases",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runFetchVersions(cmd.Context(), ghclient.NewFromEnv(),
				filepath.Join(root, "versions.json"), dryRun, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing versions.json")
	return cmd
}

func runFetchVersions(ctx context.Context, gh *ghclient.Client, path string, dryRun bool, stderr io.Writer) error {
	versions, err := schema.LoadVersions(path)
	if err != nil {
		return err
	}

	changed := false

	for _, minor := range versions.Supported {
		entry := versions.Kubernetes[minor]

		latest, err := gh.LatestPatch(ctx, "kubernetes", "kubernetes", minor)
		if err != nil {
			return fmt.Errorf("fetch-versions: kubernetes %s: %w", minor, err)
		}
		if latest != "" && latest != entry.Version {
			fmt.Fprintf(stderr, "kubernetes %s: %s -> %s\n", minor, entry.Version, latest)
			entry.Version = latest
			changed = true
		} else {
			fmt.Fprintf(stderr, "kubernetes %s: %s (up to date)\n", minor, entry.Version)
		}

		for _, sig := range schema.Sigs {
			current := entry.Sigs.Get(sig)
			owner := schema.SigOwner[sig]

			latest, err := gh.LatestPatch(ctx, owner, sig, minorOf(current))
			if err != nil {
				return fmt.Errorf("fetch-versions: %s %s: %w", sig, minor, err)
			}
			if latest != "" && latest != current {
				fmt.Fprintf(stderr, "  %s %s: %s -> %s\n", sig, minor, current, latest)
				entry.Sigs.Set(sig, latest)
				changed = true
			} else {
				fmt.Fprintf(stderr, "  %s %s: %s (up to date)\n", sig, minor, current)
			}
		}

		versions.Kubernetes[minor] = entry
	}

	if !changed {
		fmt.Fprintln(stderr, "versions.json: no changes")
		return nil
	}
	if dryRun {
		fmt.Fprintln(stderr, "versions.json: changes found (dry run, not writing)")
		return nil
	}

	if err := schema.SaveVersions(path, versions); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Wrote %s\n", path)
	return nil
}

// minorOf strips the trailing patch component off a semver-ish "X.Y.Z"
// string, mirroring fetch-versions.nix's `sig_minor="${current%.*}"`.
func minorOf(version string) string {
	idx := strings.LastIndex(version, ".")
	if idx < 0 {
		return version
	}
	return version[:idx]
}
