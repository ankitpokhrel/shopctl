//nolint:mnd
package list

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"golang.design/x/clipboard"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/pkg/fmtout"
	"github.com/ankitpokhrel/shopctl/pkg/search"
	"github.com/ankitpokhrel/shopctl/pkg/tui/table"
)

const (
	helpText = `List orders in a store.`

	examples = `$ shopctl order list

# List order by order number/name
$ shopctl order list --name 1003

# Search for orders with specific text anywhere in the order
$ shopctl order list "text in title or description" --limit 20

# List orders created in 2025 and has 'on-sale' tags but no 'summer' tag
$ shopctl order list --created ">=2025-01-01" --tags on-sale,-summer

# List orders processed in today
$ shopctl order list --processed ">=$(date +%Y-%m-%d)"

# List fulfilled orders with pending payments
$ shopctl order list --fulfillment-status fulfilled --payment-status pending

# List orders as a csv
$ shopctl order list --csv

# List orders in a plain table view without headers
$ shopctl order list --plain --no-headers

# Only list some columns of the in a plain table view
$ shopctl order list --plain --columns id,name,total

# List all orders processed in May 2025 using a raw query
# See https://shopify.dev/docs/api/usage/search-syntax
$ shopctl order list "(processed_at:>=2025-05-01 AND processed_at:<=2025-05-31)"`
)

type flag struct {
	searchText        string
	idRange           string
	name              string
	email             *string
	tags              []string
	sku               *string
	status            string
	gateway           string
	paymentID         string
	paymentStatus     string
	fulfillmentStatus string
	processed         string
	created           string
	updated           string
	limit             int16
	plain             bool
	csv               bool
	noHeaders         bool
	columns           []string
	printQuery        bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		f.searchText = args[0]
	}

	isset := func(field string) bool {
		fl := cmd.Flags().Lookup(field)
		return fl != nil && fl.Changed
	}

	idRange, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)

	name, err := cmd.Flags().GetString("name")
	cmdutil.ExitOnErr(err)

	if isset("email") {
		email, err := cmd.Flags().GetString("email")
		cmdutil.ExitOnErr(err)

		f.email = &email
	}

	if isset("sku") {
		sku, err := cmd.Flags().GetString("sku")
		cmdutil.ExitOnErr(err)

		f.sku = &sku
	}

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	status, err := cmd.Flags().GetString("status")
	cmdutil.ExitOnErr(err)

	gateway, err := cmd.Flags().GetString("gateway")
	cmdutil.ExitOnErr(err)

	paymentID, err := cmd.Flags().GetString("payment-id")
	cmdutil.ExitOnErr(err)

	paymentStatus, err := cmd.Flags().GetString("payment-status")
	cmdutil.ExitOnErr(err)

	fulfillmentStatus, err := cmd.Flags().GetString("fulfillment-status")
	cmdutil.ExitOnErr(err)

	processed, err := cmd.Flags().GetString("processed")
	cmdutil.ExitOnErr(err)

	created, err := cmd.Flags().GetString("created")
	cmdutil.ExitOnErr(err)

	updated, err := cmd.Flags().GetString("updated")
	cmdutil.ExitOnErr(err)

	limit, err := cmd.Flags().GetInt16("limit")
	cmdutil.ExitOnErr(err)

	plain, err := cmd.Flags().GetBool("plain")
	cmdutil.ExitOnErr(err)

	csv, err := cmd.Flags().GetBool("csv")
	cmdutil.ExitOnErr(err)

	noHeaders, err := cmd.Flags().GetBool("no-headers")
	cmdutil.ExitOnErr(err)

	columns, err := cmd.Flags().GetString("columns")
	cmdutil.ExitOnErr(err)

	printQuery, err := cmd.Flags().GetBool("print-query")
	cmdutil.ExitOnErr(err)

	f.idRange = idRange
	f.name = name
	f.tags = func() []string {
		if tags != "" {
			return strings.Split(tags, ",")
		}
		return []string{}
	}()
	f.status = status
	f.gateway = gateway
	f.paymentID = paymentID
	f.paymentStatus = paymentStatus
	f.fulfillmentStatus = fulfillmentStatus
	f.processed = processed
	f.created = created
	f.updated = updated
	f.limit = min(limit, 250)
	f.plain = plain
	f.csv = csv
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
	f.printQuery = printQuery
}

