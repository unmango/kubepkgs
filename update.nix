{ pkgs, gomod2nixPkg }:
{ src, outdir }:
pkgs.writeShellApplication {
  name = "update-gomod2nix";
  runtimeInputs = [ gomod2nixPkg ];
  text = ''
    gomod2nix --dir "${src}" --outdir "''${1:-${outdir}}"
  '';
}
