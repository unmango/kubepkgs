{
  buildGoModule,
  lib,
  nix-update-script,
  version,
  commit,
  src,
}:
let
  # Mirrors hack/lib/version.sh kube::version::ldflags — injects version info
  # into both k8s.io/client-go/pkg/version and k8s.io/component-base/version.
  # gitCommit left empty and buildDate pinned to epoch for reproducibility.
  versionLdflags =
    let
      xFlag =
        pkg: key: val:
        "-X '${pkg}.${key}=${val}'";
      both = key: val: [
        (xFlag "k8s.io/client-go/pkg/version" key val)
        (xFlag "k8s.io/component-base/version" key val)
      ];
    in
    (both "gitVersion" "v${version}")
    ++ (both "gitMajor" (lib.versions.major version))
    ++ (both "gitMinor" (lib.versions.minor version))
    ++ (both "gitCommit" commit)
    ++ (both "gitTreeState" "clean")
    ++ (both "buildDate" "1970-01-01T00:00:00Z");

  mkBin =
    pname: subPkg: static: extraMeta:
    buildGoModule (
      {
        inherit pname version src;

        # K8s ships a complete vendor/ dir. Use it directly; no download needed.
        # GOWORK=off: go.work at repo root conflicts with -mod=vendor in Go 1.22+.
        vendorHash = null;
        GOWORK = "off";

        subPackages = [ subPkg ];
        doCheck = false;
        ldflags = [
          "-w"
          "-s"
        ]
        ++ versionLdflags
        ++ lib.optionals static [
          "-extldflags '-static'"
          "-installsuffix static"
        ];

        passthru.updateScript = nix-update-script { };

        meta =
          with lib;
          {
            homepage = "https://kubernetes.io";
            license = licenses.asl20;
            maintainers = with maintainers; [ UnstoppableMango ];
          }
          // extraMeta;
        env = lib.optionalAttrs static { CGO_ENABLED = "0"; };
      }
    );
in
{
  # kubectl is NOT in KUBE_STATIC_BINARIES — dynamically linked on Linux.
  kubectl = mkBin "kubectl" "cmd/kubectl" false {
    description = "Run commands against Kubernetes clusters";
    mainProgram = "kubectl";
  };
  # Remaining binaries are in KUBE_STATIC_BINARIES per hack/lib/golang.sh.
  kubeadm = mkBin "kubeadm" "cmd/kubeadm" true {
    description = "Bootstrap a Kubernetes cluster";
    mainProgram = "kubeadm";
  };
  kubelet = mkBin "kubelet" "cmd/kubelet" true {
    description = "Primary node agent for Kubernetes";
    mainProgram = "kubelet";
  };
  kube-apiserver = mkBin "kube-apiserver" "cmd/kube-apiserver" true {
    description = "Kubernetes API server";
    mainProgram = "kube-apiserver";
  };
  kube-controller-manager = mkBin "kube-controller-manager" "cmd/kube-controller-manager" true {
    description = "Kubernetes controller manager";
    mainProgram = "kube-controller-manager";
  };
  kube-scheduler = mkBin "kube-scheduler" "cmd/kube-scheduler" true {
    description = "Kubernetes cluster scheduler";
    mainProgram = "kube-scheduler";
  };
  kube-proxy = mkBin "kube-proxy" "cmd/kube-proxy" true {
    description = "Kubernetes network proxy";
    mainProgram = "kube-proxy";
  };
}
