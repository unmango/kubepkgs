# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version gets a package set containing core binaries (kubectl, kubelet, kube-apiserver, etc.) selected SIG projects (cluster-api, cluster-autoscaler, descheduler, external-dns, kind, kube-state-metrics, kustomize, metrics-server, node-feature-discovery, secrets-store-csi-driver), and etcd.

Packages exposed as `legacyPackages.kubernetes."1.XX".<pkg>` and `legacyPackages.kubernetes.latest.<pkg>`.

## Commands

```bash
# Build default package (latest kube-apiserver)
make build           # nix build .#

# Lint / check
make check           # nix flake check

# Format
make fmt             # nix fmt  (nixfmt)

# Update flake inputs
make update          # nix flake update

# Start tracking the newest Kubernetes minor upstream, retiring the oldest
make add-minor

# Regenerate the README table after an edit to packages.json that it reflects
make sync-docs

# Bump tracked patch versions in packages.json from upstream releases
make fetch-versions

# Fetch srcHash/commit for tracked versions missing them
make generate-hashes

# Resolve vendorHash for every tracked SIG version that lacks one
make vendor-hashes

# All three, in order
make update-releases
```

Narrower runs go through the CLI directly, e.g. `nix run .#update -- generate-hashes --target cluster-api` or `nix run .#update -- vendor-hashes --sig cluster-api --version 1.10.10`.

Build a specific package directly:

```bash
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.37".kubectl'
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.34".sigs.cluster-api'
```

## Architecture

**`packages.json`**: the single source of truth. Holds the supported minors, `latest`, each minor's Kubernetes version + `srcHash` + `commit`, and one record per tracked SIG.

A SIG record carries `owner` and `path` (so the roster of SIG packages is data, not a hardcoded attrset), a `minors` map pinning a SIG version per Kubernetes minor, and a `versions` map of hashes keyed by *the SIG's own version*. Five optional fields cover projects that do not fit the common shape: `repo` when the repository is not named after the project (cluster-autoscaler lives in `kubernetes/autoscaler`), `tagPrefix` when a release tag is not `v` + version (`cluster-autoscaler-1.36.1`, `kustomize/v5.8.1`), `subdir` when the module sits inside a shared repository, `modRoot` when the module to build is inside the source but the rest of the repository must stay (a module that `replace`s its siblings by relative path needs this and *not* `subdir`, which would narrow the source and break those replaces), and `go` when the nixpkgs default toolchain does not build it. A `kubernetes` record takes `go` too.

**Rosters.** Tracked packages live in one of two arrays: `sigs` for Kubernetes SIG projects, `deps` for things a control plane needs that Kubernetes does not own. They differ only in meaning and in which directory holds the definitions (`sigs/` or `deps/`); resolution, hashing, pruning, the README table, and the consistency check are written once and parameterised by the roster. Names must be unique across both, since the CLI's `--target` and `--sig` flags are a flat namespace. Packages are exposed as `legacyPackages.kubernetes."1.XX".sigs.<name>` and `.deps.<name>`. Two minors that pin the same SIG version therefore share one hash record, and `vendorHash` is resolved once per SIG version rather than once per minor.

**`releases.nix`**: reads `packages.json` and resolves it per minor, looking each SIG's hashes up by the version that minor pins.

**`mk-release.nix`**: takes one release entry from `releases.nix`, fetches the kubernetes/kubernetes source, calls `core/default.nix` for core binaries, then maps each SIG entry through `callPackage (./sigs + "/${sig.path}")`. It fetches each SIG's source itself and passes the result down as `src`, the same way core receives one, so the package definitions carry no fetching or tag-scheme logic.

