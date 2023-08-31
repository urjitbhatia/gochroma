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

var _ = FDescribe("Collection", func() {
	It("generates Embeddings", func() {
		err := chroma.GenerateEmbeddings(chroma.Document{
			ID:         "testDoc1",
			Embeddings: nil,
			Metadata:   nil,
			Content:    "Hello, how are you?",
		}, testEmbedder{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("adds documents", func() {
		Expect(testCollection.Name).To(Equal("collections-unit-test"))

		err := testCollection.Add([]chroma.Document{{
			ID:         "testDoc1",
			Embeddings: nil,
			Metadata:   nil,
			Content:    "Hello, how are you?",
		}}, testEmbedder{})
		Expect(err).ToNot(HaveOccurred())
	})
})
