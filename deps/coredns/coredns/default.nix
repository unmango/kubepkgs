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
  # Matches the upstream Makefile, which builds with cgo off and without the
  # grpc tracing machinery.
  env.CGO_ENABLED = "0";
  tags = [ "grpcnotrace" ];
  # The only injectable symbol. Version comes from coremain/version.go, which
  # the release tag sets, so `coredns -version` reports what the tag says.
  ldflags = [
    "-w"
    "-s"
    "-X github.com/coredns/coredns/coremain.GitCommit=${commit}"
  ];
  meta = with lib; {
    description = "DNS server that chains plugins";
    homepage = "https://coredns.io";
    downloadPage = "https://github.com/coredns/${repo}";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "coredns";
  };
}
