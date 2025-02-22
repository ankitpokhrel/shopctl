package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/registry"
)

const (
	helpText = `List lists all available backups from the configured location.`
	tabWidth = 8
)

// NewCmdList creates a new list command.
func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List lists available backups",
		Long:    helpText,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shopCfg := cmd.Context().Value("shopCfg").(*config.ShopConfig)
			ctx := cmd.Context().Value("context").(*config.StoreContext)

			cmdutil.ExitOnErr(run(cmd, shopCfg, ctx))
			return nil
		},
	}
}

func run(cmd *cobra.Command, shopCfg *config.ShopConfig, ctx *config.StoreContext) error {
	strategy, err := cmdutil.GetStrategy(cmd, ctx, shopCfg)
	if err != nil {
		return err
	}

	file, err := registry.GetAllInDir(strategy.BkpDir, ".tar.gz")
	if err != nil {
		return err
	}

	metaItems := make([]config.RootMetaItems, 0)

	for f := range file {
		if f.Err != nil {
			return f.Err
		}
		byte, err := registry.ReadFromTarGZ(f.Path, "metadata.json")
		if err != nil {
			return err
		}

		var meta config.RootMetaItems
		if err := json.Unmarshal(byte, &meta); err != nil {
			return fmt.Errorf("unable to decode metadata for %s", f.Path)
		}
		metaItems = append(metaItems, meta)
	}

	formatTime := func(u int64) string {
		t := time.Unix(u, 0)
		return t.Format("2006-01-02 15:04:05")
	}

	if len(metaItems) == 0 {
		fmt.Println("no backups found")
		return nil
	}

	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, tabWidth, 1, '\t', 0)

	_, _ = fmt.Fprintf(w, "%s\t %s\t%s\t%s\t%s\t%s\n", "ID", "RESOURCES", "FILTERS", "TIME START", "TIME END", "STATUS")
	for _, m := range metaItems {
		resources := make([]string, 0, len(m.Resources))
		filters := make([]string, 0, len(m.Resources))
		for _, r := range m.Resources {
			resources = append(resources, r.Resource)
			if r.Query != "" {
				filters = append(filters, fmt.Sprintf("%s=%q", r.Resource, r.Query))
			}
		}
		_, _ = fmt.Fprintf(
			w, "%s\t %s\t%s\t%s\t%s\t%s\n",
			m.ID, strings.Join(resources, ","), strings.Join(filters, ","),
			formatTime(m.TimeStart), formatTime(m.TimeEnd), m.Status,
		)
	}

	if err := w.Flush(); err != nil {
		return err
	}
	return cmdutil.PagerOut(b.String())
}
