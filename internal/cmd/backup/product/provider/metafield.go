package provider

import (
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

type MetaField struct {
	Client    *api.GQLClient
	Logger    *tlog.Logger
	ProductID string
}

func (m *MetaField) Handle() (any, error) {
	m.Logger.V(tlog.VL1).Infof("Product %s: processing meta fields", m.ProductID)

	metafields, err := m.Client.GetProductMetaFields(m.ProductID)
	if err != nil {
		m.Logger.Error("error when fetching metafield", "", m.ProductID, "error", err)
		return nil, err
	}
	return metafields.Data.Product, nil
}
