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
  subPackages = [ "cmd/secrets-store-csi-driver" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X sigs.k8s.io/secrets-store-csi-driver/pkg/version.BuildVersion=v${version}"
    "-X sigs.k8s.io/secrets-store-csi-driver/pkg/version.Vcs=${commit}"
    "-X sigs.k8s.io/secrets-store-csi-driver/pkg/version.BuildTime=1970-01-01T00:00:00Z"
  ];
  meta = with lib; {
    description = "Mounts secrets from external secret stores into Kubernetes pods as volumes";
    homepage = "https://secrets-store-csi-driver.sigs.k8s.io";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "secrets-store-csi-driver";
  };
}
