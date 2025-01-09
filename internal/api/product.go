package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ankitpokhrel/shopctl/schema"
)

type Error struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Extensions struct {
		Value any `json:"value"`
	} `json:"extensions"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

type Errors []Error

// Error implements the error interface.
func (e Errors) Error() string {
	errs := make([]string, 0, len(e))
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, ", ")
}

type ProductResponse struct {
	Data struct {
		Product schema.Product `json:"product"`
	} `json:"data"`
	Errors Errors `json:"errors"`
}

type ProductIdentifierResponse struct {
	Data struct {
		Product schema.Product `json:"productByIdentifier"`
	} `json:"data"`
	Errors Errors `json:"errors"`
}

type ProductsResponse struct {
	Data struct {
		Products ProductData `json:"products"`
	} `json:"data"`
}

type ProductData struct {
	Edges []struct {
		Node schema.Product `json:"node"`
	} `json:"edges"`
	PageInfo schema.PageInfo `json:"pageInfo"`
}

type ProductVariantsResponse struct {
	Data struct {
		Product struct {
			ID       string             `json:"id"`
			Variants ProductVariantData `json:"variants"`
		} `json:"product"`
	} `json:"data"`
}

type ProductVariantData struct {
	Edges []struct {
		Node schema.ProductVariant `json:"node"`
	} `json:"edges"`
}

type ProductMediasResponse struct {
	Data struct {
		Product struct {
			ID    string           `json:"id"`
			Media ProductMediaData `json:"media"`
		} `json:"product"`
	} `json:"data"`
}

type ProductMediaData struct {
	Edges []struct {
		Node struct {
			ID               string                   `json:"id"`
			Status           schema.MediaStatus       `json:"status"`
			Preview          schema.MediaPreviewImage `json:"preview"`
			MediaContentType schema.MediaContentType  `json:"mediaContentType"`
			MediaErrors      []any                    `json:"mediaErrors"`
			MediaWarnings    []any                    `json:"mediaWarnings"`
		} `json:"node"`
	} `json:"edges"`
}

type ProductCreateResponse struct {
	Product    schema.Product     `json:"product"`
	UserErrors []schema.UserError `json:"userErrors"`
	Errors     Errors             `json:"errors"`
}

// CheckProductByID fetches a product by ID without additional details.
func (c GQLClient) CheckProductByID(id string) (*ProductResponse, error) {
	productsQuery := map[string]string{
		"query": fmt.Sprintf(`{
  product(id: "%s") {
  	id
  }
}`, id),
	}

	query, err := json.Marshal(productsQuery)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var out *ProductResponse

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}

// GetProductByID fetches a product by ID.
func (c GQLClient) GetProductByID(id string) (*ProductResponse, error) {
	productsQuery := map[string]string{
		"query": fmt.Sprintf(`{
  product(id: "%s") {
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
}`, id),
	}

	query, err := json.Marshal(productsQuery)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var out *ProductResponse

	err = json.NewDecoder(res.Body).Decode(&out)
	if out.Data.Product.ID == "" {
		return nil, fmt.Errorf("product not found")
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out, err
}

func (c GQLClient) GetProductByHandle(handle string) (*ProductIdentifierResponse, error) {
	productsQuery := `
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

	variables := map[string]any{
		"identifier": map[string]string{
			"handle": handle,
		},
	}

	// Create the request body
	reqBody := GraphQLRequest{
		Query:     productsQuery,
		Variables: variables,
	}

	query, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var out *ProductIdentifierResponse

	err = json.NewDecoder(res.Body).Decode(&out)
	if out.Data.Product.ID == "" {
		return nil, fmt.Errorf("product not found")
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out, err
}

// GetProducts fetches n number of products after a cursor.
func (c GQLClient) GetProducts(limit int, after *string) (*ProductsResponse, error) {
	var params string

	if after != nil {
		params = fmt.Sprintf(`first: %d, after: "%s"`, limit, *after)
	} else {
		params = fmt.Sprintf(`first: %d`, limit)
	}

	// Get all Products.
	productsQuery := map[string]string{
		"query": fmt.Sprintf(`{
  products(%s) {
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
}`, params),
	}

	query, err := json.Marshal(productsQuery)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var out *ProductsResponse

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}

// GetProductVariants fetches variants of a product.
//
// Shopify limits 100 variants per product so we should be good to fetch them all at once.
// We will revisit this if we run into any issues even with the limit.
func (c GQLClient) GetProductVariants(productID string) (*ProductVariantsResponse, error) {
	variantsQuery := map[string]string{
		"query": fmt.Sprintf(`{
  product(id: "%s") {
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
}`, productID),
	}

	query, err := json.Marshal(variantsQuery)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var out *ProductVariantsResponse

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}

// GetProductMedias fetches medias of a product.
//
// Shopify limits 250 medias per product and the response size seems ok.
// We'll fetch them all at once for now. We will revisit this if we run
// into any issue due to the response size.
func (c GQLClient) GetProductMedias(productID string) (*ProductMediasResponse, error) {
	mediaQuery := map[string]string{
		"query": fmt.Sprintf(`{
  product(id: "%s") {
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
          mediaErrors {
            details
          }
          mediaWarnings {
            message
          }
        }
      }
    }
  }
}`, productID),
	}

	query, err := json.Marshal(mediaQuery)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var out *ProductMediasResponse

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}

// CreateProduct creates a product.
func (c GQLClient) CreateProduct(input schema.ProductCreateInput) (*ProductCreateResponse, error) {
	createQuery := `
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

	var out struct {
		Data struct {
			ProductCreate ProductCreateResponse `json:"productCreate"`
		} `json:"data"`
	}

	err := c.executeGQLMutation(context.Background(), createQuery, map[string]any{"input": input}, &out)
	return &out.Data.ProductCreate, err
}

// UpdateProduct updates a product.
func (c GQLClient) UpdateProduct(input schema.ProductUpdateInput) (*ProductCreateResponse, error) {
	updateQuery := `
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

	var out struct {
		Data struct {
			ProductUpdate ProductCreateResponse `json:"productUpdate"`
		} `json:"data"`
	}

	err := c.executeGQLMutation(context.Background(), updateQuery, map[string]any{"input": input}, &out)
	return &out.Data.ProductUpdate, err
}
