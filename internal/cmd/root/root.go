package root

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/auth"
	"github.com/ankitpokhrel/shopctl/internal/cmd/config"
	"github.com/ankitpokhrel/shopctl/internal/cmd/customer"
	"github.com/ankitpokhrel/shopctl/internal/cmd/export"
	"github.com/ankitpokhrel/shopctl/internal/cmd/ingest"
	"github.com/ankitpokhrel/shopctl/internal/cmd/product"
	"github.com/ankitpokhrel/shopctl/internal/cmd/version"
)

// NewCmdRoot constructs a root command.
func NewCmdRoot() *cobra.Command {
	cmd := cobra.Command{
		Use:   "shopctl <cmd> <subcommand>",
		Short: "Manage Shopify data directly from your terminal",
		Long:  "shopctl helps you manage Shopify data directly from your terminal.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP(
		"context", "c", "",
		"Override current-context",
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
		customer.NewCmdCustomer(),
		export.NewCmdExport(),
		ingest.NewCmdImport(),
		version.NewCmdVersion(),
	)
}
