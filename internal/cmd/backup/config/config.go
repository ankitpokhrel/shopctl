package config

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/config/add"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/config/list"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup/config/remove"
)

const helpText = `Config helps you manage backup configuration.`

// NewCmdConfig creates a new config command.
func NewCmdConfig() *cobra.Command {
	cmd := cobra.Command{
		Use:     "config",
		Short:   "Config manages backup config",
		Long:    helpText,
		Aliases: []string{"cfg", "configure"},
		RunE:    config,
	}

	cmd.AddCommand(
		add.NewCmdAdd(),
		list.NewCmdList(),
		remove.NewCmdRemove(),
	)

	return &cmd
}

func config(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
