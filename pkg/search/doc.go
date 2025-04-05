// Package search provides a query builder for constructing Shopify search syntax queries.
// It offers a simple and fluent API to build complex queries by chaining conditions.
// See https://shopify.dev/docs/api/usage/search-syntax for Shopify search syntax details.
//
// Example usage:
//
//	// Query: title:"red shirt"
//	q1 := search.New().Equal("title", "red shirt")
//	fmt.Println(q1.Build())  // Output: title:"red shirt"
//
//	// Query: title:"red shirt" AND price:>10
//	q2 := search.New().
//	        Equal("title", "red shirt").
//	        And().
//	        GreaterThan("price", 10)
//	fmt.Println(q2.Build())  // Output: title:"red shirt" AND price:>10
//
//	// Query: (product_type:t-shirt OR product_type:sweater)
//	q3 := search.New().In("product_type", "t-shirt", "sweater")
//	fmt.Println(q3.Build())  // Output: (product_type:"t-shirt" OR product_type:"sweater")
//
//	// Query: (title:"red shirt" AND description:*cotton*) AND price:<=20
//	q4 := search.New().
//	        Group(func(sub *search.Query) {
//	            sub.Equal("title", "red shirt").
//	                And().
//	                Contains("description", "cotton")
//	        }).
//	        And().
//	        LessThanOrEqual("price", 20)
//	fmt.Println(q4.Build())  // Output: ((title:"red shirt" AND description:*cotton*)) AND price:<=20
//
//	// Query: -status:"sold out" OR status:available
//	q5 := search.New().
//	        NotEqual("status", "sold out").
//	        Or().
//	        Equal("status", "available")
//	fmt.Println(q5.Build())  // Output: -status:"sold out" OR status:available
package search
