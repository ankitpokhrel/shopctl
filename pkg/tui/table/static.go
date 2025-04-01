package table

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// StaticTable constructs static table object.
type StaticTable struct {
	table     *table.Table
	colWidths []int
	noHeaders bool
	columns   []string
}

// StaticTableOption is a functional opt for StaticTable.
type StaticTableOption func(*StaticTable)

// NewStaticTable builds a new static table.
func NewStaticTable(cols []Column, rows []Row, opts ...StaticTableOption) *StaticTable {
	var (
		headers  []string
		widths   []int
		keepCols []int
	)

	t := StaticTable{noHeaders: true}
	for _, o := range opts {
		o(&t)
	}

	for i, c := range cols {
		if slices.Contains(t.columns, keyme(c.Title)) {
			headers = append(headers, c.Title)
			widths = append(widths, c.Width)
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

	tbl := table.New().Rows(contents...)
	if !t.noHeaders {
		tbl.Headers(headers...)
	}
	t.table = tbl
	t.colWidths = widths

	return &t
}

// WithNoHeaders sets noHeaders option.
func WithNoHeaders(ok bool) StaticTableOption {
	return func(t *StaticTable) {
		t.noHeaders = ok
	}
}

// WithTableColumns sets columns to display.
func WithTableColumns(cols []string) StaticTableOption {
	return func(t *StaticTable) {
		t.columns = cols
	}
}

// Render renders the final table.
func (t *StaticTable) Render() error {
	gray := lipgloss.Color("245")
	headerStyle := lipgloss.NewStyle().Foreground(gray).Bold(true)
	cellStyle := lipgloss.NewStyle().Padding(0, 0)

	t.table.
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle.Width(t.colWidths[col])
		})

	_, err := fmt.Println(t.table)
	return err
}

func keyme(s string) string {
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, " ", "_")
}
