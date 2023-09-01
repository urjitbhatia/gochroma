package embeddings

// TODO: Support batch call
type Embedder interface {
	GetEmbeddings(id string, content string) ([]float32, error)
}
