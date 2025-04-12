package product

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/create"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/delete"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/list"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/media"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/option"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/peek"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/update"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product/variant"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Interact with the products data on your store.`

// NewCmdProduct builds a new product command.
func NewCmdProduct() *cobra.Command {
	cmd := cobra.Command{
		Use:         "product",
		Short:       "Interact with the products data",
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

	cmd.AddCommand(
		list.NewCmdList(),
		peek.NewCmdPeek(),
		create.NewCmdCreate(),
		update.NewCmdUpdate(),
		delete.NewCmdDelete(),
		option.NewCmdOption(),
		variant.NewCmdVariant(),
		media.NewCmdMedia(),
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

	gqlClient := api.NewGQLClient(ctx.Store)
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyContext, ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyGQLClient, gqlClient))

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
