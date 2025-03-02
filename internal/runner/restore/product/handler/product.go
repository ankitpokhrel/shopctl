package handler

import (
	"encoding/json"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Product struct {
	Client *api.GQLClient
	Logger *tlog.Logger
	File   registry.File
}

func (h *Product) Handle() (any, error) {
	product, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	var prod schema.Product
	if err = json.Unmarshal(product, &prod); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	res, err := createOrUpdateProduct(&prod, h.Client, h.Logger)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func createOrUpdateProduct(product *schema.Product, client *api.GQLClient, lgr *tlog.Logger) (*api.ProductCreateResponse, error) {
	res, err := client.CheckProductByID(product.ID)
	if err != nil {
		return nil, err
	}

	// TODO: Compare and extract fields that are actually updated.

	var (
		category *string

		options           = make([]any, 0, len(product.Options))
		redirectNewHandle = true // Auto create redirect if handle is changed.
	)

	if product.Category != nil {
		category = &product.Category.ID
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

	input := schema.ProductInput{
		Handle:                 &product.Handle,
		Title:                  &product.Title,
		DescriptionHtml:        &product.DescriptionHtml,
		ProductType:            &product.ProductType,
		Category:               category,
		Tags:                   product.Tags,
		Vendor:                 &product.Vendor,
		Seo:                    &schema.SEOInput{Title: product.Seo.Title, Description: product.Seo.Description},
		Status:                 &product.Status,
		TemplateSuffix:         product.TemplateSuffix,
		GiftCardTemplateSuffix: product.GiftCardTemplateSuffix,
		GiftCard:               &product.IsGiftCard,
		CollectionsToJoin:      nil, // Not supported yet.
		CollectionsToLeave:     nil, // Not supported yet.
		RedirectNewHandle:      &redirectNewHandle,
		CombinedListingRole:    product.CombinedListingRole,
		RequiresSellingPlan:    &product.RequiresSellingPlan,
		ClaimOwnership:         nil, // No way to get this value.
	}

	if res.Data.Product.ID != "" {
		input.ID = &product.ID

		lgr.Warn("Product already exists, updating", "id", product.ID)
		return client.UpdateProduct(input)
	}

	// Some fields can only be specified during create.
	input.ProductOptions = options // Note that UI may not display all options unless there is a variant with that option.

	lgr.Info("Creating product", "id", product.ID)
	return client.CreateProduct(input)
}
