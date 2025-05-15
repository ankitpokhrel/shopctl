package api

import (
	"context"
	"fmt"

	"github.com/ankitpokhrel/shopctl/pkg/gql/client"
	"github.com/ankitpokhrel/shopctl/schema"
)

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
