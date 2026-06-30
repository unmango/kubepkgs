{
  buildGoApplication,
  fetchFromGitHub,
  lib,
  mkGomod2nixUpdater,
  nix-update-script,
  version,
  commit,
  hash,
  modules,
}:
let
  src = fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "metrics-server";
    rev = "v${version}";
    inherit hash;
  };
  majorMinor = "${lib.versions.major version}.${lib.versions.minor version}";
in
buildGoApplication {
  pname = "metrics-server";
  inherit version src modules;
  subPackages = [ "cmd/metrics-server" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X k8s.io/client-go/pkg/version.gitCommit=${commit}"
  ];
  passthru = {
    updateScript = nix-update-script { };
    updateGomod2nix = mkGomod2nixUpdater {
      inherit src;
      outdir = "sigs/instrumentation/metrics-server/${majorMinor}";
    };
  };
  meta = with lib; {
    description = "Scalable and efficient source of container resource metrics for Kubernetes built-in autoscaling pipelines";
    homepage = "https://github.com/kubernetes-sigs/metrics-server";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "metrics-server";
  };
}
