// Package client provides a GraphQL client for making requests to a GraphQL server.
//
// Example usage:
//
//	client := client.NewClient("https://example.com/graphql", "your-access-token")
//	headers := client.Header{
//	    "Custom-Header": "value",
//	}
//	body := []byte(`{ "query": "{ exampleQuery { field } }" }`)
//	resp, err := client.Request(context.Background(), body, headers)
//	if err != nil {
//	    log.Fatalf("request error: %v", err)
//	}
//	defer resp.Body.Close()
//	// Process response...
package client
