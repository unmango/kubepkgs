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
    # etcd 3.7 ships a root go.work above modRoot, which would otherwise change
    # module resolution out from under the pinned vendorHash. Unlike core, which
    # deliberately leaves GOWORK on for Kubernetes' workspace-mode vendor tree,
    # etcd vendors nothing and each module resolves on its own.
    GOWORK = "off";
  };
  # The server module's main package is the module root, so the binary lands as
  # `server`.
  preInstall = ''
    mv "$GOPATH/bin/server" "$GOPATH/bin/etcd"
  '';
  # The only injectable symbol. Version is baked into api/version/version.go at
  # release time and is deliberately not overridden here, so `etcd --version`
  # always reports what the tag actually says.
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
