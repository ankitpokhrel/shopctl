//nolint:mnd
package list

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/pkg/tui/table"
	"github.com/ankitpokhrel/shopctl/schema"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

const (
	helpText = `List products in a store.`

	examples = `$ shopctl product variant list

# List variants created before April but updated after April
$ shopctl product list --sku ABC --created "<2025-04-01" --updated ">2025-04-01"

# Search for variants with price > 200 and regular price <= 200
$ shopctl product variant list --price ">200" --regular-price "<=200"`
)

type flag struct {
	productID    string
	price        *string
	unitCost     *string
	regularPrice *string
	barcode      *string
	sku          *string
	created      string
	updated      string
	limit        int16
	plain        bool
	noHeaders    bool
	columns      []string
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	isset := func(field string) bool {
		fl := cmd.Flags().Lookup(field)
		return fl != nil && fl.Changed
	}

	if isset("price") {
		price, err := cmd.Flags().GetString("price")
		cmdutil.ExitOnErr(err)

		f.price = &price
	}
	if isset("unit-cost") {
		unitCost, err := cmd.Flags().GetString("unit-cost")
		cmdutil.ExitOnErr(err)

		f.unitCost = &unitCost
	}
	if isset("regular-price") {
		regularPrice, err := cmd.Flags().GetString("regular-price")
		cmdutil.ExitOnErr(err)

		f.regularPrice = &regularPrice
	}
	if isset("barcode") {
		barcode, err := cmd.Flags().GetString("barcode")
		cmdutil.ExitOnErr(err)

		f.barcode = &barcode
	}
	if isset("sku") {
		sku, err := cmd.Flags().GetString("sku")
		cmdutil.ExitOnErr(err)

		f.sku = &sku
	}

	created, err := cmd.Flags().GetString("created")
	cmdutil.ExitOnErr(err)

	updated, err := cmd.Flags().GetString("updated")
	cmdutil.ExitOnErr(err)

	limit, err := cmd.Flags().GetInt16("limit")
	cmdutil.ExitOnErr(err)

	plain, err := cmd.Flags().GetBool("plain")
	cmdutil.ExitOnErr(err)

	noHeaders, err := cmd.Flags().GetBool("no-headers")
	cmdutil.ExitOnErr(err)

	columns, err := cmd.Flags().GetString("columns")
	cmdutil.ExitOnErr(err)

	f.productID = shopctl.ShopifyProductID(args[0])
	f.created = created
	f.updated = updated
	f.limit = min(limit, 250)
	f.plain = plain
	f.noHeaders = noHeaders
	f.columns = func() []string {
		if columns != "" {
			return strings.Split(columns, ",")
		}
		return []string{}
	}()
	for _, c := range f.columns {
		if !slices.Contains(validColumns(), c) {
			cmdutil.Fail("Error: column names should be one of: %s", strings.Join(validColumns(), ", "))
			os.Exit(1)
		}
	}
}

