package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
