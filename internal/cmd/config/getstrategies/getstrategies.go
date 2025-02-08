package getstrategies

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Displays one or many strategies for current context from the shopconfig file.`
	tabWidth = 8
)

// NewCmdGetStrategies is a cmd to get all available strategies for current context.
func NewCmdGetStrategies() *cobra.Command {
	return &cobra.Command{
		Use:     "get-strategies",
		Short:   "Displays one or many strategies for current context from the shopconfig file",
		Long:    helpText,
		Aliases: []string{"get-strategy"},
		RunE:    getStrategies,
	}
}

func getStrategies(cmd *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	ctx := cfg.GetContext(cfg.CurrentContext())

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

	_, _ = fmt.Fprintf(w, "%s\t %s\t%s\t%s\t%s\n", "NAME", "TYPE", "BACKUP DIR", "PREFIX", "RESOURCES")
	for _, s := range out {
		name := s.Name
		if name == currentStrategy {
			name += "*"
		}
		_, _ = fmt.Fprintf(w, "%s\t %s\t%s\t%s\t%s\n", name, s.Kind, s.BkpDir, s.BkpPrefix, strings.Join(s.Resources, ","))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Println(b.String())
	return nil
}
