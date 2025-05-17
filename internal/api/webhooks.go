package api

import (
	"context"
	"fmt"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

// ErrAddrTaken is thrown if the webhook subscription address is already registered.
var ErrAddrTaken = fmt.Errorf("Address for this topic has already been taken") //nolint:staticcheck

// GetWebhooks fetches n number of webhooks after a cursor.
func (c GQLClient) GetWebhooks(limit int, after *string, topics []schema.WebhookSubscriptionTopic, query *string) ([]schema.WebhookSubscription, error) {
	var out struct {
		Data struct {
			WebhookSubscriptions WebhookData `json:"webhookSubscriptions"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	webhookQuery := fmt.Sprintf(`query WebhookSubscriptionsList($first: Int!, $after: String, $topics: [WebhookSubscriptionTopic!], $query: String, $format: WebhookSubscriptionFormat, $sortKey: WebhookSubscriptionSortKeys!, $reverse: Boolean!) {
  webhookSubscriptions(first: $first, after: $after, topics: $topics, query: $query, format: $format, sortKey: $sortKey, reverse: $reverse) {
    nodes {
      %s
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`, fieldsWebhook)

	req := client.GQLRequest{
		Query: webhookQuery,
		Variables: client.QueryVars{
			"first":   limit,
			"after":   after,
			"query":   query,
			"topics":  topics,
			"format":  schema.WebhookSubscriptionFormatJson,
			"sortKey": "CREATED_AT",
			"reverse": true,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return out.Data.WebhookSubscriptions.Nodes, nil
}

// SubscribeWebhook subscribes to a webhook.
func (c GQLClient) SubscribeWebhook(topic string, endpoint string) (*schema.WebhookSubscription, error) {
	var out struct {
		Data struct {
			WebhookSubscriptionCreate WebhookSyncResponse `json:"webhookSubscriptionCreate"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := `
    mutation webhookSubscriptionCreate($topic: WebhookSubscriptionTopic!, $webhookSubscription: WebhookSubscriptionInput!) {
      webhookSubscriptionCreate(topic: $topic, webhookSubscription: $webhookSubscription) {
        webhookSubscription {
          id
          topic
          apiVersion { handle }
          endpoint {
            __typename ... on WebhookHttpEndpoint { callbackUrl }
          }
        }
        userErrors {
          field
          message
        }
      }
    }`

	format := schema.WebhookSubscriptionFormatJson

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"topic": topic,
			"webhookSubscription": schema.WebhookSubscriptionInput{
				CallbackURL: &endpoint,
				Format:      &format,
			},
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	if len(out.Data.WebhookSubscriptionCreate.UserErrors) > 0 {
		usrErr := out.Data.WebhookSubscriptionCreate.UserErrors.Error()
		if usrErr == ErrAddrTaken.Error() {
			return &out.Data.WebhookSubscriptionCreate.WebhookSubscription, ErrAddrTaken
		}
		return nil, fmt.Errorf("webhookSubscriptionCreate: The operation failed with user error: %s", usrErr)
	}
	return &out.Data.WebhookSubscriptionCreate.WebhookSubscription, nil
}

// GetWebhookByID fetches webhook by its ID.
func (c GQLClient) GetWebhookByID(id string) (*schema.WebhookSubscription, error) {
	var out struct {
		Data struct {
			WebhookSubscription schema.WebhookSubscription `json:"webhookSubscription"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := fmt.Sprintf(`query WebhookSubscription($id: ID!) {
        webhookSubscription(id: $id) {
          %s
      }
    }`, fieldsWebhook)

	req := client.GQLRequest{
		Query:     query,
		Variables: client.QueryVars{"id": id},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	return &out.Data.WebhookSubscription, nil
}

// DeleteWebhook deletes a Webhook by ID.
func (c GQLClient) DeleteWebhook(id string) (*WebhookDeleteResponse, error) {
	var out struct {
		Data struct {
			WebhookSubscriptionDelete WebhookDeleteResponse `json:"webhookSubscriptionDelete"`
		} `json:"data"`
		Errors Errors `json:"errors"`
	}

	query := `
    mutation webhookSubscriptionDelete($id: ID!) {
      webhookSubscriptionDelete(id: $id) {
        deletedWebhookSubscriptionId
        userErrors {
          field
          message
        }
      }
    }`

	req := client.GQLRequest{
		Query: query,
		Variables: client.QueryVars{
			"id": id,
		},
	}
	if err := c.Execute(context.Background(), req, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("%s", out.Errors)
	}
	if len(out.Data.WebhookSubscriptionDelete.UserErrors) > 0 {
		return nil, fmt.Errorf("webhookSubscriptionDelete: The operation failed with user error: %s", out.Data.WebhookSubscriptionDelete.UserErrors.Error())
	}
	return &out.Data.WebhookSubscriptionDelete, nil
}
