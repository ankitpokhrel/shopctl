package config

import (
	"sync"
)

const (
	KeyStatus    = "status"
	KeyResources = "resources"
	KeyTimeStart = "time_start"
	KeyTimeEnd   = "time_end"

	keyID       = "id"
	keyUser     = "user"
	keyTimeInit = "time_initiated"

	metaConfigFile = "metadata"
)

// MetaItems defines item in a metadata file.
type RootMetaItems struct {
	ID        string
	Store     string
	TimeInit  int64
	TimeStart int64
	TimeEnd   int64
	Resources []string
	Kind      string
	Status    string
	User      string
}

// RootMeta is a root metadata for the initiated backup.
type RootMeta struct {
	*config
	mux  *sync.Mutex
	data RootMetaItems
}

// NewRootMeta builds a RootMeta object for the initiated backup.
func NewRootMeta(loc string, items RootMetaItems) *RootMeta {
	return &RootMeta{
		config: newConfig(loc, metaConfigFile, fileTypeJson),
		mux:    &sync.Mutex{},
		data:   items,
	}
}

// Set saves allowed key val to the file.
func (r *RootMeta) Set(keyAndVal map[string]any) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for k, v := range keyAndVal {
		// ID and store could only be set once during meta creation.
		if k != keyID && k != keyStore {
			r.setData(k, v)
			r.writer.Set(k, v)
		}
	}
	return r.writer.WriteConfig()
}

// Save writes metadata to the file.
func (r *RootMeta) Save() error {
	if err := ensureConfigFile(r.dir, metaConfigFile, r.kind, false); err != nil {
		return err
	}
	return r.writeAll()
}

func (r *RootMeta) setData(key string, val any) {
	switch key {
	case KeyStatus:
		r.data.Status = val.(string)
	case KeyResources:
		r.data.Resources = val.([]string)
	case KeyTimeStart:
		r.data.TimeStart = val.(int64)
	case KeyTimeEnd:
		r.data.TimeEnd = val.(int64)
	}
}

func (r *RootMeta) writeAll() error {
	count := 0

	// Prevent replacing ID and store if its already set.
	id := r.writer.GetString(keyID)
	if id == "" {
		count++
		r.writer.Set(keyID, r.data.ID)
	}
	store := r.writer.GetString(keyStore)
	if store == "" {
		count++
		r.writer.Set(keyStore, r.data.Store)
	}

	if r.data.Status != "" {
		count++
		r.writer.Set(KeyStatus, r.data.Status)
	}
	if len(r.data.Resources) > 0 {
		count++
		r.writer.Set(KeyResources, r.data.Resources)
	}
	if r.data.Kind != "" {
		count++
		r.writer.Set(keyKind, r.data.Kind)
	}
	if r.data.TimeInit != 0 {
		count++
		r.writer.Set(keyTimeInit, r.data.TimeInit)
	}
	if r.data.TimeStart != 0 {
		count++
		r.writer.Set(KeyTimeStart, r.data.TimeStart)
	}
	if r.data.TimeEnd != 0 {
		count++
		r.writer.Set(KeyTimeEnd, r.data.TimeEnd)
	}
	if r.data.User != "" {
		count++
		r.writer.Set(keyUser, r.data.User)
	}

	if count > 0 {
		return r.writer.WriteConfig()
	}
	return nil
}
