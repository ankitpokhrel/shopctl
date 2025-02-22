package backup

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/run"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Backup initiates backup process for a Shopify store.

You can either backup an entire store or a filtered subset, including products, customers and orders.

Supports advanced options for incremental backups and output customization.`

// NewCmdBackup creates a new backup command.
func NewCmdBackup() *cobra.Command {
	cmd := cobra.Command{
		Use:         "backup",
		Short:       "Backup initiates backup process for a shopify store",
		Long:        helpText,
		Aliases:     []string{"bkp", "dump"},
		Annotations: map[string]string{"cmd:main": "true"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(preRun(cmd, args))
			return nil
		},
		RunE: backup,
	}

	cmd.PersistentFlags().StringP(
		"strategy", "s", "",
		"Override current-strategy",
	)
	cmd.PersistentFlags().Bool(
		"quiet", false,
		"Do not print anything to stdout",
	)

	cmd.AddCommand(
		run.NewCmdRun(),
	)

	return &cmd
}

func preRun(cmd *cobra.Command, _ []string) error {
	v, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		return err
	}

	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}
	lgr := tlog.New(tlog.VerboseLevel(v), quiet)

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
		// Let's pass empty strategy if we fail to retrieve strategy.
		// This value will be checked and properly set downstream.
		strategy = nil
	}

	gqlClient := api.NewGQLClient(ctx.Store, api.LogRequest(lgr))
	cmd.SetContext(context.WithValue(cmd.Context(), "context", ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), "strategy", strategy))
	cmd.SetContext(context.WithValue(cmd.Context(), "gqlClient", gqlClient))
	cmd.SetContext(context.WithValue(cmd.Context(), "logger", lgr))

	return nil
}

func backup(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
