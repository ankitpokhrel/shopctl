package edit

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const helpText = `Edit opens a config file in the configured editor for you to edit.

Note that this will skip any validations, and you need to make sure
that the file is valid after updating.`

type flag struct {
	store string
	alias string
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	alias, err := cmd.Flags().GetString("alias")
	cmdutil.ExitOnErr(err)

	f.store = store
	f.alias = alias
}

// NewCmdEdit creates a new config edit command.
func NewCmdEdit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "edit",
		Short:   "Edit lets you edit a config file",
		Long:    helpText,
		Aliases: []string{"udpate", "modify"},
		RunE:    edit,
	}
	cmd.Flags().StringP("alias", "a", "", "Alias of the config to delete")

	return &cmd
}

func edit(cmd *cobra.Command, _ []string) error {
	flag := &flag{}
	flag.parse(cmd)

	editor, args := cmdutil.GetEditor()
	if editor == "" {
		cmdutil.Fail("Unable to locate any editors; You can set prefered editor using `SHOPIFY_EDITOR` or `EDITOR` env")
		os.Exit(1)
	}
	file, err := config.GetPresetLoc(flag.store, flag.alias)
	if err != nil {
		cmdutil.Fail("Preset with alias '%s' couldn't be found for store '%s'", flag.alias, flag.store)
		os.Exit(1)
	}

	ex := exec.Command(editor, append(args, file)...)
	ex.Stdin = os.Stdin
	ex.Stdout = os.Stdout
	ex.Stderr = os.Stderr

	if err := ex.Run(); err != nil {
		return err
	}
	cmdutil.Success("Config with alias '%s' for store '%s' was updated successfully", flag.alias, flag.store)
	return nil
}
