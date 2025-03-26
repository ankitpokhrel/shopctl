package update

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

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

// Flag wraps available command flags.
type flag struct {
	id       string
	handle   string
	title    string
	descHtml string
	typ      string
	category string
	tags     []string
	vendor   string
	status   string
	web      bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	id := args[0]

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

	web, err := cmd.Flags().GetBool("web")
	cmdutil.ExitOnErr(err)

	f.id = id
	f.handle = handle
	f.title = title
	f.descHtml = desc
	f.typ = typ
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
		Short:   "Update lets you update a product",
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
	cmd.Flags().String("tags", "", "Comma separated product tags")
	cmd.Flags().String("vendor", "", "Product vendor")
	cmd.Flags().String("status", "", "Product status (ACTIVE, ARCHIVED, DRAFT)")
	cmd.Flags().Bool("web", false, "Open in web browser after successful update")

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	product, err := client.GetProductByID(cmdutil.ShopifyProductID(flag.id))
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
		ctx.Alias, cmdutil.ExtractNumericID(res.Product.ID),
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
	if !strings.HasPrefix(id, "gid://") {
		id = "gid://shopify/Product/" + id
	}

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

	input := schema.ProductInput{ID: &id}

	if f.handle != "" {
		input.Handle = &f.handle
	}
	if f.title != "" {
		input.Title = &f.title
	}
	if f.descHtml != "" {
		input.DescriptionHtml = &f.descHtml
	}
	if f.typ != "" {
		input.ProductType = &f.typ
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

	if !hasAnythingToUpdate(input) {
		cmdutil.Warn("Nothing to update")
		os.Exit(0)
	}
	return &input
}

func hasAnythingToUpdate(input schema.ProductInput) bool {
	return input.Handle != nil ||
		input.Title != nil ||
		input.DescriptionHtml != nil ||
		input.ProductType != nil ||
		input.Category != nil ||
		len(input.Tags) == 0 ||
		input.Vendor != nil ||
		input.Status != nil
}
