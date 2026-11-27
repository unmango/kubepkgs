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
  # Mirrors kube::version::ldflags in hack/lib/version.sh, with gitCommit empty and buildDate at epoch for reproducibility.
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

      # vendor/modules.txt is workspace-generated, so GOWORK stays on; GOWORK=off fails go's vendor consistency check ("not marked as replaced").
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

  # kubectl is dynamically linked on Linux and useful off a cluster node, so it has no platform restriction.
  kubectlBin = mkBin "kubectl" "cmd/kubectl" false {
    description = "Run commands against Kubernetes clusters";
    mainProgram = "kubectl";
  };
in
{
  kubectl = kubectlBin;
  # Node and control-plane components, so Linux only, except kube-apiserver.
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
    # darwin carries it for controller-runtime's envtest, as setup-envtest's darwin archives do.
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

  # build/pause/Makefile's flags minus -static, which needs a static libc the default stdenv lacks; nixpkgs links pause dynamically too.
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

  # kube-addons-main.sh sources kube-addons.sh from the working directory or /opt, so the wrapper runs from the directory holding both rather than patching the lookup.
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
