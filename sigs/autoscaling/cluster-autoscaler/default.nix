{
  buildGoModule,
  fetchFromGitHub,
  lib,
  nix-update-script,
  version,
  commit,
  hash,
  modules ? null,
  vendorHash,
}:
let
  repo = fetchFromGitHub {
    owner = "kubernetes";
    repo = "autoscaler";
    rev = "cluster-autoscaler-${version}";
    inherit hash;
  };
  src = "${repo}/cluster-autoscaler";
  majorMinor = "${lib.versions.major version}.${lib.versions.minor version}";
in
buildGoModule {
  pname = "cluster-autoscaler";
  inherit version src vendorHash;
  subPackages = [ "." ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X k8s.io/autoscaler/cluster-autoscaler/version.ClusterAutoscalerVersion=${version}"
  ];
  passthru = {
    updateScript = nix-update-script { };
  };
  meta = with lib; {
    description = "Automatically adjusts the size of a Kubernetes cluster based on the utilization of Pods";
    homepage = "https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "cluster-autoscaler";
  };
}
