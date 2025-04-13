package handler

import (
	"encoding/json"
	"fmt"

	"github.com/ankitpokhrel/shopctl"
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
	var realCustomerID string
	if id, ok := data.(string); ok {
		realCustomerID = id
	} else {
		return nil, fmt.Errorf("unable to figure out real cusotmer ID")
	}

	metaRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}
	h.Summary.Count += 1

	var meta api.CustomerMetafieldsData
	if err = json.Unmarshal(metaRaw, &meta); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		h.Summary.Failed += 1
		return nil, err
	}
	if len(meta.Metafields.Nodes) == 0 {
		return nil, nil
	}

	keyme := func(namespace, key string) string {
		return fmt.Sprintf("%s.%s", namespace, key)
	}

	toAdd := make([]*schema.Metafield, 0)
	toDelete := make([]*schema.Metafield, 0)

	// Get upstream metafields.
	currentMetafields, _ := h.Client.GetCustomerMetaFieldsByEmailOrPhoneOrID(
		&meta.Email, &meta.Phone,
		shopctl.ExtractNumericID(realCustomerID),
	)
	if currentMetafields != nil {
		currentMetaNode := currentMetafields.Data.Customers.Nodes[0]

		currentMetafieldsMap := make(map[string]*schema.Metafield, len(currentMetaNode.Metafields.Nodes))
		for _, m := range currentMetaNode.Metafields.Nodes {
			key := keyme(m.Namespace, m.Key)
			currentMetafieldsMap[key] = &m
		}

		backupMetafieldsMap := make(map[string]*schema.Metafield, len(meta.Metafields.Nodes))
		for _, m := range meta.Metafields.Nodes {
			key := keyme(m.Namespace, m.Key)
			backupMetafieldsMap[key] = &m
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

	attemptSync := func(cid string) error {
		if _, err := h.handleCustomerMetaFieldsSet(cid, toAdd); err != nil {
			return err
		}
		if _, err := h.handleCustomerMetaFieldsDelete(cid, toDelete); err != nil {
			return err
		}
		return nil
	}
	if h.DryRun {
		h.Logger.V(tlog.VL2).Infof("Customer metafields to sync - add: %d, remove: %d", len(toAdd), len(toDelete))
		h.Logger.V(tlog.VL3).Warn("Skipping customer metafields sync")
		h.Summary.Passed += 1
		return nil, nil
	}
	err = attemptSync(realCustomerID)
	if err != nil {
		h.Logger.Error("Failed to sync customer metafields", "oldID", meta.CustomerID, "upstreamID", realCustomerID)
		h.Summary.Failed += 1
		return nil, err
	}
	h.Summary.Passed += 1
	return nil, err
}

func (h Metafield) handleCustomerMetaFieldsSet(customerID string, toAdd []*schema.Metafield) (*api.MetafieldSetResponse, error) {
	if len(toAdd) == 0 {
		return nil, nil
	}

	metafields := make([]schema.MetafieldsSetInput, 0, len(toAdd))
	for _, m := range toAdd {
		metafields = append(metafields, schema.MetafieldsSetInput{
			Namespace: &m.Namespace,
			Key:       m.Key,
			Value:     m.Value,
			OwnerID:   customerID,
			Type:      &m.Type,
		})
	}
	h.Logger.V(tlog.VL2).Info("Attempting to set customer metafields", "id", customerID)
	return h.Client.SetMetafields(metafields)
}

func (h Metafield) handleCustomerMetaFieldsDelete(customerID string, toDelete []*schema.Metafield) (*api.MetafieldDeleteResponse, error) {
	if len(toDelete) == 0 {
		return nil, nil
	}

	metafields := make([]schema.MetafieldIdentifierInput, 0, len(toDelete))
	for _, m := range toDelete {
		metafields = append(metafields, schema.MetafieldIdentifierInput{
			Key:       m.Key,
			Namespace: m.Namespace,
			OwnerID:   customerID,
		})
	}
	h.Logger.V(tlog.VL2).Info("Attempting to delete customer metafields", "id", customerID)
	return h.Client.DeleteMetafields(metafields)
}
