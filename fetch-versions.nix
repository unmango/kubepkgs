{ pkgs }:
pkgs.writeShellApplication {
  name = "fetch-versions";
  runtimeInputs = with pkgs; [
    gh
    jq
  ];
  text = ''
    REPO_ROOT="$(git rev-parse --show-toplevel)"
    VERSIONS_JSON="''${REPO_ROOT}/versions.json"
    TMPFILE="$(mktemp)"
    trap 'rm -f "''${TMPFILE}"' EXIT

    versions="$(cat "''${VERSIONS_JSON}")"

    sig_owner() {
      case "$1" in
        cluster-api)        echo "kubernetes-sigs" ;;
        kube-state-metrics) echo "kubernetes" ;;
        metrics-server)     echo "kubernetes-sigs" ;;
        external-dns)       echo "kubernetes-sigs" ;;
        *) echo "Unknown SIG: $1" >&2; exit 1 ;;
      esac
    }

    # Latest non-prerelease tag with the given minor prefix, sorted by patch number.
    # Fetches up to 100 releases (sufficient for any tracked minor series).
    latest_patch() {
      local owner="$1" repo="$2" minor_prefix="$3"
      gh api "repos/''${owner}/''${repo}/releases?per_page=100" \
        --jq "[.[] | select(.prerelease == false and .draft == false)
               | .tag_name | ltrimstr(\"v\")
               | select(startswith(\"''${minor_prefix}.\"))]
              | sort_by(split(\".\")[-1] | tonumber)
              | last // empty"
    }

    # Update kubernetes core patch versions
    while IFS= read -r k8s_minor; do
      current="$(echo "''${versions}" | jq -r ".kubernetes.\"''${k8s_minor}\".version")"
      latest="$(latest_patch kubernetes kubernetes "''${k8s_minor}")"
      if [[ -n "''${latest}" && "''${latest}" != "''${current}" ]]; then
        echo "kubernetes ''${k8s_minor}: ''${current} -> ''${latest}" >&2
        versions="$(echo "''${versions}" | jq ".kubernetes.\"''${k8s_minor}\".version = \"''${latest}\"")"
      else
        echo "kubernetes ''${k8s_minor}: ''${current} (up to date)" >&2
      fi
    done < <(echo "''${versions}" | jq -r '.supported[]')

    # Update SIG patch versions within each tracked minor series
    while IFS= read -r k8s_minor; do
      for sig in cluster-api kube-state-metrics metrics-server external-dns; do
        current="$(echo "''${versions}" | jq -r ".kubernetes.\"''${k8s_minor}\".sigs.\"''${sig}\"")"
        sig_minor="''${current%.*}"
        owner="$(sig_owner "''${sig}")"
        latest="$(latest_patch "''${owner}" "''${sig}" "''${sig_minor}")"
        if [[ -n "''${latest}" && "''${latest}" != "''${current}" ]]; then
          echo "  ''${sig} ''${k8s_minor}: ''${current} -> ''${latest}" >&2
          versions="$(echo "''${versions}" | jq ".kubernetes.\"''${k8s_minor}\".sigs.\"''${sig}\" = \"''${latest}\"")"
        else
          echo "  ''${sig} ''${k8s_minor}: ''${current} (up to date)" >&2
        fi
      done
    done < <(echo "''${versions}" | jq -r '.supported[]')

    echo "''${versions}" | jq '.' > "''${TMPFILE}"
    mv "''${TMPFILE}" "''${VERSIONS_JSON}"
    echo "Wrote ''${VERSIONS_JSON}" >&2
  '';
}
