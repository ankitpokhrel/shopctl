package compare

import (
	"sort"
	"strings"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
)

// Render renders the diff.
func Render(diffs map[string]string, sortOrder []string) error {
	var out strings.Builder

	// Group diffs by parent field.
	grouped := make(map[string][]string)
	for k := range diffs {
		parent := extractParent(k)
		grouped[parent] = append(grouped[parent], k)
	}

	// Ensure the keys in each group are sorted.
	for _, keys := range grouped {
		sort.Strings(keys)
	}

	// Append items from sortOrder.
	seen := make(map[string]struct{})
	for _, field := range sortOrder {
		if d, ok := diffs[field]; ok {
			seen[field] = struct{}{}
			out.WriteString(d)
			out.WriteRune('\n')
		}

		if keys, ok := grouped[field]; ok {
			for _, key := range keys {
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out.WriteString(diffs[key])
				out.WriteRune('\n')
			}
		}
	}

	// Append any remaining items.
	for _, keys := range grouped {
		for _, key := range keys {
			if _, ok := seen[key]; ok {
				continue
			}
			out.WriteString(diffs[key])
			out.WriteRune('\n')
		}
	}

	return cmdutil.DiffOut(out.String())
}

func extractParent(key string) string {
	if idx := strings.IndexAny(key, ".["); idx != -1 {
		return key[:idx]
	}
	return key
}
