package introspect

import (
	"strings"
)

// GetIntrospectionTypes extracts GQL type definition from OBJECT and INPUT_FIELDS.
func GetIntrospectionTypes(schema IntrospectionSchema) map[string]TypeRef {
	ofTypes := make(map[string]TypeRef)

	parseField := func(fieldType TypeRef) {
		if fieldType.OfType != nil {
			if fieldType.OfType.Name != "" {
				ofTypes[fieldType.OfType.Name] = *fieldType.OfType
			} else if fieldType.Name != "" {
				ofTypes[fieldType.Name] = *fieldType.OfType
			}
		} else if fieldType.Name != "" {
			ofTypes[fieldType.Name] = fieldType
		}
	}

	for _, gqlType := range schema.Types {
		for _, field := range gqlType.Fields {
			parseField(field.Type)
		}
		for _, field := range gqlType.InputFields {
			parseField(field.Type)
		}
	}

	return ofTypes
}

// gqlTypeToGoType converts a GraphQL type to a Go type.
func gqlTypeToGoType(ref TypeRef) string {
	switch ref.Kind {
	case SCALAR:
		switch ref.Name {
		case "Int":
			return "int"
		case "Float":
			fallthrough
		case "Decimal":
			return "float64"
		case "String":
			return "string"
		case "Boolean":
			return "bool"
		case "ID":
			return "string"
		default:
			return "string"
		}
	case OBJECT:
		return ref.Name
	case NON_NULL:
		if ref.OfType != nil {
			return gqlTypeToGoType(*ref.OfType)
		}
		return ref.Name
	case NULL:
		return gqlTypeToGoType(*ref.OfType)
	case LIST:
		if ref.OfType != nil && ref.OfType.Name != "" {
			return "[]" + gqlTypeToGoType(*ref.OfType)
		}
		return "[]any"
	case ENUM:
		return ref.Name
	case INPUT_OBJECT:
		return ref.Name
	case INTERFACE:
		return ref.Name
	case UNION:
		return "any"
	}
	return "any" // Default fallback type.
}

// Convert a string to camel case, with the first letter capitalized and special cases handled.
func capitalize(input string) string {
	if input == "" {
		return input
	}

	// Special cases.
	special := map[string]string{
		"id":  "ID",
		"uri": "URI",
		"url": "URL",
		"api": "API",
	}

	// Exceptions.
	exceptions := []string{
		"valid",
		"invalid",
		"liquid",
	}

	hasSuffix := func(s, suffix string) bool { return strings.HasSuffix(strings.ToLower(s), suffix) }
	toUpper := func(s string) string { return strings.ToUpper(s[:1]) + s[1:] }
	capitalizeWord := func(word string) string {
		// Bail early if we find an exact special word.
		if val, ok := special[word]; ok {
			return val
		}

		// Ignore replacement if suffix is in the exception list.
		for _, v := range exceptions {
			if len(word) >= len(v) && hasSuffix(word, v) {
				return word
			}
		}

		// If the input ends with special suffix, replace it with the defined value.
		for k, v := range special {
			if len(word) > len(k) && hasSuffix(word, k) {
				return word[:len(word)-len(k)] + v
			}
		}

		return word
	}

	words := strings.Split(input, "_")
	for i, word := range words {
		words[i] = toUpper(capitalizeWord(word))
	}
	return strings.Join(words, "")
}
