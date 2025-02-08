package currentcontext

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Display the current-context.`

// NewCmdCurrentContext is a cmd to display current context.
func NewCmdCurrentContext() *cobra.Command {
	return &cobra.Command{
		Use:   "current-context",
		Short: "Display the current-context",
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

	ctx := cfg.CurrentContext()
	if ctx == "" {
		return fmt.Errorf("current-context is not set")
	}

	fmt.Println(ctx)
	return nil
}
