{ lib, ... }:
let
  releases = import ./releases.nix;
  latestVersion = releases.latest;
in
{
  perSystem =
    {
      pkgs,
      inputs',
      ...
    }:
    let
      inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication;
      callPackage = lib.callPackageWith ({ inherit buildGoApplication callPackage; } // pkgs);
      mkRelease = callPackage ./mk-release.nix { inherit callPackage; };

      releaseData = builtins.removeAttrs releases [
        "supported"
        "latest"
      ];
      versionedSets = lib.mapAttrs (
        v: info:
        mkRelease {
          version = v;
          inherit (info)
            srcHash
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
    };
}
