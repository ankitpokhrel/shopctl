package edit

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Edit product options.`

	examples = `$ shopct product option edit 8856145494 Size -l

    # We can remove option values with '-'
    # Edit option to remove size 'xs' and add size 's'
    $ shopctl product option edit 8856145494 Size -l"-xs" -ls

    # Set variant strategy to MANAGE; default is LEAVE_AS_IS
    # With '--manage' flag, variants are created and deleted according to the option values to add and delete
    # See https://shopify.dev/docs/api/admin-graphql/latest/enums/ProductOptionUpdateVariantStrategy
    $ shopctl product option edit 8856145494 Style -nMood -lCasual -lFormal -l"-Informal" --manage`
)

// Flag wraps available command flags.
type flag struct {
	id       string
	toEdit   string
	name     string
	position *int
	values   []string
	manage   schema.ProductOptionUpdateVariantStrategy
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	var (
		position *int
		err      error
	)

	id := args[0]
	toEdit := args[1]

	name, err := cmd.Flags().GetString("name")
	cmdutil.ExitOnErr(err)

	posFlag := cmd.Flags().Lookup("position")
	if posFlag != nil && posFlag.Changed {
		pos, err := cmd.Flags().GetInt("position")
		cmdutil.ExitOnErr(err)
		position = &pos
	}

	values, err := cmd.Flags().GetStringArray("value")
	cmdutil.ExitOnErr(err)

	manage, err := cmd.Flags().GetBool("manage")
	cmdutil.ExitOnErr(err)

	strategy := schema.ProductOptionUpdateVariantStrategyLeaveAsIs
	if manage {
		strategy = schema.ProductOptionUpdateVariantStrategyManage
	}

	f.id = cmdutil.ShopifyProductID(id)
	f.toEdit = toEdit
	f.name = name
	f.position = position
	f.values = values
	f.manage = strategy
}

// NewCmdEdit constructs a new product option edit command.
func NewCmdEdit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "edit PRODUCT_ID OPTION_NAME",
		Short:   "Edit product options",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(2),
		Aliases: []string{"update"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "New option name")
	cmd.Flags().IntP("position", "p", 0, "Option position")
	cmd.Flags().StringArrayP("value", "l", []string{}, "Option values")
	cmd.Flags().Bool("manage", false, "Option update variant strategy")

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	if !hasAnythingToUpdate(flag) {
		cmdutil.Warn("Nothing to update")
		os.Exit(0)
	}

	productOptions, err := client.GetProductOptions(flag.id)
	if err != nil {
		return err
	}

	var optToEdit *schema.ProductOption
	for _, opt := range productOptions.Data.Product.Options {
		if !strings.EqualFold(opt.Name, flag.toEdit) {
			continue
		}
		optToEdit = &opt
	}

	if optToEdit == nil {
		cmdutil.Fail("Option with name %q not found", flag.toEdit)
		os.Exit(1)
	}

	name := optToEdit.Name
	if flag.name != "" {
		name = flag.name
	}
	pos := optToEdit.Position
	if flag.position != nil {
		pos = *flag.position
	}

	opt := schema.OptionUpdateInput{
		ID:       optToEdit.ID,
		Name:     &name,
		Position: &pos,
	}

	currentOptionValuesMap := make(map[string]*schema.ProductOptionValue, 0)
	for _, v := range optToEdit.OptionValues {
		currentOptionValuesMap[strings.ToLower(v.Name)] = &v
	}

	optionValuesToCreate := make([]schema.OptionValueCreateInput, 0)
	optionValuesToUpdate := make([]schema.OptionValueUpdateInput, 0)
	optionValuesToDelete := make([]string, 0)

	for _, v := range flag.values {
		if strings.HasPrefix(v, "-") {
			k := strings.ToLower(v[1:])
			if val, ok := currentOptionValuesMap[k]; ok {
				optionValuesToDelete = append(optionValuesToDelete, val.ID)
			}
		} else {
			k := strings.ToLower(v)
			if val, ok := currentOptionValuesMap[k]; ok {
				optionValuesToUpdate = append(optionValuesToUpdate, schema.OptionValueUpdateInput{
					ID:   val.ID,
					Name: &v,
				})
			} else {
				optionValuesToCreate = append(optionValuesToCreate, schema.OptionValueCreateInput{
					Name: &v,
				})
			}
		}
	}

	res, err := client.UpdateProductOptions(
		flag.id, &opt,
		optionValuesToCreate,
		optionValuesToUpdate,
		optionValuesToDelete,
		flag.manage,
	)
	if err != nil {
		return err
	}

	cmdutil.Success("Option %q of product %s updated successfully", flag.toEdit, res.Product.ID)
	return nil
}

func hasAnythingToUpdate(f *flag) bool {
	return f.name != "" || f.position != nil || len(f.values) != 0
}
