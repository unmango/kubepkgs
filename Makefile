.DEFAULT_GOAL := build

build:
	nix build .#

update:
	nix flake update

check lint:
	nix flake check

format fmt:
	nix fmt

# Run the three stages below afterwards to fill in the hashes it leaves empty.
add-minor:
	nix run '.#update' -- add-minor

fetch-versions:
	nix run '.#update' -- fetch-versions

generate-hashes:
	nix run '.#update' -- generate-hashes

vendor-hashes:
	nix run '.#update' -- vendor-hashes

sync-docs:
	nix run '.#update' -- sync-docs

update-releases: fetch-versions generate-hashes vendor-hashes

.PHONY: add-minor sync-docs fetch-versions generate-hashes vendor-hashes update-releases
.PHONY: build update check lint format fmt
