package chroma_test

import (
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chroma "github.com/urjitbhatia/gochroma"
	"github.com/urjitbhatia/gochroma/embeddings"
	"net/http"
	"net/http/httptest"
)

type testEmbedder struct{}

func (e testEmbedder) EmbedQuery(_ context.Context, content string) ([]float32, error) {
	// implements a really simple embedding scheme which lets us test consistently
	// since the output depends on the input in some way
	return []float32{float32(len(content)), 1.1, 2.2}, nil
}

func (e testEmbedder) EmbedDocuments(_ context.Context, content []string) ([][]float32, error) {
	// implements a really simple embedding scheme which lets us test consistently
	// since the output depends on the input in some way
	var embedVectors [][]float32
	for _, text := range content {
		embedVectors = append(embedVectors, []float32{float32(len(text)), 1.1, 2.2})
	}
	return embedVectors, nil
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

	Describe("embeddings", func() {
		It("getEmbeddings", func() {
			// Start a local HTTP server
			embeddingsResponseObject := json.RawMessage(`
			{
			  "object": "list",
			  "data": [
				{
				  "object": "embedding",
				  "embedding": [
					0.0023064255,
					-0.009327292,
					-0.0028842222
				  ],
				  "index": 0
				},
				{
				  "object": "embedding",
				  "embedding": [
					1.0023064255,
					2.009327292,
					3.0028842222
				  ],
				  "index": 0
				}
			  ],
			  "model": "text-embedding-ada-002",
			  "usage": {
				"prompt_tokens": 8,
				"total_tokens": 8
			  }
			}
			`)
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				defer GinkgoRecover()
				// Test request parameters
				Expect(req.URL.String()).To(Equal("/embeddings"))
				// Send response to be tested
				rw.Write(embeddingsResponseObject)
			}))
			// Close the server when test finishes
			defer server.Close()

			openai := embeddings.NewOpenAIClientWithHTTP(server.URL, "", server.Client())
			e, err := openai.EmbedQuery(context.Background(), "foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(e)).To(Equal(3))

			// test batch
			ee, err := openai.EmbedDocuments(context.Background(), []string{"foo", "bar"})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(ee)).To(Equal(2))
			Expect(len(ee[0])).To(Equal(3))
			Expect(len(ee[1])).To(Equal(3))
			Expect(ee[1][0]).To(BeNumerically(">=", 1.0))
			Expect(ee[1][1]).To(BeNumerically(">=", 2.0))
			Expect(ee[1][2]).To(BeNumerically(">=", 3.0))
		})
	})

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

		It("counts documents in the collection", func() {
			count, err := testCollection.Count()
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(Equal(2))
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
			td := testDocument1
			td.Distance = 4
			Expect(docs[0]).To(Equal(td))
			td = testDocument2
			td.Distance = 64
			Expect(docs[1]).To(Equal(td))
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
			td := testDocument2
			td.Distance = 81
			Expect(docs[0]).To(Equal(td))
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
			td := testDocument1
			td.Distance = 289
			Expect(docs[0]).To(Equal(td))
		})

	})

	It("Slice batcher", func() {
		flatten := func(nested [][]int) []int {
			var res []int
			for _, item := range nested {
				res = append(res, item...)
			}
			return res
		}

		items := []int{1, 2, 3, 4, 5, 6, 7}
		batches := chroma.SliceBatch(items, 1)
		Expect(len(batches)).To(Equal(7))
		Expect(flatten(batches)).To(Equal(items))

		batches = chroma.SliceBatch(items, 3)
		Expect(len(batches)).To(Equal(3))
		Expect(flatten(batches)).To(Equal(items))

		batches = chroma.SliceBatch(items, 4)
		Expect(len(batches)).To(Equal(2))
		Expect(flatten(batches)).To(Equal(items))

		batches = chroma.SliceBatch(items, 10)
		Expect(len(batches)).To(Equal(1))
		Expect(flatten(batches)).To(Equal(items))
	})
})
