package setstrategy

import (
	"fmt"
	"strings"

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
	kind      string
	bkpDir    string
	bkpPrefix string
	resources []string
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	name := args[0]

	kind, err := cmd.Flags().GetString("type")
	cmdutil.ExitOnErr(err)

	if kind == "" {
		kind = string(engine.BackupTypeIncremental)
	}

	dir, err := cmd.Flags().GetString("dir")
	cmdutil.ExitOnErr(err)

	prefix, err := cmd.Flags().GetString("prefix")
	cmdutil.ExitOnErr(err)

	resources, err := cmd.Flags().GetString("resources")
	cmdutil.ExitOnErr(err)

	if dir == "" || resources == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: backup directory and resources to backup are required.

Usage:
  $ shopctl config set-strategy daily -d /path/to/bkp-dir -r=product,customer

See 'shopctl config set-strategy --help' for more info.`),
		)
	}

	f.name = name
	f.kind = kind
	f.bkpDir = dir
	f.bkpPrefix = prefix
	f.resources = strings.Split(resources, ",")
}

// NewCmdSetStrategy is a cmd to add/update a backup strategy.
func NewCmdSetStrategy() *cobra.Command {
	cmd := cobra.Command{
		Use:   "set-strategy STRATEGY_NAME",
		Short: "Add/update a backup strategy",
		Long:  helpText,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
	cmd.Flags().StringP("dir", "d", "", "Root directory to save backups to")
	cmd.Flags().StringP("prefix", "p", "", "Prefix for the main backup directory")
	cmd.Flags().StringP("resources", "r", "", "Resource types to backup (comma separated)")
	cmd.Flags().StringP("type", "t", "", "Backup type (full or incremental)")

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
	ctx := shopCfg.GetContext(shopCfg.CurrentContext())
	if ctx == nil {
		return fmt.Errorf("current-context is not set")
	}

	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return err
	}

	storeCfg.SetStoreBackupStrategy(&config.BackupStrategy{
		Name:      flag.name,
		Kind:      flag.kind,
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
