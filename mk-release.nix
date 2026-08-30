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
in
core
// {
  # Each SIG carries its own source location in packages.json, so the set of
  # SIG packages follows the data rather than a hand-maintained attrset.
  sigs = builtins.mapAttrs (
    _: sig: callPackage (./sigs + "/${sig.path}") (builtins.removeAttrs sig [ "path" ])
  ) sigs;
}
