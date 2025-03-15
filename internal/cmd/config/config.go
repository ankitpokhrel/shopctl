package config

import (
	"github.com/spf13/cobra"

	cc "github.com/ankitpokhrel/shopctl/internal/cmd/config/current-context"
	cs "github.com/ankitpokhrel/shopctl/internal/cmd/config/current-strategy"
	dc "github.com/ankitpokhrel/shopctl/internal/cmd/config/delete-context"
	ds "github.com/ankitpokhrel/shopctl/internal/cmd/config/delete-strategy"
	ec "github.com/ankitpokhrel/shopctl/internal/cmd/config/edit-context"
	gc "github.com/ankitpokhrel/shopctl/internal/cmd/config/get-contexts"
	gs "github.com/ankitpokhrel/shopctl/internal/cmd/config/get-strategies"
	rc "github.com/ankitpokhrel/shopctl/internal/cmd/config/rename-context"
	rs "github.com/ankitpokhrel/shopctl/internal/cmd/config/rename-strategy"
	ss "github.com/ankitpokhrel/shopctl/internal/cmd/config/set-strategy"
	uc "github.com/ankitpokhrel/shopctl/internal/cmd/config/use-context"
	us "github.com/ankitpokhrel/shopctl/internal/cmd/config/use-strategy"
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
		uc.NewCmdUseContext(),
		cc.NewCmdCurrentContext(),
		dc.NewCmdDeleteContext(),
		ec.NewCmdEditContext(),
		gc.NewCmdGetContexts(),
		rc.NewCmdRenameContext(),
		ss.NewCmdSetStrategy(),
		us.NewCmdUseStrategy(),
		cs.NewCmdCurrentStrategy(),
		ds.NewCmdDeleteStrategy(),
		gs.NewCmdGetStrategies(),
		rs.NewCmdRenameStrategy(),
	)

	return &cmd
}

func config(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
