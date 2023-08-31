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

	serverBaseAddr string
}

type Document struct {
	ID         string
	Embeddings []float32
	Metadata   map[string]any
	Content    string
}

var (
	openai = embeddings.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"), http.DefaultClient)
)

type chromaCollectionObject struct {
	Embeddings [][]float32      `json:"embeddings"`
	Metadatas  []map[string]any `json:"metadatas"`
	Documents  []string         `json:"documents"`
	IDs        []string         `json:"ids"`
}

func (c chromaCollectionObject) asDocuments() []Document {
	docs := make([]Document, len(c.IDs))
	for i := 0; i < len(c.IDs); i++ {
		docs[i].ID = c.IDs[i]
		docs[i].Content = c.Documents[i]
		if c.Embeddings != nil {
			docs[i].Embeddings = c.Embeddings[i]
		}
		if c.Metadatas != nil {
			docs[i].Metadata = c.Metadatas[i]
		}
	}
	return docs
}

func (c Collection) Add(docs []Document, embedder embeddings.Embedder) error {
	addReq := chromaCollectionObject{
		Embeddings: make([][]float32, len(docs)),
		Metadatas:  make([]map[string]any, len(docs)),
		Documents:  make([]string, len(docs)),
		IDs:        make([]string, len(docs)),
	}
	for i, doc := range docs {
		// create embeddings if they don't exist
		if doc.Embeddings == nil || len(doc.Embeddings) == 0 {
			if e, err := embedder.GetEmbeddings(doc.ID, doc.Content); err != nil {
				return err
			} else {
				doc.Embeddings = e
			}
		}
		addReq.Embeddings[i] = doc.Embeddings
		addReq.Metadatas[i] = doc.Metadata
		addReq.Documents[i] = doc.Content
		addReq.IDs[i] = doc.ID
	}

	body, err := json.Marshal(addReq)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Post(
		fmt.Sprintf("%s/collections/%s/add", c.serverBaseAddr, c.ID),
		"application/json",
		bytes.NewBuffer(body))

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBuf, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error adding documents: Unable to read response body: %w", err)
		}
		return fmt.Errorf("error adding documents: %s", string(bodyBuf))
	}
	return nil
}

func (c Collection) Get(ids []string, where map[string]any, documents map[string]any) ([]Document, error) {
	payload := map[string]any{
		"ids":            ids,
		"where":          where,
		"where_document": documents,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		c.serverBaseAddr+"/collections/"+c.ID+"/get",
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respObj := chromaCollectionObject{}
	err = json.Unmarshal(bodyBuf, &respObj)
	if err != nil {
		return nil, err
	}
	return respObj.asDocuments(), nil
}
