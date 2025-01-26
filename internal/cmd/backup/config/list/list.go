package list

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `List lists available backup configurations for a store.`
	tabWidth = 8
)

type flag struct {
	store string
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	f.store = store
}

// NewCmdList creates a new config list command.
func NewCmdList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list",
		Short:   "List lists backup configs for a store",
		Long:    helpText,
		Aliases: []string{"ls"},
		RunE:    list,
	}

	return &cmd
}

func list(cmd *cobra.Command, _ []string) error {
	flag := &flag{}
	flag.parse(cmd)

	files, err := config.ListPresets(flag.store)
	if err != nil {
		if os.IsNotExist(err) {
			cmdutil.Fail("Error: couldn't find config for the given store")
			os.Exit(1)
		}
		return err
	}

	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, tabWidth, 1, '\t', 0)

	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Alias", "Type", "Backup Location", "Resources")
	for _, f := range files {
		base := path.Base(f)
		name := strings.TrimSuffix(base, path.Ext(base))

		item, err := config.ReadAllPreset(flag.store, name)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(
			w, "%s\t%s\t%s\t%s\n",
			item.Alias, item.Kind, item.BkpDir, strings.Join(item.Resources, ", "),
		)
	}

	if err := w.Flush(); err != nil {
		return err
	}
	return cmdutil.PagerOut(b.String())
}
