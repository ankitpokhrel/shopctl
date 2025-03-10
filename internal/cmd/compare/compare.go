package compare

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/compare/product"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Compare lets you compare data in the store with the backup

You can quickly compare data in the store with a local copy to see what changed.`

// NewCmdCompare creates a new compare command.
func NewCmdCompare() *cobra.Command {
	cmd := cobra.Command{
		Use:         "compare",
		Short:       "Compare lets you compare data in the store with the backup",
		Long:        helpText,
		Aliases:     []string{"cmp", "diff"},
		Annotations: map[string]string{"cmd:main": "true"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(preRun(cmd, args))
			return nil
		},
		RunE: run,
	}

	cmd.PersistentFlags().StringP(
		"strategy", "s", "",
		"Override current-strategy",
	)

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
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyGQLClient, gqlClient))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyStrategy, strategy))

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
