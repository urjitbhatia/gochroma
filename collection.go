package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urjitbhatia/gochroma/embeddings"
	"io"
	"net/http"
	"os"
)

type Collection struct {
	Name     string         `json:"name"`
	ID       string         `json:"id"`
	Metadata map[string]any `json:"metadata"`

	cc *Client
}

type Document struct {
	ID         string
	Embeddings []float32
	Metadata   map[string]any
	Content    string
}

var (
	openai                = embeddings.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"), http.DefaultClient)
	collectionsAddPathFmt = "collections/%s/add"
)

func GenerateEmbeddings(d Document, embedder embeddings.Embedder) error {
	var err error
	d.Embeddings, err = embedder.GetEmbeddings(d.ID, d.Content)
	return err
}

type chromaCollectionAddRequest struct {
	Embeddings [][]float32      `json:"embeddings"`
	Metadatas  []map[string]any `json:"metadatas"`
	Documents  []string         `json:"documents"`
	IDs        []string         `json:"ids"`
}

func (c Collection) Add(docs []Document, embedder embeddings.Embedder) error {
	addReq := chromaCollectionAddRequest{
		Embeddings: make([][]float32, len(docs)),
		Metadatas:  make([]map[string]any, len(docs)),
		Documents:  make([]string, len(docs)),
		IDs:        make([]string, len(docs)),
	}
	for i, doc := range docs {
		// create embeddings if they don't exist
		if doc.Embeddings == nil || len(doc.Embeddings) == 0 {
			if err := GenerateEmbeddings(doc, embedder); err != nil {
				return err
			}
		}
		addReq.Embeddings[i] = doc.Embeddings
		addReq.Metadatas[i] = doc.Metadata
		addReq.Documents[i] = doc.Content
		addReq.IDs[i] = doc.ID
	}

	path := fmt.Sprintf(collectionsAddPathFmt, c.Name)
	body, err := json.Marshal(addReq)
	if err != nil {
		return err
	}

	resp, err := c.cc.httpClient.Post(
		fmt.Sprintf("%s/%s", c.cc.url, path),
		"application/json",
		bytes.NewBuffer(body))

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	r := map[string]any{}
	err = json.Unmarshal(bodyBuf, &r)
	if err != nil {
		return fmt.Errorf("error decoding response from collection add documents. Err: %w Response: %s",
			err,
			string(bodyBuf))
	}
	return nil
}
