package add

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Add product variants.`

	examples = `# Add a variant 'xs' for color 'Blue'
# In the example below, option 'Color' and 'Size' must exist
$ shopctl product variant add 8856145494 -o"Color:Blue" -o"Size:xs"

# Add a variant of price 20, unit cost 10 and compare at price of 30
$ shopctl product variant add 8856145494 -o"Color:Black" -o"Size:s" --price 20 --unit-cost 10 --regular-price 30

# Add a variant with SKU and barcode that requires shipping with inventory tracked
$ shopctl product variant add 8856145494 -o"Color:Red" -o"Size:xl" -p20 --weight "GRAMS:100" --sku 123 --barcode 456 --tracked --requires-shipping`
)

// Flag wraps available command flags.
type flag struct {
	id               string
	sku              string
	price            string
	regularPrice     string
	options          []string
	unitCost         float64
	barcode          string
	weightUnit       schema.WeightUnit
	weightValue      float64
	isTracked        bool
	requiresShipping bool
	backorder        bool
	isTaxable        bool
	taxcode          string
	strategy         schema.ProductVariantsBulkCreateStrategy
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	sku, err := cmd.Flags().GetString("sku")
	cmdutil.ExitOnErr(err)

	price, err := cmd.Flags().GetFloat64("price")
	cmdutil.ExitOnErr(err)

	regularPrice, err := cmd.Flags().GetFloat64("regular-price")
	cmdutil.ExitOnErr(err)

	options, err := cmd.Flags().GetStringArray("option")
	cmdutil.ExitOnErr(err)

	cost, err := cmd.Flags().GetFloat64("unit-cost")
	cmdutil.ExitOnErr(err)

	barcode, err := cmd.Flags().GetString("barcode")
	cmdutil.ExitOnErr(err)

	weight, err := cmd.Flags().GetString("weight")
	cmdutil.ExitOnErr(err)

	var (
		weightUnit  schema.WeightUnit
		weightValue float64

		units = []string{
			string(schema.WeightUnitKilograms),
			string(schema.WeightUnitGrams),
			string(schema.WeightUnitPounds),
			string(schema.WeightUnitOunces),
		}
	)
	if weight != "" {
		unit, val, err := cmdutil.SplitKeyVal(weight)
		if err != nil {
			cmdutil.ExitOnErr(
				cmdutil.HelpErrorf("Weight should be in the folowing format Unit:Value, eg: GRAMS:3.14", examples),
			)
		}
		unit = strings.ToUpper(unit)
		if !slices.Contains(units, unit) {
			cmdutil.Fail("Weight unit should be one of: %s", strings.Join(units, ", "))
			os.Exit(1)
		}
		weightUnit = schema.WeightUnit(unit)
		weightValue, err = strconv.ParseFloat(val, 64)
		if err != nil {
			cmdutil.ExitOnErr(
				cmdutil.HelpErrorf("Weight value is invalid", examples),
			)
		}
	}

	isTracked, err := cmd.Flags().GetBool("tracked")
	cmdutil.ExitOnErr(err)

	requiresShipping, err := cmd.Flags().GetBool("requires-shipping")
	cmdutil.ExitOnErr(err)

	backorder, err := cmd.Flags().GetBool("allow-backorder")
	cmdutil.ExitOnErr(err)

	isTaxable, err := cmd.Flags().GetBool("taxable")
	cmdutil.ExitOnErr(err)

	taxcode, err := cmd.Flags().GetString("taxcode")
	cmdutil.ExitOnErr(err)

	rmStandalone, err := cmd.Flags().GetBool("remove-standalone")
	cmdutil.ExitOnErr(err)

	strategy := schema.ProductVariantsBulkCreateStrategyDefault
	if rmStandalone {
		strategy = schema.ProductVariantsBulkCreateStrategyRemoveStandaloneVariant
	}

	f.id = cmdutil.ShopifyProductID(args[0])
	f.sku = sku
	f.price = strconv.FormatFloat(price, 'f', -1, 64)
	f.regularPrice = strconv.FormatFloat(regularPrice, 'f', -1, 64)
	f.options = options
	f.unitCost = cost
	f.barcode = barcode
	f.weightUnit = weightUnit
	f.weightValue = weightValue
	f.isTracked = isTracked
	f.requiresShipping = requiresShipping
	f.backorder = backorder
	f.isTaxable = isTaxable
	f.taxcode = taxcode
	f.strategy = strategy
}

// NewCmdAdd constructs a new product variant add command.
func NewCmdAdd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "add PRODUCT_ID",
		Short:   "Add product variants",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"create"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringArrayP("option", "o", []string{}, "Variant option values")
	cmd.Flags().String("sku", "", "SKU (Stock Keeping Unit)")
	cmd.Flags().Float64P("price", "p", 0, "Price of the variant")
	cmd.Flags().Float64("regular-price", 0, "Compare at price of the variant")
	cmd.Flags().Float64("unit-cost", 0, "Cost per item")
	cmd.Flags().Bool("tracked", false, "Is the inventory item tracked?")
	cmd.Flags().Bool("requires-shipping", false, "Does the item requires shipping?")
	cmd.Flags().String("barcode", "", "Barcode (ISBN, UPC, GTIN, etc.)")
	cmd.Flags().String("weight", "", "The weight of the inventory item in format Unit:Value (eg: GRAMS:100)")
	cmd.Flags().Bool("allow-backorder", false, "Allow out-of-stock purchases (sets inventory policy to CONTINUE)")
	cmd.Flags().Bool("taxable", false, "Is the variant taxable?")
	cmd.Flags().String("taxcode", "", "The tax code associated with the variant")
	cmd.Flags().Bool("remove-standalone", false, "Remove the standalone variant when creating variants")

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	options := make([]any, 0, len(flag.options))
	for _, o := range flag.options {
		optName, name, err := cmdutil.SplitKeyVal(o)
		if err != nil {
			return fmt.Errorf("option values should be in the following format, OptionName:Value")
		}
		options = append(options, schema.VariantOptionValueInput{
			Name:       &name,
			OptionName: &optName,
		})
	}

	inventoryPolicy := schema.ProductVariantInventoryPolicyDeny
	if flag.backorder {
		inventoryPolicy = schema.ProductVariantInventoryPolicyContinue
	}

	var (
		measurement schema.InventoryItemMeasurementInput
		weight      schema.WeightInput
	)
	if flag.weightUnit != "" {
		weight.Unit = schema.WeightUnit(flag.weightUnit)
		weight.Value = flag.weightValue
		measurement = schema.InventoryItemMeasurementInput{
			Weight: &weight,
		}
	}

	input := schema.ProductVariantsBulkInput{
		Barcode:         &flag.barcode,
		Price:           &flag.price,
		CompareAtPrice:  &flag.regularPrice,
		InventoryPolicy: &inventoryPolicy,
		InventoryItem: &schema.InventoryItemInput{
			Sku:              &flag.sku,
			Cost:             &flag.unitCost,
			Tracked:          &flag.isTracked,
			RequiresShipping: &flag.requiresShipping,
			Measurement:      &measurement,
		},
		OptionValues: options,
		Taxable:      &flag.isTaxable,
		TaxCode:      &flag.taxcode,
	}
	res, err := client.CreateProductVariants(flag.id, []schema.ProductVariantsBulkInput{input}, flag.strategy)
	if err != nil {
		return err
	}

	cmdutil.Success("Variant added successfully to product: %s", res.Product.ID)
	return nil
}
