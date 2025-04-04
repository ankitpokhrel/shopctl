package product

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	cmputil "github.com/ankitpokhrel/shopctl/internal/cmdutil/compare"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/structdiff"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Product lets you compare two products.

Use this command to quickly compare data from the upstream with your latest backup data
or, to compare data from two versions of the backup.`

	examples = `# Compare with the upstream
$ shopctl cmp product --id <product_id> --with </path/to/bkp>

# Compare data between two backups
$ shopctl cmp product --id <product_id> --from </path/to/bkp/v1> --with </path/to/bkp/v2>`
)

var (
	sortOrder = []string{
		"ID",
		"Status",
		"Title",
		"Handle",
		"Description",
		"DescriptionHtml",
		"ProductType",
		"Category",
		"Tags",
		"Vendor",
		"Options",
		"Variants",
	}

	ignoreFields = []string{
		"ID",
		"LegacyResourceID",
		"DefaultCursor",
		"Description",
		"MediaCount",
		"VariantsCount",
		"Edges",
		"PageInfo",
	}
)

// Flag wraps available command flags.
type flag struct {
	id   string
	from string
	with string
}

func (f *flag) parse(cmd *cobra.Command) {
	id, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)

	from, err := cmd.Flags().GetString("from")
	cmdutil.ExitOnErr(err)

	with, err := cmd.Flags().GetString("with")
	cmdutil.ExitOnErr(err)

	id = cmdutil.ShopifyProductID(id)

	usage := `Usage:
  $ shopctl cmp product --id <product_id> --with <bkp_id>
  $ shopctl cmp product --id <product_id> --from </path/to/source/bkp> --with <bkp_id>

See 'shopctl cmp product --help' for more info.`

	if id == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf("error: a valid product ID is required. \n\n%s", usage),
		)
	}
	if with == "" {
		cmdutil.ExitOnErr(
			fmt.Errorf("error: path to the backup to compare with is required. \n\n%s", usage),
		)
	}

	f.id = id
	f.from = from
	f.with = with
}

// NewCmdProduct creates a new product restore command.
func NewCmdProduct() *cobra.Command {
	cmd := cobra.Command{
		Use:     "product",
		Short:   "Product lets you compare two products",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(compare(cmd, client))
			return nil
		},
	}
	cmd.Flags().String("id", "", "Compare by product ID")
	cmd.Flags().StringP("with", "w", "", "Compare with product from given path")
	cmd.Flags().StringP("from", "f", "", "Look for product in the given path instead of upstream")

	cmd.Flags().SortFlags = false

	return &cmd
}

func compare(cmd *cobra.Command, client *api.GQLClient) error {
	var (
		product *schema.Product
		reg     *registry.Registry
		err     error
	)

	flag := &flag{}
	flag.parse(cmd)

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

	// TODO: Fix compare cmd.
	path, err := registry.LookForDirWithSuffix(flag.with, flag.from)
	if err != nil {
		return err
	}

	reg, err = registry.NewRegistry(path)
	if err != nil {
		return err
	}
	backedProduct, err := reg.GetProductByID(cmdutil.ExtractNumericID(flag.id))
	if err != nil {
		return err
	}

	diff := structdiff.New(*product, *backedProduct, structdiff.WithIgnoreList(ignoreFields))
	diffs := diff.Get()

	if len(diffs) == 0 {
		fmt.Println("No differences found.")
		return nil
	}
	return cmputil.Render(diffs, sortOrder)
}
