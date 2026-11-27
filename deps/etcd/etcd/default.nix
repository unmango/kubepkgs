{
  buildGoModule,
  lib,
  version,
  commit,
  src,
  vendorHash,
  modRoot,
  repo,
}:
buildGoModule {
  pname = "etcd";
  inherit
    version
    src
    vendorHash
    modRoot
    ;
  doCheck = false;
  env = {
    # Matches scripts/build_lib.sh, which builds every etcd binary with cgo off.
    CGO_ENABLED = "0";
    # etcd 3.7 ships a root go.work above modRoot that would change module resolution under the pinned vendorHash.
    GOWORK = "off";
  };
  preInstall = ''
    mv "$GOPATH/bin/server" "$GOPATH/bin/etcd"
  '';
  # Version is baked in at release time and not overridden, so `etcd --version` reports what the tag says.
  ldflags = [
    "-w"
    "-s"
    "-X go.etcd.io/etcd/api/v3/version.GitSHA=${commit}"
  ];
  meta = with lib; {
    description = "Distributed reliable key-value store for the most critical data of a distributed system";
    homepage = "https://etcd.io";
    downloadPage = "https://github.com/etcd-io/${repo}";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "etcd";
    # darwin carries the server alone, for controller-runtime's envtest.
    platforms = platforms.linux ++ platforms.darwin;
  };
}
