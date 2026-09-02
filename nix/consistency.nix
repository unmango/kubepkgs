{
  lib,
  runCommand,
  python3,
  globset,
}:
let
  root = ../.;
in
runCommand "kubepkgs-consistency"
  {
    src = lib.fileset.toSource {
      inherit root;
      fileset = globset.lib.globs root [
        "packages.json"
        "README.md"
        "CLAUDE.md"
        ".github/copilot-instructions.md"
        "nix/check-consistency.py"
        "sigs/**/default.nix"
      ];
    };
    nativeBuildInputs = [ python3 ];
  }
  ''
    python3 "$src/nix/check-consistency.py" "$src"
    touch "$out"
  ''
