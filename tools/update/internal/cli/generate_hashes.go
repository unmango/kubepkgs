package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/ghclient"
	"github.com/unmango/kubepkgs/tools/update/internal/nixtool"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// allTargets is "kubernetes" plus every tracked SIG, the default scope for
// generate-hashes.
var allTargets = append([]string{"kubernetes"}, schema.Sigs...)

func newGenerateHashesCmd() *cobra.Command {
	var targets []string
	var minors []string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "generate-hashes",
		Short: "Populate hashes.json srcHash/commit (and vendorHash placeholders) from versions.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runGenerateHashes(cmd.Context(), ghclient.NewFromEnv(), root, targets, minors, dryRun, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringSliceVar(&targets, "target", nil,
		"targets to refresh (default: all of "+strings.Join(allTargets, ", ")+")")
	cmd.Flags().StringSliceVar(&minors, "minor", nil,
		"minors to refresh (default: all of versions.json's supported minors)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing hashes.json")
	return cmd
}

type hashFailure struct {
	target string
	minor  string
	err    error
}

func runGenerateHashes(ctx context.Context, gh *ghclient.Client, root string, targets, minors []string, dryRun bool, stderr io.Writer) error {
	versionsPath := filepath.Join(root, "versions.json")
	hashesPath := filepath.Join(root, "hashes.json")

	versions, err := schema.LoadVersions(versionsPath)
	if err != nil {
		return err
	}
	hashes, err := schema.LoadHashes(hashesPath)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		targets = allTargets
	}
	if len(minors) == 0 {
		minors = versions.Supported
	}

	var failures []hashFailure

	for _, minor := range minors {
		entry, ok := versions.Kubernetes[minor]
		if !ok {
			failures = append(failures, hashFailure{"kubernetes", minor, fmt.Errorf("minor %q not found in versions.json", minor)})
			continue
		}

		for _, target := range targets {
			owner, repoName, version := "kubernetes", "kubernetes", entry.Version
			if target != "kubernetes" {
				owner, repoName, version = schema.SigOwner[target], target, entry.Sigs.Get(target)
			}

			tag := "v" + version
			srcHash, err := nixtool.PrefetchGithub(ctx, owner, repoName, tag)
			if err != nil {
				failures = append(failures, hashFailure{target, minor, err})
				continue
			}
			commit, err := gh.ResolveCommit(ctx, owner, repoName, tag)
			if err != nil {
				failures = append(failures, hashFailure{target, minor, err})
				continue
			}

			if target == "kubernetes" {
				hashes.Kubernetes[minor] = schema.CoreHash{Version: version, SrcHash: srcHash, Commit: commit}
			} else {
				vendorHash := schema.FakeVendorHash
				if existing, ok := hashes.Sigs[target][minor]; ok && existing.VendorHash != "" {
					vendorHash = existing.VendorHash
				}
				hashes.Sigs[target][minor] = schema.SigHash{Version: version, Commit: commit, VendorHash: vendorHash, SrcHash: srcHash}
			}
			fmt.Fprintf(stderr, "%s %s: %s\n", target, minor, version)
		}
	}

	for _, f := range failures {
		fmt.Fprintf(stderr, "FAILED %s %s: %v\n", f.target, f.minor, f.err)
	}

	if dryRun {
		fmt.Fprintln(stderr, "hashes.json: dry run, not writing")
	} else if err := schema.SaveHashes(hashesPath, hashes, versions.Supported); err != nil {
		return err
	} else {
		fmt.Fprintf(stderr, "Wrote %s\n", hashesPath)
	}

	if len(failures) > 0 {
		return fmt.Errorf("generate-hashes: %d target(s) failed", len(failures))
	}
	return nil
}
