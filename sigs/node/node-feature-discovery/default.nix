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
  subPackages = [ "cmd/nfd-master" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X sigs.k8s.io/node-feature-discovery/pkg/version.version=v${version}"
  ];
  meta = with lib; {
    description = "Detects hardware features and configuration and labels nodes accordingly";
    homepage = "https://kubernetes-sigs.github.io/node-feature-discovery";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "nfd-master";
  };
}
