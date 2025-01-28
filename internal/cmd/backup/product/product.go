package product

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	pr "github.com/ankitpokhrel/shopctl/internal/runner/backup/product"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const (
	helpText = `Product initiates a product backup process for a shopify store.

Use this command to back up the entire product catalog or a subset based on filters like category, tags, or date.`

	examples = `$ shopctl backup product

# Back up first 100 products
$ shopctl backup product --limit 100

# Back up products from a specific category (e.g., "Winter Collection")
$ shopctl backup product --category "Winter Collection"

# Back up products with specific tags (e.g., "Sale" and "New")
$ shopctl backup product --tags "Sale,New"

# Back up products created or updated after a certain date
$ shopctl backup product --since "2024-01-01"

# Perform a dry run to see which products would be backed up
$ shopctl backup product --dry-run

# Perform an incremental backup (only new or updated products)
$ shopctl backup product --incremental`
)

// NewCmdProduct creates a new product backup command.
func NewCmdProduct(store string) *cobra.Command {
	return &cobra.Command{
		Use:     "product",
		Short:   "Product initiates product backup",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return product(store, cmd, args)
		},
	}
}

func product(store string, cmd *cobra.Command, _ []string) error {
	eng := engine.New(engine.NewBackup(store))
	client := cmd.Context().Value("gqlClient").(*api.GQLClient)
	logger := cmd.Context().Value("logger").(*tlog.Logger)

	runner := pr.NewRunner(eng, client, logger)
	return runner.Run()
}
