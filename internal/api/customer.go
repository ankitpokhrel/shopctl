package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
)

// CheckCustomerByEmailOrPhone fetches a customer by email or phone without additional details.
func (c GQLClient) CheckCustomerByEmailOrPhone(email *string, phone *string) (*CustomersResponse, error) {
	var (
		query = `query CheckCustomerByEmailOrPhone($query: String!) {
          customers(first: 1, query: $query) {
            nodes {
              id
              email
              phone
            }
          }
        }`

		exp []string
		out *CustomersResponse
	)

	if email != nil {
		exp = append(exp, fmt.Sprintf("email:%s", *email))
	}
	if phone != nil {
		exp = append(exp, fmt.Sprintf("phone:%s", *phone))
	}

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"query": strings.Join(exp, " OR ")},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCustomerMetaFields fetches metafields of a customer by its ID.
//
// Shopify limits 200 metafields per customer and the response size seems ok.
// We'll fetch them all at once for now. We will revisit this if we run
// into any issue due to the response size.
func (c GQLClient) GetCustomerMetaFields(customerID string) (*CustomerMetaFieldsResponse, error) {
	var out *CustomerMetaFieldsResponse

	query := fmt.Sprintf(`query GetCustomerMetaFields($id: ID!) {
  customer(id: $id) {
    id
    email
    phone
    metafields(first: 200) {
      nodes {
        %s
      }
    }
  }
}`, fieldsMetafields)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": customerID},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCustomerMetaFieldsByEmailOrPhone fetches metafields of a customer by email or phone.
func (c GQLClient) GetCustomerMetaFieldsByEmailOrPhone(email *string, phone *string) (*CustomersMetaFieldsResponse, error) {
	var (
		query = fmt.Sprintf(`query GetCustomerMetaFieldsByEmailOrPhone($query: String!) {
      customers(first: 1, query: $query) {
        nodes {
          id
          email
          phone
          metafields(first: 200) {
            nodes {
              %s
            }
          }
        }
      }
}`, fieldsMetafields)

		exp []string
		out *CustomersMetaFieldsResponse
	)

	if email != nil {
		exp = append(exp, fmt.Sprintf("email:%s", *email))
	}
	if phone != nil {
		exp = append(exp, fmt.Sprintf("phone:%s", *phone))
	}

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"query": strings.Join(exp, " OR ")},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Data.Customers.Nodes) == 0 {
		return nil, fmt.Errorf("customer metafield not found")
	}
	return out, nil
}

// GetAllCustomers fetches customers in a batch and streams the response to a channel.
func (c GQLClient) GetAllCustomers(ch chan *CustomersResponse, limit int, after *string) error {
	var out *CustomersResponse

	query := fmt.Sprintf(`query GetCustomers($first: Int!, $after: String) {
  customers(first: $first, after: $after) {
    nodes {
      %s
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`, fieldsCustomer)

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"first": limit,
			"after": after,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return err
	}
	if len(out.Errors) > 0 {
		return fmt.Errorf("%s", out.Errors)
	}

	ch <- out

	if out.Data.Customers.PageInfo.HasNextPage {
		return c.GetAllCustomers(ch, limit, out.Data.Customers.PageInfo.EndCursor)
	}
	return nil
}
