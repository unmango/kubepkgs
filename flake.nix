{
  description = "Kubernetes packages by release";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";
    globset.url = "github:pdtpartners/globset";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
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
          inputs',
          ...
        }:
        let
          mkRelease = pkgs.callPackage ./mk-release.nix { };

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

          update = pkgs.callPackage ./nix/updater.nix {
            inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication;
            globset = inputs.globset;
          };
        in
        {
          legacyPackages.kubernetes = versionedSets // {
            inherit latest;
          };

          packages = {
            default = latest.kube-apiserver;
            inherit update;
          };

          checks =
            (lib.mapAttrs' (n: lib.nameValuePair "core-${n}") (
              lib.filterAttrs (_: lib.isDerivation) (removeAttrs latest [ "sigs" ])
            ))
            // (lib.mapAttrs' (n: lib.nameValuePair "sig-${n}") latest.sigs)
            // {
              inherit update;
            };

          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              gnumake
              nixfmt
              nix-prefetch-github
              go
              inputs'.gomod2nix.packages.default
            ];
          };

          treefmt.programs = {
            nixfmt.enable = true;
          };
        };
    };
}
