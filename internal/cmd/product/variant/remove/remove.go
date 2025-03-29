package remove

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

const (
	helpText = `Remove product variant by its id or title.`

	examples = `$ shopctl product variant remove 8856145494 "Red / XS"

# Accepts multiple variant IDs and/or title
$ shopctl product variant remove 8856145494 471883718 "Black / XL"`
)

// NewCmdRemove constructs a new product option remove command.
func NewCmdRemove() *cobra.Command {
	return &cobra.Command{
		Use:     "remove PRODUCT_ID",
		Short:   "Remove product variant",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(2),
		Aliases: []string{"delete", "del", "rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
}

func run(_ *cobra.Command, args []string, client *api.GQLClient) error {
	productID := cmdutil.ShopifyProductID(args[0])
	variantIDs := args[1:]

	variantsToDelete := make([]string, 0, len(variantIDs))
	variantsSkipped := make([]string, 0)
	for _, id := range variantIDs {
		vid := cmdutil.ShopifyProductVariantID(id)
		if vid == "" {
			parts := strings.Split(id, "/")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			title := strings.Join(parts, " / ")

			variant, err := client.GetProductVariantByTitle(productID, title, false)
			if err != nil {
				variantsSkipped = append(variantsSkipped, id)
			} else {
				variantsToDelete = append(variantsToDelete, variant.ID)
			}
		} else {
			variant, err := client.CheckProductVariantByID(vid)
			if err != nil {
				variantsSkipped = append(variantsSkipped, vid)
			} else {
				variantsToDelete = append(variantsToDelete, variant.ID)
			}
		}
	}

	if len(variantsSkipped) > 0 {
		cmdutil.Warn("Some varaints were skipped: %s", strings.Join(variantsSkipped, ", "))
	}
	if len(variantsToDelete) == 0 {
		cmdutil.Warn("Nothing to delete")
		return nil
	}

	res, err := client.DeleteProductVariants(productID, variantsToDelete)
	if err != nil {
		return err
	}
	cmdutil.Success(
		"Variants %q deleted successfully for product: %s", strings.Join(variantsToDelete, ", "), res.Product.ID,
	)
	return nil
}
