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
	"github.com/ankitpokhrel/shopctl/pkg/fmtout"
	"github.com/ankitpokhrel/shopctl/pkg/search"
	"github.com/ankitpokhrel/shopctl/pkg/tui/table"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `List registered events/webhooks in a store.`

	examples = `$ shopctl event list`
)

type flag struct {
	idRange   string
	topics    []schema.WebhookSubscriptionTopic
	limit     int16
	plain     bool
	csv       bool
	noHeaders bool
	columns   []string
	created   string
	updated   string
}

func (f *flag) parse(cmd *cobra.Command, _ []string) {
	idRange, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)

	topics, err := cmd.Flags().GetString("topics")
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

	created, err := cmd.Flags().GetString("created")
	cmdutil.ExitOnErr(err)

	updated, err := cmd.Flags().GetString("updated")
	cmdutil.ExitOnErr(err)

	f.idRange = idRange
	f.topics = func() []schema.WebhookSubscriptionTopic {
		if topics == "" {
			return nil
		}
		whTopics := strings.Split(topics, ",")
		if len(whTopics) == 0 {
			return nil
		}
		subTopics := make([]schema.WebhookSubscriptionTopic, 0, len(whTopics))
		for _, t := range whTopics {
			subTopics = append(subTopics, schema.WebhookSubscriptionTopic(t))
		}
		return subTopics
	}()
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
	f.created = created
	f.updated = updated
}

// NewCmdList constructs a new webhook list command.
func NewCmdList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list [QUERY]",
		Short:   "List registered webhooks in a store",
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
	cmd.Flags().String("topics", "", "Filter by webhook topics")
	cmd.Flags().String("created", "", "Filter by the created date")
	cmd.Flags().String("updated", "", "Filter by the updated date")
	cmd.Flags().Int16("limit", 50, "Number of entries to fetch (max 250)")
	cmd.Flags().Bool("plain", false, "Show output in properly formatted plain text")
	cmd.Flags().Bool("csv", false, "Print output in csv")
	cmd.Flags().Bool("no-headers", false, "Don't print table headers (works only with --plain)")
	cmd.Flags().String("columns", "", "Comma separated list of columns to print (works only with --plain)")

	cmd.Flags().SortFlags = false

	return &cmd
}

//nolint:gocyclo
func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	query := buildSearchQuery(flag)

	webhooks, err := client.GetWebhooks(int(flag.limit), nil, flag.topics, query)
	if err != nil {
		return err
	}

	cols := []table.Column{
		{Title: "ID", Width: 15},
		{Title: "Topic", Width: 25},
		{Title: "Endpoint", Width: 60},
		{Title: "API Version", Width: 15},
		{Title: "Format", Width: 15},
		{Title: "Created", Width: 25},
		{Title: "Updated", Width: 25},
	}

	rows := make([]table.Row, 0)
	for _, wh := range webhooks {
		url := ""
		endpoint, ok := wh.Endpoint.(map[string]any)
		if ok {
			url = endpoint["callbackUrl"].(string)
		}
		row := table.Row{
			shopctl.ExtractNumericID(wh.ID),
			string(wh.Topic),
			url,
			wh.ApiVersion.Handle,
			string(wh.Format),
			cmdutil.FormatDateTime(wh.CreatedAt, ""),
			cmdutil.FormatDateTime(wh.UpdatedAt, ""),
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		cmdutil.Warn("No webhooks found for the given criteria")
		os.Exit(0)
	}

	if flag.plain || flag.csv {
		defaultCols := []string{"id", "topic", "endpoint", "api_version", "created"}
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
		"c/C: Copy numeric or full event ID",
		"q/CTRL+c/ESC: Quit",
	}
	footerTexts := []string{
		fmt.Sprintf("Showing %d results for store %q", len(rows), ctx.Store),
	}

	tbl := table.NewInteractiveTable(
		cols, rows,
		table.WithHelpTexts(helpTexts),
		table.WithFooterTexts(footerTexts),
		table.WithCopyFunc(func(id string, key string) error {
			if key == "C" {
				id = shopctl.ShopifyWebhookSubscriptionID(id)
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
		if f.idRange != "" {
			sub.Eq("id", f.idRange)
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
		"endpoint",
		"topic",
		"format",
		"api_version",
		"created",
		"updated",
	}
}
