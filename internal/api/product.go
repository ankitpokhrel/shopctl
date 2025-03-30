package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// CheckProductByHandle fetches a product by Handle without additional details.
func (c GQLClient) CheckProductByHandle(handle string) (*schema.Product, error) {
	var out struct {
		Data struct {
			Product schema.Product `json:"productByIdentifier"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := `query GetProductByHandle($identifier: ProductIdentifierInput!) {
  productByIdentifier(identifier: $identifier) {
    id
    handle
  }
}`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"identifier": map[string]string{"handle": handle},
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	return &out.Data.Product, nil
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
func (c GQLClient) GetProducts(limit int, after *string, query *string) (*ProductsResponse, error) {
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

// GetProductOptions fetches product options.
func (c GQLClient) GetProductOptions(productID string) (*ProductOptionsResponse, error) {
	var out *ProductOptionsResponse

	query := `query GetProductOptions($id: ID!) {
  product(id: $id) {
    id
    options {
      id
      name
      position
      linkedMetafield {
        key
        namespace
      }
      optionValues {
        id
        name
        hasVariants
        linkedMetafieldValue
      }
    }
  }
}`

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
      nodes {
        %s
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

// CheckProductVariantByID returns product variant without fetching all fields.
func (c GQLClient) CheckProductVariantByID(variantID string) (*schema.ProductVariant, error) {
	return c.getProductVariantByID(variantID, "id\ntitle\n")
}

// GetProductVariantByID returns product variant by its id.
func (c GQLClient) GetProductVariantByID(variantID string) (*schema.ProductVariant, error) {
	return c.getProductVariantByID(variantID, fieldsVariant)
}

func (c GQLClient) getProductVariantByID(variantID string, fields string) (*schema.ProductVariant, error) {
	var out struct {
		Data struct {
			Node *schema.ProductVariant `json:"node"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := fmt.Sprintf(`query GetProductVariantById($id: ID!) {
  node(id: $id) {
    ... on ProductVariant {
        %s
    }
  }
}`, fields)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": variantID},
	}
	if err := c.Execute(context.Background(), req, client.Header{"X-ShopCTL-Resource-ID": variantID}, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	if out.Data.Node == nil {
		return nil, fmt.Errorf("variant not found")
	}
	return out.Data.Node, nil
}

// GetProductVariantByTitle returns variant matching the given title.
//
// Shopify limits 100 variants per product so we should be good to fetch them all at once.
// We will revisit this if we run into any issues even with the limit.
func (c GQLClient) GetProductVariantByTitle(productID string, title string, fetchAll bool) (*schema.ProductVariant, error) {
	var out *ProductVariantsResponse

	query := `query GetProductVariants($id: ID!) {
  product(id: $id) {
    id
    variants(first: 100) {
      nodes {
        id
        title
      }
    }
  }
}`

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
	for _, v := range out.Data.Product.Variants.Nodes {
		if strings.EqualFold(v.Title, title) {
			if fetchAll {
				return c.GetProductVariantByID(v.ID)
			}
			return &v, nil
		}
	}
	return nil, fmt.Errorf("variant with the given title not found")
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
      nodes {
        %s
      }
    }
  }
}`, fieldsMetafields)

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
      nodes {
        %s
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
