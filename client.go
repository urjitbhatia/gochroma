package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	u.Path = "/api/v1"
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

func (c *Client) ListCollections() (map[string]any, error) {
	resp, err := c.httpClient.Get(c.url + "/collections")
	if err != nil {
		return nil, err
	}
	collections := map[string]any{}
	json.NewDecoder(resp.Body).Decode(&collections)
	return collections, nil
}
