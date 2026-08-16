// Tests in this package are white-box (package nixtool, not nixtool_test)
// so they can override runCombinedOutput and exercise the real
// stdout/exit-status parsing logic without invoking the actual nix/
// nix-prefetch-github binaries.
package nixtool

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNixtool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nixtool Suite")
}
