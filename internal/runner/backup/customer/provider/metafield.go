package provider

import (
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

type MetaField struct {
	Client     *api.GQLClient
	Logger     *tlog.Logger
	CustomerID string
}

func (m *MetaField) Handle() (any, error) {
	m.Logger.Infof("Customer %s: processing meta fields", m.CustomerID)

	metafields, err := m.Client.GetCustomerMetaFields(m.CustomerID)
	if err != nil {
		m.Logger.Error("Error when fetching metafield", "customerID", m.CustomerID, "error", err)
		return nil, err
	}
	return metafields.Data.Customer, nil
}
