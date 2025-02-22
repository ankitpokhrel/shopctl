package run

import (
	"fmt"
	"os"
	"os/user"
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
	examples = `# Run backup for configured context and strategy
$ shopctl backup run

# Specify custom context and strategy
$ shopctl backup run -c mycontext -s mystrategy

# Run adhoc backup for all products and customers
$ shopctl backup run -r product -r customer -o /path/to/bkp

# Backup all products for current context and save as mybkp in the given path
$ shopctl backup run -r product -o /path/to/bkp -n mybkp

# Backup premium on-sale products and customers created starting 2025
$ shopctl backup run -c mycontext -r product="tag:on-sale AND tag:premium" -r customer=created_at:>=2025-01-01 -o /path/to/bkp

# Dry run executes the backup without creating final backup files. This will still save files in temporary location.
# Use this option if you want to verify your config without the risk of saving data to the unintented location.
$ shopctl backup run --dry-run
$ shopctl backup run --dry-run -vvv
`

	repeatedDashes = "" +
		"-------------------------------"
	repeatedEquals = "" +
		"==============================="
)

type flag struct {
	outDir    string
	name      string
	resources []config.BackupResource
	dryRun    bool
	quiet     bool
}

func (f *flag) parse(cmd *cobra.Command) {
	dir, err := cmd.Flags().GetString("output-dir")
	cmdutil.ExitOnErr(err)

	name, err := cmd.Flags().GetString("name")
	cmdutil.ExitOnErr(err)

	resources, err := cmd.Flags().GetStringArray("resource")
	cmdutil.ExitOnErr(err)

	if len(resources) > 0 && dir == "" {
		cmdutil.ExitOnErr(helpErrorf("Error: backup directory is required for adhoc run."))
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	cmdutil.ExitOnErr(err)

	quiet, err := cmd.Flags().GetBool("quiet")
	cmdutil.ExitOnErr(err)

	f.outDir = dir
	f.name = name
	f.resources = cmdutil.ParseBackupResource(resources)
	f.dryRun = dryRun
	f.quiet = quiet
}

// NewCmdRun creates a new run command.
func NewCmdRun() *cobra.Command {
	cmd := cobra.Command{
		Use:     "run",
		Short:   "Run starts a backup process",
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

	cmd.Flags().StringP("output-dir", "o", "", "Root output directory to save backup to")
	cmd.Flags().StringP("name", "n", "", "Backup dir name")
	cmd.Flags().StringArrayP("resource", "r", []string{}, "Resources to run backup for")
	cmd.Flags().Bool("dry-run", false, "Print logs without creating an actual backup file")

	return &cmd
}

//nolint:gocyclo
func run(cmd *cobra.Command, client *api.GQLClient, shopCfg *config.ShopConfig, ctx *config.StoreContext, logger *tlog.Logger) error {
	flag := &flag{}
	flag.parse(cmd)

	isAdhocRun := flag.outDir != "" || len(flag.resources) > 0

	strategy, _ := cmdutil.GetStrategy(cmd, ctx, shopCfg)
	if strategy == nil {
		strategy = &config.BackupStrategy{
			Name:      "adhoc",
			BkpPrefix: "adhoc",
		}
	}
	if flag.outDir != "" {
		strategy.BkpDir = flag.outDir
	}
	if len(flag.resources) > 0 {
		strategy.Resources = flag.resources
	}

	if len(strategy.Resources) == 0 {
		return helpErrorf("Error: you must define resources to backup for adhoc run")
	}

	bkpEng := engine.NewBackup(
		ctx.Store,
		engine.WithBackupDir(flag.name),
		engine.WithBackupPrefix(strategy.BkpPrefix),
	)
	eng := engine.New(bkpEng)

	meta, err := saveRootMeta(bkpEng, strategy)
	if err != nil {
		cmdutil.Fail("Error: unable to create backup files; make sure that the location is writable by the user")
		os.Exit(1)
	}

	var (
		wg      sync.WaitGroup
		rnr     runner.Runner
		start   time.Time
		counter int

		runners = make([]runner.Runner, 0, len(strategy.Resources))
	)

	defer func() {
		err := meta.Set(map[string]any{
			config.KeyStatus:  string(engine.BackupStatusComplete),
			config.KeyTimeEnd: time.Now().Unix(),
		})
		if err != nil {
			logger.Errorf("Error: unable to update metadata after backup run: %s", err.Error())
		}

		if !flag.dryRun && counter > 0 {
			err := archive(bkpEng.Root(), strategy.BkpDir, bkpEng.Dir())
			if err != nil {
				logger.Errorf("Error: unable to archive: %s", err.Error())
			}
		}
		logger.Infof("Backup complete in %s", time.Since(start))

		if counter > 0 {
			if isAdhocRun {
				summarizeAdhoc(strategy, bkpEng, runners)
			} else {
				summarize(strategy, bkpEng, runners)
			}
		}
	}()

	for _, resource := range strategy.Resources {
		switch engine.ResourceType(resource.Resource) {
		case engine.Product:
			rnr = product.NewRunner(eng, client, resource.Query, logger)
		case engine.Customer:
			rnr = customer.NewRunner(eng, client, logger)
		default:
			logger.Warnf("Skipping '%s': invalid resource", resource)
			continue
		}
		runners = append(runners, rnr)
	}

	if flag.dryRun {
		logger.Warn("This is a dry run. API calls will be made but final backup files won't be created.")
	}
	logger.V(tlog.VL1).Infof("Using context %q", ctx.Alias)
	logger.V(tlog.VL1).Infof("Using store %q", ctx.Store)
	if strategy != nil {
		logger.V(tlog.VL1).Infof("Using strategy %q", strategy.Name)
	}

	start = time.Now()
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
	logger.V(tlog.VL1).Infof("We're done with fetching data. Processing results...")

	if flag.quiet {
		return nil
	}

	for _, rnr := range runners {
		stats := rnr.Stats()
		counter += stats.Count
	}

	if counter == 0 {
		logger.Info("No matching records found for the given criteria")
		logger.Warn("Backup was not created since no records were found")
	}
	return nil
}

func archive(from string, to string, name string) error {
	const modDir = 0o755
	if err := os.MkdirAll(to, modDir); err != nil {
		return err
	}
	return cmdutil.Archive(from, to, name)
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

	fmt.Println()
	title("BACKUP SUMMARY", repeatedEquals)
	fmt.Printf(`ID: %s
Strategy: %s
Store: %s
Type: %s
Path: %s
File: %s.tar.gz
`,
		bkpEng.ID(), strategy.Name, bkpEng.Store(),
		strategy.Kind, strategy.BkpDir, bkpEng.Dir(),
	)
	for _, rnr := range runners {
		fmt.Println()
		title(strings.ToTitle(string(rnr.Kind())), repeatedDashes)
		fmt.Println(rnr.Stats())
	}
}

func summarizeAdhoc(strategy *config.BackupStrategy, bkpEng *engine.Backup, runners []runner.Runner) {
	title := func(msg string, sep string) {
		fmt.Printf("%s\n%s\n%s\n", sep, msg, sep)
	}

	fmt.Println()
	title("BACKUP SUMMARY", repeatedEquals)
	fmt.Printf(`ID: %s
Store: %s
Path: %s
File: %s.tar.gz
`,
		bkpEng.ID(), bkpEng.Store(),
		strategy.BkpDir, bkpEng.Dir(),
	)
	for _, rnr := range runners {
		fmt.Println()
		title(strings.ToTitle(string(rnr.Kind())), repeatedDashes)
		fmt.Println(rnr.Stats())
	}
}

func helpErrorf(msg string) error {
	lines := strings.Split(examples, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return fmt.Errorf(msg+"\n\n\033[1mUsage:\033[0m\n\n%s", strings.Join(lines, "\n"))
}
