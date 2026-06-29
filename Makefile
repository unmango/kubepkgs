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

GOMOD2NIX_PKGS := \
	core-1.33 \
	core-1.34 \
	core-1.35 \
	core-1.36 \
	cluster-api-1.8 \
	cluster-api-1.9 \
	cluster-api-1.10 \
	kube-state-metrics-2.13 \
	kube-state-metrics-2.14 \
	metrics-server-0.7 \
	external-dns-0.14 \
	external-dns-0.15

GOMOD2NIX_TARGETS := $(addprefix update-gomod2nix-,$(GOMOD2NIX_PKGS))

SYSTEM ?= $(shell nix eval --impure --raw --expr 'builtins.currentSystem')
BUILD_TARGETS := $(addprefix build-,$(GOMOD2NIX_PKGS))

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
	nix run '.#generate-hashes'

fetch-versions:
	nix run '.#fetch-versions'

update-releases: fetch-versions generate-hashes update-all-gomod2nix

update-gomod2nix: PKG ?= $(error PKG is required)
update-gomod2nix:
	nix run '.#$(ATTR_$(PKG)).updateGomod2nix'

update-all-gomod2nix: $(GOMOD2NIX_TARGETS)

$(GOMOD2NIX_TARGETS): update-gomod2nix-%:
	nix run '.#$(ATTR_$*).updateGomod2nix'

.PHONY: generate-hashes fetch-versions update-releases
.PHONY: build build-all $(BUILD_TARGETS)
.PHONY: update-gomod2nix update-all-gomod2nix $(GOMOD2NIX_TARGETS)
