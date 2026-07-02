{
  jq,
  nix,
  writeShellApplication,
}:
writeShellApplication {
  name = "update-vendor-hash";

  runtimeInputs = [
    nix
    jq
  ];

  text = ''
    SIG="$1"
    BUILD_MINOR="$2"
    shift 2
    UPDATE_MINORS=("$@")

    REPO_ROOT="$(git rev-parse --show-toplevel)"
    HASHES_JSON="$REPO_ROOT/hashes.json"
    FAKE="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
    SYSTEM=$(nix eval --impure --raw --expr 'builtins.currentSystem')
    ATTR=".#legacyPackages.$SYSTEM.kubernetes.\"$BUILD_MINOR\".sigs.$SIG"

    BAK="$HASHES_JSON.bak"
    cp "$HASHES_JSON" "$BAK"
    trap 'mv "$BAK" "$HASHES_JSON"' ERR INT TERM

    jq --arg sig "$SIG" --arg minor "$BUILD_MINOR" --arg h "$FAKE" \
      '.sigs[$sig][$minor].vendorHash = $h' "$BAK" > "$HASHES_JSON"

    set +e
    output=$(nix build --no-link "$ATTR" 2>&1)
    set -e

    got=$(printf '%s\n' "$output" | grep 'got:' | grep -oE 'sha256-[A-Za-z0-9+/=]+' | head -1)
    [ -n "$got" ] || { mv "$BAK" "$HASHES_JSON"; exit 1; }

    minors=$(printf '%s\n' "''${UPDATE_MINORS[@]}" | jq -R . | jq -s .)
    jq --arg sig "$SIG" --arg hash "$got" --argjson minors "$minors" \
      'reduce $minors[] as $m (.; .sigs[$sig][$m].vendorHash = $hash)' "$BAK" > "$HASHES_JSON"
    rm "$BAK"
  '';
}
