package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// GetCustomerByID fetches customer data by ID.
func (c GQLClient) GetCustomerByID(id string) (*schema.Customer, error) {
	var (
		query = fmt.Sprintf(`query GetCustomerByID($id: ID!) {
          customer(id: $id) {
            %s
          }
        }`, fieldsCustomer)

		out *CustomerResponse
	)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": id},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return &out.Data.Customer, nil
}

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

// CreateCustomer creates a customer.
func (c GQLClient) CreateCustomer(input schema.CustomerInput) (*CustomerSyncResponse, error) {
	var out struct {
		Data struct {
			CustomerCreate CustomerSyncResponse `json:"customerCreate"`
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
		return nil, fmt.Errorf("customerCreate: The operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.CustomerCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("customerCreate: The operation failed with user error: %s", out.Data.CustomerCreate.UserErrors.Error())
	}
	return &out.Data.CustomerCreate, nil
}

// UpdateCustomer updates a customer.
func (c GQLClient) UpdateCustomer(input schema.CustomerInput) (*CustomerSyncResponse, error) {
	var out struct {
		Data struct {
			CustomerUpdate CustomerSyncResponse `json:"customerUpdate"`
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
		return nil, fmt.Errorf("customerUpdate: Customer %s: The operation failed with error: %s", *input.ID, out.Errors.Error())
	}
	if len(out.Data.CustomerUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("customerUpdate: Customer %s: The operation failed with user error: %s", *input.ID, out.Data.CustomerUpdate.UserErrors.Error())
	}
	return &out.Data.CustomerUpdate, nil
}

// UpdateCustomerAddress updates a customer address.
func (c GQLClient) UpdateCustomerAddress(customerID string, addressID string, input schema.MailingAddressInput, isDefault bool) (*CustomerAddressUpdateResponse, error) {
	var out struct {
		Data struct {
			CustomerAddressUpdate CustomerAddressUpdateResponse `json:"customerAddressUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := `
	mutation UpdateCustomerAddress($address: MailingAddressInput!, $addressId: ID!, $customerId: ID!, $setAsDefault: Boolean) {
	    customerAddressUpdate(address: $address, addressId: $addressId, customerId: $customerId, setAsDefault: $setAsDefault) {
			address {
			    address1
			}
			userErrors {
				field
				message
			}
		}
	}`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"address":      input,
			"addressId":    addressID,
			"customerId":   customerID,
			"setAsDefault": isDefault,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("customerAddressUpdate: Customer %s: The operation failed with error: %s", customerID, out.Errors.Error())
	}
	if len(out.Data.CustomerAddressUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("customerAddressUpdate: Customer %s: The operation failed with user error: %s", customerID, out.Data.CustomerAddressUpdate.UserErrors.Error())
	}
	return &out.Data.CustomerAddressUpdate, nil
}
