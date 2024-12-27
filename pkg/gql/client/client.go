package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"maps"
	"net"
	"net/http"
	"time"
)

const timeout = 15 * time.Second

// Header is a key, value pair for request headers.
type Header map[string]string

// Client is a GraphQL client.
type Client struct {
	server    string
	token     string
	transport http.RoundTripper
}

// NewClient creates a new GraphQL client.
func NewClient(server, token string) *Client {
	client := Client{
		server: server,
		token:  token,
		transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
		},
	}

	return &client
}

// Request sends POST request to a GraphQL server.
func (c *Client) Request(ctx context.Context, body []byte, headers Header) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.server, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	reqHeaders := Header{
		"Content-Type":           "application/json",
		"X-Shopify-Access-Token": c.token,
	}
	if headers != nil {
		maps.Copy(headers, reqHeaders)
		reqHeaders = headers
	}

	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}
	httpClient := &http.Client{Transport: c.transport}

	return httpClient.Do(req.WithContext(ctx))
}
