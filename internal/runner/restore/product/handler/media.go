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

type Media struct {
	Client  *api.GQLClient
	Logger  *tlog.Logger
	File    registry.File
	Summary *runner.Summary
	DryRun  bool
}

func (h *Media) Handle(data any) (any, error) {
	var realProductID string
	if id, ok := data.(string); ok {
		realProductID = id
	} else {
		return nil, fmt.Errorf("unable to figure out real product ID")
	}
	mediaRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}
	h.Summary.Count += 1

	var media api.ProductMediaData
	if err = json.Unmarshal(mediaRaw, &media); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		h.Summary.Failed += 1
		return nil, err
	}

	toAdd := make([]*api.ProductMediaNode, 0)
	toDelete := make([]string, 0)

	// Get upstream medias.
	currentMedia, _ := h.Client.GetProductMedias(realProductID)
	currentMediaMap := make(map[string]*api.ProductMediaNode, 0)
	if currentMedia != nil {
		for _, m := range currentMedia.Data.Product.Media.Nodes {
			currentMediaMap[m.ID] = &m
		}

		backupMediaMap := make(map[string]*api.ProductMediaNode, len(media.Media.Nodes))
		for _, m := range media.Media.Nodes {
			backupMediaMap[m.ID] = &m
		}

		for id := range currentMediaMap {
			if _, ok := backupMediaMap[id]; !ok {
				toDelete = append(toDelete, id)
			}
		}
		for id, m := range backupMediaMap {
			if _, ok := currentMediaMap[id]; !ok {
				toAdd = append(toAdd, m)
			}
		}
	} else {
		for _, m := range media.Media.Nodes {
			toAdd = append(toAdd, &m)
		}
	}

	attemptSync := func(pid string) error {
		if _, err := h.handleProductMediaDelete(pid, toDelete); err != nil {
			return err
		}
		if _, err := h.handleProductMediaAdd(pid, toAdd); err != nil {
			return err
		}
		return nil
	}

	h.Logger.V(tlog.VL2).Info("Attempting to attach product medias", "id", realProductID)
	if h.DryRun {
		h.Logger.V(tlog.VL2).Infof("Product media to sync - add: %d, remove: %d", len(toAdd), len(toDelete))
		h.Logger.V(tlog.VL3).Warn("Skipping product media sync")
		h.Summary.Passed += 1
		return nil, nil
	}
	err = attemptSync(realProductID)
	if err != nil {
		h.Logger.Error("Failed to sync product medias", "oldID", media.ProductID, "upstreamID", realProductID)
		h.Summary.Failed += 1
		return nil, err
	}
	h.Summary.Passed += 1
	return nil, nil
}

func (h Media) handleProductMediaAdd(productID string, toAdd []*api.ProductMediaNode) (*api.ProductCreateResponse, error) {
	input := schema.ProductInput{
		ID: &productID,
	}

	createMediaInput := make([]schema.CreateMediaInput, 0, len(toAdd))
	for _, m := range toAdd {
		createMediaInput = append(createMediaInput, schema.CreateMediaInput{
			OriginalSource:   m.Preview.Image.URL,
			Alt:              m.Preview.Image.AltText,
			MediaContentType: m.MediaContentType,
		})
	}
	return h.Client.UpdateProduct(input, createMediaInput)
}

func (h Media) handleProductMediaDelete(productID string, toDelete []string) (*api.FileUpdateResponse, error) {
	if len(toDelete) == 0 {
		return nil, nil
	}

	input := make([]schema.FileUpdateInput, 0, len(toDelete))
	for _, id := range toDelete {
		input = append(input, schema.FileUpdateInput{
			ID:                 id,
			ReferencesToRemove: []any{productID},
		})
	}
	h.Logger.V(tlog.VL2).Info("Attempting to detach product medias", "id", productID)
	return h.Client.DetachProductMedia(input)
}
