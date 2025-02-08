package backup

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/run"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Backup initiates backup process for a Shopify store.

You can either backup an entire store or a filtered subset, including products, customers and orders.

Supports advanced options for incremental backups and output customization.`

// NewCmdBackup creates a new backup command.
func NewCmdBackup() *cobra.Command {
	var (
		store string
		err   error
	)

	cmd := cobra.Command{
		Use:         "backup",
		Short:       "Backup initiates backup process for a shopify store",
		Long:        helpText,
		Aliases:     []string{"bkp", "dump"},
		Annotations: map[string]string{"cmd:main": "true"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().Name() != "backup" {
				return nil
			}
			if cmd.Name() == "config" {
				return nil
			}

			store, err = cmd.Flags().GetString("store")
			if err != nil {
				return err
			}

			v, _ := cmd.Flags().GetCount("verbose")
			lgr := tlog.New(tlog.VerboseLevel(v))

			gqlClient := api.NewGQLClient(store, api.LogRequest(lgr))
			cmd.SetContext(context.WithValue(cmd.Context(), "gqlClient", gqlClient))
			cmd.SetContext(context.WithValue(cmd.Context(), "logger", lgr))

			return nil
		},
		RunE: backup,
	}

	cmd.AddCommand(
		run.NewCmdRun(),
	)

	return &cmd
}

func backup(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
