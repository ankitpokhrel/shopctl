package config

import (
	"github.com/spf13/cobra"

	cc "github.com/ankitpokhrel/shopctl/internal/cmd/config/current-context"
	dc "github.com/ankitpokhrel/shopctl/internal/cmd/config/delete-context"
	gc "github.com/ankitpokhrel/shopctl/internal/cmd/config/get-contexts"
	rc "github.com/ankitpokhrel/shopctl/internal/cmd/config/rename-context"
	uc "github.com/ankitpokhrel/shopctl/internal/cmd/config/use-context"
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
		gc.NewCmdGetContexts(),
		rc.NewCmdRenameContext(),
	)

	return &cmd
}

func config(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
