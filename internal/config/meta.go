package config

import (
	"sync"
)

const (
	KeyStatus    = "status"
	KeyResources = "resources"
	KeyTimeStart = "timeStart"
	KeyTimeEnd   = "timeEnd"

	metaConfigFile = "metadata"
)

// MetaItems defines item in a metadata file.
type RootMetaItems struct {
	ID        string   `koanf:"id" json:"id"`
	Store     string   `koanf:"store" json:"store"`
	TimeInit  int64    `koanf:"timeInitiated" json:"timeInitiated"`
	TimeStart int64    `koanf:"timeStart" json:"timeStart"`
	TimeEnd   int64    `koanf:"timeEnd" json:"timeEnd"`
	Resources []string `koanf:"resources" json:"resources"`
	Kind      string   `koanf:"type" json:"type"`
	Status    string   `koanf:"status" json:"status"`
	User      string   `koanf:"user" json:"user"`
}

// RootMeta is a root metadata for the initiated backup.
type RootMeta struct {
	*config
	mux  *sync.Mutex
	data RootMetaItems
}

// NewRootMeta builds a RootMeta object for the initiated backup.
func NewRootMeta(loc string, items RootMetaItems) (*RootMeta, error) {
	cfg, err := newConfig(loc, metaConfigFile, fileTypeJson)
	if err != nil {
		return nil, err
	}

	return &RootMeta{
		config: cfg,
		mux:    &sync.Mutex{},
		data:   items,
	}, nil
}

// Set saves allowed key val to the file.
func (r *RootMeta) Set(keyAndVal map[string]any) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for k, v := range keyAndVal {
		r.setData(k, v)
	}
	return writeJSON(r.path, r.data)
}

// Save writes metadata to the file.
func (r *RootMeta) Save() error {
	return writeJSON(r.path, r.data)
}

func (r *RootMeta) setData(key string, val any) {
	// Some keys like ID, store and time initiated can only
	// be set once during meta creation so we skip them.
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
