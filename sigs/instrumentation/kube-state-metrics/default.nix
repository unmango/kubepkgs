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
    owner = "kubernetes";
    repo = "kube-state-metrics";
    rev = "v${version}";
    inherit hash;
  };
  majorMinor = "${lib.versions.major version}.${lib.versions.minor version}";
in
buildGoApplication {
  pname = "kube-state-metrics";
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
      outdir = "sigs/instrumentation/kube-state-metrics/${majorMinor}";
    };
  };
  meta = with lib; {
    description = "Add-on agent to generate and expose cluster-level metrics from the Kubernetes API";
    homepage = "https://github.com/kubernetes/kube-state-metrics";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "kube-state-metrics";
  };
}
