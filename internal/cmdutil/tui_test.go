package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{
			name:     "it pads the string with spaces",
			input:    "test",
			length:   10,
			expected: "test      ",
		},
		{
			name:     "it returns the string as is",
			input:    "test",
			length:   4,
			expected: "test",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, Pad(tc.input, tc.length))
		})
	}
}

func TestShortenAndPad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{
			name:     "it shortens the string and pads with ellipsis",
			input:    "this is a long string",
			length:   10,
			expected: "this is a…",
		},
		{
			name:     "it returns the string as is",
			input:    "short text",
			length:   10,
			expected: "short text",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, ShortenAndPad(tc.input, tc.length))
		})
	}
}

func TestIsDumbTerminal(t *testing.T) {
	{
		t.Setenv("TERM", "")
		assert.True(t, IsDumbTerminal())
	}

	{
		t.Setenv("TERM", "dumb")
		assert.True(t, IsDumbTerminal())
	}

	{
		t.Setenv("TERM", "foo")
		assert.False(t, IsDumbTerminal())
	}

	{
		t.Setenv("WT_SESSION", "foo")
		assert.False(t, IsDumbTerminal())
	}
}

func TestGetEditor(t *testing.T) {
	// SHOPIFY_EDITOR and VISUAL is set.
	{
		t.Setenv("SHOPIFY_EDITOR", "nvim")
		t.Setenv("VISUAL", "nano")

		editor, args := GetEditor()
		assert.Equal(t, "nvim", editor)
		assert.Empty(t, args)

		t.Setenv("SHOPIFY_EDITOR", "")
		t.Setenv("VISUAL", "")
	}

	// VISUAL is set.
	{
		t.Setenv("VISUAL", "nvim -n")

		editor, args := GetEditor()
		assert.Equal(t, "nvim", editor)
		assert.Equal(t, []string{"-n"}, args)

		t.Setenv("VISUAL", "")
	}

	// EDITOR is set.
	{
		t.Setenv("EDITOR", "nano")

		editor, args := GetEditor()
		assert.Equal(t, "nano", editor)
		assert.Empty(t, args)

		t.Setenv("EDITOR", "")
	}

	// Env not set.
	{
		editor, args := GetEditor()
		assert.Equal(t, "/usr/bin/vim", editor)
		assert.Empty(t, args)
	}
}

func TestGetPager(t *testing.T) {
	// TERM is xterm, SHOPIFY_PAGER is not set, PAGER is set.
	{
		t.Setenv("TERM", "xterm")

		t.Setenv("PAGER", "")
		assert.Equal(t, "less", GetPager())

		t.Setenv("PAGER", "more")
		assert.Equal(t, "more", GetPager())

		t.Setenv("PAGER", "")
	}

	// TERM is set, SHOPIFY_PAGER is not set, PAGER is unset.
	{
		t.Setenv("TERM", "dumb")
		assert.Equal(t, "cat", GetPager())

		t.Setenv("TERM", "")
		assert.Equal(t, "cat", GetPager())

		t.Setenv("TERM", "xterm")
		assert.Equal(t, "less", GetPager())
	}

	// TERM is set, SHOPIFY_PAGER is set, PAGER is unset.
	{
		t.Setenv("SHOPIFY_PAGER", "bat")

		t.Setenv("TERM", "dumb")
		assert.Equal(t, "cat", GetPager())

		t.Setenv("TERM", "")
		assert.Equal(t, "cat", GetPager())

		t.Setenv("TERM", "xterm")
		assert.Equal(t, "bat", GetPager())
	}

	// TERM gets precedence if both PAGER and TERM are set.
	{
		t.Setenv("TERM", "")
		t.Setenv("PAGER", "")
		t.Setenv("SHOPIFY_PAGER", "")
		assert.Equal(t, "cat", GetPager())

		t.Setenv("PAGER", "more")
		t.Setenv("TERM", "dumb")
		assert.Equal(t, "cat", GetPager())

		t.Setenv("PAGER", "more")
		t.Setenv("TERM", "xterm")
		assert.Equal(t, "more", GetPager())
	}
}
