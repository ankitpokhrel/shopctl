package update

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Update lets you update a product.`

	examples = `$ shopctl product update --title "Product title"

# Update and activate a product
$ shopctl product update 8856145494 -tTitle -d"Product description" --status active

# Update product to add tag1 and remove tag2
$ shopctl product update 8856145494 -tTitle -d"Product description" --tags tag1,-tag2

# Update product in another store
$ shopctl product update 8856145494 -c store2 -tTitle -d"Product description" --type Bags`
)

type flag struct {
	id       string
	handle   string
	title    string
	descHtml *string
	typ      *string
	category string
	tags     []string
	vendor   string
	status   string
	seoTitle *string
	seoDesc  *string
	web      bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	isset := func(item string) bool {
		fl := cmd.Flags().Lookup(item)
		return fl != nil && fl.Changed
	}

	handle, err := cmd.Flags().GetString("handle")
	cmdutil.ExitOnErr(err)

	title, err := cmd.Flags().GetString("title")
	cmdutil.ExitOnErr(err)

	if isset("desc") {
		desc, err := cmd.Flags().GetString("desc")
		cmdutil.ExitOnErr(err)

		f.descHtml = &desc
	}

	if isset("type") {
		typ, err := cmd.Flags().GetString("type")
		cmdutil.ExitOnErr(err)

		f.typ = &typ
	}

	category, err := cmd.Flags().GetString("category")
	cmdutil.ExitOnErr(err)

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	vendor, err := cmd.Flags().GetString("vendor")
	cmdutil.ExitOnErr(err)

	status, err := cmd.Flags().GetString("status")
	cmdutil.ExitOnErr(err)

	if isset("seo-title") {
		seoTitle, err := cmd.Flags().GetString("seo-title")
		cmdutil.ExitOnErr(err)

		f.seoTitle = &seoTitle
	}

	if isset("seo-desc") {
		seoDesc, err := cmd.Flags().GetString("seo-desc")
		cmdutil.ExitOnErr(err)

		f.seoDesc = &seoDesc
	}

	web, err := cmd.Flags().GetBool("web")
	cmdutil.ExitOnErr(err)

	f.id = shopctl.ShopifyProductID(args[0])
	f.handle = handle
	f.title = title
	f.category = category
	f.tags = strings.Split(tags, ",")
	f.vendor = vendor
	f.status = status
	f.web = web
}

// NewCmdUpdate constructs a new product update command.
func NewCmdUpdate() *cobra.Command {
	cmd := cobra.Command{
		Use:     "update PRODUCT_ID",
		Short:   "Update a product",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"edit"},
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
	cmd.Flags().String("tags", "", "Comma separated list of product tags")
	cmd.Flags().String("vendor", "", "Product vendor")
	cmd.Flags().String("seo-title", "", "SEO title of the product")
	cmd.Flags().String("seo-desc", "", "SEO description of the product")
	cmd.Flags().String("status", "", "Product status (ACTIVE, ARCHIVED, DRAFT)")
	cmd.Flags().Bool("web", false, "Open in web browser after successful update")

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	if !hasAnythingToUpdate(cmd) {
		cmdutil.Warn("Nothing to update")
		os.Exit(0)
	}

	product, err := client.GetProductByID(flag.id)
	if err != nil {
		return fmt.Errorf("product not found")
	}
	input := getInput(*flag, product)

	res, err := client.UpdateProduct(*input, nil)
	if err != nil {
		return err
	}

	adminURL := fmt.Sprintf(
		"https://admin.shopify.com/store/%s/products/%s",
		ctx.Alias, shopctl.ExtractNumericID(res.Product.ID),
	)
	if flag.web {
		_ = browser.Browse(adminURL)
	}

	cmdutil.Success("Product updated successfully: %s", res.Product.Handle)
	fmt.Println(adminURL)

	return nil
}

func getInput(f flag, p *schema.Product) *schema.ProductInput {
	id := f.id

	tagSet := make(map[string]struct{})
	for _, tag := range p.Tags {
		tagSet[tag.(string)] = struct{}{}
	}
	for _, tag := range f.tags {
		if strings.HasPrefix(tag, "-") {
			delete(tagSet, strings.TrimPrefix(tag, "-"))
		} else {
			tagSet[tag] = struct{}{}
		}
	}
	tags := make([]any, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}

	input := schema.ProductInput{ID: &id, Seo: &schema.SEOInput{
		Title:       p.Seo.Title,
		Description: p.Seo.Description,
	}}

	if f.handle != "" {
		input.Handle = &f.handle
	}
	if f.title != "" {
		input.Title = &f.title
	}
	if f.descHtml != nil {
		input.DescriptionHtml = f.descHtml
	}
	if f.typ != nil {
		input.ProductType = f.typ
	}
	if f.category != "" {
		input.Category = &f.category
	}
	if len(tags) > 0 {
		input.Tags = tags
	}
	if f.vendor != "" {
		input.Vendor = &f.vendor
	}
	if f.status != "" {
		status := schema.ProductStatus(strings.ToTitle(f.status))
		input.Status = &status
	}
	if f.seoTitle != nil {
		input.Seo.Title = f.seoTitle
	}
	if f.seoDesc != nil {
		input.Seo.Description = f.seoDesc
	}

	return &input
}

func hasAnythingToUpdate(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("handle") ||
		cmd.Flags().Changed("title") ||
		cmd.Flags().Changed("desc") ||
		cmd.Flags().Changed("type") ||
		cmd.Flags().Changed("category") ||
		cmd.Flags().Changed("tags") ||
		cmd.Flags().Changed("vendor") ||
		cmd.Flags().Changed("status") ||
		cmd.Flags().Changed("seo-title") ||
		cmd.Flags().Changed("seo-desc")
}
