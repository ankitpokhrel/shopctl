package remove

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Remove deletes a backup configuration/preset.

This action is irreversible.`

type flag struct {
	store string
	alias string
	force bool
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	alias, err := cmd.Flags().GetString("alias")
	cmdutil.ExitOnErr(err)

	if alias == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: config alias is required.

Usage:
  $ shopctl backup config remove -s <store> -a <alias>

See 'shopctl backup config remove --help' for more info.`),
		)
	}

	force, err := cmd.Flags().GetBool("force")
	cmdutil.ExitOnErr(err)

	f.store = store
	f.alias = alias
	f.force = force
}

// NewCmdRemove creates a new config remove command.
func NewCmdRemove() *cobra.Command {
	cmd := cobra.Command{
		Use:     "remove",
		Short:   "Remove deletes a backup config",
		Long:    helpText,
		Aliases: []string{"rm", "delete", "del"},
		RunE:    remove,
	}
	cmd.Flags().StringP("alias", "a", "", "Alias of the config to delete")
	cmd.Flags().Bool("force", false, "Delete without confirmation")

	cmd.Flags().SortFlags = false

	return &cmd
}

func remove(cmd *cobra.Command, _ []string) error {
	flag := &flag{}
	flag.parse(cmd)

	if err := config.DeletePreset(flag.store, flag.alias, flag.force); err != nil {
		if errors.Is(err, config.ErrNoConfig) {
			cmdutil.Fail("Preset with alias '%s' couldn't be found for store '%s'", flag.alias, flag.store)
			os.Exit(1)
		}
		if errors.Is(err, config.ErrActionAborted) {
			cmdutil.Fail("Action aborted")
			os.Exit(1)
		}
		return err
	}
	cmdutil.Success("Config with alias '%s' for store '%s' was deleted successfully", flag.alias, flag.store)
	return nil
}
