{
  buildGoModule,
  fetchFromGitHub,
  lib,
  nix-update-script,
  version,
  commit,
  hash,
  vendorHash,
}:
let
  src = fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "external-dns";
    rev = "v${version}";
    inherit hash;
  };
in
buildGoModule {
  pname = "external-dns";
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
