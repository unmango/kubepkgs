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
  subPackages = [ "cmd/clusterctl" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X sigs.k8s.io/cluster-api/version.gitMajor=${lib.versions.major version}"
    "-X sigs.k8s.io/cluster-api/version.gitMinor=${lib.versions.minor version}"
    "-X sigs.k8s.io/cluster-api/version.gitVersion=v${version}"
    "-X sigs.k8s.io/cluster-api/version.gitCommit=${commit}"
  ];
  meta = with lib; {
    description = "Declarative APIs and tooling for provisioning, upgrading, and operating Kubernetes clusters";
    homepage = "https://cluster-api.sigs.k8s.io";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "clusterctl";
    platforms = platforms.linux;
  };
}
