package engine

import (
	"time"
)

// Restore is a restore engine.
type Restore struct {
	timestamp time.Time
}

// NewRestore creates a new restore engine.
func NewRestore() *Restore {
	return &Restore{
		timestamp: time.Now(),
	}
}

// Do starts the restoration process.
// Implements `engine.Doer` interface.
func (r *Restore) Do(rs Resource) error {
	_, err := rs.Handler.Handle()
	return err
}
