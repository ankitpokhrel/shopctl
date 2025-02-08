package currentstrategy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Displays the current-strategy for a context.`

// NewCmdCurrentStrategy is a cmd to display current strategy.
func NewCmdCurrentStrategy() *cobra.Command {
	return &cobra.Command{
		Use:   "current-strategy",
		Short: "Displays the current-strategy",
		Long:  helpText,
		RunE:  currentStrategy,
	}
}

func currentStrategy(cmd *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	strategy := cfg.CurrentStrategy()
	if strategy == "" {
		cmdutil.ExitOnErr(fmt.Errorf("current-strategy is not set"))
	}

	fmt.Println(strategy)
	return nil
}
