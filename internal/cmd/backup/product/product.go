package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
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

	batchSize = 100
)

var lgr *tlog.Logger

// NewCmdProduct creates a new product backup command.
func NewCmdProduct(bkp *engine.Backup) *cobra.Command {
	return &cobra.Command{
		Use:     "product",
		Short:   "Product initiates product backup",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value("gqlClient").(*api.GQLClient)

			v, _ := cmd.Flags().GetCount("verbose")
			lgr = tlog.New(tlog.VerboseLevel(v))

			return product(bkp, client)
		},
	}
}

func product(bkp *engine.Backup, client *api.GQLClient) error {
	bkp.Register(engine.Product)

	backupStart := time.Now()
	go func() {
		defer bkp.Done(engine.Product)

		for res := range bkp.Do(engine.Product) {
			if res.Err != nil {
				lgr.Errorf("Failed to backup resource %s with id %s: %v\n", res.ResourceType, res.ResourceID, res.Err)
			} else {
				lgr.Infof("Resource %s with id %s backed up successfully\n", res.ResourceType, res.ResourceID)
			}
		}
	}()

	// TODO: Handle/log error.
	_ = backupProduct(bkp, client, batchSize, nil)

	lgr.V(tlog.VL3).Infof(
		"Product backup complete in %v",
		time.Since(backupStart),
	)
	return nil
}

func backupProduct(bkp *engine.Backup, client *api.GQLClient, limit int, after *string) error {
	cursor := "<nil>"
	if after != nil {
		cursor = *after
	}

	productsFn := func() (*api.ProductsResponse, error) {
		return client.GetProducts(limit, after)
	}
	products, err := timeit(productsFn, "Request to fetch %d products after %v", limit, cursor)()
	if err != nil {
		lgr.Error("Unable to fetch products", "after", cursor, "error", err)
		return err
	}

	for _, product := range products.Data.Products.Edges {
		pid := engine.ExtractNumericID(product.Node.ID)
		hash := engine.GetHashDir(pid)

		productFn := func() (any, error) {
			lgr.Infof("Product %s: processing started", pid)
			return product.Node, nil
		}

		variantFn := func() (any, error) {
			lgr.V(tlog.VL1).Infof("Product %s: processing variants", pid)

			variants, err := client.GetProductVariants(product.Node.ID)
			if err != nil {
				lgr.Error("error when fetching variants", "productId", pid, "error", err)
				return nil, err
			}
			return variants.Data.Product.Variants.Edges, nil
		}

		mediaFn := func() (any, error) {
			lgr.V(tlog.VL1).Infof("Product %s: processing media items", pid)

			medias, err := client.GetProductMedias(product.Node.ID)
			if err != nil {
				lgr.Error("error when fetching media", "", pid, "error", err)
				return nil, err
			}
			return medias.Data.Product.Media.Edges, err
		}

		created, err := time.Parse(time.RFC3339, product.Node.CreatedAt)
		if err != nil {
			return err
		}
		path := filepath.Join(fmt.Sprint(created.Year()), fmt.Sprintf("%d", created.Month()), hash, pid)
		lgr.V(tlog.VL2).Infof("Product %s: registering backup to path %s/%s", pid, bkp.Dir(), path)

		bkp.Add(engine.Product, engine.NewResourceCollection(
			pid, path,
			engine.NewResource(engine.Product, productFn),
			engine.NewResource(engine.ProductVariant, timeit(variantFn, "Product %s: fetching variants", pid)),
			engine.NewResource(engine.ProductMedia, timeit(mediaFn, "Product %s: fetching media items", pid)),
		))
	}

	if products.Data.Products.PageInfo.HasNextPage {
		return backupProduct(bkp, client, limit, products.Data.Products.PageInfo.EndCursor)
	}
	return nil
}

// timeit is a higher-order function that wraps around a function and times its execution.
func timeit[T any](fn func() (T, error), msg string, args ...any) func() (T, error) {
	return func() (T, error) {
		start := time.Now()
		result, err := fn()

		msg = fmt.Sprintf(msg, args...)
		lgr.V(tlog.VL3).Infof("%s took %v", msg, time.Since(start))

		return result, err
	}
}
