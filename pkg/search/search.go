package search

import (
	"fmt"
	"strings"
)

// Query is a Shopify search query builder.
type Query struct {
	conditions []string
}

// New returns a new instance of Query.
func New() *Query {
	return &Query{
		conditions: []string{},
	}
}

// Add appends a raw condition to the query.
func (q *Query) Add(condition string) *Query {
	q.conditions = append(q.conditions, condition)
	return q
}

// Eq adds an equality condition for a field.
func (q *Query) Eq(field, value string) *Query {
	condition := fmt.Sprintf("%s:%s", field, escape(value))
	q.conditions = append(q.conditions, condition)
	return q
}

// Neq adds a not-equal condition (using a minus prefix) for a field.
func (q *Query) Neq(field, value string) *Query {
	condition := fmt.Sprintf("-%s:%s", field, escape(value))
	q.conditions = append(q.conditions, condition)
	return q
}

// Gt adds a greater-than condition.
// For example, Gt("price", 10) produces "price:>10".
func (q *Query) Gt(field string, value any) *Query {
	condition := fmt.Sprintf("%s:>%v", field, value)
	q.conditions = append(q.conditions, condition)
	return q
}

// Lt adds a less-than condition.
func (q *Query) Lt(field string, value any) *Query {
	condition := fmt.Sprintf("%s:<%v", field, value)
	q.conditions = append(q.conditions, condition)
	return q
}

// Gte adds a greater-than-or-equal condition.
func (q *Query) Gte(field string, value any) *Query {
	condition := fmt.Sprintf("%s:>=%v", field, value)
	q.conditions = append(q.conditions, condition)
	return q
}

// Lte adds a less-than-or-equal condition.
func (q *Query) Lte(field string, value any) *Query {
	condition := fmt.Sprintf("%s:<=%v", field, value)
	q.conditions = append(q.conditions, condition)
	return q
}

// Contains adds a condition that checks if a field contains the given substring.
// It simulates a “contains” search by wrapping the value with wildcards.
func (q *Query) Contains(field, value string) *Query {
	condition := fmt.Sprintf("%s:*%s*", field, value)
	q.conditions = append(q.conditions, condition)
	return q
}

// In adds a condition that checks if a field matches any one of the provided values.
// It groups the OR conditions together. For example, In("product_type", "shirt", "sweater")
// produces: (product_type:shirt OR product_type:sweater)
func (q *Query) In(field string, values ...string) *Query {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%s:%s", field, escape(v)))
	}
	group := fmt.Sprintf("(%s)", strings.Join(parts, " OR "))
	q.conditions = append(q.conditions, group)
	return q
}

// And adds an explicit AND operator.
func (q *Query) And() *Query {
	q.conditions = append(q.conditions, "AND")
	return q
}

// Or adds an explicit OR operator.
func (q *Query) Or() *Query {
	q.conditions = append(q.conditions, "OR")
	return q
}

// Group groups conditions together by accepting a function that builds a sub-query.
// The grouped conditions are wrapped in parentheses.
func (q *Query) Group(fn func(sub *Query)) *Query {
	subQuery := New()
	fn(subQuery)
	group := fmt.Sprintf("(%s)", subQuery.Build())
	q.conditions = append(q.conditions, group)
	return q
}

// Build constructs the final query string.
func (q *Query) Build() string {
	return strings.TrimSpace(strings.Join(q.conditions, " "))
}

// escape checks if the value contains whitespace and, if so,
// wraps it in quotes. It also escapes any inner double quotes.
func escape(value string) string {
	if strings.ContainsAny(value, " \t") {
		escaped := strings.ReplaceAll(value, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return value
}
