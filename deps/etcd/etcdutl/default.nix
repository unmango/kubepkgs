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
  pname = "etcdutl";
  inherit
    version
    src
    vendorHash
    modRoot
    ;
  doCheck = false;
  env = {
    CGO_ENABLED = "0";
    GOWORK = "off";
  };
  ldflags = [
    "-w"
    "-s"
    "-X go.etcd.io/etcd/api/v3/version.GitSHA=${commit}"
  ];
  meta = with lib; {
    description = "Administration utility for etcd, operating directly on data files";
    homepage = "https://etcd.io";
    downloadPage = "https://github.com/etcd-io/${repo}";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "etcdutl";
    platforms = platforms.linux;
  };
}
