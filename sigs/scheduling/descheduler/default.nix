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
  # descheduler ships a vendor/ dir, which buildGoModule refuses to build
  # against alongside a vendorHash. Dropping it keeps every SIG on the same
  # resolved-vendorHash model rather than special-casing this one.
  deleteVendor = true;
  subPackages = [ "cmd/descheduler" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X sigs.k8s.io/descheduler/pkg/version.version=v${version}"
    "-X sigs.k8s.io/descheduler/pkg/version.gitsha1=${commit}"
    "-X sigs.k8s.io/descheduler/pkg/version.buildDate=1970-01-01T00:00:00Z"
  ];
  meta = with lib; {
    description = "Evicts pods so the scheduler can place them on more suitable nodes";
    homepage = "https://sigs.k8s.io/descheduler";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "descheduler";
  };
}