// NewCmdList constructs a new order list command.
func NewCmdList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list [QUERY]",
		Short:   "List orders in a store",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}

	cmd.Flags().String("id", "", "Filter by id range")
	cmd.Flags().StringP("name", "n", "", "Filter by name")
	cmd.Flags().StringP("email", "e", "", "Filter by email")
	cmd.Flags().String("tags", "", "Filter by tags (comma separated)")
	cmd.Flags().String("sku", "", "Filter by variant sku")
	cmd.Flags().String("status", "", "Filter by order status")
	cmd.Flags().String("gateway", "", "Filter by payment gateway")
	cmd.Flags().String("payment-id", "", "Filter by payment id associated with the order")
	cmd.Flags().StringP("payment-status", "P", "", "Filter by payment status of the order")
	cmd.Flags().StringP("fulfillment-status", "F", "", "Filter by fulfillment status of the order")
	cmd.Flags().String("processed", "", "Filter by the order processed date")
	cmd.Flags().String("created", "", "Filter by the created date")
	cmd.Flags().String("updated", "", "Filter by the updated date")
	cmd.Flags().Int16("limit", 50, "Number of entries to fetch (max 250)")
	cmd.Flags().Bool("plain", false, "Show output in properly formatted plain text")
	cmd.Flags().Bool("csv", false, "Print output in csv")
	cmd.Flags().Bool("no-headers", false, "Don't print table headers (works only with --plain)")
	cmd.Flags().String("columns", "", "Comma separated list of columns to print (works only with --plain)")
	cmd.Flags().Bool("print-query", false, "Print parsed raw Shopify search query")

	cmd.Flags().SortFlags = false

	return &cmd
}

