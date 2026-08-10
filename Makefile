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

# sig-name, build-k8s-minor, update-minors... (same version shared across minors gets one build)
VENDOR_HASH_ARGS_cluster-api-1.8         := cluster-api 1.33 1.33
VENDOR_HASH_ARGS_cluster-api-1.9         := cluster-api 1.34 1.34 1.35
VENDOR_HASH_ARGS_cluster-api-1.10        := cluster-api 1.36 1.36
VENDOR_HASH_ARGS_kube-state-metrics-2.13 := kube-state-metrics 1.33 1.33 1.34
VENDOR_HASH_ARGS_kube-state-metrics-2.14 := kube-state-metrics 1.35 1.35 1.36
VENDOR_HASH_ARGS_metrics-server-0.7      := metrics-server 1.33 1.33 1.34 1.35 1.36
VENDOR_HASH_ARGS_external-dns-0.14       := external-dns 1.33 1.33
VENDOR_HASH_ARGS_external-dns-0.15       := external-dns 1.34 1.34 1.35 1.36

VENDOR_HASH_PKGS := \
	cluster-api-1.8 \
	cluster-api-1.9 \
	cluster-api-1.10 \
	kube-state-metrics-2.13 \
	kube-state-metrics-2.14 \
	metrics-server-0.7 \
	external-dns-0.14 \
	external-dns-0.15

VENDOR_HASH_TARGETS := $(addprefix update-vendor-hash-,$(VENDOR_HASH_PKGS))

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

generate-hashes:
	@for minor in $$(jq -r '.supported[]' versions.json); do \
	  for target in kubernetes cluster-api kube-state-metrics metrics-server external-dns; do \
	    nix run '.#generate-hashes' -- $$target $$minor; \
	  done; \
	done

fetch-versions:
	nix run '.#fetch-versions'

update-vendor-hash: PKG ?= $(error PKG is required)
update-vendor-hash: update-vendor-hash-$(PKG)

update-all-vendor-hashes: $(VENDOR_HASH_TARGETS)

$(VENDOR_HASH_TARGETS): update-vendor-hash-%:
	nix run '.#update-vendor-hash' -- $(VENDOR_HASH_ARGS_$*)

update-releases: fetch-versions generate-hashes update-all-vendor-hashes

.PHONY: generate-hashes fetch-versions update-releases
.PHONY: update-vendor-hash update-all-vendor-hashes $(VENDOR_HASH_TARGETS)
.PHONY: build build-all update check lint format fmt $(BUILD_TARGETS)
