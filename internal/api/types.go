package api

import (
	"github.com/ankitpokhrel/shopctl/schema"
)

type ProductResponse struct {
	Data struct {
		Product schema.Product `json:"product"`
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
