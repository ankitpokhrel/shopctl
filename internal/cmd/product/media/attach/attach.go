package attach

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Attach product media from a publicly accessible source.`

	examples = `$ shopctl product media attach 8856145494 --url "https://example.com/file.png" --alt "File attached from the CLI"
    $ shopctl product media attach 8856145494 --url "https://youtu.be/dQw4w9WgXcQ" --media-type EXTERNAL_VIDEO`
)

// Flag wraps available command flags.
type flag struct {
	id  string
	url string
	alt string
	typ schema.MediaContentType
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	id := args[0]

	url, err := cmd.Flags().GetString("url")
	cmdutil.ExitOnErr(err)

	if url == "" {
		cmdutil.ExitOnErr(cmdutil.HelpErrorf("Link to the media is required", examples))
	}

	alt, err := cmd.Flags().GetString("alt")
	cmdutil.ExitOnErr(err)

	mediaType, err := cmd.Flags().GetString("media-type")
	cmdutil.ExitOnErr(err)
	mediaType = strings.ToUpper(mediaType)

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

	f.id = cmdutil.ShopifyProductID(id)
	f.url = url
	f.alt = alt
	f.typ = schema.MediaContentType(mediaType)
}

// NewCmdAttach constructs a new product attach command.
func NewCmdAttach() *cobra.Command {
	cmd := cobra.Command{
		Use:     "attach PRODUCT_ID",
		Short:   "Attach product media",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"link"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringP("url", "l", "", "Link to a publicly accessible media")
	cmd.Flags().StringP("alt", "a", "", "Alt text for the media")
	cmd.Flags().StringP("media-type", "t", "IMAGE", "Media content type; one of: IMAGE, VIDEO, EXTERNAL_VIDEO, MODEL_3D")

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	input := schema.ProductInput{
		ID: &flag.id,
	}
	createMediaInput := schema.CreateMediaInput{
		OriginalSource:   flag.url,
		Alt:              &flag.alt,
		MediaContentType: flag.typ,
	}

	res, err := client.UpdateProduct(input, []schema.CreateMediaInput{createMediaInput})
	if err != nil {
		return err
	}

	cmdutil.Success("Media attached successfully to product: %s", res.Product.ID)
	return nil
}
