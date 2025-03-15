package customer

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/internal/runner/restore/customer/handler"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

// Runner is a product restore runner.
type Runner struct {
	path     string
	eng      *engine.Engine
	rstEng   *engine.Restore
	client   *api.GQLClient
	logger   *tlog.Logger
	stats    map[engine.ResourceType]*runner.Summary
	filters  *runner.RestoreFilter
	isDryRun bool
}

// NewRunner constructs a new restore runner.
func NewRunner(path string, eng *engine.Engine, client *api.GQLClient, logger *tlog.Logger, filters *runner.RestoreFilter, isDryRun bool) *Runner {
	rstEng := eng.Doer().(*engine.Restore)

	stats := make(map[engine.ResourceType]*runner.Summary)
	for _, rt := range engine.GetCustomerResourceTypes() {
		stats[rt] = &runner.Summary{}
	}

	return &Runner{
		path:     path,
		eng:      eng,
		rstEng:   rstEng,
		client:   client,
		logger:   logger,
		stats:    stats,
		filters:  filters,
		isDryRun: isDryRun,
	}
}

// Kind returns runner type; implements `runner.Runner` interface.
func (r *Runner) Kind() engine.ResourceType {
	return engine.Customer
}

// Run executes customer restoration process; implements `runner.Runner` interface.
func (r *Runner) Run() error {
	r.eng.Register(engine.Customer)
	restoreStart := time.Now()

	go func() {
		defer r.eng.Done(engine.Customer)

		// TODO: Handle/log error.
		_ = r.restore()
	}()

	for res := range r.eng.Run(engine.Customer) {
		if res.Err != nil && !errors.Is(res.Err, engine.ErrSkipChildren) {
			r.stats[res.ResourceType].Failed += 1
			r.logger.Errorf("Failed to restore resource %s: %v\n", res.ResourceType, res.Err)
		}
	}

	r.logger.V(tlog.VL3).Infof(
		"Customer restore complete in %s",
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
	const maxNumResources = 2

	// When adding resource to the resource collection we need to maintain
	// following order: Customer -> Metafields
	const (
		Customer = iota
		Metafields
	)

	// Initialize resources with fixed slots for ordering.
	resources := make(map[string][][]engine.Resource)

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
			resources[currentID] = make([][]engine.Resource, maxNumResources)
		}

		switch filepath.Base(f.Path) {
		case "customer.json":
			customerFn := &handler.Customer{Client: r.client, File: f, Filter: r.filters, Logger: r.logger, Summary: r.stats[engine.Customer], DryRun: r.isDryRun}
			resources[currentID][Customer] = append(
				resources[currentID][Customer],
				engine.NewResource(engine.Customer, r.path, customerFn),
			)
		case "customer_metafields.json":
			metafieldFn := &handler.Metafield{Client: r.client, File: f, Logger: r.logger, Summary: r.stats[engine.CustomerMetaField], DryRun: r.isDryRun}
			resources[currentID][Metafields] = append(
				resources[currentID][Metafields],
				engine.NewResource(engine.CustomerMetaField, r.path, metafieldFn),
			)

		}
	}

	// Flatten resources for each currentID in the defined order.
	for _, orderedResources := range resources {
		var flattened engine.ResourceCollection

		if len(orderedResources[Customer]) > 0 {
			flattened.Parent = &orderedResources[Customer][0]
		}

		for idx, rc := range orderedResources {
			if idx == Customer {
				continue
			}
			flattened.Children = append(flattened.Children, rc...)
		}
		if flattened.Parent != nil {
			r.eng.Add(engine.Customer, flattened)
		}
	}
	return nil
}

// Stats returns runner stats.
func (r *Runner) Stats() map[engine.ResourceType]*runner.Summary {
	return r.stats
}

func extractID(path string) (string, error) {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

	if len(parts) < 2 {
		return "", fmt.Errorf("path does not have enough elements")
	}
	return parts[len(parts)-2], nil
}
