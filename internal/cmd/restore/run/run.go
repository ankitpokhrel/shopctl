package run

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/internal/runner/restore/customer"
	"github.com/ankitpokhrel/shopctl/internal/runner/restore/product"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Run starts a data restoration process based on the given config and backup id.`

type flag struct {
	id        string
	all       bool
	resources []string
}

func (f *flag) parse(cmd *cobra.Command) {
	id, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)

	if id == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: config alias and backup id is required.

Usage:
  $ shopctl restore run -s <store> -a <alias> --id <backup id>

See 'shopctl restore run --help' for more info.`),
		)
	}

	resource, err := cmd.Flags().GetString("resource")
	cmdutil.ExitOnErr(err)

	var resources []string
	if resource != "" {
		resources = strings.Split(resource, ",")
	}

	all, err := cmd.Flags().GetBool("all")
	cmdutil.ExitOnErr(err)

	f.id = id
	f.all = all
	f.resources = resources
}

// NewCmdRun creates a new run command.
func NewCmdRun() *cobra.Command {
	cmd := cobra.Command{
		Use:     "run",
		Short:   "Run starts a data restoration process",
		Long:    helpText,
		Aliases: []string{"start", "exec"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value("context").(*config.StoreContext)
			strategy := cmd.Context().Value("strategy").(*config.BackupStrategy)
			client := cmd.Context().Value("gqlClient").(*api.GQLClient)
			logger := cmd.Context().Value("logger").(*tlog.Logger)

			cmdutil.ExitOnErr(run(cmd, client, ctx, strategy, logger))
			return nil
		},
	}
	cmd.Flags().String("id", "", "ID of the bakcup to restore")
	cmd.Flags().StringP("resource", "r", "", "Resource types to restore (comma separated)")
	cmd.Flags().Bool("all", false, "Restore all resources configured in the backup config")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, client *api.GQLClient, ctx *config.StoreContext, strategy *config.BackupStrategy, logger *tlog.Logger) error {
	flag := &flag{}
	flag.parse(cmd)

	resources := flag.resources
	if flag.all && len(resources) == 0 {
		resources = strategy.Resources
	}
	if len(resources) == 0 {
		return fmt.Errorf(`Error: please specify resources to restore or use '--all' to restore everything.

Usage:
  # Restore products and customers.
  $ shopctl restore run  -a <alias> --id <backup id> --resource=product,customer

  # Restore everything based on backup config.
  $ shopctl restore run -s <store> -a <alias> --id <backup id> --all

See 'shopctl restore run --help' for more info.`)
	}

	var (
		wg  sync.WaitGroup
		rnr runner.Runner

		runners = make([]runner.Runner, 0, len(resources))
	)

	path, err := registry.LookForDirWithSuffix(flag.id, strategy.BkpDir)
	if err != nil {
		cmdutil.Fail("Error: Unable to find backup with id %q in strategy %q of context %q", flag.id, strategy.Name, ctx.Alias)
		os.Exit(1)
	}
	eng := engine.New(engine.NewRestore(ctx.Store))

	// TODO:
	// We need to maintain the order in which products, customers and orders are restored.
	// The order should always be Product -> Customer -> Order.

	for _, resource := range resources {
		switch engine.ResourceType(resource) {
		case engine.Product:
			rnr = product.NewRunner(path, eng, client, logger)
		case engine.Customer:
			rnr = customer.NewRunner(path, eng, client, logger)
		default:
			logger.V(tlog.VL1).Warnf("Skipping '%s': Invalid resource", resource)
			continue
		}
		runners = append(runners, rnr)
	}

	start := time.Now()
	for _, rnr := range runners {
		wg.Add(1)

		go func(r runner.Runner) {
			defer wg.Done()

			if err := r.Run(); err != nil {
				logger.Errorf("%s runner exited with err: %s", r.Kind(), err.Error())
			}
		}(rnr)
	}

	wg.Wait()
	logger.Infof("Restore complete in %s", time.Since(start))
	return nil
}
