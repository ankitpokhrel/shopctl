package run

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/runner/backup/product"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Run starts a backup process based on the given config.`

type flag struct {
	store string
	alias string
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	alias, err := cmd.Flags().GetString("alias")
	cmdutil.ExitOnErr(err)

	if alias == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: config alias is required.

Usage:
  $ shopctl backup run -s <store> -a <alias>

See 'shopctl backup run --help' for more info.`),
		)
	}

	f.store = store
	f.alias = alias
}

// NewCmdRun creates a new run command.
func NewCmdRun() *cobra.Command {
	cmd := cobra.Command{
		Use:     "run",
		Short:   "Run starts a backup process",
		Long:    helpText,
		Aliases: []string{"start", "exec"},
		RunE:    run,
	}
	cmd.Flags().StringP("alias", "a", "", "Alias of the config to run")

	return &cmd
}

func run(cmd *cobra.Command, _ []string) error {
	flag := &flag{}
	flag.parse(cmd)

	preset, err := config.ReadAllPreset(flag.store, flag.alias)
	if err != nil {
		cmdutil.Fail("Preset with alias '%s' couldn't be found for store '%s'", flag.alias, flag.store)
		os.Exit(1)
	}

	eng := engine.New(
		engine.NewBackup(
			engine.WithBackupDir(preset.BkpDir),
			engine.WithBackupPrefix(preset.BkpPrefix),
		),
	)
	client := cmd.Context().Value("gqlClient").(*api.GQLClient)
	logger := cmd.Context().Value("logger").(*tlog.Logger)

	var wg sync.WaitGroup

	for _, resource := range preset.Resources {
		switch engine.ResourceType(resource) {
		case engine.Product:
			wg.Add(1)
			pr := product.NewRunner(eng, client, logger)

			go func() {
				defer wg.Done()

				if err := pr.Run(); err != nil {
					logger.Errorf("Product runner exited with err: %s", err.Error())
				}
			}()
		default:
			logger.V(tlog.VL1).Warnf("Skipping '%s': Invalid resource", resource)
		}
	}

	wg.Wait()
	return nil
}
