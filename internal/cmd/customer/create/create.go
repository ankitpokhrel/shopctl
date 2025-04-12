package create

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/pkg/browser"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Create lets you create a customer.`

	examples = `$ shopctl customer create --first-name Jon --last-name Doe

# Create customer with tags and a note
$ shopctl customer create --email janedoe@example.com --note "Example user" --tags example,dummy

# Create customer with multiple addresses (accepts tagged fields)
# See https://shopify.dev/docs/api/admin-graphql/latest/input-objects/MailingAddressInput for the list of accepted tags
$ shopctl customer create -lHolmes --address "221B Baker Street country:GB city:London zip:NW1" --address "country:NP firstname:Jon lastname:Doe"

# Create customer with metafields (accepts tagged fields)
# See https://shopify.dev/docs/apps/build/custom-data/metafields/list-of-data-types#supported-types for valid metafield types
$ shopctl customer create -fJane -lDoe --meta "custom.preferred_color:#95BF47 type:color"`
)

type address struct {
	address1     string
	address2     string
	city         string
	company      string
	countryCode  *schema.CountryCode
	firstName    string
	lastName     string
	phone        string
	provinceCode string
	zip          string
}

type metafield struct {
	namespace string
	key       string
	value     string
	typ       string
}

type flag struct {
	fname         string
	lname         string
	email         string
	phone         string
	note          string
	locale        string
	tags          []string
	address       []address
	metafields    []metafield
	taxExempt     bool
	taxExemptions []string
	web           bool
}

func (f *flag) parse(cmd *cobra.Command, _ []string) {
	fname, err := cmd.Flags().GetString("first-name")
	cmdutil.ExitOnErr(err)

	lname, err := cmd.Flags().GetString("last-name")
	cmdutil.ExitOnErr(err)

	email, err := cmd.Flags().GetString("email")
	cmdutil.ExitOnErr(err)

	phone, err := cmd.Flags().GetString("phone")
	cmdutil.ExitOnErr(err)

	if fname == "" && lname == "" && email == "" && phone == "" {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Customer must have a first name, last name, phone number or email address", examples),
		)
	}

	note, err := cmd.Flags().GetString("note")
	cmdutil.ExitOnErr(err)

	locale, err := cmd.Flags().GetString("locale")
	cmdutil.ExitOnErr(err)

	tags, err := cmd.Flags().GetString("tags")
	cmdutil.ExitOnErr(err)

	addr, err := cmd.Flags().GetStringArray("address")
	cmdutil.ExitOnErr(err)

	addresses := make([]address, 0, len(addr))
	for _, ad := range addr {
		a, err := parseAddress(ad)
		if err != nil {
			cmdutil.ExitOnErr(
				cmdutil.HelpErrorf("Invalid address format: "+err.Error(), examples),
			)
		}
		addresses = append(addresses, *a)
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

	isTaxExempt, err := cmd.Flags().GetBool("tax-exempt")
	cmdutil.ExitOnErr(err)

	taxExemptions, err := cmd.Flags().GetString("tax-exemptions")
	cmdutil.ExitOnErr(err)

	web, err := cmd.Flags().GetBool("web")
	cmdutil.ExitOnErr(err)

	f.fname = fname
	f.lname = lname
	f.email = email
	f.phone = phone
	f.note = note
	f.locale = locale
	f.tags = strings.Split(tags, ",")
	f.address = addresses
	f.metafields = metafields
	f.taxExempt = isTaxExempt
	f.taxExemptions = func() []string {
		if strings.TrimSpace(taxExemptions) != "" {
			return strings.Split(taxExemptions, ",")
		}
		return nil
	}()
	f.web = web
}

// NewCmdCreate constructs a new customer create command.
func NewCmdCreate() *cobra.Command {
	cmd := cobra.Command{
		Use:     "create",
		Short:   "Create a customer",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"add"},
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
	cmd.Flags().StringArrayP("address", "a", []string{}, "Customer address (supports tagged fields)")
	cmd.Flags().StringArray("meta", []string{}, "Customer metafields (supports tagged fields)")
	cmd.Flags().Bool("tax-exempt", false, "Is the customer exempt from paying taxes on their order")
	cmd.Flags().String("tax-exemptions", "", "Comma separated list of tax exemptions to apply")
	cmd.Flags().Bool("web", false, "Open in web browser after successful creation")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string, ctx *config.StoreContext, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	tags := make([]any, len(flag.tags))
	for _, t := range flag.tags {
		tags = append(tags, t)
	}
	var exemptions []any
	if flag.taxExemptions != nil {
		for _, e := range flag.taxExemptions {
			exemptions = append(exemptions, e)
		}
	}

	addresses := make([]any, 0, len(flag.address))
	for _, a := range flag.address {
		addresses = append(addresses, schema.MailingAddressInput{
			Address1:     &a.address1,
			Address2:     &a.address2,
			FirstName:    &a.firstName,
			LastName:     &a.lastName,
			City:         &a.city,
			CountryCode:  a.countryCode,
			Company:      &a.company,
			Phone:        &a.phone,
			ProvinceCode: &a.provinceCode,
			Zip:          &a.zip,
		})
	}
	metafields := make([]any, 0, len(flag.metafields))
	for _, m := range flag.metafields {
		metafields = append(metafields, schema.MetafieldInput{
			Namespace: &m.namespace,
			Key:       &m.key,
			Value:     &m.value,
			Type:      &m.typ,
		})
	}

	input := schema.CustomerInput{
		FirstName:             &flag.fname,
		LastName:              &flag.lname,
		Email:                 &flag.email,
		Phone:                 &flag.phone,
		Addresses:             addresses,
		Metafields:            metafields,
		Locale:                &flag.locale,
		Note:                  &flag.note,
		Tags:                  tags,
		EmailMarketingConsent: nil, // We are not going to set marketing consents.
		SmsMarketingConsent:   nil,
		TaxExempt:             &flag.taxExempt,
		TaxExemptions:         exemptions,
	}

	res, err := client.CreateCustomer(input)
	if err != nil {
		return err
	}

	adminURL := fmt.Sprintf(
		"https://admin.shopify.com/store/%s/customers/%s",
		ctx.Alias, cmdutil.ExtractNumericID(res.Customer.ID),
	)
	if flag.web {
		_ = browser.Browse(adminURL)
	}

	cmdutil.Success("Customer created successfully: %s", res.Customer.ID)
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
				addr.address1 = value
			case "address2":
				addr.address2 = value
			case "city":
				addr.city = value
			case "company":
				addr.company = value
			case "country", "countrycode":
				if value != "" {
					cc := schema.CountryCode(value)
					addr.countryCode = &cc
				}
			case "firstname", "first", "fname":
				addr.firstName = value
			case "lastname", "last", "lname":
				addr.lastName = value
			case "phone":
				addr.phone = value
			case "province", "provincecode", "state":
				addr.provinceCode = strings.ToUpper(value)
			case "zip", "post", "postcode":
				addr.zip = value
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
		addr.address1 = strings.Join(addr1Tokens, " ")
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
