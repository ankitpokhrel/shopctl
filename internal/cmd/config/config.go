package config

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/config/current-context"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/current-strategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/delete-context"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/delete-strategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/get-contexts"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/get-strategies"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/rename-context"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/rename-strategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/set-strategy"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/use-context"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config/use-strategy"
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
		renamecontext.NewCmdRenameContext(),
		setstrategy.NewCmdSetStrategy(),
		usestrategy.NewCmdUseStrategy(),
		currentstrategy.NewCmdCurrentStrategy(),
		deletestrategy.NewCmdDeleteStrategy(),
		getstrategies.NewCmdGetStrategies(),
		renamestrategy.NewCmdRenameStrategy(),
	)

	return &cmd
}

func config(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
