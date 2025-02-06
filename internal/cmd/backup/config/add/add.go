package add

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/engine"
)

const helpText = `Add creates a backup configuration.

You can create multiple backup config/preset for a store.`

type flag struct {
	store     string
	alias     string
	kind      string
	bkpDir    string
	bkpPrefix string
	resources []string
	force     bool
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	alias, err := cmd.Flags().GetString("alias")
	cmdutil.ExitOnErr(err)

	kind, err := cmd.Flags().GetString("type")
	cmdutil.ExitOnErr(err)

	if kind == "" {
		kind = string(engine.BackupTypeIncremental)
	}

	dir, err := cmd.Flags().GetString("dir")
	cmdutil.ExitOnErr(err)

	prefix, err := cmd.Flags().GetString("prefix")
	cmdutil.ExitOnErr(err)

	resource, err := cmd.Flags().GetString("resource")
	cmdutil.ExitOnErr(err)

	force, err := cmd.Flags().GetBool("force")
	cmdutil.ExitOnErr(err)

	var resources []string
	if resource == "" {
		resources = []string{string(engine.Product)}
	} else {
		resources = strings.Split(resource, ",")
	}

	f.store = store
	f.alias = alias
	f.kind = kind
	f.bkpDir = dir
	f.bkpPrefix = prefix
	f.resources = resources
	f.force = force
}

// NewCmdAdd creates a new config add command.
func NewCmdAdd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "add",
		Short:   "Add creates a backup config",
		Long:    helpText,
		Aliases: []string{"create"},
		RunE:    add,
	}
	cmd.Flags().StringP("alias", "a", "", "Unique alias for the config")
	cmd.Flags().StringP("dir", "d", "", "Root directory to save backups to")
	cmd.Flags().StringP("prefix", "p", "", "Prefix for the main backup directory")
	cmd.Flags().StringP("resource", "r", "", "Resource types to backup (comma separated)")
	cmd.Flags().StringP("type", "t", "", "Backup time (full or incremental)")
	cmd.Flags().Bool("force", false, "Replace config if it already exist")

	cmd.Flags().SortFlags = false

	return &cmd
}

func add(cmd *cobra.Command, _ []string) error {
	flag := &flag{}
	flag.parse(cmd)

	presetCfg, err := config.NewPresetConfig(flag.store, config.PresetItems{
		Alias:     flag.alias,
		Kind:      flag.kind,
		BkpDir:    flag.bkpDir,
		BkpPrefix: flag.bkpPrefix,
		Resources: flag.resources,
		Force:     flag.force,
	})
	if err != nil {
		return err
	}

	if err := presetCfg.Save(); err != nil {
		if errors.Is(err, config.ErrConfigExist) {
			cmdutil.Fail(
				"Error: config with the same name already exist for the store; either use a unique alias or `--force` to replace current config.",
			)
			os.Exit(1)
		}
		return err
	}

	if flag.force {
		cmdutil.Success("Config replaced successfully")
	} else {
		cmdutil.Success("Config created successfully")
	}
	return nil
}
