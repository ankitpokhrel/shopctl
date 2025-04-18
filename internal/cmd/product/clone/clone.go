package clone

import (
	"fmt"
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
	helpText = `Clone lets you duplicate a product.`

	examples = `$ shopctl product clone 8856145494

# Clone product with custom title and handle
$ shopctl product clone 8856145494 --title "Cloned product" --handle "cloned-product"

# Clone product along with its variants and media and set status to DRAFT
$ shopctl product clone 8856145494 --variants --media --status draft

# Clone product and replace some strings in title and description
$ shopctl product clone 8856145494 -H "find-me:replace-me" -H "also-find-me:and-replace-me"

# Clone product to another store and add tag 'cloned' and 'store1'
$ shopctl product clone 8856145494 --to store2 --tags cloned,store1`
)

type flag struct {
	id       string
	handle   string
	title    string
	status   string
	tags     []string
	variants bool
	media    bool
	to       string
	replace  []string
	web      bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	handle, err := cmd.Flags().GetString("handle")
	cmdutil.ExitOnErr(err)

	title, err := cmd.Flags().GetString("title")
	cmdutil.ExitOnErr(err)

	status, err := cmd.Flags().GetString("status")
	cmdutil.ExitOnErr(err)

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	variants, err := cmd.Flags().GetBool("variants")
	cmdutil.ExitOnErr(err)

	media, err := cmd.Flags().GetBool("media")
	cmdutil.ExitOnErr(err)

	to, err := cmd.Flags().GetString("to")
	cmdutil.ExitOnErr(err)

	replace, err := cmd.Flags().GetStringArray("replace")
	cmdutil.ExitOnErr(err)

	web, err := cmd.Flags().GetBool("web")
	cmdutil.ExitOnErr(err)

	f.id = shopctl.ShopifyProductID(args[0])
	f.handle = handle
	f.title = title
	f.status = strings.ToUpper(status)
	f.tags = strings.Split(tags, ",")
	f.variants = variants
	f.media = media
	f.to = to
	f.replace = replace
	f.web = web
}

// NewCmdClone constructs a new product clone command.
func NewCmdClone() *cobra.Command {
	cmd := cobra.Command{
		Use:     "clone",
		Short:   "Clone a product",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}
	cmd.Flags().String("handle", "", "Unique product handle")
	cmd.Flags().StringP("title", "t", "", "Override product title")
	cmd.Flags().String("status", "", "Override product status (ACTIVE, ARCHIVED, DRAFT)")
	cmd.Flags().String("tags", "", "Additional comma separated product tags to add")
	cmd.Flags().Bool("variants", false, "Whether to clone variants")
	cmd.Flags().Bool("media", false, "Whether to clone media")
	cmd.Flags().StringArrayP("replace", "H", []string{}, "Replace strings in title and description; Format <search>:<replace>, eg: \"find-me:replace-with-me\"")
	cmd.Flags().String("to", "", "Store to clone product to")
	cmd.Flags().Bool("web", false, "Open in web browser after successful cloning")

	cmd.Flags().SortFlags = false

	return &cmd
}

//nolint:gocyclo
func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	product, err := client.GetProductByID(flag.id)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	var (
		handle   *string
		category *string
		title    = product.Title
		body     = product.DescriptionHtml
		status   = product.Status

		options           = make([]any, 0, len(product.Options))
		redirectNewHandle = true
	)

	if flag.handle != "" {
		handle = &flag.handle
	}
	if flag.title != "" {
		title = flag.title
	}
	if flag.status != "" {
		status = schema.ProductStatus(flag.status)
	}
	if product.Category != nil {
		category = &product.Category.ID
	}
	if len(flag.replace) > 0 {
		for _, r := range flag.replace {
			parts := strings.Split(r, ":")
			if len(parts) != 2 {
				fmt.Println()
				cmdutil.Fail("Replace string must be in the following format <find>:<replace>. Skipping replacement...")
			} else {
				from, to := parts[0], parts[1]

				title = strings.ReplaceAll(title, from, to)
				body = strings.ReplaceAll(body, from, to)
			}
		}
	}

	for _, opt := range product.Options {
		values := make([]any, 0, len(opt.Values))
		for _, v := range opt.OptionValues {
			values = append(values, schema.OptionValueCreateInput{
				Name: &v.Name,
			})
		}
		options = append(options, schema.OptionCreateInput{
			Name:     &opt.Name,
			Position: &opt.Position,
			Values:   values,
		})
	}
	tagSet := make(map[string]struct{})
	for _, tag := range product.Tags {
		tagSet[tag.(string)] = struct{}{}
	}
	for _, tag := range flag.tags {
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

	input := schema.ProductInput{
		Handle:                 handle,
		Title:                  &title,
		DescriptionHtml:        &body,
		ProductType:            &product.ProductType,
		ProductOptions:         options,
		Category:               category,
		Tags:                   tags,
		Vendor:                 &product.Vendor,
		Seo:                    &schema.SEOInput{Title: product.Seo.Title, Description: product.Seo.Description},
		Status:                 &status,
		TemplateSuffix:         product.TemplateSuffix,
		GiftCardTemplateSuffix: product.GiftCardTemplateSuffix,
		GiftCard:               &product.IsGiftCard,
		CollectionsToJoin:      nil,
		CollectionsToLeave:     nil,
		RedirectNewHandle:      &redirectNewHandle,
		CombinedListingRole:    product.CombinedListingRole,
		RequiresSellingPlan:    &product.RequiresSellingPlan,
		ClaimOwnership:         nil,
	}

	alias := ctx.Alias
	if flag.to != "" {
		alias = flag.to
		client = api.NewGQLClient(fmt.Sprintf("%s.myshopify.com", alias))
	}

	res, err := client.CreateProduct(input)
	if err != nil {
		return err
	}

	if flag.variants {
		_, err := cloneProductVariants(res.Product.ID, product.Variants.Nodes, client)
		if err != nil {
			cmdutil.Fail("There were some errors when cloning product variants")
		}
	}
	if flag.media {
		_, err := cloneProductMedia(res.Product.ID, tags, product.Media.Nodes, client)
		if err != nil {
			cmdutil.Fail("There were some errors when cloning product media")
		}
	}

	adminURL := fmt.Sprintf(
		"https://admin.shopify.com/store/%s/products/%s",
		alias, shopctl.ExtractNumericID(res.Product.ID),
	)
	if flag.web {
		_ = browser.Browse(adminURL)
	}

	cmdutil.Success("Product created successfully: %s", res.Product.Handle)
	fmt.Println(adminURL)

	return nil
}

