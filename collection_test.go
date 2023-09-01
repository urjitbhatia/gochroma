package chroma_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chroma "github.com/urjitbhatia/gochroma"
)

type testEmbedder struct {
}

func (e testEmbedder) GetEmbeddings(_ string, content string) ([]float32, error) {
	return []float32{float32(len(content)), 1.1, 2.2}, nil
}

var _ = Describe("Collection", func() {
	testDocument1 := chroma.Document{
		ID:         "testDoc1",
		Embeddings: nil,
		Metadata:   map[string]any{"source": "unittest_doc_1"},
		Content:    "Hello, how are you?",
	}
	testDocument2 := chroma.Document{
		ID:         "testDoc2",
		Embeddings: nil,
		Metadata:   map[string]any{"source": "unittest_doc_2"},
		Content:    "I am well",
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
			err := testCollection.Add([]chroma.Document{testDocument1}, testEmbedder{})
			Expect(err).ToNot(HaveOccurred())

			err = testCollection.Add([]chroma.Document{testDocument2}, testEmbedder{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("gets documents", func() {

			docs, err := testCollection.Get(nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(docs)).To(Equal(2))
			Expect(docs[0]).To(Equal(testDocument1))
		})

		It("query documents by text", func() {
			docs, err := testCollection.Query(
				"Hello, how are yu",
				2,
				nil,
				nil,
				[]chroma.QueryEnum{chroma.WithDocuments, chroma.WithMetadatas, chroma.WithDistances},
				testEmbedder{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(docs)).To(Equal(2))
			// document 1 will be close since our test embedding generator depends on content length
			// and doc1's content length is closer
			Expect(docs[0]).To(Equal(testDocument1))
			Expect(docs[1]).To(Equal(testDocument2))
		})

		It("restrict query by metadata", func() {
			docs, err := testCollection.Query(
				"",
				2,
				map[string]any{"source": "unittest_doc_2"},
				nil,
				[]chroma.QueryEnum{chroma.WithDocuments, chroma.WithMetadatas, chroma.WithDistances},
				testEmbedder{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(docs)).To(Equal(1))
			Expect(docs[0]).To(Equal(testDocument2))
		})

		It("restrict query by where document", func() {
			docs, err := testCollection.Query(
				"y?",
				2,
				nil,
				map[string]interface{}{"$contains": "you"},
				[]chroma.QueryEnum{chroma.WithDocuments, chroma.WithMetadatas, chroma.WithDistances},
				testEmbedder{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(docs)).To(Equal(1))
			Expect(docs[0]).To(Equal(testDocument1))
		})
	})
})
