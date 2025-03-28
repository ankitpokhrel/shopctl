// Usage:
//
//	$ export SHOPIFY_ACCESS_TOKEN=<token>
//	$ introspect -s mystore.myshopify.com -t Product -p schema
//	$ go run ./cmd/introspect -s mystore.myshopify.com -t Product -p schema
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/pkg/gql/introspect"
)

// Flag wraps available command flags.
type Flag struct {
	store string
	pkg   string
	typ   string
}

func (f *Flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	exitOnErr(err)

	typ, err := cmd.Flags().GetString("type")
	exitOnErr(err)

	pkg, err := cmd.Flags().GetString("package")
	exitOnErr(err)

	if store == "" {
		store = os.Getenv("SHOPIFY_STORE")
	}
	if store == "" {
		_, _ = fmt.Fprint(
			os.Stderr,
			"Error: store url is missing: you can either use \"-s\" flag or set a \"SHOPIFY_STORE\" env variable\n\n",
		)
		_ = cmd.Usage()
		os.Exit(1)
	}

	f.store = store
	f.pkg = pkg
	f.typ = typ
}

// SchemaParser parses GQL schema.
type SchemaParser struct {
	client    *client.Client
	processed map[string]struct{}
}

// NewSchemaParser constructs a schema parser.
func NewSchemaParser(server string) *SchemaParser {
	gqlClient := client.NewClient(
		server,
		os.Getenv("SHOPIFY_ACCESS_TOKEN"),
	)

	return &SchemaParser{
		client:    gqlClient,
		processed: make(map[string]struct{}),
	}
}

func main() {
	cmd := &cobra.Command{
		Use:   "introspect [flags]",
		Short: "Fetch and generate Go types from Shopify GQL introspection schema",
		Long:  "introspect fetches introspection schema from Shopify GQL API and generates corresponding Go types",
		RunE: func(cmd *cobra.Command, args []string) error {
			return introspectSchemaToGoType(cmd)
		},
	}

	cmd.Flags().StringP(
		"store", "s", "",
		"Shopify store to connect to",
	)
	cmd.Flags().StringP(
		"type", "t", "",
		"Top-level resource type to introspect",
	)
	cmd.Flags().StringP(
		"package", "p", "",
		"Name of the package",
	)
	cmd.Flags().SortFlags = false

	exitOnErr(cmd.MarkFlagRequired("type"))
	exitOnErr(cmd.MarkFlagRequired("package"))

	exitOnErr(cmd.Execute())
}

func introspectSchemaToGoType(cmd *cobra.Command) error {
	flag := &Flag{}
	flag.parse(cmd)

	parser := NewSchemaParser(gqlServerURL(flag.store))
	nodes := introspect.NewNodes()

	if err := resolveType(flag.typ, parser, nodes); err != nil {
		return err
	}
	nodes.Link()

	fmt.Printf("// Code generated by introspect; DO NOT EDIT.\n\n")
	fmt.Printf("package %s\n\n", flag.pkg)
	fmt.Printf("%s", nodes.ToGoTypes())

	return nil
}

func resolveType(gqlTyp string, parser *SchemaParser, nodes *introspect.Nodes) error {
	if _, ok := parser.processed[gqlTyp]; ok {
		return nil
	}

	var (
		query  []byte
		res    *http.Response
		result introspect.Query
		err    error
	)

	// Send a request and process response.
	{
		if query, err = json.Marshal(getQuery(gqlTyp)); err != nil {
			return err
		}

		if res, err = parser.client.Request(context.Background(), query, nil); err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()

		if res.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("unauthorized: either the token is expired or invalid")
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("received unexpected response with code: %d", res.StatusCode)
		}

		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return err
		}
	}

	// Build and collect node.
	node := introspect.NewNode(result.Data.Type)
	nodes.Collect(node)

	// Mark node as visited.
	parser.processed[gqlTyp] = struct{}{}

	// Process children.
	fields := introspect.GetIntrospectionTypes(
		introspect.IntrospectionSchema{
			Types: []introspect.Type{result.Data.Type},
		},
	)
	for gqlType, f := range fields {
		if f.Kind == introspect.SCALAR {
			continue
		}
		if err := resolveType(gqlType, parser, nodes); err != nil {
			return err
		}
	}
	return nil
}

func gqlServerURL(store string) string {
	protocol := "https://"
	apiPath := "admin/api/2024-10/graphql.json"

	// Strip http:// or https:// if present.
	if store[:7] == "http://" {
		store = store[7:]
	} else if store[:8] == "https://" {
		store = store[8:]
	}

	return fmt.Sprintf("%s%s/%s", protocol, store, apiPath)
}

func exitOnErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
