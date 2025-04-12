package peek

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Peek lets you peek into the product data.

Use this command to quickly look into the upstream or local product data.`

	examples = `# Peek by id
$ shopctl peek product <product_id>

# Peek a product from the import folder
# Context and strategy is skipped for direct path
$ shopctl peek product <product_id> --from </path/to/backup>

# Render json output
$ shopctl peek product <product_id> --json`
)

// Flag wraps available command flags.
type flag struct {
	id   string
	from string
	json bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	id := cmdutil.ShopifyProductID(args[0])
	if id == "" {
		cmdutil.ExitOnErr(fmt.Errorf("invalid product id"))
	}

	from, err := cmd.Flags().GetString("from")
	cmdutil.ExitOnErr(err)

	jsonOut, err := cmd.Flags().GetBool("json")
	cmdutil.ExitOnErr(err)

	f.id = id
	f.from = from
	f.json = jsonOut
}

// NewCmdPeek creates a new product restore command.
// TODO: Implement `--from` option.
func NewCmdPeek() *cobra.Command {
	cmd := cobra.Command{
		Use:     "peek PRODUCT_ID",
		Short:   "Peek into product data",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"view"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}
	cmd.Flags().StringP("from", "f", "", "Direct path to the backup to look into")
	cmd.Flags().Bool("json", false, "Output in JSON format")

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	var (
		product *schema.Product
		reg     *registry.Registry
		err     error
	)

	flag := &flag{}
	flag.parse(cmd, args)

	if flag.from != "" {
		reg, err = registry.NewRegistry(flag.from)
		if err != nil {
			return err
		}
		product, err = reg.GetProductByID(cmdutil.ExtractNumericID(flag.id))
	} else {
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
