package currentcontext

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Displays the current-context.`

// NewCmdCurrentContext is a cmd to display current context.
func NewCmdCurrentContext() *cobra.Command {
	return &cobra.Command{
		Use:   "current-context",
		Short: "Displays the current-context",
		Long:  helpText,
		RunE:  useContext,
	}
}

func useContext(cmd *cobra.Command, args []string) error {
	cfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}

	ctx := cfg.CurrentContext()
	if ctx == "" {
		cmdutil.ExitOnErr(fmt.Errorf("current-context is not set"))
	}

	fmt.Println(ctx)
	return nil
}
