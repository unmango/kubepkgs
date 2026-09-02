{
  buildGoModule,
  lib,
  version,
  commit,
  src,
  vendorHash,
  repo,
}:
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
  meta = with lib; {
    description = "Automatically adjusts the size of a Kubernetes cluster based on the utilization of Pods";
    homepage = "https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "cluster-autoscaler";
  };
}