// NewCmdList constructs a new product list command.
func NewCmdList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list [QUERY]",
		Short:   "List products in a store",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}

	cmd.Flags().String("sku", "", "Filter by sku")
	cmd.Flags().String("barcode", "", "Filter by barcode")
	cmd.Flags().String("price", "", "Filter by price")
	cmd.Flags().String("unit-cost", "", "Filter by unit cost")
	cmd.Flags().String("regular-price", "", "Filter by regular/compare-at price")
	cmd.Flags().String("created", "", "Filter by created date")
	cmd.Flags().String("updated", "", "Filter by updated date")
	cmd.Flags().Int16("limit", 50, "Number of entries to fetch")
	cmd.Flags().Bool("plain", false, "Show output in plain text instead of TUI")
	cmd.Flags().Bool("no-headers", false, "Don't print table headers (works only with --plain)")
	cmd.Flags().String("columns", "", "Comma separated list of columns to print (works only with --plain)")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	product, err := client.GetProductVariants(flag.productID)
	if err != nil {
		return err
	}

	cols := []table.Column{
		{Title: "ID", Width: 15},
		{Title: "Title", Width: 50},
		{Title: "SKU", Width: 25},
		{Title: "Price", Width: 20},
		{Title: "Unit Cost", Width: 20},
		{Title: "Regular Price", Width: 20},
		{Title: "Barcode", Width: 25},
		{Title: "Created", Width: 25},
		{Title: "Updated", Width: 25},
	}

	rows := make([]table.Row, 0)
	for _, v := range product.Variants.Nodes {
		if !fulfillsCriteria(v, flag) {
			continue
		}
		unitCost := ""
		if v.InventoryItem != nil && v.InventoryItem.UnitCost != nil {
			unitCost = fmt.Sprintf("%.2f", v.InventoryItem.UnitCost.Amount)
		}
		regularPrice := ""
		if v.InventoryItem != nil && v.InventoryItem.Variant != nil && v.InventoryItem.Variant.CompareAtPrice != nil {
			regularPrice = *v.InventoryItem.Variant.CompareAtPrice
		}
		sku := ""
		if v.Sku != nil {
			sku = *v.Sku
		}
		barcode := ""
		if v.Barcode != nil {
			barcode = *v.Barcode
		}
		rows = append(rows, table.Row{
			shopctl.ExtractNumericID(v.ID),
			v.Title,
			sku,
			v.Price,
			unitCost,
			regularPrice,
			barcode,
			cmdutil.FormatDateTime(v.CreatedAt, ""),
			cmdutil.FormatDateTime(v.UpdatedAt, ""),
		})
	}

	if len(rows) == 0 {
		cmdutil.Warn("No variants found for the given criteria")
		os.Exit(0)
	}

	if flag.plain {
		if len(flag.columns) == 0 {
			// Default columns in plain mode.
			flag.columns = []string{"id", "title", "sku", "price", "created"}
		}
		tbl := table.NewStaticTable(
			cols, rows,
			table.WithNoHeaders(flag.noHeaders),
			table.WithTableColumns(flag.columns),
		)
		return tbl.Render()
	}

	helpTexts := []string{
		"↑ k/j ↓: Navigate top & down",
		"← h/l →: Navigate left & right",
		"m: Toggle distraction free mode",
		"c/C: Copy numeric or full product ID",
		"q/CTRL+c/ESC: Quit",
	}
	footerTexts := []string{
		fmt.Sprintf("Showing %d results for product %q in store %q", len(rows), flag.productID, ctx.Store),
	}

	tbl := table.NewInteractiveTable(
		cols, rows,
		table.WithHelpTexts(helpTexts),
		table.WithFooterTexts(footerTexts),
		table.WithEnterFunc(func(id string) error {
			url := fmt.Sprintf(
				"http://admin.shopify.com/store/%s/products/%s/variants/%s",
				ctx.Alias,
				shopctl.ExtractNumericID(flag.productID),
				id,
			)
			return browser.Browse(url)
		}),
		table.WithCopyFunc(func(id string, key string) error {
			if key == "C" {
				id = shopctl.ShopifyProductVariantID(id)
			}
			if err := clipboard.Init(); err == nil {
				_ = clipboard.Write(clipboard.FmtText, []byte(id))
			}
			return nil
		}),
	)
	return tbl.Render()
}

func fulfillsCriteria(v schema.ProductVariant, f *flag) bool {
	matches := make([]bool, 0)

	if f.sku != nil {
		matches = append(matches, compareString(v.Sku, f.sku))
	}
	if f.barcode != nil {
		matches = append(matches, compareString(v.Barcode, f.barcode))
	}
	if f.price != nil {
		matches = append(matches, compareFloat(*f.price, v.Price))
	}
	if f.unitCost != nil {
		if v.InventoryItem != nil && v.InventoryItem.UnitCost != nil {
			matches = append(matches, compareFloat(*f.unitCost, fmt.Sprintf("%f", v.InventoryItem.UnitCost.Amount)))
		} else {
			matches = append(matches, false)
		}
	}
	if f.regularPrice != nil {
		if v.InventoryItem != nil && v.InventoryItem.Variant != nil && v.InventoryItem.Variant.CompareAtPrice != nil {
			matches = append(matches, compareFloat(*f.regularPrice, *v.InventoryItem.Variant.CompareAtPrice))
		} else {
			matches = append(matches, false)
		}
	}
	if f.created != "" {
		matches = append(matches, compareDatetime(f.created, v.CreatedAt))
	}
	if f.updated != "" {
		matches = append(matches, compareDatetime(f.updated, v.UpdatedAt))
	}

	for _, b := range matches {
		if !b {
			return false
		}
	}
	return true
}

