package fmtout

import (
	"encoding/csv"
	"io"
	"slices"
	"strings"
)

// CSVFormatter constructs formatter for csv.
type CSVFormatter struct {
	noHeaders bool
	columns   []string
	contents  [][]string
}

// CSVOption is a functional opt for CSVFormatter.
type CSVOption func(*CSVFormatter)

// NewCSV builds a new csv formatter.
func NewCSV(cols []string, rows [][]string, opts ...CSVOption) *CSVFormatter {
	var (
		headers  []string
		keepCols []int
	)

	csvfmt := CSVFormatter{noHeaders: true}
	for _, o := range opts {
		o(&csvfmt)
	}

	for i, c := range cols {
		if slices.Contains(csvfmt.columns, keyme(c)) {
			headers = append(headers, c)
			keepCols = append(keepCols, i)
		}
	}

	contents := make([][]string, len(rows))
	for i, row := range rows {
		var newRow []string
		for _, j := range keepCols {
			if j < len(row) {
				newRow = append(newRow, row[j])
			}
		}
		contents[i] = newRow
	}
	if !csvfmt.noHeaders {
		contents = append([][]string{headers}, contents...)
	}
	csvfmt.contents = contents

	return &csvfmt
}

// WithNoHeaders sets noHeaders option.
func WithNoHeaders(ok bool) CSVOption {
	return func(t *CSVFormatter) {
		t.noHeaders = ok
	}
}

// WithColumns sets columns to display.
func WithColumns(cols []string) CSVOption {
	return func(t *CSVFormatter) {
		t.columns = cols
	}
}

// Format formats the csv output.
func (t *CSVFormatter) Format(w io.Writer) error {
	wrt := csv.NewWriter(w)
	return wrt.WriteAll(t.contents)
}

func keyme(s string) string {
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, " ", "_")
}
