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

type Metafield struct {
	Client  *api.GQLClient
	Logger  *tlog.Logger
	File    registry.File
	Summary *runner.Summary
	DryRun  bool
}

func (h Metafield) Handle(data any) (any, error) {
	var realProductID string
	if id, ok := data.(string); ok {
		realProductID = id
	} else {
		return nil, fmt.Errorf("unable to figure out real product ID")
	}

	metaRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}
	h.Summary.Count += 1

	var meta api.ProductMetafieldsData
	if err = json.Unmarshal(metaRaw, &meta); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		h.Summary.Failed += 1
		return nil, err
	}

	toAdd := make([]*schema.Metafield, 0)
	toDelete := make([]*schema.Metafield, 0)

	// Get upstream metafields.
	currentMetafields, _ := h.Client.GetProductMetaFields(realProductID)
	if currentMetafields != nil {
		currentMetafieldsMap := make(map[string]*schema.Metafield, len(currentMetafields.Data.Product.Metafields.Nodes))
		for _, opt := range currentMetafields.Data.Product.Metafields.Nodes {
			currentMetafieldsMap[opt.ID] = &opt
		}

		backupMetafieldsMap := make(map[string]*schema.Metafield, len(meta.Metafields.Nodes))
		for _, m := range meta.Metafields.Nodes {
			backupMetafieldsMap[m.ID] = &m
		}

		for id, cm := range currentMetafieldsMap {
			if m, ok := backupMetafieldsMap[id]; ok {
				toAdd = append(toAdd, m)
			} else {
				toDelete = append(toDelete, cm)
			}
		}
		for id, m := range backupMetafieldsMap {
			if _, ok := currentMetafieldsMap[id]; !ok {
				toAdd = append(toAdd, m)
			}
		}
	} else {
		for _, m := range meta.Metafields.Nodes {
			toAdd = append(toAdd, &m)
		}
	}

	attemptSync := func(pid string) error {
		if _, err := h.handleProductMetaFieldsSet(pid, toAdd); err != nil {
			return err
		}
		if _, err := h.handleProductMetaFieldsDelete(pid, toDelete); err != nil {
			return err
		}
		return nil
	}
	h.Logger.V(1).Info("Attempting to sync product metafileds", "oldID", meta.ProductID, "upstreamID", realProductID)
	if h.DryRun {
		h.Logger.V(tlog.VL2).Infof("Product metafields to sync - add: %d, remove: %d", len(toAdd), len(toDelete))
		h.Logger.V(tlog.VL3).Warn("Skipping product metafields sync")
		h.Summary.Passed += 1
		return nil, nil
	}
	err = attemptSync(realProductID)
	if err != nil {
		h.Logger.Error("Failed to sync product metafields", "oldID", meta.ProductID, "upstreamID", realProductID)
		h.Summary.Failed += 1
		return nil, err
	}
	h.Summary.Passed += 1
	return nil, nil
}

func (h Metafield) handleProductMetaFieldsSet(productID string, toAdd []*schema.Metafield) (*api.MetafieldSetResponse, error) {
	if len(toAdd) == 0 {
		return nil, nil
	}

	metafields := make([]schema.MetafieldsSetInput, 0, len(toAdd))
	for _, m := range toAdd {
		metafields = append(metafields, schema.MetafieldsSetInput{
			Namespace: &m.Namespace,
			Key:       m.Key,
			Value:     m.Value,
			OwnerID:   productID,
			Type:      &m.Type,
		})
	}
	h.Logger.V(tlog.VL2).Info("Attempting to set product metafields", "id", productID)
	return h.Client.SetMetafields(metafields)
}

func (h Metafield) handleProductMetaFieldsDelete(productID string, toDelete []*schema.Metafield) (*api.MetafieldDeleteResponse, error) {
	if len(toDelete) == 0 {
		return nil, nil
	}

	metafields := make([]schema.MetafieldIdentifierInput, 0, len(toDelete))
	for _, m := range toDelete {
		metafields = append(metafields, schema.MetafieldIdentifierInput{
			Key:       m.Key,
			Namespace: m.Namespace,
			OwnerID:   productID,
		})
	}
	h.Logger.V(tlog.VL2).Info("Attempting to delete product metafields", "id", productID)
	return h.Client.DeleteMetafields(metafields)
}
