package compare

import (
	"strings"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

// Render renders the diff.
func Render(diffs map[string]string, sortOrder []string) error {
	var out strings.Builder

	// Append items from sortOrder.
	seen := make(map[string]struct{})
	for _, field := range sortOrder {
		if d, ok := diffs[field]; ok {
			seen[field] = struct{}{}
			out.WriteString(d)
			out.WriteRune('\n')
		}
	}

	// Append rest of the items.
	for k, d := range diffs {
		if _, ok := seen[k]; ok {
			continue
		}
		out.WriteString(d)
		out.WriteRune('\n')
	}

	return cmdutil.DiffOut(out.String())
}