func compareString(value *string, filter *string) bool {
	if value == nil {
		return false
	}
	if strings.HasPrefix(*filter, "-") {
		return *value != (*filter)[1:]
	}
	return *value == *filter
}

func compareFloat(query string, val string) bool {
	query = strings.TrimSpace(query)
	query = strings.TrimPrefix(query, ":")

	var op, opr string

	switch {
	case strings.HasPrefix(query, ">=") || strings.HasPrefix(query, "<="):
		op = query[:2]
		opr = query[2:]
	case strings.HasPrefix(query, ">") || strings.HasPrefix(query, "<") || strings.HasPrefix(query, "="):
		op = query[:1]
		opr = query[1:]
	default:
		op = "="
		opr = query
	}

	operand, err := strconv.ParseFloat(strings.TrimSpace(opr), 64)
	if err != nil {
		return false
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil {
		return false
	}

	switch op {
	case "=":
		return value == operand
	case ">":
		return value > operand
	case ">=":
		return value >= operand
	case "<":
		return value < operand
	case "<=":
		return value <= operand
	default:
		return false
	}
}

func compareDatetime(query string, val string) bool {
	query = strings.TrimSpace(query)
	query = strings.TrimPrefix(query, ":")

	var op, opr string
	switch {
	case strings.HasPrefix(query, ">=") || strings.HasPrefix(query, "<="):
		op = query[:2]
		opr = query[2:]
	case strings.HasPrefix(query, ">") || strings.HasPrefix(query, "<") || strings.HasPrefix(query, "="):
		op = query[:1]
		opr = query[1:]
	default:
		op = "="
		opr = query
	}

	operand, operandHasTime, err := parseTimeInfo(strings.TrimSpace(opr))
	if err != nil {
		return false
	}
	value, _, err := parseTimeInfo(strings.TrimSpace(val))
	if err != nil {
		return false
	}

	// Ignore time parts if only date is provided.
	if !operandHasTime {
		operand = time.Date(operand.Year(), operand.Month(), operand.Day(), 0, 0, 0, 0, operand.Location())
		value = time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	}

	switch op {
	case "=":
		return value.Equal(operand)
	case ">":
		return value.After(operand)
	case ">=":
		return value.After(operand) || value.Equal(operand)
	case "<":
		return value.Before(operand)
	case "<=":
		return value.Before(operand) || value.Equal(operand)
	default:
		return false
	}
}

// parseTimeInfo attempts to parse a datetime string using multiple layouts.
// It returns the parsed time, a flag indicating whether the time component was
// provided (true if the layout includes a time), and an error if parsing fails.
func parseTimeInfo(s string) (time.Time, bool, error) {
	s = strings.TrimSpace(s)
	layouts := []struct {
		layout  string
		hasTime bool
	}{
		{time.RFC3339, true},
		{"2006-01-02", false},
		{"2006-01-02 15:04:05", true},
		{"2006/01/02", false},
		{"2006/01/02 15:04:05", true},
	}

	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.Parse(l.layout, s)
		if err == nil {
			return t, l.hasTime, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("unable to parse time: %s", s)
}

func validColumns() []string {
	return []string{
		"id",
		"title",
		"sku",
		"barcode",
		"price",
		"unit_cost",
		"regular_price",
		"created",
		"updated",
	}
}
