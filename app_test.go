package shopctl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShopifyProductID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "valid shopify product id",
			id:   "gid://shopify/Product/8737843085536",
			want: "gid://shopify/Product/8737843085536",
		},
		{
			name: "numeric id",
			id:   "8737843085536",
			want: "gid://shopify/Product/8737843085536",
		},
		{
			name: "non numeric id",
			id:   "invalid",
			want: "",
		},
		{
			name: "invalid shopify product id",
			id:   "gid://shopify/8737843085536",
			want: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ShopifyProductID(tc.id))
		})
	}
}

func TestExtractNumericID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "shopify product id",
			id:   "gid://shopify/Product/8737843085536",
			want: "8737843085536",
		},
		{
			name: "shopify product variant id",
			id:   "gid://shopify/ProductVariant/8737843085536",
			want: "8737843085536",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ExtractNumericID(tc.id))
		})
	}
}
