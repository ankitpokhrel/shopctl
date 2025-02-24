package restore

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/restore/run"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Restore initiates a data restoration process.

You can either restore an entire store or a filtered subset, including products, customers and orders.`

// NewCmdRestore creates a new restore command.
func NewCmdRestore() *cobra.Command {
	cmd := cobra.Command{
		Use:         "restore",
		Short:       "Restore initiates a data restoration process",
		Long:        helpText,
		Annotations: map[string]string{"cmd:main": "true"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(preRun(cmd, args))
			return nil
		},
		RunE: restore,
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
	if cmd.Parent().Name() != "restore" {
		return nil
	}

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
	cmd.SetContext(context.WithValue(cmd.Context(), "shopCfg", cfg))
	cmd.SetContext(context.WithValue(cmd.Context(), "context", ctx))
	cmd.SetContext(context.WithValue(cmd.Context(), "gqlClient", gqlClient))
	cmd.SetContext(context.WithValue(cmd.Context(), "logger", lgr))

	return nil
}

func restore(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
