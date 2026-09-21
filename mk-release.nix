{
  callPackage,
  fetchFromGitHub,
  buildGoModule,
  lib,
  pkgs,
}:
{
  version,
  srcHash,
  commit,
  go ? null,
  sigs,
  deps ? { },
}:
let
  resolveGo = callPackage ./nix/go-version.nix { inherit pkgs; };

  # A record pinning a Go version gets a buildGoModule bound to it; everything
  # else gets the default, so the pin stays invisible to the packages that do
  # not need one.
  goModuleFor =
    label: requested:
    let
      go = resolveGo label requested;
    in
    if go == null then buildGoModule else buildGoModule.override { inherit go; };

  src = fetchFromGitHub {
    owner = "kubernetes";
    repo = "kubernetes";
    rev = "v${version}";
    hash = srcHash;
  };
  core = callPackage ./core {
    inherit version src commit;
    buildGoModule = goModuleFor "kubernetes ${lib.versions.majorMinor version}" go;
  };

  # Where a package's source comes from is data, so the definitions receive a
  # ready source the way core does and stay free of fetching concerns. subdir
  # narrows the result for projects that share a repository with their siblings.
  packageSource =
    sig:
    let
      repo = fetchFromGitHub {
        inherit (sig) owner repo;
        rev = sig.tag;
        hash = sig.srcHash;
      };
    in
    if sig.subdir == "" then repo else "${repo}/${sig.subdir}";
  # Each package carries its own source location in packages.json, so a roster
  # follows the data rather than a hand-maintained attrset. sigs and deps
  # differ only in which directory holds the definitions.
  mkRoster =
    dir:
    builtins.mapAttrs (
      name: sig:
      callPackage (dir + "/${sig.path}") (
        builtins.removeAttrs sig [
          "path"
          "owner"
          "tag"
          "srcHash"
          "subdir"
          "modRoot"
          "go"
        ]
        // {
          src = packageSource sig;
          buildGoModule = goModuleFor "${name} ${sig.version}" sig.go;
        }
        # Only the packages that declare a modRoot receive one, so the
        # definitions that build from the source root keep the smaller
        # argument set.
        // lib.optionalAttrs (sig.modRoot != "") { inherit (sig) modRoot; }
      )
    );
in
core
// {
  # Drop-in for nixpkgs' `kubernetes`, the shape services.kubernetes.package
  # expects: every server binary plus kube-addons under one bin/, and the
  # sandbox shim as a `pause` attribute rather than in bin/.
  kubernetes = pkgs.symlinkJoin {
    name = "kubernetes-${version}";
    paths = lib.attrValues (lib.filterAttrs (_: lib.isDerivation) (removeAttrs core [ "pause" ]));
    passthru = {
      inherit version;
      inherit (core) pause;
    };
    meta = {
      description = "Kubernetes core binaries in the layout of nixpkgs' kubernetes package";
      homepage = "https://kubernetes.io";
      license = lib.licenses.asl20;
      platforms = lib.platforms.linux;
    };
  };

  sigs = mkRoster ./sigs sigs;
  deps = mkRoster ./deps deps;
}
