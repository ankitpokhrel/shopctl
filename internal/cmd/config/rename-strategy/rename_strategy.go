package renamestrategy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Rename a backup strategy for current-context.

    OLD_STRATEGY_NAME is the strategy name that you wish to change
    NEW_STRATEGY_NAME is the new name for the strategy

If the strategy being renamed is the 'current-strategy', it will get updated too.`
)

type flag struct {
	oldName string
	newName string
}

func (f *flag) parse(_ *cobra.Command, args []string) {
	f.oldName = args[0]
	f.newName = args[1]
}

// NewCmdRenameStrategy is a cmd to rename a backup strategy.
func NewCmdRenameStrategy() *cobra.Command {
	return &cobra.Command{
		Use:   "rename-strategy OLD_STRATEGY_NAME NEW_STRATEGY_NAME",
		Short: "Rename a backup strategy for current-context",
		Long:  helpText,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
}

func run(cmd *cobra.Command, args []string) error {
	flag := &flag{}
	flag.parse(cmd, args)

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	currentCtx := shopCfg.CurrentContext()
	if currentCtx == "" {
		return fmt.Errorf("current-context is not set")
	}
	ctx := shopCfg.GetContext(currentCtx)

	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return err
	}
	if !storeCfg.HasBackupStrategy(flag.oldName) {
		return fmt.Errorf("no strategy exists for context %q with the name: %q", ctx.Alias, flag.oldName)
	}
	if storeCfg.HasBackupStrategy(flag.newName) {
		return fmt.Errorf("strategy already exists for context %q with the name: %q", ctx.Alias, flag.oldName)
	}

	storeCfg.RenameStrategy(flag.oldName, flag.newName)
	if err := storeCfg.Save(); err != nil {
		return err
	}
	if flag.oldName == shopCfg.CurrentStrategy() {
		if err := shopCfg.SetCurrentStrategy(flag.newName); err != nil {
			return err
		}
		if err := shopCfg.Save(); err != nil {
			return err
		}
	}

	cmdutil.Success("Context %q renamed to %q", flag.oldName, flag.newName)
	return nil
}
