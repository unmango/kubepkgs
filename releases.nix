{
  supported = [
    "1.33"
    "1.34"
    "1.35"
    "1.36"
  ];

  latest = "1.36";

  "1.33" = {
    srcHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
    modules = ./core/1.33/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.8.3";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.8/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.13.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/kube-state-metrics/2.13/gomod2nix.toml;
      };
      metrics-server = {
        version = "0.7.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/metrics-server/0.7/gomod2nix.toml;
      };
      external-dns = {
        version = "0.14.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/network/external-dns/0.14/gomod2nix.toml;
      };
    };
  };

  "1.34" = {
    srcHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
    modules = ./core/1.34/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.9.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.9/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.13.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/kube-state-metrics/2.13/gomod2nix.toml;
      };
      metrics-server = {
        version = "0.7.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/metrics-server/0.7/gomod2nix.toml;
      };
      external-dns = {
        version = "0.15.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/network/external-dns/0.15/gomod2nix.toml;
      };
    };
  };

  "1.35" = {
    srcHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
    modules = ./core/1.35/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.9.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.9/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.14.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/kube-state-metrics/2.14/gomod2nix.toml;
      };
      metrics-server = {
        version = "0.7.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/metrics-server/0.7/gomod2nix.toml;
      };
      external-dns = {
        version = "0.15.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/network/external-dns/0.15/gomod2nix.toml;
      };
    };
  };

  "1.36" = {
    srcHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
    modules = ./core/1.36/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.10.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.10/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.14.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/kube-state-metrics/2.14/gomod2nix.toml;
      };
      metrics-server = {
        version = "0.7.2";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/instrumentation/metrics-server/0.7/gomod2nix.toml;
      };
      external-dns = {
        version = "0.15.0";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        modules = ./sigs/network/external-dns/0.15/gomod2nix.toml;
      };
    };
  };
}
