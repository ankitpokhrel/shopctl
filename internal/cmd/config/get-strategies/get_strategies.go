package getstrategies

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Display one or many backup strategies defined for current context.`
	tabWidth = 8
)

// NewCmdGetStrategies is a cmd to get all available strategies for current context.
func NewCmdGetStrategies() *cobra.Command {
	return &cobra.Command{
		Use:     "get-strategies",
		Short:   "Display one or many backup strategies defined for current context",
		Long:    helpText,
		Aliases: []string{"get-strategy"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	ctx, err := cmdutil.GetContext(cmd, cfg)
	if err != nil {
		return err
	}

	storeCfg, err := config.NewStoreConfig(ctx.Store, ctx.Alias)
	if err != nil {
		return err
	}

	var out []config.BackupStrategy

	allStrategies := storeCfg.Strategies()
	if len(allStrategies) == 0 {
		return fmt.Errorf("no strategy defined for the context %q", ctx.Alias)
	}
	currentStrategy := cfg.CurrentStrategy()

	givenStrategy := ""
	if len(args) > 0 {
		givenStrategy = args[0]
	}

	if givenStrategy == "" {
		out = allStrategies
	} else {
		for _, x := range allStrategies {
			if x.Name == givenStrategy {
				out = append(out, x)
				break
			}
		}
	}

	if len(out) == 0 {
		return fmt.Errorf("strategy %q not found for the context %q", givenStrategy, ctx.Alias)
	}

	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, tabWidth, 1, '\t', 0)

	_, _ = fmt.Fprintf(w, "%s\t %s\t%s\t%s\t%s\t%s\n", "NAME", "TYPE", "BACKUP DIR", "PREFIX", "RESOURCES", "FILTERS")
	for _, s := range out {
		name := s.Name
		if name == currentStrategy {
			name += "*"
		}
		resources := make([]string, 0, len(s.Resources))
		filters := make([]string, 0, len(s.Resources))
		for _, r := range s.Resources {
			resources = append(resources, r.Resource)
			if r.Query != "" {
				filters = append(filters, fmt.Sprintf("%s=%q", r.Resource, r.Query))
			}
		}
		_, _ = fmt.Fprintf(w, "%s\t %s\t%s\t%s\t%s\t%s\n", name, s.Kind, s.BkpDir, s.BkpPrefix, strings.Join(resources, ","), strings.Join(filters, ","))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Println(b.String())
	return nil
}
