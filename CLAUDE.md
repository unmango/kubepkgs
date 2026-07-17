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

# Regenerate the vendor hash for one package (PKG= one of the ATTR_/VENDOR_HASH_PKGS values below)
make update-vendor-hash PKG=core-1.36

# Regenerate vendor hashes for all packages
make update-all-vendor-hashes

# Run fetch-versions + generate-hashes + update-all-vendor-hashes
make update-releases
```

Valid `PKG` values: `core-1.33`, `core-1.34`, `core-1.35`, `core-1.36`, `cluster-api-1.8`, `cluster-api-1.9`, `cluster-api-1.10`, `kube-state-metrics-2.13`, `kube-state-metrics-2.14`, `metrics-server-0.7`, `external-dns-0.14`, `external-dns-0.15`.

Build a specific package directly:

```bash
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.36".kubectl'
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.33".sigs.cluster-api'
```

## Architecture

**`releases.nix`** — single source of truth. Declares each supported K8s minor version with `srcHash`, `version`, and per-SIG `{ version, hash, modules }` entries. Add new versions here first.

**`mk-release.nix`** — takes one release entry from `releases.nix`, fetches the kubernetes/kubernetes source, calls `core/default.nix` for core binaries, then calls each SIG `default.nix` for the sigs set.

**`flake.nix`** — maps `releases.nix` through `mkRelease` to produce `legacyPackages.kubernetes`, wires up `treefmt` (nixfmt), exposes `checks` that build every `latest` core binary and SIG package, and exposes `devShells.default` (gnumake + nixfmt + gh + jq + nix-prefetch-github).

**`core/default.nix`** — builds all core K8s binaries via a shared `mkBin` helper using `buildGoModule`, with `vendorHash = null` against K8s's own vendored `vendor/` dir (no download needed). `vendor/modules.txt` is workspace-generated (`## workspace` header) so `GOWORK` must stay on (default) — forcing it off breaks Go's vendor consistency check against the workspace-style modules.txt.

**`sigs/<category>/<project>/default.nix`** — each SIG package fetches its own GitHub source and builds with `buildGoModule` against a real `vendorHash`.

**`versions.json`** — supported K8s minors and, per minor, the upstream version for each package (k8s core + sigs).

**`hashes.json`** — per package per minor: `srcHash`, `commit`, and (for sigs) `vendorHash`. Populated by the `fetch-versions`/`generate-hashes`/`update-vendor-hash` flake packages (wrapped by the Makefile targets above) and consumed by `releases.nix`.

## Adding a new Kubernetes version

1. Add the minor to `versions.json` (or run `make fetch-versions` if it's now the latest upstream release).
2. Run `make generate-hashes` and `make update-all-vendor-hashes` (or `make update-releases` to run the whole fetch/generate/hash flow).
3. Add `ATTR_` entries and a `CORE_PKGS`/`VENDOR_HASH_PKGS`/`VENDOR_HASH_ARGS_` entry to `Makefile` for the new minor.

## Adding a new SIG package

1. Create `sigs/<category>/<project>/default.nix` following the pattern of existing SIG packages.
2. Wire it into `mk-release.nix` under the `sigs` attrset.
3. Add `sigs.<project>` entries to relevant versions in `releases.nix`.
4. Add `ATTR_` and `VENDOR_HASH_PKGS`/`VENDOR_HASH_ARGS_` entries to `Makefile`.

## Dev environment

`direnv` + `use flake` provides the dev shell automatically. GITHUB_TOKEN is exported via `gh auth token` in `.envrc`.

CI runs `nix flake check` then `nix build .#` on every PR/push to main.
