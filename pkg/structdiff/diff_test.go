package structdiff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDiff(t *testing.T) {
	cases := []struct {
		name     string
		a, b     any
		expected map[string]string
	}{
		{
			name:     "empty",
			a:        nil,
			b:        nil,
			expected: map[string]string{},
		},
		{
			name:     "different types",
			a:        struct{ v string }{v: "test"},
			b:        struct{ x bool }{x: true},
			expected: map[string]string{},
		},
		{
			name:     "same",
			a:        struct{ v string }{v: "test"},
			b:        struct{ v string }{v: "test"},
			expected: map[string]string{},
		},
		{
			name: "different data types",
			a: struct {
				str string
				num int
				b   bool
				f32 float32
				f64 float64
			}{str: "test", num: 1, b: true, f64: 3.14},
			b: struct {
				str string
				num int
				b   bool
				f32 float32
				f64 float64
			}{str: "test 123", num: 5, b: true, f32: 1.23, f64: 3.159},
			expected: map[string]string{
				"str": `--- a/str
+++ b/str
@@ -1,1 +0,0 @@
-test
@@ -0,0 +1,1 @@
+test 123
`,
				"num": `--- a/num
+++ b/num
@@ -1,1 +0,0 @@
-1
@@ -0,0 +1,1 @@
+5
`,
				"f32": `--- a/f32
+++ b/f32
@@ -1,1 +0,0 @@
-0
@@ -0,0 +1,1 @@
+1.23
`,
				"f64": `--- a/f64
+++ b/f64
@@ -1,1 +0,0 @@
-3.14
@@ -0,0 +1,1 @@
+3.159
`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Get(tc.b, tc.a))
		})
	}
}

func TestIsEmptyDiff(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "empty",
			input: `--- CreatedAt
		+++ b/CreatedAt`,
			expected: true,
		},
		{
			name: "not empty",
			input: `--- ProductType
+++ b/ProductType
@@ -1,1 +0,0 @@
-accessories
@@ -0,0 +1,1 @@
+T-Shirts`,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isEmptyDiff(tc.input))
		})
	}
}
