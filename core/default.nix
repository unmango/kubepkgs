{
  buildGoApplication,
  lib,
  mkGomod2nixUpdater,
  nix-update-script,
  version,
  src,
  modules,
}:
let
  updateGomod2nix = mkGomod2nixUpdater {
    inherit src;
    outdir = "core/${version}";
  };
  mkBin =
    pname: subPkg: extraMeta:
    buildGoApplication {
      inherit
        pname
        version
        src
        modules
        ;

      subPackages = [ subPkg ];
      doCheck = false;
      ldflags = [
        "-w"
        "-s"
      ];

      passthru = {
        updateScript = nix-update-script { };
        inherit updateGomod2nix;
      };

      meta =
        with lib;
        {
          homepage = "https://kubernetes.io";
          license = licenses.asl20;
          maintainers = with maintainers; [ UnstoppableMango ];
        }
        // extraMeta;
    };
in
{
  kubectl = mkBin "kubectl" "cmd/kubectl" {
    description = "Run commands against Kubernetes clusters";
    mainProgram = "kubectl";
  };
  kubeadm = mkBin "kubeadm" "cmd/kubeadm" {
    description = "Bootstrap a Kubernetes cluster";
    mainProgram = "kubeadm";
  };
  kubelet = mkBin "kubelet" "cmd/kubelet" {
    description = "Primary node agent for Kubernetes";
    mainProgram = "kubelet";
  };
  kube-apiserver = mkBin "kube-apiserver" "cmd/kube-apiserver" {
    description = "Kubernetes API server";
    mainProgram = "kube-apiserver";
  };
  kube-controller-manager = mkBin "kube-controller-manager" "cmd/kube-controller-manager" {
    description = "Kubernetes controller manager";
    mainProgram = "kube-controller-manager";
  };
  kube-scheduler = mkBin "kube-scheduler" "cmd/kube-scheduler" {
    description = "Kubernetes cluster scheduler";
    mainProgram = "kube-scheduler";
  };
  kube-proxy = mkBin "kube-proxy" "cmd/kube-proxy" {
    description = "Kubernetes network proxy";
    mainProgram = "kube-proxy";
  };
}
