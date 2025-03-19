package getcontexts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Display one or many contexts defined in the shopconfig file.`
	tabWidth = 8
)

// NewCmdGetContexts is a cmd to get all available contexts.
func NewCmdGetContexts() *cobra.Command {
	return &cobra.Command{
		Use:     "get-contexts",
		Short:   "Display one or many contexts defined in the shopconfig file",
		Long:    helpText,
		Aliases: []string{"get-context"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
}

func run(_ *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	var out []config.StoreContext

	allContexts := cfg.Contexts()
	if len(allContexts) == 0 {
		return fmt.Errorf("no contexts found")
	}
	currentCtx := cfg.CurrentContext()

	givenCtx := ""
	if len(args) > 0 {
		givenCtx = args[0]
	}

	if givenCtx == "" {
		out = allContexts
	} else {
		for _, x := range allContexts {
			if x.Alias == givenCtx {
				out = append(out, x)
				break
			}
		}
	}

	if len(out) == 0 {
		return fmt.Errorf("context not found: %q", givenCtx)
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
