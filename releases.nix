{ lib }:
let
  data = builtins.fromJSON (builtins.readFile ./packages.json);

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
      repo = sig.repo or sig.name;
      tag = (sig.tagPrefix or "v") + version;
      subdir = sig.subdir or "";
      # modRoot rather than subdir for a module depending on its siblings by relative path.
      modRoot = sig.modRoot or "";
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
