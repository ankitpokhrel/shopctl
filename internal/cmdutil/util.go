package cmdutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/spf13/cobra"
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

// GetContext gets current context details from the config.
func GetContext(cmd *cobra.Command, cfg *config.ShopConfig) (*config.StoreContext, error) {
	usrCtx, err := cmd.Flags().GetString("context")
	if err != nil {
		return nil, err
	}

	if usrCtx == "" {
		currCtx := cfg.CurrentContext()
		if currCtx == "" {
			return nil, fmt.Errorf("current-context is not set; either set a context with %q or use %q flag", "shopctl use-context context-name", "-c")
		}
		usrCtx = currCtx
	}

	ctx := cfg.GetContext(usrCtx)
	if ctx == nil {
		return nil, fmt.Errorf("no context exists with the name: %q", usrCtx)
	}

	return ctx, nil
}

// GetStrategy gets current backup strategy details from the config.
func GetStrategy(cmd *cobra.Command, ctx *config.StoreContext, cfg *config.ShopConfig) (*config.BackupStrategy, error) {
	usrStrategy, err := cmd.Flags().GetString("strategy")
	if err != nil {
		return nil, err
	}

	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return nil, err
	}

	if usrStrategy == "" {
		currStrategy := cfg.CurrentStrategy()
		if currStrategy == "" {
			return nil, fmt.Errorf("current-strategy is not set; either set a strategy with %q or use %q flag", "shopctl use-strategy strategy-name", "-s")
		}
		usrStrategy = currStrategy
	}

	return storeCfg.GetBackupStrategy(usrStrategy), nil
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
