package product

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/internal/runner/restore/product/handler"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

// Runner is a product restore runner.
type Runner struct {
	path   string
	eng    *engine.Engine
	rstEng *engine.Restore
	client *api.GQLClient
	logger *tlog.Logger
	stats  map[engine.ResourceType]*runner.Summary
}

// NewRunner constructs a new restore runner.
func NewRunner(path string, eng *engine.Engine, client *api.GQLClient, logger *tlog.Logger) *Runner {
	rstEng := eng.Doer().(*engine.Restore)

	stats := make(map[engine.ResourceType]*runner.Summary)
	for _, rt := range engine.GetProductResourceTypes() {
		stats[rt] = &runner.Summary{}
	}

	return &Runner{
		path:   path,
		eng:    eng,
		rstEng: rstEng,
		client: client,
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

// Run executes product restoration process; implements `runner.Runner` interface.
func (r *Runner) Run() error {
	r.eng.Register(engine.Product)
	restoreStart := time.Now()

	go func() {
		defer r.eng.Done(engine.Product)

		// TODO: Handle/log error.
		_ = r.restore()
	}()

	for res := range r.eng.Run(engine.Product) {
		if res.Err != nil {
			r.stats[res.ResourceType].Failed += 1
			r.logger.Errorf("Failed to restore resource %s: %v\n", res.ResourceType, res.Err)
		} else {
			r.stats[res.ResourceType].Passed += 1
		}
	}

	r.logger.V(tlog.VL3).Infof(
		"Product restore complete in %s",
		time.Since(restoreStart),
	)
	return nil
}

func (r *Runner) restore() error {
	foundFiles, err := registry.GetAllInDir(r.path, ".json")
	if err != nil {
		return err
	}

	// This is the max number of resources we're expecting to process.
	const maxNumResources = 5

	// When adding resource to the resource collection we need to maintain
	// following order: Product -> Options -> Metafields -> Variants -> Media
	const (
		Product = iota
		Options
		Metafields
		Variants
		Media
	)

	// Initialize resources with fixed slots for ordering.
	resources := make(map[string][]engine.ResourceCollection)

	for f := range foundFiles {
		if f.Err != nil {
			r.logger.Warn("Skipping file due to read err", "file", f.Path, "error", f.Err)
			continue
		}

		currentID, err := extractID(f.Path)
		if err != nil {
			return err
		}

		if _, exists := resources[currentID]; !exists {
			resources[currentID] = make([]engine.ResourceCollection, maxNumResources)
		}

		switch filepath.Base(f.Path) {
		case "product.json":
			r.stats[engine.Product].Count += 1

			productFn := &handler.Product{Client: r.client, File: f, Logger: r.logger}
			optionsFn := &handler.Option{Client: r.client, File: f, Logger: r.logger}
			resources[currentID][Product] = append(
				resources[currentID][Product],
				engine.NewResource(engine.Product, r.path, productFn),
				engine.NewResource(engine.ProductOption, r.path, optionsFn),
			)
		case "product_metafields.json":
			r.stats[engine.ProductMetaField].Count += 1

			metafieldFn := &handler.Metafield{Client: r.client, File: f, Logger: r.logger}
			resources[currentID][Metafields] = append(
				resources[currentID][Metafields],
				engine.NewResource(engine.ProductMetaField, r.path, metafieldFn),
			)
		case "product_variants.json":
			r.stats[engine.ProductVariant].Count += 1

			variantFn := &handler.Variant{Client: r.client, File: f, Logger: r.logger}
			resources[currentID][Variants] = append(
				resources[currentID][Variants],
				engine.NewResource(engine.ProductVariant, r.path, variantFn),
			)
		case "product_media.json":
			r.stats[engine.ProductMedia].Count += 1

			mediaFn := &handler.Media{Client: r.client, File: f, Logger: r.logger}
			resources[currentID][Media] = append(
				resources[currentID][Media],
				engine.NewResource(engine.ProductMedia, r.path, mediaFn),
			)
		}
	}

	// Flatten resources for each currentID in the defined order.
	for _, orderedResources := range resources {
		var flattened engine.ResourceCollection
		for _, rc := range orderedResources {
			flattened = append(flattened, rc...)
		}
		r.eng.Add(engine.Product, flattened)
	}
	return nil
}

func extractID(path string) (string, error) {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

	if len(parts) < 2 {
		return "", fmt.Errorf("path does not have enough elements")
	}
	return parts[len(parts)-2], nil
}
