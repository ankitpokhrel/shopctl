package subscribe

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

const (
	helpText = `Subscribe lets you subscribe to a webhook event.`

	examples = `$ shopctl webhook subscribe --topic PRODUCTS_CREATE --url https://example.com/products/create

# Subscribe webhook for customers update event
$ shopctl webhook subscribe --topic CUSTOMERS_UPDATE --url https://example.com:8080/products/update`
)

type flag struct {
	topic string
	url   string
}

func (f *flag) parse(cmd *cobra.Command) {
	topic, err := cmd.Flags().GetString("topic")
	cmdutil.ExitOnErr(err)

	url, err := cmd.Flags().GetString("url")
	cmdutil.ExitOnErr(err)

	f.topic = topic
	f.url = url
}

// NewCmdSubscribe constructs a new webhook subscription command.
func NewCmdSubscribe() *cobra.Command {
	cmd := cobra.Command{
		Use:     "subscribe",
		Short:   "Subscribe to a webhook event",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"sub", "create"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}

	cmd.Flags().StringP("topic", "t", "", "Webhook topic to listen to")
	cmd.Flags().String("url", "", "Endpoint for the webhook registration")

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, _ []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd)

	res, err := client.SubscribeWebhook(flag.topic, flag.url)
	if err != nil {
		return err
	}

	cmdutil.Success("Webhook subscribed successfully: %s", res.ID)
	return nil
}
