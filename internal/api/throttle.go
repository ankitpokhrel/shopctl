// See https://shopify.dev/docs/api/usage/rate-limits
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
)

const (
	shopifyAPIMaxCredits  = 2000
	shopifyAPIRestoreRate = 100

	// HeaderShopifyGQLQueryCost is the header key for the requested Shopify query cost.
	HeaderShopifyGQLQueryCost = "X-Shopify-GQL-Request-Cost"
)

// DefaultShopifyTransport is the default implementation of ShopifyTransport
// that extends HTTP transport with Shopify specific rate limiting.
var DefaultShopifyTransport http.RoundTripper = &ShopifyTransport{
	Transport: client.DefaultTransport,
	Throttler: NewShopifyThrottler(),
}

// ShopifyTransport extends HTTP transport with a Shopify specific rate limiting.
type ShopifyTransport struct {
	Transport http.RoundTripper
	Throttler *ShopifyThrottler

	latestResponseAt time.Time
}

// RoundTrip implements http.RoundTripper interface.
func (t *ShopifyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cost := req.Header.Get(HeaderShopifyGQLQueryCost)
	if cost == "" {
		return t.Transport.RoundTrip(req)
	}

	queryCost, err := strconv.Atoi(cost)
	if err != nil {
		return nil, err
	}

	// Wait until required credits are available.
	if err := t.Throttler.Wait(req.Context(), queryCost); err != nil {
		return nil, err
	}

	// Proceed with the request.
	res, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Date returned is supposed to be in UTC (same as time.Now()).
	processedAt, err := time.Parse(time.RFC1123, res.Header.Get("Date"))
	if err != nil {
		processedAt = time.Now()
	}

	// We only care about the latest response.
	if t.latestResponseAt.Before(processedAt) {
		t.latestResponseAt = processedAt

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(body)) // Rewind.

		var e struct {
			Extensions Extensions `json:"extensions"`
		}
		if err := json.Unmarshal(body, &e); err == nil {
			t.Throttler.UpdateStatus(e.Extensions.Cost.ThrottleStatus)
		}
	}

	return res, nil
}

// ShopifyThrottler is a Shopify specific throttler.
type ShopifyThrottler struct {
	mux     sync.RWMutex
	limiter *rate.Limiter
}

// NewShopifyThrottler creates a new Shopify based throttler.
func NewShopifyThrottler() *ShopifyThrottler {
	return &ShopifyThrottler{
		mux: sync.RWMutex{},
		limiter: rate.NewLimiter(
			rate.Limit(shopifyAPIRestoreRate), // Tokens per second.
			int(shopifyAPIMaxCredits),         // Max number of tokens a bucket can hold.
		),
	}
}

// Wait blocks until the required tokens are available.
func (t *ShopifyThrottler) Wait(ctx context.Context, n int) error {
	return t.limiter.WaitN(ctx, n)
}

// UpdateStatus updates the throttler status.
func (t *ShopifyThrottler) UpdateStatus(cost ThrottleStatus) {
	t.mux.Lock()
	defer t.mux.Unlock()

	// We'll try to match limiters tokens with Shopify's available tokens.
	//
	// In order to properly assign available tokens we get from the API response,
	// we need to drain the current tokens and reserve the available tokens.
	fulfill := cost.CurrentlyAvailable - t.limiter.Tokens()
	t.limiter.ReserveN(time.Now(), -int(fulfill))
}
