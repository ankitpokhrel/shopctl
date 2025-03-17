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
			name: "diffs with all elements in sort order",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			order:    []string{"str", "num", "f64"},
			expected: "--- a/str\n--- a/num\n--- a/f64\n",
		},
		{
			name: "diffs with some elements in sort order",
			diffs: map[string]string{
				"str": "--- a/str",
				"num": "--- a/num",
				"f64": "--- a/f64",
			},
			order:    []string{"num", "str"},
			expected: "--- a/num\n--- a/str\n--- a/f64\n",
		},
	}

	// Set environment variable.
	t.Setenv("SHOPIFY_DIFF_TOOL", "cat")

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
