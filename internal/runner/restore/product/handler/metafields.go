package handler

import (
	"encoding/json"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Metafield struct {
	Client *api.GQLClient
	Logger *tlog.Logger
	File   registry.File
}

func (h Metafield) Handle() (any, error) {
	metaRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	var meta api.ProductMetafieldsData
	if err = json.Unmarshal(metaRaw, &meta); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	// Get upstream options.
	currentMetafields, err := h.Client.GetProductMetaFields(meta.ProductID)
	if err != nil {
		return nil, err
	}

	currentMetafieldsMap := make(map[string]*schema.Metafield, len(currentMetafields.Data.Product.Metafields.Nodes))
	for _, opt := range currentMetafields.Data.Product.Metafields.Nodes {
		currentMetafieldsMap[opt.ID] = &opt
	}

	backupMetafieldsMap := make(map[string]*schema.Metafield, len(meta.Metafields.Nodes))
	for _, m := range meta.Metafields.Nodes {
		backupMetafieldsMap[m.ID] = &m
	}

	toAdd := make([]*schema.Metafield, 0)
	toDelete := make([]*schema.Metafield, 0)

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

	attemptSync := func(pid string) error {
		if _, err := h.handleProductMetaFieldsSet(pid, toAdd); err != nil {
			return err
		}
		if _, err := h.handleProductMetaFieldsDelete(pid, toDelete); err != nil {
			return err
		}
		return nil
	}
	err = attemptSync(meta.ProductID)
	if err != nil {
		h.Logger.Error("Failed to sync product metafields", "id", meta.ProductID)
	}
	return nil, err
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
	h.Logger.V(2).Info("Attempting to set product metafields", "id", productID)
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
	h.Logger.V(2).Info("Attempting to delete product metafields", "id", productID)
	return h.Client.DeleteMetafields(metafields)
}
