package config

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/config/setstrategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/usecontext"
)

const helpText = `Config helps you manage shopctl configuration.`

// NewCmdConfig creates a new config command.
func NewCmdConfig() *cobra.Command {
	cmd := cobra.Command{
		Use:         "config",
		Short:       "Config manages shopctl config",
		Long:        helpText,
		Annotations: map[string]string{"cmd:main": "true"},
		Aliases:     []string{"cfg"},
		RunE:        config,
	}

	cmd.AddCommand(
		usecontext.NewCmdUseContext(),
		setstrategy.NewCmdSetStrategy(),
	)

	return &cmd
}

func config(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
