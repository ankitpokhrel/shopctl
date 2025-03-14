package handler

import (
	"encoding/json"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Customer struct {
	Client *api.GQLClient
	Logger *tlog.Logger
	File   registry.File
	DryRun bool
}

func (h *Customer) Handle(data any) (any, error) {
	customerRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	var customer schema.Customer
	if err = json.Unmarshal(customerRaw, &customer); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		return nil, err
	}

	if h.DryRun {
		h.Logger.V(tlog.VL3).Warn("Skipping customer sync")
		return &api.CustomerCreateResponse{}, nil
	}
	res, err := createOrUpdateCustomer(&customer, h.Client, h.Logger)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func createOrUpdateCustomer(customer *schema.Customer, client *api.GQLClient, lgr *tlog.Logger) (*api.CustomerCreateResponse, error) {
	res, err := client.CheckCustomerByEmailOrPhone(customer.Email, customer.Phone)
	if err != nil {
		return nil, err
	}

	var addresses []any
	for _, address := range customer.AddressesV2.Nodes {
		if addressMap, ok := address.(map[string]any); ok {
			data, err := json.Marshal(addressMap)
			if err != nil {
				continue
			}

			var a schema.MailingAddress
			if err := json.Unmarshal(data, &a); err != nil {
				continue
			}

			addresses = append(addresses, schema.MailingAddressInput{
				Address1:     a.Address1,
				Address2:     a.Address2,
				FirstName:    a.FirstName,
				LastName:     a.LastName,
				City:         a.City,
				CountryCode:  a.CountryCodeV2,
				Company:      a.Company,
				Phone:        a.Phone,
				ProvinceCode: a.ProvinceCode,
				Zip:          a.Zip,
			})
		}
	}

	input := schema.CustomerInput{
		FirstName:             customer.FirstName,
		LastName:              customer.LastName,
		Email:                 customer.Email,
		Phone:                 customer.Phone,
		Addresses:             addresses,
		Locale:                &customer.Locale,
		Note:                  customer.Note,
		Tags:                  customer.Tags,
		EmailMarketingConsent: nil, // We are not going to reset marketing consents.
		SmsMarketingConsent:   nil,
		TaxExempt:             &customer.TaxExempt,
		TaxExemptions:         customer.TaxExemptions,
	}

	if len(res.Data.Customers.Nodes) > 0 && res.Data.Customers.Nodes[0].ID != "" {
		input.ID = &res.Data.Customers.Nodes[0].ID

		lgr.Warn("Customer already exists, updating", "oldID", customer.ID, "upstreamID", *input.ID)
		return client.UpdateCustomer(input)
	}

	lgr.Info("Creating customer", "id", customer.ID)
	return client.CreateCustomer(input)
}
