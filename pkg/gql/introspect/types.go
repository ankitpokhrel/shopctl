// See https://github.com/facebook/graphql/blob/master/spec/Section%204%20--%20Introspection.md#schema-introspection

package introspect

import (
	"fmt"
	"strconv"
)

// GraphQL type kinds.
const (
	NULL TypeKind = iota
	NON_NULL
	SCALAR
	OBJECT
	INTERFACE
	UNION
	ENUM
	INPUT_OBJECT
	LIST
)

// Directive locations.
const (
	DirectiveLocationQuery                DirectiveLocation = "QUERY"
	DirectiveLocationMutation             DirectiveLocation = "MUTATION"
	DirectiveLocationSubscription         DirectiveLocation = "SUBSCRIPTION"
	DirectiveLocationField                DirectiveLocation = "FIELD"
	DirectiveLocationFragmentDefinition   DirectiveLocation = "FRAGMENT_DEFINITION"
	DirectiveLocationFragmentSpread       DirectiveLocation = "FRAGMENT_SPREAD"
	DirectiveLocationInlineFragment       DirectiveLocation = "INLINE_FRAGMENT"
	DirectiveLocationVariableDefinition   DirectiveLocation = "VARIABLE_DEFINITION"
	DirectiveLocationSchema               DirectiveLocation = "SCHEMA"
	DirectiveLocationScalar               DirectiveLocation = "SCALAR"
	DirectiveLocationObject               DirectiveLocation = "OBJECT"
	DirectiveLocationFieldDefinition      DirectiveLocation = "FIELD_DEFINITION"
	DirectiveLocationArgumentDefinition   DirectiveLocation = "ARGUMENT_DEFINITION"
	DirectiveLocationInterface            DirectiveLocation = "INTERFACE"
	DirectiveLocationUnion                DirectiveLocation = "UNION"
	DirectiveLocationEnum                 DirectiveLocation = "ENUM"
	DirectiveLocationEnumValue            DirectiveLocation = "ENUM_VALUE"
	DirectiveLocationInputObject          DirectiveLocation = "INPUT_OBJECT"
	DirectiveLocationInputFieldDefinition DirectiveLocation = "INPUT_FIELD_DEFINITION"
)

// String is a custom type for string.
type String *string

// TypeKind represents the kind of GraphQL type.
type TypeKind int

// UnmarshalJSON unmarshal the JSON data into a TypeKind.
func (k *TypeKind) UnmarshalJSON(kind []byte) error {
	kindStr, _ := strconv.Unquote(string(kind))
	switch kindStr {
	case "NULL":
		*k = NULL
	case "NON_NULL":
		*k = NON_NULL
	case "SCALAR":
		*k = SCALAR
	case "OBJECT":
		*k = OBJECT
	case "INTERFACE":
		*k = INTERFACE
	case "UNION":
		*k = UNION
	case "ENUM":
		*k = ENUM
	case "INPUT_OBJECT":
		*k = INPUT_OBJECT
	case "LIST":
		*k = LIST
	default:
		return fmt.Errorf("unknown type kind: %s", kind)
	}
	return nil
}

// DirectiveLocation represents the location of a directive.
type DirectiveLocation string

// Schema represents the root introspection query result.
type Schema struct {
	Data struct {
		Schema IntrospectionSchema `json:"__schema"`
	} `json:"data"`
}

type Query struct {
	Data struct {
		Type Type `json:"__type"`
	} `json:"data"`
}

// IntrospectionSchema represents the GraphQL schema structure.
type IntrospectionSchema struct {
	Types            []Type      `json:"types"`
	QueryType        TypeRef     `json:"queryType"`
	MutationType     *TypeRef    `json:"mutationType,omitempty"`
	SubscriptionType *TypeRef    `json:"subscriptionType,omitempty"`
	Directives       []Directive `json:"directives"`
}

// Type represents a type in the GraphQL schema.
type Type struct {
	Kind          TypeKind     `json:"kind"`
	Name          string       `json:"name"`
	Description   String       `json:"description,omitempty"`
	Fields        []Field      `json:"fields,omitempty"`
	InputFields   []InputField `json:"inputFields,omitempty"`
	Interfaces    []TypeRef    `json:"interfaces,omitempty"`
	EnumValues    []EnumValue  `json:"enumValues,omitempty"`
	PossibleTypes []TypeRef    `json:"possibleTypes,omitempty"`
}

// Field represents a field in a type.
type Field struct {
	Name              string       `json:"name"`
	Description       String       `json:"description,omitempty"`
	Args              []InputField `json:"args"`
	Type              TypeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason String       `json:"deprecationReason,omitempty"`
}

// InputField represents an input field (used in arguments or input types).
type InputField struct {
	Name         string  `json:"name"`
	Description  String  `json:"description,omitempty"`
	Type         TypeRef `json:"type"`
	DefaultValue String  `json:"defaultValue,omitempty"`
}

// EnumValue represents an enum value in a type.
type EnumValue struct {
	Name              string `json:"name"`
	Description       String `json:"description,omitempty"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason String `json:"deprecationReason,omitempty"`
}

// TypeRef represents a reference to another type.
type TypeRef struct {
	Kind   TypeKind `json:"kind"`
	Name   string   `json:"name,omitempty"`
	OfType *TypeRef `json:"ofType,omitempty"`
}

// Directive represents a directive in the schema.
type Directive struct {
	Name        string       `json:"name"`
	Description String       `json:"description,omitempty"`
	Locations   []string     `json:"locations"`
	Args        []InputField `json:"args"`
}
