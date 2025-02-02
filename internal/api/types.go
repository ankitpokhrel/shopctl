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
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
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
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type ProductVariantData struct {
	Edges []struct {
		Node schema.ProductVariant `json:"node"`
	} `json:"edges"`
	PageInfo schema.PageInfo `json:"pageInfo"`
}

type ProductMetaFieldsResponse struct {
	Data struct {
		Product struct {
			ID         string         `json:"id"`
			MetaFields MetaFieldsData `json:"metafields"`
		} `json:"product"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type MetaFieldsData struct {
	Edges []struct {
		Node schema.Metafield `json:"node"`
	} `json:"edges"`
}

type ProductMediasResponse struct {
	Data struct {
		Product struct {
			ID    string           `json:"id"`
			Media ProductMediaData `json:"media"`
		} `json:"product"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
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
	Product    schema.Product `json:"product"`
	UserErrors UserErrors     `json:"userErrors"`
}

type CustomerCreateResponse struct {
	Customer   schema.Customer `json:"customer"`
	UserErrors UserErrors      `json:"userErrors"`
}

type CustomerResponse struct {
	Data struct {
		Customer schema.Customer `json:"customer"`
	} `json:"data"`
	Errors Errors `json:"errors"`
}

type CustomersResponse struct {
	Data struct {
		Customers CustomerData `json:"customers"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type CustomerData struct {
	Edges []struct {
		Node schema.Customer `json:"node"`
	} `json:"edges"`
	PageInfo schema.PageInfo `json:"pageInfo"`
}

type CustomerMetaFieldsResponse struct {
	Data struct {
		Customer struct {
			ID         string         `json:"id"`
			MetaFields MetaFieldsData `json:"metafields"`
		} `json:"customer"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}
