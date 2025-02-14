package peek

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/peek/product"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Peek lets you view data in the store.

You can quickly view data in the store, including products, customers and orders.`

// NewCmdPeek creates a new peek command.
func NewCmdPeek() *cobra.Command {
	cmd := cobra.Command{
		Use:         "peek",
		Short:       "Peek lets you view data in the store",
		Long:        helpText,
		Annotations: map[string]string{"cmd:main": "true"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(preRun(cmd, args))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}

	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")

	cmd.AddCommand(
		product.NewCmdProduct(),
	)

	return &cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	var usrCtx string

	usrCtx, err := cmd.Flags().GetString("context")
	if err != nil {
		return err
	}

	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	if usrCtx == "" {
		currCtx := cfg.CurrentContext()
		if currCtx == "" {
			return fmt.Errorf("current-context is not set; either set a context with %q or use %q flag", "shopctl use-context context-name", "-c")
		}
		usrCtx = currCtx
	}
	ctx := cfg.GetContext(usrCtx)
	if ctx == nil {
		return fmt.Errorf("no context exists with the name: %q", usrCtx)
	}

	gqlClient := api.NewGQLClient(ctx.Store)
	cmd.SetContext(context.WithValue(cmd.Context(), "gqlClient", gqlClient))
	cmd.SetContext(context.WithValue(cmd.Context(), "store", ctx.Store))

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
