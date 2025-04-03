package engine

import (
	"strings"
)

// ExtractNumericID extracts numeric part of a Shopify ID.
// Ex: gid://shopify/Product/8737842954464 -> 8737842954464.
func ExtractNumericID(shopifyID string) string {
	parts := strings.Split(shopifyID, "/")
	return parts[len(parts)-1]
}
