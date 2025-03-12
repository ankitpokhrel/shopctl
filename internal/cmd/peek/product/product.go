package product

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Product lets you peek into the product data.

Use this command to quickly look into the upstream or local product data.`

	examples = `# Peek by id
$ shopctl peek product <product_id>

# Peek a product from a backup using backup id
$ shopctl peek product <product_id> -b <backup_id>

# Peek a product from the backup folder
# Context and strategy is skipped for direct path
$ shopctl peek product <product_id> --from </path/to/backup>`
)

// Flag wraps available command flags.
type flag struct {
	id    string
	bkpID string
	from  string
	json  bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	id := cmdutil.ShopifyProductID(args[0])
	if id == "" {
		cmdutil.ExitOnErr(fmt.Errorf("invalid product id"))
	}

	bkpID, err := cmd.Flags().GetString("backup-id")
	cmdutil.ExitOnErr(err)

	from, err := cmd.Flags().GetString("from")
	cmdutil.ExitOnErr(err)

	jsonOut, err := cmd.Flags().GetBool("json")
	cmdutil.ExitOnErr(err)

	f.id = id
	f.bkpID = bkpID
	f.from = from
	f.json = jsonOut
}

// NewCmdProduct creates a new product restore command.
func NewCmdProduct() *cobra.Command {
	cmd := cobra.Command{
		Use:     "product PRODUCT_ID",
		Short:   "Product lets you peek into product data",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"products"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			strategy := cmd.Context().Value(cmdutil.KeyStrategy).(*config.BackupStrategy)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, strategy, client))
			return nil
		},
	}
	cmd.Flags().StringP("backup-id", "b", "", "Backup id to look into")
	cmd.Flags().StringP("from", "f", "", "Direct path to the backup to look into")

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, strategy *config.BackupStrategy, client *api.GQLClient) error {
	var (
		product *schema.Product
		reg     *registry.Registry
		err     error
	)

	flag := &flag{}
	flag.parse(cmd, args)

	switch {
	case flag.bkpID != "":
		var path string

		path, err = registry.LookForDirWithSuffix(flag.bkpID, strategy.BkpDir)
		if err != nil {
			if errors.Is(err, registry.ErrNoTargetFound) {
				return fmt.Errorf("couldn't find backup with id %q in %q", flag.bkpID, strategy.BkpDir)
			}
			return err
		}

		reg, err = registry.NewRegistry(path)
		if err != nil {
			return err
		}
		product, err = reg.GetProductByID(cmdutil.ExtractNumericID(flag.id))
	case flag.from != "":
		reg, err = registry.NewRegistry(flag.from)
		if err != nil {
			return err
		}
		product, err = reg.GetProductByID(cmdutil.ExtractNumericID(flag.id))
	default:
		product, err = client.GetProductByID(flag.id)
	}
	if err != nil {
		return err
	}

	if flag.json {
		s, err := json.MarshalIndent(product, "", "  ")
		if err != nil {
			return err
		}
		return cmdutil.PagerOut(string(s))
	}

	// Convert to Markdown.
	r := NewFormatter(ctx.Store, product)
	return r.Render()
}
