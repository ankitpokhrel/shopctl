package cmdutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/archives"
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/config"
)

// ContextValue is a string type to use as a key for `context.SetValue`.
// This helps avoid conflicts with the default `string` type.
type ContextValue string

const (
	KeyContext    ContextValue = "context"
	KeyStrategy   ContextValue = "strategy"
	KeyGQLClient  ContextValue = "gqlClient"
	KeyShopConfig ContextValue = "shopCfg"
	KeyLogger     ContextValue = "logger"
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

	strategy := storeCfg.GetBackupStrategy(usrStrategy)
	if strategy == nil {
		return nil, fmt.Errorf("strategy not found; please select a valid strategy with %q or use %q flag", "shopctl use-strategy strategy-name", "-s")
	}
	return strategy, nil
}

// Archive archives the source and saves it to the destination.
func Archive(src string, dest string, dir string) error {
	ctx := context.Background()

	files, err := archives.FilesFromDisk(ctx, nil, map[string]string{
		src: ".",
	})
	if err != nil {
		return err
	}

	zipFile := filepath.Join(dest, fmt.Sprintf("%s.tar.gz", dir))
	out, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}
	return format.Archive(ctx, out, files)
}

// ParseBackupResource parses raw resource string.
func ParseBackupResource(resources []string) []config.BackupResource {
	bkpResources := make([]config.BackupResource, 0, len(resources))
	for _, resource := range resources {
		piece := strings.SplitN(resource, "=", 2)
		res := config.BackupResource{
			Resource: piece[0],
		}
		if len(piece) == 2 {
			res.Query = piece[1]
		}
		bkpResources = append(bkpResources, res)
	}
	return bkpResources
}

// GetBackupIDFromName extracts backup id from the file name.
func GetBackupIDFromName(name string) string {
	name = strings.TrimSuffix(name, ".tar.gz")
	pattern := regexp.MustCompile(`^.+_(\d{4}_\d{2}_\d{2}_\d{2}_\d{2}_\d{2})_(.+)$`)
	matches := pattern.FindStringSubmatch(name)
	if matches == nil {
		return ""
	}
	if len(matches) < 3 {
		return ""
	}
	return matches[2]
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
