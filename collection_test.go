package chroma_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chroma "github.com/urjitbhatia/gochroma"
)

type testEmbedder struct {
}

func (e testEmbedder) GetEmbeddings(_ string, _ string) ([]float32, error) {
	return []float32{0.0, 1.1, 2.2}, nil
}

var _ = Describe("Collection", func() {
	testDocument := chroma.Document{
		ID:         "testDoc1",
		Embeddings: nil,
		Metadata:   map[string]any{"source": "unittest"},
		Content:    "Hello, how are you?",
	}

	Describe("add, fetch, delete sequence", Ordered, func() {

		var testCollection chroma.Collection
		BeforeAll(func() {
			testClient.DeleteCollection("collections-unit-test")
			// this can error if the reset was called previously in the tests,
			// so we can ignore the error here

			tc, err := testClient.CreateCollection("collections-unit-test", "l2", nil)
			Expect(err).ToNot(HaveOccurred())
			testCollection = tc
		})

		It("adds documents", func() {
			Expect(testCollection.Name).To(Equal("collections-unit-test"))

			err := testCollection.Add([]chroma.Document{testDocument}, testEmbedder{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("gets documents", func() {
			Expect(testCollection.Name).To(Equal("collections-unit-test"))

			docs, err := testCollection.Get(nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(docs)).To(Equal(1))
			Expect(docs[0]).To(Equal(testDocument))
		})
	})
})
