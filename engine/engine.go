package engine

import "sync"

const (
	// TODO: Decide number of workers.
	numWorkers = 3
)

// Result is a result of the engine operation.
type Result struct {
	ResourceType ResourceType
	Err          error
}

// Doer is an interface that defines the handler of the engine.
type Doer interface {
	Do(Resource) error
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
		for _, r := range rc {
			err := e.doer.Do(r)
			out <- Result{ResourceType: r.Type, Err: err}
		}
	}

	out := make(chan Result, numWorkers)
	for i := 0; i < numWorkers; i++ {
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
