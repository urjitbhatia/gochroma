package embeddings

type Embedder interface {
	GetEmbeddings(id string, content string) ([]float32, error)
}
