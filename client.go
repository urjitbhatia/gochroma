package chroma

import (
	"encoding/json"
	"fmt"
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
