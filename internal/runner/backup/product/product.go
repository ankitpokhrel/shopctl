package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/runner/backup/product/provider"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const batchSize = 250

// Runner is a product backup runner.
type Runner struct {
	eng    *engine.Engine
	bkpEng *engine.Backup
	client *api.GQLClient
	logger *tlog.Logger
}

// NewRunner constructs a new backup runner.
func NewRunner(eng *engine.Engine, client *api.GQLClient, logger *tlog.Logger) *Runner {
	bkpEng := eng.Doer().(*engine.Backup)

	return &Runner{
		eng:    eng,
		bkpEng: bkpEng,
		client: client,
		logger: logger,
	}
}

// Run implements `runner.Runner` interface.
func (r *Runner) Run() error {
	r.eng.Register(engine.Product)
	backupStart := time.Now()

	go func() {
		defer r.eng.Done(engine.Product)
		r.backup(batchSize, nil)
	}()

	for res := range r.eng.Run(engine.Product) {
		if res.Err != nil {
			r.logger.Errorf("Failed to backup resource %s: %v\n", res.ResourceType, res.Err)
		}
	}

	r.logger.V(tlog.VL3).Infof(
		"Product backup complete in %v",
		time.Since(backupStart),
	)
	return nil
}

func (r *Runner) backup(limit int, after *string) {
	productsCh := make(chan *api.ProductsResponse, batchSize)

	go func() {
		defer close(productsCh)

		if err := r.client.GetAllProducts(productsCh, limit, after); err != nil {
			r.logger.Error("error when fetching products", "limit", limit, "after", after, "error", err)
		}
	}()

	for products := range productsCh {
		for _, product := range products.Data.Products.Edges {
			pid := engine.ExtractNumericID(product.Node.ID)
			hash := engine.GetHashDir(pid)

			created, err := time.Parse(time.RFC3339, product.Node.CreatedAt)
			if err != nil {
				r.logger.Error("error when parsing created time", "productId", pid, "error", err)
				continue
			}
			path := filepath.Join(fmt.Sprint(created.Year()), fmt.Sprintf("%d", created.Month()), hash, pid)
			r.logger.V(tlog.VL2).Infof("Product %s: registering backup to path %s/%s", pid, r.bkpEng.Dir(), path)

			productFn := &provider.Product{Product: &product.Node}
			variantFn := &provider.Variant{Client: r.client, Logger: r.logger, ProductID: product.Node.ID}
			mediaFn := &provider.Media{Client: r.client, Logger: r.logger, ProductID: product.Node.ID}
			metafieldFn := &provider.MetaField{Client: r.client, Logger: r.logger, ProductID: product.Node.ID}

			r.eng.Add(engine.Product, engine.ResourceCollection{
				engine.NewResource(engine.Product, path, productFn),
				engine.NewResource(engine.ProductVariant, path, variantFn),
				engine.NewResource(engine.ProductMedia, path, mediaFn),
				engine.NewResource(engine.ProductMetaField, path, metafieldFn),
			})
		}
	}
}
