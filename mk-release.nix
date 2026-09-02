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

  # Where a SIG's source comes from is data, so the package definitions receive
  # a ready source the way core does and stay free of fetching concerns. subdir
  # narrows the result for projects that share a repository with their siblings.
  sigSource =
    sig:
    let
      repo = fetchFromGitHub {
        inherit (sig) owner repo;
        rev = sig.tag;
        hash = sig.srcHash;
      };
    in
    if sig.subdir == "" then repo else "${repo}/${sig.subdir}";
in
core
// {
  # Each SIG carries its own source location in packages.json, so the set of
  # SIG packages follows the data rather than a hand-maintained attrset.
  sigs = builtins.mapAttrs (
    name: sig:
    callPackage (./sigs + "/${sig.path}") (
      builtins.removeAttrs sig [
        "path"
        "owner"
        "tag"
        "srcHash"
        "subdir"
        "go"
      ]
      // {
        src = sigSource sig;
        buildGoModule = goModuleFor "${name} ${sig.version}" sig.go;
      }
    )
  ) sigs;
}
