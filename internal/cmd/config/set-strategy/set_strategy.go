package setstrategy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/engine"
)

const (
	helpText = `Add/update a backup strategy for the current-context.

You can create multiple backup strategies for a store. Each strategy must have a unique name. 
If a strategy with the name already exists, then it will be updated instead.`

	example = `# Daily product backup
$ shopctl config set-strategy daily -t full -d "/path/to/backups/daily" -r product

# Add a strategy for full weekly backup
$ shopctl config set-strategy weekly -t full -d "/path/to/backups/weekly" -r product,customer`
)

type flag struct {
	name      string
	bkpDir    string
	bkpPrefix string
	resources []config.BackupResource
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	name := args[0]

	dir, err := cmd.Flags().GetString("dir")
	cmdutil.ExitOnErr(err)

	prefix, err := cmd.Flags().GetString("prefix")
	cmdutil.ExitOnErr(err)

	resources, err := cmd.Flags().GetStringArray("resources")
	cmdutil.ExitOnErr(err)

	usage := `Usage:
  $ shopctl config set-strategy daily -d /path/to/bkp-dir -r product=tag:premium -r customer

See 'shopctl config set-strategy --help' for more info.`

	if dir == "" || len(resources) == 0 {
		cmdutil.ExitOnErr(
			fmt.Errorf("error: backup directory and resources to backup are required.\n\n%s", usage),
		)
	}

	f.name = name
	f.bkpDir = dir
	f.bkpPrefix = prefix
	f.resources = cmdutil.ParseBackupResource(resources)
}

// NewCmdSetStrategy is a cmd to add/update a backup strategy.
func NewCmdSetStrategy() *cobra.Command {
	cmd := cobra.Command{
		Use:     "set-strategy STRATEGY_NAME",
		Short:   "Add/update a backup strategy",
		Long:    helpText,
		Example: example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
	cmd.Flags().StringP("dir", "d", "", "Root directory to save backups to")
	cmd.Flags().StringP("prefix", "p", "", "Prefix for the main backup directory")
	cmd.Flags().StringArrayP("resources", "r", []string{}, "Resource types to backup (format: resourcetype=filter)")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string) error {
	flag := &flag{}
	flag.parse(cmd, args)

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	ctx, err := cmdutil.GetContext(cmd, shopCfg)
	if err != nil {
		return err
	}
	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return err
	}

	storeCfg.SetStoreBackupStrategy(&config.BackupStrategy{
		Name:      flag.name,
		Kind:      string(engine.BackupTypeFull),
		BkpDir:    flag.bkpDir,
		BkpPrefix: flag.bkpPrefix,
		Resources: flag.resources,
	})

	if err := storeCfg.Save(); err != nil {
		return err
	}
	cmdutil.Success("Strategy updated successfully")
	return nil
}
