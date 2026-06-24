{
  supported = [
    "1.33"
    "1.34"
    "1.35"
    "1.36"
  ];

  latest = "1.36";

  "1.33" = {
    version = "1.33.0";
    srcHash = "sha256-5MlMBsYf8V7BvV6xaeRMVSRaE+TpG8xJkMwVGm/fVdo=";
    modules = ./core/1.33/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.8.3";
        hash = "sha256-zvMjfaEq6EOWVqjVOoS2nb1fuGyEljcNVfTpAWUsiL8=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.8/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.13.0";
        hash = "sha256-7lI1RRC/Lw3OcYs3RA3caNvLYS7xEaCoxCM/ioa0goY=";
        modules = ./sigs/instrumentation/kube-state-metrics/2.13/gomod2nix.toml;
      };
      metrics-server = {
        version = "0.7.2";
        hash = "sha256-EdVph0HgOp6rIv/m3RtquSZq+43X5O8GJzg0zPXFWFI=";
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
    version = "1.34.0";
    srcHash = "sha256-rKy4X01pX+kovJ8b2JHV0KuzHJ7PYZ08eDEO3GeuPoc=";
    modules = ./core/1.34/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.9.2";
        hash = "sha256-H86EkdGmzvQDGC/a+J6ISB0aYkJabBjE2P6Ab5FRlv4=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.9/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.13.0";
        hash = "sha256-qLn+2znmfIdBkoVkCJ0tFAPVRYc+qAJWKbDP2FqMocg=";
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
    version = "1.35.0";
    srcHash = "sha256-AT1/4RhnVK/mAoNVqPIfSwbzD8VNRqKumOpE0fidJ74=";
    modules = ./core/1.35/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.9.2";
        hash = "sha256-04ytG4U8Luc5yh5VAbS1AQpjjapKsWWZSSB3IU5Rf6U=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.9/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.14.0";
        hash = "sha256-qLn+2znmfIdBkoVkCJ0tFAPVRYc+qAJWKbDP2FqMocg=";
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
    version = "1.36.0";
    srcHash = "sha256-6waSybeFc6xMIT93WLR1OPN/bhcHzvUzJGZliEuEQIM=";
    modules = ./core/1.36/gomod2nix.toml;
    sigs = {
      cluster-api = {
        version = "1.10.0";
        hash = "sha256-04ytG4U8Luc5yh5VAbS1AQpjjapKsWWZSSB3IU5Rf6U=";
        modules = ./sigs/cluster-lifecycle/cluster-api/1.10/gomod2nix.toml;
      };
      kube-state-metrics = {
        version = "2.14.0";
        hash = "sha256-EdVph0HgOp6rIv/m3RtquSZq+43X5O8GJzg0zPXFWFI=";
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
