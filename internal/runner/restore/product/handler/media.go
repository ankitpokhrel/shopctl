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
}

func (h *Media) Handle() (any, error) {
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

	h.Logger.V(tlog.VL2).Info("Attempting to set product media", "id", media.ProductID)
	return updateProductMedia(&media, h.Client)
}

func updateProductMedia(media *api.ProductMediaData, client *api.GQLClient) (*api.ProductCreateResponse, error) {
	input := schema.ProductInput{
		ID: &media.ProductID,
	}

	createMediaInput := make([]schema.CreateMediaInput, 0, len(media.Media.Nodes))
	for _, m := range media.Media.Nodes {
		createMediaInput = append(createMediaInput, schema.CreateMediaInput{
			OriginalSource:   m.Preview.Image.URL,
			Alt:              m.Preview.Image.AltText,
			MediaContentType: m.MediaContentType,
		})
	}
	return client.UpdateProduct(input, createMediaInput)
}
