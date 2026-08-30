{
  buildGoModule,
  fetchFromGitHub,
  lib,
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
    "-X github.com/prometheus/common/version.Revision=${commit}"
  ];
  meta = with lib; {
    description = "Add-on agent to generate and expose cluster-level metrics from the Kubernetes API";
    homepage = "https://github.com/kubernetes/kube-state-metrics";
    license = licenses.asl20;
    maintainers = with maintainers; [ UnstoppableMango ];
    mainProgram = "kube-state-metrics";
  };
}
