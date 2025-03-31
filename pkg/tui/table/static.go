package table

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// StaticTable constructs static table object.
type StaticTable struct {
	table     *table.Table
	colWidths []int
	noHeaders bool
}

// StaticTableOption is a functional opt for StaticTable.
type StaticTableOption func(*StaticTable)

// NewStaticTable builds a new static table.
func NewStaticTable(cols []Column, rows []Row, opts ...StaticTableOption) *StaticTable {
	headers := make([]string, len(cols))
	widths := make([]int, len(cols))
	for i, c := range cols {
		headers[i] = c.Title
		widths[i] = c.Width
	}

	contents := make([][]string, len(rows))
	for i, row := range rows {
		for _, r := range row {
			contents[i] = append(contents[i], r)
		}
	}

	t := StaticTable{colWidths: widths, noHeaders: true}
	for _, o := range opts {
		o(&t)
	}

	tbl := table.New().Rows(contents...)
	if !t.noHeaders {
		tbl.Headers(headers...)
	}
	t.table = tbl

	return &t
}

// WithNoHeaders sets noHeaders option.
func WithNoHeaders(ok bool) StaticTableOption {
	return func(t *StaticTable) {
		t.noHeaders = ok
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
