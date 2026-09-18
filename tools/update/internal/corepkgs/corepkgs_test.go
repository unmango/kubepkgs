package corepkgs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/kubepkgs/tools/update/internal/corepkgs"
)

// source mirrors the shape of core/default.nix: `let` bindings sharing the
// result set's indentation, a nested `let ... in`, and packages built both
// through the shared helper and directly.
const source = `{ buildGoModule, stdenv, lib }:
let
  versionLdflags =
    let
      xFlag = pkg: key: val: "-X '${pkg}.${key}=${val}'";
    in
    [ (xFlag "a" "b" "c") ];

  mkBin = pname: subPkg: buildGoModule { inherit pname; };
in
{
  kubectl = mkBin "kubectl" "cmd/kubectl";

  # A comment between packages.
  kube-proxy = mkBin "kube-proxy" "cmd/kube-proxy";

  pause = stdenv.mkDerivation {
    pname = "pause";
    meta = {
      mainProgram = "pause";
    };
  };
}
`

var _ = Describe("Names", func() {
	It("returns the packages in declaration order", func() {
		Expect(corepkgs.Names(source)).To(Equal([]string{"kubectl", "kube-proxy", "pause"}))
	})

	It("ignores the let bindings above the result set", func() {
		Expect(corepkgs.Names(source)).NotTo(ContainElement("mkBin"))
		Expect(corepkgs.Names(source)).NotTo(ContainElement("versionLdflags"))
	})

	It("reports a file with no top-level result set", func() {
		_, err := corepkgs.Names("{ lib }:\n{\n}\n")
		Expect(err).To(MatchError(ContainSubstring("no top-level `in`")))
	})

	It("reports a result set with no packages", func() {
		_, err := corepkgs.Names("{ lib }:\nlet\n  x = 1;\nin\n{\n}\n")
		Expect(err).To(MatchError(ContainSubstring("declares no packages")))
	})
})
