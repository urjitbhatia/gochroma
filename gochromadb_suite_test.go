package chroma_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGochromadb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gochromadb Suite")
}
