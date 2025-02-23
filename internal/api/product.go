package api

import (
	"context"
	"fmt"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// CheckProductByID fetches a product by ID without additional details.
func (c GQLClient) CheckProductByID(id string) (*ProductResponse, error) {
	var (
		query = `query CheckProductByID($id: ID!) { product(id: $id) { id } }`

		out *ProductResponse
		err error
	)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": id},
	}
	if err = c.Execute(context.Background(), req, client.Header{"X-ShopCTL-Resource-ID": id}, &out); err != nil {
		return nil, err
	}
	return out, err
}

// GetProductByID fetches a product by ID.
func (c GQLClient) GetProductByID(id string) (*schema.Product, error) {
	var out *ProductResponse

	query := fmt.Sprintf(`query GetProductByID($id: ID!) {
  product(id: $id) {
    %s
    variants(first: 100) {
      nodes {
        %s
      }
    }
    media(first: 250) {
	  nodes {
	    %s
	  }
    }
  }
}`, fieldsProduct, fieldsVariant, fieldsMedia)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": id},
	}
	if err := c.Execute(context.Background(), req, client.Header{"X-ShopCTL-Resource-ID": id}, &out); err != nil {
		return nil, err
	}
	if out.Data.Product.ID == "" {
		return nil, fmt.Errorf("product not found")
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return &out.Data.Product, nil
}

func (c GQLClient) GetProductByHandle(handle string) (*schema.Product, error) {
	var out struct {
		Data struct {
			Product schema.Product `json:"productByIdentifier"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := fmt.Sprintf(`
	query GetProductByHandle($identifier: ProductIdentifierInput!) {
  productByIdentifier(identifier: $identifier) {
    %s
    variants(first: 100) {
      nodes {
        %s
      }
    }
    media(first: 250) {
      nodes {
      	%s
      }
    }
  }
}`, fieldsProduct, fieldsVariant, fieldsMedia)

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"identifier": map[string]string{"handle": handle},
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if out.Data.Product.ID == "" {
		return nil, fmt.Errorf("product not found")
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return &out.Data.Product, nil
}

// GetProducts fetches n number of products after a cursor.
func (c GQLClient) GetProducts(limit int, after *string) (*ProductsResponse, error) {
	var out *ProductsResponse

	query := fmt.Sprintf(`query GetProducts($first: Int!, $after: String) {
  products(first: $first, after: $after) {
    edges {
      node {
      	%s
        variantsCount {
          count
        }
        mediaCount {
          count
        }
      }
    }
    pageInfo {
    	hasNextPage
     	endCursor
    }
  }
}`, fieldsProduct)

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"first": limit,
			"after": after,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAllProducts fetches products in a batch and streams the response to a channel.
func (c GQLClient) GetAllProducts(ch chan *ProductsResponse, limit int, after *string, query *string) error {
	var out *ProductsResponse

	productQuery := fmt.Sprintf(`query GetProducts($first: Int!, $after: String, $query: String) {
        products(first: $first, after: $after, query: $query) {
    edges {
      node {
      	%s
        variantsCount {
          count
        }
        mediaCount {
          count
        }
      }
    }
    pageInfo {
    	hasNextPage
     	endCursor
    }
  }
}`, fieldsProduct)

	req := client.GQLRequest{
		Query: productQuery,
		Variables: client.QueryVars{
			"first": limit,
			"after": after,
			"query": query,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return err
	}
	if len(out.Errors) > 0 {
		return fmt.Errorf("%s", out.Errors)
	}

	ch <- out

	if out.Data.Products.PageInfo.HasNextPage {
		return c.GetAllProducts(ch, limit, out.Data.Products.PageInfo.EndCursor, query)
	}
	return nil
}

// GetProductVariants fetches variants of a product.
//
// Shopify limits 100 variants per product so we should be good to fetch them all at once.
// We will revisit this if we run into any issues even with the limit.
func (c GQLClient) GetProductVariants(productID string) (*ProductVariantsResponse, error) {
	var out *ProductVariantsResponse

	query := fmt.Sprintf(`query GetProductVariants($id: ID!) {
  product(id: $id) {
    id
    variants(first: 100) {
      edges {
        node {
          %s
        }
      }
    }
  }
}`, fieldsVariant)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": productID},
	}
	if err := c.Execute(context.Background(), req, client.Header{"X-ShopCTL-Resource-ID": productID}, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out, nil
}

// GetProductMetaFields fetches medias of a product.
//
// Shopify limits 200 metafields per product and the response size seems ok.
// We'll fetch them all at once for now. We will revisit this if we run
// into any issue due to the response size.
func (c GQLClient) GetProductMetaFields(productID string) (*ProductMetaFieldsResponse, error) {
	var out *ProductMetaFieldsResponse

	query := fmt.Sprintf(`query GetProductMetaFields($id: ID!) {
  product(id: $id) {
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
		Variables: client.QueryVars{"id": productID},
	}
	if err := c.Execute(context.Background(), req, client.Header{"X-ShopCTL-Resource-ID": productID}, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out, nil
}

// GetProductMedias fetches medias of a product.
//
// Shopify limits 250 medias per product and the response size seems ok.
// We'll fetch them all at once for now. We will revisit this if we run
// into any issue due to the response size.
func (c GQLClient) GetProductMedias(productID string) (*ProductMediasResponse, error) {
	var out *ProductMediasResponse

	query := fmt.Sprintf(`query GetProductMedias($id: ID!) {
  product(id: $id) {
id
media(first: 250) {
      edges {
        node {
          %s
        }
      }
    }
  }
}`, fieldsMedia)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": productID},
	}
	if err := c.Execute(context.Background(), req, client.Header{"X-ShopCTL-Resource-ID": productID}, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out, nil
}

// CreateProduct creates a product.
func (c GQLClient) CreateProduct(input schema.ProductInput) (*ProductCreateResponse, error) {
	var out struct {
		Data struct {
			ProductCreate ProductCreateResponse `json:"productCreate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
	mutation productCreate($input: ProductInput!) {
		productCreate(input: $input) {
			product {
				id
				title
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
		return nil, fmt.Errorf("Product: The operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("Product: The operation failed with user error: %s", out.Data.ProductCreate.UserErrors.Error())
	}
	return &out.Data.ProductCreate, nil
}

// UpdateProduct updates a product.
func (c GQLClient) UpdateProduct(input schema.ProductInput) (*ProductCreateResponse, error) {
	var out struct {
		Data struct {
			ProductUpdate ProductCreateResponse `json:"productUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
	mutation productUpdate($input: ProductInput!) {
		productUpdate(input: $input) {
			product {
				id
				title
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
		return nil, fmt.Errorf("Product %s: The operation failed with error: %s", *input.ID, out.Errors.Error())
	}
	if len(out.Data.ProductUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("Product %s: The operation failed with user error: %s", *input.ID, out.Data.ProductUpdate.UserErrors.Error())
	}
	return &out.Data.ProductUpdate, nil
}
