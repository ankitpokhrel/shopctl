package introspect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodes_ToGoTypes(t *testing.T) {
	cases := []struct {
		name     string
		schema   IntrospectionSchema
		expected string
	}{
		{
			name: "simple schema",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Name: "ID", Kind: SCALAR}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String"}},
							{Name: "age", Type: TypeRef{Kind: SCALAR, Name: "Int", OfType: &TypeRef{Name: "Int", Kind: SCALAR}}},
							{Name: "grade", Type: TypeRef{Kind: SCALAR, Name: "Float", OfType: &TypeRef{Name: "Float", Kind: SCALAR}}},
							{Name: "isStudent", Type: TypeRef{Kind: SCALAR, Name: "Boolean", OfType: &TypeRef{Name: "Boolean", Kind: SCALAR}}},
							{Name: "hobbies", Type: TypeRef{Kind: LIST, OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "interests", Type: TypeRef{Kind: LIST, Name: "String"}},
						},
					},
				},
			},
			expected: `type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name *string ` + "`json:\"name,omitempty\"`" + `
	Age int ` + "`json:\"age\"`" + `
	Grade float64 ` + "`json:\"grade\"`" + `
	IsStudent bool ` + "`json:\"isStudent\"`" + `
	Hobbies []string ` + "`json:\"hobbies\"`" + `
	Interests *[]any ` + "`json:\"interests,omitempty\"`" + `
}
`,
		},
		{
			name: "schema with nested types",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Name: "ID", Kind: SCALAR}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Name: "String", Kind: SCALAR}}},
							{Name: "address", Type: TypeRef{Kind: OBJECT, Name: "Address"}},
						},
					},
					{
						Name: "Address",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "street", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Name: "String", Kind: SCALAR}}},
							{Name: "city", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Name: "String", Kind: SCALAR}}},
						},
					},
				},
			},
			expected: `type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Address *Address ` + "`json:\"address,omitempty\"`" + `
}

type Address struct {
	Street string ` + "`json:\"street\"`" + `
	City string ` + "`json:\"city\"`" + `
}
`,
		},
		{
			name: "schema with built-in types",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "user", Type: TypeRef{Kind: OBJECT, Name: "User", OfType: &TypeRef{Name: "User", Kind: OBJECT}}},
						},
					},
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Name: "ID", Kind: SCALAR}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Name: "String", Kind: SCALAR}}},
						},
					},
				},
			},
			expected: `type Query struct {
	User User ` + "`json:\"user\"`" + `
}

type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`,
		},
		{
			name: "schema with built-in types and custom types",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "users", Type: TypeRef{Kind: LIST, OfType: &TypeRef{Kind: OBJECT, Name: "User"}}},
						},
					},
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "address", Type: TypeRef{Kind: OBJECT, Name: "Address", OfType: &TypeRef{Kind: OBJECT, Name: "Address"}}},
						},
					},
					{
						Name: "Address",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "street", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "city", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
						},
					},
				},
			},
			expected: `type Query struct {
	Users []User ` + "`json:\"users\"`" + `
}

type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Address Address ` + "`json:\"address\"`" + `
}

type Address struct {
	Street string ` + "`json:\"street\"`" + `
	City string ` + "`json:\"city\"`" + `
}
`,
		},
		{
			name: "schema with built-in types and custom types (with nested custom types)",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "user", Type: TypeRef{Kind: OBJECT, Name: "User", OfType: &TypeRef{Kind: OBJECT, Name: "User"}}},
						},
					},
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "address", Type: TypeRef{Kind: OBJECT, Name: "Address"}},
						},
					},
					{
						Name: "Address",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "street", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "city", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "country", Type: TypeRef{Kind: ENUM, Name: "Country", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
						},
					},
					{
						Name: "Country",
						Kind: ENUM,
						EnumValues: []EnumValue{
							{
								Name: "DE",
								Description: func() String {
									desc := "Germany"
									return &desc
								}(),
							},
							{
								Name: "NP",
								Description: func() String {
									desc := "Nepal"
									return &desc
								}(),
							},
							{
								Name: "TH",
								Description: func() String {
									desc := "Thailand"
									return &desc
								}(),
							},
						},
					},
				},
			},
			expected: `type Query struct {
	User User ` + "`json:\"user\"`" + `
}

type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Address *Address ` + "`json:\"address,omitempty\"`" + `
}

type Address struct {
	Street string ` + "`json:\"street\"`" + `
	City string ` + "`json:\"city\"`" + `
	Country Country ` + "`json:\"country\"`" + `
}

type Country string

const (
	CountryDe Country = "DE"
	CountryNp Country = "NP"
	CountryTh Country = "TH"
)
`,
		},
		{
			name: "schema with cycle",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Product",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "variant", Type: TypeRef{Kind: OBJECT, Name: "ProductVariant", OfType: &TypeRef{Kind: OBJECT, Name: "ProductVariant"}}},
						},
					},
					{
						Name: "ProductVariant",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "barcode", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "inventoryItem", Type: TypeRef{Kind: NON_NULL, Name: "", OfType: &TypeRef{Kind: OBJECT, Name: "InventoryItem"}}},
						},
					},
					{
						Name: "InventoryItem",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "harmonizedSystemCode", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "variant", Type: TypeRef{Kind: NON_NULL, Name: "", OfType: &TypeRef{Kind: OBJECT, Name: "ProductVariant"}}},
						},
					},
				},
			},
			expected: `type Product struct {
	Name string ` + "`json:\"name\"`" + `
	Variant *ProductVariant ` + "`json:\"variant\"`" + `
}

type ProductVariant struct {
	ID string ` + "`json:\"id\"`" + `
	Barcode string ` + "`json:\"barcode\"`" + `
	InventoryItem *InventoryItem ` + "`json:\"inventoryItem\"`" + `
}

type InventoryItem struct {
	HarmonizedSystemCode string ` + "`json:\"harmonizedSystemCode\"`" + `
	Variant *ProductVariant ` + "`json:\"variant\"`" + `
}
`,
		},
		{
			name: "schema with non-null type",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "user", Type: TypeRef{Kind: NON_NULL, OfType: &TypeRef{Kind: OBJECT, Name: "User"}}},
						},
					},
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
						},
					},
				},
			},
			expected: `type Query struct {
	User User ` + "`json:\"user\"`" + `
}

type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`,
		},
		{
			name: "schema with enum type",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "status", Type: TypeRef{Kind: ENUM, Name: "Status", OfType: &TypeRef{Kind: ENUM, Name: "Status"}}},
						},
					},
					{
						Name: "Status",
						Kind: ENUM,
						EnumValues: []EnumValue{
							{Name: "ACTIVE"},
							{Name: "INACTIVE"},
						},
					},
				},
			},
			expected: `type Query struct {
	Status Status ` + "`json:\"status\"`" + `
}

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)
`,
		},
		{
			name: "schema with non-null list type",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "users", Type: TypeRef{Kind: NON_NULL, OfType: &TypeRef{Kind: LIST, OfType: &TypeRef{Kind: OBJECT, Name: "User"}}}},
						},
					},
					{
						Name: "User",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
						},
					},
				},
			},
			expected: `type Query struct {
	Users []User ` + "`json:\"users\"`" + `
}

type User struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`,
		},
		{
			name: "schema with input object type",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "user", Type: TypeRef{Kind: OBJECT, Name: "UserInput", OfType: &TypeRef{Kind: OBJECT, Name: "UserInput"}}},
						},
					},
					{
						Name: "UserInput",
						Kind: INPUT_OBJECT,
						InputFields: []InputField{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String", OfType: &TypeRef{Kind: SCALAR, Name: "String"}}},
						},
					},
				},
			},
			expected: `type Query struct {
	User UserInput ` + "`json:\"user\"`" + `
}

type UserInput struct {
	ID string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`,
		},
		{
			name: "schema with interface type",
			schema: IntrospectionSchema{
				Types: []Type{
					{
						Name: "Query",
						Kind: OBJECT,
						Fields: []Field{
							{Name: "user", Type: TypeRef{Kind: OBJECT, Name: "User", OfType: &TypeRef{Kind: OBJECT, Name: "User"}}},
						},
					},
					{
						Name: "User",
						Kind: INTERFACE,
						Fields: []Field{
							{Name: "id", Type: TypeRef{Kind: SCALAR, Name: "ID"}},
							{Name: "name", Type: TypeRef{Kind: SCALAR, Name: "String"}},
						},
					},
				},
			},
			expected: `type Query struct {
	User User ` + "`json:\"user\"`" + `
}

type User interface {
	ID() string
	Name() string
}
`,
		},
	}

	for _, tc := range cases {
		nodes := NewNodes()
		for _, t := range tc.schema.Types {
			nodes.Collect(NewNode(t))
		}
		nodes.Link()

		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, nodes.ToGoTypes())
		})
	}
}

