{
  gh,
  jq,
  nix-prefetch-github,
  writeShellApplication,
}:
writeShellApplication {
  name = "generate-hashes";

  runtimeInputs = [
    gh
    jq
    nix-prefetch-github
  ];

  text = ''
    TARGET="$1"
    MINOR="$2"

    REPO_ROOT="$(git rev-parse --show-toplevel)"
    HASHES_JSON="$REPO_ROOT/hashes.json"
    versions="$(cat "$REPO_ROOT/versions.json")"
    FAKE="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

    fetch_commit() {
      local ref_data
      ref_data="$(gh api "repos/$1/$2/git/refs/tags/$3")"
      if [[ "$(echo "$ref_data" | jq -r '.object.type')" == "tag" ]]; then
        gh api "repos/$1/$2/git/tags/$(echo "$ref_data" | jq -r '.object.sha')" | jq -r '.object.sha'
      else
        echo "$ref_data" | jq -r '.object.sha'
      fi
    }

    if [[ "$TARGET" == "kubernetes" ]]; then
      owner="kubernetes"; repo="kubernetes"
      version="$(echo "$versions" | jq -r ".kubernetes.\"$MINOR\".version")"
      path=".kubernetes.\"$MINOR\""
      vendor="{}"
    else
      case "$TARGET" in
        kube-state-metrics) owner="kubernetes" ;;
        cluster-api|metrics-server|external-dns) owner="kubernetes-sigs" ;;
        *) exit 1 ;;
      esac
      repo="$TARGET"
      version="$(echo "$versions" | jq -r ".kubernetes.\"$MINOR\".sigs.\"$TARGET\"")"
      path=".sigs.\"$TARGET\".\"$MINOR\""
      vendor="{vendorHash: ($path.vendorHash // \$fake)}"
    fi

    src_hash="$(nix-prefetch-github --json --rev "v$version" "$owner" "$repo" | jq -r '.hash')"
    commit="$(fetch_commit "$owner" "$repo" "v$version")"

    jq --arg v "$version" --arg s "$src_hash" --arg c "$commit" --arg fake "$FAKE" \
      "$path = {version: \$v, srcHash: \$s, commit: \$c} + $vendor" \
      "$HASHES_JSON" > "$HASHES_JSON.tmp"
    mv "$HASHES_JSON.tmp" "$HASHES_JSON"
  '';
}
