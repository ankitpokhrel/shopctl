package handler

import (
	"encoding/json"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Media struct {
	Client *api.GQLClient
	Logger *tlog.Logger
	File   registry.File
	DryRun bool
}

func (h *Media) Handle(data any) (any, error) {
	var realProductID string
	if id, ok := data.(string); ok {
		realProductID = id
	}
	mediaRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	var media api.ProductMediaData
	if err = json.Unmarshal(mediaRaw, &media); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	// Get upstream medias.
	currentMedias, err := h.Client.GetProductMedias(realProductID)
	if err != nil {
		return nil, err
	}

	currentMediasMap := make(map[string]*api.ProductMediaNode, len(currentMedias.Data.Product.Media.Nodes))
	for _, opt := range currentMedias.Data.Product.Media.Nodes {
		currentMediasMap[opt.ID] = &opt
	}

	backupMediasMap := make(map[string]*api.ProductMediaNode, len(media.Media.Nodes))
	for _, opt := range media.Media.Nodes {
		backupMediasMap[opt.ID] = &opt
	}

	toAdd := make([]*api.ProductMediaNode, 0)
	toDelete := make([]string, 0)

	for id := range currentMediasMap {
		if _, ok := backupMediasMap[id]; !ok {
			toDelete = append(toDelete, id)
		}
	}
	for id, m := range backupMediasMap {
		if _, ok := currentMediasMap[id]; !ok {
			toAdd = append(toAdd, m)
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
		return &api.ProductCreateResponse{}, nil
	}
	err = attemptSync(realProductID)
	if err != nil {
		h.Logger.Error("Failed to sync product medias", "oldID", media.ProductID, "upstreamID", realProductID)
	}
	return nil, err
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

func (h Media) handleProductMediaDelete(productID string, toDelete []string) (*api.ProductMediaDeleteResponse, error) {
	if len(toDelete) == 0 {
		return nil, nil
	}
	h.Logger.V(tlog.VL2).Info("Attempting to detach product medias", "id", productID)
	return h.Client.DeleteProductMedias(productID, toDelete)
}
