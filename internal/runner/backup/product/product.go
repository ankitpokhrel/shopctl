package product

import (
	"path/filepath"
	"time"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/internal/runner/backup/product/provider"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const batchSize = 250

// Runner is a product backup runner.
type Runner struct {
	eng    *engine.Engine
	bkpEng *engine.Backup
	client *api.GQLClient
	filter *string
	logger *tlog.Logger
	stats  map[engine.ResourceType]*runner.Summary
}

// NewRunner constructs a new backup runner.
func NewRunner(eng *engine.Engine, client *api.GQLClient, filter string, logger *tlog.Logger) *Runner {
	bkpEng := eng.Doer().(*engine.Backup)

	var f *string
	if filter != "" {
		f = &filter
	}

	stats := make(map[engine.ResourceType]*runner.Summary)
	for _, rt := range engine.GetProductResourceTypes() {
		stats[rt] = &runner.Summary{}
	}

	return &Runner{
		eng:    eng,
		bkpEng: bkpEng,
		client: client,
		filter: f,
		logger: logger,
		stats:  stats,
	}
}

// Kind returns runner type; implements `runner.Runner` interface.
func (r *Runner) Kind() engine.ResourceType {
	return engine.Product
}

// Stats returns runner stats.
func (r *Runner) Stats() map[engine.ResourceType]*runner.Summary {
	return r.stats
}

// Run executes product backup; implements `runner.Runner` interface.
func (r *Runner) Run() error {
	r.eng.Register(engine.Product)
	backupStart := time.Now()

	go func() {
		defer r.eng.Done(engine.Product)
		r.backup(batchSize, nil, r.filter)
	}()

	for res := range r.eng.Run(engine.Product) {
		if res.Err != nil {
			r.stats[res.ResourceType].Failed += 1
			r.logger.Errorf("Failed to backup resource %s: %v\n", res.ResourceType, res.Err)
		} else if res.ResourceType == engine.Product {
			r.stats[res.ResourceType].Passed += 1
		}
	}

	r.logger.V(tlog.VL3).Infof(
		"Product backup complete in %s",
		time.Since(backupStart),
	)
	return nil
}

func (r *Runner) backup(limit int, after *string, query *string) {
	productsCh := make(chan *api.ProductsResponse, batchSize)

	go func() {
		defer close(productsCh)

		if err := r.client.GetAllProducts(productsCh, limit, after, query); err != nil {
			r.logger.Error("Failed to fetch products", "limit", limit, "after", after, "error", err)
		}
	}()

	for products := range productsCh {
		r.stats[r.Kind()].Count += len(products.Data.Products.Edges)

		for _, product := range products.Data.Products.Edges {
			pid := engine.ExtractNumericID(product.Node.ID)

			path := filepath.Join(engine.Product.RootDir(), pid)
			r.logger.V(tlog.VL2).Infof("Product %s: registering backup to path %s/%s", pid, r.bkpEng.Dir(), path)

			productFn := &provider.Product{Product: &product.Node}
			variantFn := &provider.Variant{Client: r.client, Logger: r.logger, ProductID: product.Node.ID}
			mediaFn := &provider.Media{Client: r.client, Logger: r.logger, ProductID: product.Node.ID}
			metafieldFn := &provider.MetaField{Client: r.client, Logger: r.logger, ProductID: product.Node.ID}

			parent := engine.NewResource(engine.Product, path, productFn)

			r.eng.Add(engine.Product, engine.ResourceCollection{
				Parent: &parent,
				Children: []engine.Resource{
					engine.NewResource(engine.ProductVariant, path, variantFn),
					engine.NewResource(engine.ProductMedia, path, mediaFn),
					engine.NewResource(engine.ProductMetaField, path, metafieldFn),
				},
			})
		}
	}
}
