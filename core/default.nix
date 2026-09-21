{
  buildGoModule,
  stdenv,
  lib,
  makeWrapper,
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
    buildGoModule ({
      inherit pname version src;

      # K8s ships a complete vendor/ dir, generated in workspace mode (vendor/modules.txt
      # starts with "## workspace"). Leave GOWORK on so -mod=vendor resolves against it;
      # forcing GOWORK=off makes go's vendor consistency check fail against the workspace-style
      # vendor/modules.txt (replace directives get flagged as "not marked as replaced").
      vendorHash = null;

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

      meta = coreMeta // extraMeta;
      env = lib.optionalAttrs static { CGO_ENABLED = "0"; };
    });
  coreMeta = {
    homepage = "https://kubernetes.io";
    license = lib.licenses.asl20;
    maintainers = with lib.maintainers; [ UnstoppableMango ];
  };

  # kubectl is NOT in KUBE_STATIC_BINARIES — dynamically linked on Linux. It is
  # the one core binary with a use off a cluster node, so it carries no platform
  # restriction. Bound here so kube-addons can point the addon manager at it.
  kubectlBin = mkBin "kubectl" "cmd/kubectl" false {
    description = "Run commands against Kubernetes clusters";
    mainProgram = "kubectl";
  };
in
{
  kubectl = kubectlBin;
  # Remaining binaries are in KUBE_STATIC_BINARIES per hack/lib/golang.sh. They
  # are node and control-plane components, so they are offered on Linux only,
  # except kube-apiserver.
  kubeadm = mkBin "kubeadm" "cmd/kubeadm" true {
    description = "Bootstrap a Kubernetes cluster";
    mainProgram = "kubeadm";
    platforms = lib.platforms.linux;
  };
  kubelet = mkBin "kubelet" "cmd/kubelet" true {
    description = "Primary node agent for Kubernetes";
    mainProgram = "kubelet";
    platforms = lib.platforms.linux;
  };
  kube-apiserver = mkBin "kube-apiserver" "cmd/kube-apiserver" true {
    description = "Kubernetes API server";
    mainProgram = "kube-apiserver";
    # darwin carries it for controller-runtime's envtest, as setup-envtest's
    # own darwin archives do.
    platforms = lib.platforms.linux ++ lib.platforms.darwin;
  };
  kube-controller-manager = mkBin "kube-controller-manager" "cmd/kube-controller-manager" true {
    description = "Kubernetes controller manager";
    mainProgram = "kube-controller-manager";
    platforms = lib.platforms.linux;
  };
  kube-scheduler = mkBin "kube-scheduler" "cmd/kube-scheduler" true {
    description = "Kubernetes cluster scheduler";
    mainProgram = "kube-scheduler";
    platforms = lib.platforms.linux;
  };
  kube-proxy = mkBin "kube-proxy" "cmd/kube-proxy" true {
    description = "Kubernetes network proxy";
    mainProgram = "kube-proxy";
    platforms = lib.platforms.linux;
  };

  # The sandbox shim is a small C program rather than a Go binary, so it is
  # built with stdenv straight from build/pause. The flags come from
  # build/pause/Makefile minus its -static, which needs a static libc the
  # default stdenv does not carry; nixpkgs links its pause dynamically too.
  pause = stdenv.mkDerivation {
    pname = "pause";
    inherit version src;

    dontConfigure = true;

    buildPhase = ''
      runHook preBuild
      $CC -Os -Wall -Werror -DVERSION=v${version} -o pause build/pause/linux/pause.c
      runHook postBuild
    '';

    installPhase = ''
      runHook preInstall
      install -D pause -t $out/bin
      runHook postInstall
    '';

    meta = coreMeta // {
      description = "Kubernetes pod sandbox shim";
      mainProgram = "pause";
      platforms = lib.platforms.linux;
    };
  };

  # The addon manager is a pair of shell scripts. kube-addons-main.sh sources
  # kube-addons.sh from the working directory or /opt, so the wrapper runs it
  # from the directory holding both instead of patching the lookup. Both read
  # KUBECTL_BIN for the kubectl to drive.
  kube-addons = stdenv.mkDerivation {
    pname = "kube-addons";
    inherit version src;

    nativeBuildInputs = [ makeWrapper ];

    dontConfigure = true;
    dontBuild = true;

    installPhase = ''
      runHook preInstall
      dir=$out/libexec/kube-addons
      install -Dm755 cluster/addons/addon-manager/kube-addons-main.sh -t $dir
      install -Dm644 cluster/addons/addon-manager/kube-addons.sh -t $dir
      patchShebangs $dir
      makeWrapper $dir/kube-addons-main.sh $out/bin/kube-addons \
        --chdir $dir \
        --set KUBECTL_BIN ${lib.getExe kubectlBin}
      runHook postInstall
    '';

    meta = coreMeta // {
      description = "Kubernetes addon manager";
      mainProgram = "kube-addons";
      platforms = lib.platforms.linux;
    };
  };
}
