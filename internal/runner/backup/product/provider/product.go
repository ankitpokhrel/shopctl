package provider

import "github.com/ankitpokhrel/shopctl/schema"

type Product struct {
	Product *schema.Product
}

func (p *Product) Handle(_ any) (any, error) {
	return p.Product, nil
}
