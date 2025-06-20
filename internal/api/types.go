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

type ProductOptionsResponse struct {
	Data struct {
		Product ProductOptionsData `json:"product"`
	} `json:"data"`
	Errors Errors `json:"errors"`
}

type ProductOptionsData struct {
	ProductID string                 `json:"id"`
	Options   []schema.ProductOption `json:"options"`
}

type ProductVariantsResponse struct {
	Data struct {
		Product ProductVariantData `json:"product"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type ProductVariantData struct {
	ProductID string `json:"id"`
	Variants  struct {
		Nodes []schema.ProductVariant `json:"nodes"`
	} `json:"variants"`
}

type ProductMetaFieldsResponse struct {
	Data struct {
		Product ProductMetafieldsData `json:"product"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type ProductMetafieldsData struct {
	ProductID  string `json:"id"`
	Metafields struct {
		Nodes []schema.Metafield `json:"nodes"`
	} `json:"metafields"`
}

type ProductMediasResponse struct {
	Data struct {
		Product ProductMediaData `json:"product"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type ProductMediaNode struct {
	ID               string                   `json:"id"`
	Status           schema.MediaStatus       `json:"status"`
	Preview          schema.MediaPreviewImage `json:"preview"`
	MediaContentType schema.MediaContentType  `json:"mediaContentType"`
	MediaErrors      []any                    `json:"mediaErrors,omitempty"`
	MediaWarnings    []any                    `json:"mediaWarnings,omitempty"`
}

type ProductMediaData struct {
	ProductID string `json:"id"`
	Media     struct {
		Nodes []ProductMediaNode `json:"nodes"`
	} `json:"media"`
}

type FileUpdateResponse struct {
	Files []struct {
		ID string `json:"id"`
	} `json:"files"`
	UserErrors UserErrors `json:"userErrors"`
}

type ProductCreateResponse struct {
	Product    schema.Product `json:"product"`
	UserErrors UserErrors     `json:"userErrors"`
}

type ProductDeleteResponse struct {
	DeletedProductID string     `json:"deletedProductId"`
	UserErrors       UserErrors `json:"userErrors"`
}

type ProductOptionSyncResponse struct {
	Product struct {
		ID      string                 `json:"id"`
		Options []schema.ProductOption `json:"options"`
	} `json:"product"`
	UserErrors UserErrors `json:"userErrors"`
}

type ProductVariantsSyncResponse struct {
	Product struct {
		ID string `json:"id"`
	} `json:"product"`
	Variants   []schema.ProductVariant `json:"productVariants"`
	UserErrors UserErrors              `json:"userErrors"`
}

type MetafieldSetResponse struct {
	Metafields []schema.Metafield `json:"metafields"`
	UserErrors UserErrors         `json:"userErrors"`
}

type MetafieldDeleteResponse struct {
	Metafields []schema.MetafieldIdentifier `json:"deletedMetafields"`
	UserErrors UserErrors                   `json:"userErrors"`
}

type CustomerCreateResponse struct {
	Customer   schema.Customer `json:"customer"`
	UserErrors UserErrors      `json:"userErrors"`
}

type CustomerSyncResponse struct {
	Customer struct {
		ID string `json:"id"`
	} `json:"customer"`
	UserErrors UserErrors `json:"userErrors"`
}

type CustomerAddressUpdateResponse struct {
	Address struct {
		Address1 string `json:"address1"`
	} `json:"address"`
	UserErrors UserErrors `json:"userErrors"`
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

type CustomerDeleteResponse struct {
	Shop struct {
		ID string `json:"id"`
	} `json:"shop"`
	UserErrors UserErrors `json:"userErrors"`
}

type CustomerData struct {
	Nodes    []schema.Customer `json:"nodes"`
	PageInfo schema.PageInfo   `json:"pageInfo"`
}

type CustomerMetafieldsData struct {
	CustomerID string `json:"id"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Metafields struct {
		Nodes []schema.Metafield `json:"nodes"`
	} `json:"metafields"`
}

type CustomerMetaFieldsResponse struct {
	Data struct {
		Customer CustomerMetafieldsData `json:"customer"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type CustomersMetaFieldsResponse struct {
	Data struct {
		Customers struct {
			Nodes []CustomerMetafieldsData `json:"nodes"`
		} `json:"customers"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type OrdersResponse struct {
	Data struct {
		Orders OrderData `json:"orders"`
	} `json:"data"`
	Errors     Errors     `json:"errors"`
	Extensions Extensions `json:"extensions"`
}

type OrderData struct {
	Nodes    []schema.Order  `json:"nodes"`
	PageInfo schema.PageInfo `json:"pageInfo"`
}

type WebhookData struct {
	Nodes    []schema.WebhookSubscription `json:"nodes"`
	PageInfo schema.PageInfo              `json:"pageInfo"`
}

type WebhookSyncResponse struct {
	WebhookSubscription schema.WebhookSubscription `json:"webhookSubscription"`
	UserErrors          UserErrors                 `json:"userErrors"`
}

type WebhookDeleteResponse struct {
	DeletedWebhookSubscriptionID string     `json:"deletedWebhookSubscriptionId"`
	UserErrors                   UserErrors `json:"userErrors"`
}
