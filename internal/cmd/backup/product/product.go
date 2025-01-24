package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/product/provider"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const (
	helpText = `Product initiates a product backup process for a shopify store.

Use this command to back up the entire product catalog or a subset based on filters like category, tags, or date.`

	examples = `$ shopctl backup product

# Back up first 100 products
$ shopctl backup product --limit 100

# Back up products from a specific category (e.g., "Winter Collection")
$ shopctl backup product --category "Winter Collection"

# Back up products with specific tags (e.g., "Sale" and "New")
$ shopctl backup product --tags "Sale,New"

# Back up products created or updated after a certain date
$ shopctl backup product --since "2024-01-01"

# Perform a dry run to see which products would be backed up
$ shopctl backup product --dry-run

# Perform an incremental backup (only new or updated products)
$ shopctl backup product --incremental`

	batchSize = 250
)

var lgr *tlog.Logger

// NewCmdProduct creates a new product backup command.
func NewCmdProduct(eng *engine.Engine) *cobra.Command {
	return &cobra.Command{
		Use:     "product",
		Short:   "Product initiates product backup",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Type assert engine.doer with engine.Backup type.
			bkpEng := eng.Doer().(*engine.Backup)

			client := cmd.Context().Value("gqlClient").(*api.GQLClient)
			lgr = cmd.Context().Value("logger").(*tlog.Logger)

			return product(eng, bkpEng, client)
		},
	}
}

func product(eng *engine.Engine, bkpEng *engine.Backup, client *api.GQLClient) error {
	eng.Register(engine.Product)
	backupStart := time.Now()

	go func() {
		defer eng.Done(engine.Product)
		backupProduct(eng, bkpEng, client, batchSize, nil)
	}()

	for res := range eng.Run(engine.Product) {
		if res.Err != nil {
			lgr.Errorf("Failed to backup resource %s: %v\n", res.ResourceType, res.Err)
		}
	}

	lgr.V(tlog.VL3).Infof(
		"Product backup complete in %v",
		time.Since(backupStart),
	)
	return nil
}

func backupProduct(eng *engine.Engine, bkpEng *engine.Backup, client *api.GQLClient, limit int, after *string) {
	productsCh := make(chan *api.ProductsResponse, batchSize)

	go func() {
		defer close(productsCh)

		if err := client.GetAllProducts(productsCh, limit, after); err != nil {
			lgr.Error("error when fetching products", "limit", limit, "after", after, "error", err)
		}
	}()

	for products := range productsCh {
		for _, product := range products.Data.Products.Edges {
			pid := engine.ExtractNumericID(product.Node.ID)
			hash := engine.GetHashDir(pid)

			created, err := time.Parse(time.RFC3339, product.Node.CreatedAt)
			if err != nil {
				lgr.Error("error when parsing created time", "productId", pid, "error", err)
				continue
			}
			path := filepath.Join(fmt.Sprint(created.Year()), fmt.Sprintf("%d", created.Month()), hash, pid)
			lgr.V(tlog.VL2).Infof("Product %s: registering backup to path %s/%s", pid, bkpEng.Dir(), path)

			productFn := &provider.Product{Product: &product.Node}
			variantFn := &provider.Variant{Client: client, Logger: lgr, ProductID: product.Node.ID}
			mediaFn := &provider.Media{Client: client, Logger: lgr, ProductID: product.Node.ID}
			metafieldFn := &provider.MetaField{Client: client, Logger: lgr, ProductID: product.Node.ID}

			eng.Add(engine.Product, engine.ResourceCollection{
				engine.NewResource(engine.Product, path, productFn),
				engine.NewResource(engine.ProductVariant, path, variantFn),
				engine.NewResource(engine.ProductMedia, path, mediaFn),
				engine.NewResource(engine.ProductMetaField, path, metafieldFn),
			})
		}
	}
}
