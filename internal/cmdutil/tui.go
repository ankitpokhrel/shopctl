package cmdutil

import (
	"fmt"
	"os"
)

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
