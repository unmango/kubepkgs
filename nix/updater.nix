{
  buildGoApplication,
  globset,
  makeWrapper,
  lib,
  git,
  nix,
  nix-prefetch-github,
}:
let
  root = ../tools/update;
in
buildGoApplication {
  pname = "kubepkgs-update";
  version = "0.1.0";
  subPackages = [ "cmd/kubepkgs-update" ];
  src = lib.fileset.toSource {
    inherit root;
    fileset = globset.lib.globs root [
      "go.mod"
      "go.sum"
      "**/*.go"
      "**/testdata/**"
    ];
  };
  modules = ../tools/update/gomod2nix.toml;
  nativeBuildInputs = [ makeWrapper ];
  postFixup = ''
    wrapProgram $out/bin/kubepkgs-update \
      --prefix PATH : ${
        lib.makeBinPath [
          git
          nix
          nix-prefetch-github
        ]
      }
  '';
  meta.mainProgram = "kubepkgs-update";
}
