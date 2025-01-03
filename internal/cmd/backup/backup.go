package backup

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/product"
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
			store, err := cmd.Flags().GetString("store")
			if err != nil {
				return err
			}

			gqlClient := api.NewGQLClient(store)
			cmd.SetContext(context.WithValue(cmd.Context(), "gqlClient", gqlClient))

			return nil
		},
		RunE: backup,
	}

	bkpEngine := engine.New(engine.NewBackup())
	cmd.AddCommand(
		product.NewCmdProduct(bkpEngine),
	)

	return &cmd
}

func backup(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
