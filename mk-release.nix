{
  callPackage,
  fetchFromGitHub,
}:
{
  version,
  srcHash,
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
  core = callPackage ./core { inherit version src modules; };
in
core
// {
  sigs = {
    cluster-api = callPackage ./sigs/cluster-lifecycle/cluster-api sigs.cluster-api;
    kube-state-metrics = callPackage ./sigs/instrumentation/kube-state-metrics sigs.kube-state-metrics;
    metrics-server = callPackage ./sigs/instrumentation/metrics-server sigs.metrics-server;
    external-dns = callPackage ./sigs/network/external-dns sigs.external-dns;
  };
}
