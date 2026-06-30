{
  description = "Kubernetes packages by release";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.inputs.systems.follows = "systems";
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
          inputs',
          pkgs,
          lib,
          ...
        }:
        let
          inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication gomod2nix;

          mkGomod2nixUpdater = import ./update.nix {
            inherit pkgs;
            gomod2nixPkg = gomod2nix;
          };

          callPackage = lib.callPackageWith (
            {
              inherit buildGoApplication callPackage mkGomod2nixUpdater;
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
                modules
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
          #
          # Add packages/overlayAttrs aliases here once real source hashes
          # and gomod2nix.toml files are generated for each version.
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
              gomod2nix
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
