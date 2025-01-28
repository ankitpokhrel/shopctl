package engine

import "testing"

func TestGetHashDir(t *testing.T) {
	tests := []struct {
		name string
		of   string
		want string
	}{
		{
			name: "random string",
			of:   "test",
			want: "09",
		},
		{
			name: "product id",
			of:   "8737843085536",
			want: "e1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHashDir(tt.of); got != tt.want {
				t.Errorf("GetHashDir() = %v, want %v", got, tt.want)
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
