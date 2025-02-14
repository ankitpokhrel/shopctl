package run

import (
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

// NewCmdRun creates a new run command.
func NewCmdRun() *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Short:   "Run starts a backup process",
		Long:    helpText,
		Aliases: []string{"start", "exec"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value("context").(*config.StoreContext)
			strategy := cmd.Context().Value("strategy").(*config.BackupStrategy)
			client := cmd.Context().Value("gqlClient").(*api.GQLClient)
			logger := cmd.Context().Value("logger").(*tlog.Logger)

			cmdutil.ExitOnErr(run(client, ctx, strategy, logger))
			return nil
		},
	}
}

func run(client *api.GQLClient, ctx *config.StoreContext, strategy *config.BackupStrategy, logger *tlog.Logger) error {
	bkpEng := engine.NewBackup(
		ctx.Store,
		engine.WithBackupDir(strategy.BkpDir),
		engine.WithBackupPrefix(strategy.BkpPrefix),
	)
	eng := engine.New(bkpEng)

	meta, err := saveRootMeta(bkpEng, strategy)
	if err != nil {
		cmdutil.Fail("Error: Unable to create backup files; make sure that the location is writable by the user")
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

		runners = make([]runner.Runner, 0, len(strategy.Resources))
	)

	for _, resource := range strategy.Resources {
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

func saveRootMeta(bkpEng *engine.Backup, strategy *config.BackupStrategy) (*config.RootMeta, error) {
	u, _ := user.Current()

	meta, err := config.NewRootMeta(bkpEng.Dir(), config.RootMetaItems{
		ID:        bkpEng.ID(),
		Store:     bkpEng.Store(),
		TimeInit:  bkpEng.Timestamp().Unix(),
		TimeStart: time.Now().Unix(),
		Status:    string(engine.BackupStatusRunning),
		Resources: strategy.Resources,
		Kind:      strategy.Kind,
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
