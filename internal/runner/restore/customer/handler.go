package customer

import (
	"encoding/json"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type customerHandler struct {
	client *api.GQLClient
	logger *tlog.Logger
	file   registry.File
}

func (h *customerHandler) Handle() (any, error) {
	customer, err := registry.ReadFileContents(h.file.Path)
	if err != nil {
		h.logger.Error("Unable to read contents", "file", h.file.Path, "error", err)
		return nil, err
	}

	var cust schema.Customer
	if err = json.Unmarshal(customer, &cust); err != nil {
		h.logger.Error("Unable to marshal contents", "file", h.file.Path, "error", err)
		return nil, err
	}

	res, err := createOrUpdateCustomer(&cust, h.client, h.logger)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func createOrUpdateCustomer(cust *schema.Customer, client *api.GQLClient, lgr *tlog.Logger) (*api.CustomerCreateResponse, error) {
	res, err := client.CheckCustomerByID(cust.ID)
	if err != nil {
		return nil, err
	}

	// TODO: Compare and extract fields that are actually updated.

	if res.Data.Customer.ID != "" {
		input := schema.CustomerInput{
			ID:            &cust.ID,
			FirstName:     cust.FirstName,
			LastName:      cust.LastName,
			Email:         cust.Email,
			Phone:         cust.Phone,
			Note:          cust.Note,
			Tags:          cust.Tags,
			TaxExempt:     &cust.TaxExempt,
			TaxExemptions: cust.TaxExemptions,
		}
		lgr.Warn("Customer already exists, updating", "id", cust.ID)
		return client.UpdateCustomer(input)
	}

	input := schema.CustomerInput{
		FirstName:     cust.FirstName,
		LastName:      cust.LastName,
		Email:         cust.Email,
		Phone:         cust.Phone,
		Note:          cust.Note,
		Tags:          cust.Tags,
		TaxExempt:     &cust.TaxExempt,
		TaxExemptions: cust.TaxExemptions,
	}
	lgr.Info("Creating customer", "id", cust.ID)
	return client.CreateCustomer(input)
}
