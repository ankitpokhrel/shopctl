package cmdutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ExitOnErr exits the program if an error is not nil.
func ExitOnErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

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

// ExtractNumericID extracts numeric part of a Shopify ID.
// Ex: gid://shopify/Product/8737842954464 -> 8737842954464.
func ExtractNumericID(shopifyID string) string {
	parts := strings.Split(shopifyID, "/")
	return parts[len(parts)-1]
}

// FormatDateTimeHuman formats date time in human readable format.
func FormatDateTimeHuman(dt, format string) string {
	t, err := time.Parse(format, dt)
	if err != nil {
		return dt
	}
	return t.Format("Mon, 02 Jan 06")
}

// GetStoreSlug gets the ID of a Shopify store given a store URL.
func GetStoreSlug(store string) string {
	store = stripProtocol(store)
	slug := store

	pieces := strings.SplitN(store, ".", 2)
	if len(pieces) > 0 {
		slug = pieces[0]
	}
	return slug
}

// stripProtocol strips the http protocol from a URL.
func stripProtocol(url string) string {
	if len(url) < 8 /* Max protocol length */ { //nolint:mnd
		return url
	}

	if url[:7] == "http://" {
		return url[7:]
	}
	if url[:8] == "https://" {
		return url[8:]
	}

	return url
}
