package embeddings

import "context"

// Embedder is an interface matching github.com/tmc/langchaingo embedder
// used for generating embeddings for a single document and a batch
type Embedder interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}
