// Package introspect provides utilities for converting GraphQL introspection schemas
// into Go type definitions. It includes functionality to handle various GraphQL types
// such as OBJECT, INPUT_OBJECT, INTERFACE and ENUM, and generate corresponding Go code.
// The support for other types will be added as needed.
//
// This intention of this package is to generate Go code from GraphQL introspection schemas
// in case the public GraphQL schema is not available for whatever reason.
//
// See https://github.com/facebook/graphql/blob/master/spec/Section%204%20--%20Introspection.md
package introspect
