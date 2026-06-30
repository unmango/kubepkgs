# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version gets a package set containing core binaries (kubectl, kubelet, kube-apiserver, etc.) and selected SIG projects (cluster-api, kube-state-metrics, metrics-server, external-dns), all built with `gomod2nix`.

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

# Regenerate gomod2nix.toml for one package (PKG= one of GOMOD2NIX_PKGS)
make update-gomod2nix PKG=core-1.36

# Regenerate all gomod2nix.toml files
make update-all-gomod2nix
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

**`flake.nix`** — maps `releases.nix` through `mkRelease` to produce `legacyPackages.kubernetes`, wires up `treefmt` (nixfmt), and exposes `devShells.default` (gnumake + nixfmt + gomod2nix).

**`core/default.nix`** — builds all core K8s binaries via a shared `mkBin` helper using `buildGoApplication`. Has a critical `preConfigure` step that restores `vendor/modules.txt` from source because gomod2nix's hook replaces it, breaking the K8s workspace vendor setup.

**`sigs/<category>/<project>/default.nix`** — each SIG package fetches its own GitHub source and builds with `buildGoApplication`. All expose `passthru.updateGomod2nix` for gomod2nix regeneration.

**`update.nix`** — factory for `updateGomod2nix` shell scripts. Takes `{ src, outdir }` and produces a script that runs `gomod2nix --dir <src> --outdir <outdir>`.

**`gomod2nix.toml` files** — one per package per version, located at `core/<minor>/gomod2nix.toml` and `sigs/<category>/<project>/<minor>/gomod2nix.toml`. Must be regenerated whenever upstream Go dependencies change.

## Adding a new Kubernetes version

1. Add entry to `releases.nix` with `version`, `srcHash`, `modules` path, and `sigs` block.
2. Create `core/<minor>/gomod2nix.toml` by running `make update-gomod2nix PKG=core-<minor>`.
3. Create each SIG's `gomod2nix.toml` similarly.
4. Add `ATTR_` and `GOMOD2NIX_PKGS` entries to `Makefile`.

## Adding a new SIG package

1. Create `sigs/<category>/<project>/default.nix` following the pattern of existing SIG packages.
2. Wire it into `mk-release.nix` under the `sigs` attrset.
3. Add `sigs.<project>` entries to relevant versions in `releases.nix`.
4. Create `gomod2nix.toml` files and `ATTR_` / `GOMOD2NIX_PKGS` entries in `Makefile`.

## Dev environment

`direnv` + `use flake` provides the dev shell automatically. GITHUB_TOKEN is exported via `gh auth token` in `.envrc`.

CI runs `nix flake check` then `nix build .#` on every PR/push to main.
