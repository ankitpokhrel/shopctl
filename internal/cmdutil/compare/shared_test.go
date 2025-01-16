package compare

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	cases := []struct {
		name     string
		diffs    map[string]string
		order    []string
		expected string
	}{
		{
			name:     "no diffs",
			diffs:    map[string]string{},
			order:    []string{},
			expected: "",
		},
		{
			name: "diffs with no sort order",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			order:    []string{},
			expected: "--- a/str\n--- a/num\n--- a/f64\n",
		},
		{
			name: "diffs with sort order",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			order:    []string{"num", "str"},
			expected: "--- a/num\n--- a/str\n--- a/f64\n",
		},
	}

	// Set environment variable
	_ = os.Setenv("SHOPIFY_DIFF_TOOL", "cat")
	defer func() { _ = os.Unsetenv("SHOPIFY_DIFF_TOOL") }()

	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = stdout }()

			err := Render(tc.diffs, tc.order)

			_ = w.Close()

			var buf bytes.Buffer
			_, errRead := buf.ReadFrom(r)
			assert.NoError(t, errRead)

			// Verify results
			assert.NoError(t, err)
			assert.Contains(t, tc.expected, buf.String())
		})
	}
}

func TestTrim(t *testing.T) {
	cases := []struct {
		name     string
		diffs    map[string]string
		cutset   []string
		expected map[string]string
	}{
		{
			name:     "no diffs",
			diffs:    map[string]string{},
			cutset:   []string{"str", "f64"},
			expected: map[string]string{},
		},
		{
			name: "nothing to trim",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			cutset: []string{},
			expected: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
		},
		{
			name: "item to trim not in the set",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			cutset: []string{"random", "invalid"},
			expected: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
		},
		{
			name: "it trims the diff",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			cutset: []string{"str", "f64"},
			expected: map[string]string{
				"num": "--- a/num",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			Trim(tc.diffs, tc.cutset)
			assert.Equal(t, tc.expected, tc.diffs)
		})
	}
}
