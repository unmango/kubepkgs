package nixtool

import (
	"context"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeExitOutput overrides runCombinedOutput to run a real subprocess that
// prints text and exits with code, producing a genuine *exec.ExitError
// (rather than a hand-constructed one) so ResolveVendorHash's
// errors.As(err, &exitErr) check is exercised for real.
func fakeExitOutput(text string, code int) {
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("cat >&2 <<'EOF'\n%s\nEOF\nexit %d", text, code))
		return cmd.CombinedOutput()
	}
}

var _ = Describe("ResolveVendorHash", func() {
	AfterEach(func() {
		runCombinedOutput = defaultRunCombinedOutput
	})

	DescribeTable("parsing nix build output",
		func(output string, exitCode int, wantHash string, wantFound bool) {
			fakeExitOutput(output, exitCode)

			hash, found, err := ResolveVendorHash(context.Background(), ".#some.attr")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(Equal(wantFound))
			Expect(hash).To(Equal(wantHash))
		},
		Entry("real nix hash-mismatch failure",
			"error: hash mismatch in fixed-output derivation:\n"+
				"         specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n"+
				"            got:    sha256-nEtYpsLMUbVFL6cD9pylrzv18r15pYppKHiAcSfG1ew=\n",
			1, "sha256-nEtYpsLMUbVFL6cD9pylrzv18r15pYppKHiAcSfG1ew=", true,
		),
		Entry("no got: line present (the actual failure case)",
			"error: builder for '/nix/store/xyz.drv' failed with exit code 1\n",
			1, "", false,
		),
		Entry("succeeds with no output at all",
			"", 0, "", false,
		),
	)

	It("returns a real error when nix itself can't be run", func() {
		runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return nil, exec.ErrNotFound
		}

		_, found, err := ResolveVendorHash(context.Background(), ".#some.attr")
		Expect(err).To(HaveOccurred())
		Expect(found).To(BeFalse())
	})
})

var defaultRunCombinedOutput = runCombinedOutput
