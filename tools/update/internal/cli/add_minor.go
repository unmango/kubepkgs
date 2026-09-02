package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/unmango/kubepkgs/tools/update/internal/ghclient"
	"github.com/unmango/kubepkgs/tools/update/internal/readme"
	"github.com/unmango/kubepkgs/tools/update/internal/schema"
)

// exampleDocs are the files carrying version-pinned examples that stop
// resolving once a minor leaves the supported window. They are documentation
// only; nix/check-consistency.py fails the build if one is left behind.
var exampleDocs = []string{
	"README.md",
	"CLAUDE.md",
	filepath.Join(".github", "copilot-instructions.md"),
}

func newAddMinorCmd() *cobra.Command {
	var dryRun bool
	var noRetire bool

	cmd := &cobra.Command{
		Use:   "add-minor [minor]",
		Short: "Start tracking a new Kubernetes minor, retiring the oldest",
		Long: "Start tracking a new Kubernetes minor, retiring the oldest.\n\n" +
			"Without an argument the newest minor upstream is adopted, and the command is a no-op " +
			"when that is one already tracked. The new minor inherits its SIG pins verbatim from the " +
			"newest minor already supported, so moving a SIG to a new version stays a separate change. " +
			"Hashes are left empty for generate-hashes to fill; the tree does not pass `nix flake check` " +
			"until generate-hashes and vendor-hashes have run.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRoot()
			if err != nil {
				return err
			}
			var minor string
			if len(args) == 1 {
				minor = args[0]
			}
			return runAddMinor(cmd.Context(), ghclient.NewFromEnv(),
				root, minor, dryRun, noRetire, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing packages.json or the docs")
	cmd.Flags().BoolVar(&noRetire, "no-retire", false, "keep the oldest minor instead of retiring it, widening the supported window")
	return cmd
}

func runAddMinor(
	ctx context.Context,
	gh *ghclient.Client,
	root, minor string,
	dryRun, noRetire bool,
	stderr io.Writer,
) error {
	path := packagesPath(root)
	f, err := schema.Load(path)
	if err != nil {
		return err
	}

	if minor == "" {
		if minor, err = gh.LatestMinor(ctx, "kubernetes", "kubernetes"); err != nil {
			return fmt.Errorf("add-minor: discovering the newest minor: %w", err)
		}
		if minor == "" {
			return fmt.Errorf("add-minor: kubernetes/kubernetes has no published stable release to discover a minor from")
		}
		if _, tracked := f.Kubernetes[minor]; tracked {
			fmt.Fprintf(stderr, "kubernetes %s: newest upstream, already tracked\n", minor)
			fmt.Fprintln(stderr, "packages.json: no new minor upstream")
			return nil
		}
	}

	version, err := gh.LatestPatch(ctx, "kubernetes", "kubernetes", "v", minor)
	if err != nil {
		return fmt.Errorf("add-minor: kubernetes %s: %w", minor, err)
	}
	if version == "" {
		return fmt.Errorf("add-minor: kubernetes %s has no published release", minor)
	}

	previousLatest := f.Latest
	inherit, _ := f.Newest()
	if err := f.AddMinor(minor, version); err != nil {
		return fmt.Errorf("add-minor: %w", err)
	}
	fmt.Fprintf(stderr, "kubernetes %s: %s (new, inheriting SIG pins from %s)\n", minor, version, inherit)
	for _, sig := range f.Packages {
		fmt.Fprintf(stderr, "  %s %s: %s\n", sig.Name, minor, sig.Minors[minor])
	}

	moves := map[string]string{}
	if f.Latest != previousLatest {
		moves[previousLatest] = f.Latest
	}

	if !noRetire {
		retired, ok := f.Oldest()
		if !ok {
			return fmt.Errorf("add-minor: no minor to retire")
		}
		if err := f.RetireMinor(retired); err != nil {
			return fmt.Errorf("add-minor: %w", err)
		}
		fmt.Fprintf(stderr, "kubernetes %s: retired\n", retired)
		if oldest, ok := f.Oldest(); ok {
			moves[retired] = oldest
		}
		if pruned := f.Prune(); pruned > 0 {
			fmt.Fprintf(stderr, "packages.json: pruned %d orphaned SIG version record(s)\n", pruned)
		}
	}

	docs, err := renderDocs(root, f, moves)
	if err != nil {
		return fmt.Errorf("add-minor: %w", err)
	}

	if dryRun {
		fmt.Fprintln(stderr, "packages.json: changes found (dry run, not writing)")
		return nil
	}

	if err := schema.Save(path, f); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Wrote %s\n", path)

	for _, doc := range docs {
		if err := os.WriteFile(doc.path, []byte(doc.content), 0o644); err != nil {
			return fmt.Errorf("add-minor: writing %s: %w", doc.path, err)
		}
		fmt.Fprintf(stderr, "Wrote %s\n", doc.path)
	}
	return nil
}

type renderedDoc struct {
	path    string
	content string
}

// renderDocs regenerates the README's supported-versions table and retargets
// the version-pinned examples in every doc that carries them, returning only
// those whose content actually changed. Nothing is written here, so a dry run
// still exercises the rendering. Errors are left unprefixed for the calling
// command to attribute, since more than one command renders docs.
func renderDocs(root string, f *schema.File, moves map[string]string) ([]renderedDoc, error) {
	var changed []renderedDoc
	for _, name := range exampleDocs {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		content := readme.Retarget(string(data), moves)
		if name == "README.md" {
			if content, err = readme.Render(content, f); err != nil {
				return nil, err
			}
		}
		if content != string(data) {
			changed = append(changed, renderedDoc{path: path, content: content})
		}
	}
	return changed, nil
}
