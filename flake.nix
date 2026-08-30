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

          # Every core binary for the latest minor, plus kubectl for each
          # older supported minor: enough to catch a bad source hash or a
          # build break on any tracked Kubernetes release without compiling
          # the full binary set four times over.
          coreChecks =
            (lib.mapAttrs' (name: lib.nameValuePair "core-${latestVersion}-${name}") (
              lib.filterAttrs (_: lib.isDerivation) (removeAttrs latest [ "sigs" ])
            ))
            // lib.listToAttrs (
              map (minor: lib.nameValuePair "core-${minor}-kubectl" versionedSets.${minor}.kubectl) (
                lib.filter (minor: minor != latestVersion) releases.supported
              )
            );

          # One build per distinct SIG version rather than per minor: minors
          # pinning the same version produce the same derivation, and the
          # attribute name collapses them. Derived from the data, so a
          # version bump changes what CI covers without anyone editing a list.
          sigChecks = lib.listToAttrs (
            lib.concatMap (
              minor:
              lib.mapAttrsToList (
                name: pkg: lib.nameValuePair "sig-${name}-${releaseData.${minor}.sigs.${name}.version}" pkg
              ) versionedSets.${minor}.sigs
            ) releases.supported
          );

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
            coreChecks
            // sigChecks
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
