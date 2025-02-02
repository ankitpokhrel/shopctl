package run

import (
	"fmt"
	"os"
	"strings"
	"sync"

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
	store     string
	alias     string
	id        string
	all       bool
	resources []string
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	alias, err := cmd.Flags().GetString("alias")
	cmdutil.ExitOnErr(err)

	id, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)

	if alias == "" || id == "" {
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

	f.store = store
	f.alias = alias
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
		RunE:    run,
	}
	cmd.Flags().StringP("alias", "a", "", "Alias of the config to run")
	cmd.Flags().String("id", "", "ID of the bakcup to restore")
	cmd.Flags().StringP("resource", "r", "", "Resource types to restore (comma separated)")
	cmd.Flags().Bool("all", false, "Restore all resources configured in the backup config")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, _ []string) error {
	flag := &flag{}
	flag.parse(cmd)

	preset, err := config.ReadAllPreset(flag.store, flag.alias)
	if err != nil {
		cmdutil.Fail("Error: Preset with alias '%s' couldn't be found for store '%s'", flag.alias, flag.store)
		os.Exit(1)
	}

	eng := engine.New(engine.NewRestore(flag.store))
	client := cmd.Context().Value("gqlClient").(*api.GQLClient)
	logger := cmd.Context().Value("logger").(*tlog.Logger)

	if err != nil {
		cmdutil.Fail("Error: Unable to create backup files; make sure that the location is writable by the user")
		os.Exit(1)
	}

	resources := flag.resources
	if flag.all && len(resources) == 0 {
		resources = preset.Resources
	}
	if len(resources) == 0 {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: please specify resources to restore or use '--all' to restore everything.

Usage:
  # Restore products and customers.
  $ shopctl restore run -s <store> -a <alias> --id <backup id> --resource=product,customer

  # Restore everything based on backup config.
  $ shopctl restore run -s <store> -a <alias> --id <backup id> --all

See 'shopctl restore run --help' for more info.`),
		)
	}

	var (
		wg  sync.WaitGroup
		rnr runner.Runner

		runners = make([]runner.Runner, 0, len(resources))
	)

	path, err := registry.LookForDirWithSuffix(flag.id, preset.BkpDir)
	if err != nil {
		cmdutil.Fail("Error: Unable to find backup with id '%s' in config '%s' of store '%s'", flag.id, flag.alias, flag.store)
		os.Exit(1)
	}

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
	return nil
}
