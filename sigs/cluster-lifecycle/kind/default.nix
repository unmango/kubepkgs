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
    # versionCore is a const, so it is baked in by the tag and not overridable;
    # only the commit is injectable here.
    "-X sigs.k8s.io/kind/pkg/cmd/kind/version.gitCommit=${commit}"
  ];
  meta = with lib; {
    description = "Runs local Kubernetes clusters using Docker container nodes";
    homepage = "https://kind.sigs.k8s.io";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "kind";
    platforms = platforms.linux;
  };
}
