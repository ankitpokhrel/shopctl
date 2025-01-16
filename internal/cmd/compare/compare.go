package compare

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/compare/product"
)

const helpText = `Compare lets you compare data in the store with the backup

You can quickly compare data in the store with a local copy to see what changed.`

// NewCmdCompare creates a new compare command.
func NewCmdCompare() *cobra.Command {
	cmd := cobra.Command{
		Use:         "compare",
		Short:       "Compare lets you compare data in the store with the backup",
		Long:        helpText,
		Aliases:     []string{"cmp", "diff"},
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
		RunE: compare,
	}

	cmd.AddCommand(
		product.NewCmdProduct(),
	)

	return &cmd
}

func compare(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
