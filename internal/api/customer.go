package api

import (
	"context"
	"fmt"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// CheckCustomerByID fetches a customer by ID without additional details.
func (c GQLClient) CheckCustomerByID(id string) (*CustomerResponse, error) {
	var (
		query = `query CheckCustomerByID($id: ID!) { customer(id: $id) { id } }`

		out *CustomerResponse
		err error
	)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": id},
	}
	if err = c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	return out, err
}

// GetAllCustomers fetches customers in a batch and streams the response to a channel.
func (c GQLClient) GetAllCustomers(ch chan *CustomersResponse, limit int, after *string) error {
	var out *CustomersResponse

	query := fmt.Sprintf(`query GetCustomers($first: Int!, $after: String) {
  customers(first: $first, after: $after) {
    edges {
      node {
        %s
      }
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

// GetCustomerMetaFields fetches medias of a product.
//
// Shopify limits 200 metafields per customer and the response size seems ok.
// We'll fetch them all at once for now. We will revisit this if we run
// into any issue due to the response size.
func (c GQLClient) GetCustomerMetaFields(customerID string) (*CustomerMetaFieldsResponse, error) {
	var out *CustomerMetaFieldsResponse

	query := fmt.Sprintf(`query GetCustomerMetaFields($id: ID!) {
  customer(id: $id) {
id
metafields(first: 200) {
      edges {
        node {
          %s
        }
      }
    }
  }
}`, fieldsMetaFields)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": customerID},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateCustomer creates a customer.
func (c GQLClient) CreateCustomer(input schema.CustomerInput) (*CustomerCreateResponse, error) {
	var out struct {
		Data struct {
			CustomerCreate CustomerCreateResponse `json:"customerCreate"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := `
	mutation customerCreate($input: CustomerInput!) {
		customerCreate(input: $input) {
			customer {
				id
			}
			userErrors {
				field
				message
			}
		}
	}`

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"input": input},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("Customer: The operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.CustomerCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("Customer: The operation failed with user error: %s", out.Data.CustomerCreate.UserErrors.Error())
	}
	return &out.Data.CustomerCreate, nil
}

// UpdateCustomer updates a customer.
func (c GQLClient) UpdateCustomer(input schema.CustomerInput) (*CustomerCreateResponse, error) {
	var out struct {
		Data struct {
			CustomerUpdate CustomerCreateResponse `json:"customerUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := `
	mutation customerUpdate($input: CustomerInput!) {
		customerUpdate(input: $input) {
			customer {
				id
			}
			userErrors {
				field
				message
			}
		}
	}`

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"input": input},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("Customer %s: The operation failed with error: %s", *input.ID, out.Errors.Error())
	}
	if len(out.Data.CustomerUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("Customer %s: The operation failed with user error: %s", *input.ID, out.Data.CustomerUpdate.UserErrors.Error())
	}
	return &out.Data.CustomerUpdate, nil
}
