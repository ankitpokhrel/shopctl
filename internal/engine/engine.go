package engine

import (
	"fmt"
	"sync"
)

const (
	// TODO: Decide number of workers.
	numWorkers = 3
)

// ErrSkipChildren is an event that signals the engine to skip processing children.
var ErrSkipChildren = fmt.Errorf("skip processing children")

// Result is a result of the engine operation.
type Result struct {
	ParentResourceType ResourceType
	ResourceType       ResourceType
	Err                error
}

// Doer is an interface that defines the handler of the engine.
type Doer interface {
	Do(Resource, any) (any, error)
}

// Engine is an execution engine.
type Engine struct {
	mux  sync.Mutex
	doer Doer
	jobs map[ResourceType]chan ResourceCollection
}

// New creates a new engine.
func New(d Doer) *Engine {
	return &Engine{
		mux:  sync.Mutex{},
		doer: d,
		jobs: make(map[ResourceType]chan ResourceCollection, 0),
	}
}

// Doer returns the doer.
func (e *Engine) Doer() Doer {
	return e.doer
}

// Register registers a resource type to execute.
func (e *Engine) Register(rt ResourceType) {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.jobs[rt] = make(chan ResourceCollection) // TODO: Decide buffer size based on store size.
}

// Add adds a resource to the engine.
func (e *Engine) Add(rt ResourceType, rc ResourceCollection) {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.jobs[rt] <- rc
}

// Run starts the engine.
func (e *Engine) Run(rt ResourceType) chan Result {
	var wg sync.WaitGroup

	run := func(rc ResourceCollection, out chan<- Result) {
		data, err := e.doer.Do(*rc.Parent, nil)
		out <- Result{ResourceType: rc.Parent.Type, Err: err}
		if err == nil {
			for _, r := range rc.Children {
				_, err := e.doer.Do(r, data)
				out <- Result{ParentResourceType: rc.Parent.Type, ResourceType: r.Type, Err: err}
			}
		}
	}

	out := make(chan Result, numWorkers)
	for range numWorkers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for rc := range e.jobs[rt] {
				run(rc, out)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Done marks the job for a resource type as done
// by closing its receiving channel.
func (e *Engine) Done(rt ResourceType) {
	close(e.jobs[rt])
}
