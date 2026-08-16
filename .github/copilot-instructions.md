# Copilot instructions for kubepkgs

Nix flake exposing versioned Kubernetes package sets. Each Kubernetes minor version gets a
package set with core binaries (kubectl, kubelet, kube-apiserver, kube-controller-manager,
kube-scheduler, kube-proxy, kubeadm) plus selected SIG projects (cluster-api,
kube-state-metrics, metrics-server, external-dns), all built with `buildGoModule` against
pinned source/vendor hashes in `hashes.json`. (`gomod2nix`/`buildGoApplication` is used
separately, only to package this repo's own `tools/update` Go CLI — see below.)

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

There is no unit test suite for the K8s core/SIG packages themselves; `doCheck = false` for
all core + SIG `buildGoModule` packages. "Testing" one of those means building it. (The
`tools/update` Go CLI has its own ginkgo/gomega unit test suite — see below.) CI
(`.github/workflows/ci.yml`) runs `nix flake check` then `nix develop -c make build-all` on
every PR/push to `main`.

## Dependency / version update pipeline

Versions and hashes are data-driven, not hand-edited into `.nix` files, and the pipeline that
maintains them is implemented by a Go CLI at `tools/update` (packaged via
`gomod2nix`/`buildGoApplication`, exposed as the flake package `update`) rather than shell
scripts:

- `versions.json` — tracked minor versions (`supported`, `latest`) and the resolved
  core + SIG semver for each. This is the file you edit by hand to add/track a version.
- `hashes.json` — generated `srcHash`/`commit`/`vendorHash` per package. Marked "Do not edit
  manually".
- `make fetch-versions` (`nix run .#update -- fetch-versions`) — bumps patch versions in
  `versions.json` from GitHub releases.
- `make generate-hashes` (`nix run .#update -- generate-hashes`, or automatically as a
  `hashes.json: versions.json` file-target dependency of the vendor-hash targets below) —
  populates `hashes.json` via `nix-prefetch-github` + the GitHub API.
- `make update-vendor-hash PKG=<sig>` (`nix run .#update -- vendor-hashes --sig <sig>`) —
  resolves the real `vendorHash` for one SIG's Go module, across every distinct version it
  has among the supported minors (grouping is derived automatically from `versions.json`,
  not hand-maintained). `make update-all-vendor-hashes` does every SIG.
- `make update-releases` — runs fetch-versions → update-all-vendor-hashes (which pulls in
  generate-hashes as a prerequisite if `versions.json` changed).

Valid `PKG` values: `cluster-api`, `kube-state-metrics`, `metrics-server`, `external-dns`.
For a single SIG-version group rather than all of a SIG's groups, use
`nix run .#update -- vendor-hashes --sig <sig> --minor <build-minor>` directly.

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
  (gnumake, nixfmt, nix-prefetch-github, go, the `gomod2nix` CLI), and exposes the `update`
  package (`nix/updater.nix`, the `tools/update` Go CLI, built via `gomod2nix`'s
  `buildGoApplication`).
- **`core/default.nix`** — builds all core binaries via a shared `mkBin` helper using
  `buildGoModule` with `vendorHash = null` against K8s's own vendored `vendor/` dir. Injects
  version info through `ldflags` mirroring `hack/lib/version.sh` into both
  `k8s.io/client-go/pkg/version` and `k8s.io/component-base/version`.
- **`sigs/<category>/<project>/default.nix`** — each SIG fetches its own GitHub source via
  `fetchFromGitHub` and builds with `buildGoModule` against a real `vendorHash` from
  `hashes.json`.
- **`nix/updater.nix`** — `buildGoApplication` derivation for the `tools/update` Go CLI;
  `src` is filtered to just `go.mod`/`go.sum`/`**/*.go`/`**/testdata/**` via the `globset`
  flake input + `lib.fileset.toSource`, so unrelated file changes don't trigger a rebuild.

## Conventions specific to this repo

- **`core/default.nix`'s `GOWORK` must stay on (the default).** K8s ships a complete
  `vendor/` dir generated in workspace mode (`vendor/modules.txt` starts with
  `## workspace`). Forcing `GOWORK = "off"` breaks Go's vendor consistency check against
  that workspace-style `modules.txt` (replace directives get flagged as "not marked as
  replaced"). Don't reintroduce it.
- **`kubectl` is dynamically linked; every other core binary is static.** This mirrors
  `KUBE_STATIC_BINARIES` in upstream `hack/lib/golang.sh` — the third `mkBin` arg toggles
  `-extldflags '-static'` + `CGO_ENABLED = 0`.
- **Reproducibility pins:** core `ldflags` set `buildDate` to the epoch and `gitTreeState` to
  `clean`; `commit` comes from `hashes.json`.
- **Adding a K8s minor:** add it to `supported` (and possibly `latest`) plus a `kubernetes.<minor>`
  entry in `versions.json`, run `make generate-hashes` + `make update-all-vendor-hashes` (or
  `make update-releases` for the full fetch/generate/hash flow — vendor-hash grouping is
  derived automatically from `versions.json`, no per-minor mapping to hand-maintain), then
  add `ATTR_*` and a `CORE_PKGS` entry to the Makefile.
- **Adding a SIG package:** create `sigs/<category>/<project>/default.nix`, wire it into
  `mk-release.nix`'s `sigs` attrset, add `sigs.<project>` entries to relevant
  `versions.json` minors, then add `ATTR_*` and `VENDOR_HASH_PKGS` entries to the Makefile
  (used by `build-<pkg>`; vendor-hash resolution itself needs no per-package Makefile entry).
- **Formatting:** nixfmt via `nix fmt`. `.editorconfig` enforces final newline + trimmed
  trailing whitespace.
