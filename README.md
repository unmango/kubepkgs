# kubepkgs

[![CI](https://github.com/unmango/kubepkgs/actions/workflows/ci.yml/badge.svg)](https://github.com/unmango/kubepkgs/actions/workflows/ci.yml)
[![NixOS](https://img.shields.io/badge/NixOS-unstable-5277C3?logo=nixos&logoColor=white)](https://nixos.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> [!WARNING]
> This project is a work in progress. Expect breaking changes.

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version ships a package set containing core binaries and selected SIG projects, all built reproducibly with `buildGoModule` against the source and vendor hashes pinned in `packages.json`.

## Supported versions

| Kubernetes        | cluster-api | kube-state-metrics | metrics-server | external-dns |
| ----------------- | ----------- | ------------------ | -------------- | ------------ |
| **1.37** (latest) | 1.10        | 2.14               | 0.7            | 0.15         |
| **1.36**          | 1.10        | 2.14               | 0.7            | 0.15         |
| **1.35**          | 1.9         | 2.14               | 0.7            | 0.15         |
| **1.34**          | 1.9         | 2.13               | 0.7            | 0.15         |
Exact patch versions are pinned in `packages.json` (the table above shows only major/minor).


## Usage

Add the flake to your inputs:

```nix
inputs.kubepkgs.url = "github:unmango/kubepkgs";
```

Then reference packages via `legacyPackages`:

```nix
# Latest Kubernetes version
kubepkgs.legacyPackages.x86_64-linux.kubernetes.latest.kubectl

# Specific minor version
kubepkgs.legacyPackages.x86_64-linux.kubernetes."1.36".kubectl
kubepkgs.legacyPackages.x86_64-linux.kubernetes."1.34".sigs.cluster-api
```

### Available core packages

`kubectl`, `kubeadm`, `kubelet`, `kube-apiserver`, `kube-controller-manager`, `kube-scheduler`, `kube-proxy`

### Available SIG packages

`sigs.cluster-api`, `sigs.kube-state-metrics`, `sigs.metrics-server`, `sigs.external-dns`

## Development

```bash
make build                          # build default package (latest kube-apiserver)
make check                          # nix flake check (build matrix + consistency check)
make fmt                            # format with nixfmt
make update                         # update flake inputs
make add-minor                      # track the newest Kubernetes minor upstream, retire the oldest
make fetch-versions                 # bump packages.json patch versions from upstream releases
make generate-hashes                # fetch srcHash/commit for versions missing them
make vendor-hashes                  # resolve vendorHash for versions missing one
make update-releases                # all three, in order
```

The update lifecycle is implemented by the `tools/update` Go CLI (packaged via `gomod2nix`, `nix run .#update -- <subcommand>`), wired into the `make` targets above. `packages.json` is the single source of truth for tracked versions and pinned hashes; SIG hashes are keyed by the SIG's own version, so a version shared across Kubernetes minors is recorded and resolved once.

`direnv` + `use flake` provides the dev shell automatically.
