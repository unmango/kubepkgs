{ pkgs }:
pkgs.writeShellApplication {
  name = "update-vendor-hash";
  runtimeInputs = with pkgs; [
    nix
    jq
  ];
  text = ''
    # Usage: update-vendor-hash <sig-name> <build-k8s-minor> <update-minor> [<update-minor> ...]
    # Computes vendorHash for sigs.<sig-name> by building against <build-k8s-minor>,
    # then writes the result into hashes.json for every <update-minor> listed.
    SIG="$1"
    BUILD_MINOR="$2"
    shift 2
    UPDATE_MINORS=("$@")

    REPO_ROOT="$(git rev-parse --show-toplevel)"
    HASHES_JSON="$REPO_ROOT/hashes.json"
    FAKE="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
    SYSTEM=$(nix eval --impure --raw --expr 'builtins.currentSystem')
    ATTR=".#legacyPackages.$SYSTEM.kubernetes.\"$BUILD_MINOR\".sigs.$SIG"

    BAK="$HASHES_JSON.update-bak"
    cp "$HASHES_JSON" "$BAK"
    restore() { mv "$BAK" "$HASHES_JSON"; }
    trap restore ERR INT TERM

    # Write fake hash so the build fails with the real hash in the error output
    jq --arg sig "$SIG" --arg minor "$BUILD_MINOR" --arg h "$FAKE" \
      '.sigs[$sig][$minor].vendorHash = $h' "$BAK" > "$HASHES_JSON"

    set +e
    output=$(nix build --no-link "$ATTR" 2>&1)
    set -e

    got=$(printf '%s\n' "$output" | grep 'got:' | grep -oE 'sha256-[A-Za-z0-9+/=]+' | head -1)

    if [ -z "$got" ]; then
      restore
      printf 'Could not extract vendorHash from nix output:\n%s\n' "$output" >&2
      exit 1
    fi

    tmp=$(mktemp)
    cp "$BAK" "$tmp"
    for minor in "''${UPDATE_MINORS[@]}"; do
      jq --arg sig "$SIG" --arg minor "$minor" --arg hash "$got" \
        '.sigs[$sig][$minor].vendorHash = $hash' "$tmp" > "$HASHES_JSON"
      cp "$HASHES_JSON" "$tmp"
      echo "Updated .sigs.$SIG.$minor.vendorHash = $got"
    done
    rm "$tmp" "$BAK"
  '';
}
