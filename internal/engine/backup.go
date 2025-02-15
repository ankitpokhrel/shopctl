package engine

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"

	BackupStatusPending  BackupStatus = "pending"
	BackupStatusRunning  BackupStatus = "running"
	BackupStatusComplete BackupStatus = "complete"
	BackupStatusFailed   BackupStatus = "failed"

	modeDir  = 0o755
	modeFile = 0o644
)

// BackupType represents the type of a backup.
type BackupType string

// BackupStatus is a current status of the initiated backup.
type BackupStatus string

// Backup is a backup engine.
type Backup struct {
	id        string
	store     string
	root      string
	dir       string
	prefix    string
	timestamp time.Time
}

// Option is a functional opt for Backup.
type Option func(*Backup)

// NewBackup creates a new backup engine.
func NewBackup(store string, opts ...Option) *Backup {
	now := time.Now()
	id := genBackupID(store, now.Unix())
	bkp := Backup{
		id:        id,
		store:     store,
		root:      os.TempDir(),
		timestamp: now,
	}

	for _, opt := range opts {
		opt(&bkp)
	}

	if bkp.prefix == "" {
		bkp.prefix = "backup"
	}

	bkp.dir = fmt.Sprintf("%s_%s_%s", bkp.prefix, bkp.timestamp.Format("2006_01_02_15_04_05"), id)
	bkp.root = filepath.Join(bkp.root, bkp.dir)

	return &bkp
}

// WithBackupDir sets root backup dir.
func WithBackupDir(dir string) Option {
	return func(b *Backup) {
		b.root = dir
	}
}

// WithBackupPrefix sets prefix for backup dir name.
func WithBackupPrefix(prefix string) Option {
	return func(b *Backup) {
		b.prefix = prefix
	}
}

// ID returns the backup ID.
func (b *Backup) ID() string {
	return b.id
}

// Store returns the store this backup will run for.
func (b *Backup) Store() string {
	return b.store
}

// Root returns root backup directory.
func (b *Backup) Root() string {
	return b.root
}

// Dir returns backup directory name.
func (b *Backup) Dir() string {
	return b.dir
}

// Timestamp returns backup timestamp.
func (b *Backup) Timestamp() time.Time {
	return b.timestamp
}

// Do starts the backup process.
// Implements `engine.Doer` interface.
func (b *Backup) Do(rs Resource) error {
	dir := filepath.Join(b.root, rs.Path)
	if err := os.MkdirAll(dir, modeDir); err != nil {
		return err
	}
	dest := filepath.Join(dir, rs.Type.File())

	data, err := rs.Handler.Handle()
	if err != nil {
		return err
	}

	return b.saveJSON(dest, data)
}

// saveJSON saves data to a JSON file.
func (b *Backup) saveJSON(path string, data any) error {
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

func genBackupID(store string, timestamp int64) string {
	data := []byte(fmt.Sprintf("%s-%d", store, timestamp))
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:5])
}
