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
    owner = "kubernetes";
    repo = "kube-state-metrics";
    rev = "v${version}";
    inherit hash;
  };
in
buildGoModule {
  pname = "kube-state-metrics";
  inherit version src vendorHash;
  subPackages = [ "." ];
  doCheck = false;
  ldflags = [
    "-w"
    "-s"
    "-X github.com/prometheus/common/version.Revision=${commit}"
  ];
  passthru.updateScript = nix-update-script { };
  meta = with lib; {
    description = "Add-on agent to generate and expose cluster-level metrics from the Kubernetes API";
    homepage = "https://github.com/kubernetes/kube-state-metrics";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "kube-state-metrics";
  };
}
