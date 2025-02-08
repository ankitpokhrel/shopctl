package usestrategy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Sets the current-strategy for the context in a shopconfig file.`

// NewCmdUseStrategy is a cmd to update current strategy for the context.
func NewCmdUseStrategy() *cobra.Command {
	return &cobra.Command{
		Use:   "use-strategy STRATEGY_NAME",
		Short: "Sets the current-strategy in a shopconfig file",
		Long:  helpText,
		Args:  cobra.MinimumNArgs(1),
		RunE:  useStrategy,
	}
}

func useStrategy(cmd *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	strategy := args[0]
	if len(strategy) == 0 {
		return fmt.Errorf("empty strategy names are not allowed")
	}

	if err := cfg.SetCurrentStrategy(strategy); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	cmdutil.Success("Switched to strategy %q", strategy)
	return nil
}
