{ lib }:
let
  data = builtins.fromJSON (builtins.readFile ./packages.json);

  # mkPackage resolves one tracked package at one Kubernetes minor into the
  # argument set <roster>/<path>/default.nix expects: the version this minor pins, plus the
  # hashes recorded once for that version, plus where to fetch it from.
  mkPackage =
    minor: sig:
    let
      version = sig.minors.${minor};
    in
    {
      inherit version;
      inherit (sig.versions.${version})
        srcHash
        commit
        vendorHash
        ;
      inherit (sig) owner path;
      # A SIG's repository, release tag, and location within that repository
      # are all data, so the package definitions stay free of per-project
      # special cases. The defaults cover the common shape.
      repo = sig.repo or sig.name;
      tag = (sig.tagPrefix or "v") + version;
      subdir = sig.subdir or "";
      # subdir narrows the source; modRoot picks a module inside it, which a
      # module depending on its siblings by relative path needs instead.
      modRoot = sig.modRoot or "";
      # null means "whatever nixpkgs defaults to", which is the common case.
      go = sig.go or null;
    };

  mkRoster =
    minor: packages:
    lib.listToAttrs (map (sig: lib.nameValuePair sig.name (mkPackage minor sig)) packages);

  mkEntry =
    minor:
    let
      core = data.kubernetes.${minor};
    in
    {
      inherit (core) version srcHash commit;
      go = core.go or null;
      sigs = mkRoster minor data.sigs;
      deps = mkRoster minor (data.deps or [ ]);
    };
in
{
  inherit (data) supported latest;
}
// lib.genAttrs data.supported mkEntry
