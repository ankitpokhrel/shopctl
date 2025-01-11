package api

import "strings"

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
	return e.Message
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
