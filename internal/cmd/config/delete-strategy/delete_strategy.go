package deletestrategy

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Delete the specified strategy from the shopconfig.`

type flag struct {
	name  string
	force bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	name := args[0]

	force, err := cmd.Flags().GetBool("force")
	cmdutil.ExitOnErr(err)

	f.name = name
	f.force = force
}

// NewCmdDeleteStrategy cmd allows you to delete a context.
func NewCmdDeleteStrategy() *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete-strategy STRATEGY_NAME",
		Short: "Delete the specified strategy from the shopconfig",
		Long:  helpText,
		Args:  cobra.MinimumNArgs(1),
		RunE:  deleteStrategy,
	}

	cmd.Flags().Bool("force", false, "Delete without confirmation")

	return &cmd
}

func deleteStrategy(cmd *cobra.Command, args []string) error {
	flag := &flag{}
	flag.parse(cmd, args)

	strategy := flag.name

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	ctx := shopCfg.CurrentContext()

	if flag.force {
		return del(shopCfg, ctx, strategy)
	}

	fmt.Printf("You are about to delete strategy %q for the context %q. This action is irreversible. Are you sure? (y/N): ", strategy, ctx)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		return del(shopCfg, ctx, strategy)
	}

	cmdutil.ExitOnErr(config.ErrActionAborted)
	return nil
}

func del(shopCfg *config.ShopConfig, ctx string, strategy string) error {
	storeCfg, err := config.NewStoreConfig("", ctx)
	if err != nil {
		return err
	}
	if !storeCfg.HasBackupStrategy(strategy) {
		return fmt.Errorf("no strategy exists for context %q with the name: %q", ctx, strategy)
	}

	if shopCfg.CurrentStrategy() == strategy {
		cmdutil.Warn("WARN: This removed your active strategy, use \"shopctl config use-strategy\" to select a different one")
		shopCfg.UnsetCurrentStrategy()
	}

	storeCfg.UnsetStrategy(strategy)
	if err := storeCfg.Save(); err != nil {
		return err
	}

	cmdutil.Success("Deleted strategy %q for the context %q in path %s", strategy, ctx, storeCfg.Path())
	return nil
}
