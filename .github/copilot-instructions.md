# Copilot instructions for kubepkgs

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version gets a
package set with core binaries (kubectl, kubelet, kube-apiserver, kube-controller-manager,
kube-scheduler, kube-proxy, kubeadm) plus selected SIG projects (cluster-api,
kube-state-metrics, metrics-server, external-dns), all built with `gomod2nix` /
`buildGoApplication`.

Packages are exposed as `legacyPackages.kubernetes."1.XX".<pkg>`,
`legacyPackages.kubernetes."1.XX".sigs.<pkg>`, and `legacyPackages.kubernetes.latest.<pkg>`.
The flake `packages.default` is `latest.kube-apiserver`.

## Commands

```bash
make build                 # nix build .#            (default = latest kube-apiserver)
make build-all             # build kubectl (per tracked K8s minor) + each tracked SIG package for current system
make check                 # nix flake check         (also: make lint)
make fmt                   # nix fmt (treefmt -> nixfmt)   (also: make format)
make update                # nix flake update
```

Build one package directly (quote the attr path):

```bash
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.36".kubectl'
nix build '.#legacyPackages.x86_64-linux.kubernetes."1.33".sigs.cluster-api'
# or via the Makefile helper (uses current system automatically):
make build-core-1.36
make build-cluster-api-1.10
```

There is no unit test suite; `doCheck = false` for all core + SIG `buildGoApplication` packages. "Testing" a package
means building it. CI (`.github/workflows/ci.yml`) runs `nix flake check` then
`nix develop -c make build-all` on every PR/push to `main`.

## Dependency / version update pipeline

Versions and hashes are data-driven, not hand-edited into `.nix` files:

- `versions.json` — tracked minor versions (`supported`, `latest`) and the resolved
  core + SIG semver for each. This is the file you edit by hand to add/track a version.
- `hashes.json` — generated `srcHash` + `commit` per package. Marked "Do not edit manually".
- `make fetch-versions` (`nix run .#fetch-versions`) — bumps patch versions in `versions.json`
  from GitHub releases.
- `make generate-hashes` (`nix run .#generate-hashes`) — populates `hashes.json` via
  `nix-prefetch-github`.
- `make update-gomod2nix PKG=<pkg>` — regenerate one `gomod2nix.toml`.
  `make update-all-gomod2nix` — regenerate all.
- `make update-releases` — runs fetch-versions → generate-hashes → update-all-gomod2nix.

Valid `PKG` values (see `GOMOD2NIX_PKGS` in the Makefile): `core-1.33`, `core-1.34`,
`core-1.35`, `core-1.36`, `cluster-api-1.8`, `cluster-api-1.9`, `cluster-api-1.10`,
`kube-state-metrics-2.13`, `kube-state-metrics-2.14`, `metrics-server-0.7`,
`external-dns-0.14`, `external-dns-0.15`.

`gomod2nix.toml` files live at `core/<minor>/gomod2nix.toml` and
`sigs/<category>/<project>/<minor>/gomod2nix.toml`, and must be regenerated whenever
upstream Go dependencies change.

## Architecture (read these together)

- **`versions.json` + `hashes.json`** — the source of truth for what gets built. All other
  `.nix` files read from these via `releases.nix`.
- **`releases.nix`** — maps `versions.json`/`hashes.json` into per-minor release entries
  (`version`, `srcHash`, `commit`, `modules` path, and a `sigs` block). SIG base paths and
  the K8s-minor → SIG-version wiring live here.
- **`mk-release.nix`** — takes one release entry, fetches kubernetes/kubernetes source,
  calls `core/default.nix` for core binaries, then calls each SIG `default.nix` under `sigs`.
- **`flake.nix`** — flake-parts entry point. Maps `releases.nix` through `mkRelease` to build
  `legacyPackages.kubernetes`, wires `treefmt` (nixfmt), defines `devShells.default`
  (gnumake, nixfmt, gomod2nix, gh, jq, nix-prefetch-github), and exposes the
  `generate-hashes` / `fetch-versions` packages.
- **`core/default.nix`** — builds all core binaries via a shared `mkBin` helper. Injects
  version info through `ldflags` mirroring `hack/lib/version.sh` into both
  `k8s.io/client-go/pkg/version` and `k8s.io/component-base/version`.
- **`sigs/<category>/<project>/default.nix`** — each SIG fetches its own GitHub source and
  builds with `buildGoApplication`.
- **`update.nix`** — factory producing the per-package `updateGomod2nix` shell scripts,
  exposed via `passthru.updateGomod2nix`.

## Conventions specific to this repo

- **Core vendor restore is load-bearing.** `core/default.nix` has a `preConfigure` that copies
  `${src}/vendor/modules.txt` back over the vendor dir. gomod2nix's hook replaces vendor with
  a module-mode layout that lacks the `## workspace` header the k8s go-workspace build needs.
  Do not remove this step.
- **`kubectl` is dynamically linked; every other core binary is static.** This mirrors
  `KUBE_STATIC_BINARIES` in upstream `hack/lib/golang.sh` — the third `mkBin` arg toggles
  `-extldflags '-static'` + `CGO_ENABLED = 0`.
- **Reproducibility pins:** core `ldflags` set `buildDate` to the epoch and `gitTreeState` to
  `clean`; `commit` comes from `hashes.json`.
- **Adding a K8s minor:** add it to `supported` (and possibly `latest`) plus a `kubernetes.<minor>`
  entry in `versions.json`, run fetch-versions + generate-hashes, create
  `core/<minor>/gomod2nix.toml` and each SIG's `gomod2nix.toml`, then add `ATTR_*` and
  `GOMOD2NIX_PKGS` entries to the Makefile.
- **Adding a SIG package:** create `sigs/<category>/<project>/default.nix`, register its base
  path in `releases.nix` (`sigBasePath`) and its build in `mk-release.nix` (`sigs` attrset),
  add the SIG version to relevant `versions.json` entries, generate `gomod2nix.toml`s, and add
  Makefile entries.
- **Formatting:** nixfmt via `nix fmt`. `.editorconfig` enforces final newline + trimmed
  trailing whitespace.
