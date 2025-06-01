package api

import (
	"context"
	"fmt"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// GetOrders fetches n number of orders after a cursor.
func (c GQLClient) GetOrders(limit int, after *string, query *string) ([]schema.Order, error) {
	var out *OrdersResponse

	ordersQuery := fmt.Sprintf(`query GetOrders($first: Int!, $after: String, $query: String, $sortKey: OrderSortKeys!, $reverse: Boolean!) {
  orders(first: $first, after: $after, query: $query, sortKey: $sortKey, reverse: $reverse) {
    nodes {
      %s
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`, fieldsOrder)

	req := client.GQLRequest{
		Query: ordersQuery,
		Variables: client.QueryVars{
			"first":   limit,
			"after":   after,
			"query":   query,
			"sortKey": "PROCESSED_AT",
			"reverse": true,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out.Data.Orders.Nodes, nil
}
