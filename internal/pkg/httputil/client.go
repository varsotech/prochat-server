package httputil

import (
	"net/http"
	"time"
)

type Client struct {
	client    http.Client
	userAgent string
}

func NewClient() *Client {
	c := http.Client{
		Timeout:   10 * time.Second,
		Transport: PublicOnlyTransport(),
	}

	return &Client{
		client:    c,
		userAgent: "prochat-sdk",
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.userAgent)
	return c.client.Do(req)
}
