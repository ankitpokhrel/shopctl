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
	"github.com/ankitpokhrel/shopctl/schema"
)

// ContextValue is a string type to use as a key for `context.SetValue`.
// This helps avoid conflicts with the default `string` type.
type ContextValue string

const (
	KeyContext    ContextValue = "context"
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

// ShopifyProductVariantID formats Shopify product variant ID.
func ShopifyProductVariantID(id string) string {
	prefix := "gid://shopify/ProductVariant"
	if strings.HasPrefix(id, prefix) {
		return id
	}
	if _, err := strconv.Atoi(id); err != nil {
		return "" // Not an integer id.
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
		return "" // Not an integer id.
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

// ParseRestoreFilters parses resource filters.
func ParseRestoreFilters(input string) (map[string][]string, []string, error) {
	grammar := regexp.MustCompile(`^(\w+:(?:'[^']*'|"[^"]*"|[\w,-]+))( (?i)(AND|OR) \w+:(?:'[^']*'|"[^"]*"|[\w,-]+))*$`)
	if !grammar.MatchString(input) {
		return nil, nil, fmt.Errorf("invalid input format: %s", input)
	}

	conditionRegex := regexp.MustCompile(`(\w+):(?:'([^']*)'|"([^"]*)"|([\w,-]+))`)
	separatorRegex := regexp.MustCompile(` (?i)(AND|OR) `)

	parts := separatorRegex.Split(input, -1)
	separators := separatorRegex.FindAllString(input, -1)

	result := make(map[string][]string)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		matches := conditionRegex.FindStringSubmatch(part)
		if len(matches) < 5 { //nolint:mnd
			return nil, nil, fmt.Errorf("invalid filter format: %s", part)
		}

		key := matches[1]
		var value string
		switch {
		case matches[2] != "":
			// Single-quoted value.
			value = matches[2]
		case matches[3] != "":
			// Double-quoted value.
			value = matches[3]
		default:
			// Unquoted value.
			value = matches[4]
		}

		values := strings.Split(value, ",")
		for _, v := range values {
			result[key] = append(result[key], strings.TrimSpace(v))
		}
	}

	for i := range separators {
		separators[i] = strings.ToUpper(strings.TrimSpace(separators[i]))
	}
	return result, separators, nil
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

// HelpErrorf prepares error message by appending its usage.
func HelpErrorf(msg string, examples string) error {
	lines := strings.Split(examples, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return fmt.Errorf(msg+"\n\n\033[1mUsage:\033[0m\n\n%s", strings.Join(lines, "\n"))
}

// SplitKeyVal splits string input separated by a colon.
func SplitKeyVal(items string) (string, string, error) {
	parts := strings.SplitN(items, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("values should be in the following format, Key:Value")
	}
	return parts[0], strings.TrimSpace(parts[1]), nil
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
