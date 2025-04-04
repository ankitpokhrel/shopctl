package api

import (
	"context"
	"fmt"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// SetMetafields sets a metafield to the resource.
func (c GQLClient) SetMetafields(metafields []schema.MetafieldsSetInput) (*MetafieldSetResponse, error) {
	var out struct {
		Data struct {
			MetafieldsSet MetafieldSetResponse `json:"metafieldsSet"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation metafieldsSet($metafields: [MetafieldsSetInput!]!) {
        metafieldsSet(metafields: $metafields) {
            metafields {
              id
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
			"metafields": metafields,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productMetafieldsSet: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.MetafieldsSet.UserErrors) > 0 {
		return nil, fmt.Errorf("productMetafieldsSet: the operation failed with user error: %s", out.Data.MetafieldsSet.UserErrors.Error())
	}
	return &out.Data.MetafieldsSet, nil
}

// DeleteMetafields deletes metafield attached to a resource.
func (c GQLClient) DeleteMetafields(metafields []schema.MetafieldIdentifierInput) (*MetafieldDeleteResponse, error) {
	var out struct {
		Data struct {
			MetafieldsDelete MetafieldDeleteResponse `json:"metafieldsDelete"`
		} `json:"data"`
		Errors Errors `json:"errors,omitempty"`
	}

	query := `
    mutation metafieldsDelete($metafields: [MetafieldIdentifierInput!]!) {
        metafieldsDelete(metafields: $metafields) {
            deletedMetafields {
              key
              namespace
              ownerId
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
			"metafields": metafields,
		},
	}

	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("productMetafieldsDelete: the operation failed with error: %s", out.Errors.Error())
	}
	if len(out.Data.MetafieldsDelete.UserErrors) > 0 {
		return nil, fmt.Errorf("productMetafieldsDelete: the operation failed with user error: %s", out.Data.MetafieldsDelete.UserErrors.Error())
	}
	return &out.Data.MetafieldsDelete, nil
}
