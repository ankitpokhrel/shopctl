package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// CheckProductByID fetches a product by ID without additional details.
func (c GQLClient) CheckProductByID(id string) (*schema.Product, error) {
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
	return &out.Data.Product, err
}

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

	productQuery := fmt.Sprintf(`query GetProducts($first: Int!, $after: String, $query: String, $sortKey: ProductSortKeys!, $reverse: Boolean!) {
products(first: $first, after: $after, query: $query, sortKey: $sortKey, reverse: $reverse) {
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
			"first":   limit,
			"after":   after,
			"query":   query,
			"sortKey": "CREATED_AT",
			"reverse": true,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
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
func (c GQLClient) GetProductVariants(productID string) (*ProductVariantData, error) {
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
	return &out.Data.Product, nil
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
                handle
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
		return nil, fmt.Errorf("productCreate: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("productCreate: the operation failed with user error: %s", out.Data.ProductCreate.UserErrors.Error())
	}
	return &out.Data.ProductCreate, nil
}

// UpdateProduct updates a product.
func (c GQLClient) UpdateProduct(input schema.ProductInput, media []schema.CreateMediaInput) (*ProductCreateResponse, error) {
	var out struct {
		Data struct {
			ProductUpdate ProductCreateResponse `json:"productUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
	mutation productUpdate($input: ProductInput!, $media: [CreateMediaInput!]) {
        productUpdate(input: $input, media: $media) {
			product {
				id
                handle
				title
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
			"input": input,
			"media": media,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productUpdate: Product %s: The operation failed with error: %s", *input.ID, out.Errors.Error())
	}
	if len(out.Data.ProductUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("productUpdate: Product %s: The operation failed with user error: %s", *input.ID, out.Data.ProductUpdate.UserErrors.Error())
	}
	return &out.Data.ProductUpdate, nil
}

// DeleteProduct deletes a product.
func (c GQLClient) DeleteProduct(productID string) (*ProductDeleteResponse, error) {
	var out struct {
		Data struct {
			ProductDelete ProductDeleteResponse `json:"productDelete"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation deleteProduct($productId: ID!) {
      productDelete(input: { id: $productId }) {
        deletedProductId
        userErrors {
          field
          message
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId": productID,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productDelete: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductDelete.UserErrors) > 0 {
		return nil, fmt.Errorf("productDelete: the operation failed with user error: %s", out.Data.ProductDelete.UserErrors.Error())
	}
	return &out.Data.ProductDelete, nil
}

// CreateProductOptions creates one or more product options.
//
//nolint:dupl
func (c GQLClient) CreateProductOptions(
	productID string,
	options []schema.OptionCreateInput,
	strategy schema.ProductOptionCreateVariantStrategy,
) (*ProductOptionSyncResponse, error) {
	var out struct {
		Data struct {
			ProductOptionCreate ProductOptionSyncResponse `json:"productOptionsCreate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation createOptions($productId: ID!, $options: [OptionCreateInput!]!, $strategy: ProductOptionCreateVariantStrategy) {
      productOptionsCreate(productId: $productId, options: $options, variantStrategy: $strategy) {
        product {
          id
          options {
            id
            name
            optionValues { name }
          }
        }
        userErrors {
          field
          message
          code
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId": productID,
			"options":   options,
			"strategy":  strategy,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productOptionsCreate: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductOptionCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("productOptionsCreate: the operation failed with user error: %s", out.Data.ProductOptionCreate.UserErrors.Error())
	}
	return &out.Data.ProductOptionCreate, nil
}

// UpdateProductOptions updates product options.
func (c GQLClient) UpdateProductOptions(
	productID string,
	option *schema.OptionUpdateInput,
	optionsToAdd []schema.OptionValueCreateInput,
	optionsToUpdate []schema.OptionValueUpdateInput,
	optionsToDelete []string,
	strategy schema.ProductOptionUpdateVariantStrategy,
) (*ProductOptionSyncResponse, error) {
	var out struct {
		Data struct {
			ProductOptionUpdate ProductOptionSyncResponse `json:"productOptionUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation updateOption($productId: ID!, $option: OptionUpdateInput!, $optionValuesToAdd: [OptionValueCreateInput!], $optionValuesToUpdate: [OptionValueUpdateInput!], $optionValuesToDelete: [ID!], $variantStrategy: ProductOptionUpdateVariantStrategy) {
      productOptionUpdate(productId: $productId, option: $option, optionValuesToAdd: $optionValuesToAdd, optionValuesToUpdate: $optionValuesToUpdate, optionValuesToDelete: $optionValuesToDelete, variantStrategy: $variantStrategy) {
        product {
          id
          options {
            id
            name
          }
        }
        userErrors {
          field
          message
          code
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId":            productID,
			"option":               option,
			"optionValuesToAdd":    optionsToAdd,
			"optionValuesToUpdate": optionsToUpdate,
			"optionValuesToDelete": optionsToDelete,
			"variantStrategy":      strategy,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productOptionUpdate: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductOptionUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("productOptionUpdate: the operation failed with user error: %s", out.Data.ProductOptionUpdate.UserErrors.Error())
	}
	return &out.Data.ProductOptionUpdate, nil
}

// DeleteProductOptions removes one or more product options.
func (c GQLClient) DeleteProductOptions(productID string, options []string) (*ProductOptionSyncResponse, error) {
	var out struct {
		Data struct {
			ProductOptionDelete ProductOptionSyncResponse `json:"productOptionsDelete"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation deleteOptions($productId: ID!, $options: [ID!]!, $strategy: ProductOptionDeleteStrategy) {
      productOptionsDelete(productId: $productId, options: $options, strategy: $strategy) {
        deletedOptionsIds
        product {
          id
          options {
            id
            name
          }
        }
        userErrors {
          field
          message
          code
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId": productID,
			"options":   options,
			"strategy":  "POSITION",
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productOptionsDelete: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductOptionDelete.UserErrors) > 0 {
		return nil, fmt.Errorf("productOptionsDelete: the operation failed with user error: %s", out.Data.ProductOptionDelete.UserErrors.Error())
	}
	return &out.Data.ProductOptionDelete, nil
}

// CreateProductVariants creates one or more product variants.
//
//nolint:dupl
func (c GQLClient) CreateProductVariants(
	productID string,
	variants []schema.ProductVariantsBulkInput,
	strategy schema.ProductVariantsBulkCreateStrategy,
) (*ProductVariantsSyncResponse, error) {
	var out struct {
		Data struct {
			ProductVariantsBulkCreate ProductVariantsSyncResponse `json:"productVariantsBulkCreate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation ProductVariantsCreate($productId: ID!, $variants: [ProductVariantsBulkInput!]!, $strategy: ProductVariantsBulkCreateStrategy) {
      productVariantsBulkCreate(productId: $productId, variants: $variants, strategy: $strategy) {
        product {
          id
        }
        userErrors {
          field
          message
          code
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId": productID,
			"variants":  variants,
			"strategy":  strategy,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productVariantsCreate: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductVariantsBulkCreate.UserErrors) > 0 {
		return nil, fmt.Errorf("productVariantsCreate: the operation failed with user error: %s", out.Data.ProductVariantsBulkCreate.UserErrors.Error())
	}
	return &out.Data.ProductVariantsBulkCreate, nil
}

// UpdateProductVariants creates one or more product variants.
func (c GQLClient) UpdateProductVariants(productID string, variants []schema.ProductVariantsBulkInput, partialUpdate bool) (*ProductVariantsSyncResponse, error) {
	var out struct {
		Data struct {
			ProductVariantsBulkUpdate ProductVariantsSyncResponse `json:"productVariantsBulkUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation ProductVariantsUpdate($productId: ID!, $variants: [ProductVariantsBulkInput!]!, $allowPartialUpdates: Boolean!) {
      productVariantsBulkUpdate(productId: $productId, variants: $variants, allowPartialUpdates: $allowPartialUpdates) {
        product {
          id
        }
        userErrors {
          field
          message
          code
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId":           productID,
			"variants":            variants,
			"allowPartialUpdates": partialUpdate,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productVariantsUpdate: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductVariantsBulkUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("productVariantsUpdate: the operation failed with user error: %s", out.Data.ProductVariantsBulkUpdate.UserErrors.Error())
	}
	return &out.Data.ProductVariantsBulkUpdate, nil
}

// DeleteProductVariants deletes one or more product variants.
func (c GQLClient) DeleteProductVariants(productID string, variants []string) (*ProductVariantsSyncResponse, error) {
	var out struct {
		Data struct {
			ProductVariantBulkDelete ProductVariantsSyncResponse `json:"productVariantsBulkDelete"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation productVariantsDelete($productId: ID!, $variantsIds: [ID!]!) {
      productVariantsBulkDelete(productId: $productId, variantsIds: $variantsIds) {
        product {
          id
          options {
            id
            name
          }
        }
        userErrors {
          field
          message
          code
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"productId":   productID,
			"variantsIds": variants,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productVariantsDelete: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.ProductVariantBulkDelete.UserErrors) > 0 {
		return nil, fmt.Errorf("productVariantsDelete: the operation failed with user error: %s", out.Data.ProductVariantBulkDelete.UserErrors.Error())
	}
	return &out.Data.ProductVariantBulkDelete, nil
}

// DetachProductMedia detaches one or more product media.
func (c GQLClient) DetachProductMedia(input []schema.FileUpdateInput) (*FileUpdateResponse, error) {
	var out struct {
		Data struct {
			FileUpdate FileUpdateResponse `json:"fileUpdate"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation productMediaDetach($input: [FileUpdateInput!]!) {
      fileUpdate(files: $input) {
        files {
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
		return nil, fmt.Errorf("productMediaDetach: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.FileUpdate.UserErrors) > 0 {
		return nil, fmt.Errorf("productMediaDelete: the operation failed with user error: %s", out.Data.FileUpdate.UserErrors.Error())
	}
	return &out.Data.FileUpdate, nil
}
