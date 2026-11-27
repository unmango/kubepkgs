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
        // lib.optionalAttrs (sig.modRoot != "") { inherit (sig) modRoot; }
      )
    );
in
core
// {
  # Drop-in for nixpkgs' `kubernetes`, the shape services.kubernetes.package expects, with `pause` as an attribute rather than in bin/.
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
