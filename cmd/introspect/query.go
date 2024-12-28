package main

import "fmt"

func getQuery(gqlType string) map[string]string {
	query := fmt.Sprintf(`{
	  __type(name: "%s") {
		name
		kind
		description
		fields {
		  name
		  description
		  type {
			name
			kind
			ofType {
			  name
			  kind
			}
		  }
		  args {
			name
			description
			type {
			  name
			  kind
			  ofType {
				name
				kind
			  }
			}
			defaultValue
		  }
		}
		enumValues {
		  name
		  description
		}
	  }
	}`, gqlType)

	return map[string]string{
		"query": query,
	}
}
