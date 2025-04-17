package ingest

import (
	"context"
	"fmt"
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
	helpText = `Import imports exported Shopify data.`
	examples = `$ shopctl import --resource product --from /path/to/import/dir

# Restore some products on status DRAFT for the context from the latest backup
$ shopctl import -r product="id:id1,id2,id3 AND status:DRAFT" --from /path/to/import/dir

# Restore specific products and verified customers from the latest backup
$ shopctl import -r product="tags:premium,on-sale" -r customer="verifiedemail:true" --from /path/to/import/dir

# Restore products and customers directly from the given backup path
$ shopctl import -r product -r customer --from /path/to/import/dir

# Dry run executes the restoration process and print logs without making an actual API call
$ shopctl import -r product --from /path/to/import/dir --dry-run -vvv
`
)

var verbosity int

type flag struct {
	from      string
	resources []config.BackupResource
	dryRun    bool
	quiet     bool
}

func (f *flag) parse(cmd *cobra.Command) {
	from, err := cmd.Flags().GetString("from")
	cmdutil.ExitOnErr(err)

	resources, err := cmd.Flags().GetStringArray("resource")
	cmdutil.ExitOnErr(err)

	dryRun, err := cmd.Flags().GetBool("dry-run")
	cmdutil.ExitOnErr(err)

	quiet, err := cmd.Flags().GetBool("quiet")
	cmdutil.ExitOnErr(err)

	if len(resources) == 0 || from == "" {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Error: path to the dir to import from and resources to import is required.", examples),
		)
	}

	f.from = from
	f.resources = cmdutil.ParseBackupResource(resources)
	f.dryRun = dryRun
	f.quiet = quiet
}

// NewCmdImport creates a new import command.
func NewCmdImport() *cobra.Command {
	cmd := cobra.Command{
		Use:         "import",
		Short:       "Import Shopify resources",
		Long:        helpText,
		Example:     examples,
		Annotations: map[string]string{"cmd:main": "true"},
		Aliases:     []string{"ingest"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(preRun(cmd, args))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)
			logger := cmd.Context().Value(cmdutil.KeyLogger).(*tlog.Logger)

			cmdutil.ExitOnErr(run(cmd, client, ctx, logger))
			return nil
		},
	}
	cmd.Flags().StringP("from", "f", "", "Path of the import folder to restore from")
	cmd.Flags().StringArrayP("resource", "r", []string{}, "Resource types to restore")
	cmd.Flags().Bool("dry-run", false, "Print logs without creating an actual backup file")
	cmd.Flags().Bool("quiet", false, "Do not print anything to stdout")
	cmd.Flags().CountVarP(&verbosity, "verbose", "v", "Set the verbosity level (e.g., -v, -vv, -vvv)")

	cmd.Flags().SortFlags = false

	return &cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	ctx, err := cmdutil.GetContext(cmd, cfg)
	if err != nil {
		return err
	}

	v, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		return err
	}

	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}
	lgr := tlog.New(tlog.VerboseLevel(v), quiet)

	gqlClient := api.NewGQLClient(ctx.Store)
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyContext, ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyGQLClient, gqlClient))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyLogger, lgr))

	return nil
}

func run(cmd *cobra.Command, client *api.GQLClient, ctx *config.StoreContext, logger *tlog.Logger) error {
	flag := &flag{}
	flag.parse(cmd)

	var (
		wg      sync.WaitGroup
		rnr     runner.Runner
		counter int

		runners = make([]runner.Runner, 0, len(flag.resources))
	)

	eng := engine.New(engine.NewRestore(ctx.Store))

	dirPath := flag.from
	if strings.HasSuffix(flag.from, ".tar.gz") {
		logger.V(tlog.VL1).Info("Extracting backup folder to temp location")

		tmpPath, err := registry.ExtractZipToTemp(flag.from, cmdutil.GetBackupIDFromName(filepath.Base(flag.from)))
		if err != nil {
			return err
		}
		dirPath = tmpPath

		logger.V(tlog.VL2).Infof("Export folder %q was extracted to %q", flag.from, tmpPath)
	}

	// TODO:
	// We need to maintain the order in which products, customers and orders are restored.
	// The order should always be Product -> Customer -> Order.

	toRestore := make([]string, 0, len(flag.resources))
	for _, resource := range flag.resources {
		var filters runner.RestoreFilter
		if resource.Query != "" {
			conditions, separators, err := cmdutil.ParseRestoreFilters(resource.Query)
			if err != nil {
				return err
			}
			filters.Filters = conditions
			filters.Separators = separators
		}

		toRestore = append(toRestore, resource.Resource)
		switch engine.ResourceType(resource.Resource) {
		case engine.Product:
			rnr = product.NewRunner(dirPath, eng, client, logger, &filters, flag.dryRun)
		case engine.Customer:
			rnr = customer.NewRunner(dirPath, eng, client, logger, &filters, flag.dryRun)
		default:
			logger.V(tlog.VL1).Warnf("Skipping '%s': Invalid resource", resource)
			continue
		}
		runners = append(runners, rnr)
	}

	if flag.dryRun {
		logger.Warn("This is a dry run. CRUD mutations won't be executed.")
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

	for _, rnr := range runners {
		stats := rnr.Stats()
		counter += stats[rnr.Kind()].Count
	}

	if !flag.quiet && counter > 0 {
		summarize(ctx.Store, flag.from, runners)
	} else if counter == 0 {
		logger.Info("No matching records found for the given criteria")
	}
	return nil
}

func summarize(store string, bkpPath string, runners []runner.Runner) {
	resources := make([]string, 0, len(runners))
	for _, rnr := range runners {
		resources = append(resources, string(rnr.Kind()))
	}

	fmt.Println()
	cmdutil.SummaryTitle("RESTORE SUMMARY", cmdutil.RepeatedEquals)
	fmt.Printf(`Store: %s
Path used: %s
Resources: %s
`,
		store, bkpPath,
		strings.Join(resources, ","),
	)

	for _, rnr := range runners {
		fmt.Println()
		stats := rnr.Stats()
		for _, rt := range engine.GetAllResourceTypes() {
			// ProductOption count is not accurate as
			// we are handling them separately now.
			// TODO: Better summary implementation.
			if rt == engine.ProductOption {
				continue
			}
			st, ok := stats[rt]
			if !ok {
				continue
			}
			if rt.IsPrimary() {
				cmdutil.SummaryTitle(strings.ToTitle(string(rt)), cmdutil.RepeatedDashes)
			} else {
				cmdutil.SummarySubtitle(strings.ToTitle(string(rt)), cmdutil.RepeatedDashesSM)
			}
			fmt.Println(st.String())
			fmt.Println()
		}
	}
}
