package customer

import (
	"fmt"
	"path/filepath"
	"time"

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
	stats  *runner.Summary
}

// NewRunner constructs a new backup runner.
func NewRunner(eng *engine.Engine, client *api.GQLClient, logger *tlog.Logger) *Runner {
	bkpEng := eng.Doer().(*engine.Backup)

	return &Runner{
		eng:    eng,
		bkpEng: bkpEng,
		client: client,
		logger: logger,
		stats:  &runner.Summary{},
	}
}

// Kind returns runner type; implements `runner.Runner` interface.
func (r *Runner) Kind() engine.ResourceType {
	return engine.Customer
}

// Stats returns runner stats.
func (r *Runner) Stats() *runner.Summary {
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
			r.stats.Failed += 1
			r.logger.Errorf("Failed to backup resource %s: %v\n", res.ResourceType, res.Err)
		} else if res.ResourceType == engine.Customer {
			r.stats.Passed += 1
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
		r.stats.Count += len(customers.Data.Customers.Edges)

		for _, customer := range customers.Data.Customers.Edges {
			cid := engine.ExtractNumericID(customer.Node.ID)
			hash := engine.GetHashDir(cid)

			created, err := time.Parse(time.RFC3339, customer.Node.CreatedAt)
			if err != nil {
				r.logger.Error("error when parsing created time", "customerId", cid, "error", err)
				continue
			}
			path := filepath.Join(engine.Customer.RootDir(), fmt.Sprint(created.Year()), fmt.Sprintf("%d", created.Month()), hash, cid)
			r.logger.V(tlog.VL2).Infof("Customer %s: registering backup to path %s/%s", cid, r.bkpEng.Dir(), path)

			customerFn := &provider.Customer{Customer: &customer.Node}
			metafieldFn := &provider.MetaField{Client: r.client, Logger: r.logger, CustomerID: customer.Node.ID}

			r.eng.Add(engine.Customer, engine.ResourceCollection{
				engine.NewResource(engine.Customer, path, customerFn),
				engine.NewResource(engine.CustomerMetaField, path, metafieldFn),
			})
		}
	}
}
