package handler

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/internal/runner"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
)

type Customer struct {
	Client  *api.GQLClient
	Logger  *tlog.Logger
	File    registry.File
	Filter  *runner.RestoreFilter
	Summary *runner.Summary
	DryRun  bool
}

func (h *Customer) Handle(data any) (any, error) {
	customerRaw, err := registry.ReadFileContents(h.File.Path)
	if err != nil {
		h.Logger.Error("Unable to read contents", "file", h.File.Path, "error", err)
		return nil, err
	}
	h.Summary.Count += 1

	var customer schema.Customer
	if err = json.Unmarshal(customerRaw, &customer); err != nil {
		h.Logger.Error("Unable to marshal contents", "file", h.File.Path, "error", err)
		h.Summary.Failed += 1
		return nil, err
	}

	// Filter customer.
	if len(h.Filter.Filters) > 0 {
		matched, err := matchesFilters(&customer, h.Filter)
		if err != nil || !matched {
			h.Summary.Skipped += 1
			return nil, engine.ErrSkipChildren
		}
	}

	if h.DryRun {
		h.Logger.V(tlog.VL3).Warn("Skipping customer sync")
		h.Summary.Passed += 1
		return customer.ID, nil
	}
	res, err := createOrUpdateCustomer(&customer, h.Client, h.Logger)
	if err != nil {
		h.Summary.Failed += 1
		return nil, err
	}
	h.Summary.Passed += 1
	return res.Customer.ID, nil
}

func createOrUpdateCustomer(customer *schema.Customer, client *api.GQLClient, lgr *tlog.Logger) (*api.CustomerSyncResponse, error) {
	cust, _ := client.CheckCustomerByEmailOrPhoneOrID(
		customer.Email,
		customer.Phone,
		shopctl.ExtractNumericID(customer.ID),
	)

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

	if cust != nil && cust.ID != "" {
		input.ID = &cust.ID

		lgr.Warn("Customer already exists, updating", "oldID", customer.ID, "upstreamID", *input.ID)
		return client.UpdateCustomer(input)
	}

	lgr.Info("Creating customer", "id", customer.ID)
	return client.CreateCustomer(input)
}

//nolint:gocyclo
func matchesFilters(customer *schema.Customer, rf *runner.RestoreFilter) (bool, error) {
	containsAny := func(s []any, v any) bool { return slices.Contains(s, v) } //nolint:gocritic

	results := []bool{}
	for key, values := range rf.Filters {
		switch strings.ToLower(key) {
		case "id":
			matched := slices.Contains(values, shopctl.ExtractNumericID(customer.ID))
			results = append(results, matched)
		case "email":
			matched := false
			if customer.Email != nil {
				matched = slices.Contains(values, *customer.Email)
			}
			results = append(results, matched)
		case "phone":
			matched := slices.Contains(values, *customer.Phone)
			results = append(results, matched)
		case "firstname":
			matched := false
			if customer.FirstName != nil {
				matched = slices.Contains(values, *customer.FirstName)
			}
			results = append(results, matched)
		case "lastname":
			matched := false
			if customer.FirstName != nil {
				matched = slices.Contains(values, *customer.LastName)
			}
			results = append(results, matched)
		case "tags":
			matched := false
			for _, tag := range values {
				if containsAny(customer.Tags, tag) {
					matched = true
					break
				}
			}
			results = append(results, matched)
		case "state":
			matched := slices.Contains(values, strings.ToLower(string(customer.State)))
			results = append(results, matched)
		case "verifiedemail":
			matched := slices.Contains(values, "true")
			results = append(results, matched)
		default:
			return false, fmt.Errorf("unsupported filter key: %s", key)
		}
	}

	if len(results) == 0 {
		return false, nil
	}

	// Combine results using separators
	finalResult := results[0]
	for i, sep := range rf.Separators {
		switch strings.ToLower(sep) {
		case "and":
			finalResult = finalResult && results[i+1]
		case "or":
			finalResult = finalResult || results[i+1]
		default:
			return false, fmt.Errorf("unsupported separator: %s", sep)
		}
	}
	return finalResult, nil
}
