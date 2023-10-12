package chroma_test

import (
	chroma "github.com/urjitbhatia/gochroma"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testClient chroma.Chroma

func TestGochromadb(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		c, err := chroma.NewClient("http://localhost:8000")
		Expect(err).ToNot(HaveOccurred())
		testClient = c
	})

	RunSpecs(t, "Gochromadb Suite")
}
