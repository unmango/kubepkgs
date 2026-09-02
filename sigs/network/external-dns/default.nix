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
  ];
  meta = with lib; {
    description = "Configure external DNS servers dynamically from Kubernetes resources";
    homepage = "https://github.com/kubernetes-sigs/external-dns";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "external-dns";
  };
}
