package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Request(t *testing.T) {
	var errResponse bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "12345", r.Header.Get("X-Shopify-Access-Token"))

		var (
			resp []byte
			err  error
		)

		if errResponse {
			resp, err = os.ReadFile("./testdata/error.json")
		} else {
			resp, err = os.ReadFile("./testdata/shop.json")
		}
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "12345")
	headers := Header{
		"Content-Type":           "application/json",
		"X-Shopify-Access-Token": "12345",
	}
	body := []byte(`{ "query": "{ shop { name }" }`)
	resp, err := client.Request(context.Background(), body, headers)
	assert.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	errResponse = true

	body = []byte(`{ "query": "{ shop { names }" }`)
	resp, err = client.Request(context.Background(), body, headers)
	assert.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
