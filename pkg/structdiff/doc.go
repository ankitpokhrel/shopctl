// Package structdiff generates diff between two structs.
//
// This package performs a field-by-field comparison of two structs of same type and
// generate diff using the Myers algorithm: http://www.xmailserver.org/diff2.pdf
//
// This package utilizes `pkg/diff` under the hood and the output generated
// is compatible with tools that understand git diff output format.
//
// Example usage:
//
//	diff := structdiff.Get(a, b) // a and b must be struct of same type.
package structdiff
