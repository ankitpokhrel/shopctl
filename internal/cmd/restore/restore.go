package restore

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/restore/product"
)

const helpText = `Restore initiates data restoration process.

You can either restore an entire store or a filtered subset, including products, customers and orders.`

// NewCmdRestore creates a new restore command.
func NewCmdRestore() *cobra.Command {
	cmd := cobra.Command{
		Use:         "restore",
		Short:       "Restore initiates a data restoration process",
		Long:        helpText,
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
		RunE: restore,
	}

	rstEngine := engine.New(engine.NewRestore())
	cmd.AddCommand(
		product.NewCmdProduct(rstEngine),
	)

	return &cmd
}

func restore(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
