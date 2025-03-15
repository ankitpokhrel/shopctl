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

const (
	helpText = `Delete the specified strategy from the shopconfig file.`
	example  = `# Delete a strategy called 'weekly'
$ shopctl config delete-strategy weekly`
)

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
		Use:     "delete-strategy STRATEGY_NAME",
		Short:   "Delete the specified strategy from the shopconfig file",
		Long:    helpText,
		Example: example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.ExitOnErr(run(cmd, args))
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Delete without confirmation")

	return &cmd
}

func run(cmd *cobra.Command, args []string) error {
	flag := &flag{}
	flag.parse(cmd, args)

	strategy := flag.name

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	ctx := shopCfg.CurrentContext()
	if ctx == "" {
		return fmt.Errorf("current-context is not set")
	}

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
	return config.ErrActionAborted
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

	cmdutil.Success("Deleted strategy %q for the context %q from %s", strategy, ctx, storeCfg.Path())
	return nil
}
