# Resolves the optional "go" field in packages.json to the nixpkgs attribute
# providing that toolchain, so a package that does not build with the default
# Go can say so as data rather than by pinning the whole flake to a nixpkgs
# that happens to ship the right one.
#
# This selects among the versions nixpkgs offers. A package needing a Go newer
# than any of them still needs the nixpkgs input to move.
{ lib, pkgs }:
label: requested:
if requested == null then
  null
else
  let
    attr = "go_" + lib.replaceStrings [ "." ] [ "_" ] requested;
    available = lib.filter (n: lib.hasPrefix "go_" n && lib.match "go_[0-9_]+" n != null) (
      lib.attrNames pkgs
    );
  in
  pkgs.${attr} or (throw ''
    ${label} pins Go ${requested}, which this nixpkgs does not provide.
    Wanted attribute: ${attr}
    Available: ${lib.concatStringsSep ", " available}
    Either pin a version nixpkgs ships, or update the nixpkgs input.'')
