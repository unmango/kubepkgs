{
  buildGoApplication,
  fetchFromGitHub,
  lib,
  nix-update-script,
  version,
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
in
buildGoApplication {
  pname = "metrics-server";
  inherit version src modules;
  subPackages = [ "cmd/metrics-server" ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
  ];
  passthru.updateScript = nix-update-script { };
  meta = with lib; {
    description = "Scalable and efficient source of container resource metrics for Kubernetes built-in autoscaling pipelines";
    homepage = "https://github.com/kubernetes-sigs/metrics-server";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "metrics-server";
  };
}
