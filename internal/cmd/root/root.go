package root

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/auth"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup"
	"github.com/ankitpokhrel/shopctl/internal/cmd/compare"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config"
	"github.com/ankitpokhrel/shopctl/internal/cmd/export"
	"github.com/ankitpokhrel/shopctl/internal/cmd/ingest"
	"github.com/ankitpokhrel/shopctl/internal/cmd/peek"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product"
	"github.com/ankitpokhrel/shopctl/internal/cmd/restore"
	"github.com/ankitpokhrel/shopctl/internal/cmd/version"
)

var verbosity int

// NewCmdRoot constructs a root command.
func NewCmdRoot() *cobra.Command {
	cmd := cobra.Command{
		Use:   "shopctl <cmd> <subcommand>",
		Short: "CLI to manage Shopify backup, restore, migration and sync",
		Long:  "shopctl controls the Shopify backup, restore, migration and sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP(
		"context", "c", "",
		"Override current-context",
	)
	cmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose", "v",
		"Set the verbosity level (e.g., -v, -vv, -vvv)",
	)

	cmd.SetHelpFunc(helpFunc)

	addChildCommands(&cmd)

	return &cmd
}

func addChildCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		auth.NewCmdAuth(),
		config.NewCmdConfig(),
		product.NewCmdProduct(),
		export.NewCmdExport(),
		ingest.NewCmdImport(),
		backup.NewCmdBackup(),
		restore.NewCmdRestore(),
		peek.NewCmdPeek(),
		compare.NewCmdCompare(),
		version.NewCmdVersion(),
	)
}
