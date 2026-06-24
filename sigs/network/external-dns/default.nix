{
  buildGoApplication,
  fetchFromGitHub,
  lib,
  mkGomod2nixUpdater,
  nix-update-script,
  version,
  hash,
  modules,
}:
let
  src = fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "external-dns";
    rev = "v${version}";
    inherit hash;
  };
  majorMinor = "${lib.versions.major version}.${lib.versions.minor version}";
in
buildGoApplication {
  pname = "external-dns";
  inherit version src modules;
  subPackages = [ "." ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
  ];
  passthru = {
    updateScript = nix-update-script { };
    updateGomod2nix = mkGomod2nixUpdater {
      inherit src;
      outdir = "sigs/network/external-dns/${majorMinor}";
    };
  };
  meta = with lib; {
    description = "Configure external DNS servers dynamically from Kubernetes resources";
    homepage = "https://github.com/kubernetes-sigs/external-dns";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "external-dns";
  };
}
