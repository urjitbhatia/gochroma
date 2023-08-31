package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	url        string
	httpClient http.Client
}

type ChromaClient interface {
	Heartbeat() (int, error)
}

func NewClient(serverURL string) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	u = u.JoinPath("api/v1")
	c := &Client{url: u.String(), httpClient: http.Client{}}
	return c, err
}

func (c *Client) Heartbeat() (int, error) {
	resp, err := c.httpClient.Get(c.url + "/heartbeat")
	if err != nil {
		return -1, err
	}
	value := -1
	if resp.StatusCode != http.StatusOK {
		return value,
			fmt.Errorf("error getting server heartbeat. StatusCode: %d Response: %s",
				resp.StatusCode, resp.Body)
	}
	response := map[string]int{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response["nanosecond heartbeat"], err
}

func (c *Client) Reset() (bool, error) {
	resp, err := c.httpClient.Post(c.url+"/reset", "", nil)
	// Chroma returns just the string "true/false" if reset is enabled otherwise a json object with
	// an error string :facepalm:

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	switch string(body) {
	case "false":
		return false, nil
	case "true":
		return true, nil
	default:
		// it might be a json value
		response := map[string]any{}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return false, err
		}
		if errStr, ok := response["error"]; ok {
			err = fmt.Errorf("error reseting the db. ServerError: %s", errStr)
			return false, err
		}
	}
	return false, err
}

func (c *Client) GetVersion() (string, error) {
	resp, err := c.httpClient.Get(c.url + "/version")
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	// server response string is wrapped in double quotes, remove that
	body = bytes.ReplaceAll(body, []byte("\""), []byte(""))
	return string(body), err
}

func (c *Client) ListCollections() ([]Collection, error) {
	resp, err := c.httpClient.Get(c.url + "/collections")
	if err != nil {
		return nil, err
	}

	collections := []Collection{}
	err = json.NewDecoder(resp.Body).Decode(&collections)
	if err != nil {
		return nil, err
	}
	for _, collection := range collections {
		collection.cc = c
	}
	return collections, err
}

func (c *Client) GetOrCreateCollection(name string, distanceFn string, metadata map[string]any) (Collection, error) {
	return c.createCollection(name, distanceFn, metadata, true)
}

func (c *Client) CreateCollection(name string, distanceFn string, metadata map[string]any) (Collection, error) {
	return c.createCollection(name, distanceFn, metadata, false)
}
func (c *Client) createCollection(name string, distanceFn string, metadata map[string]any, getOrCreate bool) (Collection, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if distanceFn == "" {
		metadata["hnsw:space"] = "l2"
	} else {
		metadata["hnsw:space"] = strings.ToLower(distanceFn)
	}
	data := map[string]any{
		"name": name, "metadata": metadata, "get_or_create": getOrCreate,
	}
	collection := Collection{cc: c}

	reqBody, err := json.Marshal(data)
	if err != nil {
		return collection, err
	}

	resp, err := c.httpClient.Post(c.url+"/collections", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return collection, err
	}
	bodyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return collection, err
	}

	respJSON := map[string]any{}
	err = json.Unmarshal(bodyBuf, &respJSON)
	if err != nil {
		return collection, err
	}

	// Check response type
	if errStr, ok := respJSON["error"]; ok {
		return collection, fmt.Errorf("error while creating collection: %s", errStr)
	}
	// not an error, convert to collection type
	err = json.Unmarshal(bodyBuf, &collection)
	return collection, err
}

func (c *Client) DeleteCollection(name string) error {
	req, err := http.NewRequest(http.MethodDelete, c.url+"/collections/"+name, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	deleteResp := map[string]any{}
	err = json.NewDecoder(resp.Body).Decode(&deleteResp)
	if err != nil {
		return err
	}
	if errStr, ok := deleteResp["error"]; ok {
		return fmt.Errorf("error deleting collection: %s", errStr)
	}
	return nil
}

func (c *Client) GetCollection(name string) (Collection, error) {
	resp, err := c.httpClient.Get(c.url + "/collections/" + name)
	if err != nil {
		return Collection{}, err
	}

	collection := Collection{cc: c}
	err = json.NewDecoder(resp.Body).Decode(&collection)
	if err != nil {
		return Collection{}, err
	}
	return collection, nil
}
