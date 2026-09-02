package readme_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReadme(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Readme Suite")
}
