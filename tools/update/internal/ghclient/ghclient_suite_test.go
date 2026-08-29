package ghclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGhclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ghclient Suite")
}
