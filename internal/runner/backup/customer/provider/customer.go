package provider

import "github.com/ankitpokhrel/shopctl/schema"

type Customer struct {
	Customer *schema.Customer
}

func (c *Customer) Handle() (any, error) {
	return c.Customer, nil
}
