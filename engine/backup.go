package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// TODO: Decide number of workers based on store size.
	// We'll collect some store meta on setup.
	numWorkers = 3

	modeDir  = 0o755
	modeFile = 0o644
)

// Result is a result of the backup operation.
type Result struct {
	ResourceType ResourceType
	ResourceID   string
	Err          error
}

// Backup is a backup engine.
type Backup struct {
	dir       string
	jobs      map[ResourceType]chan *ResourceCollection
	timestamp time.Time
}

// Option is a functional opt for Backup.
type Option func(*Backup)

// NewBackup creates a new backup engine.
func NewBackup(opts ...Option) *Backup {
	bkp := Backup{
		jobs:      make(map[ResourceType]chan *ResourceCollection, 0),
		timestamp: time.Now(),
	}

	for _, opt := range opts {
		opt(&bkp)
	}

	if bkp.dir == "" {
		bkp.dir = fmt.Sprintf("backup_%s", bkp.timestamp.Format("2006_01_02_15_04_05"))
	}
	return &bkp
}

// WithBackupDir sets backup dir.
func WithBackupDir(dir string) Option {
	return func(b *Backup) {
		b.dir = dir
	}
}

// Dir returns backup directory.
func (b *Backup) Dir() string {
	return b.dir
}

// Register registers a resource type for backup.
func (b *Backup) Register(rt ResourceType) {
	b.jobs[rt] = make(chan *ResourceCollection, 3) // TODO: Decide buffer size based on store size.
}

// Add adds a resource to the backup job.
func (b *Backup) Add(rt ResourceType, rc *ResourceCollection) {
	b.jobs[rt] <- rc
}

// Do starts the backup process.
func (b *Backup) Do(rt ResourceType) chan Result {
	var wg sync.WaitGroup

	run := func(rc *ResourceCollection, out chan<- Result) {
		for _, r := range rc.Resources {
			err := b.execute(r, rc.Path)
			if err != nil {
				out <- Result{ResourceType: r.Type, ResourceID: rc.RootID, Err: err}
			}
		}
	}

	out := make(chan Result, numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for rc := range b.jobs[rt] {
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
func (b *Backup) Done(rt ResourceType) {
	close(b.jobs[rt])
}

// SaveJSON saves data to a JSON file.
func (b *Backup) SaveJSON(path string, data any) error {
	var (
		jsonData []byte
		err      error
	)

	if jsonData, err = json.Marshal(data); err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	file := fmt.Sprintf("%s.json", path)
	if err := os.WriteFile(file, jsonData, modeFile); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (b *Backup) execute(rsc Resource, path string) error {
	dir := filepath.Join(b.dir, path)
	if err := os.MkdirAll(dir, modeDir); err != nil {
		return err
	}
	dest := filepath.Join(dir, rsc.Type.File())

	data, err := rsc.BackupFn()
	if err != nil {
		return err
	}

	if err := b.SaveJSON(dest, data); err != nil {
		return err
	}
	return nil
}
