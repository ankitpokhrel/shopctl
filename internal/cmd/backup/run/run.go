package run

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

const (
	helpText = `Run starts a backup process based on the given config.`

	repeatedDashes = "" +
		"-------------------------------"
	repeatedEquals = "" +
		"==============================="
)

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

			cmdutil.ExitOnErr(run(cmd, client, ctx, strategy, logger))
			return nil
		},
	}
}

func run(cmd *cobra.Command, client *api.GQLClient, ctx *config.StoreContext, strategy *config.BackupStrategy, logger *tlog.Logger) error {
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	bkpEng := engine.NewBackup(
		ctx.Store,
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
			logger.Warnf("Skipping '%s': Invalid resource", resource)
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
	logger.V(tlog.VL1).Infof("We're done with fetching data. Archiving...")

	if err := cmdutil.Archive(bkpEng.Root(), strategy.BkpDir, bkpEng.Dir()); err != nil {
		return err
	}
	logger.Infof("Backup complete in %s", time.Since(start))

	if !quiet {
		summarize(strategy, bkpEng, runners)
	}
	return nil
}

func saveRootMeta(bkpEng *engine.Backup, strategy *config.BackupStrategy) (*config.RootMeta, error) {
	u, _ := user.Current()

	meta, err := config.NewRootMeta(bkpEng.Root(), config.RootMetaItems{
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

func summarize(strategy *config.BackupStrategy, bkpEng *engine.Backup, runners []runner.Runner) {
	title := func(msg string, sep string) {
		fmt.Printf("%s\n%s\n%s\n", sep, msg, sep)
	}

	title("BACKUP SUMMARY", repeatedEquals)
	fmt.Printf(`ID: %s
Strategy: %s
Store: %s
Type: %s
Path: %s
`,
		bkpEng.ID(), strategy.Name, bkpEng.Store(),
		strategy.Kind, filepath.Join(strategy.BkpDir, bkpEng.Dir()),
	)
	for _, rnr := range runners {
		fmt.Println()
		title(strings.ToTitle(string(rnr.Kind())), repeatedDashes)
		fmt.Println(rnr.Stats())
	}
}
