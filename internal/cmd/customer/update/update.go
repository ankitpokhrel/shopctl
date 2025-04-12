package update

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Update lets you update a customer.`

	examples = `$ shopctl customer update 8856145494 --first-name Jon --last-name Doe

# Update customer tags and note
$ shopctl customer update 8856145494 --email janedoe@example.com --note "Example user updated" --tags example,dummy

# Update customer default address (accepts tagged fields)
# See https://shopify.dev/docs/api/admin-graphql/latest/input-objects/MailingAddressInput for the list of accepted tags
$ shopctl customer update 8856145494 -lHolmes --default-address "221B Baker Street country:GB city:London zip:NW1"

# Update customer with metafields (accepts tagged fields)
# See https://shopify.dev/docs/apps/build/custom-data/metafields/list-of-data-types#supported-types for valid metafield types
$ shopctl customer update 8856145494 -fJane -lDoe --meta "custom.preferred_color:#95BF47 type:color"`
)

type address struct {
	address1     *string
	address2     *string
	city         *string
	company      *string
	countryCode  *schema.CountryCode
	firstName    *string
	lastName     *string
	phone        *string
	provinceCode *string
	zip          *string
}

type metafield struct {
	namespace string
	key       string
	value     string
	typ       string
}

type flag struct {
	id            string
	fname         *string
	lname         *string
	email         *string
	phone         *string
	note          *string
	locale        *string
	tags          []string
	address       *address
	metafields    []metafield
	taxExempt     *bool
	taxExemptions []string
	web           bool
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	isset := func(field string) bool {
		fl := cmd.Flags().Lookup(field)
		return fl != nil && fl.Changed
	}

	if isset("first-name") {
		fname, err := cmd.Flags().GetString("first-name")
		cmdutil.ExitOnErr(err)

		f.fname = &fname
	}

	if isset("last-name") {
		lname, err := cmd.Flags().GetString("last-name")
		cmdutil.ExitOnErr(err)

		f.lname = &lname
	}

	if isset("email") {
		email, err := cmd.Flags().GetString("email")
		cmdutil.ExitOnErr(err)

		f.email = &email
	}

	if isset("phone") {
		phone, err := cmd.Flags().GetString("phone")
		cmdutil.ExitOnErr(err)

		f.phone = &phone
	}

	if isset("note") {
		note, err := cmd.Flags().GetString("note")
		cmdutil.ExitOnErr(err)

		f.note = &note
	}

	if isset("locale") {
		locale, err := cmd.Flags().GetString("locale")
		cmdutil.ExitOnErr(err)

		f.locale = &locale
	}

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	address, err := cmd.Flags().GetString("default-address")
	cmdutil.ExitOnErr(err)

	addr, err := parseAddress(address)
	if err != nil {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Invalid address format: "+err.Error(), examples),
		)
	}

	meta, err := cmd.Flags().GetStringArray("meta")
	cmdutil.ExitOnErr(err)

	metafields := make([]metafield, 0, len(meta))
	for _, mf := range meta {
		m, err := parseMetafield(mf)
		if err != nil {
			cmdutil.ExitOnErr(
				cmdutil.HelpErrorf("Invalid metafield format: "+err.Error(), examples),
			)
		}
		metafields = append(metafields, *m)
	}

	if isset("tax-exempt") {
		isTaxExempt, err := cmd.Flags().GetBool("tax-exempt")
		cmdutil.ExitOnErr(err)

		f.taxExempt = &isTaxExempt
	}

	taxExemptions, err := cmd.Flags().GetString("tax-exemptions")
	cmdutil.ExitOnErr(err)

	web, err := cmd.Flags().GetBool("web")
	cmdutil.ExitOnErr(err)

	f.id = shopctl.ShopifyCustomerID(args[0])
	f.tags = strings.Split(tags, ",")
	f.address = addr
	f.metafields = metafields
	f.taxExemptions = func() []string {
		if strings.TrimSpace(taxExemptions) != "" {
			return strings.Split(taxExemptions, ",")
		}
		return nil
	}()
	f.web = web
}

// NewCmdUpdate constructs a new customer update command.
func NewCmdUpdate() *cobra.Command {
	cmd := cobra.Command{
		Use:     "update CUSTOMER_ID",
		Short:   "Update a customer",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"edit"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context().Value(cmdutil.KeyContext).(*config.StoreContext)
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, ctx, client))
			return nil
		},
	}
	cmd.Flags().StringP("first-name", "f", "", "Customer's first name")
	cmd.Flags().StringP("last-name", "l", "", "Customer's last name")
	cmd.Flags().StringP("email", "e", "", "The unique email address of the customer")
	cmd.Flags().StringP("phone", "p", "", "The unique phone number for the customer")
	cmd.Flags().String("note", "", "A note about the customer")
	cmd.Flags().String("locale", "", "Customer's locale")
	cmd.Flags().String("tags", "", "Comma separated list of tags associated with the customer")
	cmd.Flags().StringP("default-address", "a", "", "Customer default address (supports tagged fields)")
	cmd.Flags().StringArray("meta", []string{}, "Customer metafields (supports tagged fields)")
	cmd.Flags().Bool("tax-exempt", false, "Is the customer exempt from paying taxes on their order")
	cmd.Flags().String("tax-exemptions", "", "Comma separated list of tax exemptions to apply")
	cmd.Flags().Bool("web", false, "Open in web browser after successful creation")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	var defaultAddressID string

	flag := &flag{}
	flag.parse(cmd, args)

	customer, err := client.GetCustomerByID(flag.id)
	if err != nil {
		return fmt.Errorf("customer not found")
	}
	if customer.DefaultAddress != nil {
		defaultAddressID = customer.DefaultAddress.ID
	} else {
		flag.address = nil
	}
	input := getCustomerInput(*flag, customer)

	res, err := client.UpdateCustomer(*input)
	if err != nil {
		return err
	}
	if flag.address != nil && defaultAddressID != "" {
		addressInput := getMailingAddressInput(flag.address)
		_, err := client.UpdateCustomerAddress(flag.id, defaultAddressID, *addressInput, true)
		if err != nil {
			return fmt.Errorf("customer updated but default address update falied: %w", err)
		}
	}

	adminURL := fmt.Sprintf(
		"https://admin.shopify.com/store/%s/customers/%s",
		ctx.Alias, shopctl.ExtractNumericID(res.Customer.ID),
	)
	if flag.web {
		_ = browser.Browse(adminURL)
	}

	cmdutil.Success("Customer updated successfully: %s", res.Customer.ID)
	fmt.Println(adminURL)

	return nil
}

func parseAddress(input string) (*address, error) {
	var addr address

	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("address input is empty")
	}

	isTag := false
	var addr1Tokens []string

	for _, token := range tokens {
		switch {
		case strings.Contains(token, ":"):
			isTag = true
			parts := strings.SplitN(token, ":", 2)
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])

			switch key {
			case "address1":
				addr.address1 = &value
			case "address2":
				addr.address2 = &value
			case "city":
				addr.city = &value
			case "company":
				addr.company = &value
			case "country", "countrycode":
				if value != "" {
					cc := schema.CountryCode(value)
					addr.countryCode = &cc
				}
			case "firstname", "first", "fname":
				addr.firstName = &value
			case "lastname", "last", "lname":
				addr.lastName = &value
			case "phone":
				addr.phone = &value
			case "province", "provincecode", "state":
				val := strings.ToUpper(value)
				addr.provinceCode = &val
			case "zip", "post", "postcode":
				addr.zip = &value
			default:
				return nil, fmt.Errorf("unrecognized address field: %s", key)
			}
		case !isTag:
			addr1Tokens = append(addr1Tokens, token)
		default:
			return nil, fmt.Errorf("unexpected untagged token after tagged fields: %s", token)
		}
	}

	if len(addr1Tokens) > 0 {
		addr1 := strings.Join(addr1Tokens, " ")
		addr.address1 = &addr1
	}
	return &addr, nil
}

func parseMetafield(input string) (*metafield, error) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("metafield input is empty")
	}

	mf := metafield{}

	// Process remaining tagged tokens
	for _, t := range tokens {
		if !strings.Contains(t, ":") {
			return nil, fmt.Errorf("invalid tagged field: %s", t)
		}
		kv := strings.SplitN(t, ":", 2)
		k := strings.ToLower(kv[0])
		v := kv[1]

		switch {
		case strings.Contains(k, "."):
			nv := strings.SplitN(k, ".", 2)
			mf.namespace = nv[0]
			mf.key = nv[1]
			mf.value = v
		case k == "type":
			mf.typ = v
		default:
			return nil, fmt.Errorf("unsupported field: %s", k)
		}
	}
	return &mf, nil
}

func getCustomerInput(f flag, c *schema.Customer) *schema.CustomerInput {
	tagSet := make(map[string]struct{})
	for _, tag := range c.Tags {
		tagSet[tag.(string)] = struct{}{}
	}
	for _, tag := range f.tags {
		if strings.HasPrefix(tag, "-") {
			delete(tagSet, strings.TrimPrefix(tag, "-"))
		} else {
			tagSet[tag] = struct{}{}
		}
	}
	tags := make([]any, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}

	var exemptions []any
	if f.taxExemptions != nil {
		for _, e := range f.taxExemptions {
			exemptions = append(exemptions, e)
		}
	}

	metafields := make([]any, 0, len(f.metafields))
	for _, m := range f.metafields {
		metafields = append(metafields, schema.MetafieldInput{
			Namespace: &m.namespace,
			Key:       &m.key,
			Value:     &m.value,
			Type:      &m.typ,
		})
	}

	input := schema.CustomerInput{
		ID:                    &f.id,
		FirstName:             f.fname,
		LastName:              f.lname,
		Email:                 f.email,
		Phone:                 f.phone,
		Metafields:            metafields,
		Locale:                f.locale,
		Note:                  f.note,
		Tags:                  tags,
		EmailMarketingConsent: nil, // We are not going to set marketing consents.
		SmsMarketingConsent:   nil,
		TaxExempt:             f.taxExempt,
		TaxExemptions:         exemptions,
	}
	return &input
}

func getMailingAddressInput(a *address) *schema.MailingAddressInput {
	return &schema.MailingAddressInput{
		Address1:     a.address1,
		Address2:     a.address2,
		City:         a.city,
		Company:      a.company,
		CountryCode:  a.countryCode,
		FirstName:    a.firstName,
		LastName:     a.lastName,
		Phone:        a.phone,
		ProvinceCode: a.provinceCode,
		Zip:          a.zip,
	}
}
