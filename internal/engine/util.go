package engine

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// GetHashDir returns a hash directory for a given string.
func GetHashDir(of string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(of)))
	return hash[:2]
}

// ExtractNumericID extracts numeric part of a Shopify ID.
// Ex: gid://shopify/Product/8737842954464 -> 8737842954464.
func ExtractNumericID(shopifyID string) string {
	parts := strings.Split(shopifyID, "/")
	return parts[len(parts)-1]
}
