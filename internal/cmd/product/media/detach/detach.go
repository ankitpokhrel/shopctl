package detach

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Detach one or more media from the product.`

	examples = `# Detach media of type IMAGE from a product
$ shopctl product media detach 8856145494 365299811616

# Detach multiple media of same type from a product
$ shopctl product media detach 8856145494 365299811616 365299811617 365299811618 -tVIDEO

# Detach multiple media of different type from a product
$ shopctl product media detach 8856145494 gid://shopify/MediaImage/365299811616 gid://shopify/Video/365299811617`
)

type flag struct {
	productID string
	mediaID   []string
	mediaType schema.MediaContentType
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	mediaType, err := cmd.Flags().GetString("media-type")
	cmdutil.ExitOnErr(err)

	validMediaTypes := []string{
		string(schema.MediaContentTypeImage),
		string(schema.MediaContentTypeVideo),
		string(schema.MediaContentTypeExternalVideo),
		string(schema.MediaContentTypeModel3d),
	}
	if mediaType != "" && !slices.Contains(validMediaTypes, mediaType) {
		cmdutil.ExitOnErr(cmdutil.HelpErrorf(
			fmt.Sprintf("Media type must be one of: %s", strings.Join(validMediaTypes, ", ")), examples),
		)
	}

	f.mediaType = schema.MediaContentType(mediaType)
	f.productID = shopctl.ShopifyProductID(args[0])

	f.mediaID = make([]string, 0, len(args[1:]))
	for _, id := range args[1:] {
		mid := shopctl.ShopifyMediaID(id, f.mediaType)
		f.mediaID = append(f.mediaID, mid)
	}
}

// NewCmdDetach constructs a new product attach command.
func NewCmdDetach() *cobra.Command {
	cmd := cobra.Command{
		Use:     "detach PRODUCT_ID MEDIA_ID...",
		Short:   "Detach one or more media from the product",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(2),
		Aliases: []string{"unlink"},
		Annotations: map[string]string{
			"help:args": `PRODUCT_ID  Shopify full or numeric Product ID, eg: 8856145494 or gid://shopify/Product/8856145494
MEDIA_ID    List of Shopify full or numeric Media IDs, eg: 365299811616 or gid://shopify/MediaImage/365299811616`,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringP("media-type", "t", "IMAGE", "Media content type; one of: IMAGE, VIDEO, EXTERNAL_VIDEO, MODEL_3D")

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	input := make([]schema.FileUpdateInput, 0, len(flag.mediaID))
	for _, id := range flag.mediaID {
		input = append(input, schema.FileUpdateInput{
			ID:                 id,
			ReferencesToRemove: []any{flag.productID},
		})
	}

	_, err := client.DetachProductMedia(input)
	if err != nil {
		return err
	}

	cmdutil.Success("Media detached successfully from product: %s", flag.productID)
	return nil
}
