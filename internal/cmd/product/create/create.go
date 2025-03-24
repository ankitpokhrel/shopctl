package create

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Create lets you create a product.`

	examples = `$ shopctl product create --title "Product title"

# Create active product in the current context
$ shopctl product create -tTitle -d"Product description" --status active

# Create active product with tags in the current context
$ shopctl product create -tTitle -d"Product description" --tags tag1,tag2

# Create product in another store
$ shopctl product create -c store2 -tTitle -d"Product description" --type Bags`
)

// Flag wraps available command flags.
type flag struct {
	handle     string
	title      string
	descHtml   string
	typ        string
	category   string
	tags       []string
	vendor     string
	status     string
	isGiftCard bool
	web        bool
}

func (f *flag) parse(cmd *cobra.Command, _ []string) {
	handle, err := cmd.Flags().GetString("handle")
	cmdutil.ExitOnErr(err)

	title, err := cmd.Flags().GetString("title")
	cmdutil.ExitOnErr(err)

	desc, err := cmd.Flags().GetString("desc")
	cmdutil.ExitOnErr(err)

	typ, err := cmd.Flags().GetString("type")
	cmdutil.ExitOnErr(err)

	category, err := cmd.Flags().GetString("category")
	cmdutil.ExitOnErr(err)

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	vendor, err := cmd.Flags().GetString("vendor")
	cmdutil.ExitOnErr(err)

	status, err := cmd.Flags().GetString("status")
	cmdutil.ExitOnErr(err)

	isGiftCard, err := cmd.Flags().GetBool("gift-card")
	cmdutil.ExitOnErr(err)

	web, err := cmd.Flags().GetBool("web")
	cmdutil.ExitOnErr(err)

	f.handle = handle
	f.title = title
	f.descHtml = desc
	f.typ = typ
	f.category = category
	f.tags = strings.Split(tags, ",")
	f.vendor = vendor
	f.status = status
	f.isGiftCard = isGiftCard
	f.web = web
}

// NewCmdCreate constructs a new product create command.
func NewCmdCreate() *cobra.Command {
	cmd := cobra.Command{
		Use:     "create",
		Short:   "Create lets you create a product",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"add"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}
	cmd.Flags().String("handle", "", "Product handle")
	cmd.Flags().StringP("title", "t", "", "Product title")
	cmd.Flags().StringP("desc", "d", "", "Product description")
	cmd.Flags().String("type", "", "Product type")
	cmd.Flags().StringP("category", "y", "", "Product category id")
	cmd.Flags().String("tags", "", "Comma separated product tags")
	cmd.Flags().String("vendor", "", "Product vendor")
	cmd.Flags().String("status", string(schema.ProductStatusDraft), "Product status (ACTIVE, ARCHIVED, DRAFT)")
	cmd.Flags().Bool("gift-card", false, "Is gift card?")
	cmd.Flags().Bool("web", false, "Open in web browser after successful creation")

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	tags := make([]any, len(flag.tags))
	for _, t := range flag.tags {
		tags = append(tags, t)
	}
	status := schema.ProductStatusDraft
	if flag.status != "" {
		status = schema.ProductStatus(strings.ToTitle(flag.status))
	}
	var category *string
	if flag.category != "" {
		category = &flag.category
	}

	input := schema.ProductInput{
		Handle:          &flag.handle,
		Title:           &flag.title,
		DescriptionHtml: &flag.descHtml,
		ProductType:     &flag.typ,
		Category:        category,
		Tags:            tags,
		Vendor:          &flag.vendor,
		Status:          &status,
		GiftCard:        &flag.isGiftCard,
	}

	res, err := client.CreateProduct(input)
	if err != nil {
		return err
	}

	adminURL := fmt.Sprintf(
		"https://admin.shopify.com/store/%s/products/%s",
		ctx.Alias, cmdutil.ExtractNumericID(res.Product.ID),
	)
	if flag.web {
		_ = browser.Browse(adminURL)
	}

	cmdutil.Success("Product created successfully: %s", res.Product.Handle)
	fmt.Println(adminURL)

	return nil
}
