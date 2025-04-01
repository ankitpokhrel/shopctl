//nolint:mnd
package list

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/search"
	"github.com/ankitpokhrel/shopctl/pkg/tui/table"
	"github.com/spf13/cobra"
)

const (
	helpText = `List products in a store.`

	examples = `$ shopctl product list

# Search for products with specific text anywhere in the product
$ shopctl product list "text in title or description" --limit 20

# List products in status DRAFT and ARCHIVED that are created in 2025
$ shopctl product list -sDRAFT,ARCHIVED --created ">=2025-01-01"

# List products with tag 'on-sale' but without tag 'summer'
$ shopctl product list --tags on-sale,-summer

# Get gift cards in category Clothing published on 2025
$ shopctl product list --gift-card -yClothing --published ">=2025-01-01"

# List products in a plain table view without headers
$ shopctl product list --plain --no-headers

# Only list some columns of the in a plain table view
$ shopctl product list --plain --columns key,assignee,status

# Get products with empty sku and non-empty product type
$ shopctl product list --sku "" --type -

# List products using raw query
# See https://shopify.dev/docs/api/usage/search-syntax
$ shopctl product list "(title:Caramel Apple) OR (inventory_total:>500 inventory_total:<=1000)"`
)

// Flag wraps available command flags.
type flag struct {
	searchText  string
	productType *string
	categoryID  *string
	tags        []string
	vendor      *string
	price       *float64
	barcode     *string
	sku         *string
	giftCard    *bool
	status      []string
	created     string
	updated     string
	limit       int16
	plain       bool
	noHeaders   bool
	columns     []string
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		f.searchText = args[0]
	}

	isset := func(field string) bool {
		fl := cmd.Flags().Lookup(field)
		return fl != nil && fl.Changed
	}

	status, err := cmd.Flags().GetString("status")
	cmdutil.ExitOnErr(err)

	if isset("type") {
		productType, err := cmd.Flags().GetString("type")
		cmdutil.ExitOnErr(err)

		f.productType = &productType
	}
	if isset("category-id") {
		categoryID, err := cmd.Flags().GetString("category-id")
		cmdutil.ExitOnErr(err)

		f.categoryID = &categoryID
	}
	if isset("vendor") {
		vendor, err := cmd.Flags().GetString("vendor")
		cmdutil.ExitOnErr(err)

		f.vendor = &vendor
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
	if isset("gift-card") {
		giftCard, err := cmd.Flags().GetBool("gift-card")
		cmdutil.ExitOnErr(err)

		f.giftCard = &giftCard
	}
	if isset("price") {
		price, err := cmd.Flags().GetFloat64("price")
		cmdutil.ExitOnErr(err)

		f.price = &price
	}

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

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

	f.status = func() []string {
		if status != "" {
			return strings.Split(status, ",")
		}
		return []string{}
	}()
	f.tags = func() []string {
		if tags != "" {
			return strings.Split(tags, ",")
		}
		return []string{}
	}()
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
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}

	cmd.Flags().StringP("status", "s", "", "Filter products by status (ACTIVE, DRAFT or ARCHIVED)")
	cmd.Flags().String("created", "", "Filter by created date")
	cmd.Flags().String("updated", "", "Filter by updated date")
	cmd.Flags().String("type", "", "Filter by product type")
	cmd.Flags().StringP("category-id", "y", "", "Filter by category ID") // TODO: Check if we can use name instead of ID.
	cmd.Flags().String("tags", "", "Filter by tags (comma separated)")
	cmd.Flags().String("vendor", "", "Filter by vendor")
	cmd.Flags().Float64("price", 0, "Filter by variant price")
	cmd.Flags().String("barcode", "", "Filter by variant barcode")
	cmd.Flags().String("sku", "", "Filter by variant sku")
	cmd.Flags().Bool("gift-card", false, "Filter gift cards")
	cmd.Flags().Int16("limit", 50, "Number of entries to fetch")
	cmd.Flags().Bool("plain", false, "Show output in plain text instead of TUI")
	cmd.Flags().Bool("no-headers", false, "Don't print table headers (works only with --plain)")
	cmd.Flags().String("columns", "", "Comma separated list of columns to print (works only with --plain)")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string, _ *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	query := buildSearchQuery(flag)
	products, err := client.GetProducts(int(flag.limit), nil, query)
	if err != nil {
		return err
	}

	cols := []table.Column{
		{Title: "ID", Width: 15},
		{Title: "Title", Width: 50},
		{Title: "Options", Width: 25},
		{Title: "Product Type", Width: 20},
		{Title: "Category", Width: 25},
		{Title: "Tags", Width: 25},
		{Title: "Vendor", Width: 20},
		{Title: "Variants", Width: 8},
		{Title: "Media", Width: 8},
		{Title: "Status", Width: 10},
		{Title: "Created", Width: 25},
		{Title: "Updated", Width: 25},
	}

	rows := make([]table.Row, 0)
	for _, p := range products.Data.Products.Edges {
		id := cmdutil.ExtractNumericID(p.Node.ID)
		options := make([]string, 0, len(p.Node.Options))
		for _, o := range p.Node.Options {
			options = append(options, o.Name)
		}
		tags := make([]string, 0, len(p.Node.Tags))
		for _, t := range p.Node.Tags {
			tags = append(tags, t.(string))
		}
		category := ""
		if p.Node.Category != nil {
			category = p.Node.Category.Name
		}
		rows = append(rows, table.Row{
			id,
			p.Node.Title,
			strings.Join(options, ","),
			p.Node.ProductType,
			category,
			strings.Join(tags, ","),
			p.Node.Vendor,
			fmt.Sprintf("%d", p.Node.VariantsCount.Count),
			fmt.Sprintf("%d", p.Node.MediaCount.Count),
			string(p.Node.Status),
			formatDateTime(p.Node.CreatedAt, ""),
			formatDateTime(p.Node.UpdatedAt, ""),
		})
	}

	if len(rows) == 0 {
		cmdutil.Warn("No products found for the given criteria")
		os.Exit(0)
	}

	if flag.plain {
		if len(flag.columns) == 0 {
			// Default columns in plain mode.
			flag.columns = []string{"id", "title", "category", "created"}
		}
		tbl := table.NewStaticTable(
			cols, rows,
			table.WithNoHeaders(flag.noHeaders),
			table.WithTableColumns(flag.columns),
		)
		return tbl.Render()
	}
	tbl := table.NewInteractiveTable(cols, rows)
	return tbl.Render()
}

