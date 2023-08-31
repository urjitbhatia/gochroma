package chroma_test

import (
	chroma "github.com/urjitbhatia/gochroma"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testClient *chroma.Client
var testCollection chroma.Collection

func TestGochromadb(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		c, err := chroma.NewClient("http://localhost:8001")
		Expect(err).ToNot(HaveOccurred())
		testClient = c

		tc, err := testClient.
			GetOrCreateCollection("collections-unit-test", "l2", nil)
		Expect(err).ToNot(HaveOccurred())
		testCollection = tc
	})
	RunSpecs(t, "Gochromadb Suite")
}