func cloneProductVariants(productID string, toAdd []schema.ProductVariant, client *api.GQLClient) (*api.ProductVariantsSyncResponse, error) {
	if len(toAdd) == 0 {
		return nil, nil
	}

	variantsInput := make([]schema.ProductVariantsBulkInput, 0, len(toAdd))
	for _, v := range toAdd {
		var inventoryItem *schema.InventoryItemInput

		if v.InventoryItem != nil {
			var cost float64
			if v.InventoryItem.UnitCost != nil {
				cost = v.InventoryItem.UnitCost.Amount
			}
			inventoryItem = &schema.InventoryItemInput{
				Sku:                  v.InventoryItem.Sku,
				Cost:                 &cost,
				CountryCodeOfOrigin:  v.InventoryItem.CountryCodeOfOrigin,
				HarmonizedSystemCode: v.InventoryItem.HarmonizedSystemCode,
				ProvinceCodeOfOrigin: v.InventoryItem.ProvinceCodeOfOrigin,
				Tracked:              &v.InventoryItem.Tracked,
				Measurement: &schema.InventoryItemMeasurementInput{
					Weight: &schema.WeightInput{
						Value: v.InventoryItem.Measurement.Weight.Value,
						Unit:  v.InventoryItem.Measurement.Weight.Unit,
					},
				},
				RequiresShipping: &v.InventoryItem.RequiresShipping,
			}
		}

		input := schema.ProductVariantsBulkInput{
			Barcode:            v.Barcode,
			CompareAtPrice:     v.CompareAtPrice,
			InventoryPolicy:    &v.InventoryPolicy,
			InventoryItem:      inventoryItem,
			OptionValues:       getOptions(v.SelectedOptions),
			Price:              &v.Price,
			Taxable:            &v.Taxable,
			TaxCode:            v.TaxCode,
			RequiresComponents: &v.RequiresComponents,
		}
		variantsInput = append(variantsInput, input)
	}

	return client.CreateProductVariants(productID, variantsInput, schema.ProductVariantsBulkCreateStrategyRemoveStandaloneVariant)
}

func cloneProductMedia(productID string, tags []any, toAdd []any, client *api.GQLClient) (*api.ProductCreateResponse, error) {
	if len(toAdd) == 0 {
		return nil, nil
	}

	// Tags are replaced if we don't pass anything.
	// So, let's send them again as a workaround.
	input := schema.ProductInput{
		ID:   &productID,
		Tags: tags,
	}

	createMediaInput := make([]schema.CreateMediaInput, 0, len(toAdd))
	for _, node := range toAdd {
		m := node.(map[string]any)
		prv := m["preview"].(map[string]any)
		img := prv["image"].(map[string]any)

		url := img["url"].(string)
		alt := img["altText"].(string)
		mct := m["mediaContentType"].(string)

		createMediaInput = append(createMediaInput, schema.CreateMediaInput{
			OriginalSource:   url,
			Alt:              &alt,
			MediaContentType: schema.MediaContentType(mct),
		})
	}
	return client.UpdateProduct(input, createMediaInput)
}

func getOptions(options []any) []any {
	optionValues := make([]any, 0, len(options))

	for _, o := range options {
		opt, ok := o.(map[string]any)
		if !ok {
			continue
		}
		optValues, ok := opt["optionValue"].(map[string]any)
		if !ok {
			continue
		}
		name := opt["name"].(string)
		value := opt["value"].(string)
		linkedMetafield, ok := optValues["linkedMetafieldValue"].(string)
		if ok && linkedMetafield != "" {
			optionValues = append(optionValues, schema.VariantOptionValueInput{
				Name:                 &value,
				OptionName:           &name,
				LinkedMetafieldValue: &linkedMetafield,
			})
		} else {
			optionValues = append(optionValues, schema.VariantOptionValueInput{
				Name:       &value,
				OptionName: &name,
			})
		}
	}
	return optionValues
}
