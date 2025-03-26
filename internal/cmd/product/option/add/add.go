package add

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Add product options.`

	examples = `$ shopct product option add 8856145494 --name Title --value "Special product"

    # Option with multiple values
    $ shopctl product option add 8856145494 -nSize -lxs -lsm -lxl

    # Set variant strategy to CREATE; default is LEAVE_AS_IS
    # With '--create' flag, existing variants are updated with the first option value
    # See https://shopify.dev/docs/api/admin-graphql/latest/enums/ProductOptionCreateVariantStrategy
    $ shopctl product option add 8856145494 -nStyle -lCasual -lInformal --create`
)

// Flag wraps available command flags.
type flag struct {
	id       string
	name     string
	position *int
	values   []string
	create   schema.ProductOptionCreateVariantStrategy
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	var (
		position *int
		err      error
	)

	id := args[0]

	name, err := cmd.Flags().GetString("name")
	cmdutil.ExitOnErr(err)

	if name == "" {
		cmdutil.ExitOnErr(cmdutil.HelpErrorf("Product option name cannot be blank", examples))
	}

	posFlag := cmd.Flags().Lookup("position")
	if posFlag != nil && posFlag.Changed {
		pos, err := cmd.Flags().GetInt("position")
		cmdutil.ExitOnErr(err)
		position = &pos
	}

	values, err := cmd.Flags().GetStringArray("value")
	cmdutil.ExitOnErr(err)

	create, err := cmd.Flags().GetBool("create")
	cmdutil.ExitOnErr(err)

	strategy := schema.ProductOptionCreateVariantStrategyLeaveAsIs
	if create {
		strategy = schema.ProductOptionCreateVariantStrategyCreate
	}

	f.id = id
	f.name = name
	f.position = position
	f.values = values
	f.create = strategy
}

// NewCmdAdd constructs a new product option add command.
func NewCmdAdd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "add PRODUCT_ID",
		Short:   "Add product options",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "Option name")
	cmd.Flags().IntP("position", "p", 0, "Option position")
	cmd.Flags().StringArrayP("value", "l", []string{}, "Option values")
	cmd.Flags().Bool("create", false, "Option create variant strategy")

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	opt := schema.OptionCreateInput{
		Name:     &flag.name,
		Position: flag.position,
		Values:   make([]any, 0, len(flag.values)),
	}
	for _, v := range flag.values {
		opt.Values = append(opt.Values, schema.OptionValueCreateInput{
			Name: &v,
		})
	}

	res, err := client.CreateProductOptions(cmdutil.ShopifyProductID(flag.id), []schema.OptionCreateInput{opt}, flag.create)
	if err != nil {
		return err
	}

	cmdutil.Success("Option added successfully to product: %s", res.Product.ID)
	return nil
}
