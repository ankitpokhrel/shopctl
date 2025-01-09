package cmdutil

import (
	"testing"
	"time"

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
			if got := ShopifyProductID(tc.id); got != tc.want {
				t.Errorf("ShopifyProductID() = %v, want %v", got, tc.want)
			}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractNumericID(tt.id); got != tt.want {
				t.Errorf("ExtractNumericID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDateTimeHuman(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		format   func() string
		expected string
	}{
		{
			name: "it returns input date for invalid date input",
			format: func() string {
				return FormatDateTimeHuman("2024-12-03 10:00:00", time.RFC3339)
			},
			expected: "2024-12-03 10:00:00",
		},
		{
			name: "it returns input date for invalid input format",
			format: func() string {
				return FormatDateTimeHuman("2025-01-10 10:00:00", "invalid")
			},
			expected: "2025-01-10 10:00:00",
		},
		{
			name: "it format input date from RFC3339 date format",
			format: func() string {
				return FormatDateTimeHuman("2025-01-10T16:12:00.000Z", time.RFC3339)
			},
			expected: "Fri, 10 Jan 25",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.format())
		})
	}
}

func TestGetStoreSlug(t *testing.T) {
	tests := []struct {
		name  string
		store string
		want  string
	}{
		{
			name:  "empty store url",
			store: "",
			want:  "",
		},
		{
			name:  "valid store url without protocol",
			store: "store1.myshopify.com",
			want:  "store1",
		},
		{
			name:  "valid store url with http protocol",
			store: "http://store2.myshopify.com",
			want:  "store2",
		},
		{
			name:  "valid store url with https protocol",
			store: "https://store3.myshopify.com",
			want:  "store3",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, GetStoreSlug(tc.store))
		})
	}
}
