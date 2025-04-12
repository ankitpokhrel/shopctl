package introspect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIntrospectionTypes(t *testing.T) {
	schema := IntrospectionSchema{
		Types: []Type{
			{
				Name: "User",
				Fields: []Field{
					{Name: "field1", Type: TypeRef{Name: "ID", Kind: SCALAR}},
					{Name: "field2", Type: TypeRef{OfType: &TypeRef{Name: "Int", Kind: SCALAR}}},
					{Name: "field3", Type: TypeRef{Kind: LIST, OfType: &TypeRef{Name: "String", Kind: SCALAR}}},
					{Name: "field4", Type: TypeRef{Kind: NON_NULL, OfType: &TypeRef{Name: "Boolean", Kind: SCALAR}}},
				},
				InputFields: []InputField{
					{Name: "input1", Type: TypeRef{Name: "Boolean", Kind: SCALAR, OfType: &TypeRef{Name: "", Kind: SCALAR}}},
					{Name: "input2", Type: TypeRef{Name: "Float", Kind: SCALAR}},
				},
			},
			{
				Name: "CustomType",
				Fields: []Field{
					{Name: "field5", Type: TypeRef{Name: "Custom", Kind: OBJECT}},
				},
			},
			{
				Name: "InterfaceType",
				Kind: INTERFACE, // Only OBJECT and INPUT_OBJECT are processed.
				Fields: []Field{
					{Name: "field6", Type: TypeRef{Name: "String", Kind: SCALAR}},
				},
			},
		},
	}

	expected := map[string]TypeRef{
		"ID":      {Name: "ID", Kind: SCALAR},
		"Int":     {Name: "Int", Kind: SCALAR},
		"String":  {Name: "String", Kind: SCALAR},
		"Boolean": {Name: "", Kind: SCALAR},
		"Float":   {Name: "Float", Kind: SCALAR},
		"Custom":  {Name: "Custom", Kind: OBJECT},
	}

	result := GetIntrospectionTypes(schema)
	assert.Equal(t, expected, result)
}

func TestGQLTypeToGoType(t *testing.T) {
	cases := []struct {
		name     string
		input    TypeRef
		expected string
	}{
		{
			name:     "scalar int",
			input:    TypeRef{Kind: SCALAR, Name: "Int"},
			expected: "int",
		},
		{
			name:     "scalar float",
			input:    TypeRef{Kind: SCALAR, Name: "Float"},
			expected: "float64",
		},
		{
			name:     "scalar string",
			input:    TypeRef{Kind: SCALAR, Name: "String"},
			expected: "string",
		},
		{
			name:     "scalar boolean",
			input:    TypeRef{Kind: SCALAR, Name: "Boolean"},
			expected: "bool",
		},
		{
			name:     "scalar id",
			input:    TypeRef{Kind: SCALAR, Name: "ID"},
			expected: "string",
		},
		{
			name:     "object type",
			input:    TypeRef{Kind: OBJECT, Name: "CustomType"},
			expected: "CustomType",
		},
		{
			name:     "input object type",
			input:    TypeRef{Kind: INPUT_OBJECT, Name: "InputType"},
			expected: "InputType",
		},
		{
			name:     "null type",
			input:    TypeRef{Kind: NULL, OfType: &TypeRef{Kind: SCALAR, Name: "String"}},
			expected: "string", // Nullable types will be prepended with '*' later in the process.
		},
		{
			name:     "non-null type",
			input:    TypeRef{Kind: NON_NULL, OfType: &TypeRef{Kind: SCALAR, Name: "String"}},
			expected: "string",
		},
		{
			name:     "list type",
			input:    TypeRef{Kind: LIST, OfType: &TypeRef{Kind: SCALAR, Name: "Int"}},
			expected: "[]int",
		},
		{
			name:     "enum type",
			input:    TypeRef{Kind: ENUM, Name: "EnumType"},
			expected: "EnumType",
		},
		{
			name:     "interface type",
			input:    TypeRef{Kind: INTERFACE, Name: "InterfaceType"},
			expected: "InterfaceType",
		},
		{
			name:     "union type",
			input:    TypeRef{Kind: UNION},
			expected: "any",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := gqlTypeToGoType(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCapitalize(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "lowercase",
			input:    "name",
			expected: "Name",
		},
		{
			name:     "uppercase",
			input:    "NAME",
			expected: "NAME",
		},
		{
			name:     "mixed case",
			input:    "nAmE",
			expected: "NAmE",
		},
		{
			name:     "multiple words",
			input:    "firstName",
			expected: "FirstName",
		},
		{
			name:     "special words",
			input:    "id",
			expected: "ID",
		},
		{
			name:     "special words as suffix",
			input:    "structId",
			expected: "StructID",
		},
		{
			name:     "special words as suffix",
			input:    "onlineStoreUrl",
			expected: "OnlineStoreURL",
		},
		{
			name:     "valid words with special char at the end",
			input:    "statusValid",
			expected: "StatusValid",
		},
		{
			name:     "valid words with underscore",
			input:    "all_valid",
			expected: "AllValid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := capitalize(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
