//nolint:mnd
package list

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/pkg/fmtout"
	"github.com/ankitpokhrel/shopctl/pkg/search"
	"github.com/ankitpokhrel/shopctl/pkg/tui/table"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

const (
	helpText = `List customers in a store.`

	examples = `$ shopctl customer list

# Search for customers with specific case-insensitive text
$ shopctl customer list "text from multiple fields" --limit 20

# List customers with atleast 1 order
$ shopctl customer list --orders-count ">0"

# List customers from Germany with who spent min 100 and agreed to receive marketing email
$ shopctl customer list --total-spent ">=100" --country Germany --accepts-marketing

# List customers as a csv
$ shopctl customer list --csv

# List customers in a plain table view without headers
$ shopctl customer list --plain --no-headers

# List customers using raw query
# See https://shopify.dev/docs/api/usage/search-syntax
$ shopctl customer list "first_name:Jane orders_count:>1 country:Germany"`
)

type flag struct {
	searchText             string
	firstName              *string
	lastName               *string
	email                  *string
	phone                  *string
	country                *string
	state                  *string
	tags                   []string
	totalSpent             *string
	acceptsMarketing       *bool
	ordersCount            *string
	orderDate              string
	lastAbandonedOrderDate string
	created                string
	updated                string
	limit                  int16
	plain                  bool
	csv                    bool
	noHeaders              bool
	columns                []string
	printQuery             bool
	withSensitiveData      bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		f.searchText = args[0]
	}

	isset := func(field string) bool {
		fl := cmd.Flags().Lookup(field)
		return fl != nil && fl.Changed
	}

	if isset("first-name") {
		fname, err := cmd.Flags().GetString("first-name")
		cmdutil.ExitOnErr(err)

		f.firstName = &fname
	}
	if isset("last-name") {
		lname, err := cmd.Flags().GetString("last-name")
		cmdutil.ExitOnErr(err)

		f.firstName = &lname
	}
	if isset("email") {
		email, err := cmd.Flags().GetString("email")
		cmdutil.ExitOnErr(err)

		f.email = &email
	}
	if isset("phone") {
		phone, err := cmd.Flags().GetString("phone")
		cmdutil.ExitOnErr(err)

		f.phone = &phone
	}
	if isset("country") {
		country, err := cmd.Flags().GetString("country")
		cmdutil.ExitOnErr(err)

		f.country = &country
	}
	if isset("state") {
		state, err := cmd.Flags().GetString("state")
		cmdutil.ExitOnErr(err)

		f.state = &state
	}
	if isset("orders-count") {
		ordersCount, err := cmd.Flags().GetString("orders-count")
		cmdutil.ExitOnErr(err)

		f.ordersCount = &ordersCount
	}
	if isset("total-spent") {
		totalSpent, err := cmd.Flags().GetString("total-spent")
		cmdutil.ExitOnErr(err)

		f.totalSpent = &totalSpent
	}
	if isset("accepts-marketing") {
		acceptsMarketing, err := cmd.Flags().GetBool("accepts-marketing")
		cmdutil.ExitOnErr(err)

		f.acceptsMarketing = &acceptsMarketing
	}

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	lastAbandonedOrderDate, err := cmd.Flags().GetString("last-abandoned-order-date")
	cmdutil.ExitOnErr(err)

	orderDate, err := cmd.Flags().GetString("order-date")
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

	withSensitiveData, err := cmd.Flags().GetBool("with-sensitive-data")
	cmdutil.ExitOnErr(err)

	f.tags = func() []string {
		if tags != "" {
			return strings.Split(tags, ",")
		}
		return []string{}
	}()
	f.orderDate = orderDate
	f.lastAbandonedOrderDate = lastAbandonedOrderDate
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
	f.withSensitiveData = withSensitiveData
}

