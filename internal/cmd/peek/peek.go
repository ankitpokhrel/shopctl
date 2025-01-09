package peek

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/peek/product"
)

const helpText = `Peek lets you view data in the store

You can quickly view data in the store, including products, customers and orders.`

// NewCmdPeek creates a new peek command.
func NewCmdPeek() *cobra.Command {
	cmd := cobra.Command{
		Use:         "peek",
		Short:       "Peek lets you view data in the store",
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
		RunE: peek,
	}

	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")

	cmd.AddCommand(
		product.NewCmdProduct(),
	)

	return &cmd
}

func peek(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
