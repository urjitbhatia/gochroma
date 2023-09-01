package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

var openAIURL = "https://api.openai.com/v1/"
var openAIEmbeddingsURL = openAIURL + "/embeddings"
var openAIAuth string

type embeddingResponse struct {
	Data struct {
		Object    string
		Embedding []float32
	}
	Model string
	Usage struct {
		PromptTokens int
		TotalTokens  int
	}
}

type OpenAIClient struct {
	client     *http.Client
	authHeader string
}

func NewOpenAIClient(key string) OpenAIClient {
	return OpenAIClient{
		client:     http.DefaultClient,
		authHeader: fmt.Sprintf("Bearer: %s", key),
	}
}

func (o *OpenAIClient) GetEmbeddings(id string, content string) ([]float32, error) {
	body, err := json.Marshal(map[string]string{
		"model": "text-embedding-ada-002",
		"input": content,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, openAIEmbeddingsURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", openAIAuth)

	resp, err := http.DefaultClient.Do(req)
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
	err = json.NewDecoder(resp.Body).Decode(&er)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling openai embeddings response: %w", err)
	}

	log.Debug().
		Str("documentID", id).
		Str("embeddingModelUsed", er.Model).
		Int("promptTokensUsed", er.Usage.PromptTokens).
		Int("totalTokensUsed", er.Usage.TotalTokens).
		Msg("openai embedding token usage")

	return er.Data.Embedding, nil
}
