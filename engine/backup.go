package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"

	modeDir  = 0o755
	modeFile = 0o644
)

// BackupType represents the type of a backup.
type BackupType string

// Backup is a backup engine.
type Backup struct {
	dir       string
	timestamp time.Time
}

// Option is a functional opt for Backup.
type Option func(*Backup)

// NewBackup creates a new backup engine.
func NewBackup(opts ...Option) *Backup {
	bkp := Backup{
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

// Do starts the backup process.
// Implements `engine.Doer` interface.
func (b *Backup) Do(rs Resource) error {
	dir := filepath.Join(b.dir, rs.Path)
	if err := os.MkdirAll(dir, modeDir); err != nil {
		return err
	}
	dest := filepath.Join(dir, rs.Type.File())

	data, err := rs.Handler.Handle()
	if err != nil {
		return err
	}

	if err := b.saveJSON(dest, data); err != nil {
		return err
	}
	return nil
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
