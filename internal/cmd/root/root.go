package root

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/auth"
	"github.com/ankitpokhrel/shopctl/internal/cmd/backup"
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
		"store", "s", "",
		"Shopify store to look into",
	)
	cmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose", "v",
		"Set the verbosity level (e.g., -v, -vv, -vvv)",
	)

	_ = cmd.MarkPersistentFlagRequired("store")
	cmd.SetHelpFunc(helpFunc)

	addChildCommands(&cmd)

	return &cmd
}

func addChildCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		auth.NewCmdAuth(),
		backup.NewCmdBackup(),
		restore.NewCmdRestore(),
		version.NewCmdVersion(),
	)
}
