package run

import (
	"fmt"
	"os"
	"path/filepath"
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

const (
	helpText = `Run starts a data restoration process based on the given config and backup id.`
	examples = `# Restore everything for the context from the latest backup
$ shopctl restore run --latest --all

# Restore some products for the context from the latest backup
$ shopctl restore run --latest -r product="id1,id2,id3"

# Restore everything for the context from the given backup id
$ shopctl restore run --backup-id 3820045c0c --all

# Restore all products from the given context and backup id
$ shopctl restore run -c mycontext --backup-id 3820045c0c -r product

# Restore specific products from the latest backup of the given context and strategy
$ shopctl restore run -c mycontext -s mystrategy --latest -r product="id:id1,id2,id3"

# Restore specific products and all customers from the latest backup
$ shopctl restore run --latest -r product="id:id1,id2,id3" -r customer

# Restore products and customers directly from the given backup path
$ shopctl restore run --backup-path /path/to/unzipped/bkp -r product -r customer

# Dry run executes the restoration process and print logs without making an actual API call
$ shopctl restore run --latest --all --dry-run
$ shopctl restore run -r product --backup-id 3820045c0c --dry-run -vvv
`
)

type flag struct {
	id        string
	path      string
	all       bool
	latest    bool
	resources []config.BackupResource
	dryRun    bool
	quiet     bool
}

func (f *flag) parse(cmd *cobra.Command) {
	id, err := cmd.Flags().GetString("backup-id")
	cmdutil.ExitOnErr(err)

	path, err := cmd.Flags().GetString("backup-path")
	cmdutil.ExitOnErr(err)

	resources, err := cmd.Flags().GetStringArray("resource")
	cmdutil.ExitOnErr(err)

	all, err := cmd.Flags().GetBool("all")
	cmdutil.ExitOnErr(err)

	latest, err := cmd.Flags().GetBool("latest")
	cmdutil.ExitOnErr(err)

	dryRun, err := cmd.Flags().GetBool("dry-run")
	cmdutil.ExitOnErr(err)

	quiet, err := cmd.Flags().GetBool("quiet")
	cmdutil.ExitOnErr(err)

	if !latest && id == "" && path == "" {
		cmdutil.ExitOnErr(helpErrorf("Error: either '--backup-id', '--backup-path' or the '--latest' flag is required."))
	}

	// Reset other options based on precedence to prevent accidental mixups.
	// Precedence: latest --> backup-id --> backup-path
	if latest {
		id = ""
		path = ""
	} else if id != "" {
		path = ""
	}

	if path != "" && len(resources) == 0 {
		cmdutil.ExitOnErr(helpErrorf("Error: the '--backup-path' option require resources to process; use '--resource' option to provide a list of resources to restore."))
	}

	if len(resources) == 0 && !all {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: please specify resources to restore or use '--all' to restore everything.

Usage:

  # Restore everything from the latest backup
  $ shopctl restore run --latest --all

  # Restore all products and customers from the given backup id
  $ shopctl restore run --backup-id 3820045c0c -r product -r customer

  # Restore some products and all customers from the given backup id
  $ shopctl restore run --backup-id 3820045c0c -r product="id:id1,id2,id3" -r customer

  See 'shopctl restore run --help' for more info.`),
		)
	}

	f.id = id
	f.path = path
	f.all = all
	f.latest = latest
	f.resources = cmdutil.ParseBackupResource(resources)
	f.dryRun = dryRun
	f.quiet = quiet
}

// NewCmdRun creates a new run command.
func NewCmdRun() *cobra.Command {
	cmd := cobra.Command{
		Use:     "run",
		Short:   "Run starts a data restoration process",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"start", "exec"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shopCfg := cmd.Context().Value("shopCfg").(*config.ShopConfig)
			ctx := cmd.Context().Value("context").(*config.StoreContext)
			client := cmd.Context().Value("gqlClient").(*api.GQLClient)
			logger := cmd.Context().Value("logger").(*tlog.Logger)

			cmdutil.ExitOnErr(run(cmd, client, shopCfg, ctx, logger))
			return nil
		},
	}
	cmd.Flags().StringP("backup-id", "b", "", "ID of the backup to restore from")
	cmd.Flags().StringP("backup-path", "p", "", "Path of the backup folder to restore from")
	cmd.Flags().StringArrayP("resource", "r", []string{}, "Resource types to restore")
	cmd.Flags().Bool("latest", false, "Restore from the latest backup")
	cmd.Flags().Bool("all", false, "Restore all resources configured in the backup config")
	cmd.Flags().Bool("dry-run", false, "Print logs without creating an actual backup file")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, client *api.GQLClient, shopCfg *config.ShopConfig, ctx *config.StoreContext, logger *tlog.Logger) error {
	var strategy *config.BackupStrategy

	flag := &flag{}
	flag.parse(cmd)

	if flag.latest || flag.id != "" {
		var err error
		if strategy, err = cmdutil.GetStrategy(cmd, ctx, shopCfg); err != nil {
			return err
		}
	}

	if flag.path == "" && strategy == nil {
		return fmt.Errorf("strategy not found")
	}

	if flag.all && len(flag.resources) == 0 {
		flag.resources = strategy.Resources
	}

	var (
		wg  sync.WaitGroup
		rnr runner.Runner

		runners = make([]runner.Runner, 0, len(flag.resources))
	)

	id, path, err := getBackupIDAndPath(flag, strategy)
	if err != nil {
		return fmt.Errorf("context: %q: %w", ctx.Alias, err)
	}
	eng := engine.New(engine.NewRestore(ctx.Store))

	if flag.latest || flag.id != "" {
		logger.Infof("Using backup with id %s and name %s", id, filepath.Base(path))
	}

	bkpPath := path
	if strings.HasSuffix(path, ".tar.gz") {
		logger.V(tlog.VL1).Info("Extracting backup folder to temp location")

		tmpPath, err := registry.ExtractZipToTemp(path, id)
		if err != nil {
			return err
		}
		bkpPath = tmpPath

		logger.V(tlog.VL2).Infof("Backup folder %q was extracted to %q", path, tmpPath)
	}

	// TODO:
	// We need to maintain the order in which products, customers and orders are restored.
	// The order should always be Product -> Customer -> Order.

	toRestore := make([]string, 0, len(flag.resources))
	for _, resource := range flag.resources {
		toRestore = append(toRestore, resource.Resource)
		switch engine.ResourceType(resource.Resource) {
		case engine.Product:
			rnr = product.NewRunner(bkpPath, eng, client, logger)
		case engine.Customer:
			rnr = customer.NewRunner(bkpPath, eng, client, logger)
		default:
			logger.V(tlog.VL1).Warnf("Skipping '%s': Invalid resource", resource)
			continue
		}
		runners = append(runners, rnr)
	}

	logger.Infof("Starting restore for store: %s", ctx.Store)
	logger.Infof("Resources to restore: %s", strings.Join(toRestore, ","))

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

func getBackupIDAndPath(flag *flag, strategy *config.BackupStrategy) (string, string, error) {
	var id, path string

	switch {
	case flag.id != "":
		id = flag.id
		path, _ = registry.LookForDirWithSuffix(id, strategy.BkpDir)
	case flag.path != "":
		path = flag.path
		id = cmdutil.GetBackupIDFromName(filepath.Base(path))
		if id == "" {
			cmdutil.Fail("Error: invalid backup path")
			os.Exit(1)
		}
	case flag.latest:
		file, suffix, err := registry.GetLatestInDir(strategy.BkpDir)
		if err != nil {
			return "", "", err
		}
		id = suffix
		path = file.Path
	}

	if path == "" {
		return "", "", fmt.Errorf("unable to find backup with id %q in strategy %q", flag.id, strategy.Name)
	}
	return id, path, nil
}

func helpErrorf(msg string) error {
	lines := strings.Split(examples, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return fmt.Errorf(msg+"\n\n\033[1mUsage:\033[0m\n\n%s", strings.Join(lines, "\n"))
}
