{
  description = "Kubernetes packages by release";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;
      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        {
          pkgs,
          lib,
          ...
        }:
        let
          callPackage = lib.callPackageWith (
            {
              inherit callPackage;
            }
            // pkgs
          );

          mkRelease = callPackage ./mk-release.nix { inherit callPackage; };

          releases = import ./releases.nix { inherit lib; };
          latestVersion = releases.latest;

          releaseData = builtins.removeAttrs releases [
            "supported"
            "latest"
          ];

          versionedSets = lib.mapAttrs (
            v: info:
            mkRelease {
              inherit (info)
                version
                srcHash
                commit
                sigs
                ;
            }
          ) releaseData;

          latest = versionedSets.${latestVersion};
        in
        {
          # Versioned package sets: legacyPackages.kubernetes."1.36".kubectl
          #                         legacyPackages.kubernetes."1.36".sigs.cluster-api
          #                         legacyPackages.kubernetes.latest.kubectl
          legacyPackages.kubernetes = versionedSets // {
            inherit latest;
          };

          packages = {
            default = latest.kube-apiserver;
            generate-hashes = pkgs.callPackage ./generate-hashes.nix { };
            fetch-versions = pkgs.callPackage ./fetch-versions.nix { };
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              gnumake
              nixfmt
              gh
              jq
              nix-prefetch-github
            ];
          };

          treefmt.programs = {
            nixfmt.enable = true;
          };
        };
    };
}
