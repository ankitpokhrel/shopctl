package renamecontext

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Rename a context from the shopconfig file.

    OLD_CONTEXT_NAME is the context name that you wish to change
    NEW_CONTEXT_NAME is the new name for the context

If the context being renamed is the 'current-context', it will get updated too.`
)

type flag struct {
	oldName string
	newName string
}

func (f *flag) parse(_ *cobra.Command, args []string) {
	f.oldName = args[0]
	f.newName = args[1]
}

// NewCmdRenameContext is a cmd to rename a context.
func NewCmdRenameContext() *cobra.Command {
	return &cobra.Command{
		Use:   "rename-context OLD_CONTEXT_NAME NEW_CONTEXT_NAME",
		Short: "Rename a context from the shopconfig file",
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

	if !shopCfg.HasContext(flag.oldName) {
		return fmt.Errorf("no context exists with the name: %q", flag.oldName)
	}
	if shopCfg.HasContext(flag.newName) {
		return fmt.Errorf("context already exists with the name: %q", flag.newName)
	}
	ctx := shopCfg.GetContext(flag.oldName)

	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return err
	}

	shopCfg.RenameContext(flag.oldName, flag.newName)
	if shopCfg.CurrentContext() == flag.oldName {
		shopCfg.SetCurrentContext(flag.newName)
	}
	if err := shopCfg.Save(); err != nil {
		return err
	}
	if err := storeCfg.Rename(flag.newName); err != nil {
		return err
	}

	cmdutil.Success("Context %q renamed to %q", flag.oldName, flag.newName)
	return nil
}
