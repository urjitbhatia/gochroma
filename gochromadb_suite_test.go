package chroma_test

import (
	chroma "github.com/urjitbhatia/gochroma"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testClient *chroma.Client

func TestGochromadb(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		c, err := chroma.NewClient("http://localhost:8001")
		Expect(err).ToNot(HaveOccurred())
		testClient = c
	})
	RunSpecs(t, "Gochromadb Suite")
}