func TestNodes_Link(t *testing.T) {
	// Helper function to create a node.
	createNode := func(name string, fields ...*FieldMeta) *Node {
		return &Node{
			Name:   name,
			Kind:   OBJECT,
			Fields: fields,
		}
	}

	// Initialize nodes with helper function.
	n1 := createNode("Node1",
		&FieldMeta{Name: "Node 2", Type: "Node2", JSONTag: "node_2"},
		&FieldMeta{Name: "Node 3", Type: "Node3", JSONTag: "node_3"},
		&FieldMeta{Name: "Node 5", Type: "Node5", JSONTag: "node_5"},
	)
	n2 := createNode("Node2",
		&FieldMeta{Name: "Node 4", Type: "Node4", JSONTag: "node_4"},
	)
	n3 := createNode("Node3",
		&FieldMeta{Name: "Node 5", Type: "Node5", JSONTag: "node_5"},
	)
	n4 := createNode("Node4")
	n5 := createNode("Node5",
		&FieldMeta{Name: "Node 1", Type: "Node1", JSONTag: "node_1"},
	)

	nodes := NewNodes()
	for _, nd := range []*Node{n1, n2, n3, n4, n5} {
		nodes.Collect(nd)
	}
	nodes.Link()

	cases := []struct {
		name     string
		start    *Node
		target   *Node
		expected bool
	}{
		{"No Cycle N1 -> N2", n1, n2, false},
		{"Cycle N1 -> N3 -> N5 -> N1", n1, n3, true},
		{"Cycle N1 -> N5 -> N1", n1, n5, true},
		{"No Cycle N2 -> N4", n2, n4, false},
		{"Cycle N3 -> N5 -> N1 -> N3", n3, n5, true},
		{"Cycle N5 -> N1 -> N5", n5, n1, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seen := make(map[string]struct{})
			assert.Equal(t, tc.expected, tc.start.hasCycleWith(tc.target, seen))
		})
	}

	// Check if the nodes fields property has isCycle set to true.
	assert.True(t, n1.Fields[1].hasCycle)
	assert.True(t, n1.Fields[2].hasCycle)
	assert.True(t, n3.Fields[0].hasCycle)
	assert.True(t, n5.Fields[0].hasCycle)

	// Check if the nodes fields property has isCycle set to false.
	assert.False(t, n1.Fields[0].hasCycle)
	assert.False(t, n2.Fields[0].hasCycle)
}