// NewCmdList constructs a new customer list command.
func NewCmdList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list [QUERY]",
		Short:   "List customers in a store",
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

	cmd.Flags().StringP("first-name", "f", "", "Filter by first name")
	cmd.Flags().StringP("last-name", "l", "", "Filter by last name")
	cmd.Flags().StringP("email", "e", "", "Filter by email")
	cmd.Flags().StringP("phone", "p", "", "Filter by phone")
	cmd.Flags().String("country", "", "Filter by country name")
	cmd.Flags().String("state", "", "Filter by state/province (name or the two-letter code)")
	cmd.Flags().String("tags", "", "Filter by tags (comma separated)")
	cmd.Flags().String("total-spent", "", "Filter by money spent across all orders")
	cmd.Flags().Bool("accepts-marketing", false, "Filter by customers who has consented to receive marketing material")
	cmd.Flags().String("orders-count", "", "Filter by the total number of orders a customer has placed")
	cmd.Flags().String("order-date", "", "Filter by the date & time the order was placed by the customer")
	cmd.Flags().String("last-abandoned-order-date", "", "Filter by the customer's most recent abandoned checkout")
	cmd.Flags().String("created", "", "Filter by the created date")
	cmd.Flags().String("updated", "", "Filter by the updated date")
	cmd.Flags().Int16("limit", 50, "Number of entries to fetch (max 250)")
	cmd.Flags().Bool("plain", false, "Show output in properly formatted plain text")
	cmd.Flags().Bool("csv", false, "Print output in csv")
	cmd.Flags().Bool("no-headers", false, "Don't print table headers (works only with --plain)")
	cmd.Flags().String("columns", "", "Comma separated list of columns to print (works only with --plain)")
	cmd.Flags().Bool("print-query", false, "Print parsed raw Shopify search query")
	cmd.Flags().Bool("with-sensitive-data", false, "Include protected/sensitive fields like email and phone in tui")

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

	customers, err := client.GetCustomers(int(flag.limit), nil, query)
	if err != nil {
		return err
	}

	cols := []table.Column{
		{Title: "ID", Width: 15},
		{Title: "First Name", Width: 25},
	}
	if flag.withSensitiveData {
		cols = append(cols,
			[]table.Column{
				{Title: "Last Name", Width: 25},
				{Title: "Email", Width: 25},
				{Title: "Phone", Width: 25},
			}...,
		)
	}
	cols = append(cols,
		[]table.Column{
			{Title: "Country", Width: 25},
			{Title: "Tags", Width: 25},
			{Title: "Valid Email", Width: 15},
			{Title: "Created", Width: 25},
			{Title: "Updated", Width: 25},
		}...,
	)

	getStr := func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}

	rows := make([]table.Row, 0)
	for _, c := range customers {
		id := shopctl.ExtractNumericID(c.ID)
		tags := make([]string, 0, len(c.Tags))
		for _, t := range c.Tags {
			tags = append(tags, t.(string))
		}
		country := ""
		if c.DefaultAddress != nil {
			country = getStr(c.DefaultAddress.Country)
		}
		row := table.Row{
			id,
			getStr(c.FirstName),
		}
		if flag.withSensitiveData {
			row = append(row,
				getStr(c.LastName),
				getStr(c.Email),
				getStr(c.Phone),
			)
		}
		validEmail := "No"
		if c.ValidEmailAddress {
			validEmail = "Yes"
		}
		row = append(row,
			country,
			strings.Join(tags, ","),
			validEmail,
			cmdutil.FormatDateTime(c.CreatedAt, ""),
			cmdutil.FormatDateTime(c.UpdatedAt, ""),
		)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		cmdutil.Warn("No customers found for the given criteria")
		os.Exit(0)
	}

	if flag.plain || flag.csv {
		defaultCols := []string{"id", "first_name", "country", "valid_email", "created"}
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
		"c/C: Copy numeric or full customer ID",
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
			url := fmt.Sprintf("http://admin.shopify.com/store/%s/customers/%s", ctx.Alias, id)
			return browser.Browse(url)
		}),
		table.WithCopyFunc(func(id string, key string) error {
			if key == "C" {
				id = shopctl.ShopifyCustomerID(id)
			}
			if err := clipboard.Init(); err == nil {
				_ = clipboard.Write(clipboard.FmtText, []byte(id))
			}
			return nil
		}),
	)
	return tbl.Render()
}

// Filter by state and two character country code doesn't work even though the
// documentation says it does. Seem to be a bug on Shopify API.
//
// TODO: Check filter by state and 2 char country code later.
//
//nolint:gocyclo
func buildSearchQuery(f *flag) *string {
	q := search.New()

	q.Group(func(sub *search.Query) {
		if f.searchText != "" {
			sub.Add(f.searchText)
		}
		if f.firstName != nil {
			sub.Eq("first_name", *f.firstName)
		}
		if f.lastName != nil {
			sub.Eq("last_name", *f.lastName)
		}
		if f.email != nil {
			sub.Eq("email", *f.email)
		}
		if f.phone != nil {
			sub.Eq("phone", *f.phone)
		}
		if f.country != nil {
			sub.Eq("country", *f.country)
		}
		if f.state != nil {
			sub.Eq("state", *f.state)
		}
		if f.totalSpent != nil {
			sub.Eq("total_spent", *f.totalSpent)
		}
		if f.acceptsMarketing != nil {
			sub.Eq("accepts_marketing", fmt.Sprintf("%t", *f.acceptsMarketing))
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
		if f.ordersCount != nil {
			sub.Eq("orders_count", *f.ordersCount)
		}
		if f.orderDate != "" {
			sub.Eq("order_date", f.orderDate)
		}
		if f.lastAbandonedOrderDate != "" {
			sub.Eq("last_abandoned_order_date", f.lastAbandonedOrderDate)
		}
		if f.created != "" {
			sub.Eq("customer_date", f.created)
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
		"first_name",
		"last_name",
		"email",
		"phone",
		"country",
		"tags",
		"valid_email",
		"created",
		"updated",
	}
}