//nolint:gocyclo
func buildSearchQuery(f *flag) *string {
	q := search.New()

	// TODO: AND is missing
	q.Group(func(sub *search.Query) {
		if f.searchText != "" {
			sub.Add(f.searchText)
		}
		if len(f.status) > 0 {
			var (
				inc []string
				exc []string
			)
			for _, s := range f.status {
				if strings.HasPrefix(s, "-") {
					exc = append(exc, s[1:])
				} else {
					inc = append(inc, s)
				}
			}
			if len(inc) > 0 {
				sub.In("status", inc...)
			}
			if len(exc) > 0 {
				sub.In("-status", exc...)
			}
		}
		if f.productType != nil {
			k, v := exp("product_type", *f.productType)
			sub.Eq(k, wrapEmpty(v))
		}
		if f.categoryID != nil {
			sub.Eq("category_id", *f.categoryID)
		}
		if f.vendor != nil { // Vendor cannot be empty.
			k, v := exp("vendor", *f.vendor)
			sub.Eq(k, v)
		}
		if f.barcode != nil { // Shopify doesn't seem to allow search by empty barcode.
			k, v := exp("barcode", *f.barcode)
			sub.Eq(k, v)
		}
		if f.sku != nil {
			k, v := exp("sku", *f.sku)
			if v == "" {
				sub.Add(k)
			} else {
				sub.Eq(k, v)
			}
		}
		if f.price != nil {
			sub.Eq("price", fmt.Sprintf("%f", *f.price))
		}
		if f.giftCard != nil {
			sub.Eq("gift_card", fmt.Sprintf("%t", *f.giftCard))
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
		"title",
		"options",
		"product_type",
		"category",
		"tags",
		"vendor",
		"variants",
		"media",
		"status",
		"created",
		"updated",
	}
}

func exp(f string, s string) (string, string) {
	if strings.HasPrefix(s, "-") {
		return "-" + f, s[1:]
	}
	return f, s
}

func wrapEmpty(s string) string {
	if s != "" {
		return s
	}
	return fmt.Sprintf("%q", s)
}

func formatDateTime(dt, tz string) string {
	t, err := time.Parse(time.RFC3339, dt)
	if err != nil {
		return dt
	}
	if tz == "" {
		return t.Format("2006-01-02 15:04:05")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return dt
	}
	return t.In(loc).Format("2006-01-02 15:04:05")
}
