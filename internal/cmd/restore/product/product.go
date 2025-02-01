package product

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	pr "github.com/ankitpokhrel/shopctl/internal/runner/restore/product"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const (
	helpText = `Product initiates a product restoration process.

Use this command to restore the entire product catalog or a subset of data.`

	examples = `$ shopctl restore product --from </path/to/bkp>`
)

// NewCmdProduct creates a new product restore command.
func NewCmdProduct(store string) *cobra.Command {
	return &cobra.Command{
		Use:     "product BACKUP_PATH",
		Short:   "Product initiates a product restoration process",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return restore(store, cmd, args)
		},
	}
}

func restore(store string, cmd *cobra.Command, args []string) error {
	eng := engine.New(engine.NewRestore(store))
	client := cmd.Context().Value("gqlClient").(*api.GQLClient)
	logger := cmd.Context().Value("logger").(*tlog.Logger)

	runner := pr.NewRunner(args[0], eng, client, logger)
	return runner.Run()
}
