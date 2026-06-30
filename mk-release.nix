{
  callPackage,
  fetchFromGitHub,
}:
{
  version,
  srcHash,
  commit,
  modules,
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
    inherit
      version
      src
      modules
      commit
      ;
  };
in
core
// {
  sigs = {
    cluster-api = callPackage ./sigs/cluster-lifecycle/cluster-api sigs.cluster-api;
    cluster-autoscaler = callPackage ./sigs/autoscaling/cluster-autoscaler sigs.cluster-autoscaler;
    kube-state-metrics = callPackage ./sigs/instrumentation/kube-state-metrics sigs.kube-state-metrics;
    metrics-server = callPackage ./sigs/instrumentation/metrics-server sigs.metrics-server;
    external-dns = callPackage ./sigs/network/external-dns sigs.external-dns;
  };
}
