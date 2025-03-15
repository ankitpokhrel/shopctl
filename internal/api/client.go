package api

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const (
	version = "2025-01"
)

// GQLClient is a GraphQL client.
type GQLClient struct {
	*client.Client
	logger *tlog.Logger
}

// GQLClientFunc is a functional opt for GQLClient.
type GQLClientFunc func(*GQLClient)

// NewGQLClient constructs new GraphQL client for a store.
func NewGQLClient(store string, opts ...GQLClientFunc) *GQLClient {
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

	c := GQLClient{}
	for _, opt := range opts {
		opt(&c)
	}
	if c.logger == nil {
		c.logger = tlog.New(tlog.VerboseLevel(tlog.VL1), false)
	}
	c.Client = client.NewClient(server, token, client.WithLogger(c.logger))

	return &c
}

// LogRequest sets custom logger for the client.
func LogRequest(lgr *tlog.Logger) GQLClientFunc {
	return func(c *GQLClient) {
		c.logger = lgr
	}
}
