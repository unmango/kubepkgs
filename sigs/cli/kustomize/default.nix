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
  subPackages = [ "." ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X sigs.k8s.io/kustomize/api/provenance.version=v${version}"
    "-X sigs.k8s.io/kustomize/api/provenance.gitCommit=${commit}"
    "-X sigs.k8s.io/kustomize/api/provenance.buildDate=1970-01-01T00:00:00Z"
  ];
  meta = with lib; {
    description = "Customizes Kubernetes YAML configurations without templates";
    homepage = "https://kustomize.io";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "kustomize";
  };
}
