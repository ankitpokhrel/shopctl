package runner

import "github.com/ankitpokhrel/shopctl/internal/engine"

// Runner is a runner interface.
type Runner interface {
	Run() error
	Kind() engine.ResourceType
}
