package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuild(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Query
		expected string
	}{
		{
			name: "simple query",
			build: func() *Query {
				return New().Eq("title", "red shirt")
			},
			expected: `title:"red shirt"`,
		},
		{
			name: "AND and greater than condition",
			build: func() *Query {
				return New().
					Eq("title", "red shirt").
					And().
					Gt("price", 10)
			},
			expected: `title:"red shirt" AND price:>10`,
		},
		{
			name: "in clause",
			build: func() *Query {
				return New().In("product_type", "t-shirt", "sweater", "sport shoes")
			},
			expected: `(product_type:t-shirt OR product_type:sweater OR product_type:"sport shoes")`,
		},
		{
			name: "complex query with grouping",
			build: func() *Query {
				return New().
					Group(func(sub *Query) {
						sub.Eq("title", "red shirt").
							And().
							Contains("description", "cotton")
					}).
					And().
					Lte("price", 20)
			},
			// The outer group wraps the subquery.
			expected: `(title:"red shirt" AND description:*cotton*) AND price:<=20`,
		},
		{
			name: "not equal with OR",
			build: func() *Query {
				return New().
					Neq("status", "sold out").
					Or().
					Eq("status", "available")
			},
			expected: `-status:"sold out" OR status:available`,
		},
		{
			name: "query with GreaterThanOrEqual and LessThan condition",
			build: func() *Query {
				return New().
					Gte("rating", 4).
					And().
					Lt("rating", 5)
			},
			expected: `rating:>=4 AND rating:<5`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.build()

			assert.Equal(t, tc.expected, q.Build())
		})
	}
}
