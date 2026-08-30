# Copilot instructions for kubepkgs

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version gets a
package set with core binaries (kubectl, kubelet, kube-apiserver, kube-controller-manager,
kube-scheduler, kube-proxy, kubeadm) plus selected SIG projects (cluster-api,
kube-state-metrics, metrics-server, external-dns), all built with `buildGoModule` against
the source and vendor hashes pinned in `packages.json`. (`gomod2nix`/`buildGoApplication` is used
separately, only to package this repo's own `tools/update` Go CLI (see below).)

Packages are exposed as `legacyPackages.kubernetes."1.XX".<pkg>`,
`legacyPackages.kubernetes."1.XX".sigs.<pkg>`, and `legacyPackages.kubernetes.latest.<pkg>`.
The flake `packages.default` is `latest.kube-apiserver`.

## Commands

```bash
make build                 # nix build .#            (default = latest kube-apiserver)
make check                 # nix flake check         (also: make lint) — the whole build matrix
make fmt                   # nix fmt (treefmt -> nixfmt)   (also: make format)
make update                # nix flake update
```

Build one package directly (quote the attr path):

```bash
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.36".kubectl'
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.33".sigs.cluster-api'
# a single check by name (quote it, the names contain dots):
nix build '.#checks.x86_64-linux."core-1.33-kubectl"'
nix build '.#checks.x86_64-linux."sig-cluster-api-1.9.11"'
```

There is no unit test suite for the K8s core/SIG packages themselves; `doCheck = false` for
all core + SIG `buildGoModule` packages. "Testing" one of those means building it. (The
`tools/update` Go CLI has its own ginkgo/gomega unit test suite, see below.) CI
(`.github/workflows/ci.yml`) runs `nix flake check` on every PR/push to `main`. That is the
whole build matrix: `flake.nix` derives the checks from `packages.json` (every core binary
for `latest`, kubectl for each older supported minor, one build per distinct SIG version), so
a version bump changes what CI covers with no list to maintain. The Makefile holds no
variables.

## Dependency / version update pipeline

Versions and hashes are data-driven, not hand-edited into `.nix` files, and the pipeline that
maintains them is implemented by a Go CLI at `tools/update` (packaged via
`gomod2nix`/`buildGoApplication`, exposed as the flake package `update`) rather than shell
scripts:

- `packages.json`: the single source of truth. Supported minors, `latest`, each minor's
  Kubernetes version + `srcHash` + `commit`, and one record per tracked SIG. Hand-edit it to
  add or retire a minor, or to move a package to a new minor series.
- `make fetch-versions` (`nix run .#update -- fetch-versions`): bumps patch versions from
  GitHub releases, staying inside the minor series each package is already pinned to.
- `make generate-hashes` (`nix run .#update -- generate-hashes`): fills in `srcHash`/`commit`
  via `nix-prefetch-github` + the GitHub API.
- `make vendor-hashes` (`nix run .#update -- vendor-hashes`): resolves real `vendorHash`
  values by building with a fake hash and parsing the reported one.
- `make update-releases`: all three, in order.

Each stage does only outstanding work. A Kubernetes record's `srcHash`/`commit` are written
with the version they describe and cleared by `fetch-versions` on a bump; SIG records are
keyed by the version they describe. So an empty hash means "needs fetching" and populated
records are current by construction. `--force` (generate-hashes) and `--all` (vendor-hashes)
override.

A SIG record holds `owner` and `path` (so the roster of SIG packages is data), a `minors` map
pinning a SIG version per Kubernetes minor, and a `versions` map of hashes keyed by *the SIG's
own version*. Minors pinning the same SIG version therefore share one hash record, and each
SIG version is fetched and vendor-resolved once rather than once per minor.

Narrower runs go through the CLI: `nix run .#update -- generate-hashes --target cluster-api`,
`nix run .#update -- vendor-hashes --sig cluster-api --version 1.10.10`.

## Architecture (read these together)

- **`packages.json`**: the source of truth for what gets built. All other `.nix` files read
  from it via `releases.nix`.
- **`releases.nix`**: resolves `packages.json` per minor, looking each SIG's hashes up by the
  version that minor pins.
- **`mk-release.nix`**: takes one release entry, fetches kubernetes/kubernetes source,
  calls `core/default.nix` for core binaries, then maps each SIG entry through
  `callPackage (./sigs + "/${sig.path}")`. There is no hardcoded SIG attrset.
- **`flake.nix`**: flake-parts entry point. Maps `releases.nix` through `mkRelease` to build
  `legacyPackages.kubernetes`, wires `treefmt` (nixfmt), defines `devShells.default`
  (gnumake, nixfmt, nix-prefetch-github, go, the `gomod2nix` CLI), and exposes the `update`
  package (`nix/updater.nix`, the `tools/update` Go CLI, built via `gomod2nix`'s
  `buildGoApplication`). Derives `checks` from the same data rather than a hand-written list.
- **`core/default.nix`**: builds all core binaries via a shared `mkBin` helper using
  `buildGoModule` with `vendorHash = null` against K8s's own vendored `vendor/` dir. Injects
  version info through `ldflags` mirroring `hack/lib/version.sh` into both
  `k8s.io/client-go/pkg/version` and `k8s.io/component-base/version`.
- **`sigs/<category>/<project>/default.nix`**: each SIG fetches its own GitHub source via
  `fetchFromGitHub` and builds with `buildGoModule` against a real `vendorHash`. `owner` and
  `repo` are arguments, supplied from `packages.json`, not hardcoded.
- **`tools/update/internal/schema`**: owns `packages.json` and is its only writer. Emits
  deterministic key order (minors per `supported`, SIG versions per semver) so a regenerated
  file diffs cleanly.
- **`nix/consistency.nix`** + **`nix/check-consistency.py`**: the `consistency` check.
  Validates `packages.json` against itself (complete entries, no placeholder vendorHash, no
  orphan version records), against the tree (every SIG `path` has a `default.nix`), and
  against the README's supported-versions table. Runs standalone as
  `python3 nix/check-consistency.py .`.
- **`nix/updater.nix`**: `buildGoApplication` derivation for the `tools/update` Go CLI;
  `src` is filtered to just `go.mod`/`go.sum`/`**/*.go`/`**/testdata/**` via the `globset`
  flake input + `lib.fileset.toSource`, so unrelated file changes don't trigger a rebuild.

## Conventions specific to this repo

- **`core/default.nix`'s `GOWORK` must stay on (the default).** K8s ships a complete
  `vendor/` dir generated in workspace mode (`vendor/modules.txt` starts with
  `## workspace`). Forcing `GOWORK = "off"` breaks Go's vendor consistency check against
  that workspace-style `modules.txt` (replace directives get flagged as "not marked as
  replaced"). Don't reintroduce it.
- **`kubectl` is dynamically linked; every other core binary is static.** This mirrors
  `KUBE_STATIC_BINARIES` in upstream `hack/lib/golang.sh` (the third `mkBin` arg toggles
  `-extldflags '-static'` + `CGO_ENABLED = 0`).
- **Reproducibility pins:** core `ldflags` set `buildDate` to the epoch and `gitTreeState` to
  `clean`; `commit` comes from `packages.json`.
- **Adding a K8s minor:** add it to `supported` and `kubernetes`, add it to every SIG's
  `minors` map, update `latest` if it is now the newest, then run `make generate-hashes` and
  `make vendor-hashes`. `make fetch-versions` will not add a minor for you.
- **Adding a SIG package:** create `sigs/<category>/<project>/default.nix` taking
  `owner`/`repo` as arguments, append an entry to `sigs` in `packages.json` with `name`,
  `owner`, `path`, and a `minors` map, leave `versions` as `{}`, then run
  `make generate-hashes` and `make vendor-hashes`. No `.nix` file needs editing.
- **No `passthru.updateScript`.** `nix-update-script` cannot work here: these derivations
  take `version` as an argument rather than embedding it, and updates go through
  `packages.json`. Don't reintroduce it.
- **The README's supported-versions table is checked, not decorative.** Bumping a pin means
  updating that table, or `nix flake check` fails.
- **Formatting:** nixfmt via `nix fmt`. `.editorconfig` enforces final newline + trimmed
  trailing whitespace.
