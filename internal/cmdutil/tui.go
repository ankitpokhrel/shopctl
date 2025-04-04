package cmdutil

import (
	"fmt"
	"os"
)

const (
	RepeatedDashes = "" +
		"-------------------------------"
	RepeatedDashesSM = "" +
		"-------------------"
	RepeatedEquals = "" +
		"==============================="
)

// SummaryTitle displays msg between separators.
func SummaryTitle(msg string, sep string) {
	fmt.Printf("%s\n%s\n%s\n", sep, msg, sep)
}

// SummarySubtitle displays msg and a separator underneath.
func SummarySubtitle(msg string, sep string) {
	fmt.Printf("%s\n%s\n", msg, sep)
}

// Success prints success message in stdout.
func Success(msg string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, fmt.Sprintf("\u001B[0;32m✔\u001B[0m %s\n", msg), args...)
}

// Warn prints warning message in stderr.
func Warn(msg string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\u001B[0;33m%s\u001B[0m\n", msg), args...)
}

// Fail prints failure message in stderr.
func Fail(msg string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\u001B[0;31m✗\u001B[0m %s\n", msg), args...)
}
