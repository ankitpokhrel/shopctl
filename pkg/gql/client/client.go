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

	"github.com/hashicorp/go-retryablehttp"
)

const (
	maxIdleConns    = 100
	connTimeout     = 15 * time.Second
	idleConnTimeout = 90 * time.Second
	handsakeTimeout = 10 * time.Second

	minWait  = 1 * time.Second
	maxWait  = 60 * time.Second
	retryMax = 5
)

// DefaultTransport is the default HTTP transport for the client.
var DefaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	TLSClientConfig: &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	},
	DialContext:           (&net.Dialer{Timeout: connTimeout}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          maxIdleConns,
	IdleConnTimeout:       idleConnTimeout,
	TLSHandshakeTimeout:   handsakeTimeout,
	ExpectContinueTimeout: 1 * time.Second,
}

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
	http      *retryablehttp.Client
	transport http.RoundTripper
	logger    retryablehttp.LeveledLogger
}

// ClientFunc is a functional option for Client.
type ClientFunc func(*Client)

// NewClient creates a new GraphQL client.
func NewClient(server, token string, opts ...ClientFunc) *Client {
	c := Client{
		server: server,
		token:  token,
		http:   retryablehttp.NewClient(),
	}

	for _, opt := range opts {
		opt(&c)
	}

	if c.transport == nil {
		c.http.HTTPClient.Transport = DefaultTransport
	} else {
		c.http.HTTPClient.Transport = c.transport
	}

	c.http.Logger = c.logger
	c.http.RequestLogHook = func(l retryablehttp.Logger, r *http.Request, n int) {
		if n > 0 {
			l.Printf("Retrying as request was throttled or server didn't respond in time, attempt #%d", n)
		}
	}
	c.http.ResponseLogHook = func(l retryablehttp.Logger, r *http.Response) {
		if r.StatusCode == http.StatusTooManyRequests {
			retryAfter := r.Header.Get("Retry-After")
			l.Printf("Request was throttled by the server, will retry after %s seconds", retryAfter)
		}
	}
	c.http.RetryMax = retryMax
	c.http.RetryWaitMin = minWait
	c.http.RetryWaitMax = maxWait
	return &c
}

// WithTransport sets custom transport for the client.
func WithTransport(transport http.RoundTripper) ClientFunc {
	return func(c *Client) {
		c.transport = transport
	}
}

// WithLogger sets custom logger for the client.
func WithLogger(logger retryablehttp.LeveledLogger) ClientFunc {
	return func(c *Client) {
		c.logger = logger
	}
}

// Request sends POST request to a GraphQL server.
func (c *Client) Request(ctx context.Context, body []byte, headers Header) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(http.MethodPost, c.server, bytes.NewReader(body))
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
	return c.http.Do(req.WithContext(ctx))
}

// Execute sends a GraphQL request and decodes the response to the given result.
func (c Client) Execute(ctx context.Context, payload GQLRequest, headers Header, result any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL query: %w", err)
	}

	res, err := c.Request(ctx, data, headers)
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
