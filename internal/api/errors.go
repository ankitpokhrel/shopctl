package api

import (
	"strings"

	"github.com/ankitpokhrel/shopctl/schema"
)

// Error represents an error response from the Shopify API.
type Error struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Extensions struct {
		Value any `json:"value"`
	} `json:"extensions"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return clean(e.Message)
}

// Errors is a list of errors.
type Errors []Error

// Error implements the error interface.
func (e Errors) Error() string {
	errs := make([]string, 0, len(e))
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, ", ")
}

// UserErrors is a list of user errors.
type UserErrors []schema.UserError

// Err implements the error interface.
func (e UserErrors) Error() string {
	errs := make([]string, 0, len(e))
	for _, err := range e {
		errs = append(errs, clean(err.Message))
	}
	return strings.Join(errs, ", ")
}

// Extensions is the extensions returned by the Shopify API.
type Extensions struct {
	Cost QueryCost `json:"cost"`
}

// QueryCost is the cost of the query returned by the Shopify API.
type QueryCost struct {
	RequestedQueryCost float64        `json:"requestedQueryCost"`
	ActualQueryCost    float64        `json:"actualQueryCost"`
	ThrottleStatus     ThrottleStatus `json:"throttleStatus"`
}

// ThrottleStatus is the status of the throttle returned by the Shopify API.
type ThrottleStatus struct {
	MaximumAvailable   float64 `json:"maximumAvailable"`
	CurrentlyAvailable float64 `json:"currentlyAvailable"`
	RestoreRate        float64 `json:"restoreRate"`
}

func clean(input string) string {
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	return input
}
