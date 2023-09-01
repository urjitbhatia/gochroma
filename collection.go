package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urjitbhatia/gochroma/embeddings"
	"io"
	"net/http"
	"os"
	"strconv"
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
	openai = embeddings.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
)

type chromaCollectionObject struct {
	Embeddings [][]float32      `json:"embeddings"`
	Metadatas  []map[string]any `json:"metadatas"`
	Documents  []string         `json:"documents"`
	IDs        []string         `json:"ids"`
}
type chromaQueryResultObject struct {
	Embeddings [][][]float32      `json:"embeddings"`
	Metadatas  [][]map[string]any `json:"metadatas"`
	Documents  [][]string         `json:"documents"`
	IDs        [][]string         `json:"ids"`
}

func (o chromaQueryResultObject) asFlattenedDocuments() []Document {
	docs := []Document{}
	for i := 0; i < len(o.IDs); i++ {
		for j := 0; j < len(o.IDs[i]); j++ {
			d := Document{
				ID:      o.IDs[i][j],
				Content: o.Documents[i][j],
			}
			if len(o.Embeddings) > 0 && len(o.Embeddings[i]) > 0 {
				d.Embeddings = o.Embeddings[i][j]
			}
			if len(o.Metadatas) > 0 && len(o.Metadatas[i]) > 0 {
				d.Metadata = o.Metadatas[i][j]
			}
			docs = append(docs, d)
		}
	}
	return docs
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
		if len(doc.Embeddings) == 0 {
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

type QueryEnum string

const (
	WithDocuments  QueryEnum = "documents"
	WithEmbeddings QueryEnum = "embeddings"
	WithMetadatas  QueryEnum = "metadatas"
	WithDistances  QueryEnum = "distances"
)

/*
Query fetches results for a single query. TODO: bulk query implementation
This calculates the embeddings for the query automatically. TODO: allow search by embeddings
*/
func (c Collection) Query(query string, numResults int32, where map[string]interface{},
	whereDocument map[string]interface{}, include []QueryEnum, embedder embeddings.Embedder) ([]Document, error) {

	if len(include) == 0 {
		include = []QueryEnum{WithDocuments, WithEmbeddings, WithDistances, WithMetadatas}
	}
	queryEmbeddings, err := embedder.GetEmbeddings("", query)
	if err != nil {
		return nil, fmt.Errorf("error generating embeddings for query. Error: %w", err)
	}
	payload := map[string]any{
		"query_embeddings": [][]float32{queryEmbeddings},
		"query_texts":      query,
		"n_results":        numResults,
		"where":            where,
		"where_document":   whereDocument,
		"include":          include,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		c.serverBaseAddr+"/collections/"+c.ID+"/query",
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

	respObj := chromaQueryResultObject{}
	err = json.Unmarshal(bodyBuf, &respObj)
	if err != nil {
		return nil, err
	}
	return respObj.asFlattenedDocuments(), nil
}

func (c Collection) Count() (int, error) {
	resp, err := http.DefaultClient.Get(c.serverBaseAddr + "/collections/" + c.ID + "/count")
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	count, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(string(count))
}
