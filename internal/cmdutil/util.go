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

// HelpErrorf prepares error message by appending its usage.
func HelpErrorf(msg string, examples string) error {
	lines := strings.Split(examples, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return fmt.Errorf(msg+"\n\n\033[1mUsage:\033[0m\n\n%s", strings.Join(lines, "\n"))
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
