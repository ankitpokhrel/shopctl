package edit

import (
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
	helpText = `Edit product variants.`

	examples = `# Edit a variant 'xs' for color 'Blue'
$ shopctl product variant edit 8856145494 --title "Blue / xs"

# Edit a variant using variant id
$ shopctl product variant edit 8856145494 --id 471883718 --price 20

# Set a variant price to 20, unit cost to 10 and compare at price to 30
$ shopctl product variant edit 8856145494 -t"Black / s" --price 20 --unit-cost 10 --regular-price 30

# Edit SKU and barcode of a variant and enable shipping with inventory tracked
$ shopctl product variant add 8856145494 -t"Red / xl" -p20 --weight "GRAMS:100" --sku 123 --barcode 456 --tracked --requires-shipping

# Set tracked quantity of a variant to false
$ shopctl product variant edit 8856145494 -id 471883718 --tracked=false`
)

// Flag wraps available command flags.
type flag struct {
	id               string
	variantID        string
	title            string
	sku              *string
	price            *string
	regularPrice     *string
	unitCost         *float64
	barcode          *string
	weightUnit       *schema.WeightUnit
	weightValue      *float64
	isTracked        *bool
	requiresShipping *bool
	backorder        *bool
	isTaxable        *bool
	taxcode          *string
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	f.id = cmdutil.ShopifyProductID(args[0])

	isset := func(item string) bool {
		fl := cmd.Flags().Lookup(item)
		return fl != nil && fl.Changed
	}

	variantID, err := cmd.Flags().GetString("id")
	cmdutil.ExitOnErr(err)
	f.variantID = cmdutil.ShopifyProductVariantID(variantID)

	title, err := cmd.Flags().GetString("title")
	cmdutil.ExitOnErr(err)

	parts := strings.Split(title, "/")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	f.title = strings.Join(parts, " / ")

	if f.variantID == "" && f.title == "" {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Either variant ID or title is required", examples),
		)
	}

	if isset("sku") {
		sku, err := cmd.Flags().GetString("sku")
		cmdutil.ExitOnErr(err)

		f.sku = &sku
	}
	if isset("price") {
		price, err := cmd.Flags().GetFloat64("price")
		cmdutil.ExitOnErr(err)

		p64 := strconv.FormatFloat(price, 'f', -1, 64)
		f.price = &p64
	}
	if isset("regular-price") {
		regularPrice, err := cmd.Flags().GetFloat64("regular-price")
		cmdutil.ExitOnErr(err)

		p64 := strconv.FormatFloat(regularPrice, 'f', -1, 64)
		f.regularPrice = &p64
	}
	if isset("unit-cost") {
		cost, err := cmd.Flags().GetFloat64("unit-cost")
		cmdutil.ExitOnErr(err)

		f.unitCost = &cost
	}
	if isset("barcode") {
		barcode, err := cmd.Flags().GetString("barcode")
		cmdutil.ExitOnErr(err)

		f.barcode = &barcode
	}
	if isset("tracked") {
		isTracked, err := cmd.Flags().GetBool("tracked")
		cmdutil.ExitOnErr(err)

		f.isTracked = &isTracked
	}
	if isset("requires-shipping") {
		requiresShipping, err := cmd.Flags().GetBool("requires-shipping")
		cmdutil.ExitOnErr(err)

		f.requiresShipping = &requiresShipping
	}
	if isset("allow-backorder") { // Inventory policy.
		backorder, err := cmd.Flags().GetBool("allow-backorder")
		cmdutil.ExitOnErr(err)

		f.backorder = &backorder
	}
	if isset("taxable") {
		isTaxable, err := cmd.Flags().GetBool("taxable")
		cmdutil.ExitOnErr(err)

		f.isTaxable = &isTaxable
	}
	if isset("taxcode") {
		taxcode, err := cmd.Flags().GetString("taxcode")
		cmdutil.ExitOnErr(err)

		f.taxcode = &taxcode
	}

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
		f.weightUnit = &weightUnit
		f.weightValue = &weightValue
	}
}

// NewCmdEdit constructs a new product variant edit command.
func NewCmdEdit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "edit PRODUCT_ID",
		Short:   "Edit product variants",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"update"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().String("id", "", "ID of the variant to edit")
	cmd.Flags().StringP("title", "t", "", "Title of the variant to edit")
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

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	var (
		variant *schema.ProductVariant
		err     error
	)

	if flag.variantID != "" {
		variant, err = client.GetProductVariantByID(flag.variantID)
	} else {
		variant, err = client.GetProductVariantByTitle(flag.id, flag.title, true)
	}
	if err != nil {
		return err
	}
	flag.variantID = variant.ID

	input := getInput(*flag, variant)
	res, err := client.UpdateProductVariants(
		flag.id, []schema.ProductVariantsBulkInput{input},
		false, /* We don't allow partial update */
	)
	if err != nil {
		return err
	}

	cmdutil.Success("Variant %q of product %s updated successfully", variant.Title, res.Product.ID)
	return nil
}

func getInput(f flag, v *schema.ProductVariant) schema.ProductVariantsBulkInput {
	getInventoryData := func() *schema.InventoryItemInput {
		inventory := schema.InventoryItemInput{}
		isdirty := false
		if f.sku != nil {
			isdirty = true
			inventory.Sku = f.sku
		}
		if f.unitCost != nil {
			isdirty = true
			inventory.Cost = f.unitCost
		}
		if f.isTracked != nil {
			isdirty = true
			inventory.Tracked = f.isTracked
		}
		if f.requiresShipping != nil {
			isdirty = true
			inventory.RequiresShipping = f.requiresShipping
		}
		if f.weightUnit != nil || f.weightValue != nil {
			isdirty = true
			var (
				unit schema.WeightUnit
				val  float64
			)
			if v.InventoryItem != nil {
				unit = v.InventoryItem.Measurement.Weight.Unit
				val = v.InventoryItem.Measurement.Weight.Value
			}
			if f.weightUnit != nil {
				unit = *f.weightUnit
			}
			if f.weightValue != nil {
				val = *f.weightValue
			}
			inventory.Measurement = &schema.InventoryItemMeasurementInput{
				Weight: &schema.WeightInput{
					Unit:  unit,
					Value: val,
				},
			}
		}

		if !isdirty {
			return nil
		}
		return &inventory
	}

	input := schema.ProductVariantsBulkInput{ID: &f.variantID}

	if f.barcode != nil {
		input.Barcode = f.barcode
	}
	if f.price != nil {
		input.Price = f.price
	}
	if f.regularPrice != nil {
		input.CompareAtPrice = f.regularPrice
	}
	if f.backorder != nil && *f.backorder {
		policy := schema.ProductVariantInventoryPolicyContinue
		input.InventoryPolicy = &policy
	}

	if inventory := getInventoryData(); inventory != nil {
		input.InventoryItem = inventory
	}
	return input
}
