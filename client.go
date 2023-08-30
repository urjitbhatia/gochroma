package chroma

import "net/url"

type Client struct {
	url *url.URL
}

func NewClient(serverURL string) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	c := &Client{url: u}
	return c, err
}
