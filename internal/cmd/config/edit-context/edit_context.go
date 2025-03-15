package editcontext

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Edit opens the context config file in the configured editor for you to edit.

Note that this will skip any validations, and you need to make sure
that the file is valid after updating.`

// NewCmdEditContext creates a new config edit command.
func NewCmdEditContext() *cobra.Command {
	return &cobra.Command{
		Use:   "edit-context",
		Short: "Edit lets you edit a context config directly",
		Long:  helpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
}

func run(cmd *cobra.Command, _ []string) error {
	editor, args := cmdutil.GetEditor()
	if editor == "" {
		return fmt.Errorf("unable to locate any editors; You can set prefered editor using `SHOPIFY_EDITOR` or `EDITOR` env")
	}

	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	ctx, err := cmdutil.GetContext(cmd, cfg)
	if err != nil {
		return err
	}
	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return err
	}

	ex := exec.Command(editor, append(args, storeCfg.Path())...)
	ex.Stdin = os.Stdin
	ex.Stdout = os.Stdout
	ex.Stderr = os.Stderr

	if err := ex.Run(); err != nil {
		return err
	}
	cmdutil.Success("Config for the context %q was updated successfully", ctx.Alias)
	return nil
}
