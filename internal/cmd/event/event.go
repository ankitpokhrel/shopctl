package event

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/event/list"
	"github.com/ankitpokhrel/shopctl/internal/cmd/event/listen"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

// NewCmdEvent responds to the Shopify webhooks.
func NewCmdEvent() *cobra.Command {
	cmd := cobra.Command{
		Use:         "event",
		Short:       "Interact with the Shopify webhooks",
		Long:        "Interact with the webhook topics provided by the Shopify.",
		Aliases:     []string{"ev", "events", "webhook", "webhooks", "wh"},
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
		listen.NewCmdListen(),
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
