package customer

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/customer/create"
	"github.com/ankitpokhrel/shopctl/internal/cmd/customer/delete"
	"github.com/ankitpokhrel/shopctl/internal/cmd/customer/list"
	"github.com/ankitpokhrel/shopctl/internal/cmd/customer/update"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Interact with the customers data on your store.`

// NewCmdCustomer builds a new product command.
func NewCmdCustomer() *cobra.Command {
	cmd := cobra.Command{
		Use:         "customer",
		Short:       "Interact with the customer data",
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
		create.NewCmdCreate(),
		update.NewCmdUpdate(),
		delete.NewCmdDelete(),
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

	gqlClient := api.NewGQLClient(ctx)
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyContext, ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyGQLClient, gqlClient))

	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
