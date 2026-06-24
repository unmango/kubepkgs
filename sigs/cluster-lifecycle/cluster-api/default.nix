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
    repo = "cluster-api";
    rev = "v${version}";
    inherit hash;
  };
  majorMinor = "${lib.versions.major version}.${lib.versions.minor version}";
in
buildGoApplication {
  pname = "cluster-api";
  inherit version src modules;
  subPackages = [ "cmd/clusterctl" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X sigs.k8s.io/cluster-api/version.gitMajor=${lib.versions.major version}"
    "-X sigs.k8s.io/cluster-api/version.gitMinor=${lib.versions.minor version}"
    "-X sigs.k8s.io/cluster-api/version.gitVersion=v${version}"
  ];
  passthru = {
    updateScript = nix-update-script { };
    updateGomod2nix = mkGomod2nixUpdater {
      inherit src;
      outdir = "sigs/cluster-lifecycle/cluster-api/${majorMinor}";
    };
  };
  meta = with lib; {
    description = "Declarative APIs and tooling for provisioning, upgrading, and operating Kubernetes clusters";
    homepage = "https://cluster-api.sigs.k8s.io";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "clusterctl";
  };
}
