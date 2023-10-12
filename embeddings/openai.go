package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

var openAIURL = "https://api.openai.com/v1/"
var openAIEmbeddingsPath = "/embeddings"

type embeddingResponse struct {
	Data []struct {
		Object    string    `json:"object,omitempty"`
		Index     int       `json:"index,omitempty"`
		Embedding []float32 `json:"embedding,omitempty"`
	} `json:"data"`
	Model string `json:"model,omitempty"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens,omitempty"`
		TotalTokens  int `json:"total_tokens,omitempty"`
	} `json:"usage"`
}

type OpenAIClient struct {
	client         *http.Client
	authHeader     string
	openAIEndpoint string
}

func NewOpenAIClient(key string) OpenAIClient {
	return NewOpenAIClientWithHTTP(openAIURL, key, http.DefaultClient)
}

func NewOpenAIClientWithHTTP(openAIEndpoint, key string, client *http.Client) OpenAIClient {
	if openAIEndpoint == "" {
		openAIEndpoint = openAIURL
	}
	return OpenAIClient{
		client:         client,
		authHeader:     fmt.Sprintf("Bearer %s", key),
		openAIEndpoint: openAIEndpoint,
	}
}

func (o *OpenAIClient) EmbedQuery(ctx context.Context, content string) ([]float32, error) {
	embeddings, err := o.EmbedDocuments(ctx, []string{content})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

func (o *OpenAIClient) EmbedDocuments(_ context.Context, content []string) ([][]float32, error) {
	body, err := json.Marshal(map[string]any{
		"model": "text-embedding-ada-002",
		"input": content,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, o.openAIEndpoint+openAIEmbeddingsPath, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", o.authHeader)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error getting openai embeddings and unable to read response body. status: %s", resp.Status)
		}
		return nil, fmt.Errorf("error getting openai embeddings. Status: %s Response: %s", resp.Status, string(body))
	}

	er := embeddingResponse{}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading openai embeddings response body: %w", err)
	}

	err = json.Unmarshal(respBody, &er)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling openai embeddings response: %w\nresponse body: %s", err, string(respBody))
	}

	log.Debug().
		Str("endpoint", o.openAIEndpoint).
		Str("embeddingModelUsed", er.Model).
		Int("promptTokensUsed", er.Usage.PromptTokens).
		Int("totalTokensUsed", er.Usage.TotalTokens).
		Msg("openai embedding token usage")

	if len(er.Data) == 0 {
		return nil, fmt.Errorf("something went wrong, got no embeddings from openai")
	}
	var embeddings [][]float32
	for _, data := range er.Data {
		embeddings = append(embeddings, data.Embedding)
	}
	return embeddings, nil
}
