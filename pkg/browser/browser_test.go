package browser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserENVPrecedence(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(t *testing.T)
		expected string
	}{
		{
			name: "it uses SHOPIFY_BROWSER env",
			setup: func(t *testing.T) {
				t.Setenv("SHOPIFY_BROWSER", "firefox")
			},
			expected: "firefox",
		},
		{
			name: "it uses BROWSER env",
			setup: func(t *testing.T) {
				t.Setenv("BROWSER", "chrome")
			},
			expected: "chrome",
		},
		{
			name: "SHOPIFY_BROWSER gets precedence over BROWSER env if both are set",
			setup: func(t *testing.T) {
				t.Setenv("BROWSER", "chrome")
				t.Setenv("SHOPIFY_BROWSER", "firefox")
			},
			expected: "firefox",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			assert.Equal(t, tc.expected, getBrowserFromENV())
		})
	}
}
