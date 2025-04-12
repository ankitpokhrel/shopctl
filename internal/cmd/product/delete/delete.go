package delete

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

const (
	helpText = `Delete lets you delete a product.`

	examples = `$ shopctl product delete 123456789
$ shopctl product delete gid://shopify/Product/123456789`
)

// NewCmdDelete constructs a new product create command.
func NewCmdDelete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete PRODUCT_ID",
		Short:   "Delete a product",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"del", "rm", "remove"},
		Annotations: map[string]string{
			"help:args": `PRODUCT_ID full or numeric Product ID, eg: 88561444456 or gid://shopify/Product/88561444456`,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
}

func run(_ *cobra.Command, args []string, client *api.GQLClient) error {
	productID := shopctl.ShopifyProductID(args[0])

	_, err := client.DeleteProduct(productID)
	if err != nil {
		return err
	}

	cmdutil.Success("Product deleted successfully")
	return nil
}
