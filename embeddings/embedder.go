package embeddings

type Embedder interface {
	GetEmbeddings(content string) ([]float32, error)
	GetEmbeddingsBatch(content []string) ([][]float32, error)
}
