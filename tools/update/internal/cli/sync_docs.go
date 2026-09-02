package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

func newSyncDocsCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync-docs",
		Short: "Regenerate the README's supported-versions table from packages.json",
		Long: "Regenerate the README's supported-versions table from packages.json.\n\n" +
			"add-minor already does this for the changes it makes. Run this after any other edit " +
			"to packages.json that the table reflects, such as adding a SIG or moving one to a new " +
			"version. nix/check-consistency.py stays the independent verifier of the result.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			return runSyncDocs(root, dryRun, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing the docs")
	return cmd
}

func runSyncDocs(root string, dryRun bool, stderr io.Writer) error {
	f, err := schema.Load(packagesPath(root))
	if err != nil {
		return err
	}

	docs, err := renderDocs(root, f, nil)
	if err != nil {
		return fmt.Errorf("sync-docs: %w", err)
	}
	if len(docs) == 0 {
		fmt.Fprintln(stderr, "docs: already in sync")
		return nil
	}
	if dryRun {
		for _, doc := range docs {
			fmt.Fprintf(stderr, "%s: out of date\n", doc.path)
		}
		fmt.Fprintln(stderr, "docs: changes found (dry run, not writing)")
		return nil
	}

	for _, doc := range docs {
		if err := os.WriteFile(doc.path, []byte(doc.content), 0o644); err != nil {
			return fmt.Errorf("sync-docs: writing %s: %w", doc.path, err)
		}
		fmt.Fprintf(stderr, "Wrote %s\n", doc.path)
	}
	return nil
}
