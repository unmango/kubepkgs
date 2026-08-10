{ lib }:
let
  versions = builtins.fromJSON (builtins.readFile ./versions.json);
  hashes = builtins.fromJSON (builtins.readFile ./hashes.json);

  mkSig =
    k8sMinor: sigName: sigVersion:
    let
      sigHashes = hashes.sigs.${sigName}.${k8sMinor};
    in
    {
      version = sigVersion;
      srcHash = sigHashes.srcHash;
      commit = sigHashes.commit;
      vendorHash = sigHashes.vendorHash;
    };

  mkEntry =
    k8sMinor:
    let
      info = versions.kubernetes.${k8sMinor};
      k8sVer = info.version;
      k8sHashes = hashes.kubernetes.${k8sMinor};
    in
    {
      version = k8sVer;
      commit = k8sHashes.commit;
      srcHash = k8sHashes.srcHash;
      sigs = builtins.mapAttrs (mkSig k8sMinor) info.sigs;
    };
in
{
  inherit (versions) supported latest;
}
// lib.genAttrs versions.supported mkEntry
