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
	if err = c.Execute(context.Background(), req, &out); err != nil {
		return nil, err
	}
	return out, err
}

// GetProductByID fetches a product by ID.
func (c GQLClient) GetProductByID(id string) (*schema.Product, error) {
	var out *ProductResponse

	query := `query GetProductByID($id: ID!) {
  product(id: $id) {
	  id
	  title
	  handle
	  description
	  descriptionHtml
	  productType
	  isGiftCard
	  status
	  category {
		  id
          name
          fullName
	  }
	  tags
	  totalInventory
	  tracksInventory
	  createdAt
	  updatedAt
	  publishedAt
	  combinedListingRole
	  defaultCursor
	  giftCardTemplateSuffix
	  hasOnlyDefaultVariant
	  hasOutOfStockVariants
	  hasVariantsThatRequiresComponents
	  legacyResourceId
	  onlineStorePreviewUrl
	  onlineStoreUrl
	  requiresSellingPlan
	  templateSuffix
	  vendor
	  options {
	    name
	    values
	    position
	    optionValues {
	      id
	      name
	      hasVariants
	    }
	  }
	  variants(first: 100) {
	    nodes {
	      id
	      title
	      displayName
	      price
	      sku
	      position
	      availableForSale
	      barcode
	      compareAtPrice
	      inventoryQuantity
	      sellableOnlineQuantity
	      requiresComponents
	      taxable
	      taxCode
	      createdAt
	      updatedAt
	    }
	  }
	  media(first: 250) {
		nodes {
		  id
		  alt
		  status
		  mediaContentType
		  preview {
		    image {
		      url
		    }
		  }
		}
	  }
	}
}`

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": id},
	}
	if err := c.Execute(context.Background(), req, &out); err != nil {
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

	query := `
	query GetProductByHandle($identifier: ProductIdentifierInput!) {
  productByIdentifier(identifier: $identifier) {
    id
    title
    handle
    description
    descriptionHtml
    productType
    isGiftCard
    status
    category {
      id
      name
      fullName
    }
    tags
    totalInventory
    tracksInventory
    createdAt
    updatedAt
    publishedAt
    combinedListingRole
    defaultCursor
    giftCardTemplateSuffix
    hasOnlyDefaultVariant
    hasOutOfStockVariants
    hasVariantsThatRequiresComponents
    legacyResourceId
    onlineStorePreviewUrl
    onlineStoreUrl
    requiresSellingPlan
    templateSuffix
    vendor
    options {
      name
      values
      position
      optionValues {
        id
        name
        hasVariants
      }
    }
    variants(first: 100) {
      nodes {
        id
        title
        displayName
        price
        sku
        position
        availableForSale
        barcode
        compareAtPrice
        inventoryQuantity
        sellableOnlineQuantity
        requiresComponents
        taxable
        taxCode
        createdAt
        updatedAt
      }
    }
    media(first: 250) {
      nodes {
        id
        alt
        status
        mediaContentType
        preview {
          image {
            url
          }
        }
      }
    }
  }
}`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"identifier": map[string]string{"handle": handle},
		},
	}

	if err := c.Execute(context.Background(), req, &out); err != nil {
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

	query := `query GetProducts($first: Int!, $after: String) {
  products(first: $first, after: $after) {
    edges {
      node {
      	id
       	title
       	handle
        description
        descriptionHtml
        productType
        isGiftCard
        status
        tags
        totalInventory
        tracksInventory
        createdAt
        updatedAt
        publishedAt
        combinedListingRole
        defaultCursor
        giftCardTemplateSuffix
        hasOnlyDefaultVariant
        hasOutOfStockVariants
        hasVariantsThatRequiresComponents
        legacyResourceId
        onlineStorePreviewUrl
        onlineStoreUrl
        requiresSellingPlan
        templateSuffix
        vendor
        options {
          name
          values
        }
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
}`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"first": limit,
			"after": after,
		},
	}

	if err := c.Execute(context.Background(), req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProductVariants fetches variants of a product.
//
// Shopify limits 100 variants per product so we should be good to fetch them all at once.
// We will revisit this if we run into any issues even with the limit.
func (c GQLClient) GetProductVariants(productID string) (*ProductVariantsResponse, error) {
	var out *ProductVariantsResponse

	query := `query GetProductVariants($id: ID!) {
  product(id: $id) {
    id
    variants(first: 100) {
      edges {
        node {
          id
          displayName
          availableForSale
          barcode
          compareAtPrice
          createdAt
          image {
            id
            altText
            url
            height
            width
            metafields(first: 5) {
              edges {
                node {
                  id
                  description
                }
              }
              nodes {
                id
              }
              pageInfo {
                hasNextPage
              }
            }
          }
        }
      }
    }
  }
}`

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": productID},
	}
	if err := c.Execute(context.Background(), req, &out); err != nil {
		return nil, err
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

	query := `query GetProductMedias($id: ID!) {
  product(id: $id) {
id
media(first: 250) {
      edges {
        node {
          id
          status
          preview {
            image {
              altText
              url
              height
              width
            }
            status
          }
          mediaContentType
          mediaErrors { details }
          mediaWarnings { message }
        }
      }
    }
  }
}`

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": productID},
	}
	if err := c.Execute(context.Background(), req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateProduct creates a product.
func (c GQLClient) CreateProduct(input schema.ProductCreateInput) (*ProductCreateResponse, error) {
	var out struct {
		Data struct {
			ProductCreate ProductCreateResponse `json:"productCreate"`
		} `json:"data"`
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
	if err := c.Execute(context.Background(), req, &out); err != nil {
		return nil, err
	}
	return &out.Data.ProductCreate, nil
}

// UpdateProduct updates a product.
func (c GQLClient) UpdateProduct(input schema.ProductUpdateInput) (*ProductCreateResponse, error) {
	var out struct {
		Data struct {
			ProductUpdate ProductCreateResponse `json:"productUpdate"`
		} `json:"data"`
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

	if err := c.Execute(context.Background(), req, &out); err != nil {
		return nil, err
	}
	return &out.Data.ProductUpdate, nil
}
