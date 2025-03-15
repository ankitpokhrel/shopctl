package backup

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/list"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/run"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Backup initiates backup process for a Shopify store.

You can either backup an entire store or a filtered subset, including products, customers and orders.`

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
		list.NewCmdList(),
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

	gqlClient := api.NewGQLClient(ctx.Store, api.LogRequest(lgr))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyShopConfig, cfg))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyContext, ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyGQLClient, gqlClient))
	cmd.SetContext(context.WithValue(cmd.Context(), cmdutil.KeyLogger, lgr))

	return nil
}

func backup(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
