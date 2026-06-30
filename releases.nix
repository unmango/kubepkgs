{ lib }:
let
  versions = builtins.fromJSON (builtins.readFile ./versions.json);
  hashes = builtins.fromJSON (builtins.readFile ./hashes.json);

  sigBasePath = {
    cluster-api = ./sigs/cluster-lifecycle/cluster-api;
    cluster-autoscaler = ./sigs/autoscaling/cluster-autoscaler;
    kube-state-metrics = ./sigs/instrumentation/kube-state-metrics;
    metrics-server = ./sigs/instrumentation/metrics-server;
    external-dns = ./sigs/network/external-dns;
  };

  mkSig =
    k8sMinor: sigName: sigVersion:
    let
      mm = lib.versions.majorMinor sigVersion;
      sigHashes = hashes.sigs.${sigName}.${k8sMinor};
      basePath = sigBasePath.${sigName};
    in
    {
      version = sigVersion;
      hash = sigHashes.hash;
      commit = sigHashes.commit;
    }
    // lib.optionalAttrs (!(sigHashes ? vendorHash)) {
      modules = basePath + "/${mm}/gomod2nix.toml";
    }
    // lib.optionalAttrs (sigHashes ? vendorHash) {
      vendorHash = sigHashes.vendorHash;
    };

  mkEntry =
    k8sMinor:
    let
      info = versions.kubernetes.${k8sMinor};
      k8sVer = info.version;
      mm = lib.versions.majorMinor k8sVer;
      k8sHashes = hashes.kubernetes.${k8sMinor};
    in
    {
      version = k8sVer;
      commit = k8sHashes.commit;
      srcHash = k8sHashes.srcHash;
      modules = ./core + "/${mm}/gomod2nix.toml";
      sigs = builtins.mapAttrs (mkSig k8sMinor) info.sigs;
    };
in
{
  inherit (versions) supported latest;
}
// lib.genAttrs versions.supported mkEntry
