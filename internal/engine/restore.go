package engine

import (
	"time"
)

// Restore is a restore engine.
type Restore struct {
	store     string
	timestamp time.Time
}

// NewRestore creates a new restore engine.
func NewRestore(store string) *Restore {
	return &Restore{
		store:     store,
		timestamp: time.Now(),
	}
}

// Do starts the restoration process.
// Implements `engine.Doer` interface.
func (r *Restore) Do(rs Resource, data any) (any, error) {
	data, err := rs.Handler.Handle(data)
	return data, err
}
