ATTR_core-1.33               := kubernetes."1.33".kubectl
ATTR_core-1.34               := kubernetes."1.34".kubectl
ATTR_core-1.35               := kubernetes."1.35".kubectl
ATTR_core-1.36               := kubernetes."1.36".kubectl
ATTR_cluster-api-1.8         := kubernetes."1.33".sigs.cluster-api
ATTR_cluster-api-1.9         := kubernetes."1.34".sigs.cluster-api
ATTR_cluster-api-1.10        := kubernetes."1.36".sigs.cluster-api
ATTR_kube-state-metrics-2.13 := kubernetes."1.33".sigs.kube-state-metrics
ATTR_kube-state-metrics-2.14 := kubernetes."1.35".sigs.kube-state-metrics
ATTR_metrics-server-0.7      := kubernetes."1.33".sigs.metrics-server
ATTR_external-dns-0.14       := kubernetes."1.33".sigs.external-dns
ATTR_external-dns-0.15       := kubernetes."1.34".sigs.external-dns

VENDOR_HASH_PKGS := \
	cluster-api-1.8 \
	cluster-api-1.9 \
	cluster-api-1.10 \
	kube-state-metrics-2.13 \
	kube-state-metrics-2.14 \
	metrics-server-0.7 \
	external-dns-0.14 \
	external-dns-0.15

CORE_PKGS := \
	core-1.33 \
	core-1.34 \
	core-1.35 \
	core-1.36

BUILD_PKGS := $(CORE_PKGS) $(VENDOR_HASH_PKGS)

SYSTEM ?= $(shell nix eval --impure --raw --expr 'builtins.currentSystem')
BUILD_TARGETS := $(addprefix build-,$(BUILD_PKGS))

.DEFAULT_GOAL := build

build:
	nix build .#

build-all: $(BUILD_TARGETS)

$(BUILD_TARGETS): build-%:
	nix build '.#legacyPackages.$(SYSTEM).$(ATTR_$*)'

update:
	nix flake update

check lint:
	nix flake check

format fmt:
	nix fmt

.PHONY: fetch-versions
fetch-versions:
	nix run '.#update' -- fetch-versions

hashes.json: versions.json
	nix run '.#update' -- generate-hashes

.PHONY: generate-hashes
generate-hashes: hashes.json

update-vendor-hash: PKG ?=
update-vendor-hash: hashes.json
	nix run '.#update' -- vendor-hashes $(if $(PKG),--sig $(PKG))

update-all-vendor-hashes: hashes.json
	nix run '.#update' -- vendor-hashes

update-releases: fetch-versions update-all-vendor-hashes

.PHONY: update-vendor-hash update-all-vendor-hashes update-releases
.PHONY: build build-all update check lint format fmt $(BUILD_TARGETS)
