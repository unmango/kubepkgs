# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version gets a package set containing core binaries (kubectl, kubelet, kube-apiserver, etc.) and selected SIG projects (cluster-api, kube-state-metrics, metrics-server, external-dns).

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

# Refresh supported k8s minors/versions in versions.json
make fetch-versions

# Regenerate hashes.json entries for all supported minors
make generate-hashes

# Regenerate the vendor hash for one SIG (PKG= one of: cluster-api, kube-state-metrics, metrics-server, external-dns)
# Resolves every group for that SIG (i.e. every distinct version it has across supported minors).
# For a single group only, use `nix run .#update -- vendor-hashes --sig <sig> --minor <build-minor>` directly.
make update-vendor-hash PKG=cluster-api

# Regenerate vendor hashes for all SIGs
make update-all-vendor-hashes

# Run fetch-versions + update-all-vendor-hashes (hashes.json regenerates automatically
# as a dependency of update-all-vendor-hashes if versions.json changed)
make update-releases
```

`ATTR_*`/`CORE_PKGS`/`build-<pkg>` in the `Makefile` still use pseudo-ids like `core-1.33`, `cluster-api-1.10`, etc. Those are unrelated to `PKG` above and unchanged; see `make build-all`.

Build a specific package directly:

```bash
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.36".kubectl'
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.33".sigs.cluster-api'
```

## Architecture

**`versions.json` + `hashes.json`**: source of truth for tracked versions and pinned hashes (`srcHash`, `commit`, and SIG `vendorHash`).

**`releases.nix`**: derives release entries from `versions.json`/`hashes.json`, wiring per-minor core and SIG metadata used by builds.

**`mk-release.nix`**: takes one release entry from `releases.nix`, fetches the kubernetes/kubernetes source, calls `core/default.nix` for core binaries, then calls each SIG `default.nix` for the sigs set.

**`flake.nix`**: maps `releases.nix` through `mkRelease` to produce `legacyPackages.kubernetes`, wires up `treefmt` (nixfmt), exposes `checks` that build every `latest` core binary and SIG package, exposes the `update` package (`nix/updater.nix`, the `tools/update` Go CLI packaged via `gomod2nix`'s `buildGoApplication`), and exposes `devShells.default` (gnumake + nixfmt + nix-prefetch-github + go + the `gomod2nix` CLI).

**`tools/update/`**: first-party Go module implementing the version/hash update lifecycle (`fetch-versions`, `generate-hashes`, `vendor-hashes` cobra subcommands of a single `kubepkgs-update` binary), replacing the previous shell-based `fetch-versions.nix`/`generate-hashes.nix`/`update-vendor-hash.nix`. Packaged separately from core/sigs via `gomod2nix`/`buildGoApplication` (unrelated to the `buildGoModule` setup used for core/sigs, that migration, per commit `5025cf4`, is settled and unaffected by this).

**`nix/updater.nix`**: the `buildGoApplication` derivation for `tools/update`; filters `src` down to `go.mod`/`go.sum`/`**/*.go`/`**/testdata/**` via the `globset` flake input + `lib.fileset.toSource`, so unrelated file changes in `tools/update` don't trigger a rebuild.

**`core/default.nix`**: builds all core K8s binaries via a shared `mkBin` helper using `buildGoModule`, with `vendorHash = null` against K8s's own vendored `vendor/` dir (no download needed). `vendor/modules.txt` is workspace-generated (`## workspace` header) so `GOWORK` must stay on (default); forcing it off breaks Go's vendor consistency check against the workspace-style modules.txt.

**`sigs/<category>/<project>/default.nix`**: each SIG package fetches its own GitHub source and builds with `buildGoModule` against a real `vendorHash`.

**`versions.json`**: supported K8s minors and, per minor, the upstream version for each package (k8s core + sigs).

**`hashes.json`**: per package per minor: `srcHash`, `commit`, and (for sigs) `vendorHash`. Populated by the `tools/update` Go CLI's `generate-hashes`/`vendor-hashes` subcommands (the `update` flake package, wrapped by the Makefile targets above) and consumed by `releases.nix`.

## Adding a new Kubernetes version

1. Add the minor to `versions.json` (or run `make fetch-versions` if it's now the latest upstream release).
2. Run `make generate-hashes` and `make update-all-vendor-hashes` (or `make update-releases` to run the whole fetch/generate/hash flow). Vendor-hash grouping is derived automatically from `versions.json` (no per-minor mapping to hand-maintain).
3. Add `ATTR_` and a `CORE_PKGS` entry to `Makefile` for the new minor.

## Adding a new SIG package

1. Create `sigs/<category>/<project>/default.nix` following the pattern of existing SIG packages.
2. Wire it into `mk-release.nix` under the `sigs` attrset.
3. Add `sigs.<project>` entries to relevant versions in `releases.nix`.
4. Add `ATTR_` and `VENDOR_HASH_PKGS` entries to `Makefile` (used by `build-<pkg>`; vendor-hash resolution itself needs no per-package entry).

## Dev environment

`direnv` + `use flake` provides the dev shell automatically. GITHUB_TOKEN is exported via `gh auth token` in `.envrc`.

CI runs `nix flake check` then `nix build .#` on every PR/push to main.
