package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	// Secret used for signing
	secret := "my-secret-key"

	// Simulated query parameters
	params := url.Values{
		"code":      []string{"ABC3EZB97"},
		"shop":      []string{"example-shop.myshopify.com"},
		"timestamp": []string{"1234567890"},
	}

	// Compute the expected HMAC signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(params.Encode()))
	expectedHMAC := hex.EncodeToString(h.Sum(nil))

	// Add the HMAC to the query parameters to simulate a Shopify request
	params.Set("hmac", expectedHMAC)

	// Create a fake HTTP request with the query parameters
	req := httptest.NewRequest("GET", "/?"+params.Encode(), nil)

	assert.True(t, verifySignature(req, secret))

	// Test an invalid case: modify the HMAC in the request
	params.Set("hmac", "invalid-hmac")
	req = httptest.NewRequest("GET", "/?"+params.Encode(), nil)

	// Call verifySignature
	assert.False(t, verifySignature(req, secret))
}

func TestGenerateState(t *testing.T) {
	state, err := generateState(3)
	assert.NoError(t, err)
	assert.Len(t, state, 6)

	state, err = generateState(8)
	assert.NoError(t, err)
	assert.Len(t, state, 16)

	state, err = generateState(16)
	assert.NoError(t, err)
	assert.Len(t, state, 32)
}

func TestValidateShopURL(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid shop URL",
			url:      "https://example-shop.myshopify.com",
			expected: true,
		},
		{
			name:     "valid shop URL with trailing slash",
			url:      "https://example-shop.myshopify.com/",
			expected: true,
		},
		{
			name:     "invalid shop URL",
			url:      "https://invalid-shop_.myshopify.com",
			expected: false,
		},
		{
			name:     "invalid shop URL with trailing slash",
			url:      "https://invalid-shop_.myshopify.com/",
			expected: false,
		},
		{
			name:     "valid URL but with invalid scheme",
			url:      "ftp://example.myshopify.com/",
			expected: false,
		},
		{
			name:     "invalid URL",
			url:      "https://shop.example.com",
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, validateShopURL(tc.url))
		})
	}
}
