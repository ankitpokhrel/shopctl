package peek

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
)

const (
	helpText = `Peek lets you peek into the product data.

Use this command to quickly look into the upstream or local product data.`

	examples = `$ shopctl peek product <product_id>`
)

// Flag wraps available command flags.
type flag struct {
	id   string
	json bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	id := cmdutil.ShopifyProductID(args[0])
	if id == "" {
		cmdutil.ExitOnErr(fmt.Errorf("invalid product id"))
	}

	jsonOut, err := cmd.Flags().GetBool("json")
	cmdutil.ExitOnErr(err)

	f.id = id
	f.json = jsonOut
}

// NewCmdPeek creates a new product restore command.
// TODO: Implement `--from` option.
func NewCmdPeek() *cobra.Command {
	cmd := cobra.Command{
		Use:     "peek PRODUCT_ID",
		Short:   "Peek lets you peek into product data",
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
	cmd.Flags().Bool("json", false, "Output in JSON format")

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	product, err := client.GetProductByID(flag.id)
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
