// Package cli wires the packages.json update lifecycle (add-minor,
// fetch-versions, generate-hashes, vendor-hashes, sync-docs) into cobra
// subcommands of a single kubepkgs-update binary.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/repo"
)

var repoRoot string

// Execute runs the kubepkgs-update root command.
func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubepkgs-update",
		Short:         "Update packages.json for the kubepkgs flake",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.PersistentFlags().StringVar(&repoRoot, "repo-root", "",
		"repository root (default: auto-detected via git rev-parse --show-toplevel)")

	cmd.AddCommand(newAddMinorCmd())
	cmd.AddCommand(newFetchVersionsCmd())
	cmd.AddCommand(newGenerateHashesCmd())
	cmd.AddCommand(newSyncDocsCmd())
	cmd.AddCommand(newVendorHashesCmd())
	return cmd
}

func resolveRepoRoot() (string, error) {
	if repoRoot != "" {
		return repoRoot, nil
	}
	return repo.FindRoot()
}
