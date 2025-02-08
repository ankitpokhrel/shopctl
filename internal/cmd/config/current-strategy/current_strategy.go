package currentstrategy

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Display the current-strategy for a context.`

// NewCmdCurrentStrategy is a cmd to display current strategy.
func NewCmdCurrentStrategy() *cobra.Command {
	return &cobra.Command{
		Use:   "current-strategy",
		Short: "Display the current-strategy",
		Long:  helpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}
}

func run(_ *cobra.Command, _ []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	strategy := cfg.CurrentStrategy()
	if strategy == "" {
		return fmt.Errorf("current-strategy is not set")
	}

	fmt.Println(strategy)
	return nil
}
