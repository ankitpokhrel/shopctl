package product

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Product lets you peek into the product data.

Use this command to quickly look into the upstream or local product data.`

	examples = `$ shopctl peek product --id <product_id>
$ shopctl peek product --handle <product_handle>
$ shopctl peek product --id <product_id> --from </path/to/bkp>`
)

// Flag wraps available command flags.
type flag struct {
	store  string
	id     string
	handle string
	from   string
	json   bool
}

func (f *flag) parse(cmd *cobra.Command) {
	store, err := cmd.Flags().GetString("store")
	cmdutil.ExitOnErr(err)

	id, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)

	handle, err := cmd.Flags().GetString("handle")
	cmdutil.ExitOnErr(err)

	from, err := cmd.Flags().GetString("from")
	cmdutil.ExitOnErr(err)

	jsonOut, err := cmd.Flags().GetBool("json")
	cmdutil.ExitOnErr(err)

	id = cmdutil.ShopifyProductID(id)

	if id == "" && handle == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf(`Error: either a valid product ID or handle is required.

Usage:
  $ shopctl peek product --id <product_id>
  $ shopctl peek product --handle <product_handle>

See 'shopctl peek product --help' for more info.`),
		)
	}

	f.store = store
	f.id = id
	f.handle = handle
	f.from = from
	f.json = jsonOut
}

// NewCmdProduct creates a new product restore command.
func NewCmdProduct() *cobra.Command {
	cmd := cobra.Command{
		Use:     "product",
		Short:   "Product lets you peek into product data",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value("gqlClient").(*api.GQLClient)
			return peek(cmd, client)
		},
	}
	cmd.Flags().String("id", "", "Peek by product ID")
	cmd.Flags().StringP("handle", "l", "", "Peek by product handle")
	cmd.Flags().StringP("from", "f", "", "Fetch from local backup")

	cmd.Flags().SortFlags = false

	return &cmd
}

func peek(cmd *cobra.Command, client *api.GQLClient) error {
	var (
		product *schema.Product
		err     error
	)

	flag := &flag{}
	flag.parse(cmd)

	if flag.from != "" {
		// TODO: Set source to local backup.
	}

	if flag.id != "" {
		product, err = client.GetProductByID(flag.id)
	} else {
		product, err = client.GetProductByHandle(flag.handle)
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
	r := NewFormatter(flag.store, product)
	return r.Render()
}
