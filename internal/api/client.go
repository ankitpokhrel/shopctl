package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/zalando/go-keyring"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
)

const (
	version = "2025-01"
)

// GraphQLRequest is a GraphQL request.
type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// GQLClient is a GraphQL client.
type GQLClient struct {
	*client.Client
}

// NewGQLClient constructs new GraphQL client for a store.
func NewGQLClient(store string) *GQLClient {
	var (
		token   string
		err     error
		defined bool

		server  = fmt.Sprintf("https://%s/admin/api/%s/graphql.json", store, version)
		service = fmt.Sprintf("shopctl:%s", cmdutil.GetStoreSlug(store))
	)

	// The `SHOPIFY_ACCESS_TOKEN` env has the highest priority when looking for a token.
	// We will then look into other secure storages like system's keyring/keychain.
	// Finally, we'll fallback to read from insecure storage like config files.
	if token, defined = os.LookupEnv("SHOPIFY_ACCESS_TOKEN"); !defined {
		token, err = keyring.Get(service, store)
		if err != nil || token == "" {
			token = config.GetToken(store)
		}
	}

	return &GQLClient{
		Client: client.NewClient(server, token),
	}
}

func (c GQLClient) executeGQLMutation(ctx context.Context, mutation string, variables map[string]any, output any) error {
	payload := GraphQLRequest{
		Query:     mutation,
		Variables: variables,
	}

	query, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := c.Request(ctx, query, nil)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(output)
}
