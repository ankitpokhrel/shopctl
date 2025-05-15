package shopctl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	AppConfigVersion  = "v0"
	ShopifyApiVersion = "2025-04"
)

// ShopifyProductID formats Shopify product ID.
func ShopifyProductID(id string) string {
	prefix := "gid://shopify/Product"
	if strings.HasPrefix(id, prefix) {
		return id
	}
	if _, err := strconv.Atoi(id); err != nil {
		return "" // Not an integer id.
	}
	return fmt.Sprintf("%s/%s", prefix, id)
}

// ShopifyProductVariantID formats Shopify product variant ID.
func ShopifyProductVariantID(id string) string {
	prefix := "gid://shopify/ProductVariant"
	if strings.HasPrefix(id, prefix) {
		return id
	}
	if _, err := strconv.Atoi(id); err != nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", prefix, id)
}

// ShopifyMediaID formats Shopify product media ID.
func ShopifyMediaID(id string, typ schema.MediaContentType) string {
	validPrefixes := map[schema.MediaContentType]string{
		schema.MediaContentTypeImage:         "gid://shopify/MediaImage",
		schema.MediaContentTypeVideo:         "gid://shopify/Video",
		schema.MediaContentTypeModel3d:       "gid://shopify/Model3d",
		schema.MediaContentTypeExternalVideo: "gid://shopify/ExternalVideo",
	}
	for _, p := range validPrefixes {
		if strings.HasPrefix(id, p) {
			return id
		}
	}

	prefix, ok := validPrefixes[typ]
	if !ok {
		return ""
	}
	if _, err := strconv.Atoi(id); err != nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", prefix, id)
}

// ShopifyCustomerID formats Shopify customer ID.
func ShopifyCustomerID(id string) string {
	prefix := "gid://shopify/Customer"
	if strings.HasPrefix(id, prefix) {
		return id
	}
	if _, err := strconv.Atoi(id); err != nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", prefix, id)
}

// ShopifyWebhookSubscriptionID formats Shopify product variant ID.
func ShopifyWebhookSubscriptionID(id string) string {
	prefix := "gid://shopify/WebhookSubscription"
	if strings.HasPrefix(id, prefix) {
		return id
	}
	if _, err := strconv.Atoi(id); err != nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", prefix, id)
}

// ExtractNumericID extracts numeric part of a Shopify ID.
// Ex: gid://shopify/Product/8737842954464 -> 8737842954464.
func ExtractNumericID(shopifyID string) string {
	parts := strings.Split(shopifyID, "/")
	return parts[len(parts)-1]
}
