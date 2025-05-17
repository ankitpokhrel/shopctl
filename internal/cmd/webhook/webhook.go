package webhook

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/webhook/delete"
	"github.com/ankitpokhrel/shopctl/internal/cmd/webhook/list"
	"github.com/ankitpokhrel/shopctl/internal/cmd/webhook/listen"
	"github.com/ankitpokhrel/shopctl/internal/cmd/webhook/register"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

// NewCmdWebhook responds to the Shopify webhooks.
func NewCmdWebhook() *cobra.Command {
	cmd := cobra.Command{
		Use:         "webhook",
		Short:       "Interact with the Shopify webhooks",
		Long:        "Interact with the webhook topics provided by the Shopify.",
		Aliases:     []string{"wh", "hook", "event"},
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
		register.NewCmdRegister(),
		listen.NewCmdListen(),
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
