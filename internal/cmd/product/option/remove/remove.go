package remove

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Remove product options.`

	examples = `$ shopct product option remove 8856145494 -nSize -nTitle -nStyle`
)

// Flag wraps available command flags.
type flag struct {
	id      string
	options []string
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	options, err := cmd.Flags().GetStringArray("name")
	cmdutil.ExitOnErr(err)

	if len(options) == 0 {
		cmdutil.ExitOnErr(cmdutil.HelpErrorf("Name of options to delete is required", examples))
	}

	f.id = cmdutil.ShopifyProductID(args[0])
	f.options = options
}

// NewCmdRemove constructs a new product option remove command.
func NewCmdRemove() *cobra.Command {
	cmd := cobra.Command{
		Use:     "remove PRODUCT_ID",
		Short:   "Remove product options",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"delete", "del", "rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringArrayP("name", "n", []string{}, "Name of options to remove")

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	productOptions, err := client.GetProductOptions(flag.id)
	if err != nil {
		return err
	}

	productOptionsMap := make(map[string]*schema.ProductOption, 0)
	for _, o := range productOptions.Data.Product.Options {
		productOptionsMap[strings.ToLower(o.Name)] = &o
	}

	optionsToDelete := make([]string, 0)
	optionsProcessed := make([]string, 0)
	for _, n := range flag.options {
		if o, ok := productOptionsMap[strings.ToLower(n)]; ok {
			optionsToDelete = append(optionsToDelete, o.ID)
			optionsProcessed = append(optionsProcessed, o.Name)
		}
	}

	if len(optionsToDelete) == 0 {
		cmdutil.Warn("Nothing to delete")
		os.Exit(0)
	}

	res, err := client.DeleteProductOptions(flag.id, optionsToDelete)
	if err != nil {
		return err
	}

	cmdutil.Success(
		"Options %q deleted successfully for product: %s", strings.Join(optionsProcessed, ", "), res.Product.ID,
	)
	return nil
}
