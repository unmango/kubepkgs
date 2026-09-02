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
  pname = repo;
  inherit version src vendorHash;
  subPackages = [ "cmd/metrics-server" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X k8s.io/client-go/pkg/version.gitCommit=${commit}"
  ];
  meta = with lib; {
    description = "Scalable and efficient source of container resource metrics for Kubernetes built-in autoscaling pipelines";
    homepage = "https://github.com/kubernetes-sigs/metrics-server";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "metrics-server";
  };
}
