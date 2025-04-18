package handler

import (
	"encoding/json"
	"fmt"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Variant struct {
	Client  *api.GQLClient
	Logger  *tlog.Logger
	File    registry.File
	Summary *runner.Summary
	DryRun  bool
}

func (h *Variant) Handle(data any) (any, error) {
	var realProductID string
	if id, ok := data.(string); ok {
		realProductID = id
	} else {
		return nil, fmt.Errorf("unable to figure out real product ID")
	}
	variantsRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}
	h.Summary.Count += 1

	var product api.ProductVariantData
	if err = json.Unmarshal(variantsRaw, &product); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		h.Summary.Failed += 1
		return nil, err
	}

	toAdd := make([]*schema.ProductVariant, 0)
	toUpdate := make([]*schema.ProductVariant, 0)
	toDelete := make([]string, 0)

	// Get upstream variants.
	currentVariants, _ := h.Client.GetProductVariants(realProductID)
	if currentVariants != nil {
		currentVariantsMap := make(map[string]*schema.ProductVariant, len(currentVariants.Variants.Nodes))
		for _, v := range currentVariants.Variants.Nodes {
			currentVariantsMap[keyme(v.Title)] = &v
		}

		backupVariantsMap := make(map[string]*schema.ProductVariant, len(product.Variants.Nodes))
		for _, v := range product.Variants.Nodes {
			backupVariantsMap[keyme(v.Title)] = &v
		}

		for k := range currentVariantsMap {
			id := currentVariantsMap[k].ID
			if v, ok := backupVariantsMap[k]; ok {
				v.ID = id
				toUpdate = append(toUpdate, v)
			} else {
				toDelete = append(toDelete, id)
			}
		}
		for k, v := range backupVariantsMap {
			if _, ok := currentVariantsMap[k]; !ok {
				toAdd = append(toAdd, v)
			}
		}
	} else {
		for _, v := range product.Variants.Nodes {
			toAdd = append(toAdd, &v)
		}
	}

	attemptSync := func(pid string) error {
		if _, err := h.handleProductVariantDelete(pid, toDelete); err != nil {
			return err
		}
		if _, err := h.handleProductVariantAdd(pid, toAdd); err != nil {
			return err
		}
		if _, err := h.handleProductVariantUpdate(pid, toUpdate); err != nil {
			return err
		}
		return nil
	}

	h.Logger.V(1).Info("Attempting to sync product variants", "oldID", product.ProductID, "upstreamID", realProductID)
	if h.DryRun {
		h.Logger.V(tlog.VL2).Infof("Product variants to sync - add: %d, update: %d", len(toAdd), len(toUpdate))
		h.Logger.V(tlog.VL3).Warn("Skipping product variants sync")
		h.Summary.Passed += 1
		return nil, nil
	}
	err = attemptSync(realProductID)
	if err != nil {
		h.Logger.Error("Failed to sync product variants", "oldID", product.ProductID, "upstreamID", realProductID)
		h.Summary.Failed += 1
		return nil, err
	}
	h.Summary.Passed += 1
	return nil, nil
}

func (h Variant) handleProductVariantDelete(productID string, toDelete []string) (*api.ProductVariantsSyncResponse, error) {
	if len(toDelete) == 0 {
		return nil, nil
	}
	return h.Client.DeleteProductVariants(productID, toDelete)
}

func (h Variant) handleProductVariantAdd(productID string, toAdd []*schema.ProductVariant) (*api.ProductVariantsSyncResponse, error) {
	return h.createOrUpdateProductVariants(productID, toAdd, false)
}

func (h Variant) handleProductVariantUpdate(productID string, toUpdate []*schema.ProductVariant) (*api.ProductVariantsSyncResponse, error) {
	return h.createOrUpdateProductVariants(productID, toUpdate, true)
}

func (h Variant) createOrUpdateProductVariants(productID string, variants []*schema.ProductVariant, isUpdate bool) (*api.ProductVariantsSyncResponse, error) {
	if len(variants) == 0 {
		return nil, nil
	}

	variantsInput := make([]schema.ProductVariantsBulkInput, 0, len(variants))
	for _, v := range variants {
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
		if isUpdate {
			input.ID = &v.ID
		}
		variantsInput = append(variantsInput, input)
	}

	if isUpdate {
		h.Logger.V(tlog.VL2).Info("Attempting to update product variant", "id", productID)
		return h.Client.UpdateProductVariants(productID, variantsInput, false)
	}
	h.Logger.V(tlog.VL2).Info("Attempting to create product variant", "id", productID)
	return h.Client.CreateProductVariants(productID, variantsInput, schema.ProductVariantsBulkCreateStrategyRemoveStandaloneVariant)
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
