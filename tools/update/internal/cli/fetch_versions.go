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
		Short: "Bump packages.json patch versions from upstream GitHub releases",
		Long: "Bump packages.json patch versions from upstream GitHub releases.\n\n" +
			"Each tracked package stays within the minor series it is already pinned to: " +
			"Kubernetes within its own minor, each SIG within the minor series that minor pins. " +
			"Moving a package to a new minor series, and adding or retiring a Kubernetes minor, " +
			"are deliberate edits to packages.json rather than something this command does.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runFetchVersions(cmd.Context(), ghclient.NewFromEnv(),
				packagesPath(root), dryRun, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing packages.json")
	return cmd
}

func runFetchVersions(ctx context.Context, gh *ghclient.Client, path string, dryRun bool, stderr io.Writer) error {
	f, err := schema.Load(path)
	if err != nil {
		return err
	}

	changed := false

	for _, minor := range f.Supported {
		entry := f.Kubernetes[minor]

		latest, err := gh.LatestPatch(ctx, "kubernetes", "kubernetes", minor)
		if err != nil {
			return fmt.Errorf("fetch-versions: kubernetes %s: %w", minor, err)
		}
		if latest != "" && latest != entry.Version {
			fmt.Fprintf(stderr, "kubernetes %s: %s -> %s\n", minor, entry.Version, latest)
			entry.Version = latest
			// The recorded hashes describe the version being replaced. Clear
			// them so they can't be mistaken for the new version's, and so
			// generate-hashes knows this entry is the one that needs work.
			entry.SrcHash, entry.Commit = "", ""
			changed = true
		} else {
			fmt.Fprintf(stderr, "kubernetes %s: %s (up to date)\n", minor, entry.Version)
		}
		f.Kubernetes[minor] = entry

		for i := range f.Sigs {
			sig := &f.Sigs[i]
			current := sig.Minors[minor]

			latest, err := gh.LatestPatch(ctx, sig.Owner, sig.Name, minorOf(current))
			if err != nil {
				return fmt.Errorf("fetch-versions: %s %s: %w", sig.Name, minor, err)
			}
			if latest != "" && latest != current {
				fmt.Fprintf(stderr, "  %s %s: %s -> %s\n", sig.Name, minor, current, latest)
				sig.Minors[minor] = latest
				changed = true
			} else {
				fmt.Fprintf(stderr, "  %s %s: %s (up to date)\n", sig.Name, minor, current)
			}
		}
	}

	if !changed {
		fmt.Fprintln(stderr, "packages.json: no changes")
		return nil
	}
	if dryRun {
		fmt.Fprintln(stderr, "packages.json: changes found (dry run, not writing)")
		return nil
	}

	// Versions that are no longer pinned by any minor keep stale hashes alive
	// and would be rebuilt by vendor-hashes for no reason.
	f.Prune()

	if err := schema.Save(path, f); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Wrote %s\n", path)
	return nil
}

// minorOf strips the trailing patch component off a semver-ish "X.Y.Z"
// string, yielding the minor series to search for a newer patch in.
func minorOf(version string) string {
	idx := strings.LastIndex(version, ".")
	if idx < 0 {
		return version
	}
	return version[:idx]
}

func packagesPath(root string) string {
	return filepath.Join(root, "packages.json")
}
