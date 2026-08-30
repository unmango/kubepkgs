{
  buildGoModule,
  fetchFromGitHub,
  lib,
  nix-update-script,
  version,
  commit,
  srcHash,
  vendorHash,
  owner,
  repo,
}:
let
  src = fetchFromGitHub {
    inherit owner repo;
    rev = "v${version}";
    hash = srcHash;
  };
in
buildGoModule {
  pname = repo;
  inherit version src vendorHash;
  subPackages = [ "." ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
  ];
  passthru.updateScript = nix-update-script { };
  meta = with lib; {
    description = "Configure external DNS servers dynamically from Kubernetes resources";
    homepage = "https://github.com/kubernetes-sigs/external-dns";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "external-dns";
  };
}
