package handler

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Product struct {
	Client  *api.GQLClient
	Logger  *tlog.Logger
	File    registry.File
	Filter  *runner.RestoreFilter
	Summary *runner.Summary
	DryRun  bool
}

func (h *Product) Handle(data any) (any, error) {
	productRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}
	h.Summary.Count += 1

	var product schema.Product
	if err = json.Unmarshal(productRaw, &product); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		h.Summary.Failed += 1
		return nil, err
	}

	// Filter product.
	if len(h.Filter.Filters) > 0 {
		matched, err := matchesFilters(&product, h.Filter)
		if err != nil || !matched {
			h.Summary.Skipped += 1
			return nil, engine.ErrSkipChildren
		}
	}

	if h.DryRun {
		h.Logger.V(tlog.VL3).Warn("Skipping product sync")
		h.Summary.Passed += 1
		return product.ID, nil
	}
	res, err := createOrUpdateProduct(&product, h.Client, h.Logger)
	if err != nil {
		h.Summary.Failed += 1
		return nil, err
	}
	h.Summary.Passed += 1
	return res.Product.ID, nil
}

func createOrUpdateProduct(product *schema.Product, client *api.GQLClient, lgr *tlog.Logger) (*api.ProductCreateResponse, error) {
	res, err := client.CheckProductByHandle(product.Handle)
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

	if res.ID != "" {
		input.ID = &res.ID

		lgr.Warn("Product already exists, updating", "id", res.ID, "handle", res.Handle)
		return client.UpdateProduct(input, nil)
	}

	// Some fields can only be specified during create.
	input.ProductOptions = options // Note that UI may not display all options unless there is a variant with that option.

	lgr.Info("Creating product", "oldID", product.ID, "handle", product.Handle)
	return client.CreateProduct(input)
}

func matchesFilters(product *schema.Product, rf *runner.RestoreFilter) (bool, error) {
	containsAny := func(s []any, v any) bool { return slices.Contains(s, v) } //nolint:gocritic

	results := []bool{}
	for key, values := range rf.Filters {
		switch strings.ToLower(key) {
		case "id":
			matched := slices.Contains(values, shopctl.ExtractNumericID(product.ID))
			results = append(results, matched)
		case "handle":
			matched := slices.Contains(values, string(product.Handle))
			results = append(results, matched)
		case "title":
			matched := slices.Contains(values, string(product.Title))
			results = append(results, matched)
		case "tags":
			matched := false
			for _, tag := range values {
				if containsAny(product.Tags, tag) {
					matched = true
					break
				}
			}
			results = append(results, matched)
		case "status":
			matched := slices.Contains(values, strings.ToLower(string(product.Status)))
			results = append(results, matched)
		case "producttype":
			matched := slices.Contains(values, strings.ToLower(string(product.ProductType)))
			results = append(results, matched)
		case "category":
			matched := false
			if product.Category != nil {
				matched = slices.Contains(values, strings.ToLower(product.Category.Name))
			}
			results = append(results, matched)
		default:
			return false, fmt.Errorf("unsupported filter key: %s", key)
		}
	}

	if len(results) == 0 {
		return false, nil
	}

	// Combine results using separators
	finalResult := results[0]
	for i, sep := range rf.Separators {
		switch strings.ToLower(sep) {
		case "and":
			finalResult = finalResult && results[i+1]
		case "or":
			finalResult = finalResult || results[i+1]
		default:
			return false, fmt.Errorf("unsupported separator: %s", sep)
		}
	}
	return finalResult, nil
}
