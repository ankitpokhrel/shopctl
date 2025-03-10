package peek

import (
	"context"

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

	cmd.PersistentFlags().StringP(
		"strategy", "s", "",
		"Override current-strategy",
	)
	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")

	cmd.AddCommand(
		product.NewCmdProduct(),
	)

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

	strategy, err := cmdutil.GetStrategy(cmd, ctx, cfg)
	if err != nil {
		return err
	}

	gqlClient := api.NewGQLClient(ctx.Store)
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyContext, ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyStrategy, strategy))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyGQLClient, gqlClient))

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