**`flake.nix`**: maps `releases.nix` through `mkRelease` to produce `legacyPackages.kubernetes`, wires up `treefmt` (nixfmt), derives `checks` from the same data (every core binary for `latest`, kubectl for each older supported minor, one build per distinct package version per roster so minors sharing a version are not built twice, plus `consistency`), exposes the `update` package (`nix/updater.nix`, the `tools/update` Go CLI packaged via `gomod2nix`'s `buildGoApplication`), and exposes `devShells.default` (gnumake + nixfmt + nix-prefetch-github + go + the `gomod2nix` CLI).

**`tools/update/`**: first-party Go module implementing the update lifecycle (`add-minor`, `fetch-versions`, `generate-hashes`, `vendor-hashes` cobra subcommands of a single `kubepkgs-update` binary). Packaged separately from core/sigs via `gomod2nix`/`buildGoApplication` (unrelated to the `buildGoModule` setup used for core/sigs, that migration, per commit `5025cf4`, is settled and unaffected by this).

`internal/schema` owns `packages.json`: it is the only writer, and it emits deterministic key order (minors per `supported`, SIG versions per semver) so a regenerated file diffs cleanly.

Each stage does only outstanding work. A Kubernetes record's `srcHash`/`commit` are written together with the version they describe and cleared by `fetch-versions` when that version is bumped, so a populated record is current by construction; SIG records are keyed by the version they describe, so the same holds without any clearing. `generate-hashes` therefore skips populated records and `vendor-hashes` skips resolved ones, making `make update-releases` cheap when little changed. `--force` and `--all` override this.

`fetch-versions` keeps every package inside the minor series it is already pinned to. It does not add a Kubernetes minor, retire one, or move `latest`, that is `add-minor`'s job; moving a SIG to a new minor series stays a deliberate edit to `packages.json`.

`add-minor` discovers the newest minor upstream via `ghclient.LatestMinor`, which shares `LatestPatch`'s prerelease filter, so a minor is only adopted once its stable `.0` ships. It writes the version and pins but leaves `srcHash`/`commit` empty, which is exactly `CoreEntry.NeedsFetch`, so `generate-hashes` picks the record up with no special casing and `add-minor` never has to touch `nixtool`.

**`internal/readme`**: renders the README's supported-versions table from `packages.json` and retargets version-pinned examples across `README.md`, `CLAUDE.md`, and `.github/copilot-instructions.md`. `nix/check-consistency.py` stays the independent verifier of both, so a generator bug is still caught.

**`nix/go-version.nix`**: resolves a `go` pin like `"1.25"` to the matching nixpkgs attribute (`go_1_25`), throwing with the package name, the requested version, and the list nixpkgs actually offers when there is no match. `mk-release.nix` binds the result through `buildGoModule.override`, so a record without a pin is unaffected. This selects among the versions nixpkgs ships; a package needing a Go newer than any of them still needs the nixpkgs input to move.

**`nix/consistency.nix`** + **`nix/check-consistency.py`**: the `consistency` check. Validates that `packages.json` agrees with itself (every supported minor has an entry, every pinned SIG version has a complete hash record, no placeholder vendorHash, no orphan records), with the tree (every SIG `path` has a `default.nix`), and with the README's supported-versions table. Runs standalone as `python3 nix/check-consistency.py .`.

**`nix/updater.nix`**: the `buildGoApplication` derivation for `tools/update`; filters `src` down to `go.mod`/`go.sum`/`**/*.go`/`**/testdata/**` via the `globset` flake input + `lib.fileset.toSource`, so unrelated file changes in `tools/update` don't trigger a rebuild.

**`core/default.nix`**: builds all core K8s binaries via a shared `mkBin` helper using `buildGoModule`, with `vendorHash = null` against K8s's own vendored `vendor/` dir (no download needed). `vendor/modules.txt` is workspace-generated (`## workspace` header) so `GOWORK` must stay on (default); forcing it off breaks Go's vendor consistency check against the workspace-style modules.txt.

**`deps/etcd/{etcd,etcdctl,etcdutl}/default.nix`**: etcd is three Go modules in one repository (`./server`, `./etcdctl`, `./etcdutl`), each with its own dependency graph, so it is three records over the same source rather than one. That gives each a `vendorHash` of its own with no schema special case, and keeps every record yielding exactly one derivation. They use `modRoot`, not `subdir`: the submodules `replace` each other by relative path, so narrowing the source breaks them. `GOWORK = "off"` because etcd 3.7 ships a root `go.work` that would otherwise change module resolution under `modRoot` (the opposite of core, which needs `GOWORK` on). Versions follow what Kubernetes itself pins in `build/dependencies.yaml` per minor.

**`sigs/<category>/<project>/default.nix`**: each SIG package receives a ready `src` and builds it with `buildGoModule` against a real `vendorHash`. A project that ships its own `vendor/` dir sets `deleteVendor = true` rather than `vendorHash = null`, keeping every SIG on one model.

## Adding a new Kubernetes version

1. `make add-minor` adopts the newest minor upstream, or `nix run .#update -- add-minor 1.38` pins a specific one. It adds the minor to `supported` and `kubernetes`, inherits every SIG pin from the newest minor already supported, moves `latest`, retires the oldest minor (pass `--no-retire` to widen the window instead), regenerates the README table, and retargets version-pinned doc examples.
2. Run `make generate-hashes` then `make vendor-hashes` to fill the hashes `add-minor` deliberately leaves empty.

Moving a SIG to a new version stays a separate, deliberate edit to `packages.json`. The `update` workflow runs step 1 and 2 weekly and opens a PR.

## Adding a new package

1. Create `sigs/<category>/<project>/default.nix`, or `deps/<project>/default.nix` for a non-Kubernetes dependency, following the pattern of the existing packages. It takes `src` and `repo` as arguments rather than fetching anything itself.
2. Append an entry to `sigs` or `deps` in `packages.json` with `name`, `owner`, `path` (relative to `sigs/`), and a `minors` map pinning a version per supported minor. Add `repo`, `tagPrefix`, `subdir`, or `go` if the project does not follow the common shape. Leave `versions` as `{}`.
3. Run `make generate-hashes`, then `make vendor-hashes`, then `make sync-docs` to add the README column.

No Nix needs editing: `mk-release.nix` maps over whatever `packages.json` declares.

## Dev environment

`direnv` + `use flake` provides the dev shell automatically. GITHUB_TOKEN is exported via `gh auth token` in `.envrc`.

`.github/workflows/update.yml` runs weekly (and on demand), opening one PR for a new minor and a separate one for patch bumps. It does not run `nix flake check` itself; the PR it opens triggers `ci.yml`, which does.

CI runs `nix flake check` on every PR/push to main. That is the whole build matrix: the checks are derived from `packages.json`, so a version bump changes what CI covers with no list to update.
