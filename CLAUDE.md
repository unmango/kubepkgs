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

# Bump tracked patch versions in packages.json from upstream releases
make fetch-versions

# Refresh srcHash/commit for every tracked version in packages.json
make generate-hashes

# Resolve vendorHash for every tracked SIG version that lacks one
make vendor-hashes

# All three, in order
make update-releases
```

Narrower runs go through the CLI directly, e.g. `nix run .#update -- generate-hashes --target cluster-api` or `nix run .#update -- vendor-hashes --sig cluster-api --version 1.10.10`.

Build a specific package directly:

```bash
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.36".kubectl'
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.33".sigs.cluster-api'
```

## Architecture

**`packages.json`**: the single source of truth. Holds the supported minors, `latest`, each minor's Kubernetes version + `srcHash` + `commit`, and one record per tracked SIG.

A SIG record carries `owner` and `path` (so the roster of SIG packages is data, not a hardcoded attrset), a `minors` map pinning a SIG version per Kubernetes minor, and a `versions` map of hashes keyed by *the SIG's own version*. Two minors that pin the same SIG version therefore share one hash record, and `vendorHash` is resolved once per SIG version rather than once per minor.

**`releases.nix`**: reads `packages.json` and resolves it per minor, looking each SIG's hashes up by the version that minor pins.

**`mk-release.nix`**: takes one release entry from `releases.nix`, fetches the kubernetes/kubernetes source, calls `core/default.nix` for core binaries, then maps each SIG entry through `callPackage (./sigs + "/${sig.path}")`.

**`flake.nix`**: maps `releases.nix` through `mkRelease` to produce `legacyPackages.kubernetes`, wires up `treefmt` (nixfmt), exposes `checks` that build every `latest` core binary and SIG package, exposes the `update` package (`nix/updater.nix`, the `tools/update` Go CLI packaged via `gomod2nix`'s `buildGoApplication`), and exposes `devShells.default` (gnumake + nixfmt + nix-prefetch-github + go + the `gomod2nix` CLI).

**`tools/update/`**: first-party Go module implementing the update lifecycle (`fetch-versions`, `generate-hashes`, `vendor-hashes` cobra subcommands of a single `kubepkgs-update` binary). Packaged separately from core/sigs via `gomod2nix`/`buildGoApplication` (unrelated to the `buildGoModule` setup used for core/sigs, that migration, per commit `5025cf4`, is settled and unaffected by this).

`internal/schema` owns `packages.json`: it is the only writer, and it emits deterministic key order (minors per `supported`, SIG versions per semver) so a regenerated file diffs cleanly.

`fetch-versions` keeps every package inside the minor series it is already pinned to. It does not add a Kubernetes minor, retire one, move `latest`, or move a SIG to a new minor series; those are deliberate edits to `packages.json`.

**`nix/updater.nix`**: the `buildGoApplication` derivation for `tools/update`; filters `src` down to `go.mod`/`go.sum`/`**/*.go`/`**/testdata/**` via the `globset` flake input + `lib.fileset.toSource`, so unrelated file changes in `tools/update` don't trigger a rebuild.

**`core/default.nix`**: builds all core K8s binaries via a shared `mkBin` helper using `buildGoModule`, with `vendorHash = null` against K8s's own vendored `vendor/` dir (no download needed). `vendor/modules.txt` is workspace-generated (`## workspace` header) so `GOWORK` must stay on (default); forcing it off breaks Go's vendor consistency check against the workspace-style modules.txt.

**`sigs/<category>/<project>/default.nix`**: each SIG package fetches its own GitHub source and builds with `buildGoModule` against a real `vendorHash`.

## Adding a new Kubernetes version

1. Add the minor to `supported` and `kubernetes` in `packages.json`, and add it to every SIG's `minors` map. Update `latest` if it is now the newest. `make fetch-versions` will not do this for you.
2. Run `make generate-hashes` then `make vendor-hashes`.

## Adding a new SIG package

1. Create `sigs/<category>/<project>/default.nix` following the pattern of existing SIG packages. It takes `owner`/`repo` as arguments rather than hardcoding them.
2. Append an entry to `sigs` in `packages.json` with `name`, `owner`, `path` (relative to `sigs/`), and a `minors` map pinning a version per supported minor. Leave `versions` as `{}`.
3. Run `make generate-hashes` then `make vendor-hashes`.

No Nix needs editing: `mk-release.nix` maps over whatever `packages.json` declares.

## Dev environment

`direnv` + `use flake` provides the dev shell automatically. GITHUB_TOKEN is exported via `gh auth token` in `.envrc`.

CI runs `nix flake check` then `nix build .#` on every PR/push to main.
