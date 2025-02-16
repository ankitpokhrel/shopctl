package runner

import (
	"fmt"

	"github.com/ankitpokhrel/shopctl/internal/engine"
)

// Runner is a runner interface.
type Runner interface {
	Run() error
	Kind() engine.ResourceType
	Stats() string
}

// Summary aggregate runner stats.
type Summary struct {
	Count   int
	Passed  int
	Failed  int
	Skipped int
}

// String implements `fmt.Stringer` interface.
func (s Summary) String() string {
	return fmt.Sprintf(`Processed: %d
Succeeded: %d
Skipped: %d
Failed: %d`,
		s.Count, s.Passed,
		s.Skipped, s.Failed,
	)
}
