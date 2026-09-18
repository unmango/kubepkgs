.DEFAULT_GOAL := build

# Everything CI builds is a flake check, derived from packages.json; see
# `nix flake show` for the list.
build:
	nix build .#

update:
	nix flake update

check lint:
	nix flake check

format fmt:
	nix fmt

# Start tracking the newest Kubernetes minor upstream, retiring the oldest. A
# no-op when the newest minor is one already tracked. Run the three stages
# below afterwards to fill in the hashes it deliberately leaves empty.
add-minor:
	nix run '.#update' -- add-minor

# The three stages of the packages.json update lifecycle, in order. Each is a
# thin wrapper; pass flags to the CLI directly for anything narrower.
fetch-versions:
	nix run '.#update' -- fetch-versions

generate-hashes:
	nix run '.#update' -- generate-hashes

vendor-hashes:
	nix run '.#update' -- vendor-hashes

# Regenerate the README's version tables and package inventory lists after an
# edit that they reflect, such as adding a SIG or a core binary. add-minor
# already does this for the changes it makes.
sync-docs:
	nix run '.#update' -- sync-docs

update-releases: fetch-versions generate-hashes vendor-hashes

.PHONY: add-minor sync-docs fetch-versions generate-hashes vendor-hashes update-releases
.PHONY: build update check lint format fmt
