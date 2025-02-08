package usecontext

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Set current-context in the shopconfig file.`

// NewCmdUseContext is a cmd to update current context.
func NewCmdUseContext() *cobra.Command {
	return &cobra.Command{
		Use:     "use-context CONTEXT_NAME",
		Short:   "Set current-context in the shopconfig file",
		Long:    helpText,
		Aliases: []string{"use"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
}

func run(_ *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	ctx := args[0]
	if len(ctx) == 0 {
		return fmt.Errorf("empty context names are not allowed")
	}

	if err := cfg.SetCurrentContext(ctx); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	cmdutil.Success("Switched to context %q", ctx)
	return nil
}
