package handler

import (
	"encoding/json"
	"errors"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Option struct {
	ProductID string
	Client    *api.GQLClient
	Logger    *tlog.Logger
	File      registry.File
	DryRun    bool
}

func (h Option) Handle(data any) (any, error) {
	var realProductID string
	if id, ok := data.(string); ok {
		realProductID = id
	}
	productRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	var product schema.Product
	if err = json.Unmarshal(productRaw, &product); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	// Get upstream options.
	currentOptions, err := h.Client.GetProductOptions(realProductID)
	if err != nil {
		return nil, err
	}

	currentOptionsMap := make(map[string]*schema.ProductOption, len(currentOptions.Data.Product.Options))
	for _, opt := range currentOptions.Data.Product.Options {
		currentOptionsMap[opt.ID] = &opt
	}

	backupOptionsMap := make(map[string]*schema.ProductOption, len(product.Options))
	for _, opt := range product.Options {
		backupOptionsMap[opt.ID] = &opt
	}

	toAdd := make([]*schema.ProductOption, 0)
	toUpdate := make([]*schema.ProductOption, 0)
	toDelete := make([]string, 0)

	for id := range currentOptionsMap {
		if opt, ok := backupOptionsMap[id]; ok {
			toUpdate = append(toUpdate, opt)
		} else {
			toDelete = append(toDelete, id)
		}
	}
	for id, opt := range backupOptionsMap {
		if _, ok := currentOptionsMap[id]; !ok {
			toAdd = append(toAdd, opt)
		}
	}

	attemptSync := func(pid string) error {
		if _, err := h.handleProductOptionDelete(pid, toDelete); err != nil {
			return err
		}
		if _, err := h.handleProductOptionAdd(pid, toAdd); err != nil {
			return err
		}
		if _, err := h.handlProductOptionUpdate(pid, currentOptionsMap, toUpdate); err != nil {
			return err
		}
		return nil
	}

	h.Logger.V(1).Info("Attempting to sync product options", "oldID", product.ID, "upstreamID", realProductID)
	if h.DryRun {
		h.Logger.V(tlog.VL2).Infof("Product options to sync - add: %d, update: %d, remove: %d", len(toAdd), len(toUpdate), len(toDelete))
		h.Logger.V(tlog.VL3).Warn("Skipping product options sync")
		return nil, nil
	}
	err = attemptSync(realProductID)
	if err != nil {
		h.Logger.Error("Failed to sync product options", "oldID", product.ID, "upstreamID", realProductID)
	}
	return nil, err
}

func (h Option) handleProductOptionAdd(productID string, toAdd []*schema.ProductOption) (*api.ProductOptionSyncResponse, error) {
	if len(toAdd) == 0 {
		return nil, nil
	}

	options := make([]schema.OptionCreateInput, 0, len(toAdd))
	for _, opt := range toAdd {
		optionValues := make([]any, 0, len(opt.OptionValues))
		linkedOptionValues := make([]any, 0, len(optionValues))
		for _, val := range opt.OptionValues {
			if val.LinkedMetafieldValue == nil {
				optionValues = append(optionValues, schema.OptionValueCreateInput{Name: &val.Name})
			} else {
				linkedOptionValues = append(linkedOptionValues, val.LinkedMetafieldValue)
			}
		}
		if opt.LinkedMetafield != nil {
			linkedMetaField := &schema.LinkedMetafieldCreateInput{
				Key:       *opt.LinkedMetafield.Key,
				Namespace: *opt.LinkedMetafield.Namespace,
				Values:    linkedOptionValues,
			}
			options = append(options, schema.OptionCreateInput{
				Name:            &opt.Name,
				Position:        &opt.Position,
				LinkedMetafield: linkedMetaField,
			})
		} else {
			options = append(options, schema.OptionCreateInput{
				Name:     &opt.Name,
				Position: &opt.Position,
				Values:   optionValues,
			})
		}
	}
	h.Logger.V(tlog.VL2).Info("Attempting to create product options", "id", productID)
	return h.Client.CreateProductOptions(productID, options, schema.ProductOptionCreateVariantStrategyLeaveAsIs)
}

func (h Option) handlProductOptionUpdate(productID string, currentOptionsMap map[string]*schema.ProductOption, toUpdate []*schema.ProductOption) (*api.ProductOptionSyncResponse, error) {
	if len(toUpdate) == 0 {
		return nil, nil
	}

	var (
		updateResponses = make([]*api.ProductOptionSyncResponse, 0)
		updateErrors    = make([]error, 0)
	)

	h.Logger.V(tlog.VL2).Info("Attempting to update product options", "id", productID)
	for _, opt := range toUpdate {
		option := schema.OptionUpdateInput{
			ID:       opt.ID,
			Name:     &opt.Name,
			Position: &opt.Position,
		}
		currentOptionValues := currentOptionsMap[opt.ID].OptionValues
		newOptionValues := opt.OptionValues

		currentOptionValuesMap := make(map[string]*schema.ProductOptionValue, 0)
		for _, v := range currentOptionValues {
			currentOptionValuesMap[v.ID] = &v
		}
		newOptionValuesMap := make(map[string]*schema.ProductOptionValue, 0)
		for _, v := range newOptionValues {
			newOptionValuesMap[v.ID] = &v
		}

		optionValuesToAdd := make([]schema.OptionValueCreateInput, 0)
		optionValuesToUpdate := make([]schema.OptionValueUpdateInput, 0)
		optionValuesToDelete := make([]string, 0)

		for _, nv := range newOptionValues {
			if _, ok := currentOptionValuesMap[nv.ID]; !ok {
				if opt.LinkedMetafield != nil {
					optionValuesToAdd = append(optionValuesToAdd, schema.OptionValueCreateInput{
						LinkedMetafieldValue: nv.LinkedMetafieldValue,
					})
				} else {
					optionValuesToAdd = append(optionValuesToAdd, schema.OptionValueCreateInput{
						Name: &nv.Name,
					})
				}
			}
		}

		for _, cv := range currentOptionValues {
			if nv, ok := newOptionValuesMap[cv.ID]; ok {
				if opt.LinkedMetafield != nil {
					optionValuesToUpdate = append(optionValuesToUpdate, schema.OptionValueUpdateInput{
						ID:                   cv.ID,
						LinkedMetafieldValue: nv.LinkedMetafieldValue,
					})
				} else {
					optionValuesToUpdate = append(optionValuesToUpdate, schema.OptionValueUpdateInput{
						ID:   cv.ID,
						Name: &nv.Name,
					})
				}
			} else {
				optionValuesToDelete = append(optionValuesToDelete, cv.ID)
			}
		}

		out, err := h.Client.UpdateProductOptions(productID, &option, optionValuesToAdd, optionValuesToUpdate, optionValuesToDelete, schema.ProductOptionUpdateVariantStrategyManage)
		if err != nil {
			updateErrors = append(updateErrors, err)
		} else {
			updateResponses = append(updateResponses, out)
		}
	}

	if len(updateErrors) > 0 {
		return nil, errors.Join(updateErrors...)
	}
	// We don't really care about update response atm.
	return updateResponses[0], nil
}

func (h Option) handleProductOptionDelete(productID string, toDelete []string) (*api.ProductOptionSyncResponse, error) {
	if len(toDelete) == 0 {
		return nil, nil
	}
	h.Logger.V(tlog.VL2).Info("Attempting to delete product options", "id", productID)
	return h.Client.DeleteProductOptions(productID, toDelete)
}
