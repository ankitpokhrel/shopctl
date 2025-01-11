package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"net/http"
	"time"
)

const timeout = 15 * time.Second

// Header is a key, value pair for request headers.
type Header map[string]string

// QueryVars is a map of query variables.
type QueryVars map[string]any

// GQLRequest is a GraphQL request.
type GQLRequest struct {
	Query     string    `json:"query"`
	Variables QueryVars `json:"variables,omitempty"`
}

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

// Execute sends a GraphQL request and decodes the response to the given result.
func (c Client) Execute(ctx context.Context, payload GQLRequest, result any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL query: %w", err)
	}

	res, err := c.Request(ctx, data, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if res == nil {
		return fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}
