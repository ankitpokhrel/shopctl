package provider

import (
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

type Variant struct {
	Client    *api.GQLClient
	Logger    *tlog.Logger
	ProductID string
}

func (v *Variant) Handle(_ any) (any, error) {
	v.Logger.Infof("Product %s: processing variants", v.ProductID)

	variants, err := v.Client.GetProductVariants(v.ProductID)
	if err != nil {
		v.Logger.Error("Error when fetching variants", "productId", v.ProductID, "error", err)
		return nil, err
	}
	return variants, nil
}
