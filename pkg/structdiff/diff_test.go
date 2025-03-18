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
		{
			name: "slice",
			a: struct {
				str   string
				slice []int
			}{str: "test", slice: []int{1, 2, 3}},
			b: struct {
				str   string
				slice []int
			}{str: "test", slice: []int{3, 4, 5}},
			expected: map[string]string{
				"slice": `--- a/slice
+++ b/slice
@@ -3,3 +3,3 @@
-1
+3
-2
+4
-3
+5
`,
			},
		},
		{
			name: "pointers",
			a: struct {
				str    string
				ptr    *string
				nilptr *int
				custom *struct{ B bool }
			}{
				str:    "test",
				ptr:    func() *string { test := "failed"; return &test }(),
				nilptr: func() *int { test := 42; return &test }(),
			},
			b: struct {
				str    string
				ptr    *string
				nilptr *int
				custom *struct{ B bool }
			}{
				str:    "test",
				ptr:    func() *string { test := "passed"; return &test }(),
				nilptr: nil,
				custom: func() *struct{ B bool } { test := struct{ B bool }{B: true}; return &test }(),
			},
			expected: map[string]string{
				"ptr": `--- a/ptr
+++ b/ptr
@@ -1,1 +0,0 @@
-failed
@@ -0,0 +1,1 @@
+passed
`,
				"nilptr": `--- a/nilptr
+++ b/nilptr
@@ -1,1 +0,0 @@
-42
`,
				"custom": `--- a/custom
+++ b/custom
@@ -0,0 +1,1 @@
+B: true
`,
			},
		},
		{
			name: "direct nested types",
			a: struct {
				str    string
				num    int
				nested struct{ B bool }
			}{str: "test", num: 1, nested: struct{ B bool }{B: false}},
			b: struct {
				str    string
				num    int
				nested struct{ B bool }
			}{str: "test 123", num: 5, nested: struct{ B bool }{B: true}},
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
				"nested": `--- a/nested
+++ b/nested
@@ -1,1 +0,0 @@
-B: false
@@ -0,0 +1,1 @@
+B: true
`,
			},
		},
		{
			name: "indirect nested types",
			a: struct {
				Nested []struct {
					B bool
					c rune
					D float64
				}
			}{Nested: []struct {
				B bool
				c rune
				D float64
			}{{B: false, D: 3.14}, {B: true, D: 42}}},
			b: struct {
				Nested []struct {
					B bool
					c rune
					D float64
				}
			}{
				Nested: []struct {
					B bool
					c rune
					D float64
				}{{B: false}},
			},
			expected: map[string]string{
				"Nested": `--- a/Nested
+++ b/Nested
@@ -3,3 +1,1 @@
-D: 3.14
+D: 0
-B: true
-D: 42
`,
			},
		},
		{
			name: "indirect nested types 2",
			a: struct {
				str        string
				num        int
				unexported struct{ B bool }
				Nested     []struct {
					B bool
					c rune
					D float64
				}
			}{str: "test", num: 1, Nested: []struct {
				B bool
				c rune
				D float64
			}{{B: false, D: 3.14}}},
			b: struct {
				str        string
				num        int
				unexported struct{ B bool }
				Nested     []struct {
					B bool
					c rune
					D float64
				}
			}{
				str: "test 123",
				num: 5,
				Nested: []struct {
					B bool
					c rune
					D float64
				}{
					{B: true, c: 'c', D: 3.159}, {B: false},
				},
			},
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
				"Nested": `--- a/Nested
+++ b/Nested
@@ -2,2 +4,4 @@
-B: false
-D: 3.14
+B: true
+D: 3.159
+B: false
+D: 0
`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := New(tc.b, tc.a)
			assert.Equal(t, tc.expected, d.Get())
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
