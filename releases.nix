{ lib }:
let
  data = builtins.fromJSON (builtins.readFile ./packages.json);

  # mkSig resolves one SIG at one Kubernetes minor into the argument set
  # sigs/<path>/default.nix expects: the version this minor pins, plus the
  # hashes recorded once for that version, plus where to fetch it from.
  mkSig =
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
      repo = sig.name;
    };

  mkEntry =
    minor:
    let
      core = data.kubernetes.${minor};
    in
    {
      inherit (core) version srcHash commit;
      sigs = lib.listToAttrs (map (sig: lib.nameValuePair sig.name (mkSig minor sig)) data.sigs);
    };
in
{
  inherit (data) supported latest;
}
// lib.genAttrs data.supported mkEntry
