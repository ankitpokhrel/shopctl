package provider

import (
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

type Media struct {
	Client    *api.GQLClient
	Logger    *tlog.Logger
	ProductID string
}

func (m *Media) Handle() (any, error) {
	m.Logger.Infof("Product %s: processing media items", m.ProductID)

	medias, err := m.Client.GetProductMedias(m.ProductID)
	if err != nil {
		m.Logger.Error("error when fetching media", "", m.ProductID, "error", err)
		return nil, err
	}
	return medias.Data.Product, nil
}
