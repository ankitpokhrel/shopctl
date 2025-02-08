package getcontexts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Displays one or many contexts from the shopconfig file.`
	tabWidth = 8
)

// NewCmdGetContexts is a cmd to get all available contexts.
func NewCmdGetContexts() *cobra.Command {
	return &cobra.Command{
		Use:     "get-contexts",
		Short:   "Displays one or many contexts from the shopconfig file",
		Long:    helpText,
		Aliases: []string{"get-context"},
		RunE:    getContexts,
	}
}

func getContexts(cmd *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	var out []config.StoreContext

	allContexts := cfg.Contexts()
	currentCtx := cfg.CurrentContext()

	if len(args) == 0 {
		out = allContexts
	} else {
		ctx := args[0]
		for _, x := range allContexts {
			if x.Alias == ctx {
				out = append(out, x)
				break
			}
		}
	}

	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, tabWidth, 1, '\t', 0)

	_, _ = fmt.Fprintf(w, "%s\t %s\n", "NAME", "STORE")
	for _, x := range out {
		name := x.Alias
		if name == currentCtx {
			name += "*"
		}
		_, _ = fmt.Fprintf(w, "%s\t %s\n", name, x.Store)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Println(b.String())
	return nil
}
