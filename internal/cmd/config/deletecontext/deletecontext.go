package deletecontext

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Delete the specified context from the shopconfig.`

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

// NewCmdDeleteContext cmd allows you to delete a context.
func NewCmdDeleteContext() *cobra.Command {
	cmd := cobra.Command{
		Use:     "delete-context CONTEXT_NAME",
		Short:   "Delete the specified context from the shopconfig",
		Long:    helpText,
		Aliases: []string{"use"},
		Args:    cobra.MinimumNArgs(1),
		RunE:    deleteContext,
	}

	cmd.Flags().Bool("force", false, "Delete without confirmation")

	return &cmd
}

func deleteContext(cmd *cobra.Command, args []string) error {
	flag := &flag{}
	flag.parse(cmd, args)

	ctx := flag.name

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	if !shopCfg.HasContext(ctx) {
		return fmt.Errorf("no context exists with the name: %q", ctx)
	}

	if flag.force {
		return del(shopCfg, ctx)
	}

	fmt.Printf("You are about to delete context %q. This action is irreversible. Are you sure? (y/N): ", ctx)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		return del(shopCfg, ctx)
	}

	cmdutil.ExitOnErr(config.ErrActionAborted)
	return nil
}

func del(shopCfg *config.ShopConfig, ctx string) error {
	storeCfg, err := config.NewStoreConfig("", ctx)
	if err != nil {
		return err
	}

	if shopCfg.CurrentContext() == ctx {
		cmdutil.Warn("WARN: This removed your active context and strategies, use \"shopctl config use-context\" to select a different one")
		shopCfg.UnsetCurrentContext()
	}

	shopCfg.DeleteContext(ctx)
	if err := shopCfg.Save(); err != nil {
		return err
	}
	if err := storeCfg.Delete(); err != nil {
		return err
	}

	cmdutil.Success("Deleted context %q from %s", ctx, shopCfg.Path())
	return nil
}
