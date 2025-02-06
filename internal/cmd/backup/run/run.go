package run

import (
	"fmt"
	"os"
	"os/user"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/internal/runner/backup/customer"
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
		cmdutil.Fail("Error: Preset with alias '%s' couldn't be found for store '%s'", flag.alias, flag.store)
		os.Exit(1)
	}

	bkpEng := engine.NewBackup(
		flag.store,
		engine.WithBackupDir(preset.BkpDir),
		engine.WithBackupPrefix(preset.BkpPrefix),
	)
	eng := engine.New(bkpEng)
	client := cmd.Context().Value("gqlClient").(*api.GQLClient)
	logger := cmd.Context().Value("logger").(*tlog.Logger)

	meta, err := saveRootMeta(bkpEng, preset)
	if err != nil {
		cmdutil.Fail("Error: Unable to create backup files; make sure that the location is writable by the user", flag.alias, flag.store)
		os.Exit(1)
	}

	defer func() {
		err := meta.Set(map[string]any{
			config.KeyStatus:  string(engine.BackupStatusComplete),
			config.KeyTimeEnd: time.Now().Unix(),
		})
		if err != nil {
			logger.Errorf("Unable to update metadata after backup run: %s", err.Error())
		}
	}()

	var (
		wg  sync.WaitGroup
		rnr runner.Runner

		runners = make([]runner.Runner, 0, len(preset.Resources))
	)

	for _, resource := range preset.Resources {
		switch engine.ResourceType(resource) {
		case engine.Product:
			rnr = product.NewRunner(eng, client, logger)
		case engine.Customer:
			rnr = customer.NewRunner(eng, client, logger)
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
	logger.Infof("Backup complete in %s", time.Since(start))
	return nil
}

func saveRootMeta(bkpEng *engine.Backup, preset *config.PresetItems) (*config.RootMeta, error) {
	u, _ := user.Current()

	meta, err := config.NewRootMeta(bkpEng.Dir(), config.RootMetaItems{
		ID:        bkpEng.ID(),
		Store:     bkpEng.Store(),
		TimeInit:  bkpEng.Timestamp().Unix(),
		TimeStart: time.Now().Unix(),
		Status:    string(engine.BackupStatusRunning),
		Resources: preset.Resources,
		Kind:      preset.Kind,
		User:      u.Username,
	})
	if err != nil {
		return nil, err
	}
	if err := meta.Save(); err != nil {
		return nil, err
	}
	return meta, nil
}
