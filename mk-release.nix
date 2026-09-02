{
  callPackage,
  fetchFromGitHub,
}:
{
  version,
  srcHash,
  commit,
  sigs,
}:
let
  src = fetchFromGitHub {
    owner = "kubernetes";
    repo = "kubernetes";
    rev = "v${version}";
    hash = srcHash;
  };
  core = callPackage ./core {
    inherit version src commit;
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
    _: sig:
    callPackage (./sigs + "/${sig.path}") (
      builtins.removeAttrs sig [
        "path"
        "owner"
        "tag"
        "srcHash"
        "subdir"
      ]
      // {
        src = sigSource sig;
      }
    )
  ) sigs;
}
