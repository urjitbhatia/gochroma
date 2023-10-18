package chroma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/urjitbhatia/gochroma/embeddings"
	"io"
	"net/http"
	"strconv"
)

type Collection struct {
	Name       string         `json:"name"`
	ID         string         `json:"id"`
	Metadata   map[string]any `json:"metadata"`
	DistanceFn string         `json:"distanceFn"`

	server Server
}

// CollectionWithSrv creates a collection with the given chroma server as the backend
// mainly used for unit testing without having to run a full chromadb server
func CollectionWithSrv(s Server) Collection {
	return Collection{server: s}
}

type Document struct {
	ID         string
	Embeddings []float32
	Metadata   map[string]any
	Content    string
	Distance   float32 // set when documents are returned from a similarity search - distance to query
}

type chromaCollectionObject struct {
	Embeddings [][]float32      `json:"embeddings"`
	Metadatas  []map[string]any `json:"metadatas"`
	Documents  []string         `json:"documents"`
	IDs        []string         `json:"ids"`
}

type chromaQueryResultObject struct {
	Embeddings [][][]float32      `json:"embeddings"`
	Distances  [][]float64        `json:"distances"`
	Metadatas  [][]map[string]any `json:"metadatas"`
	Documents  [][]string         `json:"documents"`
	IDs        [][]string         `json:"ids"`
}

func (o chromaQueryResultObject) asFlattenedDocuments() []Document {
	var docs []Document
	for i := 0; i < len(o.IDs); i++ {
		for j := 0; j < len(o.IDs[i]); j++ {
			d := Document{
				ID:       o.IDs[i][j],
				Content:  o.Documents[i][j],
				Distance: float32(o.Distances[i][j]),
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
		Embeddings: [][]float32{},
		Metadatas:  []map[string]any{},
		Documents:  []string{},
		IDs:        []string{},
	}

	var docsToFetchEmbeddings []Document
	var docsWithEmbeddings []Document
	for _, doc := range docs {
		if len(doc.Embeddings) == 0 {
			docsToFetchEmbeddings = append(docsToFetchEmbeddings, doc)
		} else {
			docsWithEmbeddings = append(docsWithEmbeddings, doc)
		}
	}

	// get embeddings for docs that are missing embeddings using batch calls for efficiency
	for _, batch := range SliceBatch(docsToFetchEmbeddings, 10) {
		contents := make([]string, len(batch))
		for i, doc := range batch {
			contents[i] = doc.Content
		}
		addReq.Documents = append(addReq.Documents, contents...)

		embedVectors, err := embedder.EmbedDocuments(context.Background(), contents)
		if err != nil {
			return err
		}
		addReq.Embeddings = append(addReq.Embeddings, embedVectors...)

		for _, doc := range batch {
			addReq.Metadatas = append(addReq.Metadatas, doc.Metadata)
			addReq.IDs = append(addReq.IDs, doc.ID)
		}
	}

	for i, doc := range docsWithEmbeddings {
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
		fmt.Sprintf("%s/collections/%s/add", c.server.BaseUrl(), c.ID),
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
		c.server.BaseUrl()+"/collections/"+c.ID+"/get",
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
	queryEmbeddings, err := embedder.EmbedQuery(context.Background(), query)
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
		c.server.BaseUrl()+"/collections/"+c.ID+"/query",
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
		return nil, fmt.Errorf("failed unpacking chroma query result: %w", err)
	}
	return respObj.asFlattenedDocuments(), nil
}

func (c Collection) Count() (int, error) {
	resp, err := http.DefaultClient.Get(c.server.BaseUrl() + "/collections/" + c.ID + "/count")
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

func SliceBatch[T any](items []T, chunkSize int) [][]T {
	batches := make([][]T, 0, (len(items)+chunkSize-1)/chunkSize)
	for chunkSize < len(items) {
		items, batches = items[chunkSize:], append(batches, items[0:chunkSize:chunkSize])
	}
	return append(batches, items)
}
