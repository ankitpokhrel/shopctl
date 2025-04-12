package delete

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

const (
	helpText = `Delete lets you delete a customer by ID, email or phone.`

	examples = `# Delete customer by its ID
$ shopctl customer delete 8370159190
$ shopctl customer delete gid://shopify/Customer/8370159190

# Delete customer by its email
$ shopctl customer delete --email example@domain.com

# Delete customer by its phone number
$ shopctl customer delete --phone +1234567890`
)

type flag struct {
	id    string
	email string
	phone string
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	email, err := cmd.Flags().GetString("email")
	cmdutil.ExitOnErr(err)

	phone, err := cmd.Flags().GetString("phone")
	cmdutil.ExitOnErr(err)

	if len(args) > 0 {
		f.id = cmdutil.ShopifyCustomerID(args[0])
	}
	f.email = email
	f.phone = phone

	if f.id == "" && f.email == "" && f.phone == "" {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Either a valid id, email or phone of the customer to delete is required.", examples),
		)
	}
}

// NewCmdDelete constructs a new customer create command.
func NewCmdDelete() *cobra.Command {
	cmd := cobra.Command{
		Use:     "delete [CUSTOMER_ID]",
		Short:   "Delete a customer",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"del", "rm", "remove"},
		Annotations: map[string]string{
			"help:args": `CUSTOMER_ID full or numeric Customer ID, eg: 88561444456 or gid://shopify/Customer/88561444456`,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
	cmd.Flags().StringP("email", "e", "", "The email address of the customer")
	cmd.Flags().StringP("phone", "p", "", "The phone number of the customer")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	var customerID string

	switch {
	case flag.id != "":
		customerID = flag.id
	case flag.email != "":
		customer, err := client.CheckCustomerByEmailOrPhone(&flag.email, nil)
		if err != nil {
			return err
		}
		customerID = customer.ID
	case flag.phone != "":
		customer, err := client.CheckCustomerByEmailOrPhone(nil, &flag.phone)
		if err != nil {
			return err
		}
		customerID = customer.ID
	}

	_, err := client.DeleteCustomer(customerID)
	if err != nil {
		return err
	}

	cmdutil.Success("Customer deleted successfully: %s", customerID)
	return nil
}
