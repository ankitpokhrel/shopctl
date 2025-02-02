package product

import (
	"encoding/json"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type productHandler struct {
	client *api.GQLClient
	logger *tlog.Logger
	file   registry.File
}

func (h *productHandler) Handle() (any, error) {
	product, err := registry.ReadFileContents(h.file.Path)
	if err != nil {
		h.logger.Error("Unable to read contents", "file", h.file.Path, "error", err)
		return nil, err
	}

	var prod schema.Product
	if err = json.Unmarshal(product, &prod); err != nil {
		h.logger.Error("Unable to marshal contents", "file", h.file.Path, "error", err)
		return nil, err
	}

	// TODO: Handle/log error.
	res, err := createOrUpdateProduct(&prod, h.client, h.logger)
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

	var category *string
	if product.Category != nil {
		category = &product.Category.ID
	}

	if res.Data.Product.ID != "" {
		// Product exists, execute update mutation.
		input := schema.ProductInput{
			ID:                     &product.ID,
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
			CollectionsToJoin:      nil,
			RequiresSellingPlan:    &product.RequiresSellingPlan,
		}
		lgr.Warn("Product already exists, updating", "id", product.ID)
		return client.UpdateProduct(input)
	}

	//	Product does not exist, execute create mutation.
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
		CollectionsToJoin:      nil,
		CombinedListingRole:    product.CombinedListingRole,
		RequiresSellingPlan:    &product.RequiresSellingPlan,
	}
	lgr.Info("Creating product", "id", product.ID)
	return client.CreateProduct(input)
}
