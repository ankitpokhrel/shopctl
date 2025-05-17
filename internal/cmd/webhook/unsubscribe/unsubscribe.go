package unsubscribe

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

const (
	helpText = `Unsubscribe lets you delete a webhook subscription.`

	examples = `$ shopctl webhook unsubscribe 123456789
$ shopctl webhook unsubscribe gid://shopify/WebhookSubscription/123456789`
)

// NewCmdUnsubscribe constructs a new webhook subscription delete command.
func NewCmdUnsubscribe() *cobra.Command {
	return &cobra.Command{
		Use:     "unsubscribe WEBHOOK_ID",
		Short:   "Delete a webhook subscription",
		Long:    helpText,
		Example: examples,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"unsub", "delete", "del"},
		Annotations: map[string]string{
			"help:args": `WEBHOOK_ID full or numeric webhook subscription ID, eg: 88561444456 or gid://shopify/WebhookSubscription/88561444456`,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}
}

func run(_ *cobra.Command, args []string, client *api.GQLClient) error {
	subID := shopctl.ShopifyWebhookSubscriptionID(args[0])

	_, err := client.DeleteWebhook(subID)
	if err != nil {
		return err
	}

	cmdutil.Success("Webhook unsubscribed successfully")
	return nil
}
