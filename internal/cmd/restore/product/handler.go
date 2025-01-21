package product

import (
	"encoding/json"
	"fmt"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/schema"
)

type productHandler struct {
	client *api.GQLClient
	file   registry.File
}

func (h *productHandler) Handle() (any, error) {
	product, err := registry.ReadFileContents(h.file.Path)
	if err != nil {
		lgr.Error("Unable to read contents", "file", h.file.Path, "error", err)
		return nil, err
	}

	var prod schema.Product
	if err = json.Unmarshal(product, &prod); err != nil {
		lgr.Error("Unable to marshal contents", "file", h.file.Path, "error", err)
		return nil, err
	}

	// TODO: Handle/log error.
	res, err := createOrUpdateProduct(&prod, h.client)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return res, fmt.Errorf("errors occurred while restoring product: %+v", res.Errors)
	}
	return res, nil
}

func createOrUpdateProduct(product *schema.Product, client *api.GQLClient) (*api.ProductCreateResponse, error) {
	res, err := client.CheckProductByID(product.ID)
	if err != nil {
		return nil, err
	}

	// TODO: Compare and extract fields that are actually updated.

	var category *string
	if product.Category != nil {
		category = &product.Category.Name
	}

	if res.Data.Product.ID != "" {
		// Product exists, execute update mutation.
		input := schema.ProductUpdateInput{
			ID:                     &product.ID,
			DescriptionHtml:        &product.DescriptionHtml,
			Handle:                 &product.Handle,
			Seo:                    &schema.SEOInput{Title: product.Seo.Title, Description: product.Seo.Description},
			ProductType:            &product.ProductType,
			Category:               category,
			Tags:                   product.Tags,
			TemplateSuffix:         product.TemplateSuffix,
			GiftCardTemplateSuffix: product.GiftCardTemplateSuffix,
			Title:                  &product.Title,
			Vendor:                 &product.Vendor,
			CollectionsToJoin:      nil,
			Status:                 &product.Status,
			RequiresSellingPlan:    &product.RequiresSellingPlan,
		}
		lgr.Warn("Product already exists, updating", "id", product.ID)
		return client.UpdateProduct(input)
	}

	// Product does not exist, execute create mutation.
	input := schema.ProductCreateInput{
		DescriptionHtml:        &product.DescriptionHtml,
		Handle:                 &product.Handle,
		Seo:                    &schema.SEOInput{Title: product.Seo.Title, Description: product.Seo.Description},
		ProductType:            &product.ProductType,
		Category:               category,
		Tags:                   product.Tags,
		TemplateSuffix:         product.TemplateSuffix,
		GiftCardTemplateSuffix: product.GiftCardTemplateSuffix,
		Title:                  &product.Title,
		Vendor:                 &product.Vendor,
		GiftCard:               &product.IsGiftCard,
		CollectionsToJoin:      nil,
		CombinedListingRole:    product.CombinedListingRole,
		Status:                 &product.Status,
		RequiresSellingPlan:    &product.RequiresSellingPlan,
	}
	lgr.Info("Creating product", "id", product.ID)
	return client.CreateProduct(input)
}
