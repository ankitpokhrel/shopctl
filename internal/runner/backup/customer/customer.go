package customer

import (
	"path/filepath"
	"time"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/internal/runner/backup/customer/provider"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const batchSize = 250

// Runner is a customer backup runner.
type Runner struct {
	eng    *engine.Engine
	bkpEng *engine.Backup
	client *api.GQLClient
	logger *tlog.Logger
	stats  map[engine.ResourceType]*runner.Summary
}

// NewRunner constructs a new backup runner.
func NewRunner(eng *engine.Engine, client *api.GQLClient, logger *tlog.Logger) *Runner {
	bkpEng := eng.Doer().(*engine.Backup)

	stats := make(map[engine.ResourceType]*runner.Summary)
	for _, rt := range engine.GetCustomerResourceTypes() {
		stats[rt] = &runner.Summary{}
	}

	return &Runner{
		eng:    eng,
		bkpEng: bkpEng,
		client: client,
		logger: logger,
		stats:  stats,
	}
}

// Kind returns runner type; implements `runner.Runner` interface.
func (r *Runner) Kind() engine.ResourceType {
	return engine.Customer
}

// Stats returns runner stats.
func (r *Runner) Stats() map[engine.ResourceType]*runner.Summary {
	return r.stats
}

// Run executes customer backup; implements `runner.Runner` interface.
func (r *Runner) Run() error {
	r.eng.Register(engine.Customer)
	backupStart := time.Now()

	go func() {
		defer r.eng.Done(engine.Customer)
		r.backup(batchSize, nil)
	}()

	for res := range r.eng.Run(engine.Customer) {
		if res.Err != nil {
			r.stats[res.ResourceType].Failed += 1
			r.logger.Errorf("Failed to backup resource %s: %v\n", res.ResourceType, res.Err)
		} else if res.ResourceType == engine.Customer {
			r.stats[res.ResourceType].Passed += 1
		}
	}

	r.logger.V(tlog.VL3).Infof(
		"Customer backup complete in %s",
		time.Since(backupStart),
	)
	return nil
}

func (r *Runner) backup(limit int, after *string) {
	customersCh := make(chan *api.CustomersResponse, batchSize)

	go func() {
		defer close(customersCh)

		if err := r.client.GetAllCustomers(customersCh, limit, after); err != nil {
			r.logger.Error("error when fetching customres", "limit", limit, "after", after, "error", err)
		}
	}()

	for customers := range customersCh {
		r.stats[r.Kind()].Count += len(customers.Data.Customers.Nodes)

		for _, customer := range customers.Data.Customers.Nodes {
			cid := shopctl.ExtractNumericID(customer.ID)

			path := filepath.Join(engine.Customer.RootDir(), cid)
			r.logger.V(tlog.VL2).Infof("Customer %s: registering backup to path %s/%s", cid, r.bkpEng.Dir(), path)

			customerFn := &provider.Customer{Customer: &customer}
			metafieldFn := &provider.MetaField{Client: r.client, Logger: r.logger, CustomerID: customer.ID}

			parent := engine.NewResource(engine.Customer, path, customerFn)

			r.eng.Add(engine.Customer, engine.ResourceCollection{
				Parent: &parent,
				Children: []engine.Resource{
					engine.NewResource(engine.CustomerMetaField, path, metafieldFn),
				},
			})
		}
	}
}
