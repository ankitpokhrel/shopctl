package config

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/config/currentcontext"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/currentstrategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/deletecontext"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/deletestrategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/getcontexts"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/getstrategies"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/setstrategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/usecontext"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/usestrategy"
)

const helpText = `Modify shopconfig files using commands like "shopctl config set-context my-context".`

// NewCmdConfig creates a new config command.
func NewCmdConfig() *cobra.Command {
	cmd := cobra.Command{
		Use:         "config",
		Short:       "Modify shopconfig files",
		Long:        helpText,
		Annotations: map[string]string{"cmd:main": "true"},
		Aliases:     []string{"cfg"},
		RunE:        config,
	}

	cmd.AddCommand(
		usecontext.NewCmdUseContext(),
		currentcontext.NewCmdCurrentContext(),
		deletecontext.NewCmdDeleteContext(),
		getcontexts.NewCmdGetContexts(),
		setstrategy.NewCmdSetStrategy(),
		usestrategy.NewCmdUseStrategy(),
		currentstrategy.NewCmdCurrentStrategy(),
		deletestrategy.NewCmdDeleteStrategy(),
		getstrategies.NewCmdGetStrategies(),
	)

	return &cmd
}

func config(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
