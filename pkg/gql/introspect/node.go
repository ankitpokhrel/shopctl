package introspect

import (
	"fmt"
	"strings"
)

// Node represents a node in a GraphQL schema.
type Node struct {
	Name     string
	Kind     TypeKind
	Fields   []*FieldMeta
	Children []*Node
}

// FieldMeta holds field meta definition within a Go type.
type FieldMeta struct {
	Name    string
	Type    string
	JSONTag string

	hasCycle   bool
	isNullable bool
}

// NewNode creates a new Node from a GraphQL Type.
func NewNode(schema Type) *Node {
	node := Node{
		Name:     schema.Name,
		Kind:     schema.Kind,
		Children: make([]*Node, 0),
	}

	switch schema.Kind {
	case OBJECT, INTERFACE, UNION:
		for _, f := range schema.Fields {
			node.Fields = append(node.Fields, &FieldMeta{
				Name:       f.Name,
				Type:       gqlTypeToGoType(f.Type),
				JSONTag:    f.Name,
				isNullable: f.Type.Kind == NULL || f.Type.OfType == nil,
			})
		}
	case INPUT_OBJECT:
		for _, f := range schema.InputFields {
			node.Fields = append(node.Fields, &FieldMeta{
				Name:       f.Name,
				Type:       gqlTypeToGoType(f.Type),
				JSONTag:    f.Name,
				isNullable: f.Type.Kind == NULL || f.Type.OfType == nil,
			})
		}
	case ENUM:
		for _, v := range schema.EnumValues {
			node.Fields = append(node.Fields, &FieldMeta{
				Name:       v.Name,
				Type:       schema.Name,
				JSONTag:    v.Name,
				isNullable: false,
			})
		}
	}

	return &node
}

// String implements the fmt.Stringer interface.
func (n *Node) String() string {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("\n%s (%d)\n", n.Name, n.Kind))
	for _, c := range n.Children {
		out.WriteString(fmt.Sprintf("  -> %s", c.Name))
	}

	return out.String()
}

// ToGoTypes converts a collection of Node into Go types.
func (n *Nodes) ToGoTypes() string {
	out := make([]string, 0, len(n.hashMap))

	for _, v := range n.sortOrder {
		nd := n.hashMap[v]

		switch nd.Kind {
		case OBJECT, INPUT_OBJECT, UNION:
			out = append(out, nd.toGoStruct())
		case INTERFACE:
			out = append(out, nd.toGoInterface())
		case ENUM:
			out = append(out, nd.toGoEnum())
		default:
			panic(fmt.Sprintf("unsupported type: %d", nd.Kind))
		}
	}

	return strings.Join(out, "\n")
}

// toGoStruct converts a Node into a Go struct definition.
func (n *Node) toGoStruct() string {
	var (
		out       strings.Builder
		fieldTmpl string
	)

	out.WriteString(fmt.Sprintf("type %s struct {\n", n.Name))
	for _, f := range n.Fields {
		fieldTmpl = "\t%s %s `json:\"%s\"`\n"
		if f.isNullable {
			fieldTmpl = "\t%s *%s `json:\"%s,omitempty\"`\n"
		}
		if f.hasCycle {
			fieldTmpl = "\t%s *%s `json:\"%s\"`\n"
		}
		out.WriteString(fmt.Sprintf(fieldTmpl, capitalize(f.Name), f.Type, f.JSONTag))
	}
	out.WriteString("}\n")

	return out.String()
}

// toGoInterface converts a Node into a Go interface definition.
func (n *Node) toGoInterface() string {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("type %s interface {\n", n.Name))
	for _, f := range n.Fields {
		out.WriteString(fmt.Sprintf("\t%s() %s\n", capitalize(f.Name), f.Type))
	}
	out.WriteString("}\n")

	return out.String()
}

// toGoEnum converts a Node into a Go enum definition.
func (n *Node) toGoEnum() string {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("type %s string\n\n", n.Name))
	out.WriteString("const (\n")
	for _, f := range n.Fields {
		out.WriteString(fmt.Sprintf("\t%s%s %s = \"%s\"\n", n.Name, capitalize(strings.ToLower(f.Name)), n.Name, f.Name))
	}
	out.WriteString(")\n")

	return out.String()
}

// markCycles detects and marks fields that have cycles with other Nodes.
func (n *Node) markCycles() {
	for _, c := range n.Children {
		seen := make(map[string]struct{})
		if !n.hasCycleWith(c, seen) {
			continue
		}
		for _, f := range n.Fields {
			if f.Type == c.Name {
				f.hasCycle = true
			}
		}
	}
}

// hasCycle checks if a Node has a cycle with or via another Node.
func (n *Node) hasCycleWith(n1 *Node, seen map[string]struct{}) bool {
	if _, ok := seen[n.Name]; ok {
		return true
	}
	if n1.Kind != OBJECT && n1.Kind != INPUT_OBJECT {
		return false
	}
	seen[n.Name] = struct{}{}
	for _, c := range n1.Children {
		if n1.hasCycleWith(c, seen) {
			return true
		}
	}
	return false
}

// Nodes is a collection of Node.
type Nodes struct {
	hashMap   map[string]*Node
	sortOrder []string
}

// NewNodes instantiate Nodes type.
func NewNodes() *Nodes {
	return &Nodes{
		hashMap:   make(map[string]*Node),
		sortOrder: make([]string, 0),
	}
}

// Collect appends a Node to the collection.
func (n *Nodes) Collect(nd *Node) {
	if _, ok := n.hashMap[nd.Name]; !ok {
		n.hashMap[nd.Name] = nd
		n.sortOrder = append(n.sortOrder, nd.Name)
	}
}

// Link connects the children of each Node and mark any cyclic nodes.
func (n *Nodes) Link() {
	for _, nd := range n.hashMap {
		seen := make(map[string]struct{})
		for _, f := range nd.Fields {
			if _, ok := n.hashMap[f.Type]; !ok {
				continue
			}
			if _, ok := seen[f.Type]; !ok {
				nd.Children = append(nd.Children, n.hashMap[f.Type])
			}
			seen[f.Type] = struct{}{}
		}
	}

	// Mark cyclic nodes.
	for _, o := range n.sortOrder {
		nd := n.hashMap[o]
		nd.markCycles()
	}
}