//nolint:gocyclo
func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	query := buildSearchQuery(flag)

	if flag.printQuery {
		if query != nil && *query != "()" {
			fmt.Printf("%s", *query)
		}
		return nil
	}

	orders, err := client.GetOrders(int(flag.limit), nil, query)
	if err != nil {
		return err
	}

	cols := []table.Column{
		{Title: "ID", Width: 15},
		{Title: "Name", Width: 15},
		{Title: "Date", Width: 25},
		{Title: "Customer", Width: 25},
		{Title: "Total", Width: 15},
		{Title: "Payment", Width: 15},
		{Title: "Fulfillment", Width: 15},
		{Title: "Country", Width: 25},
		{Title: "Tags", Width: 25},
		{Title: "Note", Width: 40},
		{Title: "Processed", Width: 25},
	}

	getStr := func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}

	rows := make([]table.Row, 0)
	for _, o := range orders {
		tags := make([]string, 0, len(o.Tags))
		for _, t := range o.Tags {
			tags = append(tags, t.(string))
		}
		note := ""
		if o.Note != nil {
			note = *o.Note
		}
		name := ""
		if o.Customer != nil && o.Customer.FirstName != nil {
			name = getStr(o.Customer.FirstName)
		}
		finStatus := ""
		if o.DisplayFinancialStatus != nil {
			finStatus = string(*o.DisplayFinancialStatus)
		}
		country := ""
		if o.ShippingAddress != nil {
			country = getStr(o.ShippingAddress.Country)
		}
		row := table.Row{
			shopctl.ExtractNumericID(o.ID),
			o.Name,
			cmdutil.FormatDateTime(o.CreatedAt, ""),
			name,
			fmt.Sprintf("%.2f", o.TotalPriceSet.ShopMoney.Amount),
			finStatus,
			string(o.DisplayFulfillmentStatus),
			country,
			strings.Join(tags, ","),
			note,
			cmdutil.FormatDateTime(o.ProcessedAt, ""),
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		cmdutil.Warn("No orders found for the given criteria")
		os.Exit(0)
	}

	if flag.plain || flag.csv {
		defaultCols := []string{"id", "name", "date", "customer", "total", "payment", "country"}

		if len(flag.columns) == 0 {
			flag.columns = defaultCols
		}
	}

	if flag.plain {
		tbl := table.NewStaticTable(
			cols, rows,
			table.WithNoHeaders(flag.noHeaders),
			table.WithTableColumns(flag.columns),
		)
		return tbl.Render()
	}
	if flag.csv {
		csvfmt := fmtout.NewCSV(
			table.ColsToString(cols),
			table.RowsToString(rows),
			fmtout.WithNoHeaders(flag.noHeaders),
			fmtout.WithColumns(flag.columns),
		)
		return csvfmt.Format(os.Stdout)
	}

	helpTexts := []string{
		"↑ k/j ↓: Navigate top & down",
		"← h/l →: Navigate left & right",
		"m: Toggle distraction free mode",
		"c/C: Copy numeric or full order ID",
		"q/CTRL+c/ESC: Quit",
	}
	footerTexts := []string{
		fmt.Sprintf("Showing %d results for store %q", len(rows), ctx.Store),
	}
	if query != nil && *query != "" && *query != "()" {
		footerTexts = append(footerTexts, fmt.Sprintf("Query: %s", *query))
	}

	tbl := table.NewInteractiveTable(
		cols, rows,
		table.WithHelpTexts(helpTexts),
		table.WithFooterTexts(footerTexts),
		table.WithEnterFunc(func(id string) error {
			url := fmt.Sprintf("http://admin.shopify.com/store/%s/orders/%s", ctx.Alias, id)
			return browser.Browse(url)
		}),
		table.WithCopyFunc(func(id string, key string) error {
			if key == "C" {
				id = shopctl.ShopifyOrderID(id)
			}
			if err := clipboard.Init(); err == nil {
				_ = clipboard.Write(clipboard.FmtText, []byte(id))
			}
			return nil
		}),
	)
	return tbl.Render()
}

//nolint:gocyclo
func buildSearchQuery(f *flag) *string {
	q := search.New()

	q.Group(func(sub *search.Query) {
		if f.searchText != "" {
			sub.Add(f.searchText)
		}
		if f.idRange != "" {
			sub.Eq("id", f.idRange)
		}
		if f.name != "" {
			sub.Eq("name", f.name)
		}
		if f.email != nil {
			sub.Eq("email", *f.email)
		}
		if f.sku != nil {
			sub.Eq("sku", *f.sku)
		}
		if f.gateway != "" {
			sub.Eq("gateway", f.gateway)
		}
		if f.paymentID != "" {
			sub.Eq("payment_id", f.paymentID)
		}
		if f.paymentStatus != "" {
			sub.Eq("financial_status", f.paymentStatus)
		}
		if f.fulfillmentStatus != "" {
			sub.Eq("fulfillment_status", f.fulfillmentStatus)
		}
		if len(f.tags) > 0 {
			var (
				inc []string
				exc []string
			)
			for _, tag := range f.tags {
				if strings.HasPrefix(tag, "-") {
					exc = append(exc, tag[1:])
				} else {
					inc = append(inc, tag)
				}
			}
			if len(inc) > 0 {
				sub.InAnd("tag", inc...)
			}
			if len(exc) > 0 {
				sub.In("-tag", exc...)
			}
		}
		if f.processed != "" {
			sub.Eq("processed_at", f.processed)
		}
		if f.created != "" {
			sub.Eq("created_at", f.created)
		}
		if f.updated != "" {
			sub.Eq("updated_at", f.updated)
		}
	})
	query := q.Build()
	return &query
}

func validColumns() []string {
	return []string{
		"id",
		"date",
		"name",
		"total",
		"payment",
		"fulfillment",
		"country",
		"tags",
		"note",
		"processed",
	}
}
